import collections
import unittest
from collections.abc import Iterable

from geo_test import script

names_us = script("names-us")


class Firsts(unittest.TestCase):
	def test_a_name_given_to_fewer_than_the_floor_under_a_sex_goes_from_that_sex_only(self) -> None:
		counts = collections.Counter({("Anna", "f"): 9, ("Anna", "m"): 2, ("Bo", "m"): 5})
		self.assertEqual(names_us.firsts(counts, 5), [{"name": "Anna", "sex": "f", "count": 9}, {"name": "Bo", "sex": "m", "count": 5}])


def nobody(labels: Iterable[str], language: str, country: str, cache: str) -> dict[str, int]:
	return {}


class Lasts(unittest.TestCase):
	def test_a_surname_under_the_floor_or_with_a_suppressed_count_goes(self) -> None:
		rows = [["OBRIEN", "1", "150"], ["MCDONALD", "2", "99"], ["SMITH", "3", "2376206"], ["ALL OTHER NAMES", "4", "5000"], ["ZED", "5", "(S)"]]
		self.assertEqual(names_us.lasts(rows, 100, nobody, "cache"), [{"name": "Obrien", "count": 150}, {"name": "Smith", "count": 2376206}])

	def test_a_surname_is_written_as_most_american_bearers_write_it_where_wikidata_settles_it(self) -> None:
		asked: list[tuple[str, str]] = []

		def bearers(labels: Iterable[str], language: str, country: str, cache: str) -> dict[str, int]:
			asked.append((language, country))
			return {"Obrien": 6, "O'Brien": 2320, "Dangelo": 2, "D'Angelo": 2, "Delacruz": 1, "De La Cruz": 4, "De la Cruz": 4}

		rows = [["OBRIEN", "1", "150"], ["DANGELO", "2", "150"], ["DELACRUZ", "3", "150"], ["SMITH", "4", "150"]]
		self.assertEqual(
			names_us.lasts(rows, 100, bearers, "cache"),
			[{"name": "Dangelo", "count": 150}, {"name": "Delacruz", "count": 150}, {"name": "O'Brien", "count": 150}, {"name": "Smith", "count": 150}],
		)
		self.assertEqual(asked, [("en", "Q30")])


class Particled(unittest.TestCase):
	def test_a_name_is_offered_with_an_apostrophe_a_space_or_a_capital_after_each_particle_it_starts_with(self) -> None:
		self.assertIn("O'Brien", names_us.particled("Obrien"))
		self.assertIn("D'Angelo", names_us.particled("Dangelo"))
		self.assertTrue({"De La Cruz", "De la Cruz", "DeLaCruz"} <= set(names_us.particled("Delacruz")))
		self.assertIn("St John", names_us.particled("Stjohn"))
		self.assertIn("VanDyke", names_us.particled("Vandyke"))
		self.assertIn("MacDonald", names_us.particled("Macdonald"))

	def test_a_name_with_no_particle_or_too_little_after_it_is_offered_nothing(self) -> None:
		self.assertEqual(names_us.particled("Smith"), [])
		self.assertEqual(names_us.particled("Dean"), [])
