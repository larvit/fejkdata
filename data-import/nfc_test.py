import unicodedata
import unittest
from pathlib import Path

DATA = Path(__file__).resolve().parent.parent / "data"


class ShippedText(unittest.TestCase):
    def test_every_shipped_value_is_in_nfc_so_a_length_counts_what_a_reader_sees(self):
        files = sorted(p for p in DATA.rglob("*") if p.suffix in (".json", ".tsv"))
        self.assertTrue(files, f"no shipped data under {DATA}")
        outside = []
        for p in files:
            text = p.read_text(encoding="utf-8")
            if not unicodedata.is_normalized("NFC", text):
                outside += [f"{p.relative_to(DATA)}:{n}: {line}" for n, line in enumerate(text.splitlines(), 1)
                            if not unicodedata.is_normalized("NFC", line)]
        self.assertEqual(outside, [])


if __name__ == "__main__":
    unittest.main()
