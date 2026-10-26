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

import changelog
import patch_release
import publish_release
from fixtures import COMPOSE, DOCKERFILE, GO_MOD, GORELEASER, MAIN, NEW_GO_MOD, NOTES, PACKAGES, RELEASED, SECTION, carrying


class NextPatch(unittest.TestCase):
	def test_the_patch_after_a_tag(self):
		self.assertEqual(patch_release.next_patch("v0.1.9"), "0.1.10")

	def test_a_tag_not_of_the_form_vxyz_is_refused(self):
		with self.assertRaises(ValueError):
			patch_release.next_patch("0.1.0")


class Prepare(unittest.TestCase):
	def setUp(self):
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		self.tree = tmp.name

	def files(self, *files):
		for name, text in files:
			with open(os.path.join(self.tree, name), "w", encoding="utf-8") as f:
				f.write(text)

	def read(self, name):
		with open(os.path.join(self.tree, name), encoding="utf-8") as f:
			return f.read()

	def test_writes_the_patch_s_go_and_modules_into_the_release_tree(self):
		pin = DOCKERFILE.format("1.27.1") + "FROM goreleaser/goreleaser:v2.18.3 AS packaging\n"
		self.files(("Dockerfile", pin), ("compose.yaml", COMPOSE.format("1.27.1")), ("go.mod", GO_MOD.format("1.22", "v0.29.0")), ("go.sum", "sum\n"))
		patch_release.prepare(self.tree, carrying(), carrying("1.27.2", "v2.19.0", NEW_GO_MOD, "new sum\n"))
		self.assertEqual(self.read("Dockerfile"), pin.replace("1.27.1", "1.27.2"))
		self.assertEqual(self.read("compose.yaml"), COMPOSE.format("1.27.2"))
		self.assertEqual((self.read("go.mod"), self.read("go.sum")), (NEW_GO_MOD, "new sum\n"))

	def test_leaves_what_the_patch_does_not_change_and_drops_an_emptied_go_sum(self):
		self.files(("go.mod", GO_MOD.format("1.22", "v0.29.0")), ("go.sum", "sum\n"))
		patch_release.prepare(self.tree, carrying(), carrying(go_mod=NEW_GO_MOD, go_sum=""))
		self.assertEqual(os.listdir(self.tree), ["go.mod"])


def report(*ids, module="stdlib", version="v1.27.1"):
	messages = [{"osv": {"id": i, "aliases": [f"CVE-{i[3:]}"], "summary": "A flaw"}} for i in ids]
	messages += [{"finding": {"osv": i, "trace": [{"module": module, "version": version}]}} for i in ids]
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
		os.makedirs("release-tooling")
		self.write(
			("Dockerfile", DOCKERFILE.format("1.27.1")), ("compose.yaml", COMPOSE.format("1.27.1")), ("CHANGELOG.md", RELEASED),
			("go.mod", GO_MOD.format("1.22", "v0.29.0")), ("release-tooling/goreleaser.yaml", GORELEASER),
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
		self.module_findings = {}
		self.scanned, self.outputs, self.notes_read = [], [], []
		self.published = mock.Mock(return_value=(0, "v0.1.1"))
		self.effects = patch_release.Effects(self.scan, self.ship, self.published, self.outputs.append, self.notes)

	def scan(self, tree, go, into):
		go_mod = self.read(os.path.join(tree, "go.mod"))
		self.scanned.append((go, go_mod))
		os.makedirs(into)
		with open(os.path.join(into, "report.json"), "w", encoding="utf-8") as f:
			f.write(self.reports[go] + "\n" + self.module_findings.get(go_mod, ""))
		return os.path.join(into, "report.json")

	def ship(self, tree, version):
		self.shipped = {
			name: self.read(os.path.join(tree, name)) for name in ("Dockerfile", "compose.yaml", "CHANGELOG.md", "go.mod")
			+ (("go.sum",) if os.path.exists(os.path.join(tree, "go.sum")) else ())
		}
		self.shipped_workflows = {
			name: self.read(os.path.join(tree, ".github/workflows", name))
			for name in sorted(os.listdir(os.path.join(tree, ".github/workflows")))
		}
		return "abc123"

	def notes(self, minor):
		self.notes_read.append(minor)
		return NOTES

	def cut(self):
		with redirect_stdout(io.StringIO()) as out:
			code = patch_release.cut(os.getcwd(), "v0.1.0", "2026-10-14", self.effects)
		return code, out.getvalue()

	def worktrees(self):
		return len(git("worktree", "list").splitlines())

	def test_cuts_a_patch_of_the_last_release_on_main_s_go(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")))
		code, _ = self.cut()
		self.assertEqual((code, self.outputs), (0, ["tag=v0.1.1"]))
		self.assertEqual(self.shipped["Dockerfile"], DOCKERFILE.format("1.27.2"))
		self.assertEqual(self.shipped["compose.yaml"], COMPOSE.format("1.27.2"))
		self.assertNotIn("go.sum", self.shipped)
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
		self.assertEqual(self.worktrees(), 1)

	def test_a_go_minor_names_the_systems_it_stops_supporting_as_breaking(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.28.0")))
		self.assertEqual(self.cut()[0], 0)
		self.assertEqual(self.notes_read, ["1.28"])
		self.assertIn(
			"## [0.1.1] - 2026-10-14\n\n### Breaking\n\n"
			"- Go 1.28 requires macOS 14 Sonoma or later, and so do the binaries.\n"
			"- Go 1.28 now requires at least Windows 10.1 or Windows Server 2019, and so do the binaries.\n\n"
			"### Changed\n\n- The binaries and image are built with Go 1.28.0.\n\n",
			self.shipped["CHANGELOG.md"],
		)

	def test_a_go_minor_whose_notes_it_cannot_read_is_refused(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.28.0")))
		self.effects = self.effects._replace(go_notes=lambda minor: NOTES.replace("now requires", "drops everything below"))
		with self.assertRaisesRegex(ValueError, "^Go 1.28's release notes say under Windows: "):
			self.cut()
		self.published.assert_not_called()

	def test_main_building_other_targets_than_the_release_is_refused(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")), ("release-tooling/goreleaser.yaml", GORELEASER.replace("windows", "freebsd")))
		with self.assertRaisesRegex(ValueError, "^main's release-tooling/goreleaser.yaml builds for other targets than v0.1.0"):
			self.cut()
		self.published.assert_not_called()

	def test_a_release_tree_with_no_goreleaser_yaml_is_not_compared(self):
		git("rm", "-q", "release-tooling/goreleaser.yaml")
		git("commit", "-qm", "Drop the config")
		git("tag", "-d", "v0.1.0")
		git("tag", "v0.1.0")
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")))
		os.makedirs("release-tooling", exist_ok=True)
		self.write(("release-tooling/goreleaser.yaml", GORELEASER))
		self.assertEqual(self.cut()[0], 0)

	def test_cuts_a_patch_carrying_main_s_modules_and_names_the_module_vulnerabilities_it_drops(self):
		self.write(("go.mod", NEW_GO_MOD), ("go.sum", "new sum\n"))
		self.reports["1.27.1"] = report()
		self.module_findings[GO_MOD.format("1.22", "v0.29.0")] = report("GO-2026-0003", module="golang.org/x/text", version="v0.29.0")
		self.assertEqual(self.cut()[0], 0)
		self.assertEqual((self.shipped["go.mod"], self.shipped["go.sum"]), (NEW_GO_MOD, "new sum\n"))
		self.assertEqual(self.shipped["Dockerfile"], DOCKERFILE.format("1.27.1"))
		self.assertEqual(self.scanned, [("1.27.1", GO_MOD.format("1.22", "v0.29.0")), ("1.27.1", NEW_GO_MOD)])
		self.assertIn("- The binaries carry golang.org/x/text v0.30.0, up from v0.29.0.\n", self.shipped["CHANGELOG.md"])
		self.assertIn(
			"### Security\n\n- The binaries no longer carry 0.1.0's known vulnerabilities: GO-2026-0003 (CVE-2026-0003).\n\n## [0.1.0]",
			self.shipped["CHANGELOG.md"],
		)

	def test_cuts_a_patch_carrying_main_s_goreleaser(self):
		self.write((".github/workflows/packages.yml", PACKAGES.format("v2.19.0")))
		self.reports["1.27.1"] = report()
		self.assertEqual(self.cut()[0], 0)
		self.assertEqual(self.shipped_workflows["packages.yml"], PACKAGES.format("v2.19.0"))
		self.assertIn(
			"### Changed\n\n- The binaries, packages and image are built with GoReleaser v2.19.0.\n\n## [0.1.0]",
			self.shipped["CHANGELOG.md"],
		)

	def test_a_patch_that_fails_the_gate_is_refused_and_leaves_no_worktree(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")))
		self.effects = self.effects._replace(ship=mock.Mock(side_effect=ValueError("`docker build` failed: vet")))
		with self.assertRaisesRegex(ValueError, "^`docker build` failed: vet$"):
			self.cut()
		self.published.assert_not_called()
		self.assertEqual(self.worktrees(), 1)

	def test_a_last_release_tag_the_checkout_lacks_is_refused(self):
		with self.assertRaisesRegex(ValueError, "^`git rev-parse` failed"):
			patch_release.cut(os.getcwd(), "v0.9.0", "2026-10-14", self.effects)

	def main_cut(self):
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")), ("output", ""))
		env = {"GITHUB_OUTPUT": os.path.abspath("output")}
		with mock.patch.object(patch_release, "last_release", return_value="v0.1.0"), \
			mock.patch.object(patch_release, "scan_tree", lambda root, tree, go, into: self.scan(tree, go, into)), \
			mock.patch.object(patch_release, "build_and_push", self.ship), \
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

	def test_a_patched_tree_still_vulnerable_in_the_release_s_modules_is_refused_and_leaves_no_worktree(self):
		held_back = NEW_GO_MOD.replace("example.com/new v0.1.0 // indirect", "example.com/new v0.1.0")
		self.write(("Dockerfile", DOCKERFILE.format("1.27.2")), ("go.mod", held_back))
		self.module_findings[GO_MOD.format("1.22", "v0.29.0")] = report("GO-2026-0003", module="golang.org/x/text", version="v0.29.0")
		with self.assertRaisesRegex(ValueError, "^the patched tree still carries a known vulnerability; "):
			self.cut()
		self.assertEqual(self.scanned[-1], ("1.27.2", GO_MOD.format("1.22", "v0.29.0")))
		self.published.assert_not_called()
		self.assertEqual(self.worktrees(), 1)


class Plan(Repository):
	def setUp(self):
		super().setUp()
		git("checkout", "-q", "v0.1.0")
		self.write(("CHANGELOG.md", changelog.insert_section(RELEASED, SECTION)))
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
		self.code = 0

	def run_stub(self, *argv, cwd=None, ok=(0,)):
		self.calls.append((argv, cwd))
		if self.code not in ok:
			raise ValueError(f"`{' '.join(argv[:2])}` failed: vet")
		return subprocess.CompletedProcess(argv, self.code, stdout="abc123\n", stderr="")

	def test_scan_tree_builds_the_tree_s_report_through_main_s_dockerfile(self):
		with mock.patch.object(patch_release, "run", self.run_stub):
			self.assertEqual(patch_release.scan_tree("/main", "/tree", "1.27.2", "/into"), "/into/report.json")
		self.assertEqual(
			self.calls,
			[((
				"docker", "build", "-f", "/main/Dockerfile", "--no-cache-filter", "vulnerabilities",
				"--build-arg", "GO_VERSION=1.27.2", "--target", "vulnerability-report", "--output", "/into", "/tree",
			), None)],
		)

	def test_build_and_push_builds_the_tree_then_commits_and_pushes_it_to_its_release_branch(self):
		env = {"FORGE_SERVER_URL": "https://forge.example", "FORGE_REPOSITORY": "owner/repo", "FORGE_TOKEN": "t0ken"}
		with mock.patch.object(patch_release, "run", self.run_stub), mock.patch.dict(os.environ, env):
			self.assertEqual(patch_release.build_and_push("/tree", "0.1.1"), "abc123")
		remote = "https://x-access-token:t0ken@forge.example/owner/repo.git"
		self.assertEqual(
			self.calls,
			[
				(("docker", "build", "."), "/tree"),
				(("git", "add", "-A"), "/tree"),
				(("git", "commit", "-qm", "Release 0.1.1"), "/tree"),
				(("git", "push", "-q", remote, ":refs/heads/release/0.1.1"), "/tree"),
				(("git", "push", "-q", remote, "HEAD:refs/heads/release/0.1.1"), "/tree"),
				(("git", "rev-parse", "HEAD"), "/tree"),
			],
		)

	def test_a_tree_that_fails_the_gate_is_neither_committed_nor_pushed(self):
		self.code = 1
		with mock.patch.object(patch_release, "run", self.run_stub), self.assertRaisesRegex(ValueError, "^`docker build` failed"):
			patch_release.build_and_push("/tree", "0.1.1")
		self.assertEqual(self.calls, [(("docker", "build", "."), "/tree")])


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
