#!/usr/bin/env python3
"""Rebuild data/geo/SE/*.tsv from SCB (CC0), GeoNames (CC BY 4.0) and Trafikverket NVDB (CC0).

    TRAFIKVERKET_API_KEY=… data-import/geo-se.py [--key-file FILE] [--cache DIR] [--min-streets N] [--street-share F] [--out DIR]
"""
import collections
import csv
import io
import json
import math
import os
import re
import sys
import urllib.request
from pathlib import Path

import geo
import source
import xlsx

CODES = "https://www.scb.se/contentassets/7a89e48960f741e08918e489ea36354a/kommunlankod-2026.xlsx"
POPULATION = "https://api.scb.se/OV0104/v1/doris/sv/ssd/START/BE/BE0101/BE0101A/BefolkningNy"
POPULATION_QUERY = {
    "query": [
        {"code": "Region", "selection": {"filter": "all", "values": ["*"]}},
        {"code": "ContentsCode", "selection": {"filter": "item", "values": ["BE0101N1"]}},
        {"code": "Tid", "selection": {"filter": "top", "values": ["1"]}},
    ],
    "response": {"format": "json"},
}
TATORTER = "https://geodata.scb.se/geoserver/stat/wfs?service=WFS&version=2.0.0&request=GetFeature&typeNames=stat:Tatorter_2023&outputFormat=csv&srsName=EPSG:4326&propertyName=tatort,kommun,bef,sp_geometry"
# SCB's open data gives a småort, 50 to 199 people, a code and no name, so a postort is matched to one by position.
# srsName asks for WGS84, the coordinates the postal codes use.
SMAORTER = "https://geodata.scb.se/geoserver/stat/wfs?service=WFS&version=2.0.0&request=GetFeature&typeNames=stat:Smaorter_2023&outputFormat=csv&srsName=EPSG:4326&propertyName=kommun,bef,sp_geometry"
NVDB = "https://api.trafikinfo.trafikverket.se/v2/data.json"
NVDB_PAGE = 50000
TIMEZONE = "Europe/Stockholm"
# This script's reading of the "exceptions in big cities" in docs/research/research-geo-se.md, Postal codes → Structure.
ONE_POSITION = {"Stockholm", "Göteborg", "Malmö"}
# A tätort's floor: "minst 200 invånare", docs/research/research-geo-se.md, Lantmäteriet → Ortnamn. The weight of
# a postort that is no tätort, is not named as its kommun and lies in no småort; docs/decisions.md says why it stays.
UNMATCHED_POPULATION = 200
# Not a street: a road number, a ferry route, a cycle path's or road association's label, a placeholder, or two names
# or a description in one.
NOT_A_STREET = re.compile(r"^(Europaväg|Länsväg|Riksväg) \d|^Färjeled |\b(GC|SMF|VSF)\b|^NN$|,")
PARTICLES = {"af", "de", "von"}


def scb_codes(cache):
    regions, municipalities = {}, {}
    for cells in xlsx.rows(source.fetch(CODES, cache, "kommunlankod.xlsx", magic=b"PK")):
        if len(cells) < 2 or not re.fullmatch(r"\d{2}|\d{4}", cells[0]):
            continue
        (regions if len(cells[0]) == 2 else municipalities)[cells[0]] = cells[1].strip()
    return regions, municipalities


def scb_population(cache):
    body = json.dumps(POPULATION_QUERY).encode()
    data = source.fetch(POPULATION, cache, "befolkning.json", data=body, headers={"Content-Type": "application/json"})
    return {row["key"][0]: row["values"][0] for row in json.loads(data.decode("utf-8-sig"))["data"]}


def wkt_rings(wkt):
    """The rings of a WKT polygon, each a list of (lon, lat). A ring is any parenthesised run, so a hole is a ring too."""
    return [[tuple(map(float, p.split())) for p in ring.split(",")] for ring in re.findall(r"\(([^()]+)\)", wkt)]


def scb_tatorter(cache):
    """Each tätort name's (municipality, population) pairs, and every tätort as (name, municipality, rings)."""
    csv.field_size_limit(sys.maxsize)
    text = source.fetch(TATORTER, cache, "tatorter-areas.csv", magic=b"FID").decode("utf-8")
    by_name, areas = collections.defaultdict(list), []
    for r in csv.DictReader(io.StringIO(text)):
        by_name[r["tatort"]].append((r["kommun"], int(r["bef"])))
        areas.append((r["tatort"], r["kommun"], wkt_rings(r["sp_geometry"])))
    return by_name, areas


def scb_smaorter(cache):
    """Each municipality's småorter as (population, rings)."""
    csv.field_size_limit(sys.maxsize)
    text = source.fetch(SMAORTER, cache, "smaorter.csv", magic=b"FID").decode("utf-8")
    by_municipality = collections.defaultdict(list)
    for r in csv.DictReader(io.StringIO(text)):
        by_municipality[r["kommun"]].append((int(r["bef"]), wkt_rings(r["sp_geometry"])))
    return by_municipality


def inside(lon, lat, rings):
    """Whether the point lies in the polygons made of rings: a ray running east from it crosses an odd number of their
    edges there, and a hole's ring, crossed too, puts the point back outside."""
    odd = False
    for ring in rings:
        for (x1, y1), (x2, y2) in zip(ring, ring[1:]):
            if (y1 > lat) != (y2 > lat) and lon < (x2 - x1) * (lat - y1) / (y2 - y1) + x1:
                odd = not odd
    return odd


class Areas:
    """Finds the area holding a point, each area's edges filed by latitude band, so checking a point reads only its own band."""

    BAND = 0.01

    def __init__(self, areas):
        self.at = collections.defaultdict(list)
        for value, rings in areas:
            edges = collections.defaultdict(list)
            for ring in rings:
                for a, b in zip(ring, ring[1:]):
                    for band in range(int(min(a[1], b[1]) // self.BAND), int(max(a[1], b[1]) // self.BAND) + 1):
                        edges[band].append((a, b))
            lons = [p[0] for ring in rings for p in ring]
            for band, held in edges.items():
                self.at[band].append((min(lons), max(lons), value, held))

    def find(self, lon, lat):
        for west, east, value, edges in self.at[int(lat // self.BAND)]:
            if west <= lon <= east and inside(lon, lat, edges):
                return value
        return None


def nvdb_segments(cache, key):
    path = cache / "nvdb-gatunamn.tsv"
    if not path.exists():
        with open(path.with_suffix(".part"), "w", encoding="utf-8") as out:
            change = "0"
            while True:
                query = (
                    f'<REQUEST><LOGIN authenticationkey="{key}"/>'
                    f'<QUERY objecttype="Gatunamn" namespace="vägdata.nvdb_dk_o" schemaversion="1.0" limit="{NVDB_PAGE}" changeid="{change}">'
                    "<FILTER><EQ name=\"Deleted\" value=\"false\"/></FILTER>"
                    "<INCLUDE>Namn</INCLUDE><INCLUDE>Geometry.WKT-WGS84-3D</INCLUDE></QUERY></REQUEST>"
                )
                req = urllib.request.Request(NVDB, data=query.encode(), headers={"Content-Type": "text/xml"})
                with urllib.request.urlopen(req, timeout=600) as r:
                    result = json.load(r)["RESPONSE"]["RESULT"][0]
                rows = result.get("Gatunamn", [])
                for row in rows:
                    m = re.match(r"LINESTRING Z \(([-\d.]+) ([-\d.]+) ", row.get("Geometry", {}).get("WKT-WGS84-3D", ""))
                    name = " ".join(row.get("Namn", "").split())
                    if m and name:
                        out.write(f"{name}\t{m.group(1)}\t{m.group(2)}\n")
                change = result["INFO"]["LASTCHANGEID"]
                if len(rows) < NVDB_PAGE:
                    break
        path.with_suffix(".part").rename(path)
    for line in path.read_text(encoding="utf-8").splitlines():
        name, lon, lat = line.split("\t")
        yield name, float(lat), float(lon)


def street_delivery(name, codes):
    """The codes delivered to a street. A postort's own prefix is three digits where one three-digit prefix starts at
    least half its codes, else two. Outside ONE_POSITION's cities, the digit after that prefix is 0 or 1 for boxes and 8
    for reply mail; inside them, a second digit 0 marks boxes. The box and reply-mail digits are per
    docs/research/research-geo-se.md, Postal codes → Structure, and docs/decisions.md says why digit 9 stays."""
    if name in ONE_POSITION:
        return [c for c in codes if c[1] != "0"]
    largest = collections.Counter(c[:3] for c in codes).most_common(1)[0][1]
    length = 3 if largest * 2 >= len(codes) else 2
    return [c for c in codes if c[length] not in "018"]


def municipality_of(name, rows, point, tatorter, municipalities, smaorter):
    named = [code for code, n in municipalities.items() if n == name]
    if named:
        return named[0], "kommun"
    voted = collections.Counter(r["municipality"] for r in rows if r["municipality"] in municipalities)
    matches = tatorter.get(name, [])
    if matches:
        in_vote = [m for m in matches if voted and m[0] == voted.most_common(1)[0][0]]
        return max(in_vote or matches, key=lambda m: m[1])[0], "tatort"
    held = smaort_municipality(point, municipalities, smaorter)
    if held:
        return held, "smaort"
    if voted:
        return voted.most_common(1)[0][0], "codes"
    return None, "unplaced"


def smaort_municipality(point, municipalities, smaorter):
    """The kommun of the småort holding point, or None."""
    if point is None:
        return None
    lat, lon = point
    return next((m for m, places in smaorter.items() if m in municipalities for _, rings in places if inside(lon, lat, rings)), None)


def population_of(name, municipality, point, tatorter, municipalities, population, smaorter):
    matches = [m for m in tatorter.get(name, []) if m[0] == municipality]
    if matches:
        return str(max(m[1] for m in matches))
    if municipalities[municipality] == name:
        return population[municipality]
    lat, lon = point
    return str(next((n for n, rings in smaorter.get(municipality, []) if inside(lon, lat, rings)), UNMATCHED_POPULATION))


def well_cased(name):
    """Whether each word of name is capitalised: GeoNames also spells some postorter, such as STOCKHOLM and örebro, in
    capitals or lower case."""
    return all(part[:1].isupper() and (len(part) == 1 or not part.isupper()) for part in re.split(r"[ -]", name))


def localities(codes, tatorter, municipalities, population, smaorter):
    """Each postort with its municipality, weight, centroid and street-delivery codes."""
    by_locality = collections.defaultdict(list)
    for r in codes:
        by_locality[r["locality"]].append(r)
    out, how = {}, collections.Counter()
    for name, rows in by_locality.items():
        point = geo.centroid(rows)
        municipality, method = municipality_of(name, rows, point, tatorter, municipalities, smaorter)
        how[method] += 1
        kept = street_delivery(name, [r["code"].replace(" ", "") for r in rows])
        if municipality is None or not kept or point is None or not well_cased(name):
            continue
        lat, lon = point
        out[name] = {"name": name, "municipality": municipality, "population": population_of(name, municipality, point, tatorter, municipalities, population, smaorter), "lat": f"{lat:.4f}", "lon": f"{lon:.4f}", "codes": kept}
    print(f"municipality by {dict(how)}; {len(by_locality) - len(out)} postorter dropped before streets", file=sys.stderr)
    return out


def street_name(name):
    opening = name[0].isupper() or name.split()[0] in PARTICLES
    return opening and not NOT_A_STREET.search(name) and not geo.in_capitals(name)


def reach_distance(locality, lat, lon):
    """The distance to a postort's centre over its reach, the fourth root of its codes. docs/decisions.md says why
    the fourth root."""
    scale = math.cos(math.radians(lat))
    return math.hypot(float(locality["lat"]) - lat, (float(locality["lon"]) - lon) * scale) / len(locality["codes"]) ** 0.25


def streets(segments, codes, localities, tatorter, share, minimum):
    """The names with most segments per locality. A segment goes to its nearest code centroid, unless it lies in a tätort
    named as a postort of the same kommun and that postort is nearer by reach_distance: GeoNames puts most of a big
    postort's codes on few points, so plain nearest leaves it only the streets around those points."""
    nearest = geo.Nearest((r["lat"], r["lon"], r["locality"]) for r in codes if r["lat"] is not None and r["locality"] in localities)
    namesakes = Areas((name, rings) for name, municipality, rings in tatorter if name in localities and localities[name]["municipality"] == municipality)
    count = collections.Counter()
    for name, lat, lon in segments:
        if street_name(name):
            near, own = nearest.find(lat, lon), namesakes.find(lon, lat)
            wins = own and reach_distance(localities[own], lat, lon) <= reach_distance(localities[near], lat, lon)
            count[(own if wins else near, name)] += 1
    return geo.top_streets(count, share, minimum, "segments")


def main():
    a = geo.parser(__doc__, "SE", 1.0, key_file={"help": "file holding the Trafikverket API key; TRAFIKVERKET_API_KEY otherwise"}).parse_args()
    key = Path(a.key_file).read_text().strip() if a.key_file else os.environ.get("TRAFIKVERKET_API_KEY")
    if not key:
        sys.exit("set TRAFIKVERKET_API_KEY or pass --key-file")
    cache, out = geo.directories(a)

    region_names, municipality_names = scb_codes(cache)
    population = scb_population(cache)
    codes = geo.geonames(cache, "SE")
    tatorter, areas = scb_tatorter(cache)
    places = localities(codes, tatorter, municipality_names, population, scb_smaorter(cache))
    named = streets(nvdb_segments(cache, key), codes, places, areas, a.street_share, a.min_streets)
    regions = {c: {"code": c, "name": n, "population": population[c], "timezone": TIMEZONE} for c, n in region_names.items()}
    municipalities = {c: {"code": c, "name": n, "region": c[:2], "population": population[c]} for c, n in municipality_names.items()}
    kept, dropped = geo.fence(regions, municipalities, places, named, {k: p["codes"] for k, p in places.items()})
    geo.log_dropped_localities(dropped["locality"])
    geo.exit_on_dropped_parents(dropped)
    postal_codes = [{"code": f"{c[:3]} {c[3:]}", "locality": loc["name"]} for loc in kept["locality"].values() for c in loc["codes"]]
    geo.write_tables(out, kept, postal_codes, named, postal_code=["code", "locality"], street=["name", "locality", "segments"])


if __name__ == "__main__":
    main()
