import unittest

import go_release_notes
from fixtures import GORELEASER, NOTES

BUILT_FOR = {"darwin", "linux", "windows", "amd64", "arm64"}


class Systems(unittest.TestCase):
	def test_are_goreleaser_yaml_s_goos_and_goarch(self):
		self.assertEqual(go_release_notes.systems(GORELEASER), BUILT_FOR)

	def test_a_goreleaser_yaml_naming_no_goos_is_refused(self):
		with self.assertRaisesRegex(ValueError, "goos"):
			go_release_notes.systems("builds:\n  - goarch: [amd64]\n")


class Requirements(unittest.TestCase):
	def test_names_what_a_new_minor_requires_of_a_built_for_system(self):
		self.assertEqual(
			go_release_notes.requirements(NOTES, "1.28", BUILT_FOR),
			["Go 1.28 requires macOS 14 Sonoma or later", "Go 1.28 now requires at least Windows 10.1 or Windows Server 2019"],
		)

	def test_notes_with_no_ports_section_name_nothing(self):
		self.assertEqual(go_release_notes.requirements("# Go 1.28 Release Notes\n", "1.28", BUILT_FOR), [])

	def test_a_requirement_in_another_shape_is_refused(self):
		for sentence in (
			"Go 1.28 drops the arm64 port on Windows.",
			"Go 1.28 raises its minimum requirement to ARMv8.1.",
			"Go 1.28 removes the arm64 port on Windows.",
		):
			with self.subTest(sentence=sentence):
				notes = NOTES.replace("### FreeBSD", f"### ARM64\n\n{sentence}\n\n### FreeBSD")
				with self.assertRaisesRegex(ValueError, f"under ARM64: '{sentence}'"):
					go_release_notes.requirements(notes, "1.28", BUILT_FOR)

	def test_a_sentence_in_another_shape_after_a_last_release_one_is_refused(self):
		notes = NOTES.replace("Go 1.29 will require Linux kernel version 4.4 or later.", "Go 1.28 no longer supports the 386 port.")
		with self.assertRaisesRegex(ValueError, "under Linux: 'Go 1.28 no longer supports the 386 port.'"):
			go_release_notes.requirements(notes, "1.28", BUILT_FOR)

	def test_a_port_the_notes_add_is_no_requirement(self):
		notes = NOTES.replace("### FreeBSD", "### ARM64\n\nGo 1.28 adds support for a new arm64 feature.\n\n### FreeBSD")
		self.assertEqual(
			go_release_notes.requirements(notes, "1.28", BUILT_FOR),
			["Go 1.28 requires macOS 14 Sonoma or later", "Go 1.28 now requires at least Windows 10.1 or Windows Server 2019"],
		)

	def test_a_heading_names_its_target_in_any_case(self):
		self.assertIn("Go 1.28 requires FreeBSD 14 or later", go_release_notes.requirements(NOTES, "1.28", BUILT_FOR | {"freebsd"}))


class Plain(unittest.TestCase):
	def test_drops_comments_and_link_targets_and_folds_whitespace(self):
		self.assertEqual(
			go_release_notes.plain("<!-- go.dev/issue/1 -->\nAs [announced](go1.27#darwin) in\nthe notes."),
			"As announced in the notes.",
		)


class BreakingLines(unittest.TestCase):
	def test_reads_the_notes_of_each_minor_above_the_released_go(self):
		read = []

		def notes(minor):
			read.append(minor)
			return NOTES if minor == "1.28" else ""

		self.assertEqual(
			go_release_notes.breaking_lines((1, 27, 3), (1, 29, 0), BUILT_FOR, notes),
			[
				"Go 1.28 requires macOS 14 Sonoma or later, and so do the binaries.",
				"Go 1.28 now requires at least Windows 10.1 or Windows Server 2019, and so do the binaries.",
			],
		)
		self.assertEqual(read, ["1.28", "1.29"])

	def test_a_go_patch_reads_no_notes(self):
		self.assertEqual(go_release_notes.breaking_lines((1, 27, 1), (1, 27, 2), BUILT_FOR, None), [])


if __name__ == "__main__":
	unittest.main()
