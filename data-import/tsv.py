"""A table written as the loader admits it."""
import re
import sys
from pathlib import Path


def write(path, columns, rows):
    """Write the rows as a TSV; every cell must be non-empty and free of tabs, newlines and braces."""
    lines = ["\t".join(columns)]
    for row in rows:
        cells = [str(row[c]) for c in columns]
        if not all(cells) or any(re.search(r"[\t\n{}]", c) for c in cells):
            raise ValueError(f"{path}: a cell is empty or holds a tab, newline or brace: {row}")
        lines.append("\t".join(cells))
    Path(path).write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"{path}: {len(lines) - 1} rows", file=sys.stderr)
