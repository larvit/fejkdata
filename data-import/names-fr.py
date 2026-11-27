"""Rebuild data/fr_FR/first-name.tsv and last-name.tsv from INSEE's name files (Licence Ouverte 2.0, "Source : Insee") and Wikidata's family names (CC0).

    data-import/names-fr.py [--first-min N] [--last-min N] [--cache DIR] [--out DIR]

- The first names are the "Fichier des prénoms": for each year, the children of each sex who got the name as
  their first one. A name's count is the sum over the births of 1944 to 2023, the last 80 years, so a name gone
  out of use weighs less than in the whole file. A name that girls and boys both bear has a row under each sex.
  INSEE publishes a name only if 20 children got it in 1900 to 1945 or in 1946 to 2023, and only for a year in
  which 3 did: the rest sit in `_PRENOMS_RARES` and in rows without a year, which are dropped, so a count runs
  slightly low.
- The last names are the "Fichier des noms": the births of 1891 to 2000, each name's counts over its decades
  added up. INSEE writes them without accents, so a count holds every spelling of a name. A name takes
  one accent where more than half of Wikidata's French bearers write it so, and at least 5 do: `Lefèvre` for `Lefevre`.
- Both drop a name that `tsv.NAME` says is not a name in letters, and a first name borne by fewer than --first-min people or a last name borne by fewer than --last-min.
- INSEE writes names in capitals. Each part of a name, split at a hyphen, a space or an apostrophe, is
  capitalised, but in a last name `da`, `de` and `dos` stay in lower case before another part.
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
import wikidata_names

FIRST = "https://www.insee.fr/fr/statistiques/fichier/8205621/prenoms-2023-nat_csv.zip"
LAST = "https://www.insee.fr/fr/statistiques/fichier/3536630/noms2008nat_txt.zip"
YEARS = range(1944, 2024)
OUT = Path(__file__).resolve().parent.parent / "data" / "fr_FR"
CACHE = Path(__file__).resolve().parent / "cache"
SEX = {"1": "m", "2": "f"}
FRANCE = "Q142"
PARTICLES = {"da", "de", "dos"}
ACCENTS = {"e": "éèêë", "a": "àâä", "u": "ùûü", "o": "ôö", "i": "îï", "c": "ç", "E": "ÉÈÊË"}
RARE = "_PRENOMS_RARES"
OTHERS = "AUTRES NOMS"


def unzipped(url: str, cache: str, name: str) -> bytes:
	"""The one file of a zip downloaded from INSEE."""
	with zipfile.ZipFile(io.BytesIO(source.fetch(url, cache, name, magic=b"PK"))) as z:
		(member,) = z.namelist()
		return z.read(member)


def capitalised(name: str) -> str:
	"""INSEE's capitals as a name is written: each part starts with a capital."""
	return "".join(p.capitalize() if re.match(r"[^\W\d_]", p) else p for p in re.split(r"([ '-])", tsv.straight(name).lower()))


def surname(name: str) -> str:
	"""A last name as it is written: a particle stays in lower case before another part."""
	parts = capitalised(name).split(" ")
	return " ".join(p.lower() if i < len(parts) - 1 and p.lower() in PARTICLES else p for i, p in enumerate(parts))


def firsts(cache: str, floor: int) -> list[dict[str, object]]:
	"""The rows (name, sex, count) of the first names borne by at least floor people under that sex."""
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
		out += [{"name": n, "sex": sex, "count": c} for n, c in tsv.common(((n, c) for (n, s), c in counts.items() if s == sex), floor)]
	return sorted(out, key=lambda r: (r["name"], r["sex"]))


def accented(name: str) -> list[str]:
	"""The forms of a name with one letter given an accent, `Lefevre` as `Lefèvre`."""
	return [name[:i] + accent + name[i + 1 :] for i, letter in enumerate(name) for accent in ACCENTS.get(letter, "")]


def lasts(cache: str, floor: int, bearers: wikidata_names.Bearers) -> list[dict[str, object]]:
	"""The rows (name, count) of the last names borne by at least floor people.
	A name is written as most of Wikidata's French bearers write it, where that is a majority."""
	lines = unzipped(LAST, cache, "insee-noms2008nat.zip").decode("ascii").splitlines()
	if lines[0].split("\t")[:2] != ["NOM", "_1891_1900"]:
		sys.exit(f"the header is {lines[0]!r}, not NOM then the decade 1891-1900")
	counts = {}
	for line in lines[1:]:
		name, *decades = line.split("\t")
		if name != OTHERS and tsv.NAME.match(name):
			counts[name] = sum(int(d) for d in decades)
	named = [(surname(n), c) for n, c in tsv.common(counts.items(), floor)]
	spelling = wikidata_names.written({n: accented(n) for n, _ in named}, "fr", FRANCE, cache, bearers)
	return [{"name": n, "count": c} for n, c in sorted((spelling.get(n, n), c) for n, c in named)]


FIRST_MIN = 5
LAST_MIN = 200


def main() -> None:
	p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
	p.add_argument("--cache", default=str(CACHE))
	p.add_argument("--first-min", type=int, default=FIRST_MIN, help="bearers a first name needs")
	p.add_argument("--last-min", type=int, default=LAST_MIN, help="bearers a last name needs")
	p.add_argument("--out", default=str(OUT))
	a = p.parse_args()
	out = Path(a.out)
	out.mkdir(parents=True, exist_ok=True)
	tsv.write(out / "first-name.tsv", ["name", "sex", "count"], firsts(a.cache, a.first_min))
	tsv.write(out / "last-name.tsv", ["name", "count"], lasts(a.cache, a.last_min, wikidata_names.bearers))


if __name__ == "__main__":
	main()
