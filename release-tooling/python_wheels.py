"""Builds the Python package's wheels: one per platform GoReleaser built, each carrying that platform's binary."""

import base64
import hashlib
import pathlib
import re
import sys
import zipfile

# A wheel installs where pip finds one of its tags. The Linux binaries are static, so they run on glibc and musl alike.
# macOS 13 is the oldest Go 1.27 runs on (https://go.dev/doc/go1.27, Ports).
PLATFORMS = {
	"linux_amd64": "manylinux_2_17_x86_64.manylinux2014_x86_64.musllinux_1_1_x86_64",
	"linux_arm64": "manylinux_2_17_aarch64.manylinux2014_aarch64.musllinux_1_1_aarch64",
	"darwin_amd64": "macosx_13_0_x86_64",
	"darwin_arm64": "macosx_13_0_arm64",
	"windows_amd64": "win_amd64",
	"windows_arm64": "win_arm64",
}
LICENCES = ("LICENSE", "DATA-LICENSES.md", "data/en_US/WORDNET-LICENSE.txt")
# The oldest Python still supported (https://devguide.python.org/versions/) is 3.11.
METADATA = """\
Metadata-Version: 2.4
Name: fejkdata
Version: {version}
Summary: Fake data for tests. Related values match, like a postcode and its city.
Project-URL: Homepage, https://fejkdata.dev
Project-URL: Source, https://github.com/larvit/fejkdata
Project-URL: Changelog, https://github.com/larvit/fejkdata/blob/main/CHANGELOG.md
Author: Lilleman auf Larv
License-Expression: MIT
{licence_files}
Classifier: Programming Language :: Python :: 3
Classifier: Topic :: Software Development :: Testing
Classifier: Typing :: Typed
Requires-Python: >=3.11
Description-Content-Type: text/markdown

{readme}"""
WHEEL = "Wheel-Version: 1.0\nGenerator: fejkdata release-tooling\nRoot-Is-Purelib: false\n{tags}\n"


def binary(dist: pathlib.Path, platform: str) -> pathlib.Path:
	"""GoReleaser suffixes a build's folder with its CPU variant, as in fejkdata_linux_amd64_v1."""
	found = [path for path in dist.glob(f"fejkdata_{platform}_*/fejkdata*") if path.stem == "fejkdata"]
	if len(found) != 1:
		raise ValueError(f"{dist} holds {len(found)} binaries for {platform}, want one")
	return found[0]


def entry(name: str, executable: bool = False) -> zipfile.ZipInfo:
	"""A fixed date makes two builds of one release the same bytes."""
	info = zipfile.ZipInfo(name, (1980, 1, 1, 0, 0, 0))
	info.compress_type = zipfile.ZIP_DEFLATED
	info.external_attr = (0o100755 if executable else 0o100644) << 16
	return info


def record(name: str, data: bytes) -> str:
	digest = base64.urlsafe_b64encode(hashlib.sha256(data).digest()).rstrip(b"=").decode()
	return f"{name},sha256={digest},{len(data)}\n"


def files(src: pathlib.Path, version: str, tag: str) -> dict[str, bytes]:
	"""Every file of a wheel but the binary and RECORD, by its name in the wheel."""
	package = src / "python-package"
	info = f"fejkdata-{version}.dist-info"
	held = {f"fejkdata/{path.name}": path.read_bytes() for path in sorted((package / "fejkdata").iterdir()) if path.is_file()}
	licences = [pathlib.Path(name).name for name in LICENCES]
	held[f"{info}/METADATA"] = METADATA.format(
		version=version,
		licence_files="\n".join(f"License-File: {name}" for name in licences),
		readme=(package / "README.md").read_text(encoding="utf-8"),
	).encode()
	held[f"{info}/WHEEL"] = WHEEL.format(tags="\n".join(f"Tag: py3-none-{part}" for part in tag.split("."))).encode()
	for name in LICENCES:
		held[f"{info}/licenses/{pathlib.Path(name).name}"] = (src / name).read_bytes()
	return held


def build(src: pathlib.Path, dist: pathlib.Path, version: str, out: pathlib.Path) -> list[pathlib.Path]:
	"""Writes into out a wheel per platform, from the release's checkout src and GoReleaser's dist, and returns them."""
	if not re.fullmatch(r"\d+\.\d+\.\d+", version):
		raise ValueError(f"version {version!r} is not a release's, such as 0.9.0")
	binaries = {platform: binary(dist, platform) for platform in PLATFORMS}
	out.mkdir(parents=True, exist_ok=True)
	wheels = []
	for platform, tag in PLATFORMS.items():
		held = files(src, version, tag)
		held[f"fejkdata/bin/{binaries[platform].name}"] = binaries[platform].read_bytes()
		wheels.append(out / f"fejkdata-{version}-py3-none-{tag}.whl")
		with zipfile.ZipFile(wheels[-1], "w") as wheel:
			for name, data in held.items():
				wheel.writestr(entry(name, executable=name.startswith("fejkdata/bin/")), data)
			recorded = "".join(record(name, data) for name, data in held.items())
			wheel.writestr(entry(f"fejkdata-{version}.dist-info/RECORD"), recorded + f"fejkdata-{version}.dist-info/RECORD,,\n")
	return wheels


if __name__ == "__main__":
	src, dist, version, out = sys.argv[1:]
	for wheel in build(pathlib.Path(src), pathlib.Path(dist), version, pathlib.Path(out)):
		print(wheel)
