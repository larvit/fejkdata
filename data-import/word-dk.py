"""Rebuild data/da_DK/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from COR, the Danish central word register.

    data-import/word-dk.py [--cor URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

COR₁ (Dansk Sprognævn, CC0) lists the forms of each lexeme of the Danish orthographic dictionary, tagged. A row is a
noun, a verb, an adjective or an adverb that COR spells in lower-case letters alone, with a form of the same lexeme: a
noun's definite singular, a verb's active present and an adjective's definite singular. An adverb carries no form. Only the
forms COR marks as the norm (N) count. A noun, verb or adjective lacking its form is dropped, and so is one whose word or form is on LDNOOBW's Danish list
of rude words. COR gives no frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import argparse
from collections.abc import Collection
from pathlib import Path

import source
import wordtables

COR = "https://ordregister.dk/files/cor1.5.1.0.tsv"
OUT = Path(__file__).resolve().parent.parent / "data" / "da_DK"
CACHE = Path(__file__).resolve().parent / "cache"
FORMS = {"noun": "definite", "verb": "present", "adj": "definite", "adv": ""}
# COR's tag of a word and of its form, per part of speech; a noun's tag also names its gender, fk or itk.
TAGS = {
    "noun": ({"sb", "sg", "ubest"}, {"sb", "sg", "best"}),
    "verb": ({"vb", "inf", "akt"}, {"vb", "præs", "akt"}),
    "adj": ({"adj", "sg", "ubest", "fk"}, {"adj", "sg", "best"}),
    "adv": ({"adv"}, set()),
}


def tagged(tag: str, wanted: set[str]) -> bool:
    held = set(tag.split("."))
    return bool(wanted) and wanted <= held and len(held - wanted) <= (1 if "sb" in wanted else 0)


def lexemes(cor: str) -> dict[tuple[str, str], tuple[str, str]]:
    """Per lexeme and part of speech, in COR's order, its word and the first of its forms, where the norm holds them."""
    found: dict[tuple[str, str], tuple[str, str]] = {}
    for line in cor.split("\n"):
        cells = line.split("\t")
        if len(cells) != 6 or cells[5] != "N":
            continue
        lexeme, lemma, tag, form = cells[0].split(".")[1], cells[1], cells[3], cells[4]
        for part, (own, inflected) in TAGS.items():
            word, held = found.get((lexeme, part), ("", ""))
            if tagged(tag, own) and form == lemma:
                found[(lexeme, part)] = (form, held)
            elif tagged(tag, inflected) and not held:
                found[(lexeme, part)] = (word, form)
    return found


def tables(cor: str, rude: Collection[str]) -> dict[str, list[dict[str, str]]]:
    rows: dict[str, dict[str, dict[str, str]]] = {part: {} for part in FORMS}
    for (_, part), (word, form) in lexemes(cor).items():
        column = FORMS[part]
        if not word.isalpha() or not word.islower() or word in rude:
            continue
        if not column:
            rows[part].setdefault(word, {"word": word})
        elif form.isalpha() and form.islower() and form not in rude:
            rows[part].setdefault(word, {"word": word, column: form})
    return {part: [held[w] for w in sorted(held)] for part, held in rows.items()}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--cache", default=str(CACHE))
    parser.add_argument("--cor", default=COR)
    parser.add_argument("--rude", default="")
    parser.add_argument("--out", default=str(OUT))
    args = parser.parse_args()
    cor = source.fetch(args.cor, args.cache, Path(args.cor).name, magic=b"COR.").decode("utf-8")
    wordtables.write(Path(args.out), FORMS, tables(cor, wordtables.rude("da", args.cache, args.rude)))


if __name__ == "__main__":
    main()
