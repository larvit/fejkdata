"""Rebuild data/geo/LT/*.tsv and address.json from the Centre of Registers' Address Register and the State Data
Agency's count of residents.

    data-import/geo-lt.py [--cache DIR] [--min-streets N] [--street-share F] [--out DIR]

The register's open data lists each apskritis, savivaldybė, gyvenamoji vietovė and street with its code and the code of
the row it lies in, each address with its gyvenamoji vietovė, its street, its number and its postal code, and each of
the premises given an address, such as an apartment, with the address it lies at. An apskritis is a region, a
savivaldybė a municipality and a gyvenamoji vietovė a locality, each under its code in the register. The register names
an apskritis and a savivaldybė in the genitive, `Vilniaus`, so each is named with its kind, `Vilniaus apskritis`. A
locality keeps its nominative name, `Vilnius`, and a street its genitive name and the short form of its kind,
`Laisvės pr.`, as an address writes them.

A postal code and a street weigh their addresses: an address is one, and each of the premises at it one more. An address
with no postal code is not counted. A municipality weighs its residents at the start of YEAR in the agency's table
M3010202, and a region its municipalities'. A locality weighs its share, by its addresses, of its municipality's
residents. An address with no street is a number in its village, so a locality with no address on a
street is dropped. The files hold no point, so a locality has none. address.json weighs each length of house number,
with a letter or without, by the addresses that carry one.
"""
import collections
import csv
import io
import json
import re
from collections.abc import Iterable, Iterator, Mapping
from pathlib import Path
from typing import Any

import geo
import source

FILE = "https://www.registrucentras.lt/aduomenys/?byla={}.csv"
# The State Data Agency's count of each municipality's residents at the start of YEAR, table M3010202 of indicator S3R167.
YEAR = 2026
RESIDENTS = f"https://osp-rs.stat.gov.lt/rest_json/data/S3R167_M3010202?startPeriod={YEAR}&endPeriod={YEAR}"
# The table's codes for men and women together and for every age together.
BOTH_SEXES, EVERY_AGE = "0", "g000g999"
TIMEZONE = "Europe/Vilnius"
NUMBER = re.compile(r"^([1-9]\d{0,3})([A-Z]?)$")


class Register:
    """The register's places, each kind by its code, and each street's name by the street's code."""

    def __init__(self, apskritys: Iterable[Mapping[str, str]], savivaldybes: Iterable[Mapping[str, str]], vietoves: Iterable[Mapping[str, str]], gatves: Iterable[Mapping[str, str]]) -> None:
        self.regions: dict[str, geo.Row] = {r["ADM_KODAS"]: {"code": r["ADM_KODAS"], "name": f"{r['VARDAS_K']} {r['TIPAS']}", "timezone": TIMEZONE} for r in apskritys}
        self.municipalities: dict[str, geo.Row] = {r["SAV_KODAS"]: {"code": r["SAV_KODAS"], "name": f"{r['VARDAS_K']} {r['TIPAS']}", "region": r["ADM_KODAS"]} for r in savivaldybes}
        self.localities: dict[str, geo.Row] = {r["GYV_KODAS"]: {"code": r["GYV_KODAS"], "name": r["VARDAS"], "municipality": r["SAV_KODAS"]} for r in vietoves}
        self.streets = {r["GAT_KODAS"]: f"{r['VARDAS_K']} {r['TIPO_SANTRUMPA']}" for r in gatves}


def house_number(number: str) -> tuple[int, bool] | None:
    """A house number and whether a letter follows it, or None where the text is no house number."""
    return geo.house_number(NUMBER, number)


class Tally(geo.Counts):
    def add(self, row: Mapping[str, str], register: Register, n: int) -> None:
        """Counts an address as n, where it has a postal code and the register holds its locality."""
        locality = row["GYV_KODAS"]
        if not row["PASTO_KODAS"] or locality not in register.localities:
            return
        self.count(locality, row["PASTO_KODAS"], n, register.streets.get(row["GAT_KODAS"]), house_number(row["NR"]))


def premises(rows: Iterable[Mapping[str, str]]) -> collections.Counter[str]:
    """The premises at each address, by the address's code."""
    return collections.Counter(r["AOB_KODAS"] for r in rows)


def tally(rows: Iterable[Mapping[str, str]], register: Register, premises: Mapping[str, int]) -> Tally:
    """The tally of the addresses' rows: an address counts as one, and one more for each of the premises at it."""
    counted = Tally()
    for row in rows:
        counted.add(row, register, 1 + premises.get(row["AOB_KODAS"], 0))
    return counted


def residents(stat: Mapping[str, Any]) -> dict[str, int]:
    """Each territory's residents in the agency's answer, men and women of every age together, by its code, which for a
    savivaldybė is its code in the register. A territory the answer gives no count for is left out."""
    sexes, ages, territories = (stat["structure"]["dimensions"]["observation"][i]["values"] for i in (0, 1, 2))
    both = next(i for i, sex in enumerate(sexes) if sex["id"] == BOTH_SEXES)
    every = next(i for i, age in enumerate(ages) if age["id"] == EVERY_AGE)
    out: dict[str, int] = {}
    for i, territory in enumerate(territories):
        n = stat["dataSets"][0]["observations"].get(f"{both}:{every}:{i}:0:0", [None])[0]
        if n is not None:
            out[territory["id"]] = n
    return out


def weighed(register: Register, counted: Tally, residents: Mapping[str, int]) -> None:
    """Gives each municipality its residents, each region those of its municipalities, and each locality with an address
    its share, by addresses, of its municipality's. A municipality the table gives no count for raises ValueError."""
    geo.every_counted(register.municipalities, residents)
    municipality_of = {code: row["municipality"] for code, row in register.localities.items()}
    for code, n in geo.populations(counted.addresses, municipality_of, {}, residents).items():
        register.localities[code]["population"] = n
    geo.sum_people(register.regions, register.municipalities, residents)


def table(cache: Path, name: str, first_column: str) -> Iterator[dict[str, str]]:
    """The rows of one of the register's files, each column under its name."""
    text = source.fetch(FILE.format(name), cache, f"{name}.csv", magic=first_column.encode()).decode("utf-8")
    return csv.DictReader(io.StringIO(text, newline=""), delimiter="|")


def main() -> None:
    a = geo.parser(__doc__, "LT", 1.0).parse_args()
    cache, out = geo.directories(a)
    register = Register(table(cache, "adr_apskritys", "ADM_KODAS"), table(cache, "adr_savivaldybes", "SAV_KODAS"), table(cache, "adr_gyvenamosios_vietoves", "GYV_KODAS"), table(cache, "adr_gatves", "GAT_KODAS"))
    counted = tally(table(cache, "adr_stat_lr", "SAV_KODAS"), register, premises(table(cache, "adr_pat_lr", "SAV_KODAS")))
    weighed(register, counted, residents(json.loads(source.fetch(RESIDENTS, cache, f"municipality-residents-{YEAR}.json", magic=b"{"))))
    geo.write_counted(out, a, register.regions, register.municipalities, register.localities, counted)


if __name__ == "__main__":
    main()
