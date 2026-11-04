#!/usr/bin/env python3
"""Rebuild data/sv_SE/bankaccount.tsv from Bankinfrastruktur i Sverige AB's list of clearing numbers.

    data-import/bankaccount-se.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

One row per clearing number of each listed range, with the template that draws an account number
its bank's check digit accepts, per "Bankernas kontonummer" (Bankinfrastruktur i Sverige AB,
2024-02-22): type 1 comment 1 weighs the clearing number without its first digit and seven digits
under the 11-modul, comment 2 the whole clearing number and seven digits, type 2 comment 1 is ten
digits under the 10-modul (Luhn), comment 2 nine digits under the 11-modul, and comment 3 ten
digits under the 10-modul. Comment 4 is not published, so its ranges are left out.
"""
import argparse
import csv
import io
import re
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://www.bankinfrastruktur.se/media/sqfjsvkp/clearingnummertabell-for-nedladdning.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE" / "bankaccount.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["clearing", "bank", "number"]
RANGE = re.compile(r"^(\d{4}) - (\d{4})$")
UNPUBLISHED = 4
EXPECTED_RANGES = 58
EXPECTED_ROWS = 7730


def number(clearing, typ, comment):
    if comment == 1 and typ == 1:
        return "{digits(5)}{bgmod11(" + clearing[1:] + ")}"
    if comment == 2 and typ == 1:
        return "{digits(5)}{bgmod11(" + clearing + ")}"
    if comment == 2 and typ == 2:
        return "{digits(7)}{bgmod11(0)}"
    if comment in (1, 3) and typ == 2:
        return "{digits(9)}{luhn()}"
    sys.exit(f"{clearing}: type {typ}, comment {comment} has no rule here")


def rows(text):
    ranges = 0
    for r in csv.reader(io.StringIO(text), delimiter=";"):
        m = RANGE.match(r[0].strip()) if r else None
        if not m:
            continue
        ranges += 1
        typ, comment = (int(x) for x in r[4].split(":"))
        if comment == UNPUBLISHED:
            continue
        for c in range(int(m[1]), int(m[2]) + 1):
            clearing = f"{c:04d}"
            yield {"clearing": clearing, "bank": r[1].strip(), "number": number(clearing, typ, comment)}
    if ranges != EXPECTED_RANGES:
        sys.exit(f"{ranges} ranges, not {EXPECTED_RANGES}; the list has changed")


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    text = source.fetch(a.source, a.cache, "bankaccount-clearing.csv").decode("utf-8-sig")
    table = list(rows(text))
    if len(table) != EXPECTED_ROWS:
        sys.exit(f"{len(table)} clearing numbers, not {EXPECTED_ROWS}; the list has changed")
    tsv.write(a.out, COLUMNS, table, templates=("number",))


if __name__ == "__main__":
    main()
