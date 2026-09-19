#!/usr/bin/env python3
"""Rebuild data/misc/httpmethod.tsv from the IANA HTTP Method Registry.

    data-import/httpmethod.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

A method ships when the register cites the HTTP core specification for it, RFC 9110
section 9.3 or RFC 5789, which is the nine a request carries; the WebDAV and DeltaV
extensions the register also holds do not. The register's `yes` and `no` ship as `true`
and `false`, which a consumer's boolean reads.
"""
import argparse
import csv
import io
import re
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://www.iana.org/assignments/http-methods/methods.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "httpmethod.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["idempotent", "method", "safe"]
CORE = re.compile(r"\[RFC9110, Section 9\.3\.|\[RFC5789, Section 2\]")
BOOLEAN = {"yes": "true", "no": "false"}
EXPECTED = 9


def rows(text):
    for r in csv.DictReader(io.StringIO(text)):
        method = (r["Method Name"] or "").strip()
        if not CORE.search(r["Reference"] or ""):
            continue
        answers = {}
        for column in ("Safe", "Idempotent"):
            answer = (r[column] or "").strip()
            if answer not in BOOLEAN:
                sys.exit(f"{method}: {column} {answer!r} is neither yes nor no")
            answers[column] = BOOLEAN[answer]
        yield {"idempotent": answers["Idempotent"], "method": method, "safe": answers["Safe"]}


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = sorted(rows(source.fetch(a.source, a.cache, "http-methods.csv").decode("utf-8")), key=lambda r: r["method"])
    if len(table) != EXPECTED:
        sys.exit(f"{len(table)} methods cite the core specification, not {EXPECTED}; the register's Reference column has moved")
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
