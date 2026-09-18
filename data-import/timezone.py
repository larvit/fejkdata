#!/usr/bin/env python3
"""Rebuild data/misc/timezone.tsv from the IANA tzdb tarball (public domain).

    data-import/timezone.py [--source URL_OR_FILE] [--cache DIR] [--out FILE] [--territories FILE]

zone.tab names one zone per territory, so Europe/Stockholm ships where zone1970.tab
would spell Sweden Europe/Berlin. The offset is the zone's standard offset, the first
field of its Zone rule's last continuation line, with a Link resolved to its target.
"""
import argparse
import csv
import io
import re
import sys
import tarfile
from pathlib import Path

import source
import tsv

SOURCE = "https://data.iana.org/time-zones/tzdata-latest.tar.gz"
OUT = Path(__file__).resolve().parent.parent / "data" / "misc" / "timezone.tsv"
TERRITORIES = Path(__file__).resolve().parent.parent / "data" / "misc" / "territory.tsv"
CACHE = Path(__file__).resolve().parent / "cache"
COLUMNS = ["offset", "territory", "zone"]
REGIONS = ["africa", "antarctica", "asia", "australasia", "backward", "etcetera", "europe", "northamerica", "southamerica"]


def offsets(tar):
    """Every zone's standard offset, and every link's target."""
    std, links = {}, {}
    for name in REGIONS:
        zone = None
        for raw in tar.extractfile(name).read().decode("utf-8").splitlines():
            line = raw.split("#")[0].rstrip()
            if not line.strip():
                continue
            if line.startswith("Zone"):
                f = line.split()
                zone, std[f[1]] = f[1], f[2]
            elif line.startswith("Link"):
                f = line.split()
                links[f[2]], zone = f[1], None
            elif line[0] in " \t" and zone:
                std[zone] = line.split()[0]
            else:
                zone = None
    return std, links


def resolve(zone, std, links):
    for _ in range(10):
        if zone in std:
            return std[zone]
        if zone not in links:
            return None
        zone = links[zone]
    return None


def utc_offset(raw):
    """±HH:MM from a tzdb STDOFF field; a zone still off by seconds is not one we can spell."""
    m = re.match(r"^(-)?(\d{1,2}):(\d{2})(?::(\d{2}))?$", raw)
    if not m or (m.group(4) or "00") != "00":
        return None
    return f"{'-' if m.group(1) else '+'}{int(m.group(2)):02d}:{m.group(3)}"


def rows(tar, territories):
    std, links = offsets(tar)
    for line in tar.extractfile("zone.tab").read().decode("utf-8").splitlines():
        if line.startswith("#") or not line.strip():
            continue
        territory, _, zone = line.split("\t")[:3]
        if territory not in territories:
            continue
        raw = resolve(zone, std, links)
        if raw is None:
            sys.exit(f"{zone}: the tarball gives it no Zone rule and no Link to one")
        offset = utc_offset(raw)
        if offset is None:
            sys.exit(f"{zone}: standard offset {raw!r} is not a whole number of minutes")
        yield {"offset": offset, "territory": territory, "zone": zone}


def main():
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--source", default=SOURCE)
    p.add_argument("--out", default=str(OUT))
    p.add_argument("--territories", default=str(TERRITORIES))
    a = p.parse_args()
    shipped = {r["alpha2"] for r in csv.DictReader(io.StringIO(Path(a.territories).read_text(encoding="utf-8")), delimiter="\t")}
    body = source.fetch(a.source, a.cache, "tzdata-latest.tar.gz", magic=b"\x1f\x8b")
    with tarfile.open(fileobj=io.BytesIO(body)) as tar:
        table = sorted(rows(tar, shipped), key=lambda r: r["zone"])
        print(f"tzdb {tar.extractfile('version').read().decode('utf-8').strip()}", file=sys.stderr)
    linked = {r["territory"] for r in table}
    if missing := shipped - linked:
        sys.exit(f"no zone for {sorted(missing)}; a parent row without a child is a load error")
    tsv.write(a.out, COLUMNS, table)


if __name__ == "__main__":
    main()
