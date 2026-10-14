#!/usr/bin/env python3
"""Publish the release that CHANGELOG.md's top heading names, tagging SHA.

A top heading of `[Unreleased]`, or a version already released, publishes nothing.
Env: FORGE_API_URL, FORGE_REPOSITORY (owner/repo), FORGE_TOKEN, SHA, and GITHUB_OUTPUT,
to which it appends `tag=vX.Y.Z` when it publishes a release.
"""

import json
import os
import re
import sys
import urllib.error
import urllib.request

SECTION = re.compile(r"^## \[(?P<name>[^\]]+)\][^\n]*\n?(?P<body>.*?)(?=^## \[|\Z)", re.M | re.S)
"""A changelog section: its heading's name, Unreleased or a version, and its body."""
VERSION = re.compile(r"\d+\.\d+\.\d+")


def request(path: str, data: dict | None = None):
	req = urllib.request.Request(
		f"{os.environ['FORGE_API_URL']}/repos/{os.environ['FORGE_REPOSITORY']}/{path}",
		data=json.dumps(data).encode() if data else None,
		headers={"Authorization": f"token {os.environ['FORGE_TOKEN']}", "Content-Type": "application/json"},
	)
	with urllib.request.urlopen(req) as resp:
		return json.load(resp)


def publish(changelog: str, sha: str) -> tuple[int, str | None]:
	"""Publish the release the changelog's top heading names at the commit, returning the exit code and the tag it published, if any."""
	top = SECTION.search(changelog)
	if top is None:
		print("CHANGELOG.md has no `## [...]` heading", file=sys.stderr)
		return 1, None
	version, body = top["name"], top["body"].strip()
	if version == "Unreleased":
		print("top heading is Unreleased; nothing to publish")
		return 0, None
	if not VERSION.fullmatch(version):
		print(f"top heading `[{version}]` is neither Unreleased nor X.Y.Z", file=sys.stderr)
		return 1, None
	tag = f"v{version}"
	try:
		at = request(f"git/ref/tags/{tag}")["object"]
		if at["type"] == "tag":
			at = request(f"git/tags/{at['sha']}")["object"]
		if at["sha"] != sha:
			print(f"{tag} exists at {at['sha']}, not {sha}; the version is burnt, bump the heading", file=sys.stderr)
			return 1, None
	except urllib.error.HTTPError as e:
		if e.code != 404:
			raise
	try:
		print(f"{tag} already published: {request(f'releases/tags/{tag}')['html_url']}")
		return 0, None
	except urllib.error.HTTPError as e:
		if e.code != 404:
			raise
	release = request("releases", {"body": body, "name": tag, "tag_name": tag, "target_commitish": sha})
	print(release["html_url"])
	return 0, tag


def main() -> int:
	with open("CHANGELOG.md", encoding="utf-8") as f:
		code, tag = publish(f.read(), os.environ["SHA"])
	if tag:
		with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as f:
			f.write(f"tag={tag}\n")
	return code


if __name__ == "__main__":
	sys.exit(main())
