import io
import unittest
import urllib.error
from contextlib import redirect_stderr, redirect_stdout

from publish_release import publish, top_heading


class Forge:
	def __init__(self, refs=None, releases=None, annotated=None):
		self.refs = dict(refs or {})
		self.releases = dict(releases or {})
		self.annotated = dict(annotated or {})
		self.posts = []

	def __call__(self, path, data=None):
		if data is not None:
			self.posts.append(path)
			if path == "git/refs":
				self.refs[data["ref"].removeprefix("refs/tags/")] = data["sha"]
				return {}
			self.releases[data["tag_name"]] = {"html_url": f"https://forge/{data['tag_name']}"}
			return self.releases[data["tag_name"]]
		kind, _, name = path.partition("/tags/")
		if kind == "git/ref" and name in self.refs:
			return {"object": {"type": "commit", "sha": self.refs[name]}}
		if kind == "git/ref" and name in self.annotated:
			return {"object": {"type": "tag", "sha": "tagobject"}}
		if path == "git/tags/tagobject":
			return {"object": {"sha": next(iter(self.annotated.values()))}}
		if kind == "releases" and name in self.releases:
			return self.releases[name]
		raise urllib.error.HTTPError(path, 404, "Not Found", {}, None)


def run(forge):
	with redirect_stdout(io.StringIO()), redirect_stderr(io.StringIO()):
		return publish("0.1.0", "notes", [".", "cmd/fejkdata"], "abc", forge)


class Publish(unittest.TestCase):
	def test_fresh_release_tags_every_module_then_releases(self):
		forge = Forge()
		self.assertEqual(run(forge), (0, "v0.1.0"))
		self.assertEqual(forge.refs, {"v0.1.0": "abc", "cmd/fejkdata/v0.1.0": "abc"})
		self.assertEqual(forge.posts, ["git/refs", "git/refs", "releases"])

	def test_rerun_finishes_a_partial_release(self):
		forge = Forge(refs={"v0.1.0": "abc"})
		self.assertEqual(run(forge), (0, "v0.1.0"))
		self.assertEqual(forge.posts, ["git/refs", "releases"])

	def test_push_after_release_publishes_nothing(self):
		forge = Forge(refs={"v0.1.0": "old", "cmd/fejkdata/v0.1.0": "old"}, releases={"v0.1.0": {"html_url": "u"}})
		self.assertEqual(run(forge), (0, None))
		self.assertEqual(forge.posts, [])

	def test_tag_at_another_commit_creates_nothing(self):
		forge = Forge(refs={"cmd/fejkdata/v0.1.0": "zzz"})
		self.assertEqual(run(forge), (1, None))
		self.assertEqual(forge.posts, [])

	def test_annotated_tag_counts_at_its_commit(self):
		forge = Forge(annotated={"v0.1.0": "abc"})
		self.assertEqual(run(forge), (0, "v0.1.0"))
		self.assertEqual(forge.posts, ["git/refs", "releases"])


class TopHeading(unittest.TestCase):
	def test_reads_the_top_section(self):
		text = "# Changelog\n\n## [0.2.0] - 2026-11-01\n\n- new\n\n## [0.1.0]\n\n- old\n"
		self.assertEqual(top_heading(text), ("0.2.0", "- new"))

	def test_unreleased(self):
		self.assertEqual(top_heading("## [Unreleased]\n\n- x\n"), ("Unreleased", "- x"))

	def test_none(self):
		self.assertIsNone(top_heading("# Changelog\n"))


if __name__ == "__main__":
	unittest.main()
