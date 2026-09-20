#!/usr/bin/env python3
"""Rebuild data/misc/loglevel.tsv from RFC 5424's severity table.

    data-import/loglevel.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

Table 2 is the whole set, read between the facility table's caption and its own.
"""
import argparse
import re
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://www.rfc-editor.org/rfc/rfc5424.txt"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "loglevel.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["code", "severity"]
TABLE = re.compile(r"Table 1\.\s+Syslog Message Facilities(.*?)Table 2\.\s+Syslog Message Severities", re.S)
ROW = re.compile(r"^ +(\d) +([A-Z][a-z]+): ", re.M)
EXPECTED = 8


def rows(text):
    table = TABLE.search(text)
    if not table:
        sys.exit("the RFC holds no text between the facility and the severity caption; its tables have moved")
    for code, severity in ROW.findall(table.group(1)):
        yield {"code": code, "severity": severity}


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = sorted(rows(source.fetch(a.source, a.cache, "rfc5424.txt").decode("utf-8")), key=lambda r: r["code"])
    codes = [r["code"] for r in table]
    if codes != [str(i) for i in range(EXPECTED)]:
        sys.exit(f"the severity table numbers {codes}, not 0 through {EXPECTED - 1}")
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
