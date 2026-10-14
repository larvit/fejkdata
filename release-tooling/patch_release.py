#!/usr/bin/env python3
"""Cut a patch release from the last release when a push to main raises Go by a patch.

  patch_release.py go <Dockerfile before> <Dockerfile after>
      Prints the new Go version when the push raised it by a patch, else nothing.
  patch_release.py version <last release's tag>
      Prints the patch release's version.
  patch_release.py apply <release checkout> <main's CHANGELOG.md> <version> <Go version>
      Sets the checkout's Go to the new version, puts main's section for the version atop its
      CHANGELOG.md, and prints that section's body as the release notes.
"""

import re
import sys

GO_VERSION = re.compile(r"^ARG GO_VERSION=(\d+)\.(\d+)\.(\d+)$", re.M)
SECTION = re.compile(r"^## \[(?P<version>[^\]]+)\][^\n]*\n(?P<body>.*?)(?=^## \[|\Z)", re.M | re.S)


def go_patch(before: str, after: str) -> str | None:
	old, new = (tuple(map(int, GO_VERSION.search(text).groups())) for text in (before, after))
	if old[:2] != new[:2] or new[2] <= old[2]:
		return None
	return ".".join(map(str, new))


def next_version(tag: str) -> str:
	match = re.fullmatch(r"v(\d+)\.(\d+)\.(\d+)", tag)
	if match is None:
		raise ValueError(f"the last release's tag {tag!r} is no vX.Y.Z")
	major, minor, patch = match.groups()
	return f"{major}.{minor}.{int(patch) + 1}"


def set_go(text: str, version: str) -> str:
	text = re.sub(r"^ARG GO_VERSION=\S+$", f"ARG GO_VERSION={version}", text, flags=re.M)
	return re.sub(r"golang:\$\{GO_VERSION:-[^}]+\}", f"golang:${{GO_VERSION:-{version}}}", text)


def carry_section(main: str, released: str, version: str) -> tuple[str, str]:
	"""Return the released CHANGELOG.md with main's section for the version atop it, and that section's body."""
	section = next((m for m in SECTION.finditer(main) if m["version"] == version), None)
	if section is None:
		raise ValueError(f"main's CHANGELOG.md has no section for {version}")
	first = next(m for m in SECTION.finditer(released) if m["version"] != "Unreleased")
	return released[: first.start()] + section[0] + released[first.start() :], section["body"].strip()


def main(argv: list[str]) -> int:
	match argv:
		case ["go", before, after]:
			with open(before, encoding="utf-8") as b, open(after, encoding="utf-8") as a:
				print(go_patch(b.read(), a.read()) or "")
		case ["version", tag]:
			print(next_version(tag))
		case ["apply", checkout, main_changelog, version, go]:
			for name in ("Dockerfile", "compose.yaml"):
				path = f"{checkout}/{name}"
				with open(path, encoding="utf-8") as f:
					text = set_go(f.read(), go)
				with open(path, "w", encoding="utf-8") as f:
					f.write(text)
			with open(main_changelog, encoding="utf-8") as m, open(f"{checkout}/CHANGELOG.md", encoding="utf-8") as r:
				changelog, notes = carry_section(m.read(), r.read(), version)
			with open(f"{checkout}/CHANGELOG.md", "w", encoding="utf-8") as f:
				f.write(changelog)
			print(notes)
		case _:
			print(__doc__, file=sys.stderr)
			return 2
	return 0


if __name__ == "__main__":
	sys.exit(main(sys.argv[1:]))
