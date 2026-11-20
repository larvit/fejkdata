import collections
import unittest

from geo_test import script

geo_fr = script("geo-fr")

HEADER = ["REG", "Région", "DEP", "COM", "Commune", "PMUN"]
POPULATION = [dict(zip(HEADER, row, strict=True)) for row in (
    ("84", "Auvergne-Rhône-Alpes", "01", "01001", "L'Abergement-Clémenciat", "860"),
    ("11", "Île-de-France", "75", "75101", "Paris 1er Arrondissement", "15000"),
    ("11", "Île-de-France", "75", "75102", "Paris 2e Arrondissement", "20000"),
    ("44", "Grand Est", "55", "55050", "Bezonvaux", "0"),
    ("32", "Hauts-de-France", "02", "02001", "Abbécourt", "500"),
)]
DEPARTMENTS = {"01": "Ain", "02": "Aisne", "55": "Meuse", "75": "Paris", "974": "La Réunion"}
REGIONS = {"11": "Île-de-France", "32": "Hauts-de-France", "44": "Grand Est", "84": "Auvergne-Rhône-Alpes", "04": "La Réunion"}


def address(commune: str, postal_code: str, number: str = "1", suffix: str = "", lat: str = "46.0", lon: str = "5.0") -> dict[str, str]:
    """A row of BAN's address file, with the columns the import reads."""
    return {"code_insee": commune, "code_postal": postal_code, "numero": number, "rep": suffix, "lat": lat, "lon": lon}


class City(unittest.TestCase):
    def test_an_arrondissement_is_its_city(self) -> None:
        self.assertEqual([geo_fr.city(c) for c in ("75101", "75120", "69381", "69389", "13201", "13216")], ["75056", "75056", "69123", "69123", "13055", "13055"])

    def test_any_other_commune_is_itself(self) -> None:
        self.assertEqual([geo_fr.city(c) for c in ("01001", "75056", "13117", "2A004")], ["01001", "75056", "13117", "2A004"])


class Tally(unittest.TestCase):
    def test_a_commune_counts_its_addresses_under_each_postal_code(self) -> None:
        t = geo_fr.tally([address("01001", "01400"), address("01001", "01400"), address("01001", "01990")])
        self.assertEqual(t.codes["01001"], {"01400": 2, "01990": 1})

    def test_an_arrondissements_addresses_count_for_its_city(self) -> None:
        t = geo_fr.tally([address("75101", "75001"), address("75116", "75016"), address("75116", "75116")])
        self.assertEqual(dict(t.codes), {"75056": {"75001": 1, "75016": 1, "75116": 1}})

    def test_an_address_without_a_postal_code_is_left_out(self) -> None:
        self.assertEqual(dict(geo_fr.tally([address("01001", "")]).codes), {})

    def test_a_house_number_counts_by_its_length_bare_or_with_bis(self) -> None:
        t = geo_fr.tally([address("01001", "01400", "7"), address("01001", "01400", "12"), address("01001", "01400", "12", "bis"), address("01001", "01400", "630", "Bis")])
        self.assertEqual(dict(t.numbers), {(0, False): 1, (1, False): 1, (1, True): 1, (2, True): 1})

    def test_a_number_with_another_suffix_or_out_of_range_counts_for_no_length(self) -> None:
        t = geo_fr.tally([address("01001", "01400", "12", "ter"), address("01001", "01400", "12", "a"), address("01001", "01400", "0"), address("01001", "01400", "99999"), address("01001", "01400", "")])
        self.assertEqual(dict(t.numbers), {})
        self.assertEqual(t.codes["01001"], {"01400": 5})

    def test_a_communes_centre_is_the_mean_of_its_addresses_points(self) -> None:
        t = geo_fr.tally([address("01001", "01400", lat="46.0", lon="5.0"), address("01001", "01400", lat="47.0", lon="6.0"), address("01001", "01400", lat="", lon="")])
        self.assertEqual(t.centre("01001"), ("46.5000", "5.5000"))

    def test_half_a_point_is_no_point(self) -> None:
        t = geo_fr.tally([address("01001", "01400", lat="46.0", lon=""), address("01001", "01400", lat="", lon="5.0")])
        self.assertIsNone(t.centre("01001"))

    def test_asking_for_a_centre_adds_no_commune(self) -> None:
        t = geo_fr.tally([])
        self.assertIsNone(t.centre("01001"))
        self.assertEqual(dict(t.points), {})


class Communes(unittest.TestCase):
    def test_a_commune_holds_its_name_departement_region_and_municipal_population(self) -> None:
        self.assertEqual(geo_fr.communes(POPULATION)["01001"], {"code": "01001", "name": "L'Abergement-Clémenciat", "municipality": "01", "region": "84", "population": 860})

    def test_a_citys_arrondissements_are_one_commune_named_after_the_city_with_all_their_people(self) -> None:
        got = geo_fr.communes(POPULATION)
        self.assertEqual(got["75056"], {"code": "75056", "name": "Paris", "municipality": "75", "region": "11", "population": 35000})
        self.assertNotIn("75101", got)


class Places(unittest.TestCase):
    def places(self) -> tuple[list[dict[str, object]], list[dict[str, object]], list[dict[str, object]], list[dict[str, object]]]:
        t = geo_fr.tally([address("01001", "01400"), address("01001", "01400"), address("01001", "01990"), address("75101", "75001"), address("55050", "55100"), address("97501", "97500")])
        return geo_fr.places(t, geo_fr.communes(POPULATION), DEPARTMENTS, REGIONS)  # type: ignore[no-any-return]

    def test_a_commune_with_people_and_an_address_is_a_locality_keyed_by_its_code(self) -> None:
        _, _, localities, _ = self.places()
        self.assertEqual(localities, [
            {"code": "01001", "name": "L'Abergement-Clémenciat", "municipality": "01", "population": 860, "lat": "46.0000", "lon": "5.0000"},
            {"code": "75056", "name": "Paris", "municipality": "75", "population": 35000, "lat": "46.0000", "lon": "5.0000"},
        ])

    def test_a_postal_code_weighs_its_addresses_in_the_commune(self) -> None:
        _, _, _, postal_codes = self.places()
        self.assertEqual(postal_codes, [
            {"code": "01400", "locality": "01001", "addresses": 2},
            {"code": "01990", "locality": "01001", "addresses": 1},
            {"code": "75001", "locality": "75056", "addresses": 1},
        ])

    def test_a_departement_and_a_region_hold_their_localities_people_and_ship_only_with_one(self) -> None:
        regions, municipalities, _, _ = self.places()
        self.assertEqual(municipalities, [
            {"code": "01", "name": "Ain", "region": "84", "population": 860},
            {"code": "75", "name": "Paris", "region": "11", "population": 35000},
        ])
        self.assertEqual(regions, [
            {"code": "11", "name": "Île-de-France", "population": 35000, "timezone": "Europe/Paris"},
            {"code": "84", "name": "Auvergne-Rhône-Alpes", "population": 860, "timezone": "Europe/Paris"},
        ])

    def test_an_overseas_region_keeps_its_own_time(self) -> None:
        t = geo_fr.tally([address("97411", "97400")])
        reunion = {"97411": {"code": "97411", "name": "Saint-Denis", "municipality": "974", "region": "04", "population": 150000}}
        regions, _, _, _ = geo_fr.places(t, reunion, DEPARTMENTS, REGIONS)
        self.assertEqual(regions, [{"code": "04", "name": "La Réunion", "population": 150000, "timezone": "Indian/Reunion"}])


class AddressJson(unittest.TestCase):
    def test_the_number_leads_an_invented_street_bis_follows_it_and_a_length_no_address_has_is_left_out(self) -> None:
        self.assertEqual(geo_fr.address_json(collections.Counter({(0, False): 5, (1, True): 3})), """{
  "format": "{street-number} {street}\\n{postal-code} {locality}",
  "street": "{.locality as l}{/geo.FR.street}",
  "street-number": [
    { "format": "{int(1,9)}", "weight": 5 },
    { "format": "{int(10,99)} bis", "weight": 3 }
  ],
  "postal-code": "{l.postal-code.code}",
  "locality": "{l.name}",
  "municipality": "{l..municipality.name}",
  "region": "{l..municipality..region.name}"
}
""")


if __name__ == "__main__":
    unittest.main()
