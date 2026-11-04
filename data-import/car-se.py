#!/usr/bin/env python3
"""Rebuild data/sv_SE/car.tsv from Mobility Sweden's report of new passenger car registrations (Källa: Mobility Sweden).

    data-import/car-se.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

The source is the December 2025 report, sheet "PB - Fabrikat och modell": each make with the models it
registered and their registrations from January to December 2025. A row ships for each model with 500 or
more registrations. The sheet's "Övriga" rows, the models a make leaves ungrouped, are dropped. The sheet
writes a model in capitals after its make, `VW ID.4`: the script drops that prefix, which is the make's
name, `Volkswagen` as `VW` and `Mercedes-Benz` as `MERCEDES`, and gives every run of four or more
capital letters an initial capital, `TIGUAN` as `Tiguan`. Shorter runs, such as `XC60` and `RAV 4`, keep
the sheet's spelling.
"""
import argparse
import re
import sys
from pathlib import Path

import source
import tsv
import xlsx

SOURCE = "https://mobilitysweden.se/storage/B9226CA61408A37CBD098A0E05A5C7AC74D47E6A241F84862CD2E228B08E36DA/b87ae3b4143c419593cc8777e2174f5b/xlsx/media/92ad70f7a5dd444fb2680306bc6016fb/Ma%CC%8Anadsrapport%20Nyregistreringar%20december%202025.xlsx"
SHEET = "PB - Fabrikat och modell"
OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE" / "car.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
MINIMUM = 500
TOTAL = 272998
PREFIXES = {"Volkswagen": ["VW"], "Mercedes-Benz": ["MERCEDES"]}
WORD = re.compile(r"[A-ZÅÄÖ]{4,}")
HEADER = "Fabrikat (Antal reg. YTD↓ )"


def models(sheet):
    """The (make, model, registrations) of each model row, below the sheet's header, until its grand total."""
    make = None
    for r in sheet:
        r = r + [""] * (11 - len(r))
        if r[0] and r[0] != HEADER and r[2] == "Total":
            make = r[0]
        elif r[2] == "Totalt":
            if int(r[10]) != TOTAL:
                sys.exit(f"the sheet's grand total is {r[10]}, not {TOTAL}")
            return
        elif make and r[2] and r[2] != "Total":
            yield make, r[2], int(r[10] or 0)
    sys.exit("the sheet has no row 'Totalt'; Mobility Sweden has changed it")


def spelled(make, model):
    """The model without its make's prefix, its long capital words in initial capitals."""
    for prefix in [make.upper(), make.upper().split("-")[0]] + PREFIXES.get(make, []):
        if model.startswith(prefix + " "):
            return WORD.sub(lambda m: m.group().capitalize(), model[len(prefix) + 1:])
    sys.exit(f"{model!r} does not start with the name of its make, {make}")


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    sheet = list(xlsx.rows(source.fetch(a.source, a.cache, "mobility-sweden-2025-12.xlsx", magic=b"PK"), SHEET))
    rows = [{"make": make, "model": spelled(make, model), "registrations": n} for make, model, n in models(sheet) if n >= MINIMUM and "ÖVRIGA" not in model]
    tsv.write(a.out, ["make", "model", "registrations"], rows)


if __name__ == "__main__":
    main()
