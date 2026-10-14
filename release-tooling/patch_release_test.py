import io
import os
import tempfile
import unittest
from contextlib import redirect_stdout

import patch_release

DOCKERFILE = "ARG GO_VERSION={}\nFROM golang:${{GO_VERSION}} AS portable\n"
COMPOSE = "x-go: &go\n  image: golang:${{GO_VERSION:-{}}}\n"
MAIN = """# Changelog

## [Unreleased]

### Added

- A thing.

## [0.1.1] - 2026-10-14

### Security

- Build the binaries and image with Go 1.27.2.

## [0.1.0] - 2026-10-13

The first release.
"""
RELEASED = """# Changelog

## [0.1.0] - 2026-10-13

The first release.
"""


class GoPatch(unittest.TestCase):
	def test_a_patch_bump_names_the_new_version(self):
		self.assertEqual(patch_release.go_patch(DOCKERFILE.format("1.27.1"), DOCKERFILE.format("1.27.2")), "1.27.2")

	def test_a_minor_bump_is_no_patch(self):
		self.assertIsNone(patch_release.go_patch(DOCKERFILE.format("1.27.2"), DOCKERFILE.format("1.28.0")))

	def test_an_unchanged_version_is_no_patch(self):
		self.assertIsNone(patch_release.go_patch(DOCKERFILE.format("1.27.2"), DOCKERFILE.format("1.27.2")))

	def test_a_downgrade_is_no_patch(self):
		self.assertIsNone(patch_release.go_patch(DOCKERFILE.format("1.27.10"), DOCKERFILE.format("1.27.9")))


class NextVersion(unittest.TestCase):
	def test_the_next_patch_of_a_tag(self):
		self.assertEqual(patch_release.next_version("v0.1.9"), "0.1.10")

	def test_a_tag_that_is_no_version_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.next_version("latest")


class Apply(unittest.TestCase):
	def test_sets_go_in_the_dockerfile_and_compose(self):
		self.assertEqual(patch_release.set_go(DOCKERFILE.format("1.27.1"), "1.27.2"), DOCKERFILE.format("1.27.2"))
		self.assertEqual(patch_release.set_go(COMPOSE.format("1.27.1"), "1.27.2"), COMPOSE.format("1.27.2"))

	def test_takes_the_release_section_from_main_atop_the_released_changelog(self):
		changelog, notes = patch_release.carry_section(MAIN, RELEASED, "0.1.1")
		self.assertEqual(
			changelog,
			"# Changelog\n\n## [0.1.1] - 2026-10-14\n\n### Security\n\n- Build the binaries and image with Go 1.27.2.\n\n"
			"## [0.1.0] - 2026-10-13\n\nThe first release.\n",
		)
		self.assertEqual(notes, "### Security\n\n- Build the binaries and image with Go 1.27.2.")

	def test_a_version_main_does_not_record_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.carry_section(MAIN, RELEASED, "0.1.2")

	def test_main_writes_both_files_and_prints_the_notes(self):
		cwd = os.getcwd()
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		os.chdir(tmp.name)
		self.addCleanup(os.chdir, cwd)
		os.mkdir("release")
		files = {
			"main.md": MAIN,
			"release/CHANGELOG.md": RELEASED,
			"release/Dockerfile": DOCKERFILE.format("1.27.1"),
			"release/compose.yaml": COMPOSE.format("1.27.1"),
		}
		for name, text in files.items():
			with open(name, "w", encoding="utf-8") as f:
				f.write(text)
		out = io.StringIO()
		with redirect_stdout(out):
			self.assertEqual(patch_release.main(["apply", "release", "main.md", "0.1.1", "1.27.2"]), 0)
		self.assertEqual(out.getvalue(), "### Security\n\n- Build the binaries and image with Go 1.27.2.\n")
		with open("release/Dockerfile", encoding="utf-8") as f:
			self.assertEqual(f.read(), DOCKERFILE.format("1.27.2"))
		with open("release/CHANGELOG.md", encoding="utf-8") as f:
			self.assertTrue(f.read().startswith("# Changelog\n\n## [0.1.1] - 2026-10-14\n"))


if __name__ == "__main__":
	unittest.main()
