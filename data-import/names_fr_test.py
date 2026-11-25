import unittest

from geo_test import script

names_fr = script("names-fr")


class Capitalised(unittest.TestCase):
    def test_a_name_with_a_one_letter_part_before_an_apostrophe_comes_out_as_it_is_written_in_either_apostrophe(self) -> None:
        for source in ["O'BRIEN", "O’BRIEN"]:
            self.assertEqual(names_fr.capitalised(source), "O'Brien")
        self.assertEqual(names_fr.capitalised("N'DIAYE"), "N'Diaye")
        self.assertEqual(names_fr.capitalised("PRUD'HOMME"), "Prud'Homme")


class Surname(unittest.TestCase):
    def test_a_particle_stays_in_lower_case_beside_an_apostrophe_name(self) -> None:
        self.assertEqual(names_fr.surname("DE L'ISLE"), "de L'Isle")
        self.assertEqual(names_fr.surname("D’ANGELO"), "D'Angelo")
