"""A source fetched once into the cache."""
import re
import sys
import time
import urllib.request
from pathlib import Path


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
