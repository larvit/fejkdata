#!/usr/bin/env python3
"""Publish the release that CHANGELOG.md's top heading names: tag SHA for every Go module, then the release.

A top heading of `[Unreleased]`, or a version already released, publishes nothing. A tag of the version
at another commit burns it, and publishes nothing either. Each tag and the release are made only where
missing, so a rerun finishes what an interrupted run began.
Env: FORGE_API_URL, FORGE_REPOSITORY (owner/repo), FORGE_TOKEN, SHA; GITHUB_OUTPUT, when set, gets
`tag=vX.Y.Z` for a version this run published.
"""

import json
import os
import re
import subprocess
import sys
import urllib.error
import urllib.request

HEADING = re.compile(r"^## \[([^\]]+)\].*?$\n?(.*?)(?=^## \[|\Z)", re.M | re.S)


def request(path: str, data: dict | None = None):
	req = urllib.request.Request(
		f"{os.environ['FORGE_API_URL']}/repos/{os.environ['FORGE_REPOSITORY']}/{path}",
		data=json.dumps(data).encode() if data else None,
		headers={"Authorization": f"token {os.environ['FORGE_TOKEN']}", "Content-Type": "application/json"},
	)
	with urllib.request.urlopen(req, timeout=30) as resp:
		return json.load(resp)


def top_heading(changelog: str) -> tuple[str, str] | None:
	"""The top `## [...]` heading's version and its section's body, or None."""
	top = HEADING.search(changelog)
	return (top.group(1), top.group(2).strip()) if top else None


def module_dirs() -> list[str]:
	"""The directory of every tracked go.mod, `.` for the root; a gate test holds go.work to the same set."""
	out = subprocess.run(["git", "ls-files", "-z", "--", "go.mod", "*/go.mod"], capture_output=True, check=True, text=True)
	return sorted(os.path.dirname(p) or "." for p in out.stdout.split("\0") if p)


def found(req, path: str):
	"""The object at path, or None where the forge answers 404."""
	try:
		return req(path)
	except urllib.error.HTTPError as e:
		if e.code != 404:
			raise
		return None


def tagged_at(req, tag: str) -> str | None:
	"""The commit `tag` points at, or None where no such tag exists."""
	ref = found(req, f"git/ref/tags/{tag}")
	if ref is None:
		return None
	obj = ref["object"]
	return req(f"git/tags/{obj['sha']}")["object"]["sha"] if obj["type"] == "tag" else obj["sha"]


def publish(version: str, body: str, dirs: list[str], sha: str, req) -> tuple[int, str | None]:
	"""Tag sha for each module in dirs, then publish the release; return the exit code and, when this run published it, the tag."""
	tag = f"v{version}"
	release = found(req, f"releases/tags/{tag}")
	if release is not None:
		print(f"{tag} already published: {release['html_url']}")
		return 0, None
	tags = [tag if d == "." else f"{d}/{tag}" for d in dirs]
	at = {t: tagged_at(req, t) for t in tags}
	burnt = [f"{t} exists at {c}, not {sha}" for t, c in at.items() if c not in (None, sha)]
	if burnt:
		print("; ".join(burnt) + "; the version is burnt, bump the heading", file=sys.stderr)
		return 1, None
	for t in tags:
		if at[t] is None:
			req("git/refs", {"ref": f"refs/tags/{t}", "sha": sha})
			print(f"tagged {t}")
	print(req("releases", {"body": body, "name": tag, "tag_name": tag, "target_commitish": sha})["html_url"])
	return 0, tag


def main() -> int:
	with open("CHANGELOG.md", encoding="utf-8") as f:
		top = top_heading(f.read())
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
	code, tag = publish(version, body, module_dirs(), os.environ["SHA"], request)
	if tag and "GITHUB_OUTPUT" in os.environ:
		with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as f:
			f.write(f"tag={tag}\n")
	return code


if __name__ == "__main__":
	sys.exit(main())
