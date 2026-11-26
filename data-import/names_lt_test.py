import unittest

from geo_test import script

names_lt = script("names-lt")

REGISTER = '{"_data":[{"nr":2,"vardas":"ANTANAS","viso_vardas":96104},{"nr":1,"vardas":"JONAS","viso_vardas":127263},{"nr":3,"vardas":"JONAS-PETRAS","viso_vardas":5}]}'


class Published(unittest.TestCase):
	def test_a_name_keeps_the_count_the_register_published_for_it_in_the_order_of_its_rank(self) -> None:
		self.assertEqual(names_lt.published(REGISTER), [("Jonas", 127263), ("Antanas", 96104), ("Jonas-Petras", 5)])


class Ranked(unittest.TestCase):
	def test_a_recalled_name_weighs_the_smallest_published_count_times_the_published_names_over_its_rank_and_the_rarer_ones_share_the_last_rank(self) -> None:
		rows = names_lt.ranked([("Aa", 100), ("Bb", 50)], "Cc Dd\n=\nXx Yy", "names")
		self.assertEqual(rows, [("Aa", 100), ("Bb", 50), ("Cc", 33), ("Dd", 25), ("Xx", 25), ("Yy", 25)])

	def test_a_recalled_name_that_is_published_too_is_refused(self) -> None:
		with self.assertRaises(SystemExit):
			names_lt.ranked([("Aa", 100)], "Aa", "names")


class Feminine(unittest.TestCase):
	def test_a_masculine_last_name_takes_the_unmarried_and_the_married_ending_lithuanian_gives_it(self) -> None:
		for masculine, feminine in [
			("Kazlauskas", ["Kazlauskaitė", "Kazlauskienė"]), ("Stankevičius", ["Stankevičiūtė", "Stankevičienė"]),
			("Butkus", ["Butkutė", "Butkienė"]), ("Kairys", ["Kairytė", "Kairienė"]), ("Jonaitis", ["Jonaitytė", "Jonaitienė"]),
			("Mažeika", ["Mažeikaitė", "Mažeikienė"]), ("Balčiūnas", ["Balčiūnaitė", "Balčiūnienė"]),
		]:
			with self.subTest(masculine=masculine):
				self.assertEqual(names_lt.feminine(masculine), feminine)

	def test_a_russian_or_polish_last_name_takes_the_feminine_ending_of_its_own_language(self) -> None:
		for masculine, feminine in [("Ivanov", ["Ivanova"]), ("Kozlov", ["Kozlova"]), ("Kowalski", ["Kowalska"]), ("Kovalenko", ["Kovalenko"])]:
			with self.subTest(masculine=masculine):
				self.assertEqual(names_lt.feminine(masculine), feminine)


class Lasts(unittest.TestCase):
	def test_a_recalled_last_name_has_a_row_under_m_and_its_count_shared_by_the_feminine_forms_under_f(self) -> None:
		rows = {(str(r["name"]), str(r["sex"])): int(str(r["count"])) for r in names_lt.lasts()}
		self.assertEqual(rows[("Kazlauskas", "m")], 100000)
		self.assertEqual(rows[("Kazlauskaitė", "f")], 50000)
		self.assertEqual(rows[("Kazlauskienė", "f")], 50000)
		self.assertEqual(rows[("Petrauskaitė", "f")], 35355)
		self.assertEqual(rows[("Kovalenko", "m")], rows[("Kovalenko", "f")])
		self.assertNotIn(("Kazlauskas", "f"), rows)


if __name__ == "__main__":
	unittest.main()
