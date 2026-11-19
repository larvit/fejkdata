import json
import sys
import tempfile
import unittest
from collections import Counter
from pathlib import Path
from unittest import mock

from geo_test import script

company_lv = script("company-lv")

HEADER = "regcode;name;regtype;type;terminated\n"
WORDS = company_lv.Words({"ozols", "kalniņš"}, {"jānis", "anna"})


def register_csv(rows: list[tuple[str, str, str, str]]) -> bytes:
    return ("﻿" + HEADER + "".join(f'{i};"{name.replace(chr(34), chr(34) * 2)}";{regtype};{form};{terminated}\n' for i, (name, regtype, form, terminated) in enumerate(rows))).encode()


class Register(unittest.TestCase):
    def test_a_merchant_of_the_commercial_register_not_terminated_counts_for_its_form_and_every_other_is_left_out(self) -> None:
        counts, names = company_lv.register(
            register_csv([
                ('SIA "Alfa"', "K", "SIA", ""), ('SIA "Beta"', "K", "SIA", "2019-01-14"), ('IK "Gamma"', "K", "IK", ""),
                ('AS "Delta"', "K", "AS", ""), ('PS "Epsilon"', "K", "PS", ""), ('KS "Zeta"', "K", "KS", ""),
                ('Biedrība "Eta"', "B", "BDR", ""), ('Teta SIA filiāle', "K", "FIL", ""), ('Ārvalsts "Jota"', "K", "AKF", ""),
                ('Zemnieku saimniecība "Kapa"', "Z", "ZEM", ""), ('SIA "Lambda"', "Z", "SIA", ""),
            ])
        )
        self.assertEqual(dict(counts), {"SIA": 1, "IK": 1, "AS": 1, "PS": 1, "KS": 1})
        self.assertEqual(names, [("SIA", 'SIA "Alfa"'), ("IK", 'IK "Gamma"'), ("AS", 'AS "Delta"'), ("PS", 'PS "Epsilon"'), ("KS", 'KS "Zeta"')])

    def test_a_name_keeps_its_quotes_and_a_semicolon_inside_them(self) -> None:
        _, names = company_lv.register(register_csv([('SIA "Alfa; Beta"', "K", "SIA", "")]))
        self.assertEqual(names, [("SIA", 'SIA "Alfa; Beta"')])


class Placed(unittest.TestCase):
    def test_a_name_places_its_form_before_or_after_it_and_quoted_or_bare(self) -> None:
        self.assertEqual(company_lv.placed("SIA", 'SIA "Alfa Būve"'), ("before", True, ["Alfa", "Būve"]))
        self.assertEqual(company_lv.placed("SIA", "SIA Alfa"), ("before", False, ["Alfa"]))
        self.assertEqual(company_lv.placed("SIA", "Alfa SIA"), ("after", False, ["Alfa"]))
        self.assertEqual(company_lv.placed("SIA", '"Alfa" SIA'), ("after", True, ["Alfa"]))

    def test_the_form_counts_as_its_abbreviation_or_its_words_in_any_case_and_with_or_without_a_dot(self) -> None:
        self.assertEqual(company_lv.placed("SIA", 'Sabiedrība ar ierobežotu atbildību "Alfa"'), ("before", True, ["Alfa"]))
        self.assertEqual(company_lv.placed("SIA", 'SABIEDRĪBA AR IEROBEŽOTU ATBILDĪBU "ALFA"'), ("before", True, ["ALFA"]))
        self.assertEqual(company_lv.placed("AS", "Alfa akciju sabiedrība"), ("after", False, ["Alfa"]))
        self.assertEqual(company_lv.placed("IK", "Individuālais komersants Alfa"), ("before", False, ["Alfa"]))
        self.assertEqual(company_lv.placed("PS", 'Pilnsabiedrība "Alfa"'), ("before", True, ["Alfa"]))
        self.assertEqual(company_lv.placed("KS", 'komandītsabiedrība "Alfa"'), ("before", True, ["Alfa"]))
        self.assertEqual(company_lv.placed("AS", "Alfa, A.S."), None)
        self.assertEqual(company_lv.placed("AS", "Alfa, AS."), ("after", False, ["Alfa"]))

    def test_a_name_carrying_another_form_or_none_or_only_its_form_or_the_form_inside_is_not_placed(self) -> None:
        self.assertIsNone(company_lv.placed("SIA", 'AS "Alfa"'))
        self.assertIsNone(company_lv.placed("SIA", "Alfa Būve"))
        self.assertIsNone(company_lv.placed("SIA", "SIA"))
        self.assertIsNone(company_lv.placed("SIA", "Alfa SIA Beta"))


class Shapes(unittest.TestCase):
    def shape(self, name: str) -> tuple[str, ...]:
        shape: tuple[str, ...] = company_lv.shape_of(name.split(), WORDS, {"būve"})
        return shape

    def test_each_word_is_a_name_a_number_a_connecting_word_a_trade_word_or_coined(self) -> None:
        self.assertEqual(self.shape("Ozols un Anna 12 Būve Zorba &"), ("S", "un", "F", "N", "T", "X", "&"))

    def test_a_first_name_before_a_last_name_is_one_person(self) -> None:
        self.assertEqual(self.shape("Jānis Ozols Būve"), ("FS", "T"))
        self.assertEqual(self.shape("Jānis Ozols & Anna Kalniņš"), ("FS", "&", "FS"))

    def test_a_name_is_compared_in_lower_case_and_a_word_that_is_both_names_is_a_last_name_at_the_end(self) -> None:
        words = company_lv.Words({"ozols", "anna"}, {"anna"})
        self.assertEqual(company_lv.kind("ANNA", words, False), "F")
        self.assertEqual(company_lv.kind("Anna", words, True), "S")
        self.assertEqual(company_lv.kind("OZOLS", words, False), "S")
        self.assertIsNone(company_lv.kind("Zorba", words, True))

    def test_a_legal_word_is_left_out_of_a_shape_with_or_without_a_dot(self) -> None:
        shapes, _ = company_lv.count_names([("SIA", "Sia. Zorba SIA"), ("SIA", "Zorba SIA")], WORDS, set())
        self.assertEqual(dict(shapes["SIA"]), {("X",): 2})

    def test_a_name_of_legal_words_alone_is_not_counted(self) -> None:
        shapes, places = company_lv.count_names([("SIA", "Sabiedrība ar ierobežotu atbildību SIA")], WORDS, set())
        self.assertEqual((dict(shapes["SIA"]), dict(places["SIA"])), ({}, {}))

    def test_a_word_that_is_both_names_is_a_first_name_inside_a_name_and_a_last_name_at_its_end_in_a_shape(self) -> None:
        words = company_lv.Words({"ozols", "anna"}, {"anna", "jānis"})
        self.assertEqual(company_lv.shape_of(["Anna", "Zorba"], words, set()), ("F", "X"))
        self.assertEqual(company_lv.shape_of(["Zorba", "Anna"], words, set()), ("X", "S"))
        self.assertEqual(company_lv.shape_of(["Jānis", "Anna"], words, set()), ("FS",))

    def test_a_name_counts_its_shape_and_where_it_places_its_form(self) -> None:
        names = [("SIA", 'SIA "Zorba"'), ("SIA", "Zorba SIA"), ("SIA", "Zorba"), ("AS", 'AS "Jānis Ozols"')]
        shapes, places = company_lv.count_names(names, WORDS, set())
        self.assertEqual(dict(shapes["SIA"]), {("X",): 2})
        self.assertEqual(dict(places["SIA"]), {("before", True): 1, ("after", False): 1})
        self.assertEqual(dict(shapes["AS"]), {("FS",): 1})
        self.assertEqual(dict(places["AS"]), {("before", True): 1})

    def test_a_shape_opening_with_a_number_or_a_connecting_word_does_not_ship(self) -> None:
        shapes = Counter({("X",): 50, ("N", "X"): 30, ("&", "X"): 10, ("un", "T"): 10})
        self.assertEqual(company_lv.shipped(shapes), [(("X",), 50)])

    def test_each_person_of_a_shape_is_drawn_once_under_a_name_of_its_own(self) -> None:
        self.assertEqual(
            company_lv.slots(("FS", "&", "FS")),
            "{.person as p}{p.first} {p.last} & {.person as q}{q.first} {q.last}",
        )
        self.assertEqual(company_lv.slots(("S", "un", "F", "N", "T", "X")), "{.last-name} un {.first-name} {int(1,30)} {.company-word} {.company-coined}")


class Bare(unittest.TestCase):
    def test_a_name_that_places_its_form_loses_the_form_and_its_quotes_and_any_other_name_is_dropped(self) -> None:
        names = [("SIA", 'Sabiedrība ar ierobežotu atbildību "Alfa Būve"'), ("AS", "Beta AS"), ("IK", "Gamma"), ("SIA", "SIA")]
        self.assertEqual(company_lv.bare(names), [("SIA", "Alfa Būve"), ("AS", "Beta")])


class Coined(unittest.TestCase):
    def coined(self, word: str, names: int, ends: int, trade: set[str] | None = None) -> dict[str, int]:
        rows = [("SIA", f"Pluss {word}" if i < ends else f"{word} Pluss") for i in range(names)]
        found: dict[str, int] = company_lv.coined(company_lv.word_usage(rows, lambda w: False, company_lv.VOCABULARY), trade or set())
        return found

    def test_a_coined_word_is_carried_by_COINED_MIN_names_and_ends_at_least_COINED_ENDS_MIN_of_them(self) -> None:
        self.assertEqual(self.coined("Zorba", 40, 2)["Zorba"], 40)
        self.assertNotIn("Zorba", self.coined("Zorba", 39, 2))
        self.assertNotIn("Zorba", self.coined("Zorba", 40, 1))

    def test_a_coined_word_has_four_letters_or_more(self) -> None:
        self.assertNotIn("Zor", self.coined("Zor", 40, 2))
        self.assertIn("Arhitektūra", self.coined("Arhitektūra", 40, 2))

    def test_a_coined_word_of_four_letters_counts(self) -> None:
        self.assertIn("Zorb", self.coined("Zorb", 40, 2))

    def test_a_coined_word_keeps_each_latvian_letter(self) -> None:
        for letter in "āčēģīķļņšūž":
            self.assertIn(f"Zorb{letter}", self.coined(f"Zorb{letter}", 40, 2))

    def test_a_coined_word_spelled_in_capitals_with_a_latvian_vowel_is_spelled_in_lower_case(self) -> None:
        for vowel in "āēīū":
            self.assertIn(vowel.upper() + vowel * 3, self.coined(vowel.upper() * 4, 40, 2))

    def test_a_trade_word_is_not_coined(self) -> None:
        self.assertNotIn("Zorba", self.coined("Zorba", 40, 2, {"zorba"}))


class Trade(unittest.TestCase):
    def test_a_trade_word_of_one_form_alone_counts_since_sia_names_are_nearly_every_name(self) -> None:
        rows = [("SIA", "Alfa Būve" if i < 40 else "Būve Alfa") for i in range(100)]
        trade = company_lv.trade_words(company_lv.word_usage(rows, lambda w: False, company_lv.VOCABULARY), company_lv.VOCABULARY)
        self.assertEqual(trade["būve"], ("Būve", 100))

    def test_a_trade_word_has_three_letters_or_more(self) -> None:
        rows = [("SIA", "Art Ab" if i < 50 else "Ab Art") for i in range(100)]
        trade = company_lv.trade_words(company_lv.word_usage(rows, lambda w: False, company_lv.VOCABULARY), company_lv.VOCABULARY)
        self.assertEqual(sorted(trade), ["art"])


SHAPES = {"SIA": {("X",): 200, ("T", "X"): 1}, "IK": {("FS",): 10}, "AS": {("X",): 1}, "PS": {("X",): 1}, "KS": {("X",): 1}}
PLACES = {
    "SIA": Counter({("before", True): 6, ("before", False): 4, ("after", False): 3, ("after", True): 1}),
    "IK": Counter({("before", False): 2}),
    "AS": Counter({("after", True): 1}),
    "PS": Counter({("before", True): 1}),
    "KS": Counter({("after", False): 1}),
}


class Output(unittest.TestCase):
    def company(self) -> dict[str, dict[str, object]]:
        data: dict[str, dict[str, object]] = json.loads(company_lv.company_json({"IK": 50, "SIA": 80, "AS": 20, "PS": 2, "KS": 1}, SHAPES, PLACES))
        return data

    def test_the_forms_weigh_their_merchants(self) -> None:
        registered = self.company()["registered"]
        assert isinstance(registered, list)
        self.assertEqual([(r["format"], r["legal"], r["weight"]) for r in registered], [
            ("{.company.sia}", "SIA", 80), ("{.company.ik}", "IK", 50), ("{.company.as}", "AS", 20), ("{.company.ps}", "PS", 2), ("{.company.ks}", "KS", 1),
        ])

    def test_a_form_places_its_abbreviation_and_its_quotes_as_the_names_of_that_form_do_and_ships_its_shapes_of_share_min_or_more(self) -> None:
        self.assertEqual(self.company()["sia"], {
            "format": "{name as sia_name}{placed}",
            "name": [{"format": "{.company-coined}", "weight": 200}],
            "hidden": {"placed": [
                {"format": 'SIA "{sia_name}"', "weight": 6}, {"format": "SIA {sia_name}", "weight": 4},
                {"format": "{sia_name} SIA", "weight": 3}, {"format": '"{sia_name}" SIA', "weight": 1},
            ]},
        })

    def test_a_form_whose_names_never_place_it_one_way_does_not_draw_that_way(self) -> None:
        self.assertEqual(self.company()["ps"]["hidden"], {"placed": [{"format": 'PS "{ps_name}"', "weight": 1}]})

    def test_an_ik_draws_the_person_its_names_carry(self) -> None:
        self.assertEqual(self.company()["ik"]["name"], [{"format": "{.person as p}{p.first} {p.last}", "weight": 10}])


class Main(unittest.TestCase):
    def test_the_script_reads_a_register_file_and_writes_the_company_files_from_its_names_and_the_shipped_tables(self) -> None:
        rows = [('SIA "Alfa Būve"', "K", "SIA", "")] * 100 + [('SIA "Bērziņš"', "K", "SIA", "")] * 60 + [('SIA "Zorba"', "K", "SIA", "")] * 40
        with tempfile.TemporaryDirectory() as folder:
            register = Path(folder) / "register.csv"
            register.write_bytes(register_csv(rows))
            with mock.patch.object(sys, "argv", ["company-lv.py", "--register", str(register), "--out", folder]):
                company_lv.main()
            company = json.loads((Path(folder) / "company.json").read_text(encoding="utf-8"))
            self.assertEqual(company["registered"][0], {"format": "{.company.sia}", "legal": "SIA", "weight": 200})
            self.assertEqual(
                company["sia"]["name"],
                [{"format": "{.company-coined} {.company-word}", "weight": 100}, {"format": "{.last-name}", "weight": 60}, {"format": "{.company-coined}", "weight": 40}],
            )
            self.assertEqual((Path(folder) / "company-coined.tsv").read_text(encoding="utf-8").splitlines(), ["word\tcount", "Zorba\t40"])
            self.assertEqual((Path(folder) / "company-word.tsv").read_text(encoding="utf-8").splitlines(), ["word\tcount", "Būve\t100"])


if __name__ == "__main__":
    unittest.main()
