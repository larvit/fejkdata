"""Rebuild data/geo/FR/*.tsv and address.json from the Base Adresse Nationale (BAN), INSEE's legal populations and a
language model's recall of street names.

    data-import/geo-fr.py [--cache DIR] [--out DIR]

- BAN's address file holds every address in France, each with its commune, code postal, house number and point. A
  commune is a locality, keyed by its INSEE code, since many communes share a name. Its codes postaux are the ones its
  addresses carry, each weighted by those addresses, and its point is the mean of theirs. BAN files an address of
  Paris, Lyon or Marseille under its arrondissement; all of each city is one place, as a letter addresses it.
  address.json weighs each length of house number, bare or followed by bis, by BAN's addresses. An address with another
  suffix, such as ter or a, counts for no length.
- INSEE's file of legal populations names each commune, with its département and région, and gives its municipal
  population. A département is a municipality and a région a region, and each weighs the people of its communes
  here. A commune without people, such as the villages destroyed at Verdun, ships no place. Neither does a commune the
  file leaves out: those of Mayotte, and those outside the régions, such as Saint-Pierre-et-Miquelon's.
- The streets are a language model's recall of street names common in France, until todo.md item 271 imports BAN's.
  geo.invented_streets checks them. A street belongs to no place, and each name weighs the same.
"""
import argparse
import collections
import csv
import gzip
import io
import sys
import zipfile
from collections.abc import Iterable, Iterator, Mapping
from dataclasses import dataclass, field
from pathlib import Path

import geo
import source
import tsv

BAN = "https://adresse.data.gouv.fr/data/ban/adresses/latest/csv/adresses-france.csv.gz"
POPULATION = "https://www.insee.fr/fr/statistiques/fichier/8680726/ensemble.zip"
TIMEZONE = "Europe/Paris"
# The overseas régions, by INSEE code, which keep their own time.
OVERSEAS_TIMEZONES = {"01": "America/Guadeloupe", "02": "America/Martinique", "03": "America/Cayenne", "04": "Indian/Reunion"}
CITY_NAMES = {"75056": "Paris", "69123": "Lyon", "13055": "Marseille"}
# Each arrondissement's INSEE code, mapped to its city's: Paris's 20, Lyon's 9 and Marseille's 16.
CITY_OF = {
    **{f"751{n:02}": "75056" for n in range(1, 21)},
    **{f"6938{n}": "69123" for n in range(1, 10)},
    **{f"132{n:02}": "13055" for n in range(1, 17)},
}
STREETS_RECALLED = """
Allée des Acacias
Allée des Bouleaux
Allée des Cèdres
Allée des Chênes
Allée des Érables
Allée des Marronniers
Allée des Peupliers
Allée des Pins
Allée des Roses
Allée des Tilleuls
Allée du Château
Allée du Parc
Avenue Foch
Avenue Georges Clemenceau
Avenue Jean Jaurès
Avenue Pasteur
Avenue Victor Hugo
Avenue de Verdun
Avenue de la Gare
Avenue de la Libération
Avenue de la Paix
Avenue de la République
Avenue du Général Leclerc
Avenue du Général de Gaulle
Avenue du Stade
Boulevard Carnot
Boulevard Foch
Boulevard Gambetta
Boulevard Jean Jaurès
Boulevard Pasteur
Boulevard Victor Hugo
Boulevard Voltaire
Boulevard de la Liberté
Boulevard de la République
Chemin Vert
Chemin de Halage
Chemin de la Chapelle
Chemin de la Croix
Chemin de la Fontaine
Chemin des Carrières
Chemin des Crêtes
Chemin des Écoliers
Chemin des Prés
Chemin des Vignes
Chemin du Bois
Chemin du Cimetière
Chemin du Lavoir
Chemin du Moulin
Chemin du Stade
Cours Victor Hugo
Cours de la République
Grande Rue
Impasse de l'Église
Impasse de la Fontaine
Impasse des Acacias
Impasse des Cerisiers
Impasse des Jardins
Impasse des Lilas
Impasse des Mésanges
Impasse des Roses
Impasse des Tilleuls
Impasse des Vignes
Impasse du Château
Impasse du Lavoir
Impasse du Moulin
Impasse du Verger
Place Carnot
Place Jean Jaurès
Place Victor Hugo
Place de l'Église
Place de la Fontaine
Place de la Gare
Place de la Liberté
Place de la Mairie
Place de la République
Place du Champ de Foire
Place du Général de Gaulle
Place du Marché
Quai Gambetta
Quai de la République
Quai du Port
Route Nationale
Route de Lyon
Route de Paris
Route de la Forêt
Route de la Gare
Route des Vignes
Route du Moulin
Rue Albert Camus
Rue Alphonse Daudet
Rue Ampère
Rue Anatole France
Rue Aristide Briand
Rue Auguste Renoir
Rue Basse
Rue Carnot
Rue Charles de Gaulle
Rue Claude Debussy
Rue Claude Monet
Rue Colbert
Rue Danton
Rue Denis Papin
Rue Descartes
Rue Diderot
Rue Émile Zola
Rue Frédéric Mistral
Rue Gabriel Péri
Rue Gambetta
Rue George Sand
Rue Georges Bizet
Rue Georges Clemenceau
Rue Gustave Flaubert
Rue Guy Môquet
Rue Haute
Rue Hector Berlioz
Rue Henri Barbusse
Rue Hoche
Rue Honoré de Balzac
Rue Jean Jaurès
Rue Jean Macé
Rue Jean Monnet
Rue Jean Moulin
Rue Jean-Jacques Rousseau
Rue Jeanne d'Arc
Rue Jules Ferry
Rue Jules Guesde
Rue Jules Verne
Rue Kléber
Rue Lafayette
Rue Lamartine
Rue Lavoisier
Rue Léon Blum
Rue Louis Blériot
Rue Louis Pasteur
Rue Louise Michel
Rue Marceau
Rue Marie Curie
Rue Maurice Ravel
Rue Mirabeau
Rue Molière
Rue Montaigne
Rue Nationale
Rue Neuve
Rue Notre-Dame
Rue Pascal
Rue Pasteur
Rue Paul Bert
Rue Paul Cézanne
Rue Paul Éluard
Rue Paul Verlaine
Rue Pierre Brossolette
Rue Pierre Curie
Rue Pierre et Marie Curie
Rue Principale
Rue Racine
Rue René Cassin
Rue Robert Schuman
Rue Roger Salengro
Rue Roland Garros
Rue Saint-Exupéry
Rue Saint-Jacques
Rue Saint-Jean
Rue Saint-Louis
Rue Saint-Martin
Rue Saint-Michel
Rue Saint-Nicolas
Rue Saint-Pierre
Rue Saint-Roch
Rue Sainte-Anne
Rue Sainte-Catherine
Rue Salvador Allende
Rue Thiers
Rue Traversière
Rue Victor Hugo
Rue Voltaire
Rue d'Alsace
Rue d'Anjou
Rue de Bordeaux
Rue de Bourgogne
Rue de Bretagne
Rue de Champagne
Rue de Lorraine
Rue de Lyon
Rue de Nantes
Rue de Normandie
Rue de Paris
Rue de Picardie
Rue de Provence
Rue de Strasbourg
Rue de Verdun
Rue de l'Abbaye
Rue de l'Égalité
Rue de l'Église
Rue de l'Est
Rue de l'Étang
Rue de l'Hôpital
Rue de l'Hôtel de Ville
Rue de l'Industrie
Rue de l'Ouest
Rue de la Chapelle
Rue de la Colline
Rue de la Côte
Rue de la Croix
Rue de la Fontaine
Rue de la Forêt
Rue de la Forge
Rue de la Fraternité
Rue de la Gare
Rue de la Libération
Rue de la Liberté
Rue de la Mairie
Rue de la Mer
Rue de la Montagne
Rue de la Paix
Rue de la Plage
Rue de la Plaine
Rue de la Poste
Rue de la République
Rue de la Résistance
Rue de la Rivière
Rue de la Source
Rue de la Tour
Rue de la Tuilerie
Rue de la Vallée
Rue de la Verrerie
Rue des Acacias
Rue des Alouettes
Rue des Artisans
Rue des Bleuets
Rue des Bouleaux
Rue des Bruyères
Rue des Carrières
Rue des Cerisiers
Rue des Champs
Rue des Chênes
Rue des Coquelicots
Rue des Écoles
Rue des Érables
Rue des Fauvettes
Rue des Forgerons
Rue des Fossés
Rue des Genêts
Rue des Glycines
Rue des Halles
Rue des Hirondelles
Rue des Hortensias
Rue des Iris
Rue des Jardins
Rue des Jonquilles
Rue des Lilas
Rue des Marronniers
Rue des Merles
Rue des Mésanges
Rue des Meuniers
Rue des Mimosas
Rue des Noyers
Rue des Peupliers
Rue des Pins
Rue des Pinsons
Rue des Platanes
Rue des Pommiers
Rue des Prés
Rue des Primevères
Rue des Remparts
Rue des Roses
Rue des Rossignols
Rue des Sapins
Rue des Saules
Rue des Sports
Rue des Tanneurs
Rue des Tilleuls
Rue des Tisserands
Rue des Tulipes
Rue des Vignerons
Rue des Vignes
Rue des Violettes
Rue du 8 Mai 1945
Rue du 11 Novembre
Rue du 14 Juillet
Rue du 19 Mars 1962
Rue du Bois
Rue du Bourg
Rue du Calvaire
Rue du Canal
Rue du Centre
Rue du Château
Rue du Château d'Eau
Rue du Cimetière
Rue du Collège
Rue du Commerce
Rue du Coteau
Rue du Couvent
Rue du Four
Rue du Général Leclerc
Rue du Général de Gaulle
Rue du Lac
Rue du Lavoir
Rue du Levant
Rue du Marché
Rue du Maréchal Foch
Rue du Maréchal Joffre
Rue du Maréchal Leclerc
Rue du Midi
Rue du Moulin
Rue du Nord
Rue du Parc
Rue du Pont
Rue du Port
Rue du Pré
Rue du Presbytère
Rue du Puits
Rue du Ruisseau
Rue du Soleil
Rue du Stade
Rue du Temple
Rue du Vallon
Rue du Village
Ruelle de l'Église
Ruelle du Four
Ruelle du Moulin
"""


@dataclass
class Tally:
    """What BAN's addresses add up to."""

    # Each commune's addresses under each of its codes postaux.
    codes: collections.defaultdict[str, collections.Counter[str]] = field(default_factory=lambda: collections.defaultdict(collections.Counter))
    numbers: collections.Counter[tuple[int, bool]] = field(default_factory=collections.Counter)
    # Each commune's latitudes summed, its longitudes summed, and how many points that is.
    points: collections.defaultdict[str, list[float]] = field(default_factory=lambda: collections.defaultdict(lambda: [0.0, 0.0, 0]))

    def centre(self, commune: str) -> tuple[str, str] | None:
        """The mean of a commune's points as lat and lon, or None where no address of it has one."""
        lat, lon, n = self.points.get(commune, (0.0, 0.0, 0))
        return (f"{lat / n:.4f}", f"{lon / n:.4f}") if n else None


def city(commune: str) -> str:
    """A commune's INSEE code, or its city's where the code is an arrondissement's."""
    return CITY_OF.get(commune, commune)


def number_key(number: str, suffix: str) -> tuple[int, bool] | None:
    """The key geo.number_range counts a house number under, a number with bis counting as one with a letter; None for a
    number outside 1 to 9999 or with another suffix."""
    if not number.isdigit() or not 1 <= int(number) <= geo.NUMBER_RANGES[-1][1] or suffix.lower() not in ("", "bis"):
        return None
    return geo.number_range(int(number), bool(suffix))


def tally(addresses: Iterable[Mapping[str, str]]) -> Tally:
    """addresses are the rows of BAN's file, each column under its name. One without a code postal is left out."""
    t = Tally()
    for a in addresses:
        if not a["code_postal"]:
            continue
        commune = city(a["code_insee"])
        t.codes[commune][a["code_postal"]] += 1
        if a["lat"] and a["lon"]:
            point = t.points[commune]
            point[0], point[1], point[2] = point[0] + float(a["lat"]), point[1] + float(a["lon"]), point[2] + 1
        key = number_key(a["numero"], a["rep"])
        if key:
            t.numbers[key] += 1
    return t


def communes(rows: Iterable[Mapping[str, str]]) -> dict[str, geo.Row]:
    """Each commune by its code, with its name, département, région and municipal population. rows are those of INSEE's
    donnees_communes.csv, which lists a city's arrondissements and not the city: their people count for the city."""
    out: dict[str, geo.Row] = {}
    for r in rows:
        code = city(r["COM"])
        commune = out.setdefault(code, {"code": code, "name": CITY_NAMES.get(code, r["Commune"]), "municipality": r["DEP"], "region": r["REG"], "population": 0})
        commune["population"] += int(r["PMUN"])
    return out


def places(t: Tally, communes: Mapping[str, geo.Row], departments: Mapping[str, str], regions: Mapping[str, str]) -> tuple[list[geo.Row], list[geo.Row], list[geo.Row], list[geo.Row]]:
    """The region, municipality, locality and postal-code rows, each sorted by key. communes is what the function of
    that name returns; departments and regions map a code to its name."""
    localities, region_of = [], {}
    in_department: collections.Counter[str] = collections.Counter()
    in_region: collections.Counter[str] = collections.Counter()
    for code, c in sorted(communes.items()):
        centre = t.centre(code)
        if not c["population"] or not centre:
            continue
        localities.append({"code": code, "name": c["name"], "municipality": c["municipality"], "population": c["population"], "lat": centre[0], "lon": centre[1]})
        region_of[c["municipality"]] = c["region"]
        in_department[c["municipality"]] += c["population"]
        in_region[c["region"]] += c["population"]
    postal_codes = sorted(({"code": code, "locality": row["code"], "addresses": n} for row in localities for code, n in t.codes[row["code"]].items()), key=lambda r: (r["code"], r["locality"]))
    municipalities = [{"code": d, "name": departments[d], "region": region_of[d], "population": in_department[d]} for d in sorted(in_department)]
    region_rows = [{"code": r, "name": regions[r], "population": in_region[r], "timezone": OVERSEAS_TIMEZONES.get(r, TIMEZONE)} for r in sorted(in_region)]
    return region_rows, municipalities, localities, postal_codes


def address_json(numbers: Mapping[tuple[int, bool], int]) -> str:
    """France's address, its street-number choice weighted by numbers, which counts house numbers by number_key."""
    items = ",\n".join(
        f'    {{ "format": "{{int({low},{top})}}{" bis" if bis else ""}", "weight": {numbers[(i, bis)]} }}'
        for bis in (False, True) for i, (low, top) in enumerate(geo.NUMBER_RANGES) if numbers[(i, bis)]
    )
    return (
        '{\n  "format": "{street-number} {street}\\n{postal-code} {locality}",\n  "street": "{/geo.FR.street}",\n'
        f'  "street-number": [\n{items}\n  ],\n'
        '  "postal-code": "{.locality as l}{l.postal-code.code}",\n  "locality": "{l.name}",\n'
        '  "municipality": "{l..municipality.name}",\n  "region": "{l..municipality..region.name}"\n}\n'
    )


def ban(cache: Path) -> Iterator[dict[str, str]]:
    """Each row of BAN's address file. The file holds a few bytes that are not UTF-8, in columns the import does
    not read, so those are replaced."""
    body = source.fetch(BAN, cache, "ban-adresses-france.csv.gz", magic=b"\x1f\x8b")
    with gzip.open(io.BytesIO(body), "rt", encoding="utf-8", errors="replace", newline="") as text:
        yield from csv.DictReader(text, delimiter=";")


def insee(cache: Path, table: str) -> list[dict[str, str]]:
    """The rows of one table of INSEE's file of legal populations."""
    z = zipfile.ZipFile(io.BytesIO(source.fetch(POPULATION, cache, "insee-populations-2023.zip", magic=b"PK")))
    return list(csv.DictReader(io.StringIO(z.read(f"donnees_{table}.csv").decode("utf-8-sig")), delimiter=";"))


def main() -> None:
    p = argparse.ArgumentParser(description=" ".join(__doc__.split("\n\n")[0].split()))
    p.add_argument("--cache", default=str(geo.CACHE))
    p.add_argument("--out", default=str(geo.DATA / "FR"))
    cache, out = geo.directories(p.parse_args())
    t = tally(ban(cache))
    populated = communes(insee(cache, "communes"))
    unplaced = sorted(set(t.codes) - set(populated))
    print(f"{len(unplaced)} communes of BAN are missing from INSEE's file, dropped: {', '.join(unplaced)}", file=sys.stderr)
    departments = {r["DEP"]: r["Département"] for r in insee(cache, "departements")}
    region_names = {r["REG"]: r["Région"] for r in insee(cache, "regions")}
    regions, municipalities, localities, postal_codes = places(t, populated, departments, region_names)
    print(f"{len(populated) - len(localities)} communes of INSEE's file have no people or no address in BAN, dropped", file=sys.stderr)
    tsv.write(out / "region.tsv", geo.COLUMNS["region"], regions)
    tsv.write(out / "municipality.tsv", geo.COLUMNS["municipality"], municipalities)
    tsv.write(out / "locality.tsv", ["code", *geo.COLUMNS["locality"]], localities)
    tsv.write(out / "postal-code.tsv", geo.COLUMNS["postal_code"], postal_codes)
    tsv.write(out / "street.tsv", ["name"], geo.invented_streets(STREETS_RECALLED))
    (out / "address.json").write_text(address_json(t.numbers), encoding="utf-8")


if __name__ == "__main__":
    main()
