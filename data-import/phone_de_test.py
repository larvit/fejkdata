import io
import unittest
import zipfile

from geo_test import script

phone_de = script("phone-de")


def archive(csv: str) -> bytes:
	out = io.BytesIO()
	with zipfile.ZipFile(out, "w") as z:
		z.writestr("onb.csv", csv.encode("utf-8-sig"))
	return out.getvalue()


class MobileRows(unittest.TestCase):
	def test_a_block_of_three_four_or_five_digits_leaves_a_subscriber_number_of_the_plans_length(self) -> None:
		rows = phone_de.mobile_rows(b"RNB;Name\n160;A\n1511;B\n15591 C\n")
		self.assertEqual(
			rows,
			[
				{"area": "01511", "ndc": "1511", "subscriber": "{.phone-subscriber.digits-7}"},
				{"area": "015591", "ndc": "15591", "subscriber": "{.phone-subscriber.digits-6}"},
				{"area": "0160", "ndc": "160", "subscriber": "{.phone-subscriber.digits-7-8}"},
			],
		)

	def test_a_line_that_is_no_block_is_refused(self) -> None:
		with self.assertRaises(SystemExit):
			phone_de.mobile_rows(b"RNB;Name\nA;160\n")

	def test_a_block_of_six_digits_is_refused(self) -> None:
		with self.assertRaises(SystemExit):
			phone_de.mobile_rows(b"RNB;Name\n155911;A\n")


class FixedRows(unittest.TestCase):
	def test_an_active_code_leaves_a_subscriber_number_that_makes_the_total_ten_or_eleven_digits(self) -> None:
		rows = phone_de.fixed_rows(archive("Ortsnetzkennzahl;Ortsnetzname;KennzeichenAktiv\n30;Berlin;1\n201;Essen;1\n33052;Inaktiv;0\n"))
		self.assertEqual(
			rows,
			[
				{"area": "0201", "ndc": "201", "name": "Essen", "subscriber": "{.phone-subscriber.digits-8}"},
				{"area": "030", "ndc": "30", "name": "Berlin", "subscriber": "{.phone-subscriber.digits-8}"},
			],
		)

	def test_four_and_five_digit_codes_leave_seven_and_six_digits(self) -> None:
		rows = phone_de.fixed_rows(archive("Ortsnetzkennzahl;Ortsnetzname;KennzeichenAktiv\n2302;A;1\n33051;B;1\n"))
		self.assertEqual([r["subscriber"] for r in rows], ["{.phone-subscriber.digits-7}", "{.phone-subscriber.digits-6}"])

	def test_a_code_that_is_not_two_to_five_digits_without_a_zero_first_is_refused(self) -> None:
		for code in ("0123", "1", "123456"):
			with self.subTest(code=code), self.assertRaises(SystemExit):
				phone_de.fixed_rows(archive(f"Ortsnetzkennzahl;Ortsnetzname;KennzeichenAktiv\n{code};A;1\n"))


if __name__ == "__main__":
	unittest.main()
