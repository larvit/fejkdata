#!/usr/bin/env python3
"""Rebuild data/en_US/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from WordNet 3.0.

    data-import/word-us.py [--wordnet URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

A row is a noun, verb, adjective or adverb that WordNet's index lists in lower-case letters alone, so a phrase, a proper
noun and a word with a hyphen or an apostrophe are left out. A verb carries its third-person present, the regular
spelling or one of REGULAR's exceptions. A word on LDNOOBW's English list of rude words is dropped. WordNet's sense
counts cover a quarter of its words, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import argparse
import re
import tarfile
import io
from pathlib import Path

import source
import tsv

WORDNET = "https://wordnetcode.princeton.edu/3.0/WNdb-3.0.tar.gz"
RUDE = "https://raw.githubusercontent.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words/master/en"
OUT = Path(__file__).resolve().parent.parent / "data" / "en_US"
CACHE = Path(__file__).resolve().parent / "cache"
LETTERS = re.compile(r"[a-z]+")
PARTS = ("noun", "verb", "adj", "adv")
# The verbs whose third person the spelling rule gets wrong.
REGULAR = {"be": "is", "have": "has", "quiz": "quizzes", "whiz": "whizzes"}


def present(verb):
    if verb in REGULAR:
        return REGULAR[verb]
    if re.search(r"[^aeiou]y$", verb):
        return verb[:-1] + "ies"
    if re.search(r"(s|x|z|ch|sh|o)$", verb):
        return verb + "es"
    return verb + "s"


def lemmas(tarball, part):
    """The lemmas of dict/index.<part>, whose license-header lines open with two spaces."""
    with tarfile.open(fileobj=io.BytesIO(tarball)) as archive:
        text = archive.extractfile(f"dict/index.{part}").read().decode("utf-8")
    return (line.split(" ", 1)[0] for line in text.split("\n") if line and not line.startswith("  "))


def tables(tarball, rude):
    rows = {}
    for part in PARTS:
        words = sorted({w for w in lemmas(tarball, part) if LETTERS.fullmatch(w) and w not in rude})
        rows[part] = [{"word": w, "present": present(w)} if part == "verb" else {"word": w} for w in words]
    return rows


def word_json(counts):
    items = ",\n".join(f'    {{ "format": "{{.word-{part}}}", "weight": {n} }}' for part, n in counts.items())
    fields = ",\n".join(
        '  "verb": { "format": "{.word-verb as verbword}{verbword}", "present": "{verbword.present}" }'
        if part == "verb"
        else f'  "{part}": "{{.word-{part}}}"'
        for part in PARTS
    )
    return '{\n  "format": "{any}",\n  "any": [\n' + items + "\n  ],\n" + fields + "\n}\n"


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--cache", default=str(CACHE))
    parser.add_argument("--wordnet", default=WORDNET)
    parser.add_argument("--rude", default=RUDE)
    parser.add_argument("--out", default=str(OUT))
    args = parser.parse_args()
    tarball = source.fetch(args.wordnet, args.cache, "WNdb-3.0.tar.gz", magic=b"\x1f\x8b")
    listed = source.fetch(args.rude, args.cache, "ldnoobw-en.txt").decode("utf-8").split("\n")
    rows = tables(tarball, {w.strip().lower() for w in listed})
    out = Path(args.out)
    for part in PARTS:
        tsv.write(out / f"word-{part}.tsv", ["word", "present"] if part == "verb" else ["word"], rows[part])
        (out / f"word-{part}.json").write_text(f'{{ "format": "{{word}}", "rows": "word-{part}.tsv" }}\n', encoding="utf-8")
    (out / "word.json").write_text(word_json({part: len(held) for part, held in rows.items()}), encoding="utf-8")


if __name__ == "__main__":
    main()
