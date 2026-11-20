"""Rebuild data/lv_LV/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from Tēzaurs.lv's wordforms.

    data-import/word-lv.py [--wordforms URL_OR_FILE] [--cache DIR] [--out DIR]

Tēzaurs.lv (CC BY-SA 4.0) gives each Latvian lemma its part of speech, its usage labels and its inflected forms. A row is a
noun, verb, adjective or adverb whose lemma is lower-case letters alone and carries no label that marks it as regional or
out of date. A verb is an infinitive, ending in `-t` or `-ties`, since Tēzaurs.lv files some participles as verbs. A noun
carries its nominative plural, and one without it is dropped. No word is dropped for being rude. Tēzaurs.lv gives no
frequency, so each word is drawn evenly, and `word` weighs a part of speech by its row count. The wordforms file is 14 GB
unzipped, so it is read as a stream.
"""
import io
import zipfile
from collections.abc import Iterator
from pathlib import Path

import wordtables

OUT = Path(__file__).resolve().parent.parent / "data" / "lv_LV"
URL = "https://repository.clarin.lv/repository/xmlui/bitstream/handle/20.500.12574/156/tezaurs_2026_2_wordforms_tei.xml.zip?sequence=5&isAllowed=y"
FORMS = {"noun": "plural", "verb": "", "adj": "", "adv": ""}
PARTS = {"Lietvārds": "noun", "Darbības vārds": "verb", "Īpašības vārds": "adj", "Apstākļa vārds": "adv"}
INFINITIVE = ("t", "ties")
OUT_OF_USE = {"Apvidvārds", "Dialektisms", "Novecojis", "Vēsturisks", "Neaktuāls", "Sens, reti lietots vārds"}


def text(line: bytes) -> str:
    """What an element written on one line holds, such as `mājas` of `<orth>mājas</orth>`."""
    return line[line.index(b">") + 1 : line.rindex(b"<")].decode("utf-8")


def kind(line: bytes) -> str:
    """The `type` of a `<gram type="…">` line."""
    return line[12 : line.index(b'"', 12)].decode("utf-8")


def candidates(dump: bytes) -> Iterator[wordtables.Candidate]:
    """A part of speech, lemma and nominative plural per noun, verb, adjective or adverb of the wordforms zip that no label marks as regional or out of date, and per verb only an infinitive; the plural is empty for any other part and for a noun with none.

    Each tag stands on a line of its own. An entry's lemma and properties come first, then an `inflection` form per case and number and a `derivative` form per participle or adverb it makes. Each form has its own `<orth>` and part of speech, so only the first of each is the lemma's."""
    lemma = part = plural = case = number = form = ""
    labels: set[str] = set()
    inflected = False
    with zipfile.ZipFile(io.BytesIO(dump)) as z, z.open(z.infolist()[0]) as lines:
        for raw in lines:
            line = raw.strip()
            if line.startswith(b"<entry "):
                lemma = part = plural = ""
                labels = set()
                inflected = False
            elif line == b'<form type="inflection">':
                inflected = True
                form = case = number = ""
            elif line == b"</form>" and inflected:
                inflected = False
                if case == "Nominatīvs" and number == "Daudzskaitlis" and not plural:
                    plural = form
            elif line.startswith(b"<orth"):
                if inflected:
                    form = text(line)
                elif not lemma:
                    lemma = text(line)
            elif line.startswith(b"<gram "):
                gram, value = kind(line), text(line)
                if inflected and gram == "Locījums":
                    case = value
                elif inflected and gram == "Skaitlis":
                    number = value
                elif gram == "Vārdšķira" and not part:
                    part = value
                elif gram == "Lietojums" and not inflected:
                    labels.add(value)
            elif line == b"</entry>" and part in PARTS and not labels & OUT_OF_USE:
                kept = PARTS[part]
                if kept != "verb" or lemma.endswith(INFINITIVE):
                    yield kept, lemma, plural if FORMS[kept] else ""


if __name__ == "__main__":
    source = wordtables.Source(flag="wordforms", url=URL, name="tezaurs-wordforms.xml.zip", magic=b"PK", candidates=candidates)
    wordtables.main(__doc__, "", OUT, FORMS, [source], wordtables.lower_letters)
