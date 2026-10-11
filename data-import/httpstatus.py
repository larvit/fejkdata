#!/usr/bin/env python3
"""Rebuild data/misc/httpstatus.tsv from the IANA HTTP Status Code Registry.

    data-import/httpstatus.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

What ships is the codes in use: a range, an unassigned code and a qualified reason are all skipped.
"""
import argparse
import csv
import io
from pathlib import Path

import source
import tsv

SOURCE = "https://www.iana.org/assignments/http-status-codes/http-status-codes-1.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "httpstatus.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["code", "reason"]


def rows(text):
    for r in csv.DictReader(io.StringIO(text)):
        code, reason = r["Value"].strip(), r["Description"].strip()
        if code.isdigit() and reason and "(" not in reason and reason != "Unassigned":
            yield {"code": code, "reason": reason}


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = rows(source.fetch(a.source, a.cache, "http-status-codes-1.csv").decode("utf-8"))
    tsv.write(a.out, COLUMNS, sorted(table, key=lambda r: int(r["code"])))


if __name__ == "__main__":
    main()
