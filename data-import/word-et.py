"""Rebuild data/et_EE/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Wikidata's Estonian lexemes.

    data-import/word-et.py [--lexemes URL_OR_FILE] [--cache DIR] [--out DIR]

Wikidata's lexemes (CC0) give each Estonian word its lexical category and its forms, each tagged with grammatical
features. A row is a noun, a verb, an adjective or an adverb whose lemma is lower-case letters alone. Wikidata lemmatises a
verb as its `-ma` infinitive. A noun carries its first nominative plural, and one without a plural in lower-case letters alone is
dropped. No word is dropped for being rude. Wikidata gives no frequency, so each word is drawn evenly, and `word` weighs a
part of speech by its row count.
"""
from functools import partial
from pathlib import Path

import wikidata_lexemes
import wordtables

OUT = Path(__file__).resolve().parent.parent / "data" / "et_EE"
FORMS = {"noun": "plural", "verb": "", "adj": "", "adv": ""}
candidates = partial(wikidata_lexemes.candidates, language="Q9072", code="et", forms=FORMS)

if __name__ == "__main__":
    source = wordtables.Source(flag="lexemes", url=wikidata_lexemes.DUMP, name="wikidata-lexemes.json.gz", magic=b"\x1f\x8b", candidates=candidates)
    wordtables.main(__doc__, "", OUT, FORMS, [source], wordtables.lower_letters)
