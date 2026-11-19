import unittest

from geo_test import script

names_lv = script("names-lv")

TIED = '\ufeffVardi,Dzimums,Skaits\n"BAIBA","SIEVIETE",3\n"ZANE","SIEVIETE",2\n"AIGA","SIEVIETE",2\n"ANDRIS","VĪRIETIS",1\n'
PMLP = '﻿Vardi,Dzimums,Skaits\n"",SIEVIETE,1\n"-","VĪRIETIS",4\n"JĀNIS","VĪRIETIS",5\n"JĀNIS ANDRIS","VĪRIETIS",3\n"ANNA-MARIJA","SIEVIETE",2\n"JĀNIS","SIEVIETE",1\n"ZIGMARS","VĪRIETIS",1\n"ALEKSS","NAV NORĀDĪTS",9\n'


class Firsts(unittest.TestCase):
    def rows(self, top: int) -> list[tuple[str, str, int]]:
        return [(str(r["name"]), str(r["sex"]), int(str(r["count"]))) for r in names_lv.firsts(PMLP, top)]

    def test_a_person_counts_once_under_the_first_of_their_given_names_and_a_label_that_is_no_name_or_a_person_of_no_stated_sex_is_dropped(self) -> None:
        self.assertEqual(self.rows(10), [("Anna-Marija", "f", 2), ("Jānis", "f", 1), ("Jānis", "m", 8), ("Zigmars", "m", 1)])

    def test_each_sex_keeps_its_top_names_by_bearers(self) -> None:
        self.assertEqual(self.rows(1), [("Anna-Marija", "f", 2), ("Jānis", "m", 8)])

    def test_a_tie_at_the_cut_keeps_the_name_that_sorts_first_and_the_rows_sort_by_name_then_sex(self) -> None:
        rows = [(str(r["name"]), str(r["sex"]), int(str(r["count"]))) for r in names_lv.firsts(TIED, 2)]
        self.assertEqual(rows, [("Aiga", "f", 2), ("Andris", "m", 1), ("Baiba", "f", 3)])


class Feminine(unittest.TestCase):
    def test_a_masculine_last_name_takes_the_feminine_ending_latvian_gives_it(self) -> None:
        for masculine, feminine in [
            ("Bērziņš", "Bērziņa"), ("Balodis", "Balode"), ("Dombrovskis", "Dombrovska"), ("Ozols", "Ozola"), ("Jansons", "Jansone"),
            ("Millers", "Millere"), ("Zvejnieks", "Zvejniece"), ("Kalējs", "Kalēja"), ("Ivanovs", "Ivanova"), ("Liepa", "Liepa"),
            ("Jaunzems", "Jaunzeme"), ("Bondars", "Bondare"),
        ]:
            with self.subTest(masculine=masculine):
                self.assertEqual(names_lv.feminine(masculine), feminine)


class Lasts(unittest.TestCase):
    def test_a_recalled_last_name_has_a_row_under_each_sex_with_one_count(self) -> None:
        rows = {(str(r["name"]), str(r["sex"])): int(str(r["count"])) for r in names_lv.lasts()}
        self.assertEqual(rows[("Bērziņš", "m")], 100000)
        self.assertEqual(rows[("Bērziņa", "f")], 100000)
        self.assertEqual(rows[("Liepa", "m")], rows[("Liepa", "f")])
        self.assertNotIn(("Bērziņš", "f"), rows)


if __name__ == "__main__":
    unittest.main()
