"""Wikidata's family names (CC0): the written forms of a name that a register spells in plain letters, and how many people of a country bear each."""
import concurrent.futures
import hashlib
import json
import urllib.parse
from collections.abc import Callable, Iterable, Mapping, Sequence

import source

ENDPOINT = "https://query.wikidata.org/sparql"
BATCH = 4000
WORKERS = 4
MIN_BEARERS = 5

Bearers = Callable[[Iterable[str], str, str, str], dict[str, int]]


def question(labels: Sequence[str], language: str, country: str) -> bytes:
	"""A SPARQL question, as a form body, for the people of the country that each family name labelled so in the language has."""
	values = " ".join(json.dumps(label, ensure_ascii=False) + "@" + language for label in labels)
	return urllib.parse.urlencode({"query": f"""
SELECT ?label (COUNT(DISTINCT ?p) AS ?people) WHERE {{
  VALUES ?label {{ {values} }}
  ?n wdt:P31 wd:Q101352; rdfs:label ?label.
  OPTIONAL {{ ?p wdt:P734 ?n; wdt:P27 wd:{country}. }}
}} GROUP BY ?label"""}).encode()


def bearers(labels: Iterable[str], language: str, country: str, cache: str) -> dict[str, int]:
	"""The people of the country (a Wikidata item such as `Q142`) bearing each family name that is labelled so in the language; a label no family name has is missing."""
	asked = sorted(set(labels))
	bodies = [question(asked[i : i + BATCH], language, country) for i in range(0, len(asked), BATCH)]

	def answer(body: bytes) -> dict[str, int]:
		name = "wikidata-names-" + hashlib.sha1(body).hexdigest()[:16] + ".json"
		got = source.fetch(ENDPOINT, cache, name, magic=b"{", data=body, headers={"Accept": "application/sparql-results+json"})
		return {b["label"]["value"]: int(b["people"]["value"]) for b in json.loads(got)["results"]["bindings"]}

	with concurrent.futures.ThreadPoolExecutor(WORKERS) as pool:
		return {label: people for found in pool.map(answer, bodies) for label, people in found.items()}


def spelled(default: str, people: Mapping[str, int], floor: int) -> str:
	"""The form of a name that more than half of its bearers write, if at least floor people do; else the default."""
	best = max(people, key=lambda form: (people[form], form), default=default)
	if people.get(best, 0) >= floor and 2 * people[best] > sum(people.values()):
		return best
	return default


def written(forms: Mapping[str, Sequence[str]], language: str, country: str, cache: str, ask: Bearers) -> dict[str, str]:
	"""Each name, mapped to its written form: the name, or one of the forms offered for it that Wikidata's people of the country settle on."""
	offered = {name: [name, *others] for name, others in forms.items() if others}
	found = ask({form for each in offered.values() for form in each}, language, country, cache)
	return {name: spelled(name, {form: found.get(form, 0) for form in each}, MIN_BEARERS) for name, each in offered.items()}
