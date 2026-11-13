import collections
import importlib.util
import tempfile
import types
import unittest
from pathlib import Path

import geo


def script(name: str) -> types.ModuleType:
    """The import script of that file name, which is no module name, as a module."""
    spec = importlib.util.spec_from_file_location(name.replace("-", "_"), Path(__file__).with_name(name + ".py"))
    assert spec and spec.loader
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


geo_dk = script("geo-dk")
geo_fi = script("geo-fi")


def baf(code: str = "", street: str = "", swedish: str = "", low: str = "", low_letter: str = "", high: str = "", high_letter: str = "", kunta: str = "091") -> str:
    """A line of Posti's BAF, each field at the position Posti's record description gives."""
    return " " * 13 + f"{code:5}" + " " * 84 + f"{street:30}{swedish:30}" + " " * 25 + f"{low:>5}{low_letter:1}" + " " * 7 + f"{high:>5}{high_letter:1}" + " " * 7 + kunta


def pcf(code: str, town: str, kind: str) -> str:
    """A line of Posti's PCF, each field at the position Posti's record description gives."""
    return " " * 13 + f"{code:5}{town:30}" + " " * 62 + kind


class Homes(unittest.TestCase):
    def test_a_locality_lives_where_most_of_its_addresses_are_and_shares_each_municipalitys_people(self) -> None:
        in_municipality = {("Alby", "0101"): 30, ("Alby", "0147"): 10, ("Borg", "0147"): 10}
        self.assertEqual(geo.homes(in_municipality, {"0101": 300, "0147": 200}), {"Alby": ("0101", 400), "Borg": ("0147", 100)})

    def test_the_lower_code_wins_a_tie(self) -> None:
        self.assertEqual(geo.homes({("Alby", "0200"): 5, ("Alby", "0100"): 5}, {"0100": 10, "0200": 10}), {"Alby": ("0100", 20)})

    def test_a_locality_holds_at_least_one_person(self) -> None:
        self.assertEqual(geo.homes({("Alby", "0100"): 1, ("Borg", "0100"): 999}, {"0100": 100})["Alby"], ("0100", 1))


class HomeMunicipalities(unittest.TestCase):
    def test_a_locality_lives_where_most_of_its_addresses_are(self) -> None:
        self.assertEqual(geo.home_municipalities({("Alby", "0101"): 30, ("Alby", "0147"): 10, ("Borg", "0147"): 10}), {"Alby": "0101", "Borg": "0147"})

    def test_the_lower_code_wins_a_tie(self) -> None:
        self.assertEqual(geo.home_municipalities({("Alby", "0200"): 5, ("Alby", "0100"): 5}), {"Alby": "0100"})


class AddressJson(unittest.TestCase):
    def test_a_lettered_number_draws_its_letter_with_the_template_given(self) -> None:
        self.assertIn('{ "format": "{int(1,9)}{lower(1)}", "weight": 3 }', geo.address_json(collections.Counter({(0, True): 3}), "{lower(1)}"))

    def test_a_lettered_number_draws_an_upper_case_letter_where_no_template_is_given(self) -> None:
        self.assertIn('{ "format": "{int(1,9)}{upper(1)}", "weight": 3 }', geo.address_json(collections.Counter({(0, True): 3})))


class NumberRange(unittest.TestCase):
    def test_a_number_counts_under_the_range_it_falls_in(self) -> None:
        for number, lettered, want in [(9, False, (0, False)), (10, True, (1, True)), (999, False, (2, False)), (1000, True, (3, True))]:
            with self.subTest(number=number):
                self.assertEqual(geo.number_range(number, lettered), want)


class WriteTables(unittest.TestCase):
    def test_each_table_is_sorted_and_a_keyword_replaces_a_tables_columns(self) -> None:
        kept = {
            "region": {"2": {"code": "2", "name": "Syd", "population": 5, "timezone": "Europe/Copenhagen"}, "1": {"code": "1", "name": "Nord", "population": 7, "timezone": "Europe/Copenhagen"}},
            "municipality": {"0101": {"code": "0101", "name": "By", "region": "1", "population": 12}},
            "locality": {"Borg": {"name": "Borg", "municipality": "0101", "population": 4, "lat": "1", "lon": "2"}, "Alby": {"name": "Alby", "municipality": "0101", "population": 8, "lat": "3", "lon": "4"}},
        }
        postal_codes = [{"code": "2000", "locality": "Borg"}, {"code": "1000", "locality": "Alby"}]
        streets = {"Borg": [{"name": "Torvet", "locality": "Borg", "addresses": 3}], "Alby": [{"name": "Gade", "locality": "Alby", "addresses": 9}]}
        with tempfile.TemporaryDirectory() as d:
            out = Path(d)
            geo.write_tables(out, kept, postal_codes, streets, postal_code=["code", "locality"])
            tables = {t: (out / f"{t}.tsv").read_text(encoding="utf-8") for t in ("region", "municipality", "locality", "postal-code", "street")}
        self.assertEqual(tables, {
            "region": "code\tname\tpopulation\ttimezone\n1\tNord\t7\tEurope/Copenhagen\n2\tSyd\t5\tEurope/Copenhagen\n",
            "municipality": "code\tname\tregion\tpopulation\n0101\tBy\t1\t12\n",
            "locality": "name\tmunicipality\tpopulation\tlat\tlon\nAlby\t0101\t8\t3\t4\nBorg\t0101\t4\t1\t2\n",
            "postal-code": "code\tlocality\n1000\tAlby\n2000\tBorg\n",
            "street": "name\tlocality\taddresses\nGade\tAlby\t9\nTorvet\tBorg\t3\n",
        })


class Parsed(unittest.TestCase):
    def test_an_address_is_its_postnummer_postdistrikt_kommune_street_number_and_letter(self) -> None:
        self.assertEqual(geo_dk.parsed("Dalsgårdsvej 7B, Dalsgård, 8963 Auning", "7B", "0706-0287"), ("8963", "Auning", "0706", "Dalsgårdsvej", 7, "B"))

    def test_a_label_that_does_not_open_with_a_street_and_its_number_is_none(self) -> None:
        for label, number in [("74, 8660 Skanderborg", "74"), ("Vestergade 12, 8000 Aarhus C", "14"), ("Vestergade 12-14, 8000 Aarhus C", "12-14"), ("Vestergade 12", "12")]:
            with self.subTest(label=label):
                self.assertIsNone(geo_dk.parsed(label, number, "0751-0001"))


class Recased(unittest.TestCase):
    def test_a_postitoimipaikka_is_written_as_a_place_name(self) -> None:
        for posti, want in [("HELSINKI", "Helsinki"), ("HEINOLA KK", "Heinola kk"), ("ALAVUS AS", "Alavus as"), ("KOSKI TL", "Koski Tl"), ("INHAN TEHTAAT", "Inhan tehtaat"), ("ALA-HONKAJOKI", "Ala-Honkajoki"), ("KUOPPA-AHO", "Kuoppa-aho"), ("YLI-II", "Yli-Ii"), ("SAARIN-NIEMI", "Saarin-Niemi")]:
            with self.subTest(posti=posti):
                self.assertEqual(geo_fi.recased(posti, {"II"}), want)


class HouseNumbers(unittest.TestCase):
    def baf(self, low: str, low_letter: str, high: str, high_letter: str) -> str:
        return baf(low=low, low_letter=low_letter, high=high, high_letter=high_letter)

    def test_a_range_counts_its_two_ends_with_their_letters(self) -> None:
        self.assertEqual(sorted(geo_fi.house_numbers(self.baf("2", "a", "40", ""))), [(2, "a"), (40, "")])

    def test_a_range_of_one_building_counts_it_once(self) -> None:
        self.assertEqual(geo_fi.house_numbers(self.baf("7", "", "7", "")), [(7, "")])

    def test_a_street_without_numbers_counts_none(self) -> None:
        self.assertEqual(geo_fi.house_numbers(self.baf("", "", "", "")), [])


class StreetName(unittest.TestCase):
    def test_a_highway_a_postinumero_and_an_island_are_no_street(self) -> None:
        for name in ["Valtatie 23", "41520", "Mustalampi, Lammassaari", "Haapasaari (Vuosaari)", "Makkarasaari/Lapinsaari", ""]:
            with self.subTest(name=name):
                self.assertFalse(geo_fi.street_name(name))

    def test_a_street_may_hold_a_dot_a_colon_and_an_apostrophe(self) -> None:
        for name in ["J.A. Palménin tie", "VPK:n kuja", "Uudenmaantie"]:
            with self.subTest(name=name):
                self.assertTrue(geo_fi.street_name(name))


class Capitalised(unittest.TestCase):
    def test_a_name_opens_in_upper_case_and_keeps_the_rest_as_posti_writes_it(self) -> None:
        for posti, want in [("kuivalahdentie", "Kuivalahdentie"), ("asema-aukio", "Asema-aukio"), ("grevinnans väg", "Grevinnans väg"), ("äkäslompolontie", "Äkäslompolontie"), ("VPK:n kuja", "VPK:n kuja"), ("", "")]:
            with self.subTest(posti=posti):
                self.assertEqual(geo_fi.capitalised(posti), want)


class NormalCodes(unittest.TestCase):
    def test_only_a_normal_postinumero_is_kept_under_its_postitoimipaikka_as_a_place_name(self) -> None:
        lines = [pcf("04400", "JÄRVENPÄÄ", "1"), pcf("04401", "JÄRVENPÄÄ", "2"), pcf("91101", "YLI-II", "1")]
        self.assertEqual(geo_fi.normal_codes(lines, {"139": {"name": "Ii", "region": "17"}}), {"04400": "Järvenpää", "91101": "Yli-Ii"})


class Mainland(unittest.TestCase):
    def test_aland_and_its_kunnat_are_left_out(self) -> None:
        regions, municipalities = geo_fi.mainland({"01": "Uusimaa", "21": "Ahvenanmaa"}, {"186": {"name": "Järvenpää", "region": "01"}, "478": {"name": "Maarianhamina", "region": "21"}}, {"MK01": 1700, "MK21": 30, "KU186": 46, "KU478": 11})
        self.assertEqual(regions, {"01": {"code": "01", "name": "Uusimaa", "population": 1700, "timezone": "Europe/Helsinki"}})
        self.assertEqual(municipalities, {"186": {"code": "186", "name": "Järvenpää", "region": "01", "population": 46}})


class Tally(unittest.TestCase):
    CODES = {"04400": "Järvenpää", "99800": "Inari"}

    def test_a_range_without_a_swedish_name_counts_none(self) -> None:
        lines = [baf("04400", "Laurilantie"), baf("04400", "Laurilantie"), baf("04400", "Laurilantie", "Laurilavägen")]
        streets, swedish, *_ = geo_fi.tally(lines, self.CODES)
        self.assertEqual(streets, {("Järvenpää", "Laurilantie"): 3})
        self.assertEqual(swedish[("Järvenpää", "Laurilantie")], {"Laurilavägen": 1})

    def test_a_highway_an_island_and_a_range_under_no_normal_postinumero_count_nowhere(self) -> None:
        lines = [baf("04400", "Valtatie 23"), baf("04400", "Makkarasaari/Lapinsaari"), baf("04401", "Laurilantie", "Laurilavägen", "1", "", "9", "")]
        self.assertEqual(geo_fi.tally(lines, self.CODES), ({}, {}, {}, {}, {}))

    def test_a_street_opens_in_upper_case_in_finnish_and_in_swedish(self) -> None:
        lines = [baf("99800", "kuivalahdentie"), baf("99800", "Kuivalahdentie"), baf("04400", "Kreivittären tie", "grevinnans väg")]
        streets, swedish, *_ = geo_fi.tally(lines, self.CODES)
        self.assertEqual(streets, {("Inari", "Kuivalahdentie"): 2, ("Järvenpää", "Kreivittären tie"): 1})
        self.assertEqual(swedish[("Järvenpää", "Kreivittären tie")], {"Grevinnans väg": 1})

    def test_a_range_counts_its_kunta_its_postinumero_and_the_house_numbers_at_its_ends(self) -> None:
        lines = [baf("04400", "Laurilantie", "", "2", "a", "40", "", "186"), baf("04400", "Sibeliuksenkatu", "", "12000", "", "", "", "186")]
        _, _, in_municipality, postal_codes, numbers = geo_fi.tally(lines, self.CODES)
        self.assertEqual(in_municipality, {("Järvenpää", "186"): 2})
        self.assertEqual(postal_codes, {"Järvenpää": {"04400"}})
        self.assertEqual(numbers, {(0, True): 1, (1, False): 1})


class SwedishName(unittest.TestCase):
    def test_it_is_the_swedish_name_posti_gives_most_often(self) -> None:
        self.assertEqual(geo_fi.swedish_name(collections.Counter({"Storängsväg": 1, "Storängsvägen": 2}), "Isonniityntie"), "Storängsvägen")

    def test_the_first_in_the_alphabet_wins_a_tie(self) -> None:
        self.assertEqual(geo_fi.swedish_name(collections.Counter({"Storängsvägen": 1, "Storängsväg": 1}), "Isonniityntie"), "Storängsväg")

    def test_it_is_the_finnish_name_where_no_range_gives_a_swedish_one(self) -> None:
        self.assertEqual(geo_fi.swedish_name(collections.Counter(), "Laurilantie"), "Laurilantie")


class Localities(unittest.TestCase):
    def test_a_locality_holds_its_postinumerot_with_residents_and_sums_them(self) -> None:
        points = {"00100": {"lat": 60.0, "lon": 25.0}, "00101": {"lat": 62.0, "lon": 27.0}, "00102": {"lat": 61.0, "lon": 26.0}}
        codes_of = {"Helsinki": {"00100", "00101", "00102"}, "Tyhjälä": {"99990"}, "Pisteetön": {"99991"}}
        places, postal_codes = geo_fi.localities(codes_of, {"00100": 10, "00102": 5, "99991": 7}, {"Helsinki": "091", "Tyhjälä": "148", "Pisteetön": "148"}, points)
        self.assertEqual(places, {"Helsinki": {"name": "Helsinki", "municipality": "091", "population": 15, "lat": "60.5000", "lon": "25.5000"}})
        self.assertEqual(postal_codes, {"Helsinki": {"00100": 10, "00102": 5}})


if __name__ == "__main__":
    unittest.main()
