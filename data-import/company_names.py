"""What the company imports share: the words of a register's company names, the trade words among them, and the shapes and coined words their names make.

A trade word is neither a name nor a place, is carried by WORDS_MIN or more names, ends a name in at least a
third of its uses, and is carried by FORM_MIN names in each of at least form_spread of the forms a Vocabulary
names, so a word one company family repeats stays out. A Vocabulary's form_spread is three unless it sets
another. A word's spelling is the one most names that are not in capitals use. A word in capitals in such a name counts as written only if it is an acronym, such as VVS: it has
at most the vocabulary's acronym_max letters, or no vowel.

A shape is the kind of each word of a name: a first and last name drawn as one person (FS), a last name (S), a first name (F), a number
(N), a connecting word of the vocabulary, a trade word (T), or any other word (X), which draws a coined word. A coined word is no trade word,
has four letters or more, is carried by COINED_MIN or more names, and ends at least COINED_ENDS_MIN of them.
"""
import re
from collections import Counter, defaultdict
from collections.abc import Callable, Collection, Iterable, Mapping
from dataclasses import dataclass
from pathlib import Path
from typing import NamedTuple

SHARE_MIN = 0.01
WORDS_MIN = 100
FORM_MIN = 10
COINED_MIN = 40
COINED_ENDS_MIN = 0.05
VOWELS = "aeiouyåäöæøéüõāēīūąęėįų"
PERSONS = "pqrstuvw"
SLOTS = {"X": "{.company-coined}", "T": "{.company-word}", "S": "{.last-name}", "F": "{.first-name}", "N": "{int(1,30)}"}
# A name's kinds of word in order, each a key of SLOTS, a connecting word, or FS for a person.
type Shape = tuple[str, ...]


@dataclass(frozen=True)
class Vocabulary:
    """What a register's language makes of a name's words."""

    word: re.Pattern[str]
    function_words: frozenset[str]
    legal_words: frozenset[str]
    word_forms: frozenset[str]
    acronym_max: int
    connecting_words: tuple[str, ...]
    form_spread: int = 3


class Words(NamedTuple):
    surnames: Collection[str]
    firsts: Collection[str]


class Usage:
    """How the names of the register use one word."""

    def __init__(self) -> None:
        self.names, self.begins, self.ends = 0, 0, 0
        self.forms: Counter[str] = Counter()
        self.spellings: Counter[str] = Counter()


def column(path: str | Path) -> set[str]:
    """The first column of a shipped table, lower case."""
    return {line.split("\t")[0].lower() for line in Path(path).read_text(encoding="utf-8").splitlines()[1:]}


def split(name: str) -> list[str]:
    return [w for w in re.split(r"[\s,]+", name) if w]


def common[K](counts: Mapping[K, int]) -> list[tuple[K, int]]:
    """The keys with at least SHARE_MIN of the total, by count then key."""
    total = sum(counts.values())
    return [(k, c) for k, c in sorted(counts.items(), key=lambda kc: (-kc[1], kc[0])) if c >= SHARE_MIN * total]


def quoted(s: str) -> str:
    return '"' + s.replace("\\", "\\\\").replace('"', '\\"') + '"'


def weighted(fmt: str, weight: float) -> str:
    return "{ " + f'"format": {quoted(fmt)}, "weight": {weight}' + " }"


def legal_item(fmt: str, legal: str, weight: float) -> str:
    return "{ " + f'"format": {quoted(fmt)}, "legal": {quoted(legal)}, "weight": {weight}' + " }"


def word_usage(names: Iterable[tuple[str, str]], named: Callable[[str], object], vocabulary: Vocabulary) -> dict[str, Usage]:
    """The Usage of each word of the vocabulary's letters that is no name, place, legal or function word, by lower case. A name counts a word once."""
    usage: dict[str, Usage] = defaultdict(Usage)
    for form, name in names:
        words = [w for w in split(name) if w.lower() not in vocabulary.legal_words]
        seen: set[str] = set()
        for i, w in enumerate(words):
            low = w.lower()
            if vocabulary.word.match(low) and low not in vocabulary.function_words and not named(low):
                u = usage[low]
                if low not in seen:
                    seen.add(low)
                    u.names += 1
                    u.forms[form] += 1
                    if not name.isupper() and not (w.isupper() and len(w) > vocabulary.acronym_max and any(c in VOWELS for c in low)):
                        u.spellings[w[0].upper() + w[1:]] += 1
                u.begins += i == 0
                u.ends += i == len(words) - 1
    return usage


def trade_words(usage: Mapping[str, Usage], vocabulary: Vocabulary) -> dict[str, tuple[str, int]]:
    """The trade words, each with its spelling most names use and its count, by the rule in the module docstring."""
    return {
        w: (u.spellings.most_common(1)[0][0] if u.spellings else w.capitalize(), u.names)
        for w, u in usage.items()
        if u.names >= WORDS_MIN and 3 * u.ends >= u.names and sum(u.forms[f] >= FORM_MIN for f in vocabulary.word_forms) >= vocabulary.form_spread
    }


def kind(word: str, words: Words, last: bool) -> str | None:
    """S or F for a last name or a first name; else None. A word that is both is a last name at the end of a name."""
    w = word.lower()
    if w in words.surnames and (last or w not in words.firsts):
        return "S"
    return "F" if w in words.firsts else None


def shape_of(name: list[str], words: Words, trade: Collection[str], vocabulary: Vocabulary) -> Shape:
    """The kinds of the words, as SLOTS keys or connecting words, with FS for a first name followed by a last name."""
    shape: list[str] = []
    for i, word in enumerate(name):
        low = word.lower()
        if low.isdigit():
            shape.append("N")
        elif low in vocabulary.connecting_words:
            shape.append(low)
        else:
            k = kind(word, words, i == len(name) - 1) or ("T" if low in trade else "X")
            if k == "S" and shape and shape[-1] == "F":
                shape[-1] = "FS"
            else:
                shape.append(k)
    return tuple(shape)


def shipped(shapes: Mapping[Shape, int], vocabulary: Vocabulary) -> list[tuple[Shape, int]]:
    """The common shapes that open with no number or connecting word."""
    return [(s, c) for s, c in common(shapes) if s[0] != "N" and s[0] not in vocabulary.connecting_words]


def slots(shape: Shape, vocabulary: Vocabulary) -> str:
    """The format of a shape: its words' slots, each person drawn once under the next name of PERSONS."""
    persons = iter(PERSONS)
    out = []
    for k in shape:
        if k == "FS":
            p = next(persons)
            out.append(f"{{.person as {p}}}{{{p}.first}} {{{p}.last}}")
        elif k in vocabulary.connecting_words:
            out.append(k)
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


# Which side of the name carries the form, and whether the register quotes the rest of the name.
type Placement = tuple[str, bool]

# Each form's path under company and the words that spell it: its abbreviation first, then its words, in lower case.
type NamedForms = Mapping[str, tuple[str, tuple[str, ...]]]

LAYOUTS = {("before", True): '{F} "{n}"', ("before", False): "{F} {n}", ("after", False): "{n} {F}", ("after", True): '"{n}" {F}'}


def spelled(forms: NamedForms, form: str, words: list[str]) -> int:
    """How many words at the start of `words` spell the form, 0 where none."""
    low = [w.rstrip(".").lower() for w in words]
    for spelling in forms[form][1]:
        n = len(spelling.split())
        if low[:n] == spelling.split():
            return n
    return 0


def placed(forms: NamedForms, form: str, name: str) -> tuple[str, bool, list[str]] | None:
    """("before", quoted, the other words) where the name opens with its form, ("after", …) where it ends in it, else None."""
    words = split(name.replace('"', " "))
    quotes = '"' in name
    before = spelled(forms, form, words)
    if before and len(words) > before:
        return "before", quotes, words[before:]
    for n in range(1, len(words)):
        if spelled(forms, form, words[n:]) == len(words) - n:
            return "after", quotes, words[:n]
    return None


def bare(forms: NamedForms, names: Iterable[tuple[str, str]]) -> list[tuple[str, str]]:
    """The form and name of each name that places its form, without the form and its quotes."""
    out = []
    for form, name in names:
        found = placed(forms, form, name)
        if found:
            out.append((form, " ".join(found[2])))
    return out


def count_names(
    forms: NamedForms, names: Iterable[tuple[str, str]], words: Words, trade: Collection[str], vocabulary: Vocabulary
) -> tuple[dict[str, Counter[Shape]], dict[str, Counter[Placement]]]:
    """Per form: the shapes of the names that carry it, and where those names place it."""
    shapes: dict[str, Counter[Shape]] = {f: Counter() for f in forms}
    places: dict[str, Counter[Placement]] = {f: Counter() for f in forms}
    for form, name in names:
        found = placed(forms, form, name)
        if found:
            rest = [w for w in found[2] if w.rstrip(".").lower() not in vocabulary.legal_words]
            if rest:
                shapes[form][shape_of(rest, words, trade, vocabulary)] += 1
                places[form][found[0], found[1]] += 1
    return shapes, places


def form_block(forms: NamedForms, form: str, shapes: Mapping[Shape, int], places: Mapping[Placement, int], vocabulary: Vocabulary) -> str:
    """One form's object: its name by shape, and its abbreviation and quotes where its names put them."""
    path = forms[form][0]
    items = ",\n".join("      " + weighted(slots(shape, vocabulary), c) for shape, c in shipped(shapes, vocabulary))
    placements = ", ".join(weighted(fmt.replace("{F}", form).replace("{n}", "{" + path + "_name}"), places[layout]) for layout, fmt in LAYOUTS.items() if places.get(layout))
    return (
        f'  "{path}": {{\n    "format": {quoted("{name as " + path + "_name}{placed}")},\n    "name": [\n{items}\n    ],\n'
        f'    "hidden": {{ "placed": [{placements}] }}\n  }}'
    )


def company_json(
    forms: NamedForms, registered: Mapping[str, int], shapes: Mapping[str, Mapping[Shape, int]], places: Mapping[str, Mapping[Placement, int]], vocabulary: Vocabulary
) -> str:
    order = sorted(registered, key=lambda f: (-registered[f], f))
    items = [legal_item("{.company." + forms[f][0] + "}", f, registered[f]) for f in order]
    blocks = [form_block(forms, f, shapes[f], places[f], vocabulary) for f in forms]
    return '{\n  "format": "{registered as r}{r}",\n  "legal": "{r.legal}",\n  "registered": [\n    ' + ",\n    ".join(items) + "\n  ],\n" + ",\n".join(blocks) + "\n}\n"
