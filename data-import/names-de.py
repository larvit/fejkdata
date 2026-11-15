"""Rebuild data/de_DE/first-name.tsv and last-name.tsv from the city of Munich's newborn names and a language model's recall.

    data-import/names-de.py [--cache DIR] [--out DIR]

- The first names are the city of Munich's open data "Vornamen von Neugeborenen" (Datenlizenz Deutschland -
  Namensnennung 2.0, "Datenquelle: Landeshauptstadt München, Statistisches Amt"), the children of each sex that
  people living in Munich bore at the end of each year from 2013 to 2025. A name's count is the sum of the numbers
  the files publish for it. The 2016 file ends in a row of a DOS end-of-file mark and nothing else, which is dropped. A file replaces a count of 4 or 5 or fewer with a phrase, so a name that is never
  published as a number has no row. A name that girls and boys both bear has a row under each sex. A label that is
  not a name in letters, as `tsv.NAME` says, is dropped.
- The last names are a language model's recall of names common in Germany, as no open register publishes them.
  The first of them run from the most to the least common, and each has the count 700000 over the square root of
  its rank. The rarer ones that follow, in no order, share the count of the last rank. These counts are
  estimates, no published figure.
"""
import argparse
import csv
import io
import sys
from pathlib import Path

import source
import tsv

BASE = "https://opendata.muenchen.de/dataset/99ad40ec-9d7b-4a2e-87eb-9bac783fb57a/resource/"
FILES = {
    2013: "17c4915d-6b8b-470f-a6b6-31f61b10a285/download/vornamen_2013.csv",
    2014: "5d581ba3-4c7e-4057-8be1-04d56bb0d8d6/download/vornamen_2014.csv",
    2015: "c11e8fef-94c8-4e16-a769-70dc6b98b69b/download/vornamen_2015.csv",
    2016: "6655ec3c-251a-4a5e-9a57-d60c4ad59ca3/download/vornamen2016.csv",
    2017: "2c70d289-8b1a-4071-9739-866c5e532b77/download/vornamen2017.csv",
    2018: "23fdd53e-485a-46bd-abb6-d7a51b2ebb18/download/vornamen_2018.csv",
    2019: "523d5860-25eb-4bea-96e3-193d1dacfb8f/download/vornamen2019.csv",
    2020: "48a4c5fd-f5f2-4c0c-bcb3-222fd9ebda67/download/vornamen_2020.csv",
    2021: "dc6170db-f7f4-4bdc-b790-13df55f0cf64/download/vornamen_2021.csv",
    2022: "5f11d3a0-4779-4f64-b113-b59326e6d839/download/open_data_portal_2022.csv",
    2023: "02ab322e-d33e-4447-a9ab-23df63dfa7e1/download/vornamen-muenchen-2023.csv",
    2024: "9f393a00-aa64-4733-8c8c-80eabc18f95a/download/vornamen_muenchen2024.csv",
    2025: "6388c83a-266d-437c-824a-7bbcb7ceec63/download/vornamen_muenchen2025.csv",
}
OUT = Path(__file__).resolve().parent.parent / "data" / "de_DE"
CACHE = Path(__file__).resolve().parent / "cache"
SEX = {"w": "f", "m": "m"}

LAST_RECALLED = """
Müller Schmidt Schneider Fischer Weber Meyer Wagner Becker Schulz Hoffmann
Schäfer Koch Bauer Richter Klein Wolf Schröder Neumann Schwarz Zimmermann
Braun Krüger Hofmann Hartmann Lange Schmitt Werner Schmitz Krause Meier
Lehmann Schmid Schulze Maier Köhler Herrmann König Walter Mayer Huber
Kaiser Fuchs Peters Lang Scholz Möller Weiß Jung Hahn Schubert
Vogel Friedrich Keller Günther Frank Berger Winkler Roth Beck Lorenz
Baumann Franke Albrecht Schuster Simon Ludwig Böhm Winter Kraus Martin
Schumacher Krämer Vogt Stein Jäger Otto Sommer Groß Seidel Heinrich
Brandt Haas Schreiber Graf Schulte Dietrich Ziegler Kuhn Pohl Engel
Horn Busch Bergmann Thomas Voigt Sauer Arnold Wolff Pfeiffer Hoffmeister
Ernst Lindner Bayer Hübner Kühn Schütz Kaufmann Ebert Götz Hartung
Kremer Moser Berndt Fiedler Vetter Maurer Wirth Henkel Rieger Seifert
Förster Beyer Hermann Gärtner Reuter Bender Köster Opitz Kunze Harms
Mohr Gebhardt Heinz Behrens Wendt Jansen Kurz Menzel Carstens Ritter
Naumann Lutz Schindler Brinkmann Heck Heil Bach Böttcher Schenk Zander
Nowak Kowalski Lenz Wiesner Pietsch Dörr Held Linke Wegner Burkhardt
Hess Kirchner Brückner Rauch Kasper Marx Geiger Fröhlich Hein Rose
Bock Popp Frey Dahm Kramer Heller Eichhorn Kolb Lau Michel
Blum Siegel Sander Strauß Seitz Ullrich Mertens Reimann Specht Stoll
Thiele Neubauer Schilling Rothe Burger Ahrens Haupt Wenzel Hinz Rudolph
Paul Lemke Hanke Haase Voss Stephan Jacob Bartels Möhring Kühne
Albers Wilke Schiller Janßen Peter Schade Büttner Henning Wolter Schlegel
=
Abel Ackermann Adam Adler Albert Altmann Appel Arndt Bachmann Baier Barth
Bartsch Bauch Baur Beckmann Behr Benz Berg Bergner Bernhardt Bertram
Bickel Bischoff Blank Block Bohn Bonn Borchert Bosch Brand Brauer
Braune Breuer Brenner Brunner Buchholz Bühler Busse Conrad Christ Clemens
Cordes Dallmann Decker Degen Dettmer Diehl Dittrich Dohrmann Dreher Drechsler
Dressler Dunkel Eberhardt Eckert Eder Ehlers Eichler Eisele Engelhardt Erdmann
Esser Falk Faber Feldmann Fink Fleischer Flick Foerster Franz Freitag
Freund Frenzel Fritsch Fritz Funk Gabriel Gast Gerber Gerlach Giese
Glaser Göbel Graff Greiner Grimm Gross Grunwald Haag Hagen Hahne
Hamann Hansen Hartwig Hauser Heim Heine Helbig Hellmann Henn Hensel
Herbst Hertel Hesse Heyer Hildebrandt Hirsch Hoppe Huth Ihle Illing
Jahn Jordan Jost Kahl Kalb Kaminski Karl Keil Kern Kessler
Kiefer Kimmel Klaus Klemm Kluge Knapp Knoll Koehler Kohl Konrad
Korte Kraft Kramp Krebs Kröger Kroll Kruse Kugler Kühl Lamprecht
Lampe Landmann Lehner Leitner Lindemann Link Löffler Lohmann Lukas Lüders
Mai Marquardt Matthes Mauer Mayr Merz Metzger Meißner Moll Morgenstern
Müllner Nagel Nolte Oswald Pape Pfeifer Pfaff Pieper Prinz Rabe
Raab Rau Reichert Reiter Remmert Renner Richard Riedel Riegel Röder
Roos Rosenberg Rudolf Sachs Sailer Salzmann Sattler Schaefer Scheer Schell
Scherer Schiffer Schlüter Schmale Schnell Schott Schramm Schreiner Schwab Schwarzer
Siebert Singer Sperling Spiegel Staudt Steffen Steiner Stock Stolz Storch
Straub Strobel Thal Thiel Tietz Trapp Uhlig Unger Vollmer Wacker
Wagener Wahl Wallner Weidner Weigel Weise Weller Wendel Westphal Wetzel
Wichmann Wiese Winkelmann Wittmann Wolfram Wurm Zeller Zimmer Zink Zorn
"""


def firsts(cache: str) -> list[dict[str, object]]:
    """The rows (name, sex, count): each name and sex with the sum of the numbers that Munich publishes for it."""
    counts: dict[tuple[str, str], int] = {}
    for year, path in FILES.items():
        text = source.fetch(BASE + path, cache, f"munich-vornamen-{year}.csv", magic=b"").decode("utf-8-sig")
        rows = list(csv.reader(io.StringIO(text)))
        if [c.strip() for c in rows[0]] != ["vorname", "anzahl", "geschlecht"]:
            sys.exit(f"{year}: the columns are {rows[0]}, not vorname, anzahl and geschlecht")
        for row in rows[1:]:
            if not any(c.strip("\x1a ") for c in row):
                continue
            name, count, sex = (c.strip() for c in row)
            if sex not in SEX:
                sys.exit(f"{year}: {name!r} has the sex {sex!r}, not w or m")
            if count.isdigit() and tsv.NAME.match(name):
                counts[(name, SEX[sex])] = counts.get((name, SEX[sex]), 0) + int(count)
    return [{"name": name, "sex": sex, "count": n} for (name, sex), n in sorted(counts.items())]


def lasts() -> list[dict[str, object]]:
    """The rows (name, count) of the recalled last names, the ranked ones with the count 700000 over the square root of their rank."""
    ordered, _, rarer = (part.split() for part in LAST_RECALLED.partition("\n=\n"))
    names = ordered + rarer
    dupes = sorted({n for n in names if names.count(n) > 1})
    if dupes:
        sys.exit(f"last names: listed twice: {', '.join(dupes)}")
    for name in names:
        if not tsv.NAME.match(name):
            sys.exit(f"last names: {name!r} is not a name in letters")
    rows = [(name, round(700000 / rank ** 0.5)) for rank, name in enumerate(ordered, start=1)]
    rows += [(name, round(700000 / len(ordered) ** 0.5)) for name in rarer]
    return [{"name": name, "count": n} for name, n in sorted(rows)]


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    out = Path(a.out)
    out.mkdir(parents=True, exist_ok=True)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], firsts(a.cache))
    tsv.write(out / "last-name.tsv", ["name", "count"], lasts())


if __name__ == "__main__":
    main()
