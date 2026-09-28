#!/usr/bin/env python3
"""Replace calibration records by candidate_id with a reviewed follow-up."""
from __future__ import annotations

import argparse
import csv
from pathlib import Path


def read(path: Path) -> tuple[list[str], list[dict[str, str]]]:
    with path.open(encoding="utf-8", newline="") as handle:
        reader = csv.DictReader(handle, delimiter="\t")
        return list(reader.fieldnames or []), list(reader)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base", type=Path, required=True)
    parser.add_argument("--followup", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    fields, base = read(args.base)
    follow_fields, followup = read(args.followup)
    if fields != follow_fields:
        raise ValueError("base and follow-up schemas differ")
    base_by_id = {row["candidate_id"]: row for row in base}
    if len(base_by_id) != len(base):
        raise ValueError("duplicate candidate_id in base")
    follow_by_id = {row["candidate_id"]: row for row in followup}
    if len(follow_by_id) != len(followup):
        raise ValueError("duplicate candidate_id in follow-up")
    unknown = set(follow_by_id) - set(base_by_id)
    if unknown:
        raise ValueError(f"follow-up contains unknown candidate IDs: {sorted(unknown)}")
    base_by_id.update(follow_by_id)
    output = sorted(base_by_id.values(), key=lambda row: (row["panel"], row["candidate_id"]))
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fields, delimiter="\t", lineterminator="\n")
        writer.writeheader()
        writer.writerows(output)


if __name__ == "__main__":
    main()
