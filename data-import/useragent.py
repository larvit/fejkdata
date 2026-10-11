#!/usr/bin/env python3
"""Rebuild data/misc/useragent.tsv from microlinkhq/top-user-agents (MIT).

    data-import/useragent.py [--source URL] [--cache DIR] [--out FILE]

The desktop and mobile lists partition the whole, so `device` is read rather than
parsed. A row ships when the string names a browser and an operating system both.
"""
import argparse
import json
import re
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://raw.githubusercontent.com/microlinkhq/top-user-agents/master/src"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "useragent.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["browser", "device", "os", "ua"]
# First match wins: a Chromium fork this list does not name ships as Chrome.
BROWSERS = [("Edge", r"Edg(A|iOS)?/"), ("Opera", r"OPR/"), ("Samsung Internet", r"SamsungBrowser/"),
            ("Chrome", r"(Chrome|CriOS)/"), ("Firefox", r"(Firefox|FxiOS)/"), ("Safari", r"Version/[\d.]+ .*Safari")]
SYSTEMS = [("iOS", r"iPhone|iPad|CPU OS "), ("Android", r"Android"), ("ChromeOS", r"CrOS"),
           ("Windows", r"Windows NT"), ("macOS", r"Macintosh|Mac OS X"), ("Linux", r"X11.*Linux|Ubuntu")]


def named(table, ua):
    for name, pattern in table:
        if re.search(pattern, ua):
            return name
    return ""


def rows(desktop, mobile):
    seen, kept = set(), []
    for device, uas in (("desktop", desktop), ("mobile", mobile)):
        for ua in uas:
            browser, os = named(BROWSERS, ua), named(SYSTEMS, ua)
            if not browser or not os:
                print(f"dropped, naming no {'browser' if not browser else 'operating system'}: {ua}", file=sys.stderr)
                continue
            if ua in seen:
                sys.exit(f"{ua}: listed twice, which would draw it twice as often")
            seen.add(ua)
            kept.append({"browser": browser, "device": device, "os": os, "ua": ua})
    if len(kept) < 0.8 * (len(desktop) + len(mobile)):
        sys.exit(f"only {len(kept)} of {len(desktop) + len(mobile)} strings named both; the tokens have moved")
    return kept


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    lists = [json.loads(source.fetch(f"{a.source}/{n}.json", a.cache, f"top-user-agents-{n}.json")) for n in ("desktop", "mobile")]
    tsv.write(a.out, COLUMNS, sorted(rows(*lists), key=lambda r: (r["browser"], r["os"], r["ua"])))


if __name__ == "__main__":
    main()
