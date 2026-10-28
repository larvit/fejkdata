import unittest

import changelog
from fixtures import MAIN, NEW_GO_MOD, RELEASED, SECTION, carrying


class Section(unittest.TestCase):
	def test_a_section_names_what_changed_and_what_it_fixes(self):
		self.assertEqual(
			changelog.section("0.1.1", "2026-10-14", [], ["The binaries and image are built with Go 1.27.2."], "v0.1.0", "GO-2026-0001"),
			SECTION,
		)

	def test_a_section_fixing_nothing_names_only_what_changed(self):
		self.assertEqual(
			changelog.section("0.1.1", "2026-10-14", [], ["The binaries and image are built with Go 1.27.2."], "v0.1.0", ""),
			"## [0.1.1] - 2026-10-14\n\n### Changed\n\n- The binaries and image are built with Go 1.27.2.\n\n",
		)

	def test_a_section_names_the_systems_a_go_minor_drops_first_as_breaking(self):
		self.assertEqual(
			changelog.section(
				"0.1.1", "2026-10-14", ["Go 1.28 requires macOS 14 Sonoma or later, and so do the binaries."],
				["The binaries and image are built with Go 1.28.0."], "v0.1.0", "",
			),
			"## [0.1.1] - 2026-10-14\n\n### Breaking\n\n- Go 1.28 requires macOS 14 Sonoma or later, and so do the binaries.\n\n"
			"### Changed\n\n- The binaries and image are built with Go 1.28.0.\n\n",
		)

	def test_changes_name_go_goreleaser_and_each_module(self):
		self.assertEqual(
			changelog.changes(carrying(), carrying("1.28.0", "v2.19.0", NEW_GO_MOD)),
			[
				"The binaries and image are built with Go 1.28.0.",
				"The binaries, packages and image are built with GoReleaser v2.19.0.",
				"The binaries no longer carry example.com/gone.",
				"The binaries carry example.com/new v0.1.0.",
				"The binaries carry golang.org/x/text v0.30.0, up from v0.29.0.",
			],
		)

	def test_a_section_goes_below_unreleased(self):
		self.assertEqual(changelog.insert_section(MAIN, SECTION), MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]"))

	def test_a_section_ends_a_changelog_with_no_release(self):
		self.assertEqual(changelog.insert_section("# Changelog\n\n", SECTION), "# Changelog\n\n" + SECTION)

	def test_a_section_goes_atop_the_released_changelog(self):
		self.assertEqual(
			changelog.insert_section(RELEASED, SECTION),
			"# Changelog\n\n" + SECTION + "## [0.1.0] - 2026-10-13\n\nThe first release.\n",
		)


class Reconcile(unittest.TestCase):
	def test_a_released_section_main_lacks_goes_below_unreleased(self):
		released = changelog.insert_section(RELEASED, SECTION)
		self.assertEqual(
			changelog.reconcile(MAIN, released),
			MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]"),
		)

	def test_a_section_main_has_is_left(self):
		main = MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]")
		self.assertEqual(changelog.reconcile(main, changelog.insert_section(RELEASED, SECTION)), main)

	def test_two_missing_sections_keep_their_order(self):
		older = SECTION.replace("0.1.1", "0.1.2").replace("1.27.2", "1.27.3")
		released = changelog.insert_section(changelog.insert_section(RELEASED, SECTION), older)
		self.assertEqual(
			changelog.reconcile(MAIN, released), MAIN.replace("## [0.1.0]", older + SECTION + "## [0.1.0]")
		)

	def test_a_heading_not_of_the_form_x_y_z_is_no_release(self):
		for name in ("1.2.3-rc.1", "1.2."):
			with self.subTest(name=name):
				candidate = RELEASED.replace("## [0.1.0]", f"## [{name}] - 2026-10-14\n\nA candidate.\n\n## [0.1.0]")
				self.assertEqual([m["name"] for m in changelog.releases(candidate)], ["0.1.0"])

	def test_a_released_section_drops_the_unreleased_fixes_it_names(self):
		fixes = "### Security\n\n- Fix GO-2026-0001, which a wrapped\n  bullet names.\n- Fix GO-2026-0009.\n- Harden the parser.\n\n"
		self.assertEqual(
			changelog.reconcile(MAIN.replace("## [0.1.0]", fixes + "## [0.1.0]"), changelog.insert_section(RELEASED, SECTION)),
			MAIN.replace("## [0.1.0]", "### Security\n\n- Fix GO-2026-0009.\n- Harden the parser.\n\n" + SECTION + "## [0.1.0]"),
		)

	def test_a_security_heading_left_with_no_fix_goes(self):
		main = MAIN.replace("### Added", "### Security\n\n- Fix GO-2026-0001.\n\n### Added")
		self.assertEqual(
			changelog.reconcile(main, changelog.insert_section(RELEASED, SECTION)), MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]")
		)

	def test_main_ahead_of_the_release_is_left(self):
		main = MAIN.replace("## [Unreleased]\n\n### Added\n\n- A thing.\n\n", "## [0.2.0] - 2026-11-01\n\n- A thing.\n\n")
		self.assertEqual(changelog.reconcile(main, RELEASED), main)


class Record(unittest.TestCase):
	def test_a_fix_goes_under_unreleased_security(self):
		self.assertIn(
			"### Added\n\n- A thing.\n\n### Security\n\n- Fix GO-2026-0009.\n\n## [0.1.0]",
			changelog.record(MAIN, ["Fix GO-2026-0009."]),
		)

	def test_a_second_fix_joins_the_security_section(self):
		once = changelog.record(MAIN, ["Fix GO-2026-0009."])
		self.assertIn(
			"### Security\n\n- Fix GO-2026-0009.\n- Fix GO-2026-0010.\n\n## [0.1.0]",
			changelog.record(once, ["Fix GO-2026-0010."]),
		)

	def test_a_fix_whose_vulnerabilities_a_release_names_is_left(self):
		main = MAIN.replace("## [0.1.0]", SECTION + "## [0.1.0]")
		self.assertEqual(changelog.record(main, ["Worded otherwise: GO-2026-0001."]), main)

	def test_a_changelog_topped_by_a_release_gets_unreleased_above_it(self):
		self.assertEqual(
			changelog.record(RELEASED, ["Fix GO-2026-0009."]),
			"# Changelog\n\n## [Unreleased]\n\n### Security\n\n- Fix GO-2026-0009.\n\n## [0.1.0] - 2026-10-13\n\nThe first release.\n",
		)

	def test_two_fixes_are_two_bullets(self):
		self.assertIn(
			"### Security\n\n- Fix GO-2026-0009.\n- Fix GO-2026-0010.\n\n## [0.1.0]",
			changelog.record(MAIN, ["Fix GO-2026-0009.", "Fix GO-2026-0010."]),
		)

	def test_a_changelog_with_no_section_gets_unreleased_at_its_end(self):
		self.assertEqual(
			changelog.record("# Changelog\n", ["Fix GO-2026-0009."]),
			"# Changelog\n## [Unreleased]\n\n### Security\n\n- Fix GO-2026-0009.\n\n",
		)

	def test_a_changelog_with_only_unreleased_ends_in_one_newline(self):
		self.assertEqual(
			changelog.record("# Changelog\n\n## [Unreleased]\n", ["Fix GO-2026-0009."]),
			"# Changelog\n\n## [Unreleased]\n\n### Security\n\n- Fix GO-2026-0009.\n",
		)

	def test_nothing_fixed_leaves_the_changelog(self):
		self.assertEqual(changelog.record(MAIN, []), MAIN)


if __name__ == "__main__":
	unittest.main()
