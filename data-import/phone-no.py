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
MOBILE, FIXED = "Mobilnummer", "Fastnettnummer"
KINDS = {MOBILE: "phone-mobile.tsv", FIXED: "phone-fixed.tsv"}


def assigned(register):
    """The four-digit prefixes whose ten thousand numbers are all assigned, per kind of number."""
    thousands = {kind: set() for kind in KINDS}
    for row in csv.DictReader(io.StringIO(register.decode("utf-8-sig"))):
        if row["Status"] != "Tildelt" or row["Kategori"] not in KINDS:
            continue
        ends = [row[k].replace(" ", "") for k in ("Fra", "Til")]
        if not all(len(e) == 8 and e.isdigit() for e in ends):
            sys.exit(f"{row['Fra']} to {row['Til']} is not a range of 8-digit numbers")
        first, last = map(int, ends)
        if first % 1000 or (last + 1) % 1000:
            sys.exit(f"{row['Fra']} to {row['Til']} is not a whole number of thousands")
        thousands[row["Kategori"]].update(range(first // 1000, last // 1000 + 1))
    return {
        kind: sorted(prefix for prefix in {t // 10 for t in held} if all(prefix * 10 + i in held for i in range(10)))
        for kind, held in thousands.items()
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--cache", default=str(CACHE))
    parser.add_argument("--source", default=SOURCE)
    parser.add_argument("--out-mobile", default=str(DATA / KINDS[MOBILE]))
    parser.add_argument("--out-fixed", default=str(DATA / KINDS[FIXED]))
    args = parser.parse_args()
    prefixes = assigned(source.fetch(args.source, args.cache, "E164.csv", magic=b"\xef\xbb\xbf\"Fra\""))
    for kind, out in ((MOBILE, args.out_mobile), (FIXED, args.out_fixed)):
        tsv.write(out, ["prefix"], [{"prefix": f"{n // 100:02d} {n % 100:02d}"} for n in prefixes[kind]])


if __name__ == "__main__":
    main()
