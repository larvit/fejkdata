"""Rebuild data/fr_FR/company.json, company-coined.json and company-word.{json,tsv} from INSEE's Sirene.

    data-import/company-fr.py [--register URL_OR_FILE] [--cache DIR] [--out DIR]

The script reads Sirene's StockUniteLegale (a 1 GB zip) for counts and names, and writes no row of it. It also
reads the shipped fr_FR first and last names, so run names-fr.py first. It counts the active legal units, those
whose etatAdministratifUniteLegale is A. company.json weighs a legal form by its units, and for each form but EI
writes the shapes of its units' names. The forms are EI (category 1000), SARL (5499), SAS (5710), SNC (5202) and
SA (5599 and 5699). An EI is a person: its name is `{.first-name} {.last-name}`, as the register holds it in nom
and prenom, so no shape is counted. The units of every other category, such as the SCI and the associations,
aren't drawn.

Sirene writes a name in capitals without accents, and mostly without the legal form. A trailing legal form, SARL,
EURL, SAS, SASU, SNC, SA or S.A., is dropped from a name, and a legal word anywhere in it, such as SOCIETE or SCI,
is left out of its shape. A name counts for the form of its category.

A shape is the kind of each word of a name: a last name, a first name, a number, a connecting word such as `et`,
`de` or `la`, a trade word, or any other, which draws a coined word. A last or first name is one in the shipped
fr_FR tables, compared without accents. A shape ships when it is at least company_names.SHARE_MIN of the form's
names, and its first word is capitalised.

company-word.tsv holds the TRADE_MAX commonest trade words, each with its count and its spelling: the word
capitalised, with the accents SPELLINGS gives it, as Sirene holds none. The rule is company_names.py's, over the
names of the forms SARL, SAS, SNC and SA. SPELLINGS is written by hand, from the Dictionnaire de l'Académie
française, for the words among those, and among the roots, that take an accent or are an acronym.

company-coined.tsv holds the roots, each with the number of names that carry it: the words of ROOT_LETTERS
letters that begin ROOT_MIN or more names and end at least ROOT_ENDS_MIN but under a third of the names they are
in, so `Chez`, which only opens a name, stays out. None is a name, a trade word or a legal form's abbreviation, and
each is carried by FORM_MIN or more names in at least two of the forms, so a word one form uses stays out. A root
is a word of the register's names, so a name drawn with it may be a real company's.
"""
import argparse
import csv
import io
import re
import unicodedata
import zipfile
from collections import Counter
from collections.abc import Collection, Iterable, Mapping
from pathlib import Path

import source
import tsv
from company_names import (
    FORM_MIN,
    Usage,
    Vocabulary,
    column,
    common,
    legal_item,
    quoted,
    split,
    trade_words,
    weighted,
    word_usage,
)

REGISTER = "https://static.data.gouv.fr/resources/base-sirene-des-entreprises-et-de-leurs-etablissements-siren-siret/20261103-091933/stock-stockunitelegale-csv.zip"
ROOT = Path(__file__).resolve().parent.parent / "data"
OUT = ROOT / "fr_FR"
CACHE = Path(__file__).resolve().parent / "cache"
CATEGORY_FORM = {"1000": "EI", "5499": "SARL", "5710": "SAS", "5202": "SNC", "5599": "SA", "5699": "SA"}
FORMS = ("EI", "SARL", "SAS", "SNC", "SA")
NAMED_FORMS = FORMS[1:]
ROOT_MIN = 60
ROOT_ENDS_MIN = 0.05
ROOT_LETTERS = range(4, 7)
TRADE_MAX = 120
CONNECTING_WORDS = ("et", "de", "du", "des", "la", "le", "les", "au", "aux", "en", "sur", "sous", "pour", "par", "&")
FUNCTION_WORDS = {*CONNECTING_WORDS, "the", "and", "of", "for"}
# The abbreviations and words of the legal forms in INSEE's catégories juridiques, which are no part of a name.
LEGAL_WORDS = {
    "anonyme", "associe", "actions", "cuma", "earl", "eurl", "gaec", "geie", "gie", "limitee", "responsabilite", "sarl", "sarlu", "sas", "sasp", "sasu",
    "sca", "scea", "scic", "sci", "sccv", "scop", "scp", "scs", "selafa", "selarl", "selas", "selca", "sem", "semop", "sica", "simplifiee", "snc",
    "societe", "spfpl", "unique",
}
LEGAL_SUFFIX = re.compile(r"[\s,]+(?:S\.?A\.?R\.?L\.?|E\.?U\.?R\.?L\.?|S\.?A\.?S\.?U?\.?|S\.?N\.?C\.?|S\.?A\.?)\s*$")
VOCABULARY = Vocabulary(re.compile(r"^[a-z]{3,}$"), frozenset(FUNCTION_WORDS), frozenset(LEGAL_WORDS), frozenset(NAMED_FORMS), 0)
# The accented spelling of a trade word or a root, which Sirene writes without accents.
SPELLINGS = {
    "amenagement": "Aménagement", "associes": "Associés", "batiment": "Bâtiment", "beaute": "Beauté", "cafe": "Café", "creation": "Création",
    "deco": "Déco", "decoration": "Décoration", "developpement": "Développement", "editions": "Éditions", "electricite": "Électricité",
    "energie": "Énergie", "energies": "Énergies", "freres": "Frères", "ingenierie": "Ingénierie", "maconnerie": "Maçonnerie", "media": "Média",
    "negoce": "Négoce", "privee": "Privée", "renovation": "Rénovation", "sante": "Santé", "securite": "Sécurité",
    "acces": "Accès", "aero": "Aéro", "ariege": "Ariège", "boite": "Boîte", "cafes": "Cafés", "cine": "Ciné", "creche": "Crèche",
    "credit": "Crédit", "cles": "Clés", "declic": "Déclic", "defi": "Défi", "depot": "Dépôt", "eclat": "Éclat", "ecole": "École",
    "ecrin": "Écrin", "elysee": "Élysée", "eole": "Éole", "equip": "Équip", "etude": "Étude", "etudes": "Études", "eveil": "Éveil",
    "fleche": "Flèche", "genie": "Génie", "gite": "Gîte", "gites": "Gîtes", "hotel": "Hôtel", "idee": "Idée", "idees": "Idées", "meca": "Méca",
    "medi": "Médi", "medic": "Médic", "nego": "Négo", "pieces": "Pièces", "pole": "Pôle", "pret": "Prêt", "regie": "Régie", "reseau": "Réseau",
    "rhone": "Rhône", "tele": "Télé", "tete": "Tête", "unite": "Unité",
}
# A name's kinds of word in order, each a key of SLOTS.
type Shape = tuple[str, ...]

SLOTS = {"X": "{.company-coined}", "T": "{.company-word}", "S": "{.last-name}", "F": "{.first-name}", "N": "{int(1,30)}", **{w: w for w in CONNECTING_WORDS}}


def fold(word: str) -> str:
    """The word in lower case without accents."""
    return unicodedata.normalize("NFD", word.lower()).encode("ascii", "ignore").decode("ascii")


def register(data: bytes) -> tuple[Counter[str], list[tuple[str, str]]]:
    """(units per form, (form, name) of every active unit of NAMED_FORMS)."""
    counts: Counter[str] = Counter()
    names: list[tuple[str, str]] = []
    with zipfile.ZipFile(io.BytesIO(data)) as z, z.open(z.namelist()[0]) as f:
        for row in csv.DictReader(io.TextIOWrapper(f, encoding="utf-8", newline="")):
            form = CATEGORY_FORM.get(row["categorieJuridiqueUniteLegale"])
            if form and row["etatAdministratifUniteLegale"] == "A":
                counts[form] += 1
                if form in NAMED_FORMS and row["denominationUniteLegale"]:
                    names.append((form, LEGAL_SUFFIX.sub("", row["denominationUniteLegale"])))
    return counts, names


def kind(word: str, surnames: Collection[str], firsts: Collection[str]) -> str | None:
    """S or F for a last name or a first name; else None. A word that is both is a last name."""
    w = fold(word)
    if w in surnames:
        return "S"
    return "F" if w in firsts else None


def roots(usage: Mapping[str, Usage], trade: Collection[str]) -> dict[str, int]:
    """The roots of company-coined.tsv, each capitalised with its count, by the rule in the module docstring."""
    return {
        SPELLINGS.get(w, w.capitalize()): u.names
        for w, u in sorted(usage.items())
        if w not in trade and len(w) in ROOT_LETTERS and u.begins >= ROOT_MIN and ROOT_ENDS_MIN * u.names <= u.ends and 3 * u.ends < u.names
        and sum(u.forms[f] >= FORM_MIN for f in VOCABULARY.word_forms) >= 2
    }


def shape_of(words: list[str], surnames: Collection[str], firsts: Collection[str], trade: Collection[str]) -> Shape:
    """The kinds of the words, as SLOTS keys."""
    shape: list[str] = []
    for word in words:
        low = fold(word)
        if low.isdigit():
            shape.append("N")
        elif low in CONNECTING_WORDS:
            shape.append(low)
        else:
            shape.append(kind(word, surnames, firsts) or ("T" if low in trade else "X"))
    return tuple(shape)


def count_shapes(names: Iterable[tuple[str, str]], surnames: Collection[str], firsts: Collection[str], trade: Collection[str]) -> dict[str, Counter[Shape]]:
    """Per form: the shapes counted over its names."""
    shapes: dict[str, Counter[Shape]] = {f: Counter() for f in NAMED_FORMS}
    for form, name in names:
        words = [w for w in split(name) if fold(w).replace(".", "") not in LEGAL_WORDS]
        if words:
            shapes[form][shape_of(words, surnames, firsts, trade)] += 1
    return shapes


def commonest(trade: Mapping[str, tuple[str, int]], limit: int) -> dict[str, tuple[str, int]]:
    """The limit commonest trade words, each with its spelling from SPELLINGS."""
    top = sorted(trade.items(), key=lambda wsc: (-wsc[1][1], wsc[0]))[:limit]
    return {w: (SPELLINGS.get(w, s), c) for w, (s, c) in top}


def slots(shape: Shape) -> str:
    """The format of a shape: its words' slots, the first one capitalised."""
    text = " ".join(SLOTS[k] for k in shape)
    return text[0].upper() + text[1:]


def form_block(form: str, shapes: Mapping[Shape, int]) -> str:
    """One form's object: its name by shape, and the abbreviation after it."""
    items = ",\n".join("      " + weighted(slots(shape), c) for shape, c in common(shapes))
    return f'  "{form.lower()}": {{\n    "format": {quoted("{name} " + form)},\n    "name": [\n{items}\n    ]\n  }}'


def ei_block() -> str:
    return '  "ei": {\n    "format": "{name}",\n    "name": "{.first-name} {.last-name}"\n  }'


def company_json(registered: Mapping[str, int], shapes: Mapping[str, Mapping[Shape, int]]) -> str:
    forms = [legal_item("{.company." + f.lower() + "}", "" if f == "EI" else f, registered[f]) for f in sorted(FORMS, key=lambda f: (-registered[f], f))]
    blocks = [ei_block()] + [form_block(f, shapes[f]) for f in NAMED_FORMS]
    return '{\n  "format": "{registered as r}{r}",\n  "legal": "{r.legal}",\n  "registered": [\n    ' + ",\n    ".join(forms) + "\n  ],\n" + ",\n".join(blocks) + "\n}\n"


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--register", default=REGISTER)
    p.add_argument("--cache", default=str(CACHE))
    p.add_argument("--out", default=str(OUT))
    a = p.parse_args()
    registered, names = register(source.fetch(a.register, a.cache, "sirene-stockunitelegale.zip", magic=b"PK"))
    surnames = {fold(w) for w in column(ROOT / "fr_FR/last-name.tsv")}
    firsts = {fold(w) for w in column(ROOT / "fr_FR/first-name.tsv")}

    def named(w: str) -> bool:
        return kind(w, surnames, firsts) is not None

    usage = word_usage(names, named, VOCABULARY)
    trade = commonest(trade_words(usage, VOCABULARY), TRADE_MAX)
    out = Path(a.out)
    tsv.write(out / "company-word.tsv", ["word", "count"], [{"word": s, "count": c} for _, (s, c) in sorted(trade.items())])
    tsv.write(out / "company-coined.tsv", ["word", "count"], [{"word": w, "count": c} for w, c in roots(usage, trade).items()])
    for table in ("company-word", "company-coined"):
        (out / f"{table}.json").write_text('{ "format": "{word}", "rows": "' + table + '.tsv", "weight": "count" }\n', encoding="utf-8")
    (out / "company.json").write_text(company_json(registered, count_shapes(names, surnames, firsts, trade)), encoding="utf-8")


if __name__ == "__main__":
    main()
