#!/usr/bin/env python3
"""Publish the release that CHANGELOG.md's top heading names, tagging SHA.

A top heading of `[Unreleased]`, or a version already released, publishes nothing.
Env: FORGE_API_URL, FORGE_REPOSITORY (owner/repo), FORGE_TOKEN, SHA.
"""

import json
import os
import re
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
	with urllib.request.urlopen(req) as resp:
		return json.load(resp)


def main() -> int:
	with open("CHANGELOG.md", encoding="utf-8") as f:
		top = HEADING.search(f.read())
	if top is None:
		print("CHANGELOG.md has no `## [...]` heading", file=sys.stderr)
		return 1
	version, body = top.group(1), top.group(2).strip()
	if version == "Unreleased":
		print("top heading is Unreleased; nothing to publish")
		return 0
	if not re.fullmatch(r"\d+\.\d+\.\d+", version):
		print(f"top heading `[{version}]` is neither Unreleased nor X.Y.Z", file=sys.stderr)
		return 1
	tag = f"v{version}"
	try:
		print(f"{tag} already published: {request(f'releases/tags/{tag}')['html_url']}")
		return 0
	except urllib.error.HTTPError as e:
		if e.code != 404:
			raise
	sha = os.environ["SHA"]
	try:
		at = request(f"tags/{tag}")["commit"]["sha"]
		if at != sha:
			print(f"{tag} exists at {at}, not {sha}; the version is burnt, bump the heading", file=sys.stderr)
			return 1
	except urllib.error.HTTPError as e:
		if e.code != 404:
			raise
	release = request("releases", {"body": body, "name": tag, "tag_name": tag, "target_commitish": sha})
	print(release["html_url"])
	return 0


if __name__ == "__main__":
	sys.exit(main())
