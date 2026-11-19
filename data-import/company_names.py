"""What the company imports share: the words of a register's company names, and the trade words among them.

A trade word is neither a name nor a place, is carried by WORDS_MIN or more names, ends a name in at least a
third of its uses, and is carried by FORM_MIN names in each of at least form_spread of the forms a Vocabulary
names, so a word one company family repeats stays out. A Vocabulary's form_spread is three unless it sets
another. A word's spelling is the one most names that are not in capitals use. A word in capitals in such a name counts as written only if it is an acronym, such as VVS: it has
at most the vocabulary's acronym_max letters, or no vowel.
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
PERSONS = "pqrstuvw"
VOWELS = "aeiouyåäöæøéüõāēīū"
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
    form_spread: int = 3
    connecting_words: tuple[str, ...] = ()


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
    """The words of company-coined.tsv, each in its spelling with its count: no trade word, four letters or more, COINED_MIN or more names, ending COINED_ENDS_MIN of them."""
    return {
        (u.spellings.most_common(1)[0][0] if u.spellings else w.capitalize()): u.names
        for w, u in sorted(usage.items())
        if w not in trade and len(w) >= 4 and u.names >= COINED_MIN and u.ends >= COINED_ENDS_MIN * u.names
    }
