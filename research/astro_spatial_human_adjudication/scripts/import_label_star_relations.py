#!/usr/bin/env python3
"""Import relation attributes while enforcing frozen endpoints and completeness."""
from __future__ import annotations

import argparse
import csv
import datetime as dt
import json
import math
import xml.etree.ElementTree as ET
from pathlib import Path
from zipfile import ZipFile

from build_label_star_relations import FLAGS, PKG, digest

FIELDS = ["relation_candidate_id", "label_candidate_id", "object_candidate_id", "relation_type",
          "human_confidence", "reviewer_id", "review_timestamp", "notes"]
GEOMETRY_KEYS = {"xtl", "ytl", "xbr", "ybr", "cx", "cy", "rx", "ry", "rotation"}


def load(path: Path) -> ET.Element:
    if path.suffix.lower() == ".zip":
        with ZipFile(path) as archive:
            names = [name for name in archive.namelist() if name.endswith("annotations.xml")]
            if len(names) != 1: raise ValueError("expected exactly one annotations.xml")
            return ET.fromstring(archive.read(names[0]))
    return ET.parse(path).getroot()


def attributes(node: ET.Element) -> dict[str, str]:
    values = {a.attrib["name"]: a.text or "" for a in node.findall("attribute")}
    if len(values) != len(node.findall("attribute")): raise ValueError("duplicate attributes")
    return values


def frames(root: ET.Element) -> dict[str, ET.Element]:
    result = {}
    for frame in root.findall("image"):
        name = Path(frame.attrib["name"]).name
        if name in result: raise ValueError(f"duplicate frame: {name}")
        result[name] = frame
    return result


def validate(expected: ET.Element, reviewed: ET.Element, reviewer: str, timestamp: str, attachment: bool = False) -> list[dict[str, str]]:
    reference, actual = frames(expected), frames(reviewed)
    if set(reference) != set(actual):
        raise ValueError(f"frame mismatch: missing={sorted(set(reference)-set(actual))}, extra={sorted(set(actual)-set(reference))}")
    output = []
    seen = set()
    for filename, ref in reference.items():
        frame = actual[filename]
        if any(int(frame.attrib[k]) != int(ref.attrib[k]) for k in ("width", "height")):
            raise ValueError(f"frame dimensions changed: {filename}")
        ref_shapes = list(ref)
        shapes = list(frame)
        if len(shapes) != len(ref_shapes): raise ValueError(f"deleted or new shapes: {filename}")
        if not ref_shapes: continue
        by_role = {s.attrib.get("label", ""): s for s in shapes}
        if set(by_role) != {"LABEL_ENDPOINT", "STAR_ENDPOINT"} or len(by_role) != len(shapes):
            raise ValueError(f"endpoint class changed or duplicated: {filename}")
        for endpoint in ref_shapes:
            role = endpoint.attrib["label"]
            shape = by_role[role]
            if endpoint.tag != shape.tag: raise ValueError(f"endpoint shape type changed: {filename}/{role}")
            for key in GEOMETRY_KEYS:
                left, right = float(endpoint.attrib.get(key, 0)), float(shape.attrib.get(key, 0))
                delta = abs(left-right)
                if key == "rotation": delta = min(delta % 360, 360-delta % 360)
                # Real CVAT exports round coordinates/angles to two decimals.
                if not math.isfinite(right) or delta > .0051:
                    raise ValueError(f"frozen endpoint geometry changed: {filename}/{role}/{key}")
            expected_attrs, attrs = attributes(endpoint), attributes(shape)
            for key in ("relation_candidate_id", "label_candidate_id", "object_candidate_id", *(["panel"] if "panel" in expected_attrs else [])):
                if attrs.get(key) != expected_attrs[key]: raise ValueError(f"immutable endpoint ID changed: {filename}/{key}")
        attrs = attributes(by_role["STAR_ENDPOINT"])
        relation_id = attrs["relation_candidate_id"]
        if relation_id in seen: raise ValueError(f"duplicate relation candidate: {relation_id}")
        seen.add(relation_id)
        chosen = []
        for name, relation_type in ({} if attachment else FLAGS).items():
            value = attrs.get(name, "").lower()
            if value not in {"true", "false"}: raise ValueError(f"invalid flag: {relation_id}/{name}")
            if value == "true": chosen.append(relation_type)
        decision = attrs.get("relation_decision", "UNREVIEWED")
        if decision == "ASSIGNED":
            if not chosen and not attachment: raise ValueError(f"ASSIGNED without a spatial relation: {relation_id}")
            relation_type = "VISUAL_LABEL_OF" if attachment else ";".join(sorted(chosen))
        elif decision in {"UNASSIGNED", "UNCERTAIN"}:
            if chosen: raise ValueError(f"{decision} with selected flags: {relation_id}")
            relation_type = decision
        else:
            raise ValueError(f"unreviewed or invalid relation decision: {relation_id}/{decision}")
        confidence = attrs.get("human_confidence", "")
        if confidence not in {"LOW", "MEDIUM", "HIGH"}: raise ValueError(f"invalid confidence: {relation_id}")
        output.append({
            "relation_candidate_id": relation_id, "label_candidate_id": attrs["label_candidate_id"],
            "object_candidate_id": attrs["object_candidate_id"], "relation_type": relation_type,
            "human_confidence": confidence, "reviewer_id": reviewer, "review_timestamp": timestamp,
            "notes": attrs.get("notes", ""),
        })
    return sorted(output, key=lambda row: row["relation_candidate_id"])


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--reviewer-id", required=True)
    parser.add_argument("--timestamp", default="")
    parser.add_argument("--expected-xml", type=Path, default=PKG / "relations/cvat/label_star_relations_h1.xml")
    args = parser.parse_args()
    if args.output.exists(): raise ValueError("output already exists; use a new immutable export name")
    manifest_path = args.expected_xml.parent.parent / "RELATIONS_MANIFEST.json"
    manifest = json.loads(manifest_path.read_text())
    expected_key = str(args.expected_xml.relative_to(manifest_path.parent))
    if digest(args.expected_xml) != manifest["files"][expected_key]["sha256"]:
        raise ValueError("reference relation XML differs from the prepared package")
    for name, expected_digest in manifest["input_sha256"].items():
        if digest(PKG / name) != expected_digest:
            raise ValueError(f"frozen relation input changed: {name}")
    stamp = args.timestamp or dt.datetime.now(dt.timezone.utc).isoformat()
    rows = validate(load(args.expected_xml), load(args.input), args.reviewer_id, stamp)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, FIELDS, delimiter="\t", lineterminator="\n")
        writer.writeheader()
        writer.writerows(rows)
    print(f"Validated and imported {len(rows)} LABEL-STAR relation decisions")


if __name__ == "__main__":
    main()
