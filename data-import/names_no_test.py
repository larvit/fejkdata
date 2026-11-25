import unittest

from geo_test import script

names_no = script("names-no")


def cube(labels: list[str], values: list[int]) -> dict[str, object]:
    codes = [f"c{i}" for i in range(len(labels))]
    return {"id": ["Etternavn", "ContentsCode", "Tid"], "size": [len(codes), 1, 1], "value": values, "dimension": {"Etternavn": {"category": {"index": codes, "label": dict(zip(codes, labels))}}}}


class Counted(unittest.TestCase):
    def test_a_one_letter_part_before_an_apostrophe_stays_written_with_the_ascii_apostrophe_and_a_lone_initial_goes(self) -> None:
        got = names_no.counted(cube(["O’Brien", "D'Angelo", "O", "Hansen"], [5, 3, 99, 7]), "Etternavn")
        self.assertEqual(got, [("c0", "O'Brien", 5), ("c1", "D'Angelo", 3), ("c3", "Hansen", 7)])
