"""Rebuild data/geo/LV/*.tsv and address.json from the State Land Service's State Address Register and the Central Statistical
Bureau's population.

    data-import/geo-lv.py [--cache DIR] [--min-streets N] [--street-share F] [--out DIR]

The register's open data lists each novads, pilsēta, pagasts, ciems and iela with its code in the register, the code and
kind of the row it lies in, and its status. It lists each building's address with the row it lies in, its postal code
and a point, and each apartment with its building. Only a row in force, status EKS, is read. The register holds no level
above the novads, so a novads or a state city is a region. A pagasts or a pilsēta is a municipality, and a pilsēta or a
ciems a locality: a pilsēta is both, and a state city, such as Rīga, all three. A region and a municipality keep their
codes in ATVK, the classification of Latvia's territories, which the register carries. A ciems has none, so a locality
keeps its code in the register.

A postal code and a street weigh their addresses in force: a building is one, and each apartment in it one more. A
region, a municipality and a locality weigh their people at the start of YEAR in the bureau's table RIG010: a
municipality by its own count, and a region by its municipalities'. A ciems the table doesn't count shares, by its
addresses, the people of its pagasts that the table counts under no ciems, with the houses outside every ciems. A locality's
point is the mean of its buildings' points. An address on an iela is a house number, and one straight under a pilsēta,
a ciems or a pagasts is a house's name, such as `Riņņi`, so a locality with no address on an iela is dropped.
address.json weighs each length of house number, with a letter or without, by the buildings that carry one.
"""
import collections
import csv
import io
import json
import re
from collections.abc import Iterable, Iterator, Mapping
from pathlib import Path

import geo
import source

RESOURCE = "https://data.gov.lv/dati/dataset/6b06a7e8-dedf-4705-a47b-2a7c51177473/resource/{}/download/{}.csv"
RESOURCES = {
    "aw_novads": "c62c60bb-58d4-4f26-82c0-5b630769f9d1",
    "aw_pilseta": "ee02baa4-2bc3-4f77-a6cb-5427a3e9befe",
    "aw_pagasts": "6ba8c905-27a1-443a-b9c6-256a0777425b",
    "aw_ciems": "0d3810f4-1ac0-4fba-8b10-0188084a361b",
    "aw_iela": "3c4ab802-76cf-433c-9c1c-89215e28d833",
    "aw_eka": "a510737a-18ce-400f-ad4b-04fce5228272",
    "aw_dziv": "b83be373-f444-4f50-9b98-28741845325e",
}
# The Central Statistical Bureau's people of each unit at the start of YEAR, table RIG010.
YEAR = 2026
PEOPLE = "https://data.stat.gov.lv/api/v1/lv/OSP_PUB/POP/IR/IRD/RIG010"
QUERY = {
    "query": [
        {"code": "AllAreaLV", "selection": {"filter": "all", "values": ["*"]}},
        {"code": "AgeGroup", "selection": {"filter": "item", "values": ["TOTAL"]}},
        {"code": "SEX", "selection": {"filter": "item", "values": ["T"]}},
        {"code": "ContentsCode", "selection": {"filter": "item", "values": ["RIG010"]}},
        {"code": "TIME", "selection": {"filter": "item", "values": [str(YEAR)]}},
    ],
    "response": {"format": "csv"},
}
# A unit's row in the table: an ATVK code of seven digits, or a ciems's code in the register, of nine.
UNIT = re.compile(r"^LV(\d{7}|\d{9}) ")
# Carnikava became a pilsēta on 22 August 2026: its ATVK code and its pagasts's since, and the codes the table holds.
CARNIKAVA, CARNIKAVAS_PAGASTS, CARNIKAVA_CIEMS, CARNIKAVAS_PAGASTS_BEFORE = "0023210", "0023411", "100184532", "0023410"
TIMEZONE = "Europe/Riga"
IN_FORCE = "EKS"
# The register's kinds of row, as TIPS_CD and VKUR_TIPS name them.
STATE, PILSETA, PAGASTS, CIEMS, IELA = "101", "104", "105", "106", "107"
LAYOUT = "{street} {street-number}\\n{locality}, {postal-code}"
# A house number and its letter, then what the import drops: a second number after a slash, a korpuss after ` k-`.
NUMBER = re.compile(r"^([1-9]\d{0,3})([A-ZĀČĒĢĪĶĻŅŠŪŽ]?)(?:/\S+)?(?: k-\S+)?$")


def in_force(rows: Iterable[Mapping[str, str]]) -> Iterator[Mapping[str, str]]:
    """The rows in force."""
    return (r for r in rows if r["STATUSS"] == IN_FORCE)


class Register:
    """The register's places in force: regions and municipalities by ATVK code, localities by their code in the register,
    and each street's locality and name by the street's code. municipality_of maps a pagasts's code in the register to its
    ATVK code."""

    def __init__(self, novadi: Iterable[Mapping[str, str]], pilsetas: Iterable[Mapping[str, str]], pagasti: Iterable[Mapping[str, str]], ciemi: Iterable[Mapping[str, str]], ielas: Iterable[Mapping[str, str]]) -> None:
        self.regions: dict[str, geo.Row] = {}
        self.municipalities: dict[str, geo.Row] = {}
        self.localities: dict[str, geo.Row] = {}
        self.municipality_of: dict[str, str] = {}
        region_of: dict[str, str] = {}
        for r in in_force(novadi):
            region_of[r["KODS"]] = self.region(r)
        for r in in_force(pilsetas):
            region = self.region(r) if r["VKUR_TIPS"] == STATE else region_of[r["VKUR_CD"]]
            self.municipalities[r["ATRIB"]] = {"code": r["ATRIB"], "name": r["NOSAUKUMS"], "region": region}
            self.localities[r["KODS"]] = {"code": r["KODS"], "name": r["NOSAUKUMS"], "municipality": r["ATRIB"]}
        for r in in_force(pagasti):
            self.municipalities[r["ATRIB"]] = {"code": r["ATRIB"], "name": r["NOSAUKUMS"], "region": region_of[r["VKUR_CD"]]}
            self.municipality_of[r["KODS"]] = r["ATRIB"]
        for r in in_force(ciemi):
            if r["VKUR_CD"] in self.municipality_of:
                self.localities[r["KODS"]] = {"code": r["KODS"], "name": r["NOSAUKUMS"], "municipality": self.municipality_of[r["VKUR_CD"]]}
        self.streets = {r["KODS"]: (r["VKUR_CD"], r["NOSAUKUMS"]) for r in in_force(ielas) if r["VKUR_CD"] in self.localities}

    def region(self, row: Mapping[str, str]) -> str:
        """Adds a novads or a state city as a region, and returns its ATVK code."""
        self.regions[row["ATRIB"]] = {"code": row["ATRIB"], "name": row["NOSAUKUMS"], "timezone": TIMEZONE}
        return row["ATRIB"]


def house_number(name: str) -> tuple[int, bool] | None:
    """An address's house number and whether a letter follows it, or None where the address is a house's name."""
    m = NUMBER.match(name)
    return (int(m[1]), bool(m[2])) if m else None


class Tally(geo.Counts):
    """What geo.Counts counts, with the sum of each locality's points and the addresses outside every locality, by
    municipality."""

    def __init__(self) -> None:
        super().__init__()
        self.outside: collections.Counter[str] = collections.Counter()
        self.lat: dict[str, float] = collections.defaultdict(float)
        self.lon: dict[str, float] = collections.defaultdict(float)
        self.points: collections.Counter[str] = collections.Counter()

    def add(self, row: Mapping[str, str], register: Register, n: int) -> None:
        """Counts a building in force as n addresses, where the register holds the row it lies in."""
        parent, kind = row["VKUR_CD"], row["VKUR_TIPS"]
        if kind == PAGASTS and parent in register.municipality_of:
            self.outside[register.municipality_of[parent]] += n
        elif kind == IELA and parent in register.streets:
            locality, street = register.streets[parent]
            self.place(row, locality, n, street, house_number(row["NOSAUKUMS"]))
        elif kind in (PILSETA, CIEMS) and parent in register.localities:
            self.place(row, parent, n, None, None)

    def place(self, row: Mapping[str, str], locality: str, n: int, street: str | None, number: tuple[int, bool] | None) -> None:
        """Counts a building as n addresses by Counts.count, and adds its point to its locality's where it has one."""
        self.count(locality, row["ATRIB"], n, street, number)
        if row["DD_N"] and row["DD_E"]:
            self.points[locality] += 1
            self.lat[locality] += float(row["DD_N"])
            self.lon[locality] += float(row["DD_E"])

    def point(self, locality: str) -> tuple[str, str]:
        """The mean of a locality's points, as a latitude and a longitude of four decimals, each empty where no address has a point."""
        n = self.points[locality]
        if not n:
            return "", ""
        return f"{self.lat[locality] / n:.4f}", f"{self.lon[locality] / n:.4f}"


def apartments(rows: Iterable[Mapping[str, str]]) -> collections.Counter[str]:
    """The apartments in force in each building, by the building's code."""
    return collections.Counter(r["VKUR_CD"] for r in in_force(rows))


def tally(rows: Iterable[Mapping[str, str]], register: Register, apartments: Mapping[str, int]) -> Tally:
    """The tally of the building addresses' rows: a building counts as one address, and one more for each of its apartments."""
    counted = Tally()
    for row in in_force(rows):
        counted.add(row, register, 1 + apartments.get(row["KODS"], 0))
    return counted


def people(rows: Iterable[list[str]]) -> dict[str, int]:
    """Each unit's people in the bureau's table: a ciems under its code in the register, and a pilsēta, a pagasts and a
    novads under ATVK. A row opens with `LV`, the code and the name, and ends in the count. A row of another kind, and
    one whose count the bureau does not give, is left out. The table draws the borders of 1 January 2026, when
    Carnikava was a ciems of its pagasts, so the pilsēta takes that ciems's people and the pagasts the rest, and a
    table counting the ciems without that pagasts raises ValueError."""
    out: dict[str, int] = {}
    for row in rows:
        m = UNIT.match(row[0])
        if m and row[-1].isdecimal():
            out[m[1]] = int(row[-1])
    if CARNIKAVA_CIEMS in out:
        if CARNIKAVAS_PAGASTS_BEFORE not in out:
            raise ValueError(f"the table counts Carnikava, and gives no count for {CARNIKAVAS_PAGASTS_BEFORE}, the pagasts it was a ciems of")
        out[CARNIKAVA] = out[CARNIKAVA_CIEMS]
        out[CARNIKAVAS_PAGASTS] = out[CARNIKAVAS_PAGASTS_BEFORE] - out[CARNIKAVA_CIEMS]
    return out


def weighed(register: Register, counted: Tally, people: Mapping[str, int]) -> None:
    """Gives each locality with an address its population by geo.populations and its point, each municipality its
    people, and each region those of its municipalities. A pilsēta has no count under its code in the register, so it
    takes its municipality's, which is its own. A municipality the table gives no count for raises ValueError."""
    geo.every_counted(register.municipalities, people)
    municipality_of = {code: row["municipality"] for code, row in register.localities.items()}
    for code, n in geo.populations(counted.addresses, municipality_of, people, people, counted.outside).items():
        register.localities[code]["population"] = n
        register.localities[code]["lat"], register.localities[code]["lon"] = counted.point(code)
    geo.sum_people(register.regions, register.municipalities, people)


def table(cache: Path, name: str) -> Iterator[dict[str, str]]:
    """The rows of one of the register's files, each column under its name."""
    text = source.fetch(RESOURCE.format(RESOURCES[name], name), cache, f"{name}.csv", magic=b"\xef\xbb\xbf").decode("utf-8-sig")
    return csv.DictReader(io.StringIO(text, newline=""))


def main() -> None:
    a = geo.parser(__doc__, "LV", 1.0).parse_args()
    cache, out = geo.directories(a)
    register = Register(*(table(cache, name) for name in ("aw_novads", "aw_pilseta", "aw_pagasts", "aw_ciems", "aw_iela")))
    counted = tally(table(cache, "aw_eka"), register, apartments(table(cache, "aw_dziv")))
    answer = source.fetch(PEOPLE, cache, f"people-{YEAR}.csv", magic=b"\xef\xbb\xbf", data=json.dumps(QUERY).encode(), headers={"Content-Type": "application/json"})
    weighed(register, counted, people(csv.reader(io.StringIO(answer.decode("utf-8-sig"), newline=""))))
    geo.write_counted(out, a, register.regions, register.municipalities, register.localities, counted, layout=LAYOUT)


if __name__ == "__main__":
    main()
