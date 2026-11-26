"""Rebuild data/sv_SE/first-name.tsv and last-name.tsv from SCB's whole-population name counts (CC0, "Källa: SCB").

    data-import/names-se.py [--source URL_OR_FILE] [--first-min N] [--last-min N] [--cache DIR] [--out DIR]

A tilltalsnamn goes under each sex that carries it, so a name both sexes carry has two rows.
A first name borne by fewer than --first-min people goes, and a last name borne by fewer than --last-min.
"""
import argparse
from collections.abc import Iterable
from pathlib import Path

import source
import tsv
import xlsx

SOURCE = "https://www.scb.se/contentassets/9fe7dbb460994c72b835163dbc491ef9/namn-med-minst-tva-barare-31-december-2022.xlsx"
OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE"
CACHE = Path(__file__).resolve().parent / "cache"
SHEETS = {"f": "Tilltalsnamn kvinnor", "m": "Tilltalsnamn män"}
SURNAMES = "Efternamn"
PARTICLES = {"af", "av", "da", "de", "del", "den", "der", "di", "dos", "du", "la", "le", "van", "von"}


def cased(name: str) -> str:
    """SCB's uppercase name as it is written: each part capitalised, a particle before another part lowercased."""
    parts = tsv.straight(name).lower().split(" ")
    return " ".join(w if w in PARTICLES and len(parts) > 1 else w.title() for w in parts)


def counted(rows: Iterable[list[str]], floor: int) -> list[tuple[str, int]]:
    """The names of a sheet borne by at least floor people, the commonest first, initials and single letters dropped."""
    names = [(tsv.straight(cells[0]), int(cells[1])) for cells in rows if len(cells) >= 2 and cells[1].isdigit() and tsv.NAME.match(cells[0])]
    return tsv.common(names, floor)


FIRST_MIN = 5
LAST_MIN = 20


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--first-min", type=int, default=FIRST_MIN, help="bearers a first name needs")
    p.add_argument("--last-min", type=int, default=LAST_MIN, help="bearers a last name needs")
    p.add_argument("--out", default=str(OUT))
    p.add_argument("--source", default=SOURCE)
    a = p.parse_args()
    data = source.fetch(a.source, a.cache, "scb-namn-2022.xlsx", magic=b"PK")
    first = [{"name": cased(n), "sex": sex, "count": c} for sex, sheet in SHEETS.items() for n, c in counted(xlsx.rows(data, sheet), a.first_min)]
    last = [{"name": cased(n), "count": c} for n, c in counted(xlsx.rows(data, SURNAMES), a.last_min)]
    out = Path(a.out)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], sorted(first, key=lambda r: (r["name"], r["sex"])))
    tsv.write(out / "last-name.tsv", ["name", "count"], sorted(last, key=lambda r: r["name"]))


if __name__ == "__main__":
    main()
