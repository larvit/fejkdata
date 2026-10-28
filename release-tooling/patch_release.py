#!/usr/bin/env python3
"""Own CHANGELOG.md's release sections: cut a patch of the last release carrying what main carries, and keep main in step.

  patch_release.py last
      Prints the last release's tag, as the forge names its latest release.
  patch_release.py cut
      When main carries a newer Go, module version bumps or another GoReleaser than the last release:
      puts them in the last release's tree, builds it through `docker build .`, commits it with
      main's .github/ and a changelog section to a branch release/X.Y.Z cut from the last
      release's tag, and publishes it with publish_release.py. Writes `tag=vX.Y.Z` to
      GITHUB_OUTPUT. It exits 2, cutting nothing, while govulncheck finds a vulnerability in that
      tree, while the tree fails the build, while main builds for other targets than the release,
      and while a new Go minor's release notes say something about a built-for system it cannot
      read: each waits for a human. While main's modules wait and nothing else is due, it scans
      the last release's tree, and exits 2 where that carries a known vulnerability.
  patch_release.py plan <govulncheck report of main> <todo.md> <CHANGELOG.md>
      Adds to CHANGELOG.md each release section the last release has and main lacks, and
      drops each line under Unreleased that such a section holds, as one recorded before its
      patch was cut. Plans the report in todo.md through file_vulnerabilities.py, and records
      each fixed item's line under Unreleased unless a release section already names its
      vulnerabilities. Records under Unreleased's Breaking what each Go minor main carries
      since the last release requires of a built-for system, unless the changelog holds it.

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


class Effects(NamedTuple):
	"""The steps of `cut` that run docker, read the network or publish, each a function a test can stand in for."""

	scan: Callable[[str, str, str], str]
	build_and_push: Callable[[str, str], str]
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


def scan_tree(root: str, tree: str, go: str, into: str) -> str:
	"""Govulncheck's report on the tree built with that Go, through the Dockerfile of main's checkout at root."""
	out(
		"docker", "build", "-f", os.path.join(root, "Dockerfile"), "--no-cache-filter", "vulnerabilities",
		"--build-arg", f"GO_VERSION={go}", "--target", "vulnerability-report", "--output", into, tree,
	)
	return os.path.join(into, "report.json")


def build_and_push(tree: str, version: str) -> str:
	"""Builds the tree through its own merge gate, commits it, pushes it to its release branch, and returns its commit."""
	out("docker", "build", ".", cwd=tree)
	out("git", "add", "-A", cwd=tree)
	out("git", "commit", "-qm", f"Release {version}", cwd=tree)
	server = urllib.parse.urlsplit(os.environ["FORGE_SERVER_URL"])
	remote = f"{server.scheme}://x-access-token:{os.environ['FORGE_TOKEN']}@{server.netloc}/{os.environ['FORGE_REPOSITORY']}.git"
	# Delete a branch a failed run left, so the push creates it and the forge checks it against main.
	run("git", "push", "-q", remote, f":refs/heads/release/{version}", cwd=tree, ok=(0, 1))
	out("git", "push", "-q", remote, f"HEAD:refs/heads/release/{version}", cwd=tree)
	return out("git", "rev-parse", "HEAD", cwd=tree)


@contextlib.contextmanager
def worktree(root: str, tag: str, path: str):
	out("git", "worktree", "add", "-q", "--detach", path, tag, cwd=root)
	try:
		yield path
	finally:
		run("git", "worktree", "remove", "--force", path, cwd=root)


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


def patch_tree(
	tree: str, work: str, last: str, released: carried.Carried, patch: carried.Carried, scan: Callable[[str, str, str], str],
) -> str:
	"""Puts the patch into the release tree, and names the vulnerabilities that drops; refuses a tree still vulnerable."""
	before = scan(tree, carried.dotted(released.go), os.path.join(work, "before"))
	prepare(tree, released, patch)
	after = scan(tree, carried.dotted(patch.go), os.path.join(work, "after"))
	found = known_vulnerabilities(after)
	if found:
		raise ValueError(f"{last}'s patched tree still carries {found}; a patch waits for a release the maintainer cuts")
	return out(sys.executable, FILER, "fixed-between", before, after)


def carry_workflows(root: str, tree: str) -> None:
	"""Gives the release tree main's .github/, since the forge refuses the workflow token a push whose workflows differ from main's."""
	try:
		shutil.rmtree(os.path.join(tree, ".github"))
		shutil.copytree(os.path.join(root, ".github"), os.path.join(tree, ".github"))
	except OSError as err:
		raise ValueError(f"cannot carry main's .github/ into the release tree: {err}") from err


def built_for(root: str, last: str) -> set[str]:
	"""The targets main's goreleaser.yaml builds, which must be the release's: main's config packages the patch."""
	path = "release-tooling/goreleaser.yaml"
	targets = go_release_notes.systems(read(os.path.join(root, path)))
	released = shown(root, last, path)
	if released and go_release_notes.systems(released) != targets:
		raise ValueError(f"main's {path} builds for other targets than {last}; a patch waits for a release the maintainer cuts")
	return targets


def shown(root: str, tag: str, path: str) -> str:
	"""The file at the tag, or "" where the tag's tree lacks it."""
	return run("git", "show", f"{tag}:{path}", cwd=root, ok=(0, 128)).stdout


def known_vulnerabilities(report: str) -> str:
	"""The vulnerabilities the report names, as the filer's `check` prints them, or "" where it names none."""
	checked = run(sys.executable, FILER, "check", report, ok=(0, 1))
	return checked.stdout.strip() if checked.returncode else ""


def wait_for_modules(root: str, last: str, released: carried.Carried, scan: Callable[[str, str, str], str]) -> None:
	"""Says main's modules wait for a release the maintainer cuts; refuses a last release that a known vulnerability holds meanwhile."""
	with tempfile.TemporaryDirectory() as work, worktree(root, last, os.path.join(work, "tree")) as tree:
		found = known_vulnerabilities(scan(tree, carried.dotted(released.go), os.path.join(work, "report")))
	if found:
		raise ValueError(
			f"{last} still carries {found}; main's modules wait for a release the maintainer cuts, since its go.mod changes more than module versions"
		)
	print(f"No patch due: main's go.mod changes more than module versions since {last}, so its modules wait for a release the maintainer cuts.")


def cut(root: str, last: str, date: str, effects: Effects) -> int:
	"""Cuts the patch of the last release that the Go, modules and GoReleaser of main's checkout at root call for, if any."""
	if run("git", "rev-parse", "--verify", "-q", f"{last}^{{commit}}", cwd=root, ok=(0, 1)).returncode:
		raise ValueError(f"main's checkout holds no tag {last}, the forge's latest release")
	released = carried.of(lambda path: shown(root, last, path))
	checkout = carried.of(lambda path: read_if_any(os.path.join(root, path)))
	patch = carried.due(released, checkout)
	if patch is None:
		if carried.requires(checkout.go_mod) != carried.requires(released.go_mod):
			wait_for_modules(root, last, released, effects.scan)
		else:
			print(f"No patch due: main carries nothing a patch of {last} can take.")
		return 0
	targets = built_for(root, last)
	breaking = go_release_notes.breaking_lines(released.go, patch.go, targets, effects.go_notes)
	with tempfile.TemporaryDirectory() as work, worktree(root, last, os.path.join(work, "tree")) as tree:
		fixed = patch_tree(tree, work, last, released, patch, effects.scan)
		text = changelog.section(next_patch(last), date, breaking, changelog.changes(released, patch), last, fixed)
		path = os.path.join(tree, "CHANGELOG.md")
		write(path, changelog.insert_section(read(path), text))
		carry_workflows(root, tree)
		code, tag = effects.publish(read(path), effects.build_and_push(tree, next_patch(last)))
	if tag:
		effects.output(f"tag={tag}")
	return code


def plan(report: str, todo: str, changelog_path: str, last: str, notes: Callable[[str], str]) -> int:
	breaking = unreleased_breaking(last, notes)
	text = changelog.reconcile(read(changelog_path), out("git", "show", f"{last}:CHANGELOG.md"))
	planned = run(sys.executable, FILER, "plan", report, todo, "--patch", next_patch(last))
	sys.stderr.write(planned.stderr)
	write(changelog_path, changelog.record(text, planned.stdout.splitlines(), breaking))
	return 0


def unreleased_breaking(last: str, notes: Callable[[str], str]) -> list[str]:
	"""The Breaking lines of main's Go minors since the last release, for the release shipping them, a patch or one the maintainer cuts."""
	built_for = go_release_notes.systems(read("release-tooling/goreleaser.yaml"))
	return go_release_notes.breaking_lines(carried.go_of(out("git", "show", f"{last}:Dockerfile")), carried.go_of(read("Dockerfile")), built_for, notes)


def publish_on_forge(text: str, sha: str) -> tuple[int, str | None]:
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
				root = os.getcwd()
				today = datetime.datetime.now(datetime.UTC).date().isoformat()
				effects = Effects(functools.partial(scan_tree, root), build_and_push, publish_on_forge, github_output, go_release_notes.fetch)
				return cut(root, last_release(), today, effects)
			case ["plan", report, todo, changelog_path]:
				read(report)
				return plan(report, todo, changelog_path, last_release(), go_release_notes.fetch)
	except ValueError as err:
		print(err, file=sys.stderr)
		return 2
	print(__doc__, file=sys.stderr)
	return 2


if __name__ == "__main__":
	sys.exit(main(sys.argv[1:]))
