#!/usr/bin/env python3
"""Build the frozen, provenance-preserving human adjudication candidate layer."""
from __future__ import annotations

import csv
import hashlib
import json
import math
import shutil
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
PKG = Path(__file__).resolve().parents[1]
A = ROOT / "research/astro_spatial_annotation"
AI1 = ROOT / "research/astro_spatial_annotation_ai_b"
AI2 = ROOT / "research/astro_spatial_annotation_ai_b2"
PANELS = ("f67r1", "f67r2", "f67v1", "f68r1", "f68r2", "f68r3", "f68v1", "f68v2")
CALIBRATION = {"f68r1", "f68r3", "f68v2"}
SOURCE_TAGS = ("A", "AI1", "AI2")


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as handle:
        return list(csv.DictReader(handle, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], rows: list[dict]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        writer.writeheader()
        writer.writerows(rows)


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def load_manifest(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def verify_and_copy_images() -> list[dict]:
    m1 = load_manifest(AI1 / "AI_B_INPUT_MANIFEST.json")["files"]
    m2 = load_manifest(AI2 / "AI2_INPUT_MANIFEST.json")["files"]
    records = []
    for panel in PANELS:
        key = f"crops/{panel}.jpg"
        source1 = AI1 / "package" / key
        source2 = AI2 / "package" / key
        h1, h2 = sha256(source1), sha256(source2)
        expected1, expected2 = m1[key]["sha256"], m2[key]["sha256"]
        if not (h1 == expected1 == h2 == expected2):
            raise RuntimeError(f"canonical image mismatch for {panel}: {h1}, {expected1}, {h2}, {expected2}")
        target = PKG / "images" / f"{panel}.jpg"
        if not target.exists() or sha256(target) != h1:
            shutil.copyfile(source1, target)
        records.append({
            "panel": panel, "filename": target.name, "sha256": h1,
            "width": m1[key]["width"], "height": m1[key]["height"], "bytes": m1[key]["bytes"],
            "matches_ai1": "YES", "matches_ai2": "YES",
        })
    return records


class DSU:
    def __init__(self, nodes):
        self.parent = {n: n for n in nodes}
        self.members = {n: {n} for n in nodes}

    def find(self, node):
        while self.parent[node] != node:
            self.parent[node] = self.parent[self.parent[node]]
            node = self.parent[node]
        return node

    def union(self, left, right):
        a, b = self.find(left), self.find(right)
        if a != b:
            # One physical candidate may contain at most one annotation from
            # each source. Pairwise match triangles can otherwise collapse
            # neighbouring marks through a non-transitive chain.
            if {tag for tag, _ in self.members[a]} & {tag for tag, _ in self.members[b]}:
                return False
            keep, drop = min(a, b), max(a, b)
            self.parent[drop] = keep
            self.members[keep] |= self.members.pop(drop)
            return True
        return True


def add_frozen_matches(dsu: DSU, path: Path, left_col: str, right_col: str, left_tag: str, right_tag: str) -> None:
    for row in read_tsv(path):
        left, right = row[left_col], row[right_col]
        if row["match_status"] == "MATCHED" and left != "UNMATCHED" and right != "UNMATCHED":
            lnode, rnode = (left_tag, left), (right_tag, right)
            if lnode in dsu.parent and rnode in dsu.parent:
                dsu.union(lnode, rnode)


def bbox(record: dict) -> tuple[float, float, float, float]:
    return tuple(float(record[k]) for k in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"))


def iou(left: tuple[float, ...], right: tuple[float, ...]) -> float:
    x1, y1, x2, y2 = max(left[0], right[0]), max(left[1], right[1]), min(left[2], right[2]), min(left[3], right[3])
    inter = max(0.0, x2 - x1) * max(0.0, y2 - y1)
    la = max(0.0, left[2] - left[0]) * max(0.0, left[3] - left[1])
    ra = max(0.0, right[2] - right[0]) * max(0.0, right[3] - right[1])
    return inter / (la + ra - inter) if la + ra > inter else 0.0


def fmt(value: float) -> str:
    return f"{value:.6f}"


def source_confidence(record: dict) -> str:
    return record.get("confidence") or record.get("annotation_confidence") or "UNSPECIFIED"


def provisional_rotation(records: list[dict]) -> float:
    """Return an axial circular mean: text axes 0° and 180° are equivalent."""
    angles = [float(r.get("orientation_angle") or 0.0) % 180.0 for r in records]
    nonzero = [angle for angle in angles if min(angle, 180.0 - angle) > 0.01]
    selected = nonzero or angles
    if not selected:
        return 0.0
    sine = sum(math.sin(math.radians(2.0 * angle)) for angle in selected)
    cosine = sum(math.cos(math.radians(2.0 * angle)) for angle in selected)
    result = math.degrees(math.atan2(sine, cosine)) / 2.0
    return result % 180.0


def label_geometry(records: list[dict]) -> str:
    notes = " ".join(r.get("notes", "").lower() for r in records)
    if "short text run" in notes:
        return "ROTATED_RECTANGLE"
    ring_language = any(phrase in notes for phrase in ("continuous ring", "ring of writing", "circumference of the diagram", "round the circumference"))
    largest = max((max(bbox(r)[2] - bbox(r)[0], bbox(r)[3] - bbox(r)[1]) for r in records), default=0.0)
    return "ELLIPSE_RING" if ring_language and largest >= 400.0 else "ROTATED_RECTANGLE"


def agreement(support: set[str]) -> str:
    mapping = {
        frozenset(("A", "AI1", "AI2")): "SUPPORT_3_OF_3",
        frozenset(("A", "AI1")): "SUPPORT_A_AI1",
        frozenset(("A", "AI2")): "SUPPORT_A_AI2",
        frozenset(("AI1", "AI2")): "SUPPORT_AI1_AI2",
        frozenset(("A",)): "SUPPORT_A_ONLY",
        frozenset(("AI1",)): "SUPPORT_AI1_ONLY",
        frozenset(("AI2",)): "SUPPORT_AI2_ONLY",
    }
    return mapping[frozenset(support)]


def priority(kind: str, panel: str, support: set[str], records: list[dict]) -> tuple[str, str]:
    boxes = [bbox(r) for r in records]
    min_iou = min((iou(boxes[i], boxes[j]) for i in range(len(boxes)) for j in range(i + 1, len(boxes))), default=1.0)
    low_conf = any(source_confidence(r).upper() == "LOW" for r in records)
    ai_singleton = support in ({"AI1"}, {"AI2"})
    if ai_singleton:
        return "HIGH", "single_blind_source"
    if len(records) > 1 and min_iou < 0.15:
        return "HIGH" if kind == "LABEL" else "HIGH", "major_bbox_or_granularity_disagreement"
    if panel == "f68r1" and len(support) < 3:
        return "HIGH", "independent_reference_conflict"
    if kind == "LABEL" and support in ({"A"},):
        return "HIGH", "reference_only_label"
    if low_conf or (len(records) > 1 and min_iou < 0.40):
        return "MEDIUM", "low_confidence_or_iou"
    if support in ({"AI1", "AI2"}, {"A", "AI1", "AI2"}):
        return "LOW", "strong_consensus"
    return "MEDIUM", "other_disagreement"


def build_candidates(kind: str, datasets: dict[str, list[dict]], id_fields: dict[str, str], match_specs: list[tuple]) -> tuple[list[dict], dict[tuple[str, str], str]]:
    records_by_node = {}
    for tag, rows in datasets.items():
        for row in rows:
            records_by_node[(tag, row[id_fields[tag]])] = row
    dsu = DSU(records_by_node)
    for spec in match_specs:
        add_frozen_matches(dsu, *spec)
    components = defaultdict(list)
    for node in records_by_node:
        components[dsu.find(node)].append(node)

    candidates, source_map = [], {}
    prefix = "HSTAR" if kind == "STAR_OBJECT" else ("HLABEL" if kind == "LABEL" else "HOBJ")
    for nodes in sorted(components.values(), key=lambda ns: (records_by_node[ns[0]]["panel"], sorted(ns))):
        rows = [records_by_node[n] for n in nodes]
        panels = {r["panel"] for r in rows}
        if len(panels) != 1:
            raise RuntimeError(f"cross-panel component: {nodes}")
        panel = next(iter(panels))
        stable_key = "|".join(f"{tag}:{source_id}" for tag, source_id in sorted(nodes))
        candidate_id = f"{prefix}_{panel}_{hashlib.sha256(stable_key.encode()).hexdigest()[:12].upper()}"
        support = {tag for tag, _ in nodes}
        coords = [bbox(r) for r in rows]
        raw_display = tuple(sum(c[i] for c in coords) / len(coords) for i in range(4))
        canonical = (min(c[0] for c in coords), min(c[1] for c in coords), max(c[2] for c in coords), max(c[3] for c in coords))
        rotation = provisional_rotation(rows) if kind == "LABEL" else 0.0
        geometry = label_geometry(rows) if kind == "LABEL" else "AXIS_ALIGNED_RECTANGLE"
        if kind == "LABEL" and geometry == "ROTATED_RECTANGLE" and abs(rotation) > 0.01:
            cx, cy = (raw_display[0] + raw_display[2]) / 2.0, (raw_display[1] + raw_display[3]) / 2.0
            length = max(raw_display[2] - raw_display[0], raw_display[3] - raw_display[1])
            thickness = min(raw_display[2] - raw_display[0], raw_display[3] - raw_display[1])
            display = (cx - length / 2.0, cy - thickness / 2.0, cx + length / 2.0, cy + thickness / 2.0)
        else:
            display = raw_display
        px1, py1, px2, py2 = display
        pri, reason = priority("LABEL" if kind == "LABEL" else kind, panel, support, rows)
        by_tag = defaultdict(list)
        source_ids = defaultdict(list)
        for tag, sid in sorted(nodes):
            by_tag[tag].append(records_by_node[(tag, sid)])
            source_ids[tag].append(sid)
        def bstr(tag):
            return ";".join(",".join(fmt(v) for v in bbox(record)) for record in by_tag[tag])
        row = {
            "candidate_id": candidate_id, "panel": panel, "candidate_type": kind,
            "canonical_bbox": ",".join(fmt(v) for v in canonical),
            "canonical_center": f"{fmt((canonical[0] + canonical[2]) / 2)},{fmt((canonical[1] + canonical[3]) / 2)}",
            "provisional_display_bbox": ",".join(fmt(v) for v in display),
            "geometry_mode": geometry, "provisional_rotation": fmt(rotation),
            "bbox_x1": fmt(px1), "bbox_y1": fmt(py1), "bbox_x2": fmt(px2), "bbox_y2": fmt(py2),
            "center_x": fmt((px1 + px2) / 2), "center_y": fmt((py1 + py2) / 2),
            "support_a": "YES" if "A" in support else "NO", "support_ai1": "YES" if "AI1" in support else "NO",
            "support_ai2": "YES" if "AI2" in support else "NO", "support_count": len(support),
            "agreement_type": agreement(support),
            "source_annotation_ids": ";".join(f"{tag}:{sid}" for tag in SOURCE_TAGS for sid in source_ids[tag]),
            "bbox_a": bstr("A"), "bbox_ai1": bstr("AI1"), "bbox_ai2": bstr("AI2"),
            "orientation_a": ";".join(r.get("orientation_angle", "0") or "0" for r in by_tag["A"]),
            "orientation_ai1": ";".join(r.get("orientation_angle", "0") or "0" for r in by_tag["AI1"]),
            "orientation_ai2": ";".join(r.get("orientation_angle", "0") or "0" for r in by_tag["AI2"]),
            "confidence_summary": ";".join(f"{tag}:{source_confidence(record)}" for tag in SOURCE_TAGS for record in by_tag[tag]),
            "priority": pri, "priority_reason": reason,
        }
        candidates.append(row)
        for node in nodes:
            source_map[node] = candidate_id
    return candidates, source_map


CANDIDATE_FIELDS = [
    "candidate_id", "panel", "candidate_type", "canonical_bbox", "canonical_center", "provisional_display_bbox", "geometry_mode", "provisional_rotation",
    "bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2", "center_x", "center_y",
    "support_a", "support_ai1", "support_ai2", "support_count", "agreement_type",
    "source_annotation_ids", "bbox_a", "bbox_ai1", "bbox_ai2", "orientation_a", "orientation_ai1", "orientation_ai2", "confidence_summary", "priority", "priority_reason",
]


def build_relations(label_map: dict, object_map: dict) -> list[dict]:
    inputs = {
        "A": (read_tsv(A / "ASTRO_LABEL_OBJECT_RELATIONS.tsv"), "label_occurrence_id"),
        "AI1": (read_tsv(AI1 / "AI_B_LABEL_OBJECT_RELATIONS.tsv"), "label_id"),
        "AI2": (read_tsv(AI2 / "AI2_LABEL_OBJECT_RELATIONS.tsv"), "label_id"),
    }
    grouped = defaultdict(list)
    for tag, (rows, label_field) in inputs.items():
        for row in rows:
            lk, ok = (tag, row[label_field]), (tag, row["object_id"])
            if lk not in label_map or ok not in object_map:
                continue
            grouped[(label_map[lk], object_map[ok])].append((tag, row))
    output = []
    for (label_id, object_id), items in sorted(grouped.items()):
        types = sorted({r["relation_type"] for _, r in items})
        tags = sorted({tag for tag, _ in items})
        key = f"{label_id}|{object_id}"
        output.append({
            "relation_candidate_id": f"HREL_{hashlib.sha256(key.encode()).hexdigest()[:12].upper()}",
            "label_candidate_id": label_id, "object_candidate_id": object_id,
            "provisional_relation_types": ";".join(types), "support_count": len(tags),
            "agreement_status": "AGREE" if len(types) == 1 and len(tags) > 1 else ("DISAGREE" if len(types) > 1 else "SINGLE_SOURCE"),
            "source_relation_ids": ";".join(f"{tag}:{r.get('label_occurrence_id') or r.get('label_id')}->{r['object_id']}" for tag, r in items),
            "priority": "HIGH" if len(types) > 1 else ("LOW" if len(tags) > 1 else "MEDIUM"),
        })
    return output


def main() -> None:
    images = verify_and_copy_images()
    a_objects = read_tsv(A / "ASTRO_OBJECTS.tsv")
    ai1_objects = read_tsv(AI1 / "AI_B_OBJECTS.tsv")
    ai2_objects = read_tsv(AI2 / "AI2_OBJECTS.tsv")
    object_specs = [
        (AI1 / "AI_B_A_MATCH_OBJECTS.tsv", "a_object_id", "b_object_id", "A", "AI1"),
        (AI2 / "A_AI2_OBJECT_MATCH.tsv", "a_object_id", "ai2_object_id", "A", "AI2"),
        (AI2 / "AI1_AI2_OBJECT_MATCH.tsv", "ai1_object_id", "ai2_object_id", "AI1", "AI2"),
    ]
    all_objects, object_map = build_candidates("OBJECT", {"A": a_objects, "AI1": ai1_objects, "AI2": ai2_objects}, {t: "object_id" for t in SOURCE_TAGS}, object_specs)
    for row in all_objects:
        source_classes = []
        for token in row["source_annotation_ids"].split(";"):
            tag, sid = token.split(":", 1)
            source = next(r for r in ({"A": a_objects, "AI1": ai1_objects, "AI2": ai2_objects}[tag]) if r["object_id"] == sid)
            source_classes.append(source["object_class"])
        row["candidate_type"] = source_classes[0] if len(set(source_classes)) == 1 else "OTHER_OBJECT"
        if row["candidate_type"] == "STAR_OBJECT" and row["panel"] == "f68r1" and int(row["support_count"]) < 3:
            row["priority"], row["priority_reason"] = "HIGH", "independent_reference_conflict"
        elif row["candidate_type"] != "STAR_OBJECT" and row["priority"] == "HIGH":
            row["priority"], row["priority_reason"] = "MEDIUM", "non_star_object_disagreement"
    stars = [r for r in all_objects if r["candidate_type"] == "STAR_OBJECT"]

    label_specs = [
        (AI1 / "AI_B_A_MATCH_LABELS.tsv", "a_label_id", "b_label_id", "A", "AI1"),
        (AI2 / "A_AI2_LABEL_MATCH.tsv", "a_label_id", "ai2_label_id", "A", "AI2"),
        (AI2 / "AI1_AI2_LABEL_MATCH.tsv", "ai1_label_id", "ai2_label_id", "AI1", "AI2"),
    ]
    labels, label_map = build_candidates(
        "LABEL", {"A": read_tsv(A / "ASTRO_LABELS_SPATIAL.tsv"), "AI1": read_tsv(AI1 / "AI_B_LABELS.tsv"), "AI2": read_tsv(AI2 / "AI2_LABELS.tsv")},
        {"A": "label_occurrence_id", "AI1": "label_id", "AI2": "label_id"}, label_specs,
    )
    relations = build_relations(label_map, object_map)
    write_tsv(PKG / "HUMAN_CANDIDATE_OBJECTS.tsv", CANDIDATE_FIELDS, all_objects)
    write_tsv(PKG / "HUMAN_CANDIDATE_LABELS.tsv", CANDIDATE_FIELDS, labels)
    write_tsv(PKG / "HUMAN_CANDIDATE_RELATIONS.tsv", ["relation_candidate_id", "label_candidate_id", "object_candidate_id", "provisional_relation_types", "support_count", "agreement_status", "source_relation_ids", "priority"], relations)
    priority_rows = [{"task": "STAR", "candidate_id": r["candidate_id"], "panel": r["panel"], "priority": r["priority"], "priority_reason": r["priority_reason"]} for r in stars]
    priority_rows += [{"task": "OTHER_OBJECT", "candidate_id": r["candidate_id"], "panel": r["panel"], "priority": r["priority"], "priority_reason": r["priority_reason"]} for r in all_objects if r["candidate_type"] != "STAR_OBJECT"]
    priority_rows += [{"task": "LABEL", "candidate_id": r["candidate_id"], "panel": r["panel"], "priority": r["priority"], "priority_reason": r["priority_reason"]} for r in labels]
    write_tsv(PKG / "HUMAN_REVIEW_PRIORITY.tsv", ["task", "candidate_id", "panel", "priority", "priority_reason"], sorted(priority_rows, key=lambda r: ({"HIGH": 0, "MEDIUM": 1, "LOW": 2}[r["priority"]], r["task"], r["panel"], r["candidate_id"])))
    cal = [{"task": t, "candidate_id": r["candidate_id"], "panel": r["panel"], "phase": "H-Cal"} for t, rows in (("STAR", stars), ("LABEL", labels)) for r in rows if r["panel"] in CALIBRATION]
    write_tsv(PKG / "HUMAN_CALIBRATION_SET.tsv", ["task", "candidate_id", "panel", "phase"], cal)
    label_followup = [r for r in labels if r["panel"] in CALIBRATION and r["geometry_mode"] == "ELLIPSE_RING"]
    write_tsv(PKG / "HUMAN_LABEL_CALIBRATION_FOLLOWUP.tsv", CANDIDATE_FIELDS, label_followup)
    write_tsv(PKG / "IMAGE_MANIFEST.tsv", ["panel", "filename", "sha256", "width", "height", "bytes", "matches_ai1", "matches_ai2"], images)
    summary = {
        "canonical_image_bytes_match_ai1": True, "canonical_image_bytes_match_ai2": True,
        "source_counts": {"ai1_star_objects": sum(r["object_class"] == "STAR_OBJECT" for r in ai1_objects), "ai1_labels": len(read_tsv(AI1 / "AI_B_LABELS.tsv")), "ai2_star_objects": sum(r["object_class"] == "STAR_OBJECT" for r in ai2_objects), "ai2_labels": len(read_tsv(AI2 / "AI2_LABELS.tsv"))},
        "candidate_counts": {"stars": len(stars), "labels": len(labels), "relations": len(relations)},
        "calibration_followup": {"ring_labels": len(label_followup)},
        "high_priority": {"stars": sum(r["priority"] == "HIGH" for r in stars), "labels": sum(r["priority"] == "HIGH" for r in labels)},
        "matching": "frozen_pairwise_matches_with_deterministic_one-record-per-source_component_reconciliation",
        "canonical_bbox": "union_of_source_bboxes", "display_bbox": "orientation_normalized_mean_source_bbox_ui_helper_only",
    }
    human_paths = {
        "star_calibration": PKG / "exports/HUMAN_STAR_CALIBRATION_R01.tsv",
        "label_calibration": PKG / "exports/HUMAN_LABEL_CALIBRATION_R01_FINAL.tsv",
        "star_high": PKG / "exports/HUMAN_STAR_HIGH_R01_FINAL.tsv",
        "label_high": PKG / "exports/HUMAN_LABEL_HIGH_R01_FINAL.tsv",
        "star_qc": PKG / "exports/HUMAN_STAR_CONSENSUS_QC_R01_FINAL.tsv",
        "label_qc": PKG / "exports/HUMAN_LABEL_CONSENSUS_QC_R01_FINAL.tsv",
    }
    if all(path.exists() for path in human_paths.values()):
        human_rows = {name: read_tsv(path) for name, path in human_paths.items()}
        medium_paths = {
            "star_medium": PKG / "exports/HUMAN_STAR_MEDIUM_R01.tsv",
            "label_medium": PKG / "exports/HUMAN_LABEL_MEDIUM_R01.tsv",
        }
        medium_complete = all(path.exists() for path in medium_paths.values())
        if medium_complete:
            human_rows.update({name: read_tsv(path) for name, path in medium_paths.items()})
        reviewed_stars = {
            row["candidate_id"]
            for name in human_rows if name.startswith("star_")
            for row in human_rows[name]
        }
        reviewed_labels = {
            row["candidate_id"]
            for name in human_rows if name.startswith("label_")
            for row in human_rows[name]
        }
        summary["human_review"] = {
            "consensus_qc_complete": True,
            "consensus_qc_labels": 18,
            "consensus_qc_stars": 34,
            "cumulative_unique_labels": len(reviewed_labels),
            "cumulative_unique_stars": len(reviewed_stars),
            "h_high_complete": True,
            "h_high_labels": len(human_rows["label_high"]),
            "h_high_stars": len(human_rows["star_high"]),
            "h_medium_complete": medium_complete,
            "pending_medium_labels": sum(row["priority"] == "MEDIUM" and row["candidate_id"] not in reviewed_labels for row in labels),
            "pending_medium_stars": sum(row["priority"] == "MEDIUM" and row["candidate_id"] not in reviewed_stars for row in stars),
        }
    (PKG / "BUILD_SUMMARY.json").write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
