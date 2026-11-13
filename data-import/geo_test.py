import importlib.util
import tempfile
import unittest
from pathlib import Path

import geo

spec = importlib.util.spec_from_file_location("geo_dk", Path(__file__).with_name("geo-dk.py"))
geo_dk = importlib.util.module_from_spec(spec)
spec.loader.exec_module(geo_dk)
spec = importlib.util.spec_from_file_location("geo_fi", Path(__file__).with_name("geo-fi.py"))
geo_fi = importlib.util.module_from_spec(spec)
spec.loader.exec_module(geo_fi)


class Homes(unittest.TestCase):
    def test_a_locality_lives_where_most_of_its_addresses_are_and_shares_each_municipalitys_people(self):
        in_municipality = {("Alby", "0101"): 30, ("Alby", "0147"): 10, ("Borg", "0147"): 10}
        self.assertEqual(geo.homes(in_municipality, {"0101": 300, "0147": 200}), {"Alby": ("0101", 400), "Borg": ("0147", 100)})

    def test_the_lower_code_wins_a_tie(self):
        self.assertEqual(geo.homes({("Alby", "0200"): 5, ("Alby", "0100"): 5}, {"0100": 10, "0200": 10}), {"Alby": ("0100", 20)})

    def test_a_locality_holds_at_least_one_person(self):
        self.assertEqual(geo.homes({("Alby", "0100"): 1, ("Borg", "0100"): 999}, {"0100": 100})["Alby"], ("0100", 1))


class NumberRange(unittest.TestCase):
    def test_a_number_counts_under_the_range_it_falls_in(self):
        for number, lettered, want in [(9, False, (0, False)), (10, True, (1, True)), (999, False, (2, False)), (1000, True, (3, True))]:
            with self.subTest(number=number):
                self.assertEqual(geo.number_range(number, lettered), want)


class WriteTables(unittest.TestCase):
    def test_each_table_is_sorted_and_a_keyword_replaces_a_tables_columns(self):
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
    def test_an_address_is_its_postnummer_postdistrikt_kommune_street_number_and_letter(self):
        self.assertEqual(geo_dk.parsed("Dalsgårdsvej 7B, Dalsgård, 8963 Auning", "7B", "0706-0287"), ("8963", "Auning", "0706", "Dalsgårdsvej", 7, "B"))

    def test_a_label_that_does_not_open_with_a_street_and_its_number_is_none(self):
        for label, number in [("74, 8660 Skanderborg", "74"), ("Vestergade 12, 8000 Aarhus C", "14"), ("Vestergade 12-14, 8000 Aarhus C", "12-14"), ("Vestergade 12", "12")]:
            with self.subTest(label=label):
                self.assertIsNone(geo_dk.parsed(label, number, "0751-0001"))


if __name__ == "__main__":
    unittest.main()


class Recased(unittest.TestCase):
    def test_a_postitoimipaikka_is_written_as_a_place_name(self):
        for posti, want in [("HELSINKI", "Helsinki"), ("HEINOLA KK", "Heinola kk"), ("ALAVUS AS", "Alavus as"), ("KOSKI TL", "Koski Tl"), ("INHAN TEHTAAT", "Inhan tehtaat"), ("ALA-HONKAJOKI", "Ala-Honkajoki"), ("KUOPPA-AHO", "Kuoppa-aho"), ("YLI-II", "Yli-Ii")]:
            with self.subTest(posti=posti):
                self.assertEqual(geo_fi.recased(posti, {"II"}), want)


class HouseNumbers(unittest.TestCase):
    def baf(self, low, low_letter, high, high_letter):
        return " " * 187 + f"{low:>5}{low_letter:1}" + " " * 7 + f"{high:>5}{high_letter:1}" + " " * 7 + "091"

    def test_a_range_counts_its_two_ends_with_their_letters(self):
        self.assertEqual(sorted(geo_fi.house_numbers(self.baf("2", "a", "40", ""))), [(2, "a"), (40, "")])

    def test_a_range_of_one_building_counts_it_once(self):
        self.assertEqual(geo_fi.house_numbers(self.baf("7", "", "7", "")), [(7, "")])

    def test_a_street_without_numbers_counts_none(self):
        self.assertEqual(geo_fi.house_numbers(self.baf("", "", "", "")), [])


class StreetName(unittest.TestCase):
    def test_a_highway_a_postinumero_and_an_island_are_no_street(self):
        for name in ["Valtatie 23", "41520", "Mustalampi, Lammassaari", "Haapasaari (Vuosaari)", ""]:
            with self.subTest(name=name):
                self.assertFalse(geo_fi.street_name(name))

    def test_a_street_may_hold_a_dot_a_colon_and_an_apostrophe(self):
        for name in ["J.A. Palménin tie", "VPK:n kuja", "Uudenmaantie"]:
            with self.subTest(name=name):
                self.assertTrue(geo_fi.street_name(name))
