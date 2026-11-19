"""Rebuild data/et_EE/company.json, company-coined.{json,tsv} and company-word.{json,tsv} from the e-Business Register.

    data-import/company-ee.py [--register URL_OR_FILE] [--cache DIR] [--out DIR]

The script reads the register's basic data (a 19 MB zip) for counts and names, and writes no row of it. It also
reads the shipped et_EE first and last names. It counts the businesses entered into the register, those whose
ettevotja_staatus is R, and not those in liquidation or bankruptcy. company.json weighs a legal form by its
businesses. The forms are FIE, OÜ, AS, UÜ and TÜ; the non-profits, the apartment associations, the co-operatives and
every other form aren't drawn. An FIE is a person, and most of the register's FIE names are a first and a last
name alone, so an FIE's name is one person's and no shape is counted.

A name of the other forms is counted where its last or first word spells its form, as the abbreviation, such as
OÜ, or the word, such as osaühing, in upper or lower case, with or without a trailing dot. company.json puts the
abbreviation after a name as often as the form's names end in it, and before as often as they open with it.

A shape is the kind of each word of a name, its legal words left out: a person's first and last name, a last name, a
first name, a number, `ja`, `&`, a trade word, or any other, which draws a coined word. A last or first name is
one in the shipped et_EE tables, and a word that is both is a last name at the end of a name. A first name before
a last name is one person, drawn from et_EE.person so that the last name is of the first name's sex. A shape
ships when it is at least company_names.SHARE_MIN of the form's names and opens with no number, `ja` or `&`.

company-word.tsv holds the trade words, each with its count and its most common spelling, by the rule of
company_names.py over the names of OÜ, AS, UÜ and TÜ, with a form spread of one: OÜ names are 97.8 in 100 of
those names, so few trade words reach FORM_MIN names in a second form.

company-coined.tsv holds the coined words, each with the number of names that carry it. A coined word has four
letters or more, is no name or trade word, is carried by company_names.COINED_MIN or more names, and ends at least a
company_names.COINED_ENDS_MIN share of them. Nearly half of the register's names are one word no other name carries, and a
coined word is a common word of the names, not a root that opens them. company_names.COINED_MIN keeps out the brand of a
group's project companies, such as Hepsor's. A coined word takes its most common spelling, so a name drawn with it
may be a real company's.
"""
import argparse
import csv
import io
import re
import zipfile
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

REGISTER = "https://avaandmed.ariregister.rik.ee/sites/default/files/avaandmed/ettevotja_rekvisiidid__lihtandmed.csv.zip"
ROOT = Path(__file__).resolve().parent.parent / "data"
OUT = ROOT / "et_EE"
CACHE = Path(__file__).resolve().parent / "cache"
LEGAL_FORM = {"Füüsilisest isikust ettevõtja": "FIE", "Osaühing": "OÜ", "Aktsiaselts": "AS", "Usaldusühing": "UÜ", "Täisühing": "TÜ"}
# Each named form's path under company and the words that spell it: its abbreviation and its word, in lower case.
NAMED_FORMS = {"OÜ": ("ou", {"oü", "osaühing"}), "AS": ("as", {"as", "aktsiaselts"}), "UÜ": ("uu", {"uü", "usaldusühing"}), "TÜ": ("tu", {"tü", "täisühing"})}
CONNECTING_WORDS = ("ja", "&")
FUNCTION_WORDS = {*CONNECTING_WORDS, "the", "and", "of", "for"}
LEGAL_WORDS = {"fie", "mtü"} | {w for _, words in NAMED_FORMS.values() for w in words}
VOCABULARY = Vocabulary(re.compile(r"^[a-zäöüõšž]{3,}$"), frozenset(FUNCTION_WORDS), frozenset(LEGAL_WORDS), frozenset(NAMED_FORMS), 0, CONNECTING_WORDS, form_spread=1)



def register(data: bytes) -> tuple[Counter[str], list[tuple[str, str]]]:
    """The businesses entered into the register, counted by form, and the form and name of each whose form is in NAMED_FORMS."""
    counts: Counter[str] = Counter()
    names: list[tuple[str, str]] = []
    with zipfile.ZipFile(io.BytesIO(data)) as z, z.open(z.namelist()[0]) as f:
        for row in csv.DictReader(io.TextIOWrapper(f, encoding="utf-8-sig", newline=""), delimiter=";"):
            form = LEGAL_FORM.get(row["ettevotja_oiguslik_vorm"])
            if form and row["ettevotja_staatus"] == "R":
                counts[form] += 1
                if form in NAMED_FORMS:
                    names.append((form, row["nimi"]))
    return counts, names


def spells(form: str, word: str) -> bool:
    return word.rstrip(".").lower() in NAMED_FORMS[form][1]


def placed(form: str, name: str) -> tuple[str, list[str]] | None:
    """("after", the other words) where the name ends in its form, ("before", …) where it opens with it, else None."""
    words = split(name)
    if len(words) < 2:
        return None
    if spells(form, words[-1]):
        return "after", words[:-1]
    if spells(form, words[0]):
        return "before", words[1:]
    return None


def count_names(names: Iterable[tuple[str, str]], words: Words, trade: Collection[str]) -> tuple[dict[str, Counter[Shape]], dict[str, Counter[str]]]:
    """Per form: the shapes of the names that carry it, and where those names place it."""
    shapes: dict[str, Counter[Shape]] = {f: Counter() for f in NAMED_FORMS}
    places: dict[str, Counter[str]] = {f: Counter() for f in NAMED_FORMS}
    for form, name in names:
        found = placed(form, name)
        if found:
            rest = [w for w in found[1] if w.rstrip(".").lower() not in LEGAL_WORDS]
            if rest:
                shapes[form][shape_of(rest, words, trade, VOCABULARY)] += 1
                places[form][found[0]] += 1
    return shapes, places


def form_block(form: str, shapes: Mapping[Shape, int], places: Mapping[str, int]) -> str:
    """One form's object: its name by shape, and its abbreviation after or before it as its names place it."""
    path = NAMED_FORMS[form][0]
    bound = "{" + path + "_name}"
    items = ",\n".join("      " + weighted(slots(shape, VOCABULARY), c) for shape, c in shipped(shapes, VOCABULARY))
    placing = [(f"{bound} {form}", places["after"]), (f"{form} {bound}", places["before"])]
    placements = ", ".join(weighted(fmt, c) for fmt, c in placing if c)
    return (
        f'  "{path}": {{\n    "format": {quoted("{name as " + path + "_name}{placed}")},\n    "name": [\n{items}\n    ],\n'
        f'    "hidden": {{ "placed": [{placements}] }}\n  }}'
    )


def fie_block() -> str:
    return '  "fie": {\n    "format": "{name}",\n    "name": "{.person as p}{p.first} {p.last}"\n  }'


def company_json(registered: Mapping[str, int], shapes: Mapping[str, Mapping[Shape, int]], places: Mapping[str, Mapping[str, int]]) -> str:
    path = {"FIE": "fie"} | {f: p for f, (p, _) in NAMED_FORMS.items()}
    order = sorted(registered, key=lambda f: (-registered[f], f))
    forms = [legal_item("{.company." + path[f] + "}", "" if f == "FIE" else f, registered[f]) for f in order]
    blocks = [fie_block()] + [form_block(f, shapes[f], places[f]) for f in NAMED_FORMS]
    return '{\n  "format": "{registered as r}{r}",\n  "legal": "{r.legal}",\n  "registered": [\n    ' + ",\n    ".join(forms) + "\n  ],\n" + ",\n".join(blocks) + "\n}\n"


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--register", default=REGISTER)
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    registered, names = register(source.fetch(a.register, a.cache, "ariregister-lihtandmed.csv.zip", magic=b"PK"))
    words = Words(column(ROOT / "et_EE/last-name.tsv"), column(ROOT / "et_EE/first-name.tsv"))

    def named(w: str) -> bool:
        return kind(w, words, True) is not None

    usage = word_usage(names, named, VOCABULARY)
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
