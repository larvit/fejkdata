"""Rebuild the month and weekday tables of data/sv_SE and data/nn_NO, and data/sv_SE/territory.tsv, from Unicode CLDR's names.

    data-import/cldr.py [--cldr URL_OR_DIR] [--cache DIR] [--data-dir DIR]

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
ROOT = Path(__file__).resolve().parent.parent / "data"
CACHE = Path(__file__).resolve().parent / "cache"
WEEKDAYS = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"]
LOCALES = {"sv": "sv_SE", "nn": "nn_NO"}


def load(cldr: str, cache: str, package: str, lang: str, file: str) -> Any:
    url = f"{cldr}/{package}/main/{lang}/{file}.json"
    return json.loads(source.fetch(url, cache, f"cldr-{lang}-{file}.json").decode("utf-8"))["main"][lang]


def calendar(lang: str, cldr: str, cache: str, out: Path) -> None:
    gregorian = load(cldr, cache, "cldr-dates-full", lang, "ca-gregorian")["dates"]["calendars"]["gregorian"]
    months = gregorian["months"]["format"]["wide"]
    if sorted(months, key=int) != [str(i) for i in range(1, 13)]:
        sys.exit(f"CLDR numbers the months {sorted(months, key=int)}, not 1 to 12")
    tsv.write(out / "month.tsv", ["code", "name"], [{"code": m, "name": months[m]} for m in sorted(months, key=int)])
    days = gregorian["days"]["format"]["wide"]
    if sorted(days) != sorted(WEEKDAYS):
        sys.exit(f"CLDR names the weekdays {sorted(days)}, not {sorted(WEEKDAYS)}")
    tsv.write(out / "weekday.tsv", ["code", "name"], [{"code": i + 1, "name": days[d]} for i, d in enumerate(WEEKDAYS)])


def territories(cldr: str, cache: str, misc: Path, out: Path) -> None:
    names = load(cldr, cache, "cldr-localenames-full", "sv", "territories")["localeDisplayNames"]["territories"]
    with open(misc, encoding="utf-8", newline="") as f:
        codes = [row["alpha2"] for row in csv.DictReader(f, delimiter="\t")]
    missing = [c for c in codes if c not in names]
    if missing:
        sys.exit(f"CLDR has no Swedish name for {missing}")
    tsv.write(out / "territory.tsv", ["alpha2", "name"], [{"alpha2": c, "name": names[c]} for c in codes])


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--cldr", default=CLDR)
    p.add_argument("--data-dir", default=str(ROOT))
    a = p.parse_args()
    data = Path(a.data_dir)
    for lang, locale in LOCALES.items():
        calendar(lang, a.cldr, a.cache, data / locale)
    territories(a.cldr, a.cache, data / "misc" / "territory.tsv", data / "sv_SE")


if __name__ == "__main__":
    main()
