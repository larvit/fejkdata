"""What the release scripts read from the repository: CHANGELOG.md's headings and go.work's modules."""

import re

MODULE = "github.com/larvit/fejkdata"
HEADING = re.compile(r"^## \[([^\]]+)\].*?$\n?(.*?)(?=^## \[|\Z)", re.M | re.S)


def top_heading() -> tuple[str, str] | None:
	"""CHANGELOG.md's top `## [...]` heading and its section's body, or None."""
	with open("CHANGELOG.md", encoding="utf-8") as f:
		top = HEADING.search(f.read())
	return (top.group(1), top.group(2).strip()) if top else None


def newest_version() -> str | None:
	"""The X.Y.Z of CHANGELOG.md's first versioned heading, or None."""
	with open("CHANGELOG.md", encoding="utf-8") as f:
		m = re.search(r"(?m)^## \[(\d+\.\d+\.\d+)\]", f.read())
	return m.group(1) if m else None


def workspace_dirs() -> list[str]:
	"""The module directories go.work's use directives name, `.` for the root."""
	with open("go.work", encoding="utf-8") as f:
		work = f.read()
	dirs = re.findall(r"(?m)^use\s+(\S+)\s*$", work)
	for block in re.findall(r"(?ms)^use\s*\((.*?)\)", work):
		dirs += block.split()
	return [d.removeprefix("./") or "." for d in dirs if d != "("]
