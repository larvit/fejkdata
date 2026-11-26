import base64
import hashlib
import pathlib
import stat
import tempfile
import unittest
import zipfile

import python_wheels

PLATFORMS = {
	"fejkdata_linux_amd64_v1/fejkdata": "manylinux_2_17_x86_64.manylinux2014_x86_64.musllinux_1_1_x86_64",
	"fejkdata_linux_arm64_v8.0/fejkdata": "manylinux_2_17_aarch64.manylinux2014_aarch64.musllinux_1_1_aarch64",
	"fejkdata_darwin_amd64_v1/fejkdata": "macosx_13_0_x86_64",
	"fejkdata_darwin_arm64_v8.0/fejkdata": "macosx_13_0_arm64",
	"fejkdata_windows_amd64_v1/fejkdata.exe": "win_amd64",
	"fejkdata_windows_arm64_v8.0/fejkdata.exe": "win_arm64",
}


class Build(unittest.TestCase):
	def setUp(self) -> None:
		folder = tempfile.TemporaryDirectory()
		self.addCleanup(folder.cleanup)
		self.root = pathlib.Path(folder.name)
		self.write("src/python-package/fejkdata/__init__.py", "def render(): ...\n")
		self.write("src/python-package/fejkdata/py.typed", "")
		self.write("src/python-package/fejkdata/__pycache__/stale.pyc", "stale")
		self.write("src/python-package/fejkdata/bin/fejkdata", "a developer's own build")
		self.write("src/python-package/README.md", "# fejkdata\n\nFake data.\n")
		self.write("src/python-package/render_test.py", "import fejkdata\n")
		self.write("src/LICENSE", "MIT\n")
		self.write("src/DATA-LICENSES.md", "# Data\n")
		self.write("src/data/en_US/WORDNET-LICENSE.txt", "WordNet\n")
		for binary in PLATFORMS:
			self.write(f"dist/{binary}", f"binary of {binary}")

	def write(self, name: str, text: str) -> None:
		path = self.root / name
		path.parent.mkdir(parents=True, exist_ok=True)
		path.write_text(text)

	def build(self, version: str = "0.9.0") -> list[pathlib.Path]:
		return python_wheels.build(self.root / "src", self.root / "dist", version, self.root / "wheels")

	def wheel(self, platform: str) -> zipfile.ZipFile:
		self.build()
		wheel = zipfile.ZipFile(self.root / "wheels" / f"fejkdata-0.9.0-py3-none-{platform}.whl")
		self.addCleanup(wheel.close)
		return wheel

	def test_writes_one_wheel_per_platform_goreleaser_builds(self) -> None:
		self.assertEqual(
			sorted(wheel.name for wheel in self.build()), sorted(f"fejkdata-0.9.0-py3-none-{tag}.whl" for tag in PLATFORMS.values())
		)

	def test_a_wheel_carries_its_platform_s_binary_as_an_executable(self) -> None:
		for binary, platform in PLATFORMS.items():
			with self.subTest(platform):
				wheel = self.wheel(platform)
				name = "fejkdata/bin/" + binary.split("/")[1]
				self.assertEqual(wheel.read(name), f"binary of {binary}".encode())
				self.assertTrue(wheel.getinfo(name).external_attr >> 16 & stat.S_IXOTH)

	def test_a_wheel_carries_the_package_and_nothing_else_from_its_folder(self) -> None:
		self.assertEqual(
			sorted(self.wheel("win_amd64").namelist()),
			[
				"fejkdata-0.9.0.dist-info/METADATA",
				"fejkdata-0.9.0.dist-info/RECORD",
				"fejkdata-0.9.0.dist-info/WHEEL",
				"fejkdata-0.9.0.dist-info/licenses/DATA-LICENSES.md",
				"fejkdata-0.9.0.dist-info/licenses/LICENSE",
				"fejkdata-0.9.0.dist-info/licenses/WORDNET-LICENSE.txt",
				"fejkdata/__init__.py",
				"fejkdata/bin/fejkdata.exe",
				"fejkdata/py.typed",
			],
		)

	def test_no_other_file_is_executable_and_each_is_compressed(self) -> None:
		for info in self.wheel("linux_amd64").infolist():
			with self.subTest(info.filename):
				self.assertEqual(info.compress_type, zipfile.ZIP_DEFLATED)
				if not info.filename.startswith("fejkdata/bin/"):
					self.assertEqual(stat.S_IMODE(info.external_attr >> 16), 0o644)

	def test_the_metadata_names_each_licence_file_and_the_oldest_python(self) -> None:
		metadata = self.wheel("win_amd64").read("fejkdata-0.9.0.dist-info/METADATA").decode()
		self.assertIn("\nLicense-Expression: MIT\nLicense-File: LICENSE\nLicense-File: DATA-LICENSES.md\nLicense-File: WORDNET-LICENSE.txt\n", metadata)
		self.assertIn("\nRequires-Python: >=3.11\n", metadata)

	def test_the_metadata_names_the_version_and_no_dependency(self) -> None:
		metadata = self.wheel("win_amd64").read("fejkdata-0.9.0.dist-info/METADATA").decode()
		self.assertIn("\nName: fejkdata\nVersion: 0.9.0\n", metadata)
		self.assertNotIn("Requires-Dist", metadata)
		self.assertTrue(metadata.endswith("\n\n# fejkdata\n\nFake data.\n"), metadata)

	def test_the_wheel_file_tags_each_platform_it_installs_on(self) -> None:
		wheel = self.wheel(PLATFORMS["fejkdata_linux_amd64_v1/fejkdata"]).read("fejkdata-0.9.0.dist-info/WHEEL").decode()
		self.assertIn("Root-Is-Purelib: false\n", wheel)
		self.assertEqual(
			[line for line in wheel.splitlines() if line.startswith("Tag: ")],
			["Tag: py3-none-manylinux_2_17_x86_64", "Tag: py3-none-manylinux2014_x86_64", "Tag: py3-none-musllinux_1_1_x86_64"],
		)

	def test_the_record_holds_every_other_file_s_sha256_and_size(self) -> None:
		wheel = self.wheel("win_amd64")
		recorded = dict(line.split(",", 1) for line in wheel.read("fejkdata-0.9.0.dist-info/RECORD").decode().splitlines())
		self.assertEqual(sorted(recorded), sorted(wheel.namelist()))
		self.assertEqual(recorded.pop("fejkdata-0.9.0.dist-info/RECORD"), ",")
		for name, digest_and_size in recorded.items():
			data = wheel.read(name)
			digest = base64.urlsafe_b64encode(hashlib.sha256(data).digest()).rstrip(b"=").decode()
			self.assertEqual(digest_and_size, f"sha256={digest},{len(data)}", name)

	def test_two_builds_write_the_same_bytes(self) -> None:
		first = [wheel.read_bytes() for wheel in self.build()]
		self.assertEqual([wheel.read_bytes() for wheel in self.build()], first)

	def test_a_version_pypi_would_not_order_as_a_release_is_refused(self) -> None:
		for version in ("v0.9.0", "0.9", "0.9.0-SNAPSHOT-1a2b3c4", ""):
			with self.subTest(version), self.assertRaisesRegex(ValueError, "version"):
				self.build(version)

	def test_a_dist_missing_a_platform_s_binary_is_refused(self) -> None:
		(self.root / "dist/fejkdata_windows_arm64_v8.0/fejkdata.exe").unlink()
		with self.assertRaisesRegex(ValueError, "windows_arm64"):
			self.build()

	def test_a_dist_holding_two_builds_of_one_platform_is_refused(self) -> None:
		self.write("dist/fejkdata_linux_amd64_v3/fejkdata", "another")
		with self.assertRaisesRegex(ValueError, "linux_amd64"):
			self.build()


class TheMacosTag(unittest.TestCase):
	def test_was_read_from_the_notes_of_the_go_minor_the_dockerfile_names(self) -> None:
		dockerfile = pathlib.Path(__file__).parent.with_name("Dockerfile").read_text(encoding="utf-8")
		self.assertIn(
			f"ARG GO_VERSION={python_wheels.MACOS_TAG_READ_FROM_GO}.",
			dockerfile.splitlines()[0],
			"the Go minor moved: read the oldest macOS its release notes name, and set the macosx tags and MACOS_TAG_READ_FROM_GO",
		)


if __name__ == "__main__":
	unittest.main()
