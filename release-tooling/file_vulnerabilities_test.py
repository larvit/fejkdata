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


def report(*findings, osvs=()):
	messages = [{"config": {"go_version": "go1.27.1", "scan_level": "module"}}]
	messages += [{"osv": osv} for osv in osvs]
	messages += [
		{"finding": {"osv": osv, "fixed_version": fixed, "trace": [{"module": module, "version": version}]}}
		for osv, module, version, fixed in findings
	]
	return "\n".join(json.dumps(m, indent=2) for m in messages)


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


class FileItems(unittest.TestCase):
	def test_no_finding_files_nothing(self):
		self.assertEqual(file_vulnerabilities.file_items(report(osvs=OSVS), TODO), (TODO, 0))

	def test_files_one_security_item_per_module_at_the_top_of_the_next_release(self):
		todo, found = file_vulnerabilities.file_items(report(*STDLIB, osvs=OSVS), TODO)
		self.assertEqual(found, 2)
		self.assertIn("`Next ID = 8`", todo)
		rows = [line for line in todo.splitlines() if "| **" in line]
		self.assertEqual(
			rows[0],
			"| 7 | 0.2.0 | security | **Raise Go to 1.27.10, so the release binaries carry none of"
			" its known vulnerabilities.** | 1 | 1 | 1 | 8 | 8 | 1.1 | 29.1 |",
		)
		self.assertTrue(rows[1].startswith("| 5 |"))
		detail = todo.split("## Details\n\n", 1)[1]
		self.assertTrue(
			detail.startswith(
				"### 7. Raise Go to 1.27.10, so the release binaries carry none of its known vulnerabilities.\n\n"
				"govulncheck found these in Go 1.27.1:\n\n"
				"- GO-2026-0001 (CVE-2026-1): Cross-site scripting in html/template\n"
				"- GO-2026-0002 (CVE-2026-2): Denial of service in net/http\n\n"
			),
			detail,
		)
		self.assertIn("### 5. Add a thing.", detail)

	def test_a_module_requirement_gets_its_own_item(self):
		findings = STDLIB + (("GO-2026-0003", "golang.org/x/text", "v0.3.0", "v0.3.8"),)
		todo, found = file_vulnerabilities.file_items(report(*findings, osvs=OSVS), TODO)
		self.assertEqual(found, 3)
		self.assertIn("| 8 | 0.2.0 | security | **Raise `golang.org/x/text` to v0.3.8, so the module requires none of its known vulnerabilities.** |", todo)
		self.assertIn("- GO-2026-0003: Panic in x/text\n", todo)
		self.assertIn("`Next ID = 9`", todo)

	def test_a_vulnerability_already_planned_is_not_filed_again(self):
		filed, _ = file_vulnerabilities.file_items(report(*STDLIB, osvs=OSVS), TODO)
		self.assertEqual(file_vulnerabilities.file_items(report(*STDLIB, osvs=OSVS), filed), (filed, 2))

	def test_a_plan_with_no_items_is_refused(self):
		empty = TODO.split("| 5 |")[0].rstrip("\n") + "\n\n## Details\n"
		with self.assertRaises(SystemExit):
			file_vulnerabilities.file_items(report(*STDLIB, osvs=OSVS), empty)


class Main(unittest.TestCase):
	def setUp(self):
		cwd = os.getcwd()
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		os.chdir(tmp.name)
		self.addCleanup(os.chdir, cwd)
		with open("todo.md", "w", encoding="utf-8") as f:
			f.write(TODO)

	def run_main(self, text):
		with open("report.json", "w", encoding="utf-8") as f:
			f.write(text)
		out = io.StringIO()
		with redirect_stdout(out):
			code = file_vulnerabilities.main(["report.json", "todo.md"])
		with open("todo.md", encoding="utf-8") as f:
			return code, out.getvalue(), f.read()

	def test_a_clean_report_passes_and_leaves_the_plan(self):
		self.assertEqual(self.run_main(report(osvs=OSVS)), (0, "No known vulnerabilities.\n", TODO))

	def test_a_finding_fails_and_files_it(self):
		code, out, todo = self.run_main(report(*STDLIB, osvs=OSVS))
		self.assertEqual(code, 1)
		self.assertEqual(out, "2 known vulnerabilities; todo.md item 7 plans the fix.\n")
		self.assertIn("| 7 | 0.2.0 | security |", todo)

	def test_a_finding_already_planned_still_fails(self):
		self.run_main(report(*STDLIB, osvs=OSVS))
		code, out, _ = self.run_main(report(*STDLIB, osvs=OSVS))
		self.assertEqual(code, 1)
		self.assertEqual(out, "2 known vulnerabilities, already planned in todo.md.\n")


if __name__ == "__main__":
	unittest.main()
