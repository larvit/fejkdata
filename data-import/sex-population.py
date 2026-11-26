"""Rebuild data/sv_SE/sex.tsv, nb_NO/sex.tsv, da_DK/sex.tsv, fi_FI/sex.tsv, de_DE/sex.tsv, et_EE/sex.tsv, lt_LT/sex.tsv, lv_LV/sex.tsv, fr_FR/sex.tsv and en_US/sex.tsv: the people of each sex.

    data-import/sex-population.py [--cache DIR] [--out-dir DIR]

- sv_SE: SCB's table BefolkningR1860N, the population at the end of 2024 (CC0 1.0, "Källa: SCB")
- nb_NO: Statistics Norway's StatBank table 07459, the population on 1 January 2026, all ages summed
  (CC BY 4.0, "Kilde: Statistisk sentralbyrå")
- da_DK: Statistics Denmark's table FOLK1A, the population on 1 January 2026 (CC BY 4.0, "Kilde: Danmarks Statistik")
- fi_FI: Statistics Finland's table 11rb, the population on 31 December 2025 (CC BY 4.0, "Source: Statistics Finland")
- de_DE: Eurostat's table demo_pjan, the population of Germany on 1 January 2025, from the Federal Statistical
  Office (CC BY 4.0, "Source: Eurostat")
- et_EE: Eurostat's table demo_pjan, the population of Estonia on 1 January 2025, from Statistics Estonia
  (CC BY 4.0, "Source: Eurostat")
- lt_LT: Eurostat's table demo_pjan, the population of Lithuania on 1 January 2025, from Statistics Lithuania
  (CC BY 4.0, "Source: Eurostat")
- lv_LV: Eurostat's table demo_pjan, the population of Latvia on 1 January 2025, from the Central Statistical Bureau
  of Latvia (CC BY 4.0, "Source: Eurostat")
- fr_FR: Eurostat's table demo_pjan, the population of France on 1 January 2025, from INSEE (CC BY 4.0,
  "Source: Eurostat"); Eurostat marks the figure provisional
- en_US: the Census Bureau's national population estimates by age and sex, the resident population on
  1 July 2024 (vintage 2024, public domain)

The names are ours: `kvinna` and `man`, `kvinne` and `mann`, `kvinde` and `mand`, `nainen` and `mies`, `Frau` and `Mann`, `naine` and `mees`, `moteris` and `vyras`, `sieviete` and `vīrietis`, `femme` and `homme`, `female` and `male`.
"""
import argparse
import csv
import io
import json
import sys
from pathlib import Path
from typing import Any

import source
import tsv

SCB = "https://api.scb.se/OV0104/v1/doris/en/ssd/BE/BE0101/BE0101A/BefolkningR1860N"
SSB = "https://data.ssb.no/api/v0/en/table/07459"
DST = "https://api.statbank.dk/v1/data"
EUROSTAT = "https://ec.europa.eu/eurostat/api/dissemination/statistics/1.0/data/demo_pjan?format=JSON&lang=EN&geo={geo}&age=TOTAL&sex=M&sex=F&time=2025"
STATFIN = "https://pxdata.stat.fi/PxWeb/api/v1/en/StatFin/vaerak/11rb.px"
CENSUS = "https://www2.census.gov/programs-surveys/popest/datasets/2020-2024/national/asrh/nc-est2024-agesex-res.csv"
ROOT = Path(__file__).resolve().parent.parent / "data"
CACHE = Path(__file__).resolve().parent / "cache"
JSON = {"Content-Type": "application/json"}
NAMES = {"sv_SE": ("kvinna", "man"), "nb_NO": ("kvinne", "mann"), "da_DK": ("kvinde", "mand"), "fi_FI": ("nainen", "mies"), "de_DE": ("Frau", "Mann"), "et_EE": ("naine", "mees"), "lt_LT": ("moteris", "vyras"), "lv_LV": ("sieviete", "vīrietis"), "fr_FR": ("femme", "homme"), "en_US": ("female", "male")}


def post(url: str, cache: str, name: str, query: object) -> Any:
	return json.loads(source.fetch(url, cache, name, magic=b"{", data=json.dumps(query).encode(), headers=JSON))


def sweden(cache: str) -> tuple[int, int]:
	"""(women, men) at the end of 2024."""
	query = {"query": [{"code": "Alder", "selection": {"filter": "item", "values": ["tot"]}}, {"code": "Kon", "selection": {"filter": "item", "values": ["1", "2"]}}, {"code": "Tid", "selection": {"filter": "item", "values": ["2024"]}}], "response": {"format": "json"}}
	meta = json.loads(source.fetch(SCB, cache, "scb-BefolkningR1860N-meta.json", magic=b"{"))
	kon = next(v for v in meta["variables"] if v["code"] == "Kon")
	if dict(zip(kon["values"], kon["valueTexts"])) != {"1": "men", "2": "women"}:
		sys.exit(f"SCB codes sex {dict(zip(kon['values'], kon['valueTexts']))}, not 1 for men and 2 for women")
	counts = {row["key"][1]: int(row["values"][0]) for row in post(SCB, cache, "scb-BefolkningR1860N-2024.json", query)["data"]}
	return counts["2"], counts["1"]


def norway(cache: str) -> tuple[int, int]:
	"""(women, men) on 1 January 2026."""
	query = {"query": [{"code": "Region", "selection": {"filter": "item", "values": ["0"]}}, {"code": "Kjonn", "selection": {"filter": "item", "values": ["1", "2"]}}, {"code": "Alder", "selection": {"filter": "all", "values": ["*"]}}, {"code": "ContentsCode", "selection": {"filter": "item", "values": ["Personer1"]}}, {"code": "Tid", "selection": {"filter": "item", "values": ["2026"]}}], "response": {"format": "json-stat2"}}
	cube = post(SSB, cache, "ssb-07459-2026.json", query)
	if cube["id"] != ["Region", "Kjonn", "Alder", "ContentsCode", "Tid"] or cube["size"][0] != 1 or cube["size"][3:] != [1, 1]:
		sys.exit(f"table 07459 has dimensions {cube['id']} of sizes {cube['size']}, not one region, sex, age, one count, one year")
	sexes = cube["dimension"]["Kjonn"]["category"]
	codes = sexes["index"] if isinstance(sexes["index"], list) else sorted(sexes["index"], key=sexes["index"].get)
	ages = cube["size"][2]
	if len(cube["value"]) != len(codes) * ages or None in cube["value"]:
		sys.exit("table 07459 holds a missing count or not one count for each sex and age")
	by = {code: sum(cube["value"][i * ages:(i + 1) * ages]) for i, code in enumerate(codes)}
	if {code: sexes["label"][code] for code in codes} != {"1": "Males", "2": "Females"}:
		sys.exit(f"table 07459 labels its sexes {sexes['label']}, not 1 for Males and 2 for Females")
	return by["2"], by["1"]


def denmark(cache: str) -> tuple[int, int]:
	"""(women, men) on 1 January 2026."""
	query = {"table": "FOLK1A", "format": "CSV", "delimiter": "Semicolon", "variables": [{"code": "KØN", "values": ["1", "2"]}, {"code": "TID", "values": ["2026K1"]}]}
	text = source.fetch(DST, cache, "dst-FOLK1A-2026K1.csv", magic=b"\xef\xbb\xbfK", data=json.dumps(query).encode(), headers=JSON).decode("utf-8-sig")
	rows = list(csv.reader(io.StringIO(text), delimiter=";"))
	if rows[0] != ["KØN", "TID", "INDHOLD"] or sorted(r[0] for r in rows[1:]) != ["Kvinder", "Mænd"] or {r[1] for r in rows[1:]} != {"2026K1"}:
		sys.exit(f"table FOLK1A answers {rows}, not the counts of Mænd and Kvinder in 2026K1")
	counts = {r[0]: int(r[2]) for r in rows[1:]}
	return counts["Kvinder"], counts["Mænd"]


def eurostat(cache: str, geo: str) -> tuple[int, int]:
	"""(women, men) of a country on 1 January 2025."""
	cube = json.loads(source.fetch(EUROSTAT.format(geo=geo), cache, f"eurostat-demo_pjan-{geo}-2025.json", magic=b"{"))
	sexes = cube["dimension"]["sex"]["category"]
	if cube["id"] != ["freq", "unit", "age", "sex", "geo", "time"] or cube["size"] != [1, 1, 1, 2, 1, 1] or sexes["label"] != {"M": "Males", "F": "Females"}:
		sys.exit(f"table demo_pjan answers {cube['id']} of sizes {cube['size']} with sexes {sexes['label']}, not the Males and Females of {geo} in 2025")
	counts = {code: cube["value"][str(i)] for code, i in sexes["index"].items()}
	return counts["F"], counts["M"]


def germany(cache: str) -> tuple[int, int]:
	return eurostat(cache, "DE")


def estonia(cache: str) -> tuple[int, int]:
	return eurostat(cache, "EE")


def lithuania(cache: str) -> tuple[int, int]:
	return eurostat(cache, "LT")


def latvia(cache: str) -> tuple[int, int]:
	return eurostat(cache, "LV")


def france(cache: str) -> tuple[int, int]:
	return eurostat(cache, "FR")


def finland(cache: str) -> tuple[int, int]:
	"""(women, men) on 31 December 2025."""
	query = {"query": [{"code": "timeperiod_y", "selection": {"filter": "item", "values": ["2025"]}}, {"code": "sukupuoli_9_20180101", "selection": {"filter": "item", "values": ["1", "2"]}}, {"code": "contentscode", "selection": {"filter": "item", "values": ["vaerak-vaesto"]}}], "response": {"format": "json-stat2"}}
	cube = post(STATFIN, cache, "statfin-11rb-2025.json", query)
	sexes = cube["dimension"]["sukupuoli_9_20180101"]["category"]
	if cube["id"] != ["timeperiod_y", "sukupuoli_9_20180101", "contentscode"] or cube["size"] != [1, 2, 1] or sexes["label"] != {"1": "Males", "2": "Females"} or list(sexes["index"]) != ["1", "2"]:
		sys.exit(f"table 11rb answers {cube['id']} of sizes {cube['size']} with sexes {sexes['label']}, not the Males and Females of 2025")
	men, women = cube["value"]
	return women, men


def america(cache: str) -> tuple[int, int]:
	"""(women, men) on 1 July 2024."""
	text = source.fetch(CENSUS, cache, "census-nc-est2024-agesex-res.csv").decode("utf-8")
	counts = {r["SEX"]: int(r["POPESTIMATE2024"]) for r in csv.DictReader(io.StringIO(text)) if r["AGE"] == "999"}
	if counts["0"] != counts["1"] + counts["2"]:
		sys.exit(f"the Census Bureau's total {counts['0']} is not its men {counts['1']} and women {counts['2']}")
	return counts["2"], counts["1"]


def main() -> None:
	p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
	p.add_argument("--cache", default=str(CACHE))
	p.add_argument("--out-dir", default=str(ROOT))
	a = p.parse_args()
	for locale, read in (("sv_SE", sweden), ("nb_NO", norway), ("da_DK", denmark), ("fi_FI", finland), ("de_DE", germany), ("et_EE", estonia), ("lt_LT", lithuania), ("lv_LV", latvia), ("fr_FR", france), ("en_US", america)):
		(women, men), (woman, man) = read(a.cache), NAMES[locale]
		tsv.write(Path(a.out_dir) / locale / "sex.tsv", ["code", "name", "count"], [{"code": "f", "name": woman, "count": women}, {"code": "m", "name": man, "count": men}])


if __name__ == "__main__":
	main()
