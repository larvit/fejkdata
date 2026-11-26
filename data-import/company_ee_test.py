import io
import json
import re
import unittest
import zipfile
from collections import Counter

from geo_test import script

company_ee = script("company-ee")

HEADER = "nimi;ariregistri_kood;ettevotja_oiguslik_vorm;ettevotja_staatus\n"
WORDS = company_ee.Words({"tamm", "kask", "ivanova"}, {"jaan", "mari"})


def basic_data(rows: list[tuple[str, str, str]]) -> bytes:
	buffer = io.BytesIO()
	with zipfile.ZipFile(buffer, "w") as z:
		z.writestr("ettevotja_rekvisiidid__lihtandmed.csv", "﻿" + HEADER + "".join(f"{name};{i};{form};{status}\n" for i, (name, form, status) in enumerate(rows)))
	return buffer.getvalue()


class Register(unittest.TestCase):
	def test_a_legal_person_entered_into_the_register_counts_for_its_form_and_every_other_is_left_out(self) -> None:
		counts, names = company_ee.register(
			basic_data([
				("Jaan Tamm", "Füüsilisest isikust ettevõtja", "R"), ("Alfa OÜ", "Osaühing", "R"), ("Beeta OÜ", "Osaühing", "L"),
				("Gamma AS", "Aktsiaselts", "N"), ("Delta UÜ", "Usaldusühing", "R"), ("Eesti TÜ", "Täisühing", "R"),
				("AS Zeta", "Aktsiaselts", "R"), ("Kase MTÜ", "Mittetulundusühing", "R"), ("Kaare tn 2 korteriühistu", "Korteriühistu", "R"),
			])
		)
		self.assertEqual(dict(counts), {"FIE": 1, "OÜ": 1, "UÜ": 1, "TÜ": 1, "AS": 1})
		self.assertEqual(names, [("OÜ", "Alfa OÜ"), ("UÜ", "Delta UÜ"), ("TÜ", "Eesti TÜ"), ("AS", "AS Zeta")])


class Placed(unittest.TestCase):
	def test_a_name_ending_in_its_form_places_it_after_and_one_opening_with_it_before(self) -> None:
		self.assertEqual(company_ee.placed("OÜ", "Alfa Ehitus OÜ"), ("after", ["Alfa", "Ehitus"]))
		self.assertEqual(company_ee.placed("AS", "AS Tallinna Vesi"), ("before", ["Tallinna", "Vesi"]))

	def test_the_form_counts_as_its_abbreviation_or_its_word_in_any_case_and_with_a_dot_or_comma(self) -> None:
		self.assertEqual(company_ee.placed("OÜ", "Osaühing Alfa"), ("before", ["Alfa"]))
		self.assertEqual(company_ee.placed("OÜ", "ALFA OSAÜHING"), ("after", ["ALFA"]))
		self.assertEqual(company_ee.placed("AS", "Alfa, A.S."), None)
		self.assertEqual(company_ee.placed("AS", "Alfa, AS."), ("after", ["Alfa"]))
		self.assertEqual(company_ee.placed("TÜ", "täisühing Alfa"), ("before", ["Alfa"]))

	def test_a_name_carrying_another_form_or_none_or_only_its_form_is_not_placed(self) -> None:
		self.assertIsNone(company_ee.placed("OÜ", "Alfa AS"))
		self.assertIsNone(company_ee.placed("OÜ", "Alfa Ehitus"))
		self.assertIsNone(company_ee.placed("OÜ", "OÜ"))


class Shapes(unittest.TestCase):
	def test_a_legal_word_is_left_out_of_a_shape_with_or_without_a_dot(self) -> None:
		shapes, _ = company_ee.count_names([("OÜ", "Osaühing Zorba OÜ"), ("OÜ", "Osaühing. Zorba OÜ")], WORDS, set())
		self.assertEqual(dict(shapes["OÜ"]), {("X",): 2})

	def test_a_name_of_legal_words_alone_is_not_counted(self) -> None:
		shapes, places = company_ee.count_names([("OÜ", "Osaühing OÜ")], WORDS, set())
		self.assertEqual((dict(shapes["OÜ"]), dict(places["OÜ"])), ({}, {}))

	def test_the_vocabulary_names_the_connecting_words_of_its_language(self) -> None:
		shapes, _ = company_ee.count_names([("OÜ", "Kask ja Mari OÜ")], WORDS, set())
		self.assertEqual(dict(shapes["OÜ"]), {("S", "ja", "F"): 1})

	def test_a_name_counts_its_shape_and_where_it_places_its_form(self) -> None:
		shapes, places = company_ee.count_names([("OÜ", "Zorba OÜ"), ("OÜ", "OÜ Zorba"), ("OÜ", "Zorba"), ("AS", "AS Jaan Tamm")], WORDS, set())
		self.assertEqual(dict(shapes["OÜ"]), {("X",): 2})
		self.assertEqual(dict(places["OÜ"]), {"after": 1, "before": 1})
		self.assertEqual(dict(shapes["AS"]), {("FS",): 1})
		self.assertEqual(dict(places["AS"]), {"before": 1})


class Coined(unittest.TestCase):
	def coined(self, word: str, names: int, ends: int, trade: set[str] | None = None) -> dict[str, int]:
		rows = [("OÜ", f"Pluss {word}" if i < ends else f"{word} Pluss") for i in range(names)]
		found: dict[str, int] = company_ee.coined(company_ee.word_usage(rows, lambda w: False, company_ee.VOCABULARY), trade or set())
		return found

	def test_a_coined_word_keeps_its_estonian_letters(self) -> None:
		self.assertEqual(self.coined("ÕNNE", 40, 2)["Õnne"], 40)
		self.assertIn("Šokk", self.coined("Šokk", 40, 2))


class Trade(unittest.TestCase):
	def test_a_trade_word_of_one_form_alone_counts_since_ou_names_are_nearly_every_name(self) -> None:
		rows = [("OÜ", "Alfa Ehitus" if i < 40 else "Ehitus Alfa") for i in range(100)]
		trade = company_ee.trade_words(company_ee.word_usage(rows, lambda w: False, company_ee.VOCABULARY), company_ee.VOCABULARY)
		self.assertEqual(trade["ehitus"], ("Ehitus", 100))

	def test_a_trade_word_spans_three_forms_where_a_vocabulary_names_no_spread(self) -> None:
		vocabulary = company_ee.Vocabulary(re.compile(r"^[a-z]{3,}$"), frozenset(), frozenset(), frozenset({"A", "B", "C"}), 0, ())
		rows = [("AB"[i % 2], "Alfa Ehitus" if i < 40 else "Ehitus Alfa") for i in range(100)]
		self.assertNotIn("ehitus", company_ee.trade_words(company_ee.word_usage(rows, lambda w: False, vocabulary), vocabulary))
		rows = [("ABC"[i % 3], name) for i, (_, name) in enumerate(rows)]
		self.assertIn("ehitus", company_ee.trade_words(company_ee.word_usage(rows, lambda w: False, vocabulary), vocabulary))


SHAPES = {"OÜ": {("X",): 200, ("T", "X"): 1}, "AS": {("FS",): 10}, "UÜ": {("X",): 1}, "TÜ": {("X",): 1}}
PLACES = {"OÜ": Counter({"after": 3, "before": 1}), "AS": Counter({"before": 2}), "UÜ": Counter({"after": 1}), "TÜ": Counter({"after": 1})}


class Output(unittest.TestCase):
	def company(self) -> dict[str, dict[str, object]]:
		data: dict[str, dict[str, object]] = json.loads(company_ee.company_json({"FIE": 50, "OÜ": 80, "AS": 20, "UÜ": 2, "TÜ": 1}, SHAPES, PLACES))
		return data

	def test_the_forms_weigh_their_businesses_with_no_abbreviation_on_the_fie(self) -> None:
		registered = self.company()["registered"]
		assert isinstance(registered, list)
		self.assertEqual([(r["format"], r["legal"], r["weight"]) for r in registered], [
			("{.company.ou}", "OÜ", 80), ("{.company.fie}", "", 50), ("{.company.as}", "AS", 20), ("{.company.uu}", "UÜ", 2), ("{.company.tu}", "TÜ", 1),
		])

	def test_an_fie_is_one_person(self) -> None:
		self.assertEqual(self.company()["fie"], {"format": "{name}", "name": "{.person as p}{p.first} {p.last}"})

	def test_a_form_puts_its_abbreviation_where_the_names_put_the_form_and_ships_its_shapes_above_share_min(self) -> None:
		self.assertEqual(self.company()["ou"], {
			"format": "{name as ou_name}{placed}",
			"name": [{"format": "{.company-coined}", "weight": 200}],
			"hidden": {"placed": [{"format": "{ou_name} OÜ", "weight": 3}, {"format": "OÜ {ou_name}", "weight": 1}]},
		})

	def test_a_form_whose_names_never_place_it_one_way_does_not_draw_that_way(self) -> None:
		self.assertEqual(self.company()["as"]["hidden"], {"placed": [{"format": "AS {as_name}", "weight": 2}]})


if __name__ == "__main__":
	unittest.main()
