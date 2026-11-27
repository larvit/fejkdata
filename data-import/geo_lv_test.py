import collections
import unittest
from typing import Any

import geo
from geo_test import script

geo_lv = script("geo-lv")

STATE, NOVADS, PILSETA, PAGASTS, CIEMS, IELA = "101", "113", "104", "105", "106", "107"


def place(code: str, kind: str, name: str, parent: str, parent_kind: str, atvk: str = "", status: str = "EKS") -> dict[str, str]:
	"""A row of a file of places or streets, with the columns the import reads."""
	return {"KODS": code, "TIPS_CD": kind, "NOSAUKUMS": name, "VKUR_CD": parent, "VKUR_TIPS": parent_kind, "ATRIB": atvk, "STATUSS": status}


NOVADI = [place("100016606", NOVADS, "Limbažu nov.", "100000000", STATE, "0035000"), place("100016390", NOVADS, "Talsu nov.", "100000000", STATE, "0051000")]
PILSETAS = [place("100003003", PILSETA, "Rīga", "100000000", STATE, "0001000"), place("100003366", PILSETA, "Ainaži", "100016606", NOVADS, "0035200")]
PAGASTI = [place("100013922", PAGASTS, "Abavas pag.", "100016390", NOVADS, "0051400")]
CIEMI = [place("100195988", CIEMS, "Dreimaņciems", "100013922", PAGASTS), place("100187337", CIEMS, "Lielvirbi", "100013922", PAGASTS)]
IELAS = [
	place("100366498", IELA, "Aizsaules iela", "100003366", PILSETA),
	place("100462850", IELA, "Akmeņu iela", "100003366", PILSETA),
	place("100500001", IELA, "Liepu iela", "100195988", CIEMS),
]


def address(number: str, parent: str, parent_kind: str, postal_code: str = "LV-4035", lat: str = "57.85", lon: str = "24.36", status: str = "EKS", code: str = "101000034") -> dict[str, str]:
	"""A row of the file of building addresses, with the columns the import reads."""
	return {"KODS": code, "NOSAUKUMS": number, "VKUR_CD": parent, "VKUR_TIPS": parent_kind, "ATRIB": postal_code, "DD_N": lat, "DD_E": lon, "STATUSS": status}


def register() -> Any:
	"""The places and streets of the fixtures above."""
	return geo_lv.Register(NOVADI, PILSETAS, PAGASTI, CIEMI, IELAS)


def tally(*rows: dict[str, str], apartments: dict[str, int] | None = None) -> Any:
	"""The tally of the address rows, placed by the fixtures' register, each building holding the apartments given for its code."""
	return geo_lv.tally(rows, register(), collections.Counter(apartments or {}))


class RegisterOfPlaces(unittest.TestCase):
	def test_a_novads_is_a_region_under_its_atvk_code(self) -> None:
		self.assertEqual(register().regions["0051000"], {"code": "0051000", "name": "Talsu nov.", "timezone": "Europe/Riga"})

	def test_a_state_city_is_a_region_a_municipality_and_a_locality(self) -> None:
		r = register()
		self.assertEqual(r.regions["0001000"], {"code": "0001000", "name": "Rīga", "timezone": "Europe/Riga"})
		self.assertEqual(r.municipalities["0001000"], {"code": "0001000", "name": "Rīga", "region": "0001000"})
		self.assertEqual(r.localities["100003003"], {"code": "100003003", "name": "Rīga", "municipality": "0001000"})

	def test_a_town_of_a_novads_is_a_municipality_and_a_locality_under_it(self) -> None:
		r = register()
		self.assertNotIn("0035200", r.regions)
		self.assertEqual(r.municipalities["0035200"], {"code": "0035200", "name": "Ainaži", "region": "0035000"})
		self.assertEqual(r.localities["100003366"], {"code": "100003366", "name": "Ainaži", "municipality": "0035200"})

	def test_a_pagasts_is_a_municipality_and_its_ciems_a_locality(self) -> None:
		r = register()
		self.assertEqual(r.municipalities["0051400"], {"code": "0051400", "name": "Abavas pag.", "region": "0051000"})
		self.assertEqual(r.localities["100195988"], {"code": "100195988", "name": "Dreimaņciems", "municipality": "0051400"})

	def test_a_street_belongs_to_its_town_or_ciems(self) -> None:
		r = register()
		self.assertEqual(r.streets["100366498"], ("100003366", "Aizsaules iela"))
		self.assertEqual(r.streets["100500001"], ("100195988", "Liepu iela"))

	def test_a_street_geo_refuses_has_no_name(self) -> None:
		doubled = place("100500002", IELA, "Liepu  iela", "100195988", CIEMS)
		remark = place("100500003", IELA, "Ozolu iela (Nr.2)", "100195988", CIEMS)
		r = geo_lv.Register(NOVADI, PILSETAS, PAGASTI, CIEMI, [*IELAS, doubled, remark])
		self.assertEqual((r.streets["100500002"], r.streets["100500003"]), (("100195988", None), ("100195988", None)))

	def test_a_row_no_longer_in_force_is_no_place(self) -> None:
		gone = place("100015645", NOVADS, "Aglonas nov.", "100000000", STATE, "0604300", status="DEL")
		wrong = place("100480957", IELA, "1", "100195988", CIEMS, status="ERR")
		r = geo_lv.Register([*NOVADI, gone], PILSETAS, PAGASTI, CIEMI, [*IELAS, wrong])
		self.assertNotIn("0604300", r.regions)
		self.assertNotIn("100480957", r.streets)

	def test_a_ciems_whose_pagasts_is_not_in_force_is_no_place(self) -> None:
		r = geo_lv.Register(NOVADI, PILSETAS, PAGASTI, [*CIEMI, place("100195377", CIEMS, "Abavmala", "100014748", PAGASTS)], IELAS)
		self.assertNotIn("100195377", r.localities)


class HouseNumber(unittest.TestCase):
	def test_a_number(self) -> None:
		self.assertEqual(geo_lv.house_number("18"), (18, False))

	def test_a_letter_follows_the_number(self) -> None:
		self.assertEqual(geo_lv.house_number("4A"), (4, True))

	def test_a_korpuss_and_a_second_number_are_dropped(self) -> None:
		self.assertEqual([geo_lv.house_number(n) for n in ("195 k-1", "25D k-1", "11/13", "18/20 k-2")], [(195, False), (25, True), (11, False), (18, False)])

	def test_a_houses_name_is_no_number(self) -> None:
		for name in ("Amatnieki", "Dzirnavas 1", "0", "12345"):
			self.assertIsNone(geo_lv.house_number(name), name)


class TallyOfAddresses(unittest.TestCase):
	def test_an_address_on_a_street_counts_for_the_street_its_locality_and_its_postal_code(self) -> None:
		counted = tally(address("18", "100366498", IELA), address("4A", "100366498", IELA), address("2", "100462850", IELA, "LV-4036"))
		self.assertEqual(counted.streets, {("100003366", "Aizsaules iela"): 2, ("100003366", "Akmeņu iela"): 1})
		self.assertEqual(counted.addresses, {"100003366": 3})
		self.assertEqual(counted.postal_codes, {("100003366", "LV-4035"): 2, ("100003366", "LV-4036"): 1})

	def test_an_address_on_a_street_geo_refuses_counts_for_its_locality_and_for_no_street(self) -> None:
		remark = place("100500003", IELA, "Ozolu iela (Nr.2)", "100195988", CIEMS)
		counted = geo_lv.tally([address("5", "100500003", IELA, "LV-3294")], geo_lv.Register(NOVADI, PILSETAS, PAGASTI, CIEMI, [*IELAS, remark]), collections.Counter())
		self.assertEqual((counted.addresses, counted.streets), ({"100195988": 1}, {}))

	def test_a_named_house_counts_for_its_locality_and_for_no_street(self) -> None:
		counted = tally(address("Auliki", "100003366", PILSETA), address("Amatnieki", "100195988", CIEMS, "LV-3294"))
		self.assertEqual(counted.addresses, {"100003366": 1, "100195988": 1})
		self.assertEqual(counted.streets, {})

	def test_a_house_outside_every_ciems_counts_for_its_pagasts_alone(self) -> None:
		counted = tally(address("Riņņi", "100013922", PAGASTS, "LV-3294"))
		self.assertEqual(counted.addresses, {})
		self.assertEqual(counted.outside, {"0051400": 1})

	def test_an_address_no_longer_in_force_or_under_no_place_is_not_counted(self) -> None:
		counted = tally(address("18", "100366498", IELA, status="DEL"), address("18", "100999999", IELA), address("Auliki", "100999999", CIEMS))
		self.assertEqual((counted.addresses, counted.outside, counted.streets), ({}, {}, {}))

	def test_house_numbers_count_by_their_length_and_letter(self) -> None:
		counted = tally(address("18", "100366498", IELA), address("4A", "100366498", IELA), address("7", "100462850", IELA), address("195 k-1", "100462850", IELA))
		self.assertEqual(counted.numbers, {(1, False): 1, (0, True): 1, (0, False): 1, (2, False): 1})

	def test_a_localitys_point_is_the_mean_of_its_addresses(self) -> None:
		counted = tally(address("18", "100366498", IELA, lat="57.80", lon="24.30"), address("Auliki", "100003366", PILSETA, lat="57.90", lon="24.40"))
		self.assertEqual(counted.point("100003366"), ("57.8500", "24.3500"))

	def test_an_address_with_no_point_counts_and_the_point_is_the_mean_of_the_others(self) -> None:
		counted = tally(address("18", "100366498", IELA, lat="57.80", lon="24.30"), address("Auliki", "100003366", PILSETA, lat="", lon=""))
		self.assertEqual(counted.addresses, {"100003366": 2})
		self.assertEqual(counted.point("100003366"), ("57.8000", "24.3000"))

	def test_a_building_counts_once_and_once_more_for_each_apartment_in_it(self) -> None:
		counted = tally(address("18", "100366498", IELA, code="101000034"), address("Riņņi", "100013922", PAGASTS, code="101000042"), apartments={"101000034": 3, "101000042": 1})
		self.assertEqual(counted.addresses, {"100003366": 4})
		self.assertEqual(counted.postal_codes, {("100003366", "LV-4035"): 4})
		self.assertEqual(counted.streets, {("100003366", "Aizsaules iela"): 4})
		self.assertEqual(counted.outside, {"0051400": 2})
		self.assertEqual(counted.numbers, {(1, False): 1})


class Apartments(unittest.TestCase):
	def test_a_building_holds_its_apartments_in_force(self) -> None:
		rows = [{"VKUR_CD": "101000034", "STATUSS": "EKS"}, {"VKUR_CD": "101000034", "STATUSS": "EKS"}, {"VKUR_CD": "101000034", "STATUSS": "ERR"}, {"VKUR_CD": "101000042", "STATUSS": "EKS"}]
		self.assertEqual(geo_lv.apartments(rows), {"101000034": 2, "101000042": 1})


class People(unittest.TestCase):
	def test_a_village_counts_under_its_register_code_and_a_town_a_pagasts_and_a_novads_under_atvk(self) -> None:
		rows = [["Teritoriālā vienība", "2026"], ["LV Latvija", "1860565"], ["LV0051000 Talsu novads", "34000"], ["LV0051400 ..Abavas pagasts", "1100"], ["LV100195988 ...Dreimaņciems", "140"]]
		self.assertEqual(geo_lv.people(rows), {"0051000": 34000, "0051400": 1100, "100195988": 140})

	def test_a_count_the_bureau_does_not_give_and_a_row_of_another_kind_are_left_out(self) -> None:
		rows = [["LV100187337 ...Lielvirbi", "…"], ["LVDPA0001 Rīga", "600000"], ["LVRIG01 Centrs", "30000"], ["UNK_LV Nezināma", "12"]]
		self.assertEqual(geo_lv.people(rows), {})

	def test_carnikava_and_its_pagasts_split_the_pagasts_the_town_was_a_village_of(self) -> None:
		got = geo_lv.people([["LV0023410 ..Carnikavas pagasts", "10669"], ["LV100184532 ...Carnikava", "5078"]])
		self.assertEqual((got["0023210"], got["0023411"]), (5078, 5591))

	def test_carnikava_counted_without_the_pagasts_it_was_a_village_of_stops_the_import(self) -> None:
		with self.assertRaisesRegex(ValueError, "no count for 0023410"):
			geo_lv.people([["LV100184532 ...Carnikava", "5078"]])


class Weighed(unittest.TestCase):
	def test_a_village_a_municipality_and_a_town_take_their_people_and_a_region_sums_its_municipalities(self) -> None:
		r = register()
		geo_lv.weighed(r, tally(address("18", "100366498", IELA), address("3", "100500001", IELA, "LV-3294")), {"0035200": 700, "0051400": 1100, "100195988": 140, "0001000": 600000})
		self.assertEqual([r.localities[c]["population"] for c in ("100003366", "100195988")], [700, 140])
		self.assertEqual([r.municipalities[c]["population"] for c in ("0035200", "0051400", "0001000")], [700, 1100, 600000])
		self.assertEqual([r.regions[c]["population"] for c in ("0035000", "0051000", "0001000")], [700, 1100, 600000])
		self.assertNotIn("population", r.localities["100187337"])

	def test_a_municipality_with_no_count_stops_the_import_and_is_named(self) -> None:
		r = register()
		with self.assertRaisesRegex(ValueError, "no count for municipality 0001000, 0051400$"):
			geo_lv.weighed(r, geo_lv.tally([], r, collections.Counter()), {"0035200": 700})

	def test_a_region_sums_every_municipality_in_it(self) -> None:
		arlava = place("100013930", PAGASTS, "Ārlavas pag.", "100016390", NOVADS, "0051420")
		r = geo_lv.Register(NOVADI, PILSETAS, [*PAGASTI, arlava], CIEMI, IELAS)
		geo_lv.weighed(r, geo_lv.tally([], r, collections.Counter()), {"0035200": 700, "0051400": 1100, "0051420": 300, "0001000": 600000})
		self.assertEqual(r.regions["0051000"]["population"], 1400)

	def test_a_locality_takes_the_point_of_its_addresses(self) -> None:
		r = register()
		geo_lv.weighed(r, tally(address("18", "100366498", IELA, lat="57.85", lon="24.36")), {"0035200": 700, "0051400": 1100, "0001000": 600000})
		self.assertEqual((r.localities["100003366"]["lat"], r.localities["100003366"]["lon"]), ("57.8500", "24.3600"))

	def test_a_village_with_no_count_shares_the_rest_of_its_pagasts_with_the_houses_outside_every_village(self) -> None:
		r = register()
		counted = tally(address("3", "100500001", IELA, "LV-3294"), address("Riņņi", "100013922", PAGASTS, "LV-3294"), address("Šķetras", "100013922", PAGASTS, "LV-3294"), address("Kalni", "100187337", CIEMS, "LV-3294"))
		geo_lv.weighed(r, counted, {"0035200": 700, "0051400": 1100, "100187337": 200, "0001000": 600000})
		self.assertEqual(r.localities["100195988"]["population"], 300)


class AddressJson(unittest.TestCase):
	def test_the_postal_code_follows_the_locality_after_a_comma(self) -> None:
		got = geo.address_json(collections.Counter({(0, False): 3}), layout=geo_lv.LAYOUT)
		self.assertIn('"format": "{street} {street-number}\\n{locality}, {postal-code}"', got)


if __name__ == "__main__":
	unittest.main()
