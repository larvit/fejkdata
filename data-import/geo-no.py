#!/usr/bin/env python3
"""Rebuild data/geo/NO/*.tsv from Kartverket's address register (CC BY 4.0) and SSB (CC BY 4.0).

    data-import/geo-no.py [--addresses URL_OR_FILE] [--cache DIR] [--min-streets N] [--street-share F] [--out DIR]

Every street address (vegadresse) in the register names its street, number, postnummer, poststed and kommune.
A poststed is a locality, with each postnummer and street its addresses spell, each weighted by its addresses. A
poststed's kommune is the one holding most of its addresses. Its population is each kommune's, from SSB, shared
out over that kommune's poststeder by their addresses there: a poststed has no population of its own, and an
address in Oslo houses more people than one in a village. The register spells a poststed in capitals, and
`recased` writes it as a place name.
"""
import collections
import csv
import datetime
import io
import json
import re
import sys
import zipfile

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
# Not a street: a numbered placeholder such as Veg 7205 or Gate 2, or a name holding a European route, E6 or E-6.
NOT_A_STREET = re.compile(r"^(Gate|Veg|Vei) ?\d+$|\bE-? ?\d+\b")


def recased(name):
    """A poststed written as a place name: OSLO to Oslo, MO I RANA to Mo i Rana, NORD-TORPA to Nord-Torpa."""
    words = re.split(r"([ -])", name.lower())
    return "".join(w if w in LOWER or w in " -" else w[:1].upper() + w[1:] for w in words)


def street_name(name):
    # A word of five capitals or more is a name written in capitals, HÅKON DEN GODES GATE.
    return not NOT_A_STREET.search(name) and not any(len(w) >= 5 and w.isupper() for w in name.split())


def klass(cache, classification):
    date = datetime.date.today().isoformat()
    data = source.fetch(KLASS.format(classification, date), cache, f"klass-{classification}.json", headers={"Accept": "application/json"})
    return {c["code"]: c["name"] for c in json.loads(data)["codes"]}


def population(cache):
    body = json.dumps(POPULATION_QUERY).encode()
    data = json.loads(source.fetch(POPULATION, cache, "ssb-07459.json", data=body, headers={"Content-Type": "application/json"}))
    index = data["dimension"]["Region"]["category"]["index"]
    return {code: data["value"][i] for code, i in index.items() if data["value"][i]}


def addresses(cache, url):
    """Each street address as (postnummer, poststed, kommune, street, lat, lon)."""
    z = zipfile.ZipFile(io.BytesIO(source.fetch(url, cache, "matrikkelen-adresse.zip", magic=b"PK")))
    name = next(n for n in z.namelist() if n.endswith(".csv"))
    with z.open(name) as f:
        for r in csv.DictReader(io.TextIOWrapper(f, encoding="utf-8-sig", newline=""), delimiter=";"):
            if r["adressetype"] == "vegadresse":
                yield r["postnummer"], recased(r["poststed"]), r["kommunenummer"], " ".join(r["adressenavn"].split()), float(r["Nord"]), float(r["Øst"])


def main():
    p = geo.parser(__doc__, "NO", 1.0, addresses={"default": ADDRESSES})
    a = p.parse_args()
    cache, out = geo.directories(a)
    municipality_names, region_names, people = klass(cache, MUNICIPALITIES), klass(cache, REGIONS), population(cache)

    in_municipality, codes, streets, points = collections.Counter(), collections.defaultdict(collections.Counter), collections.Counter(), collections.defaultdict(list)
    for code, place, municipality, street, lat, lon in addresses(cache, a.addresses):
        in_municipality[(place, municipality)] += 1
        codes[place][code] += 1
        points[place].append({"lat": lat, "lon": lon})
        if street_name(street):
            streets[(place, street)] += 1
    of_municipality = collections.Counter()
    for (_, municipality), n in in_municipality.items():
        of_municipality[municipality] += n
    share = collections.Counter()
    for (place, municipality), n in in_municipality.items():
        share[place] += people.get(municipality, 0) * n / of_municipality[municipality]
    home = {}
    for (place, municipality), n in sorted(in_municipality.items(), key=lambda kv: -kv[1]):
        home.setdefault(place, municipality)

    places = {}
    for place, municipality in home.items():
        if municipality not in municipality_names:
            print(f"{place}: kommune {municipality} is not in SSB's classification, dropped", file=sys.stderr)
            continue
        lat, lon = geo.centroid(points[place])
        places[place] = {"name": place, "municipality": municipality, "population": max(1, round(share[place])), "lat": f"{lat:.4f}", "lon": f"{lon:.4f}"}
    named = geo.top_streets(streets, a.street_share, a.min_streets, "addresses")
    regions = {c: {"code": c, "name": n, "population": people[c], "timezone": TIMEZONE} for c, n in region_names.items() if c in people}
    municipalities = {c: {"code": c, "name": n, "region": c[:2], "population": people[c]} for c, n in municipality_names.items() if c in people}
    kept, dropped = geo.fence(regions, municipalities, places, named, codes)
    geo.log_dropped_localities(dropped["locality"])
    if dropped["municipality"] or dropped["region"]:
        sys.exit(f"municipalities without a locality: {sorted(dropped['municipality'])}; regions without a municipality: {sorted(dropped['region'])}")
    places = kept["locality"]

    tsv.write(out / "region.tsv", ["code", "name", "population", "timezone"], [r for _, r in sorted(kept["region"].items())])
    tsv.write(out / "municipality.tsv", ["code", "name", "region", "population"], [m for _, m in sorted(kept["municipality"].items())])
    tsv.write(out / "locality.tsv", ["name", "municipality", "population", "lat", "lon"], [l for _, l in sorted(places.items())])
    tsv.write(out / "postal-code.tsv", ["code", "locality", "addresses"], [{"code": c, "locality": l, "addresses": n} for c, l, n in sorted((c, l, n) for l in places for c, n in codes[l].items())])
    tsv.write(out / "street.tsv", ["name", "locality", "addresses"], [s for locality in sorted(places) for s in named[locality]])


if __name__ == "__main__":
    main()
