"""CHANGELOG.md's release sections: the one a patch adds, and keeping main's in step with the releases."""

import re
from typing import Callable

import carried
import publish_release

VULNERABILITY = re.compile(r"GO-\d+-\d+")
BULLET = re.compile(r"^- .*\n(?:[ \t]+\S.*\n)*", re.M)


def changes(released: carried.Carried, patch: carried.Carried) -> list[str]:
	"""The section's lines for what the patch carries that the release did not."""
	lines = []
	if patch.go != released.go:
		lines.append(f"The binaries and image are built with Go {carried.dotted(patch.go)}.")
	if patch.goreleaser != released.goreleaser:
		lines.append(f"The binaries, packages and image are built with GoReleaser {patch.goreleaser}.")
	old, new = carried.requires(released.go_mod), carried.requires(patch.go_mod)
	for path in sorted(old.keys() | new.keys()):
		if path not in new:
			lines.append(f"The binaries no longer carry {path}.")
		elif path not in old:
			lines.append(f"The binaries carry {path} {new[path]}.")
		elif old[path] != new[path]:
			lines.append(f"The binaries carry {path} {new[path]}, up from {old[path]}.")
	return lines


def section(version: str, date: str, breaking: list[str], changed: list[str], last: str, fixed: str) -> str:
	"""The patch's section: Breaking first, as CHANGELOG.md's preamble asks, then Changed, then Security."""
	security = [f"The binaries no longer carry {last.removeprefix('v')}'s known vulnerabilities: {fixed}."] if fixed else []
	text = f"## [{version}] - {date}\n\n"
	for heading, lines in (("Breaking", breaking), ("Changed", changed), ("Security", security)):
		if lines:
			text += f"### {heading}\n\n" + "".join(f"- {line}\n" for line in lines) + "\n"
	return text


def releases(changelog: str) -> list[re.Match]:
	"""The changelog's release sections, newest first."""
	return [m for m in publish_release.SECTION.finditer(changelog) if publish_release.VERSION.fullmatch(m["name"])]


def insert_section(changelog: str, text: str) -> str:
	"""The changelog with the section above its newest release's, so below any Unreleased section."""
	newest = releases(changelog)
	at = newest[0].start() if newest else len(changelog)
	return changelog[:at] + text + changelog[at:]


def unreleased(changelog: str) -> re.Match | None:
	return next((m for m in publish_release.SECTION.finditer(changelog) if m["name"] == "Unreleased"), None)


def under_unreleased(changelog: str, heading: str) -> re.Match | None:
	"""Unreleased's subsection under the heading, with what it holds and the blank lines after them."""
	section = unreleased(changelog)
	subsection = re.compile(rf"^### {heading}\n(?P<held>.*?)(?=^### |\Z)", re.M | re.S)
	return subsection.search(changelog, section.start("body"), section.end()) if section else None


def reconcile(main: str, released: str) -> str:
	"""Main's changelog with each release section the released changelog has and main lacks, and without each line under
	Unreleased that those sections hold, as one recorded before its patch was cut: a fix naming only vulnerabilities they
	name, or a Breaking line they repeat."""
	have = {m["name"] for m in releases(main)}
	missing = [m[0] for m in releases(released) if m["name"] not in have]
	for text in reversed(missing):
		main = insert_section(main, text)
	sections = "".join(missing)
	main = drop(main, "Security", lambda bullet: names_only(bullet, set(VULNERABILITY.findall(sections))))
	return drop(main, "Breaking", lambda bullet: bullet in sections)


def drop(changelog: str, heading: str, shipped: Callable[[str], bool]) -> str:
	"""The changelog without each shipped bullet under Unreleased's heading, nor the heading once nothing is left under it."""
	found = under_unreleased(changelog, heading)
	if found is None:
		return changelog
	left = BULLET.sub(lambda bullet: "" if shipped(bullet[0]) else bullet[0], found["held"])
	if left == found["held"]:
		return changelog
	kept = re.sub(r"\n{3,}", "\n\n", f"### {heading}\n" + left) if left.strip() else ""
	return changelog[: found.start()] + kept + changelog[found.end() :]


def names_only(text: str, vulnerabilities: set[str]) -> bool:
	"""Whether the text names a vulnerability, and none but those."""
	named = set(VULNERABILITY.findall(text))
	return bool(named) and named <= vulnerabilities


def record(changelog: str, fixed: list[str], breaking: list[str]) -> str:
	"""The changelog with each fixed line under Unreleased's Security, save one whose vulnerabilities a release section
	names, and each Breaking line it lacks under Unreleased's Breaking, its first heading."""
	released = set(VULNERABILITY.findall("".join(m[0] for m in releases(changelog))))
	changelog = add(changelog, "Security", [line for line in fixed if not set(VULNERABILITY.findall(line)) <= released], first=False)
	return add(changelog, "Breaking", [line for line in breaking if f"- {line}\n" not in changelog], first=True)


def add(changelog: str, heading: str, lines: list[str], first: bool) -> str:
	"""The changelog with the lines as bullets under Unreleased's heading, made its first or last heading where it lacks it."""
	if not lines:
		return changelog
	bullets = "".join(f"- {line}\n" for line in lines)
	found = under_unreleased(changelog, heading)
	return join_subsection(changelog, found, bullets) if found else new_subsection(changelog, heading, bullets, first)


def join_subsection(changelog: str, found: re.Match, bullets: str) -> str:
	"""The changelog with the bullets after the subsection's last bullet, or after its text where it holds none."""
	inside = list(BULLET.finditer(changelog, found.start("held"), found.end()))
	if inside:
		return changelog[: inside[-1].end()] + bullets + changelog[inside[-1].end() :]
	text = found["held"].strip("\n")
	return changelog[: found.start("held")] + "\n" + (text + "\n\n" if text else "") + bullets + "\n" + changelog[found.end() :]


def new_subsection(changelog: str, heading: str, bullets: str, first: bool) -> str:
	"""The changelog with the heading and its bullets as Unreleased's first or last subsection, under a new Unreleased where it
	has none."""
	section = unreleased(changelog)
	if section is None:
		return insert_section(changelog, f"## [Unreleased]\n\n### {heading}\n\n{bullets}\n")
	body = section["body"].strip("\n")
	if first:
		body = f"### {heading}\n\n{bullets}" + ("\n" + body if body else "")
	else:
		body += ("\n\n" if body else "") + f"### {heading}\n\n" + bullets
	end = section.end()
	tail = "\n\n" + changelog[end:] if end < len(changelog) else "\n"
	return changelog[: section.start("body")] + "\n" + body.rstrip("\n") + tail
