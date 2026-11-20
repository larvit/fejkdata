"""Rebuild data/geo/LT/*.tsv and address.json from the Centre of Registers' Address Register.

    data-import/geo-lt.py [--cache DIR] [--min-streets N] [--street-share F] [--out DIR]

The register's open data lists each apskritis, savivaldybė, gyvenamoji vietovė and street with its code and the code of
the row it lies in, each address with its gyvenamoji vietovė, its street, its number and its postal code, and each of
the premises given an address, such as an apartment, with the address it lies at. An apskritis is a region, a
savivaldybė a municipality and a gyvenamoji vietovė a locality, each under its code in the register. The register names
an apskritis and a savivaldybė in the genitive, `Vilniaus`, so each is named with its kind, `Vilniaus apskritis`. A
locality keeps its nominative name, `Vilnius`, and a street its genitive name and the short form of its kind,
`Laisvės pr.`, as an address writes them.

Each table weighs a row by its addresses: an address is one, and each of the premises at it one more. An address with
no postal code is not counted. An address with no street is a number in its village, so a locality with no address on a
street is dropped. The files hold no point, so a locality has none. address.json weighs each length of house number,
with a letter or without, by the addresses that carry one.
"""
import collections
import csv
import io
import re
from collections.abc import Iterable, Iterator, Mapping
from pathlib import Path

import geo
import source

FILE = "https://www.registrucentras.lt/aduomenys/?byla={}.csv"
TIMEZONE = "Europe/Vilnius"
NUMBER = re.compile(r"^([1-9]\d{0,3})([A-ZĄČĘĖĮŠŲŪŽ]?)$")


class Register:
    """The register's places, each kind by its code, and each street's name by the street's code."""

    def __init__(self, apskritys: Iterable[Mapping[str, str]], savivaldybes: Iterable[Mapping[str, str]], vietoves: Iterable[Mapping[str, str]], gatves: Iterable[Mapping[str, str]]) -> None:
        self.regions: dict[str, geo.Row] = {r["ADM_KODAS"]: {"code": r["ADM_KODAS"], "name": f"{r['VARDAS_K']} {r['TIPAS']}", "timezone": TIMEZONE} for r in apskritys}
        self.municipalities: dict[str, geo.Row] = {r["SAV_KODAS"]: {"code": r["SAV_KODAS"], "name": f"{r['VARDAS_K']} {r['TIPAS']}", "region": r["ADM_KODAS"]} for r in savivaldybes}
        self.localities: dict[str, geo.Row] = {r["GYV_KODAS"]: {"code": r["GYV_KODAS"], "name": r["VARDAS"], "municipality": r["SAV_KODAS"]} for r in vietoves}
        self.streets = {r["GAT_KODAS"]: f"{r['VARDAS_K']} {r['TIPO_SANTRUMPA']}" for r in gatves}


def house_number(number: str) -> tuple[int, bool] | None:
    """A house number and whether a letter follows it, or None where it is neither."""
    m = NUMBER.match(number)
    return (int(m[1]), bool(m[2])) if m else None


class Tally:
    """What the addresses count: per locality, its addresses, its postal codes and its streets; and over them all, the
    house numbers by geo.number_range."""

    def __init__(self) -> None:
        self.addresses: collections.Counter[str] = collections.Counter()
        self.postal_codes: collections.Counter[tuple[str, str]] = collections.Counter()
        self.streets: collections.Counter[tuple[str, str]] = collections.Counter()
        self.numbers: collections.Counter[tuple[int, bool]] = collections.Counter()

    def add(self, row: Mapping[str, str], register: Register, n: int) -> None:
        """Counts an address as n, where it has a postal code and the register holds its locality."""
        locality = row["GYV_KODAS"]
        if not row["PASTO_KODAS"] or locality not in register.localities:
            return
        self.addresses[locality] += n
        self.postal_codes[(locality, row["PASTO_KODAS"])] += n
        number = house_number(row["NR"])
        if number and row["GAT_KODAS"] in register.streets:
            self.streets[(locality, register.streets[row["GAT_KODAS"]])] += n
            self.numbers[geo.number_range(*number)] += 1


def premises(rows: Iterable[Mapping[str, str]]) -> collections.Counter[str]:
    """The premises at each address, by the address's code."""
    return collections.Counter(r["AOB_KODAS"] for r in rows)


def tally(rows: Iterable[Mapping[str, str]], register: Register, premises: Mapping[str, int]) -> Tally:
    """The tally of the addresses' rows: an address counts as one, and one more for each of the premises at it."""
    counted = Tally()
    for row in rows:
        counted.add(row, register, 1 + premises.get(row["AOB_KODAS"], 0))
    return counted


def weighed(register: Register, counted: Tally) -> None:
    """Gives each locality with an address its addresses, each municipality those of its localities, and each region
    those of its municipalities."""
    for row in (*register.regions.values(), *register.municipalities.values()):
        row["addresses"] = 0
    for code, n in counted.addresses.items():
        locality = register.localities[code]
        locality["addresses"] = n
        register.municipalities[locality["municipality"]]["addresses"] += n
    for municipality in register.municipalities.values():
        register.regions[municipality["region"]]["addresses"] += municipality["addresses"]


def table(cache: Path, name: str, first_column: str) -> Iterator[dict[str, str]]:
    """The rows of one of the register's files, each column under its name."""
    text = source.fetch(FILE.format(name), cache, f"{name}.csv", magic=first_column.encode()).decode("utf-8")
    return csv.DictReader(io.StringIO(text, newline=""), delimiter="|")


def main() -> None:
    a = geo.parser(__doc__, "LT", 1.0).parse_args()
    cache, out = geo.directories(a)
    register = Register(table(cache, "adr_apskritys", "ADM_KODAS"), table(cache, "adr_savivaldybes", "SAV_KODAS"), table(cache, "adr_gyvenamosios_vietoves", "GYV_KODAS"), table(cache, "adr_gatves", "GAT_KODAS"))
    counted = tally(table(cache, "adr_stat_lr", "SAV_KODAS"), register, premises(table(cache, "adr_pat_lr", "SAV_KODAS")))
    weighed(register, counted)
    streets = geo.top_streets(counted.streets, a.street_share, a.min_streets, "addresses")
    geo.write_by_addresses(out, register.regions, register.municipalities, {c: row for c, row in register.localities.items() if "addresses" in row}, counted.postal_codes, streets)
    (out / "address.json").write_text(geo.address_json(counted.numbers), encoding="utf-8")


if __name__ == "__main__":
    main()
