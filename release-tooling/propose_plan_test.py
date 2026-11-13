import json
import os
import stat
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import patch_release
import propose_plan

REPOSITORY = "owner/repo"
SHA = "a" * 40
OURS = {"id": 7, "conclusion": "action_required", "head_repository": {"full_name": REPOSITORY}}
FORK = {"id": 9, "conclusion": "action_required", "head_repository": {"full_name": "fork/repo"}}
DELETED_FORK = {"id": 8, "conclusion": "action_required", "head_repository": None}
IS_OPEN = ("gh", "pr", "list", "--head", "security-plan", "--state", "open", "--json", "number")


def runs_path(sha):
	return f"repos/{REPOSITORY}/actions/runs?event=pull_request&head_sha={sha}"


class Forge:
	"""Stands in for git and gh, recording each command and answering only the queries the script should ask."""

	def __init__(self, changed=True, open_prs=(), runs=((OURS,),)):
		self.changed, self.open_prs, self.runs = changed, list(open_prs), list(runs)
		self.commands = []
		self.slept = []

	def run(self, *argv, ok=(0,)):
		self.commands.append(argv)
		code, stdout = 0, ""
		if argv[:2] == ("git", "diff"):
			code = 1 if self.changed else 0
		elif argv == IS_OPEN:
			stdout = json.dumps(self.open_prs)
		elif argv[:3] == ("gh", "pr", "create"):
			self.open_prs.append({"number": 1})
		elif argv == ("gh", "api", runs_path(SHA)):
			stdout = json.dumps({"workflow_runs": list(self.runs.pop(0) if self.runs else ())})
		elif argv == ("git", "rev-parse", "HEAD"):
			stdout = SHA + "\n"
		return subprocess.CompletedProcess(argv, code, stdout, "")

	def sleep(self, seconds):
		self.slept.append(seconds)

	def propose(self):
		propose_plan.propose(self.run, self.sleep, REPOSITORY, "token")

	def ran(self, *prefix):
		return [c for c in self.commands if c[: len(prefix)] == prefix]


class NoChange(unittest.TestCase):
	def test_closes_a_pull_request_still_open(self) -> None:
		forge = Forge(changed=False, open_prs=[{"number": 1}])
		forge.propose()
		self.assertEqual(forge.ran("gh", "pr", "close"), [("gh", "pr", "close", "security-plan", "--delete-branch", "--comment", "main already holds what this proposed.")])
		self.assertEqual(forge.ran("git", "push"), [])

	def test_does_nothing_with_no_pull_request_open(self) -> None:
		forge = Forge(changed=False)
		forge.propose()
		self.assertEqual(forge.ran("gh", "pr", "close"), [])
		self.assertEqual(forge.ran("git", "push"), [])


class Change(unittest.TestCase):
	def test_pushes_opens_approves_and_merges(self) -> None:
		forge = Forge()
		forge.propose()
		self.assertEqual(forge.ran("git", "push"), [("git", "push", "-qf", f"https://x-access-token:token@github.com/{REPOSITORY}.git", "security-plan")])
		self.assertEqual(forge.ran("gh", "pr", "create"), [("gh", "pr", "create", "--head", "security-plan", "--title", propose_plan.TITLE, "--body", propose_plan.BODY)])
		self.assertEqual(forge.ran("gh", "api", "-X"), [("gh", "api", "-X", "POST", f"repos/{REPOSITORY}/actions/runs/7/approve")])
		self.assertEqual(forge.ran("gh", "pr", "merge"), [("gh", "pr", "merge", "security-plan", "--auto", "--rebase")])

	def test_leaves_a_pull_request_already_open_and_refreshes_its_branch(self) -> None:
		forge = Forge(open_prs=[{"number": 1}])
		forge.propose()
		self.assertEqual(forge.ran("gh", "pr", "create"), [])
		self.assertEqual(len(forge.ran("git", "push")), 1)

	def test_approves_no_run_that_needs_no_approval(self) -> None:
		forge = Forge(runs=[[{**OURS, "conclusion": None}]])
		forge.propose()
		self.assertEqual(forge.ran("gh", "api", "-X"), [])
		self.assertEqual(len(forge.ran("gh", "pr", "merge")), 1)

	def test_asks_again_until_the_run_is_listed_and_ignores_a_forks(self) -> None:
		forge = Forge(runs=[[], [FORK], [FORK, DELETED_FORK, OURS]])
		forge.propose()
		self.assertEqual(forge.slept, [5, 5])
		self.assertEqual(forge.ran("gh", "api", "-X"), [("gh", "api", "-X", "POST", f"repos/{REPOSITORY}/actions/runs/7/approve")])

	def test_gives_up_when_no_run_is_ever_listed_and_merges_nothing(self) -> None:
		forge = Forge(runs=[])
		with self.assertRaisesRegex(ValueError, f"no pull_request run for {SHA}"):
			forge.propose()
		self.assertEqual(len(forge.slept), propose_plan.RUN_ATTEMPTS)
		self.assertEqual(forge.ran("gh", "pr", "merge"), [])


def git(*argv):
	return patch_release.run("git", *argv).stdout.strip()


class RealCheckout(unittest.TestCase):
	"""Runs propose's git commands for real, in a checkout whose gh and push are stood in for."""

	def setUp(self) -> None:
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		home = Path(tmp.name, "home")
		home.mkdir()
		patch = mock.patch.dict(os.environ, {"HOME": str(home)})
		patch.start()
		self.addCleanup(patch.stop)
		os.environ.pop("GIT_CONFIG_GLOBAL", None)
		checkout = Path(tmp.name, "checkout")
		checkout.mkdir()
		self.addCleanup(os.chdir, os.getcwd())
		os.chdir(checkout)
		git("init", "-q")
		for name in ("todo.md", "CHANGELOG.md", "other.md"):
			Path(name).write_text("first\n")
		git("add", ".")
		git("-c", "user.name=Author", "-c", "user.email=author@example.com", "commit", "-qm", "first")
		self.first = git("rev-parse", "HEAD")
		self.first_branch = git("branch", "--show-current")
		self.stood_in = []

	def run_forge(self, *argv, ok=(0,)):
		if argv[0] != "gh" and argv[:2] != ("git", "push"):
			return patch_release.run(*argv, ok=ok)
		self.stood_in.append(argv)
		stdout = ""
		if argv == IS_OPEN:
			stdout = "[]"
		elif argv[:2] == ("gh", "api") and argv[2] != "-X":
			stdout = json.dumps({"workflow_runs": [OURS]})
		return subprocess.CompletedProcess(argv, 0, stdout, "")

	def propose(self):
		propose_plan.propose(self.run_forge, lambda _seconds: None, REPOSITORY, "token")

	def assert_proposes(self, changed):
		Path(changed).write_text("second\n")
		self.propose()
		self.assertEqual(git("branch", "--show-current"), "security-plan")
		self.assertEqual(git("rev-parse", "HEAD~1"), self.first)
		self.assertEqual(git("show", "--name-only", "--format=", "HEAD"), changed)
		self.assertEqual(git("log", "-1", "--format=%s|%an|%ae"), f"{propose_plan.TITLE}|github-actions[bot]|41898282+github-actions[bot]@users.noreply.github.com")
		self.assertIn(("gh", "api", runs_path(git("rev-parse", "HEAD"))), self.stood_in)

	def test_proposes_a_change_to_the_changelog_alone(self) -> None:
		self.assert_proposes("CHANGELOG.md")

	def test_proposes_a_change_to_the_plan_alone(self) -> None:
		self.assert_proposes("todo.md")

	def test_a_change_to_any_other_file_proposes_nothing(self) -> None:
		Path("other.md").write_text("second\n")
		self.propose()
		self.assertEqual(git("rev-parse", "HEAD"), self.first)
		self.assertEqual(git("branch", "--show-current"), self.first_branch)
		self.assertEqual(self.stood_in, [IS_OPEN])


STAND_IN = """#!/bin/sh
echo "$0 $*" >> "$LOG"
[ -z "$GH_FAILS" ] || [ "${0##*/}" != gh ] || { echo "$GH_FAILS" >&2; exit 1; }
case "$1 $2" in
	"diff --quiet") exit 1 ;;
	"rev-parse HEAD") echo "$SHA" ;;
	"pr list") echo '[]' ;;
	"api repos/"*) echo "$RUNS" ;;
esac
"""


class Script(unittest.TestCase):
	"""Runs the script itself, with stand-ins for git and gh first on PATH."""

	def setUp(self) -> None:
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		self.log = Path(tmp.name, "log")
		self.env = {
			**os.environ,
			"PATH": f"{tmp.name}{os.pathsep}{os.environ['PATH']}",
			"LOG": str(self.log),
			"SHA": SHA,
			"RUNS": json.dumps({"workflow_runs": [OURS]}),
			"GITHUB_REPOSITORY": REPOSITORY,
			"GH_TOKEN": "secret",
		}
		for name in ("git", "gh"):
			stand_in = Path(tmp.name, name)
			stand_in.write_text(STAND_IN)
			stand_in.chmod(stand_in.stat().st_mode | stat.S_IXUSR)

	def script(self):
		return subprocess.run([sys.executable, str(Path(__file__).with_name("propose_plan.py"))], env=self.env, capture_output=True, text=True)

	def test_exits_0_after_proposing_with_the_repository_and_token_it_was_given(self) -> None:
		done = self.script()
		self.assertEqual(done.returncode, 0, done.stderr)
		self.assertIn(f"push -qf https://x-access-token:secret@github.com/{REPOSITORY}.git security-plan", self.log.read_text())
		self.assertIn("pr merge security-plan --auto --rebase", self.log.read_text())

	def test_exits_2_naming_the_step_it_cannot_complete(self) -> None:
		self.env["GH_FAILS"] = "refused"
		done = self.script()
		self.assertEqual(done.returncode, 2)
		self.assertEqual(done.stderr.strip(), "`gh pr` failed: refused")


if __name__ == "__main__":
	unittest.main()
