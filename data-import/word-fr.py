"""Rebuild data/fr_FR/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from the English Wiktionary's French entries.

    data-import/word-fr.py [--wiktionary URL_OR_FILE] [--cache DIR] [--out DIR]

kaikki.org extracts those entries (CC BY-SA 4.0), one JSON object a line, each with its part of speech, head template and
forms. A row is a headword in lower-case letters alone, whose head template calls it a noun, verb, adjective or adverb, not
a form or a participle, so a verb is an infinitive and never a conjugated form, and which has a sense that is not a form or
variant of another entry.
A noun carries its first plural, or its own spelling where Wiktionary marks it invariable; a noun with neither is dropped.
"""
import io
import json
from collections.abc import Iterator, Mapping
from pathlib import Path
from typing import Any

import wordtables

WIKTIONARY = "https://kaikki.org/dictionary/French/kaikki.org-dictionary-French.jsonl"
OUT = Path(__file__).resolve().parent.parent / "data" / "fr_FR"
FORMS = {"noun": "plural", "verb": "", "adj": "", "adv": ""}
HEADS = {
    "noun": {"fr-noun", "noun", "nouns"},
    "verb": {"fr-verb", "verb"},
    "adj": {"fr-adj", "fr-adjective", "adj", "adjective"},
    "adv": {"fr-adv", "fr-adverb", "adverb"},
}


def heads(entry: Mapping[str, Any]) -> set[str]:
    """What each head template of the entry calls it: the template's name, or for the generic `head` its second argument, such as `verb form`."""
    return {str(t.get("args", {}).get("2", "")) if t.get("name") == "head" else str(t.get("name", "")) for t in entry.get("head_templates", [])}


def headword(entry: Mapping[str, Any]) -> bool:
    """Whether the entry's head template calls it a word of its part of speech, not a form or a participle, and some sense stands on its own, not as a form or variant of another entry."""
    return bool(heads(entry) & HEADS[entry["pos"]]) and any("form_of" not in sense and "alt_of" not in sense for sense in entry.get("senses", []))


def plural(entry: Mapping[str, Any]) -> str:
    for form in entry.get("forms", []):
        if "plural" in form.get("tags", []):
            return str(form["form"])
    if any("invariable" in sense.get("tags", []) for sense in entry.get("senses", [])):
        return str(entry["word"])
    return ""


def candidates(dump: bytes) -> Iterator[wordtables.Candidate]:
    for line in io.BytesIO(dump):
        entry = json.loads(line)
        part = entry.get("pos", "")
        if part in HEADS and headword(entry):
            yield part, entry.get("word", ""), plural(entry) if FORMS[part] else ""


if __name__ == "__main__":
    wordtables.main(__doc__, "", OUT, FORMS, [wordtables.Source(flag="wiktionary", url=WIKTIONARY, name="kaikki-french.jsonl", magic=b"{", candidates=candidates)], wordtables.lower_letters)
