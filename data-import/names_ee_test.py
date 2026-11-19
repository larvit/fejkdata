import unittest

import recalled_names
from geo_test import script

names_ee = script("names-ee")


class Recalled(unittest.TestCase):
    def test_a_ranked_name_weighs_100000_over_the_square_root_of_its_rank_and_the_rarer_ones_share_the_last(self) -> None:
        rows = names_ee.recalled("names", "Aa Bb Cc Dd Ee Ff Gg Hh Ii\n=\nXx Yy")
        self.assertEqual(
            rows,
            [("Aa", 100000), ("Bb", 70711), ("Cc", 57735), ("Dd", 50000), ("Ee", 44721), ("Ff", 40825), ("Gg", 37796), ("Hh", 35355), ("Ii", 33333), ("Xx", 33333), ("Yy", 33333)],
        )


class Lasts(unittest.TestCase):
    def rows(self) -> dict[tuple[str, str], int]:
        return {(str(r["name"]), str(r["sex"])): int(str(r["count"])) for r in names_ee.lasts()}

    def test_a_russian_name_in_ov_or_ev_takes_an_a_for_a_woman_and_the_others_stay_as_they_are(self) -> None:
        rows = self.rows()
        self.assertEqual(rows[("Ivanov", "m")], rows[("Ivanova", "f")])
        self.assertEqual(rows[("Aleksejev", "m")], rows[("Aleksejeva", "f")])
        self.assertEqual(rows[("Tamm", "m")], rows[("Tamm", "f")])
        self.assertNotIn(("Ivanov", "f"), rows)
        self.assertNotIn(("Tamma", "f"), rows)


class Split(unittest.TestCase):
    def test_the_ranked_names_come_apart_from_the_rarer_ones_at_the_equals_line(self) -> None:
        self.assertEqual(recalled_names.split("names", "Aa Bb\nCc\n=\nDd\nEe"), (["Aa", "Bb", "Cc"], ["Dd", "Ee"]))

    def test_a_name_listed_twice_or_fixed_and_listed_ends_the_run(self) -> None:
        with self.assertRaises(SystemExit) as raised:
            recalled_names.split("names", "Aa Bb\n=\nBb", ["Aa"])
        self.assertEqual(raised.exception.code, "names: listed twice: Aa, Bb")

    def test_a_name_that_is_not_letters_ends_the_run(self) -> None:
        with self.assertRaises(SystemExit) as raised:
            recalled_names.split("names", "Aa B1\n=\nCc")
        self.assertEqual(raised.exception.code, "names: 'B1' is not a name in letters")


if __name__ == "__main__":
    unittest.main()
