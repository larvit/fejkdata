import collections
import importlib.util
import tempfile
import types
import unittest
from pathlib import Path
from typing import ClassVar

import geo


def script(name: str) -> types.ModuleType:
    """The import script `name`.py as a module. A script's hyphen keeps `import` from naming it; a module here takes an underscore."""
    spec = importlib.util.spec_from_file_location(name.replace("-", "_"), Path(__file__).with_name(name + ".py"))
    assert spec and spec.loader
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


geo_dk = script("geo-dk")
geo_fi = script("geo-fi")
geo_de = script("geo-de")


def baf(code: str = "", street: str = "", swedish: str = "", low: str = "", low_letter: str = "", high: str = "", high_letter: str = "", kunta: str = "091") -> str:
    """A line of Posti's BAF, each field at the position Posti's record description gives."""
    return " " * 13 + f"{code:5}" + " " * 84 + f"{street:30}{swedish:30}" + " " * 25 + f"{low:>5}{low_letter:1}" + " " * 7 + f"{high:>5}{high_letter:1}" + " " * 7 + kunta


def pcf(code: str, town: str, kind: str, swedish: str = "") -> str:
    """A line of Posti's PCF, each field at the position Posti's record description gives."""
    return " " * 13 + f"{code:5}{town:30}{swedish:30}" + " " * 32 + kind


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


class InventedStreets(unittest.TestCase):
    def test_a_street_is_a_row_with_its_name_alone(self) -> None:
        rows = geo.invented_streets("Am Markt\nStraße des 17. Juni\nSt.-Anna-Weg\nBahnhofstraße\n")
        self.assertEqual(rows, [{"name": "Am Markt"}, {"name": "Bahnhofstraße"}, {"name": "St.-Anna-Weg"}, {"name": "Straße des 17. Juni"}])

    def test_a_name_listed_twice_exits_naming_it(self) -> None:
        with self.assertRaisesRegex(SystemExit, "Am Markt"):
            geo.invented_streets("Am Markt\nBahnhofstraße\nAm Markt\n")

    def test_a_name_that_is_no_street_exits_naming_it(self) -> None:
        for bad in ("17. Juni", "Markt {x}", "Am\tMarkt", "_Weg"):
            with self.subTest(bad=bad), self.assertRaises(SystemExit) as raised:
                geo.invented_streets(f"Bahnhofstraße\n{bad}\n")
            self.assertIn(f"{bad!r} is not a street name", str(raised.exception))

    def test_no_name_exits(self) -> None:
        with self.assertRaises(SystemExit):
            geo.invented_streets("\n\n")


def geonames_de(code: str, place: str, land: str, kreis_code: str, kreis: str, lat: str = "50.0", lon: str = "10.0", accuracy: str = "4") -> str:
    """A line of GeoNames' DE.txt, each field at the position GeoNames' readme gives."""
    return f"DE\t{code}\t{place}\t\t{land}\t\t00\t{kreis}\t{kreis_code}\t{lat}\t{lon}\t{accuracy}"


class GermanPlaces(unittest.TestCase):
    LINES: ClassVar[list[str]] = [
        geonames_de("01945", "Tettau", "BB", "12066", "Landkreis Oberspreewald-Lausitz", "51.0", "13.0"),
        geonames_de("01945", "Guteborn", "BB", "12066", "Landkreis Oberspreewald-Lausitz", "51.4167", "13.9333"),
        geonames_de("01946", "Tettau", "BB", "12066", "Landkreis Oberspreewald-Lausitz", "52.0", "15.0"),
        geonames_de("83362", "Surberg", "BY", "09189", "Landkreis Traunstein"),
        geonames_de("83308", "Trostberg", "BY", "09189", "Landkreis Traunstein"),
        geonames_de("83309", "Trostberg", "BY", "09189", "Landkreis Traunstein"),
        geonames_de("87491", "Jungholz", "", "", ""),
    ]

    def test_a_land_is_named_in_german_and_weighs_its_postal_code_rows(self) -> None:
        regions, _, _, _ = geo_de.places(self.LINES)
        self.assertEqual(regions, [
            {"code": "BB", "name": "Brandenburg", "postal-codes": 3, "timezone": "Europe/Berlin"},
            {"code": "BY", "name": "Bayern", "postal-codes": 3, "timezone": "Europe/Berlin"},
        ])

    def test_a_kreis_weighs_the_postal_codes_of_its_localities(self) -> None:
        _, municipalities, _, _ = geo_de.places(self.LINES)
        self.assertEqual(municipalities, [
            {"code": "09189", "name": "Landkreis Traunstein", "region": "BY", "postal-codes": 3},
            {"code": "12066", "name": "Landkreis Oberspreewald-Lausitz", "region": "BB", "postal-codes": 3},
        ])

    def test_a_kreis_spelt_two_ways_exits_naming_both(self) -> None:
        lines = [*self.LINES, geonames_de("83310", "Trostberg", "BY", "09189", "Traunstein")]
        with self.assertRaises(SystemExit) as raised:
            geo_de.places(lines)
        self.assertIn("'Landkreis Traunstein' and 'Traunstein'", str(raised.exception))

    def test_a_kreis_is_named_as_a_german_writes_it(self) -> None:
        lines = [
            geonames_de("38226", "Salzgitter", "NI", "03102", "Salzgitter, Stadt"),
            geonames_de("90402", "Nürnberg", "BY", "09564", "Nuremberg"),
        ]
        _, municipalities, _, _ = geo_de.places(lines)
        self.assertEqual([r["name"] for r in municipalities], ["Stadt Salzgitter", "Kreisfreie Stadt Nürnberg"])

    def test_a_locality_is_a_place_name_in_a_kreis_at_the_mean_of_its_rows_points(self) -> None:
        _, _, localities, _ = geo_de.places(self.LINES)
        self.assertEqual([(r["code"], r["name"], r["municipality"], r["postal-codes"]) for r in localities], [
            ("09189-surberg", "Surberg", "09189", 1), ("09189-trostberg", "Trostberg", "09189", 2),
            ("12066-guteborn", "Guteborn", "12066", 1), ("12066-tettau", "Tettau", "12066", 2),
        ])
        self.assertEqual({k: localities[3][k] for k in ("lat", "lon")}, {"lat": "51.5000", "lon": "14.0000"})

    def test_a_locality_is_keyed_by_its_name_in_ascii(self) -> None:
        for place, code in (("Müllrose", "mullrose"), ("Groß Köris", "gross-koris"), ("Jessen (Elster)", "jessen-elster")):
            with self.subTest(place=place):
                self.assertEqual(geo_de.slug(place), code)

    def test_two_places_of_a_kreis_with_one_code_exit_naming_both(self) -> None:
        lines = [geonames_de("01945", "Müller", "BB", "12066", "Landkreis"), geonames_de("01946", "Muller", "BB", "12066", "Landkreis")]
        with self.assertRaises(SystemExit) as raised:
            geo_de.places(lines)
        self.assertIn("'Muller' and 'Müller'", str(raised.exception))

    def test_every_place_of_a_city_state_is_the_city(self) -> None:
        lines = [
            geonames_de("22767", "Hamburg Altona-Altstadt", "HH", "02000", "Hamburg, Freie und Hansestadt"),
            geonames_de("20095", "Hamburg", "HH", "02000", "Hamburg, Freie und Hansestadt"),
            geonames_de("13403", "Reinickendorf", "BE", "11000", "Berlin"),
        ]
        _, _, localities, postal_codes = geo_de.places(lines)
        self.assertEqual([(r["code"], r["name"], r["postal-codes"]) for r in localities], [("02000-hamburg", "Hamburg", 2), ("11000-berlin", "Berlin", 1)])
        self.assertEqual(len(postal_codes), 3)

    def test_a_postal_code_lists_every_locality_it_holds_once(self) -> None:
        _, _, _, postal_codes = geo_de.places(self.LINES)
        self.assertEqual(postal_codes, [
            {"code": "01945", "locality": "12066-guteborn"}, {"code": "01945", "locality": "12066-tettau"}, {"code": "01946", "locality": "12066-tettau"},
            {"code": "83308", "locality": "09189-trostberg"}, {"code": "83309", "locality": "09189-trostberg"}, {"code": "83362", "locality": "09189-surberg"},
        ])

    def test_a_row_with_no_land_is_dropped(self) -> None:
        regions, municipalities, localities, postal_codes = geo_de.places(self.LINES)
        self.assertNotIn("Jungholz", [r["name"] for r in localities])
        self.assertNotIn("87491", [r["code"] for r in postal_codes])
        self.assertEqual((len(regions), len(municipalities)), (2, 2))

    def test_a_row_with_no_accuracy_is_a_postal_code_of_a_firm_or_an_office_and_is_dropped(self) -> None:
        lines = [*self.LINES, geonames_de("01968", "Agentur für Arbeit Senftenberg", "BB", "12066", "Landkreis Oberspreewald-Lausitz", accuracy="")]
        regions, _, localities, postal_codes = geo_de.places(lines)
        self.assertNotIn("Agentur für Arbeit Senftenberg", [r["name"] for r in localities])
        self.assertNotIn("01968", [r["code"] for r in postal_codes])
        self.assertEqual(regions[0]["postal-codes"], 3)

    def test_an_unknown_land_code_exits_naming_it(self) -> None:
        with self.assertRaisesRegex(SystemExit, "XX"):
            geo_de.places([geonames_de("01945", "Tettau", "XX", "12066", "Landkreis")])

    def test_a_row_in_a_land_with_no_kreis_exits_naming_its_postal_code(self) -> None:
        with self.assertRaisesRegex(SystemExit, "01945 Tettau: names no Kreis"):
            geo_de.places([geonames_de("01945", "Tettau", "BB", "", "")])


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

    def test_a_swedish_abbreviation_stays_lower_case(self) -> None:
        for posti, want in [("INGÅ ST", "Ingå st"), ("SJUNDEÅ KBY", "Sjundeå kby"), ("HANGÖ NORRA", "Hangö Norra")]:
            with self.subTest(posti=posti):
                self.assertEqual(geo_fi.recased(posti, set()), want)


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


class SwedishPlaces(unittest.TestCase):
    KUNNAT: ClassVar[dict[str, dict[str, str]]] = {"139": {"name": "Ii", "swedish-name": "Ijo", "region": "17"}}

    def places(self, lines: list[str]) -> dict[str, str]:
        return dict(geo_fi.swedish_places(lines, geo_fi.normal_codes(lines, self.KUNNAT), self.KUNNAT))

    def test_a_postitoimipaikka_takes_the_swedish_name_posti_gives_it_written_as_a_place_name(self) -> None:
        lines = [pcf("00100", "HELSINKI", "1", "HELSINGFORS"), pcf("91101", "YLI-II", "1", "YLI-II"), pcf("10600", "INKOO AS", "1", "INGÅ ST")]
        self.assertEqual(self.places(lines), {"Helsinki": "Helsingfors", "Yli-Ii": "Yli-Ii", "Inkoo as": "Ingå st"})

    def test_a_postitoimipaikka_without_a_swedish_name_keeps_its_finnish_one(self) -> None:
        self.assertEqual(self.places([pcf("04400", "JÄRVENPÄÄ", "1")]), {"Järvenpää": "Järvenpää"})

    def test_a_postinumero_that_is_not_normal_names_nothing(self) -> None:
        self.assertEqual(self.places([pcf("00101", "HELSINKI", "2", "HELSINGFORS")]), {})

    def test_the_swedish_name_most_postinumerot_give_wins(self) -> None:
        lines = [pcf("00100", "HELSINKI", "1", "HELSINGFORS"), pcf("00101", "HELSINKI", "1", "HELSINGFORS"), pcf("00102", "HELSINKI", "1", "HELSINKI")]
        self.assertEqual(self.places(lines), {"Helsinki": "Helsingfors"})


class Mainland(unittest.TestCase):
    def test_aland_and_its_kunnat_are_left_out(self) -> None:
        regions, municipalities = geo_fi.mainland({"01": "Uusimaa", "21": "Ahvenanmaa"}, {"186": {"name": "Järvenpää", "swedish-name": "Träskända", "region": "01"}, "478": {"name": "Maarianhamina", "swedish-name": "Mariehamn", "region": "21"}}, {"MK01": 1700, "MK21": 30, "KU186": 46, "KU478": 11})
        self.assertEqual(regions, {"01": {"code": "01", "name": "Uusimaa", "population": 1700, "timezone": "Europe/Helsinki"}})
        self.assertEqual(municipalities, {"186": {"code": "186", "name": "Järvenpää", "swedish-name": "Träskända", "region": "01", "population": 46}})


class Tally(unittest.TestCase):
    CODES: ClassVar[dict[str, str]] = {"04400": "Järvenpää", "99800": "Inari"}

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
        places, postal_codes = geo_fi.localities(codes_of, {"00100": 10, "00102": 5, "99991": 7}, {"Helsinki": "091", "Tyhjälä": "148", "Pisteetön": "148"}, points, {"Helsinki": "Helsingfors"})
        self.assertEqual(places, {"Helsinki": {"name": "Helsinki", "swedish-name": "Helsingfors", "municipality": "091", "population": 15, "lat": "60.5000", "lon": "25.5000"}})
        self.assertEqual(postal_codes, {"Helsinki": {"00100": 10, "00102": 5}})


if __name__ == "__main__":
    unittest.main()
