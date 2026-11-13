"""A table written as the loader admits it."""
import re
import sys
from collections.abc import Collection, Iterable, Mapping, Sequence
from pathlib import Path
from typing import Any

# A person's name: parts of two or more letters, joined by a space, an apostrophe or a hyphen.
NAME = re.compile(r"^[^\W\d_]{2,}([ '-][^\W\d_]{2,})*$")

SELECTOR = '[]{}"|'  # mirrors inSelector in internal/grammar/path.go


def write(path: str | Path, columns: Sequence[str], rows: Iterable[Mapping[str, Any]], templates: Collection[str] = ()) -> None:
    """Write the rows as a TSV of a header and at least one row, each cell stripped, non-empty and free of tabs, newlines and braces; a column named in templates may hold braces."""
    lines = ["\t".join(columns)]
    for row in rows:
        cells = [str(row[c]).strip() for c in columns]
        if not all(cells) or any(re.search(r"[\t\n]" if c in templates else r"[\t\n{}]", cell) for c, cell in zip(columns, cells)):
            raise ValueError(f"{path}: a cell is empty or holds a tab, newline or brace: {row}")
        lines.append("\t".join(cells))
    if len(lines) < 2:
        raise ValueError(f"{path}: has no rows below its header; a table holds at least one row")
    Path(path).write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"{path}: {len(lines) - 1} rows", file=sys.stderr)
