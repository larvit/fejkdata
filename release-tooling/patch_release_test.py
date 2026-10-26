import datetime
import io
import json
import os
import subprocess
import tempfile
import unittest
import urllib.error
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
PACKAGES = "    steps:\n      - uses: goreleaser/goreleaser-action@v7\n        with:\n          version: {}\n"
GO_MOD = "module example.com/m\n\ngo {}\n\nrequire (\n\tgolang.org/x/text {}\n\texample.com/gone v1.0.0 // indirect\n)\n"
NEW_GO_MOD = "module example.com/m\n\ngo 1.22\n\nrequire golang.org/x/text v0.30.0\n\nrequire example.com/new v0.1.0 // indirect\n"
SECTION = (
	"## [0.1.1] - 2026-10-14\n\n### Changed\n\n- The binaries and image are built with Go 1.27.2.\n\n"
	"### Security\n\n- The binaries no longer carry 0.1.0's known vulnerabilities: GO-2026-0001.\n\n"
)
NOTES = """# Go 1.28 Release Notes

## Ports {#ports}

### Darwin {#darwin}

<!-- go.dev/issue/1 -->
As [announced](go1.27#darwin) in the Go 1.27 release notes,
Go 1.28 requires macOS 14 Sonoma or later;
support for previous versions has been discontinued.

### Linux {#linux}

Go 1.28 is the last release that requires Linux kernel version 3.2 or later. Go 1.29 will require Linux kernel version 4.4 or later.

### Windows

Go 1.28 requires at least Windows 10.1 or Windows Server 2019.

### FreeBSD

Go 1.28 requires FreeBSD 14 or later.

## Standard library

Go 1.28 requires nothing new of Darwin.
"""


def carried(go="1.27.1", goreleaser="v2.18.3", go_mod=GO_MOD.format("1.22", "v0.29.0"), go_sum="sum\n"):
	return patch_release.Carried(patch_release.go_of(DOCKERFILE.format(go)), goreleaser, go_mod, go_sum)


class Due(unittest.TestCase):
	def test_main_s_go_a_patch_above_the_last_release_s_is_carried(self):
		self.assertEqual(patch_release.due(carried("1.27.1"), carried("1.27.2")), carried("1.27.2"))

	def test_main_s_go_a_minor_above_the_last_release_s_is_carried(self):
		self.assertEqual(patch_release.due(carried("1.27.2"), carried("1.28.0")), carried("1.28.0"))

	def test_main_carrying_what_the_release_does_calls_for_nothing(self):
		self.assertIsNone(patch_release.due(carried(), carried()))

	def test_an_older_go_calls_for_nothing(self):
		self.assertIsNone(patch_release.due(carried("1.27.10"), carried("1.27.9")))

	def test_main_s_goreleaser_is_carried(self):
		self.assertEqual(patch_release.due(carried(), carried(goreleaser="v2.19.0")), carried(goreleaser="v2.19.0"))

	def test_main_s_modules_are_carried_with_its_go_sum(self):
		main = carried(go_mod=NEW_GO_MOD, go_sum="new sum\n")
		self.assertEqual(patch_release.due(carried(), main), main)

	def test_modules_wait_while_main_raises_the_lowest_go(self):
		self.assertIsNone(patch_release.due(carried(), carried(go_mod=GO_MOD.format("1.23", "v0.30.0"))))

	def test_a_go_sum_alone_calls_for_nothing(self):
		self.assertIsNone(patch_release.due(carried(), carried(go_sum="other sum\n")))

	def test_a_dockerfile_with_no_go_version_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.go_of("FROM scratch\n")


class NextPatch(unittest.TestCase):
	def test_the_patch_after_a_tag(self):
		self.assertEqual(patch_release.next_patch("v0.1.9"), "0.1.10")

	def test_a_tag_not_of_the_form_vxyz_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.next_patch("0.1.0")


class Requires(unittest.TestCase):
	def test_reads_a_require_block_and_single_lines(self):
		self.assertEqual(
			patch_release.requires(GO_MOD.format("1.22", "v0.29.0") + NEW_GO_MOD),
			{"golang.org/x/text": "v0.30.0", "example.com/gone": "v1.0.0", "example.com/new": "v0.1.0"},
		)

	def test_a_go_mod_requiring_nothing(self):
		self.assertEqual(patch_release.requires("module example.com/m\n\ngo 1.22\n"), {})


class Platforms(unittest.TestCase):
	def test_names_what_a_new_minor_requires_of_a_platform_the_binaries_are_built_for(self):
		self.assertEqual(
			patch_release.platforms_dropped(NOTES, "1.28"),
			["Go 1.28 requires macOS 14 Sonoma or later", "Go 1.28 requires at least Windows 10.1 or Windows Server 2019"],
		)

	def test_notes_with_no_ports_section_name_nothing(self):
		self.assertEqual(patch_release.platforms_dropped("# Go 1.28 Release Notes\n", "1.28"), [])

	def test_reads_the_notes_of_each_minor_above_the_released_go(self):
		read = []

		def notes(minor):
			read.append(minor)
			return NOTES if minor == "1.28" else ""

		self.assertEqual(
			patch_release.platform_lines((1, 27, 3), (1, 29, 0), notes),
			[
				"Go 1.28 requires macOS 14 Sonoma or later, and so do the binaries.",
				"Go 1.28 requires at least Windows 10.1 or Windows Server 2019, and so do the binaries.",
			],
		)
		self.assertEqual(read, ["1.28", "1.29"])

	def test_a_go_patch_reads_no_notes(self):
		self.assertEqual(patch_release.platform_lines((1, 27, 1), (1, 27, 2), None), [])


class Apply(unittest.TestCase):
	def test_sets_go_in_the_dockerfile_and_compose(self):
		self.assertEqual(patch_release.set_go(DOCKERFILE.format("1.27.1"), "1.27.2"), DOCKERFILE.format("1.27.2"))
		self.assertEqual(patch_release.set_go(COMPOSE.format("1.27.1"), "1.27.2"), COMPOSE.format("1.27.2"))

	def test_a_file_naming_no_go_version_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.set_go("FROM scratch\n", "1.27.2")

	def test_a_section_names_what_changed_and_what_it_fixes(self):
		self.assertEqual(
			patch_release.section("0.1.1", "2026-10-14", ["The binaries and image are built with Go 1.27.2."], "v0.1.0", "GO-2026-0001"),
			SECTION,
		)

	def test_a_section_fixing_nothing_names_only_what_changed(self):
		self.assertEqual(
			patch_release.section("0.1.1", "2026-10-14", ["The binaries and image are built with Go 1.27.2."], "v0.1.0", ""),
			"## [0.1.1] - 2026-10-14\n\n### Changed\n\n- The binaries and image are built with Go 1.27.2.\n\n",
		)

	def test_changes_name_go_its_platforms_goreleaser_and_each_module(self):
		patch = carried("1.28.0", "v2.19.0", NEW_GO_MOD)
		self.assertEqual(
			patch_release.changes(carried(), patch, ["Go 1.28 requires macOS 14 Sonoma or later, and so do the binaries."]),
			[
				"The binaries and image are built with Go 1.28.0.",
				"Go 1.28 requires macOS 14 Sonoma or later, and so do the binaries.",
				"The binaries, packages and image are built with GoReleaser v2.19.0.",
				"The binaries no longer carry example.com/gone.",
				"The binaries carry example.com/new v0.1.0.",
				"The binaries carry golang.org/x/text v0.30.0, up from v0.29.0.",
			],
		)

	def test_prepare_writes_what_the_patch_carries_into_the_release_tree(self):
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		os.makedirs(os.path.join(tmp.name, "release-tooling"))
		pin = "FROM goreleaser/goreleaser:{} AS packaging\n"
		for name, text in (
			("Dockerfile", DOCKERFILE.format("1.27.1") + pin.format("v2.18.3")),
			("compose.yaml", COMPOSE.format("1.27.1")),
			("release-tooling/package-repos.Dockerfile", pin.format("v2.18.3")),
			("go.mod", GO_MOD.format("1.22", "v0.29.0")),
			("go.sum", "sum\n"),
		):
			with open(os.path.join(tmp.name, name), "w", encoding="utf-8") as f:
				f.write(text)
		patch_release.prepare(tmp.name, carried(), carried("1.27.2", "v2.19.0", NEW_GO_MOD, "new sum\n"))
		for name, text in (
			("Dockerfile", DOCKERFILE.format("1.27.2") + pin.format("v2.19.0")),
			("compose.yaml", COMPOSE.format("1.27.2")),
			("release-tooling/package-repos.Dockerfile", pin.format("v2.19.0")),
			("go.mod", NEW_GO_MOD),
			("go.sum", "new sum\n"),
		):
			with open(os.path.join(tmp.name, name), encoding="utf-8") as f:
				self.assertEqual(f.read(), text)

	def test_prepare_leaves_what_the_patch_does_not_change(self):
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		with open(os.path.join(tmp.name, "go.mod"), "w", encoding="utf-8") as f:
			f.write(NEW_GO_MOD)
		patch_release.prepare(tmp.name, carried(), carried(go_mod=NEW_GO_MOD, go_sum=""))
		self.assertEqual(os.listdir(tmp.name), ["go.mod"])

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

	def test_a_heading_not_of_the_form_x_y_z_is_no_release(self):
		for name in ("1.2.3-rc.1", "1.2."):
			with self.subTest(name=name):
				candidate = RELEASED.replace("## [0.1.0]", f"## [{name}] - 2026-10-14\n\nA candidate.\n\n## [0.1.0]")
				self.assertEqual([m["name"] for m in patch_release.releases(candidate)], ["0.1.0"])

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


TODO = """# todo

## Scoring

`Score = 100 * V / (2**S + 8)`

`Bar = 1`

`Next ID = 7`

## Items

| ID | Release | Blocked by | Exempt | Item | L | E | I | R | S | Moves | Score |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 5 | 0.2.0 |  |  | **Add a thing.** | 2 | 2 | 2 | 4 | 2 | 1.1 by 7 | 17.9 |

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
		os.makedirs(".github/workflows")
		self.write(
			("Dockerfile", DOCKERFILE.format("1.27.1")), ("compose.yaml", COMPOSE.format("1.27.1")), ("CHANGELOG.md", RELEASED),
			("go.mod", GO_MOD.format("1.22", "v0.29.0")), ("go.sum", "sum\n"),
			(".github/workflows/packages.yml", PACKAGES.format("v2.18.3")), (".github/workflows/gone.yml", "on: push\n"),
		)
		git("add", ".")
		git("commit", "-qm", "Release 0.1.0")
		git("tag", "v0.1.0")
		os.remove(".github/workflows/gone.yml")
		self.write(("CHANGELOG.md", MAIN), (".github/workflows/test.yml", "on: [push, schedule]\n"))
		git("add", "-A")
		git("commit", "-qm", "Add a thing")

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
		self.reports = {"1.27.1": report("GO-2026-0001", "GO-2026-0002"), "1.27.2": report(), "1.28.0": report()}
		self.scanned, self.outputs, self.notes_read = [], [], []
		self.published = mock.Mock(return_value=(0, "v0.1.1"))
		self.effects = patch_release.Effects(self.scan, self.ship, self.published, self.outputs.append, self.notes)

	def scan(self, tree, go, into):
		self.scanned.append((go, self.read(os.path.join(tree, "go.mod"))))
		os.makedirs(into)
		with open(os.path.join(into, "report.json"), "w", encoding="utf-8") as f:
			f.write(self.reports[go])
		return os.path.join(into, "report.json")

	def ship(self, checkout, version):
		self.shipped = {
			name: self.read(os.path.join(checkout, name)) for name in ("Dockerfile", "compose.yaml", "CHANGELOG.md", "go.mod", "go.sum")
		}
		self.shipped_workflows = {
			name: self.read(os.path.join(checkout, ".github/workflows", name))
			for name in sorted(os.listdir(os.path.join(checkout, ".github/workflows")))
		}
		return "abc123"

	def notes(self, minor):
		self.notes_read.append(minor)
		return NOTES

	def cut(self):
		with redirect_stdout(io.StringIO()) as out:
			code = patch_release.cut("v0.1.0", "2026-10-14", self.effects)
		return code, out.getvalue()

	def test_cuts_a_patch_of_the_last_release_on_main_s_go(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")))
		code, _ = self.cut()
		self.assertEqual((code, self.outputs), (0, ["tag=v0.1.1"]))
		self.assertEqual(self.shipped["Dockerfile"], DOCKERFILE.format("1.27.2"))
		self.assertEqual(self.shipped["compose.yaml"], COMPOSE.format("1.27.2"))
		self.assertEqual(
			self.shipped["CHANGELOG.md"],
			"# Changelog\n\n## [0.1.1] - 2026-10-14\n\n### Changed\n\n- The binaries and image are built with Go 1.27.2.\n\n"
			"### Security\n\n- The binaries no longer carry 0.1.0's known vulnerabilities: GO-2026-0001 (CVE-2026-0001),"
			" GO-2026-0002 (CVE-2026-0002).\n\n## [0.1.0] - 2026-10-13\n\nThe first release.\n",
		)
		self.assertEqual([go for go, _ in self.scanned], ["1.27.1", "1.27.2"])
		self.assertEqual(
			self.shipped_workflows, {"packages.yml": PACKAGES.format("v2.18.3"), "test.yml": "on: [push, schedule]\n"}
		)
		self.assertEqual(self.notes_read, [])
		self.published.assert_called_once_with(self.shipped["CHANGELOG.md"], "abc123")

	def test_a_go_minor_names_the_platforms_it_stops_supporting(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.28.0")))
		self.assertEqual(self.cut()[0], 0)
		self.assertEqual(self.notes_read, ["1.28"])
		self.assertIn(
			"### Changed\n\n- The binaries and image are built with Go 1.28.0.\n"
			"- Go 1.28 requires macOS 14 Sonoma or later, and so do the binaries.\n"
			"- Go 1.28 requires at least Windows 10.1 or Windows Server 2019, and so do the binaries.\n\n",
			self.shipped["CHANGELOG.md"],
		)

	def test_cuts_a_patch_carrying_main_s_modules(self):
		self.write(("go.mod", NEW_GO_MOD), ("go.sum", "new sum\n"))
		self.reports["1.27.1"] = report()
		self.assertEqual(self.cut()[0], 0)
		self.assertEqual((self.shipped["go.mod"], self.shipped["go.sum"]), (NEW_GO_MOD, "new sum\n"))
		self.assertEqual(self.shipped["Dockerfile"], DOCKERFILE.format("1.27.1"))
		self.assertEqual(self.scanned, [("1.27.1", GO_MOD.format("1.22", "v0.29.0")), ("1.27.1", NEW_GO_MOD)])
		self.assertIn("- The binaries carry golang.org/x/text v0.30.0, up from v0.29.0.\n", self.shipped["CHANGELOG.md"])

	def test_cuts_a_patch_carrying_main_s_goreleaser(self):
		self.write((".github/workflows/packages.yml", PACKAGES.format("v2.19.0")))
		self.reports["1.27.1"] = report()
		self.assertEqual(self.cut()[0], 0)
		self.assertEqual(self.shipped_workflows["packages.yml"], PACKAGES.format("v2.19.0"))
		self.assertIn(
			"### Changed\n\n- The binaries, packages and image are built with GoReleaser v2.19.0.\n\n## [0.1.0]",
			self.shipped["CHANGELOG.md"],
		)

	def main_cut(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")), ("output", ""))
		env = {"GITHUB_OUTPUT": os.path.abspath("output")}
		with mock.patch.object(patch_release, "last_release", return_value="v0.1.0"), \
			mock.patch.object(patch_release, "scan", self.scan), \
			mock.patch.object(patch_release, "ship", self.ship), \
			mock.patch.object(publish_release, "publish", self.published), \
			mock.patch.dict(os.environ, env), redirect_stdout(io.StringIO()), redirect_stderr(io.StringIO()) as err:
			code = patch_release.main(["cut"])
		return code, err.getvalue(), self.read("output")

	def test_main_hands_the_cut_tag_to_the_workflow(self):
		self.assertEqual(self.main_cut(), (0, "", "tag=v0.1.1\n"))
		today = datetime.datetime.now(datetime.UTC).date().isoformat()
		self.assertTrue(self.shipped["CHANGELOG.md"].startswith(f"# Changelog\n\n## [0.1.1] - {today}\n\n"))
		self.published.assert_called_once_with(self.shipped["CHANGELOG.md"], "abc123")

	def test_main_names_a_forge_error_while_publishing_and_exits_2(self):
		self.published.side_effect = urllib.error.URLError("forge down")
		code, err, output = self.main_cut()
		self.assertEqual((code, output), (2, ""))
		self.assertRegex(err, "^cannot publish the release: .*forge down.*\n$")

	def test_a_release_already_published_hands_no_tag_on(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")))
		self.published.return_value = (0, None)
		self.assertEqual(self.cut()[0], 0)
		self.assertEqual(self.outputs, [])

	def test_no_patch_is_due_while_main_carries_what_the_release_does(self):
		self.assertEqual(self.cut(), (0, "No patch due: main carries the Go, modules and GoReleaser v0.1.0 does.\n"))
		self.published.assert_not_called()

	def test_a_patch_still_vulnerable_is_refused_with_a_warning(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")))
		self.reports["1.27.2"] = report("GO-2026-0002")
		code, out = self.cut()
		self.assertEqual(code, 0)
		self.assertEqual(out, "::warning::a patch of v0.1.0 on main's Go, modules and GoReleaser still carries a known vulnerability; no patch is cut\n")
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
		self.assertIn("\n| 7 | 0.1.2 |  | security |", self.read("todo.md"))
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

	def test_ship_builds_the_tree_on_the_latest_and_lowest_go_then_commits_and_pushes_it(self):
		env = {"FORGE_SERVER_URL": "https://forge.example", "FORGE_REPOSITORY": "owner/repo", "FORGE_TOKEN": "t0ken"}
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		with open(os.path.join(tmp.name, "go.mod"), "w", encoding="utf-8") as f:
			f.write(GO_MOD.format("1.22", "v0.29.0"))
		with mock.patch.object(patch_release, "run", self.run_stub), mock.patch.dict(os.environ, env):
			self.assertEqual(patch_release.ship(tmp.name, "0.1.1"), "abc123")
		remote = "https://x-access-token:t0ken@forge.example/owner/repo.git"
		self.assertEqual(
			self.calls,
			[
				(("docker", "build", "."), tmp.name),
				(("docker", "build", "--build-arg", "GO_VERSION=1.22", "--target", "portable", "."), tmp.name),
				(("git", "add", "-A"), tmp.name),
				(("git", "commit", "-qm", "Release 0.1.1"), tmp.name),
				(("git", "push", "-q", remote, ":refs/heads/release/0.1.1"), tmp.name),
				(("git", "push", "-q", remote, "HEAD:refs/heads/release/0.1.1"), tmp.name),
				(("git", "rev-parse", "HEAD"), tmp.name),
			],
		)


class Main(unittest.TestCase):
	def test_last_prints_the_forge_s_latest_release(self):
		with mock.patch.object(publish_release, "request", return_value={"tag_name": "v0.1.0"}), \
			redirect_stdout(io.StringIO()) as out:
			self.assertEqual(patch_release.main(["last"]), 0)
		self.assertEqual(out.getvalue(), "v0.1.0\n")

	def test_a_failing_step_exits_2_with_its_message(self):
		with mock.patch.object(patch_release, "last_release", return_value="v0.1.0"), \
			mock.patch.object(patch_release, "cut", side_effect=ValueError("a step failed")), \
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
