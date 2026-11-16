"""Rebuild nb_NO's and nn_NO's word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Norsk ordbank.

    data-import/word-no.py [--bokmal URL_OR_FILE] [--nynorsk URL_OR_FILE] [--cache DIR] [--out DIR]

Norsk ordbank (Språkbanken, CC BY 4.0) lists every form of each lemma, tagged. A row is a common noun, a verb, an
adjective or an adverb that the ordbank spells in lower-case letters alone, with a form from the same paradigm: a noun's
definite singular, a verb's present and an adjective's definite singular. Only the forms the current norm holds count.
A lemma lacking that form is dropped, and so is one whose word or form is on LDNOOBW's Norwegian list of rude words.
The ordbank gives no frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import argparse
import io
import tarfile
from collections.abc import Collection, Iterator, Mapping
from pathlib import Path

import source
import wordtables

BOKMAL = "https://www.nb.no/sbfil/leksikalske_databaser/ordbank/20220201_norsk_ordbank_nob_2005.tar.gz"
NYNORSK = "https://www.nb.no/sbfil/leksikalske_databaser/ordbank/20220201_norsk_ordbank_nno_2012.tar.gz"
OUT = Path(__file__).resolve().parent.parent / "data"
CACHE = Path(__file__).resolve().parent / "cache"
FORMS = {"noun": "definite", "verb": "present", "adj": "definite", "adv": ""}


def parts(singular: str, definite: str) -> dict[str, tuple[set[str], set[str], set[str]]]:
    """Per part of speech, the tags its every form carries, then those of its word and of its form.

    Bokmål tags a singular `ent` and a definite `be`, nynorsk `eint` and `bu`."""
    return {
        "noun": ({"subst", "appell"}, {singular, "ub"}, {singular, definite}),
        "verb": ({"verb"}, {"inf"}, {"pres"}),
        "adj": ({"adj", "pos"}, {"m/f", "ub", singular}, {definite, singular}),
        "adv": ({"adv"}, set(), set()),
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


def paradigms(tarball: bytes, name: str, tags: Mapping[str, tuple[set[str], set[str], set[str]]]) -> dict[tuple[str, str, str], tuple[set[str], list[str]]]:
    """Per lemma, part of speech and paradigm, in the order the list first names them, the spellings of the word and those of its form.

    A form counts while the norm holds it, its end date 4000 and it normert, and a verb's passive never counts."""
    found: dict[tuple[str, str, str], tuple[set[str], list[str]]] = {}
    for _, lemma, form, tag, paradigm, _, _, until, norm in lines(tarball, name):
        held = set(tag.split())
        if until != "4000" or norm != "normert" or "pass" in held:
            continue
        for part, (every, word, inflected) in tags.items():
            if every <= held:
                words, forms = found.setdefault((lemma, part, paradigm), (set(), []))
                if word <= held:
                    words.add(form)
                if inflected and inflected <= held:
                    forms.append(form)
    return found


def tables(tarball: bytes, variety: str, rude: Collection[str]) -> dict[str, list[dict[str, str]]]:
    lemmas_file, forms_file, tags = VARIETIES[variety]
    lemmas = {lemma: word for _, lemma, word, *_ in lines(tarball, lemmas_file)}
    rows: dict[str, dict[str, dict[str, str]]] = {part: {} for part in FORMS}
    for (lemma, part, _), (words, forms) in paradigms(tarball, forms_file, tags).items():
        word, column = lemmas.get(lemma, ""), FORMS[part]
        form = forms[0] if forms else ""
        if word not in words or not word.isalpha() or not word.islower() or word in rude:
            continue
        if not column:
            rows[part].setdefault(word, {"word": word})
        elif form.isalpha() and form.islower() and form not in rude:
            rows[part].setdefault(word, {"word": word, column: form})
    return {part: [held[w] for w in sorted(held)] for part, held in rows.items()}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--cache", default=str(CACHE))
    parser.add_argument("--bokmal", default=BOKMAL)
    parser.add_argument("--nynorsk", default=NYNORSK)
    parser.add_argument("--rude", default="")
    parser.add_argument("--out", default=str(OUT))
    args = parser.parse_args()
    rude = wordtables.rude("no", args.cache, args.rude)
    for variety, url in (("nb_NO", args.bokmal), ("nn_NO", args.nynorsk)):
        tarball = source.fetch(url, args.cache, Path(url).name, magic=b"\x1f\x8b")
        wordtables.write(Path(args.out) / variety, FORMS, tables(tarball, variety, rude))


if __name__ == "__main__":
    main()
