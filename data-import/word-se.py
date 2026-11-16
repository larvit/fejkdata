"""Rebuild data/sv_SE/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from SALDO's morphology.

    data-import/word-se.py [--saldo URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

A row is a noun, verb, adjective or adverb that SALDO spells in lower-case letters alone, with the forms its table
carries: a noun's definite singular, a verb's present and an adjective's definite plural. An entry lacking that form, such as
a noun with no singular or a verb only in its s-form, is dropped, and so is one whose word or form is on LDNOOBW's Swedish
list of rude words. SALDO gives no frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import io
from collections.abc import Iterator
from pathlib import Path
from xml.etree import ElementTree

import wordtables

SALDO = "https://svn.spraakbanken.gu.se/sb-arkiv/pub/lmf/saldom/saldom.xml"
OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE"
FORMS = {"noun": "definite", "verb": "present", "adj": "definite", "adv": ""}
# SALDO's partOfSpeech: the path, and SALDO's msd (its morphological tag) for the part's form and for the lemma.
PARTS = {
    "nn": ("noun", "sg def nom", "sg indef nom"),
    "vb": ("verb", "pres ind aktiv", "inf aktiv"),
    "av": ("adj", "pos def pl nom", "pos indef sg u nom"),
    "ab": ("adv", "", ""),
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


def candidates(saldo: bytes) -> Iterator[wordtables.Candidate]:
    for word, pos, forms in entries(saldo):
        if pos in PARTS:
            part, msd, own = PARTS[pos]
            yield part, word, forms.get(msd, "") if forms.get(own) == word else ""


if __name__ == "__main__":
    wordtables.main(__doc__, "sv", OUT, FORMS, [wordtables.Source("saldo", SALDO, "saldom.xml", b"<LexicalResource", "", candidates)])
