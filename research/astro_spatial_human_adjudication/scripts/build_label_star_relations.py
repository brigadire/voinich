#!/usr/bin/env python3
"""Prepare neutral per-pair CVAT views from frozen, confirmed human endpoints."""
from __future__ import annotations

import argparse
import csv
import hashlib
import json
import math
import shutil
import xml.etree.ElementTree as ET
from pathlib import Path
from zipfile import ZIP_DEFLATED, ZIP_STORED, ZipFile, ZipInfo

from PIL import Image, __version__ as PIL_VERSION
from export_to_cvat import add_attr, add_meta_attribute

PKG = Path(__file__).resolve().parents[1]
FLAGS = {"adjacent_to": "ADJACENT_TO", "nearest_object": "NEAREST_OBJECT",
         "between_objects": "BETWEEN_OBJECTS", "on_object": "ON_OBJECT",
         "inside_object": "INSIDE_OBJECT"}
INPUTS = ("HUMAN_STAR_ADJUDICATION.tsv", "HUMAN_LABEL_ADJUDICATION.tsv",
          "HUMAN_CANDIDATE_RELATIONS.tsv", "IMAGE_MANIFEST.tsv",
          "HUMAN_ADJUDICATION_PROTOCOL_FROZEN.md")


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as handle:
        return list(csv.DictReader(handle, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], rows: list[dict]) -> None:
    with path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fields, delimiter="\t", lineterminator="\n")
        writer.writeheader()
        writer.writerows(rows)


def attribute(name: str, kind: str, default: str, values: list[str], mutable: bool = True) -> dict:
    return {"name": name, "input_type": kind, "mutable": mutable,
            "default_value": default, "values": values}


def schema() -> list[dict]:
    ids = [attribute(name, "text", "", [""], False) for name in
           ("relation_candidate_id", "label_candidate_id", "object_candidate_id")]
    return [
        {"name": "LABEL_ENDPOINT", "type": "any", "color": "#03a9f4", "attributes": ids},
        {"name": "STAR_ENDPOINT", "type": "rectangle", "color": "#ff9800", "attributes": ids + [
            attribute("relation_decision", "select", "UNREVIEWED", ["UNREVIEWED", "ASSIGNED", "UNASSIGNED", "UNCERTAIN"]),
            *[attribute(name, "checkbox", "false", ["false"]) for name in FLAGS],
            attribute("human_confidence", "select", "HIGH", ["LOW", "MEDIUM", "HIGH"]),
            attribute("notes", "text", "", [""]),
        ]},
    ]


def envelope(row: dict[str, str]) -> tuple[float, float, float, float]:
    x1, y1, x2, y2 = (float(row[k]) for k in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"))
    if row.get("geometry_type") == "ELLIPSE":
        return x1, y1, x2, y2
    angle = math.radians(float(row.get("rotation", 0)))
    cx, cy = (x1+x2)/2, (y1+y2)/2
    points = [(cx+(x-cx)*math.cos(angle)-(y-cy)*math.sin(angle),
               cy+(x-cx)*math.sin(angle)+(y-cy)*math.cos(angle))
              for x in (x1, x2) for y in (y1, y2)]
    return min(x for x,y in points), min(y for x,y in points), max(x for x,y in points), max(y for x,y in points)


def shape(parent: ET.Element, row: dict[str, str], role: str, pair: dict, offset: tuple[int, int], group: int) -> None:
    ox, oy = offset
    x1, y1, x2, y2 = (float(row[k]) for k in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"))
    common = {"label": role, "source": "manual", "occluded": "0", "z_order": "0", "group_id": str(group)}
    if row.get("geometry_type") == "ELLIPSE":
        node = ET.SubElement(parent, "ellipse", common | {
            "cx": f"{(x1+x2)/2-ox:.6f}", "cy": f"{(y1+y2)/2-oy:.6f}",
            "rx": f"{(x2-x1)/2:.6f}", "ry": f"{(y2-y1)/2:.6f}", "rotation": row.get("rotation", "0"),
        })
    else:
        node = ET.SubElement(parent, "box", common | {
            "xtl": f"{x1-ox:.6f}", "ytl": f"{y1-oy:.6f}",
            "xbr": f"{x2-ox:.6f}", "ybr": f"{y2-oy:.6f}", "rotation": row.get("rotation", "0"),
        })
    for name in ("relation_candidate_id", "label_candidate_id", "object_candidate_id"):
        add_attr(node, name, pair[name])
    if role == "STAR_ENDPOINT":
        for attr in schema()[1]["attributes"][3:]:
            add_attr(node, attr["name"], attr["default_value"])


def archive(path: Path, entries: list[tuple[str, Path]]) -> None:
    with ZipFile(path, "w") as output:
        for name, source in entries:
            info = ZipInfo(name, date_time=(1980,1,1,0,0,0))
            info.compress_type = ZIP_STORED if source.suffix.lower() in {".png", ".jpg"} else ZIP_DEFLATED
            info.external_attr = 0o644 << 16
            output.writestr(info, source.read_bytes())


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=Path, default=PKG / "relations")
    args = parser.parse_args()
    out = args.output_dir
    if (out / "RELATIONS_MANIFEST.json").exists():
        raise ValueError("completed relation package already exists; rebuild into a new output directory")
    frozen = json.loads((PKG / "manifest.json").read_text())
    input_hashes = {name: digest(PKG / name) for name in INPUTS}
    for name, actual in input_hashes.items():
        if frozen["files"][name]["sha256"] != actual:
            raise ValueError(f"input differs from frozen package: {name}")
    stars = {r["candidate_id"]: r for r in read(PKG / INPUTS[0])}
    labels = {r["candidate_id"]: r for r in read(PKG / INPUTS[1])}
    confirmed = lambda row: row and row["human_decision"] in {"ACCEPT", "MODIFY"}
    selected, excluded = [], []
    for pair in read(PKG / INPUTS[2]):
        label, star = labels.get(pair["label_candidate_id"]), stars.get(pair["object_candidate_id"])
        reasons = []
        if not confirmed(label): reasons.append("LABEL_NOT_CONFIRMED")
        if not confirmed(star) or star["final_class"] != "STAR_OBJECT": reasons.append("STAR_NOT_CONFIRMED")
        if reasons:
            excluded.append({"relation_candidate_id": pair["relation_candidate_id"], "label_candidate_id": pair["label_candidate_id"],
                             "object_candidate_id": pair["object_candidate_id"], "reason": ";".join(reasons)})
        else:
            if label["panel"] != star["panel"]: raise ValueError("cross-panel source pair")
            selected.append(pair)
    selected.sort(key=lambda row: row["relation_candidate_id"])
    if len({r["relation_candidate_id"] for r in selected}) != len(selected): raise ValueError("duplicate relation ID")
    (out / "images").mkdir(parents=True, exist_ok=True)
    (out / "cvat").mkdir(exist_ok=True)
    (out / "CVAT_LABEL_STAR_RELATION_SCHEMA.json").write_text(json.dumps(schema(), indent=2)+"\n")
    root = ET.Element("annotations")
    ET.SubElement(root, "version").text = "1.1"
    task = ET.SubElement(ET.SubElement(root, "meta"), "task")
    ET.SubElement(task, "name").text = "label-star-relations-h1"
    ET.SubElement(task, "mode").text = "annotation"
    ET.SubElement(task, "overlap").text = "0"
    ET.SubElement(task, "flipped").text = "False"
    meta_labels = ET.SubElement(task, "labels")
    for label in schema():
        node = ET.SubElement(meta_labels, "label")
        for key in ("name", "type", "color"): ET.SubElement(node, key).text = label[key]
        attrs = ET.SubElement(node, "attributes")
        for attr in label["attributes"]:
            add_meta_attribute(attrs, attr["name"], attr["mutable"], attr["input_type"], attr["default_value"], attr["values"])
    source_info = {r["panel"]: r for r in read(PKG / INPUTS[3])}
    panels = sorted({labels[r["label_candidate_id"]]["panel"] for r in selected})
    images = {}
    for panel in panels:
        path = PKG / "images" / source_info[panel]["filename"]
        if digest(path) != source_info[panel]["sha256"]: raise ValueError(f"canonical image changed: {panel}")
        images[panel] = Image.open(path).convert("RGB")
        input_hashes[f"images/{path.name}"] = digest(path)
    queue = []
    pixel_manifest = {}
    for index, pair in enumerate(selected):
        label, star = labels[pair["label_candidate_id"]], stars[pair["object_candidate_id"]]
        panel = label["panel"]
        le, se = envelope(label), envelope(star)
        image = images[panel]
        # Translation only; context margin is at least 256 canonical pixels.
        distance = math.hypot(float(label["center_x"])-float(star["center_x"]), float(label["center_y"])-float(star["center_y"]))
        margin = max(256, math.ceil(distance))
        crop = (max(0, math.floor(min(le[0],se[0])-margin)), max(0, math.floor(min(le[1],se[1])-margin)),
                min(image.width, math.ceil(max(le[2],se[2])+margin)), min(image.height, math.ceil(max(le[3],se[3])+margin)))
        pixels = image.crop(crop)
        filename = pair["relation_candidate_id"]+".png"
        pixels.save(out / "images" / filename, compress_level=6)
        pixel_manifest[filename] = {"decoded_rgb_sha256": hashlib.sha256(pixels.tobytes()).hexdigest()}
        node = ET.SubElement(root, "image", {"id": str(index), "name": filename, "width": str(pixels.width), "height": str(pixels.height)})
        shape(node, label, "LABEL_ENDPOINT", pair, crop[:2], index+1)
        shape(node, star, "STAR_ENDPOINT", pair, crop[:2], index+1)
        queue.append({**{k:pair[k] for k in ("relation_candidate_id", "label_candidate_id", "object_candidate_id")},
                      "panel": panel, "filename": filename, "crop_x": crop[0], "crop_y": crop[1],
                      "width": pixels.width, "height": pixels.height})
        if (index+1)%50 == 0: print(f"Prepared {index+1}/{len(selected)} pairs", flush=True)
    for panel in panels:
        filename = f"ZZ_CONTEXT_{panel}.jpg"
        shutil.copyfile(PKG / "images" / source_info[panel]["filename"], out / "images" / filename)
        image = images[panel]
        ET.SubElement(root, "image", {"id": str(len(queue)+panels.index(panel)), "name": filename,
                                     "width": str(image.width), "height": str(image.height)})
    ET.SubElement(task, "size").text = str(len(queue)+len(panels))
    ET.indent(root, space="  ")
    xml_path = out / "cvat/label_star_relations_h1.xml"
    ET.ElementTree(root).write(xml_path, encoding="utf-8", xml_declaration=True)
    archive(out / "cvat/label_star_relations_h1.zip", [("annotations.xml", xml_path)])
    archive(out / "LABEL_STAR_RELATIONS_IMAGES.zip", [(p.name,p) for p in sorted((out / "images").iterdir())])
    write_tsv(out / "LABEL_STAR_RELATION_QUEUE.tsv", list(queue[0]), queue)
    write_tsv(out / "LABEL_STAR_RELATION_EXCLUSIONS.tsv", ["relation_candidate_id", "label_candidate_id", "object_candidate_id", "reason"], excluded)
    manifest = {"version": "1.0", "pass": "H1", "pairs": len(queue), "context_frames": len(panels),
                "excluded_pairs": len(excluded), "input_sha256": input_hashes, "pillow_version": PIL_VERSION,
                "crop_transform": "integer translation only; PNG pixels equal canonical JPEG decoded RGB crop",
                "pixels": pixel_manifest, "new_geometric_pairs_created": 0,
                "files": {str(p.relative_to(out)): {"sha256":digest(p), "bytes":p.stat().st_size}
                          for p in sorted(out.rglob('*')) if p.is_file()}}
    (out / "RELATIONS_MANIFEST.json").write_text(json.dumps(manifest, indent=2, sort_keys=True)+"\n")
    print(json.dumps({"pairs":len(queue), "excluded":len(excluded), "images_zip_bytes":(out / "LABEL_STAR_RELATIONS_IMAGES.zip").stat().st_size}), flush=True)


if __name__ == "__main__":
    main()
