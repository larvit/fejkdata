import io
import os
import tempfile
import unittest
from contextlib import redirect_stderr

import patch_release

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


class Apply(unittest.TestCase):
	def test_sets_go_in_the_dockerfile_and_compose(self):
		self.assertEqual(patch_release.set_go(DOCKERFILE.format("1.27.1"), "1.27.2"), DOCKERFILE.format("1.27.2"))
		self.assertEqual(patch_release.set_go(COMPOSE.format("1.27.1"), "1.27.2"), COMPOSE.format("1.27.2"))

	def test_a_file_naming_no_go_version_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.set_go("FROM scratch\n", "1.27.2")

	def test_a_section_names_the_go_version_and_each_fixed_line(self):
		lines = ["The binaries no longer carry Go 1.27.1's known vulnerabilities: GO-2026-0001."]
		self.assertEqual(patch_release.section("0.1.1", "2026-10-14", "1.27.2", lines), SECTION)

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


class Main(unittest.TestCase):
	def test_reconcile_writes_main_s_changelog(self):
		cwd = os.getcwd()
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		os.chdir(tmp.name)
		self.addCleanup(os.chdir, cwd)
		for name, text in (("main.md", MAIN), ("released.md", patch_release.insert_section(RELEASED, SECTION))):
			with open(name, "w", encoding="utf-8") as f:
				f.write(text)
		self.assertEqual(patch_release.main(["reconcile", "main.md", "released.md"]), 0)
		with open("main.md", encoding="utf-8") as f:
			self.assertEqual(f.read(), MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]"))

	def test_an_unknown_command_exits_2_with_the_usage(self):
		with redirect_stderr(io.StringIO()) as err:
			self.assertEqual(patch_release.main(["publish"]), 2)
		self.assertIn("patch_release.py cut", err.getvalue())


if __name__ == "__main__":
	unittest.main()
