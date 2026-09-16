#!/usr/bin/env python3
"""Create the Gitea release for a tag, its body the tag's CHANGELOG.md section.

Env: GITEA_API_URL, GITEA_REPOSITORY (owner/repo), GITEA_TOKEN, TAG (vX.Y.Z).
"""

import json
import os
import re
import sys
import urllib.request


def section(changelog: str, version: str) -> str | None:
	m = re.search(rf"^## \[{re.escape(version)}\].*?$\n(.*?)(?=^## \[|\Z)", changelog, re.M | re.S)
	return m.group(1).strip() if m else None


def main() -> int:
	tag = os.environ["TAG"]
	if not re.fullmatch(r"v\d+\.\d+\.\d+", tag):
		print(f"{tag}: not a release tag; a release is vX.Y.Z", file=sys.stderr)
		return 1
	with open("CHANGELOG.md", encoding="utf-8") as f:
		body = section(f.read(), tag[1:])
	if body is None:
		print(f"CHANGELOG.md has no `## [{tag[1:]}]` heading; add the section, then tag", file=sys.stderr)
		return 1
	req = urllib.request.Request(
		f"{os.environ['GITEA_API_URL']}/repos/{os.environ['GITEA_REPOSITORY']}/releases",
		data=json.dumps({"body": body, "name": tag, "tag_name": tag}).encode(),
		headers={"Authorization": f"token {os.environ['GITEA_TOKEN']}", "Content-Type": "application/json"},
	)
	with urllib.request.urlopen(req) as resp:
		print(json.load(resp)["html_url"])
	return 0


if __name__ == "__main__":
	sys.exit(main())
