#!/usr/bin/env python3
"""Import visual-caption attachment v2 without reinterpreting spatial v1 results."""
from __future__ import annotations

import argparse
import csv
import copy
import datetime as dt
import json
from pathlib import Path

from build_label_star_attachment_v2 import V2_FIELDS
from build_label_star_relations import FLAGS, PKG, digest
from import_label_star_relations import attributes, frames, load, validate as validate_endpoints
from export_to_cvat import add_attr


def validate(expected, reviewed, reviewer, timestamp, allow_unused_legacy_schema=False, reviewer_overrides=None):
    overrides = {}
    if reviewer_overrides:
        reviewed = copy.deepcopy(reviewed)
        endpoints = {attributes(s).get("relation_candidate_id"):s for s in reviewed.findall(".//*[@label='STAR_ENDPOINT']")}
        for override in reviewer_overrides:
            cid = override["relation_candidate_id"]
            if cid in overrides or cid not in endpoints: raise ValueError(f"duplicate or unknown reviewer override: {cid}")
            if override["reviewer_id"]!=reviewer: raise ValueError("override reviewer differs from import reviewer")
            decision = override["relation_decision"]
            if decision not in {"ASSIGNED","UNASSIGNED","UNCERTAIN"}: raise ValueError("invalid reviewer override decision")
            stamp = dt.datetime.fromisoformat(override["review_timestamp"].replace("Z","+00:00"))
            if stamp.utcoffset()!=dt.timedelta(0): raise ValueError("override timestamp must be UTC")
            node = endpoints[cid].find("attribute[@name='relation_decision']")
            if node is None or node.text!="UNREVIEWED": raise ValueError("override may only complete an unreviewed raw pair; historical corrections need a new revision")
            node.text = decision
            overrides[cid] = override
    adapted = False
    if allow_unused_legacy_schema:
        reviewed = copy.deepcopy(reviewed)
        reference = frames(expected)
        for filename, frame in frames(reviewed).items():
            ref = reference.get(filename)
            if ref is None: continue  # strict endpoint validation rejects extra frames
            by_role = {s.attrib.get("label"):s for s in ref}
            for endpoint in frame:
                role = endpoint.attrib.get("label")
                if role not in by_role: continue
                attrs = attributes(endpoint)
                legacy = set(attrs) & set(FLAGS)
                if any(attrs[name].lower() != "false" for name in legacy):
                    raise ValueError("active or invalid spatial-v1 flags cannot be adapted as attachment-v2 decisions")
                for node in list(endpoint.findall("attribute")):
                    if node.attrib["name"] in legacy: endpoint.remove(node)
                if legacy: adapted = True
                if "panel" not in attrs:
                    add_attr(endpoint,"panel",attributes(by_role[role])["panel"])
                    adapted = True
    for endpoint in reviewed.findall(".//*[@label='STAR_ENDPOINT']"):
        if set(attributes(endpoint)) & set(FLAGS):
            raise ValueError("spatial-v1 checkbox attributes are not attachment-v2 decisions; create a new task with the v2 schema")
    rows = validate_endpoints(expected, reviewed, reviewer, timestamp, attachment=True)
    for row in rows:
        row["relation_protocol_version"] = "2"
        row["decision_origin"] = "PAIR_REVIEW_LEGACY_SCHEMA_ADAPTED" if adapted else "PAIR_REVIEW"
        if row["relation_candidate_id"] in overrides:
            override = overrides[row["relation_candidate_id"]]
            row["decision_origin"] = "REVIEWER_POST_EXPORT_CONFIRMATION"
            row["review_timestamp"] = override["review_timestamp"]
            row["notes"] = "; ".join(value for value in (row["notes"],override.get("notes","")) if value)
    return rows


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--reviewer-id", required=True)
    parser.add_argument("--subset", choices=("main","f68v2"), default="main")
    parser.add_argument("--timestamp", default="")
    parser.add_argument("--reviewer-overrides", type=Path,
                        help="Explicit reviewer confirmations for otherwise UNREVIEWED raw pairs; applied to a copy and audited in output")
    parser.add_argument("--allow-unused-legacy-schema", action="store_true",
                        help="Explicitly discard only false legacy spatial flags and derive missing panel from the frozen frame mapping; IDs/geometry/decisions remain strict")
    parser.add_argument("--package", type=Path, default=PKG / "relations_v2")
    args = parser.parse_args()
    if args.output.exists(): raise ValueError("output exists; use a new immutable export name")
    manifest = json.loads((args.package / "RELATIONS_V2_MANIFEST.json").read_text())
    for name, expected in manifest["input_sha256"].items():
        if digest(PKG / name)!=expected: raise ValueError(f"frozen attachment input changed: {name}")
    expected_xml = args.package / args.subset / "annotations.xml"
    if digest(expected_xml)!=manifest["files"][f"{args.subset}/annotations.xml"]["sha256"]:
        raise ValueError("reference task XML changed")
    stamp = args.timestamp or dt.datetime.now(dt.timezone.utc).isoformat()
    overrides = None
    if args.reviewer_overrides:
        with args.reviewer_overrides.open(encoding="utf-8",newline="") as handle:
            overrides = list(csv.DictReader(handle,delimiter="\t"))
    rows = validate(load(expected_xml),load(args.input),args.reviewer_id,stamp,args.allow_unused_legacy_schema,overrides)
    args.output.parent.mkdir(parents=True,exist_ok=True)
    with args.output.open("w",encoding="utf-8",newline="") as handle:
        writer = csv.DictWriter(handle,V2_FIELDS,delimiter="\t",lineterminator="\n")
        writer.writeheader(); writer.writerows(rows)
    print(f"Imported {len(rows)} visual attachment v2 decisions ({args.subset})")


if __name__ == "__main__": main()
