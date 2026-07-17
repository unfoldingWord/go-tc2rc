#!/usr/bin/env python3
"""Fetch translationCore repos from DCS and clone them into a directory.

Pages through /api/v1/repos/search?metadataType=tc and shallow-clones each
repo's clone_url into <dest>/<owner>/<repo>. Existing clones are skipped.

Usage:
    python3 scripts/fetch_tc_repos.py --dest /tmp/tc_repos [--limit N] [--api-url URL]
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any

DEFAULT_API_URL = "https://git.door43.org/api/v1/repos/search"
USER_AGENT = "go-tc2rc-fetch/0.1"
PAGE_SIZE = 50


def fetch_page(api_url: str, page: int, timeout: float) -> list[dict[str, Any]]:
    params = urllib.parse.urlencode(
        {"metadataType": "tc", "limit": PAGE_SIZE, "page": page}
    )
    req = urllib.request.Request(
        f"{api_url}?{params}",
        headers={"Accept": "application/json", "User-Agent": USER_AGENT},
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        payload = json.loads(resp.read().decode("utf-8"))
    data = payload.get("data")
    if not isinstance(data, list):
        raise ValueError("search response is missing a list 'data' property")
    return [entry for entry in data if isinstance(entry, dict)]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dest", required=True, help="directory to clone repos into")
    parser.add_argument("--limit", type=int, default=0, help="max repos to clone (0 = all)")
    parser.add_argument("--api-url", default=DEFAULT_API_URL)
    parser.add_argument("--timeout", type=float, default=30.0)
    args = parser.parse_args()

    dest = Path(args.dest)
    dest.mkdir(parents=True, exist_ok=True)

    cloned = skipped = failed = 0
    page = 1
    while True:
        entries = fetch_page(args.api_url, page, args.timeout)
        if not entries:
            break
        for entry in entries:
            full_name = entry.get("full_name", "")
            clone_url = entry.get("clone_url", "")
            if not full_name or not clone_url:
                continue
            if args.limit and cloned + skipped >= args.limit:
                print(f"done: {cloned} cloned, {skipped} skipped, {failed} failed")
                return 1 if failed else 0
            target = dest / full_name
            if target.exists():
                skipped += 1
                continue
            target.parent.mkdir(parents=True, exist_ok=True)
            result = subprocess.run(
                ["git", "clone", "--quiet", "--depth", "1", clone_url, str(target)],
                capture_output=True,
                text=True,
            )
            if result.returncode == 0:
                cloned += 1
                print(f"cloned {full_name}")
            else:
                failed += 1
                print(f"FAILED {full_name}: {result.stderr.strip()}", file=sys.stderr)
        page += 1

    print(f"done: {cloned} cloned, {skipped} skipped, {failed} failed")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
