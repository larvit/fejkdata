import json
import re
import unittest
from collections import Counter

from company_names import (
    Vocabulary,
    Words,
    bare,
    coined,
    company_json,
    count_names,
    kind,
    placed,
    shape_of,
    shipped,
    slots,
    word_usage,
)

VOCABULARY = Vocabulary(re.compile(r"^[a-z]{3,}$"), frozenset(), frozenset(), frozenset({"A"}), 0, ("and", "&"))
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

    def test_a_coined_word_takes_its_most_common_spelling(self) -> None:
        rows = [("A", "Pluss ZorBa" if i < 25 else "Pluss Zorba") for i in range(40)]
        found = coined(word_usage(rows, lambda w: False, VOCABULARY), set())
        self.assertEqual(found["ZorBa"], 40)
        self.assertNotIn("Zorba", found)

    def test_a_trade_word_is_not_coined(self) -> None:
        self.assertNotIn("Zorba", self.coined("Zorba", 40, 2, {"zorba"}))


FORMS = {"UAB": ("uab", ("uab", "uždaroji akcinė bendrovė")), "MB": ("mb", ("mb", "mažoji bendrija")), "AB": ("ab", ("ab", "akcinė bendrovė"))}
PLACING = Vocabulary(re.compile(r"^[a-ząčęėįšųūž]{3,}$"), frozenset(), frozenset({"uab", "mb", "ab"}), frozenset(FORMS), 0, ("ir", "&"))
PEOPLE = Words({"kazlauskas"}, {"jonas"})


class Placed(unittest.TestCase):
    def test_a_name_places_its_form_before_or_after_it_and_quoted_or_bare(self) -> None:
        self.assertEqual(placed(FORMS, "UAB", 'UAB "Alfa Statyba"'), ("before", True, ["Alfa", "Statyba"]))
        self.assertEqual(placed(FORMS, "MB", "MB Alfa"), ("before", False, ["Alfa"]))
        self.assertEqual(placed(FORMS, "UAB", "Alfa, UAB"), ("after", False, ["Alfa"]))
        self.assertEqual(placed(FORMS, "UAB", '"Alfa" UAB'), ("after", True, ["Alfa"]))

    def test_the_form_counts_as_its_abbreviation_or_its_words_in_any_case_and_with_or_without_a_dot(self) -> None:
        self.assertEqual(placed(FORMS, "UAB", 'Uždaroji akcinė bendrovė "Alfa"'), ("before", True, ["Alfa"]))
        self.assertEqual(placed(FORMS, "UAB", 'UŽDAROJI AKCINĖ BENDROVĖ "ALFA"'), ("before", True, ["ALFA"]))
        self.assertEqual(placed(FORMS, "AB", "Alfa akcinė bendrovė"), ("after", False, ["Alfa"]))
        self.assertEqual(placed(FORMS, "AB", "Alfa, AB."), ("after", False, ["Alfa"]))

    def test_a_name_carrying_another_form_or_none_or_only_its_form_or_the_form_inside_is_not_placed(self) -> None:
        self.assertIsNone(placed(FORMS, "UAB", 'AB "Alfa"'))
        self.assertIsNone(placed(FORMS, "UAB", "Alfa Statyba"))
        self.assertIsNone(placed(FORMS, "UAB", "UAB"))
        self.assertIsNone(placed(FORMS, "UAB", "Alfa UAB Beta"))
        self.assertIsNone(placed(FORMS, "AB", "Alfa, A.B."))


class Counting(unittest.TestCase):
    def test_a_legal_word_of_every_form_is_left_out_of_a_shape_with_or_without_a_dot(self) -> None:
        for form, spelling in (("UAB", "Uab."), ("MB", "Mb."), ("AB", "Ab.")):
            with self.subTest(form=form):
                shapes, _ = count_names(FORMS, [(form, f"{spelling} Zorba {form}")], PEOPLE, set(), PLACING)
                self.assertEqual(dict(shapes[form]), {("X",): 1})

    def test_a_name_of_legal_words_alone_is_not_counted(self) -> None:
        shapes, places = count_names(FORMS, [("UAB", "Uždaroji akcinė bendrovė UAB")], PEOPLE, set(), PLACING)
        self.assertEqual((dict(shapes["UAB"]), dict(places["UAB"])), ({}, {}))

    def test_a_name_counts_its_shape_and_where_it_places_its_form(self) -> None:
        names = [("UAB", 'UAB "Zorba"'), ("UAB", "Zorba UAB"), ("UAB", "Zorba"), ("AB", 'AB "Jonas Kazlauskas"')]
        shapes, places = count_names(FORMS, names, PEOPLE, set(), PLACING)
        self.assertEqual(dict(shapes["UAB"]), {("X",): 2})
        self.assertEqual(dict(places["UAB"]), {("before", True): 1, ("after", False): 1})
        self.assertEqual(dict(shapes["AB"]), {("FS",): 1})
        self.assertEqual(dict(places["AB"]), {("before", True): 1})

    def test_a_name_that_places_its_form_loses_the_form_and_its_quotes_and_any_other_name_is_dropped(self) -> None:
        names = [("UAB", 'Uždaroji akcinė bendrovė "Alfa Statyba"'), ("AB", "Beta AB"), ("MB", "Gama"), ("UAB", "UAB")]
        self.assertEqual(bare(FORMS, names), [("UAB", "Alfa Statyba"), ("AB", "Beta")])


SHAPES: dict[str, dict[tuple[str, ...], int]] = {"UAB": {("X",): 200, ("T", "X"): 1}, "MB": {("FS",): 10}, "AB": {("X",): 1}}
PLACES = {
    "UAB": Counter({("before", True): 6, ("before", False): 4, ("after", False): 3, ("after", True): 1}),
    "MB": Counter({("before", False): 2}),
    "AB": Counter({("after", True): 1}),
}


class Company(unittest.TestCase):
    def company(self) -> dict[str, dict[str, object]]:
        data: dict[str, dict[str, object]] = json.loads(company_json(FORMS, {"MB": 50, "UAB": 80, "AB": 2}, SHAPES, PLACES, PLACING))
        return data

    def test_the_forms_weigh_their_entities_most_first(self) -> None:
        registered = self.company()["registered"]
        assert isinstance(registered, list)
        self.assertEqual([(r["format"], r["legal"], r["weight"]) for r in registered], [("{.company.uab}", "UAB", 80), ("{.company.mb}", "MB", 50), ("{.company.ab}", "AB", 2)])

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
        self.assertEqual(self.company()["ab"]["hidden"], {"placed": [{"format": '"{ab_name}" AB', "weight": 1}]})

    def test_a_form_whose_names_carry_a_person_draws_it(self) -> None:
        self.assertEqual(self.company()["mb"]["name"], [{"format": "{.person as p}{p.first} {p.last}", "weight": 10}])
