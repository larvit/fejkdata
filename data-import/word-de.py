"""Rebuild data/de_DE/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Wikidata's German lexemes.

    data-import/word-de.py [--lexemes URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

Wikidata's lexemes (CC0) give each German word its lexical category and its forms, each tagged with grammatical
features. A row is a noun, a verb, an adjective or an adverb whose lemma is letters alone, a noun's capitalised and
the others' lower-case. A noun carries its first nominative plural, and one without a plural in letters alone is
dropped. A word on LDNOOBW's German list of rude words, compared in lower case, is dropped, and so is a noun whose plural is on
it. Wikidata gives no frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
from functools import partial
from pathlib import Path

import wikidata_lexemes
import wordtables

OUT = Path(__file__).resolve().parent.parent / "data" / "de_DE"
FORMS = {"noun": "plural", "verb": "", "adj": "", "adv": ""}


def spelled(word: str, part: str) -> bool:
	"""A noun is a capital and lower-case letters, any other part lower-case letters alone."""
	head, tail = (word[:1], word[1:]) if part == "noun" else ("", word)
	return word.isalpha() and (not head or head.isupper()) and tail.islower()


candidates = partial(wikidata_lexemes.candidates, language="Q188", code="de", forms=FORMS)


if __name__ == "__main__":
	source = wordtables.Source(flag="lexemes", url=wikidata_lexemes.DUMP, name="wikidata-lexemes.json.gz", magic=b"\x1f\x8b", candidates=candidates)
	wordtables.main(__doc__, "de", OUT, FORMS, [source], spelled)
