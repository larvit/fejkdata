import io
import unittest
import zipfile

from geo_test import script

word_lv = script("word-lv")

NOUN, VERB, ADJ, ADV = "Lietvārds", "Darbības vārds", "Īpašības vārds", "Apstākļa vārds"


def gram(kind: str, value: str) -> str:
	return f'<gram type="{kind}">{value}</gram>'


def inflection(orth: str, case: str, number: str) -> str:
	return f'<form type="inflection"><orth>{orth}</orth><gramGrp type="properties">{gram("Locījums", case)}{gram("Skaitlis", number)}</gramGrp></form>'


def derivative(orth: str, pos: str) -> str:
	return f'<form type="derivative"><orth>{orth}</orth><gramGrp type="properties">{gram("Vārdšķira", pos)}</gramGrp></form>'


def entry(pos: str, lemma: str, forms: list[tuple[str, str, str]] | None = None, labels: tuple[str, ...] = (), derived: list[tuple[str, str]] | None = None) -> str:
	"""One wordform entry as Tēzaurs.lv writes it: the lemma with its properties, then each of its inflections and derivatives."""
	props = gram("Vārdšķira", pos) + "".join(gram("Lietojums", label) for label in labels)
	shown = "".join(inflection(*f) for f in forms or []) + "".join(derivative(*d) for d in derived or [])
	return f'<entry type="supplemental"><form><orth type="lemma">{lemma}</orth><gramGrp type="properties">{props}</gramGrp>{shown}</form></entry>'


def dump(*entries: str) -> bytes:
	"""Tēzaurs.lv's wordform file as the zip it is published in, each tag on a line of its own."""
	xml = '<?xml version="1.0"?>\n<TEI>\n<standOff>\n' + "\n".join(entries).replace("><", ">\n<") + "\n</standOff>\n</TEI>\n"
	made = io.BytesIO()
	with zipfile.ZipFile(made, "w") as z:
		z.writestr("tezaurs_wordforms_tei.xml", xml)
	return made.getvalue()


def drawn(*entries: str) -> list[tuple[str, str, str]]:
	return list(word_lv.candidates(dump(*entries)))


HOUSE = [("māja", "Nominatīvs", "Vienskaitlis"), ("mājas", "Ģenitīvs", "Vienskaitlis"), ("mājas", "Nominatīvs", "Daudzskaitlis"), ("māju", "Ģenitīvs", "Daudzskaitlis")]


class Candidates(unittest.TestCase):
	def test_a_noun_carries_its_nominative_plural(self) -> None:
		got = drawn(entry(NOUN, "māja", [("māju", "Ģenitīvs", "Daudzskaitlis")] + HOUSE))
		self.assertEqual(got, [("noun", "māja", "mājas")])

	def test_a_noun_carries_its_first_nominative_plural(self) -> None:
		forms = [("vīri", "Nominatīvs", "Daudzskaitlis"), ("vīrie", "Nominatīvs", "Daudzskaitlis")]
		self.assertEqual(drawn(entry(NOUN, "vīrs", forms)), [("noun", "vīrs", "vīri")])

	def test_a_noun_with_no_nominative_plural_has_none(self) -> None:
		forms = [("gaiss", "Nominatīvs", "Vienskaitlis"), ("gaisa", "Ģenitīvs", "Vienskaitlis")]
		self.assertEqual(drawn(entry(NOUN, "gaiss", forms), entry(NOUN, "ūdens")), [("noun", "gaiss", ""), ("noun", "ūdens", "")])

	def test_a_noun_used_in_the_plural_alone_is_its_own_plural(self) -> None:
		self.assertEqual(drawn(entry(NOUN, "brilles", [("brilles", "Nominatīvs", "Daudzskaitlis")])), [("noun", "brilles", "brilles")])

	def test_a_verb_adjective_and_adverb_carry_no_plural(self) -> None:
		plural = [("labi", "Nominatīvs", "Daudzskaitlis")]
		got = drawn(entry(VERB, "skriet", plural), entry(ADJ, "labs", plural), entry(ADV, "bieži"))
		self.assertEqual(got, [("verb", "skriet", ""), ("adj", "labs", ""), ("adv", "bieži", "")])

	def test_a_verb_is_an_infinitive_and_a_participle_filed_as_a_verb_is_left_out(self) -> None:
		got = drawn(entry(VERB, "pamācīt"), entry(VERB, "baidīties"), entry(VERB, "pamācošs"), entry(VERB, "rakstāms"), entry(VERB, "cietušais"), entry(VERB, "nodarīts"))
		self.assertEqual(got, [("verb", "pamācīt", ""), ("verb", "baidīties", "")])

	def test_a_noun_or_adjective_need_not_end_as_an_infinitive(self) -> None:
		self.assertEqual(drawn(entry(NOUN, "māja"), entry(ADJ, "labs")), [("noun", "māja", ""), ("adj", "labs", "")])

	def test_a_part_of_speech_the_tables_do_not_hold_is_left_out(self) -> None:
		pos = ("Saīsinājums", "Reziduālis", "Divdabis", "Vietniekvārds", "Prievārds", "Saiklis", "Partikula", "Izsauksmes vārds", "Skaitļa vārds")
		self.assertEqual(drawn(*[entry(p, "kā") for p in pos]), [])

	def test_a_word_marked_as_regional_or_out_of_date_is_left_out(self) -> None:
		labels = ("Apvidvārds", "Dialektisms", "Novecojis", "Vēsturisks", "Neaktuāls", "Sens, reti lietots vārds")
		self.assertEqual(drawn(*[entry(NOUN, "abāds", labels=(label,)) for label in labels]), [])

	def test_a_word_marked_as_regional_beside_another_label_is_left_out(self) -> None:
		self.assertEqual(drawn(entry(VERB, "abižot", labels=("Sarunvaloda", "Apvidvārds"))), [])

	def test_a_colloquial_slang_or_rude_word_stays(self) -> None:
		labels = ("Sarunvaloda", "Žargonisms", "Slengs", "Vulgārisms", "Lamuvārds", "Barbarisms")
		self.assertEqual(drawn(*[entry(ADV, f"w{i}", labels=(label,)) for i, label in enumerate(labels)]), [("adv", f"w{i}", "") for i in range(len(labels))])

	def test_the_label_of_one_entry_does_not_reach_the_next(self) -> None:
		self.assertEqual(drawn(entry(NOUN, "abāds", labels=("Apvidvārds",)), entry(NOUN, "māja")), [("noun", "māja", "")])

	def test_the_plural_of_one_entry_does_not_reach_the_next(self) -> None:
		self.assertEqual(drawn(entry(NOUN, "māja", HOUSE), entry(NOUN, "ūdens")), [("noun", "māja", "mājas"), ("noun", "ūdens", "")])

	def test_a_form_whose_case_or_number_differs_is_not_the_plural(self) -> None:
		forms = [("mājas", "Ģenitīvs", "Vienskaitlis"), ("māju", "Ģenitīvs", "Daudzskaitlis"), ("māja", "Nominatīvs", "Vienskaitlis")]
		self.assertEqual(drawn(entry(NOUN, "māja", forms)), [("noun", "māja", "")])

	def test_a_derivative_keeps_neither_its_text_nor_its_part_of_speech_from_the_lemma(self) -> None:
		got = drawn(entry(VERB, "abacilēt", derived=[("abacilējoši", ADV), ("abacilējošs", ADJ)]))
		self.assertEqual(got, [("verb", "abacilēt", "")])

	def test_a_dump_with_no_entry_has_no_candidate(self) -> None:
		self.assertEqual(drawn(), [])


if __name__ == "__main__":
	unittest.main()
