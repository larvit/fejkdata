import unittest

from geo_test import script

names_fr = script("names-fr")


class Capitalised(unittest.TestCase):
    def test_each_part_of_a_name_starts_with_a_capital_and_an_apostrophe_comes_out_ascii_whichever_it_came_in(self) -> None:
        for source, written in [("O'BRIEN", "O'Brien"), ("O’BRIEN", "O'Brien"), ("N'DIAYE", "N'Diaye"), ("PRUD'HOMME", "Prud'Homme")]:
            with self.subTest(source):
                self.assertEqual(names_fr.capitalised(source), written)


class Surname(unittest.TestCase):
    def test_a_particle_stays_in_lower_case_before_a_part_with_an_apostrophe_and_the_apostrophe_comes_out_ascii(self) -> None:
        self.assertEqual(names_fr.surname("DE L'ISLE"), "de L'Isle")
        self.assertEqual(names_fr.surname("D’ANGELO"), "D'Angelo")
