#!/usr/bin/env python3
"""Rebuild data/sv_SE/month.tsv, weekday.tsv and territory.tsv from Unicode CLDR's Swedish names.

    data-import/cldr-sv.py [--dates URL_OR_FILE] [--territories URL_OR_FILE] [--cache DIR] [--out-dir DIR] [--misc FILE]

A month is numbered 1 to 12 and a weekday 1 (Monday) to 7 (Sunday), as ISO 8601 does. The
territory table holds one row for each alpha-2 code in data/misc/territory.tsv; when CLDR has no
Swedish name for a code, the script exits with an error before it writes territory.tsv.
"""
import argparse
import csv
import json
import sys
from pathlib import Path
from typing import Any

import source
import tsv

CLDR = "https://raw.githubusercontent.com/unicode-org/cldr-json/48.0.0/cldr-json"
DATES = f"{CLDR}/cldr-dates-full/main/sv/ca-gregorian.json"
TERRITORIES = f"{CLDR}/cldr-localenames-full/main/sv/territories.json"
ROOT = Path(__file__).resolve().parent.parent / "data"
CACHE = Path(__file__).resolve().parent / "cache"
WEEKDAYS = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"]


def load(path: str, cache: str, name: str) -> Any:
    return json.loads(source.fetch(path, cache, name).decode("utf-8"))["main"]["sv"]


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--dates", default=DATES)
    p.add_argument("--territories", default=TERRITORIES)
    p.add_argument("--misc", default=str(ROOT / "misc" / "territory.tsv"))
    p.add_argument("--out-dir", default=str(ROOT / "sv_SE"))
    a = p.parse_args()
    out = Path(a.out_dir)
    gregorian = load(a.dates, a.cache, "cldr-sv-ca-gregorian.json")["dates"]["calendars"]["gregorian"]
    months = gregorian["months"]["format"]["wide"]
    if sorted(months, key=int) != [str(i) for i in range(1, 13)]:
        sys.exit(f"CLDR numbers the months {sorted(months, key=int)}, not 1 to 12")
    tsv.write(out / "month.tsv", ["code", "name"], [{"code": m, "name": months[m]} for m in sorted(months, key=int)])
    days = gregorian["days"]["format"]["wide"]
    if sorted(days) != sorted(WEEKDAYS):
        sys.exit(f"CLDR names the weekdays {sorted(days)}, not {sorted(WEEKDAYS)}")
    tsv.write(out / "weekday.tsv", ["code", "name"], [{"code": i + 1, "name": days[d]} for i, d in enumerate(WEEKDAYS)])
    names = load(a.territories, a.cache, "cldr-sv-territories.json")["localeDisplayNames"]["territories"]
    with open(a.misc, encoding="utf-8", newline="") as f:
        codes = [row["alpha2"] for row in csv.DictReader(f, delimiter="\t")]
    missing = [c for c in codes if c not in names]
    if missing:
        sys.exit(f"CLDR has no Swedish name for {missing}")
    tsv.write(out / "territory.tsv", ["alpha2", "name"], [{"alpha2": c, "name": names[c]} for c in codes])


if __name__ == "__main__":
    main()
