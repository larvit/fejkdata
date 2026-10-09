#!/usr/bin/env python3
"""Publish the release that CHANGELOG.md's top heading names: tag SHA for every go.work module, then the release.

A top heading of `[Unreleased]` publishes nothing, and a tag already at another commit publishes nothing either.
Each tag and the release are made only where missing, so a rerun finishes what an interrupted run began.
Env: FORGE_API_URL, FORGE_REPOSITORY (owner/repo), FORGE_TOKEN, SHA; GITHUB_OUTPUT, when set, gets `tag=vX.Y.Z`.
"""

import json
import os
import re
import sys
import urllib.error
import urllib.request

from workspace import top_heading, workspace_dirs


def request(path: str, data: dict | None = None):
	req = urllib.request.Request(
		f"{os.environ['FORGE_API_URL']}/repos/{os.environ['FORGE_REPOSITORY']}/{path}",
		data=json.dumps(data).encode() if data else None,
		headers={"Authorization": f"token {os.environ['FORGE_TOKEN']}", "Content-Type": "application/json"},
	)
	with urllib.request.urlopen(req) as resp:
		return json.load(resp)


def found(path: str):
	"""The object at path, or None where the forge answers 404."""
	try:
		return request(path)
	except urllib.error.HTTPError as e:
		if e.code != 404:
			raise
		return None


def tagged_at(tag: str) -> str | None:
	"""The commit tag points at, or None where it does not exist."""
	ref = found(f"git/ref/tags/{tag}")
	if ref is None:
		return None
	obj = ref["object"]
	return request(f"git/tags/{obj['sha']}")["object"]["sha"] if obj["type"] == "tag" else obj["sha"]


def main() -> int:
	top = top_heading()
	if top is None:
		print("CHANGELOG.md has no `## [...]` heading", file=sys.stderr)
		return 1
	version, body = top
	if version == "Unreleased":
		print("top heading is Unreleased; nothing to publish")
		return 0
	if not re.fullmatch(r"\d+\.\d+\.\d+", version):
		print(f"top heading `[{version}]` is neither Unreleased nor X.Y.Z", file=sys.stderr)
		return 1
	tag, sha = f"v{version}", os.environ["SHA"]
	tags = [tag if d == "." else f"{d}/{tag}" for d in workspace_dirs()]
	at = {t: tagged_at(t) for t in tags}
	burnt = [f"{t} exists at {c}, not {sha}" for t, c in at.items() if c not in (None, sha)]
	if burnt:
		print("; ".join(burnt) + "; the version is burnt, bump the heading", file=sys.stderr)
		return 1
	for t in tags:
		if at[t] is None:
			request("git/refs", {"ref": f"refs/tags/{t}", "sha": sha})
			print(f"tagged {t}")
	release = found(f"releases/tags/{tag}")
	if release is None:
		release = request("releases", {"body": body, "name": tag, "tag_name": tag, "target_commitish": sha})
	print(release["html_url"])
	if "GITHUB_OUTPUT" in os.environ:
		with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as f:
			f.write(f"tag={tag}\n")
	return 0


if __name__ == "__main__":
	sys.exit(main())
