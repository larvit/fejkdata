#!/usr/bin/env python3
"""Rebuild data/sv_SE/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from SALDO's morphology.

    data-import/word-se.py [--saldo URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

A row is a noun, verb, adjective or adverb that SALDO spells in lower-case letters alone, with the forms its table
carries: a noun's definite singular, a verb's present and an adjective's definite plural. An entry lacking that form, such as
a noun with no singular or a verb only in its s-form, is dropped, and so is one whose word or form is on LDNOOBW's Swedish
list of rude words. SALDO gives no frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import argparse
import io
import re
import xml.etree.ElementTree as ElementTree
from collections.abc import Collection, Iterator, Mapping
from pathlib import Path

import source
import tsv

SALDO = "https://svn.spraakbanken.gu.se/sb-arkiv/pub/lmf/saldom/saldom.xml"
RUDE = "https://raw.githubusercontent.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words/master/sv"
OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE"
CACHE = Path(__file__).resolve().parent / "cache"
LETTERS = re.compile(r"[^\W\d_A-ZÅÄÖ]+")
# SALDO's partOfSpeech: the path, the form's column, and SALDO's msd (its morphological tag) for that form and for the lemma.
PARTS = {
    "nn": ("noun", "definite", "sg def nom", "sg indef nom"),
    "vb": ("verb", "present", "pres ind aktiv", "inf aktiv"),
    "av": ("adj", "definite", "pos def pl nom", "pos indef sg u nom"),
    "ab": ("adv", "", "", ""),
}


def feats(element: ElementTree.Element) -> dict[str, str]:
    return {f.attrib["att"]: f.attrib["val"] for f in element.findall("feat")}


def entries(saldo: bytes) -> Iterator[tuple[str, str, dict[str, str]]]:
    """Each entry as its lemma, part of speech and first written form per msd."""
    for _, entry in ElementTree.iterparse(io.BytesIO(saldo)):
        if entry.tag != "LexicalEntry":
            continue
        lemma = feats(entry.findall("Lemma/FormRepresentation")[0])
        forms: dict[str, str] = {}
        for form in entry.findall("WordForm"):
            written = feats(form)
            forms.setdefault(written["msd"], written["writtenForm"])
        yield lemma["writtenForm"], lemma["partOfSpeech"], forms
        entry.clear()


def tables(saldo: bytes, rude: Collection[str]) -> dict[str, list[dict[str, str]]]:
    rows: dict[str, dict[str, dict[str, str]]] = {part: {} for part, *_ in PARTS.values()}
    for word, pos, forms in entries(saldo):
        if pos not in PARTS or not LETTERS.fullmatch(word) or word in rude:
            continue
        part, column, msd, own = PARTS[pos]
        if not column:
            rows[part].setdefault(word, {"word": word})
        elif forms.get(own) == word and msd in forms and LETTERS.fullmatch(forms[msd]) and forms[msd] not in rude:
            rows[part].setdefault(word, {"word": word, column: forms[msd]})
    return {part: [held[w] for w in sorted(held)] for part, held in rows.items()}


def word_json(counts: Mapping[str, int]) -> str:
    items = ",\n".join(f'    {{ "format": "{{.word-{part}}}", "weight": {n} }}' for part, n in counts.items())
    fields = []
    for part, column, *_ in PARTS.values():
        form = f', "{column}": "{{{part}word.{column}}}"' if column else ""
        fields.append(f'  "{part}": {{ "format": "{{.word-{part} as {part}word}}{{{part}word}}"{form} }}')
    return '{\n  "format": "{any}",\n  "any": [\n' + items + "\n  ],\n" + ",\n".join(fields) + "\n}\n"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--cache", default=str(CACHE))
    parser.add_argument("--saldo", default=SALDO)
    parser.add_argument("--rude", default=RUDE)
    parser.add_argument("--out", default=str(OUT))
    args = parser.parse_args()
    saldo = source.fetch(args.saldo, args.cache, "saldom.xml", magic=b"<LexicalResource")
    listed = source.fetch(args.rude, args.cache, "ldnoobw-sv.txt").decode("utf-8").split("\n")
    rows = tables(saldo, {w.strip().lower() for w in listed})
    out = Path(args.out)
    for part, column, *_ in PARTS.values():
        tsv.write(out / f"word-{part}.tsv", ["word", column] if column else ["word"], rows[part])
        (out / f"word-{part}.json").write_text(f'{{ "format": "{{word}}", "rows": "word-{part}.tsv" }}\n', encoding="utf-8")
    (out / "word.json").write_text(word_json({part: len(held) for part, held in rows.items()}), encoding="utf-8")


if __name__ == "__main__":
    main()
