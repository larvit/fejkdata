#!/usr/bin/env python3
"""Rebuild data/misc/currency.tsv from datasets/currency-codes (PDDL) and CLDR's symbols (Unicode).

    data-import/currency.py [--source URL_OR_FILE] [--symbols URL_OR_FILE ...] [--cache DIR] [--out FILE]

The symbols come from the first locale file that has one, narrow symbols before wide.
"""
import argparse
import csv
import io
import xml.etree.ElementTree as ET
from pathlib import Path

import source
import tsv

SOURCE = "https://raw.githubusercontent.com/datasets/currency-codes/main/data/codes-all.csv"
SYMBOLS = [
    "https://raw.githubusercontent.com/unicode-org/cldr/main/common/main/en.xml",
    "https://raw.githubusercontent.com/unicode-org/cldr/main/common/main/root.xml",
]
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "currency.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["code", "decimals", "name", "numeric", "symbol"]


def symbols(xml_texts):
    """CLDR's symbol per code: the first locale's narrow symbol, else the first locale's wide one."""
    narrow, wide = {}, {}
    for text in xml_texts:
        for c in ET.fromstring(text).iter("currency"):
            for s in c.findall("symbol"):
                into = narrow if s.get("alt") == "narrow" else wide if s.get("alt") is None else None
                if into is not None:
                    into.setdefault(c.get("type"), s.text)
    return {**wide, **narrow}


def rows(text, symbol):
    seen = set()
    for r in csv.DictReader(io.StringIO(text)):
        code = r["AlphabeticCode"]
        if not code or code in seen or r["WithdrawalDate"] or r["Entity"].startswith("ZZ") or not r["MinorUnit"].isdigit():
            continue
        seen.add(code)
        yield {
            "code": code,
            "decimals": r["MinorUnit"],
            "name": r["Currency"],
            "numeric": r["NumericCode"].zfill(3),
            "symbol": symbol.get(code, code),
        }


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--symbols", nargs="+", default=SYMBOLS)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    symbol = symbols(source.fetch(s, a.cache, Path(s).name).decode("utf-8") for s in a.symbols)
    table = rows(source.fetch(a.source, a.cache, "codes-all.csv").decode("utf-8"), symbol)
    tsv.write(a.out, COLUMNS, sorted(table, key=lambda r: r["code"]))


if __name__ == "__main__":
    main()
