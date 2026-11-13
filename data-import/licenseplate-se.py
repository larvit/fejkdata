#!/usr/bin/env python3
"""Rebuild data/sv_SE/licenseplate-letters.tsv: every three-letter combination of a Swedish plate's 23 letters but the blocked ones.

    data-import/licenseplate-se.py [--out FILE]

The 99 blocked combinations are the list that Transportstyrelsen's page "Spärrade bokstavskombinationer"
held when it was updated on 2026-03-31. Its update of 2026-08-30 removed them, so they are kept here.
"""
import argparse
import itertools
import sys
from pathlib import Path

import tsv

OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE" / "licenseplate-letters.tsv"
ALPHABET = "ABCDEFGHJKLMNOPRSTUWXYZ"
BLOCKED = """
APA ARG ASS BAJ BSS CUC CUK CUM DUM ETA ETT FAG FAN FEG FEL FEM FES FET FNL FUC FUK FUL GAM GAY GEJ GEY GHB
GUD GYN HAT HBT HKH HOR HOT KGB KKK KUC KUF KUG KUK KYK LAM LAT LEM LOJ LSD LUS MAD MAO MEN MES MLB MUS NAZ
NRP NSF NYP OND OOO ORM PAJ PKK PLO PMS PUB RAP RAS ROM RPS RUS SEG SEX SJU SOS SPY SUG SUP SUR TBC TOA TOK
TRE TYP UFO USA WAM WAR WWW XTC XTZ XUK XXL XXX ZEX ZOG ZPY ZUG ZUP ZOO
""".split()
EXPECTED = 99


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    if len(BLOCKED) != EXPECTED or len(set(BLOCKED)) != EXPECTED:
        sys.exit(f"{len(BLOCKED)} blocked combinations, {len(set(BLOCKED))} of them distinct, not {EXPECTED}")
    if stray := [c for c in BLOCKED if len(c) != 3 or set(c) - set(ALPHABET)]:
        sys.exit(f"blocked combinations that no plate could carry: {stray}")
    rows = [{"letters": "".join(c)} for c in itertools.product(ALPHABET, repeat=3) if "".join(c) not in BLOCKED]
    tsv.write(a.out, ["letters"], rows)


if __name__ == "__main__":
    main()
