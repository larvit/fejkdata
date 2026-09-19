#!/usr/bin/env python3
"""Rebuild data/misc/port.tsv from the IANA Service Name and Transport Protocol Port Number Registry.

    data-import/port.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

What ships is the TCP assignments in use: a row needs a service name, a port number and
a description the registry has filled in. A port the registry lists more than once keeps
the first service, so a number selects one row.
"""
import argparse
import csv
import io
from pathlib import Path

import source
import tsv

SOURCE = "https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "port.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["name", "number"]
UNUSED = ("Reserved", "Unassigned", "IANA assigned this well-formed service name to replace an unregistered or squatted port.")


def rows(text):
    seen = set()
    for r in csv.DictReader(io.StringIO(text)):
        name, number = r["Service Name"].strip(), r["Port Number"].strip()
        described = (r["Description"] or "").strip()
        if r["Transport Protocol"].strip() != "tcp" or not name or not number.isdigit():
            continue
        if not described or described in UNUSED or number in seen:
            continue
        seen.add(number)
        yield {"name": name, "number": number}


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = sorted(rows(source.fetch(a.source, a.cache, "service-names-port-numbers.csv").decode("utf-8")), key=lambda r: int(r["number"]))
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
