"""What the company imports share: the words of a register's company names, and the trade words among them.

A trade word is neither a name nor a place, is carried by WORDS_MIN or more names, ends a name in at least a
third of its uses, and is carried by FORM_MIN names in each of at least three of the forms a Vocabulary names,
so a word one company family repeats stays out. A word's spelling is the one most names that are not in
capitals use; a word in capitals there is an acronym, such as VVS, only up to the vocabulary's acronym_max letters or where it holds no vowel.
"""
import re
from collections import Counter, defaultdict
from dataclasses import dataclass
from pathlib import Path

SHARE_MIN = 0.01
WORDS_MIN = 100
FORM_MIN = 10
VOWELS = "aeiouyåäöæøéü"


@dataclass(frozen=True)
class Vocabulary:
    """What a register's language makes of a name's words."""

    word: re.Pattern
    function_words: frozenset
    legal_words: frozenset
    word_forms: frozenset
    acronym_max: int


class Usage:
    """How the names of the register use one word."""

    def __init__(self):
        self.names, self.begins, self.ends = 0, 0, 0
        self.forms, self.spellings = Counter(), Counter()


def column(path):
    """The first column of a shipped table, lower case."""
    return {line.split("\t")[0].lower() for line in Path(path).read_text(encoding="utf-8").splitlines()[1:]}


def split(name):
    return [w for w in re.split(r"[\s,]+", name) if w]


def common(counts):
    """The keys with at least SHARE_MIN of the total, by count then key."""
    total = sum(counts.values())
    return [(k, c) for k, c in sorted(counts.items(), key=lambda kc: (-kc[1], kc[0])) if c >= SHARE_MIN * total]


def quoted(s):
    return '"' + s.replace("\\", "\\\\").replace('"', '\\"') + '"'


def weighted(fmt, weight):
    return "{ " + f'"format": {quoted(fmt)}, "weight": {weight}' + " }"


def word_usage(names, named, vocabulary):
    """The Usage of each word of the vocabulary's letters that is no name, place, legal or function word, by lower case."""
    usage = defaultdict(Usage)
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


def trade_words(usage, vocabulary):
    """The trade words, each with its spelling most names use and its count, by the rule in the module docstring."""
    return {
        w: (u.spellings.most_common(1)[0][0] if u.spellings else w.capitalize(), u.names)
        for w, u in usage.items()
        if u.names >= WORDS_MIN and 3 * u.ends >= u.names and sum(u.forms[f] >= FORM_MIN for f in vocabulary.word_forms) >= 3
    }
