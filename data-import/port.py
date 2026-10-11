#!/usr/bin/env python3
"""Rebuild data/misc/port.tsv from the IANA Service Name and Transport Protocol Port Number Registry.

    data-import/port.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

A row needs a service name and a numeric TCP port. A port the registry lists more than
once keeps the first service it describes, or the first of them where it describes none,
so a number selects one row.
"""
import argparse
import csv
import io
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "port.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["number", "service"]
FLOOR = 4000


def rows(text):
    best, described = {}, set()
    for r in csv.DictReader(io.StringIO(text)):
        service, number = (r["Service Name"] or "").strip(), (r["Port Number"] or "").strip()
        transport, description = (r["Transport Protocol"] or "").strip(), (r["Description"] or "").strip()
        if transport != "tcp" or not service or not number.isdigit():
            continue
        if not 1 <= int(number) <= 65535:
            sys.exit(f"{service}: port {number} is outside 1-65535")
        if number not in best or (description and number not in described):
            best[number] = {"number": number, "service": service}
        if description:
            described.add(number)
    return sorted(best.values(), key=lambda r: int(r["number"]))


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = rows(source.fetch(a.source, a.cache, "service-names-port-numbers.csv").decode("utf-8"))
    if len(table) < FLOOR:
        sys.exit(f"only {len(table)} ports named a TCP service; the registry's columns have moved")
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
