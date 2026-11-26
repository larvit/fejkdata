import unittest

import tsv


class Name(unittest.TestCase):
    def test_a_one_letter_part_before_an_apostrophe_is_a_name_with_either_apostrophe(self) -> None:
        for name in ["O'BRIEN", "O’BRIEN", "D'ANGELO", "N'DIAYE", "O'Brien", "Le Gall", "Anne-Marie", "De L'Isle", "DE L’ISLE", "PRUD’HOMME"]:
            with self.subTest(name):
                self.assertTrue(tsv.NAME.match(name))

    def test_a_lone_initial_or_a_one_letter_part_not_directly_before_an_apostrophe_is_no_name(self) -> None:
        for name in ["O", "A B", "Anna B", "J-P", "O'", "'BRIEN", "O'B", "Jean P", "O''BRIEN"]:
            with self.subTest(name):
                self.assertFalse(tsv.NAME.match(name))


class Common(unittest.TestCase):
    def test_a_name_under_the_floor_goes_and_the_rest_come_commonest_first_then_by_name(self) -> None:
        rows = [("Anna", 7), ("Berit", 5), ("Cilla", 4), ("Dora", 7)]
        self.assertEqual(tsv.common(rows, 5), [("Anna", 7), ("Dora", 7), ("Berit", 5)])

    def test_the_default_floor_is_five_bearers(self) -> None:
        self.assertEqual(tsv.MIN_BEARERS, 5)


class Straight(unittest.TestCase):
    def test_a_typographic_apostrophe_becomes_the_ascii_one(self) -> None:
        self.assertEqual(tsv.straight("O’Brien D'Angelo"), "O'Brien D'Angelo")
