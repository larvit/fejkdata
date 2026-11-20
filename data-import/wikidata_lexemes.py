"""Wikidata's lexemes (CC0), read for one language into a word table's candidates."""
import gzip
import io
import json
from collections.abc import Iterator, Mapping
from typing import Any

import wordtables

DUMP = "https://dumps.wikimedia.org/wikidatawiki/entities/latest-lexemes.json.gz"
CATEGORIES = {"Q1084": "noun", "Q24905": "verb", "Q34698": "adj", "Q380057": "adv"}
NOMINATIVE_PLURAL = {"Q131105", "Q146786"}


def lexemes(dump: bytes, language: str) -> Iterator[Mapping[str, Any]]:
    """Each lexeme of the language, a Wikidata item such as `Q188`, in the dump: a JSON array written one lexeme per line."""
    with gzip.open(io.BytesIO(dump), "rt", encoding="utf-8") as lines:
        for line in lines:
            line = line.strip().rstrip(",")
            if line.startswith("{"):
                lexeme = json.loads(line)
                if lexeme.get("language") == language:
                    yield lexeme


def plural(lexeme: Mapping[str, Any], code: str) -> str:
    for form in lexeme.get("forms", []):
        if NOMINATIVE_PLURAL <= set(form["grammaticalFeatures"]):
            return str(form["representations"].get(code, {}).get("value", ""))
    return ""


def candidates(dump: bytes, language: str, code: str, forms: Mapping[str, str]) -> Iterator[wordtables.Candidate]:
    """A part of speech, lemma and first nominative plural per lexeme of the language, whose lemmas and forms are written in `code`, such as `de`; the plural is empty where `forms` gives the part no column."""
    for lexeme in lexemes(dump, language):
        part = CATEGORIES.get(lexeme["lexicalCategory"], "")
        if part:
            yield part, lexeme["lemmas"].get(code, {}).get("value", ""), plural(lexeme, code) if forms[part] else ""
