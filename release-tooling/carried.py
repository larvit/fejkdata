"""What a release carries, as one tree names it: its Go, the modules go.mod requires, and GoReleaser."""

import re
from dataclasses import dataclass
from typing import Callable

GO_VERSION = re.compile(r"^ARG GO_VERSION=(\d+)\.(\d+)\.(\d+)$", re.M)
GORELEASER = re.compile(r"goreleaser/goreleaser-action@\S+\s+with:\s+version: (\S+)")


@dataclass(frozen=True)
class Carried:
	go: tuple[int, int, int]
	goreleaser: str | None
	"""As packages.yml names it, since the Packages workflow builds every release with its own commit's GoReleaser."""
	go_mod: str
	go_sum: str


def of(show: Callable[[str], str]) -> Carried:
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


def requires(go_mod: str) -> dict[str, str]:
	"""The modules go.mod requires, by path, each at its version."""
	return _parse(go_mod)[0]


def bumps_only(released: str, main: str) -> bool:
	"""Whether main's go.mod differs from the release's only in module versions: the same modules required directly, and every other line the same."""
	(old, old_direct, old_rest), (new, new_direct, new_rest) = _parse(released), _parse(main)
	return new != old and new_direct == old_direct and new_rest == old_rest


def _parse(go_mod: str) -> tuple[dict[str, str], set[str], list[str]]:
	"""go.mod's requires by path, the paths it requires directly, and its other lines but blank ones."""
	found, direct, rest, block = {}, set(), [], False
	for line in go_mod.splitlines():
		code, _, comment = line.partition("//")
		words = code.split()
		if words == ["require", "("] or (block and words == [")"]):
			block = words != [")"]
			continue
		if words[:1] == ["require"] and len(words) == 3:
			words = words[1:]
		elif not (block and len(words) == 2):
			if line.strip():
				rest.append(line)
			continue
		found[words[0]] = words[1]
		if comment.strip() != "indirect":
			direct.add(words[0])
	return found, direct, rest


def due(released: Carried, main: Carried) -> Carried | None:
	"""What a patch of the release carries, main's wherever main moved on, or None while that is what the release carries.

	main's modules wait while its go.mod changes more than their versions, which is work merged since: a module
	required directly that the release doesn't, or another line, such as a `go` line raising the lowest supported Go.
	"""
	if main.goreleaser is None:
		raise ValueError("main's .github/workflows/packages.yml names no GoReleaser `version:` for goreleaser-action")
	modules = bumps_only(released.go_mod, main.go_mod)
	patch = Carried(
		max(released.go, main.go),
		main.goreleaser,
		main.go_mod if modules else released.go_mod,
		main.go_sum if modules else released.go_sum,
	)
	return None if patch == released else patch


def set_go(text: str, version: str) -> str:
	changed = re.sub(r"^ARG GO_VERSION=\S+$", f"ARG GO_VERSION={version}", text, flags=re.M)
	changed = re.sub(r"golang:\$\{GO_VERSION:-[^}]+\}", f"golang:${{GO_VERSION:-{version}}}", changed)
	if changed == text:
		raise ValueError("a file names no Go version to set")
	return changed
