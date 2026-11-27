"""Wikidata's family names (CC0): the written forms of a name that a register spells in letters alone, and how many people of a country bear each."""
import hashlib
import json
import urllib.parse
from collections.abc import Callable, Iterable, Mapping, Sequence

import source

ENDPOINT = "https://query.wikidata.org/sparql"
BATCH = 4000
MIN_BEARERS = 5

Bearers = Callable[[Iterable[str], str, str, str], dict[str, int]]


def question(labels: Sequence[str], language: str, country: str) -> bytes:
	"""A SPARQL query, as a form body, counting for each label the people of the country whose family name carries that label in the language or in `mul`.
	A label is also asked with `’` for `'`, and both count as the label with `'`.
	The query runs in the order written: Wikidata's optimizer would start from every kind of family name, and time out."""
	forms = (form for label in labels for form in dict.fromkeys([label, label.replace("'", "’")]))
	values = " ".join(json.dumps(form, ensure_ascii=False) + "@" + tag for form in forms for tag in (language, "mul"))
	return urllib.parse.urlencode({"query": f"""
SELECT ?s (COUNT(DISTINCT ?p) AS ?people) WHERE {{
  hint:Query hint:optimizer "None".
  VALUES ?label {{ {values} }}
  ?n rdfs:label ?label; wdt:P31/wdt:P279* wd:Q101352.
  BIND(REPLACE(STR(?label), "’", "'") AS ?s)
  OPTIONAL {{ ?p wdt:P734 ?n; wdt:P27 wd:{country}. }}
}} GROUP BY ?s"""}).encode()


def bearers(labels: Iterable[str], language: str, country: str, cache: str) -> dict[str, int]:
	"""How many people of the country (a Wikidata item such as `Q142`) bear a family name carrying each label; a label that no family name carries is missing."""
	asked = sorted(set(labels))
	bodies = [question(asked[i : i + BATCH], language, country) for i in range(0, len(asked), BATCH)]

	def answer(body: bytes) -> dict[str, int]:
		name = "wikidata-names-" + hashlib.sha1(body).hexdigest()[:16] + ".json"
		got = source.fetch(ENDPOINT, cache, name, magic=b"{", data=body, headers={"Accept": "application/sparql-results+json"})
		return {b["s"]["value"]: int(b["people"]["value"]) for b in json.loads(got)["results"]["bindings"]}

	return {label: people for body in bodies for label, people in answer(body).items()}


def spelled(default: str, people: Mapping[str, int], floor: int) -> str:
	"""The form of a name that more than half of the people counted bear, if at least floor people do; else the default."""
	best = max(people, key=lambda form: (people[form], form), default=default)
	if people.get(best, 0) >= floor and 2 * people[best] > sum(people.values()):
		return best
	return default


def written(forms: Mapping[str, Sequence[str]], language: str, country: str, cache: str, ask: Bearers) -> dict[str, str]:
	"""Each name offered another form, mapped to the form that at least MIN_BEARERS of Wikidata's people of the country, and more than half of those counted, bear; else to the name itself."""
	offered = {name: [name, *others] for name, others in forms.items() if others}
	found = ask({form for each in offered.values() for form in each}, language, country, cache)
	return {name: spelled(name, {form: found.get(form, 0) for form in each}, MIN_BEARERS) for name, each in offered.items()}
