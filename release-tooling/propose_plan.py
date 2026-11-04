#!/usr/bin/env python3
"""Open a pull request to main that holds what `patch_release.py plan` changed in todo.md and CHANGELOG.md, and merges itself.

With no change, closes a pull request still proposing one. Run in a checkout of main, with
GH_TOKEN and GITHUB_REPOSITORY set. Exits 2 on a step it cannot complete, naming it.
"""

import json
import os
import sys
import time
from typing import Callable

import patch_release

BRANCH = "security-plan"
TITLE = "Plan known vulnerabilities, and record fixed ones, the last release and new Go minors' Breaking lines"
BODY = (
	"Plans each module's new known vulnerabilities in todo.md, scored 0.0 until rank_todo.py scores them; "
	"moves fixed ones, any release section main lacks and each new Go minor's Breaking lines into CHANGELOG.md. "
	"Merges itself once the tests pass."
)
RUN_ATTEMPTS = 24
RUN_ATTEMPT_SECONDS = 5


def is_open(run: Callable) -> bool:
	found = run("gh", "pr", "list", "--head", BRANCH, "--state", "open", "--json", "number").stdout
	return json.loads(found) != []


def pull_request_run(run: Callable, sleep: Callable[[int], None], repository: str, sha: str) -> dict:
	"""The pull_request run of sha in this repository, which the forge may take a while to list."""
	path = f"repos/{repository}/actions/runs?event=pull_request&head_sha={sha}"
	for _ in range(RUN_ATTEMPTS):
		for found in json.loads(run("gh", "api", path).stdout)["workflow_runs"]:
			if (found["head_repository"] or {}).get("full_name") == repository:
				return found
		sleep(RUN_ATTEMPT_SECONDS)
	raise ValueError(f"no pull_request run for {sha}; its required check will never report")


def propose(run: Callable, sleep: Callable[[int], None], repository: str, token: str) -> None:
	if run("git", "diff", "--quiet", "todo.md", "CHANGELOG.md", ok=(0, 1)).returncode == 0:
		if is_open(run):
			run("gh", "pr", "close", BRANCH, "--delete-branch", "--comment", "main already holds what this proposed.")
		return
	run("git", "config", "--global", "user.name", "github-actions[bot]")
	run("git", "config", "--global", "user.email", "41898282+github-actions[bot]@users.noreply.github.com")
	run("git", "switch", "-qC", BRANCH)
	run("git", "commit", "-qam", TITLE)
	run("git", "push", "-qf", f"https://x-access-token:{token}@github.com/{repository}.git", BRANCH)
	if not is_open(run):
		run("gh", "pr", "create", "--head", BRANCH, "--title", TITLE, "--body", BODY)
	# GitHub holds the pull_request run of a pull request this token opens until that run is approved.
	found = pull_request_run(run, sleep, repository, run("git", "rev-parse", "HEAD").stdout.strip())
	if found["conclusion"] == "action_required":
		run("gh", "api", "-X", "POST", f"repos/{repository}/actions/runs/{found['id']}/approve")
	run("gh", "pr", "merge", BRANCH, "--auto", "--rebase")


def main() -> int:
	try:
		propose(patch_release.run, time.sleep, os.environ["GITHUB_REPOSITORY"], os.environ["GH_TOKEN"])
	except ValueError as err:
		print(err, file=sys.stderr)
		return 2
	return 0


if __name__ == "__main__":
	sys.exit(main())
