import io
import os
import tempfile
import unittest
import urllib.error
from contextlib import redirect_stderr, redirect_stdout
from unittest import mock

import publish_release

SHA = "a" * 40
OTHER = "b" * 40


def not_found(path):
	return urllib.error.HTTPError(path, 404, "Not Found", None, None)


class TagGuard(unittest.TestCase):
	def setUp(self):
		cwd = os.getcwd()
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		os.chdir(tmp.name)
		self.addCleanup(os.chdir, cwd)
		with open("CHANGELOG.md", "w", encoding="utf-8") as f:
			f.write("# Changelog\n\n## [1.2.3]\n\n- A change.\n")
		env = mock.patch.dict(os.environ, {"SHA": SHA})
		env.start()
		self.addCleanup(env.stop)

	def publish(self, responses):
		created = []

		def request(path, data=None):
			if data is not None:
				created.append(data)
				return {"html_url": "https://forge.example/release"}
			response = responses.get(path)
			if response is None:
				raise not_found(path)
			return response

		with mock.patch.object(publish_release, "request", request), \
			redirect_stdout(io.StringIO()), redirect_stderr(io.StringIO()):
			code = publish_release.main()
		return code, created

	def test_a_tag_at_another_commit_publishes_nothing(self):
		code, created = self.publish({"git/ref/tags/v1.2.3": {"object": {"sha": OTHER, "type": "commit"}}})
		self.assertEqual((code, created), (1, []))

	def test_an_annotated_tag_at_another_commit_publishes_nothing(self):
		code, created = self.publish({
			"git/ref/tags/v1.2.3": {"object": {"sha": "c" * 40, "type": "tag"}},
			"git/tags/" + "c" * 40: {"object": {"sha": OTHER, "type": "commit"}},
		})
		self.assertEqual((code, created), (1, []))

	def test_an_annotated_tag_at_this_commit_publishes(self):
		code, created = self.publish({
			"git/ref/tags/v1.2.3": {"object": {"sha": "c" * 40, "type": "tag"}},
			"git/tags/" + "c" * 40: {"object": {"sha": SHA, "type": "commit"}},
		})
		self.assertEqual((code, [c["tag_name"] for c in created]), (0, ["v1.2.3"]))

	def test_no_tag_publishes_one_at_this_commit(self):
		code, created = self.publish({})
		self.assertEqual(code, 0)
		self.assertEqual(created, [{"body": "- A change.", "name": "v1.2.3", "tag_name": "v1.2.3", "target_commitish": SHA}])

	def test_a_published_release_publishes_nothing(self):
		code, created = self.publish({"releases/tags/v1.2.3": {"html_url": "https://forge.example/release"}})
		self.assertEqual((code, created), (0, []))


if __name__ == "__main__":
	unittest.main()
