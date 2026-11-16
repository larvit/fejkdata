"""Rebuild data/da_DK/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from COR, the Danish central word register.

    data-import/word-dk.py [--cor URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

COR₁ (Dansk Sprognævn, CC0) lists the forms of each lexeme of the Danish orthographic dictionary, tagged. A row is a
noun, a verb, an adjective or an adverb that COR spells in lower-case letters alone, with a form of the same lexeme: a
noun's definite singular, a verb's active present and an adjective's definite singular. An adverb carries no form. Only the
forms COR marks as the norm (N) count. A noun, verb or adjective lacking its form is dropped, and so is one whose word or form is on LDNOOBW's Danish list
of rude words. COR gives no frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
from collections.abc import Iterator
from pathlib import Path

import wordtables

COR = "https://ordregister.dk/files/cor1.5.1.0.tsv"
OUT = Path(__file__).resolve().parent.parent / "data" / "da_DK"
FORMS = {"noun": "definite", "verb": "present", "adj": "definite", "adv": ""}
# COR's tag of a word and of its form, per part of speech; a noun's tag also names its gender, fk or itk.
TAGS = {
    "noun": ({"sb", "sg", "ubest"}, {"sb", "sg", "best"}),
    "verb": ({"vb", "inf", "akt"}, {"vb", "præs", "akt"}),
    "adj": ({"adj", "sg", "ubest", "fk"}, {"adj", "sg", "best"}),
    "adv": ({"adv"}, set()),
}


def tagged(tag: str, wanted: set[str]) -> bool:
    tags = set(tag.split("."))
    return bool(wanted) and wanted <= tags and len(tags - wanted) <= (1 if "sb" in wanted else 0)


def lexemes(cor: str) -> dict[tuple[str, str], tuple[str, str]]:
    """Per lexeme and part of speech, in COR's order, its word and the first of its forms, where the norm holds them."""
    found: dict[tuple[str, str], tuple[str, str]] = {}
    for line in cor.split("\n"):
        cells = line.split("\t")
        if len(cells) != 6 or cells[5] != "N":
            continue
        lexeme, lemma, tag, form = cells[0].split(".")[1], cells[1], cells[3], cells[4]
        for part, (own, inflected) in TAGS.items():
            word, first_form = found.get((lexeme, part), ("", ""))
            if tagged(tag, own) and form == lemma:
                found[(lexeme, part)] = (form, first_form)
            elif tagged(tag, inflected) and not first_form:
                found[(lexeme, part)] = (word, form)
    return found


def candidates(cor: bytes) -> Iterator[wordtables.Candidate]:
    for (_, part), (word, form) in lexemes(cor.decode("utf-8")).items():
        yield part, word, form


if __name__ == "__main__":
    wordtables.main(__doc__, "da", OUT, FORMS, [wordtables.Source(flag="cor", url=COR, magic=b"COR.", candidates=candidates)])
