"""Rebuild data/lt_LT/company.json, company-coined.{json,tsv} and company-word.{json,tsv} from the Register of Legal Entities.

    data-import/company-lt.py [--register URL_OR_FILE] [--cache DIR] [--out DIR]

The script reads the Centre of Registers' open data on the legal entities in the register (a 41 MB CSV) for counts
and names, and writes no row of it. It also reads the shipped lt_LT first and last names. It counts the entities by
their legal form. company.json weighs a legal form by its entities. The forms are UAB, MB, AB and ŽŪB, the forms of
a business whose name the register publishes. The register names an individual enterprise, a limited partnership and
the like by its form alone, and every non-profit form, such as the public body VšĮ, isn't a company; none is drawn.

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
from collections.abc import Collection, Iterable, Mapping
from pathlib import Path

import source
import tsv
from company_names import (
    Shape,
    Vocabulary,
    Words,
    coined,
    column,
    kind,
    legal_item,
    quoted,
    shape_of,
    shipped,
    slots,
    split,
    trade_words,
    weighted,
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
# Which side of the name carries the form, and whether the register quotes the rest of the name.
type Placement = tuple[str, bool]

LAYOUTS = {("before", True): '{F} "{n}"', ("before", False): "{F} {n}", ("after", False): "{n} {F}", ("after", True): '"{n}" {F}'}


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


def spelled(form: str, words: list[str]) -> int:
    """How many words at the start of `words` spell the form, 0 where none."""
    low = [w.rstrip(".").lower() for w in words]
    for spelling in NAMED_FORMS[form][1]:
        n = len(spelling.split())
        if low[:n] == spelling.split():
            return n
    return 0


def placed(form: str, name: str) -> tuple[str, bool, list[str]] | None:
    """("before", quoted, the other words) where the name opens with its form, ("after", …) where it ends in it, else None."""
    words = split(name.replace('"', " "))
    quotes = '"' in name
    before = spelled(form, words)
    if before and len(words) > before:
        return "before", quotes, words[before:]
    for n in range(1, len(words)):
        if spelled(form, words[n:]) == len(words) - n:
            return "after", quotes, words[:n]
    return None


def bare(names: Iterable[tuple[str, str]]) -> list[tuple[str, str]]:
    """The form and name of each name that places its form, without the form and its quotes."""
    out = []
    for form, name in names:
        found = placed(form, name)
        if found:
            out.append((form, " ".join(found[2])))
    return out


def count_names(names: Iterable[tuple[str, str]], words: Words, trade: Collection[str]) -> tuple[dict[str, Counter[Shape]], dict[str, Counter[Placement]]]:
    """Per form: the shapes of the names that carry it, and where those names place it."""
    shapes: dict[str, Counter[Shape]] = {f: Counter() for f in NAMED_FORMS}
    places: dict[str, Counter[Placement]] = {f: Counter() for f in NAMED_FORMS}
    for form, name in names:
        found = placed(form, name)
        if found:
            rest = [w for w in found[2] if w.rstrip(".").lower() not in LEGAL_WORDS]
            if rest:
                shapes[form][shape_of(rest, words, trade, VOCABULARY)] += 1
                places[form][found[0], found[1]] += 1
    return shapes, places


def form_block(form: str, shapes: Mapping[Shape, int], places: Mapping[Placement, int]) -> str:
    """One form's object: its name by shape, and its abbreviation and quotes where its names put them."""
    path = NAMED_FORMS[form][0]
    items = ",\n".join("      " + weighted(slots(shape, VOCABULARY), c) for shape, c in shipped(shapes, VOCABULARY))
    placements = ", ".join(weighted(fmt.replace("{F}", form).replace("{n}", "{" + path + "_name}"), places[layout]) for layout, fmt in LAYOUTS.items() if places.get(layout))
    return (
        f'  "{path}": {{\n    "format": {quoted("{name as " + path + "_name}{placed}")},\n    "name": [\n{items}\n    ],\n'
        f'    "hidden": {{ "placed": [{placements}] }}\n  }}'
    )


def company_json(registered: Mapping[str, int], shapes: Mapping[str, Mapping[Shape, int]], places: Mapping[str, Mapping[Placement, int]]) -> str:
    order = sorted(registered, key=lambda f: (-registered[f], f))
    forms = [legal_item("{.company." + NAMED_FORMS[f][0] + "}", f, registered[f]) for f in order]
    blocks = [form_block(f, shapes[f], places[f]) for f in NAMED_FORMS]
    return '{\n  "format": "{registered as r}{r}",\n  "legal": "{r.legal}",\n  "registered": [\n    ' + ",\n    ".join(forms) + "\n  ],\n" + ",\n".join(blocks) + "\n}\n"


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

    usage = word_usage(bare(names), named, VOCABULARY)
    trade = trade_words(usage, VOCABULARY)
    shapes, places = count_names(names, words, trade)
    out = Path(a.out)
    tsv.write(out / "company-word.tsv", ["word", "count"], [{"word": s, "count": c} for _, (s, c) in sorted(trade.items())])
    tsv.write(out / "company-coined.tsv", ["word", "count"], [{"word": w, "count": c} for w, c in coined(usage, trade).items()])
    for table in ("company-word", "company-coined"):
        (out / f"{table}.json").write_text('{ "format": "{word}", "rows": "' + table + '.tsv", "weight": "count" }\n', encoding="utf-8")
    (out / "company.json").write_text(company_json(registered, shapes, places), encoding="utf-8")


if __name__ == "__main__":
    main()
