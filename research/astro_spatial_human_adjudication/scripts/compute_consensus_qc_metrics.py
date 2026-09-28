#!/usr/bin/env python3
"""Compute complete deterministic Consensus-QC metrics from frozen review files."""
from __future__ import annotations

import argparse
import csv
import json
import xml.etree.ElementTree as ET
from collections import Counter
from pathlib import Path


def read(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as handle:
        return list(csv.DictReader(handle, delimiter="\t"))


def sample_ids(path: Path) -> set[str]:
    root = ET.parse(path).getroot()
    return {
        node.text or ""
        for node in root.findall(".//attribute[@name='candidate_id']")
    }


def metrics(sample: set[str], paths: list[Path]) -> dict[str, object]:
    reviewed: dict[str, dict[str, str]] = {}
    for path in paths:
        for row in read(path):
            candidate_id = row["candidate_id"]
            if candidate_id in reviewed and reviewed[candidate_id]["human_decision"] != row["human_decision"]:
                raise ValueError(f"conflicting review decisions for {candidate_id}")
            reviewed[candidate_id] = row
    missing = sample - set(reviewed)
    if missing:
        raise ValueError(f"unreviewed Consensus-QC candidates: {sorted(missing)}")
    rows = [reviewed[candidate_id] for candidate_id in sample]
    decisions = Counter(row["human_decision"] for row in rows)
    total = len(rows)
    positive = decisions["ACCEPT"] + decisions["MODIFY"]
    return {
        "sample_records": total,
        "decision_counts": dict(sorted(decisions.items())),
        "existence_positive_rate": positive / total,
        "reject_rate": decisions["REJECT"] / total,
        "uncertain_rate": decisions["UNCERTAIN"] / total,
        "geometry_or_class_modification_rate": decisions["MODIFY"] / total,
    }


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--star-sample", type=Path, required=True)
    parser.add_argument("--label-sample", type=Path, required=True)
    parser.add_argument("--star-reviews", type=Path, nargs="+", required=True)
    parser.add_argument("--label-reviews", type=Path, nargs="+", required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    result = {
        "STAR": metrics(sample_ids(args.star_sample), args.star_reviews),
        "LABEL": metrics(sample_ids(args.label_sample), args.label_reviews),
        "note": "ACCEPT and MODIFY confirm existence; MODIFY also records a geometry or class correction; consensus is never promoted automatically",
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
