import contextlib
import datetime
import io
import pathlib
import re
import subprocess
import tempfile
import unittest
import unittest.mock

import fejkdata

DAY = "2026-01-15"


class Render(unittest.TestCase):
	def test_returns_what_the_cli_prints(self) -> None:
		self.assertEqual(fejkdata.render("{/misc.territory[SE].alpha2}!\n", seed=1, today=DAY), "SE!\n")

	def test_the_same_seed_and_day_render_the_same_text(self) -> None:
		first = fejkdata.render("{/sv_SE.person}", seed=7, today=DAY)
		self.assertEqual(fejkdata.render("{/sv_SE.person}", seed=7, today=DAY), first)
		self.assertNotEqual(fejkdata.render("{/sv_SE.person}", seed=8, today=DAY), first)

	def test_today_is_the_day_builtins_read(self) -> None:
		self.assertEqual(fejkdata.render("{date(today,today,'2006-01-02')}", seed=1, today=DAY), DAY)

	def test_today_takes_a_date(self) -> None:
		self.assertEqual(fejkdata.render("{date(today,today,'2006-01-02')}", seed=1, today=datetime.date(2025, 3, 9)), "2025-03-09")

	def test_today_takes_the_day_of_a_datetime(self) -> None:
		moment = datetime.datetime(2025, 3, 9, 10, 30)
		self.assertEqual(fejkdata.render("{date(today,today,'2006-01-02')}", seed=1, today=moment), "2025-03-09")

	def test_repeat_renders_the_template_that_many_times(self) -> None:
		self.assertEqual(fejkdata.render("{/misc.territory[SE].alpha2} ", seed=1, today=DAY, repeat=3), "SE SE SE ")

	def test_format_writes_records(self) -> None:
		template = '{"format": "", "code": "{/misc.territory[SE].alpha2}"}'
		self.assertEqual(fejkdata.render(template, seed=1, today=DAY, repeat=2, format="csv"), "code\nSE\nSE\n")

	def test_table_names_the_sql_target(self) -> None:
		template = '{"format": "", "code": "{/misc.territory[SE].alpha2}"}'
		self.assertIn("INSERT INTO countries", fejkdata.render(template, seed=1, today=DAY, format="sql", table="countries"))

	def test_table_names_the_pgcopy_target(self) -> None:
		template = '{"format": "", "code": "{/misc.territory[SE].alpha2}"}'
		self.assertIn("COPY countries", fejkdata.render(template, seed=1, today=DAY, format="pgcopy", table="countries"))

	def test_separator_goes_between_repeated_renders(self) -> None:
		self.assertEqual(fejkdata.render("x", seed=1, today=DAY, repeat=3, separator=","), "x,x,x")

	def test_a_separator_that_looks_like_a_flag_is_a_value(self) -> None:
		self.assertEqual(fejkdata.render("x", seed=1, today=DAY, repeat=2, separator="--"), "x--x")

	def test_an_empty_separator_is_the_default(self) -> None:
		self.assertEqual(fejkdata.render("x", seed=1, today=DAY, repeat=2, separator=""), "xx")

	def test_null_text_is_what_csv_writes_for_a_null(self) -> None:
		template = '{"format": "", "a": "x", "deleted_at": null}'
		self.assertEqual(fejkdata.render(template, seed=1, today=DAY, format="csv", null_text="NULL"), "a,deleted_at\nx,NULL\n")

	def test_a_flag_the_format_refuses_raises_the_cli_s_error(self) -> None:
		with self.assertRaises(fejkdata.FejkdataError) as raised:
			fejkdata.render("x", seed=1, today=DAY, null_text="NULL")
		self.assertEqual(raised.exception.returncode, 2)

	def test_data_path_adds_the_categories_of_a_folder(self) -> None:
		with tempfile.TemporaryDirectory() as folder:
			(pathlib.Path(folder) / "pets.json").write_text('["pixelfox"]', encoding="utf-8")
			self.assertEqual(fejkdata.render("{/pets}", seed=1, today=DAY, data_path=folder), "pixelfox")

	def test_data_path_takes_a_pathlib_path(self) -> None:
		with tempfile.TemporaryDirectory() as folder:
			(pathlib.Path(folder) / "pets.json").write_text('["pixelfox"]', encoding="utf-8")
			self.assertEqual(fejkdata.render("{/pets}", seed=1, today=DAY, data_path=pathlib.Path(folder)), "pixelfox")

	def test_data_path_takes_several_folders_and_the_last_wins(self) -> None:
		with tempfile.TemporaryDirectory() as team, tempfile.TemporaryDirectory() as project:
			(pathlib.Path(team) / "pets.json").write_text('["pixelfox"]', encoding="utf-8")
			(pathlib.Path(team) / "tools.json").write_text('["hammer"]', encoding="utf-8")
			(pathlib.Path(project) / "pets.json").write_text('["snork"]', encoding="utf-8")
			self.assertEqual(fejkdata.render("{/pets} {/tools}", seed=1, today=DAY, data_path=[team, pathlib.Path(project)]), "snork hammer")
			self.assertEqual(fejkdata.render("{/pets} {/tools}", seed=1, today=DAY, data_path=(project, team)), "pixelfox hammer")

	def test_no_shipped_data_loads_only_the_data_path_folders(self) -> None:
		with tempfile.TemporaryDirectory() as folder:
			(pathlib.Path(folder) / "pets.json").write_text('["pixelfox"]', encoding="utf-8")
			self.assertEqual(fejkdata.render("{/pets}", seed=1, today=DAY, data_path=folder, no_shipped_data=True), "pixelfox")
			with self.assertRaises(fejkdata.FejkdataError) as raised:
				fejkdata.render("{/misc.territory[SE].alpha2}", seed=1, today=DAY, data_path=folder, no_shipped_data=True)
			self.assertEqual(raised.exception.returncode, 2)

	def test_a_folder_the_cli_cannot_read_raises_its_error(self) -> None:
		with self.assertRaises(fejkdata.FejkdataError):
			fejkdata.render("x", seed=1, today=DAY, data_path="/nonexistent-folder-for-fejkdata")

	def test_a_flag_left_out_is_not_passed(self) -> None:
		with unittest.mock.patch.object(subprocess, "run", wraps=subprocess.run) as run:
			fejkdata.render("x", seed=1, today=DAY)
		self.assertEqual(
			run.call_args.args[0],
			[fejkdata._BINARY, "--seed", "1", "--today", DAY, "--repeat", "1", "--format", "text"],
		)

	def test_text_outside_ascii_comes_back_whole(self) -> None:
		self.assertEqual(fejkdata.render("Malmö {/misc.territory[SE].alpha2}", seed=1, today=DAY), "Malmö SE")

	def test_a_template_that_does_not_compile_raises_the_cli_s_error_text(self) -> None:
		with self.assertRaises(fejkdata.FejkdataError) as raised:
			fejkdata.render("{/nope.x}", seed=1, today=DAY)
		self.assertEqual(raised.exception.returncode, 2)
		self.assertEqual(str(raised.exception), 'fejkdata: template: reference {/nope.x}: no entry "nope"')

	def test_an_option_the_cli_refuses_raises_its_error_text(self) -> None:
		with self.assertRaises(fejkdata.FejkdataError) as raised:
			fejkdata.render("x", seed=-1, today=DAY)
		self.assertEqual(raised.exception.returncode, 2)
		self.assertEqual(str(raised.exception), 'fejkdata: --seed needs an unsigned integer, got "-1"')

	def test_a_runtime_error_raises_its_exit_code_and_its_text_without_the_newline(self) -> None:
		with tempfile.TemporaryDirectory() as folder:
			failing = pathlib.Path(folder) / "fejkdata"
			failing.write_text("#!/bin/sh\necho 'fejkdata: the disk is full' >&2\nexit 1\n", encoding="utf-8")
			failing.chmod(0o755)
			with unittest.mock.patch.object(fejkdata, "_BINARY", failing), self.assertRaises(fejkdata.FejkdataError) as raised:
				fejkdata.render("x", seed=1, today=DAY)
		self.assertEqual(raised.exception.returncode, 1)
		self.assertEqual(str(raised.exception), "fejkdata: the disk is full")


class ReadmeExamples(unittest.TestCase):
	"""Runs each `python` block of a README, and holds what it prints to the `text` block after it."""

	def run_blocks(self, markdown: str) -> int:
		names: dict[str, object] = {}
		blocks = re.findall(r"^```(\w+)\n(.*?)^```$", markdown, re.DOTALL | re.MULTILINE)
		ran = 0
		for (language, code), (after, printed) in zip(blocks, [*blocks[1:], ("", "")], strict=True):
			if language != "python":
				continue
			self.assertEqual(after, "text", code)
			with contextlib.redirect_stdout(io.StringIO()) as out:
				exec(code, names)  # noqa: S102 the code is the repository's own README
			self.assertEqual(out.getvalue(), printed, code)
			ran += 1
		return ran

	def test_the_package_s_readme(self) -> None:
		readme = pathlib.Path(__file__).with_name("README.md").read_text(encoding="utf-8")
		self.assertEqual(self.run_blocks(readme), 2)

	def test_the_repository_s_readme(self) -> None:
		readme = pathlib.Path(__file__).parent.with_name("README.md").read_text(encoding="utf-8")
		self.assertEqual(self.run_blocks(readme), 1)

	def test_a_python_block_with_no_text_block_after_it_fails(self) -> None:
		with self.assertRaises(AssertionError):
			self.run_blocks("```python\nprint(1)\n```\n\nIt prints 1.\n")

	def test_a_python_block_that_prints_something_else_fails(self) -> None:
		with self.assertRaises(AssertionError):
			self.run_blocks("```python\nprint(1)\n```\n\n```text\n2\n```\n")


if __name__ == "__main__":
	unittest.main()
