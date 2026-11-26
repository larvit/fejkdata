"""Rebuild data/fi_FI/first-name.tsv and last-name.tsv from DVV's name statistics (CC BY 4.0, "Lähde: Digi- ja väestötietovirasto").

    data-import/names-fi.py [--etunimet URL_OR_FILE] [--sukunimet URL_OR_FILE] [--first-min N] [--last-min N] [--cache DIR] [--out DIR]

The first names are the sheets `Miehet ens` and `Naiset ens` of the "Etunimitilasto": the first given names
of the Finnish citizens in the population register. The last names are the sheet `Nimet` of the "Sukunimitilasto". A name
that women and men both bear has two rows. A label that is not a name in letters, as `tsv.NAME` says, and a first name borne by fewer than --first-min people or a last name borne by fewer than --last-min are dropped.
"""
import argparse
from collections.abc import Iterable
from pathlib import Path

import source
import tsv
import xlsx

FIRST = "https://avoindata.suomi.fi/data/dataset/57282ad6-3ab1-48fb-983a-8aba5ff8d29a/resource/08c89936-a230-42e9-a9fc-288632e234f5/download/etunimitilasto-2026-08-04-dvv.xlsx"
LAST = "https://avoindata.suomi.fi/data/dataset/57282ad6-3ab1-48fb-983a-8aba5ff8d29a/resource/26c2d0fc-c27d-4c8a-b646-95155f9b6e59/download/sukunimitilasto-2026-08-04-dvv.xlsx"
OUT = Path(__file__).resolve().parent.parent / "data" / "fi_FI"
CACHE = Path(__file__).resolve().parent / "cache"
SHEETS = {"f": "Naiset ens", "m": "Miehet ens"}
LAST_SHEET = "Nimet"


def counted(rows: Iterable[list[str]], floor: int) -> list[tuple[str, int]]:
	"""The names of a sheet borne by at least floor people, the commonest first, whose header row and any label that is no name are dropped."""
	names = [(tsv.straight(cells[0]), int(cells[1])) for cells in rows if len(cells) >= 2 and cells[1].isdigit() and tsv.NAME.match(cells[0])]
	return tsv.common(names, floor)


FIRST_MIN = 5
LAST_MIN = 5


def main() -> None:
	p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
	p.add_argument("--cache", default=str(CACHE))
	p.add_argument("--etunimet", default=FIRST)
	p.add_argument("--sukunimet", default=LAST)
	p.add_argument("--first-min", type=int, default=FIRST_MIN, help="bearers a first name needs")
	p.add_argument("--last-min", type=int, default=LAST_MIN, help="bearers a last name needs")
	p.add_argument("--out", default=str(OUT))
	a = p.parse_args()
	etunimet = source.fetch(a.etunimet, a.cache, "dvv-etunimitilasto-2026-08-04.xlsx", magic=b"PK")
	sukunimet = source.fetch(a.sukunimet, a.cache, "dvv-sukunimitilasto-2026-08-04.xlsx", magic=b"PK")
	first = [{"name": n, "sex": sex, "count": c} for sex, sheet in SHEETS.items() for n, c in counted(xlsx.rows(etunimet, sheet), a.first_min)]
	last = [{"name": n, "count": c} for n, c in counted(xlsx.rows(sukunimet, LAST_SHEET), a.last_min)]
	out = Path(a.out)
	out.mkdir(parents=True, exist_ok=True)
	tsv.write(out / "first-name.tsv", ["name", "sex", "count"], sorted(first, key=lambda r: (r["name"], r["sex"])))
	tsv.write(out / "last-name.tsv", ["name", "count"], sorted(last, key=lambda r: r["name"]))


if __name__ == "__main__":
	main()
