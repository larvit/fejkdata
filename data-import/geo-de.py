"""Rebuild data/geo/DE/*.tsv from GeoNames' postal codes of Germany and a language model's recall of street names.

    data-import/geo-de.py [--cache DIR] [--out DIR]

- The regions, municipalities, localities and postal codes are GeoNames' DE.txt (CC BY 4.0). A region is a Land, named
  as ISO 3166-2:DE names it, a municipality a Kreis, and a locality a place name within a Kreis. A Kreis takes the
  name most of its rows spell, as the file spells some Kreise two ways. A row that names no Land, 87491 Jungholz,
  which is Austrian, is dropped. So is a row with no accuracy, which is mostly a firm's or an office's own postal
  code, such as "Agentur für Arbeit Cottbus". GeoNames publishes no population, so every table weighs a row by the
  postal codes it holds, which is an estimate of its people.
- The streets are a language model's recall of street names common in Germany, as no open register covers the country
  whole: the Länder publish theirs on terms of their own, and Bavaria's are not open. geo.invented_streets checks
  them. A street belongs to no place, and each name weighs the same.
"""
import argparse
import collections
import io
import sys
import zipfile
from collections.abc import Iterable
from pathlib import Path

import geo
import source
import tsv

TIMEZONE = "Europe/Berlin"
# Each Land's code in GeoNames and ISO 3166-2:DE, and its name in German.
LANDS = {
    "BB": "Brandenburg", "BE": "Berlin", "BW": "Baden-Württemberg", "BY": "Bayern", "HB": "Bremen", "HE": "Hessen",
    "HH": "Hamburg", "MV": "Mecklenburg-Vorpommern", "NI": "Niedersachsen", "NW": "Nordrhein-Westfalen",
    "RP": "Rheinland-Pfalz", "SH": "Schleswig-Holstein", "SL": "Saarland", "SN": "Sachsen", "ST": "Sachsen-Anhalt",
    "TH": "Thüringen",
}

STREETS_RECALLED = """
Hauptstraße
Bahnhofstraße
Dorfstraße
Schulstraße
Gartenstraße
Bergstraße
Kirchstraße
Poststraße
Waldstraße
Ringstraße
Lindenstraße
Feldstraße
Rosenstraße
Friedhofstraße
Industriestraße
Gewerbestraße
Fabrikstraße
Hafenstraße
Neue Straße
Alte Straße
Lange Straße
Kurze Straße
Breite Straße
Obere Straße
Untere Straße
Mittelstraße
Querstraße
Oststraße
Weststraße
Nordstraße
Südstraße
Sonnenstraße
Mondstraße
Sternstraße
Quellenstraße
Teichstraße
Seestraße
Uferstraße
Brückenstraße
Burgstraße
Schlossstraße
Klosterstraße
Mauerstraße
Turmstraße
Torstraße
Marktstraße
Rathausstraße
Kaiserstraße
Königstraße
Jägerstraße
Gärtnerstraße
Bäckerstraße
Fleischerstraße
Schmiedestraße
Müllerstraße
Schneiderstraße
Schusterstraße
Weberstraße
Töpferstraße
Zimmerstraße
Färberstraße
Mühlenstraße

Goethestraße
Schillerstraße
Lessingstraße
Mozartstraße
Beethovenstraße
Bismarckstraße
Kantstraße
Uhlandstraße
Hebelstraße
Eichendorffstraße
Heinestraße
Brahmsstraße
Schubertstraße
Haydnstraße
Wagnerstraße
Bachstraße
Humboldtstraße
Gutenbergstraße
Röntgenstraße
Kopernikusstraße
Keplerstraße
Einsteinstraße
Daimlerstraße
Benzstraße
Dieselstraße
Siemensstraße
Zeppelinstraße
Lilienthalstraße
Kleiststraße
Fontanestraße
Mörikestraße
Hölderlinstraße
Rilkestraße
Brechtstraße
Ludwigstraße
Wilhelmstraße
Friedrichstraße
Karlstraße
Marienstraße
Josefstraße
Annastraße
Luisenstraße
Charlottenstraße
Elisabethstraße
Paulstraße
Peterstraße
Martinstraße
Michaelstraße
Georgstraße
Johannisstraße
Katharinenstraße
Margaretenstraße
Barbarastraße
Albrechtstraße
Heinrichstraße
Hermannstraße
Ernststraße
Gustavstraße
Rudolfstraße
Viktoriastraße

Otto-Hahn-Straße
Max-Planck-Straße
Albert-Schweitzer-Straße
Konrad-Adenauer-Straße
Willy-Brandt-Straße
Theodor-Heuss-Straße
Friedrich-Ebert-Straße
Karl-Marx-Straße
Rosa-Luxemburg-Straße
Thomas-Mann-Straße
Gerhart-Hauptmann-Straße
Hermann-Löns-Straße
Wilhelm-Busch-Straße
Käthe-Kollwitz-Straße
Clara-Schumann-Straße
Robert-Koch-Straße
Paul-Ehrlich-Straße
Sebastian-Kneipp-Straße
Marie-Curie-Straße
Carl-Benz-Straße
Robert-Bosch-Straße
Ernst-Reuter-Straße
Kurt-Schumacher-Straße
Erich-Kästner-Straße
Bertolt-Brecht-Straße
Heinrich-Heine-Straße
Johann-Sebastian-Bach-Straße
Georg-Büchner-Straße
Friedrich-Schiller-Straße
Johann-Wolfgang-von-Goethe-Straße

Berliner Straße
Münchner Straße
Hamburger Straße
Kölner Straße
Frankfurter Straße
Leipziger Straße
Dresdner Straße
Stuttgarter Straße
Bremer Straße
Hannoversche Straße
Nürnberger Straße
Bonner Straße
Mainzer Straße
Wiesbadener Straße
Lübecker Straße
Magdeburger Straße
Potsdamer Straße
Rostocker Straße
Kasseler Straße
Würzburger Straße

Am Markt
Am Bahnhof
Am Berg
Am Bach
Am Teich
Am See
Am Wald
Am Park
Am Sportplatz
Am Rathaus
Am Schloss
Am Hang
Am Graben
Am Hügel
Am Anger
Am Brunnen
Am Kamp
Am Ring
Am Mühlenbach
Am Waldrand
Am Feldrain
Am Sonnenhang
Am Rosengarten

An der Kirche
An der Mühle
An der Linde
An der Eiche
An der Schule
An der Weide
An der Bahn
An der Au
An der Heide
An der Burg
An der Stadtmauer
An der Promenade
An den Teichen
An den Wiesen
An den Linden

Auf dem Berg
Auf der Höhe
Auf der Heide
Auf dem Anger
Auf der Wiese
Auf dem Hügel
Auf der Platte
Auf der Au
Auf dem Kamp
Auf der Schanze
Auf dem Rücken
Auf der Hardt

In der Aue
In den Gärten
In der Wanne
Im Winkel
Im Grund
Im Tal
Im Hagen
Im Feld
Im Brühl
Im Dorf
Im Moos
Im Sande
Im Busch
Im Rosengarten
Im Wiesengrund
Im Eichenhain
Im Birkenwäldchen
Im Neubruch
Im Lehmkamp

Zur Mühle
Zum Sportplatz
Zur Heide
Zum Bahnhof
Zum Wald
Zur Sonne
Zum Stausee
Zur Quelle
Zum Hafen
Zur Schule
Zum Anger
Zur Ziegelei
Zum Steinbruch
Zum Weiher
Zum Hohenstein

Eichenweg
Buchenweg
Tannenweg
Fichtenweg
Kiefernweg
Ahornweg
Lindenweg
Ulmenweg
Erlenweg
Weidenweg
Birkenweg
Rosenweg
Tulpenweg
Nelkenweg
Veilchenweg
Lilienweg
Fliederweg
Jasminweg
Primelweg
Akazienweg
Heckenweg
Sandweg
Lehmweg
Steinweg
Kiesweg
Waldweg
Feldweg
Wiesenweg
Heideweg
Moorweg
Dammweg
Schulweg
Kirchweg
Postweg
Bahnweg
Parkweg
Gartenweg
Hohlweg
Fußweg
Radweg
Bergweg
Talweg
Höhenweg
Hangweg
Rainweg
Sonnenweg
Panoramaweg
Herrenweg
Mühlenweg
Lerchenweg
Finkenweg
Amselweg
Drosselweg
Meisenweg
Starenweg
Schwalbenweg
Storchenweg
Adlerweg
Falkenweg
Fuchsweg
Hasenweg
Rehweg
Hirschweg
Wolfsweg
Bärenweg
Igelweg
Dachsweg

Kastanienallee
Lindenallee
Birkenallee
Eichenallee
Buchenallee
Platanenallee
Ahornallee
Kirschenallee
Apfelallee
Pappelallee
Parkallee
Schlossallee
Gartenallee
Bahnhofsallee
Hindenburgallee
Stadtallee

Marktplatz
Bahnhofsplatz
Kirchplatz
Schlossplatz
Rathausplatz
Marienplatz
Dorfplatz
Festplatz
Messeplatz
Europaplatz
Friedensplatz
Freiheitsplatz
Lindenplatz
Brunnenplatz
Theaterplatz
Neumarkt
Altmarkt
Holzmarkt
Kornmarkt
Fischmarkt

Ahrstraße
Moselstraße
Saarstraße
Neckarstraße
Mainstraße
Donaustraße
Elbstraße
Weserstraße
Emsstraße
Rheinstraße
Harzstraße
Rhönstraße
Eifelstraße
Taunusstraße
Spessartstraße
Odenwaldstraße
Allgäustraße
Alpenstraße
Schwarzwaldstraße
Frankenstraße
Ostseestraße
Nordseestraße
Lahnstraße
Isarstraße
Lechstraße
Innstraße
Spreestraße
Havelstraße
Oderstraße
Werrastraße
Fuldastraße
Leinestraße
Ruhrstraße
Lippestraße
Wuppertalstraße
Sauerlandstraße
Westerwaldstraße
Hunsrückstraße
Vogelsbergstraße
Thüringer Straße
Sächsische Straße
Bayerische Straße
Schlesische Straße
Pommersche Straße
Holsteiner Straße
Mecklenburger Straße
Friesenstraße
Hessenstraße
Rheinlandstraße
"""


def lines(cache: Path) -> list[str]:
    """The lines of GeoNames' DE.txt."""
    z = zipfile.ZipFile(io.BytesIO(source.fetch(geo.GEONAMES.format("DE"), cache, "DE.zip", magic=b"PK")))
    return z.read("DE.txt").decode("utf-8").splitlines()


def kreis_names(counts: collections.Counter[tuple[str, str]]) -> dict[str, str]:
    """Each Kreis code's name: the one most rows spell, the first in order in a tie. counts maps (code, name) to rows."""
    best: dict[str, tuple[int, str]] = {}
    for (code, name), n in sorted(counts.items()):
        if code not in best or n > best[code][0]:
            best[code] = (n, name)
    return {code: name for code, (_, name) in best.items()}


def places(raw: Iterable[str]) -> tuple[list[geo.Row], list[geo.Row], list[geo.Row], list[geo.Row]]:
    """The region, municipality, locality and postal-code rows of GeoNames' lines, each sorted by key. A line's fields
    are GeoNames' postal code file: the postal code is the second, the place the third, the Land's code the fifth, the
    Kreis its eighth (name) and ninth (code), and the point its tenth and eleventh."""
    land_of: dict[str, str] = {}
    names: collections.Counter[tuple[str, str]] = collections.Counter()
    points: dict[tuple[str, str], list[geo.Row]] = collections.defaultdict(list)
    codes: dict[tuple[str, str], set[str]] = collections.defaultdict(set)
    for line in raw:
        f = line.split("\t")
        land, kreis, place = f[4], f[8], f[2]
        if not land or not f[11]:
            continue
        if land not in LANDS:
            sys.exit(f"{f[1]} {place}: the Land {land!r} is none of {', '.join(sorted(LANDS))}")
        if not kreis:
            sys.exit(f"{f[1]} {place}: names no Kreis")
        land_of[kreis] = land
        names[(kreis, f[7])] += 1
        points[(kreis, place)].append({"lat": float(f[9]), "lon": float(f[10])})
        codes[(kreis, place)].add(f[1])
    kreise = kreis_names(names)
    ordinal: dict[str, int] = collections.Counter()
    localities: list[geo.Row] = []
    postal_codes: list[geo.Row] = []
    for kreis, place in sorted(codes):
        ordinal[kreis] += 1
        code = f"{kreis}-{ordinal[kreis]:03d}"
        lat, lon = geo.centroid(points[(kreis, place)]) or (0.0, 0.0)
        localities.append({"code": code, "name": place, "municipality": kreis, "postal-codes": len(codes[(kreis, place)]), "lat": f"{lat:.4f}", "lon": f"{lon:.4f}"})
        postal_codes += [{"code": c, "locality": code} for c in codes[(kreis, place)]]
    weight: collections.Counter[str] = collections.Counter()
    in_land: collections.Counter[str] = collections.Counter()
    for row in localities:
        weight[row["municipality"]] += row["postal-codes"]
        in_land[land_of[row["municipality"]]] += row["postal-codes"]
    municipalities = [{"code": k, "name": kreise[k], "region": land_of[k], "postal-codes": weight[k]} for k in sorted(kreise)]
    regions = [{"code": c, "name": LANDS[c], "postal-codes": in_land[c], "timezone": TIMEZONE} for c in sorted(in_land)]
    return regions, municipalities, localities, sorted(postal_codes, key=lambda r: (r["code"], r["locality"]))


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(geo.CACHE))
    p.add_argument("--out", default=str(geo.DATA / "DE"))
    a = p.parse_args()
    cache, out = geo.directories(a)
    regions, municipalities, localities, postal_codes = places(lines(cache))
    tsv.write(out / "region.tsv", ["code", "name", "postal-codes", "timezone"], regions)
    tsv.write(out / "municipality.tsv", ["code", "name", "region", "postal-codes"], municipalities)
    tsv.write(out / "locality.tsv", ["code", "name", "municipality", "postal-codes", "lat", "lon"], localities)
    tsv.write(out / "postal-code.tsv", ["code", "locality"], postal_codes)
    tsv.write(out / "street.tsv", ["name"], geo.invented_streets(STREETS_RECALLED))


if __name__ == "__main__":
    main()
