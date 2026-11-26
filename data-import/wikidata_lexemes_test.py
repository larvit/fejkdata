import gzip
import json
import unittest
from typing import Any

import wikidata_lexemes

NOMINATIVE = "Q131105"
PLURAL = "Q146786"
SINGULAR = "Q110786"
PARTITIVE = "Q857325"
FORMS = {"noun": "plural", "verb": "", "adj": "", "adv": ""}


def form(value: str, *features: str, code: str = "et") -> dict[str, Any]:
	return {"representations": {code: {"language": code, "value": value}}, "grammaticalFeatures": list(features)}


def lexeme(category: str, lemma: str, forms: list[dict[str, Any]] | None = None, language: str = "Q9072", code: str = "et") -> dict[str, Any]:
	return {"type": "lexeme", "lemmas": {code: {"language": code, "value": lemma}}, "lexicalCategory": category, "language": language, "forms": forms or []}


def dump(*lexemes: dict[str, Any]) -> bytes:
	"""Wikidata's lexeme dump: a JSON array written one lexeme a line, gzipped."""
	lines = "[\n" + ",\n".join(json.dumps(x) for x in lexemes) + "\n]\n"
	return gzip.compress(lines.encode())


def drawn(*lexemes: dict[str, Any], language: str = "Q9072", code: str = "et") -> list[tuple[str, str, str]]:
	return list(wikidata_lexemes.candidates(dump(*lexemes), language, code, FORMS))


class Candidates(unittest.TestCase):
	def test_a_noun_carries_its_nominative_plural(self) -> None:
		forms = [form("maja", NOMINATIVE, SINGULAR), form("majasid", PARTITIVE, PLURAL), form("majad", NOMINATIVE, PLURAL)]
		self.assertEqual(drawn(lexeme("Q1084", "maja", forms)), [("noun", "maja", "majad")])

	def test_a_noun_with_no_nominative_plural_has_none(self) -> None:
		forms = [form("majasid", PARTITIVE, PLURAL), form("maja", NOMINATIVE, SINGULAR)]
		self.assertEqual(drawn(lexeme("Q1084", "maja", forms), lexeme("Q1084", "tennis")), [("noun", "maja", ""), ("noun", "tennis", "")])

	def test_a_verb_adjective_and_adverb_carry_no_plural(self) -> None:
		got = drawn(lexeme("Q24905", "oskama"), lexeme("Q34698", "hell", [form("hellad", NOMINATIVE, PLURAL)]), lexeme("Q380057", "alati"))
		self.assertEqual(got, [("verb", "oskama", ""), ("adj", "hell", ""), ("adv", "alati", "")])

	def test_a_lexeme_of_another_category_is_left_out(self) -> None:
		self.assertEqual(drawn(lexeme("Q147276", "tallinn"), lexeme("Q36224", "mina"), lexeme("Q4833830", "sellest")), [])

	def test_a_lexeme_of_another_language_is_left_out(self) -> None:
		self.assertEqual(drawn(lexeme("Q1084", "house", language="Q1860"), lexeme("Q1084", "talo", language="Q1412")), [])

	def test_the_language_and_its_code_are_the_caller_s(self) -> None:
		german = lexeme("Q1084", "Haus", [form("Häuser", NOMINATIVE, PLURAL, code="de")], language="Q188", code="de")
		self.assertEqual(drawn(german, lexeme("Q1084", "maja"), language="Q188", code="de"), [("noun", "Haus", "Häuser")])

	def test_a_dump_with_no_lexeme_has_no_candidate(self) -> None:
		self.assertEqual(drawn(), [])


if __name__ == "__main__":
	unittest.main()
