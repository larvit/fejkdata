"""What a release carries, as one tree names it: its Go, the modules go.mod requires, and GoReleaser."""

import re
from dataclasses import dataclass
from typing import Callable

GO_VERSION = re.compile(r"^ARG GO_VERSION=(\d+)\.(\d+)\.(\d+)$", re.M)
GO_LINE = re.compile(r"^go (\S+)$", re.M)
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


def due(released: Carried, main: Carried) -> Carried | None:
	"""What a patch of the release carries, main's wherever main moved on, or None while that is what the release carries.

	main's modules wait while its go.mod raises the lowest supported Go, which only a major may do.
	"""
	if main.goreleaser is None:
		raise ValueError("main's .github/workflows/packages.yml names no GoReleaser `version:` for goreleaser-action")
	modules = requires(main.go_mod) != requires(released.go_mod) and go_line(main.go_mod) == go_line(released.go_mod)
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
