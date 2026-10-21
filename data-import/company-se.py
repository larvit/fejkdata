#!/usr/bin/env python3
"""Rebuild data/sv_SE/company-word.tsv from Bolagsverket's company register, and print what company.json weighs by.

    data-import/company-se.py [--register URL_OR_FILE] [--statistics URL_OR_FILE] [--cache DIR] [--out DIR]

The register (bolagsverket_bulkfil.zip, 250 MB) names no licence; the script reads it for counts and
writes no row of it. A word ships when its count is at least WORDS_MIN and it sits in the names of at
least three of the legal forms AB, HB, KB and E, FORM_MIN times in each, so a word one company family
repeats stays out. The printed counts are company.json's weights: the forms from the statistics
(CC BY 2.5 SE), then for each form the legal-form wording and the shapes of the names, a shape being
the kinds of their words.
"""
import argparse
import csv
import io
import re
import sys
import zipfile
from collections import Counter, defaultdict
from pathlib import Path

import source
import tsv

REGISTER = "https://vardefulla-datamangder.bolagsverket.se/bolagsverket/bolagsverket_bulkfil.zip"
STATISTICS = "https://www.bolagsverket.se/statistik/ftgstat_oppna.csv"
ROOT = Path(__file__).resolve().parent.parent / "data"
OUT = ROOT / "sv_SE"
CACHE = Path(__file__).resolve().parent / "cache"
FORMS = ("AB", "HB", "KB", "BRF", "EK", "E")
STATISTICS_YEAR = "2025"
WORDS_MIN = 100
FORM_MIN = 10
WORD_FORMS = {"AB", "HB", "KB", "E"}
WORD = re.compile(r"^[a-zåäöéü]{3,}$")
FUNCTION_WORDS = {"och", "the", "and", "of", "for", "för", "med", "på", "till", "av", "ett", "den", "det", "mot", "from", "with"}
LEGAL_WORDS = {
    "ab", "aktiebolag", "aktiebolaget", "hb", "handelsbolag", "handelsbolaget", "kb", "kommanditbolag", "kommanditbolaget",
    "bostadsrättsförening", "bostadsrättsföreningen", "ekonomisk", "ekonomiska", "förening", "föreningen", "publ",
}
CONNECTIVES = {"och", "&", "i"}


def column(path):
    """The first column of a shipped table, lower case."""
    return {line.split("\t")[0].lower() for line in Path(path).read_text(encoding="utf-8").splitlines()[1:]}


def registrations(data):
    """Each legal form's new registrations in STATISTICS_YEAR: handelse 1 of Bolagsverket's statistics."""
    counts = Counter()
    for row in csv.DictReader(io.StringIO(data.decode("latin-1"))):
        if row["ar"] == STATISTICS_YEAR and row["handelse"] == "1":
            for form in FORMS + ("FL", "I"):
                counts[form] += int(row[form].strip() or 0)
    return counts


def register_names(data):
    """(legal form, company name) of every active company: the first name of its register row with a form in FORMS."""
    z = zipfile.ZipFile(io.BytesIO(data))
    rows = csv.reader(io.TextIOWrapper(z.open(z.namelist()[0]), encoding="utf-8", newline=""), delimiter=";")
    next(rows)
    for row in rows:
        form = row[4].removesuffix("-ORGFO")
        if row[5] or form not in FORMS:
            continue
        for part in row[3].split("|"):
            if "$FORETAGSNAMN-ORGNAM$" in part:
                yield form, part.split("$")[0]
                break


def split(name):
    return [w for w in re.split(r"[\s,]+", name) if w]


def frequent_words(names, known):
    """The words of the names that are neither a name nor a place, with their counts, by the rule in the docstring."""
    counts, forms = Counter(), defaultdict(Counter)
    for form, name in names:
        for w in split(name):
            w = w.lower()
            if w not in known and WORD.match(w) and w not in FUNCTION_WORDS | LEGAL_WORDS:
                counts[w] += 1
                forms[w][form] += 1
    return {w: c for w, c in counts.items() if c >= WORDS_MIN and sum(forms[w][f] >= FORM_MIN for f in WORD_FORMS) >= 3}


def legal_wording(form, words):
    """(wording, the words left) of a name whose legal-form wording sits where the form's names put it, else None."""
    low = [w.lower() for w in words]
    if form == "AB" and low[-1:] in (["ab"], ["aktiebolag"]):
        return low[-1], words[:-1]
    if form in ("HB", "KB"):
        word = {"HB": "handelsbolag", "KB": "kommanditbolag"}[form]
        if low[-1:] == [word]:
            return "after", words[:-1]
        if low[:1] == [word + "et"]:
            return "before", words[1:]
    if form == "BRF" and low[:1] == ["bostadsrättsföreningen"]:
        return "before", words[1:]
    if form == "EK" and low[-2:] == ["ekonomisk", "förening"]:
        return "ekonomisk förening", words[:-2]
    if form == "EK" and low[-2:] == ["ek.", "för."]:
        return "ek. för.", words[:-2]
    if form == "E":
        return "", words
    return None


def kind(word, surnames, firsts, places, trade):
    w = word.lower()
    if w in surnames:
        return "S"
    if w.endswith("s") and w[:-1] in surnames:
        return "Ss"
    if w in firsts:
        return "F"
    if w in places:
        return "P"
    if w in trade:
        return "T"
    if w in CONNECTIVES:
        return w
    return "N" if w.isdigit() else "X"


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--register", default=REGISTER)
    p.add_argument("--statistics", default=STATISTICS)
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    statistics = source.fetch(a.statistics, a.cache, "bolagsverket-ftgstat.csv")
    register = source.fetch(a.register, a.cache, "bolagsverket-bulkfil.zip", magic=b"PK")
    surnames, firsts, places = (column(ROOT / path) for path in ("sv_SE/last-name.tsv", "sv_SE/first-name.tsv", "geo/SE/locality.tsv"))
    names = list(register_names(register))
    trade = frequent_words(names, surnames | firsts | places)
    tsv.write(Path(a.out) / "company-word.tsv", ["word", "count"], [{"word": w.capitalize(), "count": c} for w, c in sorted(trade.items())])

    print(f"new registrations {STATISTICS_YEAR}: {dict(registrations(statistics))}", file=sys.stderr)
    wording, shapes = {f: Counter() for f in FORMS}, {f: Counter() for f in FORMS}
    for form, name in names:
        words = split(name)
        found = legal_wording(form, words) if words else None
        if found:
            wording[form][found[0]] += 1
            shapes[form][" ".join(kind(w, surnames, firsts, places, trade) for w in found[1])] += 1
    for form in FORMS:
        total = sum(shapes[form].values())
        print(f"{form}: {total} names, wording {dict(wording[form])}", file=sys.stderr)
        for shape, count in shapes[form].most_common(16):
            print(f"  {count:7} {count / total:5.1%} {shape}", file=sys.stderr)


if __name__ == "__main__":
    main()
