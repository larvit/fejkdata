import json
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from company_names import count_names, placed
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


class Forms(unittest.TestCase):
	def test_a_legal_word_of_every_form_is_left_out_of_a_shape(self) -> None:
		for form, (_, (abbreviation, _)) in company_lt.NAMED_FORMS.items():
			with self.subTest(form=form):
				shapes, _ = count_names(company_lt.NAMED_FORMS, [(form, f"{form} Zorba {abbreviation}")], WORDS, set(), company_lt.VOCABULARY)
				self.assertEqual(dict(shapes[form]), {("X",): 1})

	def test_every_form_is_placed_by_its_abbreviation_and_by_its_words(self) -> None:
		for form, (_, spellings) in company_lt.NAMED_FORMS.items():
			for spelling in spellings:
				with self.subTest(form=form, spelling=spelling):
					self.assertEqual(placed(company_lt.NAMED_FORMS, form, f'{spelling} "Alfa"'), ("before", True, ["Alfa"]))

	def test_the_vocabulary_names_the_connecting_words_of_its_language(self) -> None:
		shapes, _ = count_names(company_lt.NAMED_FORMS, [("UAB", "Kazlauskas ir Ona UAB")], WORDS, set(), company_lt.VOCABULARY)
		self.assertEqual(dict(shapes["UAB"]), {("S", "ir", "F"): 1})


class Coined(unittest.TestCase):
	def coined(self, word: str, names: int, ends: int, trade: set[str] | None = None) -> dict[str, int]:
		rows = [("UAB", f"Pluss {word}" if i < ends else f"{word} Pluss") for i in range(names)]
		found: dict[str, int] = company_lt.coined(company_lt.word_usage(rows, lambda w: False, company_lt.VOCABULARY), trade or set())
		return found

	def test_a_coined_word_keeps_each_lithuanian_letter(self) -> None:
		for letter in "ąčęėįšųūž":
			with self.subTest(letter=letter):
				self.assertIn(f"Zorb{letter}", self.coined(f"Zorb{letter}", 40, 2))

	def test_a_coined_word_spelled_in_capitals_with_a_lithuanian_vowel_is_spelled_in_lower_case(self) -> None:
		for vowel in "ąęėįųū":
			with self.subTest(vowel=vowel):
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
