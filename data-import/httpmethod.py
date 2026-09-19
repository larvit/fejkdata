#!/usr/bin/env python3
"""Rebuild data/misc/httpmethod.tsv from the IANA HTTP Method Registry.

    data-import/httpmethod.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

A method whose name is not letters and hyphens does not ship, which drops the `*` the
registry holds to stop anyone registering it.
"""
import argparse
import csv
import io
import re
from pathlib import Path

import source
import tsv

SOURCE = "https://www.iana.org/assignments/http-methods/methods.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "httpmethod.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["idempotent", "method", "safe"]


def rows(text):
    for r in csv.DictReader(io.StringIO(text)):
        method, safe, idempotent = r["Method Name"].strip(), r["Safe"].strip(), r["Idempotent"].strip()
        if re.match(r"^[A-Za-z][A-Za-z-]*$", method) and safe and idempotent:
            yield {"idempotent": idempotent, "method": method, "safe": safe}


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = sorted(rows(source.fetch(a.source, a.cache, "http-methods.csv").decode("utf-8")), key=lambda r: r["method"])
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
