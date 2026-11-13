#!/usr/bin/env python3
"""Rebuild data/sv_SE/bankaccount.tsv from Bankinfrastruktur i Sverige AB's list of clearing numbers.

    data-import/bankaccount-se.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

One row per clearing number of each listed range, with the template that draws an account number
passing the check Bankgirot's "Bankernas kontonummer" (2024-02-22) gives the range. The list's
"Konto-typ" column holds the account type and a comment number:

- type 1, comment 1: seven digits, 11-modul over the clearing number without its first digit and them
- type 1, comment 2: seven digits, 11-modul over the whole clearing number and them
- type 2, comment 1 or 3: ten digits, 10-modul (Luhn)
- type 2, comment 2: nine digits, 11-modul
- comment 4: not published, so its ranges are left out
"""
import argparse
import csv
import io
import re
import sys
from collections.abc import Iterator
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


def number(clearing: str, typ: int, comment: int) -> str:
    if comment == 1 and typ == 1:
        return "{digits(5)}{bgmod11(" + clearing[1:] + ")}"
    if comment == 2 and typ == 1:
        return "{digits(5)}{bgmod11(" + clearing + ")}"
    if comment == 2 and typ == 2:
        return "{digits(7)}{bgmod11(0)}"
    if comment in (1, 3) and typ == 2:
        return "{digits(9)}{luhn()}"
    sys.exit(f"{clearing}: type {typ}, comment {comment} has no rule here")


def rows(text: str) -> Iterator[dict[str, str]]:
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


def main() -> None:
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
