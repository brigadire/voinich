#!/usr/bin/env python3
"""Union completed blind review phases without promoting unreviewed candidates."""
from __future__ import annotations

import argparse
import csv
import io
import json
import math
from collections import Counter
from pathlib import Path

from import_from_cvat import LABEL_FIELDS, STAR_FIELDS

PKG = Path(__file__).resolve().parents[1]


def read(path: Path) -> tuple[list[str], list[dict[str, str]]]:
    with path.open(encoding="utf-8", newline="") as handle:
        reader = csv.DictReader(handle, delimiter="\t")
        return list(reader.fieldnames or []), list(reader)


def render(fields: list[str], rows: list[dict[str, str]]) -> str:
    buffer = io.StringIO(newline="")
    writer = csv.DictWriter(buffer, fields, delimiter="\t", lineterminator="\n")
    writer.writeheader()
    writer.writerows(rows)
    return buffer.getvalue()


def build(package: Path) -> dict[str, str]:
    outputs = {}
    unreviewed = []
    for task, fields in (("STAR", STAR_FIELDS), ("LABEL", LABEL_FIELDS)):
        candidate_file = "HUMAN_CANDIDATE_OBJECTS.tsv" if task == "STAR" else "HUMAN_CANDIDATE_LABELS.tsv"
        _, candidate_rows = read(package / candidate_file)
        candidates = {
            row["candidate_id"]: row for row in candidate_rows
            if task == "LABEL" or row["candidate_type"] == "STAR_OBJECT"
        }
        calibration = "HUMAN_STAR_CALIBRATION_R01.tsv" if task == "STAR" else "HUMAN_LABEL_CALIBRATION_R01_FINAL.tsv"
        phase_files = [
            calibration,
            f"HUMAN_{task}_HIGH_R01_FINAL.tsv",
            f"HUMAN_{task}_CONSENSUS_QC_R01_FINAL.tsv",
            f"HUMAN_{task}_MEDIUM_R01.tsv",
        ]
        merged: dict[str, dict[str, str]] = {}
        for filename in phase_files:
            observed_fields, rows = read(package / "exports" / filename)
            if observed_fields != fields:
                raise ValueError(f"unexpected schema: {filename}")
            seen = set()
            for row in rows:
                cid = row["candidate_id"]
                if cid in seen or cid not in candidates:
                    raise ValueError(f"duplicate or unknown candidate in {filename}: {cid}")
                seen.add(cid)
                if row["panel"] != candidates[cid]["panel"]:
                    raise ValueError(f"panel mismatch: {cid}")
                if cid in merged and row != merged[cid]:
                    raise ValueError(f"conflicting frozen review records: {cid}")
                if row["human_confidence"] != "HIGH":
                    raise ValueError(f"unexpected confidence: {cid}")
                decisions = {"ACCEPT", "REJECT", "MODIFY", "UNCERTAIN"}
                if task == "LABEL":
                    decisions |= {"SPLIT", "MERGE"}
                if row["human_decision"] not in decisions:
                    raise ValueError(f"invalid decision: {cid}")
                valid_classes = {"STAR_OBJECT", "OTHER_OBJECT"} if task == "STAR" else {"LABEL"}
                if row["final_class"] not in valid_classes:
                    raise ValueError(f"invalid class: {cid}")
                coordinates = [float(row[field]) for field in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2", "center_x", "center_y")]
                if not all(math.isfinite(value) for value in coordinates):
                    raise ValueError(f"non-finite geometry: {cid}")
                x1, y1, x2, y2, cx, cy = coordinates
                if x2 <= x1 or y2 <= y1 or abs(cx - (x1+x2)/2) > .001 or abs(cy - (y1+y2)/2) > .001:
                    raise ValueError(f"invalid bbox or center: {cid}")
                merged[cid] = row
        remaining = [row for cid, row in candidates.items() if cid not in merged]
        if any(row["priority"] != "LOW" for row in remaining):
            raise ValueError(f"unreviewed HIGH/MEDIUM candidates remain for {task}")
        for row in remaining:
            unreviewed.append({
                "task": task, "candidate_id": row["candidate_id"], "panel": row["panel"],
                "priority": row["priority"], "status": "NOT_REVIEWED",
            })
        final = sorted(merged.values(), key=lambda row: (row["panel"], row["candidate_id"]))
        outputs[f"HUMAN_{task}_ADJUDICATION.tsv"] = render(fields, final)
        print(json.dumps({"task": task, "reviewed": len(final), "not_reviewed": len(remaining),
                          "decisions": dict(sorted(Counter(row["human_decision"] for row in final).items()))}, sort_keys=True))
    outputs["HUMAN_PRODUCTION_UNREVIEWED.tsv"] = render(
        ["task", "candidate_id", "panel", "priority", "status"],
        sorted(unreviewed, key=lambda row: (row["task"], row["panel"], row["candidate_id"])),
    )
    return outputs


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--package", type=Path, default=PKG)
    parser.add_argument("--output-dir", type=Path)
    args = parser.parse_args()
    output_dir = args.output_dir or args.package
    outputs = build(args.package)
    # Validate all existing targets before writing anything. Completed snapshots
    # may only be regenerated byte-identically, never silently overwritten.
    for filename, contents in outputs.items():
        path = output_dir / filename
        if path.exists() and path.read_text(encoding="utf-8").count("\n") > 1 and path.read_text(encoding="utf-8") != contents:
            raise ValueError(f"completed snapshot differs; create a new revision: {path}")
    output_dir.mkdir(parents=True, exist_ok=True)
    for filename, contents in outputs.items():
        (output_dir / filename).write_text(contents, encoding="utf-8")


if __name__ == "__main__":
    main()
