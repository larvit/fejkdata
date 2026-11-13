#!/usr/bin/env python3
"""Rebuild data/sv_SE/company-word.tsv and company.json from Bolagsverket's company register and statistics.

    data-import/company-se.py [--register URL_OR_FILE] [--statistics URL_OR_FILE] [--cache DIR] [--out DIR]

The script reads the register (bolagsverket_bulkfil.zip, 250 MB) for counts and writes no row of it.
company.json weighs a legal form by its new registrations in STATISTICS_YEAR. For each form it writes the
wordings that the names of the form's active companies carry, and the shapes of those names. A shape is the
kind of each word: a surname, its genitive, a given name, a place, a number, `och`, `&` or `i`, a trade word,
or any other, which draws a coined word. A wording or shape ships when it is at least company_names.SHARE_MIN of
the form's names. `registered` holds one item per form and wording, so `legal` reads the wording its company prints.

company-word.tsv holds the trade words, each with its count and its most common spelling, by the rule of
company_names.py over the names of the forms AB, HB, KB and E.
"""
import argparse
import csv
import io
import re
import zipfile
from collections import Counter
from collections.abc import Collection, Iterable, Iterator, Mapping
from pathlib import Path

import source
import tsv
from company_names import Vocabulary, column, common, legal_item, split, trade_words, weighted, word_usage

REGISTER = "https://vardefulla-datamangder.bolagsverket.se/bolagsverket/bolagsverket_bulkfil.zip"
STATISTICS = "https://www.bolagsverket.se/statistik/ftgstat_oppna.csv"
ROOT = Path(__file__).resolve().parent.parent / "data"
OUT = ROOT / "sv_SE"
CACHE = Path(__file__).resolve().parent / "cache"
STATISTICS_YEAR = "2025"
FORMS = ("AB", "E", "HB", "BRF", "EK", "KB")
FUNCTION_WORDS = {"och", "the", "and", "of", "for", "för", "med", "på", "till", "av", "ett", "den", "det", "mot", "from", "with"}
LEGAL_WORDS = {
    "ab", "aktiebolag", "aktiebolaget", "hb", "handelsbolag", "handelsbolaget", "kb", "kommanditbolag", "kommanditbolaget",
    "bostadsrättsförening", "bostadsrättsföreningen", "brf", "ekonomisk", "ekonomiska", "förening", "föreningen", "publ",
}
VOCABULARY = Vocabulary(re.compile(r"^[a-zåäöéü]{3,}$"), frozenset(FUNCTION_WORDS), frozenset(LEGAL_WORDS), frozenset({"AB", "HB", "KB", "E"}), 3)
ENDINGS = {
    "AB": [["ab"], ["aktiebolag"]],
    "HB": [["hb"], ["handelsbolag"]],
    "KB": [["kb"], ["kommanditbolag"]],
    "EK": [["ekonomisk", "förening"], ["ek.", "för."]],
}
BEGINNINGS = {"HB": ["handelsbolaget"], "KB": ["kommanditbolaget"], "BRF": ["bostadsrättsföreningen", "brf"]}
SLOTS = {
    "X": "{.company-coined}", "T": "{.company-word}", "S": "{.last-name}", "Ss": "{.last-name}s", "F": "{.first-name}",
    "P": "{/geo.SE.locality}", "N": "{int(1,30)}", "i": "i", "&": "&", "och": "och",
}

# A name's kinds of word in order, each a key of SLOTS.
type Shape = tuple[str, ...]
# Where a name writes its legal form, "before" or "after", and the legal form as the name spells it.
type Wording = tuple[str, str]


def registrations(data: bytes) -> Counter[str]:
    """Each legal form's new registrations in STATISTICS_YEAR: handelse 1 of Bolagsverket's statistics."""
    counts: Counter[str] = Counter()
    for row in csv.DictReader(io.StringIO(data.decode("latin-1"))):
        if row["ar"] == STATISTICS_YEAR and row["handelse"] == "1":
            for form in FORMS:
                counts[form] += int(row[form].strip() or 0)
    return counts


def register_names(data: bytes) -> Iterator[tuple[str, str]]:
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


def kind(word: str, surnames: Collection[str], firsts: Collection[str], places: Collection[str]) -> str | None:
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


def legal_wording(form: str, words: list[str]) -> tuple[str, str, list[str]] | None:
    """(position, wording as written, the words left) of a name carrying its form's wording where the form's names put it, else None."""
    low = [w.lower() for w in words]
    if form == "E":
        return "", "", words
    for end in ENDINGS.get(form, []):
        if low[-len(end):] == end:
            return "after", " ".join(words[-len(end):]), words[:-len(end)]
    if low[0] in BEGINNINGS.get(form, []):
        return "before", words[0], words[1:]
    return None


def shape_of(words: Iterable[str], surnames: Collection[str], firsts: Collection[str], places: Collection[str], trade: Collection[str]) -> Shape:
    """The kinds of the words, as SLOTS keys."""
    shape: list[str] = []
    for w in words:
        low = w.lower()
        if low.isdigit():
            shape.append("N")
        elif low in ("och", "&", "i"):
            shape.append(low)
        else:
            shape.append(kind(low, surnames, firsts, places) or ("T" if low in trade else "X"))
    return tuple(shape)


def count_names(names: Iterable[tuple[str, str]], surnames: Collection[str], firsts: Collection[str], places: Collection[str], trade: Collection[str]) -> tuple[dict[str, Counter[Wording]], dict[str, Counter[Shape]]]:
    """Per form: the wordings counted over names not in capitals, and the shapes counted over every name."""
    wordings: dict[str, Counter[Wording]] = {f: Counter() for f in FORMS}
    shapes: dict[str, Counter[Shape]] = {f: Counter() for f in FORMS}
    for form, name in names:
        words = split(name)
        found = legal_wording(form, words) if words else None
        if found and found[2]:
            if not name.isupper():
                wordings[form][(found[0], found[1])] += 1
            shapes[form][shape_of(found[2], surnames, firsts, places, trade)] += 1
    return wordings, shapes


def lines(key: str, items: Iterable[str]) -> list[str]:
    return [f'    "{key}": [', ",\n".join("      " + i for i in items), "    ]"]


def written_format(form: str, where: str, wording: str) -> str:
    ref = "{.company.%s.name}" % form.lower()
    return f"{ref} {wording}" if where == "after" else f"{wording} {ref}"


def registered_items(form: str, wordings: Mapping[Wording, int] | None, registered: int) -> list[str]:
    """The items of `registered` for one form: its registrations split by how often each wording is written."""
    if not wordings:
        return [legal_item("{.company.%s}" % form.lower(), "", registered)]
    kept = common(wordings)
    total = sum(c for _, c in kept)
    return [legal_item(written_format(form, where, w), w, round(registered * c / total, 2)) for (where, w), c in kept]


def form_block(form: str, wordings: Mapping[Wording, int] | None, shapes: Mapping[Shape, int]) -> str:
    """One form's object: its names by shape, and, when it has wordings, the company as written."""
    low = form.lower()
    names = lines("name", [weighted(" ".join(SLOTS[k] for k in shape), c) for shape, c in common(shapes)])
    if not wordings:
        return "\n".join([f'  "{low}": {{', '    "format": "{name}",', *names, "  }"])
    written = lines("written", [weighted(written_format(form, where, w), c) for (where, w), c in common(wordings)])
    return "\n".join([f'  "{low}": {{', '    "format": "{written}",', *names[:-1], names[-1] + ",", *written, "  }"])


def company_json(registered: Mapping[str, int], wordings: Mapping[str, Mapping[Wording, int]], shapes: Mapping[str, Mapping[Shape, int]]) -> str:
    forms = [i for f in sorted(FORMS, key=lambda f: -registered[f]) for i in registered_items(f, wordings[f] if f != "E" else None, registered[f])]
    blocks = [form_block(f, wordings[f] if f != "E" else None, shapes[f]) for f in FORMS]
    return '{\n  "format": "{registered as r}{r}",\n  "legal": "{r.legal}",\n  "registered": [\n    ' + ",\n    ".join(forms) + "\n  ],\n" + ",\n".join(blocks) + "\n}\n"


def main() -> None:
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
    trade = trade_words(word_usage(names, lambda w: kind(w, surnames, firsts, places), VOCABULARY), VOCABULARY)
    out = Path(a.out)
    tsv.write(out / "company-word.tsv", ["word", "count"], [{"word": s, "count": c} for _, (s, c) in sorted(trade.items())])
    wordings, shapes = count_names(names, surnames, firsts, places, trade)
    (out / "company.json").write_text(company_json(registrations(statistics), wordings, shapes), encoding="utf-8")


if __name__ == "__main__":
    main()
