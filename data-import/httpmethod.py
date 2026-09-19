#!/usr/bin/env python3
"""Rebuild data/misc/httpmethod.tsv from the IANA HTTP Method Registry.

    data-import/httpmethod.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

A row needs a name of letters and hyphens, which drops the registry's `*`, and both of
the answers the registry gives for it.
"""
import argparse
import csv
import io
import re
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://www.iana.org/assignments/http-methods/methods.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "httpmethod.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["idempotent", "method", "safe"]
ANSWERS = ("yes", "no")


def rows(text):
    for r in csv.DictReader(io.StringIO(text)):
        method = (r["Method Name"] or "").strip()
        safe, idempotent = (r["Safe"] or "").strip(), (r["Idempotent"] or "").strip()
        if not re.fullmatch(r"[A-Za-z][A-Za-z-]*", method) or not safe or not idempotent:
            continue
        for column, answer in (("Safe", safe), ("Idempotent", idempotent)):
            if answer not in ANSWERS:
                sys.exit(f"{method}: {column} {answer!r} is neither yes nor no")
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
