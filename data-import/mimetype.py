#!/usr/bin/env python3
"""Rebuild data/misc/mimetype.tsv from mime-db (MIT), the IANA media type registry with extensions.

    data-import/mimetype.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

Only IANA-registered types with a filename extension ship; the first extension is the one listed.
"""
import argparse
import json
from pathlib import Path

import source
import tsv

SOURCE = "https://raw.githubusercontent.com/jshttp/mime-db/master/db.json"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "mimetype.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["ext", "type"]


def rows(db):
    for mimetype, entry in db.items():
        if entry.get("source") == "iana" and entry.get("extensions"):
            yield {"ext": "." + entry["extensions"][0], "type": mimetype}


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    table = rows(json.loads(source.fetch(a.source, a.cache, "mime-db.json").decode("utf-8")))
    tsv.write(a.out, COLUMNS, sorted(table, key=lambda r: r["type"]))


if __name__ == "__main__":
    main()
