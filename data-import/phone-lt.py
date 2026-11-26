"""Rebuild data/lt_LT/phone-fixed-prefix.tsv from Lithuania's numbering plan.

    data-import/phone-lt.py [--out FILE]

The plan is the one the Communications Regulatory Authority (RRT) adopted on 1 September 2023 and
published through the ITU on 6 March 2025. Its table of the public fixed telephone service
(3000 0000 to 5999 9999) numbers an area code for each city or district and marks the rest "Reserved". A
row is one such code, with the trunk 0 in front as `area`, the code as `ndc` and a `subscriber` template
for the digits behind it. A number has 8 digits: Vilnius's code is 5, Kaunas's, Šiauliai's, Panevėžys's
and Klaipėda's have 2 digits, and a district's has 3. Where the plan fixes the first subscriber digits,
as in Vilnius's 20 to 27, the template holds them as a range.
"""
import argparse
from pathlib import Path

import tsv

DATA = Path(__file__).resolve().parent.parent / "data" / "lt_LT"

AREAS = [
	("5", "Vilnius"), ("37", "Kaunas"), ("41", "Šiauliai"), ("45", "Panevėžys"), ("46", "Klaipėda"),
	("310", "Varėna"), ("313", "Druskininkai"), ("315", "Alytus"), ("318", "Lazdijai"),
	("319", "Prienai and Birštonas"), ("340", "Ukmergė"), ("342", "Vilkaviškis"),
	("343", "Marijampolė"), ("345", "Šakiai"), ("346", "Kaišiadorys"), ("347", "Kėdainiai"),
	("349", "Jonava"), ("380", "Šalčininkai"), ("381", "Anykščiai"), ("382", "Širvintos"),
	("383", "Molėtai"), ("385", "Zarasai"), ("386", "Ignalina and Visaginas"), ("387", "Švenčionys"),
	("389", "Utena"), ("421", "Pakruojis"), ("422", "Radviliškis"), ("425", "Akmenė"),
	("426", "Joniškis"), ("427", "Kelmė"), ("428", "Raseiniai"), ("440", "Skuodas"),
	("441", "Šilutė"), ("443", "Mažeikiai"), ("444", "Telšiai"), ("445", "Kretinga"),
	("446", "Tauragė"), ("447", "Jurbarkas"), ("448", "Plungė"), ("449", "Šilalė"),
	("450", "Biržai"), ("451", "Pasvalys"), ("458", "Rokiškis"), ("459", "Kupiškis"),
	("460", "Palanga"), ("469", "Neringa"), ("528", "Trakai"),
]
# The subscriber digits the plan fixes after a code, as a range, and how many digits that is.
LEADS = {"5": ("2{int(0,7)}", 2), "45": ("{int(4,5)}", 1), "46": ("{int(2,4)}", 1)}
# The digits of a subscriber number that the first group holds, by the length of the area code.
FIRST_GROUP = {1: 3, 2: 2, 3: 5}


def subscriber(ndc_len: int, lead: str = "", lead_len: int = 0) -> str:
	"""The template of a subscriber number: the digits the plan fixes, then free ones, in the groups it is written in."""
	total = 8 - ndc_len
	first = FIRST_GROUP[ndc_len]
	groups = [lead + f"{{digits({first - lead_len})}}"]
	if total > first:
		groups.append(f"{{digits({total - first})}}")
	return " ".join(groups)


def rows() -> list[dict[str, str]]:
	return [
		{"area": "0" + code, "name": name, "ndc": code, "subscriber": subscriber(len(code), *LEADS.get(code, ("", 0)))}
		for code, name in AREAS
	]


def main() -> None:
	parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
	parser.add_argument("--out", default=str(DATA / "phone-fixed-prefix.tsv"))
	args = parser.parse_args()
	tsv.write(args.out, ["area", "name", "ndc", "subscriber"], rows(), templates={"subscriber"})


if __name__ == "__main__":
	main()
