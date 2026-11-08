#!/usr/bin/env python3
"""Rebuild data/en_US/word.json and word-noun.tsv, word-verb.tsv, word-adj.tsv and word-adv.tsv from WordNet 3.0.

    data-import/word-us.py [--wordnet URL_OR_FILE] [--rude URL_OR_FILE] [--cache DIR] [--out DIR]

A row is a noun, verb, adjective or adverb that WordNet's data files spell in lower-case letters alone, so a phrase, a
proper noun, an acronym and a word with a hyphen or an apostrophe are left out. WordNet's index files lower-case every
lemma, so the spelling comes from the data files. A verb carries its third-person present, by the spelling rule or
one of IRREGULAR. A word is dropped when every one of its senses is in WordNet's usage domain of disparagement,
ethnic slurs or obscenity, or when it is on LDNOOBW's English list of rude words. WordNet's sense counts cover a quarter of
its words, so each word is drawn evenly, and `word` weighs a part of speech by its row count.
"""
import argparse
import io
import re
import tarfile
from pathlib import Path

import source
import tsv

WORDNET = "https://wordnetcode.princeton.edu/3.0/WNdb-3.0.tar.gz"
RUDE = "https://raw.githubusercontent.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words/master/en"
OUT = Path(__file__).resolve().parent.parent / "data" / "en_US"
CACHE = Path(__file__).resolve().parent / "cache"
LETTERS = re.compile(r"[a-z]+")
RUDE_DOMAINS = {"disparagement", "ethnic_slur", "obscenity"}
PARTS = ("noun", "verb", "adj", "adv")
# The verbs whose third person the spelling rule gets wrong.
IRREGULAR = {"be": "is", "have": "has", "quiz": "quizzes", "stomach": "stomachs", "whiz": "whizzes"}
IRREGULAR.update({v: v + "s" for v in "bravo bunco crescendo decrescendo demo disco ditto kayo mambo solo tango velcro".split()})


def present(verb):
    if verb in IRREGULAR:
        return IRREGULAR[verb]
    if re.search(r"[^aeiou]y$", verb):
        return verb[:-1] + "ies"
    if re.search(r"([^aeiou]o|s|x|z|ch|sh)$", verb):
        return verb + "es"
    return verb + "s"


def synsets(tarball, part):
    """Each synset of dict/data.<part> as its offset, its words and the offsets it points to as `;u`, a usage domain."""
    with tarfile.open(fileobj=io.BytesIO(tarball)) as archive:
        text = archive.extractfile(f"dict/data.{part}").read().decode("utf-8")
    for line in text.split("\n"):
        if not line or line.startswith("  "):
            continue
        field = line.split("|")[0].split(" ")
        words = [re.sub(r"\(\w+\)$", "", field[4 + 2 * i]) for i in range(int(field[3], 16))]
        pointers = field[4 + 2 * len(words) :]
        domains = {pointers[k + 1] for k in range(1, 4 * int(pointers[0]), 4) if pointers[k] == ";u"}
        yield field[0], words, domains


def rude_domains(tarball):
    """The offsets of the usage domains RUDE_DOMAINS names, which are nouns."""
    return {offset for offset, words, _ in synsets(tarball, "noun") if RUDE_DOMAINS & set(words)}


def tables(tarball, rude):
    domains = rude_domains(tarball)
    rows = {}
    for part in PARTS:
        senses = [(w, d) for _, words, d in synsets(tarball, part) for w in words]
        kept = {w for w, d in senses if not d & domains}
        words = sorted({w for w, _ in senses if LETTERS.fullmatch(w) and w in kept and w not in rude})
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
