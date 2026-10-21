#!/usr/bin/env python3
"""Rebuild data/sv_SE/company-word.tsv and company.json from Bolagsverket's company register and statistics.

    data-import/company-se.py [--register URL_OR_FILE] [--statistics URL_OR_FILE] [--cache DIR] [--out DIR]

The register (bolagsverket_bulkfil.zip, 250 MB) names no licence; the script reads it for counts and
writes no row of it. The statistics are CC BY 2.5 SE. company.json weighs a legal form by its new
registrations in STATISTICS_YEAR. For each form it writes the wordings its active companies' names carry
and the shapes of those names, a shape being the kind of each word: a surname, a given name, a place, a
number, a trade word or a coined one. A wording or shape ships when it is at least SHARE_MIN of the form's
names. company-word.tsv holds the trade words: a word that is neither a name nor a place, that WORDS_MIN or
more names carry, that ends a name in a third of its uses, and that FORM_MIN names carry in each of at
least three of the forms AB, HB, KB and E, so a word one company family repeats stays out.
"""
import argparse
import csv
import io
import re
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
STATISTICS_YEAR = "2025"
FORMS = ("AB", "E", "HB", "BRF", "EK", "KB")
SHARE_MIN = 0.01
WORDS_MIN = 100
FORM_MIN = 10
WORD_FORMS = {"AB", "HB", "KB", "E"}
WORD = re.compile(r"^[a-zåäöéü]{3,}$")
FUNCTION_WORDS = {"och", "the", "and", "of", "for", "för", "med", "på", "till", "av", "ett", "den", "det", "mot", "from", "with"}
LEGAL_WORDS = {
    "ab", "aktiebolag", "aktiebolaget", "hb", "handelsbolag", "handelsbolaget", "kb", "kommanditbolag", "kommanditbolaget",
    "bostadsrättsförening", "bostadsrättsföreningen", "ekonomisk", "ekonomiska", "förening", "föreningen", "publ",
}
ENDINGS = {"AB": [["ab"], ["aktiebolag"]], "HB": [["handelsbolag"]], "KB": [["kommanditbolag"]], "EK": [["ekonomisk", "förening"], ["ek.", "för."]]}
BEGINNINGS = {"HB": "handelsbolaget", "KB": "kommanditbolaget", "BRF": "bostadsrättsföreningen"}
SLOTS = {
    "X": "{.company-coined}", "T": "{.company-word}", "S": "{.last-name}", "Ss": "{.last-name}s", "F": "{.first-name}",
    "P": "{/geo.SE.locality}", "N": "{int(1,30)}", "i": "i", "&": "&", "och": "och",
}


def column(path):
    """The first column of a shipped table, lower case."""
    return {line.split("\t")[0].lower() for line in Path(path).read_text(encoding="utf-8").splitlines()[1:]}


def registrations(data):
    """Each legal form's new registrations in STATISTICS_YEAR: handelse 1 of Bolagsverket's statistics."""
    counts = Counter()
    for row in csv.DictReader(io.StringIO(data.decode("latin-1"))):
        if row["ar"] == STATISTICS_YEAR and row["handelse"] == "1":
            for form in FORMS:
                counts[form] += int(row[form].strip() or 0)
    return counts


def register_names(data):
    """(legal form, company name) of every active company of FORMS: the first name of its register row."""
    z = zipfile.ZipFile(io.BytesIO(data))
    for row in csv.DictReader(io.TextIOWrapper(z.open(z.namelist()[0]), encoding="utf-8", newline=""), delimiter=";"):
        form = row["organisationsform"].removesuffix("-ORGFO")
        if row["avregistreringsdatum"] or form not in FORMS:
            continue
        for part in row["organisationsnamn"].split("|"):
            if "$FORETAGSNAMN-ORGNAM$" in part:
                yield form, part.split("$")[0]
                break


def split(name):
    return [w for w in re.split(r"[\s,]+", name) if w]


def kind(word, surnames, firsts, places):
    """S, Ss, F or P for a surname, its genitive, a given name or a place; else None."""
    w = word.lower()
    if w in surnames:
        return "S"
    if w.endswith("s") and w[:-1] in surnames:
        return "Ss"
    if w in firsts:
        return "F"
    if w in places:
        return "P"
    return None


def trade_words(names, named):
    """The trade words, each with its spelling most names use and its count, by the rule in the docstring."""
    counts, ending, forms, spellings = Counter(), Counter(), defaultdict(Counter), defaultdict(Counter)
    for form, name in names:
        words = [w for w in split(name) if w.lower() not in LEGAL_WORDS]
        for i, w in enumerate(words):
            low = w.lower()
            if WORD.match(low) and low not in FUNCTION_WORDS and not named(low):
                counts[low] += 1
                forms[low][form] += 1
                ending[low] += i == len(words) - 1
                if not name.isupper():
                    spellings[low][w[0].upper() + w[1:]] += 1
    return {
        w: (spellings[w].most_common(1)[0][0] if spellings[w] else w.capitalize(), c)
        for w, c in counts.items()
        if c >= WORDS_MIN and 3 * ending[w] >= c and sum(forms[w][f] >= FORM_MIN for f in WORD_FORMS) >= 3
    }


def legal_wording(form, words):
    """(position, wording as written, the words left) of a name carrying its form's wording where the form's names put it, else None."""
    low = [w.lower() for w in words]
    if form == "E":
        return "", "", words
    for end in ENDINGS.get(form, []):
        if low[-len(end):] == end:
            return "after", " ".join(words[-len(end):]), words[:-len(end)]
    if form in BEGINNINGS and low[0] == BEGINNINGS[form]:
        return "before", words[0], words[1:]
    return None


def shape_of(words, surnames, firsts, places, trade):
    """The kinds of the words, as SLOTS keys."""
    shape = []
    for w in words:
        low = w.lower()
        if low.isdigit():
            shape.append("N")
        elif low in ("och", "&", "i"):
            shape.append(low)
        else:
            shape.append(kind(low, surnames, firsts, places) or ("T" if low in trade else "X"))
    return tuple(shape)


def count_names(names, surnames, firsts, places, trade):
    """Per form: the wordings counted over names not in capitals, and the shapes counted over every name."""
    wordings = {f: Counter() for f in FORMS}
    shapes = {f: Counter() for f in FORMS}
    for form, name in names:
        words = split(name)
        found = legal_wording(form, words) if words else None
        if found and found[2]:
            if not name.isupper():
                wordings[form][(found[0], found[1])] += 1
            shapes[form][shape_of(found[2], surnames, firsts, places, trade)] += 1
    return wordings, shapes


def common(counts):
    """The keys with at least SHARE_MIN of the total, by count then key."""
    total = sum(counts.values())
    return [(k, c) for k, c in sorted(counts.items(), key=lambda kc: (-kc[1], kc[0])) if c >= SHARE_MIN * total]


def quoted(s):
    return '"' + s.replace("\\", "\\\\").replace('"', '\\"') + '"'


def weighted(fmt, weight):
    return "{ " + f'"format": {quoted(fmt)}, "weight": {weight}' + " }"


def lines(key, items):
    return [f'    "{key}": [', ",\n".join("      " + i for i in items), "    ]"]


def form_block(form, wordings, shapes):
    """One form's object: its names by shape, and, when it has wordings, the company as written."""
    low = form.lower()
    names = lines("name", [weighted(" ".join(SLOTS[k] for k in shape), c) for shape, c in common(shapes)])
    if not wordings:
        return "\n".join([f'  "{low}": {{', '    "format": "{name}",', *names, "  }"])
    ref = "{.company.%s.name}" % low
    written = lines("written", [weighted(f"{ref} {w}" if where == "after" else f"{w} {ref}", c) for (where, w), c in common(wordings)])
    return "\n".join([f'  "{low}": {{', '    "format": "{written}",', *names[:-1], names[-1] + ",", *written, "  }"])


def company_json(registered, wordings, shapes):
    forms = [weighted("{.company.%s}" % f.lower(), registered[f]) for f in sorted(FORMS, key=lambda f: -registered[f])]
    blocks = [form_block(f, wordings[f] if f != "E" else None, shapes[f]) for f in FORMS]
    return '{\n  "format": "{registered}",\n  "registered": [\n    ' + ",\n    ".join(forms) + "\n  ],\n" + ",\n".join(blocks) + "\n}\n"


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
    trade = trade_words(names, lambda w: kind(w, surnames, firsts, places))
    out = Path(a.out)
    tsv.write(out / "company-word.tsv", ["word", "count"], [{"word": s, "count": c} for _, (s, c) in sorted(trade.items())])
    wordings, shapes = count_names(names, surnames, firsts, places, trade)
    (out / "company.json").write_text(company_json(registrations(statistics), wordings, shapes), encoding="utf-8")


if __name__ == "__main__":
    main()
