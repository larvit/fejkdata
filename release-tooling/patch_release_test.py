import io
import json
import os
import subprocess
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout
from unittest import mock

import patch_release
import publish_release

DOCKERFILE = "ARG GO_VERSION={}\nFROM golang:${{GO_VERSION}} AS portable\n"
COMPOSE = "x-go: &go\n  image: golang:${{GO_VERSION:-{}}}\n"
RELEASED = """# Changelog

## [0.1.0] - 2026-10-13

The first release.
"""
MAIN = """# Changelog

## [Unreleased]

### Added

- A thing.

## [0.1.0] - 2026-10-13

The first release.
"""
SECTION = (
	"## [0.1.1] - 2026-10-14\n\n### Security\n\n- The binaries and image are built with Go 1.27.2.\n"
	"- The binaries no longer carry Go 1.27.1's known vulnerabilities: GO-2026-0001.\n\n"
)


class Due(unittest.TestCase):
	def test_main_s_go_a_patch_above_the_last_release_s_calls_for_a_patch(self):
		self.assertEqual(
			patch_release.due("v0.1.0", DOCKERFILE.format("1.27.1"), DOCKERFILE.format("1.27.2")), ("0.1.1", "1.27.2")
		)

	def test_a_newer_go_minor_waits_for_a_release_the_maintainer_cuts(self):
		self.assertIsNone(patch_release.due("v0.1.0", DOCKERFILE.format("1.27.2"), DOCKERFILE.format("1.28.0")))

	def test_the_same_go_calls_for_nothing(self):
		self.assertIsNone(patch_release.due("v0.1.1", DOCKERFILE.format("1.27.2"), DOCKERFILE.format("1.27.2")))

	def test_an_older_go_calls_for_nothing(self):
		self.assertIsNone(patch_release.due("v0.1.0", DOCKERFILE.format("1.27.10"), DOCKERFILE.format("1.27.9")))

	def test_the_patch_after_a_two_digit_patch(self):
		self.assertEqual(patch_release.due("v0.1.9", DOCKERFILE.format("1.27.1"), DOCKERFILE.format("1.27.2"))[0], "0.1.10")

	def test_a_dockerfile_with_no_go_version_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.due("v0.1.0", "FROM scratch\n", DOCKERFILE.format("1.27.2"))

	def test_a_tag_not_of_the_form_vxyz_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.due("latest", DOCKERFILE.format("1.27.1"), DOCKERFILE.format("1.27.2"))


class NextPatch(unittest.TestCase):
	def test_the_patch_after_a_tag(self):
		self.assertEqual(patch_release.next_patch("v0.1.9"), "0.1.10")

	def test_a_tag_not_of_the_form_vxyz_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.next_patch("0.1.0")


class Apply(unittest.TestCase):
	def test_sets_go_in_the_dockerfile_and_compose(self):
		self.assertEqual(patch_release.set_go(DOCKERFILE.format("1.27.1"), "1.27.2"), DOCKERFILE.format("1.27.2"))
		self.assertEqual(patch_release.set_go(COMPOSE.format("1.27.1"), "1.27.2"), COMPOSE.format("1.27.2"))

	def test_a_file_naming_no_go_version_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.set_go("FROM scratch\n", "1.27.2")

	def test_a_section_names_the_go_version_and_what_it_fixes(self):
		self.assertEqual(patch_release.section("0.1.1", "2026-10-14", "1.27.2", "1.27.1", "GO-2026-0001"), SECTION)

	def test_a_section_fixing_nothing_names_only_the_go_version(self):
		self.assertEqual(
			patch_release.section("0.1.1", "2026-10-14", "1.27.2", "1.27.1", ""),
			"## [0.1.1] - 2026-10-14\n\n### Security\n\n- The binaries and image are built with Go 1.27.2.\n\n",
		)

	def test_prepare_writes_the_go_version_and_the_section_into_the_release_tree(self):
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		for name, text in (
			("Dockerfile", DOCKERFILE.format("1.27.1")),
			("compose.yaml", COMPOSE.format("1.27.1")),
			("CHANGELOG.md", RELEASED),
		):
			with open(os.path.join(tmp.name, name), "w", encoding="utf-8") as f:
				f.write(text)
		patch_release.prepare(tmp.name, "1.27.2", SECTION)
		for name, text in (
			("Dockerfile", DOCKERFILE.format("1.27.2")),
			("compose.yaml", COMPOSE.format("1.27.2")),
			("CHANGELOG.md", patch_release.insert_section(RELEASED, SECTION)),
		):
			with open(os.path.join(tmp.name, name), encoding="utf-8") as f:
				self.assertEqual(f.read(), text)

	def test_a_section_goes_below_unreleased(self):
		self.assertEqual(patch_release.insert_section(MAIN, SECTION), MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]"))

	def test_a_section_ends_a_changelog_with_no_release(self):
		self.assertEqual(patch_release.insert_section("# Changelog\n\n", SECTION), "# Changelog\n\n" + SECTION)

	def test_a_section_goes_atop_the_released_changelog(self):
		self.assertEqual(
			patch_release.insert_section(RELEASED, SECTION),
			"# Changelog\n\n" + SECTION + "## [0.1.0] - 2026-10-13\n\nThe first release.\n",
		)


class Reconcile(unittest.TestCase):
	def test_a_released_section_main_lacks_goes_below_unreleased(self):
		released = patch_release.insert_section(RELEASED, SECTION)
		self.assertEqual(
			patch_release.reconcile(MAIN, released),
			MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]"),
		)

	def test_a_section_main_has_is_left(self):
		main = MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]")
		self.assertEqual(patch_release.reconcile(main, patch_release.insert_section(RELEASED, SECTION)), main)

	def test_two_missing_sections_keep_their_order(self):
		older = SECTION.replace("0.1.1", "0.1.2").replace("1.27.2", "1.27.3")
		released = patch_release.insert_section(patch_release.insert_section(RELEASED, SECTION), older)
		self.assertEqual(
			patch_release.reconcile(MAIN, released), MAIN.replace("## [0.1.0]", older + SECTION + "## [0.1.0]")
		)

	def test_main_ahead_of_the_release_is_left(self):
		main = MAIN.replace("## [Unreleased]\n\n### Added\n\n- A thing.\n\n", "## [0.2.0] - 2026-11-01\n\n- A thing.\n\n")
		self.assertEqual(patch_release.reconcile(main, RELEASED), main)


class Record(unittest.TestCase):
	def test_a_fix_goes_under_unreleased_security(self):
		self.assertIn(
			"### Added\n\n- A thing.\n\n### Security\n\n- Fix GO-2026-0009.\n\n## [0.1.0]",
			patch_release.record(MAIN, ["Fix GO-2026-0009."]),
		)

	def test_a_second_fix_joins_the_security_section(self):
		once = patch_release.record(MAIN, ["Fix GO-2026-0009."])
		self.assertIn(
			"### Security\n\n- Fix GO-2026-0009.\n- Fix GO-2026-0010.\n\n## [0.1.0]",
			patch_release.record(once, ["Fix GO-2026-0010."]),
		)

	def test_a_fix_whose_vulnerabilities_a_release_names_is_left(self):
		main = MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]")
		self.assertEqual(patch_release.record(main, ["Worded otherwise: GO-2026-0001."]), main)

	def test_a_changelog_topped_by_a_release_gets_unreleased_above_it(self):
		self.assertEqual(
			patch_release.record(RELEASED, ["Fix GO-2026-0009."]),
			"# Changelog\n\n## [Unreleased]\n\n### Security\n\n- Fix GO-2026-0009.\n\n## [0.1.0] - 2026-10-13\n\nThe first release.\n",
		)

	def test_two_fixes_are_two_bullets(self):
		self.assertIn(
			"### Security\n\n- Fix GO-2026-0009.\n- Fix GO-2026-0010.\n\n## [0.1.0]",
			patch_release.record(MAIN, ["Fix GO-2026-0009.", "Fix GO-2026-0010."]),
		)

	def test_a_changelog_with_no_section_gets_unreleased_at_its_end(self):
		self.assertEqual(
			patch_release.record("# Changelog\n", ["Fix GO-2026-0009."]),
			"# Changelog\n## [Unreleased]\n\n### Security\n\n- Fix GO-2026-0009.\n\n",
		)

	def test_a_changelog_with_only_unreleased_ends_in_one_newline(self):
		self.assertEqual(
			patch_release.record("# Changelog\n\n## [Unreleased]\n", ["Fix GO-2026-0009."]),
			"# Changelog\n\n## [Unreleased]\n\n### Security\n\n- Fix GO-2026-0009.\n",
		)

	def test_nothing_fixed_leaves_the_changelog(self):
		self.assertEqual(patch_release.record(MAIN, []), MAIN)


def report(*ids):
	messages = [{"osv": {"id": i, "aliases": [f"CVE-{i[3:]}"], "summary": "A flaw"}} for i in ids]
	messages += [{"finding": {"osv": i, "trace": [{"module": "stdlib", "version": "v1.27.1"}]}} for i in ids]
	return "\n".join(json.dumps(m) for m in messages)


TODAY = __import__("datetime").datetime.now(__import__("datetime").UTC).date().isoformat()
TODO = """# todo

## Scoring

`Score = -R - S/4 + 2*A + 2*G*W`

`Bar = 9`

`Next ID = 7`

| Goal | W |
|---|---|
| 1.1 | 1.00 |

## Items

| ID | Release | Exempt | Item | R | S | C | A | G | Goals | Score |
|---|---|---|---|---|---|---|---|---|---|---|
| 5 | 0.2.0 |  | **Add a thing.** | 2 | 2 | 2 | 7 | 8 | 1.1 | 26.5 |

## Details

### 5. Add a thing.

Why.
"""


def git(*argv):
	return subprocess.run(["git", *argv], check=True, capture_output=True, text=True).stdout.strip()


class Repository(unittest.TestCase):
	"""A git repository whose v0.1.0 holds Go 1.27.1, and whose main has moved on since."""

	def setUp(self):
		cwd = os.getcwd()
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		os.chdir(tmp.name)
		self.addCleanup(os.chdir, cwd)
		git("init", "-q")
		git("config", "user.email", "test@example.com")
		git("config", "user.name", "Test")
		self.write(("Dockerfile", DOCKERFILE.format("1.27.1")), ("compose.yaml", COMPOSE.format("1.27.1")), ("CHANGELOG.md", RELEASED))
		git("add", ".")
		git("commit", "-qm", "Release 0.1.0")
		git("tag", "v0.1.0")
		self.write(("CHANGELOG.md", MAIN))
		git("commit", "-qam", "Add a thing")

	def write(self, *files):
		for name, text in files:
			with open(name, "w", encoding="utf-8") as f:
				f.write(text)

	def read(self, name):
		with open(name, encoding="utf-8") as f:
			return f.read()


class Cut(Repository):
	def setUp(self):
		super().setUp()
		self.reports = {"1.27.1": report("GO-2026-0001", "GO-2026-0002"), "1.27.2": report()}
		self.scanned, self.outputs = [], []
		self.published = mock.Mock(return_value=(0, "v0.1.1"))

	def scan(self, tree, go, into):
		self.scanned.append(tree)
		os.makedirs(into)
		with open(os.path.join(into, "report.json"), "w", encoding="utf-8") as f:
			f.write(self.reports[go])
		return os.path.join(into, "report.json")

	def ship(self, checkout, version, go):
		self.shipped = {name: self.read(os.path.join(checkout, name)) for name in ("Dockerfile", "compose.yaml", "CHANGELOG.md")}
		self.shipped_from = checkout
		return "abc123"

	def cut(self):
		with redirect_stdout(io.StringIO()) as out:
			code = patch_release.cut("v0.1.0", self.scan, self.ship, self.published, self.outputs.append)
		return code, out.getvalue()

	def test_cuts_a_patch_of_the_last_release_on_main_s_go(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")))
		code, _ = self.cut()
		self.assertEqual((code, self.outputs), (0, ["tag=v0.1.1"]))
		self.assertEqual(self.shipped["Dockerfile"], DOCKERFILE.format("1.27.2"))
		self.assertEqual(self.shipped["compose.yaml"], COMPOSE.format("1.27.2"))
		self.assertEqual(
			self.shipped["CHANGELOG.md"],
			f"# Changelog\n\n## [0.1.1] - {TODAY}\n\n### Security\n\n- The binaries and image are built with Go 1.27.2.\n"
			"- The binaries no longer carry Go 1.27.1's known vulnerabilities: GO-2026-0001 (CVE-2026-0001),"
			" GO-2026-0002 (CVE-2026-0002).\n\n## [0.1.0] - 2026-10-13\n\nThe first release.\n",
		)
		self.assertEqual(self.scanned, [self.shipped_from, self.shipped_from])
		self.published.assert_called_once_with(self.shipped["CHANGELOG.md"], "abc123")

	def test_no_patch_is_due_on_the_last_release_s_go(self):
		self.assertEqual(self.cut(), (0, "No patch due: main's Go is no patch above v0.1.0's.\n"))
		self.published.assert_not_called()

	def test_a_release_tree_still_vulnerable_on_main_s_go_is_refused_with_a_warning(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")))
		self.reports["1.27.2"] = report("GO-2026-0002")
		code, out = self.cut()
		self.assertEqual(code, 0)
		self.assertEqual(out, "::warning::v0.1.0's tree on Go 1.27.2 still carries a known vulnerability; no patch is cut\n")
		self.published.assert_not_called()


class Plan(Repository):
	def setUp(self):
		super().setUp()
		git("checkout", "-q", "v0.1.0")
		self.write(("CHANGELOG.md", patch_release.insert_section(RELEASED, SECTION)))
		git("commit", "-qam", "Release 0.1.1")
		git("tag", "v0.1.1")
		git("checkout", "-q", "-")
		self.write(("todo.md", TODO))

	def plan(self, *ids):
		self.write(("report.json", report(*ids).replace("v1.27.1", "v1.27.2")))
		with redirect_stderr(io.StringIO()):
			return patch_release.plan("report.json", "todo.md", "CHANGELOG.md", "v0.1.1")

	def test_files_under_the_next_patch_and_reconciles_the_last_release(self):
		self.assertEqual(self.plan("GO-2026-0009"), 0)
		self.assertIn("\n| 7 | 0.1.2 | security |", self.read("todo.md"))
		self.assertEqual(self.read("CHANGELOG.md"), MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]"))

	def test_records_a_fixed_item_under_unreleased(self):
		self.plan("GO-2026-0009")
		self.plan()
		self.assertNotIn("\n| 7 |", self.read("todo.md"))
		self.assertIn(
			"### Security\n\n- The binaries no longer carry Go 1.27.2's known vulnerabilities: GO-2026-0009 (CVE-2026-0009).\n\n## [0.1.1]",
			self.read("CHANGELOG.md"),
		)


class Commands(unittest.TestCase):
	def setUp(self):
		self.calls = []

	def run_stub(self, *argv, cwd=None, ok=(0,)):
		self.calls.append((argv, cwd))
		return subprocess.CompletedProcess(argv, 0, stdout="abc123\n", stderr="")

	def test_scan_builds_the_tree_s_report_through_main_s_dockerfile(self):
		with mock.patch.object(patch_release, "run", self.run_stub):
			self.assertEqual(patch_release.scan("/tree", "1.27.2", "/into"), "/into/report.json")
		self.assertEqual(
			self.calls,
			[((
				"docker", "build", "-f", os.path.abspath("Dockerfile"), "--no-cache-filter", "vulnerabilities",
				"--build-arg", "GO_VERSION=1.27.2", "--target", "vulnerability-report", "--output", "/into", "/tree",
			), None)],
		)

	def test_ship_builds_commits_and_pushes_the_tree_to_its_release_branch(self):
		env = {"FORGE_SERVER_URL": "https://forge.example", "FORGE_REPOSITORY": "owner/repo", "FORGE_TOKEN": "t0ken"}
		with mock.patch.object(patch_release, "run", self.run_stub), mock.patch.dict(os.environ, env):
			self.assertEqual(patch_release.ship("/tree", "0.1.1", "1.27.2"), "abc123")
		self.assertEqual(
			self.calls,
			[
				(("docker", "build", "."), "/tree"),
				(("git", "commit", "-qam", "Release 0.1.1, built with Go 1.27.2"), "/tree"),
				(("git", "push", "-qf", "https://x-access-token:t0ken@forge.example/owner/repo.git", "HEAD:refs/heads/release/0.1.1"), "/tree"),
				(("git", "rev-parse", "HEAD"), "/tree"),
			],
		)


class Main(unittest.TestCase):
	def test_last_prints_the_forge_s_latest_release(self):
		with mock.patch.object(publish_release, "request", return_value={"tag_name": "v0.1.0"}), \
			redirect_stdout(io.StringIO()) as out:
			self.assertEqual(patch_release.main(["last"]), 0)
		self.assertEqual(out.getvalue(), "v0.1.0\n")

	def test_a_failing_step_exits_2_with_its_message(self):
		with mock.patch.object(patch_release, "cut", side_effect=ValueError("a step failed")), \
			redirect_stderr(io.StringIO()) as err:
			self.assertEqual(patch_release.main(["cut"]), 2)
		self.assertEqual(err.getvalue(), "a step failed\n")

	def test_a_file_it_cannot_read_exits_2_with_a_message(self):
		with redirect_stderr(io.StringIO()) as err:
			self.assertEqual(patch_release.main(["plan", "/nonexistent/report.json", "todo.md", "CHANGELOG.md"]), 2)
		self.assertEqual(err.getvalue(), "cannot read /nonexistent/report.json: No such file or directory\n")

	def test_an_unknown_command_exits_2_with_the_usage(self):
		with redirect_stderr(io.StringIO()) as err:
			self.assertEqual(patch_release.main(["publish"]), 2)
		self.assertIn("patch_release.py plan", err.getvalue())


if __name__ == "__main__":
	unittest.main()
