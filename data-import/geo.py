"""The steps a country's geo import may share with another's."""
import argparse
import collections
import io
import math
import sys
import zipfile
from pathlib import Path

import source
import tsv

DATA = Path(__file__).resolve().parent.parent / "data" / "geo"
CACHE = Path(__file__).resolve().parent / "cache"
GEONAMES = "https://download.geonames.org/export/zip/{}.zip"
NUMBER_RANGES = ((1, 9), (10, 99), (100, 999), (1000, 9999))
COLUMNS = {
    "region": ["code", "name", "population", "timezone"],
    "municipality": ["code", "name", "region", "population"],
    "locality": ["name", "municipality", "population", "lat", "lon"],
    "postal_code": ["code", "locality", "addresses"],
    "street": ["name", "locality", "addresses"],
}


def parser(doc, country, street_share, **own):
    """A parser for the options every country's import takes, plus the script's own: each keyword names one option, with
    "_" standing for "-"."""
    options = {"cache": {"default": str(CACHE)}, "out": {"default": str(DATA / country)}, "min_streets": {"type": int, "default": 10}, "street_share": {"type": float, "default": street_share}, **own}
    p = argparse.ArgumentParser(description=doc.splitlines()[0])
    for name in sorted(options):
        p.add_argument("--" + name.replace("_", "-"), **options[name])
    return p


def directories(args):
    """The cache and out directories of the parsed options, both created."""
    cache, out = Path(args.cache), Path(args.out)
    cache.mkdir(parents=True, exist_ok=True)
    out.mkdir(parents=True, exist_ok=True)
    return cache, out


def centroid(rows):
    """The mean lat and lon of the rows with a point, or None where no row has one."""
    points = [r for r in rows if r["lat"] is not None]
    if not points:
        return None
    return sum(r["lat"] for r in points) / len(points), sum(r["lon"] for r in points) / len(points)


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


def top_streets(count, share, minimum, column):
    """count maps (locality, street name) to a count. Returns, per locality, rows for its most-counted names, share of
    them rounded up and at least minimum, ties broken by name; each row holds its count under column."""
    of = collections.defaultdict(list)
    for (locality, name), n in count.items():
        of[locality].append((n, name))
    return {locality: [{"name": name, "locality": locality, column: n} for n, name in sorted(named, key=lambda s: (-s[0], s[1]))[:max(minimum, math.ceil(len(named) * share))]] for locality, named in of.items()}


def fence(regions, municipalities, localities, streets, postal_codes):
    """Meet the loader's fence that every parent row has a child row. Keeps each locality with entries in both streets
    and postal_codes, then each municipality and region left with a child.

    regions and municipalities map a code to a row; a municipality row names its region under "region". localities,
    streets and postal_codes map a locality's key to its row, its streets and its postal codes. A locality row holds
    name, municipality, population, lat and lon, plus its country's own keys.

    Returns kept, then dropped: each maps "locality", "municipality" and "region" to that table's rows."""
    kept = {"locality": {k: l for k, l in localities.items() if streets.get(k) and postal_codes.get(k)}}
    municipality_codes = {l["municipality"] for l in kept["locality"].values()}
    kept["municipality"] = {c: m for c, m in municipalities.items() if c in municipality_codes}
    region_codes = {m["region"] for m in kept["municipality"].values()}
    kept["region"] = {c: r for c, r in regions.items() if c in region_codes}
    every = {"locality": localities, "municipality": municipalities, "region": regions}
    dropped = {table: {k: row for k, row in rows.items() if k not in kept[table]} for table, rows in every.items()}
    return kept, dropped


def exit_on_dropped_parents(dropped):
    """Exits where fence dropped a municipality or a region, naming each; a country whose every one must ship calls it."""
    if dropped["municipality"] or dropped["region"]:
        sys.exit(f"municipalities without a locality: {sorted(dropped['municipality'])}; regions without a municipality: {sorted(dropped['region'])}")


def in_capitals(name):
    """Whether a word of name is five capitals or more, as in a street name a source writes in capitals. A shorter
    word in capitals is an acronym or a Roman numeral: IOGT-vägen, Karl XII gatan."""
    return any(len(w) >= 5 and w.isupper() for w in name.split())


def log_dropped_localities(rows):
    """Logs each locality row as dropped, by its key, plus its name where that differs."""
    for key, row in rows.items():
        label = key if key == row["name"] else f"{key} {row['name']}"
        print(f"{label}: no street or no postal code, dropped", file=sys.stderr)


def geonames(cache, country):
    """GeoNames' postal codes of a country (CC BY 4.0), each with its place name, municipality code and point."""
    z = zipfile.ZipFile(io.BytesIO(source.fetch(GEONAMES.format(country), cache, f"{country}.zip", magic=b"PK")))
    rows = []
    for line in z.read(f"{country}.txt").decode("utf-8").splitlines():
        f = line.split("\t")
        lat, lon = (float(f[9]), float(f[10])) if f[9] and f[10] else (None, None)
        rows.append({"code": f[1], "locality": f[2], "municipality": f[6], "lat": lat, "lon": lon})
    return rows


def homes(in_municipality, people):
    """Each locality's municipality and population. in_municipality maps (locality, municipality) to the addresses they
    share, and people maps a municipality to its population. A locality's municipality holds most of its addresses, the
    lower code winning a tie. Its population is each municipality's, shared out over that municipality's localities by
    their addresses there."""
    of_municipality, share, home = collections.Counter(), collections.Counter(), {}
    for (_, municipality), n in in_municipality.items():
        of_municipality[municipality] += n
    for (place, municipality), n in sorted(in_municipality.items(), key=lambda kv: (-kv[1], kv[0][1])):
        share[place] += people.get(municipality, 0) * n / of_municipality[municipality]
        home.setdefault(place, municipality)
    return {place: (municipality, max(1, round(share[place]))) for place, municipality in home.items()}


def number_range(number, lettered):
    """The key address_json counts a house number under: its range's index in NUMBER_RANGES, and whether a letter follows."""
    return next(i for i, (_, top) in enumerate(NUMBER_RANGES) if number <= top), lettered


def address_json(numbers, letter="{upper(1)}"):
    """A country's address, its street-number choice weighted by numbers, which counts house numbers by number_range,
    and drawing a number's letter with the template letter."""
    items = ",\n".join(
        f'    {{ "format": "{{int({low},{top})}}{letter if lettered else ""}", "weight": {numbers[(i, lettered)]} }}'
        for lettered in (False, True) for i, (low, top) in enumerate(NUMBER_RANGES) if numbers[(i, lettered)]
    )
    return (
        '{\n  "format": "{street} {street-number}\\n{postal-code} {locality}",\n  "street": "{.locality as l}{l.street.name}",\n'
        f'  "street-number": [\n{items}\n  ],\n'
        '  "postal-code": "{l.postal-code.code}",\n  "locality": "{l.name}"\n}\n'
    )


def write_tables(out, kept, postal_codes, streets, **columns):
    """Writes out's region, municipality, locality, postal-code and street TSVs: the rows fence kept, sorted by key,
    postal_codes' rows sorted by code, and streets' rows for each kept locality. Each table takes COLUMNS' columns,
    unless a keyword names the table, with "_" standing for "-"."""
    columns = {**COLUMNS, **columns}
    for table in ("region", "municipality", "locality"):
        tsv.write(out / f"{table}.tsv", columns[table], [row for _, row in sorted(kept[table].items())])
    tsv.write(out / "postal-code.tsv", columns["postal_code"], sorted(postal_codes, key=lambda r: r["code"]))
    tsv.write(out / "street.tsv", columns["street"], [s for locality in sorted(kept["locality"]) for s in streets[locality]])
