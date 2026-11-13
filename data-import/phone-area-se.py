#!/usr/bin/env python3
"""Rebuild data/sv_SE/phone-area.tsv from PTS's numbering plan.

    data-import/phone-area-se.py [--source URL_OR_FILE] [--cache DIR] [--out FILE]

A row of the plan reads `<ndc> (NDC) <max> <min> Geographic number - Area code for <name>`, where max and min count the
national number, so a subscriber number is the NDC's length shorter. The plan is a PDF: its page contents are Flate
streams that place runs of Windows-1252 text at positions, read here without a PDF library.
"""
import argparse
import collections
import re
import sys
import zlib
from pathlib import Path

import source
import tsv

SOURCE = "https://pts.se/globalassets/globala-block/nummertillstand/the-swedish-numbering-plan-for-telephony-according-to-itu---2024-01-08.pdf"
OUT = Path(__file__).resolve().parent.parent / "data" / "sv_SE" / "phone-area.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
AREA_CODES = 264
STREAM = re.compile(rb"(?<!end)stream\r?\n(.*?)endstream", re.S)
PLACED_TEXT = re.compile(rb"([\d.-]+) ([\d.-]+) ?Tm(.*?)ET", re.S)
ARRAY = re.compile(rb"\[(.*?)\]\s*TJ", re.S)
STRING = re.compile(rb"\(((?:\\.|[^\\)])*)\)", re.S)
# A row's name runs up to the next row or the date heading the next page. One row, 586's, has no hyphen before "Area".
ROW = re.compile(r"(\d+) \(NDC\) (\d+) (\d+) Geographic number(?: -)? Area code for (.+?)(?= \d+ \(NDC\)| \d{4}-\d{2}-\d{2} )")


def unescaped(string: bytes) -> str:
    string = re.sub(rb"\\(\d{3})", lambda m: bytes([int(m[1], 8)]), string)
    return re.sub(rb"\\(.)", rb"\1", string, flags=re.S).decode("cp1252", "replace")


def page_lines(content: bytes) -> list[str]:
    """A page's text, top line first, each line's runs left to right."""
    lines = collections.defaultdict(list)
    for placed in PLACED_TEXT.finditer(content):
        text = "".join(unescaped(s) for array in ARRAY.findall(placed[3]) for s in STRING.findall(array))
        lines[round(float(placed[2]))].append((float(placed[1]), text))
    return ["".join(text for _, text in sorted(lines[y])) for y in sorted(lines, reverse=True)]


def plan_text(pdf: bytes) -> str:
    lines = []
    for stream in STREAM.findall(pdf):
        try:
            content = zlib.decompress(stream)
        except zlib.error:
            continue
        if b" Tm" in content:
            lines += page_lines(content)
    return re.sub(r"\s+", " ", " ".join(lines))


def area_rows(text: str) -> list[dict[str, str]]:
    rows = []
    for ndc, longest, shortest, name in ROW.findall(text):
        rows.append({
            "area": "0" + ndc,
            "ndc": ndc,
            "name": name.replace("- ", "-"),
            "subscriber": f"{{.phone-subscriber.digits-{int(shortest) - len(ndc)}-{int(longest) - len(ndc)}}}",
        })
    if len(rows) != AREA_CODES:
        sys.exit(f"the plan yields {len(rows)} geographic area codes, and states {AREA_CODES}")
    return sorted(rows, key=lambda r: r["area"])


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--cache", default=str(CACHE))
    parser.add_argument("--source", default=SOURCE)
    parser.add_argument("--out", default=str(OUT))
    args = parser.parse_args()
    pdf = source.fetch(args.source, args.cache, "pts-numbering-plan.pdf", magic=b"%PDF-")
    tsv.write(args.out, ["area", "ndc", "name", "subscriber"], area_rows(plan_text(pdf)), templates=("subscriber",))


if __name__ == "__main__":
    main()
