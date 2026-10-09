#!/usr/bin/env python3
"""Publish the release that CHANGELOG.md's top heading names.

On main no module requires another, since go.work joins them. A release adds a commit on top of SHA that
writes each module's requires of the others at the version, tags that commit for every Go module, and
publishes the release on it.

A top heading of `[Unreleased]`, or a version already released, publishes nothing. A tag of the version
on any commit but a release commit of SHA burns it, and publishes nothing either. Each tag and the
release are made only where missing, so a rerun finishes what an interrupted run began.

Usage: publish_release.py PACKAGES, PACKAGES holding a line per package of every module, its module's
path and then its dependencies' import paths, as the release job's `go list` prints them.
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


def modules() -> dict[str, tuple[str, str]]:
	"""Every tracked go.mod's module path and text, by its directory, `.` for the root; a gate test holds go.work to the same set."""
	out = subprocess.run(["git", "ls-files", "-z", "--", "go.mod", "*/go.mod"], capture_output=True, check=True, text=True)
	mods = {}
	for p in sorted(p for p in out.stdout.split("\0") if p):
		with open(p, encoding="utf-8") as f:
			text = f.read()
		mods[os.path.dirname(p) or "."] = (re.search(r"(?m)^module\s+(\S+)", text).group(1), text)
	return mods


def release_go_mods(mods: dict[str, tuple[str, str]], packages: list[tuple[str, list[str]]], tag: str) -> dict[str, str]:
	"""The go.mod of every module importing another, by its path in the repository, requiring those modules at tag."""
	dirs = {path: d for d, (path, _) in mods.items()}
	needs: dict[str, set[str]] = {}
	for mod, deps in packages:
		for dep in deps:
			owner = max((p for p in dirs if dep == p or dep.startswith(p + "/")), key=len, default=None)
			if owner not in (None, mod):
				needs.setdefault(mod, set()).add(owner)
	files = {}
	for mod, owners in sorted(needs.items()):
		d = dirs[mod]
		lines = "".join(f"\t{o} {tag}\n" for o in sorted(owners))
		files["go.mod" if d == "." else f"{d}/go.mod"] = mods[d][1].rstrip("\n") + f"\n\nrequire (\n{lines})\n"
	return files


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


def release_commit(req, sha: str, files: dict[str, str], tag: str) -> str:
	"""A new commit on sha writing files."""
	entries = [{"path": p, "mode": "100644", "type": "blob", "content": c} for p, c in sorted(files.items())]
	tree = req("git/trees", {"base_tree": req(f"git/commits/{sha}")["tree"]["sha"], "tree": entries})["sha"]
	return req("git/commits", {"message": f"Release {tag}: require each module's dependencies at {tag}", "tree": tree, "parents": [sha]})["sha"]


def publish(version: str, body: str, mods, packages, sha: str, req) -> tuple[int, str | None]:
	"""Commit the requires on sha, tag that commit for every module and publish the release; return the exit code and, when this run published it, the tag."""
	tag = f"v{version}"
	release = found(req, f"releases/tags/{tag}")
	if release is not None:
		print(f"{tag} already published: {release['html_url']}")
		return 0, None
	tags = [tag if d == "." else f"{d}/{tag}" for d in mods]
	at = {t: tagged_at(req, t) for t in tags}
	tagged = sorted({c for c in at.values() if c})
	if len(tagged) > 1 or any([p["sha"] for p in req(f"git/commits/{c}")["parents"]] != [sha] for c in tagged):
		burnt = ", ".join(f"{t} at {c}" for t, c in at.items() if c)
		print(f"{burnt}: not one release commit of {sha}; the version is burnt, bump the heading", file=sys.stderr)
		return 1, None
	target = tagged[0] if tagged else release_commit(req, sha, release_go_mods(mods, packages, tag), tag)
	for t in tags:
		if at[t] is None:
			req("git/refs", {"ref": f"refs/tags/{t}", "sha": target})
			print(f"tagged {t}")
	print(req("releases", {"body": body, "name": tag, "tag_name": tag, "target_commitish": target})["html_url"])
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
	with open(sys.argv[1], encoding="utf-8") as f:
		packages = [(fields[0], fields[1:]) for fields in map(str.split, f) if fields]
	code, tag = publish(version, body, modules(), packages, os.environ["SHA"], request)
	if tag and "GITHUB_OUTPUT" in os.environ:
		with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as f:
			f.write(f"tag={tag}\n")
	return code


if __name__ == "__main__":
	sys.exit(main())
