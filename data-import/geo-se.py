#!/usr/bin/env python3
"""Rebuild data/geo/SE/*.tsv from SCB (CC0), GeoNames (CC BY 4.0) and Trafikverket NVDB (CC0).

    TRAFIKVERKET_API_KEY=… data-import/geo-se.py [--key-file FILE] [--cache DIR] [--streets-per-locality N] [--out DIR]

A locality is a GeoNames postort, placed in the municipality it names, else of its
tätort, else of most of its codes, and weighted by its tätort's population, else its
municipality's, else 200. Box codes are dropped by the digit after the postort's own
prefix. Each NVDB street segment goes to the nearest postal code centroid; a locality
keeps the N names with most segments.
"""
import argparse
import collections
import csv
import io
import json
import math
import os
import re
import sys
import urllib.request
import xml.etree.ElementTree as ET
import zipfile
from pathlib import Path

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
TATORTER = "https://geodata.scb.se/geoserver/stat/wfs?service=WFS&version=2.0.0&request=GetFeature&typeNames=stat:Tatorter_2023&outputFormat=csv&propertyName=tatort,kommun,bef"
POSTAL_CODES = "https://download.geonames.org/export/zip/SE.zip"
NVDB = "https://api.trafikinfo.trafikverket.se/v2/data.json"
NVDB_PAGE = 50000
OUT = Path(__file__).resolve().parent.parent / "data" / "geo" / "SE"
CACHE = Path(__file__).resolve().parent / "cache"
TIMEZONE = "Europe/Stockholm"
ONE_POSITION = {"Stockholm", "Göteborg", "Malmö"}
UNMATCHED_POPULATION = 200
XLSX_NS = {"m": "http://schemas.openxmlformats.org/spreadsheetml/2006/main"}


def fetch(url, cache, name, data=None, headers=None):
    path = cache / name
    if not path.exists():
        req = urllib.request.Request(url, data=data, headers=headers or {})
        with urllib.request.urlopen(req, timeout=300) as r:
            path.write_bytes(r.read())
    return path.read_bytes()


def xlsx_rows(data):
    z = zipfile.ZipFile(io.BytesIO(data))
    strings = ["".join(t.text or "" for t in si.iter("{%s}t" % XLSX_NS["m"])) for si in ET.fromstring(z.read("xl/sharedStrings.xml")).findall("m:si", XLSX_NS)]
    sheet = ET.fromstring(z.read("xl/worksheets/sheet1.xml"))
    for row in sheet.findall(".//m:row", XLSX_NS):
        cells = []
        for c in row.findall("m:c", XLSX_NS):
            v = c.find("m:v", XLSX_NS)
            cells.append("" if v is None else strings[int(v.text)] if c.get("t") == "s" else v.text)
        yield cells


def scb_codes(cache):
    regions, municipalities = {}, {}
    for cells in xlsx_rows(fetch(CODES, cache, "kommunlankod.xlsx")):
        if len(cells) < 2 or not re.fullmatch(r"\d{2}|\d{4}", cells[0]):
            continue
        (regions if len(cells[0]) == 2 else municipalities)[cells[0]] = cells[1].strip()
    return regions, municipalities


def scb_population(cache):
    body = json.dumps(POPULATION_QUERY).encode()
    data = fetch(POPULATION, cache, "befolkning.json", data=body, headers={"Content-Type": "application/json"})
    return {row["key"][0]: row["values"][0] for row in json.loads(data.decode("utf-8-sig"))["data"]}


def scb_tatorter(cache):
    text = fetch(TATORTER, cache, "tatorter.csv").decode("utf-8")
    by_name = collections.defaultdict(list)
    for r in csv.DictReader(io.StringIO(text)):
        by_name[r["tatort"]].append((r["kommun"], int(r["bef"])))
    return by_name


def geonames(cache):
    z = zipfile.ZipFile(io.BytesIO(fetch(POSTAL_CODES, cache, "SE.zip")))
    rows = []
    for line in z.read("SE.txt").decode("utf-8").splitlines():
        f = line.split("\t")
        lat, lon = (float(f[9]), float(f[10])) if f[9] and f[10] else (None, None)
        rows.append({"code": f[1], "locality": f[2], "municipality": f[6], "lat": lat, "lon": lon})
    return rows


def nvdb_segments(cache, key):
    path = cache / "nvdb-gatunamn.tsv"
    if not path.exists():
        with open(path, "w", encoding="utf-8") as out:
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
    for line in path.read_text(encoding="utf-8").splitlines():
        name, lon, lat = line.split("\t")
        yield name, float(lat), float(lon)


class Nearest:
    """Nearest point by an equirectangular distance, over a degree grid."""

    def __init__(self, points, cell=0.05):
        self.cell = cell
        self.grid = collections.defaultdict(list)
        for lat, lon, value in points:
            self.grid[(int(lat // cell), int(lon // cell))].append((lat, lon, value))

    def find(self, lat, lon):
        ci, cj = int(lat // self.cell), int(lon // self.cell)
        best, best_d = None, math.inf
        ring = 0
        while ring < 400:
            for i in range(ci - ring, ci + ring + 1):
                for j in range(cj - ring, cj + ring + 1):
                    if max(abs(i - ci), abs(j - cj)) != ring:
                        continue
                    for plat, plon, value in self.grid.get((i, j), ()):
                        d = (plat - lat) ** 2 + ((plon - lon) * math.cos(math.radians(lat))) ** 2
                        if d < best_d:
                            best, best_d = value, d
            if best is not None and math.sqrt(best_d) < ring * self.cell * math.cos(math.radians(lat)):
                return best
            ring += 1
        return best


def street_delivery(name, codes):
    """The codes delivered to a street: the digit after the postort's own prefix says box, company or reply."""
    if name in ONE_POSITION:
        return [c for c in codes if c[1] != "0"]
    largest = collections.Counter(c[:3] for c in codes).most_common(1)[0][1]
    length = 3 if largest * 2 >= len(codes) else 2
    return [c for c in codes if c[length] not in "018"]


def municipality_of(name, rows, tatorter, municipalities):
    named = [code for code, n in municipalities.items() if n == name]
    if named:
        return named[0], "kommun"
    voted = collections.Counter(r["municipality"] for r in rows if r["municipality"] in municipalities)
    matches = tatorter.get(name, [])
    if matches:
        in_vote = [m for m in matches if voted and m[0] == voted.most_common(1)[0][0]]
        return max(in_vote or matches, key=lambda m: m[1])[0], "tatort"
    if voted:
        return voted.most_common(1)[0][0], "codes"
    return None, "unplaced"


def population_of(name, municipality, tatorter, municipalities, population):
    matches = [m for m in tatorter.get(name, []) if m[0] == municipality]
    if matches:
        return str(max(m[1] for m in matches))
    return population[municipality] if municipalities[municipality] == name else str(UNMATCHED_POPULATION)


def well_cased(name):
    return all(part[:1].isupper() and (len(part) == 1 or not part.isupper()) for part in re.split(r"[ -]", name))


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
    p.add_argument("--key-file", help="file holding the Trafikverket API key; TRAFIKVERKET_API_KEY otherwise")
    p.add_argument("--out", default=str(OUT))
    p.add_argument("--streets-per-locality", type=int, default=10)
    a = p.parse_args()
    key = Path(a.key_file).read_text().strip() if a.key_file else os.environ.get("TRAFIKVERKET_API_KEY")
    if not key:
        sys.exit("set TRAFIKVERKET_API_KEY or pass --key-file")
    cache, out = Path(a.cache), Path(a.out)
    cache.mkdir(parents=True, exist_ok=True)
    out.mkdir(parents=True, exist_ok=True)

    regions, municipalities = scb_codes(cache)
    population = scb_population(cache)
    tatorter = scb_tatorter(cache)
    codes = geonames(cache)

    by_locality = collections.defaultdict(list)
    for r in codes:
        by_locality[r["locality"]].append(r)
    localities, how = {}, collections.Counter()
    for name, rows in by_locality.items():
        municipality, method = municipality_of(name, rows, tatorter, municipalities)
        how[method] += 1
        kept = street_delivery(name, [r["code"].replace(" ", "") for r in rows])
        with_point = [r for r in rows if r["lat"] is not None]
        if municipality is None or not kept or not with_point or not well_cased(name):
            continue
        lat = sum(r["lat"] for r in with_point) / len(with_point)
        lon = sum(r["lon"] for r in with_point) / len(with_point)
        localities[name] = {"name": name, "municipality": municipality, "population": population_of(name, municipality, tatorter, municipalities, population), "lat": f"{lat:.4f}", "lon": f"{lon:.4f}", "codes": kept}
    print(f"municipality by {dict(how)}; {len(by_locality) - len(localities)} postorter dropped", file=sys.stderr)

    centroids = [(r["lat"], r["lon"], r["locality"]) for r in codes if r["lat"] is not None and r["locality"] in localities]
    nearest_locality = Nearest(centroids)
    segments = collections.Counter()
    for name, lat, lon in nvdb_segments(cache, key):
        segments[(nearest_locality.find(lat, lon), name)] += 1
    streets_of = collections.defaultdict(list)
    for (locality, name), n in segments.items():
        streets_of[locality].append((n, name))
    streets = []
    for locality in sorted(streets_of):
        for n, name in sorted(streets_of[locality], key=lambda s: (-s[0], s[1]))[: a.streets_per_locality]:
            streets.append({"name": name, "locality": locality, "segments": n})
    for name in [l for l in localities if l not in streets_of]:
        del localities[name]

    empty = sorted(m for m in municipalities if not any(l["municipality"] == m for l in localities.values()))
    if empty:
        sys.exit(f"municipalities without a locality: {empty}")
    write(out / "region.tsv", ["code", "name", "population", "timezone"], [{"code": c, "name": n, "population": population[c], "timezone": TIMEZONE} for c, n in sorted(regions.items())])
    write(out / "municipality.tsv", ["code", "name", "region", "population"], [{"code": c, "name": n, "region": c[:2], "population": population[c]} for c, n in sorted(municipalities.items())])
    write(out / "locality.tsv", ["name", "municipality", "population", "lat", "lon"], [l for _, l in sorted(localities.items())])
    write(out / "postal-code.tsv", ["code", "locality"], sorted(({"code": f"{c[:3]} {c[3:]}", "locality": l["name"]} for l in localities.values() for c in l["codes"]), key=lambda r: r["code"]))
    write(out / "street.tsv", ["name", "locality", "segments"], streets)


if __name__ == "__main__":
    main()
