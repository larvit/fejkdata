import unittest

import carried
from fixtures import COMPOSE, DOCKERFILE, GO_MOD, NEW_GO_MOD, PACKAGES, carrying


class Of(unittest.TestCase):
	def test_reads_go_goreleaser_and_the_modules_from_a_tree(self):
		files = {
			"Dockerfile": DOCKERFILE.format("1.27.1"), ".github/workflows/packages.yml": PACKAGES.format("v2.18.3"),
			"go.mod": GO_MOD.format("1.22", "v0.29.0"), "go.sum": "sum\n",
		}
		self.assertEqual(carried.of(lambda path: files.get(path, "")), carrying(go_sum="sum\n"))

	def test_a_tree_whose_packages_yml_names_no_goreleaser_carries_none(self):
		self.assertIsNone(carried.of(lambda path: DOCKERFILE.format("1.27.1") if path == "Dockerfile" else "").goreleaser)

	def test_a_dockerfile_with_no_go_version_is_refused(self):
		with self.assertRaises(ValueError):
			carried.of(lambda path: "FROM scratch\n")


class Due(unittest.TestCase):
	def test_main_s_go_a_patch_above_the_last_release_s_is_carried(self):
		self.assertEqual(carried.due(carrying("1.27.1"), carrying("1.27.2")), carrying("1.27.2"))

	def test_main_s_go_a_minor_above_the_last_release_s_is_carried(self):
		self.assertEqual(carried.due(carrying("1.27.2"), carrying("1.28.0")), carrying("1.28.0"))

	def test_main_carrying_what_the_release_does_calls_for_nothing(self):
		self.assertIsNone(carried.due(carrying(), carrying()))

	def test_an_older_go_calls_for_nothing(self):
		self.assertIsNone(carried.due(carrying("1.27.10"), carrying("1.27.9")))

	def test_main_s_goreleaser_is_carried(self):
		self.assertEqual(carried.due(carrying(), carrying(goreleaser="v2.19.0")), carrying(goreleaser="v2.19.0"))

	def test_a_release_whose_tree_names_no_goreleaser_takes_main_s(self):
		self.assertEqual(carried.due(carrying(goreleaser=None), carrying()), carrying())

	def test_main_naming_no_goreleaser_is_refused(self):
		with self.assertRaisesRegex(ValueError, "packages.yml names no GoReleaser"):
			carried.due(carrying(), carrying(goreleaser=None))

	def test_main_s_modules_are_carried_with_its_go_sum(self):
		main = carrying(go_mod=NEW_GO_MOD, go_sum="new sum\n")
		self.assertEqual(carried.due(carrying(), main), main)

	def test_modules_wait_while_main_raises_the_lowest_go(self):
		self.assertIsNone(carried.due(carrying(), carrying(go_mod=GO_MOD.format("1.23", "v0.30.0"))))

	def test_a_go_sum_alone_calls_for_nothing(self):
		self.assertIsNone(carried.due(carrying(), carrying(go_sum="other sum\n")))

	def test_modules_wait_while_main_requires_another_module_directly(self):
		main = NEW_GO_MOD + "\nrequire example.com/feature v1.0.0\n"
		self.assertIsNone(carried.due(carrying(), carrying(go_mod=main)))

	def test_modules_wait_while_main_s_go_mod_changes_a_line_besides_its_requires(self):
		main = NEW_GO_MOD + "\nreplace golang.org/x/text => example.com/text v0.30.1\n"
		self.assertIsNone(carried.due(carrying(), carrying(go_mod=main)))

	def test_modules_wait_while_main_s_go_mod_excludes_a_version(self):
		main = NEW_GO_MOD + "\nexclude golang.org/x/text v0.29.0\n"
		self.assertIsNone(carried.due(carrying(), carrying(go_mod=main)))

	def test_an_indirect_module_a_bump_adds_or_drops_is_carried(self):
		self.assertEqual(carried.due(carrying(), carrying(go_mod=NEW_GO_MOD)).go_mod, NEW_GO_MOD)


class Requires(unittest.TestCase):
	def test_reads_a_require_block_and_single_lines(self):
		self.assertEqual(
			carried.requires(GO_MOD.format("1.22", "v0.29.0") + NEW_GO_MOD),
			{"golang.org/x/text": "v0.30.0", "example.com/gone": "v1.0.0", "example.com/new": "v0.1.0"},
		)

	def test_a_go_mod_requiring_nothing(self):
		self.assertEqual(carried.requires("module example.com/m\n\ngo 1.22\n"), {})


class SetGo(unittest.TestCase):
	def test_sets_go_in_the_dockerfile_and_compose(self):
		self.assertEqual(carried.set_go(DOCKERFILE.format("1.27.1"), "1.27.2"), DOCKERFILE.format("1.27.2"))
		self.assertEqual(carried.set_go(COMPOSE.format("1.27.1"), "1.27.2"), COMPOSE.format("1.27.2"))

	def test_a_file_naming_no_go_version_is_refused(self):
		with self.assertRaises(ValueError):
			carried.set_go("FROM scratch\n", "1.27.2")


if __name__ == "__main__":
	unittest.main()
