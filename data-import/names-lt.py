"""Rebuild data/lt_LT/first-name.tsv and last-name.tsv: the first names from the State Enterprise Centre of Registers' top 50 and a language model's recall, the last names from the recall.

    data-import/names-lt.py [--cache DIR] [--out DIR] [--men URL_OR_FILE] [--women URL_OR_FILE]

- The 50 commonest men's names and women's names keep the counts the Centre of Registers published for the people in
  the Population Register, born since 1900 (CC BY 4.0, "Valstybės įmonė Registrų centras", data.gov.lt, the datasets
  `vyru_vardai` and `moteru_vardai`). The register holds the foreigners' names too, spelt as it spells them.
- Every other first name is a language model's recall of a name common in Lithuania, and so is every last name: no open
  file of Lithuanian last names was found, and the Centre publishes no first name below its top 50. A list holds the
  commoner names from the most to the least common, then a line `=`, then the rarer ones. A commoner first name has the
  smallest published count times 50 over its rank, and the rarer ones share the count of the last rank. A commoner last
  name has the count 100000 over the square root of its rank, and the rarer ones share the count of the last rank.
  These counts are estimates, not published figures.
- A last name has a row under `m` and, under `f`, a row for each feminine form of it: the unmarried and the married.
"""
import argparse
import json
from collections.abc import Sequence
from pathlib import Path

import recalled_names
import source
import tsv

REGISTER = "https://get.data.gov.lt/datasets/gov/rc/gr/{table}"
OUT = Path(__file__).resolve().parent.parent / "data" / "lt_LT"
CACHE = Path(__file__).resolve().parent / "cache"
TOP = 100000

MEN_RECALLED = """
Rokas Dainius Giedrius Ramūnas Ričardas Robertas Valdas Vidas Evaldas Remigijus
Aurimas Dovydas Jurgis Justinas Karolis Titas Nerijus Linas Rytis Donatas
Albertas Alfredas Henrikas Rolandas Vilius Eimantas Tadas Zigmas Žydrūnas Edgaras
Simonas Deividas Ernestas Laimonas Sigitas Raimondas Vaidotas Audrius Julius Ąžuolas
Dominykas Benas Augustas Modestas Vygantas Kristupas Eligijus Šarūnas Arvydas Edmundas
=
Adomas Aivaras Algis Almantas Aloyzas Anicetas Aistis Bernardas Dalius Dangiras
Domantas Eitvydas Erikas Gytis Gvidas Jaunius Jokūbas Juras Kajus Kornelijus
Kostas Laurynas Liudas Mykolas Nojus Ovidijus Pijus Raimundas Ramutis Rimvydas
Stepas Svajūnas Tautvydas Valentinas Vaidas Vainius Vilmantas Visvaldas Zenonas Žilvinas
Žygimantas
Aidas Alvydas Andrejus Arminas Armandas Arnoldas Aurelijus Bartas Beniaminas Bonifacas
Daumantas Daugirdas Deimantas Dobilas Eugenijus Faustas Feliksas Gabrielius Gedas Gerardas
Gilvydas Gražvydas Gvidonas Irmantas Izidorius Jaroslavas Jeronimas Juozapas Justas Kęstas
Kipras Klemensas Kristijonas Laimis Lauras Leonas Leonardas Lionginas Marijus Mečislovas
Nauris Nedas Norbertas Normantas Olegas Osvaldas Pranciškus Rapolas Rimas Rimgaudas
Romas Šimas Teodoras Valdemaras Valerijus Vasilijus Vidmantas Virgilijus Vitalijus Vitas
Vladislavas Zigmantas
"""

WOMEN_RECALLED = """
Gintarė Giedrė Dalia Vilma Eglė Agnė Aušra Ingrida Loreta Rūta
Laima Violeta Sandra Indrė Neringa Dovilė Justina Greta Simona Kamilė
Austėja Sofija Liepa Milda Gabija Edita Renata Audronė Skaidrė Raimonda
Lijana Roma Sigita Gražina Vanda Teresė Vaida Aistė Egidija Dainora
Rimantė Vaiva Karolina Paulina Akvilė Samanta Meda Ugnė Goda Jūratė
=
Agota Alma Alvyda Angelė Asta Audra Aurelija Aurika Bernadeta Dainė
Dorota Eimantė Eugenija Evelina Gerda Gita Gytė Ilona Jovita Judita
Kotryna Laura Lukrecija Mildutė Miglė Morta Odeta Palmira Ramunė Rimutė
Rožė Saulė Sigitė Silva Skirmantė Solveiga Unė Virgilija Vitalija Viltė
Žaneta Živilė
Agnieška Albina Angelina Arūnė Benita Bernardina Dalė Dana Danguolė Dominyka
Ernesta Eleonora Elžbieta Erika Felicija Gailė Gabrielė Genė Ilma Janė
Joana Juzefa Kazimiera Kornelija Laimutė Leokadija Liucija Liudvika Lidija Lolita
Marytė Marta Martyna Natalija Nerija Nida Nina Oksana Onutė Otilija
Paulė Pranė Raminta Rimgailė Rosita Rugilė Salomėja Sara Severija Sima
Sonata Svajonė Šarūnė Teofilė Terezija Toma Urtė Valdonė Vanesa Vera
Vilė Vilija Zina Žydrė
"""

LAST_RECALLED = """
Kazlauskas Petrauskas Jankauskas Stankevičius Vasiliauskas Žukauskas Butkus Paulauskas Urbonas Kavaliauskas
Sakalauskas Pocius Navickas Baranauskas Ramanauskas Mikalauskas Rimkus Gudaitis Šimkus Jonaitis
Morkūnas Adomaitis Savickas Balčiūnas Tamošiūnas Bagdonas Radzevičius Kairys Mockus Rutkauskas
Černiauskas Vaitkus Lukošius Matulevičius Girdzijauskas Čepulis Grigas Petkevičius Jakubauskas Šimkevičius
Zdanavičius Ivanauskas Jasiūnas Vaičiūnas Mažeika Žilinskas Kučinskas Venckus Sirvydas Kubilius
Norkus Jucevičius Kriščiūnas Dambrauskas Pranckevičius Stonkus Bernotas Šalkauskas Maksimavičius Budrys
Daugėla Aleksandravičius Gaidys Lapinskas Staškevičius Juška Ambrazevičius Barkauskas Mačiulis Rimša
Vaišvila Dapkus Kaminskas Bartkus Gedvilas Kasparavičius Raudonis Kriaučiūnas Gudelis Lekavičius
Šukys Samuolis Kvedaras Sinkevičius Misevičius Juškevičius Martinkus Stasiulis Marcinkevičius Valaitis
Šidlauskas Sabaliauskas Tamulis Bielskis Vyšniauskas Laurinaitis Karalius Jurevičius Rakauskas Žemaitis
=
Alekna Armonas Balsys Banys Baltrušaitis Bružas Butrimas Dobrovolskis Dvarionas Gaižauskas
Gedminas Giedraitis Glemža Grinius Jakimavičius Juodis Kalinauskas Kalvaitis Kasputis Kazakevičius
Kniukšta Kuncaitis Lileikis Mikulėnas Naujokas Paškevičius Petraitis Pivoriūnas Raškauskas Sadauskas
Stankus Šeškus Tamašauskas Vaitiekūnas Vilkas Vizgirda Ščerbinskas
Ivanov Kovalenko Kowalski Kozlov Michailov Orlov Petrov Smirnov Sokolov
"""


def published(text: str) -> list[tuple[str, int]]:
    """The rows (name, count) of a register table, in the order of the rank it gives them."""
    rows = sorted(json.loads(text)["_data"], key=lambda r: r["nr"])
    return [(r["vardas"].title(), int(r["viso_vardas"])) for r in rows]


def ranked(top: Sequence[tuple[str, int]], recalled: str, what: str) -> list[tuple[str, int]]:
    """The rows (name, count) of a list: the published names as published, the recalled names in order, then the rarer ones, which share a count."""
    ordered, rarer = recalled_names.split(what, recalled, [name for name, _ in top])
    smallest = min(n for _, n in top)
    rows = list(top)
    for rank, name in enumerate(ordered, start=len(top) + 1):
        rows.append((name, smallest * len(top) // rank))
    last = smallest * len(top) // (len(top) + len(ordered))
    return rows + [(name, last) for name in rarer]


FEMININE = (("ius", "iūtė"), ("us", "utė"), ("ys", "ytė"), ("is", "ytė"), ("as", "aitė"), ("a", "aitė"))


def feminine(name: str) -> list[str]:
    """The feminine forms of a masculine last name, the unmarried one first and the married one second; a Russian or Polish one takes its own language's ending, and any other is one form."""
    for ending, unmarried in FEMININE:
        if name.endswith(ending):
            stem = name[: -len(ending)]
            return [stem + unmarried, stem + "ienė"]
    if name.endswith(("ov", "ev")):
        return [name + "a"]
    if name.endswith("ski"):
        return [name[:-1] + "a"]
    return [name]


def lasts() -> list[dict[str, object]]:
    """The rows (name, sex, count) of the last names: each recalled name under m, and each feminine form of it under f."""
    rows = recalled_names.weighted("last names", LAST_RECALLED, TOP)
    return [{"name": form, "sex": sex, "count": n} for name, n in rows for sex, forms in (("f", feminine(name)), ("m", [name])) for form in forms]


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--men", default=REGISTER.format(table="vyru_vardai/VyroVardas"))
    p.add_argument("--women", default=REGISTER.format(table="moteru_vardai/MotersVardas"))
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    men = published(source.fetch(a.men, a.cache, "rc-vyru-vardai.json", magic=b"{").decode("utf-8"))
    women = published(source.fetch(a.women, a.cache, "rc-moteru-vardai.json", magic=b"{").decode("utf-8"))
    first = [{"name": n, "sex": "m", "count": c} for n, c in ranked(men, MEN_RECALLED, "men's names")]
    first += [{"name": n, "sex": "f", "count": c} for n, c in ranked(women, WOMEN_RECALLED, "women's names")]
    out = Path(a.out)
    out.mkdir(parents=True, exist_ok=True)
    tsv.write(out / "first-name.tsv", ["name", "sex", "count"], sorted(first, key=lambda r: (str(r["name"]), str(r["sex"]))))
    tsv.write(out / "last-name.tsv", ["name", "sex", "count"], sorted(lasts(), key=lambda r: (str(r["name"]), str(r["sex"]))))


if __name__ == "__main__":
    main()
