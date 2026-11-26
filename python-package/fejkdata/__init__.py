"""Fake data for tests, rendered by the fejkdata CLI binary this package carries."""

import datetime
import os
import pathlib
import subprocess

__all__ = ["FejkdataError", "render"]

_BINARY = pathlib.Path(__file__).parent / "bin" / ("fejkdata.exe" if os.name == "nt" else "fejkdata")


class FejkdataError(Exception):
	"""The CLI refused a render: its message is what the CLI printed, and returncode its exit code."""

	def __init__(self, message: str, returncode: int) -> None:
		super().__init__(message)
		self.returncode = returncode


def render(template: str, *, seed: int, today: datetime.date | str, repeat: int = 1, format: str = "text") -> str:
	"""Renders template as `fejkdata --seed … --today … --repeat … --format …` does, and returns what it prints.

	today is a date or its YYYY-MM-DD. A render the CLI refuses raises FejkdataError.
	"""
	done = subprocess.run(
		[_BINARY, "--seed", str(seed), "--today", str(today), "--repeat", str(repeat), "--format", format],
		input=template.encode(),
		capture_output=True,
		check=False,
	)
	if done.returncode != 0:
		raise FejkdataError(done.stderr.decode().strip(), done.returncode)
	return done.stdout.decode()
