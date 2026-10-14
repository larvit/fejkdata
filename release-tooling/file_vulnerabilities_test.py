import io
import json
import os
import tempfile
import unittest
from contextlib import redirect_stdout

import file_vulnerabilities

TODO = """# todo

## Scoring

`Score = -R - S/4 + 2*A + 2*G*W`

`Bar = 9`

`Next ID = 7`

| Goal | W |
|---|---|
| 1 | 1.00 |
| 1.1 | 0.90 |

## Items

| ID | Release | Exempt | Item | R | S | C | A | G | Goals | Score |
|---|---|---|---|---|---|---|---|---|---|---|
| 5 | 0.2.0 |  | **Add a thing.** | 2 | 2 | 2 | 7 | 8 | 1 | 26.5 |

## Details

### 5. Add a thing.

Why.
"""

OSVS = (
	{"id": "GO-2026-0002", "summary": "Denial of service in net/http", "aliases": ["CVE-2026-2"]},
	{"id": "GO-2026-0001", "summary": "Cross-site scripting in html/template", "aliases": ["CVE-2026-1"]},
	{"id": "GO-2026-0003", "summary": "Panic in x/text", "aliases": []},
)
STDLIB = (
	("GO-2026-0002", "stdlib", "v1.27.1", "v1.27.2"),
	("GO-2026-0001", "stdlib", "v1.27.1", "v1.27.10"),
	("GO-2026-0001", "stdlib", "v1.27.1", "v1.27.10"),
)
TEXT = (("GO-2026-0003", "golang.org/x/text", "v0.3.0", "v0.3.8"),)
GO_TAGLINE = "Raise Go to 1.27.10, so the release binaries carry none of Go 1.27.1's known vulnerabilities."


def report(*findings):
	messages = [{"config": {"go_version": "go1.27.1", "scan_level": "module"}}]
	messages += [{"osv": osv} for osv in OSVS]
	for osv, module, version, fixed in findings:
		finding = {"osv": osv, "trace": [{"module": module, "version": version}]}
		if fixed:
			finding["fixed_version"] = fixed
		messages.append({"finding": finding})
	return "\n".join(json.dumps(m, indent=2) for m in messages)


def file(todo, *findings):
	return file_vulnerabilities.file_items(file_vulnerabilities.vulnerable_modules(report(*findings)), todo)


def rows(todo):
	return [line for line in todo.splitlines() if "| **" in line]


class FileItems(unittest.TestCase):
	def test_no_finding_files_nothing(self):
		self.assertEqual(file(TODO), (TODO, []))

	def test_files_one_security_item_per_module_atop_the_next_release(self):
		todo, filed = file(TODO, *STDLIB)
		self.assertEqual(filed, [7])
		self.assertIn("`Next ID = 8`", todo)
		self.assertEqual(rows(todo)[0], f"| 7 | 0.2.0 | security | **{GO_TAGLINE}** | 1 | 1 | 1 | 8 | 8 | 1.1 | 29.1 |")
		self.assertTrue(rows(todo)[1].startswith("| 5 |"))
		self.assertTrue(
			todo.split("## Details\n\n", 1)[1].startswith(
				f"### 7. {GO_TAGLINE}\n\n"
				"govulncheck found these in Go 1.27.1:\n\n"
				"- GO-2026-0001 (CVE-2026-1): Cross-site scripting in html/template\n"
				"- GO-2026-0002 (CVE-2026-2): Denial of service in net/http\n\n"
				"### 5. Add a thing."
			)
		)

	def test_a_required_module_gets_its_own_item(self):
		todo, filed = file(TODO, *STDLIB, *TEXT)
		self.assertEqual(filed, [7, 8])
		self.assertIn(
			"| 8 | 0.2.0 | security | **Raise `golang.org/x/text` to v0.3.8, so fejkdata requires no version"
			" of it with a known vulnerability.** |",
			todo,
		)
		self.assertIn("govulncheck found these in `golang.org/x/text` v0.3.0:\n\n- GO-2026-0003: Panic in x/text\n", todo)
		self.assertIn("`Next ID = 9`", todo)

	def test_a_vulnerability_already_planned_is_not_filed_again(self):
		planned, _ = file(TODO, *STDLIB)
		self.assertEqual(file(planned, *STDLIB), (planned, []))

	def test_a_module_newly_vulnerable_is_filed_beside_one_already_planned(self):
		planned, _ = file(TODO, *STDLIB)
		todo, filed = file(planned, *STDLIB, *TEXT)
		self.assertEqual(filed, [8])
		self.assertIn("| 8 | 0.2.0 | security | **Raise `golang.org/x/text`", todo)

	def test_a_vulnerability_no_version_fixes_names_no_target(self):
		todo, _ = file(TODO, *STDLIB, ("GO-2026-0003", "stdlib", "v1.27.1", ""))
		self.assertIn(
			"| 7 | 0.2.0 | security | **Ship no release while fejkdata carries Go 1.27.1, some of whose known"
			" vulnerabilities no version fixes yet.** |",
			todo,
		)

	def test_a_plan_with_no_items_is_refused(self):
		empty = TODO.split("| 5 |")[0].rstrip("\n") + "\n\n## Details\n"
		with self.assertRaises(ValueError):
			file(empty, *STDLIB)

	def test_the_repositorys_plan_takes_an_item_in_its_own_shape(self):
		path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "todo.md")
		with open(path, encoding="utf-8") as f:
			plan = f.read()
		todo, filed = file(plan, *STDLIB)
		header = next(line for line in plan.splitlines() if line.startswith("| ID |"))
		row = rows(todo)[0]
		self.assertEqual(row.count("|"), header.count("|"))
		self.assertEqual(row.split(" | ")[1], rows(plan)[0].split(" | ")[1])
		self.assertRegex(row, rf"^\| {filed[0]} \| .* \| \d+\.\d \|$")
		self.assertIn(f"`Next ID = {filed[0] + 1}`", todo)


class Main(unittest.TestCase):
	def setUp(self):
		cwd = os.getcwd()
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		os.chdir(tmp.name)
		self.addCleanup(os.chdir, cwd)
		with open("todo.md", "w", encoding="utf-8") as f:
			f.write(TODO)

	def run_main(self, *findings):
		with open("report.json", "w", encoding="utf-8") as f:
			f.write(report(*findings))
		out = io.StringIO()
		with redirect_stdout(out):
			code = file_vulnerabilities.main(["report.json", "todo.md"])
		with open("todo.md", encoding="utf-8") as f:
			return code, out.getvalue(), f.read()

	def test_a_clean_report_passes_and_leaves_the_plan(self):
		self.assertEqual(self.run_main(), (0, "No known vulnerabilities.\n", TODO))

	def test_a_finding_fails_and_files_it(self):
		code, out, todo = self.run_main(*STDLIB)
		self.assertEqual((code, out), (1, "2 known vulnerabilities; todo.md item 7 plans the fix.\n"))
		self.assertIn("| 7 | 0.2.0 | security |", todo)

	def test_findings_in_two_modules_name_both_items(self):
		_, out, _ = self.run_main(*STDLIB, *TEXT)
		self.assertEqual(out, "3 known vulnerabilities; todo.md items 7, 8 plan the fix.\n")

	def test_one_finding_already_planned_still_fails(self):
		self.run_main(*TEXT)
		self.assertEqual(self.run_main(*TEXT)[:2], (1, "1 known vulnerability, already planned in todo.md.\n"))


if __name__ == "__main__":
	unittest.main()
