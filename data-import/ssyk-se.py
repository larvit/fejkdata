#!/usr/bin/env python3
"""Rebuild data/sv_SE/ssyk.tsv from SCB's SSYK 2012 (Källa: SCB).

    data-import/ssyk-se.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

One row per four-digit occupation group, sorted by code, with its Swedish label. SCB's sheet pads each
label with spaces, which the script strips, and ends in empty rows, which it skips. The sheet stores the
three military groups as numbers, `110`, `210` and `310`, and the script restores their leading zero.
"""
import argparse
import re
import sys
from pathlib import Path

import source
import tsv
import xlsx

SOURCE = "https://www.scb.se/contentassets/0c0089cc085a45d49c1dc83923ad933a/ssyk-2012-koder.xlsx"
SHEET = "4-siffer"
OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE" / "ssyk.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
CODE = re.compile(r"^\d{4}$")
HEADER = ["SSYK 2012 kod", "Benämning"]
EXPECTED = 429


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    sheet = list(xlsx.rows(source.fetch(a.source, a.cache, "ssyk-2012.xlsx", magic=b"PK"), SHEET))
    if not sheet or sheet[0] != HEADER:
        sys.exit(f"{SHEET} does not open with the columns {HEADER}; SCB has changed the sheet")
    rows = [{"code": r[0].strip().zfill(4), "name": " ".join(r[1].split())} for r in sheet[1:] if len(r) == 2]
    bad = [r for r in rows if not CODE.match(r["code"])]
    if bad:
        sys.exit(f"codes that are not four digits: {bad}")
    if len(rows) != EXPECTED:
        sys.exit(f"the sheet holds {len(rows)} occupation groups, not {EXPECTED}")
    tsv.write(a.out, ["code", "name"], sorted(rows, key=lambda r: r["code"]))


if __name__ == "__main__":
    main()
