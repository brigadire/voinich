#!/usr/bin/env python3
"""Validate and normalize a reviewed pending Consensus-QC export."""
from __future__ import annotations

import argparse
import csv
import json
import xml.etree.ElementTree as ET
from collections import Counter
from pathlib import Path


def read(path: Path) -> tuple[list[str], list[dict[str, str]]]:
    with path.open(encoding="utf-8", newline="") as handle:
        reader = csv.DictReader(handle, delimiter="\t")
        return list(reader.fieldnames or []), list(reader)


def index(rows: list[dict[str, str]], name: str) -> dict[str, dict[str, str]]:
    result = {row["candidate_id"]: row for row in rows}
    if len(result) != len(rows):
        raise ValueError(f"duplicate candidate_id in {name}")
    return result


def restore_geometry(row: dict[str, str], candidate: dict[str, str], task: str) -> None:
    for field in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2", "center_x", "center_y"):
        row[field] = candidate[field]
    if task == "label":
        row["rotation"] = "0.000000" if candidate["geometry_mode"] == "ELLIPSE_RING" else candidate["provisional_rotation"]
        row["geometry_type"] = "ELLIPSE" if candidate["geometry_mode"] == "ELLIPSE_RING" else "BOX"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--task", choices=("star", "label"), required=True)
    parser.add_argument("--review", type=Path, required=True)
    parser.add_argument("--expected-xml", type=Path, required=True)
    parser.add_argument("--candidates", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()

    fields, review_rows = read(args.review)
    _, candidate_rows = read(args.candidates)
    review = index(review_rows, "review")
    candidates = index(candidate_rows, "candidates")
    expected_root = ET.parse(args.expected_xml).getroot()
    expected_ids = {
        node.text or ""
        for node in expected_root.findall(".//attribute[@name='candidate_id']")
    }
    if "" in expected_ids:
        raise ValueError("empty candidate_id in expected XML")
    if set(review) != expected_ids:
        raise ValueError(
            f"review/QC candidate mismatch: missing={sorted(expected_ids - set(review))}, "
            f"extra={sorted(set(review) - expected_ids)}"
        )
    unknown = expected_ids - set(candidates)
    if unknown:
        raise ValueError(f"unknown candidate IDs: {sorted(unknown)}")

    output: list[dict[str, str]] = []
    for candidate_id in sorted(review, key=lambda value: (review[value]["panel"], value)):
        row = dict(review[candidate_id])
        if row["human_decision"] != "MODIFY":
            restore_geometry(row, candidates[candidate_id], args.task)
        row["human_confidence"] = "HIGH"
        if args.task == "label":
            row["split_required"] = "true" if row["human_decision"] == "SPLIT" else "false"
            row["merge_required"] = "true" if row["human_decision"] == "MERGE" else "false"
            row["uncertain"] = "true" if row["human_decision"] == "UNCERTAIN" else "false"
        output.append(row)

    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fields, delimiter="\t", lineterminator="\n")
        writer.writeheader()
        writer.writerows(output)
    print(json.dumps({
        "task": args.task,
        "records": len(output),
        "decisions": dict(sorted(Counter(row["human_decision"] for row in output).items())),
        "output": str(args.output),
    }, sort_keys=True))


if __name__ == "__main__":
    main()
