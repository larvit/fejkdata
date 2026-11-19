import re
import unittest
from collections import Counter

from company_names import Vocabulary, Words, coined, kind, shape_of, shipped, slots, word_usage

VOCABULARY = Vocabulary(re.compile(r"^[a-z]{3,}$"), frozenset(), frozenset(), frozenset({"A"}), 0, connecting_words=("and", "&"))
WORDS = Words({"tamm", "kask", "ivanova"}, {"jaan", "mari"})


class Shapes(unittest.TestCase):
    def shape(self, name: str) -> tuple[str, ...]:
        shape: tuple[str, ...] = shape_of(name.split(), WORDS, {"trade"}, VOCABULARY)
        return shape

    def test_each_word_is_a_name_a_number_a_connecting_word_a_trade_word_or_coined(self) -> None:
        self.assertEqual(self.shape("Kask and Mari 12 Trade Zorba &"), ("S", "and", "F", "N", "T", "X", "&"))

    def test_the_vocabulary_names_the_connecting_words(self) -> None:
        self.assertEqual(self.shape("Zorba un Zorba"), ("X", "X", "X"))

    def test_a_first_name_before_a_last_name_is_one_person(self) -> None:
        self.assertEqual(self.shape("Jaan Tamm Trade"), ("FS", "T"))
        self.assertEqual(self.shape("Jaan Tamm & Mari Kask"), ("FS", "&", "FS"))

    def test_a_name_is_compared_in_lower_case_and_a_word_that_is_both_names_is_a_last_name_at_the_end(self) -> None:
        words = Words({"tamm", "mari"}, {"mari"})
        self.assertEqual(kind("MARI", words, False), "F")
        self.assertEqual(kind("Mari", words, True), "S")
        self.assertEqual(kind("TAMM", words, False), "S")
        self.assertIsNone(kind("Zorba", words, True))

    def test_a_word_that_is_both_names_is_a_first_name_inside_a_name_and_a_last_name_at_its_end_in_a_shape(self) -> None:
        words = Words({"tamm", "mari"}, {"mari", "jaan"})
        self.assertEqual(shape_of(["Mari", "Zorba"], words, set(), VOCABULARY), ("F", "X"))
        self.assertEqual(shape_of(["Zorba", "Mari"], words, set(), VOCABULARY), ("X", "S"))
        self.assertEqual(shape_of(["Jaan", "Mari"], words, set(), VOCABULARY), ("FS",))

    def test_a_shape_opening_with_a_number_or_a_connecting_word_does_not_ship(self) -> None:
        shapes = Counter({("X",): 50, ("N", "X"): 30, ("&", "X"): 10, ("and", "T"): 10})
        self.assertEqual(shipped(shapes, VOCABULARY), [(("X",), 50)])

    def test_each_person_of_a_shape_is_drawn_once_under_a_name_of_its_own(self) -> None:
        self.assertEqual(
            slots(("FS", "&", "FS"), VOCABULARY),
            "{.person as p}{p.first} {p.last} & {.person as q}{q.first} {q.last}",
        )
        self.assertEqual(slots(("S", "and", "F", "N", "T", "X"), VOCABULARY), "{.last-name} and {.first-name} {int(1,30)} {.company-word} {.company-coined}")


class Coined(unittest.TestCase):
    def coined(self, word: str, names: int, ends: int, trade: set[str] | None = None) -> dict[str, int]:
        rows = [("A", f"Pluss {word}" if i < ends else f"{word} Pluss") for i in range(names)]
        found: dict[str, int] = coined(word_usage(rows, lambda w: False, VOCABULARY), trade or set())
        return found

    def test_a_coined_word_is_carried_by_COINED_MIN_names_and_ends_at_least_COINED_ENDS_MIN_of_them(self) -> None:
        self.assertEqual(self.coined("Zorba", 40, 2)["Zorba"], 40)
        self.assertNotIn("Zorba", self.coined("Zorba", 39, 2))
        self.assertNotIn("Zorba", self.coined("Zorba", 40, 1))

    def test_a_name_carrying_a_word_twice_counts_it_once(self) -> None:
        rows = [("A", "Pluss Zorba Zorba" if i < 20 else "Zorba Zorba Pluss") for i in range(30)]
        zorba = word_usage(rows, lambda w: False, VOCABULARY)["zorba"]
        self.assertEqual((zorba.names, zorba.begins, zorba.ends, zorba.forms["A"]), (30, 10, 20, 30))
        self.assertNotIn("Zorba", coined(word_usage(rows[:20], lambda w: False, VOCABULARY), set()))

    def test_a_coined_word_has_four_letters_or_more(self) -> None:
        self.assertNotIn("Zor", self.coined("Zor", 40, 2))
        self.assertIn("Zorb", self.coined("Zorb", 40, 2))

    def test_a_trade_word_is_not_coined(self) -> None:
        self.assertNotIn("Zorba", self.coined("Zorba", 40, 2, {"zorba"}))
