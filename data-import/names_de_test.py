import tempfile
import unittest
from pathlib import Path

from geo_test import script

names_de = script("names-de")


class Firsts(unittest.TestCase):
	def test_a_one_letter_part_before_an_apostrophe_is_a_name_written_with_the_ascii_apostrophe_and_a_lone_initial_goes(self) -> None:
		with tempfile.TemporaryDirectory() as cache:
			for year in names_de.FILES:
				Path(cache, f"munich-vornamen-{year}.csv").write_text("vorname,anzahl,geschlecht\nO’Brien,2,m\nD'Angelo,1,m\nO,9,m\n", encoding="utf-8")
			rows = [(str(r["name"]), str(r["sex"]), int(str(r["count"]))) for r in names_de.firsts(cache)]
		n = len(names_de.FILES)
		self.assertEqual(rows, [("D'Angelo", "m", n), ("O'Brien", "m", 2 * n)])
