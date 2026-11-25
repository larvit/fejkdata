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


if __name__ == "__main__":
    unittest.main()
