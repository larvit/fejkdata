"""Rebuild data/lv_LV/company.json, company-coined.{json,tsv} and company-word.{json,tsv} from the Register of Enterprises.

    data-import/company-lv.py [--register URL_OR_FILE] [--cache DIR] [--out DIR]

The script reads the register's open data (a 129 MB CSV) for counts and names, and writes no row of it. It also
reads the shipped lv_LV first and last names. It counts the merchants of the Commercial Register, those with no
termination date, by their `type`. company.json weighs a legal form by its merchants. The forms are SIA, IK, AS, PS
and KS; the branches, the European companies and every register but the Commercial Register aren't drawn.

A name is counted where its first or last words spell its form, as the abbreviation, such as SIA, or the words,
such as sabiedrība ar ierobežotu atbildību, in upper or lower case, with or without a trailing dot. company.json
puts the abbreviation before the rest of a name as often as the form's names open with their form, and after it as
often as they end in it, and puts the rest in quotes as often as those names do, as `SIA "Alfa"` does. A name that
spells its form out counts as one with the abbreviation: the register writes `Sabiedrība ar ierobežotu atbildību
"Alfa"` for about 4 in 10 of its SIA.

A shape is the kind of each word of a name, its legal words left out: a person's first and last name, a last name, a
first name, a number, `un`, `&`, a trade word, or any other, which draws a coined word. A last or first name is
one in the shipped lv_LV tables, and a word that is both is a last name at the end of a name. A first name before
a last name is one person, drawn from lv_LV.person so that the last name is of the first name's sex. A shape
ships when it is at least company_names.SHARE_MIN of the form's names and opens with no number, `un` or `&`.

company-word.tsv holds the trade words, each with its count and its most common spelling, by the rule of
company_names.py over the names of every form, with a form spread of one: SIA names are about 9 in 10 of those names,
so few trade words reach FORM_MIN names in a second form.

company-coined.tsv holds the coined words, each with the number of names that carry it. A coined word has four
letters or more, is no name or trade word, is carried by COINED_MIN or more names, and ends at least a
COINED_ENDS_MIN share of them. A coined word is a common word of the names, not a root that opens them. A coined
word takes its most common spelling. A name drawn from these parts may be a real company's.
"""
import argparse
import csv
import io
import re
from collections import Counter
from collections.abc import Collection, Iterable, Mapping
from pathlib import Path
from typing import NamedTuple

import source
import tsv
from company_names import (
    Usage,
    Vocabulary,
    column,
    common,
    legal_item,
    quoted,
    split,
    trade_words,
    weighted,
    word_usage,
)

REGISTER = "https://data.gov.lv/dati/dataset/4de9697f-850b-45ec-8bba-61fa09ce932f/resource/25e80bf3-f107-4ab4-89ef-251b5b9374e9/download/register.csv"
ROOT = Path(__file__).resolve().parent.parent / "data"
OUT = ROOT / "lv_LV"
CACHE = Path(__file__).resolve().parent / "cache"
COMMERCIAL_REGISTER = "K"
# Each form's path under company and the words that spell it: its abbreviation and its words, in lower case.
NAMED_FORMS = {
    "SIA": ("sia", ("sia", "sabiedrība ar ierobežotu atbildību")),
    "IK": ("ik", ("ik", "individuālais komersants")),
    "AS": ("as", ("as", "akciju sabiedrība")),
    "PS": ("ps", ("ps", "pilnsabiedrība")),
    "KS": ("ks", ("ks", "komandītsabiedrība")),
}
COINED_MIN = 40
COINED_ENDS_MIN = 0.05
CONNECTING_WORDS = ("un", "&")
FUNCTION_WORDS = {*CONNECTING_WORDS, "the", "and", "of", "for"}
LEGAL_WORDS = {words[0] for _, words in NAMED_FORMS.values()} | {"pilnsabiedrība", "komandītsabiedrība"}
VOCABULARY = Vocabulary(re.compile(r"^[a-zāčēģīķļņšūž]{3,}$"), frozenset(FUNCTION_WORDS), frozenset(LEGAL_WORDS), frozenset(NAMED_FORMS), 0, form_spread=1)
PERSONS = "pqrstuvw"
# A name's kinds of word in order, each a key of SLOTS, or FS for a person.
type Shape = tuple[str, ...]
# Which side of the name carries the form, and whether the register quotes the rest of the name.
type Placement = tuple[str, bool]

SLOTS = {"X": "{.company-coined}", "T": "{.company-word}", "S": "{.last-name}", "F": "{.first-name}", "N": "{int(1,30)}", **{w: w for w in CONNECTING_WORDS}}
LAYOUTS = {("before", True): '{F} "{n}"', ("before", False): "{F} {n}", ("after", False): "{n} {F}", ("after", True): '"{n}" {F}'}


class Words(NamedTuple):
    surnames: Collection[str]
    firsts: Collection[str]


def register(data: bytes) -> tuple[Counter[str], list[tuple[str, str]]]:
    """The merchants of the Commercial Register with no termination date, counted by form, and the form and name of each."""
    counts: Counter[str] = Counter()
    names: list[tuple[str, str]] = []
    for row in csv.DictReader(io.TextIOWrapper(io.BytesIO(data), encoding="utf-8-sig", newline=""), delimiter=";"):
        form = row["type"]
        if form in NAMED_FORMS and row["regtype"] == COMMERCIAL_REGISTER and not row["terminated"]:
            counts[form] += 1
            names.append((form, row["name"]))
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


def kind(word: str, words: Words, last: bool) -> str | None:
    """S or F for a last name or a first name; else None. A word that is both is a last name at the end of a name."""
    w = word.lower()
    if w in words.surnames and (last or w not in words.firsts):
        return "S"
    return "F" if w in words.firsts else None


def shape_of(name: list[str], words: Words, trade: Collection[str]) -> Shape:
    """The kinds of the words, as SLOTS keys, with FS for a first name followed by a last name."""
    shape: list[str] = []
    for i, word in enumerate(name):
        low = word.lower()
        if low.isdigit():
            shape.append("N")
        elif low in CONNECTING_WORDS:
            shape.append(low)
        else:
            k = kind(word, words, i == len(name) - 1) or ("T" if low in trade else "X")
            if k == "S" and shape and shape[-1] == "F":
                shape[-1] = "FS"
            else:
                shape.append(k)
    return tuple(shape)


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
                shapes[form][shape_of(rest, words, trade)] += 1
                places[form][found[0], found[1]] += 1
    return shapes, places


def shipped(shapes: Mapping[Shape, int]) -> list[tuple[Shape, int]]:
    """The common shapes that open with no number, `un` or `&`."""
    return [(s, c) for s, c in common(shapes) if s[0] != "N" and s[0] not in CONNECTING_WORDS]


def slots(shape: Shape) -> str:
    """The format of a shape: its words' slots, each person drawn once under the next name of PERSONS."""
    persons = iter(PERSONS)
    out = []
    for k in shape:
        if k == "FS":
            p = next(persons)
            out.append(f"{{.person as {p}}}{{{p}.first}} {{{p}.last}}")
        else:
            out.append(SLOTS[k])
    return " ".join(out)


def coined(usage: Mapping[str, Usage], trade: Collection[str]) -> dict[str, int]:
    """The words of company-coined.tsv, each in its spelling with its count, by the rule in the module docstring."""
    return {
        (u.spellings.most_common(1)[0][0] if u.spellings else w.capitalize()): u.names
        for w, u in sorted(usage.items())
        if w not in trade and len(w) >= 4 and u.names >= COINED_MIN and u.ends >= COINED_ENDS_MIN * u.names
    }


def form_block(form: str, shapes: Mapping[Shape, int], places: Mapping[Placement, int]) -> str:
    """One form's object: its name by shape, and its abbreviation and quotes where its names put them."""
    path = NAMED_FORMS[form][0]
    items = ",\n".join("      " + weighted(slots(shape), c) for shape, c in shipped(shapes))
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
    registered, names = register(source.fetch(a.register, a.cache, "uznemumu-registrs-register.csv"))
    words = Words(column(ROOT / "lv_LV/last-name.tsv"), column(ROOT / "lv_LV/first-name.tsv"))

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
