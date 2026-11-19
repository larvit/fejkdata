"""A language model's recall of names, as a script lists it."""
import sys
from collections.abc import Collection

import tsv


def split(what: str, text: str, fixed: Collection[str] = ()) -> tuple[list[str], list[str]]:
    """The commoner names of a list, ranked, and the rarer ones below its `=` line; exits on a name listed twice, among these and `fixed`, or not in letters."""
    ordered, _, rarer = (part.split() for part in text.partition("\n=\n"))
    names = [*fixed, *ordered, *rarer]
    dupes = sorted({n for n in names if names.count(n) > 1})
    if dupes:
        sys.exit(f"{what}: listed twice: {', '.join(dupes)}")
    for name in names:
        if not tsv.NAME.match(name):
            sys.exit(f"{what}: {name!r} is not a name in letters")
    return ordered, rarer
