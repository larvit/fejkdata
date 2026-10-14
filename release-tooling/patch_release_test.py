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


class Cut(unittest.TestCase):
	def setUp(self):
		cwd = os.getcwd()
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		os.chdir(tmp.name)
		self.addCleanup(os.chdir, cwd)
		git = lambda *argv: subprocess.run(["git", *argv], check=True, capture_output=True)
		git("init", "-q")
		git("config", "user.email", "test@example.com")
		git("config", "user.name", "Test")
		for name, text in (("Dockerfile", DOCKERFILE.format("1.27.1")), ("compose.yaml", COMPOSE.format("1.27.1")), ("CHANGELOG.md", RELEASED)):
			with open(name, "w", encoding="utf-8") as f:
				f.write(text)
		git("add", ".")
		git("commit", "-qm", "Release 0.1.0")
		git("tag", "v0.1.0")
		self.reports = {"1.27.1": report("GO-2026-0001", "GO-2026-0002"), "1.27.2": report()}
		open("output", "w").close()

	def bump(self, go):
		with open("Dockerfile", "w", encoding="utf-8") as f:
			f.write(DOCKERFILE.format(go))

	def scan(self, tree, go, into):
		os.makedirs(into)
		with open(os.path.join(into, "report.json"), "w", encoding="utf-8") as f:
			f.write(self.reports[go])
		return os.path.join(into, "report.json")

	def ship(self, checkout, version, go):
		self.shipped = {}
		for name in ("Dockerfile", "compose.yaml", "CHANGELOG.md"):
			with open(os.path.join(checkout, name), encoding="utf-8") as f:
				self.shipped[name] = f.read()
		return "abc123"

	def cut(self):
		published = mock.Mock(return_value=(0, "v0.1.1"))
		with mock.patch.object(patch_release, "last_release", return_value="v0.1.0"), \
			mock.patch.object(patch_release, "scan", self.scan), \
			mock.patch.object(patch_release, "ship", self.ship), \
			mock.patch.object(publish_release, "publish", published), \
			mock.patch.dict(os.environ, {"GITHUB_OUTPUT": os.path.abspath("output")}), \
			redirect_stdout(io.StringIO()):
			code = patch_release.cut()
		with open("output", encoding="utf-8") as f:
			return code, published, f.read()

	def test_cuts_a_patch_of_the_last_release_on_main_s_go(self):
		self.bump("1.27.2")
		code, published, output = self.cut()
		self.assertEqual((code, output), (0, "tag=v0.1.1\n"))
		self.assertEqual(self.shipped["Dockerfile"], DOCKERFILE.format("1.27.2"))
		self.assertEqual(self.shipped["compose.yaml"], COMPOSE.format("1.27.2"))
		self.assertIn(
			"- The binaries no longer carry Go 1.27.1's known vulnerabilities: GO-2026-0001 (CVE-2026-0001),"
			" GO-2026-0002 (CVE-2026-0002).\n\n## [0.1.0]",
			self.shipped["CHANGELOG.md"],
		)
		published.assert_called_once_with(self.shipped["CHANGELOG.md"], "abc123")

	def test_no_patch_is_due_on_the_last_release_s_go(self):
		code, published, output = self.cut()
		self.assertEqual((code, output), (0, ""))
		published.assert_not_called()

	def test_a_release_tree_still_vulnerable_on_main_s_go_is_not_cut(self):
		self.bump("1.27.2")
		self.reports["1.27.2"] = report("GO-2026-0002")
		with self.assertRaisesRegex(ValueError, "still carries a known vulnerability"):
			self.cut()


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
