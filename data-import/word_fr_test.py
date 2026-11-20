import json
import unittest
from typing import Any

from geo_test import script

word_fr = script("word-fr")


def entry(pos: str, word: str, template: str = "", senses: list[dict[str, Any]] | None = None, forms: list[dict[str, Any]] | None = None, generic: str = "") -> dict[str, Any]:
    """An entry whose head template is `template`, or, where `generic` is given, the generic `head` template with `generic` as its second argument."""
    made: dict[str, Any] = {"pos": pos, "word": word, "senses": senses if senses is not None else [{"glosses": ["x"]}]}
    if generic:
        made["head_templates"] = [{"name": "head", "args": {"1": "fr", "2": generic}}]
    elif template:
        made["head_templates"] = [{"name": template}]
    if forms is not None:
        made["forms"] = forms
    return made


def dump(*entries: dict[str, Any]) -> bytes:
    return "".join(json.dumps(e) + "\n" for e in entries).encode()


def drawn(*entries: dict[str, Any]) -> list[tuple[str, str, str]]:
    return list(word_fr.candidates(dump(*entries)))


class Candidates(unittest.TestCase):
    def test_a_noun_carries_its_first_plural(self) -> None:
        forms = [{"form": "garennes", "tags": ["plural"]}, {"form": "autres", "tags": ["plural"]}]
        self.assertEqual(drawn(entry("noun", "garenne", "fr-noun", forms=forms)), [("noun", "garenne", "garennes")])

    def test_an_invariable_noun_is_its_own_plural(self) -> None:
        senses = [{"glosses": ["x"], "tags": ["masculine", "invariable"]}]
        self.assertEqual(drawn(entry("noun", "pays", "fr-noun", senses=senses)), [("noun", "pays", "pays")])

    def test_a_noun_with_no_plural_has_none(self) -> None:
        self.assertEqual(drawn(entry("noun", "faim", "fr-noun")), [("noun", "faim", "")])

    def test_a_verb_adjective_and_adverb_carry_no_plural(self) -> None:
        got = drawn(entry("verb", "aimer", "fr-verb"), entry("adj", "beau", "fr-adj", forms=[{"form": "beaux", "tags": ["plural"]}]), entry("adv", "souvent", "fr-adv"))
        self.assertEqual(got, [("verb", "aimer", ""), ("adj", "beau", ""), ("adv", "souvent", "")])

    def test_an_entry_the_head_template_calls_a_form_or_a_participle_is_left_out(self) -> None:
        got = drawn(
            entry("verb", "aimez", generic="verb form"),
            entry("noun", "lunettes", generic="noun form"),
            entry("adj", "murs", generic="adjective form"),
            entry("verb", "impliquant", generic="present participle"),
            entry("verb", "dit", "fr-past participle"),
        )
        self.assertEqual(got, [])

    def test_the_generic_head_template_of_a_word_of_its_part_stays(self) -> None:
        got = drawn(entry("adv", "bien", generic="adverb"), entry("verb", "lyncher", generic="verb"), entry("adj", "force", generic="adjective"), entry("noun", "papa", generic="nouns"))
        self.assertEqual(got, [("adv", "bien", ""), ("verb", "lyncher", ""), ("adj", "force", ""), ("noun", "papa", "")])

    def test_a_template_of_another_part_of_speech_is_left_out(self) -> None:
        self.assertEqual(drawn(entry("adj", "quoicoubaka", "fr-noun"), entry("verb", "bien", generic="adverb")), [])

    def test_a_part_of_speech_the_tables_do_not_hold_is_left_out(self) -> None:
        self.assertEqual(drawn(entry("name", "paris", "fr-proper noun"), entry("prep", "sur", "fr-prep")), [])

    def test_an_entry_whose_every_sense_is_a_form_or_variant_is_left_out(self) -> None:
        senses = [{"form_of": [{"word": "aimer"}]}, {"alt_of": [{"word": "aimer"}]}]
        self.assertEqual(drawn(entry("verb", "aimer", "fr-verb", senses=senses)), [])

    def test_an_entry_with_one_sense_of_its_own_stays(self) -> None:
        senses: list[dict[str, Any]] = [{"form_of": [{"word": "aimer"}]}, {"glosses": ["x"]}]
        self.assertEqual(drawn(entry("verb", "aimer", "fr-verb", senses=senses)), [("verb", "aimer", "")])

    def test_an_entry_with_no_sense_is_left_out(self) -> None:
        self.assertEqual(drawn(entry("verb", "aimer", "fr-verb", senses=[])), [])


if __name__ == "__main__":
    unittest.main()
