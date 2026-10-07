"""The steps every country's geo import takes."""
import argparse
import collections
import math
import sys
from pathlib import Path

DATA = Path(__file__).resolve().parent.parent / "data" / "geo"
CACHE = Path(__file__).resolve().parent / "cache"


def parser(doc, country):
    """The options every country's import takes; a script adds its own before parse_args."""
    p = argparse.ArgumentParser(description=doc.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--out", default=str(DATA / country))
    p.add_argument("--streets-per-locality", type=int, default=10)
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


def top_streets(count, per_locality, column):
    """Per locality, the per_locality street names counted most in count, keyed (locality, name), ties by name; each
    row holds its count in column."""
    of = collections.defaultdict(list)
    for (locality, name), n in count.items():
        of[locality].append((n, name))
    return {locality: [{"name": name, "locality": locality, column: n} for n, name in sorted(named, key=lambda s: (-s[0], s[1]))[:per_locality]] for locality, named in of.items()}


def with_streets(places, named):
    """The places named holds streets for; each other one is logged as dropped."""
    for key, place in places.items():
        if key not in named:
            label = key if key == place["name"] else f"{key} {place['name']}"
            print(f"{label}: no streets, dropped", file=sys.stderr)
    return {key: place for key, place in places.items() if key in named}
