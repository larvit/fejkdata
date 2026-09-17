#!/usr/bin/env python3
"""Rebuild data/geo/US/*.tsv from the Census Bureau's Gazetteer, population estimates, ZCTA relationships and TIGER/Line files (public domain).

    data-import/geo-us.py [--cache DIR] [--min-population N] [--streets-per-locality N] [--out DIR]

A locality is an incorporated place, or a consolidated city's balance, of at least N
people, in the county holding most of it. Its postal codes are the ZCTAs mostly inside
it, weighted by their TIGER address ranges, and its streets the N names with most
address ranges in those codes.
"""
import argparse
import collections
import concurrent.futures
import csv
import io
import re
import struct
import sys
import time
import urllib.request
import zipfile
from pathlib import Path

GAZETTEER = "https://www2.census.gov/geo/docs/maps-data/data/gazetteer/2026_Gazetteer/2026_Gaz_{}_national.zip"
POPULATION = "https://www2.census.gov/programs-surveys/popest/datasets/2020-2025/{}"
STATES = POPULATION.format("state/totals/NST-EST2025-ALLDATA.csv")
COUNTIES = POPULATION.format("counties/totals/co-est2025-alldata.csv")
PLACES = POPULATION.format("cities/totals/sub-est2025.csv")
ZCTA_PLACE = "https://www2.census.gov/geo/docs/maps-data/data/rel2020/zcta520/tab20_zcta520_place20_natl.txt"
TIGER = "https://www2.census.gov/geo/tiger/TIGER2025/{0}/tl_2025_{1}_{2}.zip"
OUT = Path(__file__).resolve().parent.parent / "data" / "geo" / "US"
CACHE = Path(__file__).resolve().parent / "cache"
ESTIMATE = "POPESTIMATE2025"
CDP = "57"
SUFFIX = re.compile(r" (city and borough|city|town|village|borough|municipality|comunidad|zona urbana|metropolitan government|metro government|consolidated government|unified government|urban county|corporation|plantation)( \(balance\))?$")
# Places whose Census name is a merged government's; the postal city is what an address carries.
NAMES = {"1303440": "Athens", "1304204": "Augusta", "1349008": "Macon", "2148006": "Louisville", "3011397": "Butte", "4732742": "Hartsville", "4752006": "Nashville"}
TIMEZONES = {
    "AK": "America/Anchorage", "AL": "America/Chicago", "AR": "America/Chicago", "AZ": "America/Phoenix",
    "CA": "America/Los_Angeles", "CO": "America/Denver", "CT": "America/New_York", "DC": "America/New_York",
    "DE": "America/New_York", "FL": "America/New_York", "GA": "America/New_York", "HI": "Pacific/Honolulu",
    "IA": "America/Chicago", "ID": "America/Boise", "IL": "America/Chicago", "IN": "America/Indiana/Indianapolis",
    "KS": "America/Chicago", "KY": "America/New_York", "LA": "America/Chicago", "MA": "America/New_York",
    "MD": "America/New_York", "ME": "America/New_York", "MI": "America/Detroit", "MN": "America/Chicago",
    "MO": "America/Chicago", "MS": "America/Chicago", "MT": "America/Denver", "NC": "America/New_York",
    "ND": "America/Chicago", "NE": "America/Chicago", "NH": "America/New_York", "NJ": "America/New_York",
    "NM": "America/Denver", "NV": "America/Los_Angeles", "NY": "America/New_York", "OH": "America/New_York",
    "OK": "America/Chicago", "OR": "America/Los_Angeles", "PA": "America/New_York", "PR": "America/Puerto_Rico",
    "RI": "America/New_York", "SC": "America/New_York", "SD": "America/Chicago", "TN": "America/Chicago",
    "TX": "America/Chicago", "UT": "America/Denver", "VA": "America/New_York", "VT": "America/New_York",
    "WA": "America/Los_Angeles", "WI": "America/Chicago", "WV": "America/New_York", "WY": "America/Denver",
}


def fetch(url, cache, name, magic=b""):
    path = cache / name
    for attempt in range(1, 6):
        if path.exists():
            return path.read_bytes()
        req = urllib.request.Request(url, headers={"User-Agent": "fejkdata data-import"})
        try:
            with urllib.request.urlopen(req, timeout=600) as r:
                data = r.read()
        except OSError:
            data = b""
        if data.startswith(magic) and b"Request Rejected" not in data[:512]:
            path.write_bytes(data)
        else:
            time.sleep(10 * attempt)
    sys.exit(f"{url}: no valid download in 5 attempts")


def text(data):
    try:
        return data.decode("utf-8-sig")
    except UnicodeDecodeError:
        return data.decode("latin-1")


def gazetteer(cache, kind):
    z = zipfile.ZipFile(io.BytesIO(fetch(GAZETTEER.format(kind), cache, f"gaz_{kind}.zip", magic=b"PK")))
    rows = text(z.read(z.namelist()[0])).splitlines()
    header = [h.strip() for h in rows[0].split("|")]
    return [dict(zip(header, (c.strip() for c in row.split("|")))) for row in rows[1:]]


def csv_rows(cache, url, name):
    return list(csv.DictReader(io.StringIO(text(fetch(url, cache, name)))))


def dbf_rows(data, wanted):
    """The records of a dBASE file, the wanted fields only."""
    count, header_len, record_len = struct.unpack("<xxxxIHH", data[:12])
    fields, pos = [], 32
    while data[pos] != 0x0D:
        name = data[pos:pos + 11].split(b"\0")[0].decode()
        fields.append((name, data[pos + 16]))
        pos += 32
    pos = header_len
    for _ in range(count):
        record = data[pos:pos + record_len]
        pos += record_len
        if record[:1] == b"*":
            continue
        row, at = {}, 1
        for name, length in fields:
            if name in wanted:
                row[name] = record[at:at + length].decode("utf-8", "replace").strip()
            at += length
        yield row


def tiger_zip(cache, kind, county):
    return fetch(TIGER.format(kind.upper(), county, kind), cache, f"tl_{county}_{kind}.zip", magic=b"PK")


def tiger(cache, kind, county, wanted):
    z = zipfile.ZipFile(io.BytesIO(tiger_zip(cache, kind, county)))
    return dbf_rows(z.read(f"tl_2025_{county}_{kind}.dbf"), wanted)


def place_name(geoid, name):
    if geoid in NAMES:
        return NAMES[geoid]
    stripped = SUFFIX.sub("", name)
    if stripped == name:
        print(f"{geoid}: no suffix stripped from {name!r}", file=sys.stderr)
    return stripped


def write(path, columns, rows):
    lines = ["\t".join(columns)]
    for row in rows:
        cells = [str(row[c]) for c in columns]
        assert not any(re.search(r"[\t\n{}]", c) for c in cells), row
        lines.append("\t".join(cells))
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"{path}: {len(rows)} rows", file=sys.stderr)


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--min-population", type=int, default=25000)
    p.add_argument("--out", default=str(OUT))
    p.add_argument("--streets-per-locality", type=int, default=10)
    a = p.parse_args()
    cache, out = Path(a.cache), Path(a.out)
    cache.mkdir(parents=True, exist_ok=True)
    out.mkdir(parents=True, exist_ok=True)

    state_population = {r["STATE"]: r[ESTIMATE] for r in csv_rows(cache, STATES, "nst-est2025.csv") if r["SUMLEV"] == "040"}
    county_population = {r["STATE"] + r["COUNTY"]: r[ESTIMATE] for r in csv_rows(cache, COUNTIES, "co-est2025.csv") if r["SUMLEV"] == "050"}
    place_population, county_part = {}, collections.defaultdict(list)
    for r in csv_rows(cache, PLACES, "sub-est2025.csv"):
        if r["SUMLEV"] == "162":
            place_population[r["STATE"] + r["PLACE"]] = int(r[ESTIMATE])
        elif r["SUMLEV"] == "157":
            county_part[r["STATE"] + r["PLACE"]].append((int(r[ESTIMATE]), r["STATE"] + r["COUNTY"]))

    regions = {}
    for r in gazetteer(cache, "state"):
        regions[r["USPS"]] = {"abbr": r["USPS"], "code": r["GEOID"], "name": r["NAME"], "population": state_population[r["GEOID"]], "timezone": TIMEZONES[r["USPS"]]}
    counties, unestimated = {}, collections.Counter()
    for r in gazetteer(cache, "counties"):
        if r["GEOID"] not in county_population:
            unestimated[r["USPS"]] += 1
            continue
        counties[r["GEOID"]] = {"code": r["GEOID"], "name": r["NAME"], "region": r["USPS"], "population": county_population[r["GEOID"]]}
    print(f"counties without a population estimate, dropped: {dict(unestimated)}", file=sys.stderr)
    localities = {}
    for r in gazetteer(cache, "place"):
        geoid, population = r["GEOID"], place_population.get(r["GEOID"], 0)
        if r["FUNCSTAT"] not in "AFN" or r["LSAD"] == CDP or population < a.min_population or not county_part.get(geoid):
            continue
        county = max(county_part[geoid])[1]
        if county not in counties:
            print(f"{geoid} {r['NAME']}: county {county} unknown, dropped", file=sys.stderr)
            continue
        localities[geoid] = {"code": geoid, "name": place_name(geoid, r["NAME"]), "municipality": county, "population": population, "lat": r["INTPTLAT"], "lon": r["INTPTLONG"]}

    zcta_of = {}
    for r in csv.DictReader(io.StringIO(text(fetch(ZCTA_PLACE, cache, "zcta-place.txt"))), delimiter="|"):
        if r["GEOID_ZCTA5_20"] and r["GEOID_PLACE_20"] in localities:
            zcta_of.setdefault(r["GEOID_ZCTA5_20"], []).append((int(r["AREALAND_PART"]), r["GEOID_PLACE_20"]))
    locality_of_zcta = {zcta: max(parts)[1] for zcta, parts in zcta_of.items()}

    needed = sorted({l["municipality"] for l in localities.values()})
    with concurrent.futures.ThreadPoolExecutor(3) as pool:
        list(pool.map(lambda c: (tiger_zip(cache, "addr", c), tiger_zip(cache, "featnames", c)), needed))
    addresses, streets_of = collections.Counter(), collections.Counter()
    for county in needed:
        zips = collections.defaultdict(set)
        for r in tiger(cache, "addr", county, {"TLID", "ZIP"}):
            if r["ZIP"] in locality_of_zcta:
                zips[r["TLID"]].add(r["ZIP"])
                addresses[r["ZIP"]] += 1
        for r in tiger(cache, "featnames", county, {"TLID", "FULLNAME", "PAFLAG"}):
            if r["PAFLAG"] == "P" and r["FULLNAME"]:
                for z in zips.get(r["TLID"], ()):
                    streets_of[(locality_of_zcta[z], r["FULLNAME"])] += 1
    by_locality = collections.defaultdict(list)
    for (locality, name), n in streets_of.items():
        by_locality[locality].append((n, name))
    streets = []
    for locality in sorted(by_locality):
        for n, name in sorted(by_locality[locality], key=lambda s: (-s[0], s[1]))[: a.streets_per_locality]:
            streets.append({"name": name, "locality": locality, "addresses": n})
    for geoid in [l for l in localities if l not in by_locality]:
        print(f"{geoid} {localities[geoid]['name']}: no streets, dropped", file=sys.stderr)
        del localities[geoid]
    postal_codes = [{"code": z, "locality": l, "addresses": addresses[z]} for z, l in sorted(locality_of_zcta.items()) if addresses[z] and l in localities]

    kept_counties = {l["municipality"] for l in localities.values()}
    kept_regions = {counties[c]["region"] for c in kept_counties}
    write(out / "region.tsv", ["abbr", "code", "name", "population", "timezone"], [r for _, r in sorted(regions.items()) if r["abbr"] in kept_regions])
    write(out / "municipality.tsv", ["code", "name", "region", "population"], [c for _, c in sorted(counties.items()) if c["code"] in kept_counties])
    write(out / "locality.tsv", ["code", "name", "municipality", "population", "lat", "lon"], [l for _, l in sorted(localities.items())])
    write(out / "postal-code.tsv", ["code", "locality", "addresses"], postal_codes)
    write(out / "street.tsv", ["name", "locality", "addresses"], streets)


if __name__ == "__main__":
    main()
