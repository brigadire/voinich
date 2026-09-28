#!/usr/bin/env python3
"""Export blind CVAT 1.1 image-annotation XML for STAR or LABEL review."""
from __future__ import annotations

import argparse
import csv
import hashlib
import json
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

PKG = Path(__file__).resolve().parents[1]
CAL_PANELS = {"f68r1", "f68r3", "f68v2"}


def rows(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as handle:
        return list(csv.DictReader(handle, delimiter="\t"))


def eligible(row: dict, phase: str) -> bool:
    if phase == "calibration":
        return row["panel"] in CAL_PANELS
    if phase == "calibration-followup":
        return row["panel"] in CAL_PANELS and row.get("geometry_mode") == "ELLIPSE_RING"
    if phase == "high":
        return row["priority"] == "HIGH"
    if phase == "medium":
        return row["priority"] == "MEDIUM"
    if phase == "consensus-qc":
        return row["support_ai1"] == row["support_ai2"] == "YES"
    return True


def add_attr(box: ET.Element, name: str, value: str) -> None:
    ET.SubElement(box, "attribute", {"name": name}).text = value


def add_meta_attribute(parent: ET.Element, name: str, mutable: bool, input_type: str, default: str, values: list[str]) -> None:
    attribute = ET.SubElement(parent, "attribute")
    ET.SubElement(attribute, "name").text = name
    ET.SubElement(attribute, "mutable").text = str(mutable)
    ET.SubElement(attribute, "input_type").text = input_type
    ET.SubElement(attribute, "default_value").text = default
    ET.SubElement(attribute, "values").text = "\n".join(values)


def add_meta_label(parent: ET.Element, name: str, task: str) -> None:
    label = ET.SubElement(parent, "label")
    ET.SubElement(label, "name").text = name
    ET.SubElement(label, "color").text = "#ff9800" if name == "STAR_OBJECT" else ("#607d8b" if name == "OTHER_OBJECT" else "#03a9f4")
    ET.SubElement(label, "type").text = "any" if task == "label" else "rectangle"
    attributes = ET.SubElement(label, "attributes")
    add_meta_attribute(attributes, "candidate_id", False, "text", "", [""])
    decisions = ["ACCEPT", "REJECT", "MODIFY", "UNCERTAIN"] + (["SPLIT", "MERGE"] if task == "label" else [])
    add_meta_attribute(attributes, "decision", True, "select", "UNCERTAIN", decisions)
    add_meta_attribute(attributes, "human_confidence", True, "select", "UNSET", ["UNSET", "LOW", "MEDIUM", "HIGH"])
    add_meta_attribute(attributes, "priority", False, "select", "UNSET", ["UNSET", "HIGH", "MEDIUM", "LOW"])
    if task == "label":
        add_meta_attribute(attributes, "split_required", True, "checkbox", "false", ["false"])
        add_meta_attribute(attributes, "merge_required", True, "checkbox", "false", ["false"])
        add_meta_attribute(attributes, "uncertain", True, "checkbox", "false", ["false"])
        add_meta_attribute(attributes, "notes", True, "text", "", [""])


def build(task: str, phase: str, pass_name: str, excluded_ids: set[str] | None = None) -> ET.ElementTree:
    manifest = {r["panel"]: r for r in rows(PKG / "IMAGE_MANIFEST.tsv")}
    source = PKG / ("HUMAN_CANDIDATE_OBJECTS.tsv" if task == "star" else "HUMAN_CANDIDATE_LABELS.tsv")
    excluded_ids = excluded_ids or set()
    candidates = [r for r in rows(source) if eligible(r, phase)]
    if task == "star":
        candidates = [r for r in candidates if r["candidate_type"] == "STAR_OBJECT"]
    if phase == "consensus-qc":
        candidates.sort(key=lambda r: (hashlib.sha256(r["candidate_id"].encode()).hexdigest(), r["candidate_id"]))
        candidates = candidates[:max(1, int(len(candidates) * 0.20))]
    candidates = [candidate for candidate in candidates if candidate["candidate_id"] not in excluded_ids]
    root = ET.Element("annotations")
    ET.SubElement(root, "version").text = "1.1"
    meta = ET.SubElement(root, "meta")
    task_meta = ET.SubElement(meta, "task")
    ET.SubElement(task_meta, "name").text = f"astronomical-{task}-{phase}-{pass_name.lower()}"
    ET.SubElement(task_meta, "size").text = str(len(manifest))
    ET.SubElement(task_meta, "mode").text = "annotation"
    ET.SubElement(task_meta, "overlap").text = "0"
    ET.SubElement(task_meta, "flipped").text = "False"
    labels_meta = ET.SubElement(task_meta, "labels")
    if task == "star":
        add_meta_label(labels_meta, "STAR_OBJECT", task)
        add_meta_label(labels_meta, "OTHER_OBJECT", task)
    else:
        add_meta_label(labels_meta, "LABEL", task)
    grouped = {panel: [] for panel in manifest}
    for candidate in candidates:
        grouped[candidate["panel"]].append(candidate)
    for index, panel in enumerate(manifest):
        info = manifest[panel]
        image = ET.SubElement(root, "image", {"id": str(index), "name": info["filename"], "width": info["width"], "height": info["height"]})
        for candidate in sorted(grouped[panel], key=lambda r: r["candidate_id"]):
            common = {"label": "STAR_OBJECT" if task == "star" else "LABEL", "source": "manual", "occluded": "0", "z_order": "0"}
            if task == "label" and candidate.get("geometry_mode") == "ELLIPSE_RING":
                x1, y1, x2, y2 = (float(candidate[k]) for k in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"))
                shape = ET.SubElement(image, "ellipse", common | {
                    "cx": f"{(x1 + x2) / 2:.6f}", "cy": f"{(y1 + y2) / 2:.6f}",
                    "rx": f"{(x2 - x1) / 2:.6f}", "ry": f"{(y2 - y1) / 2:.6f}", "rotation": "0.000000",
                })
            else:
                shape = ET.SubElement(image, "box", common | {
                    "xtl": candidate["bbox_x1"], "ytl": candidate["bbox_y1"], "xbr": candidate["bbox_x2"], "ybr": candidate["bbox_y2"],
                    "rotation": candidate.get("provisional_rotation", "0.000000") if task == "label" else "0.000000",
                })
            add_attr(shape, "candidate_id", candidate["candidate_id"])
            add_attr(shape, "decision", "UNCERTAIN")
            add_attr(shape, "human_confidence", "HIGH" if phase == "calibration-followup" else "UNSET")
            add_attr(shape, "priority", candidate["priority"] if pass_name == "H2" else "UNSET")
            if task == "label":
                add_attr(shape, "split_required", "false")
                add_attr(shape, "merge_required", "false")
                add_attr(shape, "uncertain", "false")
                add_attr(shape, "notes", "")
            # H1 is visually blind. H2 is a separate, post-freeze reconsideration export.
            if pass_name == "H2":
                add_attr(shape, "candidate_support", candidate["support_count"])
    return ET.ElementTree(root)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--task", choices=("star", "label"), required=True)
    parser.add_argument("--phase", choices=("calibration", "calibration-followup", "high", "consensus-qc", "medium", "all"), default="calibration")
    parser.add_argument("--pass-name", choices=("H1", "H2"), default="H1")
    parser.add_argument(
        "--exclude-reviewed",
        type=Path,
        nargs="*",
        default=[],
        help="review TSV files whose candidate IDs must be omitted",
    )
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    excluded_ids = {
        row["candidate_id"]
        for path in args.exclude_reviewed
        for row in rows(path)
    }
    tree = build(args.task, args.phase, args.pass_name, excluded_ids)
    ET.indent(tree, space="  ")
    args.output.parent.mkdir(parents=True, exist_ok=True)
    tree.write(args.output, encoding="utf-8", xml_declaration=True)
    zip_path = args.output.with_suffix(".zip")
    with zipfile.ZipFile(zip_path, "w", compression=zipfile.ZIP_DEFLATED) as archive:
        info = zipfile.ZipInfo("annotations.xml", date_time=(1980, 1, 1, 0, 0, 0))
        info.compress_type = zipfile.ZIP_DEFLATED
        info.external_attr = 0o644 << 16
        archive.writestr(info, args.output.read_bytes())
    print(json.dumps({
        "xml": str(args.output), "zip": str(zip_path), "task": args.task,
        "phase": args.phase, "pass": args.pass_name, "excluded_reviewed": len(excluded_ids),
    }))


if __name__ == "__main__":
    main()
