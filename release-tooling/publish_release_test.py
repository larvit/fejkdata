import io
import os
import subprocess
import sys
import tempfile
import unittest
import urllib.error
from contextlib import redirect_stderr, redirect_stdout
from email.message import Message
from typing import Any, Callable
from unittest import mock

import publish_release

SHA = "a" * 40
OTHER = "b" * 40


def not_found(path: str) -> urllib.error.HTTPError:
	return urllib.error.HTTPError(path, 404, "Not Found", Message(), None)


def forge(responses: dict[str, Any], created: list[dict[str, Any]]) -> Callable[..., Any]:
	def request(path: str, data: dict[str, Any] | None = None) -> Any:
		if data is not None:
			created.append(data)
			return {"html_url": "https://forge.example/release"}
		response = responses.get(path)
		if response is None:
			raise not_found(path)
		return response

	return request


class TagGuard(unittest.TestCase):
	def setUp(self) -> None:
		cwd = os.getcwd()
		tmp = tempfile.TemporaryDirectory()
		self.addCleanup(tmp.cleanup)
		os.chdir(tmp.name)
		self.addCleanup(os.chdir, cwd)
		with open("CHANGELOG.md", "w", encoding="utf-8") as f:
			f.write("# Changelog\n\n## 1.2.3\n\n- A change.\n")
		env = mock.patch.dict(os.environ, {"SHA": SHA})
		env.start()
		self.addCleanup(env.stop)

	def publish(self, responses: dict[str, Any]) -> tuple[int, list[dict[str, Any]]]:
		created: list[dict[str, Any]] = []
		with mock.patch.object(publish_release, "request", forge(responses, created)), \
			redirect_stdout(io.StringIO()), redirect_stderr(io.StringIO()):
			with open("CHANGELOG.md", encoding="utf-8") as f:
				code, self.tag = publish_release.publish(f.read(), SHA)
		return code, created

	def test_a_tag_at_another_commit_publishes_nothing(self) -> None:
		code, created = self.publish({"git/ref/tags/v1.2.3": {"object": {"sha": OTHER, "type": "commit"}}})
		self.assertEqual((code, created), (1, []))

	def test_an_annotated_tag_at_another_commit_publishes_nothing(self) -> None:
		code, created = self.publish({
			"git/ref/tags/v1.2.3": {"object": {"sha": "c" * 40, "type": "tag"}},
			"git/tags/" + "c" * 40: {"object": {"sha": OTHER, "type": "commit"}},
		})
		self.assertEqual((code, created), (1, []))

	def test_an_annotated_tag_at_this_commit_publishes(self) -> None:
		code, created = self.publish({
			"git/ref/tags/v1.2.3": {"object": {"sha": "c" * 40, "type": "tag"}},
			"git/tags/" + "c" * 40: {"object": {"sha": SHA, "type": "commit"}},
		})
		self.assertEqual((code, [c["tag_name"] for c in created]), (0, ["v1.2.3"]))

	def test_no_tag_publishes_one_at_this_commit(self) -> None:
		code, created = self.publish({})
		self.assertEqual(code, 0)
		self.assertEqual(created, [{"body": "- A change.", "name": "v1.2.3", "tag_name": "v1.2.3", "target_commitish": SHA}])
		self.assertEqual(self.tag, "v1.2.3")

	def test_an_unreleased_heading_publishes_nothing(self) -> None:
		with open("CHANGELOG.md", "w", encoding="utf-8") as f:
			f.write("# Changelog\n\n## Unreleased\n\n- A change.\n")
		code, created = self.publish({})
		self.assertEqual((code, created, self.tag), (0, [], None))

	def test_a_bracketed_heading_of_an_older_release_publishes(self) -> None:
		with open("CHANGELOG.md", "w", encoding="utf-8") as f:
			f.write("# Changelog\n\n## [1.2.3] - 2026-11-05\n\n- A change.\n")
		code, created = self.publish({})
		self.assertEqual((code, [c["tag_name"] for c in created]), (0, ["v1.2.3"]))

	def test_a_dated_heading_publishes_with_its_body(self) -> None:
		with open("CHANGELOG.md", "w", encoding="utf-8") as f:
			f.write("# Changelog\n\n## 1.2.3 - 2026-11-20\n\n- A change.\n\n## 1.2.2 - 2026-11-05\n\n- Older.\n")
		code, created = self.publish({})
		self.assertEqual((code, [(c["tag_name"], c["body"]) for c in created]), (0, [("v1.2.3", "- A change.")]))

	def test_a_bracketed_unreleased_heading_publishes_nothing(self) -> None:
		with open("CHANGELOG.md", "w", encoding="utf-8") as f:
			f.write("# Changelog\n\n## [Unreleased]\n\n- A change.\n\n## [1.2.3] - 2026-11-05\n")
		code, created = self.publish({})
		self.assertEqual((code, created, self.tag), (0, [], None))

	def test_the_check_names_each_heading_it_cannot_read(self) -> None:
		changelog = (
			"# Changelog\n\n## [Unreleased\n\n## Unreleased (draft)\n\n## 0.5\n\n## [0.4.0] - soon\n\n"
			"## Unreleased\n\n### Added\n\n## [Unreleased]\n\n## 0.5.0 - 2026-11-20\n\n## [0.3.0] - 2026-11-05\n\n"
			"##\tUnreleased\n\n ## 0.5.0 - 2026-11-20\n\n    ## indented code\n\n## 0.5.0 - 2026-11-20 \n"
		)
		self.assertEqual(
			publish_release.unreadable_headings(changelog),
			["## [Unreleased", "## Unreleased (draft)", "## 0.5", "## [0.4.0] - soon", "##\tUnreleased", " ## 0.5.0 - 2026-11-20"],
		)

	def test_main_check_fails_on_an_unreadable_heading_and_names_it(self) -> None:
		for text, code, error in (
			("# Changelog\n\n## Unreleased\n\n## 1.2.3 - 2026-11-05\n", 0, ""),
			("# Changelog\n\n## Unrelesed\n", 1, "## Unrelesed"),
		):
			with self.subTest(text=text):
				with open("CHANGELOG.md", "w", encoding="utf-8") as f:
					f.write(text)
				stderr = io.StringIO()
				with redirect_stderr(stderr):
					self.assertEqual(publish_release.main(["check"]), code)
				self.assertIn(error, stderr.getvalue())

	def test_an_argument_other_than_check_publishes_nothing(self) -> None:
		created: list[dict[str, Any]] = []
		with mock.patch.object(publish_release, "request", forge({}, created)), redirect_stderr(io.StringIO()):
			self.assertEqual(publish_release.main(["--dry-run"]), 2)
		self.assertEqual(created, [])

	def test_the_script_run_as_ci_runs_it_checks_the_changelog_in_its_directory(self) -> None:
		for text, code in (("# Changelog\n\n## Unreleased\n", 0), ("# Changelog\n\n## Unrelesed\n", 1)):
			with self.subTest(text=text):
				with open("CHANGELOG.md", "w", encoding="utf-8") as f:
					f.write(text)
				run = subprocess.run([sys.executable, publish_release.__file__, "check"], capture_output=True, text=True)
				self.assertEqual((run.returncode, "Unrelesed" in run.stderr), (code, code == 1))

	def test_main_names_only_a_published_tag_to_ci(self) -> None:
		for responses, code, output in (
			({}, 0, "tag=v1.2.3\n"),
			({"releases/tags/v1.2.3": {"html_url": "https://forge.example/release"}}, 0, ""),
			({"git/ref/tags/v1.2.3": {"object": {"sha": OTHER, "type": "commit"}}}, 1, ""),
		):
			open("output", "w").close()
			created: list[dict[str, Any]] = []
			with mock.patch.dict(os.environ, {"GITHUB_OUTPUT": os.path.abspath("output")}), \
				mock.patch.object(publish_release, "request", forge(responses, created)), \
				redirect_stdout(io.StringIO()), redirect_stderr(io.StringIO()):
				self.assertEqual(publish_release.main([]), code)
			with open("output", encoding="utf-8") as f:
				self.assertEqual(f.read(), output)
			self.assertEqual([c["target_commitish"] for c in created], [SHA] if output else [])

	def test_a_version_cut_from_another_commit_fails_though_published(self) -> None:
		code, created = self.publish({
			"releases/tags/v1.2.3": {"html_url": "https://forge.example/release"},
			"git/ref/tags/v1.2.3": {"object": {"sha": OTHER, "type": "commit"}},
		})
		self.assertEqual((code, created, self.tag), (1, [], None))

	def test_a_heading_not_of_the_form_x_y_z_publishes_nothing(self) -> None:
		for name in ("1.2.3-rc.1", "1.2."):
			with self.subTest(name=name):
				with open("CHANGELOG.md", "w", encoding="utf-8") as f:
					f.write(f"# Changelog\n\n## {name}\n\n- A change.\n")
				code, created = self.publish({})
				self.assertEqual((code, created, self.tag), (1, [], None))

	def test_a_published_release_publishes_nothing(self) -> None:
		code, created = self.publish({"releases/tags/v1.2.3": {"html_url": "https://forge.example/release"}})
		self.assertEqual((code, created, self.tag), (0, [], None))


if __name__ == "__main__":
	unittest.main()
