#!/usr/bin/env python3
"""Rebuild data/misc/language.tsv from datasets/language-codes (PDDL), the LoC ISO 639-2 register.

    data-import/language.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

Only the 639-2 entries carrying a 639-1 code ship, since the alpha-2 code is the key.
"""
import argparse
import csv
import io
import re
from pathlib import Path

import source
import tsv

SOURCE = "https://raw.githubusercontent.com/datasets/language-codes/main/data/language-codes-full.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "language.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["code", "code3", "name"]


def name(english):
    """The register lists every synonym; the first, without its qualifier, is the one a selector spells."""
    return re.sub(r"\s*\([^)]*\)", "", english.split(";")[0]).strip()


def rows(text):
    for r in csv.DictReader(io.StringIO(text)):
        if r["alpha2"]:
            yield {"code": r["alpha2"], "code3": r["alpha3-t"] or r["alpha3-b"], "name": name(r["English"])}


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = rows(source.fetch(a.source, a.cache, "language-codes-full.csv").decode("utf-8"))
    tsv.write(a.out, COLUMNS, sorted(table, key=lambda r: r["code"]))


if __name__ == "__main__":
    main()
