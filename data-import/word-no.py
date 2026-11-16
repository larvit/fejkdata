"""Rebuild nb_NO's and nn_NO's word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Norsk ordbank.

    data-import/word-no.py [--bokmal URL_OR_FILE] [--nynorsk URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

Norsk ordbank (Språkbanken, CC BY 4.0) lists every form of each lemma, tagged. A row is a common noun, a verb, an
adjective or an adverb that the ordbank spells in lower-case letters alone, with a form from the same paradigm: a noun's
definite singular, a verb's present and an adjective's definite singular. An adverb carries no form. Only the forms the current norm
holds count. A noun, verb or adjective lacking its form is dropped, and so is one whose word or form is on LDNOOBW's Norwegian list of rude words.
The ordbank gives no frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import io
import tarfile
from collections.abc import Iterator, Mapping
from functools import partial
from pathlib import Path
from typing import NamedTuple

import wordtables

BOKMAL = "https://www.nb.no/sbfil/leksikalske_databaser/ordbank/20220201_norsk_ordbank_nob_2005.tar.gz"
NYNORSK = "https://www.nb.no/sbfil/leksikalske_databaser/ordbank/20220201_norsk_ordbank_nno_2012.tar.gz"
OUT = Path(__file__).resolve().parent.parent / "data"
FORMS = {"noun": "definite", "verb": "present", "adj": "definite", "adv": ""}


class Tags(NamedTuple):
    """The tags every form of a part of speech carries, then those of its word and of its form."""

    every: set[str]
    word: set[str]
    form: set[str]


def parts(singular: str, definite: str) -> dict[str, Tags]:
    """Each part of speech's tags. Bokmål tags a singular `ent` and a definite `be`, nynorsk `eint` and `bu`."""
    return {
        "noun": Tags({"subst", "appell"}, {singular, "ub"}, {singular, definite}),
        "verb": Tags({"verb"}, {"inf"}, {"pres"}),
        "adj": Tags({"adj", "pos"}, {"m/f", "ub", singular}, {definite, singular}),
        "adv": Tags({"adv"}, set(), set()),
    }


VARIETIES = {"nb_NO": ("lemma.txt", "fullformsliste.txt", parts("ent", "be")), "nn_NO": ("lemma_2012.txt", "fullformer_2012.txt", parts("eint", "bu"))}


def lines(tarball: bytes, name: str) -> Iterator[list[str]]:
    """The cells of each line below the header of one file in the tarball, which is Latin-1 with CRLF line ends."""
    with tarfile.open(fileobj=io.BytesIO(tarball)) as tar:
        member = tar.extractfile(name)
        if member is None:
            raise ValueError(f"the tarball holds no file {name}")
        text = member.read().decode("latin-1").replace("\r", "")
    for line in text.split("\n")[1:]:
        if line:
            yield line.split("\t")


def paradigms(tarball: bytes, name: str, parts: Mapping[str, Tags]) -> dict[tuple[str, str, str], tuple[set[str], list[str]]]:
    """Per lemma, part of speech and paradigm, in the order the list first names them, the spellings of the word and those of its form.

    A form counts while the norm holds it: its end date is 4000 and it is normert. A verb's passive never counts."""
    found: dict[tuple[str, str, str], tuple[set[str], list[str]]] = {}
    for _, lemma, form, tag, paradigm, _, _, until, norm in lines(tarball, name):
        tags = set(tag.split())
        if until != "4000" or norm != "normert" or "pass" in tags:
            continue
        for part, wanted in parts.items():
            if wanted.every <= tags:
                words, forms = found.setdefault((lemma, part, paradigm), (set(), []))
                if wanted.word <= tags:
                    words.add(form)
                if wanted.form and wanted.form <= tags:
                    forms.append(form)
    return found


def candidates(tarball: bytes, variety: str) -> Iterator[wordtables.Candidate]:
    lemmas_file, forms_file, wanted = VARIETIES[variety]
    lemmas = {lemma: word for _, lemma, word, *_ in lines(tarball, lemmas_file)}
    for (lemma, part, _), (words, forms) in paradigms(tarball, forms_file, wanted).items():
        word = lemmas.get(lemma, "")
        if word in words:
            yield part, word, forms[0] if forms else ""


if __name__ == "__main__":
    wordtables.main(
        __doc__,
        "no",
        OUT,
        FORMS,
        [
            wordtables.Source(flag="bokmal", url=BOKMAL, magic=b"\x1f\x8b", folder="nb_NO", candidates=partial(candidates, variety="nb_NO")),
            wordtables.Source(flag="nynorsk", url=NYNORSK, magic=b"\x1f\x8b", folder="nn_NO", candidates=partial(candidates, variety="nn_NO")),
        ],
    )
