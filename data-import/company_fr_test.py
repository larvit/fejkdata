import io
import unittest
import zipfile

from geo_test import script

company_fr = script("company-fr")

HEADER = "siren,etatAdministratifUniteLegale,denominationUniteLegale,categorieJuridiqueUniteLegale\n"


def stock(rows: list[tuple[str, str, str]]) -> bytes:
    buffer = io.BytesIO()
    with zipfile.ZipFile(buffer, "w") as z:
        z.writestr("StockUniteLegale_utf8.csv", HEADER + "".join(f"{i},{state},{name},{category}\n" for i, (state, name, category) in enumerate(rows)))
    return buffer.getvalue()


class Register(unittest.TestCase):
    def test_an_active_unit_counts_for_the_form_of_its_category_and_every_other_unit_is_left_out(self) -> None:
        counts, names = company_fr.register(
            stock([("A", "", "1000"), ("A", "DUPONT FRERES", "5499"), ("C", "CLOSED", "5499"), ("A", "ALPHA", "5599"), ("A", "BETA", "5699"), ("A", "SCI DU PORT", "6540")])
        )
        self.assertEqual(dict(counts), {"EI": 1, "SARL": 1, "SA": 2})
        self.assertEqual(names, [("SARL", "DUPONT FRERES"), ("SA", "ALPHA"), ("SA", "BETA")])

    def test_a_legal_form_ending_a_name_is_dropped_and_one_inside_it_is_kept(self) -> None:
        _, names = company_fr.register(stock([("A", "ALPHA S.A.R.L.", "5499"), ("A", "BETA SASU", "5710"), ("A", "SASSA", "5710"), ("A", "SA BETA", "5710")]))
        self.assertEqual([name for _, name in names], ["ALPHA", "BETA", "SASSA", "SA BETA"])


class Shapes(unittest.TestCase):
    def shape(self, name: str) -> tuple[str, ...]:
        shape: tuple[str, ...] = company_fr.shape_of(name.split(), {"dupont"}, {"jean"}, {"services"})
        return shape

    def test_each_word_is_a_name_a_number_a_connecting_word_a_trade_word_or_coined(self) -> None:
        self.assertEqual(self.shape("JEAN DUPONT ET FILS 12 SERVICES ZORBA"), ("F", "S", "et", "X", "N", "T", "X"))

    def test_a_name_is_compared_without_accents_and_a_word_that_is_both_names_is_a_last_name(self) -> None:
        self.assertEqual(company_fr.kind("Dupônt", {"dupont"}, {"dupont"}), "S")
        self.assertEqual(company_fr.kind("JEAN", {"dupont"}, {"jean"}), "F")


class Words(unittest.TestCase):
    usage = company_fr.word_usage([("SARL", "ZORBA PLUS")] * 40 + [("SAS", "ZORBA PLUS")] * 40, lambda w: False, company_fr.VOCABULARY)

    def test_a_word_that_begins_names_in_two_forms_is_a_root_with_its_names_unless_it_is_a_trade_word(self) -> None:
        self.assertEqual(company_fr.roots(self.usage, set()), {"Zorba": 80})
        self.assertEqual(company_fr.roots(self.usage, {"zorba"}), {})

    def test_a_spelling_is_its_word_with_accents(self) -> None:
        for word, spelling in company_fr.SPELLINGS.items():
            self.assertEqual(company_fr.fold(spelling), word)
            self.assertNotEqual(spelling.lower(), word)
