"""Rebuild data/et_EE/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Wikidata's Estonian lexemes.

    data-import/word-et.py [--lexemes URL_OR_FILE] [--cache DIR] [--out DIR]

Wikidata's lexemes (CC0) give each Estonian word its lexical category and its forms, each tagged with grammatical
features. A row is a noun, a verb, an adjective or an adverb whose lemma is lower-case letters alone, so a verb is its
`-ma` infinitive. A noun carries its first nominative plural, and one without a plural in lower-case letters alone is
dropped. No word is dropped for being rude. Wikidata gives no frequency, so each word is drawn evenly, and `word` weighs a
part of speech by its row count.
"""
import gzip
import io
import json
import unicodedata
from collections.abc import Iterator, Mapping
from pathlib import Path
from typing import Any

import wordtables

LEXEMES = "https://dumps.wikimedia.org/wikidatawiki/entities/latest-lexemes.json.gz"
OUT = Path(__file__).resolve().parent.parent / "data" / "et_EE"
FORMS = {"noun": "plural", "verb": "", "adj": "", "adv": ""}
ESTONIAN = "Q9072"
CATEGORIES = {"Q1084": "noun", "Q24905": "verb", "Q34698": "adj", "Q380057": "adv"}
NOMINATIVE_PLURAL = {"Q131105", "Q146786"}


def estonian(dump: bytes) -> Iterator[Mapping[str, Any]]:
    """Each Estonian lexeme of the dump, a JSON array written one lexeme per line."""
    with gzip.open(io.BytesIO(dump), "rt", encoding="utf-8") as lines:
        for line in lines:
            line = line.strip().rstrip(",")
            if line.startswith("{"):
                lexeme = json.loads(line)
                if lexeme.get("language") == ESTONIAN:
                    yield lexeme


def spelled(word: str, _part: str) -> bool:
    """Lower-case letters alone: Unicode's Ll, which leaves out the modifier letters Python's islower() accepts."""
    return bool(word) and all(unicodedata.category(c) == "Ll" for c in word)


def plural(lexeme: Mapping[str, Any]) -> str:
    for form in lexeme.get("forms", []):
        if NOMINATIVE_PLURAL <= set(form["grammaticalFeatures"]):
            return str(form["representations"].get("et", {}).get("value", ""))
    return ""


def candidates(dump: bytes) -> Iterator[wordtables.Candidate]:
    for lexeme in estonian(dump):
        part = CATEGORIES.get(lexeme["lexicalCategory"], "")
        if part:
            yield part, lexeme["lemmas"].get("et", {}).get("value", ""), plural(lexeme) if FORMS[part] else ""


if __name__ == "__main__":
    wordtables.main(__doc__, "", OUT, FORMS, [wordtables.Source(flag="lexemes", url=LEXEMES, name="wikidata-lexemes.json.gz", magic=b"\x1f\x8b", candidates=candidates)], spelled)
