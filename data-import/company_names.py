"""What the company imports share: the words of a register's company names, and the trade words among them.

A trade word is neither a name nor a place, is carried by WORDS_MIN or more names, ends a name in at least a
third of its uses, and is carried by FORM_MIN names in each of at least three of the forms a Vocabulary names,
so a word one company family repeats stays out. A word's spelling is the one most names that are not in
capitals use. A word in capitals in such a name counts as written only if it is an acronym, such as VVS: it has
at most the vocabulary's acronym_max letters, or no vowel.
"""
import re
from collections import Counter, defaultdict
from collections.abc import Callable, Iterable, Mapping
from dataclasses import dataclass
from pathlib import Path

SHARE_MIN = 0.01
WORDS_MIN = 100
FORM_MIN = 10
VOWELS = "aeiouyåäöæøéü"


@dataclass(frozen=True)
class Vocabulary:
    """What a register's language makes of a name's words."""

    word: re.Pattern[str]
    function_words: frozenset[str]
    legal_words: frozenset[str]
    word_forms: frozenset[str]
    acronym_max: int


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
    """The Usage of each word of the vocabulary's letters that is no name, place, legal or function word, by lower case."""
    usage: dict[str, Usage] = defaultdict(Usage)
    for form, name in names:
        words = [w for w in split(name) if w.lower() not in vocabulary.legal_words]
        for i, w in enumerate(words):
            low = w.lower()
            if vocabulary.word.match(low) and low not in vocabulary.function_words and not named(low):
                u = usage[low]
                u.names += 1
                u.forms[form] += 1
                u.begins += i == 0
                u.ends += i == len(words) - 1
                if not name.isupper() and not (w.isupper() and len(w) > vocabulary.acronym_max and any(c in VOWELS for c in low)):
                    u.spellings[w[0].upper() + w[1:]] += 1
    return usage


def trade_words(usage: Mapping[str, Usage], vocabulary: Vocabulary) -> dict[str, tuple[str, int]]:
    """The trade words, each with its spelling most names use and its count, by the rule in the module docstring."""
    return {
        w: (u.spellings.most_common(1)[0][0] if u.spellings else w.capitalize(), u.names)
        for w, u in usage.items()
        if u.names >= WORDS_MIN and 3 * u.ends >= u.names and sum(u.forms[f] >= FORM_MIN for f in vocabulary.word_forms) >= 3
    }
