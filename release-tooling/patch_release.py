#!/usr/bin/env python3
"""Own CHANGELOG.md's release sections: cut a patch of the last release carrying what main carries, and keep main in step.

  patch_release.py last
      Prints the last release's tag, as the forge names its latest release.
  patch_release.py cut
      When main carries a newer Go, other modules or another GoReleaser than the last release:
      puts them in the last release's tree, and commits it with main's .github/ and a changelog
      section to a branch release/X.Y.Z cut from the last release's tag, then publishes it with
      publish_release.py. Writes `tag=vX.Y.Z` to GITHUB_OUTPUT. It warns and cuts nothing
      while govulncheck finds a vulnerability in that tree, while the tree fails `docker build .`,
      and while a new Go minor's release notes say something about a system the binaries are
      built for that it cannot read.
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
import functools
import os
import re
import shutil
import subprocess
import sys
import tempfile
import urllib.error
import urllib.parse
from typing import Callable, NamedTuple

import carried
import changelog
import go_release_notes
import publish_release

TAG = re.compile(r"v(\d+)\.(\d+)\.(\d+)")
FILER = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "vulnerability-filing", "file_vulnerabilities.py")


class Hold(Exception):
	"""Why the patch waits: `cut` warns with it and cuts nothing."""


class Effects(NamedTuple):
	"""The steps of `cut` that run docker, read the network or publish, each a function a test can stand in for."""

	scan: Callable[[str, str, str], str]
	gate: Callable[[str], None]
	ship: Callable[[str, str], str]
	publish: Callable[[str, str], tuple[int, str | None]]
	output: Callable[[str], None]
	go_notes: Callable[[str], str]


def next_patch(tag: str) -> str:
	match = TAG.fullmatch(tag)
	if match is None:
		raise ValueError(f"the last release's tag {tag!r} is not of the form vX.Y.Z")
	major, minor, patch = match.groups()
	return f"{major}.{minor}.{int(patch) + 1}"


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


def scan(main: str, tree: str, go: str, into: str) -> str:
	"""Govulncheck's report on the tree built with that Go, through main's Dockerfile."""
	out(
		"docker", "build", "-f", os.path.join(main, "Dockerfile"), "--no-cache-filter", "vulnerabilities",
		"--build-arg", f"GO_VERSION={go}", "--target", "vulnerability-report", "--output", into, tree,
	)
	return os.path.join(into, "report.json")


def gate(tree: str) -> None:
	"""Builds the tree through its own merge gate, `docker build .`."""
	built = run("docker", "build", ".", cwd=tree, ok=(0, 1))
	if built.returncode != 0:
		raise Hold(f"the patched tree fails `docker build .`: {(built.stderr.strip().splitlines() or [''])[-1]}")


def ship(tree: str, version: str) -> str:
	"""Commits the tree and pushes it to its release branch, and returns its commit."""
	out("git", "add", "-A", cwd=tree)
	out("git", "commit", "-qm", f"Release {version}", cwd=tree)
	server = urllib.parse.urlsplit(os.environ["FORGE_SERVER_URL"])
	remote = f"{server.scheme}://x-access-token:{os.environ['FORGE_TOKEN']}@{server.netloc}/{os.environ['FORGE_REPOSITORY']}.git"
	# Delete a branch a failed run left, so the push creates it and the forge checks it against main.
	run("git", "push", "-q", remote, f":refs/heads/release/{version}", cwd=tree, ok=(0, 1))
	out("git", "push", "-q", remote, f"HEAD:refs/heads/release/{version}", cwd=tree)
	return out("git", "rev-parse", "HEAD", cwd=tree)


@contextlib.contextmanager
def worktree(main: str, tag: str, path: str):
	out("git", "worktree", "add", "-q", "--detach", path, tag, cwd=main)
	try:
		yield path
	finally:
		run("git", "worktree", "remove", "--force", path, cwd=main)


def prepare(tree: str, released: carried.Carried, patch: carried.Carried) -> None:
	"""Puts the patch's Go and modules into the release tree; its GoReleaser comes with main's .github/."""
	if patch.go != released.go:
		for name in ("Dockerfile", "compose.yaml"):
			path = os.path.join(tree, name)
			write(path, carried.set_go(read(path), carried.dotted(patch.go)))
	if patch.go_mod != released.go_mod:
		write(os.path.join(tree, "go.mod"), patch.go_mod)
		if patch.go_sum:
			write(os.path.join(tree, "go.sum"), patch.go_sum)
		else:
			with contextlib.suppress(FileNotFoundError):
				os.remove(os.path.join(tree, "go.sum"))


def fixed_by(tree: str, work: str, released: carried.Carried, patch: carried.Carried, scanner) -> str:
	"""Prepares the tree, and names the vulnerabilities it loses by it; Hold while it still carries one."""
	before = scanner(tree, carried.dotted(released.go), os.path.join(work, "before"))
	prepare(tree, released, patch)
	after = scanner(tree, carried.dotted(patch.go), os.path.join(work, "after"))
	if run(sys.executable, FILER, "check", after, ok=(0, 1)).returncode != 0:
		raise Hold("the patched tree still carries a known vulnerability")
	return out(sys.executable, FILER, "fixed-between", before, after)


def carry_workflows(main: str, tree: str) -> None:
	"""Gives the release tree main's .github/, since the forge refuses the workflow token a push whose workflows differ from main's."""
	try:
		shutil.rmtree(os.path.join(tree, ".github"))
		shutil.copytree(os.path.join(main, ".github"), os.path.join(tree, ".github"))
	except OSError as err:
		raise ValueError(f"cannot carry main's .github/ into the release tree: {err}") from err


def cut(main: str, last: str, date: str, effects: Effects) -> int:
	"""Cuts the patch of the last release that the Go, modules and GoReleaser of main's checkout at `main` call for, if any."""
	released = carried.of(lambda path: run("git", "show", f"{last}:{path}", cwd=main, ok=(0, 128)).stdout)
	patch = carried.due(released, carried.of(lambda path: read_if_any(os.path.join(main, path))))
	if patch is None:
		print(f"No patch due: main carries the Go, modules and GoReleaser {last} does.")
		return 0
	version = next_patch(last)
	try:
		built_for = go_release_notes.systems(read(os.path.join(main, "release-tooling", "goreleaser.yaml")))
		platforms = go_release_notes.lines(released.go, patch.go, built_for, effects.go_notes)
		with tempfile.TemporaryDirectory() as work, worktree(main, last, os.path.join(work, "tree")) as tree:
			fixed = fixed_by(tree, work, released, patch, effects.scan)
			text = changelog.section(version, date, changelog.changes(released, patch, platforms), last, fixed)
			path = os.path.join(tree, "CHANGELOG.md")
			write(path, changelog.insert_section(read(path), text))
			carry_workflows(main, tree)
			effects.gate(tree)
			code, tag = effects.publish(read(path), effects.ship(tree, version))
	except (Hold, go_release_notes.Unreadable) as held:
		print(f"::warning::no patch of {last} is cut: {held}")
		return 0
	if tag:
		effects.output(f"tag={tag}")
	return code


def plan(report: str, todo: str, changelog_path: str, last: str) -> int:
	text = changelog.reconcile(read(changelog_path), out("git", "show", f"{last}:CHANGELOG.md"))
	planned = run(sys.executable, FILER, "plan", report, todo, "--patch", next_patch(last))
	sys.stderr.write(planned.stderr)
	write(changelog_path, changelog.record(text, planned.stdout.splitlines()))
	return 0


def publish(text: str, sha: str) -> tuple[int, str | None]:
	try:
		return publish_release.publish(text, sha)
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
				here = os.getcwd()
				today = datetime.datetime.now(datetime.UTC).date().isoformat()
				effects = Effects(functools.partial(scan, here), gate, ship, publish, github_output, go_release_notes.fetch)
				return cut(here, last_release(), today, effects)
			case ["plan", report, todo, changelog_path]:
				read(report)
				return plan(report, todo, changelog_path, last_release())
	except ValueError as err:
		print(err, file=sys.stderr)
		return 2
	print(__doc__, file=sys.stderr)
	return 2


if __name__ == "__main__":
	sys.exit(main(sys.argv[1:]))
