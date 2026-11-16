"""Rebuild data/de_DE/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Wikidata's German lexemes.

    data-import/word-de.py [--lexemes URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

Wikidata's lexemes (CC0) give each German word its lexical category and its forms, each tagged with grammatical
features. A row is a noun, a verb, an adjective or an adverb whose lemma is letters alone, a noun's capitalised and
the others' lower-case. A noun carries its first nominative plural, and one without a plural in letters alone is
dropped. A word on LDNOOBW's German list of rude words, compared in lower case, is dropped, and so is a noun whose plural is on
it. Wikidata gives no frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import argparse
import gzip
import io
import json
from collections.abc import Collection, Iterator, Mapping
from pathlib import Path
from typing import Any

import source
import wordtables

LEXEMES = "https://dumps.wikimedia.org/wikidatawiki/entities/latest-lexemes.json.gz"
OUT = Path(__file__).resolve().parent.parent / "data" / "de_DE"
CACHE = Path(__file__).resolve().parent / "cache"
FORMS = {"noun": "plural", "verb": "", "adj": "", "adv": ""}
GERMAN = "Q188"
CATEGORIES = {"Q1084": "noun", "Q24905": "verb", "Q34698": "adj", "Q380057": "adv"}
NOMINATIVE_PLURAL = {"Q131105", "Q146786"}


def german(dump: bytes) -> Iterator[Mapping[str, Any]]:
    """Each German lexeme of the dump, a JSON array written one lexeme per line."""
    with gzip.open(io.BytesIO(dump), "rt", encoding="utf-8") as lines:
        for line in lines:
            line = line.strip().rstrip(",")
            if line.startswith("{"):
                lexeme = json.loads(line)
                if lexeme.get("language") == GERMAN:
                    yield lexeme


def spelled(word: str, part: str) -> bool:
    """A noun is a capital and lower-case letters, any other part lower-case letters alone."""
    head, tail = (word[:1], word[1:]) if part == "noun" else ("", word)
    return word.isalpha() and (not head or head.isupper()) and tail.islower()


def plural(lexeme: Mapping[str, Any]) -> str:
    for form in lexeme.get("forms", []):
        if NOMINATIVE_PLURAL <= set(form["grammaticalFeatures"]):
            return str(form["representations"].get("de", {}).get("value", ""))
    return ""


def tables(dump: bytes, rude: Collection[str]) -> dict[str, list[dict[str, str]]]:
    rows: dict[str, dict[str, dict[str, str]]] = {part: {} for part in FORMS}
    for lexeme in german(dump):
        part = CATEGORIES.get(lexeme["lexicalCategory"], "")
        word = lexeme["lemmas"].get("de", {}).get("value", "")
        if not part or not spelled(word, part) or word.lower() in rude:
            continue
        if part != "noun":
            rows[part].setdefault(word, {"word": word})
        elif spelled(form := plural(lexeme), part) and form.lower() not in rude:
            rows[part].setdefault(word, {"word": word, "plural": form})
    return {part: [held[w] for w in sorted(held)] for part, held in rows.items()}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--cache", default=str(CACHE))
    parser.add_argument("--lexemes", default=LEXEMES)
    parser.add_argument("--rude", default="")
    parser.add_argument("--out", default=str(OUT))
    args = parser.parse_args()
    dump = source.fetch(args.lexemes, args.cache, "wikidata-lexemes.json.gz", magic=b"\x1f\x8b")
    wordtables.write(Path(args.out), FORMS, tables(dump, wordtables.rude("de", args.cache, args.rude)))


if __name__ == "__main__":
    main()
