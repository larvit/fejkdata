"""Rebuild data/lt_LT/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from the English Wiktionary's Lithuanian entries.

    data-import/word-lt.py [--wiktionary URL_OR_FILE] [--cache DIR] [--out DIR]

kaikki.org extracts those entries (CC BY-SA 4.0), one JSON object a line. A row is a headword in lower-case letters alone,
whose head template calls it a noun, verb, adjective or adverb, not a form or a participle, and which has a sense that is
not a form or variant of another entry. A noun carries its plural, spelled without stress marks: the nominative plural of
its declension table, or the plural of its head template where it has no such one; a noun whose two differ, or which has
neither linked, is dropped. No word is dropped for being rude.
"""
from collections.abc import Callable, Mapping
from functools import partial
from pathlib import Path
from typing import Any

import wiktionary_entries
import wordtables

WIKTIONARY = "https://kaikki.org/dictionary/Lithuanian/kaikki.org-dictionary-Lithuanian.jsonl"
OUT = Path(__file__).resolve().parent.parent / "data" / "lt_LT"
FORMS = {"noun": "plural", "verb": "", "adj": "", "adv": ""}
HEADS = {
	"noun": {"lt-noun", "noun"},
	"verb": {"lt-verb", "verb"},
	"adj": {"lt-adj", "adjective"},
	"adv": {"lt-adv", "adverb", "adverbs"},
}


def plural(entry: Mapping[str, Any]) -> str:
	"""The page a plural of the entry links to, which is the form without its stress marks and its section anchor: the one its declension table gives for the nominative and its head template gives, else whichever of them links, and none where they differ."""
	typed = linked(entry, lambda tags: tags == ["plural"])
	nominative = linked(entry, lambda tags: {"nominative", "plural"} <= set(tags))
	return "" if typed and nominative and typed != nominative else nominative or typed


def linked(entry: Mapping[str, Any], wanted: Callable[[list[str]], bool]) -> str:
	for form in entry.get("forms", []):
		if wanted(form.get("tags", [])) and form.get("links"):
			return str(form["links"][0][1]).split("#")[0]
	return ""


candidates = partial(wiktionary_entries.candidates, parts=HEADS, forms=FORMS, plural=plural)

if __name__ == "__main__":
	wordtables.main(__doc__, "", OUT, FORMS, [wordtables.Source(flag="wiktionary", url=WIKTIONARY, name="kaikki-lithuanian.jsonl", magic=b"{", candidates=candidates)], wordtables.lower_letters)
