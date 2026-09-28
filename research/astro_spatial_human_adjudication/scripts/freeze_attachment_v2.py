#!/usr/bin/env python3
"""Freeze the complete attachment v2 review, retaining explicit negative/unresolved outcomes."""
from __future__ import annotations

import argparse
import json
from collections import Counter, defaultdict
from pathlib import Path

from build_label_star_attachment_v2 import V2_FIELDS
from build_label_star_relations import PKG, digest, read
from freeze_reviewed_layers import render


def build():
    package = PKG / "relations_v2"
    manifest = json.loads((package / "RELATIONS_V2_MANIFEST.json").read_text())
    for filename, expected in manifest["input_sha256"].items():
        if digest(PKG / filename)!=expected: raise ValueError(f"frozen attachment input changed: {filename}")
    merged = {}
    parts = (
        (PKG / "exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_MAIN_R01.tsv",package / "main/QUEUE.tsv"),
        (PKG / "exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_F68V2_R01.tsv",package / "f68v2/QUEUE.tsv"),
        (package / "REVIEWER_PANEL_DECISIONS.tsv",package / "REVIEWER_PANEL_DECISIONS.tsv"),
    )
    stars = {r["candidate_id"]:r for r in read(PKG / "HUMAN_STAR_ADJUDICATION.tsv")}
    labels = {r["candidate_id"]:r for r in read(PKG / "HUMAN_LABEL_ADJUDICATION.tsv")}
    for reviewed_file, expected_file in parts:
        ref_key = str(expected_file.relative_to(package))
        if digest(expected_file)!=manifest["files"][ref_key]["sha256"]: raise ValueError("prepared queue or panel decisions changed")
        rows = read(reviewed_file)
        expected = {r["relation_candidate_id"]:r for r in read(expected_file)}
        if len({r["relation_candidate_id"] for r in rows})!=len(rows): raise ValueError("duplicate relation IDs")
        if {r["relation_candidate_id"] for r in rows}!=set(expected): raise ValueError("incomplete or unexpected review pairs")
        for row in rows:
            cid = row["relation_candidate_id"]
            if set(row)!=set(V2_FIELDS): raise ValueError("incorrect attachment output schema")
            if cid in merged: raise ValueError("overlapping review subsets")
            for key in ("label_candidate_id","object_candidate_id"):
                if row[key]!=expected[cid][key]: raise ValueError("attachment endpoint changed")
            if row["relation_protocol_version"]!="2" or row["relation_type"] not in {"VISUAL_LABEL_OF","UNASSIGNED","UNCERTAIN"}:
                raise ValueError("invalid, incomplete, or mixed-version attachment result")
            if row["human_confidence"] not in {"LOW","MEDIUM","HIGH"}: raise ValueError("unset confidence")
            label,star = labels[row["label_candidate_id"]],stars[row["object_candidate_id"]]
            if label["human_decision"] not in {"ACCEPT","MODIFY"} or star["human_decision"] not in {"ACCEPT","MODIFY"} or star["final_class"]!="STAR_OBJECT":
                raise ValueError("unconfirmed attachment endpoint")
            merged[cid] = row
    filtered_file = package / "RIGHT_RULE_FILTERED.tsv"
    if digest(filtered_file)!=manifest["files"]["RIGHT_RULE_FILTERED.tsv"]["sha256"]: raise ValueError("filter audit changed")
    filtered = {r["relation_candidate_id"] for r in read(filtered_file)}
    original = {r["relation_candidate_id"] for r in read(PKG / "relations/LABEL_STAR_RELATION_QUEUE.tsv")}
    if set(merged)&filtered or set(merged)|filtered!=original: raise ValueError("attachment queue partition mismatch")
    rows = sorted(merged.values(),key=lambda row:row["relation_candidate_id"])
    positive = [r for r in rows if r["relation_type"]=="VISUAL_LABEL_OF"]
    by_star = defaultdict(set)
    for row in positive:by_star[row["object_candidate_id"]].add(row["label_candidate_id"])
    metrics = {
        "relation_protocol_version":"2", "status":"FROZEN_COMPLETE_FOR_SELECTED_QUEUE",
        "explicit_decisions":len(rows), "individual_pair_reviews":len(rows)-manifest["reviewer_panel_unassigned"],
        "reviewer_panel_rule_decisions":manifest["reviewer_panel_unassigned"],
        "decision_counts":dict(sorted(Counter(row["relation_type"] for row in rows).items())),
        "confidence_counts":dict(sorted(Counter(row["human_confidence"] for row in rows).items())),
        "decision_origin_counts":dict(sorted(Counter(row["decision_origin"] for row in rows).items())),
        "rule_filtered_not_reviewed":len(filtered),
        "confirmed_attachment_labels":len({row["label_candidate_id"] for row in positive}),
        "confirmed_attachment_stars":len(by_star),
        "stars_with_multiple_assigned_labels":sum(len(value)>1 for value in by_star.values()),
        "note":"VISUAL_LABEL_OF is visible caption attachment, not lexical/astronomical identity or mere proximity; filtered pairs are not negative human decisions; unresolved outcomes are not promoted",
    }
    contents = render(V2_FIELDS,rows)
    return {
        "HUMAN_LABEL_OBJECT_RELATIONS.tsv":contents,
        "exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_FINAL_R01.tsv":contents,
        "exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_METRICS.json":json.dumps(metrics,indent=2,sort_keys=True)+"\n",
    },metrics


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir",type=Path,default=PKG)
    args = parser.parse_args()
    outputs,metrics = build()
    for filename,contents in outputs.items():
        target = args.output_dir / filename
        if target.exists() and target.read_text(encoding="utf-8").count("\n")>1 and target.read_text(encoding="utf-8")!=contents:
            raise ValueError(f"completed attachment snapshot differs; create a new revision: {target}")
    for filename,contents in outputs.items():
        target = args.output_dir / filename
        target.parent.mkdir(parents=True,exist_ok=True)
        target.write_text(contents,encoding="utf-8")
    print(json.dumps(metrics,sort_keys=True))


if __name__ == "__main__":main()
