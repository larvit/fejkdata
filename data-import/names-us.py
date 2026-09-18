#!/usr/bin/env python3
"""Rebuild data/en_US/first-name.tsv and last-name.tsv from SSA baby names and the Census 2010 surnames (public domain).

    data-import/names-us.py [--names URL_OR_FILE] [--surnames URL_OR_FILE] [--from-year YEAR] [--first N] [--last N] [--cache DIR] [--out DIR]

A given name is counted over the births of --from-year and later, under each sex it was given to, so a
name both sexes carry has two rows. --names takes SSA's names.zip or a merged name,sex,count,year file;
--surnames takes the Census names.zip or its Names_2010Census.csv.
"""
import argparse
import collections
import csv
import io
import re
import zipfile
from pathlib import Path

import source
import tsv

NAMES = "https://raw.githubusercontent.com/hackerb9/ssa-baby-names/master/alldata.txt"
SURNAMES = "https://www2.census.gov/topics/genealogy/2010surnames/names.zip"
OUT = Path(__file__).resolve().parent.parent / "data" / "en_US"
CACHE = Path(__file__).resolve().parent / "cache"
MC = re.compile(r"^Mc([a-z])")
PLACEHOLDERS = {"Baby", "Female", "Infant", "Male", "Notnamed", "Unknown", "Unnamed"}


def csv_or_zip_rows(data, member):
    """The rows of a CSV, or of every member of a zip named like member, name,sex,count[,year]."""
    if data.startswith(b"PK"):
        z = zipfile.ZipFile(io.BytesIO(data))
        for n in sorted(z.namelist()):
            if re.fullmatch(member, n):
                year = re.sub(r"\D", "", n)
                for r in csv.reader(io.StringIO(z.read(n).decode("utf-8-sig"))):
                    yield r + [year] if year else r
        return
    yield from csv.reader(io.StringIO(data.decode("utf-8-sig")))


def given(data, from_year):
    counts = collections.Counter()
    for r in csv_or_zip_rows(data, r"yob\d{4}\.txt"):
        if len(r) >= 4 and r[3].isdigit() and int(r[3]) >= from_year and r[0] not in PLACEHOLDERS:
            counts[(r[0], r[1].lower())] += int(r[2])
    return counts


def surname(name):
    """A Census uppercase surname as it is written: capitalised, and Mc before a capital."""
    return MC.sub(lambda m: "Mc" + m.group(1).upper(), name.title())


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--first", type=int, default=2000, help="names per sex")
    p.add_argument("--from-year", type=int, default=1930)
    p.add_argument("--last", type=int, default=5000)
    p.add_argument("--names", default=NAMES)
    p.add_argument("--out", default=str(OUT))
    p.add_argument("--surnames", default=SURNAMES)
    a = p.parse_args()
    counts = given(source.fetch(a.names, a.cache, "ssa-names.txt"), a.from_year)
    first = []
    for sex in ("f", "m"):
        top = sorted(((n, c) for (n, s), c in counts.items() if s == sex), key=lambda n: (-n[1], n[0]))[:a.first]
        first += [{"name": n, "sex": sex, "count": c} for n, c in top]
    rows = csv_or_zip_rows(source.fetch(a.surnames, a.cache, "census-surnames-2010.zip"), r"Names_2010Census\.csv")
    last = [{"name": surname(r[0]), "count": int(r[2])} for r in rows if len(r) >= 3 and r[2].isdigit() and r[0].isalpha()]
    last = sorted(last, key=lambda r: (-r["count"], r["name"]))[:a.last]
    out = Path(a.out)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], sorted(first, key=lambda r: (r["name"], r["sex"])))
    tsv.write(out / "last-name.tsv", ["name", "count"], sorted(last, key=lambda r: r["name"]))


if __name__ == "__main__":
    main()
