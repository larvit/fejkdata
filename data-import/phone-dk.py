#!/usr/bin/env python3
"""Rebuild data/da_DK/phone-mobile.tsv and phone-fixed.tsv from Denmark's numbering plan.

    data-import/phone-dk.py [--out-mobile FILE] [--out-fixed FILE]

The plan is Annex 1 of Executive Order No. 1883 of 7 December 2020 on the Overall Danish Numbering
Plan, in force from 21 December 2020. It calls each range of 8-digit numbers, by their first two
digits, mainly mobile or mainly fixed. Its note (!) lists the three-digit prefixes that are mobile
inside the fixed ranges. A row is a three-digit prefix, written as "34 2". A prefix with third digit 0
is kept for expansion, and the plan reserves 67, 68, 83-85, 94 and 95, gives 80 to freephone, 90 to
premium rate services and 37 to 12-digit M2M numbers, so none of these is a row.
"""
import argparse
from pathlib import Path

import tsv

DATA = Path(__file__).resolve().parent.parent / "data" / "da_DK"
MOBILE = "20-29 30-31 40-42 50-53 60-61 71 81 91-93"
FIXED = "32-36 38-39 43-49 54-59 62-66 69 70 72-79 82 86-89 96-99"
MOBILE_IN_FIXED = (
    "342 344-349 356-357 359 362 365-366 389 398 431 441 462 466 468 472 474 476 478 485-486 488-489 "
    "493-496 498-499 542-543 545 551-552 556 571-574 577 579 584 586-587 589 597-598 627 629 641 649 658 "
    "662-665 667 692-694 697 771-772 782-783 785-786 788-789 826-827 829"
)


def numbers(spec):
    """The integers a spec such as "20-29 71" names."""
    out = set()
    for token in spec.split():
        first, _, last = token.partition("-")
        out.update(range(int(first), int(last or first) + 1))
    return out


def series():
    """The series of each kind, as three-digit integers."""
    note = numbers(MOBILE_IN_FIXED)
    both = {
        kind: {two * 10 + c for two in numbers(spec) for c in range(1, 10)}
        for kind, spec in (("mobile", MOBILE), ("fixed", FIXED))
    }
    if not note <= both["fixed"]:
        raise SystemExit(f"note (!) names series outside the fixed ranges: {sorted(note - both['fixed'])}")
    return {"mobile": both["mobile"] | note, "fixed": both["fixed"] - note}


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--out-mobile", default=str(DATA / "phone-mobile.tsv"))
    parser.add_argument("--out-fixed", default=str(DATA / "phone-fixed.tsv"))
    args = parser.parse_args()
    held = series()
    for kind, out in (("mobile", args.out_mobile), ("fixed", args.out_fixed)):
        tsv.write(out, ["prefix"], [{"prefix": f"{n // 10} {n % 10}"} for n in sorted(held[kind])])


if __name__ == "__main__":
    main()
