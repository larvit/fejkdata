#!/usr/bin/env python3
"""Rebuild data/misc/protocol.tsv from the IANA Protocol Numbers registry.

    data-import/protocol.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

A row needs a number and a keyword, which drops the seven numbers assigned to a class of
protocols rather than to one; a keyword the registry reserves or marks deprecated goes
too. Where the registry names no protocol beside the keyword, the keyword is the name.
"""
import argparse
import csv
import io
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://www.iana.org/assignments/protocol-numbers/protocol-numbers-1.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "protocol.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["keyword", "name", "number"]
SKIP = ("Reserved", "deprecated")


def rows(text):
    for r in csv.DictReader(io.StringIO(text)):
        number, keyword = (r["Decimal"] or "").strip(), (r["Keyword"] or "").strip()
        if not number.isdigit() or not keyword or any(s in keyword for s in SKIP):
            continue
        if not 0 <= int(number) <= 255:
            sys.exit(f"{keyword}: protocol number {number} is outside 0-255")
        yield {"keyword": keyword, "name": " ".join((r["Protocol"] or "").split()) or keyword, "number": number}


def refuse_a_name_the_loader_would(table):
    """A name spelling another row's keyword, or a second row's name, is a load error for every consumer."""
    keywords, seen = {r["keyword"] for r in table}, {}
    for r in table:
        if r["name"] != r["keyword"] and r["name"] in keywords:
            sys.exit(f"{r['keyword']}: name {r['name']!r} is another protocol's keyword")
        if r["name"] in seen:
            sys.exit(f"{r['keyword']}: name {r['name']!r} repeats {seen[r['name']]}'s")
        seen[r["name"]] = r["keyword"]


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = sorted(rows(source.fetch(a.source, a.cache, "protocol-numbers-1.csv").decode("utf-8")), key=lambda r: int(r["number"]))
    refuse_a_name_the_loader_would(table)
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
