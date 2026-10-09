import io
import unittest
import urllib.error
from contextlib import redirect_stderr, redirect_stdout

from publish_release import publish, release_go_mods, top_heading

CORE = "github.com/larvit/fejkdata"
CLI = "github.com/larvit/fejkdata/cmd/fejkdata"
MODS = {
	".": (CORE, f"module {CORE}\n\ngo 1.22\n"),
	"cmd/fejkdata": (CLI, f"module {CLI}\n\ngo 1.22\n"),
}
PACKAGES = [
	(CORE, ["fmt", f"{CORE}/internal/grammar"]),
	(CLI, ["fmt", CORE, f"{CORE}/internal/grammar"]),
]
CLI_GO_MOD = f"module {CLI}\n\ngo 1.22\n\nrequire (\n\t{CORE} v0.1.0\n)\n"


class Forge:
	def __init__(self, refs=None, releases=None, annotated=None, commits=None):
		self.refs = dict(refs or {})
		self.releases = dict(releases or {})
		self.annotated = dict(annotated or {})
		self.commits = {"abc": {"tree": "t0", "parents": ["base"]}, **(commits or {})}
		self.posts = []
		self.trees = {}

	def __call__(self, path, data=None):
		if data is not None:
			self.posts.append(path)
			return self.post(path, data)
		kind, _, name = path.partition("/tags/")
		if kind == "git/ref" and name in self.refs:
			return {"object": {"type": "commit", "sha": self.refs[name]}}
		if kind == "git/ref" and name in self.annotated:
			return {"object": {"type": "tag", "sha": "tagobject"}}
		if path == "git/tags/tagobject":
			return {"object": {"sha": next(iter(self.annotated.values()))}}
		if kind == "releases" and name in self.releases:
			return self.releases[name]
		if path.startswith("git/commits/") and path.removeprefix("git/commits/") in self.commits:
			c = self.commits[path.removeprefix("git/commits/")]
			return {"tree": {"sha": c["tree"]}, "parents": [{"sha": p} for p in c["parents"]]}
		raise urllib.error.HTTPError(path, 404, "Not Found", {}, None)

	def post(self, path, data):
		if path == "git/refs":
			self.refs[data["ref"].removeprefix("refs/tags/")] = data["sha"]
			return {}
		if path == "git/trees":
			self.trees["t1"] = data
			return {"sha": "t1"}
		if path == "git/commits":
			self.commits["rel"] = {"tree": data["tree"], "parents": data["parents"]}
			return {"sha": "rel"}
		self.releases[data["tag_name"]] = {"html_url": f"https://forge/{data['tag_name']}", "target": data["target_commitish"]}
		return self.releases[data["tag_name"]]


def run(forge):
	with redirect_stdout(io.StringIO()), redirect_stderr(io.StringIO()):
		return publish("0.1.0", "notes", MODS, PACKAGES, "abc", forge)


class ReleaseGoMods(unittest.TestCase):
	def test_a_module_requires_the_workspace_modules_it_imports(self):
		self.assertEqual(release_go_mods(MODS, PACKAGES, "v0.1.0"), {"cmd/fejkdata/go.mod": CLI_GO_MOD})

	def test_a_module_whose_path_extends_another_is_its_own(self):
		packages = [(CORE, [f"{CLI}/x"]), (CLI, [f"{CLI}/x"])]
		self.assertEqual(release_go_mods(MODS, packages, "v0.1.0"), {"go.mod": f"module {CORE}\n\ngo 1.22\n\nrequire (\n\t{CLI} v0.1.0\n)\n"})


class Publish(unittest.TestCase):
	def test_fresh_release_commits_the_requires_then_tags_that_commit(self):
		forge = Forge()
		self.assertEqual(run(forge), (0, "v0.1.0"))
		self.assertEqual(forge.posts, ["git/trees", "git/commits", "git/refs", "git/refs", "releases"])
		self.assertEqual(forge.trees["t1"]["base_tree"], "t0")
		self.assertEqual(forge.trees["t1"]["tree"], [{"path": "cmd/fejkdata/go.mod", "mode": "100644", "type": "blob", "content": CLI_GO_MOD}])
		self.assertEqual(forge.commits["rel"]["parents"], ["abc"])
		self.assertEqual(forge.refs, {"v0.1.0": "rel", "cmd/fejkdata/v0.1.0": "rel"})
		self.assertEqual(forge.releases["v0.1.0"]["target"], "rel")

	def test_rerun_finishes_a_partial_release_on_its_release_commit(self):
		forge = Forge(refs={"v0.1.0": "c1"}, commits={"c1": {"tree": "t9", "parents": ["abc"]}})
		self.assertEqual(run(forge), (0, "v0.1.0"))
		self.assertEqual(forge.posts, ["git/refs", "releases"])
		self.assertEqual(forge.refs["cmd/fejkdata/v0.1.0"], "c1")

	def test_push_after_release_publishes_nothing(self):
		forge = Forge(refs={"v0.1.0": "old", "cmd/fejkdata/v0.1.0": "old"}, releases={"v0.1.0": {"html_url": "u"}})
		self.assertEqual(run(forge), (0, None))
		self.assertEqual(forge.posts, [])

	def test_tag_on_another_commits_release_creates_nothing(self):
		forge = Forge(refs={"cmd/fejkdata/v0.1.0": "zzz"}, commits={"zzz": {"tree": "t9", "parents": ["other"]}})
		self.assertEqual(run(forge), (1, None))
		self.assertEqual(forge.posts, [])

	def test_tags_on_two_commits_create_nothing(self):
		commits = {"c1": {"tree": "t", "parents": ["abc"]}, "c2": {"tree": "t", "parents": ["abc"]}}
		forge = Forge(refs={"v0.1.0": "c1", "cmd/fejkdata/v0.1.0": "c2"}, commits=commits)
		self.assertEqual(run(forge), (1, None))
		self.assertEqual(forge.posts, [])

	def test_annotated_tag_counts_at_its_commit(self):
		forge = Forge(annotated={"v0.1.0": "c1"}, commits={"c1": {"tree": "t9", "parents": ["abc"]}})
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
