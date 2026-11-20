"""Rebuild data/fr_FR/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Wiktionary's French words.

    data-import/word-fr.py [--wiktionary URL_OR_FILE] [--cache DIR] [--out DIR]

kaikki.org extracts the English Wiktionary's French entries (CC BY-SA 4.0), one JSON object a line, each with its
part of speech and its forms, tagged. A row is a noun, a verb, an adjective or an adverb whose headword is lower-case
letters alone, and which some sense does not call an inflected form or a variant of another entry. A noun carries its
first plural, and one without a plural in lower-case letters alone is dropped. No word is dropped for being rude.
Wiktionary gives no frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import io
import json
import unicodedata
from collections.abc import Iterator, Mapping
from pathlib import Path
from typing import Any

import wordtables

WIKTIONARY = "https://kaikki.org/dictionary/French/kaikki.org-dictionary-French.jsonl"
OUT = Path(__file__).resolve().parent.parent / "data" / "fr_FR"
FORMS = {"noun": "plural", "verb": "", "adj": "", "adv": ""}


def spelled(word: str, _part: str) -> bool:
    """Lower-case letters alone: Unicode's Ll, which leaves out the modifier letters Python's islower() accepts."""
    return bool(word) and all(unicodedata.category(c) == "Ll" for c in word)


def headword(entry: Mapping[str, Any]) -> bool:
    """Whether some sense of the entry stands on its own, rather than naming an inflected form or a variant of another entry."""
    return any("form_of" not in sense and "alt_of" not in sense for sense in entry.get("senses", []))


def plural(entry: Mapping[str, Any]) -> str:
    for form in entry.get("forms", []):
        if "plural" in form.get("tags", []):
            return str(form["form"])
    return ""


def candidates(dump: bytes) -> Iterator[wordtables.Candidate]:
    for line in io.BytesIO(dump):
        entry = json.loads(line)
        part = entry.get("pos", "")
        if part in FORMS and headword(entry):
            yield part, entry.get("word", ""), plural(entry) if FORMS[part] else ""


if __name__ == "__main__":
    wordtables.main(__doc__, "", OUT, FORMS, [wordtables.Source(flag="wiktionary", url=WIKTIONARY, name="kaikki-french.jsonl", magic=b"{", candidates=candidates)], spelled)
