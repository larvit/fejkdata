"""Fake data for tests, rendered by the fejkdata CLI binary this package carries."""

import datetime
import os
import pathlib
import subprocess
from collections.abc import Iterable

__all__ = ["FejkdataError", "render"]

_BINARY = pathlib.Path(__file__).parent / "bin" / ("fejkdata.exe" if os.name == "nt" else "fejkdata")


class FejkdataError(Exception):
	"""A render the CLI refused. The message is the CLI's error text, and returncode the CLI's exit code."""

	def __init__(self, message: str, returncode: int) -> None:
		super().__init__(message)
		self.returncode = returncode


def render(
	template: str,
	*,
	seed: int,
	today: datetime.date | str,
	repeat: int = 1,
	format: str = "text",
	table: str | None = None,
	data_path: str | os.PathLike[str] | Iterable[str | os.PathLike[str]] | None = None,
	no_shipped_data: bool = False,
	separator: str | None = None,
	null_text: str | None = None,
) -> str:
	"""Renders template as `fejkdata --seed … --today … --repeat … --format …` does, and returns what it prints.

	today is a date, or a string such as 2026-01-15. table, data_path, separator and null_text pass the CLI flag of their name
	unless they are None, and no_shipped_data passes its flag when True. data_path takes a folder or several, the last winning
	on a clash. A render the CLI refuses raises FejkdataError.
	"""
	day = today if isinstance(today, str) else today.strftime("%Y-%m-%d")
	folders = [data_path] if isinstance(data_path, (str, os.PathLike)) else data_path or ()
	flags = [f"--data-path={os.fspath(folder)}" for folder in folders]
	if no_shipped_data:
		flags.append("--no-shipped-data")
	for flag, value in (("table", table), ("separator", separator), ("null-text", null_text)):
		if value is not None:
			flags.append(f"--{flag}={value}")
	done = subprocess.run(
		[_BINARY, "--seed", str(seed), "--today", day, "--repeat", str(repeat), "--format", format, *flags],
		input=template.encode(),
		capture_output=True,
		check=False,
	)
	if done.returncode != 0:
		raise FejkdataError(done.stderr.decode().strip().removesuffix("\ntry 'fejkdata --help'"), done.returncode)
	return done.stdout.decode()
