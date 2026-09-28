#!/usr/bin/env python3
"""Import reviewed CVAT XML into frozen project TSV without source mutation."""
from __future__ import annotations

import argparse
import csv
import datetime as dt
import xml.etree.ElementTree as ET
from pathlib import Path
from zipfile import ZipFile

STAR_FIELDS = ["candidate_id", "panel", "human_decision", "final_class", "bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2", "center_x", "center_y", "human_confidence", "reviewer_id", "review_timestamp", "notes"]
LABEL_FIELDS = STAR_FIELDS + ["rotation", "geometry_type", "split_required", "merge_required", "uncertain"]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--task", choices=("star", "label"), required=True)
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--reviewer-id", required=True)
    parser.add_argument("--timestamp", default="")
    parser.add_argument("--human-confidence-override", choices=("LOW", "MEDIUM", "HIGH"), default="", help="Explicit reviewer-directed override applied to every imported record")
    args = parser.parse_args()
    stamp = args.timestamp or dt.datetime.now(dt.timezone.utc).isoformat()
    if args.input.suffix.lower() == ".zip":
        with ZipFile(args.input) as archive:
            xml_names = [name for name in archive.namelist() if name.endswith("annotations.xml")]
            if len(xml_names) != 1:
                raise ValueError(f"expected exactly one annotations.xml in {args.input}, found {len(xml_names)}")
            root = ET.fromstring(archive.read(xml_names[0]))
    else:
        root = ET.parse(args.input).getroot()
    output = []
    seen = set()
    for image in root.findall("image"):
        panel = Path(image.attrib["name"]).stem
        shapes = list(image.findall("box")) + list(image.findall("ellipse"))
        for box in shapes:
            attrs = {a.attrib["name"]: (a.text or "") for a in box.findall("attribute")}
            candidate_id = attrs.get("candidate_id", "")
            if not candidate_id or candidate_id in seen:
                raise ValueError(f"missing or duplicate candidate_id: {candidate_id!r}")
            seen.add(candidate_id)
            if box.tag == "ellipse":
                cx, cy, rx, ry = (float(box.attrib[k]) for k in ("cx", "cy", "rx", "ry"))
                x1, y1, x2, y2 = cx - rx, cy - ry, cx + rx, cy + ry
            else:
                x1, y1, x2, y2 = (float(box.attrib[k]) for k in ("xtl", "ytl", "xbr", "ybr"))
            decision = attrs.get("decision", "UNCERTAIN")
            if decision not in {"ACCEPT", "REJECT", "MODIFY", "UNCERTAIN", "SPLIT", "MERGE"}:
                raise ValueError(f"invalid decision for {candidate_id}: {decision}")
            row = {
                "candidate_id": candidate_id, "panel": panel, "human_decision": decision,
                "final_class": box.attrib["label"], "bbox_x1": f"{x1:.6f}", "bbox_y1": f"{y1:.6f}", "bbox_x2": f"{x2:.6f}", "bbox_y2": f"{y2:.6f}",
                "center_x": f"{(x1 + x2) / 2:.6f}", "center_y": f"{(y1 + y2) / 2:.6f}",
                "human_confidence": args.human_confidence_override or attrs.get("human_confidence", "UNSET"), "reviewer_id": args.reviewer_id,
                "review_timestamp": stamp, "notes": attrs.get("notes", ""),
                "rotation": box.attrib.get("rotation", "0"), "geometry_type": box.tag.upper(),
                "split_required": "true" if decision == "SPLIT" or attrs.get("split_required", "false").lower() == "true" else "false",
                "merge_required": "true" if decision == "MERGE" or attrs.get("merge_required", "false").lower() == "true" else "false",
                # decision is authoritative; the legacy checkbox default must
                # not make every reviewed record uncertain.
                "uncertain": "true" if decision == "UNCERTAIN" else "false",
            }
            output.append(row)
    fields = STAR_FIELDS if args.task == "star" else LABEL_FIELDS
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        writer.writeheader()
        writer.writerows(sorted(output, key=lambda r: (r["panel"], r["candidate_id"])))


if __name__ == "__main__":
    main()
