import unittest

import go_release_notes
from fixtures import GORELEASER, NOTES

BUILT_FOR = {"Darwin", "Linux", "Windows", "AMD64", "ARM64"}


class Systems(unittest.TestCase):
	def test_are_goreleaser_yaml_s_goos_and_goarch_as_the_notes_head_them(self):
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

	def test_a_requirement_in_another_shape_is_unreadable(self):
		notes = NOTES.replace("### FreeBSD", "### ARM64\n\nGo 1.28 drops the arm64 port on Windows.\n\n### FreeBSD")
		with self.assertRaisesRegex(go_release_notes.Unreadable, "under ARM64: 'Go 1.28 drops the arm64 port on Windows.'"):
			go_release_notes.requirements(notes, "1.28", BUILT_FOR)

	def test_a_port_the_notes_add_is_no_requirement(self):
		notes = NOTES.replace("### FreeBSD", "### ARM64\n\nGo 1.28 adds support for a new arm64 feature.\n\n### FreeBSD")
		self.assertEqual(len(go_release_notes.requirements(notes, "1.28", BUILT_FOR)), 2)


class Lines(unittest.TestCase):
	def test_reads_the_notes_of_each_minor_above_the_released_go(self):
		read = []

		def notes(minor):
			read.append(minor)
			return NOTES if minor == "1.28" else ""

		self.assertEqual(
			go_release_notes.lines((1, 27, 3), (1, 29, 0), BUILT_FOR, notes),
			[
				"Go 1.28 requires macOS 14 Sonoma or later, and so do the binaries.",
				"Go 1.28 now requires at least Windows 10.1 or Windows Server 2019, and so do the binaries.",
			],
		)
		self.assertEqual(read, ["1.28", "1.29"])

	def test_a_go_patch_reads_no_notes(self):
		self.assertEqual(go_release_notes.lines((1, 27, 1), (1, 27, 2), BUILT_FOR, None), [])


if __name__ == "__main__":
	unittest.main()
