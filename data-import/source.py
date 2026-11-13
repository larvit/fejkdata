"""A source fetched once into the cache."""
import re
import sys
import time
import urllib.request
from pathlib import Path


def fetch(source: str, cache: str | Path, name: str, magic: bytes = b"", data: bytes | None = None, headers: dict[str, str] | None = None) -> bytes:
    """The bytes of a URL, downloaded into cache/name once, or of a local file."""
    if not re.match(r"^https?://", source):
        return Path(source).read_bytes()
    path = Path(cache) / name
    if path.exists():
        return path.read_bytes()
    body = download(source, magic, data, headers)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(body)
    return body


def download(source: str, magic: bytes = b"", data: bytes | None = None, headers: dict[str, str] | None = None) -> bytes:
    """The bytes of a URL, in up to 5 attempts, each answer opening with magic."""
    for attempt in range(1, 6):
        req = urllib.request.Request(source, data=data, headers={"User-Agent": "fejkdata-data-import (https://github.com/larvit/fejkdata)", **(headers or {})})
        try:
            with urllib.request.urlopen(req, timeout=600) as r:
                body: bytes = r.read()
        except OSError:
            body = b""
        if body and body.startswith(magic) and b"Request Rejected" not in body[:512]:
            return body
        if attempt < 5:
            time.sleep(10 * attempt)
    sys.exit(f"{source}: no valid download in 5 attempts")
