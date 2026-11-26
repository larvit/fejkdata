import collections
import unittest

from geo_test import script

names_us = script("names-us")


class Firsts(unittest.TestCase):
    def test_a_name_given_to_fewer_than_the_floor_under_a_sex_goes_from_that_sex_only(self) -> None:
        counts = collections.Counter({("Anna", "f"): 9, ("Anna", "m"): 2, ("Bo", "m"): 5})
        self.assertEqual(names_us.firsts(counts, 5), [{"name": "Anna", "sex": "f", "count": 9}, {"name": "Bo", "sex": "m", "count": 5}])


class Lasts(unittest.TestCase):
    def test_a_surname_under_the_floor_or_with_a_suppressed_count_goes(self) -> None:
        rows = [["OBRIEN", "1", "150"], ["MCDONALD", "2", "99"], ["SMITH", "3", "2376206"], ["ALL OTHER NAMES", "4", "5"], ["ZED", "5", "(S)"]]
        self.assertEqual(names_us.lasts(rows, 100), [{"name": "Obrien", "count": 150}, {"name": "Smith", "count": 2376206}])
