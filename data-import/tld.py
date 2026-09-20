#!/usr/bin/env python3
"""Rebuild data/misc/tld.tsv from the IANA Root Zone Database.

    data-import/tld.py [--source URL_OR_FILE] [--zone URL_OR_FILE] [--cache DIR] [--out FILE]

A TLD ships when the database names a manager for it; one it marks "Not assigned" is a
delegation the root zone no longer holds, and the zone file is fetched to prove the two
agree. The key is the A-label the root zone holds; `unicode` is the form the database
displays, less the bidi marks it wraps a right-to-left label in, and proved to punycode
back to the key.
"""
import argparse
import html
import re
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://www.iana.org/domains/root/db"
ZONE = "https://data.iana.org/TLD/tlds-alpha-by-domain.txt"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "tld.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["tld", "type", "unicode"]
TYPES = ("country-code", "generic", "generic-restricted", "infrastructure", "sponsored")
BIDI = {0x200E: None, 0x200F: None}
ROW = re.compile(r'<a href="/domains/root/db/([^"]+)\.html">([^<]+)</a>.*?</td>\s*<td>([^<]*)</td>\s*<td>(.*?)</td>', re.S)


def rows(page):
    for a_label, shown, kind, manager in ROW.findall(page):
        if "Not assigned" in manager:
            continue
        kind = kind.strip()
        if kind not in TYPES:
            sys.exit(f"{a_label}: the database types it {kind!r}, which is none of {TYPES}")
        yield {"tld": "." + a_label, "type": kind, "unicode": html.unescape(shown).strip().translate(BIDI)}


def prove_each_unicode_form_is_its_key(table):
    """The displayed form is the key's own spelling, which punycode encodes back to."""
    for r in table:
        label = r["unicode"][1:]
        encoded = label if label.isascii() else "xn--" + label.encode("punycode").decode("ascii")
        if "." + encoded != r["tld"]:
            sys.exit(f"{r['tld']}: the database displays {r['unicode']!r}, which is no spelling of it")


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--zone", default=ZONE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = sorted(rows(source.fetch(a.source, a.cache, "root-db.html").decode("utf-8")), key=lambda r: r["tld"])
    prove_each_unicode_form_is_its_key(table)
    listing = source.fetch(a.zone, a.cache, "tlds-alpha-by-domain.txt").decode("utf-8")
    delegated = {"." + line.strip().lower() for line in listing.splitlines() if line.strip() and not line.startswith("#")}
    shipped = {r["tld"] for r in table}
    if len(table) != len(delegated) or shipped != delegated:
        sys.exit(f"the database assigns {len(table)} rows where the root zone holds {len(delegated)} TLDs: {sorted(shipped ^ delegated)[:10]}")
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
