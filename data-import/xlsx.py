"""The rows of an xlsx sheet."""
import io
import xml.etree.ElementTree as ET
import zipfile

NS = {"m": "http://schemas.openxmlformats.org/spreadsheetml/2006/main", "r": "http://schemas.openxmlformats.org/officeDocument/2006/relationships"}


def rows(data, sheet=None):
    """The rows of an xlsx sheet named sheet, the first sheet by default, each a list of cell texts."""
    z = zipfile.ZipFile(io.BytesIO(data))
    strings = ["".join(t.text or "" for t in si.iter("{%s}t" % NS["m"])) for si in ET.fromstring(z.read("xl/sharedStrings.xml")).findall("m:si", NS)]
    rels = {r.get("Id"): r.get("Target") for r in ET.fromstring(z.read("xl/_rels/workbook.xml.rels"))}
    sheets = {s.get("name"): rels[s.get("{%s}id" % NS["r"])] for s in ET.fromstring(z.read("xl/workbook.xml")).iter("{%s}sheet" % NS["m"])}
    target = sheets[sheet] if sheet else next(iter(sheets.values()))
    for row in ET.fromstring(z.read("xl/" + target)).findall(".//m:row", NS):
        cells = []
        for c in row.findall("m:c", NS):
            v = c.find("m:v", NS)
            cells.append("" if v is None else strings[int(v.text)] if c.get("t") == "s" else v.text)
        yield cells
