"""The requirements a Go minor's release notes set on the systems the binaries are built for."""

import re
import urllib.error
import urllib.request
from typing import Callable

URL = "https://raw.githubusercontent.com/golang/website/master/_content/doc/go{}.md"


def systems(goreleaser_yaml: str) -> set[str]:
	"""The goos and goarch goreleaser.yaml builds."""
	found = {}
	for key in ("goos", "goarch"):
		match = re.search(rf"^\s*{key}: \[([^\]]*)\]$", goreleaser_yaml, re.M)
		if match is None:
			raise ValueError(f"goreleaser.yaml has no `{key}: [...]` line")
		found[key] = [name.strip() for name in match[1].split(",")]
	return set(found["goos"]) | set(found["goarch"])


def requirements(notes: str, minor: str, built_for: set[str]) -> list[str]:
	"""Each requirement Go `minor`'s notes set on a built-for system, as the notes state it.

	It reads a system's requirements from the Ports subsection its name heads in any case, such as `Darwin` or
	`ARM64`. It reads none for a system the notes head otherwise, such as riscv64 under `RISC-V`, or wasm under
	`WebAssembly`.
	"""
	ports = re.search(r"^## Ports\b[^\n]*\n(.*?)(?=^## |\Z)", notes, re.M | re.S)
	if ports is None:
		return []
	found = []
	for system in re.finditer(r"^### (\w+)[^\n]*\n(.*?)(?=^### |\Z)", ports[1], re.M | re.S):
		if system[1].lower() in built_for:
			found += _read(system[1], plain(system[2]), minor)
	return found


def plain(markdown: str) -> str:
	"""The text with its comments and link targets dropped and its whitespace folded."""
	text = re.sub(r"\[([^\]]*)\]\([^)]*\)", r"\1", re.sub(r"<!--.*?-->", "", markdown, flags=re.S))
	return " ".join(text.split())


def _read(system: str, text: str, minor: str) -> list[str]:
	"""The requirements in one subsection; a sentence naming the minor and a requirement in any other shape is refused,
	so the changelog never misses one.

	"Go 1.N is the last release that …" announces the next minor's requirement, which its own notes set.
	"""
	requirement = re.compile(rf"\bGo {re.escape(minor)} (?:now )?requires [^;]+?(?=;|\.(?:\s|$))")
	found = []
	for sentence in re.split(r"(?<=\.)\s+(?=[A-Z])", text):
		match = requirement.search(sentence)
		if match:
			found.append(match[0])
		elif re.search(rf"\bGo {re.escape(minor)}\b", sentence) and re.search(r"requir|discontinu|no longer|drop|remov", sentence, re.I) and "last release" not in sentence:
			raise ValueError(f"Go {minor}'s release notes say under {system}: {sentence!r}, which names no requirement in the shape `Go {minor} requires …`")
	return found


def breaking_lines(old: tuple[int, int, int], new: tuple[int, int, int], built_for: set[str], notes: Callable[[str], str]) -> list[str]:
	"""A Breaking line for each requirement a Go minor above old, up to new, sets on a built-for system."""
	minors = [f"{new[0]}.{minor}" for minor in range(old[1] + 1, new[1] + 1)] if old[0] == new[0] else []
	return [f"{line}, and so do the binaries." for minor in minors for line in requirements(notes(minor), minor, built_for)]


def fetch(minor: str) -> str:
	try:
		with urllib.request.urlopen(URL.format(minor)) as resp:
			return resp.read().decode()
	except urllib.error.URLError as err:
		raise ValueError(f"cannot read Go {minor}'s release notes: {err}") from err
