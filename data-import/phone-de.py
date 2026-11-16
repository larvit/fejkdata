"""Rebuild data/de_DE/phone-mobile.tsv and phone-fixed.tsv from the Bundesnetzagentur's numbering lists.

    data-import/phone-de.py [--source-mobile URL_OR_FILE] [--source-fixed URL_OR_FILE] [--cache DIR]
                            [--out-mobile FILE] [--out-fixed FILE]

Mobile: "Zugeteilte Rufnummernblöcke (Mobile Dienste)", RNB_csv.csv, one block of the Nummernplan Mobile Dienste
per row, written without the trunk 0. Nummernplan sections 2.3 and 2.4: a block of five digits leaves a
subscriber number of six digits, one of four digits seven, one of three seven or eight. The list writes the block
15591 with a space where it has a semicolon.

Fixed: the Vorwahlverzeichnis, a ZIP of one CSV of Ortsnetzkennzahlen with their Ortsnetz and an active flag.
Only active codes are rows. Verfügung 25/2006, section 2.3: a number allocated now is ten digits long in the
four two-digit areas and eleven elsewhere, so a subscriber number is that many digits shorter than the code.
"""
import argparse
import csv
import io
import re
import sys
import zipfile
from pathlib import Path

import source
import tsv

MOBILE = "https://www.bundesnetzagentur.de/DE/Fachthemen/Telekommunikation/Nummerierung/MobileDienste/RNB_csv.csv"
FIXED = "https://www.bundesnetzagentur.de/DE/Fachthemen/Telekommunikation/Nummerierung/ONRufnr/_DL/Vorwahlverzeichnis_ONB.zip.zip"
DATA = Path(__file__).resolve().parent.parent / "data" / "de_DE"
CACHE = Path(__file__).resolve().parent / "cache"
BLOCK = re.compile(r"^(\d+)(?:;| )")
SUBSCRIBER = "{.phone-subscriber.digits-%s}"


def mobile_rows(listing: bytes) -> list[dict[str, str]]:
    """A row for each block of the list, with the digits the plan gives its subscriber number."""
    rows = []
    for line in listing.decode("iso-8859-1").splitlines()[1:]:
        match = BLOCK.match(line)
        if not match:
            sys.exit(f"{line!r} is not a block of the list")
        ndc = match[1]
        digits = {3: "7-8", 4: "7", 5: "6"}.get(len(ndc))
        if digits is None:
            sys.exit(f"block {ndc} is not 3 to 5 digits long")
        rows.append({"area": "0" + ndc, "ndc": ndc, "subscriber": SUBSCRIBER % digits})
    return sorted(rows, key=lambda r: r["area"])


def fixed_rows(archive: bytes) -> list[dict[str, str]]:
    """A row for each active Ortsnetzkennzahl, with the Ortsnetz it names."""
    with zipfile.ZipFile(io.BytesIO(archive)) as z:
        (name,) = z.namelist()
        text = z.read(name).decode("utf-8-sig")
    rows = []
    for row in csv.DictReader(io.StringIO(text), delimiter=";"):
        if row["KennzeichenAktiv"] != "1":
            continue
        ndc = row["Ortsnetzkennzahl"]
        if not re.fullmatch(r"[1-9]\d{1,4}", ndc):
            sys.exit(f"{ndc!r} is not a code of 2 to 5 digits")
        rows.append({"area": "0" + ndc, "ndc": ndc, "name": row["Ortsnetzname"], "subscriber": SUBSCRIBER % ((10 if len(ndc) == 2 else 11) - len(ndc))})
    return sorted(rows, key=lambda r: r["area"])


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--cache", default=str(CACHE))
    parser.add_argument("--source-mobile", default=MOBILE)
    parser.add_argument("--source-fixed", default=FIXED)
    parser.add_argument("--out-mobile", default=str(DATA / "phone-mobile.tsv"))
    parser.add_argument("--out-fixed", default=str(DATA / "phone-fixed.tsv"))
    args = parser.parse_args()
    listing = source.fetch(args.source_mobile, args.cache, "RNB_csv.csv", magic=b"RNB;")
    archive = source.fetch(args.source_fixed, args.cache, "Vorwahlverzeichnis_ONB.zip", magic=b"PK")
    tsv.write(args.out_mobile, ["area", "ndc", "subscriber"], mobile_rows(listing), templates=["subscriber"])
    tsv.write(args.out_fixed, ["area", "ndc", "name", "subscriber"], fixed_rows(archive), templates=["subscriber"])


if __name__ == "__main__":
    main()
