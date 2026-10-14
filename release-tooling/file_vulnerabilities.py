#!/usr/bin/env python3
"""File each module's known vulnerabilities as one security item atop todo.md's next release.

Usage: file_vulnerabilities.py <govulncheck -format json report> <todo.md>
Exits 1 while any vulnerability is known, filed now or before, and 0 on a clean report.

The item is written the way ~/.claude/bin/rank_todo.py writes one, which CI cannot run:
todo.md's `Next ID = N` gives its ID, the release of the plan's top row (the next release)
gives its release, and its Score is todo.md's `Score = …` formula over the scores below and
the weight todo.md's goal table gives goal 1.1. The test that files into the repository's
own todo.md fails the gate when the plan's shape drifts from what this script reads.

A security item scores R, S and C 1, since the fix is a version bump, and A and G 8, since a
known vulnerability in the binaries works against README goal 1.1: installing fejkdata the
way one installs other tools, which image scanners then flag.
"""

import ast
import json
import re
import sys
from dataclasses import dataclass, field

ROW = re.compile(r"^\| (\d+) \| ([^|]+?) \| [^|]*\| \*\*", re.M)
WEIGHT = re.compile(r"^\| ([\d.]+) \| ([\d.]+) \|$", re.M)
NEXT_ID = re.compile(r"`Next ID = (\d+)`")
SCORES = {"R": 1, "S": 1, "C": 1, "A": 8, "G": 8}
GOAL = "1.1"


@dataclass
class Module:
	path: str
	found_in: str
	fixed: list[str] = field(default_factory=list)
	unfixed: bool = False
	osvs: dict[str, dict] = field(default_factory=dict)

	@property
	def name(self) -> str:
		return "Go" if self.path == "stdlib" else f"`{self.path}`"

	def version(self, version: str) -> str:
		return version.removeprefix("v") if self.path == "stdlib" else version

	@property
	def carried(self) -> str:
		return f"{self.name} {self.version(self.found_in)}"


def messages(report: str):
	decoder, at = json.JSONDecoder(), 0
	while at < len(report):
		if report[at].isspace():
			at += 1
			continue
		message, at = decoder.raw_decode(report, at)
		yield message


def vulnerable_modules(report: str) -> list[Module]:
	"""Each module govulncheck found a vulnerability in, Go (stdlib) first, then by path."""
	osvs, modules = {}, {}
	for message in messages(report):
		if "osv" in message:
			osv = message["osv"]
			osvs[osv["id"]] = {"id": osv["id"], "aliases": osv.get("aliases", []), "summary": osv.get("summary", "")}
		elif "finding" in message:
			finding = message["finding"]
			frame = finding["trace"][0]
			module = modules.setdefault(frame["module"], Module(frame["module"], frame["version"]))
			module.osvs[finding["osv"]] = osvs[finding["osv"]]
			if "fixed_version" in finding:
				module.fixed.append(finding["fixed_version"])
			else:
				module.unfixed = True
	return sorted(modules.values(), key=lambda m: (m.path != "stdlib", m.path))


def version_key(version: str):
	return [int(part) for part in re.findall(r"\d+", version)]


def evaluate(node, values):
	"""The value of todo.md's score formula: numbers, letters, unary minus and + - * /."""
	match node:
		case ast.Constant(value=value):
			return value
		case ast.Name(id=name):
			return values[name]
		case ast.UnaryOp(op=ast.USub(), operand=operand):
			return -evaluate(operand, values)
		case ast.BinOp(left=left, op=ast.Add(), right=right):
			return evaluate(left, values) + evaluate(right, values)
		case ast.BinOp(left=left, op=ast.Sub(), right=right):
			return evaluate(left, values) - evaluate(right, values)
		case ast.BinOp(left=left, op=ast.Mult(), right=right):
			return evaluate(left, values) * evaluate(right, values)
		case ast.BinOp(left=left, op=ast.Div(), right=right):
			return evaluate(left, values) / evaluate(right, values)
	raise ValueError(f"todo.md's formula holds {ast.dump(node)}, which this script cannot evaluate")


def score(todo: str) -> str:
	formula = re.search(r"`Score = (.+?)`", todo).group(1)
	values = dict(SCORES, W=float(dict(WEIGHT.findall(todo))[GOAL]))
	return f"{evaluate(ast.parse(formula, mode='eval').body, values):.1f}"


def tagline(module: Module) -> str:
	if module.unfixed:
		return f"Ship no release while fejkdata carries {module.carried}, some of whose known vulnerabilities no version fixes yet."
	fixed = module.version(max(module.fixed, key=version_key))
	if module.path == "stdlib":
		return f"Raise Go to {fixed}, so the release binaries carry none of {module.carried}'s known vulnerabilities."
	return f"Raise {module.name} to {fixed}, so fejkdata requires no version of it with a known vulnerability."


def detail(module: Module) -> str:
	lines = []
	for osv in sorted(module.osvs.values(), key=lambda o: o["id"]):
		aliases = f" ({', '.join(osv['aliases'])})" if osv["aliases"] else ""
		summary = f": {osv['summary']}" if osv["summary"] else ""
		lines.append(f"- {osv['id']}{aliases}{summary}")
	return f"govulncheck found these in {module.carried}:\n\n" + "\n".join(lines) + "\n"


def file_items(modules: list[Module], todo: str) -> tuple[str, list[int]]:
	"""Return todo.md with an item for each module whose vulnerabilities it does not yet name, and the items' IDs."""
	filed = []
	for module in modules:
		if all(osv in todo for osv in module.osvs):
			continue
		first = ROW.search(todo)
		if first is None:
			raise ValueError("todo.md has no item to take the next release from")
		next_id = int(NEXT_ID.search(todo).group(1))
		line = tagline(module)
		scores = " | ".join(str(SCORES[k]) for k in "RSCAG")
		row = f"| {next_id} | {first.group(2)} | security | **{line}** | {scores} | {GOAL} | {score(todo)} |\n"
		todo = todo[: first.start()] + row + todo[first.start() :]
		todo = todo.replace(f"`Next ID = {next_id}`", f"`Next ID = {next_id + 1}`", 1)
		todo = todo.replace("## Details\n\n", f"## Details\n\n### {next_id}. {line}\n\n{detail(module)}\n", 1)
		filed.append(next_id)
	return todo, filed


def main(argv: list[str]) -> int:
	report_path, todo_path = argv
	with open(report_path, encoding="utf-8") as f:
		modules = vulnerable_modules(f.read())
	if not modules:
		print("No known vulnerabilities.")
		return 0
	count = sum(len(m.osvs) for m in modules)
	known = f"{count} known {'vulnerability' if count == 1 else 'vulnerabilities'}"
	with open(todo_path, encoding="utf-8") as f:
		todo, filed = file_items(modules, f.read())
	if not filed:
		print(f"{known}, already planned in todo.md.")
		return 1
	with open(todo_path, "w", encoding="utf-8") as f:
		f.write(todo)
	items = ", ".join(map(str, filed))
	print(f"{known}; todo.md {'item ' + items + ' plans' if len(filed) == 1 else 'items ' + items + ' plan'} the fix.")
	return 1


if __name__ == "__main__":
	sys.exit(main(sys.argv[1:]))
