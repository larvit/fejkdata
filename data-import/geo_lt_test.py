import collections
import unittest
from typing import Any

from geo_test import script

geo_lt = script("geo-lt")

APSKRITYS = [
    {"ADM_KODAS": "9", "TIPAS": "apskritis", "TIPO_SANTRUMPA": "apskr.", "VARDAS_K": "Utenos"},
    {"ADM_KODAS": "10", "TIPAS": "apskritis", "TIPO_SANTRUMPA": "apskr.", "VARDAS_K": "Vilniaus"},
]
SAVIVALDYBES = [
    {"SAV_KODAS": "13", "TIPAS": "savivaldybė", "TIPO_SANTRUMPA": "sav.", "VARDAS_K": "Vilniaus miesto", "ADM_KODAS": "10"},
    {"SAV_KODAS": "62", "TIPAS": "savivaldybė", "TIPO_SANTRUMPA": "sav.", "VARDAS_K": "Molėtų rajono", "ADM_KODAS": "9"},
]
VIETOVES = [
    {"GYV_KODAS": "31003", "TIPAS": "miestas", "TIPO_SANTRUMPA": "m.", "VARDAS_K": "Vilniaus", "VARDAS": "Vilnius", "SAV_KODAS": "13"},
    {"GYV_KODAS": "14415", "TIPAS": "kaimas", "TIPO_SANTRUMPA": "k.", "VARDAS_K": "Galinių", "VARDAS": "Galiniai", "SAV_KODAS": "62"},
]
GATVES = [
    {"GAT_KODAS": "1302576", "TIPAS": "prospektas", "TIPO_SANTRUMPA": "pr.", "VARDAS_K": "Laisvės", "GYV_KODAS": "31003"},
    {"GAT_KODAS": "1307783", "TIPAS": "gatvė", "TIPO_SANTRUMPA": "g.", "VARDAS_K": "Plento", "GYV_KODAS": "31003"},
]


def address(number: str, street: str = "1302576", locality: str = "31003", postal_code: str = "LT-04339", block: str = "", code: str = "156841097") -> dict[str, str]:
    """A row of the file of addresses, with the columns the import reads."""
    return {"AOB_KODAS": code, "GYV_KODAS": locality, "GAT_KODAS": street, "NR": number, "KORPUSO_NR": block, "PASTO_KODAS": postal_code}


def register() -> Any:
    """The places and streets of the fixtures above."""
    return geo_lt.Register(APSKRITYS, SAVIVALDYBES, VIETOVES, GATVES)


def tally(*rows: dict[str, str], premises: dict[str, int] | None = None) -> Any:
    """The tally of the address rows, placed by the fixtures' register, each address holding the premises given for its code."""
    return geo_lt.tally(rows, register(), collections.Counter(premises or {}))


class RegisterOfPlaces(unittest.TestCase):
    def test_an_apskritis_is_a_region_named_with_its_kind(self) -> None:
        self.assertEqual(register().regions["10"], {"code": "10", "name": "Vilniaus apskritis", "timezone": "Europe/Vilnius"})

    def test_a_savivaldybe_is_a_municipality_named_with_its_kind(self) -> None:
        r = register()
        self.assertEqual(r.municipalities["13"], {"code": "13", "name": "Vilniaus miesto savivaldybė", "region": "10"})
        self.assertEqual(r.municipalities["62"], {"code": "62", "name": "Molėtų rajono savivaldybė", "region": "9"})

    def test_a_gyvenamoji_vietove_is_a_locality_under_its_nominative_name(self) -> None:
        r = register()
        self.assertEqual(r.localities["31003"], {"code": "31003", "name": "Vilnius", "municipality": "13"})
        self.assertEqual(r.localities["14415"], {"code": "14415", "name": "Galiniai", "municipality": "62"})

    def test_a_street_is_its_name_and_the_short_form_of_its_kind(self) -> None:
        r = register()
        self.assertEqual(r.streets["1302576"], "Laisvės pr.")
        self.assertEqual(r.streets["1307783"], "Plento g.")


class HouseNumber(unittest.TestCase):
    def test_a_number(self) -> None:
        self.assertEqual(geo_lt.house_number("40"), (40, False))

    def test_a_letter_follows_the_number(self) -> None:
        self.assertEqual(geo_lt.house_number("17A"), (17, True))

    def test_anything_else_is_no_number(self) -> None:
        for name in ("", "0", "12345", "A"):
            self.assertIsNone(geo_lt.house_number(name), name)


class TallyOfAddresses(unittest.TestCase):
    def test_an_address_on_a_street_counts_for_the_street_its_locality_and_its_postal_code(self) -> None:
        counted = tally(address("40"), address("42A"), address("17", "1307783", postal_code="LT-04340"))
        self.assertEqual(counted.streets, {("31003", "Laisvės pr."): 2, ("31003", "Plento g."): 1})
        self.assertEqual(counted.addresses, {"31003": 3})
        self.assertEqual(counted.postal_codes, {("31003", "LT-04339"): 2, ("31003", "LT-04340"): 1})

    def test_an_address_with_no_street_counts_for_its_locality_and_for_no_street(self) -> None:
        counted = tally(address("5", "", "14415", "LT-33264"))
        self.assertEqual(counted.addresses, {"14415": 1})
        self.assertEqual(counted.postal_codes, {("14415", "LT-33264"): 1})
        self.assertEqual(counted.streets, {})
        self.assertEqual(counted.numbers, {})

    def test_an_address_with_no_postal_code_or_under_no_place_is_not_counted(self) -> None:
        counted = tally(address("40", postal_code=""), address("40", locality="99999"))
        self.assertEqual((counted.addresses, counted.postal_codes, counted.streets), ({}, {}, {}))

    def test_an_address_on_a_street_the_register_does_not_hold_counts_for_no_street(self) -> None:
        counted = tally(address("40", "9999999"))
        self.assertEqual(counted.addresses, {"31003": 1})
        self.assertEqual(counted.streets, {})

    def test_an_address_whose_number_is_no_house_number_counts_for_no_street(self) -> None:
        counted = tally(address("12-3"))
        self.assertEqual((counted.addresses, counted.postal_codes), ({"31003": 1}, {("31003", "LT-04339"): 1}))
        self.assertEqual((counted.streets, counted.numbers), ({}, {}))

    def test_house_numbers_count_by_their_length_and_letter(self) -> None:
        counted = tally(address("40"), address("4A"), address("7"), address("195", block="2"))
        self.assertEqual(counted.numbers, {(1, False): 1, (0, True): 1, (0, False): 1, (2, False): 1})

    def test_an_address_counts_once_and_once_more_for_each_of_its_premises(self) -> None:
        counted = tally(address("40", code="156841097"), address("5", "", "14415", "LT-33264", code="155218235"), premises={"156841097": 12})
        self.assertEqual(counted.addresses, {"31003": 13, "14415": 1})
        self.assertEqual(counted.postal_codes, {("31003", "LT-04339"): 13, ("14415", "LT-33264"): 1})
        self.assertEqual(counted.streets, {("31003", "Laisvės pr."): 13})
        self.assertEqual(counted.numbers, {(1, False): 1})


class Premises(unittest.TestCase):
    def test_an_address_holds_its_premises(self) -> None:
        rows = [{"AOB_KODAS": "156841097"}, {"AOB_KODAS": "156841097"}, {"AOB_KODAS": "157192713"}]
        self.assertEqual(geo_lt.premises(rows), {"156841097": 2, "157192713": 1})


def stat(*units: tuple[str, int | None, int | None]) -> dict[str, Any]:
    """The agency's answer, in SDMX-JSON: each unit's code, its residents and the index of its status. Men and women
    together are the sex at index 1."""
    return {
        "structure": {
            "dimensions": {"observation": [
                {"id": "Lytis", "values": [{"id": "1"}, {"id": "0"}]},
                {"id": "admin_vnt_M3010320", "values": [{"id": code} for code, _, _ in units]},
                {"id": "MATVNT", "values": [{"id": "asmenys"}]},
                {"id": "LAIKOTARPIS", "values": [{"id": "2026"}]},
            ]},
            "attributes": {"observation": [{"id": "OBS_DECIMALS", "values": [{"id": "0"}]}, {"id": "OBS_STATUS", "values": [{"id": "konfidencialus"}, {"id": "nebuvo"}]}]},
        },
        "dataSets": [{"observations": {f"{sex}:{i}:0:0": [n if sex else 1, 0, status] for i, (_, n, status) in enumerate(units) for sex in (0, 1)}}],
    }


class Residents(unittest.TestCase):
    def test_a_unit_holds_its_men_and_women_together_under_its_code(self) -> None:
        self.assertEqual(geo_lt.residents(stat(("13", 607404, None), ("31003", 605000, None))), {"13": 607404, "31003": 605000})

    def test_a_withheld_count_is_left_out_and_a_unit_with_no_residents_holds_none(self) -> None:
        self.assertEqual(geo_lt.residents(stat(("14415", None, 0), ("14416", None, 1))), {"14416": 0})


class Weighed(unittest.TestCase):
    def test_a_locality_and_a_municipality_take_their_residents_and_a_region_sums_its_municipalities(self) -> None:
        r = register()
        geo_lt.weighed(r, tally(address("40"), address("42"), address("5", "", "14415", "LT-33264")), {"31003": 600, "13": 610, "14415": 30, "62": 90})
        self.assertEqual([r.localities[c]["population"] for c in ("31003", "14415")], [600, 30])
        self.assertEqual([r.municipalities[c]["population"] for c in ("13", "62")], [610, 90])
        self.assertEqual([r.regions[c]["population"] for c in ("10", "9")], [610, 90])

    def test_a_region_sums_every_municipality_in_it(self) -> None:
        utena = {"SAV_KODAS": "82", "TIPAS": "savivaldybė", "TIPO_SANTRUMPA": "sav.", "VARDAS_K": "Utenos rajono", "ADM_KODAS": "9"}
        r = geo_lt.Register(APSKRITYS, [*SAVIVALDYBES, utena], VIETOVES, GATVES)
        geo_lt.weighed(r, geo_lt.tally([], r, collections.Counter()), {"13": 610, "62": 90, "82": 200})
        self.assertEqual([r.regions[c]["population"] for c in ("10", "9")], [610, 290])

    def test_a_locality_with_no_count_shares_what_its_municipalitys_counted_localities_leave(self) -> None:
        r = geo_lt.Register(APSKRITYS, SAVIVALDYBES, [*VIETOVES, {"GYV_KODAS": "14416", "VARDAS": "Gojus", "SAV_KODAS": "62"}], GATVES)
        counted = geo_lt.tally([address("5", "", "14415", "LT-33264")], r, collections.Counter())
        geo_lt.weighed(r, counted, {"13": 610, "62": 90, "14416": 50})
        self.assertEqual(r.localities["14415"]["population"], 40)

    def test_a_locality_with_no_address_gets_no_weight(self) -> None:
        r = register()
        geo_lt.weighed(r, tally(address("40")), {"31003": 600, "13": 610, "62": 90})
        self.assertNotIn("population", r.localities["14415"])
        self.assertEqual(r.municipalities["62"]["population"], 90)


if __name__ == "__main__":
    unittest.main()
