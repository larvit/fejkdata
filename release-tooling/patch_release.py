#!/usr/bin/env python3
"""Own CHANGELOG.md's release sections: cut a patch of the last release carrying main's dependencies, and keep main in step.

  patch_release.py last
      Prints the last release's tag, as the forge names its latest release.
  patch_release.py cut
      When main carries a newer Go, other modules or another GoReleaser than the last release,
      and govulncheck finds nothing in the last release's tree with them: commits them, main's
      .github/ and a changelog section to a branch release/X.Y.Z cut from the last release's
      tag, builds it on the latest and the lowest supported Go, and publishes it with
      publish_release.py. Writes `tag=vX.Y.Z` to GITHUB_OUTPUT. While that tree still carries a
      vulnerability, it warns and cuts nothing.
  patch_release.py plan <govulncheck report of main> <todo.md> <CHANGELOG.md>
      Adds to CHANGELOG.md each release section the last release has and main lacks, plans
      the report in todo.md through file_vulnerabilities.py, and records each fixed item's
      line under Unreleased unless a release section already names its vulnerabilities.

Run in a checkout of main, with the forge's env set as for publish_release.py and
FORGE_SERVER_URL naming the forge's web host, which a push goes to. Exits 2 on
input or a step it cannot complete, naming it.
"""

import contextlib
import datetime
import os
import re
import shutil
import subprocess
import sys
import tempfile
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from typing import Callable, NamedTuple

import publish_release

GO_VERSION = re.compile(r"^ARG GO_VERSION=(\d+)\.(\d+)\.(\d+)$", re.M)
GO_LINE = re.compile(r"^go (\S+)$", re.M)
GORELEASER = re.compile(r"goreleaser/goreleaser-action@\S+\s+with:\s+version: (\S+)")
GORELEASER_PIN = re.compile(r"goreleaser/goreleaser:v\S+")
TAG = re.compile(r"v(\d+)\.(\d+)\.(\d+)")
VULNERABILITY = re.compile(r"GO-\d+-\d+")
FILER = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "vulnerability-filing", "file_vulnerabilities.py")
GO_NOTES = "https://raw.githubusercontent.com/golang/website/master/_content/doc/go{}.md"
BUILT_FOR = ("Darwin", "Linux", "Windows")
"""The systems goreleaser.yaml builds binaries for, as Go's release notes head their Ports."""


@dataclass(frozen=True)
class Carried:
	"""What a release carries, as one tree names it: Go, the modules go.mod requires, and GoReleaser."""

	go: tuple[int, int, int]
	goreleaser: str | None
	go_mod: str
	go_sum: str


def carried(show: Callable[[str], str]) -> Carried:
	"""What a tree carries, read through show(path), which returns "" for a file the tree lacks."""
	packages = GORELEASER.search(show(".github/workflows/packages.yml"))
	return Carried(go_of(show("Dockerfile")), packages and packages[1], show("go.mod"), show("go.sum"))


def go_of(dockerfile: str) -> tuple[int, int, int]:
	match = GO_VERSION.search(dockerfile)
	if match is None:
		raise ValueError("a Dockerfile has no `ARG GO_VERSION=X.Y.Z` line")
	return tuple(map(int, match.groups()))


def dotted(go: tuple[int, int, int]) -> str:
	return ".".join(map(str, go))


def go_line(go_mod: str) -> str | None:
	"""The lowest supported Go, as go.mod's `go` line names it."""
	match = GO_LINE.search(go_mod)
	return match and match[1]


def requires(go_mod: str) -> dict[str, str]:
	"""The modules go.mod requires, by path, each at its version."""
	found, block = {}, False
	for line in go_mod.splitlines():
		words = line.split("//")[0].split()
		if words == ["require", "("] or words == [")"]:
			block = words != [")"]
		elif words[:1] == ["require"] and len(words) == 3:
			found[words[1]] = words[2]
		elif block and len(words) == 2:
			found[words[0]] = words[1]
	return found


def next_patch(tag: str) -> str:
	match = TAG.fullmatch(tag)
	if match is None:
		raise ValueError(f"the last release's tag {tag!r} is not of the form vX.Y.Z")
	major, minor, patch = match.groups()
	return f"{major}.{minor}.{int(patch) + 1}"


def due(released: Carried, main: Carried) -> Carried | None:
	"""What a patch of the release carries, main's wherever main moved on, or None while that is what the release carries.

	main's modules wait while its go.mod raises the lowest supported Go, which only a major may do.
	"""
	modules = requires(main.go_mod) != requires(released.go_mod) and go_line(main.go_mod) == go_line(released.go_mod)
	patch = Carried(
		max(released.go, main.go),
		main.goreleaser,
		main.go_mod if modules else released.go_mod,
		main.go_sum if modules else released.go_sum,
	)
	return None if patch == released else patch


def platforms_dropped(notes: str, minor: str) -> list[str]:
	"""Each requirement Go `minor`'s release notes set on a system the binaries are built for, as the notes state it."""
	ports = re.search(r"^## Ports\b[^\n]*\n(.*?)(?=^## |\Z)", notes, re.M | re.S)
	if ports is None:
		return []
	found = []
	for system in re.finditer(r"^### (\w+)[^\n]*\n(.*?)(?=^### |\Z)", ports[1], re.M | re.S):
		if system[1] in BUILT_FOR:
			text = re.sub(r"\[([^\]]*)\]\([^)]*\)", r"\1", re.sub(r"<!--.*?-->", "", system[2], flags=re.S))
			found += re.findall(rf"\bGo {re.escape(minor)} requires [^;]+?(?=;|\.(?:\s|$))", " ".join(text.split()))
	return found


def platform_lines(old: tuple[int, int, int], new: tuple[int, int, int], go_notes: Callable[[str], str]) -> list[str]:
	"""A changelog line for each requirement a Go minor above old, up to new, sets on a system the binaries are built for."""
	minors = [f"{new[0]}.{minor}" for minor in range(old[1] + 1, new[1] + 1)] if old[0] == new[0] else []
	return [f"{line}, and so do the binaries." for minor in minors for line in platforms_dropped(go_notes(minor), minor)]


def changes(released: Carried, patch: Carried, platforms: list[str]) -> list[str]:
	"""The changelog's lines for what the patch carries that the release did not."""
	lines = []
	if patch.go != released.go:
		lines += [f"The binaries and image are built with Go {dotted(patch.go)}.", *platforms]
	if patch.goreleaser != released.goreleaser:
		lines.append(f"The binaries, packages and image are built with GoReleaser {patch.goreleaser}.")
	old, new = requires(released.go_mod), requires(patch.go_mod)
	for path in sorted(old.keys() | new.keys()):
		if path not in new:
			lines.append(f"The binaries no longer carry {path}.")
		elif path not in old:
			lines.append(f"The binaries carry {path} {new[path]}.")
		elif old[path] != new[path]:
			lines.append(f"The binaries carry {path} {new[path]}, up from {old[path]}.")
	return lines


def set_go(text: str, version: str) -> str:
	changed = re.sub(r"^ARG GO_VERSION=\S+$", f"ARG GO_VERSION={version}", text, flags=re.M)
	changed = re.sub(r"golang:\$\{GO_VERSION:-[^}]+\}", f"golang:${{GO_VERSION:-{version}}}", changed)
	if changed == text:
		raise ValueError("a file names no Go version to set")
	return changed


def section(version: str, date: str, changed: list[str], last: str, fixed: str) -> str:
	text = f"## [{version}] - {date}\n\n### Changed\n\n" + "".join(f"- {line}\n" for line in changed) + "\n"
	if fixed:
		text += f"### Security\n\n- The binaries no longer carry {last.removeprefix('v')}'s known vulnerabilities: {fixed}.\n\n"
	return text


def releases(changelog: str) -> list[re.Match]:
	"""The changelog's release sections, newest first."""
	return [m for m in publish_release.SECTION.finditer(changelog) if publish_release.VERSION.fullmatch(m["name"])]


def insert_section(changelog: str, text: str) -> str:
	"""The changelog with the section above its newest release's, so below any Unreleased section."""
	newest = releases(changelog)
	at = newest[0].start() if newest else len(changelog)
	return changelog[:at] + text + changelog[at:]


def reconcile(main: str, released: str) -> str:
	"""Main's changelog with each release section the released changelog has and main lacks."""
	have = {m["name"] for m in releases(main)}
	for missing in reversed([m for m in releases(released) if m["name"] not in have]):
		main = insert_section(main, missing[0])
	return main


def record(changelog: str, fixed: list[str]) -> str:
	"""The changelog with each line under Unreleased's Security, save one whose vulnerabilities a release section names."""
	released = set(VULNERABILITY.findall("".join(m[0] for m in releases(changelog))))
	new = [line for line in fixed if not set(VULNERABILITY.findall(line)) <= released]
	if not new:
		return changelog
	bullets = "".join(f"- {line}\n" for line in new)
	unreleased = next((m for m in publish_release.SECTION.finditer(changelog) if m["name"] == "Unreleased"), None)
	if unreleased is None:
		return insert_section(changelog, f"## [Unreleased]\n\n### Security\n\n{bullets}\n")
	body = unreleased["body"].strip("\n")
	if "### Security" in body:
		body += "\n" + bullets
	else:
		body += ("\n\n" if body else "") + "### Security\n\n" + bullets
	end = unreleased.end()
	tail = "\n\n" + changelog[end:] if end < len(changelog) else "\n"
	return changelog[: unreleased.start("body")] + "\n" + body.rstrip("\n") + tail


def read(path: str) -> str:
	try:
		with open(path, encoding="utf-8") as f:
			return f.read()
	except OSError as err:
		raise ValueError(f"cannot read {path}: {err.strerror}") from err


def read_if_any(path: str) -> str:
	return read(path) if os.path.exists(path) else ""


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


def carry_workflows(checkout: str) -> None:
	"""Gives the release tree main's .github/, since the forge refuses the workflow token a push whose workflows differ from main's."""
	try:
		shutil.rmtree(os.path.join(checkout, ".github"))
		shutil.copytree(".github", os.path.join(checkout, ".github"))
	except OSError as err:
		raise ValueError(f"cannot carry main's .github/ into the release tree: {err}") from err


def prepare(checkout: str, released: Carried, patch: Carried) -> None:
	"""Puts the patch's Go, modules and GoReleaser pins into the release tree; packages.yml's GoReleaser comes with main's .github/."""
	if patch.go != released.go:
		for name in ("Dockerfile", "compose.yaml"):
			path = os.path.join(checkout, name)
			write(path, set_go(read(path), dotted(patch.go)))
	if patch.go_mod != released.go_mod:
		write(os.path.join(checkout, "go.mod"), patch.go_mod)
		if patch.go_sum:
			write(os.path.join(checkout, "go.sum"), patch.go_sum)
		else:
			with contextlib.suppress(FileNotFoundError):
				os.remove(os.path.join(checkout, "go.sum"))
	if patch.goreleaser != released.goreleaser:
		for name in ("Dockerfile", "release-tooling/package-repos.Dockerfile"):
			path = os.path.join(checkout, name)
			if os.path.exists(path):
				write(path, GORELEASER_PIN.sub(f"goreleaser/goreleaser:{patch.goreleaser}", read(path)))


def ship(checkout: str, version: str) -> str:
	"""Builds the release tree through the gate on the latest and the lowest supported Go, commits and pushes it, and returns its commit."""
	out("docker", "build", ".", cwd=checkout)
	lowest = go_line(read(os.path.join(checkout, "go.mod")))
	out("docker", "build", "--build-arg", f"GO_VERSION={lowest}", "--target", "portable", ".", cwd=checkout)
	out("git", "add", "-A", cwd=checkout)
	out("git", "commit", "-qm", f"Release {version}", cwd=checkout)
	server = urllib.parse.urlsplit(os.environ["FORGE_SERVER_URL"])
	remote = f"{server.scheme}://x-access-token:{os.environ['FORGE_TOKEN']}@{server.netloc}/{os.environ['FORGE_REPOSITORY']}.git"
	# Delete a branch a failed run left, so the push creates it and the forge checks it against main.
	run("git", "push", "-q", remote, f":refs/heads/release/{version}", cwd=checkout, ok=(0, 1))
	out("git", "push", "-q", remote, f"HEAD:refs/heads/release/{version}", cwd=checkout)
	return out("git", "rev-parse", "HEAD", cwd=checkout)


def go_notes(minor: str) -> str:
	try:
		with urllib.request.urlopen(GO_NOTES.format(minor)) as resp:
			return resp.read().decode()
	except urllib.error.URLError as err:
		raise ValueError(f"cannot read Go {minor}'s release notes: {err}") from err


class Effects(NamedTuple):
	"""What `cut` does outside the checkout, each a function a test can stand in for."""

	scan: Callable[[str, str, str], str]
	ship: Callable[[str, str], str]
	publish: Callable[[str, str], tuple[int, str | None]]
	output: Callable[[str], None]
	go_notes: Callable[[str], str]


def cut(last: str, date: str, effects: Effects) -> int:
	"""Cuts the patch of the last release that main's Go, modules and GoReleaser call for, if any; run in main's checkout."""
	released = carried(lambda path: run("git", "show", f"{last}:{path}", ok=(0, 128)).stdout)
	patch = due(released, carried(read_if_any))
	if patch is None:
		print(f"No patch due: main carries the Go, modules and GoReleaser {last} does.")
		return 0
	version = next_patch(last)
	checkout, scans = tempfile.mkdtemp(), tempfile.mkdtemp()
	out("git", "worktree", "add", "-q", "--detach", checkout, last)
	before = effects.scan(checkout, dotted(released.go), os.path.join(scans, "before"))
	prepare(checkout, released, patch)
	after = effects.scan(checkout, dotted(patch.go), os.path.join(scans, "after"))
	if run(sys.executable, FILER, "check", after, ok=(0, 1)).returncode != 0:
		print(f"::warning::a patch of {last} on main's Go, modules and GoReleaser still carries a known vulnerability; no patch is cut")
		return 0
	fixed = out(sys.executable, FILER, "fixed-between", before, after)
	changed = changes(released, patch, platform_lines(released.go, patch.go, effects.go_notes))
	changelog = os.path.join(checkout, "CHANGELOG.md")
	write(changelog, insert_section(read(changelog), section(version, date, changed, last, fixed)))
	carry_workflows(checkout)
	code, tag = effects.publish(read(changelog), effects.ship(checkout, version))
	if tag:
		effects.output(f"tag={tag}")
	return code


def plan(report: str, todo: str, changelog: str, last: str) -> int:
	text = reconcile(read(changelog), out("git", "show", f"{last}:CHANGELOG.md"))
	planned = run(sys.executable, FILER, "plan", report, todo, "--patch", next_patch(last))
	sys.stderr.write(planned.stderr)
	write(changelog, record(text, planned.stdout.splitlines()))
	return 0


def publish(changelog: str, sha: str) -> tuple[int, str | None]:
	try:
		return publish_release.publish(changelog, sha)
	except urllib.error.URLError as err:
		raise ValueError(f"cannot publish the release: {err}") from err


def github_output(line: str) -> None:
	with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as f:
		f.write(line + "\n")


def main(argv: list[str]) -> int:
	try:
		match argv:
			case ["last"]:
				print(last_release())
				return 0
			case ["cut"]:
				today = datetime.datetime.now(datetime.UTC).date().isoformat()
				return cut(last_release(), today, Effects(scan, ship, publish, github_output, go_notes))
			case ["plan", report, todo, changelog]:
				read(report)
				return plan(report, todo, changelog, last_release())
	except ValueError as err:
		print(err, file=sys.stderr)
		return 2
	print(__doc__, file=sys.stderr)
	return 2


if __name__ == "__main__":
	sys.exit(main(sys.argv[1:]))
