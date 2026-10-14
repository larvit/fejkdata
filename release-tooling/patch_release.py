#!/usr/bin/env python3
"""Own CHANGELOG.md's release sections: cut a Go patch of the last release, and keep main in step.

  patch_release.py last
      Prints the last release's tag, as the forge names its latest release.
  patch_release.py cut
      When main's Go is a patch above the last release's, and govulncheck finds nothing in the
      last release's tree on that Go: commits the Go version and a changelog section to a
      branch release/X.Y.Z cut from the last release's tag, builds it through `docker build .`,
      and publishes it with publish_release.py. Writes `tag=vX.Y.Z` to GITHUB_OUTPUT.
  patch_release.py plan <govulncheck report of main> <todo.md> <CHANGELOG.md>
      Adds to CHANGELOG.md each release section the last release has and main lacks, plans
      the report in todo.md through file_vulnerabilities.py, and records each fixed item's
      line under Unreleased unless a release section already names its vulnerabilities.

Run in a checkout of main, with the forge's env set as for publish_release.py. Exits 2 on
input or a step it cannot complete, naming it.
"""

import datetime
import os
import re
import subprocess
import sys
import tempfile
import urllib.error

import publish_release

GO_VERSION = re.compile(r"^ARG GO_VERSION=(\d+)\.(\d+)\.(\d+)$", re.M)
TAG = re.compile(r"v(\d+)\.(\d+)\.(\d+)")
SECTION = re.compile(r"^## \[(?P<version>\d+\.\d+\.\d+)\][^\n]*\n.*?(?=^## \[|\Z)", re.M | re.S)
UNRELEASED = "## [Unreleased]\n"
VULNERABILITY = re.compile(r"GO-\d+-\d+")
FILER = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "vulnerability-filing", "file_vulnerabilities.py")


def go_of(dockerfile: str) -> tuple[int, int, int]:
	match = GO_VERSION.search(dockerfile)
	if match is None:
		raise ValueError("a Dockerfile has no `ARG GO_VERSION=X.Y.Z` line")
	return tuple(map(int, match.groups()))


def next_patch(tag: str) -> str:
	match = TAG.fullmatch(tag)
	if match is None:
		raise ValueError(f"the last release's tag {tag!r} is not of the form vX.Y.Z")
	major, minor, patch = match.groups()
	return f"{major}.{minor}.{int(patch) + 1}"


def due(last_tag: str, released: str, main: str) -> tuple[str, str] | None:
	"""The patch release's version and Go version when main's Go is a patch above the last release's."""
	version = next_patch(last_tag)
	old, new = go_of(released), go_of(main)
	if old[:2] != new[:2] or new[2] <= old[2]:
		return None
	return version, ".".join(map(str, new))


def set_go(text: str, version: str) -> str:
	changed = re.sub(r"^ARG GO_VERSION=\S+$", f"ARG GO_VERSION={version}", text, flags=re.M)
	changed = re.sub(r"golang:\$\{GO_VERSION:-[^}]+\}", f"golang:${{GO_VERSION:-{version}}}", changed)
	if changed == text:
		raise ValueError("a file names no Go version to set")
	return changed


def section(version: str, date: str, go: str, old_go: str, fixed: str) -> str:
	lines = [f"The binaries and image are built with Go {go}."]
	if fixed:
		lines.append(f"The binaries no longer carry Go {old_go}'s known vulnerabilities: {fixed}.")
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


def record(changelog: str, fixed: list[str]) -> str:
	"""The changelog with each line under Unreleased's Security, save one whose vulnerabilities a release section names."""
	released = set(VULNERABILITY.findall("".join(m[0] for m in SECTION.finditer(changelog))))
	new = [line for line in fixed if not set(VULNERABILITY.findall(line)) <= released]
	if not new:
		return changelog
	bullets = "".join(f"- {line}\n" for line in new)
	if UNRELEASED not in changelog:
		return insert_section(changelog, f"{UNRELEASED}\n### Security\n\n{bullets}\n")
	start = changelog.index(UNRELEASED) + len(UNRELEASED)
	following = re.compile(r"^## \[", re.M).search(changelog, start)
	end = following.start() if following else len(changelog)
	body = changelog[start:end].strip("\n")
	if "### Security" in body:
		body += "\n" + bullets
	else:
		body += ("\n\n" if body else "") + "### Security\n\n" + bullets
	tail = "\n\n" + changelog[end:] if following else "\n"
	return changelog[:start] + "\n" + body.rstrip("\n") + tail


def read(path: str) -> str:
	try:
		with open(path, encoding="utf-8") as f:
			return f.read()
	except OSError as err:
		raise ValueError(f"cannot read {path}: {err.strerror}") from err


def write(path: str, text: str) -> None:
	with open(path, "w", encoding="utf-8") as f:
		f.write(text)


def run(*argv: str, cwd: str | None = None, ok: tuple[int, ...] = (0,)) -> subprocess.CompletedProcess:
	done = subprocess.run(argv, cwd=cwd, capture_output=True, text=True)
	if done.returncode not in ok:
		raise ValueError(f"`{' '.join(argv[:2])}` failed: {(done.stderr or done.stdout).strip()}")
	return done


def out(*argv: str, cwd: str | None = None) -> str:
	return run(*argv, cwd=cwd).stdout.strip()


def last_release() -> str:
	try:
		return publish_release.request("releases/latest")["tag_name"]
	except (urllib.error.URLError, KeyError) as err:
		raise ValueError(f"cannot read the last release from the forge: {err}") from err


def scan(tree: str, go: str, into: str) -> str:
	"""Govulncheck's report on the tree built with that Go, through main's Dockerfile."""
	out(
		"docker", "build", "-f", os.path.abspath("Dockerfile"), "--no-cache-filter", "vulnerabilities",
		"--build-arg", f"GO_VERSION={go}", "--target", "vulnerability-report", "--output", into, tree,
	)
	return os.path.join(into, "report.json")


def prepare(checkout: str, go: str, text: str) -> None:
	"""Sets the release tree's Go and puts the section atop its changelog."""
	for name in ("Dockerfile", "compose.yaml"):
		path = os.path.join(checkout, name)
		write(path, set_go(read(path), go))
	path = os.path.join(checkout, "CHANGELOG.md")
	write(path, insert_section(read(path), text))


def ship(checkout: str, version: str, go: str) -> str:
	"""Builds the release tree through the gate, commits and pushes it, and returns its commit."""
	out("docker", "build", ".", cwd=checkout)
	out("git", "commit", "-qam", f"Release {version}, built with Go {go}", cwd=checkout)
	remote = f"https://x-access-token:{os.environ['FORGE_TOKEN']}@github.com/{os.environ['FORGE_REPOSITORY']}.git"
	out("git", "push", "-qf", remote, f"HEAD:refs/heads/release/{version}", cwd=checkout)
	return out("git", "rev-parse", "HEAD", cwd=checkout)


def cut() -> int:
	last = last_release()
	released = out("git", "show", f"{last}:Dockerfile")
	plan = due(last, released, read("Dockerfile"))
	if plan is None:
		print(f"No patch due: main's Go is no patch above {last}'s.")
		return 0
	version, go = plan
	old_go = ".".join(map(str, go_of(released)))
	checkout, scans = tempfile.mkdtemp(), tempfile.mkdtemp()
	out("git", "worktree", "add", "-q", "--detach", checkout, last)
	before = scan(checkout, old_go, os.path.join(scans, "before"))
	after = scan(checkout, go, os.path.join(scans, "after"))
	if run(sys.executable, FILER, "check", after, ok=(0, 1)).returncode != 0:
		raise ValueError(f"{last}'s tree on Go {go} still carries a known vulnerability; no patch is cut")
	fixed = out(sys.executable, FILER, "fixed-between", before, after)
	date = datetime.datetime.now(datetime.UTC).date().isoformat()
	prepare(checkout, go, section(version, date, go, old_go, fixed))
	sha = ship(checkout, version, go)
	code, tag = publish_release.publish(read(os.path.join(checkout, "CHANGELOG.md")), sha)
	if tag:
		with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as f:
			f.write(f"tag={tag}\n")
	return code


def plan(report: str, todo: str, changelog: str) -> int:
	read(report)
	text = read(changelog)
	last = last_release()
	text = reconcile(text, out("git", "show", f"{last}:CHANGELOG.md"))
	planned = run(sys.executable, FILER, "plan", report, todo, "--patch", next_patch(last))
	sys.stderr.write(planned.stderr)
	write(changelog, record(text, planned.stdout.splitlines()))
	return 0


def main(argv: list[str]) -> int:
	try:
		match argv:
			case ["last"]:
				print(last_release())
				return 0
			case ["cut"]:
				return cut()
			case ["plan", report, todo, changelog]:
				return plan(report, todo, changelog)
	except ValueError as err:
		print(err, file=sys.stderr)
		return 2
	print(__doc__, file=sys.stderr)
	return 2


if __name__ == "__main__":
	sys.exit(main(sys.argv[1:]))
