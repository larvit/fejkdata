#!/usr/bin/env python3
"""Rebuild data/misc/territory.tsv from datasets/country-codes (PDDL).

    data-import/territory.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]
"""
import argparse
import csv
import io
import re
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://raw.githubusercontent.com/datasets/country-codes/main/data/country-codes.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "territory.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["alpha2", "alpha3", "calling-code", "capital", "country", "currency", "flag", "languages", "name", "numeric", "tld"]
SOVEREIGN = re.compile(r"^(?:Part of|Territor(?:y|ies) of|Crown dependency of|Commonwealth of|Associated with) ([A-Z]{2})$")
# Gaps in the source, keyed by alpha2.
FIXUPS = {"TR": {"currency": "TRY"}}


def flag(alpha2):
    return "".join(chr(0x1F1E6 + ord(c) - ord("A")) for c in alpha2)


def calling_code(dial):
    first = re.sub(r"[^0-9-]", "", dial.split(",")[0])
    return first if re.match(r"^\d+-\d{3}$", first) else first.split("-")[0]


def languages(field):
    return ",".join(p for p in field.split(",") if p)


def country(alpha2, independent):
    """The sovereign state the register records; a territory it records none for stands alone."""
    m = SOVEREIGN.match(independent)
    return m.group(1) if m else alpha2


def rows(text):
    for r in csv.DictReader(io.StringIO(text)):
        alpha2 = r["ISO3166-1-Alpha-2"]
        row = {
            "alpha2": alpha2,
            "alpha3": r["ISO3166-1-Alpha-3"],
            "calling-code": calling_code(r["Dial"]),
            "capital": r["Capital"],
            "country": country(alpha2, r["is_independent"]),
            "currency": r["ISO4217-currency_alphabetic_code"].split(",")[0],
            "flag": flag(alpha2),
            "languages": languages(r["Languages"]),
            "name": r["CLDR display name"] or r["official_name_en"],
            "numeric": r["ISO3166-1-numeric"].zfill(3),
            "tld": r["TLD"],
        }
        row.update(FIXUPS.get(alpha2, {}))
        if all(row[c] for c in ("alpha2", "alpha3", "calling-code", "capital", "currency", "name", "tld")):
            yield row


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = sorted(rows(source.fetch(a.source, a.cache, "country-codes.csv").decode("utf-8")), key=lambda r: r["alpha2"])
    shipped = {r["alpha2"] for r in table}
    for r in table:
        if r["country"] not in shipped:
            sys.exit(f"{r['alpha2']}: country {r['country']} is no shipped territory; a sovereign the set drops leaves its territories unreadable")
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
