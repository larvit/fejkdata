"""Rebuild data/lt_LT/phone-fixed.tsv from Lithuania's numbering plan.

    data-import/phone-lt.py [--out FILE]

The plan is the one the Communications Regulatory Authority (RRT) adopted on 1 September 2023 and
published through the ITU on 6 March 2025, in the table of numbers of the fixed public telephone service
(3000 0000 to 5999 9999). A row is an area the table numbers for a city or district: its first digits
after the trunk 0, with the area code `ndc` and the digits of a subscriber number that follow it. The
table marks the rest "Reserved", so none of that is a row. A number has 8 digits, and the area code is
1 to 3 of them: 5 for Vilnius, 2 digits for Kaunas, Šiauliai, Panevėžys and Klaipėda, 3 for a district.
"""
import argparse
from pathlib import Path

import tsv

DATA = Path(__file__).resolve().parent.parent / "data" / "lt_LT"

# The area's first digits, the length of its code and its name.
AREAS = [
    ("310", 3, "Varėna"), ("313", 3, "Druskininkai"), ("315", 3, "Alytus"), ("318", 3, "Lazdijai"),
    ("319", 3, "Prienai and Birštonas"), ("340", 3, "Ukmergė"), ("342", 3, "Vilkaviškis"),
    ("343", 3, "Marijampolė"), ("345", 3, "Šakiai"), ("346", 3, "Kaišiadorys"), ("347", 3, "Kėdainiai"),
    ("349", 3, "Jonava"), ("37", 2, "Kaunas"), ("380", 3, "Šalčininkai"), ("381", 3, "Anykščiai"),
    ("382", 3, "Širvintos"), ("383", 3, "Molėtai"), ("385", 3, "Zarasai"),
    ("386", 3, "Ignalina and Visaginas"), ("387", 3, "Švenčionys"), ("389", 3, "Utena"),
    ("41", 2, "Šiauliai"), ("421", 3, "Pakruojis"), ("422", 3, "Radviliškis"), ("425", 3, "Akmenė"),
    ("426", 3, "Joniškis"), ("427", 3, "Kelmė"), ("428", 3, "Raseiniai"), ("440", 3, "Skuodas"),
    ("441", 3, "Šilutė"), ("443", 3, "Mažeikiai"), ("444", 3, "Telšiai"), ("445", 3, "Kretinga"),
    ("446", 3, "Tauragė"), ("447", 3, "Jurbarkas"), ("448", 3, "Plungė"), ("449", 3, "Šilalė"),
    ("450", 3, "Biržai"), ("451", 3, "Pasvalys"), ("454", 2, "Panevėžys"), ("455", 2, "Panevėžys"),
    ("458", 3, "Rokiškis"), ("459", 3, "Kupiškis"), ("460", 3, "Palanga"), ("462", 2, "Klaipėda"),
    ("463", 2, "Klaipėda"), ("464", 2, "Klaipėda"), ("469", 3, "Neringa"),
    *[(f"52{d}", 1, "Vilnius") for d in range(8)], ("528", 3, "Trakai"),
]
# The digits of a subscriber number that the first group holds, by the length of the area code.
FIRST_GROUP = {1: 3, 2: 2, 3: 5}


def subscriber(area: str, ndc_len: int) -> str:
    """The template of a subscriber number: the digits the area fixes, then free ones, in the groups it is written in."""
    fixed = area[ndc_len:]
    free = 8 - ndc_len - len(fixed)
    first = FIRST_GROUP[ndc_len]
    head = f"{fixed}{{digits({first - len(fixed)})}}" if first > len(fixed) else fixed
    return head if free == first - len(fixed) else f"{head} {{digits({free - (first - len(fixed))})}}"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--out", default=str(DATA / "phone-fixed.tsv"))
    args = parser.parse_args()
    rows = [
        {"area": area, "name": name, "ndc": area[:n], "subscriber": subscriber(area, n)}
        for area, n, name in AREAS
    ]
    tsv.write(args.out, ["area", "name", "ndc", "subscriber"], rows, templates={"subscriber"})


if __name__ == "__main__":
    main()
