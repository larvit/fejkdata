#!/usr/bin/env python3
"""Rebuild data/misc/useragent.tsv from microlinkhq/top-user-agents (MIT).

    data-import/useragent.py [--source URL] [--cache DIR] [--out FILE]

The desktop and mobile lists partition the whole, so `device` is read rather than
parsed. A row ships when the string names a browser and an operating system both.
"""
import argparse
import json
import re
from pathlib import Path

import source
import tsv

SOURCE = "https://raw.githubusercontent.com/microlinkhq/top-user-agents/master/src"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "useragent.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["browser", "device", "os", "ua"]
# First match wins: Edge, Opera and Samsung Internet all carry Chrome's token too,
# and an iPhone says "like Mac OS X".
BROWSERS = [("Edge", r"Edg(iOS)?/"), ("Opera", r"OPR/"), ("Samsung Internet", r"SamsungBrowser/"),
            ("Chrome", r"(Chrome|CriOS)/"), ("Firefox", r"(Firefox|FxiOS)/"), ("Safari", r"Version/[\d.]+ .*Safari")]
SYSTEMS = [("iOS", r"iPhone|iPad|CPU OS "), ("Android", r"Android"), ("ChromeOS", r"CrOS"),
           ("Windows", r"Windows NT"), ("macOS", r"Macintosh|Mac OS X"), ("Linux", r"X11.*Linux|Ubuntu")]


def named(table, ua):
    for name, pattern in table:
        if re.search(pattern, ua):
            return name
    return ""


def rows(desktop, mobile):
    for device, uas in (("desktop", desktop), ("mobile", mobile)):
        for ua in uas:
            browser, os = named(BROWSERS, ua), named(SYSTEMS, ua)
            if browser and os:
                yield {"browser": browser, "device": device, "os": os, "ua": ua}


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
