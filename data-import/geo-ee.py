"""Rebuild data/geo/EE/*.tsv and address.json from the Land and Spatial Development Board's addresses and settlement units.

    data-import/geo-ee.py [--cache DIR] [--min-streets N] [--street-share F] [--out DIR]

The board's file of postal codes lists every address of its address data system, ADS, with its sihtnumber, its near
address, such as `Sarapuu tn 70`, its full address, such as `Harju maakond, Harku vald, Tabasalu alevik, Sarapuu tn 70`,
and a point in L-EST97. Its file of settlement units gives each asustusüksus, its vald or linn and its maakond their
EHAK codes. A maakond is a region, a vald or a linn that governs itself a municipality, and an asustusüksus a locality.
A linnaosa is a part of its linn, so its addresses are the linn's, and the linn is the locality: Tallinn is one place.
A linn whose full addresses name no asustusüksus, such as Keila linn, is a locality too, under its own EHAK code.

Each table weighs a row by its addresses in the file. A locality's point is the mean of its addresses' points. A street
is a near address that ends in a house number after one of STREET_KINDS, and a farm's name, such as `Antsu`, is none, so
a locality with no street is dropped. address.json weighs each length of house number, with a letter or without, by the
addresses that carry one.
"""
import collections
import csv
import io
import math
import re
import struct
import sys
import zipfile
from collections.abc import Iterable, Iterator, Mapping
from pathlib import Path

import geo
import source

ADDRESSES = "https://geoportaal.maaruum.ee/docs/aadress/sihtnumbrid.zip"
UNITS = "https://s3.pilw.io/rp-kemit-kataster/EHAK/asustusyksus_shp.zip"
TIMEZONE = "Europe/Tallinn"
LINNAOSA = "6"
# The words that end a street's name. The board's handbook of address data, section 27.14, abbreviates tänav, maantee
# and puiestee as tn, mnt and pst, and names tee, allee, põik, väljak and plats as the words written in full.
STREET_KINDS = ("tn", "mnt", "pst", "tee", "allee", "põik", "väljak", "plats")
# A street, a house number, its letter, then what the import drops: a second building after a slash, an apartment after a hyphen.
NEAR = re.compile(rf"^(.+ (?:{'|'.join(STREET_KINDS)})) (\d+)([a-zõäöü]?)(?:/[^-]+)?(?:-.+)?$")

# A place as a full address names it: its maakond, its vald or linn and, where it has one, its asustusüksus.
type Place = tuple[str, ...]


def dbf_rows(data: bytes) -> Iterator[dict[str, str]]:
    """The records of a dBase file, each field under its name, as text in Windows-1257."""
    count, header, record = struct.unpack("<IHH", data[4:12])
    fields = [(data[at : at + 11].split(b"\0")[0].decode("ascii"), data[at + 16]) for at in range(32, header - 1, 32)]
    for start in range(header, header + count * record, record):
        at = start + 1
        row = {}
        for name, length in fields:
            row[name] = data[at : at + length].decode("cp1257").strip()
            at += length
        yield row


def units(rows: Iterable[Mapping[str, str]]) -> tuple[dict[str, geo.Row], dict[str, geo.Row], dict[str, geo.Row], dict[Place, str]]:
    """The regions, the municipalities and the localities of the settlement units' rows, each by its EHAK code, and each
    place's locality code. A linnaosa's locality is its linn, and so is that of a linn's own name with no asustusüksus."""
    regions: dict[str, geo.Row] = {}
    municipalities: dict[str, geo.Row] = {}
    localities: dict[str, geo.Row] = {}
    code_of: dict[Place, str] = {}
    for r in rows:
        regions[r["MKOOD"]] = {"code": r["MKOOD"], "name": r["MNIMI"], "timezone": TIMEZONE}
        municipalities[r["OKOOD"]] = {"code": r["OKOOD"], "name": r["ONIMI"], "region": r["MKOOD"]}
        whole = r["TYYP"] == LINNAOSA or r["ANIMI"] == r["ONIMI"]
        code, name = (r["OKOOD"], r["ONIMI"]) if whole else (r["AKOOD"], r["ANIMI"])
        localities[code] = {"code": code, "name": name, "municipality": r["OKOOD"]}
        code_of[(r["MNIMI"], r["ONIMI"], r["ANIMI"])] = code
        if whole:
            code_of[(r["MNIMI"], r["ONIMI"])] = code
    return regions, municipalities, localities, code_of


def lat_lon(east: float, north: float) -> tuple[float, float]:
    """The latitude and longitude of a point in L-EST97, EPSG:3301: a Lambert conformal conic projection of GRS 80,
    with standard parallels at 58° and 59°20′ N, its origin at 57°31′03.19415″ N 24° E, and that origin at 500,000 m
    east and 6,375,000 m north."""
    a, f = 6378137.0, 1 / 298.257222101
    e = math.sqrt(2 * f - f * f)

    def m(phi: float) -> float:
        return math.cos(phi) / math.sqrt(1 - (e * math.sin(phi)) ** 2)

    def t(phi: float) -> float:
        return float(math.tan(math.pi / 4 - phi / 2) / ((1 - e * math.sin(phi)) / (1 + e * math.sin(phi))) ** (e / 2))

    phi1, phi2, phi0 = math.radians(58), math.radians(59 + 1 / 3), math.radians(57.51755393055556)
    n = (math.log(m(phi1)) - math.log(m(phi2))) / (math.log(t(phi1)) - math.log(t(phi2)))
    big_f = m(phi1) / (n * t(phi1) ** n)
    rho0 = a * big_f * t(phi0) ** n
    x, y = east - 500000, rho0 - (north - 6375000)
    rho = math.hypot(x, y)
    t_point = (rho / (a * big_f)) ** (1 / n)
    phi = math.pi / 2 - 2 * math.atan(t_point)
    for _ in range(8):
        phi = math.pi / 2 - 2 * math.atan(t_point * ((1 - e * math.sin(phi)) / (1 + e * math.sin(phi))) ** (e / 2))
    return math.degrees(phi), 24 + math.degrees(math.atan2(x, y) / n)


def street_of(near: str) -> tuple[str, int, bool] | None:
    """A near address's street, house number and whether a letter follows it, or None where it names no street."""
    m = NEAR.match(near)
    return (m[1], int(m[2]), bool(m[3])) if m else None


def place_of(full: str) -> Place:
    """The place a full address names: its parts before the near address."""
    return tuple(full.split(", ")[:-1])


class Tally:
    """What the addresses count: per locality, its addresses, its postal codes, its streets and the sum of its points;
    and over them all, the house numbers by geo.number_range."""

    def __init__(self) -> None:
        self.addresses: collections.Counter[str] = collections.Counter()
        self.postal_codes: collections.Counter[tuple[str, str]] = collections.Counter()
        self.streets: collections.Counter[tuple[str, str]] = collections.Counter()
        self.east: dict[str, float] = collections.defaultdict(float)
        self.north: dict[str, float] = collections.defaultdict(float)
        self.numbers: collections.Counter[tuple[int, bool]] = collections.Counter()
        self.unplaced: collections.Counter[Place] = collections.Counter()

    def add(self, row: Mapping[str, str], code_of: Mapping[Place, str]) -> None:
        """Counts one address, or notes its place as unplaced where code_of holds no locality for it."""
        place = place_of(row["TAISAADRESS"])
        if place not in code_of:
            self.unplaced[place] += 1
            return
        locality = code_of[place]
        self.addresses[locality] += 1
        self.postal_codes[(locality, row["SIHTNUMBER"])] += 1
        self.east[locality] += float(row["VIITEPUNKT_X"])
        self.north[locality] += float(row["VIITEPUNKT_Y"])
        street = street_of(row["LAHIAADRESS"])
        if street and 0 < street[1] < 10000:
            self.streets[(locality, street[0])] += 1
            self.numbers[geo.number_range(street[1], street[2])] += 1

    def point(self, locality: str) -> tuple[str, str]:
        """The mean of a locality's points, as a latitude and a longitude of four decimals."""
        n = self.addresses[locality]
        lat, lon = lat_lon(self.east[locality] / n, self.north[locality] / n)
        return f"{lat:.4f}", f"{lon:.4f}"


def tally(lines: Iterable[str], code_of: Mapping[Place, str]) -> Tally:
    """The tally of the postal code file's lines, its header first."""
    counted = Tally()
    for row in csv.DictReader(lines, delimiter=";"):
        counted.add(row, code_of)
    for place, n in sorted(counted.unplaced.items()):
        print(f"{', '.join(place)}: no such settlement unit, {n} addresses dropped", file=sys.stderr)
    return counted


def weighed(regions: Mapping[str, geo.Row], municipalities: Mapping[str, geo.Row], localities: Mapping[str, geo.Row], counted: Tally) -> None:
    """Gives each locality with an address its addresses and its point, and each municipality and region the addresses
    of its localities."""
    for row in (*regions.values(), *municipalities.values()):
        row["addresses"] = 0
    for code, n in counted.addresses.items():
        locality = localities[code]
        municipality = municipalities[locality["municipality"]]
        locality["addresses"] = n
        locality["lat"], locality["lon"] = counted.point(code)
        municipality["addresses"] += n
        regions[municipality["region"]]["addresses"] += n


def unzipped(url: str, cache: Path, member: str) -> bytes:
    """A member of the zip file at a URL."""
    name = url.rsplit("/", 1)[1]
    return zipfile.ZipFile(io.BytesIO(source.fetch(url, cache, name, magic=b"PK"))).read(member)


def main() -> None:
    a = geo.parser(__doc__, "EE", 1.0).parse_args()
    cache, out = geo.directories(a)
    regions, municipalities, localities, code_of = units(dbf_rows(unzipped(UNITS, cache, "asustusyksus.dbf")))
    counted = tally(unzipped(ADDRESSES, cache, "sihtnumbrid.csv").decode("utf-8").splitlines(), code_of)
    weighed(regions, municipalities, localities, counted)
    streets = geo.top_streets(counted.streets, a.street_share, a.min_streets, "addresses")
    postal_codes: dict[str, list[geo.Row]] = collections.defaultdict(list)
    for (locality, code), n in counted.postal_codes.items():
        postal_codes[locality].append({"code": code, "locality": locality, "addresses": n})
    kept, dropped = geo.fence(regions, municipalities, {c: row for c, row in localities.items() if "addresses" in row}, streets, postal_codes)
    geo.log_dropped_localities(dropped["locality"])
    for row in dropped["municipality"].values():
        print(f"{row['code']} {row['name']}: no locality with a street, dropped", file=sys.stderr)
    rows = sorted((row for locality in kept["locality"] for row in postal_codes[locality]), key=lambda r: r["locality"])
    geo.write_tables(out, kept, rows, streets, region=["code", "name", "addresses", "timezone"], municipality=["code", "name", "region", "addresses"], locality=["code", "name", "municipality", "addresses", "lat", "lon"])
    (out / "address.json").write_text(geo.address_json(counted.numbers, "{lower(1)}"), encoding="utf-8")


if __name__ == "__main__":
    main()
