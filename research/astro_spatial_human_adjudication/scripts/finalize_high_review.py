#!/usr/bin/env python3
"""Merge a reviewed H-High export with frozen calibration and explicit overrides."""
from __future__ import annotations

import argparse
import csv
import json
from collections import Counter
from pathlib import Path


def read(path: Path) -> tuple[list[str], list[dict[str, str]]]:
    with path.open(encoding="utf-8", newline="") as handle:
        reader = csv.DictReader(handle, delimiter="\t")
        return list(reader.fieldnames or []), list(reader)


def indexed(rows: list[dict[str, str]], name: str) -> dict[str, dict[str, str]]:
    result = {row["candidate_id"]: row for row in rows}
    if len(result) != len(rows):
        raise ValueError(f"duplicate candidate_id in {name}")
    return result


def restore_candidate_geometry(row: dict[str, str], candidate: dict[str, str], task: str) -> None:
    for field in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2", "center_x", "center_y"):
        row[field] = candidate[field]
    if task == "label":
        row["rotation"] = "0.000000" if candidate["geometry_mode"] == "ELLIPSE_RING" else candidate["provisional_rotation"]
        row["geometry_type"] = "ELLIPSE" if candidate["geometry_mode"] == "ELLIPSE_RING" else "BOX"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--task", choices=("star", "label"), required=True)
    parser.add_argument("--review", type=Path, required=True)
    parser.add_argument("--calibration", type=Path, required=True)
    parser.add_argument("--candidates", type=Path, required=True)
    parser.add_argument("--overrides", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()

    fields, review_rows = read(args.review)
    calibration_fields, calibration_rows = read(args.calibration)
    _, candidate_rows = read(args.candidates)
    _, override_rows = read(args.overrides)
    if fields != calibration_fields:
        raise ValueError("review and calibration schemas differ")

    review = indexed(review_rows, "review")
    calibration = indexed(calibration_rows, "calibration")
    candidates_all = indexed(candidate_rows, "candidates")
    candidates = {
        candidate_id: row
        for candidate_id, row in candidates_all.items()
        if row["priority"] == "HIGH" and (args.task == "label" or row["candidate_type"] == "STAR_OBJECT")
    }
    if set(review) != set(candidates):
        missing = sorted(set(candidates) - set(review))
        extra = sorted(set(review) - set(candidates))
        raise ValueError(f"review/high candidate mismatch: missing={missing}, extra={extra}")

    overrides = indexed(
        [row for row in override_rows if row["task"].lower() == args.task],
        f"{args.task} overrides",
    )
    unknown_overrides = set(overrides) - set(candidates)
    if unknown_overrides:
        raise ValueError(f"unknown override candidate IDs: {sorted(unknown_overrides)}")

    output: list[dict[str, str]] = []
    calibration_ids = set(calibration) & set(candidates)
    for candidate_id in sorted(candidates, key=lambda value: (candidates[value]["panel"], value)):
        from_calibration = candidate_id in calibration_ids
        row = dict(calibration[candidate_id] if from_calibration else review[candidate_id])
        override = overrides.get(candidate_id)
        if override:
            row["human_decision"] = override["human_decision"]
            row["notes"] = override["notes"]

        geometry_action = override["geometry_action"] if override else ""
        if geometry_action not in {"", "KEEP_REVIEWED_GEOMETRY", "RESTORE_CANONICAL_GEOMETRY"}:
            raise ValueError(f"invalid geometry action for {candidate_id}: {geometry_action}")
        if geometry_action == "RESTORE_CANONICAL_GEOMETRY" or (
            not from_calibration and row["human_decision"] != "MODIFY" and geometry_action != "KEEP_REVIEWED_GEOMETRY"
        ):
            restore_candidate_geometry(row, candidates[candidate_id], args.task)

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
        "from_calibration": len(calibration_ids),
        "from_high_review": len(output) - len(calibration_ids),
        "overrides": len(overrides),
        "decisions": dict(sorted(Counter(row["human_decision"] for row in output).items())),
        "output": str(args.output),
    }, sort_keys=True))


if __name__ == "__main__":
    main()
