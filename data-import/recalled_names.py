"""A language model's recall of names, as a script lists it."""
import sys
from collections.abc import Collection

import tsv


def split(what: str, text: str, fixed: Collection[str] = ()) -> tuple[list[str], list[str]]:
    """The commoner names of a list, ranked, and the rarer ones below its `=` line; exits on a name listed twice, among these and `fixed`, or not in letters."""
    ordered, _, rarer = (tsv.straight(part).split() for part in text.partition("\n=\n"))
    names = [*fixed, *ordered, *rarer]
    dupes = sorted({n for n in names if names.count(n) > 1})
    if dupes:
        sys.exit(f"{what}: listed twice: {', '.join(dupes)}")
    for name in names:
        if not tsv.NAME.match(name):
            sys.exit(f"{what}: {name!r} is not a name in letters")
    return ordered, rarer


def weighted(what: str, text: str, top: int) -> list[tuple[str, int]]:
    """The rows (name, count) of a recalled list: the commoner names weigh `top` over the square root of their rank, and the rarer ones the last rank's weight."""
    ordered, rarer = split(what, text)
    rows = [(name, round(top / rank ** 0.5)) for rank, name in enumerate(ordered, start=1)]
    return rows + [(name, round(top / len(ordered) ** 0.5)) for name in rarer]
