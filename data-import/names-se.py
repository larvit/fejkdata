#!/usr/bin/env python3
"""Rebuild data/sv_SE/first-name.tsv and last-name.tsv from SCB's whole-population name counts (CC0, "Källa: SCB").

    data-import/names-se.py [--source URL_OR_FILE] [--first N] [--last N] [--cache DIR] [--out DIR]

A tilltalsnamn goes under each sex that carries it, so a name both sexes carry has two rows.
"""
import argparse
import re
from pathlib import Path

import tsv

SOURCE = "https://www.scb.se/contentassets/9fe7dbb460994c72b835163dbc491ef9/namn-med-minst-tva-barare-31-december-2022.xlsx"
OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE"
CACHE = Path(__file__).resolve().parent / "cache"
SHEETS = {"f": "Tilltalsnamn kvinnor", "m": "Tilltalsnamn män"}
SURNAMES = "Efternamn"
NAME = re.compile(r"^[^\W\d_]{2,}([ '-][^\W\d_]{2,})*$")
PARTICLES = {"af", "av", "de", "den", "der", "di", "du", "la", "le", "van", "von"}


def cased(name):
    """SCB's uppercase name as it is written: each part capitalised, a particle before another part lowercased."""
    parts = name.lower().split(" ")
    return " ".join(w if w in PARTICLES and len(parts) > 1 else w.title() for w in parts)


def counted(rows, top):
    """The top names of a sheet by bearers, initials and single letters dropped."""
    names = [(cells[0], int(cells[1])) for cells in rows if len(cells) >= 2 and cells[1].isdigit() and NAME.match(cells[0])]
    return sorted(names, key=lambda n: (-n[1], n[0]))[:top]


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--first", type=int, default=2000, help="names per sex")
    p.add_argument("--last", type=int, default=5000)
    p.add_argument("--out", default=str(OUT))
    p.add_argument("--source", default=SOURCE)
    a = p.parse_args()
    data = tsv.fetch(a.source, a.cache, "scb-namn-2022.xlsx", magic=b"PK")
    first = [{"name": cased(n), "sex": sex, "count": c} for sex, sheet in SHEETS.items() for n, c in counted(tsv.xlsx_rows(data, sheet), a.first)]
    last = [{"name": cased(n), "count": c} for n, c in counted(tsv.xlsx_rows(data, SURNAMES), a.last)]
    out = Path(a.out)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], sorted(first, key=lambda r: (r["name"], r["sex"])))
    tsv.write(out / "last-name.tsv", ["name", "count"], sorted(last, key=lambda r: r["name"]))


if __name__ == "__main__":
    main()
