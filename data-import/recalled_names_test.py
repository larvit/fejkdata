import unittest

import recalled_names


class Split(unittest.TestCase):
	def test_a_one_letter_part_before_an_apostrophe_is_a_name_written_with_the_ascii_apostrophe(self) -> None:
		self.assertEqual(recalled_names.split("names", "O’Brien D'Angelo\n=\nNg"), (["O'Brien", "D'Angelo"], ["Ng"]))

	def test_a_lone_initial_is_refused(self) -> None:
		with self.assertRaises(SystemExit):
			recalled_names.split("names", "Aa O")
