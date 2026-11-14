#!/usr/bin/env python3
"""Rebuild data/nn_NO/month.tsv and weekday.tsv from Unicode CLDR's Nynorsk names.

    data-import/cldr-nn.py [--dates URL_OR_FILE] [--cache DIR] [--out-dir DIR]

A month is numbered 1 to 12 and a weekday 1 (Monday) to 7 (Sunday), as ISO 8601 does.
"""
import argparse
import json
import sys
from pathlib import Path

import source
import tsv

DATES = "https://raw.githubusercontent.com/unicode-org/cldr-json/48.0.0/cldr-json/cldr-dates-full/main/nn/ca-gregorian.json"
ROOT = Path(__file__).resolve().parent.parent / "data"
CACHE = Path(__file__).resolve().parent / "cache"
WEEKDAYS = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"]


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--dates", default=DATES)
    p.add_argument("--out-dir", default=str(ROOT / "nn_NO"))
    a = p.parse_args()
    out = Path(a.out_dir)
    raw = json.loads(source.fetch(a.dates, a.cache, "cldr-nn-ca-gregorian.json").decode("utf-8"))
    gregorian = raw["main"]["nn"]["dates"]["calendars"]["gregorian"]
    months = gregorian["months"]["format"]["wide"]
    if sorted(months, key=int) != [str(i) for i in range(1, 13)]:
        sys.exit(f"CLDR numbers the months {sorted(months, key=int)}, not 1 to 12")
    tsv.write(out / "month.tsv", ["code", "name"], [{"code": m, "name": months[m]} for m in sorted(months, key=int)])
    days = gregorian["days"]["format"]["wide"]
    if sorted(days) != sorted(WEEKDAYS):
        sys.exit(f"CLDR names the weekdays {sorted(days)}, not {sorted(WEEKDAYS)}")
    tsv.write(out / "weekday.tsv", ["code", "name"], [{"code": i + 1, "name": days[d]} for i, d in enumerate(WEEKDAYS)])


if __name__ == "__main__":
    main()
