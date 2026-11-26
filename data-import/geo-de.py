"""Rebuild data/geo/DE/*.tsv from GeoNames' postal codes of Germany and a language model's recall of street names.

    data-import/geo-de.py [--cache DIR] [--out DIR]

- The regions, municipalities, localities and postal codes are GeoNames' DE.txt (CC BY 4.0). A region is a Land, named
  as ISO 3166-2:DE names it, a municipality a Kreis, and a locality a place name within a Kreis, keyed by the Kreis's
  code and the place name in ASCII. All of Hamburg and of Berlin is one place each, as a letter addresses it. A row
  that names no Land, 87491 Jungholz, which is Austrian, is dropped. So is a row with no accuracy, GeoNames' measure of
  how exactly it knows the point, as it is mostly a firm's or an office's own postal code, such as "Agentur für Arbeit
  Cottbus". GeoNames publishes no population, so a locality weighs the postal codes it holds, and a Kreis and a Land
  the sum of theirs: a count of postal codes, which favours small places over people.
- The streets are a language model's recall of street names common in Germany, as no open register is known to cover the
  country whole. geo.invented_streets checks them. A street belongs to no place, and each name weighs the same.
"""
import argparse
import collections
import io
import re
import statistics
import sys
import unicodedata
import zipfile
from collections.abc import Iterable
from pathlib import Path
from typing import NamedTuple

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
TRANSLITERATION = str.maketrans({"ä": "ae", "ö": "oe", "ü": "ue", "Ä": "Ae", "Ö": "Oe", "Ü": "Ue", "ß": "ss"})
# The Kreis of each city state, whose places GeoNames names after the city and a district, "Hamburg Altona-Altstadt".
CITY_STATES = {"02000": "Hamburg", "11000": "Berlin"}
# A Kreis whose name GeoNames writes in English.
GERMAN_KREIS_NAMES = {"09564": "Kreisfreie Stadt Nürnberg"}

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


class Entry(NamedTuple):
	"""A GeoNames row that places a postal code in a Land."""
	postal_code: str
	place: str
	land: str
	kreis_code: str
	kreis: str
	lat: float
	lon: float


def lines(cache: Path) -> list[str]:
	"""The lines of GeoNames' DE.txt."""
	z = zipfile.ZipFile(io.BytesIO(source.fetch(geo.GEONAMES.format("DE"), cache, "DE.zip", magic=b"PK")))
	return z.read("DE.txt").decode("utf-8").splitlines()


def entries(raw: Iterable[str]) -> list[Entry]:
	"""The rows of GeoNames' lines that place a postal code in a Land. A line's fields are GeoNames' postal code file:
	the postal code is the second, the place the third, the Land's code the fifth, the Kreis its eighth (name) and ninth
	(code), the point its tenth and eleventh, and the accuracy the twelfth."""
	out: list[Entry] = []
	for line in raw:
		f = line.split("\t")
		land, accuracy = f[4], f[11]
		if not land or not accuracy:
			continue
		if land not in LANDS:
			sys.exit(f"{f[1]} {f[2]}: the Land {land!r} is none of {', '.join(sorted(LANDS))}")
		if not f[8]:
			sys.exit(f"{f[1]} {f[2]}: names no Kreis")
		out.append(Entry(f[1], CITY_STATES.get(f[8], f[2]), land, f[8], f[7], float(f[9]), float(f[10])))
	return out


def kreis_name(code: str, spelt: str) -> str:
	"""A Kreis's name as a German writes it: GeoNames' "Salzgitter, Stadt" is "Stadt Salzgitter"."""
	if code in GERMAN_KREIS_NAMES:
		return GERMAN_KREIS_NAMES[code]
	name, comma, kind = spelt.partition(", ")
	return f"{kind} {name}" if comma else spelt


def kreis_names(rows: Iterable[Entry]) -> dict[str, str]:
	"""Each Kreis code's name. A code that GeoNames spells two ways exits, as there is no telling which is right."""
	spelt: dict[str, str] = {}
	for e in rows:
		if spelt.setdefault(e.kreis_code, e.kreis) != e.kreis:
			sys.exit(f"Kreis {e.kreis_code} is spelt {spelt[e.kreis_code]!r} and {e.kreis!r}")
	return {code: kreis_name(code, name) for code, name in spelt.items()}


def slug(place: str) -> str:
	"""A place name in lower-case ASCII, the umlauts and ß written as German does without them, and each run of other
	characters a hyphen: "Müllrose" is "muellrose"."""
	ascii_name = unicodedata.normalize("NFKD", place.translate(TRANSLITERATION)).encode("ascii", "ignore").decode()
	return re.sub(r"[^a-z0-9]+", "-", ascii_name.lower()).strip("-")


def localities_of(rows: Iterable[Entry]) -> tuple[list[geo.Row], list[geo.Row]]:
	"""The locality and postal-code rows: a locality is a place name within a Kreis, at the mean of its rows' points,
	and holds every postal code any of those rows gives it, so two villages of one name in a Kreis are one place.
	A code that two places in a Kreis share exits."""
	held: dict[tuple[str, str], list[Entry]] = collections.defaultdict(list)
	for e in rows:
		held[(e.kreis_code, e.place)].append(e)
	localities: list[geo.Row] = []
	postal_codes: list[geo.Row] = []
	named: dict[str, str] = {}
	for kreis, place in sorted(held):
		code = f"{kreis}-{slug(place)}"
		if named.setdefault(code, place) != place:
			sys.exit(f"{code}: both {named[code]!r} and {place!r}")
		codes = {e.postal_code for e in held[(kreis, place)]}
		lat, lon = statistics.fmean(e.lat for e in held[(kreis, place)]), statistics.fmean(e.lon for e in held[(kreis, place)])
		localities.append({"code": code, "name": place, "municipality": kreis, "postal-codes": len(codes), "lat": f"{lat:.4f}", "lon": f"{lon:.4f}"})
		postal_codes += [{"code": c, "locality": code} for c in codes]
	return localities, sorted(postal_codes, key=lambda r: (r["code"], r["locality"]))


def places(raw: Iterable[str]) -> tuple[list[geo.Row], list[geo.Row], list[geo.Row], list[geo.Row]]:
	"""The region, municipality, locality and postal-code rows of GeoNames' lines, each sorted by key."""
	rows = entries(raw)
	kreise = kreis_names(rows)
	land_of = {e.kreis_code: e.land for e in rows}
	localities, postal_codes = localities_of(rows)
	in_kreis: collections.Counter[str] = collections.Counter()
	in_land: collections.Counter[str] = collections.Counter()
	for row in localities:
		in_kreis[row["municipality"]] += row["postal-codes"]
		in_land[land_of[row["municipality"]]] += row["postal-codes"]
	municipalities = [{"code": k, "name": kreise[k], "region": land_of[k], "postal-codes": in_kreis[k]} for k in sorted(kreise)]
	regions = [{"code": c, "name": LANDS[c], "postal-codes": in_land[c], "timezone": TIMEZONE} for c in sorted(in_land)]
	return regions, municipalities, localities, postal_codes


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
