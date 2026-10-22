#!/usr/bin/env python3
"""Rebuild data/nb_NO/first-name.tsv and last-name.tsv from Statistics Norway's name counts (CC BY 4.0, "Kilde: Statistisk sentralbyrå").

    data-import/names-no.py [--year YEAR] [--cache DIR] [--out DIR]

The first names are StatBank table 10501, the persons bearing each girls' or boys' name at the end of
the year, and the last names table 12891. Each table holds the names that 200 or more persons bear. A
name that girls and boys both bear has two rows. A label that is not a name in letters, as
`tsv.NAME` says, is dropped.
"""
import argparse
import json
import sys
from pathlib import Path

import source
import tsv

API = "https://data.ssb.no/api/v0/en/table/"
FIRST = "10501"
LAST = "12891"
OUT = Path(__file__).resolve().parent.parent / "data" / "nb_NO"
CACHE = Path(__file__).resolve().parent / "cache"
SEX = {"1": "f", "2": "m"}


def query(dimension, year):
    """A StatBank query for every category of a dimension in one year."""
    return {"query": [{"code": "Tid", "selection": {"filter": "item", "values": [year]}}, {"code": dimension, "selection": {"filter": "all", "values": ["*"]}}], "response": {"format": "json-stat2"}}


def fetched(table, year, dimension, cache):
    """A StatBank table's cube for the year, downloaded once."""
    body = source.fetch(API + table, cache, f"ssb-{table}-{year}.json", magic=b"{", data=json.dumps(query(dimension, year)).encode(), headers={"Content-Type": "application/json"})
    return json.loads(body)


def counted(cube, dimension):
    """The (category code, label, persons) of a cube of one dimension by one contents and one period, where the cube gives a count and the label is a name."""
    if cube["id"] != [dimension, "ContentsCode", "Tid"] or cube["size"][1:] != [1, 1]:
        sys.exit(f"{dimension}: the table is not one count per name, its dimensions are {cube['id']} of sizes {cube['size']}")
    categories = cube["dimension"][dimension]["category"]
    codes = categories["index"] if isinstance(categories["index"], list) else sorted(categories["index"], key=categories["index"].get)
    if len(codes) != len(cube["value"]):
        sys.exit(f"{dimension}: {len(codes)} names but {len(cube['value'])} counts")
    return [(code, categories["label"][code], persons) for code, persons in zip(codes, cube["value"]) if persons is not None and tsv.NAME.match(categories["label"][code])]


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--out", default=str(OUT))
    p.add_argument("--year", default="2025")
    a = p.parse_args()
    first = [{"name": name, "sex": SEX[code[0]], "count": persons} for code, name, persons in counted(fetched(FIRST, a.year, "Fornavn", a.cache), "Fornavn")]
    last = [{"name": name, "count": persons} for _, name, persons in counted(fetched(LAST, a.year, "Etternavn", a.cache), "Etternavn")]
    out = Path(a.out)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], sorted(first, key=lambda r: (r["name"], r["sex"])))
    tsv.write(out / "last-name.tsv", ["name", "count"], sorted(last, key=lambda r: r["name"]))


if __name__ == "__main__":
    main()
