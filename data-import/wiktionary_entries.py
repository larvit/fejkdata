"""The English Wiktionary's entries for one language, as kaikki.org extracts them (CC BY-SA 4.0), read into a word table's candidates."""
import io
import json
from collections.abc import Callable, Iterator, Mapping
from typing import Any

import wordtables

Heads = Mapping[str, set[str]]


def heads(entry: Mapping[str, Any]) -> set[str]:
	"""What each head template of the entry calls it: the template's name, or for the generic `head` its second argument, such as `verb form`."""
	return {str(t.get("args", {}).get("2", "")) if t.get("name") == "head" else str(t.get("name", "")) for t in entry.get("head_templates", [])}


def headword(entry: Mapping[str, Any], parts: Heads) -> bool:
	"""Whether the entry's head template calls it a word of its part of speech, not a form or a participle, and some sense stands on its own, not as a form or variant of another entry."""
	return bool(heads(entry) & parts[entry["pos"]]) and any("form_of" not in sense and "alt_of" not in sense for sense in entry.get("senses", []))


def candidates(dump: bytes, parts: Heads, forms: Mapping[str, str], plural: Callable[[Mapping[str, Any]], str]) -> Iterator[wordtables.Candidate]:
	"""A part of speech, headword and plural per entry of a dump, one JSON object a line; the plural is empty where `forms` gives the part no column."""
	for line in io.BytesIO(dump):
		entry = json.loads(line)
		part = entry.get("pos", "")
		if part in parts and headword(entry, parts):
			yield part, entry.get("word", ""), plural(entry) if forms[part] else ""
