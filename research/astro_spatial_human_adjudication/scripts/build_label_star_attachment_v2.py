#!/usr/bin/env python3
"""Partition the original relation queue under reviewer-confirmed attachment v2."""
from __future__ import annotations

import argparse
import copy
import json
import shutil
import xml.etree.ElementTree as ET
from collections import Counter
from pathlib import Path

from build_label_star_relations import PKG, archive, attribute, digest, envelope, read, write_tsv
from export_to_cvat import add_attr, add_meta_attribute
from import_label_star_relations import FIELDS

PROTOCOL = "HUMAN_LABEL_STAR_ATTACHMENT_PROTOCOL_V2.md"
V2_FIELDS = FIELDS + ["relation_protocol_version", "decision_origin"]


def schema():
    ids = [attribute(name, "text", "", [""], False) for name in
           ("relation_candidate_id", "label_candidate_id", "object_candidate_id", "panel")]
    return [
        {"name":"LABEL_ENDPOINT", "type":"any", "color":"#03a9f4", "attributes":ids},
        {"name":"STAR_ENDPOINT", "type":"rectangle", "color":"#ff9800", "attributes":ids+[
            attribute("relation_decision", "select", "UNREVIEWED", ["UNREVIEWED","ASSIGNED","UNASSIGNED","UNCERTAIN"]),
            attribute("human_confidence", "select", "HIGH", ["LOW","MEDIUM","HIGH"]),
            attribute("notes", "text", "", [""]),
        ]},
    ]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=Path, default=PKG / "relations_v2")
    parser.add_argument("--rule-timestamp", required=True)
    args = parser.parse_args()
    out = args.output_dir
    if out.exists(): raise ValueError("use a new output directory; original tasks must remain immutable")
    old = PKG / "relations"
    previous = json.loads((old / "RELATIONS_MANIFEST.json").read_text())
    for name, expected in previous["input_sha256"].items():
        if digest(PKG / name) != expected: raise ValueError(f"frozen input changed: {name}")
    for name in ("LABEL_STAR_RELATION_QUEUE.tsv", "cvat/label_star_relations_h1.xml"):
        if digest(old / name) != previous["files"][name]["sha256"]: raise ValueError(f"old package changed: {name}")
    labels = {r["candidate_id"]:r for r in read(PKG / "HUMAN_LABEL_ADJUDICATION.tsv")}
    stars = {r["candidate_id"]:r for r in read(PKG / "HUMAN_STAR_ADJUDICATION.tsv")}
    groups = {"main":[], "f68v2":[]}
    auto, filtered = [], []
    for row in read(old / "LABEL_STAR_RELATION_QUEUE.tsv"):
        panel = row["panel"]
        if panel in {"f67r1","f67v1","f68v1"}:
            auto.append({**{k:row[k] for k in ("relation_candidate_id","label_candidate_id","object_candidate_id")},
                         "relation_type":"UNASSIGNED", "human_confidence":"HIGH", "reviewer_id":"R01",
                         "review_timestamp":args.rule_timestamp, "notes":f"Reviewer-confirmed panel-level attachment rule: {panel}.",
                         "relation_protocol_version":"2", "decision_origin":"REVIEWER_PANEL_RULE"})
        elif panel == "f68v2": groups["f68v2"].append(row)
        elif panel in {"f68r1","f68r2"}:
            label, star = labels[row["label_candidate_id"]], stars[row["object_candidate_id"]]
            le,se = envelope(label),envelope(star)
            right = float(label["center_x"])>float(star["center_x"])
            overlap = max(le[1],se[1])<=min(le[3],se[3])
            if right and overlap: groups["main"].append(row)
            else: filtered.append({**{k:row[k] for k in ("relation_candidate_id","label_candidate_id","object_candidate_id","panel")},
                                    "reason":"LABEL_NOT_RIGHT" if not right else "NO_VERTICAL_ENVELOPE_OVERLAP",
                                    "status":"RULE_FILTERED_NOT_REVIEWED"})
        elif panel == "f68r3": groups["main"].append(row)
        else: raise ValueError(f"unhandled panel: {panel}")
    out.mkdir()
    (out / "CVAT_LABEL_STAR_ATTACHMENT_SCHEMA.json").write_text(json.dumps(schema(),indent=2)+"\n")
    old_frames = {r.attrib["name"]:r for r in ET.parse(old / "cvat/label_star_relations_h1.xml").getroot().findall("image")}
    task_info = {}
    for subset, rows in groups.items():
        base = out / subset
        (base / "images").mkdir(parents=True)
        root = ET.Element("annotations")
        ET.SubElement(root,"version").text = "1.1"
        task = ET.SubElement(ET.SubElement(root,"meta"),"task")
        ET.SubElement(task,"name").text = f"label-star-attachment-v2-{subset}"
        ET.SubElement(task,"mode").text = "annotation"
        ET.SubElement(task,"overlap").text = "0"
        ET.SubElement(task,"flipped").text = "False"
        meta_labels = ET.SubElement(task,"labels")
        for label in schema():
            node = ET.SubElement(meta_labels,"label")
            for key in ("name","type","color"): ET.SubElement(node,key).text = label[key]
            attrs = ET.SubElement(node,"attributes")
            for attr in label["attributes"]: add_meta_attribute(attrs,attr["name"],attr["mutable"],attr["input_type"],attr["default_value"],attr["values"])
        for index,row in enumerate(rows):
            filename = row["filename"]
            if digest(old / "images" / filename) != previous["files"][f"images/{filename}"]["sha256"]: raise ValueError("crop changed")
            shutil.copyfile(old / "images" / filename,base / "images" / filename)
            frame = copy.deepcopy(old_frames[filename])
            frame.set("id",str(index))
            for endpoint in frame:
                endpoint.set("group_id",str(index+1))
                for attr in list(endpoint.findall("attribute")): endpoint.remove(attr)
                for name in ("relation_candidate_id","label_candidate_id","object_candidate_id","panel"): add_attr(endpoint,name,row[name])
                if endpoint.attrib["label"]=="STAR_ENDPOINT":
                    for attr in schema()[1]["attributes"][4:]: add_attr(endpoint,attr["name"],attr["default_value"])
            root.append(frame)
        panels = sorted({r["panel"] for r in rows})
        for index,panel in enumerate(panels):
            filename = f"ZZ_CONTEXT_{panel}.jpg"
            if digest(old / "images" / filename) != previous["files"][f"images/{filename}"]["sha256"]: raise ValueError("context changed")
            shutil.copyfile(old / "images" / filename,base / "images" / filename)
            frame = copy.deepcopy(old_frames[filename]); frame.set("id",str(len(rows)+index)); root.append(frame)
        ET.SubElement(task,"size").text = str(len(rows)+len(panels))
        ET.indent(root,space="  ")
        xml = base / "annotations.xml"
        ET.ElementTree(root).write(xml,encoding="utf-8",xml_declaration=True)
        archive(base / "ANNOTATIONS.zip",[("annotations.xml",xml)])
        archive(base / "IMAGES.zip",[(p.name,p) for p in sorted((base / "images").iterdir())])
        write_tsv(base / "QUEUE.tsv",list(rows[0]),rows)
        task_info[subset] = {"pairs":len(rows), "context_frames":len(panels), "panels":dict(sorted(Counter(r["panel"] for r in rows).items())),
                             "images_zip_bytes":(base / "IMAGES.zip").stat().st_size}
    write_tsv(out / "REVIEWER_PANEL_DECISIONS.tsv",V2_FIELDS,auto)
    write_tsv(out / "RIGHT_RULE_FILTERED.tsv",list(filtered[0]),filtered)
    input_hashes = {**previous["input_sha256"], "relations/RELATIONS_MANIFEST.json":digest(old / "RELATIONS_MANIFEST.json"),
                    PROTOCOL:digest(PKG / PROTOCOL)}
    manifest = {"relation_protocol_version":"2", "input_sha256":input_hashes, "tasks":task_info,
                "reviewer_panel_unassigned":len(auto), "right_rule_filtered_not_reviewed":len(filtered),
                "rule_timestamp":args.rule_timestamp,
                "files":{str(p.relative_to(out)):{"sha256":digest(p),"bytes":p.stat().st_size} for p in sorted(out.rglob('*')) if p.is_file()}}
    (out / "RELATIONS_V2_MANIFEST.json").write_text(json.dumps(manifest,indent=2,sort_keys=True)+"\n")
    print(json.dumps({"tasks":task_info,"reviewer_panel_unassigned":len(auto),"filtered_not_reviewed":len(filtered)},sort_keys=True))


if __name__ == "__main__": main()
