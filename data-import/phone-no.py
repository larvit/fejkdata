#!/usr/bin/env python3
"""Rebuild data/nb_NO/phone-mobile.tsv and phone-fixed.tsv from Nkom's E.164 register.

    data-import/phone-no.py [--source URL_OR_FILE] [--cache DIR] [--out-mobile FILE] [--out-fixed FILE]

A row is the first four digits of an 8-digit number, written in two pairs, when Nkom has assigned
every one of the ten thousand numbers behind them to a provider ("Tildelt") as mobile
("Mobilnummer") or fixed ("Fastnettnummer"). A prefix only part assigned is dropped, so no
number drawn from the tables lies in a block Nkom has left free or blocked. Every block of the
register is a whole number of thousands.
"""
import argparse
import csv
import io
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://stenonicprdnoea01.blob.core.windows.net/enonicpubliccontainer/numsys/nkom.no/E164.csv"
DATA = Path(__file__).resolve().parent.parent / "data" / "nb_NO"
CACHE = Path(__file__).resolve().parent / "cache"
KINDS = {"Mobilnummer": "phone-mobile.tsv", "Fastnettnummer": "phone-fixed.tsv"}


def assigned(register):
    """The four-digit prefixes whose ten thousand numbers are all assigned, per kind of number."""
    thousands = {kind: set() for kind in KINDS}
    for row in csv.DictReader(io.StringIO(register.decode("utf-8-sig"))):
        if row["Status"] != "Tildelt" or row["Kategori"] not in KINDS:
            continue
        first, last = (int(row[k].replace(" ", "")) for k in ("Fra", "Til"))
        if len(row["Fra"].replace(" ", "")) != 8 or first % 1000 or (last + 1) % 1000:
            sys.exit(f"{row['Fra']} to {row['Til']} is not a whole number of thousands of 8-digit numbers")
        thousands[row["Kategori"]].update(range(first // 1000, last // 1000 + 1))
    return {
        kind: sorted(p for p in {t // 10 for t in held} if all(p * 10 + i in held for i in range(10)))
        for kind, held in thousands.items()
    }


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out-mobile", default=str(DATA / KINDS["Mobilnummer"]))
    p.add_argument("--out-fixed", default=str(DATA / KINDS["Fastnettnummer"]))
    a = p.parse_args()
    prefixes = assigned(source.fetch(a.source, a.cache, "E164.csv", magic=b"\xef\xbb\xbf\"Fra\""))
    for kind, out in (("Mobilnummer", a.out_mobile), ("Fastnettnummer", a.out_fixed)):
        tsv.write(out, ["prefix"], [{"prefix": f"{p // 100:02d} {p % 100:02d}"} for p in prefixes[kind]])


if __name__ == "__main__":
    main()
