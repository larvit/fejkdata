import unittest

from geo_test import script
from word_fr_test import dump, entry

word_lt = script("word-lt")

PUPA_PLURAL = {"form": "pùpos", "links": [["pùpos", "pupos"]], "tags": ["plural"]}


def drawn(*entries: dict[str, object]) -> list[tuple[str, str, str]]:
    return list(word_lt.candidates(dump(*entries)))


class Candidates(unittest.TestCase):
    def test_a_noun_carries_its_first_plural_without_the_stress_marks(self) -> None:
        forms = [PUPA_PLURAL, {"form": "pupų", "tags": ["genitive", "plural"]}]
        self.assertEqual(drawn(entry("noun", "pupa", "lt-noun", forms=forms)), [("noun", "pupa", "pupos")])

    def test_a_plural_names_the_page_it_links_to_and_not_its_language_section(self) -> None:
        forms = [{"form": "mãmos", "links": [["mãmos", "mamos#Lithuanian"]], "tags": ["nominative", "plural"]}]
        self.assertEqual(drawn(entry("noun", "mama", "lt-noun", forms=forms)), [("noun", "mama", "mamos")])

    def test_a_noun_whose_plural_links_nowhere_has_none(self) -> None:
        self.assertEqual(drawn(entry("noun", "pupa", "lt-noun", forms=[{"form": "pùpos", "tags": ["plural"]}])), [("noun", "pupa", "")])

    def test_a_noun_with_no_plural_has_none(self) -> None:
        self.assertEqual(drawn(entry("noun", "tenisas", "lt-noun")), [("noun", "tenisas", "")])

    def test_a_verb_adjective_and_adverb_carry_no_plural(self) -> None:
        got = drawn(entry("verb", "bėgti", "lt-verb"), entry("adj", "didelis", "lt-adj", forms=[PUPA_PLURAL]), entry("adv", "dažnai", "lt-adv"))
        self.assertEqual(got, [("verb", "bėgti", ""), ("adj", "didelis", ""), ("adv", "dažnai", "")])

    def test_an_entry_the_head_template_calls_a_form_or_a_participle_is_left_out(self) -> None:
        got = drawn(
            entry("verb", "bėgo", generic="verb form"),
            entry("noun", "namų", generic="noun form"),
            entry("adj", "didelio", generic="adjective form"),
            entry("verb", "bėgdamas", generic="participle"),
            entry("verb", "bėgantis", generic="participle form"),
            entry("verb", "bėgęs", "lt-part"),
            entry("adv", "didesnis", generic="comparative adverb"),
        )
        self.assertEqual(got, [])

    def test_the_generic_head_template_of_a_word_of_its_part_stays(self) -> None:
        got = drawn(entry("adv", "ten", generic="adverbs"), entry("verb", "bėgti", generic="verb"), entry("adj", "geras", generic="adjective"), entry("noun", "ponas", generic="noun"))
        self.assertEqual(got, [("adv", "ten", ""), ("verb", "bėgti", ""), ("adj", "geras", ""), ("noun", "ponas", "")])

    def test_a_part_of_speech_the_tables_do_not_hold_is_left_out(self) -> None:
        self.assertEqual(drawn(entry("name", "vilnius", "lt-proper noun"), entry("prep", "ant", "lt-prep")), [])


if __name__ == "__main__":
    unittest.main()
