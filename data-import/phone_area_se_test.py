import unittest

from geo_test import script

phone_area_se = script("phone-area-se")


class PageLines(unittest.TestCase):
    def test_reads_the_top_line_first_and_each_line_s_runs_left_to_right(self) -> None:
        content = (
            b"BT 1 0 0 1 120 700 Tm [(Wor) -3 (ld)] TJ ET\n"
            b"BT 1 0 0 1 50 700 Tm [(Hello )] TJ ET\n"
            b"BT 1 0 0 1 50 800.2 Tm [(Top \\(1\\))] TJ ET\n"
        )
        self.assertEqual(phone_area_se.page_lines(content), ["Top (1)", "Hello World"])


if __name__ == "__main__":
    unittest.main()
