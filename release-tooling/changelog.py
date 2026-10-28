"""CHANGELOG.md's release sections: the one a patch adds, and keeping main's in step with the releases."""

import re

import carried
import publish_release

VULNERABILITY = re.compile(r"GO-\d+-\d+")
SECURITY = re.compile(r"^### Security\n(?P<bullets>.*?)(?=^### |\Z)", re.M | re.S)
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


def reconcile(main: str, released: str) -> str:
	"""Main's changelog with each release section the released changelog has and main lacks, and without each fix under
	Unreleased's Security that those sections name, as one recorded before its patch was cut."""
	have = {m["name"] for m in releases(main)}
	missing = [m[0] for m in releases(released) if m["name"] not in have]
	for text in reversed(missing):
		main = insert_section(main, text)
	return drop_fixes(main, set(VULNERABILITY.findall("".join(missing))))


def drop_fixes(changelog: str, shipped: set[str]) -> str:
	"""The changelog without each bullet under Unreleased's Security that names only shipped vulnerabilities, nor the
	heading once it has no bullet left."""
	section = unreleased(changelog)
	security = SECURITY.search(changelog, section.start("body"), section.end()) if section else None
	if security is None:
		return changelog
	bullets = BULLET.sub(lambda bullet: "" if names_only(bullet[0], shipped) else bullet[0], security["bullets"])
	kept = "### Security\n" + bullets if BULLET.search(bullets) else ""
	return changelog[: security.start()] + kept + changelog[security.end() :]


def names_only(text: str, vulnerabilities: set[str]) -> bool:
	"""Whether the text names a vulnerability, and none but those."""
	named = set(VULNERABILITY.findall(text))
	return bool(named) and named <= vulnerabilities


def record(changelog: str, fixed: list[str]) -> str:
	"""The changelog with each line under Unreleased's Security, save one whose vulnerabilities a release section names."""
	released = set(VULNERABILITY.findall("".join(m[0] for m in releases(changelog))))
	new = [line for line in fixed if not set(VULNERABILITY.findall(line)) <= released]
	if not new:
		return changelog
	bullets = "".join(f"- {line}\n" for line in new)
	section = unreleased(changelog)
	if section is None:
		return insert_section(changelog, f"## [Unreleased]\n\n### Security\n\n{bullets}\n")
	body = section["body"].strip("\n")
	if "### Security" in body:
		body += "\n" + bullets
	else:
		body += ("\n\n" if body else "") + "### Security\n\n" + bullets
	end = section.end()
	tail = "\n\n" + changelog[end:] if end < len(changelog) else "\n"
	return changelog[: section.start("body")] + "\n" + body.rstrip("\n") + tail
