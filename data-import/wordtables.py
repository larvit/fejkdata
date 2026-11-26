"""A locale's word tables: a TSV and the JSON category reading it per part of speech, and word.json, which draws from them."""
import argparse
import json
import unicodedata
from collections.abc import Callable, Collection, Iterable, Mapping, Sequence
from dataclasses import dataclass
from pathlib import Path

import source
import tsv

RUDE = "https://raw.githubusercontent.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words/master/"
CACHE = Path(__file__).resolve().parent / "cache"

# A part of speech, a word of it and the word's form, which is empty where the part carries none.
Candidate = tuple[str, str, str]


@dataclass(frozen=True, kw_only=True)
class Source:
	"""One file an import reads: its flag, its URL, the bytes it opens with, the parser yielding its candidates,
	its name in the cache or "" for the URL's own, and the folder under --out its tables go to, "" for --out itself."""

	flag: str
	url: str
	magic: bytes
	candidates: Callable[[bytes], Iterable[Candidate]]
	name: str = ""
	folder: str = ""


def rude(language: str, cache: str | Path, url: str = "") -> set[str]:
	"""LDNOOBW's list of rude words in the language, lower-cased."""
	listed = source.fetch(url or RUDE + language, cache, f"ldnoobw-{language}.txt").decode("utf-8").split("\n")
	return {w.strip().lower() for w in listed if w.strip()}


def lower(word: str, _part: str) -> bool:
	"""Lower-case letters alone, whatever the part of speech."""
	return word.isalpha() and word.islower()


def lower_letters(word: str, _part: str) -> bool:
	"""Lower-case letters alone, as Unicode's Ll: that leaves out the modifier letters `lower` accepts."""
	return bool(word) and all(unicodedata.category(c) == "Ll" for c in word)


def rows(candidates: Iterable[Candidate], forms: Mapping[str, str], rude_words: Collection[str], spelled: Callable[[str, str], bool] = lower) -> dict[str, list[dict[str, str]]]:
	"""Per part of speech, its rows sorted by word: the first candidate of each word that is spelled and not rude,
	and whose form is too where the part carries one."""
	found: dict[str, dict[str, dict[str, str]]] = {part: {} for part in forms}
	for part, word, form in candidates:
		column = forms[part]
		if not spelled(word, part) or word.lower() in rude_words:
			continue
		if not column:
			found[part].setdefault(word, {"word": word})
		elif spelled(form, part) and form.lower() not in rude_words:
			found[part].setdefault(word, {"word": word, column: form})
	return {part: [words[w] for w in sorted(words)] for part, words in found.items()}


def inline(item: Mapping[str, str | int]) -> str:
	"""A JSON object on one line, padded inside its braces as the shipped files are."""
	return "{ " + json.dumps(item, ensure_ascii=False)[1:-1] + " }"


def word_json(forms: Mapping[str, str], counts: Mapping[str, int]) -> str:
	"""word.json: `any` weighs a part of speech by its rows, and a part with a form binds its word, so the form reads that word's own."""
	items = ",\n".join("    " + inline({"format": "{.word-" + part + "}", "weight": counts[part]}) for part in forms)
	fields = []
	for part, column in forms.items():
		name = part + "word"
		if column:
			field = inline({"format": "{.word-" + part + " as " + name + "}{" + name + "}", column: "{" + name + "." + column + "}"})
		else:
			field = json.dumps("{.word-" + part + "}")
		fields.append(f"  {json.dumps(part)}: {field}")
	return '{\n  "format": "{any}",\n  "any": [\n' + items + "\n  ],\n" + ",\n".join(fields) + "\n}\n"


def write(out: Path, forms: Mapping[str, str], table: Mapping[str, Sequence[Mapping[str, str]]]) -> None:
	"""Write each part's word-<part>.tsv and word-<part>.json, its form column after the word, and word.json."""
	for part, column in forms.items():
		tsv.write(out / f"word-{part}.tsv", ["word", column] if column else ["word"], table[part])
		(out / f"word-{part}.json").write_text(inline({"format": "{word}", "rows": f"word-{part}.tsv"}) + "\n", encoding="utf-8")
	(out / "word.json").write_text(word_json(forms, {part: len(table[part]) for part in forms}), encoding="utf-8")


def main(doc: str, language: str, out: Path, forms: Mapping[str, str], sources: Sequence[Source], spelled: Callable[[str, str], bool] = lower) -> None:
	"""Run a word import: read its flags, fetch each source, and LDNOOBW's list for the language unless the language is "", and write each source's tables."""
	parser = argparse.ArgumentParser(description=doc.splitlines()[0])
	parser.add_argument("--cache", default=str(CACHE))
	for src in sources:
		parser.add_argument(f"--{src.flag}", default=src.url)
	if language:
		parser.add_argument("--rude", default="")
	parser.add_argument("--out", default=str(out))
	args = parser.parse_args()
	listed = rude(language, args.cache, args.rude) if language else set()
	for src in sources:
		url = getattr(args, src.flag)
		body = source.fetch(url, args.cache, src.name or Path(url).name, magic=src.magic)
		write(Path(args.out) / src.folder, forms, rows(src.candidates(body), forms, listed, spelled))
