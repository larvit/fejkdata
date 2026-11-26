import io
import tempfile
import unittest
import zipfile
from pathlib import Path

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


def cached(cache: str, name: str, member: str, body: str) -> None:
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w") as z:
        z.writestr(member, body)
    (Path(cache) / name).write_bytes(buf.getvalue())


class Firsts(unittest.TestCase):
    def test_a_name_whose_births_sum_to_fewer_than_the_floor_goes(self) -> None:
        body = "sexe;preusuel;annais;nombre\n1;JEAN;2000;4\n1;JEAN;2001;3\n1;LUC;2000;4\n2;MARIE;2000;9\n"
        with tempfile.TemporaryDirectory() as cache:
            cached(cache, "insee-prenoms-2023-nat.zip", "prenoms.csv", body)
            self.assertEqual(names_fr.firsts(cache, 5), [{"name": "Jean", "sex": "m", "count": 7}, {"name": "Marie", "sex": "f", "count": 9}])


class Lasts(unittest.TestCase):
    def test_a_name_whose_decades_sum_to_fewer_than_the_floor_goes(self) -> None:
        body = "NOM\t_1891_1900\t_1901_1910\nDUPONT\t10\t10\nMARTIN\t3\t3\nROY\t1\t3\n"
        with tempfile.TemporaryDirectory() as cache:
            cached(cache, "insee-noms2008nat.zip", "noms.txt", body)
            self.assertEqual(names_fr.lasts(cache, 5), [{"name": "Dupont", "count": 20}, {"name": "Martin", "count": 6}])
