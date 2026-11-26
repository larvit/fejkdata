import unittest

import wordtables


class LowerLetters(unittest.TestCase):
	def test_lower_case_letters_alone_are_a_word(self) -> None:
		for word in ["maja", "õun", "élève", "šokolaad"]:
			self.assertTrue(wordtables.lower_letters(word, "noun"), word)

	def test_a_capital_a_hyphen_a_digit_a_modifier_letter_an_apostrophe_a_space_or_nothing_is_not(self) -> None:
		for word in ["Tallinn", "-ki", "b2b", "mᶜˣ", "l'eau", "maja maja", ""]:
			self.assertFalse(wordtables.lower_letters(word, "noun"), word)


class Rows(unittest.TestCase):
	def test_a_part_with_a_form_drops_a_word_whose_form_is_empty(self) -> None:
		got = wordtables.rows([("noun", "māja", "mājas"), ("noun", "ūdens", "")], {"noun": "plural"}, set(), wordtables.lower_letters)
		self.assertEqual(got, {"noun": [{"word": "māja", "plural": "mājas"}]})


if __name__ == "__main__":
	unittest.main()
