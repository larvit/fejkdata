"""Rebuild data/fi_FI/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Kotus's word list.

    data-import/word-fi.py [--kotus URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

Nykysuomen sanalista (Kotus, CC BY 4.0) lists the headwords of Kielitoimiston sanakirja with their word classes and
no inflected forms. A row is a noun, a verb, an adjective or an adverb that the list spells in lower-case letters alone,
under each class the list gives it, unless the word is on LDNOOBW's Finnish list of rude words. The list gives no
frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import argparse
from collections.abc import Collection
from pathlib import Path

import source
import wordtables

KOTUS = "https://kaino.kotus.fi/lataa/nykysuomensanalista2024.txt"
OUT = Path(__file__).resolve().parent.parent / "data" / "fi_FI"
CACHE = Path(__file__).resolve().parent / "cache"
FORMS = {"noun": "", "verb": "", "adj": "", "adv": ""}
CLASSES = {"substantiivi": "noun", "verbi": "verb", "adjektiivi": "adj", "adverbi": "adv"}


def tables(kotus: str, rude: Collection[str]) -> dict[str, list[dict[str, str]]]:
    words: dict[str, set[str]] = {part: set() for part in FORMS}
    for line in kotus.split("\n")[1:]:
        cells = line.split("\t")
        if len(cells) < 3 or not cells[0].isalpha() or not cells[0].islower() or cells[0] in rude:
            continue
        for named in cells[2].split(","):
            if named.strip() in CLASSES:
                words[CLASSES[named.strip()]].add(cells[0])
    return {part: [{"word": w} for w in sorted(held)] for part, held in words.items()}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--cache", default=str(CACHE))
    parser.add_argument("--kotus", default=KOTUS)
    parser.add_argument("--rude", default="")
    parser.add_argument("--out", default=str(OUT))
    args = parser.parse_args()
    kotus = source.fetch(args.kotus, args.cache, Path(args.kotus).name, magic=b"Hakusana").decode("utf-8")
    wordtables.write(Path(args.out), FORMS, tables(kotus, wordtables.rude("fi", args.cache, args.rude)))


if __name__ == "__main__":
    main()
