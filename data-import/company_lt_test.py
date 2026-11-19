import json
import sys
import tempfile
import unittest
from collections import Counter
from pathlib import Path
from unittest import mock

from geo_test import script

company_lt = script("company-lt")

HEADER = "ja_kodas|ja_pavadinimas|adresas|ja_reg_data|form_kodas|form_pavadinimas|stat_kodas|stat_pavadinimas|stat_data_nuo|formavimo_data\n"
WORDS = company_lt.Words({"kazlauskas", "petraitis"}, {"jonas", "ona"})


def register_csv(rows: list[tuple[str, str]]) -> bytes:
    return (HEADER + "".join(f'{i}|"{name.replace(chr(34), chr(34) * 2)}"|Vilnius|1991-04-25|{form}|Form|0|Status|2025-03-04|2026-11-19\n' for i, (name, form) in enumerate(rows))).encode()


class Register(unittest.TestCase):
    def test_an_entity_of_a_drawn_form_counts_for_its_form_and_every_other_is_left_out(self) -> None:
        counts, names = company_lt.register(
            register_csv([
                ('UAB "Alfa"', "310"), ('MB Beta', "960"), ('AB "Gama"', "320"), ('ŽŪB "Delta"', "410"),
                ("Individuali įmonė", "810"), ('Asociacija "Epsilon"', "630"), ('VšĮ "Zeta"', "570"), ('Komanditinė ūkinė bendrija', "220"),
            ])
        )
        self.assertEqual(dict(counts), {"UAB": 1, "MB": 1, "AB": 1, "ŽŪB": 1})
        self.assertEqual(names, [("UAB", 'UAB "Alfa"'), ("MB", "MB Beta"), ("AB", 'AB "Gama"'), ("ŽŪB", 'ŽŪB "Delta"')])

    def test_a_name_keeps_its_quotes_and_a_pipe_inside_them(self) -> None:
        _, names = company_lt.register(register_csv([('UAB "Alfa | Beta"', "310")]))
        self.assertEqual(names, [("UAB", 'UAB "Alfa | Beta"')])


class Placed(unittest.TestCase):
    def test_a_name_places_its_form_before_or_after_it_and_quoted_or_bare(self) -> None:
        self.assertEqual(company_lt.placed("UAB", 'UAB "Alfa Statyba"'), ("before", True, ["Alfa", "Statyba"]))
        self.assertEqual(company_lt.placed("MB", "MB Alfa"), ("before", False, ["Alfa"]))
        self.assertEqual(company_lt.placed("UAB", "Alfa, UAB"), ("after", False, ["Alfa"]))
        self.assertEqual(company_lt.placed("UAB", '"Alfa" UAB'), ("after", True, ["Alfa"]))

    def test_the_form_counts_as_its_abbreviation_or_its_words_in_any_case_and_with_or_without_a_dot(self) -> None:
        self.assertEqual(company_lt.placed("UAB", 'Uždaroji akcinė bendrovė "Alfa"'), ("before", True, ["Alfa"]))
        self.assertEqual(company_lt.placed("UAB", 'UŽDAROJI AKCINĖ BENDROVĖ "ALFA"'), ("before", True, ["ALFA"]))
        self.assertEqual(company_lt.placed("MB", 'Mažoji bendrija "Alfa"'), ("before", True, ["Alfa"]))
        self.assertEqual(company_lt.placed("AB", "Alfa akcinė bendrovė"), ("after", False, ["Alfa"]))
        self.assertEqual(company_lt.placed("ŽŪB", 'Žemės ūkio bendrovė "Alfa"'), ("before", True, ["Alfa"]))
        self.assertEqual(company_lt.placed("ŽŪB", "žūb Alfa"), ("before", False, ["Alfa"]))
        self.assertEqual(company_lt.placed("AB", "Alfa, AB."), ("after", False, ["Alfa"]))

    def test_a_name_carrying_another_form_or_none_or_only_its_form_or_the_form_inside_is_not_placed(self) -> None:
        self.assertIsNone(company_lt.placed("UAB", 'AB "Alfa"'))
        self.assertIsNone(company_lt.placed("UAB", "Alfa Statyba"))
        self.assertIsNone(company_lt.placed("UAB", "UAB"))
        self.assertIsNone(company_lt.placed("UAB", "Alfa UAB Beta"))


class Shapes(unittest.TestCase):
    def test_a_legal_word_is_left_out_of_a_shape_with_or_without_a_dot(self) -> None:
        shapes, _ = company_lt.count_names([("UAB", "Uab. Zorba UAB"), ("UAB", "Zorba UAB")], WORDS, set())
        self.assertEqual(dict(shapes["UAB"]), {("X",): 2})

    def test_a_name_of_legal_words_alone_is_not_counted(self) -> None:
        shapes, places = company_lt.count_names([("UAB", "Uždaroji akcinė bendrovė UAB")], WORDS, set())
        self.assertEqual((dict(shapes["UAB"]), dict(places["UAB"])), ({}, {}))

    def test_the_vocabulary_names_the_connecting_words_of_its_language(self) -> None:
        shapes, _ = company_lt.count_names([("UAB", "Kazlauskas ir Ona UAB")], WORDS, set())
        self.assertEqual(dict(shapes["UAB"]), {("S", "ir", "F"): 1})

    def test_a_name_counts_its_shape_and_where_it_places_its_form(self) -> None:
        names = [("UAB", 'UAB "Zorba"'), ("UAB", "Zorba UAB"), ("UAB", "Zorba"), ("AB", 'AB "Jonas Kazlauskas"')]
        shapes, places = company_lt.count_names(names, WORDS, set())
        self.assertEqual(dict(shapes["UAB"]), {("X",): 2})
        self.assertEqual(dict(places["UAB"]), {("before", True): 1, ("after", False): 1})
        self.assertEqual(dict(shapes["AB"]), {("FS",): 1})
        self.assertEqual(dict(places["AB"]), {("before", True): 1})


class Bare(unittest.TestCase):
    def test_a_name_that_places_its_form_loses_the_form_and_its_quotes_and_any_other_name_is_dropped(self) -> None:
        names = [("UAB", 'Uždaroji akcinė bendrovė "Alfa Statyba"'), ("AB", "Beta AB"), ("MB", "Gama"), ("UAB", "UAB")]
        self.assertEqual(company_lt.bare(names), [("UAB", "Alfa Statyba"), ("AB", "Beta")])


class Coined(unittest.TestCase):
    def coined(self, word: str, names: int, ends: int, trade: set[str] | None = None) -> dict[str, int]:
        rows = [("UAB", f"Pluss {word}" if i < ends else f"{word} Pluss") for i in range(names)]
        found: dict[str, int] = company_lt.coined(company_lt.word_usage(rows, lambda w: False, company_lt.VOCABULARY), trade or set())
        return found

    def test_a_coined_word_keeps_each_lithuanian_letter(self) -> None:
        for letter in "ąčęėįšųūž":
            self.assertIn(f"Zorb{letter}", self.coined(f"Zorb{letter}", 40, 2))

    def test_a_coined_word_spelled_in_capitals_with_a_lithuanian_vowel_is_spelled_in_lower_case(self) -> None:
        for vowel in "ąęėįųū":
            self.assertIn(vowel.upper() + vowel * 3, self.coined(vowel.upper() * 4, 40, 2))


class Trade(unittest.TestCase):
    def test_a_trade_word_of_one_form_alone_counts_since_uab_names_are_nearly_every_name(self) -> None:
        rows = [("UAB", "Alfa Statyba" if i < 40 else "Statyba Alfa") for i in range(100)]
        trade = company_lt.trade_words(company_lt.word_usage(rows, lambda w: False, company_lt.VOCABULARY), company_lt.VOCABULARY)
        self.assertEqual(trade["statyba"], ("Statyba", 100))

    def test_a_trade_word_has_three_letters_or_more(self) -> None:
        rows = [("UAB", "Art Ab" if i < 50 else "Ab Art") for i in range(100)]
        trade = company_lt.trade_words(company_lt.word_usage(rows, lambda w: False, company_lt.VOCABULARY), company_lt.VOCABULARY)
        self.assertEqual(sorted(trade), ["art"])


SHAPES = {"UAB": {("X",): 200, ("T", "X"): 1}, "MB": {("FS",): 10}, "AB": {("X",): 1}, "ŽŪB": {("X",): 1}}
PLACES = {
    "UAB": Counter({("before", True): 6, ("before", False): 4, ("after", False): 3, ("after", True): 1}),
    "MB": Counter({("before", False): 2}),
    "AB": Counter({("after", True): 1}),
    "ŽŪB": Counter({("before", True): 1}),
}


class Output(unittest.TestCase):
    def company(self) -> dict[str, dict[str, object]]:
        data: dict[str, dict[str, object]] = json.loads(company_lt.company_json({"MB": 50, "UAB": 80, "AB": 2, "ŽŪB": 1}, SHAPES, PLACES))
        return data

    def test_the_forms_weigh_their_entities_and_a_form_without_an_ascii_path_takes_one(self) -> None:
        registered = self.company()["registered"]
        assert isinstance(registered, list)
        self.assertEqual([(r["format"], r["legal"], r["weight"]) for r in registered], [
            ("{.company.uab}", "UAB", 80), ("{.company.mb}", "MB", 50), ("{.company.ab}", "AB", 2), ("{.company.zub}", "ŽŪB", 1),
        ])

    def test_a_form_places_its_abbreviation_and_its_quotes_as_the_names_of_that_form_do_and_ships_its_shapes_of_share_min_or_more(self) -> None:
        self.assertEqual(self.company()["uab"], {
            "format": "{name as uab_name}{placed}",
            "name": [{"format": "{.company-coined}", "weight": 200}],
            "hidden": {"placed": [
                {"format": 'UAB "{uab_name}"', "weight": 6}, {"format": "UAB {uab_name}", "weight": 4},
                {"format": "{uab_name} UAB", "weight": 3}, {"format": '"{uab_name}" UAB', "weight": 1},
            ]},
        })

    def test_a_form_whose_names_never_place_it_one_way_does_not_draw_that_way(self) -> None:
        self.assertEqual(self.company()["zub"]["hidden"], {"placed": [{"format": 'ŽŪB "{zub_name}"', "weight": 1}]})

    def test_an_mb_draws_the_person_its_names_carry(self) -> None:
        self.assertEqual(self.company()["mb"]["name"], [{"format": "{.person as p}{p.first} {p.last}", "weight": 10}])


class Main(unittest.TestCase):
    def test_the_script_reads_a_register_file_and_writes_the_company_files_from_its_names_and_the_shipped_tables(self) -> None:
        rows = [('UAB "Alfa Statyba"', "310")] * 100 + [('UAB "Petraitis"', "310")] * 60 + [('UAB "Zorba"', "310")] * 40
        with tempfile.TemporaryDirectory() as folder:
            register = Path(folder) / "JAR_IREGISTRUOTI.csv"
            register.write_bytes(register_csv(rows))
            with mock.patch.object(sys, "argv", ["company-lt.py", "--register", str(register), "--out", folder]):
                company_lt.main()
            company = json.loads((Path(folder) / "company.json").read_text(encoding="utf-8"))
            self.assertEqual(company["registered"][0], {"format": "{.company.uab}", "legal": "UAB", "weight": 200})
            self.assertEqual(
                company["uab"]["name"],
                [{"format": "{.company-coined} {.company-word}", "weight": 100}, {"format": "{.last-name}", "weight": 60}, {"format": "{.company-coined}", "weight": 40}],
            )
            self.assertEqual((Path(folder) / "company-coined.tsv").read_text(encoding="utf-8").splitlines(), ["word\tcount", "Zorba\t40"])
            self.assertEqual((Path(folder) / "company-word.tsv").read_text(encoding="utf-8").splitlines(), ["word\tcount", "Statyba\t100"])


if __name__ == "__main__":
    unittest.main()
