#!/usr/bin/env python3
"""Rebuild data/sv_SE/sni.tsv from SCB's SNI 2025 (Källa: SCB).

    data-import/sni-se.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

One row per detaljgrupp, the five-digit level, with its code spelled `01.110` and its Swedish label.
SCB's sheet pads some codes and labels with spaces, which the script strips.
"""
import argparse
import re
import sys
from pathlib import Path

import source
import tsv
import xlsx

SOURCE = "https://www.scb.se/globalassets/sni-2025.xlsx"
SHEET = "Detaljgrupp (Femsiffer)"
OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE" / "sni.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
CODE = re.compile(r"^\d\d\.\d{3}$")
HEADER = "SNI2025Detaljgrupp (Officiell kodstruktur)"
EXPECTED = 835


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    sheet = list(xlsx.rows(source.fetch(a.source, a.cache, "sni-2025.xlsx", magic=b"PK"), SHEET))
    if not sheet or sheet[0][:1] != [HEADER]:
        sys.exit(f"{SHEET} does not open with the column {HEADER!r}; SCB has changed the sheet")
    rows = [{"code": r[0].strip(), "name": " ".join(r[2].split())} for r in sheet[1:]]
    bad = [r for r in rows if not CODE.match(r["code"])]
    if bad:
        sys.exit(f"codes that are not NN.NNN: {bad}")
    if len(rows) != EXPECTED:
        sys.exit(f"the sheet holds {len(rows)} detaljgrupper, not {EXPECTED}")
    tsv.write(a.out, ["code", "name"], rows)


if __name__ == "__main__":
    main()
