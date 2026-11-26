import contextlib
import datetime
import io
import pathlib
import re
import unittest

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
		self.assertEqual(fejkdata.render("{date(today,today,'2006-01-02')}", seed=1, today=datetime.date(2026, 1, 15)), DAY)

	def test_repeat_renders_the_template_that_many_times(self) -> None:
		self.assertEqual(fejkdata.render("{/misc.territory[SE].alpha2} ", seed=1, today=DAY, repeat=3), "SE SE SE ")

	def test_format_writes_records(self) -> None:
		template = '{"format": "", "code": "{/misc.territory[SE].alpha2}"}'
		self.assertEqual(fejkdata.render(template, seed=1, today=DAY, repeat=2, format="csv"), "code\nSE\nSE\n")

	def test_text_outside_ascii_comes_back_whole(self) -> None:
		self.assertEqual(fejkdata.render("Malmö {/misc.territory[SE].alpha2}", seed=1, today=DAY), "Malmö SE")

	def test_a_template_that_does_not_compile_raises_the_cli_s_error_text(self) -> None:
		with self.assertRaises(fejkdata.FejkdataError) as raised:
			fejkdata.render("{/nope.x}", seed=1, today=DAY)
		self.assertEqual(raised.exception.returncode, 2)
		self.assertIn('template: reference {/nope.x}: no entry "nope"', str(raised.exception))

	def test_an_option_the_cli_refuses_raises_its_error_text(self) -> None:
		with self.assertRaises(fejkdata.FejkdataError) as raised:
			fejkdata.render("x", seed=-1, today=DAY)
		self.assertEqual(raised.exception.returncode, 2)
		self.assertIn("--seed needs an unsigned integer", str(raised.exception))


class ReadmeExamples(unittest.TestCase):
	"""Runs each `python` block of a README, and holds what it prints to the `text` block after it."""

	def run_blocks(self, markdown: str) -> int:
		names: dict[str, object] = {}
		blocks = re.findall(r"```python\n(.*?)```\n(?:\n```text\n(.*?)```\n)?", markdown, re.DOTALL)
		for code, printed in blocks:
			with self.subTest(code), contextlib.redirect_stdout(io.StringIO()) as out:
				exec(code, names)
			if printed:
				self.assertEqual(out.getvalue(), printed, code)
		return len(blocks)

	def test_the_package_s_readme(self) -> None:
		readme = pathlib.Path(__file__).with_name("README.md").read_text(encoding="utf-8")
		self.assertEqual(self.run_blocks(readme), 2)

	def test_the_project_readme_s_python_section(self) -> None:
		readme = pathlib.Path(__file__).parent.with_name("README.md").read_text(encoding="utf-8")
		self.assertEqual(self.run_blocks(readme.split("\n### Python\n")[1].split("\n## ")[0]), 3)


if __name__ == "__main__":
	unittest.main()
