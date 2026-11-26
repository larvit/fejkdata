"""Fake data for tests, rendered by the fejkdata CLI binary this package carries."""

import datetime
import os
import pathlib
import subprocess

__all__ = ["FejkdataError", "render"]

_BINARY = pathlib.Path(__file__).parent / "bin" / ("fejkdata.exe" if os.name == "nt" else "fejkdata")


class FejkdataError(Exception):
	"""A render the CLI refused. The message is the CLI's error text, and returncode the CLI's exit code."""

	def __init__(self, message: str, returncode: int) -> None:
		super().__init__(message)
		self.returncode = returncode


def render(template: str, *, seed: int, today: datetime.date | str, repeat: int = 1, format: str = "text") -> str:
	"""Renders template as `fejkdata --seed … --today … --repeat … --format …` does, and returns what it prints.

	today is a date, or a string such as 2026-01-15. A render the CLI refuses raises FejkdataError.
	"""
	day = today if isinstance(today, str) else today.strftime("%Y-%m-%d")
	done = subprocess.run(
		[_BINARY, "--seed", str(seed), "--today", day, "--repeat", str(repeat), "--format", format],
		input=template.encode(),
		capture_output=True,
		check=False,
	)
	if done.returncode != 0:
		raise FejkdataError(done.stderr.decode().strip().removesuffix("\ntry 'fejkdata --help'"), done.returncode)
	return done.stdout.decode()
