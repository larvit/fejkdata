import unittest

import tsv


class Name(unittest.TestCase):
    def test_a_one_letter_part_before_an_apostrophe_is_a_name_in_either_apostrophe(self) -> None:
        for name in ["O'BRIEN", "O’BRIEN", "D'ANGELO", "N'DIAYE", "O'Brien", "Le Gall", "Anne-Marie", "De L'Isle"]:
            self.assertTrue(tsv.NAME.match(name), name)

    def test_a_lone_initial_or_a_one_letter_part_with_no_apostrophe_is_no_name(self) -> None:
        for name in ["O", "A B", "Anna B", "J-P", "O'", "'BRIEN", "O'B", "Jean P", "O''BRIEN"]:
            self.assertFalse(tsv.NAME.match(name), name)


class Straight(unittest.TestCase):
    def test_a_typographic_apostrophe_becomes_the_ascii_one(self) -> None:
        self.assertEqual(tsv.straight("O’Brien D'Angelo"), "O'Brien D'Angelo")
