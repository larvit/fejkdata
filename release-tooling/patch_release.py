#!/usr/bin/env python3
"""Cut a patch of the last release when main's Go is a patch above the last release's.

  patch_release.py cut <govulncheck report> <todo.md>
      In a checkout of main, with the forge's env set as for publish_release.py: when main's
      Go is a patch above the last release's, commits that Go version and a changelog section
      to a branch release/X.Y.Z cut from the last release's tag, builds it through the gate's
      `docker build .`, and publishes it with publish_release.py. Prints what it did.
  patch_release.py last
      Prints the last release's tag, as the forge names its latest release.
  patch_release.py reconcile <main's CHANGELOG.md> <the last release's CHANGELOG.md>
      Adds to main's CHANGELOG.md each release section the last release has and main lacks.

Exits 2 on input it cannot read.
"""

import datetime
import os
import re
import subprocess
import sys
import tempfile

import publish_release

GO_VERSION = re.compile(r"^ARG GO_VERSION=(\d+)\.(\d+)\.(\d+)$", re.M)
TAG = re.compile(r"v(\d+)\.(\d+)\.(\d+)")
SECTION = re.compile(r"^## \[(?P<version>\d+\.\d+\.\d+)\][^\n]*\n.*?(?=^## \[|\Z)", re.M | re.S)
FILER = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "vulnerability-filing", "file_vulnerabilities.py")


def go_of(dockerfile: str) -> tuple[int, int, int]:
	match = GO_VERSION.search(dockerfile)
	if match is None:
		raise ValueError("a Dockerfile has no `ARG GO_VERSION=X.Y.Z` line")
	return tuple(map(int, match.groups()))


def due(last_tag: str, released: str, main: str) -> tuple[str, str] | None:
	"""The patch release's version and Go version when main's Go is a patch above the last release's."""
	tag = TAG.fullmatch(last_tag)
	if tag is None:
		raise ValueError(f"the last release's tag {last_tag!r} is not of the form vX.Y.Z")
	old, new = go_of(released), go_of(main)
	if old[:2] != new[:2] or new[2] <= old[2]:
		return None
	major, minor, patch = tag.groups()
	return f"{major}.{minor}.{int(patch) + 1}", ".".join(map(str, new))


def set_go(text: str, version: str) -> str:
	changed = re.sub(r"^ARG GO_VERSION=\S+$", f"ARG GO_VERSION={version}", text, flags=re.M)
	changed = re.sub(r"golang:\$\{GO_VERSION:-[^}]+\}", f"golang:${{GO_VERSION:-{version}}}", changed)
	if changed == text:
		raise ValueError("a file names no Go version to set")
	return changed


def section(version: str, date: str, go: str, fixed: list[str]) -> str:
	lines = [f"The binaries and image are built with Go {go}.", *fixed]
	return f"## [{version}] - {date}\n\n### Security\n\n" + "".join(f"- {line}\n" for line in lines) + "\n"


def insert_section(changelog: str, text: str) -> str:
	"""The changelog with the section above its newest release's, so below any Unreleased section."""
	first = SECTION.search(changelog)
	at = first.start() if first else len(changelog)
	return changelog[:at] + text + changelog[at:]


def reconcile(main: str, released: str) -> str:
	"""Main's changelog with each release section the released changelog has and main lacks."""
	have = {m["version"] for m in SECTION.finditer(main)}
	for missing in reversed([m for m in SECTION.finditer(released) if m["version"] not in have]):
		main = insert_section(main, missing[0])
	return main


def run(*argv: str, cwd: str | None = None) -> str:
	return subprocess.run(argv, cwd=cwd, check=True, capture_output=True, text=True).stdout.strip()


def last_release() -> str:
	return publish_release.request("releases/latest")["tag_name"]


def cut(report: str, todo: str) -> int:
	last = last_release()
	with open("Dockerfile", encoding="utf-8") as f:
		plan = due(last, run("git", "show", f"{last}:Dockerfile"), f.read())
	if plan is None:
		print(f"No patch due: main's Go is no patch above {last}'s.")
		return 0
	version, go = plan
	fixed = run(sys.executable, FILER, "fixed", report, todo).splitlines()
	checkout = tempfile.mkdtemp()
	run("git", "worktree", "add", "-q", "--detach", checkout, last)
	for name in ("Dockerfile", "compose.yaml"):
		path = os.path.join(checkout, name)
		with open(path, encoding="utf-8") as f:
			text = set_go(f.read(), go)
		with open(path, "w", encoding="utf-8") as f:
			f.write(text)
	path = os.path.join(checkout, "CHANGELOG.md")
	with open(path, encoding="utf-8") as f:
		text = insert_section(f.read(), section(version, datetime.date.today().isoformat(), go, fixed))
	with open(path, "w", encoding="utf-8") as f:
		f.write(text)
	subprocess.run(["docker", "build", "."], cwd=checkout, check=True)
	run("git", "commit", "-qam", f"Release {version}, built with Go {go}", cwd=checkout)
	remote = f"https://x-access-token:{os.environ['FORGE_TOKEN']}@github.com/{os.environ['FORGE_REPOSITORY']}.git"
	run("git", "push", "-qf", remote, f"HEAD:refs/heads/release/{version}", cwd=checkout)
	os.environ["SHA"] = run("git", "rev-parse", "HEAD", cwd=checkout)
	os.chdir(checkout)
	code, tag = publish_release.publish()
	if tag and "GITHUB_OUTPUT" in os.environ:
		with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as f:
			f.write(f"tag={tag}\n")
	return code


def main(argv: list[str]) -> int:
	try:
		match argv:
			case ["last"]:
				print(last_release())
				return 0
			case ["cut", report, todo]:
				return cut(report, todo)
			case ["reconcile", main_path, released_path]:
				with open(main_path, encoding="utf-8") as m, open(released_path, encoding="utf-8") as r:
					text = reconcile(m.read(), r.read())
				with open(main_path, "w", encoding="utf-8") as f:
					f.write(text)
				return 0
	except ValueError as err:
		print(err, file=sys.stderr)
		return 2
	print(__doc__, file=sys.stderr)
	return 2


if __name__ == "__main__":
	sys.exit(main(sys.argv[1:]))
