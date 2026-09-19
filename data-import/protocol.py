#!/usr/bin/env python3
"""Rebuild data/misc/protocol.tsv from the IANA Protocol Numbers registry.

    data-import/protocol.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

A number the registry leaves unassigned, reserves or marks deprecated does not ship.
Where it names no protocol beside the keyword, the keyword is the name.
"""
import argparse
import csv
import io
from pathlib import Path

import source
import tsv

SOURCE = "https://www.iana.org/assignments/protocol-numbers/protocol-numbers-1.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "protocol.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["keyword", "name", "number"]
SKIP = ("Unassigned", "Reserved", "deprecated")


def rows(text):
    for r in csv.DictReader(io.StringIO(text)):
        number, keyword = r["Decimal"].strip(), r["Keyword"].strip()
        if not number.isdigit() or not keyword or any(s in keyword for s in SKIP):
            continue
        yield {"keyword": keyword, "name": " ".join(r["Protocol"].split()) or keyword, "number": number}


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = sorted(rows(source.fetch(a.source, a.cache, "protocol-numbers-1.csv").decode("utf-8")), key=lambda r: int(r["number"]))
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
