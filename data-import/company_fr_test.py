import io
import json
import unittest
import zipfile

from geo_test import script

company_fr = script("company-fr")

HEADER = "siren,etatAdministratifUniteLegale,denominationUniteLegale,categorieJuridiqueUniteLegale\n"


def stock(rows: list[tuple[str, str, str]]) -> bytes:
	buffer = io.BytesIO()
	with zipfile.ZipFile(buffer, "w") as z:
		z.writestr("StockUniteLegale_utf8.csv", HEADER + "".join(f"{i},{state},{name},{category}\n" for i, (state, name, category) in enumerate(rows)))
	return buffer.getvalue()


class Register(unittest.TestCase):
	def test_an_active_unit_counts_for_the_form_of_its_category_and_every_other_unit_is_left_out(self) -> None:
		counts, names = company_fr.register(
			stock([("A", "", "1000"), ("A", "DUPONT FRERES", "5499"), ("C", "CLOSED", "5499"), ("A", "ALPHA", "5599"), ("A", "BETA", "5699"), ("A", "SCI DU PORT", "6540")])
		)
		self.assertEqual(dict(counts), {"EI": 1, "SARL": 1, "SA": 2})
		self.assertEqual(names, [("SARL", "DUPONT FRERES"), ("SA", "ALPHA"), ("SA", "BETA")])

	def test_a_legal_form_ending_a_name_is_dropped_and_one_inside_it_is_kept(self) -> None:
		_, names = company_fr.register(stock([("A", "ALPHA S.A.R.L.", "5499"), ("A", "BETA SASU", "5710"), ("A", "SASSA", "5710"), ("A", "SA BETA", "5710")]))
		self.assertEqual([name for _, name in names], ["ALPHA", "BETA", "SASSA", "SA BETA"])

	def test_each_category_counts_for_its_form(self) -> None:
		counts, _ = company_fr.register(stock([("A", "", "1000"), ("A", "A", "5499"), ("A", "B", "5710"), ("A", "C", "5202"), ("A", "D", "5599"), ("A", "E", "5699")]))
		self.assertEqual(dict(counts), {"EI": 1, "SARL": 1, "SAS": 1, "SNC": 1, "SA": 2})

	def test_a_trailing_eurl_sa_or_comma_form_is_dropped(self) -> None:
		_, names = company_fr.register(stock([("A", "ALPHA EURL", "5499"), ("A", "BETA SA", "5599"), ("A", "GAMMA S.A.", "5599"), ("A", '"DELTA, SAS"', "5710")]))
		self.assertEqual([name for _, name in names], ["ALPHA", "BETA", "GAMMA", "DELTA"])


class Shapes(unittest.TestCase):
	def shape(self, name: str) -> tuple[str, ...]:
		shape: tuple[str, ...] = company_fr.shape_of(name.split(), {"dupont"}, {"jean"}, {"services"})
		return shape

	def test_each_word_is_a_name_a_number_a_connecting_word_a_trade_word_or_coined(self) -> None:
		self.assertEqual(self.shape("JEAN DUPONT ET FILS 12 SERVICES ZORBA"), ("F", "S", "et", "X", "N", "T", "X"))

	def test_a_name_is_compared_without_accents_and_a_word_that_is_both_names_is_a_last_name(self) -> None:
		self.assertEqual(company_fr.kind("Dupônt", {"dupont"}, {"dupont"}), "S")
		self.assertEqual(company_fr.kind("JEAN", {"dupont"}, {"jean"}), "F")


	def test_a_legal_word_is_left_out_of_a_shape_and_a_name_of_legal_words_is_not_counted(self) -> None:
		shapes = company_fr.count_shapes([("SARL", "SOCIETE ZORBA"), ("SARL", "SCI"), ("SARL", "EARL ZORBA")], set(), set(), set())
		self.assertEqual(dict(shapes["SARL"]), {("X",): 2})

	def test_a_shape_slots_are_joined_and_the_first_is_capitalised(self) -> None:
		self.assertEqual(company_fr.slots(("la", "X", "N")), "La {.company-coined} {int(1,30)}")
		self.assertEqual(company_fr.slots(("S", "et", "F")), "{.last-name} et {.first-name}")


class Words(unittest.TestCase):
	def usage(self, word: str, begins: int, ends: int, sas: int = 50, names: int = 100) -> dict[str, object]:
		rows = []
		for i in range(names):
			form = "SAS" if i < sas else "SARL"
			if i < begins:
				rows.append((form, f"{word} PLUS"))
			elif i < begins + ends:
				rows.append((form, f"PLUS {word}"))
			else:
				rows.append((form, f"PLUS {word} PLUS"))
		found: dict[str, object] = company_fr.roots(company_fr.word_usage(rows, lambda w: False, company_fr.VOCABULARY), set())
		return found

	def test_a_trade_word_is_no_root(self) -> None:
		rows = [("SARL" if i % 2 else "SAS", "ZORBA PLUS" if i < 80 else "PLUS ZORBA") for i in range(100)]
		usage = company_fr.word_usage(rows, lambda w: False, company_fr.VOCABULARY)
		self.assertEqual(company_fr.roots(usage, set())["Zorba"], 100)
		self.assertNotIn("Zorba", company_fr.roots(usage, {"zorba"}))

	def test_a_root_begins_ROOT_MIN_names(self) -> None:
		self.assertIn("Zorba", self.usage("ZORBA", 60, 5))
		self.assertNotIn("Zorba", self.usage("ZORBA", 59, 5))

	def test_a_root_ends_under_a_third_of_its_names_and_at_least_ROOT_ENDS_MIN_of_them(self) -> None:
		self.assertIn("Zorba", self.usage("ZORBA", 60, 33))
		self.assertNotIn("Zorba", self.usage("ZORBA", 60, 34))
		self.assertIn("Zorba", self.usage("ZORBA", 60, 5))
		self.assertNotIn("Zorba", self.usage("ZORBA", 60, 4))

	def test_a_root_has_four_to_six_letters(self) -> None:
		self.assertIn("Zorb", self.usage("ZORB", 60, 5))
		self.assertIn("Zorbas", self.usage("ZORBAS", 60, 5))
		self.assertNotIn("Zor", self.usage("ZOR", 60, 5))
		self.assertNotIn("Zorbass", self.usage("ZORBASS", 60, 5))

	def test_a_root_is_carried_by_FORM_MIN_names_in_two_forms(self) -> None:
		self.assertIn("Zorba", self.usage("ZORBA", 60, 5, sas=10))
		self.assertNotIn("Zorba", self.usage("ZORBA", 60, 5, sas=9))

	def test_a_root_takes_its_spelling(self) -> None:
		self.assertIn("Hôtel", self.usage("HOTEL", 60, 5))

	def test_the_commonest_trade_words_take_their_spellings(self) -> None:
		trade = {"cafe": ("Cafe", 5), "btp": ("Btp", 9), "zzz": ("Zzz", 1)}
		self.assertEqual(company_fr.commonest(trade, 2), {"btp": ("Btp", 9), "cafe": ("Café", 5)})

	def test_a_spelling_is_its_word_with_accents(self) -> None:
		for word, spelling in company_fr.SPELLINGS.items():
			self.assertEqual(company_fr.fold(spelling), word)
			self.assertNotEqual(spelling, word.capitalize())


SHAPES = {"SARL": {("la", "X"): 200, ("X",): 2}, "SAS": {("S", "T"): 10}, "SNC": {("X",): 1}, "SA": {("X",): 1}}


class Output(unittest.TestCase):
	def test_a_form_block_holds_its_shapes_above_the_share_by_count_and_ends_in_its_abbreviation(self) -> None:
		self.assertEqual(
			company_fr.form_block("SARL", SHAPES["SARL"]),
			'  "sarl": {\n    "format": "{name} SARL",\n    "name": [\n      { "format": "La {.company-coined}", "weight": 200 }\n    ]\n  }',
		)

	def test_an_ei_is_its_owner(self) -> None:
		self.assertEqual(company_fr.ei_block(), '  "ei": {\n    "format": "{name}",\n    "name": "{.first-name} {.last-name}"\n  }')

	def test_the_forms_are_registered_by_units_with_no_abbreviation_on_the_ei(self) -> None:
		data = json.loads(company_fr.company_json({"EI": 50, "SARL": 30, "SAS": 20, "SNC": 2, "SA": 1}, SHAPES))
		self.assertEqual([r["legal"] for r in data["registered"]], ["", "SARL", "SAS", "SNC", "SA"])
		self.assertEqual([r["weight"] for r in data["registered"]], [50, 30, 20, 2, 1])
		self.assertEqual(data["sas"]["name"], [{"format": "{.last-name} {.company-word}", "weight": 10}])
