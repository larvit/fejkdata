"""A locale's word tables: a TSV and its manifest per part of speech, and word.json, which draws from them."""
from collections.abc import Mapping, Sequence
from pathlib import Path

import source
import tsv

RUDE = "https://raw.githubusercontent.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words/master/"


def rude(language: str, cache: str | Path, url: str = "") -> set[str]:
    """LDNOOBW's list of rude words in the language, lower-cased."""
    listed = source.fetch(url or RUDE + language, cache, f"ldnoobw-{language}.txt").decode("utf-8").split("\n")
    return {w.strip().lower() for w in listed if w.strip()}


def word_json(forms: Mapping[str, str], counts: Mapping[str, int]) -> str:
    """word.json: `any` weighs a part of speech by its rows, and a part with a form binds its word to read it."""
    items = ",\n".join(f'    {{ "format": "{{.word-{part}}}", "weight": {counts[part]} }}' for part in forms)
    fields = []
    for part, column in forms.items():
        if column:
            fields.append(f'  "{part}": {{ "format": "{{.word-{part} as {part}word}}{{{part}word}}", "{column}": "{{{part}word.{column}}}" }}')
        else:
            fields.append(f'  "{part}": "{{.word-{part}}}"')
    return '{\n  "format": "{any}",\n  "any": [\n' + items + "\n  ],\n" + ",\n".join(fields) + "\n}\n"


def write(out: Path, forms: Mapping[str, str], rows: Mapping[str, Sequence[Mapping[str, str]]]) -> None:
    """Write each part's word-<part>.tsv and word-<part>.json, its form column after the word, and word.json."""
    for part, column in forms.items():
        tsv.write(out / f"word-{part}.tsv", ["word", column] if column else ["word"], rows[part])
        (out / f"word-{part}.json").write_text(f'{{ "format": "{{word}}", "rows": "word-{part}.tsv" }}\n', encoding="utf-8")
    (out / "word.json").write_text(word_json(forms, {part: len(rows[part]) for part in forms}), encoding="utf-8")
