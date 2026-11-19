"""Rebuild data/fr_FR/first-name.tsv and last-name.tsv from INSEE's name files (Licence Ouverte 2.0, "Source : Insee").

    data-import/names-fr.py [--first N] [--last N] [--cache DIR] [--out DIR]

- The first names are the "Fichier des prénoms": for each year, the children of each sex who got the name as
  their first one. A name's count is the sum over the births of 1944 to 2023, the 80 years whose people can
  still be alive, so a name that has gone out of use weighs less than in the whole file. A name that girls and
  boys both bear has a row under each sex. INSEE publishes a name only if 20 children got it in 1900 to 1945
  or in 1946 to 2023, and only for a year in which 3 did: the rest sit in `_PRENOMS_RARES` and in rows without
  a year, which are dropped. The top N names of each sex by count stay (2000).
- The last names are the "Fichier des noms": the births of 1891 to 2000, each name's counts over its decades
  added up. INSEE writes them in capitals without accents, so `Lefevre` is not `Lefèvre`. The top N names by
  count stay (5000), as does a name only if `tsv.NAME` says it is a name in letters, which drops `N'Diaye`.
- INSEE writes names in capitals. A first name's parts, split at a hyphen, a space or an apostrophe, are each
  capitalised. A last name's too, but `de`, `des`, `du`, `da` and `van` stay in lower case before another part, as
  does a `le` or `la` after one of them.
"""
import argparse
import csv
import io
import re
import sys
import zipfile
from pathlib import Path

import source
import tsv

FIRST = "https://www.insee.fr/fr/statistiques/fichier/8205621/prenoms-2023-nat_csv.zip"
LAST = "https://www.insee.fr/fr/statistiques/fichier/3536630/noms2008nat_txt.zip"
YEARS = range(1944, 2024)
OUT = Path(__file__).resolve().parent.parent / "data" / "fr_FR"
CACHE = Path(__file__).resolve().parent / "cache"
SEX = {"1": "m", "2": "f"}
PARTICLES = {"de", "des", "du", "da", "van"}
ARTICLES = {"le", "la"}
RARE = "_PRENOMS_RARES"
OTHERS = "AUTRES NOMS"


def unzipped(url: str, cache: str, name: str) -> bytes:
    """The one file of a zip downloaded from INSEE."""
    with zipfile.ZipFile(io.BytesIO(source.fetch(url, cache, name, magic=b"PK"))) as z:
        (member,) = z.namelist()
        return z.read(member)


def capitalised(name: str) -> str:
    """INSEE's capitals as a name is written: each part starts with a capital."""
    return "".join(p.capitalize() if re.match(r"[^\W\d_]", p) else p for p in re.split(r"([ '-])", name.lower()))


def surname(name: str) -> str:
    """A last name as it is written: a particle, and an article right after one, in lower case before another part."""
    parts = capitalised(name).split(" ")
    for i, part in enumerate(parts[:-1]):
        if part.lower() in PARTICLES or part.lower() in ARTICLES and i > 0 and parts[i - 1].lower() in PARTICLES:
            parts[i] = part.lower()
    return " ".join(parts)


def firsts(cache: str, top: int) -> list[dict[str, object]]:
    """The rows (name, sex, count) of the top names of each sex."""
    text = unzipped(FIRST, cache, "insee-prenoms-2023-nat.zip").decode("utf-8")
    rows = list(csv.reader(io.StringIO(text), delimiter=";"))
    if rows[0] != ["sexe", "preusuel", "annais", "nombre"]:
        sys.exit(f"the columns are {rows[0]}, not sexe, preusuel, annais and nombre")
    counts: dict[tuple[str, str], int] = {}
    for sex, name, year, count in rows[1:]:
        if sex not in SEX:
            sys.exit(f"{name!r} has the sex {sex!r}, not 1 or 2")
        if year.isdigit() and int(year) in YEARS and name != RARE and tsv.NAME.match(name):
            counts[(capitalised(name), SEX[sex])] = counts.get((capitalised(name), SEX[sex]), 0) + int(count)
    out = []
    for sex in SEX.values():
        ranked = sorted(((n, c) for (n, s), c in counts.items() if s == sex), key=lambda r: (-r[1], r[0]))[:top]
        out += [{"name": n, "sex": sex, "count": c} for n, c in ranked]
    return sorted(out, key=lambda r: (r["name"], r["sex"]))


def lasts(cache: str, top: int) -> list[dict[str, object]]:
    """The rows (name, count) of the top last names."""
    lines = unzipped(LAST, cache, "insee-noms2008nat.zip").decode("ascii").splitlines()
    if lines[0].split("\t")[:2] != ["NOM", "_1891_1900"]:
        sys.exit(f"the header is {lines[0]!r}, not NOM then the decade 1891-1900")
    counts = {}
    for line in lines[1:]:
        name, *decades = line.split("\t")
        if name != OTHERS and tsv.NAME.match(name):
            counts[name] = sum(int(d) for d in decades)
    ranked = sorted(counts.items(), key=lambda r: (-r[1], r[0]))[:top]
    spelled = [(surname(n), c) for n, c in ranked]
    dupes = sorted({n for n, _ in spelled if [m for m, _ in spelled].count(n) > 1})
    if dupes:
        sys.exit(f"last names: spelled alike: {', '.join(dupes)}")
    return [{"name": n, "count": c} for n, c in sorted(spelled)]


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--first", type=int, default=2000, help="names per sex")
    p.add_argument("--last", type=int, default=5000)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    out = Path(a.out)
    out.mkdir(parents=True, exist_ok=True)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], firsts(a.cache, a.first))
    tsv.write(out / "last-name.tsv", ["name", "count"], lasts(a.cache, a.last))


if __name__ == "__main__":
    main()
