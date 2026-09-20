#!/usr/bin/env python3
"""Rebuild data/misc/loglevel.tsv from RFC 5424 and POSIX <syslog.h>.

    data-import/loglevel.py [--source URL_OR_FILE] [--posix URL_OR_FILE] [--cache DIR] [--out FILE]

Table 2 is the whole set, read between the facility table's caption and its own. The
keyword is POSIX's severity macro less its `LOG_` prefix, lowercased, taken in the
order POSIX lists them, which `LOG_UPTO` states is the severity order; each must be a
prefix of the severity RFC 5424 numbers alike, so the two sources pin one pairing.
"""
import argparse
import html
import re
import sys
from pathlib import Path

import source
import tsv

SOURCE = "https://www.rfc-editor.org/rfc/rfc5424.txt"
POSIX = "https://pubs.opengroup.org/onlinepubs/9799919799/basedefs/syslog.h.html"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "loglevel.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["code", "keyword", "severity"]
TABLE = re.compile(r"Table 1\.\s+Syslog Message Facilities(.*?)Table 2\.\s+Syslog Message Severities", re.S)
ROW = re.compile(r"^ +(\d) +([A-Z][a-z]+): ", re.M)
SEVERITIES = re.compile(r"severity level portion of the.*?macros:(.*?)The following shall be declared as functions", re.S)
MACRO = re.compile(r"\bLOG_([A-Z]+)\b")
EXPECTED = 8


def severities(text):
    table = TABLE.search(text)
    if not table:
        sys.exit("the RFC holds no text between the facility and the severity caption; its tables have moved")
    return ROW.findall(table.group(1))


def keywords(page):
    section = SEVERITIES.search(html.unescape(re.sub(r"<[^>]+>", "", page)))
    if not section:
        sys.exit("POSIX spells no severity list between its own caption and the function declarations; the page has moved")
    return [m.lower() for m in MACRO.findall(section.group(1))]


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--posix", default=POSIX)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = sorted(severities(source.fetch(a.source, a.cache, "rfc5424.txt").decode("utf-8")))
    codes = [code for code, _ in table]
    if codes != [str(i) for i in range(EXPECTED)]:
        sys.exit(f"the severity table numbers {codes}, not 0 through {EXPECTED - 1}")
    words = keywords(source.fetch(a.posix, a.cache, "posix-syslog.h.html").decode("utf-8"))
    if len(words) != EXPECTED:
        sys.exit(f"POSIX lists {len(words)} severity macros, not {EXPECTED}: {words}")
    rows = []
    for (code, severity), keyword in zip(table, words):
        if not severity.lower().startswith(keyword):
            sys.exit(f"severity {code} is {severity} in the RFC and {keyword} in POSIX, which is no prefix of it")
        rows.append({"code": code, "keyword": keyword, "severity": severity})
    tsv.write(a.out, COLUMNS, rows)


if __name__ == "__main__":
    main()
