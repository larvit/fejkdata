import json
import unittest
import urllib.parse
from collections.abc import Iterable
from unittest import mock

import source
import wikidata_names


class Spelled(unittest.TestCase):
	def test_the_form_most_bearers_write_wins_where_they_are_a_majority_and_enough(self) -> None:
		self.assertEqual(wikidata_names.spelled("Obrien", {"Obrien": 6, "O'Brien": 2320}, 3), "O'Brien")

	def test_the_default_stays_where_no_form_is_a_majority(self) -> None:
		self.assertEqual(wikidata_names.spelled("Delacruz", {"Delacruz": 4, "De La Cruz": 5, "De la Cruz": 5}, 3), "Delacruz")

	def test_the_default_stays_where_the_commonest_form_has_fewer_bearers_than_the_floor(self) -> None:
		self.assertEqual(wikidata_names.spelled("Stjohn", {"Stjohn": 0, "St John": 2}, 3), "Stjohn")

	def test_the_default_stays_where_it_is_the_commonest(self) -> None:
		self.assertEqual(wikidata_names.spelled("Smith", {"Smith": 900, "Smíth": 4}, 3), "Smith")

	def test_the_default_stays_where_no_form_has_a_bearer(self) -> None:
		self.assertEqual(wikidata_names.spelled("Zed", {}, 3), "Zed")


def answer(*rows: tuple[str, int]) -> bytes:
	return json.dumps({"results": {"bindings": [{"s": {"value": label}, "people": {"value": str(people)}} for label, people in rows]}}).encode()


def query(body: bytes) -> str:
	(text,) = urllib.parse.parse_qs(body.decode())["query"]
	return text


class Bearers(unittest.TestCase):
	def test_each_label_is_asked_once_for_its_language_and_mul_and_the_countrys_people_are_read_back(self) -> None:
		asked: list[str] = []

		def fetch(url: str, cache: str, name: str, magic: bytes = b"", data: bytes | None = None, headers: dict[str, str] | None = None) -> bytes:
			assert data is not None
			asked.append(query(data))
			return answer(("Lefèvre", 328), ("Lefevre", 63))

		with mock.patch.object(source, "fetch", fetch):
			found = wikidata_names.bearers(["Lefèvre", "Lefevre", "Lefèvre"], "fr", "Q142", "cache")
		self.assertEqual(found, {"Lefèvre": 328, "Lefevre": 63})
		self.assertEqual(len(asked), 1)
		self.assertEqual(asked[0].count('"Lefèvre"@fr'), 1)
		self.assertEqual(asked[0].count('"Lefèvre"@mul'), 1)
		self.assertIn("wd:Q142", asked[0])

	def test_a_label_with_an_apostrophe_is_also_asked_with_the_typographic_one_and_both_count_as_one_label(self) -> None:
		text = query(wikidata_names.question(["O'Brien", "Obrien"], "en", "Q30"))
		self.assertIn('"O\'Brien"@en', text)
		self.assertIn('"O’Brien"@en', text)
		self.assertIn('"O’Brien"@mul', text)
		self.assertEqual(text.count('"Obrien"@en'), 1)
		self.assertIn("""REPLACE(STR(?label), "’", "'")""", text)
		self.assertIn("GROUP BY ?s", text)

	def test_a_name_of_any_kind_of_family_name_is_asked(self) -> None:
		self.assertIn("wdt:P31/wdt:P279* wd:Q101352", query(wikidata_names.question(["Oconnell"], "en", "Q30")))

	def test_labels_over_the_batch_go_in_several_questions(self) -> None:
		asked: list[bytes] = []

		def fetch(url: str, cache: str, name: str, magic: bytes = b"", data: bytes | None = None, headers: dict[str, str] | None = None) -> bytes:
			assert data is not None
			asked.append(data)
			return answer()

		with mock.patch.object(source, "fetch", fetch), mock.patch.object(wikidata_names, "BATCH", 2):
			wikidata_names.bearers(["Aa", "Bb", "Cc", "Dd", "Ee"], "en", "Q30", "cache")
		self.assertEqual(len(asked), 3)


class Written(unittest.TestCase):
	def test_a_form_wins_where_at_least_5_people_and_more_than_half_of_those_counted_bear_it(self) -> None:
		for people, want in [
			({"Obrien": 1, "O'Brien": 4}, "Obrien"),
			({"Obrien": 1, "O'Brien": 5}, "O'Brien"),
			({"Obrien": 2, "O'Brien": 5, "OBrien": 3}, "Obrien"),
		]:

			def ask(labels: Iterable[str], language: str, country: str, cache: str, people: dict[str, int] = people) -> dict[str, int]:
				return people

			with self.subTest(people):
				self.assertEqual(wikidata_names.written({"Obrien": ["O'Brien", "OBrien"], "Smith": []}, "en", "Q30", "cache", ask), {"Obrien": want})
