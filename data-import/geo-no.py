#!/usr/bin/env python3
"""Rebuild data/geo/NO/*.tsv and address.json from Kartverket's address register (CC BY 4.0) and SSB (CC BY 4.0).

    data-import/geo-no.py [--addresses URL_OR_FILE] [--cache DIR] [--date YYYY-MM-DD] [--min-streets N] [--street-share F] [--out DIR]

Every street address (vegadresse) in the register names its street, number, postnummer, poststed and kommune.
A poststed is a locality, with each postnummer and street its addresses spell, each weighted by its addresses. A
poststed's kommune is the one holding most of its addresses. Its population is each kommune's, from SSB, shared
out over that kommune's poststeder by their addresses there: a poststed has no population of its own, and an
address in Oslo houses more people than one in a village. The register spells a poststed in capitals, and
`recased` writes it as a place name. address.json draws a house number as often as the register's addresses
carry one of its length, with a letter or without.
"""
import collections
import csv
import datetime
import io
import json
import re
import sys
import zipfile
from dataclasses import dataclass, field

import geo
import source
import tsv

ADDRESSES = "https://nedlasting.geonorge.no/geonorge/Basisdata/MatrikkelenAdresse/CSV/Basisdata_0000_Norge_4258_MatrikkelenAdresse_CSV.zip"
KLASS = "https://data.ssb.no/api/klass/v1/classifications/{}/codesAt?date={}"
MUNICIPALITIES, REGIONS = 131, 104
POPULATION = "https://data.ssb.no/api/v0/no/table/07459"
POPULATION_QUERY = {
    "query": [
        {"code": "Region", "selection": {"filter": "all", "values": ["*"]}},
        {"code": "ContentsCode", "selection": {"filter": "item", "values": ["Personer1"]}},
        {"code": "Tid", "selection": {"filter": "top", "values": ["1"]}},
    ],
    "response": {"format": "json-stat2"},
}
TIMEZONE = "Europe/Oslo"
# The words a Norwegian place name writes in lower case: Mo i Rana, Sande på Sunnmøre.
LOWER = {"i", "på"}
# Not a street: a numbered placeholder such as Veg 7205, Gate 2, Veinr. 8113 or Vegkode 81682, a temporary name, or a
# name holding a road designation, E6, E-6, Europaveg-39 or FV819.
NOT_A_STREET = re.compile(r"^(Gate|Veg|Vei) ?\d+\b|^(Veinr|Vegkode|Veg nr)\.? ?\d|^Midlertidig |\b(E|FV|Fv|RV|Rv)-? ?\d+\b|^Europaveg")
NUMBER_RANGES = ((1, 9), (10, 99), (100, 999), (1000, 9999))


@dataclass
class Tally:
    """What the register's street addresses add up to."""

    in_municipality: collections.Counter = field(default_factory=collections.Counter)
    codes: collections.defaultdict = field(default_factory=lambda: collections.defaultdict(collections.Counter))
    streets: collections.Counter = field(default_factory=collections.Counter)
    points: collections.defaultdict = field(default_factory=lambda: collections.defaultdict(list))
    numbers: collections.Counter = field(default_factory=collections.Counter)
    written: dict = field(default_factory=dict)


def recased(name):
    """A poststed written as a place name: OSLO to Oslo, MO I RANA to Mo i Rana, NORD-TORPA to Nord-Torpa."""
    words = re.split(r"([ -])", name.lower())
    return "".join(w if w in LOWER or w in " -" else w[:1].upper() + w[1:] for w in words)


def street_name(name):
    return not NOT_A_STREET.search(name) and not geo.in_capitals(name)


def norwegian_name(official, written=None):
    """The name Kartverket writes, among SSB's names in each official language: SSB names a kommune Trondheim -
    Tråante, and Kartverket writes TRONDHEIM. With no written name, the first, as every fylke with more than one
    name puts the Norwegian one first."""
    names = official.split(" - ")
    return next((n for n in names if written and n.upper() == written), names[0])


def tally(addresses):
    """The Tally of addresses, each (postnummer, poststed, kommune, kommune as written, street, number, letter, lat, lon)."""
    t = Tally()
    for code, place, municipality, written, street, number, letter, lat, lon in addresses:
        t.in_municipality[(place, municipality)] += 1
        t.codes[place][code] += 1
        t.points[place].append({"lat": lat, "lon": lon})
        t.numbers[(next(i for i, (_, top) in enumerate(NUMBER_RANGES) if number <= top), bool(letter))] += 1
        t.written.setdefault(municipality, written)
        if street_name(street):
            t.streets[(place, street)] += 1
    return t


def localities(t, people, municipalities):
    """Each poststed whose kommune SSB classifies, with that kommune, its population by the module docstring's rule and
    its centre point. Of two kommuner holding as many of a poststed's addresses, the lower code wins."""
    of_municipality, share, home = collections.Counter(), collections.Counter(), {}
    for (_, municipality), n in t.in_municipality.items():
        of_municipality[municipality] += n
    for (place, municipality), n in sorted(t.in_municipality.items(), key=lambda kv: (-kv[1], kv[0][1])):
        share[place] += people.get(municipality, 0) * n / of_municipality[municipality]
        home.setdefault(place, municipality)
    out = {}
    for place, municipality in home.items():
        if municipality not in municipalities:
            print(f"{place}: kommune {municipality} is not in SSB's classification, dropped", file=sys.stderr)
            continue
        lat, lon = geo.centroid(t.points[place])
        out[place] = {"name": place, "municipality": municipality, "population": max(1, round(share[place])), "lat": f"{lat:.4f}", "lon": f"{lon:.4f}"}
    return out


def address_json(numbers):
    """geo.NO.address, its street-number choice weighted by the register's counts of numbers, by length and letter."""
    items = ",\n".join(
        f'    {{ "format": "{{int({low},{top})}}{"{upper(1)}" if lettered else ""}", "weight": {numbers[(i, lettered)]} }}'
        for lettered in (False, True) for i, (low, top) in enumerate(NUMBER_RANGES) if numbers[(i, lettered)]
    )
    return (
        '{\n  "format": "{street} {street-number}\\n{postal-code} {locality}",\n  "street": "{.locality as l}{l.street.name}",\n'
        f'  "street-number": [\n{items}\n  ],\n'
        '  "postal-code": "{l.postal-code.code}",\n  "locality": "{l.name}"\n}\n'
    )


def klass(cache, classification, date):
    data = source.fetch(KLASS.format(classification, date), cache, f"klass-{classification}-{date}.json", headers={"Accept": "application/json"})
    return {c["code"]: c["name"] for c in json.loads(data)["codes"]}


def population(cache):
    body = json.dumps(POPULATION_QUERY).encode()
    data = json.loads(source.fetch(POPULATION, cache, "ssb-07459.json", data=body, headers={"Content-Type": "application/json"}))
    index = data["dimension"]["Region"]["category"]["index"]
    return {code: data["value"][i] for code, i in index.items() if data["value"][i]}


def addresses(cache, url):
    z = zipfile.ZipFile(io.BytesIO(source.fetch(url, cache, "matrikkelen-adresse.zip", magic=b"PK")))
    name = next(n for n in z.namelist() if n.endswith(".csv"))
    with z.open(name) as f:
        for r in csv.DictReader(io.TextIOWrapper(f, encoding="utf-8-sig", newline=""), delimiter=";"):
            if r["adressetype"] == "vegadresse":
                yield (r["postnummer"], recased(r["poststed"]), r["kommunenummer"], r["kommunenavn"], " ".join(r["adressenavn"].split()),
                       int(r["nummer"]), r["bokstav"], float(r["Nord"]), float(r["Øst"]))


def main():
    a = geo.parser(__doc__, "NO", 1.0, addresses={"default": ADDRESSES}, date={"default": datetime.date.today().isoformat(), "help": "the day SSB's classifications are read at"}).parse_args()
    cache, out = geo.directories(a)
    people = population(cache)
    t = tally(addresses(cache, a.addresses))
    municipality_names = {c: norwegian_name(n, t.written.get(c)) for c, n in klass(cache, MUNICIPALITIES, a.date).items()}
    places = localities(t, people, municipality_names)
    named = geo.top_streets(t.streets, a.street_share, a.min_streets, "addresses")
    regions = {c: {"code": c, "name": norwegian_name(n), "population": people[c], "timezone": TIMEZONE} for c, n in klass(cache, REGIONS, a.date).items() if c in people}
    municipalities = {c: {"code": c, "name": n, "region": c[:2], "population": people[c]} for c, n in municipality_names.items() if c in people}
    kept, dropped = geo.fence(regions, municipalities, places, named, t.codes)
    geo.log_dropped_localities(dropped["locality"])
    geo.exit_on_dropped_parents(dropped)
    places = kept["locality"]

    tsv.write(out / "region.tsv", ["code", "name", "population", "timezone"], [r for _, r in sorted(kept["region"].items())])
    tsv.write(out / "municipality.tsv", ["code", "name", "region", "population"], [m for _, m in sorted(kept["municipality"].items())])
    tsv.write(out / "locality.tsv", ["name", "municipality", "population", "lat", "lon"], [l for _, l in sorted(places.items())])
    tsv.write(out / "postal-code.tsv", ["code", "locality", "addresses"], [{"code": c, "locality": l, "addresses": n} for c, l, n in sorted((c, l, n) for l in places for c, n in t.codes[l].items())])
    tsv.write(out / "street.tsv", ["name", "locality", "addresses"], [s for locality in sorted(places) for s in named[locality]])
    (out / "address.json").write_text(address_json(t.numbers), encoding="utf-8")


if __name__ == "__main__":
    main()
