import unittest

from geo_test import script

names_fi = script("names-fi")


class Counted(unittest.TestCase):
	def test_a_one_letter_part_before_an_apostrophe_stays_written_with_the_ascii_apostrophe_and_a_lone_initial_goes(self) -> None:
		rows = [["Nimi", "Lkm"], ["O’Brien", "5"], ["D'Angelo", "3"], ["O", "99"], ["Virtanen", "7"]]
		self.assertEqual(names_fi.counted(rows, 3), [("Virtanen", 7), ("O'Brien", 5), ("D'Angelo", 3)])

	def test_a_name_with_fewer_bearers_than_the_floor_goes(self) -> None:
		rows = [["Nimi", "Lkm"], ["Virtanen", "7"], ["Yksi", "2"], ["Kolme", "3"]]
		self.assertEqual([n for n, _ in names_fi.counted(rows, 3)], ["Virtanen", "Kolme"])
