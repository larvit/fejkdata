#!/usr/bin/env python3
"""Rebuild data/geo/DK/*.tsv and address.json from Danmarks Adresseregister (DAR), Danmarks Statistik and GeoNames.

    data-import/geo-dk.py [--cache DIR] [--min-streets N] [--street-share F] [--out DIR]

DAR, read through Datafordeler, holds every address in force, each with its street, house number, postnummer,
postdistrikt and kommune. A postdistrikt is a locality, with each postnummer and street its addresses spell, each
weighted by its addresses. A postdistrikt's kommune and population follow geo.homes, from Danmarks Statistik's
population of each kommune. Its centre is the mean of GeoNames' points for its postnumre. Danmarks Statistik numbers
the regions and kommuner in three digits, such as 084 and 101; DAR writes a kommune in four, 0101, and DAGI, the
register of Denmark's divisions, a region as 1084, so the tables write those. Christiansø, kommune 411, ships no place:
its addresses lie in the postdistrikt Gudhjem, whose kommune is Bornholm.
"""
import collections
import csv
import io
import json
import re
import sys
import time
import urllib.request
from dataclasses import dataclass, field

import geo
import source

ADDRESSES = "https://services.datafordeler.dk/DAR/DAR/1/rest/husnummer?format=json&Status=3&MedDybde=false&pagesize={}&page={}"
PAGE = 10000
DIVISIONS = "https://www.dst.dk/klassifikationsbilag/5d18d1e0-400b-4505-92ad-6782915980a3csv_da"
POPULATION = "https://api.statbank.dk/v1/data/FOLK1A/CSV?lang=da&valuePresentation=Code&OMR%C3%85DE=*&Tid=(1)"
TIMEZONE = "Europe/Copenhagen"
NUMBER = re.compile(r"^(\d{1,4})([A-Z]?)$")
CHRISTIANSOE = "411"


@dataclass
class Tally:
    """What DAR's addresses add up to."""

    in_municipality: collections.Counter = field(default_factory=collections.Counter)
    codes: collections.defaultdict = field(default_factory=lambda: collections.defaultdict(collections.Counter))
    streets: collections.Counter = field(default_factory=collections.Counter)
    numbers: collections.Counter = field(default_factory=collections.Counter)


def parsed(label, number, road):
    """An address as (postnummer, postdistrikt, kommune, street, number, letter), from DAR's label of it, such as
    "Dalsgårdsvej 7B, Dalsgård, 8963 Auning", its house number and its road's code, such as 0706-0287; None where the
    label does not open with a street and that number."""
    parts, match = label.split(", "), NUMBER.match(number)
    code, _, place = parts[-1].partition(" ")
    if len(parts) < 2 or not match or not parts[0].endswith(" " + number) or not place:
        return None
    return code, place, road[:4], parts[0][: -len(number) - 1], int(match[1]), match[2]


def tally(addresses):
    t = Tally()
    for code, place, municipality, street, number, letter in addresses:
        t.in_municipality[(place, municipality)] += 1
        t.codes[place][code] += 1
        t.numbers[geo.number_range(number, bool(letter))] += 1
        if not geo.in_capitals(street):
            t.streets[(place, street)] += 1
    return t


def dar_pages(path):
    """Writes path as a TSV of each address's label, house number and road code, from DAR's pages of addresses."""
    with open(path.with_suffix(".part"), "w", encoding="utf-8") as out:
        page = 1
        while True:
            rows = json.loads(fetch(ADDRESSES.format(PAGE, page)))
            for r in rows:
                out.write(f"{r['adgangsadressebetegnelse']}\t{r.get('husnummertekst', '')}\t{r['vejmidte']}\n")
            print(f"DAR page {page}: {len(rows)} addresses", file=sys.stderr)
            if len(rows) < PAGE:
                break
            page += 1
    path.with_suffix(".part").rename(path)


def fetch(url):
    for attempt in range(1, 6):
        req = urllib.request.Request(url, headers={"User-Agent": "fejkdata-data-import (https://github.com/larvit/fejkdata)"})
        try:
            with urllib.request.urlopen(req, timeout=600) as r:
                return r.read()
        except OSError as e:
            print(f"{url}: {e}, attempt {attempt}", file=sys.stderr)
            time.sleep(10 * attempt)
    sys.exit(f"{url}: no valid download in 5 attempts")


def addresses(cache):
    path = cache / "dar-husnummer.tsv"
    if not path.exists():
        dar_pages(path)
    with open(path, encoding="utf-8") as f:
        for line in f:
            address = parsed(*line.rstrip("\n").split("\t"))
            if address:
                yield address


def divisions(cache):
    """Danmarks Statistik's regions and kommuner, each code mapped to its name, and each kommune's code to its region's."""
    text = source.fetch(DIVISIONS, cache, "dst-nuts.csv").decode("utf-8-sig")
    regions, municipalities, region_of, region = {}, {}, {}, None
    for r in sorted(csv.DictReader(io.StringIO(text), delimiter=";"), key=lambda r: int(r["SEKVENS"])):
        if r["NIVEAU"] == "1":
            region = r["KODE"]
            regions[region] = r["TITEL"]
        elif r["NIVEAU"] == "3":
            municipalities[r["KODE"]] = r["TITEL"]
            region_of[r["KODE"]] = region
    return regions, municipalities, region_of


def population(cache):
    text = source.fetch(POPULATION, cache, "dst-folk1a.csv").decode("utf-8-sig")
    return {r["OMRÅDE"]: int(r["INDHOLD"]) for r in csv.DictReader(io.StringIO(text), delimiter=";")}


def localities(t, people, points):
    """Each postdistrikt, with its kommune and population by geo.homes, and the centre of its postnumre's points."""
    out = {}
    for place, (municipality, n) in geo.homes(t.in_municipality, people).items():
        centre = geo.centroid([points[c] for c in t.codes[place] if c in points])
        if not centre:
            print(f"{place}: GeoNames places none of its postnumre, dropped", file=sys.stderr)
            continue
        out[place] = {"name": place, "municipality": municipality, "population": n, "lat": f"{centre[0]:.4f}", "lon": f"{centre[1]:.4f}"}
    return out


def main():
    a = geo.parser(__doc__, "DK", 1.0).parse_args()
    cache, out = geo.directories(a)
    region_names, municipality_names, region_of = divisions(cache)
    dst = population(cache)
    people = {"0" + c: dst[c] for c in municipality_names if c in dst}
    t = tally(addresses(cache))
    points = {r["code"]: r for r in geo.geonames(cache, "DK") if r["lat"] is not None}
    places = localities(t, people, points)
    named = geo.top_streets(t.streets, a.street_share, a.min_streets, "addresses")
    regions = {"1" + c: {"code": "1" + c, "name": n, "population": dst[c], "timezone": TIMEZONE} for c, n in region_names.items()}
    municipalities = {"0" + c: {"code": "0" + c, "name": n, "region": "1" + region_of[c], "population": dst[c]} for c, n in municipality_names.items() if c in dst and c != CHRISTIANSOE}
    kept, dropped = geo.fence(regions, municipalities, places, named, t.codes)
    geo.log_dropped_localities(dropped["locality"])
    geo.exit_on_dropped_parents(dropped)
    geo.write_tables(out, kept, [{"code": c, "locality": l, "addresses": n} for l in kept["locality"] for c, n in t.codes[l].items()], named)
    (out / "address.json").write_text(geo.address_json(t.numbers), encoding="utf-8")


if __name__ == "__main__":
    main()
