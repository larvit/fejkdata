#!/usr/bin/env python3
"""Rebuild data/nb_NO/company.json, company-coined.json and company-word.{json,tsv} from Enhetsregisteret.

    data-import/company-no.py [--register URL_OR_FILE] [--cache DIR] [--out DIR]

The script reads the register's bulk file (155 MB gzip) for counts and writes no row of it, and the shipped
nb_NO first and last names, so run names-no.py first. It leaves out the units that are bankrupt, being wound up
or forcibly dissolved. company.json weighs a legal form by its units, and for each form writes the shapes of
the names of the units that carry the form's abbreviation. A shape is the kind of each word: a last name, a
first name, a number, `og`, `&`, a trade word, or any other, which draws a coined word. A last or first name
is one that SSB lists, and one it doesn't, such as a rarer owner's, is any other. A shape ships when it is at
least company_names.SHARE_MIN of the form's names. The forms are AS, ASA, DA, ANS, KS, SA and ENK, which is
named for a person and carries no abbreviation. NUF is left out: its names carry the foreign enterprise's own
form, such as AB, LTD or GMBH, and no Norwegian one.

A place is a word of the poststed or kommune that a unit's business address spells; a region such as Agder
isn't one. A place is neither a trade word nor a coined one, and a shape holding one is left out until geo/NO
provides the places to draw.

company-word.tsv holds the trade words, each with its count and its most common spelling, by the rule of
company_names.py over the names of the forms AS, ENK, DA and ANS.

company-coined.json draws `{a}{b}`. `a` holds the words of ROOT_LETTERS letters that begin ROOT_MIN or more
names and end under a third of the names they are in, none of them a name, a place or a trade word. `b` holds
the endings of ENDING_LETTERS letters, starting with a consonant, that ENDING_MIN or more of the register's
places share, less one that a longer ending covers with four fifths of its places.
"""
import argparse
import csv
import gzip
import io
import json
import re
from collections import Counter, namedtuple
from pathlib import Path

import source
import tsv
from company_names import Vocabulary, column, common, quoted, split, trade_words, weighted, word_usage

REGISTER = "https://data.brreg.no/enhetsregisteret/api/enheter/lastned/csv"
ROOT = Path(__file__).resolve().parent.parent / "data"
OUT = ROOT / "nb_NO"
CACHE = Path(__file__).resolve().parent / "cache"
FORMS = ("ANS", "AS", "ASA", "DA", "ENK", "KS", "SA")
NO_WORDING = {"ENK"}
ROOT_MIN = 60
ROOT_LETTERS = range(4, 7)
ENDING_MIN = 40
ENDING_LETTERS = range(3, 6)
VOWELS = "aeiouyæøå"
FUNCTION_WORDS = {"og", "the", "and", "of", "for", "med", "på", "til", "av", "din", "min", "den", "det", "ett", "nye"}
LEGAL_WORDS = {form.lower() for form in FORMS} | {"aksjeselskap", "aksjeselskapet", "ansvarlig", "selskap", "samvirkeforetak"}
VOCABULARY = Vocabulary(re.compile(r"^[a-zæøåéü]{3,}$"), frozenset(FUNCTION_WORDS), frozenset(LEGAL_WORDS), frozenset({"AS", "ENK", "DA", "ANS"}), 0)
Words = namedtuple("Words", "surnames firsts places")
SLOTS = {"X": "{.company-coined}", "T": "{.company-word}", "S": "{.last-name}", "F": "{.first-name}", "N": "{int(1,30)}", "og": "og", "&": "&"}


def register(data):
    """(units per form, (form, name) of every active unit of FORMS, the lower-case places the register spells)."""
    counts, names, places = Counter(), [], set()
    for row in csv.DictReader(io.TextIOWrapper(gzip.GzipFile(fileobj=io.BytesIO(data)), encoding="utf-8", newline="")):
        if row["forretningsadresse.landkode"] == "NO":
            for c in ("forretningsadresse.poststed", "forretningsadresse.kommune"):
                places.update(w for w in re.split(r"[\s,]+", row[c].lower()) if w)
        form = row["organisasjonsform.kode"]
        if form in FORMS and all(row[c] == "false" for c in ("konkurs", "underAvvikling", "underTvangsavviklingEllerTvangsopplosning")):
            counts[form] += 1
            names.append((form, row["navn"]))
    return counts, names, places


def without_wording(form, words):
    """The words left when a name carries its form's abbreviation as its last word, else None."""
    if form in NO_WORDING:
        return words
    if words and words[-1].rstrip(".").upper() == form:
        return words[:-1]
    return None


def kind(word, sets, last):
    """S, F or P for a last name, a first name or a place; else None. A word that is both names is a last name at the end of a name."""
    w = word.lower().replace("é", "e").replace("ü", "u")
    if w in sets.surnames and (last or w not in sets.firsts):
        return "S"
    if w in sets.firsts:
        return "F"
    return "P" if w in sets.places else None


def roots(usage, trade):
    """The words of ROOT_LETTERS letters that begin ROOT_MIN or more names and end under a third of the names they are in, none of them a trade word."""
    return sorted(
        w.capitalize()
        for w, u in usage.items()
        if w not in trade and len(w) in ROOT_LETTERS and u.begins >= ROOT_MIN and 3 * u.ends < u.names
    )


def endings(places):
    """The endings of ENDING_LETTERS letters that ENDING_MIN or more places share and that start with a consonant, less one that a longer ending covers with four fifths of its places."""
    count = Counter(p[-n:] for p in places if re.fullmatch(r"[a-zæøå]+", p) for n in ENDING_LETTERS if len(p) > n + 2)
    kept = {e for e, c in count.items() if c >= ENDING_MIN and e[0] not in VOWELS}
    return sorted(e for e in kept if not any(longer != e and longer.endswith(e) and 5 * count[longer] >= 4 * count[e] for longer in kept))


def shape_of(words, sets, trade):
    """The kinds of the words, as SLOTS keys, or None where a word is a place."""
    shape = []
    for i, w in enumerate(words):
        low = w.lower()
        if low.isdigit():
            shape.append("N")
        elif low in ("og", "&"):
            shape.append(low)
        else:
            found = kind(low, sets, i == len(words) - 1) or ("T" if low in trade else "X")
            if found == "P":
                return None
            shape.append(found)
    return tuple(shape)


def count_shapes(names, sets, trade):
    """Per form: the shapes counted over the names that carry the form's abbreviation, less those holding a place."""
    shapes = {f: Counter() for f in FORMS}
    for form, name in names:
        left = without_wording(form, split(name))
        shape = shape_of(left, sets, trade) if left else None
        if shape:
            shapes[form][shape] += 1
    return shapes


def form_block(form, shapes):
    """One form's object: its name by shape, and the abbreviation after it."""
    items = ",\n".join("      " + weighted(" ".join(SLOTS[k] for k in shape), c) for shape, c in common(shapes))
    fmt = "{name}" if form in NO_WORDING else "{name} " + form
    return f'  "{form.lower()}": {{\n    "format": {quoted(fmt)},\n    "name": [\n{items}\n    ]\n  }}'


def company_json(registered, shapes):
    forms = [weighted("{.company.%s}" % f.lower(), registered[f]) for f in sorted(FORMS, key=lambda f: (-registered[f], f))]
    blocks = [form_block(f, shapes[f]) for f in FORMS]
    return '{\n  "format": "{registered}",\n  "registered": [\n    ' + ",\n    ".join(forms) + "\n  ],\n" + ",\n".join(blocks) + "\n}\n"


def coined_json(a, b):
    return '{\n  "format": "{a}{b}",\n  "a": ' + json.dumps(a, ensure_ascii=False) + ',\n  "b": ' + json.dumps(b, ensure_ascii=False) + "\n}\n"


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--register", default=REGISTER)
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    data = source.fetch(a.register, a.cache, "enhetsregisteret-enheter.csv.gz", magic=b"\x1f\x8b")
    registered, names, places = register(data)
    sets = Words(column(ROOT / "nb_NO/last-name.tsv"), column(ROOT / "nb_NO/first-name.tsv"), places)
    named = lambda w: kind(w, sets, True) is not None or kind(w.removesuffix("s"), sets, True) == "S"
    usage = word_usage(names, named, VOCABULARY)
    trade = trade_words(usage, VOCABULARY)
    out = Path(a.out)
    tsv.write(out / "company-word.tsv", ["word", "count"], [{"word": s, "count": c} for _, (s, c) in sorted(trade.items())])
    (out / "company-word.json").write_text('{ "format": "{word}", "rows": "company-word.tsv", "weight": "count" }\n', encoding="utf-8")
    (out / "company-coined.json").write_text(coined_json(roots(usage, trade), endings(places)), encoding="utf-8")
    (out / "company.json").write_text(company_json(registered, count_shapes(names, sets, trade)), encoding="utf-8")


if __name__ == "__main__":
    main()
