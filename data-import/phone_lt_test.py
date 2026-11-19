import re
import unittest

from geo_test import script

phone_lt = script("phone-lt")


class Subscriber(unittest.TestCase):
    def test_a_three_digit_area_code_leaves_five_digits_in_one_group(self) -> None:
        self.assertEqual(phone_lt.subscriber("310", 3), "{digits(5)}")

    def test_a_two_digit_area_code_leaves_two_and_four_digits(self) -> None:
        self.assertEqual(phone_lt.subscriber("37", 2), "{digits(2)} {digits(4)}")

    def test_the_digits_the_plan_fixes_open_the_first_group(self) -> None:
        self.assertEqual(phone_lt.subscriber("454", 2), "4{digits(1)} {digits(4)}")
        self.assertEqual(phone_lt.subscriber("520", 1), "20{digits(1)} {digits(4)}")

    def test_every_area_and_its_subscriber_number_make_eight_digits(self) -> None:
        for area, ndc_len, _ in phone_lt.AREAS:
            template = phone_lt.subscriber(area, ndc_len)
            digits = re.sub(r"\{digits\((\d)\)\}", lambda m: "x" * int(m.group(1)), template).replace(" ", "")
            self.assertEqual(len(digits), 8 - ndc_len, area)


if __name__ == "__main__":
    unittest.main()
