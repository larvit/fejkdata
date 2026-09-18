"""A source fetched once into the cache, and a table written as the loader admits it."""
import io
import re
import sys
import time
import urllib.request
import xml.etree.ElementTree as ET
import zipfile
from pathlib import Path

XLSX_NS = {"m": "http://schemas.openxmlformats.org/spreadsheetml/2006/main", "r": "http://schemas.openxmlformats.org/officeDocument/2006/relationships"}


def fetch(source, cache, name, magic=b"", data=None, headers=None):
    """The bytes of a URL, downloaded into cache/name once, or of a local file."""
    if not re.match(r"^https?://", source):
        return Path(source).read_bytes()
    path = Path(cache) / name
    for attempt in range(1, 6):
        if path.exists():
            return path.read_bytes()
        req = urllib.request.Request(source, data=data, headers={"User-Agent": "fejkdata data-import", **(headers or {})})
        try:
            with urllib.request.urlopen(req, timeout=600) as r:
                body = r.read()
        except OSError:
            body = b""
        if body and body.startswith(magic) and b"Request Rejected" not in body[:512]:
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(body)
        elif attempt < 5:
            time.sleep(10 * attempt)
    sys.exit(f"{source}: no valid download in 5 attempts")


def xlsx_rows(data, sheet=None):
    """The rows of an xlsx sheet named sheet, the first sheet by default, each a list of cell texts."""
    z = zipfile.ZipFile(io.BytesIO(data))
    strings = ["".join(t.text or "" for t in si.iter("{%s}t" % XLSX_NS["m"])) for si in ET.fromstring(z.read("xl/sharedStrings.xml")).findall("m:si", XLSX_NS)]
    rels = {r.get("Id"): r.get("Target") for r in ET.fromstring(z.read("xl/_rels/workbook.xml.rels"))}
    sheets = {s.get("name"): rels[s.get("{%s}id" % XLSX_NS["r"])] for s in ET.fromstring(z.read("xl/workbook.xml")).iter("{%s}sheet" % XLSX_NS["m"])}
    target = sheets[sheet] if sheet else next(iter(sheets.values()))
    for row in ET.fromstring(z.read("xl/" + target)).findall(".//m:row", XLSX_NS):
        cells = []
        for c in row.findall("m:c", XLSX_NS):
            v = c.find("m:v", XLSX_NS)
            cells.append("" if v is None else strings[int(v.text)] if c.get("t") == "s" else v.text)
        yield cells


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
