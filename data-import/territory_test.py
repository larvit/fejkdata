import unittest

import territory

HEADER = "ISO3166-1-Alpha-2,ISO3166-1-Alpha-3,Dial,Capital,ISO4217-currency_alphabetic_code,Languages,CLDR display name,official_name_en,ISO3166-1-numeric,is_independent,TLD"


def capital(row: str) -> str:
	[got] = territory.rows(HEADER + "\n" + row + "\n")
	return got["capital"]


class Capitals(unittest.TestCase):
	def test_kazakhstan_is_astana(self) -> None:
		self.assertEqual(capital("KZ,KAZ,7,Nur-Sultan,KZT,\"kk,ru\",Kazakhstan,Kazakhstan,398,Yes,.kz"), "Astana")

	def test_palau_is_ngerulmud(self) -> None:
		self.assertEqual(capital("PW,PLW,680,Melekeok,USD,\"en-PW,pau,ja,sov,zh-CN\",Palau,Palau,585,Yes,.pw"), "Ngerulmud")

	def test_a_capital_the_source_has_right_stays(self) -> None:
		self.assertEqual(capital("SE,SWE,46,Stockholm,SEK,\"sv-SE,se,sma,fi-SE\",Sweden,Sweden,752,Yes,.se"), "Stockholm")


class Tlds(unittest.TestCase):
	def test_western_sahara_has_none(self) -> None:
		[got] = territory.rows(HEADER + "\nEH,ESH,212,El-Aaiun,MAD,\"ar,mey\",Western Sahara,Western Sahara,732,In contention,.eh\n")
		self.assertEqual(got["tld"], "")

	def test_western_sahara_stays_where_the_source_gives_it_no_tld(self) -> None:
		[got] = territory.rows(HEADER + "\nEH,ESH,212,El-Aaiun,MAD,\"ar,mey\",Western Sahara,Western Sahara,732,In contention,\n")
		self.assertEqual((got["alpha2"], got["tld"]), ("EH", ""))

	def test_a_row_the_source_gives_no_tld_is_left_out(self) -> None:
		self.assertEqual(list(territory.rows(HEADER + "\nSE,SWE,46,Stockholm,SEK,\"sv-SE,se,sma,fi-SE\",Sweden,Sweden,752,Yes,\n")), [])

	def test_a_tld_the_source_has_right_stays(self) -> None:
		[got] = territory.rows(HEADER + "\nSE,SWE,46,Stockholm,SEK,\"sv-SE,se,sma,fi-SE\",Sweden,Sweden,752,Yes,.se\n")
		self.assertEqual(got["tld"], ".se")


if __name__ == "__main__":
	unittest.main()
