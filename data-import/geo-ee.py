"""Rebuild data/geo/EE/*.tsv and address.json from the Land and Spatial Development Board's addresses and settlement units and
Statistics Estonia's census.

    data-import/geo-ee.py [--cache DIR] [--min-streets N] [--street-share F] [--out DIR]

The board's file of postal codes lists every address of its address data system, ADS, with its sihtnumber, its near
address, such as `Sarapuu tn 70`, its full address, such as `Harju maakond, Harku vald, Tabasalu alevik, Sarapuu tn 70`,
and a point in L-EST97. Its file of settlement units gives each asustusüksus, its vald or linn and its maakond their
EHAK codes. A maakond is a region, a vald or a linn that governs itself a municipality, and an asustusüksus a locality.
A linnaosa is a part of its linn, so its addresses are the linn's, and the linn is the locality: Tallinn is one place.
A linn whose full addresses name no asustusüksus, such as Keila linn, is a locality too, under its own EHAK code.

A postal code and a street weigh their addresses in the file. A region, a municipality and a locality weigh their
people in Statistics Estonia's table RL21003, the census of 31 December 2021: a locality by its own count, and a
municipality and a region by the counts of the settlements in them. A count the census withholds is WITHHELD, and a
settlement formed since the census takes its ESTIMATES guess. A locality's point is the mean of its addresses' points. A street
is a near address that ends in a house number after one of STREET_KINDS, and a farm's name, such as `Antsu`, is none, so
a locality with no street is dropped. address.json weighs each length of house number, with a letter or without, by the
addresses that carry one.
"""
import collections
import csv
import io
import json
import math
import re
import struct
import sys
import zipfile
from collections.abc import Iterable, Iterator, Mapping
from pathlib import Path
from typing import Any

import geo
import source

ADDRESSES = "https://geoportaal.maaruum.ee/docs/aadress/sihtnumbrid.zip"
UNITS = "https://s3.pilw.io/rp-kemit-kataster/EHAK/asustusyksus_shp.zip"
# Statistics Estonia's table RL21003: the people of each settlement in the census of 31 December 2021.
CENSUS = "https://andmed.stat.ee/api/v1/et/stat/rahvaloendus/rel2021/rahvastiku_paiknemine/elukoht-ja-soo-vanusjaotus/RL21003.px"
QUERY = {
    "query": [
        {"code": "Aasta", "selection": {"filter": "item", "values": ["2021"]}},
        {"code": "Sugu", "selection": {"filter": "item", "values": ["1"]}},
        {"code": "Vanus", "selection": {"filter": "item", "values": ["1"]}},
        {"code": "Elukoht", "selection": {"filter": "all", "values": ["*"]}},
    ],
    "response": {"format": "json-stat2"},
}
# The code that stands in a place's key in the census where a level has none.
NO_CODE = "0000"
# What a place's label in the census opens with: dots for its depth and, before a settlement, its EHAK code.
LABEL = re.compile(r"^\.+(?:\d{4} )?")
# The census publishes no count of 1 or 2, so a settlement whose count it withholds holds 2.
WITHHELD = 2
# Kiltsi küla was Kiltsi alevik in the census, under another code.
RECODED = {"3091": "3092"}
# The settlements formed since the census, which counts their people in the settlements they were cut from: Keskküla,
# Kohtla-Uueküla, Koplimetsa küla and Raadi alev. Each population is a language model's guess.
ESTIMATES = {"2975": 90, "3262": 60, "3440": 150, "6674": 3000}
TIMEZONE = "Europe/Tallinn"
LINNAOSA = "6"
# The words that end a street's name. The board's handbook of address data, section 27.14, abbreviates tänav, maantee
# and puiestee as tn, mnt and pst, and names tee, allee, põik, väljak and plats as the words written in full.
STREET_KINDS = ("tn", "mnt", "pst", "tee", "allee", "põik", "väljak", "plats")
# After any väikekoht and its comma: a street, a house number, its letter, then what the import drops: a second building after a slash, an apartment after a hyphen.
NEAR = re.compile(rf"^(?:.+, )?([^,]+ (?:{'|'.join(STREET_KINDS)})) (\d+)([a-z]?)(?:/[^-]+)?(?:-.+)?$")

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
    place's locality code. A linnaosa's locality is its linn. A linn that is its own municipality is also placed by its
    name alone, with no asustusüksus."""
    regions: dict[str, geo.Row] = {}
    municipalities: dict[str, geo.Row] = {}
    localities: dict[str, geo.Row] = {}
    code_of: dict[Place, str] = {}
    for r in rows:
        regions[r["MKOOD"]] = {"code": r["MKOOD"], "name": r["MNIMI"], "timezone": TIMEZONE}
        municipalities[r["OKOOD"]] = {"code": r["OKOOD"], "name": r["ONIMI"], "region": r["MKOOD"]}
        whole = r["TYYP"] == LINNAOSA or r["AKOOD"] == r["OKOOD"]
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


def place_of(full: str, near: str) -> Place:
    """The place a full address names: its parts before the near address."""
    return tuple(full.removesuffix(near).split(", ")[:-1])


class Tally(geo.Counts):
    """What geo.Counts counts, and per locality the sum of its points."""

    def __init__(self) -> None:
        super().__init__()
        self.east: dict[str, float] = collections.defaultdict(float)
        self.north: dict[str, float] = collections.defaultdict(float)
        self.unplaced: collections.Counter[Place] = collections.Counter()

    def add(self, row: Mapping[str, str], code_of: Mapping[Place, str]) -> None:
        """Counts one address, or notes its place as unplaced where code_of holds no locality for it."""
        place = place_of(row["TAISAADRESS"], row["LAHIAADRESS"])
        if place not in code_of:
            self.unplaced[place] += 1
            return
        locality = code_of[place]
        self.east[locality] += float(row["VIITEPUNKT_X"])
        self.north[locality] += float(row["VIITEPUNKT_Y"])
        street = street_of(row["LAHIAADRESS"])
        if street and not 0 < street[1] < 10000:
            street = None
        self.count(locality, row["SIHTNUMBER"], 1, street[0] if street else None, (street[1], street[2]) if street else None)

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


class Census:
    """The 2021 census's count of each settlement, by its EHAK code, and by its maakond and name where one settlement of
    the maakond bears the name. A place's key is the EHAK codes of its maakond, its municipality and its settlement, then
    two characters whose second is its kind, as the settlement units' TYYP. A linn that holds no settlement is one itself, under its municipality's code. A linnaosa, a vald and a
    maakond are none."""

    def __init__(self, stat: Mapping[str, Any]) -> None:
        category = stat["dimension"]["Elukoht"]["category"]
        places = [(key, LABEL.sub("", category["label"][key]), stat["value"][i]) for key, i in category["index"].items() if len(key) == 14 and key[4:8] != NO_CODE and key[13] != LINNAOSA]
        holders = {key[4:8] for key, _, _ in places if key[8:12] != NO_CODE}
        self.by_code: dict[str, int] = {}
        named: dict[tuple[str, str], list[int]] = collections.defaultdict(list)
        for key, name, n in places:
            if key[8:12] == NO_CODE and key[4:8] in holders:
                continue
            people = WITHHELD if n is None else n
            self.by_code[key[4:8] if key[8:12] == NO_CODE else key[8:12]] = people
            named[(key[:4], name)].append(people)
        self.by_name = {place: counts[0] for place, counts in named.items() if len(counts) == 1}

    def of(self, code: str, name: str, region: str) -> int | None:
        """A settlement's people: the census's count under its code, or else under its name in its maakond, since EHAK
        gave some settlements a new code after the census. None where the census counts it under neither."""
        code = RECODED.get(code, code)
        if code in self.by_code:
            return self.by_code[code]
        return self.by_name.get((region, name))


def weighed(regions: Mapping[str, geo.Row], municipalities: Mapping[str, geo.Row], localities: Mapping[str, geo.Row], counted: Tally, census: Census) -> None:
    """Gives each locality that has an address its point and its population: its census count, or else its estimate,
    and at least one person. Gives each municipality and region the census counts of its localities, estimates left
    out. A locality with neither count nor estimate is left out if it has no street, and raises ValueError if it has
    one."""
    streeted = {locality for locality, _ in counted.streets}
    for row in (*regions.values(), *municipalities.values()):
        row["population"] = 0
    for code, locality in localities.items():
        municipality = municipalities[locality["municipality"]]
        count = census.of(code, locality["name"], municipality["region"])
        people = ESTIMATES.get(code) if count is None else count
        if people is None and code in streeted:
            raise ValueError(f"{code} {locality['name']} has a street, and neither a census count nor an estimate")
        municipality["population"] += count or 0
        regions[municipality["region"]]["population"] += count or 0
        if people is not None and code in counted.addresses:
            locality["population"] = max(1, people)
            locality["lat"], locality["lon"] = counted.point(code)


def unzipped(url: str, cache: Path, member: str) -> bytes:
    """A member of the zip file at a URL."""
    name = url.rsplit("/", 1)[1]
    return zipfile.ZipFile(io.BytesIO(source.fetch(url, cache, name, magic=b"PK"))).read(member)


def main() -> None:
    a = geo.parser(__doc__, "EE", 1.0).parse_args()
    cache, out = geo.directories(a)
    regions, municipalities, localities, code_of = units(dbf_rows(unzipped(UNITS, cache, "asustusyksus.dbf")))
    counted = tally(unzipped(ADDRESSES, cache, "sihtnumbrid.csv").decode("utf-8").splitlines(), code_of)
    answer = source.fetch(CENSUS, cache, "rl21003.json", magic=b"{", data=json.dumps(QUERY).encode(), headers={"Content-Type": "application/json"})
    weighed(regions, municipalities, localities, counted, Census(json.loads(answer)))
    geo.write_counted(out, a, regions, municipalities, localities, counted, letter="{lower(1)}")


if __name__ == "__main__":
    main()
