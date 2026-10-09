#!/usr/bin/env python3
"""Set fejkdata.Version, and each go.work module's require of another, to CHANGELOG.md's newest versioned heading.

Run from the repository root after heading CHANGELOG.md with the release's version.
"""

import re
import sys

from workspace import MODULE, newest_version, workspace_dirs


def main() -> int:
	version = newest_version()
	if version is None:
		print("CHANGELOG.md has no `## [X.Y.Z]` heading", file=sys.stderr)
		return 1
	tag = f"v{version}"
	rewrite("version.go", r'(?m)^(const Version = ")[^"]*(")$', rf"\g<1>{tag}\g<2>", 1)
	for d in workspace_dirs():
		rewrite(f"{d}/go.mod", rf"(?m)^(\s*(?:require\s+)?{re.escape(MODULE)}(?:/\S+)?\s+)v\S+", rf"\g<1>{tag}", 0)
	print(tag)
	return 0


def rewrite(path: str, pattern: str, repl: str, want: int) -> None:
	with open(path, encoding="utf-8") as f:
		text, n = re.subn(pattern, repl, f.read())
	if want and n != want:
		sys.exit(f"{path}: found {n} matches of {pattern}, want {want}")
	with open(path, "w", encoding="utf-8") as f:
		f.write(text)


if __name__ == "__main__":
	sys.exit(main())
