import struct
import unittest
from typing import Any

from geo_test import script

geo_ee = script("geo-ee")

FIELDS = ("ANIMI", "AKOOD", "TYYP", "ONIMI", "OKOOD", "MNIMI", "MKOOD")
UNITS = [dict(zip(FIELDS, row, strict=True)) for row in (
	("Tabasalu alevik", "8009", "7", "Harku vald", "0198", "Harju maakond", "0037"),
	("Keila linn", "0296", "4", "Keila linn", "0296", "Harju maakond", "0037"),
	("Pirita linnaosa", "0596", "6", "Tallinn", "0784", "Harju maakond", "0037"),
	("Nõmme linnaosa", "0524", "6", "Tallinn", "0784", "Harju maakond", "0037"),
	("Ruhnu küla", "7185", "8", "Ruhnu vald", "0689", "Saare maakond", "0074"),
	("Haapsalu linn", "1692", "5", "Haapsalu linn", "0184", "Lääne maakond", "0056"),
)]


def dbf(fields: list[tuple[str, int]], records: list[list[str]]) -> bytes:
	"""A dBase file of character fields, each a name and a length, with its records in Windows-1257."""
	header = 32 + 32 * len(fields) + 1
	size = 1 + sum(length for _, length in fields)
	out = struct.pack("<4xIHH20x", len(records), header, size)
	for name, length in fields:
		out += struct.pack("<11sc4xB15x", name.encode("ascii"), b"C", length)
	out += b"\r"
	for record in records:
		out += b" " + b"".join(cell.encode("cp1257").ljust(length) for cell, (_, length) in zip(record, fields, strict=True))
	return out


def address(code: str, near: str, place: str, east: str = "523306.03", north: str = "6574787.37") -> dict[str, str]:
	"""A row of the postal code file, with the columns the import reads."""
	return {"SIHTNUMBER": code, "LAHIAADRESS": near, "TAISAADRESS": f"{place}, {near}", "VIITEPUNKT_X": east, "VIITEPUNKT_Y": north}


def tally(*rows: dict[str, str]) -> Any:
	"""The tally of the rows, placed by UNITS."""
	counted = geo_ee.Tally()
	for row in rows:
		counted.add(row, geo_ee.units(UNITS)[3])
	return counted


class DbfRows(unittest.TestCase):
	def test_a_record_holds_each_field_under_its_name_without_its_padding(self) -> None:
		data = dbf([("ANIMI", 20), ("AKOOD", 4)], [["Jõelähtme küla", "2514"], ["Võõpsu alevik", "9712"]])
		self.assertEqual(list(geo_ee.dbf_rows(data)), [{"ANIMI": "Jõelähtme küla", "AKOOD": "2514"}, {"ANIMI": "Võõpsu alevik", "AKOOD": "9712"}])


class Units(unittest.TestCase):
	def test_an_asustusyksus_is_a_locality_under_its_ehak_code(self) -> None:
		_, _, localities, code_of = geo_ee.units(UNITS)
		self.assertEqual(localities["8009"], {"code": "8009", "name": "Tabasalu alevik", "municipality": "0198"})
		self.assertEqual(code_of[("Harju maakond", "Harku vald", "Tabasalu alevik")], "8009")

	def test_a_linnaosa_is_its_linn(self) -> None:
		_, _, localities, code_of = geo_ee.units(UNITS)
		self.assertEqual(localities["0784"], {"code": "0784", "name": "Tallinn", "municipality": "0784"})
		self.assertEqual([code_of[("Harju maakond", "Tallinn", part)] for part in ("Pirita linnaosa", "Nõmme linnaosa")], ["0784", "0784"])
		self.assertNotIn("0596", localities)

	def test_a_linn_with_no_asustusyksus_is_placed_by_its_own_name(self) -> None:
		self.assertEqual(geo_ee.units(UNITS)[3][("Harju maakond", "Keila linn")], "0296")

	def test_a_linn_inside_a_municipality_of_its_name_keeps_its_own_ehak_code(self) -> None:
		_, _, localities, code_of = geo_ee.units(UNITS)
		self.assertEqual(localities["1692"], {"code": "1692", "name": "Haapsalu linn", "municipality": "0184"})
		self.assertEqual(code_of[("Lääne maakond", "Haapsalu linn", "Haapsalu linn")], "1692")
		self.assertNotIn("0184", localities)
		self.assertNotIn(("Lääne maakond", "Haapsalu linn"), code_of)

	def test_a_vald_is_no_place_without_its_asustusyksus(self) -> None:
		self.assertNotIn(("Harju maakond", "Harku vald"), geo_ee.units(UNITS)[3])

	def test_a_maakond_and_a_municipality_keep_their_codes(self) -> None:
		regions, municipalities, _, _ = geo_ee.units(UNITS)
		self.assertEqual(regions["0074"], {"code": "0074", "name": "Saare maakond", "timezone": "Europe/Tallinn"})
		self.assertEqual(municipalities["0689"], {"code": "0689", "name": "Ruhnu vald", "region": "0074"})


class LatLon(unittest.TestCase):
	def test_the_projections_origin_is_its_latitude_and_meridian(self) -> None:
		lat, lon = geo_ee.lat_lon(500000, 6375000)
		self.assertAlmostEqual(lat, 57.51755393, places=6)
		self.assertAlmostEqual(lon, 24, places=6)

	def test_tallinns_town_hall_square(self) -> None:
		"""542,000 m east and 6,589,000 m north is in Tallinn's old town, near 59.437° N 24.741° E."""
		lat, lon = geo_ee.lat_lon(542000, 6589000)
		self.assertAlmostEqual(lat, 59.437, delta=0.005)
		self.assertAlmostEqual(lon, 24.741, delta=0.005)


class StreetOf(unittest.TestCase):
	def test_a_street_and_a_number(self) -> None:
		self.assertEqual(geo_ee.street_of("Sarapuu tn 70"), ("Sarapuu tn", 70, False))

	def test_a_letter_follows_the_number(self) -> None:
		self.assertEqual(geo_ee.street_of("Pargi tn 2b"), ("Pargi tn", 2, True))

	def test_an_apartment_and_a_second_building_are_dropped(self) -> None:
		self.assertEqual([geo_ee.street_of(near) for near in ("Saue tn 13-5", "Toome tn 18/2", "Reinu tee 8a/1-3")], [("Saue tn", 13, False), ("Toome tn", 18, False), ("Reinu tee", 8, True)])

	def test_a_street_in_a_vaikekoht_is_named_without_it(self) -> None:
		self.assertEqual(geo_ee.street_of("Aseri AÜ, Roosi tn 5"), ("Roosi tn", 5, False))

	def test_a_street_geo_refuses_is_none(self) -> None:
		for near in ["Sarapuu  tn 70", "Pargi (vana) tn 2", "Sarapuu tn Sarapuu tn 70"]:
			with self.subTest(near=near):
				self.assertIsNone(geo_ee.street_of(near))

	def test_each_kind_of_street(self) -> None:
		for kind in ("tn", "mnt", "pst", "tee", "allee", "põik", "väljak", "plats"):
			self.assertEqual(geo_ee.street_of(f"Vana-Narva {kind} 4"), (f"Vana-Narva {kind}", 4, False))

	def test_a_farm_a_forest_district_and_a_plot_are_no_street(self) -> None:
		for near in ("Antsu", "Aegviidu metskond 12", "Kaasiku vkt 26", "Sputniku AÜ 9", "Uueleiva tee L2", "Keila-Paldiski 1,5-4,9 km"):
			self.assertIsNone(geo_ee.street_of(near), near)


class TallyOfAddresses(unittest.TestCase):
	def test_a_locality_counts_its_addresses_under_each_postal_code(self) -> None:
		keila = "Harju maakond, Keila linn"
		counted = tally(address("76607", "Pargi tn 2b", keila), address("76607", "Metsa tn 12", keila), address("76609", "Jõe tn 30", keila))
		self.assertEqual(counted.addresses, {"0296": 3})
		self.assertEqual(counted.postal_codes, {("0296", "76607"): 2, ("0296", "76609"): 1})

	def test_a_linnaosas_addresses_count_for_its_linn(self) -> None:
		counted = tally(address("11911", "Pirita tee 26", "Harju maakond, Tallinn, Pirita linnaosa"), address("11611", "Pärnu mnt 300", "Harju maakond, Tallinn, Nõmme linnaosa"))
		self.assertEqual(counted.addresses, {"0784": 2})
		self.assertEqual(counted.streets, {("0784", "Pirita tee"): 1, ("0784", "Pärnu mnt"): 1})

	def test_a_street_counts_each_address_on_it_and_a_farm_is_no_street(self) -> None:
		ruhnu, tabasalu = "Saare maakond, Ruhnu vald, Ruhnu küla", "Harju maakond, Harku vald, Tabasalu alevik"
		counted = tally(address("93001", "Antsu", ruhnu), address("76901", "Sarapuu tn 70", tabasalu), address("76901", "Sarapuu tn 72-1", tabasalu))
		self.assertEqual(counted.streets, {("8009", "Sarapuu tn"): 2})
		self.assertEqual(counted.addresses, {"7185": 1, "8009": 2})

	def test_house_numbers_count_by_their_length_and_letter(self) -> None:
		keila = "Harju maakond, Keila linn"
		counted = tally(address("76607", "Pargi tn 2b", keila), address("76607", "Metsa tn 12", keila), address("76607", "Jõe tn 7", keila), address("76607", "Jõe tn 30000", keila))
		self.assertEqual(counted.numbers, {(0, True): 1, (1, False): 1, (0, False): 1})

	def test_an_address_in_a_vaikekoht_counts_for_its_asustusyksus(self) -> None:
		counted = tally(address("76901", "Aia AÜ, Roosi tn 5", "Harju maakond, Harku vald, Tabasalu alevik"))
		self.assertEqual((counted.addresses, counted.unplaced), ({"8009": 1}, {}))
		self.assertEqual(counted.streets, {("8009", "Roosi tn"): 1})

	def test_a_house_number_of_zero_or_of_five_digits_counts_for_no_street(self) -> None:
		keila = "Harju maakond, Keila linn"
		for near in ("Pargi tn 0", "Pargi tn 10000"):
			counted = tally(address("76607", near, keila))
			self.assertEqual((counted.addresses, counted.postal_codes), ({"0296": 1}, {("0296", "76607"): 1}), near)
			self.assertEqual((counted.streets, counted.numbers), ({}, {}), near)

	def test_an_address_in_no_settlement_unit_is_noted_and_not_counted(self) -> None:
		counted = tally(address("74114", "Kroodi raudteeharu, maatükk I", "Harju maakond, Maardu linn"))
		self.assertEqual(counted.addresses, {})
		self.assertEqual(counted.unplaced, {("Harju maakond", "Maardu linn"): 1})

	def test_a_localitys_point_is_the_mean_of_its_addresses(self) -> None:
		keila = "Harju maakond, Keila linn"
		counted = tally(address("76607", "Pargi tn 2b", keila, "541000", "6588000"), address("76607", "Jõe tn 7", keila, "543000", "6590000"))
		lat, lon = geo_ee.lat_lon(542000, 6589000)
		self.assertEqual(counted.point("0296"), (f"{lat:.4f}", f"{lon:.4f}"))


def stat(*places: tuple[str, str, int | None]) -> dict[str, Any]:
	"""The census table, in JSON-stat: each place's code, its label and its count."""
	return {
		"dimension": {"Elukoht": {"category": {"index": {code: i for i, (code, _, _) in enumerate(places)}, "label": {code: label for code, label, _ in places}}}},
		"value": [n for _, _, n in places],
	}


STAT = stat(
	("1", "Kogu Eesti", 1331824), ("L2", "Linnaline asustuspiirkond", 800000),
	("00370000000000", "HARJU MAAKOND", 510), ("003707840000L4", "..Tallinn", 400), ("003707840596L6", "....Pirita linnaosa", 30), ("003707840524M6", "....Nõmme linnaosa", 20),
	("003702960000L4", "..Keila linn", 50), ("00370198000001", "..Harku vald", 60), ("003701988008M5", "....8008 Tabasalu alevik", 60),
	("00740000000000", "SAARE MAAKOND", 0), ("00740689000001", "..Ruhnu vald", None), ("007406897185M8", "....7185 Ruhnu küla", None),
	("006009283092M5", "....3092 Kiltsi alevik", 180), ("00370141000001", "..Anija vald", 9), ("003701411001M8", "....1001 Aru küla", 4), ("003701982002M8", "....2002 Aru küla", 5),
)


class CensusOfSettlements(unittest.TestCase):
	def test_a_settlement_counts_under_its_ehak_code(self) -> None:
		self.assertEqual(geo_ee.Census(STAT).of("8008", "Tabasalu alevik", "0037"), 60)

	def test_a_linn_that_holds_no_settlement_counts_under_its_own_code(self) -> None:
		self.assertEqual([geo_ee.Census(STAT).of(c, n, "0037") for c, n in (("0784", "Tallinn"), ("0296", "Keila linn"))], [400, 50])

	def test_a_settlement_whose_code_changed_since_the_census_counts_under_its_name_in_its_maakond(self) -> None:
		self.assertEqual(geo_ee.Census(STAT).of("8009", "Tabasalu alevik", "0037"), 60)
		self.assertIsNone(geo_ee.Census(STAT).of("8009", "Tabasalu alevik", "0074"))

	def test_a_settlement_counts_under_its_code_before_its_name(self) -> None:
		census = geo_ee.Census(stat(("003701985001M8", "....5001 Vana küla", 7), ("003701985002M8", "....5002 Uus küla", 9)))
		self.assertEqual(census.of("5001", "Uus küla", "0037"), 7)

	def test_a_name_two_settlements_of_a_maakond_share_counts_for_neither(self) -> None:
		self.assertIsNone(geo_ee.Census(STAT).of("1002", "Aru küla", "0037"))

	def test_a_maakond_a_vald_and_a_linnaosa_are_no_settlement(self) -> None:
		places = (("0037", "HARJU MAAKOND"), ("0198", "Harku vald"), ("0596", "Pirita linnaosa"), ("0524", "Nõmme linnaosa"))
		self.assertEqual([geo_ee.Census(STAT).of(code, name, "0037") for code, name in places], [None, None, None, None])

	def test_a_withheld_count_is_two(self) -> None:
		self.assertEqual(geo_ee.Census(STAT).of("7185", "Ruhnu küla", "0074"), 2)

	def test_kiltsi_counts_as_the_alevik_it_was(self) -> None:
		self.assertEqual(geo_ee.Census(STAT).of("3091", "Kiltsi küla", "0060"), 180)

	def test_a_settlement_formed_since_the_census_has_no_count(self) -> None:
		self.assertIsNone(geo_ee.Census(STAT).of("6674", "Raadi alev", "0079"))


class Weighed(unittest.TestCase):
	def test_a_locality_takes_its_census_count_and_a_municipality_and_a_region_sum_theirs(self) -> None:
		regions, municipalities, localities, _ = geo_ee.units(UNITS)
		tabasalu, tallinn = "Harju maakond, Harku vald, Tabasalu alevik", "Harju maakond, Tallinn, Pirita linnaosa"
		counted = tally(address("76901", "Sarapuu tn 70", tabasalu, "500000", "6375000"), address("11911", "Pirita tee 26", tallinn), address("11911", "Pirita tee 28", tallinn))
		geo_ee.weighed(regions, municipalities, localities, counted, geo_ee.Census(STAT))
		self.assertEqual([localities[c]["population"] for c in ("8009", "0784")], [60, 400])
		self.assertEqual((localities["8009"]["lat"], localities["8009"]["lon"]), ("57.5176", "24.0000"))
		self.assertEqual([municipalities[c]["population"] for c in ("0198", "0784", "0296", "0689", "0184")], [60, 400, 50, 2, 0])
		self.assertEqual([regions[c]["population"] for c in ("0037", "0074", "0056")], [510, 2, 0])
		self.assertNotIn("population", localities["7185"])

	def test_a_locality_with_no_street_and_no_count_counts_no_one(self) -> None:
		regions, municipalities, localities, _ = geo_ee.units(UNITS)
		geo_ee.weighed(regions, municipalities, localities, tally(address("90502", "Antsu", "Lääne maakond, Haapsalu linn, Haapsalu linn")), geo_ee.Census(STAT))
		self.assertNotIn("population", localities["1692"])
		self.assertEqual(municipalities["0184"]["population"], 0)

	def test_a_locality_with_a_street_and_no_count_stops_the_import(self) -> None:
		regions, municipalities, localities, _ = geo_ee.units(UNITS)
		with self.assertRaisesRegex(ValueError, "1692 Haapsalu linn"):
			geo_ee.weighed(regions, municipalities, localities, tally(address("90502", "Posti tn 1", "Lääne maakond, Haapsalu linn, Haapsalu linn")), geo_ee.Census(STAT))

	def test_a_settlement_formed_since_the_census_takes_its_estimate_which_no_municipality_or_region_sums(self) -> None:
		formed = [dict(zip(FIELDS, row, strict=True)) for row in (
			("Keskküla", "2975", "8", "Tõrva vald", "0824", "Valga maakond", "0081"),
			("Kohtla-Uueküla", "3262", "8", "Jõhvi vald", "0250", "Ida-Viru maakond", "0045"),
			("Koplimetsa küla", "3440", "8", "Rae vald", "0653", "Harju maakond", "0037"),
			("Raadi alev", "6674", "9", "Tartu vald", "0796", "Tartu maakond", "0079"),
			("Vahi alevik", "8852", "7", "Tartu vald", "0796", "Tartu maakond", "0079"),
		)]
		regions, municipalities, localities, code_of = geo_ee.units(formed)
		counted = geo_ee.Tally()
		counted.add(address("60534", "Keskuse", "Valga maakond, Tõrva vald, Keskküla"), code_of)
		counted.add(address("30328", "Kooli", "Ida-Viru maakond, Jõhvi vald, Kohtla-Uueküla"), code_of)
		counted.add(address("75312", "Metsa", "Harju maakond, Rae vald, Koplimetsa küla"), code_of)
		counted.add(address("60534", "Nõlvaku tn 2", "Tartu maakond, Tartu vald, Raadi alev"), code_of)
		geo_ee.weighed(regions, municipalities, localities, counted, geo_ee.Census(stat(("007907968852M7", "....8852 Vahi alevik", 40))))
		self.assertEqual([localities[c]["population"] for c in ("2975", "3262", "3440", "6674")], [90, 60, 150, 3000])
		self.assertEqual([municipalities[c]["population"] for c in ("0824", "0250", "0653", "0796")], [0, 0, 0, 40])
		self.assertEqual([regions[c]["population"] for c in ("0081", "0045", "0037", "0079")], [0, 0, 0, 40])

	def test_a_locality_holds_at_least_one_person(self) -> None:
		regions, municipalities, localities, _ = geo_ee.units(UNITS)
		census = geo_ee.Census(stat(("003701988009M5", "....8009 Tabasalu alevik", 0)))
		geo_ee.weighed(regions, municipalities, localities, tally(address("76901", "Sarapuu tn 70", "Harju maakond, Harku vald, Tabasalu alevik")), census)
		self.assertEqual(localities["8009"]["population"], 1)


if __name__ == "__main__":
	unittest.main()
