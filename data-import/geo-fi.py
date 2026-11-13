#!/usr/bin/env python3
"""Rebuild data/geo/FI/*.tsv and address.json from Posti's postal code files, Statistics Finland and GeoNames.

    data-import/geo-fi.py [--cache DIR] [--min-streets N] [--street-share F] [--out DIR]

Posti's postcode file (PCF) names each postinumero's postitoimipaikka and type, and its basic address file (BAF)
every street under its postinumero, in Finnish and, where Posti has one, in Swedish, with its odd and its even
house number range. A postitoimipaikka is a locality. Its postinumerot are the normal ones under which BAF lists a
street, each weighted by its residents in Statistics Finland's Paavo; one without residents is dropped. A
locality's population is its postinumerot's residents summed, and its kunta is the one BAF places most of its
ranges in. A street is weighted by its ranges in the locality. Its Swedish name is the one Posti gives it most
often there, or its Finnish name where Posti gives none.

Statistics Finland's classification names each kunta and places it in its maakunta, and its table 11ra gives
both their populations. Posti's files hold no street on Åland, so its maakunta ships no place. Posti writes a
postitoimipaikka in capitals, and `recased` writes it as a place name. address.json weighs each length of house
number, with a delivery letter or without, by how many ends of BAF's ranges carry one.
"""
import collections
import io
import json
import re
import sys
import zipfile

import geo
import source

POSTI = "https://www.posti.fi/webpcode/"
CLASSIFICATION = "https://data.stat.fi/api/classifications/v2/correspondenceTables/kunta_1_20260101%23maakunta_1_20260101/maps?content=data&format=json&lang=fi&meta=max"
PAAVO = "https://pxdata.stat.fi/PxWeb/api/v1/fi/Postinumeroalueittainen_avoin_tieto/2026/12ey_ja2026.px"
PAAVO_QUERY = {"postinumeroalue_4_20260101": ["*"], "contentscode": ["he_vakiy"], "timeperiod_y": ["2024"]}
POPULATION = "https://pxdata.stat.fi/PxWeb/api/v1/fi/StatFin/vaerak/11ra.px"
POPULATION_QUERY = {"alue_23_20260101": ["KU*", "MK*"], "contentscode": ["vaerak-vaesto"], "timeperiod_y": ["2025"]}
TIMEZONE = "Europe/Helsinki"
NORMAL = "1"
ALAND = "21"
# Posti's abbreviations of asema and kirkonkylä, as in ALAVUS AS and HEINOLA KK.
LOWER = {"as", "kk"}
VOWELS = "aeiouyäö"
NOT_A_STREET = re.compile(r"[\d,()/]")


def posti(cache, kind):
    """The lines of Posti's latest file of a kind, PCF or BAF, as text."""
    listing = source.download(POSTI).decode("utf-8")
    name = re.search(rf'href="{re.escape(POSTI)}({kind}_\d{{8}}\.zip)"', listing)[1]
    print(f"Posti {name}", file=sys.stderr)
    z = zipfile.ZipFile(io.BytesIO(source.fetch(POSTI + name, cache, "posti-" + name, magic=b"PK")))
    return z.read(name.replace(".zip", ".dat")).decode("latin-1").splitlines()


def field(line, start, length):
    """A field of a Posti record, by the 1-based position and the length Posti's record description gives."""
    return line[start - 1 : start - 1 + length].strip()


def recased(name, kunnat):
    """A postitoimipaikka written as a place name, each word and hyphenated part capitalised, with three exceptions.
    An abbreviation stays lower case: HEINOLA KK is Heinola kk. A word after a genitive does too: INHAN TEHTAAT is
    Inhan tehtaat. So does a part opening with the vowel the part before it ends in, KUOPPA-AHO is Kuoppa-aho, unless
    the part is a kunta: YLI-II is Yli-Ii. kunnat holds the kunnat's names in capitals."""
    words = []
    for word in name.lower().split(" "):
        parts = []
        for part in word.split("-"):
            joined = parts and parts[-1][-1] == part[0] and part[0] in VOWELS and part.upper() not in kunnat
            parts.append(part if joined else part[:1].upper() + part[1:])
        after_genitive = words and words[-1].endswith("n")
        words.append(word if word in LOWER or after_genitive else "-".join(parts))
    return " ".join(words)


def stat_fi(url, query, cache, name):
    """A PxWeb table's values by the code of its first dimension, from a query mapping each dimension to its
    values, where a value ending in * selects every code it opens."""
    body = {"query": [{"code": d, "selection": {"filter": "all" if v[0].endswith("*") else "item", "values": v}} for d, v in query.items()], "response": {"format": "json-stat2"}}
    table = json.loads(source.fetch(url, cache, name, data=json.dumps(body).encode(), headers={"Content-Type": "application/json"}))
    index = table["dimension"][table["id"][0]]["category"]["index"]
    return {code: table["value"][i] for code, i in index.items() if table["value"][i] is not None}


def divisions(cache):
    """Each kunta's name and maakunta, and each maakunta's name, by code."""
    municipalities, regions = {}, {}
    for m in json.loads(source.fetch(CLASSIFICATION, cache, "statfi-kunta-maakunta.json")):
        s, r = m["sourceItem"], m["targetItem"]
        municipalities[s["code"]] = {"name": s["classificationItemNames"][0]["name"], "region": r["code"]}
        regions[r["code"]] = r["classificationItemNames"][0]["name"]
    return municipalities, regions


def normal_codes(pcf, municipality_of):
    """Each normal postinumero among PCF's lines, mapped to its postitoimipaikka written as a place name."""
    towns = {field(l, 14, 5): field(l, 19, 30) for l in pcf if field(l, 111, 1) == NORMAL}
    kunnat = {m["name"].upper() for m in municipality_of.values()}
    return {code: recased(town, kunnat) for code, town in towns.items()}


def mainland(region_names, municipality_of, people):
    """The region and the municipality rows of every maakunta but Åland, each with its population in people."""
    regions = {c: {"code": c, "name": n, "population": people["MK" + c], "timezone": TIMEZONE} for c, n in region_names.items() if c != ALAND}
    municipalities = {c: {"code": c, "name": m["name"], "region": m["region"], "population": people["KU" + c]} for c, m in municipality_of.items() if m["region"] != ALAND}
    return regions, municipalities


def street_name(name):
    """Whether a BAF name is a street's, not a highway's such as `Valtatie 23`, a postinumero, or an island's or a
    lake's such as `Mustalampi, Lammassaari` or `Makkarasaari/Lapinsaari`."""
    return bool(name) and not NOT_A_STREET.search(name)


def capitalised(name):
    """A BAF name with its first letter in upper case and the rest as Posti writes it."""
    return name[:1].upper() + name[1:]


def swedish_name(names, finnish):
    """A street's Swedish name: the one names counts most often, the first in the alphabet winning a tie, or its
    Finnish name where names counts none."""
    return min(names, key=lambda name: (-names[name], name), default=finnish)


def house_numbers(line):
    """The house numbers at the ends of a BAF range, each as (number, letter)."""
    ends = {(field(line, at, 5), field(line, at + 5, 1)) for at in (188, 201) if field(line, at, 5)}
    return [(int(n), letter) for n, letter in ends]


def tally(baf, codes):
    """From BAF's ranges under a normal postinumero: per locality, the ranges of each street and kunta, each
    street's Swedish names and the postinumerot; and over them all, the house numbers at the ranges' ends."""
    streets, swedish, in_municipality = collections.Counter(), collections.defaultdict(collections.Counter), collections.Counter()
    postal_codes, numbers = collections.defaultdict(set), collections.Counter()
    for line in baf:
        code, street, swedish_street = field(line, 14, 5), capitalised(field(line, 103, 30)), capitalised(field(line, 133, 30))
        if code not in codes or not street_name(street):
            continue
        place = codes[code]
        streets[(place, street)] += 1
        if swedish_street:
            swedish[(place, street)][swedish_street] += 1
        in_municipality[(place, field(line, 214, 3))] += 1
        postal_codes[place].add(code)
        for n, letter in house_numbers(line):
            if 0 < n < 10000:
                numbers[geo.number_range(n, bool(letter))] += 1
    return streets, swedish, in_municipality, postal_codes, numbers


def localities(codes_of, residents, home, points):
    """Each locality with a postinumero holding residents and a point, and those postinumerot with their residents."""
    out, postal_codes = {}, {}
    for place, codes in codes_of.items():
        kept = {c: residents[c] for c in sorted(codes) if c in residents}
        centre = geo.centroid([points[c] for c in kept if c in points])
        if not kept or not centre:
            print(f"{place}: no postinumero with residents and a point, dropped", file=sys.stderr)
            continue
        postal_codes[place] = kept
        out[place] = {"name": place, "municipality": home[place], "population": sum(kept.values()), "lat": f"{centre[0]:.4f}", "lon": f"{centre[1]:.4f}"}
    return out, postal_codes


def main():
    a = geo.parser(__doc__, "FI", 1.0).parse_args()
    cache, out = geo.directories(a)
    municipality_of, region_names = divisions(cache)
    people = stat_fi(POPULATION, POPULATION_QUERY, cache, "statfi-11ra.json")
    residents = {code: n for code, n in stat_fi(PAAVO, PAAVO_QUERY, cache, "statfi-paavo.json").items() if n}
    streets, swedish, in_municipality, codes_of, numbers = tally(posti(cache, "BAF"), normal_codes(posti(cache, "PCF"), municipality_of))
    points = {r["code"]: r for r in geo.geonames(cache, "FI") if r["lat"] is not None}
    places, postal_codes = localities(codes_of, residents, geo.home_municipalities(in_municipality), points)
    named = geo.top_streets(streets, a.street_share, a.min_streets, "addresses")
    for place, rows in named.items():
        for row in rows:
            row["swedish-name"] = swedish_name(swedish[(place, row["name"])], row["name"])
    regions, municipalities = mainland(region_names, municipality_of, people)
    kept, dropped = geo.fence(regions, municipalities, places, named, postal_codes)
    geo.log_dropped_localities(dropped["locality"])
    geo.exit_on_dropped_parents(dropped)
    rows = [{"code": c, "locality": l, "population": n} for l in kept["locality"] for c, n in postal_codes[l].items()]
    geo.write_tables(out, kept, rows, named, postal_code=["code", "locality", "population"], street=["name", "locality", "swedish-name", "addresses"])
    (out / "address.json").write_text(geo.address_json(numbers, "{lower(1)}"), encoding="utf-8")


if __name__ == "__main__":
    main()
