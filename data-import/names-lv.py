"""Rebuild data/lv_LV/first-name.tsv and last-name.tsv: the first names from PMLP's register statistics, the last names from a language model's recall.

    data-import/names-lv.py [--vardi URL_OR_FILE] [--first N] [--cache DIR] [--out DIR]

- The first names are PMLP's "Personu vārdi" with the sex of the bearers (CC0 1.0, "Avots: Pilsonības un migrācijas lietu pārvalde",
  data.gov.lv), the file `Vardi-dz`: each person counts once, under the first of their given names. A person of no stated sex, and a
  label that is not a name in letters, as `tsv.NAME` says, are dropped. The register holds the foreigners' names too.
- PMLP publishes no last names, so each is a language model's recall of a name common in Latvia.
  The list holds the commoner names from the most to the least common, then a line `=`, then the rarer ones, sorted. A commoner
  name has the count 100000 over the square root of its rank, and the rarer ones share the count of the last rank.
  These counts are estimates, not published figures.
- A last name has a row under each sex: the masculine form under `m` and its feminine form under `f`, by the endings of `feminine`.
"""
import argparse
import csv
import io
from collections import Counter
from pathlib import Path

import recalled_names
import source
import tsv

VARDI = "https://data.gov.lv/dati/dataset/ac246d11-d5d6-445e-a6c7-8f5013460335/resource/10f0bb84-dd18-44e9-9a80-dc1515a4b3a9/download/vardi-dz-20260701.csv"
OUT = Path(__file__).resolve().parent.parent / "data" / "lv_LV"
CACHE = Path(__file__).resolve().parent / "cache"
SEXES = {"SIEVIETE": "f", "VĪRIETIS": "m"}
TOP = 100000

LAST_RECALLED = """
Bērziņš Kalniņš Ozoliņš Jansons Ozols Liepiņš Krūmiņš Balodis Eglītis Zariņš
Pētersons Kļaviņš Vītols Vanags Siliņš Krastiņš Lācis Sproģis Jēkabsons Andersons
Ivanovs Petrovs Smirnovs Kuzņecovs Vasiļjevs Sokolovs Mihailovs Fjodorovs Pavlovs Aleksejevs
Lapiņš Avotiņš Lazdiņš Ābols Āboliņš Apinis Bondars Briedis Dzenis Freimanis
Gailis Grīnbergs Jaunzems Kalējs Kārkliņš Krauklis Krauze Lielmanis Mednis Melnis
Millers Neilands Pīlāgs Priede Puriņš Putniņš Rudzītis Rozītis Strazds Tomsons
Upenieks Vilks Zālītis Zeltiņš Rozenbergs Birznieks Cīrulis Dombrovskis Gulbis Kaktiņš
Zvejnieks Dārznieks Strautiņš Rubenis Auziņš Ansons Kristapsons Kalvītis Ulmanis Valters
Kronbergs Dūmiņš Lūsis Zaķis Ķirsis Sīlis Alksnis Bukovskis Ceplis Podnieks
Volkovs Kozlovs Orlovs Jakovļevs Zaharovs Semjonovs Nikolajevs Romanovs Morozovs Stepanovs
Popovs Novikovs Lebedevs Kovaļovs Andrejevs Jegorovs Aleksandrovs Grigorjevs Filippovs Medvedevs
Gusevs Tarasovs Zaicevs Belovs Tihonovs Vorobjovs Isajevs Kondratjevs Titovs Maksimovs
=
Antonovs Apsītis Arājs Baltiņš Baranovs Borisovs Bērzs Brencis Brūveris Bumbiers Caune
Cinis Dambis Daugavietis Ezeriņš Fedorovs Frolovs Gavrilovs Grīnvalds Ģērmanis Ilziņš
Jefimovs Jurjāns Kalns Kiseļovs Kronvalds Kuprijanovs Lapsa Lasmanis Lejnieks Liepa
Lukjanovs Lukstiņš Makarovs Mālnieks Mežaraups Mikelsons Muižnieks Ošiņš Pumpurs Raups
Roze Rozentāls Sergejevs Skuja Spuris Stūrmanis Šteins Ūdris Vaivods Vērdiņš Vilciņš
Zvirbulis
"""


def firsts(text: str, top: int) -> list[dict[str, object]]:
    """The rows (name, sex, count) of the register's first names, the `top` of each sex by bearers."""
    counts: Counter[tuple[str, str]] = Counter()
    for row in csv.DictReader(io.StringIO(text.lstrip("﻿"))):
        words = row["Vardi"].split()
        if words and row["Dzimums"] in SEXES and tsv.NAME.match(words[0]):
            counts[(words[0].title(), SEXES[row["Dzimums"]])] += int(row["Skaits"])
    rows = [{"name": name, "sex": sex, "count": n} for (name, sex), n in counts.items()]
    kept = [r for sex in SEXES.values() for r in sorted((r for r in rows if r["sex"] == sex), key=lambda r: (-int(str(r["count"])), str(r["name"])))[:top]]
    return sorted(kept, key=lambda r: (str(r["name"]), str(r["sex"])))


def feminine(name: str) -> str:
    """The feminine form of a masculine last name; one that does not end as a masculine form does is the same under both."""
    if name.endswith("skis"):
        return name[:-2] + "a"
    if name.endswith("nieks"):
        return name[:-2] + "ce"
    if name.endswith(("ars", "ems", "ers", "ons")):
        return name[:-1] + "e"
    if name.endswith("is"):
        return name[:-2] + "e"
    if name.endswith(("š", "s")):
        return name[:-1] + "a"
    return name


def lasts() -> list[dict[str, object]]:
    """The rows (name, sex, count) of the last names: each recalled name under m, and its feminine form under f."""
    rows = recalled_names.weighted("last names", LAST_RECALLED, TOP)
    return [{"name": form(name), "sex": sex, "count": n} for name, n in rows for sex, form in (("f", feminine), ("m", str))]


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--vardi", default=VARDI)
    p.add_argument("--first", type=int, default=2000, help="names per sex")
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    text = source.fetch(a.vardi, a.cache, "pmlp-vardi-dz-20260701.csv", magic=b"\xef\xbb\xbfVardi").decode("utf-8")
    out = Path(a.out)
    out.mkdir(parents=True, exist_ok=True)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], firsts(text, a.first))
    tsv.write(out / "last-name.tsv", ["name", "sex", "count"], sorted(lasts(), key=lambda r: (str(r["name"]), str(r["sex"]))))


if __name__ == "__main__":
    main()
