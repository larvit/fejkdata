import json
import subprocess
import unittest

import propose_plan

REPOSITORY = "owner/repo"
SHA = "a" * 40
OURS = {"id": 7, "conclusion": "action_required", "head_repository": {"full_name": REPOSITORY}}
FORK = {"id": 9, "conclusion": "action_required", "head_repository": {"full_name": "fork/repo"}}


class Forge:
	"""Stands in for git and gh, recording each command and answering as the test sets it up."""

	def __init__(self, changed=True, open_prs=(), runs=((OURS,),)):
		self.changed, self.open_prs, self.runs = changed, list(open_prs), list(runs)
		self.commands = []
		self.slept = 0

	def run(self, *argv, ok=(0,)):
		self.commands.append(argv)
		code, stdout = 0, ""
		if argv[:2] == ("git", "diff"):
			code = 1 if self.changed else 0
		elif argv[:3] == ("gh", "pr", "list"):
			stdout = json.dumps(self.open_prs)
		elif argv[:3] == ("gh", "pr", "create"):
			self.open_prs.append({"number": 1})
		elif argv[:2] == ("gh", "api") and argv[2] != "-X":
			stdout = json.dumps({"workflow_runs": list(self.runs.pop(0) if self.runs else ())})
		elif argv[:2] == ("git", "rev-parse"):
			stdout = SHA + "\n"
		return subprocess.CompletedProcess(argv, code, stdout, "")

	def sleep(self, _seconds):
		self.slept += 1

	def propose(self):
		propose_plan.propose(self.run, self.sleep, REPOSITORY, "token")

	def ran(self, *prefix):
		return [c for c in self.commands if c[: len(prefix)] == prefix]


class NoChange(unittest.TestCase):
	def test_closes_a_pull_request_still_open(self):
		forge = Forge(changed=False, open_prs=[{"number": 1}])
		forge.propose()
		self.assertEqual(len(forge.ran("gh", "pr", "close", "security-plan", "--delete-branch")), 1)
		self.assertEqual(forge.ran("git", "push"), [])

	def test_does_nothing_with_no_pull_request_open(self):
		forge = Forge(changed=False)
		forge.propose()
		self.assertEqual(forge.ran("gh", "pr", "close"), [])
		self.assertEqual(forge.ran("git", "push"), [])


class Change(unittest.TestCase):
	def test_pushes_opens_approves_and_merges(self):
		forge = Forge()
		forge.propose()
		self.assertEqual(forge.ran("git", "push"), [("git", "push", "-qf", f"https://x-access-token:token@github.com/{REPOSITORY}.git", "security-plan")])
		self.assertEqual(len(forge.ran("gh", "pr", "create")), 1)
		self.assertEqual(forge.ran("gh", "api", "-X"), [("gh", "api", "-X", "POST", f"repos/{REPOSITORY}/actions/runs/7/approve")])
		self.assertEqual(forge.ran("gh", "pr", "merge"), [("gh", "pr", "merge", "security-plan", "--auto", "--rebase")])

	def test_leaves_a_pull_request_already_open_and_refreshes_its_branch(self):
		forge = Forge(open_prs=[{"number": 1}])
		forge.propose()
		self.assertEqual(forge.ran("gh", "pr", "create"), [])
		self.assertEqual(len(forge.ran("git", "push")), 1)

	def test_approves_no_run_that_needs_no_approval(self):
		forge = Forge(runs=[[{**OURS, "conclusion": None}]])
		forge.propose()
		self.assertEqual(forge.ran("gh", "api", "-X"), [])
		self.assertEqual(len(forge.ran("gh", "pr", "merge")), 1)

	def test_asks_again_until_the_run_is_listed_and_ignores_a_forks(self):
		forge = Forge(runs=[[], [FORK], [FORK, OURS]])
		forge.propose()
		self.assertEqual(forge.slept, 2)
		self.assertEqual(len(forge.ran("gh", "api", "-X")), 1)

	def test_gives_up_when_no_run_is_ever_listed_and_merges_nothing(self):
		forge = Forge(runs=[])
		with self.assertRaisesRegex(ValueError, f"no pull_request run for {SHA}"):
			forge.propose()
		self.assertEqual(forge.slept, propose_plan.RUN_ATTEMPTS)
		self.assertEqual(forge.ran("gh", "pr", "merge"), [])


if __name__ == "__main__":
	unittest.main()
