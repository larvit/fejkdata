import re
import unittest

from geo_test import script

phone_lt = script("phone-lt")


def digits_of(template: str) -> int:
	template = re.sub(r"\{digits\((\d)\)\}", lambda m: "x" * int(m.group(1)), template)
	return len(re.sub(r"\{int\(\d,\d\)\}", "x", template).replace(" ", ""))


class Subscriber(unittest.TestCase):
	def test_a_three_digit_area_code_leaves_five_digits_in_one_group(self) -> None:
		self.assertEqual(phone_lt.subscriber(3), "{digits(5)}")

	def test_a_two_digit_area_code_leaves_two_and_four_digits(self) -> None:
		self.assertEqual(phone_lt.subscriber(2), "{digits(2)} {digits(4)}")

	def test_a_one_digit_area_code_leaves_three_and_four_digits(self) -> None:
		self.assertEqual(phone_lt.subscriber(1), "{digits(3)} {digits(4)}")

	def test_the_digits_the_plan_fixes_open_the_first_group_as_a_choice(self) -> None:
		self.assertEqual(phone_lt.subscriber(2, "{int(4,5)}", 1), "{int(4,5)}{digits(1)} {digits(4)}")
		self.assertEqual(
			phone_lt.subscriber(1, "2{int(0,7)}", 2), "2{int(0,7)}{digits(1)} {digits(4)}"
		)


class Rows(unittest.TestCase):
	def test_a_row_is_one_area_code_with_its_trunk_zero(self) -> None:
		rows = phone_lt.rows()
		self.assertEqual(len(rows), 47)
		for row in rows:
			self.assertEqual(row["area"], "0" + row["ndc"])

	def test_every_area_and_its_subscriber_number_make_eight_digits(self) -> None:
		for row in phone_lt.rows():
			self.assertEqual(digits_of(row["ndc"] + row["subscriber"]), 8, row["area"])


if __name__ == "__main__":
	unittest.main()
