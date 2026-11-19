import json
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from company_names import count_names
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


class Forms(unittest.TestCase):
    def test_a_legal_word_of_every_form_is_left_out_of_a_shape(self) -> None:
        for form, (_, (abbreviation, _)) in company_lv.NAMED_FORMS.items():
            with self.subTest(form=form):
                shapes, _ = count_names(company_lv.NAMED_FORMS, [(form, f"{form} Zorba {abbreviation}")], WORDS, set(), company_lv.VOCABULARY)
                self.assertEqual(dict(shapes[form]), {("X",): 1})

    def test_the_vocabulary_names_the_connecting_words_of_its_language(self) -> None:
        shapes, _ = count_names(company_lv.NAMED_FORMS, [("SIA", "Ozols un Anna SIA")], WORDS, set(), company_lv.VOCABULARY)
        self.assertEqual(dict(shapes["SIA"]), {("S", "un", "F"): 1})


class Coined(unittest.TestCase):
    def coined(self, word: str, names: int, ends: int, trade: set[str] | None = None) -> dict[str, int]:
        rows = [("SIA", f"Pluss {word}" if i < ends else f"{word} Pluss") for i in range(names)]
        found: dict[str, int] = company_lv.coined(company_lv.word_usage(rows, lambda w: False, company_lv.VOCABULARY), trade or set())
        return found

    def test_a_coined_word_keeps_each_latvian_letter(self) -> None:
        for letter in "āčēģīķļņšūž":
            self.assertIn(f"Zorb{letter}", self.coined(f"Zorb{letter}", 40, 2))

    def test_a_coined_word_spelled_in_capitals_with_a_latvian_vowel_is_spelled_in_lower_case(self) -> None:
        for vowel in "āēīū":
            self.assertIn(vowel.upper() + vowel * 3, self.coined(vowel.upper() * 4, 40, 2))


class Trade(unittest.TestCase):
    def test_a_trade_word_of_one_form_alone_counts_since_sia_names_are_nearly_every_name(self) -> None:
        rows = [("SIA", "Alfa Būve" if i < 40 else "Būve Alfa") for i in range(100)]
        trade = company_lv.trade_words(company_lv.word_usage(rows, lambda w: False, company_lv.VOCABULARY), company_lv.VOCABULARY)
        self.assertEqual(trade["būve"], ("Būve", 100))

    def test_a_trade_word_has_three_letters_or_more(self) -> None:
        rows = [("SIA", "Art Ab" if i < 50 else "Ab Art") for i in range(100)]
        trade = company_lv.trade_words(company_lv.word_usage(rows, lambda w: False, company_lv.VOCABULARY), company_lv.VOCABULARY)
        self.assertEqual(sorted(trade), ["art"])


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
