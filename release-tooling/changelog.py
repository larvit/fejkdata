"""CHANGELOG.md's release sections: the one a patch adds, and keeping main's in step with the releases."""

import re

import carried
import publish_release

VULNERABILITY = re.compile(r"GO-\d+-\d+")


def changes(released: carried.Carried, patch: carried.Carried, platforms: list[str]) -> list[str]:
	"""The section's lines for what the patch carries that the release did not."""
	lines = []
	if patch.go != released.go:
		lines += [f"The binaries and image are built with Go {carried.dotted(patch.go)}.", *platforms]
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
