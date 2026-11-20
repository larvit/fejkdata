import json
import unittest
from typing import Any

from geo_test import script

word_et = script("word-et")


def entry(pos: str, word: str, template: str = "", senses: list[dict[str, Any]] | None = None, forms: list[dict[str, Any]] | None = None, generic: str = "") -> dict[str, Any]:
    """An entry whose head template is `template`, or, where `generic` is given, the generic `head` template with `generic` as its second argument."""
    made: dict[str, Any] = {"pos": pos, "word": word, "senses": senses if senses is not None else [{"glosses": ["x"]}]}
    if generic:
        made["head_templates"] = [{"name": "head", "args": {"1": "et", "2": generic}}]
    elif template:
        made["head_templates"] = [{"name": template}]
    if forms is not None:
        made["forms"] = forms
    return made


def dump(*entries: dict[str, Any]) -> bytes:
    return "".join(json.dumps(e) + "\n" for e in entries).encode()


def drawn(*entries: dict[str, Any]) -> list[tuple[str, str, str]]:
    return list(word_et.candidates(dump(*entries)))


class Spelled(unittest.TestCase):
    def test_lower_case_letters_alone_are_a_word(self) -> None:
        self.assertTrue(word_et.spelled("maja", "noun"))
        self.assertTrue(word_et.spelled("õun", "noun"))
        self.assertTrue(word_et.spelled("šokolaad", "noun"))

    def test_a_capital_a_hyphen_a_digit_a_modifier_letter_or_nothing_is_not(self) -> None:
        for word in ["Tallinn", "-ki", "b2b", "mᶜˣ", "t'ee", ""]:
            self.assertFalse(word_et.spelled(word, "noun"), word)


class Candidates(unittest.TestCase):
    def test_a_noun_carries_its_nominative_plural(self) -> None:
        forms = [
            {"form": "maja", "tags": ["genitive"]},
            {"form": "majadele", "tags": ["allative", "plural"]},
            {"form": "majasid", "tags": ["partitive", "plural"]},
            {"form": "majad", "tags": ["accusative", "nominative", "plural"]},
            {"form": "majad", "tags": ["nominative", "plural"]},
        ]
        self.assertEqual(drawn(entry("noun", "maja", "et-noun", forms=forms)), [("noun", "maja", "majad")])

    def test_a_noun_with_no_nominative_plural_has_none(self) -> None:
        forms = [{"form": "majasid", "tags": ["partitive", "plural"]}, {"form": "maja", "tags": ["nominative", "singular"]}]
        self.assertEqual(drawn(entry("noun", "maja", "et-noun", forms=forms), entry("noun", "tennis", "et-noun")), [("noun", "maja", ""), ("noun", "tennis", "")])

    def test_a_verb_adjective_and_adverb_carry_no_plural(self) -> None:
        got = drawn(entry("verb", "oskama", "et-verb"), entry("adj", "hell", "et-adj", forms=[{"form": "helled", "tags": ["nominative", "plural"]}]), entry("adv", "alati", "et-adv"))
        self.assertEqual(got, [("verb", "oskama", ""), ("adj", "hell", ""), ("adv", "alati", "")])

    def test_an_entry_the_head_template_calls_a_form_or_a_participle_is_left_out(self) -> None:
        got = drawn(
            entry("verb", "on", generic="verb form"),
            entry("noun", "augustit", generic="noun form"),
            entry("adj", "head", generic="adjective form"),
            entry("verb", "viivitunud", generic="past active participle"),
            entry("verb", "huvitav", "et-nom", generic="present active participle"),
            entry("adj", "parem", generic="comparative adjective"),
            entry("adv", "rohkem", generic="comparative adverb"),
        )
        self.assertEqual(got, [])

    def test_the_generic_head_template_of_a_word_of_its_part_stays(self) -> None:
        got = drawn(entry("adv", "ka", generic="adverb"), entry("verb", "paisuma", generic="verb"), entry("adj", "rikas", generic="adjective"), entry("noun", "a", generic="nouns"))
        self.assertEqual(got, [("adv", "ka", ""), ("verb", "paisuma", ""), ("adj", "rikas", ""), ("noun", "a", "")])

    def test_a_template_of_another_part_of_speech_is_left_out(self) -> None:
        self.assertEqual(drawn(entry("noun", "maadlema", "et-verb"), entry("verb", "ka", generic="adverb")), [])

    def test_a_part_of_speech_the_tables_do_not_hold_is_left_out(self) -> None:
        self.assertEqual(drawn(entry("name", "tallinn", "et-proper noun"), entry("pron", "mina", "et-pron")), [])

    def test_an_entry_whose_every_sense_is_a_form_or_variant_is_left_out(self) -> None:
        senses = [{"form_of": [{"word": "oskama"}]}, {"alt_of": [{"word": "oskama"}]}]
        self.assertEqual(drawn(entry("verb", "oskama", "et-verb", senses=senses)), [])

    def test_an_entry_with_one_sense_of_its_own_stays(self) -> None:
        senses: list[dict[str, Any]] = [{"form_of": [{"word": "oskama"}]}, {"glosses": ["x"]}]
        self.assertEqual(drawn(entry("verb", "oskama", "et-verb", senses=senses)), [("verb", "oskama", "")])

    def test_an_entry_with_no_sense_is_left_out(self) -> None:
        self.assertEqual(drawn(entry("verb", "oskama", "et-verb", senses=[])), [])


if __name__ == "__main__":
    unittest.main()
