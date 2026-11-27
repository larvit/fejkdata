import json
import unittest
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
	return json.dumps({"results": {"bindings": [{"label": {"value": label}, "people": {"value": str(people)}} for label, people in rows]}}).encode()


class Bearers(unittest.TestCase):
	def test_each_label_is_asked_once_for_its_language_and_the_countrys_people_are_read_back(self) -> None:
		asked: list[str] = []

		def fetch(url: str, cache: str, name: str, magic: bytes = b"", data: bytes | None = None, headers: dict[str, str] | None = None) -> bytes:
			assert data is not None
			asked.append(data.decode())
			return answer(("Lefèvre", 328), ("Lefevre", 63))

		with mock.patch.object(source, "fetch", fetch):
			found = wikidata_names.bearers(["Lefèvre", "Lefevre", "Lefèvre"], "fr", "Q142", "cache")
		self.assertEqual(found, {"Lefèvre": 328, "Lefevre": 63})
		self.assertEqual(len(asked), 1)
		self.assertEqual(asked[0].count("%22Lef%C3%A8vre%22%40fr"), 1)
		self.assertIn("wd%3AQ142", asked[0])

	def test_labels_over_the_batch_go_in_several_questions(self) -> None:
		asked: list[bytes] = []

		def fetch(url: str, cache: str, name: str, magic: bytes = b"", data: bytes | None = None, headers: dict[str, str] | None = None) -> bytes:
			assert data is not None
			asked.append(data)
			return answer()

		with mock.patch.object(source, "fetch", fetch), mock.patch.object(wikidata_names, "BATCH", 2):
			wikidata_names.bearers(["Aa", "Bb", "Cc", "Dd", "Ee"], "en", "Q30", "cache")
		self.assertEqual(len(asked), 3)
