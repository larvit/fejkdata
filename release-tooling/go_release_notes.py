"""The requirements a Go minor's release notes set on the systems the binaries are built for."""

import re
import urllib.error
import urllib.request
from typing import Callable

URL = "https://raw.githubusercontent.com/golang/website/master/_content/doc/go{}.md"


class Unreadable(Exception):
	"""The notes say something about a built-for system that `requirements` cannot read, so the changelog would miss it."""


def systems(goreleaser_yaml: str) -> set[str]:
	"""The goos and goarch goreleaser.yaml builds, spelled as the notes head their Ports subsections."""
	found = {}
	for key in ("goos", "goarch"):
		match = re.search(rf"^\s*{key}: \[([^\]]*)\]$", goreleaser_yaml, re.M)
		if match is None:
			raise ValueError(f"goreleaser.yaml has no `{key}: [...]` line")
		found[key] = [name.strip() for name in match[1].split(",")]
	return {name.capitalize() for name in found["goos"]} | {name.upper() for name in found["goarch"]}


def requirements(notes: str, minor: str, built_for: set[str]) -> list[str]:
	"""Each requirement Go `minor`'s notes set on a built-for system, as the notes state it."""
	ports = re.search(r"^## Ports\b[^\n]*\n(.*?)(?=^## |\Z)", notes, re.M | re.S)
	if ports is None:
		return []
	found = []
	for system in re.finditer(r"^### (\w+)[^\n]*\n(.*?)(?=^### |\Z)", ports[1], re.M | re.S):
		if system[1] in built_for:
			found += _read(system[1], plain(system[2]), minor)
	return found


def plain(markdown: str) -> str:
	"""The text with its comments and link targets dropped and its whitespace folded."""
	text = re.sub(r"\[([^\]]*)\]\([^)]*\)", r"\1", re.sub(r"<!--.*?-->", "", markdown, flags=re.S))
	return " ".join(text.split())


def _read(system: str, text: str, minor: str) -> list[str]:
	"""The requirements in one subsection; a sentence naming the minor and a requirement in any other shape is Unreadable.

	"Go 1.N is the last release that …" announces the next minor's requirement, which its own notes set.
	"""
	requirement = re.compile(rf"\bGo {re.escape(minor)} (?:now )?requires [^;]+?(?=;|\.(?:\s|$))")
	found = []
	for sentence in re.split(r"(?<=\.)\s+(?=[A-Z])", text):
		match = requirement.search(sentence)
		if match:
			found.append(match[0])
		elif re.search(rf"\bGo {re.escape(minor)}\b", sentence) and re.search(r"requir|discontinu|no longer|drop|remov", sentence, re.I) and "last release" not in sentence:
			raise Unreadable(f"Go {minor}'s release notes say under {system}: {sentence!r}, which names no requirement in the shape `Go {minor} requires …`")
	return found


def lines(old: tuple[int, int, int], new: tuple[int, int, int], built_for: set[str], notes: Callable[[str], str]) -> list[str]:
	"""A changelog line for each requirement a Go minor above old, up to new, sets on a built-for system."""
	minors = [f"{new[0]}.{minor}" for minor in range(old[1] + 1, new[1] + 1)] if old[0] == new[0] else []
	return [f"{line}, and so do the binaries." for minor in minors for line in requirements(notes(minor), minor, built_for)]


def fetch(minor: str) -> str:
	try:
		with urllib.request.urlopen(URL.format(minor)) as resp:
			return resp.read().decode()
	except urllib.error.URLError as err:
		raise ValueError(f"cannot read Go {minor}'s release notes: {err}") from err
