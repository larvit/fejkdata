#!/usr/bin/env python3
"""File each module's known vulnerabilities as one security item atop todo.md's next release.

Usage: file_vulnerabilities.py <govulncheck -format json report> <todo.md>
Exits 1 while any vulnerability is known, filed now or before, and 0 on a clean report.
"""

import ast
import json
import re
import sys

ROW = re.compile(r"^\| (\d+) \| ([^|]+?) \| [^|]*\| \*\*", re.M)
WEIGHT = re.compile(r"^\| ([\d.]+) \| ([\d.]+) \|$", re.M)
SCORES = {"R": 1, "S": 1, "C": 1, "A": 8, "G": 8}
GOAL = "1.1"


def messages(report: str):
	decoder, at = json.JSONDecoder(), 0
	while at < len(report):
		if report[at].isspace():
			at += 1
			continue
		message, at = decoder.raw_decode(report, at)
		yield message


def version_key(version: str):
	return [int(part) for part in re.findall(r"\d+", version)]


def score(todo: str) -> str:
	formula = re.search(r"`Score = (.+?)`", todo).group(1)
	weights = dict(WEIGHT.findall(todo))
	values = dict(SCORES, W=float(weights[GOAL]))
	return f"{evaluate(ast.parse(formula, mode='eval').body, values):.1f}"


def evaluate(node, values):
	match node:
		case ast.Constant(value=value):
			return value
		case ast.Name(id=name):
			return values[name]
		case ast.UnaryOp(op=ast.USub(), operand=operand):
			return -evaluate(operand, values)
		case ast.BinOp(left=left, op=op, right=right):
			a, b = evaluate(left, values), evaluate(right, values)
			return {ast.Add: a + b, ast.Sub: a - b, ast.Mult: a * b, ast.Div: a / b if b else 0}[type(op)]
	raise SystemExit(f"todo.md's formula holds {ast.dump(node)}, which this script cannot evaluate")


def item(module: str, found_in: str, fixed: str, osvs: list[dict]) -> tuple[str, str]:
	if module == "stdlib":
		tagline = f"Raise Go to {fixed.lstrip('v')}, so the release binaries carry none of its known vulnerabilities."
		where = f"Go {found_in.lstrip('v')}"
	else:
		tagline = f"Raise `{module}` to {fixed}, so the module requires none of its known vulnerabilities."
		where = f"`{module}` {found_in}"
	lines = []
	for osv in sorted(osvs, key=lambda o: o["id"]):
		aliases = f" ({', '.join(osv['aliases'])})" if osv.get("aliases") else ""
		lines.append(f"- {osv['id']}{aliases}: {osv.get('summary', '')}")
	return tagline, f"govulncheck found these in {where}:\n\n" + "\n".join(lines) + "\n"


def file_items(report: str, todo: str) -> tuple[str, int]:
	"""Return todo.md with an item for each module whose vulnerabilities it does not yet name, and how many are known."""
	osvs, modules = {}, {}
	for message in messages(report):
		if "osv" in message:
			osvs[message["osv"]["id"]] = message["osv"]
		elif "finding" in message:
			finding = message["finding"]
			module, found_in = finding["trace"][0]["module"], finding["trace"][0].get("version", "")
			entry = modules.setdefault(module, {"found_in": found_in, "fixed": "", "ids": set()})
			entry["ids"].add(finding["osv"])
			if version_key(finding.get("fixed_version", "")) > version_key(entry["fixed"]):
				entry["fixed"] = finding["fixed_version"]
	known = sum(len(entry["ids"]) for entry in modules.values())
	for module in sorted(modules, key=lambda m: (m != "stdlib", m)):
		entry = modules[module]
		if all(osv in todo for osv in entry["ids"]):
			continue
		first = ROW.search(todo)
		if first is None:
			raise SystemExit("todo.md has no item to take the next release from")
		next_id = int(re.search(r"`Next ID = (\d+)`", todo).group(1))
		tagline, detail = item(module, entry["found_in"], entry["fixed"], [osvs.get(i, {"id": i}) for i in entry["ids"]])
		scores = " | ".join(str(SCORES[k]) for k in "RSCAG")
		row = f"| {next_id} | {first.group(2)} | security | **{tagline}** | {scores} | {GOAL} | {score(todo)} |\n"
		todo = todo[: first.start()] + row + todo[first.start() :]
		todo = todo.replace(f"`Next ID = {next_id}`", f"`Next ID = {next_id + 1}`", 1)
		todo = todo.replace("## Details\n\n", f"## Details\n\n### {next_id}. {tagline}\n\n{detail}\n", 1)
	return todo, known


def main(argv: list[str]) -> int:
	report_path, todo_path = argv
	with open(report_path, encoding="utf-8") as f:
		report = f.read()
	with open(todo_path, encoding="utf-8") as f:
		todo = f.read()
	filed, known = file_items(report, todo)
	if not known:
		print("No known vulnerabilities.")
		return 0
	if filed == todo:
		print(f"{known} known vulnerabilities, already planned in todo.md.")
		return 1
	with open(todo_path, "w", encoding="utf-8") as f:
		f.write(filed)
	next_id = re.compile(r"`Next ID = (\d+)`")
	ids = range(int(next_id.search(todo).group(1)), int(next_id.search(filed).group(1)))
	print(f"{known} known vulnerabilities; todo.md item {', '.join(map(str, ids))} plans the fix.")
	return 1


if __name__ == "__main__":
	sys.exit(main(sys.argv[1:]))
