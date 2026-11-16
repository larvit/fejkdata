"""Rebuild data/fi_FI/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Kotus's word list.

    data-import/word-fi.py [--kotus URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

Nykysuomen sanalista (Kotus, CC BY 4.0) lists the headwords of Kielitoimiston sanakirja with their word classes and
no inflected forms. A row is a noun, a verb, an adjective or an adverb that the list spells in lower-case letters alone,
under each class the list gives it, unless the word is on LDNOOBW's Finnish list of rude words. The list gives no
frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
from collections.abc import Iterator
from pathlib import Path

import wordtables

KOTUS = "https://kaino.kotus.fi/lataa/nykysuomensanalista2024.txt"
OUT = Path(__file__).resolve().parent.parent / "data" / "fi_FI"
FORMS = {"noun": "", "verb": "", "adj": "", "adv": ""}
CLASSES = {"substantiivi": "noun", "verbi": "verb", "adjektiivi": "adj", "adverbi": "adv"}


def candidates(kotus: bytes) -> Iterator[wordtables.Candidate]:
    for line in kotus.decode("utf-8").split("\n")[1:]:
        cells = line.split("\t")
        if len(cells) < 3:
            continue
        for named in cells[2].split(","):
            if named.strip() in CLASSES:
                yield CLASSES[named.strip()], cells[0], ""


if __name__ == "__main__":
    wordtables.main(__doc__, "fi", OUT, FORMS, [wordtables.Source("kotus", KOTUS, "", b"Hakusana", "", candidates)])
