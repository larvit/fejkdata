"""Rebuild data/lt_LT/company.json, company-coined.{json,tsv} and company-word.{json,tsv} from the Register of Legal Entities.

    data-import/company-lt.py [--register URL_OR_FILE] [--cache DIR] [--out DIR]

The script reads the Centre of Registers' open data on the legal entities in the register (a 41 MB CSV) for counts
and names, and writes no row of it. It also reads the shipped lt_LT first and last names. It counts the entities by
their legal form. company.json weighs a legal form by its entities. The forms are UAB, MB, AB and ŽŪB. The
register names an individual enterprise, a limited partnership and a general partnership by their form alone. A
branch, a representative office, a cooperative, a law partnership and every non-profit or public form are left out
too; none is drawn.

A name is counted where its first or last words spell its form, as the abbreviation, such as UAB, or the words,
such as uždaroji akcinė bendrovė, in upper or lower case, with or without a trailing dot. company.json
puts the abbreviation before the rest of a name as often as the form's names open with their form, and after it as
often as they end in it, and puts the rest in quotes as often as those names do, as `UAB "Alfa"` does. A name that
spells its form out counts as one with the abbreviation: the register writes `Uždaroji akcinė bendrovė "Alfa"` for
about 1 in 10 of its UAB.

A shape is the kind of each word of a name, its legal words left out: a person's first and last name, a last name, a
first name, a number, `ir`, `&`, a trade word, or any other, which draws a coined word. A last or first name is
one in the shipped lt_LT tables, and a word that is both is a last name at the end of a name. A first name before
a last name is one person, drawn from lt_LT.person so that the last name is of the first name's sex. A shape
ships when it is at least company_names.SHARE_MIN of the form's names and opens with no number, `ir` or `&`.

company-word.tsv holds the trade words, each with its count and its most common spelling, by the rule of
company_names.py over the names of every form, with a form spread of one: UAB names are about 6 in 10 of those names,
so few trade words reach FORM_MIN names in a second form.

company-coined.tsv holds the coined words, each with the number of names that carry it. A coined word has four
letters or more, is no name or trade word, is carried by company_names.COINED_MIN or more names, and ends at least a
company_names.COINED_ENDS_MIN share of them. A coined word is a common word of the names, not a root that opens them. A coined
word takes its most common spelling. A name drawn from these parts may be a real company's.
"""
import argparse
import csv
import io
import re
from collections import Counter
from pathlib import Path

import source
import tsv
from company_names import (
	Vocabulary,
	Words,
	bare,
	coined,
	column,
	company_json,
	count_names,
	kind,
	trade_words,
	word_usage,
)

REGISTER = "https://www.registrucentras.lt/aduomenys/?byla=JAR_IREGISTRUOTI.csv"
ROOT = Path(__file__).resolve().parent.parent / "data"
OUT = ROOT / "lt_LT"
CACHE = Path(__file__).resolve().parent / "cache"
# The register's code of each form drawn.
FORM_CODES = {"310": "UAB", "960": "MB", "320": "AB", "410": "ŽŪB"}
# Each form's path under company and the words that spell it: its abbreviation and its words, in lower case.
NAMED_FORMS = {
	"UAB": ("uab", ("uab", "uždaroji akcinė bendrovė")),
	"MB": ("mb", ("mb", "mažoji bendrija")),
	"AB": ("ab", ("ab", "akcinė bendrovė")),
	"ŽŪB": ("zub", ("žūb", "žemės ūkio bendrovė")),
}
CONNECTING_WORDS = ("ir", "&")
FUNCTION_WORDS = {*CONNECTING_WORDS, "the", "and", "of", "for"}
LEGAL_WORDS = {words[0] for _, words in NAMED_FORMS.values()}
VOCABULARY = Vocabulary(re.compile(r"^[a-ząčęėįšųūž]{3,}$"), frozenset(FUNCTION_WORDS), frozenset(LEGAL_WORDS), frozenset(NAMED_FORMS), 0, CONNECTING_WORDS, form_spread=1)
def register(data: bytes) -> tuple[Counter[str], list[tuple[str, str]]]:
	"""The legal entities of a form drawn, counted by form, and the form and name of each."""
	counts: Counter[str] = Counter()
	names: list[tuple[str, str]] = []
	for row in csv.DictReader(io.TextIOWrapper(io.BytesIO(data), encoding="utf-8-sig", newline=""), delimiter="|"):
		form = FORM_CODES.get(row["form_kodas"])
		if form:
			counts[form] += 1
			names.append((form, row["ja_pavadinimas"]))
	return counts, names


def main() -> None:
	p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
	p.add_argument("--register", default=REGISTER)
	p.add_argument("--cache", default=str(CACHE))
	p.add_argument("--out", default=str(OUT))
	a = p.parse_args()
	registered, names = register(source.fetch(a.register, a.cache, "juridiniu-asmenu-registras-iregistruoti.csv", b"ja_kodas|"))
	words = Words(column(ROOT / "lt_LT/last-name.tsv"), column(ROOT / "lt_LT/first-name.tsv"))

	def named(w: str) -> bool:
		return kind(w, words, True) is not None

	usage = word_usage(bare(NAMED_FORMS, names), named, VOCABULARY)
	trade = trade_words(usage, VOCABULARY)
	shapes, places = count_names(NAMED_FORMS, names, words, trade, VOCABULARY)
	out = Path(a.out)
	tsv.write(out / "company-word.tsv", ["word", "count"], [{"word": s, "count": c} for _, (s, c) in sorted(trade.items())])
	tsv.write(out / "company-coined.tsv", ["word", "count"], [{"word": w, "count": c} for w, c in coined(usage, trade).items()])
	for table in ("company-word", "company-coined"):
		(out / f"{table}.json").write_text('{ "format": "{word}", "rows": "' + table + '.tsv", "weight": "count" }\n', encoding="utf-8")
	(out / "company.json").write_text(company_json(NAMED_FORMS, registered, shapes, places, VOCABULARY), encoding="utf-8")


if __name__ == "__main__":
	main()
