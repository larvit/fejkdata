"""Rebuild data/en_US/first-name.tsv and last-name.tsv from SSA baby names and the Census 2010 surnames (public domain).

    data-import/names-us.py [--names URL_OR_FILE] [--surnames URL_OR_FILE] [--from-year YEAR] [--min-count N] [--cache DIR] [--out DIR]

A given name is counted over the births of --from-year and later, under each sex it was given to, so a
name both sexes carry has two rows. --names takes SSA's names.zip or a merged name,sex,count,year file;
--surnames takes the Census names.zip or its Names_2010Census.csv.
A name borne by fewer than --min-count people goes.
"""
import argparse
import collections
import csv
import io
import re
import zipfile
from collections.abc import Iterable, Iterator
from pathlib import Path
from typing import Any

import source
import tsv

NAMES = "https://raw.githubusercontent.com/hackerb9/ssa-baby-names/master/alldata.txt"
SURNAMES = "https://www2.census.gov/topics/genealogy/2010surnames/names.zip"
OUT = Path(__file__).resolve().parent.parent / "data" / "en_US"
CACHE = Path(__file__).resolve().parent / "cache"
MC = re.compile(r"^Mc([a-z])")
PLACEHOLDERS = {"Baby", "Female", "Infant", "Male", "Notnamed", "Unknown", "Unnamed"}


def csv_or_zip_rows(data: bytes, member: str) -> Iterator[list[str]]:
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


def given(data: bytes, from_year: int) -> collections.Counter[tuple[str, str]]:
    counts: collections.Counter[tuple[str, str]] = collections.Counter()
    for r in csv_or_zip_rows(data, r"yob\d{4}\.txt"):
        if len(r) >= 4 and r[3].isdigit() and int(r[3]) >= from_year and r[0] not in PLACEHOLDERS:
            counts[(r[0], r[1].lower())] += int(r[2])
    return counts


def surname(name: str) -> str:
    """A Census uppercase surname as it is written: capitalised, and Mc before a capital."""
    return MC.sub(lambda m: "Mc" + m.group(1).upper(), name.title())


def firsts(counts: collections.Counter[tuple[str, str]], floor: int) -> list[dict[str, Any]]:
    """The rows (name, sex, count) of the given names borne by at least floor people under that sex."""
    rows = [{"name": n, "sex": sex, "count": c} for sex in ("f", "m") for n, c in tsv.common(((n, c) for (n, s), c in counts.items() if s == sex), floor)]
    return sorted(rows, key=lambda r: (r["name"], r["sex"]))


def lasts(rows: Iterable[list[str]], floor: int) -> list[dict[str, Any]]:
    """The rows (name, count) of the surnames borne by at least floor people, a suppressed count and a non-name dropped."""
    named = [(surname(r[0]), int(r[2])) for r in rows if len(r) >= 3 and r[2].isdigit() and r[0].isalpha()]
    return [{"name": n, "count": c} for n, c in sorted(tsv.common(named, floor))]


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--min-count", type=int, default=tsv.MIN_BEARERS, help="bearers a name needs")
    p.add_argument("--from-year", type=int, default=1930)
    p.add_argument("--names", default=NAMES)
    p.add_argument("--out", default=str(OUT))
    p.add_argument("--surnames", default=SURNAMES)
    a = p.parse_args()
    counts = given(source.fetch(a.names, a.cache, "ssa-names.txt"), a.from_year)
    rows = csv_or_zip_rows(source.fetch(a.surnames, a.cache, "census-surnames-2010.zip"), r"Names_2010Census\.csv")
    out = Path(a.out)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], firsts(counts, a.min_count))
    tsv.write(out / "last-name.tsv", ["name", "count"], lasts(rows, a.min_count))


if __name__ == "__main__":
    main()
