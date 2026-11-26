import unittest

from geo_test import script

names_se = script("names-se")


class Cased(unittest.TestCase):
	def test_a_name_with_a_one_letter_part_before_an_apostrophe_comes_out_capitalised_with_the_ascii_apostrophe(self) -> None:
		self.assertEqual(names_se.cased("O'BRIEN"), "O'Brien")
		self.assertEqual(names_se.cased("O’BRIEN"), "O'Brien")


class Counted(unittest.TestCase):
	def test_a_one_letter_part_before_an_apostrophe_stays_written_with_the_ascii_apostrophe_and_a_lone_initial_goes(self) -> None:
		rows = [["Namn", "Antal"], ["O’BRIEN", "9"], ["D'ANGELO", "3"], ["O", "99"], ["SVENSSON", "7"]]
		self.assertEqual(names_se.counted(rows, 3), [("O'BRIEN", 9), ("SVENSSON", 7), ("D'ANGELO", 3)])

	def test_a_name_with_fewer_bearers_than_the_floor_goes(self) -> None:
		rows = [["Namn", "Antal"], ["SVENSSON", "7"], ["ENSAM", "2"], ["TRE", "3"]]
		self.assertEqual([n for n, _ in names_se.counted(rows, 3)], ["SVENSSON", "TRE"])
