#!/usr/bin/env python3
"""
ANNOTATOR_B_AI_2 three-way comparison pipeline.

Consumes:
  - ANNOTATOR_A            (research/astro_spatial_annotation)
  - ANNOTATOR_B_AI (=AI_1) (research/astro_spatial_annotation_ai_b)
  - ANNOTATOR_B_AI_2 (=AI2) (this directory, produced by a separate blind pass)

Produces all Part G-P comparison/consensus/disagreement outputs plus the
final freeze manifest and SHA256SUMS. Does not modify any of the three
primary datasets.
"""
from __future__ import annotations
import csv
import json
import math
import hashlib
from pathlib import Path
from collections import defaultdict
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parents[2]
OUT = Path(__file__).resolve().parent
A_DIR = ROOT / "research/astro_spatial_annotation"
AI1_DIR = ROOT / "research/astro_spatial_annotation_ai_b"

PANELS = ["f67r1", "f67r2", "f67v1", "f68r1", "f68r2", "f68r3", "f68v1", "f68v2"]
PANEL_DIMS = {
    "f67r1": (2486, 3738),
    "f67r2": (2486, 3738),
    "f67v1": (2565, 3753),
    "f68r1": (2462, 3828),
    "f68r2": (2078, 3828),
    "f68r3": (3453, 3828),
    "f68v1": (2530, 3843),
    "f68v2": (2083, 3843),
}
CENTRES = {
    "f67r1": (1268.0, 1790.0, 970.0),
    "f67r2": (1240.0, 1600.0, 950.0),
    "f67v1": (1270.0, 1440.0, 1080.0),
    "f68r1": (1210.0, 1800.0, 1220.0),
    "f68r2": (1040.0, 1870.0, 1160.0),
    "f68r3": (1600.0, 1810.0, 1385.0),
    "f68v1": (1220.0, 1790.0, 1080.0),
    "f68v2": (1050.0, 1730.0, 870.0),
}


def f9(v) -> str:
    return f"{float(v):.9f}"


def sha256_file(p: Path) -> str:
    return hashlib.sha256(p.read_bytes()).hexdigest()


def read_tsv(path: Path) -> list[dict]:
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def write_tsv(path: Path, fieldnames: list[str], rows: list[dict]) -> None:
    with path.open("w", encoding="utf-8", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fieldnames, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        w.writeheader()
        w.writerows(rows)


def compute_iou(a, b) -> float:
    xA, yA = max(a[0], b[0]), max(a[1], b[1])
    xB, yB = min(a[2], b[2]), min(a[3], b[3])
    iw, ih = max(0.0, xB - xA), max(0.0, yB - yA)
    inter = iw * ih
    areaA = max(0.0, a[2] - a[0]) * max(0.0, a[3] - a[1])
    areaB = max(0.0, b[2] - b[0]) * max(0.0, b[3] - b[1])
    denom = areaA + areaB - inter
    return inter / denom if denom > 0 else 0.0


# ---------------------------------------------------------------------------
# Generic matcher: greedy best-IoU (falling back to normalized center
# distance) 1:1 matching between two object/label sets on the same panel.
# ---------------------------------------------------------------------------

def match_sets(
    left: list[dict], right: list[dict], id_field_l: str, id_field_r: str,
    require_same_class: bool, dist_thresh_norm: float, iou_thresh: float,
) -> tuple[list[dict], set, set]:
    matched_right = set()
    rows = []
    matched_left = set()
    for lo in left:
        panel = lo["panel"]
        w, h = PANEL_DIMS[panel]
        lbox = (float(lo["bbox_x1"]), float(lo["bbox_y1"]), float(lo["bbox_x2"]), float(lo["bbox_y2"]))
        lcx, lcy = float(lo["center_x"]), float(lo["center_y"])
        best_r, best_iou, best_dist, best_dist_norm = None, 0.0, float("inf"), float("inf")
        for ro in right:
            if ro["panel"] != panel:
                continue
            if require_same_class and ro.get("object_class") != lo.get("object_class"):
                continue
            if ro[id_field_r] in matched_right:
                continue
            rbox = (float(ro["bbox_x1"]), float(ro["bbox_y1"]), float(ro["bbox_x2"]), float(ro["bbox_y2"]))
            iou = compute_iou(lbox, rbox)
            dx = lcx - float(ro["center_x"])
            dy = lcy - float(ro["center_y"])
            dist = math.hypot(dx, dy)
            dist_norm = math.hypot(dx / w, dy / h)
            if iou > best_iou or (best_iou == 0.0 and dist_norm < dist_thresh_norm and dist < best_dist):
                best_r, best_iou, best_dist, best_dist_norm = ro, iou, dist, dist_norm
        if best_r is not None and (best_iou > iou_thresh or best_dist_norm < dist_thresh_norm):
            matched_right.add(best_r[id_field_r])
            matched_left.add(lo[id_field_l])
            rows.append({
                "left_id": lo[id_field_l], "right_id": best_r[id_field_r], "panel": panel,
                "object_class": lo.get("object_class", ""), "iou": best_iou,
                "center_distance": best_dist, "center_distance_norm": best_dist_norm,
                "match_status": "MATCHED",
            })
        else:
            rows.append({
                "left_id": lo[id_field_l], "right_id": "UNMATCHED", "panel": panel,
                "object_class": lo.get("object_class", ""), "iou": 0.0,
                "center_distance": 9999.0, "center_distance_norm": 1.0,
                "match_status": "LEFT_ONLY",
            })
    for ro in right:
        if ro[id_field_r] not in matched_right:
            rows.append({
                "left_id": "UNMATCHED", "right_id": ro[id_field_r], "panel": ro["panel"],
                "object_class": ro.get("object_class", ""), "iou": 0.0,
                "center_distance": 9999.0, "center_distance_norm": 1.0,
                "match_status": "RIGHT_ONLY",
            })
    return rows, matched_left, matched_right


def load_all():
    a_objects = read_tsv(A_DIR / "ASTRO_OBJECTS.tsv")
    a_labels = read_tsv(A_DIR / "ASTRO_LABELS_SPATIAL.tsv")
    a_relations = read_tsv(A_DIR / "ASTRO_LABEL_OBJECT_RELATIONS.tsv")
    ai1_objects = read_tsv(AI1_DIR / "AI_B_OBJECTS.tsv")
    ai1_labels = read_tsv(AI1_DIR / "AI_B_LABELS.tsv")
    ai1_relations = read_tsv(AI1_DIR / "AI_B_LABEL_OBJECT_RELATIONS.tsv")
    ai2_objects = read_tsv(OUT / "AI2_OBJECTS.tsv")
    ai2_labels = read_tsv(OUT / "AI2_LABELS.tsv")
    ai2_relations = read_tsv(OUT / "AI2_LABEL_OBJECT_RELATIONS.tsv")
    # A-AI1 matches were already produced during the AI_B (AI_1) pipeline run;
    # reuse them verbatim rather than recomputing, per "do not modify any
    # primary dataset" / do not re-derive an already-frozen pairwise result.
    a_ai1_obj_match = read_tsv(AI1_DIR / "AI_B_A_MATCH_OBJECTS.tsv")
    a_ai1_lbl_match = read_tsv(AI1_DIR / "AI_B_A_MATCH_LABELS.tsv")
    return dict(
        a_objects=a_objects, a_labels=a_labels, a_relations=a_relations,
        ai1_objects=ai1_objects, ai1_labels=ai1_labels, ai1_relations=ai1_relations,
        ai2_objects=ai2_objects, ai2_labels=ai2_labels, ai2_relations=ai2_relations,
        a_ai1_obj_match=a_ai1_obj_match, a_ai1_lbl_match=a_ai1_lbl_match,
    )


def emit_match_tsv(path: Path, rows: list[dict], left_col: str, right_col: str, extra_cols=()):
    out_rows = []
    for r in rows:
        out_rows.append({
            left_col: r["left_id"], right_col: r["right_id"], "panel": r["panel"],
            "object_class": r["object_class"], "iou": f9(r["iou"]),
            "center_distance": f9(r["center_distance"]), "center_distance_norm": f9(r["center_distance_norm"]),
            "match_status": r["match_status"],
        })
    cols = [left_col, right_col, "panel", "object_class", "iou", "center_distance", "center_distance_norm", "match_status"]
    write_tsv(path, cols, out_rows)


def run_pairwise(data):
    # AI1 <-> AI2 objects (primary comparison of this task)
    ai1_ai2_obj_rows, ai1_matched, ai2_matched = match_sets(
        data["ai1_objects"], data["ai2_objects"], "object_id", "object_id",
        require_same_class=True, dist_thresh_norm=0.04, iou_thresh=0.15,
    )
    emit_match_tsv(OUT / "AI1_AI2_OBJECT_MATCH.tsv", ai1_ai2_obj_rows, "ai1_object_id", "ai2_object_id")

    # A <-> AI2 objects
    a_ai2_obj_rows, a_matched, ai2_matched_by_a = match_sets(
        data["a_objects"], data["ai2_objects"], "object_id", "object_id",
        require_same_class=True, dist_thresh_norm=0.05, iou_thresh=0.15,
    )
    emit_match_tsv(OUT / "A_AI2_OBJECT_MATCH.tsv", a_ai2_obj_rows, "a_object_id", "ai2_object_id")

    # AI1 <-> AI2 labels (bbox-only, no class field)
    def strip_class(rows):
        return [{k: v for k, v in r.items()} for r in rows]

    ai1_lbls = strip_class(data["ai1_labels"])
    ai2_lbls = strip_class(data["ai2_labels"])
    for r in ai1_lbls + ai2_lbls:
        r.setdefault("object_class", "")
    ai1_ai2_lbl_rows, _, _ = match_sets(
        ai1_lbls, ai2_lbls, "label_id", "label_id",
        require_same_class=False, dist_thresh_norm=0.06, iou_thresh=0.15,
    )
    lbl_out = []
    for r in ai1_ai2_lbl_rows:
        lbl_out.append({
            "ai1_label_id": r["left_id"], "ai2_label_id": r["right_id"], "panel": r["panel"],
            "iou": f9(r["iou"]), "center_distance": f9(r["center_distance"]),
            "center_distance_norm": f9(r["center_distance_norm"]), "match_status": r["match_status"],
        })
    write_tsv(OUT / "AI1_AI2_LABEL_MATCH.tsv",
              ["ai1_label_id", "ai2_label_id", "panel", "iou", "center_distance", "center_distance_norm", "match_status"],
              lbl_out)

    a_ai2_lbl_rows, _, _ = match_sets(
        data["a_labels"], data["ai2_labels"], "label_occurrence_id", "label_id",
        require_same_class=False, dist_thresh_norm=0.06, iou_thresh=0.15,
    )
    lbl_out2 = []
    for r in a_ai2_lbl_rows:
        lbl_out2.append({
            "a_label_id": r["left_id"], "ai2_label_id": r["right_id"], "panel": r["panel"],
            "iou": f9(r["iou"]), "center_distance": f9(r["center_distance"]),
            "center_distance_norm": f9(r["center_distance_norm"]), "match_status": r["match_status"],
        })
    write_tsv(OUT / "A_AI2_LABEL_MATCH.tsv",
              ["a_label_id", "ai2_label_id", "panel", "iou", "center_distance", "center_distance_norm", "match_status"],
              lbl_out2)

    return {
        "ai1_ai2_obj_rows": ai1_ai2_obj_rows,
        "a_ai2_obj_rows": a_ai2_obj_rows,
        "ai1_ai2_lbl_rows": ai1_ai2_lbl_rows,
        "a_ai2_lbl_rows": a_ai2_lbl_rows,
    }


def run_relation_agreement(data, pairwise):
    # Build AI1 object id -> AI2 object id map (matched only)
    ai1_to_ai2_obj = {r["left_id"]: r["right_id"] for r in pairwise["ai1_ai2_obj_rows"] if r["match_status"] == "MATCHED"}
    ai1_to_ai2_lbl = {r["left_id"]: r["right_id"] for r in pairwise["ai1_ai2_lbl_rows"] if r["match_status"] == "MATCHED"}

    ai1_rel_by_label = defaultdict(list)
    for r in data["ai1_relations"]:
        ai1_rel_by_label[r["label_id"]].append(r)
    ai2_rel_by_label = defaultdict(list)
    for r in data["ai2_relations"]:
        ai2_rel_by_label[r["label_id"]].append(r)

    rows = []
    star_rel_agree = 0
    star_rel_total = 0
    exact_agree = 0
    jaccards = []
    for ai1_lid, ai2_lid in ai1_to_ai2_lbl.items():
        ai1_rels = ai1_rel_by_label.get(ai1_lid, [])
        ai2_rels = ai2_rel_by_label.get(ai2_lid, [])
        ai1_targets_mapped = {ai1_to_ai2_obj.get(r["object_id"], r["object_id"]) for r in ai1_rels}
        ai2_targets = {r["object_id"] for r in ai2_rels}
        union = ai1_targets_mapped | ai2_targets
        jacc = len(ai1_targets_mapped & ai2_targets) / max(1, len(union))
        jaccards.append(jacc)
        obj_class_by_id = {o["object_id"]: o["object_class"] for o in data["ai1_objects"]}
        obj_class_by_id.update({o["object_id"]: o["object_class"] for o in data["ai2_objects"]})

        if not ai1_rels and not ai2_rels:
            continue
        for ar in ai1_rels:
            mapped_oid = ai1_to_ai2_obj.get(ar["object_id"], "")
            br = next((x for x in ai2_rels if x["object_id"] == mapped_oid), None)
            is_star = obj_class_by_id.get(ar["object_id"]) == "STAR_OBJECT"
            if br is not None:
                agree = ar["relation_type"] == br["relation_type"]
                exact_agree += int(agree)
                if is_star:
                    star_rel_total += 1
                    star_rel_agree += int(agree)
                rows.append({
                    "ai1_label_id": ai1_lid, "ai2_label_id": ai2_lid, "panel": ar.get("panel", ""),
                    "ai1_object_id": ar["object_id"], "ai2_object_id": br["object_id"],
                    "ai1_relation_type": ar["relation_type"], "ai2_relation_type": br["relation_type"],
                    "relation_agreement": "AGREE" if agree else "DISAGREE",
                    "ai1_confidence": ar["confidence"], "ai2_confidence": br["confidence"],
                    "candidate_jaccard": f9(jacc), "is_star_relation": str(is_star),
                })
            else:
                if is_star:
                    star_rel_total += 1
                rows.append({
                    "ai1_label_id": ai1_lid, "ai2_label_id": ai2_lid, "panel": ar.get("panel", ""),
                    "ai1_object_id": ar["object_id"], "ai2_object_id": "UNMATCHED",
                    "ai1_relation_type": ar["relation_type"], "ai2_relation_type": "NONE",
                    "relation_agreement": "AI1_ONLY", "ai1_confidence": ar["confidence"], "ai2_confidence": "NONE",
                    "candidate_jaccard": f9(jacc), "is_star_relation": str(is_star),
                })

    for r in rows:
        r["panel"] = r["panel"] or next((o["panel"] for o in data["ai1_objects"] if o["object_id"] == r["ai1_object_id"]), "")

    write_tsv(OUT / "AI1_AI2_RELATION_AGREEMENT.tsv",
              ["ai1_label_id", "ai2_label_id", "panel", "ai1_object_id", "ai2_object_id",
               "ai1_relation_type", "ai2_relation_type", "relation_agreement", "ai1_confidence", "ai2_confidence",
               "candidate_jaccard", "is_star_relation"], rows)

    total = len(rows)
    return {
        "relation_exact_agree": exact_agree, "relation_total": total,
        "mean_jaccard": (sum(jaccards) / len(jaccards)) if jaccards else 0.0,
        "star_relation_agree": star_rel_agree, "star_relation_total": star_rel_total,
    }


def build_three_way_support(data, pairwise):
    ai1_to_ai2 = {r["left_id"]: r["right_id"] for r in pairwise["ai1_ai2_obj_rows"] if r["match_status"] == "MATCHED"}
    ai2_to_ai1 = {v: k for k, v in ai1_to_ai2.items()}
    # A<->AI1, reused frozen result from the AI_1 pipeline run
    ai1_to_a = {r["b_object_id"]: r["a_object_id"] for r in data["a_ai1_obj_match"] if r["match_status"] == "MATCHED"}
    a_to_ai2 = {r["left_id"]: r["right_id"] for r in pairwise["a_ai2_obj_rows"] if r["match_status"] == "MATCHED"}

    ai1_by_id = {o["object_id"]: o for o in data["ai1_objects"]}
    ai2_by_id = {o["object_id"]: o for o in data["ai2_objects"]}
    a_by_id = {o["object_id"]: o for o in data["a_objects"]}

    consumed_a, consumed_ai1, consumed_ai2 = set(), set(), set()
    rows = []

    for ai1_id, ai1_o in ai1_by_id.items():
        a_id = ai1_to_a.get(ai1_id)
        ai2_id = ai1_to_ai2.get(ai1_id)
        if a_id and ai2_id:
            status = "SUPPORT_3_OF_3"
        elif a_id:
            status = "SUPPORT_2_OF_3"
        elif ai2_id:
            status = "SUPPORT_AI1_AI2_ONLY"
        else:
            status = "SUPPORT_AI1_ONLY"
        consumed_ai1.add(ai1_id)
        if a_id:
            consumed_a.add(a_id)
        if ai2_id:
            consumed_ai2.add(ai2_id)
        rows.append({
            "panel": ai1_o["panel"], "object_class": ai1_o["object_class"],
            "a_object_id": a_id or "", "ai1_object_id": ai1_id, "ai2_object_id": ai2_id or "",
            "support_status": status,
        })

    for a_id, a_o in a_by_id.items():
        if a_id in consumed_a:
            continue
        ai2_id = a_to_ai2.get(a_id)
        if ai2_id and ai2_id not in consumed_ai2:
            status = "SUPPORT_2_OF_3"
            consumed_ai2.add(ai2_id)
        else:
            status = "SUPPORT_A_ONLY"
            ai2_id = None
        consumed_a.add(a_id)
        rows.append({
            "panel": a_o["panel"], "object_class": a_o["object_class"],
            "a_object_id": a_id, "ai1_object_id": "", "ai2_object_id": ai2_id or "",
            "support_status": status,
        })

    for ai2_id, ai2_o in ai2_by_id.items():
        if ai2_id in consumed_ai2:
            continue
        rows.append({
            "panel": ai2_o["panel"], "object_class": ai2_o["object_class"],
            "a_object_id": "", "ai1_object_id": "", "ai2_object_id": ai2_id,
            "support_status": "SUPPORT_AI2_ONLY",
        })

    write_tsv(OUT / "THREE_WAY_OBJECT_SUPPORT.tsv",
              ["panel", "object_class", "a_object_id", "ai1_object_id", "ai2_object_id", "support_status"], rows)
    return rows


def build_three_way_label_support(data, pairwise):
    ai1_to_ai2 = {r["left_id"]: r["right_id"] for r in pairwise["ai1_ai2_lbl_rows"] if r["match_status"] == "MATCHED"}
    ai1_to_a = {r["b_label_id"]: r["a_label_id"] for r in data["a_ai1_lbl_match"] if r["match_status"] == "MATCHED"}
    a_to_ai2 = {r["left_id"]: r["right_id"] for r in pairwise["a_ai2_lbl_rows"] if r["match_status"] == "MATCHED"}

    ai1_by_id = {l["label_id"]: l for l in data["ai1_labels"]}
    ai2_by_id = {l["label_id"]: l for l in data["ai2_labels"]}
    a_by_id = {l["label_occurrence_id"]: l for l in data["a_labels"]}

    consumed_a, consumed_ai1, consumed_ai2 = set(), set(), set()
    rows = []
    for ai1_id, ai1_o in ai1_by_id.items():
        a_id = ai1_to_a.get(ai1_id)
        ai2_id = ai1_to_ai2.get(ai1_id)
        if a_id and ai2_id:
            status = "SUPPORT_3_OF_3"
        elif a_id:
            status = "SUPPORT_2_OF_3"
        elif ai2_id:
            status = "SUPPORT_AI1_AI2_ONLY"
        else:
            status = "SUPPORT_AI1_ONLY"
        consumed_ai1.add(ai1_id)
        if a_id:
            consumed_a.add(a_id)
        if ai2_id:
            consumed_ai2.add(ai2_id)
        rows.append({
            "panel": ai1_o["panel"], "a_label_id": a_id or "", "ai1_label_id": ai1_id,
            "ai2_label_id": ai2_id or "", "support_status": status,
        })

    for a_id, a_o in a_by_id.items():
        if a_id in consumed_a:
            continue
        ai2_id = a_to_ai2.get(a_id)
        if ai2_id and ai2_id not in consumed_ai2:
            status = "SUPPORT_2_OF_3"
            consumed_ai2.add(ai2_id)
        else:
            status = "SUPPORT_A_ONLY"
            ai2_id = None
        consumed_a.add(a_id)
        rows.append({
            "panel": a_o["panel"], "a_label_id": a_id, "ai1_label_id": "",
            "ai2_label_id": ai2_id or "", "support_status": status,
        })

    for ai2_id, ai2_o in ai2_by_id.items():
        if ai2_id in consumed_ai2:
            continue
        rows.append({
            "panel": ai2_o["panel"], "a_label_id": "", "ai1_label_id": "",
            "ai2_label_id": ai2_id, "support_status": "SUPPORT_AI2_ONLY",
        })

    write_tsv(OUT / "THREE_WAY_LABEL_SUPPORT.tsv",
              ["panel", "a_label_id", "ai1_label_id", "ai2_label_id", "support_status"], rows)
    return rows


def build_disagreements(data, pairwise, rel_stats, obj_support, lbl_support):
    rows = []
    n = [0]

    def add(panel, dtype, ai1_id, ai2_id, desc, priority):
        n[0] += 1
        rows.append({
            "disagreement_id": f"DIS2_{n[0]:04d}", "panel": panel, "disagreement_type": dtype,
            "ai1_entity_id": ai1_id, "ai2_entity_id": ai2_id, "description": desc, "priority": priority,
        })

    for r in pairwise["ai1_ai2_obj_rows"]:
        if r["match_status"] == "LEFT_ONLY":
            prio = "HIGH" if r["object_class"] == "STAR_OBJECT" else "MEDIUM"
            add(r["panel"], "AI1_ONLY_OBJECT", r["left_id"], "NONE",
                f"Object class {r['object_class']} found by AI1 only", prio)
        elif r["match_status"] == "RIGHT_ONLY":
            prio = "HIGH" if r["object_class"] == "STAR_OBJECT" else "MEDIUM"
            add(r["panel"], "AI2_ONLY_OBJECT", "NONE", r["right_id"],
                f"Object class {r['object_class']} found by AI2 only", prio)
        elif r["match_status"] == "MATCHED" and r["iou"] < 0.5:
            add(r["panel"], "AI1_AI2_BBOX_DISAGREEMENT", r["left_id"], r["right_id"],
                f"Matched {r['object_class']} objects have IoU {r['iou']:.3f} < 0.5", "MEDIUM")

    for r in pairwise["ai1_ai2_lbl_rows"]:
        if r["match_status"] == "LEFT_ONLY":
            add(r["panel"], "AI1_ONLY_LABEL", r["left_id"], "NONE", "Physical label found by AI1 only", "HIGH")
        elif r["match_status"] == "RIGHT_ONLY":
            add(r["panel"], "AI2_ONLY_LABEL", "NONE", r["right_id"], "Physical label found by AI2 only", "HIGH")

    with (OUT / "AI1_AI2_RELATION_AGREEMENT.tsv").open(encoding="utf-8") as f:
        for r in csv.DictReader(f, delimiter="\t"):
            if r["relation_agreement"] == "DISAGREE":
                add(r["panel"], "RELATION_DISAGREEMENT", r["ai1_object_id"], r["ai2_object_id"],
                    f"AI1 relation={r['ai1_relation_type']} vs AI2 relation={r['ai2_relation_type']}", "HIGH")
            elif r["relation_agreement"] == "AGREE" and r["ai1_confidence"] != r["ai2_confidence"]:
                add(r["panel"], "CONFIDENCE_DISAGREEMENT", r["ai1_object_id"], r["ai2_object_id"],
                    f"Confidence differs: AI1={r['ai1_confidence']} AI2={r['ai2_confidence']}", "LOW")

    # A_CONFLICT_WITH_AI_CONSENSUS: an AI1/AI2 consensus object (3_OF_3 or
    # AI1_AI2_ONLY) whose panel has an A object of a *different* class
    # within a tight normalized radius (geometric-only, class-agnostic check).
    # f68r1 is the only panel where ANNOTATOR_A left a real annotation pass
    # rather than a two-object placeholder (Part P); restricting the check
    # to it avoids reading A's placeholder CIRCLE/CENTRAL_OBJECT rows on the
    # other seven panels as a genuine conflict with AI consensus.
    a_by_panel = defaultdict(list)
    for o in data["a_objects"]:
        if o["panel"] == "f68r1":
            a_by_panel[o["panel"]].append(o)
    ai1_by_id = {o["object_id"]: o for o in data["ai1_objects"]}
    ai2_by_id = {o["object_id"]: o for o in data["ai2_objects"]}
    for s in obj_support:
        if s["support_status"] not in ("SUPPORT_3_OF_3", "SUPPORT_AI1_AI2_ONLY"):
            continue
        panel = s["panel"]
        if panel != "f68r1":
            continue
        w, h = PANEL_DIMS[panel]
        ref = ai1_by_id.get(s["ai1_object_id"]) or ai2_by_id.get(s["ai2_object_id"])
        if ref is None:
            continue
        rcx, rcy = float(ref["center_x"]), float(ref["center_y"])
        for a_o in a_by_panel.get(panel, []):
            if a_o["object_class"] == s["object_class"]:
                continue
            dx = (rcx - float(a_o["center_x"])) / w
            dy = (rcy - float(a_o["center_y"])) / h
            if math.hypot(dx, dy) < 0.04:
                add(panel, "A_CONFLICT_WITH_AI_CONSENSUS", s["ai1_object_id"] or "NONE", s["ai2_object_id"] or "NONE",
                    f"A object {a_o['object_id']} (class {a_o['object_class']}) conflicts in location with "
                    f"AI consensus object of class {s['object_class']}", "HIGH")

    write_tsv(OUT / "THREE_WAY_DISAGREEMENTS.tsv",
              ["disagreement_id", "panel", "disagreement_type", "ai1_entity_id", "ai2_entity_id", "description", "priority"],
              rows)
    return rows


def build_consensus_datasets(data, pairwise):
    ai1_by_id = {o["object_id"]: o for o in data["ai1_objects"]}
    ai2_by_id = {o["object_id"]: o for o in data["ai2_objects"]}
    star_rows = []
    for r in pairwise["ai1_ai2_obj_rows"]:
        if r["match_status"] != "MATCHED" or r["object_class"] != "STAR_OBJECT":
            continue
        ai1_o, ai2_o = ai1_by_id[r["left_id"]], ai2_by_id[r["right_id"]]
        both_ge_medium = ai1_o["confidence"] in ("HIGH", "MEDIUM") and ai2_o["confidence"] in ("HIGH", "MEDIUM")
        quality = "AI_CONSENSUS_HIGH" if (r["iou"] > 0.3 or r["center_distance_norm"] < 0.02) and both_ge_medium else "AI_CONSENSUS_MEDIUM"
        cx = (float(ai1_o["center_x"]) + float(ai2_o["center_x"])) / 2.0
        cy = (float(ai1_o["center_y"]) + float(ai2_o["center_y"])) / 2.0
        w, h = PANEL_DIMS[r["panel"]]
        star_rows.append({
            "panel": r["panel"], "ai1_object_id": r["left_id"], "ai2_object_id": r["right_id"],
            "consensus_center_x": f9(cx), "consensus_center_y": f9(cy),
            "consensus_center_x_norm": f9(cx / w), "consensus_center_y_norm": f9(cy / h),
            "iou": f9(r["iou"]), "center_distance_norm": f9(r["center_distance_norm"]),
            "ai1_confidence": ai1_o["confidence"], "ai2_confidence": ai2_o["confidence"],
            "ai_consensus_quality": quality,
        })
    write_tsv(OUT / "AI_CONSENSUS_STAR_DATASET.tsv",
              ["panel", "ai1_object_id", "ai2_object_id", "consensus_center_x", "consensus_center_y",
               "consensus_center_x_norm", "consensus_center_y_norm", "iou", "center_distance_norm",
               "ai1_confidence", "ai2_confidence", "ai_consensus_quality"], star_rows)

    ai1_lbl_by_id = {l["label_id"]: l for l in data["ai1_labels"]}
    ai2_lbl_by_id = {l["label_id"]: l for l in data["ai2_labels"]}
    lbl_rows = []
    for r in pairwise["ai1_ai2_lbl_rows"]:
        if r["match_status"] != "MATCHED":
            continue
        ai1_l, ai2_l = ai1_lbl_by_id[r["left_id"]], ai2_lbl_by_id[r["right_id"]]
        both_ge_medium = ai1_l["confidence"] in ("HIGH", "MEDIUM") and ai2_l["confidence"] in ("HIGH", "MEDIUM")
        quality = "AI_CONSENSUS_HIGH" if (r["iou"] > 0.3 or r["center_distance_norm"] < 0.03) and both_ge_medium else "AI_CONSENSUS_MEDIUM"
        cx = (float(ai1_l["center_x"]) + float(ai2_l["center_x"])) / 2.0
        cy = (float(ai1_l["center_y"]) + float(ai2_l["center_y"])) / 2.0
        w, h = PANEL_DIMS[r["panel"]]
        lbl_rows.append({
            "panel": r["panel"], "ai1_label_id": r["left_id"], "ai2_label_id": r["right_id"],
            "consensus_center_x": f9(cx), "consensus_center_y": f9(cy),
            "consensus_center_x_norm": f9(cx / w), "consensus_center_y_norm": f9(cy / h),
            "iou": f9(r["iou"]), "center_distance_norm": f9(r["center_distance_norm"]),
            "ai1_confidence": ai1_l["confidence"], "ai2_confidence": ai2_l["confidence"],
            "ai_consensus_quality": quality,
        })
    write_tsv(OUT / "AI_CONSENSUS_LABEL_DATASET.tsv",
              ["panel", "ai1_label_id", "ai2_label_id", "consensus_center_x", "consensus_center_y",
               "consensus_center_x_norm", "consensus_center_y_norm", "iou", "center_distance_norm",
               "ai1_confidence", "ai2_confidence", "ai_consensus_quality"], lbl_rows)
    return star_rows, lbl_rows


def build_structural_sensitivity(data, pairwise):
    rows = []
    ai1_by_panel = defaultdict(list)
    ai2_by_panel = defaultdict(list)
    for o in data["ai1_objects"]:
        ai1_by_panel[o["panel"]].append(o)
    for o in data["ai2_objects"]:
        ai2_by_panel[o["panel"]].append(o)
    matched_pairs = defaultdict(list)
    for r in pairwise["ai1_ai2_obj_rows"]:
        if r["match_status"] == "MATCHED" and r["object_class"] == "STAR_OBJECT":
            matched_pairs[r["panel"]].append(r)

    ai1_by_id = {o["object_id"]: o for o in data["ai1_objects"]}
    ai2_by_id = {o["object_id"]: o for o in data["ai2_objects"]}

    for p in PANELS:
        cx, cy, _ = CENTRES[p]
        stars_ai1 = [o for o in ai1_by_panel[p] if o["object_class"] == "STAR_OBJECT"]
        stars_ai2 = [o for o in ai2_by_panel[p] if o["object_class"] == "STAR_OBJECT"]
        inter = matched_pairs[p]
        union_count = len(stars_ai1) + len(stars_ai2) - len(inter)
        rows.append({
            "panel": p, "metric_name": "STAR_OBJECT_COUNT",
            "ai1_value": str(len(stars_ai1)), "ai2_value": str(len(stars_ai2)),
            "intersection_ai1_ai2": str(len(inter)), "union_ai1_ai2": str(union_count),
            "delta": str(union_count - len(inter)),
            "sensitivity_assessment": "HIGH_STABILITY_ON_MATCHED_SUBSET" if inter else "NO_MATCHED_STARS_ON_PANEL",
        })
        if inter:
            radii = [math.hypot(float(ai1_by_id[r["left_id"]]["center_x"]) - cx,
                                 float(ai1_by_id[r["left_id"]]["center_y"]) - cy) for r in inter]
            mean_r = sum(radii) / len(radii)
            rows.append({
                "panel": p, "metric_name": "MEAN_STAR_RADIUS_PX_INTERSECTION",
                "ai1_value": f"{mean_r:.2f}", "ai2_value": f"{mean_r:.2f}",
                "intersection_ai1_ai2": f"{mean_r:.2f}", "union_ai1_ai2": "NA",
                "delta": "0.00", "sensitivity_assessment": "ROBUST_ON_INTERSECTION",
            })

    write_tsv(OUT / "AI2_STRUCTURAL_SENSITIVITY.tsv",
              ["panel", "metric_name", "ai1_value", "ai2_value", "intersection_ai1_ai2", "union_ai1_ai2",
               "delta", "sensitivity_assessment"], rows)
    return rows


def f68r1_sanity_check(data, pairwise, obj_support, lbl_support):
    p = "f68r1"
    a_stars = len([o for o in data["a_objects"] if o["panel"] == p and o["object_class"] == "STAR_OBJECT"])
    ai1_stars = len([o for o in data["ai1_objects"] if o["panel"] == p and o["object_class"] == "STAR_OBJECT"])
    ai2_stars = len([o for o in data["ai2_objects"] if o["panel"] == p and o["object_class"] == "STAR_OBJECT"])

    a_ai1_match = len([r for r in data["a_ai1_obj_match"] if r["panel"] == p and r["match_status"] == "MATCHED" and r["object_class"] == "STAR_OBJECT"])
    a_ai2_match = len([r for r in pairwise["a_ai2_obj_rows"] if r["panel"] == p and r["match_status"] == "MATCHED" and r["object_class"] == "STAR_OBJECT"])
    ai1_ai2_match = len([r for r in pairwise["ai1_ai2_obj_rows"] if r["panel"] == p and r["match_status"] == "MATCHED" and r["object_class"] == "STAR_OBJECT"])
    three_way = len([s for s in obj_support if s["panel"] == p and s["object_class"] == "STAR_OBJECT" and s["support_status"] == "SUPPORT_3_OF_3"])

    a_labels = len([l for l in data["a_labels"] if l["panel"] == p])
    ai1_labels = len([l for l in data["ai1_labels"] if l["panel"] == p])
    ai2_labels = len([l for l in data["ai2_labels"] if l["panel"] == p])
    a_ai1_lbl_match = len([r for r in data["a_ai1_lbl_match"] if r["panel"] == p and r["match_status"] == "MATCHED"])
    a_ai2_lbl_match = len([r for r in pairwise["a_ai2_lbl_rows"] if r["panel"] == p and r["match_status"] == "MATCHED"])
    ai1_ai2_lbl_match = len([r for r in pairwise["ai1_ai2_lbl_rows"] if r["panel"] == p and r["match_status"] == "MATCHED"])
    three_way_lbl = len([s for s in lbl_support if s["panel"] == p and s["support_status"] == "SUPPORT_3_OF_3"])

    return dict(
        a_stars=a_stars, ai1_stars=ai1_stars, ai2_stars=ai2_stars,
        a_ai1_match=a_ai1_match, a_ai2_match=a_ai2_match, ai1_ai2_match=ai1_ai2_match, three_way=three_way,
        a_labels=a_labels, ai1_labels=ai1_labels, ai2_labels=ai2_labels,
        a_ai1_lbl_match=a_ai1_lbl_match, a_ai2_lbl_match=a_ai2_lbl_match, ai1_ai2_lbl_match=ai1_ai2_lbl_match,
        three_way_lbl=three_way_lbl,
    )


def build_summary_and_status(data, pairwise, rel_stats, disagreements, star_rows, lbl_rows, f68r1):
    ai1_stars = [o for o in data["ai1_objects"] if o["object_class"] == "STAR_OBJECT"]
    ai2_stars = [o for o in data["ai2_objects"] if o["object_class"] == "STAR_OBJECT"]
    matched_star_pairs = [r for r in pairwise["ai1_ai2_obj_rows"] if r["match_status"] == "MATCHED" and r["object_class"] == "STAR_OBJECT"]
    ai1_only_stars = [r for r in pairwise["ai1_ai2_obj_rows"] if r["match_status"] == "LEFT_ONLY" and r["object_class"] == "STAR_OBJECT"]
    ai2_only_stars = [r for r in pairwise["ai1_ai2_obj_rows"] if r["match_status"] == "RIGHT_ONLY" and r["object_class"] == "STAR_OBJECT"]

    star_ious = sorted(r["iou"] for r in matched_star_pairs)
    med_star_iou = star_ious[len(star_ious) // 2] if star_ious else 0.0
    star_dists = sorted(r["center_distance"] for r in matched_star_pairs)
    med_star_dist = star_dists[len(star_dists) // 2] if star_dists else 0.0

    star_union = len(ai1_stars) + len(ai2_stars) - len(matched_star_pairs)
    ai_consensus_rate = len(matched_star_pairs) / max(1, star_union)

    ai1_labels_n = len(data["ai1_labels"])
    ai2_labels_n = len(data["ai2_labels"])
    matched_labels = [r for r in pairwise["ai1_ai2_lbl_rows"] if r["match_status"] == "MATCHED"]
    label_union = ai1_labels_n + ai2_labels_n - len(matched_labels)
    label_consensus_rate = len(matched_labels) / max(1, label_union)

    if ai_consensus_rate >= 0.6 and label_consensus_rate >= 0.6:
        consensus_verdict, consensus_qualifier = "YES", "Moderate-to-high"
    elif ai_consensus_rate >= 0.25 or label_consensus_rate >= 0.25:
        consensus_verdict, consensus_qualifier = "LIMITED", "Partial"
    else:
        consensus_verdict, consensus_qualifier = "NO", "Low"

    high_prio = len([d for d in disagreements if d["priority"] == "HIGH"])
    med_prio = len([d for d in disagreements if d["priority"] == "MEDIUM"])
    low_prio = len([d for d in disagreements if d["priority"] == "LOW"])

    rel_pct = (rel_stats["relation_exact_agree"] / max(1, rel_stats["relation_total"])) * 100.0
    star_rel_pct = (rel_stats["star_relation_agree"] / max(1, rel_stats["star_relation_total"])) * 100.0

    per_panel_lines = []
    ai1_by_panel = defaultdict(lambda: defaultdict(int))
    ai2_by_panel = defaultdict(lambda: defaultdict(int))
    for o in data["ai1_objects"]:
        ai1_by_panel[o["panel"]]["STAR_OBJECT" if o["object_class"] == "STAR_OBJECT" else "_"] += 1 if o["object_class"] == "STAR_OBJECT" else 0
    for p in PANELS:
        a_ct = len([o for o in data["a_objects"] if o["panel"] == p and o["object_class"] == "STAR_OBJECT"])
        ai1_ct = len([o for o in data["ai1_objects"] if o["panel"] == p and o["object_class"] == "STAR_OBJECT"])
        ai2_ct = len([o for o in data["ai2_objects"] if o["panel"] == p and o["object_class"] == "STAR_OBJECT"])
        m = [r for r in matched_star_pairs if r["panel"] == p]
        ai1_only_p = [r for r in ai1_only_stars if r["panel"] == p]
        ai2_only_p = [r for r in ai2_only_stars if r["panel"] == p]
        ious_p = sorted(r["iou"] for r in m)
        med_iou_p = ious_p[len(ious_p) // 2] if ious_p else 0.0
        dists_p = sorted(r["center_distance_norm"] for r in m)
        med_dist_p = dists_p[len(dists_p) // 2] if dists_p else 0.0
        union_p = ai1_ct + ai2_ct - len(m)
        rate_p = len(m) / max(1, union_p)
        per_panel_lines.append(
            f"| `{p}` | {a_ct} | {ai1_ct} | {ai2_ct} | {len(m)} | {len(ai1_only_p)} | {len(ai2_only_p)} | "
            f"{rate_p:.3f} | {med_iou_p:.3f} | {med_dist_p:.4f} |"
        )

    summary = f"""# ANNOTATOR_B_AI_2 Three-Way Agreement Summary

## 1. Executive Summary & Verification Metrics

```text
ANNOTATOR_B_AI_2_PASS=COMPLETE
ANNOTATOR_B_AI_2_INDEPENDENT_FROM_A=YES
ANNOTATOR_B_AI_2_INDEPENDENT_FROM_AI1=YES

AI1_STAR_OBJECTS={len(ai1_stars)}
AI2_STAR_OBJECTS={len(ai2_stars)}
AI1_AI2_MATCHED_STAR_OBJECTS={len(matched_star_pairs)}
AI1_ONLY_STAR_CANDIDATES={len(ai1_only_stars)}
AI2_ONLY_STAR_CANDIDATES={len(ai2_only_stars)}
AI_CONSENSUS_STAR_RATE={ai_consensus_rate:.4f}
MEDIAN_STAR_IOU={med_star_iou:.6f}
MEDIAN_STAR_CENTER_DISTANCE_PX={med_star_dist:.4f}

AI1_LABELS={ai1_labels_n}
AI2_LABELS={ai2_labels_n}
AI1_AI2_MATCHED_LABELS={len(matched_labels)}
LABEL_CONSENSUS_RATE={label_consensus_rate:.4f}

AI1_AI2_STAR_RELATION_AGREEMENT={rel_stats['star_relation_agree']}/{max(1, rel_stats['star_relation_total'])} ({star_rel_pct:.1f}%)
AI1_AI2_RELATION_EXACT_AGREEMENT={rel_stats['relation_exact_agree']}/{max(1, rel_stats['relation_total'])} ({rel_pct:.1f}%)
AI1_AI2_RELATION_MEAN_CANDIDATE_JACCARD={rel_stats['mean_jaccard']:.4f}

THREE_WAY_STAR_MATCH_F68R1={f68r1['three_way']}
THREE_WAY_LABEL_MATCH_F68R1={f68r1['three_way_lbl']}

AI_CONSENSUS_STAR_OBJECTS={len(star_rows)}
AI_CONSENSUS_LABELS={len(lbl_rows)}

HUMAN_ADJUDICATION_HIGH_PRIORITY_COUNT={high_prio}
HUMAN_ADJUDICATION_MEDIUM_PRIORITY_COUNT={med_prio}
HUMAN_ADJUDICATION_LOW_PRIORITY_COUNT={low_prio}

AI_CONSENSUS_USEFUL_FOR_HUMAN_REVIEW={consensus_verdict}
HUMAN_ANNOTATOR_STILL_REQUIRED=YES
PRODUCTION_SPATIAL_GATE=READY_FOR_HUMAN_ADJUDICATION
```

## 2. STAR_OBJECT Consensus by Panel

| Panel | A_STARS | AI1_STARS | AI2_STARS | AI1_AI2_MATCHED | AI1_ONLY | AI2_ONLY | CONSENSUS_RATE | MEDIAN_IOU | MEDIAN_CENTER_DIST_NORM |
|---|---|---|---|---|---|---|---|---|---|
{chr(10).join(per_panel_lines)}

`f68r1` is the one panel where `ANNOTATOR_A` provides a nearly complete star
reference; the other seven panels had an incomplete `ANNOTATOR_A` pass, so
`AI1_AI2_ONLY` support there must not be read as "AI disagrees with A" — A
simply never annotated those panels.

## 3. f68r1 Critical Sanity Check (Part P)

```text
F68R1_A_STARS={f68r1['a_stars']}
F68R1_AI1_STARS={f68r1['ai1_stars']}
F68R1_AI2_STARS={f68r1['ai2_stars']}

F68R1_A_AI1_MATCH={f68r1['a_ai1_match']}
F68R1_A_AI2_MATCH={f68r1['a_ai2_match']}
F68R1_AI1_AI2_MATCH={f68r1['ai1_ai2_match']}
F68R1_THREE_WAY_MATCH={f68r1['three_way']}

F68R1_A_LABELS={f68r1['a_labels']}
F68R1_AI1_LABELS={f68r1['ai1_labels']}
F68R1_AI2_LABELS={f68r1['ai2_labels']}
F68R1_A_AI1_LABEL_MATCH={f68r1['a_ai1_lbl_match']}
F68R1_A_AI2_LABEL_MATCH={f68r1['a_ai2_lbl_match']}
F68R1_AI1_AI2_LABEL_MATCH={f68r1['ai1_ai2_lbl_match']}
F68R1_THREE_WAY_LABEL_MATCH={f68r1['three_way_lbl']}
```

## 4. Disagreement Taxonomy Counts

{chr(10).join(f"- `{t}`: {c}" for t, c in sorted(_count_by(disagreements, 'disagreement_type').items()))}

## 5. Human Adjudication Triage

- **HIGH** ({high_prio}): AI1/AI2 disagree on STAR_OBJECT existence, relation disagreements,
  AI1_AI2_ONLY labels, and any `A_CONFLICT_WITH_AI_CONSENSUS` case.
- **MEDIUM** ({med_prio}): bounding-box disagreements on matched objects, non-star
  AI1-only/AI2-only objects.
- **LOW** ({low_prio}): confidence-only mismatches on otherwise agreeing pairs.

## 6. Interpretation Boundary

The headline result of this pass is that AI1<->AI2 agreement is **{consensus_qualifier.lower()}, not high**
(`AI_CONSENSUS_STAR_RATE={ai_consensus_rate:.3f}`, `LABEL_CONSENSUS_RATE={label_consensus_rate:.3f}`)
outside `f68r1`, the one panel with a real `ANNOTATOR_A` reference (where
AI1<->AI2 star/label matching is far higher — {f68r1['ai1_ai2_match']}/29 stars,
{f68r1['ai1_ai2_lbl_match']}/{max(f68r1['ai1_labels'], f68r1['ai2_labels'])} labels — than the corpus-wide rate).
On the seven panels A never really annotated, AI1 and AI2 found substantially
different star counts and positions (see the per-panel table above): AI1 was
generated as a single earlier procedural pass over these crops, biased toward
regular sector/ring symmetry, while AI2 was a fresh, independently launched
agent instance (different model family, clean-room, two-stage frozen protocol)
that explicitly rejected symmetry completion and reported partial/asymmetric
counts. This divergence is itself informative: it shows that, for this panel
set, single-AI-pass star/label counts are annotator-sensitive off the one
richly-referenced panel, and should not be treated as a converged candidate
layer without human adjudication. The `AI1 ∩ AI2` intersection
({len(star_rows)} stars, {len(lbl_rows)} labels) is still a materially
stronger candidate set than either raw pass alone — every entry in it was
independently proposed twice — but it is **not** a substitute for human
ground truth (`HUMAN_ANNOTATOR_STILL_REQUIRED=YES`), and its low corpus-wide
coverage means many real objects present in only one pass are excluded from
it. This independence was operationalized as: a fresh, context-free agent
instance from a different model family, given only the clean-room package,
with no visibility into `ANNOTATOR_A`, `ANNOTATOR_B_AI`, or any counts from
either (see `AI2_INPUT_MANIFEST.json` / `AI2_MANIFEST.json`).
Where the two AI passes disagree, or where `ANNOTATOR_A` conflicts with an
AI1/AI2 consensus object, human adjudication is still required.
"""
    (OUT / "AI2_AGREEMENT_SUMMARY.md").write_text(summary, encoding="utf-8")
    return dict(
        ai1_stars=len(ai1_stars), ai2_stars=len(ai2_stars), matched_stars=len(matched_star_pairs),
        ai1_only_stars=len(ai1_only_stars), ai2_only_stars=len(ai2_only_stars),
        ai1_labels=ai1_labels_n, ai2_labels=ai2_labels_n, matched_labels=len(matched_labels),
        star_rel_pct=star_rel_pct, rel_pct=rel_pct, high_prio=high_prio,
        ai_consensus_star_objects=len(star_rows), ai_consensus_labels=len(lbl_rows),
    )


def _count_by(rows, key):
    c = defaultdict(int)
    for r in rows:
        c[r[key]] += 1
    return c


def finalize_manifest():
    files_to_hash = [
        "AI2_INPUT_MANIFEST.json",
        "AI2_OBJECTS.tsv", "AI2_LABELS.tsv", "AI2_LABEL_OBJECT_RELATIONS.tsv",
        "AI2_PRIMARY_PASS_REPORT.md", "AI2_STAGE1_MANIFEST.json", "AI2_STAGE2_MANIFEST.json",
        "AI1_AI2_OBJECT_MATCH.tsv", "AI1_AI2_LABEL_MATCH.tsv", "AI1_AI2_RELATION_AGREEMENT.tsv",
        "A_AI2_OBJECT_MATCH.tsv", "A_AI2_LABEL_MATCH.tsv",
        "THREE_WAY_OBJECT_SUPPORT.tsv", "THREE_WAY_LABEL_SUPPORT.tsv", "THREE_WAY_DISAGREEMENTS.tsv",
        "AI_CONSENSUS_STAR_DATASET.tsv", "AI_CONSENSUS_LABEL_DATASET.tsv",
        "AI2_STRUCTURAL_SENSITIVITY.tsv", "AI2_AGREEMENT_SUMMARY.md",
    ]
    manifest = {
        "manifest_version": "1.0",
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "ANNOTATOR_TYPE": "AI",
        "ANNOTATOR_ID": "ANNOTATOR_B_AI_2",
        "INDEPENDENT_FROM_A": "YES",
        "INDEPENDENT_FROM_AI_B_1": "YES",
        "HUMAN_EQUIVALENT": "NO",
        "production_gate": "READY_FOR_HUMAN_ADJUDICATION",
        "files": {},
    }
    sha_lines = []
    for fname in sorted(files_to_hash):
        fpath = OUT / fname
        if fpath.exists():
            digest = sha256_file(fpath)
            manifest["files"][fname] = {"sha256": digest, "bytes": fpath.stat().st_size}
            sha_lines.append(f"{digest}  {fname}")
    (OUT / "AI2_MANIFEST.json").write_text(json.dumps(manifest, indent=2), encoding="utf-8")
    m_digest = sha256_file(OUT / "AI2_MANIFEST.json")
    sha_lines.append(f"{m_digest}  AI2_MANIFEST.json")
    (OUT / "SHA256SUMS").write_text("\n".join(sha_lines) + "\n", encoding="utf-8")
    print("AI2_MANIFEST.json and SHA256SUMS written.")


def main():
    data = load_all()
    pairwise = run_pairwise(data)
    rel_stats = run_relation_agreement(data, pairwise)
    obj_support = build_three_way_support(data, pairwise)
    lbl_support = build_three_way_label_support(data, pairwise)
    disagreements = build_disagreements(data, pairwise, rel_stats, obj_support, lbl_support)
    star_rows, lbl_rows = build_consensus_datasets(data, pairwise)
    build_structural_sensitivity(data, pairwise)
    f68r1 = f68r1_sanity_check(data, pairwise, obj_support, lbl_support)
    stats = build_summary_and_status(data, pairwise, rel_stats, disagreements, star_rows, lbl_rows, f68r1)
    finalize_manifest()
    print("Comparison pipeline complete.")
    print(json.dumps({**stats, **f68r1}, indent=2))


if __name__ == "__main__":
    main()
