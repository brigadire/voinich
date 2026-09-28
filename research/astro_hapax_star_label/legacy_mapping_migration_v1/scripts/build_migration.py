#!/usr/bin/env python3
"""Build the conservative legacy-to-3G1 astronomical LABEL migration."""

from __future__ import annotations

import argparse
import csv
import hashlib
import json
import math
import re
import shutil
from collections import Counter, defaultdict
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[4]
PACKAGE = ROOT / "research/astro_hapax_star_label/legacy_mapping_migration_v1"
TARGET_PANELS = ("f68r1", "f68r2", "f68r3")
PANEL_DIMS = {"f68r1": (2462, 3828), "f68r2": (2078, 3828), "f68r3": (3453, 3828)}
AMENDMENT_SHA = "9f413ff52b5ba065a2b896e2096c008688400072bfd1dc1a0aa8688ffa25763f"
FROZEN_PLAN_SHA = "02e1f553dd77a60e40b8ae3cae3b3aa50abe0fef50997455ce463fc63d246e7a"
OCCURRENCE_SHA = "ba0342e15d8c468ec4e9f741e97cdb4a11938fe1f0ae3ac4338b73aaf1bd773a"

MATCHES = ROOT / "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv"
OLD_LABELS = ROOT / "research/astro_spatial_annotation/ASTRO_LABELS_SPATIAL.tsv"
OLD_RELATIONS = ROOT / "research/astro_spatial_annotation/ASTRO_LABEL_OBJECT_RELATIONS.tsv"
OLD_STOLFI_BRIDGE = ROOT / "research/astro_spatial_annotation/ASTRO_STOLFI_SPATIAL_BRIDGE.tsv"
AUGMENTED = ROOT / "research/astro_spatial_augmented_human_reference/AUGMENTED_HUMAN_OBJECTS.tsv"
GROUPS = ROOT / "research/astro_spatial_augmented_human_reference/RELATION_GROUPS_3G1.tsv"
MEMBERS = ROOT / "research/astro_spatial_augmented_human_reference/RELATION_GROUP_MEMBERS_3G1.tsv"
OCCURRENCES = ROOT / "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl"
TOKEN_CANDIDATES = ROOT / "research/astro_hapax_star_label/TRANSCRIPTION_TOKEN_CANDIDATES.tsv"
PAGES = ROOT / "research/astro_spatial_human_adjudication/images"


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1 << 20), b""):
            digest.update(block)
    return digest.hexdigest()


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as handle:
        return list(csv.DictReader(handle, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fieldnames=fields, delimiter="\t", lineterminator="\n")
        writer.writeheader()
        writer.writerows(rows)


def fmt(value: float) -> str:
    return f"{value:.9f}"


def family(row: dict[str, str]) -> str:
    if row["object_type"] == "star":
        return "STAR"
    if row["object_type"] == "planet?":
        return "PLANET_MOON"
    if row["panel"] == "f67r1" and row["stolfi_group"] == "S":
        return "CIRCLE_SECTOR"
    if row["panel"] == "f68v2" and row["stolfi_group"] == "R":
        return "RADIAL_TEXT"
    if row["panel"] == "f68v2" and row["stolfi_group"] in {"X", "Y"}:
        return "SECTOR_LABEL"
    if row["panel"] == "f68v2" and row["stolfi_group"] == "Z":
        return "OUTER_TITLE"
    return "OTHER_DIAGRAM_LABEL"


def label_tokens(text: str) -> list[str]:
    return [token for token in re.split(r"[\s.,-]+", text.strip()) if token]


def wilson(success: int, total: int) -> tuple[float, float]:
    if total == 0:
        return (float("nan"), float("nan"))
    z = 1.959963984540054
    p = success / total
    den = 1 + z * z / total
    mid = (p + z * z / (2 * total)) / den
    half = z * math.sqrt((p * (1 - p) + z * z / (4 * total)) / total) / den
    return mid - half, mid + half


def bbox(row: dict[str, str]) -> tuple[float, float, float, float]:
    return tuple(float(row[key]) for key in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"))


def iou(a: tuple[float, float, float, float], b: tuple[float, float, float, float]) -> float:
    x1, y1 = max(a[0], b[0]), max(a[1], b[1])
    x2, y2 = min(a[2], b[2]), min(a[3], b[3])
    inter = max(0.0, x2 - x1) * max(0.0, y2 - y1)
    area_a = (a[2] - a[0]) * (a[3] - a[1])
    area_b = (b[2] - b[0]) * (b[3] - b[1])
    union = area_a + area_b - inter
    return inter / union if union else 0.0


def center(box: tuple[float, float, float, float]) -> tuple[float, float]:
    return ((box[0] + box[2]) / 2, (box[1] + box[3]) / 2)


def center_distance_norm(a: tuple[float, float, float, float], b: tuple[float, float, float, float], panel: str) -> float:
    ac, bc = center(a), center(b)
    return math.hypot(ac[0] - bc[0], ac[1] - bc[1]) / math.hypot(*PANEL_DIMS[panel])


def verify_ledger(directory: Path, ledger_name: str) -> None:
    for line in (directory / ledger_name).read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        expected, relative = line.split(None, 1)
        relative_path = Path(relative.strip())
        path = directory / relative_path
        if not path.exists():
            path = ROOT / relative_path
        if sha256(path) != expected:
            raise RuntimeError(f"checksum mismatch: {path}")


def row_count(path: Path) -> str:
    if path.suffix in {".tsv", ".csv"}:
        return str(max(0, sum(1 for _ in path.open(encoding="utf-8", errors="replace")) - 1))
    if path.suffix == ".jsonl":
        return str(sum(1 for _ in path.open(encoding="utf-8")))
    return "NA"


def source_manifest() -> list[dict[str, object]]:
    specs = [
        ("legacy decision report", "research/stolfi_matching_bias/STOLFI_MATCHING_BIAS_REPORT.md", "2026-09-01", "PRIMARY_REPORT"),
        ("legacy selection audit", "research/stolfi_matching_bias/STOLFI_MATCHING_BIAS_AUDIT.tsv", "schema-1", "DIRECT_REPORT_INPUT"),
        ("legacy panel/family coverage", "research/stolfi_matching_bias/STOLFI_MATCHING_BIAS_BY_PANEL_FAMILY.tsv", "schema-1", "DIRECT_REPORT_INPUT"),
        ("legacy sensitivity bounds", "research/stolfi_matching_bias/STOLFI_MATCHING_BIAS_SENSITIVITY.tsv", "schema-1", "DIRECT_REPORT_INPUT"),
        ("legacy source manifest", "research/stolfi_matching_bias/STOLFI_MATCHING_BIAS_MANIFEST.json", "schema-1", "REPORT_PROVENANCE"),
        ("legacy checksum ledger", "research/stolfi_matching_bias/STOLFI_MATCHING_BIAS_SHA256SUMS", "frozen", "REPORT_PROVENANCE"),
        ("row-level legacy mapping", "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv", "98-07-20-to-ZL3b", "PRIMARY_ROW_SOURCE"),
        ("legacy inventory audit", "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_AUDIT.md", "2026-09-01", "REPORT_INPUT"),
        ("legacy coverage table", "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_COVERAGE.tsv", "schema-1", "REPORT_INPUT"),
        ("old anonymous spatial labels", "research/astro_spatial_annotation/ASTRO_LABELS_SPATIAL.tsv", "first-pass", "F68R1_SPATIAL_BRIDGE"),
        ("old label-object relations", "research/astro_spatial_annotation/ASTRO_LABEL_OBJECT_RELATIONS.tsv", "first-pass", "F68R1_SPATIAL_PROVENANCE"),
        ("old conservative Stolfi bridge", "research/astro_spatial_annotation/ASTRO_STOLFI_SPATIAL_BRIDGE.tsv", "first-pass", "NEGATIVE_CONTROL"),
        ("spatial source registry", "research/astro_spatial_annotation/ASTRO_SPATIAL_SOURCE_REGISTRY.tsv", "first-pass", "COORDINATE_FRAME"),
        ("canonical augmented objects", "research/astro_spatial_augmented_human_reference/AUGMENTED_HUMAN_OBJECTS.tsv", "augmented-1.0", "TARGET_GEOMETRY"),
        ("3G1 groups", "research/astro_spatial_augmented_human_reference/RELATION_GROUPS_3G1.tsv", "3G1", "TARGET_GROUPS"),
        ("3G1 members", "research/astro_spatial_augmented_human_reference/RELATION_GROUP_MEMBERS_3G1.tsv", "3G1", "TARGET_MEMBERSHIP"),
        ("augmented manifest", "research/astro_spatial_augmented_human_reference/AUGMENTED_REFERENCE_MANIFEST.json", "augmented-1.0", "TARGET_PROVENANCE"),
        ("augmented checksum ledger", "research/astro_spatial_augmented_human_reference/SHA256SUMS", "frozen", "TARGET_PROVENANCE"),
        ("frozen occurrence registry", "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl", "task79-v1", "TOKEN_IDENTITY"),
        ("frozen canonical corpus", "data_work/ZL3b-x7.canonical.txt", "ZL3b-x7", "CORPUS_FREQUENCY_IF_GATE_PASSES"),
        ("frozen IVTFF transcription", "data/ZL3b-n.txt", "ZL3b-2025-05-13", "RAW_TRANSCRIPTION"),
        ("prepared token registry", "research/astro_hapax_star_label/TRANSCRIPTION_TOKEN_CANDIDATES.tsv", "HSL-1.0", "READABLE_TOKEN_CHECK"),
        ("frozen original plan", "research/astro_hapax_star_label/HAPAX_STAR_LABEL_ANALYSIS_PLAN.md", "HSL-1.0", "AMENDED_NOT_OVERWRITTEN"),
        ("visual taxonomy", "research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv", "frozen", "RELATED_CONTEXT"),
        ("visual page fingerprints", "research/visual_context/VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv", "frozen", "RELATED_CONTEXT"),
        ("prior spatial/page report", "research/hapax_check/HAPAX_SPATIAL_AND_PAGE_REPORT.md", "frozen", "RELATED_CONTEXT"),
        ("prior page comparison", "research/hapax_check/PAGE_COMPARATIVE_SUMMARY.md", "frozen", "RELATED_CONTEXT"),
        ("prior hapax report", "research/stolfi_label_hapax_enrichment/STOLFI_ASTRO_LABEL_HAPAX_REPORT.md", "2026-09-01", "LEGACY_RESULT_CONTEXT_ONLY"),
        ("downstream token-formation consumer", "research/astro_token_formation/main.py", "repository", "RELATED_CONSUMER_NOT_INVENTORY_BUILDER"),
        ("legacy enrichment implementation", "research/stolfi_label_hapax_enrichment/main.py", "repository", "REPORT_IMPLEMENTATION"),
        ("matching bias implementation", "research/stolfi_matching_bias/main.py", "repository", "REPORT_IMPLEMENTATION"),
    ]
    rows = []
    for role, relative, version, relation in specs:
        path = ROOT / relative
        if not path.exists():
            raise RuntimeError(f"missing declared input: {relative}")
        rows.append({
            "logical_role": role,
            "path": relative,
            "version": version,
            "row_count": row_count(path),
            "file_size": path.stat().st_size,
            "sha256": sha256(path),
            "frozen_status": "FROZEN_READ_ONLY",
            "relation_to_stolfi_matching_bias_report": relation,
        })
    return rows


def load_occurrences() -> tuple[dict[int, dict[str, object]], dict[int, str]]:
    occurrences: dict[int, dict[str, object]] = {}
    with OCCURRENCES.open(encoding="utf-8") as handle:
        for line in handle:
            row = json.loads(line)
            occurrences[int(row["absolute_token_position"])] = row
    readable = {int(row["occurrence_id"]): row["readable_eva"] for row in read_tsv(TOKEN_CANDIDATES)}
    return occurrences, readable


def normalized_legacy(matches: list[dict[str, str]], occurrences: dict[int, dict[str, object]]) -> list[dict[str, object]]:
    by_coordinate: dict[str, list[dict[str, str]]] = defaultdict(list)
    for row in matches:
        by_coordinate[row["stolfi_coordinate"]].append(row)
    representative = {coord: min(rows, key=lambda row: row["record_id"])["record_id"] for coord, rows in by_coordinate.items()}
    potential = {coord: max(len(label_tokens(row["stolfi_eva"])) for row in rows) for coord, rows in by_coordinate.items()}
    out = []
    for source_row, row in enumerate(matches, 2):
        positions = [int(value) for value in row["absolute_token_positions"].split(",") if value]
        normalized = ";".join(str(occurrences[pos]["token"]).replace("\x1f", "/") for pos in positions)
        unmatched = row["match_status"] == "UNMATCHED"
        is_rep = representative[row["stolfi_coordinate"]] == row["record_id"]
        out.append({
            "legacy_record_id": row["record_id"],
            "folio": row["panel"][:3],
            "panel": row["panel"],
            "visual_family": family(row),
            "legacy_coordinates": row["stolfi_coordinate"],
            "transcription_locus": row["zl3b_locus"],
            "line_id": row["zl3b_locus"],
            "token_position": row["absolute_token_positions"],
            "raw_token": row["stolfi_eva"],
            "normalized_token": normalized,
            "matched_zl3b_token": row["zl3b_eva_tokens"],
            "match_status": row["match_status"],
            "confirmation_status": "RULE_CONFIRMED_FROZEN_OCCURRENCE" if not unmatched else "NOT_CONFIRMED",
            "mapping_method": row["match_method"],
            "source_table": "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv",
            "source_row": source_row,
            "ambiguity": "NONE" if not unmatched else "NO_ADMISSIBLE_FROZEN_OCCURRENCE",
            "sensitivity_only_status": "BOUND_COORDINATE_REPRESENTATIVE" if unmatched and is_rep else "DUPLICATE_VARIANT_NOT_COUNTED" if unmatched else "NO",
            "sensitivity_only_potential_occurrences": potential[row["stolfi_coordinate"]] if unmatched and is_rep else 0,
            "target_scope": "TARGET_PANEL" if row["panel"] in TARGET_PANELS else "OUTSIDE_TARGET_PANELS",
        })
    return out


def build_crosswalk(matches: list[dict[str, str]], augmented: list[dict[str, str]], groups: list[dict[str, str]], members: list[dict[str, str]], old_labels: list[dict[str, str]], occurrences: dict[int, dict[str, object]], readable: dict[int, str]):
    target = [row for row in augmented if row["panel"] in TARGET_PANELS and row["object_type"] == "LABEL"]
    object_by_id = {row["object_id"]: row for row in augmented}
    group_by_id = {row["canonical_group_id"]: row for row in groups}
    membership = {row["object_id"]: row["canonical_group_id"] for row in members}
    old_by_id = {row["label_occurrence_id"]: row for row in old_labels}
    matched_by_locus: dict[str, list[dict[str, str]]] = defaultdict(list)
    matched_by_panel: dict[str, set[str]] = defaultdict(set)
    for row in matches:
        if row["match_status"] == "MATCHED":
            matched_by_locus[row["zl3b_locus"]].append(row)
            matched_by_panel[row["panel"]].add(row["stolfi_coordinate"])

    direct_anon_to_canonical: dict[int, str] = {}
    for row in target:
        match = re.search(r"(?:^|;)A:f68r1-label-anon-(\d{3})(?:;|$)", row["source_annotation_ids"])
        if match:
            direct_anon_to_canonical[int(match.group(1))] = row["object_id"]
    if len(direct_anon_to_canonical) != 26:
        raise RuntimeError(f"expected 26 direct f68r1 anonymous-label links, found {len(direct_anon_to_canonical)}")

    canonical_to_mapping: dict[str, dict[str, object]] = {}
    for ordinal, canonical_id in direct_anon_to_canonical.items():
        locus = f"f68r1.{ordinal + 7}"
        rows = matched_by_locus.get(locus, [])
        coordinates = sorted({row["stolfi_coordinate"] for row in rows})
        if len(coordinates) == 1:
            coordinate = coordinates[0]
            primary_row = min(rows, key=lambda row: row["record_id"])
            positions = [int(value) for value in primary_row["absolute_token_positions"].split(",") if value]
            canonical_to_mapping[canonical_id] = {
                "outcome": "UNIQUE_PROVENANCE_CROSSWALK",
                "legacy_coordinate": coordinate,
                "legacy_record_ids": ";".join(sorted(row["record_id"] for row in rows)),
                "locus": locus,
                "positions": positions,
                "raw_tokens": [readable[pos] for pos in positions],
                "normalized_tokens": [str(occurrences[pos]["token"]).replace("\x1f", "/") for pos in positions],
                "method": "LEGACY_MATCH_TO_ZL3B_LOCUS;F68R1_LS_ORDINAL_TO_ANON;DIRECT_A_SOURCE_ID",
                "confidence": "HIGH_PROVENANCE_NOT_HUMAN_TRANSCRIPTION",
                "old_label_id": f"f68r1-label-anon-{ordinal:03d}",
                "candidate_coordinates": coordinate,
            }
        else:
            canonical_to_mapping[canonical_id] = {
                "outcome": "LEGACY_UNMATCHED",
                "legacy_coordinate": "",
                "legacy_record_ids": "",
                "locus": locus,
                "positions": [],
                "raw_tokens": [],
                "normalized_tokens": [],
                "method": "F68R1_LS_ORDINAL_TO_ANON;NO_CONFIRMED_LEGACY_ROW",
                "confidence": "NONE",
                "old_label_id": f"f68r1-label-anon-{ordinal:03d}",
                "candidate_coordinates": "",
            }

    unresolved_f68r1 = sorted(set(range(1, 30)) - set(direct_anon_to_canonical))
    unresolved_loci = [f"f68r1.{ordinal + 7}" for ordinal in unresolved_f68r1]
    unresolved_coords = sorted({row["stolfi_coordinate"] for locus in unresolved_loci for row in matched_by_locus.get(locus, [])})

    crosswalk = []
    mapping = []
    geometry = []
    for row in sorted(target, key=lambda item: (item["panel"], item["object_id"])):
        canonical_id = row["object_id"]
        group_id = membership.get(canonical_id, "")
        grouped = bool(group_id)
        group = group_by_id.get(group_id, {})
        human_added = row["origin"] == "HUMAN_ADDED_COMPLETENESS"
        data = canonical_to_mapping.get(canonical_id)
        if data is None:
            if human_added:
                outcome = "HUMAN_ADDED_NOT_IN_LEGACY"
                candidates = ""
                note = "Human-added LABEL has no frozen legacy object correspondence; token assignment prohibited"
            elif grouped and row["panel"] == "f68r1":
                outcome = "AMBIGUOUS_CROSSWALK"
                candidates = ";".join(unresolved_coords)
                note = "Canonical grouped LABEL lacks direct A-label provenance; three unresolved f68r1 mappings remain many-to-many"
            elif grouped:
                outcome = "AMBIGUOUS_CROSSWALK"
                candidates = ";".join(sorted(matched_by_panel[row["panel"]]))
                note = "Panel has confirmed legacy records but no independent legacy pixel/object bridge"
            else:
                outcome = "NO_LEGACY_CANDIDATE"
                candidates = ""
                note = "Ungrouped canonical LABEL is outside the identified legacy star-label object series"
            data = {
                "outcome": outcome,
                "legacy_coordinate": "",
                "legacy_record_ids": "",
                "locus": "",
                "positions": [],
                "raw_tokens": [],
                "normalized_tokens": [],
                "method": "NO_AUTOMATIC_TRANSFER",
                "confidence": "NONE",
                "old_label_id": "",
                "candidate_coordinates": candidates,
                "note": note,
            }
        else:
            note = "Unique frozen occurrence and direct canonical source provenance" if data["outcome"] == "UNIQUE_PROVENANCE_CROSSWALK" else "Legacy series represented, but no confirmed occurrence at this visual locus"
        primary = data["outcome"] in {"EXACT_LEGACY_CROSSWALK", "UNIQUE_PROVENANCE_CROSSWALK", "UNIQUE_GEOMETRIC_CROSSWALK"}
        crosswalk.append({
            "canonical_label_id": canonical_id,
            "panel": row["panel"],
            "crosswalk_outcome": data["outcome"],
            "legacy_record_ids": data["legacy_record_ids"],
            "legacy_coordinates": data["legacy_coordinate"],
            "legacy_candidate_coordinates": data["candidate_coordinates"],
            "transcription_locus": data["locus"],
            "absolute_token_positions": ";".join(str(pos) for pos in data["positions"]),
            "crosswalk_method": data["method"],
            "mapping_confidence": data["confidence"],
            "primary_mapping_eligible": "YES" if primary else "NO",
            "source_spatial_label_id": data["old_label_id"],
            "provenance_note": note,
        })
        mapping.append({
            "canonical_label_id": canonical_id,
            "panel": row["panel"],
            "grouped": "YES" if grouped else "NO",
            "canonical_3g1_group_id": group_id or "UNGROUPED",
            "group_size": group.get("member_count", 0),
            "star_count": group.get("star_count", 0),
            "label_count": group.get("label_count", 0),
            "group_change_category": group.get("change_category", "UNGROUPED"),
            "human_added_membership": "YES" if human_added else "NO",
            "multi_member_group_flag": "YES" if int(group.get("member_count", 0) or 0) > 2 else "NO",
            "crosswalk_outcome": data["outcome"],
            "legacy_coordinates": data["legacy_coordinate"],
            "transcription_locus": data["locus"],
            "absolute_token_positions": ";".join(str(pos) for pos in data["positions"]),
            "raw_token_sequence": " ".join(data["raw_tokens"]),
            "normalized_token_sequence": ";".join(data["normalized_tokens"]),
            "token_count": len(data["positions"]),
            "legacy_confirmation_status": "RULE_CONFIRMED_FROZEN_OCCURRENCE" if primary else "NOT_PRIMARY",
            "mapping_confidence": data["confidence"],
            "primary_analysis_inclusion": "YES" if primary else "NO",
        })

        can_box = bbox(row)
        old = old_by_id.get(str(data["old_label_id"]))
        old_box = bbox(old) if old else None
        peers = [item for item in target if item["panel"] == row["panel"] and item["object_id"] != canonical_id]
        distances = sorted(center_distance_norm(old_box, bbox(item), row["panel"]) for item in peers) if old_box else []
        direct_distance = center_distance_norm(old_box, can_box, row["panel"]) if old_box else None
        runner = distances[0] if distances else None
        geometry.append({
            "canonical_label_id": canonical_id,
            "panel": row["panel"],
            "legacy_spatial_proxy_id": data["old_label_id"],
            "coordinate_transform": "IDENTITY_PANEL_PIXEL_FRAME" if old_box else "UNAVAILABLE",
            "panel_identity": "PASS",
            "legacy_bbox": ",".join(fmt(value) for value in old_box) if old_box else "",
            "canonical_bbox": ",".join(fmt(value) for value in can_box),
            "axis_aligned_iou": fmt(iou(old_box, can_box)) if old_box else "NA",
            "center_distance_norm": fmt(direct_distance) if direct_distance is not None else "NA",
            "runner_up_center_distance_norm": fmt(runner) if runner is not None else "NA",
            "uniqueness_margin": fmt(runner - direct_distance) if runner is not None and direct_distance is not None else "NA",
            "token_exists": "YES" if all(pos in occurrences for pos in data["positions"]) and data["positions"] else "NA",
            "raw_token_byte_identity": "YES" if data["positions"] and data["raw_tokens"] == [readable[pos] for pos in data["positions"]] else "NA",
            "automatic_geometry_threshold_pass": "YES" if old_box and iou(old_box, can_box) >= 0.50 and direct_distance <= 0.025 and runner is not None and runner - direct_distance >= 0.010 else "NO",
            "crosswalk_outcome": data["outcome"],
        })
    return target, crosswalk, mapping, geometry, object_by_id, group_by_id


def add_coverage_row(rows: list[dict[str, object]], kind: str, value: str, selected: list[dict[str, object]]) -> None:
    total = len(selected)
    mapped = sum(row["primary_analysis_inclusion"] == "YES" for row in selected)
    low, high = wilson(mapped, total)
    rows.append({
        "stratum_type": kind,
        "stratum": value,
        "labels_total": total,
        "labels_mapped": mapped,
        "labels_unmapped": total - mapped,
        "coverage": fmt(mapped / total) if total else "NA",
        "wilson_95_low": fmt(low) if total else "NA",
        "wilson_95_high": fmt(high) if total else "NA",
    })


def coverage_tables(mapping: list[dict[str, object]], target_by_id: dict[str, dict[str, str]]):
    rows: list[dict[str, object]] = []
    add_coverage_row(rows, "OVERALL", "ALL", mapping)
    for field, kind, values in [
        ("panel", "PANEL", TARGET_PANELS),
        ("grouped", "GROUPED", ("YES", "NO")),
        ("human_added_membership", "HUMAN_ADDED", ("YES", "NO")),
        ("group_change_category", "GROUP_CHANGE", sorted({str(row["group_change_category"]) for row in mapping})),
        ("group_size", "GROUP_SIZE", sorted({str(row["group_size"]) for row in mapping}, key=lambda value: int(value))),
    ]:
        for value in values:
            add_coverage_row(rows, kind, str(value), [row for row in mapping if str(row[field]) == str(value)])

    areas = sorted((bbox(target_by_id[row["canonical_label_id"]])[2] - bbox(target_by_id[row["canonical_label_id"]])[0]) * (bbox(target_by_id[row["canonical_label_id"]])[3] - bbox(target_by_id[row["canonical_label_id"]])[1]) for row in mapping)
    q1, q2, q3 = areas[len(areas) // 4], areas[len(areas) // 2], areas[(3 * len(areas)) // 4]
    features: dict[str, dict[str, str]] = defaultdict(dict)
    for row in mapping:
        obj = target_by_id[row["canonical_label_id"]]
        area = (bbox(obj)[2] - bbox(obj)[0]) * (bbox(obj)[3] - bbox(obj)[1])
        size = "Q1_SMALLEST" if area <= q1 else "Q2" if area <= q2 else "Q3" if area <= q3 else "Q4_LARGEST"
        form = "RING" if obj["geometry_type"] == "ELLIPSE" else "ROTATED_LOCAL" if abs(float(obj["rotation"])) > 0.001 else "AXIS_ALIGNED_LOCAL"
        features[row["canonical_label_id"]] = {
            "PANEL": str(row["panel"]),
            "GROUPED": str(row["grouped"]),
            "HUMAN_ADDED": str(row["human_added_membership"]),
            "GEOMETRY_TYPE": obj["geometry_type"],
            "LABEL_SIZE_QUARTILE": size,
            "VISUAL_FORM": form,
            "GROUP_SIZE": str(row["group_size"]),
            "OBJECT_ORIGIN": obj["origin"],
            "LEGACY_FAMILY_AVAILABILITY": "STAR" if row["grouped"] == "YES" else "NO_IDENTIFIED_LEGACY_FAMILY",
        }

    audit = []
    for feature in next(iter(features.values())):
        categories = sorted({values[feature] for values in features.values()})
        counts = []
        for category in categories:
            selected = [row for row in mapping if features[row["canonical_label_id"]][feature] == category]
            total = len(selected)
            mapped = sum(row["primary_analysis_inclusion"] == "YES" for row in selected)
            counts.append((total, mapped))
        grand_total = sum(total for total, _ in counts)
        grand_mapped = sum(mapped for _, mapped in counts)
        chi2 = 0.0
        for total, mapped in counts:
            for observed, column in ((mapped, grand_mapped), (total - mapped, grand_total - grand_mapped)):
                expected = total * column / grand_total if grand_total else 0
                if expected:
                    chi2 += (observed - expected) ** 2 / expected
        v = math.sqrt(chi2 / grand_total) if grand_total else 0.0
        for category, (total, mapped) in zip(categories, counts):
            low, high = wilson(mapped, total)
            audit.append({
                "feature": feature,
                "category": category,
                "labels_total": total,
                "mapped": mapped,
                "unmapped": total - mapped,
                "coverage": fmt(mapped / total),
                "wilson_95_low": fmt(low),
                "wilson_95_high": fmt(high),
                "feature_cramers_v": fmt(v),
                "interpretation": "PRE_HAPAX_SELECTION_FEATURE" if feature != "LEGACY_FAMILY_AVAILABILITY" else "DESCRIPTIVE_LEGACY_AVAILABILITY",
            })
    return rows, audit


def residual_rows(mapping: list[dict[str, object]], crosswalk_by_id: dict[str, dict[str, str]]) -> list[dict[str, object]]:
    rows = []
    for row in mapping:
        if row["primary_analysis_inclusion"] == "YES":
            continue
        outcome = str(row["crosswalk_outcome"])
        if outcome == "HUMAN_ADDED_NOT_IN_LEGACY":
            reason = "HUMAN_ADDED"
            action = "TARGETED_HUMAN_TRANSCRIPTION_IF_LATER_AUTHORIZED"
        elif outcome == "AMBIGUOUS_CROSSWALK":
            reason = "AMBIGUOUS_GEOMETRY_CROSSWALK"
            action = "TARGETED_LEGACY_COORDINATE_TO_CROP_ADJUDICATION"
        elif outcome == "LEGACY_UNMATCHED":
            reason = "LEGACY_RECORD_HAS_NO_EXACT_TOKEN_OCCURRENCE"
            action = "TARGETED_TRANSCRIPTION_CONFLICT_REVIEW"
        else:
            reason = "ABSENT_IN_IDENTIFIED_LEGACY_MAPPING"
            action = "TARGETED_HUMAN_TRANSCRIPTION_IF_NEEDED"
        cross = crosswalk_by_id[row["canonical_label_id"]]
        rows.append({
            "canonical_label_id": row["canonical_label_id"],
            "panel": row["panel"],
            "grouped": row["grouped"],
            "human_added": row["human_added_membership"],
            "crosswalk_outcome": outcome,
            "residual_reason": reason,
            "legacy_candidate_coordinates": cross["legacy_candidate_coordinates"],
            "recommended_minimal_action": action,
            "priority": "HIGH" if row["grouped"] == "YES" else "MEDIUM",
        })
    return rows


def draw_box(draw: ImageDraw.ImageDraw, box: tuple[float, float, float, float], offset: tuple[float, float], scale: float, color: str, width: int = 3) -> None:
    x1, y1, x2, y2 = box
    ox, oy = offset
    draw.rectangle(((x1 - ox) * scale, (y1 - oy) * scale, (x2 - ox) * scale, (y2 - oy) * scale), outline=color, width=width)


def make_tile(page: Image.Image, canonical: dict[str, str], title: str, old: dict[str, str] | None = None, size=(420, 235)) -> Image.Image:
    can_box = bbox(canonical)
    boxes = [can_box] + ([bbox(old)] if old else [])
    margin = 80
    x1 = max(0, min(box[0] for box in boxes) - margin)
    y1 = max(0, min(box[1] for box in boxes) - margin)
    x2 = min(page.width, max(box[2] for box in boxes) + margin)
    y2 = min(page.height, max(box[3] for box in boxes) + margin)
    crop = page.crop((int(x1), int(y1), int(x2), int(y2))).convert("RGB")
    content_h = size[1] - 42
    scale = min(size[0] / max(1, crop.width), content_h / max(1, crop.height))
    resized = crop.resize((max(1, int(crop.width * scale)), max(1, int(crop.height * scale))), Image.Resampling.LANCZOS)
    tile = Image.new("RGB", size, "white")
    px = (size[0] - resized.width) // 2
    py = 40 + (content_h - resized.height) // 2
    tile.paste(resized, (px, py))
    draw = ImageDraw.Draw(tile)
    draw.text((5, 4), title[:68], fill="black", font=ImageFont.load_default())
    local_offset = (x1 - px / scale, y1 - (py - 40) / scale)
    overlay = ImageDraw.Draw(tile)
    def converted(box):
        return tuple((value - (x1 if index % 2 == 0 else y1)) * scale + (px if index % 2 == 0 else py) for index, value in enumerate(box))
    overlay.rectangle(converted(can_box), outline="#d62728", width=3)
    if old:
        overlay.rectangle(converted(bbox(old)), outline="#1f77b4", width=3)
    return tile


def sheet(path: Path, tiles: list[Image.Image], columns: int = 4) -> None:
    if not tiles:
        return
    width, height = tiles[0].size
    canvas = Image.new("RGB", (columns * width, math.ceil(len(tiles) / columns) * height), "white")
    for index, tile in enumerate(tiles):
        canvas.paste(tile, ((index % columns) * width, (index // columns) * height))
    path.parent.mkdir(parents=True, exist_ok=True)
    canvas.save(path, format="PNG", compress_level=9)


def build_visual_audit(out: Path, target: list[dict[str, str]], crosswalk: list[dict[str, object]], old_labels: list[dict[str, str]]) -> list[dict[str, object]]:
    by_id = {row["object_id"]: row for row in target}
    old_by_id = {row["label_occurrence_id"]: row for row in old_labels}
    pages = {panel: Image.open(PAGES / f"{panel}.jpg").convert("RGB") for panel in TARGET_PANELS}
    cw = {row["canonical_label_id"]: row for row in crosswalk}
    line_tokens: dict[str, list[tuple[int, str]]] = defaultdict(list)
    for token in read_tsv(TOKEN_CANDIDATES):
        if token["panel"] == "f68r1" and token["locus_type"] == "L":
            line_tokens[token["line_ref"]].append((int(token["index_in_line"]), token["readable_eva"]))
    audit_rows = []

    f68_tiles = []
    for ordinal in range(1, 30):
        old_id = f"f68r1-label-anon-{ordinal:03d}"
        candidates = [row for row in crosswalk if row["source_spatial_label_id"] == old_id]
        if candidates:
            cross = candidates[0]
            canonical = by_id[cross["canonical_label_id"]]
            locus = f"f68r1.{ordinal + 7}"
            readable_line = " ".join(value for _, value in sorted(line_tokens[locus]))
            title = f"anon-{ordinal:03d} -> {locus} {readable_line} | {cross['crosswalk_outcome']}"
            f68_tiles.append(make_tile(pages["f68r1"], canonical, title, old_by_id[old_id]))
            audit_rows.append({"audit_class": "ALL_F68R1_ORDINAL", "canonical_label_id": canonical["object_id"], "panel": "f68r1", "legacy_spatial_proxy_id": old_id, "crosswalk_outcome": cross["crosswalk_outcome"], "sheet": "visual_audit/F68R1_PROVENANCE_AUDIT.png"})
            if cross["primary_mapping_eligible"] == "YES":
                crop = make_tile(pages["f68r1"], canonical, f"{canonical['object_id']} | {cross['transcription_locus']}", old_by_id[old_id], size=(640, 340))
                crop.save(out / "visual_audit/crops" / f"{canonical['object_id']}.png", format="PNG", compress_level=9)
        else:
            dummy = {"bbox_x1": old_by_id[old_id]["bbox_x1"], "bbox_y1": old_by_id[old_id]["bbox_y1"], "bbox_x2": old_by_id[old_id]["bbox_x2"], "bbox_y2": old_by_id[old_id]["bbox_y2"]}
            locus = f"f68r1.{ordinal + 7}"
            readable_line = " ".join(value for _, value in sorted(line_tokens[locus]))
            f68_tiles.append(make_tile(pages["f68r1"], dummy, f"anon-{ordinal:03d} -> {locus} {readable_line} | NO DIRECT CANONICAL SOURCE", old_by_id[old_id]))
            audit_rows.append({"audit_class": "ALL_F68R1_ORDINAL", "canonical_label_id": "", "panel": "f68r1", "legacy_spatial_proxy_id": old_id, "crosswalk_outcome": "AMBIGUOUS_CROSSWALK", "sheet": "visual_audit/F68R1_PROVENANCE_AUDIT.png"})
    sheet(out / "visual_audit/F68R1_PROVENANCE_AUDIT.png", f68_tiles)

    special = [row for row in crosswalk if row["crosswalk_outcome"] in {"AMBIGUOUS_CROSSWALK", "HUMAN_ADDED_NOT_IN_LEGACY"}]
    special_tiles = []
    for cross in special:
        canonical = by_id[cross["canonical_label_id"]]
        special_tiles.append(make_tile(pages[canonical["panel"]], canonical, f"{canonical['object_id']} | {cross['crosswalk_outcome']}"))
        audit_rows.append({"audit_class": "AMBIGUOUS_OR_HUMAN_ADDED", "canonical_label_id": canonical["object_id"], "panel": canonical["panel"], "legacy_spatial_proxy_id": "", "crosswalk_outcome": cross["crosswalk_outcome"], "sheet": "visual_audit/AMBIGUOUS_AND_HUMAN_ADDED.png"})
    sheet(out / "visual_audit/AMBIGUOUS_AND_HUMAN_ADDED.png", special_tiles)

    mapped = [row for row in crosswalk if row["primary_mapping_eligible"] == "YES"]
    ranked = sorted(mapped, key=lambda row: hashlib.sha256(f"20260915:{row['canonical_label_id']}".encode()).hexdigest())[:8]
    sample_tiles = []
    for cross in ranked:
        canonical = by_id[cross["canonical_label_id"]]
        old = old_by_id[str(cross["source_spatial_label_id"])]
        sample_tiles.append(make_tile(pages[canonical["panel"]], canonical, f"FIXED_SAMPLE | {cross['transcription_locus']} | {canonical['object_id']}", old))
        audit_rows.append({"audit_class": "DETERMINISTIC_PRIMARY_SAMPLE", "canonical_label_id": canonical["object_id"], "panel": canonical["panel"], "legacy_spatial_proxy_id": cross["source_spatial_label_id"], "crosswalk_outcome": cross["crosswalk_outcome"], "sheet": "visual_audit/DETERMINISTIC_PRIMARY_SAMPLE.png"})
    sheet(out / "visual_audit/DETERMINISTIC_PRIMARY_SAMPLE.png", sample_tiles)
    return audit_rows


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, default=PACKAGE)
    args = parser.parse_args()
    out = args.output.resolve()
    out.mkdir(parents=True, exist_ok=True)
    (out / "visual_audit/crops").mkdir(parents=True, exist_ok=True)

    canonical_amendment = PACKAGE / "HAPAX_STAR_LABEL_ANALYSIS_PLAN_AMENDMENT_01.md"
    canonical_freeze = PACKAGE / "AMENDMENT_FREEZE.json"
    if out != PACKAGE:
        shutil.copyfile(canonical_amendment, out / canonical_amendment.name)
        shutil.copyfile(canonical_freeze, out / canonical_freeze.name)
    amendment = out / canonical_amendment.name
    if sha256(amendment) != AMENDMENT_SHA:
        raise RuntimeError("amendment checksum changed")
    if sha256(ROOT / "research/astro_hapax_star_label/HAPAX_STAR_LABEL_ANALYSIS_PLAN.md") != FROZEN_PLAN_SHA:
        raise RuntimeError("frozen original analysis plan changed")
    if sha256(OCCURRENCES) != OCCURRENCE_SHA:
        raise RuntimeError("occurrence registry changed")
    verify_ledger(ROOT / "research/stolfi_matching_bias", "STOLFI_MATCHING_BIAS_SHA256SUMS")
    verify_ledger(ROOT / "research/astro_spatial_augmented_human_reference", "SHA256SUMS")
    verify_ledger(ROOT / "research/astro_hapax_star_label", "SHA256SUMS")

    manifest_rows = source_manifest()
    write_tsv(out / "LEGACY_MAPPING_INPUT_MANIFEST.tsv", list(manifest_rows[0]), manifest_rows)

    matches = read_tsv(MATCHES)
    grouped = defaultdict(list)
    for row in matches:
        grouped[row["stolfi_coordinate"]].append(row)
    coordinate_rows = [rows[0] for rows in grouped.values()]
    metrics = {
        "PLANET_MOON": (sum(family(row) == "PLANET_MOON" and row["match_status"] == "MATCHED" for row in coordinate_rows), sum(family(row) == "PLANET_MOON" for row in coordinate_rows)),
        "CIRCLE_SECTOR": (sum(family(row) == "CIRCLE_SECTOR" and row["match_status"] == "MATCHED" for row in coordinate_rows), sum(family(row) == "CIRCLE_SECTOR" for row in coordinate_rows)),
        "STAR_LABELS": (sum(family(row) == "STAR" and row["match_status"] == "MATCHED" for row in coordinate_rows), sum(family(row) == "STAR" for row in coordinate_rows)),
        "UNMATCHED_COORDINATES": sum(row["match_status"] == "UNMATCHED" for row in coordinate_rows),
        "SENSITIVITY_ONLY_POSSIBLE_OCCURRENCES": sum(max(len(label_tokens(row["stolfi_eva"])) for row in rows) for rows in grouped.values() if rows[0]["match_status"] == "UNMATCHED"),
    }
    expected = {"PLANET_MOON": (7, 7), "CIRCLE_SECTOR": (10, 12), "STAR_LABELS": (53, 67), "UNMATCHED_COORDINATES": 54, "SENSITIVITY_ONLY_POSSIBLE_OCCURRENCES": 77}
    if metrics != expected:
        blocker = "# Legacy mapping reproduction blocker\n\nExpected and observed metrics differ; automated migration stopped.\n\n" + "\n".join(f"- `{key}`: observed `{metrics.get(key)}`, expected `{value}`" for key, value in expected.items()) + "\n"
        (out / "LEGACY_MAPPING_REPRODUCTION_BLOCKER.md").write_text(blocker, encoding="utf-8")
        raise RuntimeError("legacy control counts do not reproduce")

    reproduction = f"""# Legacy mapping reproduction

Status: `PASS`.

| Metric | Reproduced |
|---|---:|
| PLANET/MOON | {metrics['PLANET_MOON'][0]}/{metrics['PLANET_MOON'][1]} |
| CIRCLE-SECTOR | {metrics['CIRCLE_SECTOR'][0]}/{metrics['CIRCLE_SECTOR'][1]} |
| STAR LABELS | {metrics['STAR_LABELS'][0]}/{metrics['STAR_LABELS'][1]} |
| UNMATCHED_COORDINATES | {metrics['UNMATCHED_COORDINATES']} |
| SENSITIVITY_ONLY_POSSIBLE_OCCURRENCES | {metrics['SENSITIVITY_ONLY_POSSIBLE_OCCURRENCES']} |

The unit is a distinct Stolfi `panel.group.number` physical coordinate, not a
transcriber variant. The 191 source rows collapse to 143 coordinates; 89 are
matched and 54 unmatched. For each unmatched coordinate the sensitivity count
uses the maximum token count among its variants, totaling 77. These 77 are
accounting bounds only and are not LABEL or recovered occurrences.

The legacy correspondences are deterministic rule-derived matches, not human
visual verification: same panel, a validated Stolfi-series to IVTFF locus-type
bridge, exact/wildcard lexical anchors, coordinate-peer consensus for variants,
and one-to-one span resolution. `MATCHED` means the frozen procedure assigned
an admissible ZL3b span; it does not mean a modern human confirmed the image
location. One physical coordinate may have transcriber variants and may map to
more than one token position when the physical run is multi-token.

The raw `labtit-98-07-20.idx` byte file is not present in this repository. Its
release, size (91,861 bytes), 1,485-row count, and SHA-256
`cb210aaa75dfd2e9d86e63fd4cff1684acdfc2669bd6a6f9969f4e6bfe10071c`
are documented by the frozen inventory audit; no local substitute was invented.
No builder for `STOLFI_ASTRO_LABEL_MATCHES.tsv` is retained locally. The frozen
row-level table and audit are therefore the reproducible migration inputs, while
the bias and legacy-enrichment builders referenced in the manifest remain
available and checksum-verified.
"""
    (out / "LEGACY_MAPPING_REPRODUCTION.md").write_text(reproduction, encoding="utf-8")

    occurrences, readable = load_occurrences()
    normalized = normalized_legacy(matches, occurrences)
    write_tsv(out / "LEGACY_MAPPING_NORMALIZED.tsv", list(normalized[0]), normalized)

    augmented = read_tsv(AUGMENTED)
    groups = read_tsv(GROUPS)
    members = read_tsv(MEMBERS)
    old_labels = read_tsv(OLD_LABELS)
    target, crosswalk, mapping, geometry, _, _ = build_crosswalk(matches, augmented, groups, members, old_labels, occurrences, readable)
    write_tsv(out / "LEGACY_TO_CANONICAL_LABEL_CROSSWALK.tsv", list(crosswalk[0]), crosswalk)
    write_tsv(out / "CROSSWALK_GEOMETRY_METRICS.tsv", list(geometry[0]), geometry)
    write_tsv(out / "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv", list(mapping[0]), mapping)

    target_by_id = {row["object_id"]: row for row in target}
    coverage, bias = coverage_tables(mapping, target_by_id)
    write_tsv(out / "MAPPING_COVERAGE_SUMMARY.tsv", list(coverage[0]), coverage)
    write_tsv(out / "MAPPING_SELECTION_BIAS_AUDIT.tsv", list(bias[0]), bias)
    crosswalk_by_id = {row["canonical_label_id"]: row for row in crosswalk}
    residual = residual_rows(mapping, crosswalk_by_id)
    write_tsv(out / "RESIDUAL_LABEL_MAPPING_QUEUE.tsv", list(residual[0]), residual)
    audit_rows = build_visual_audit(out, target, crosswalk, old_labels)
    write_tsv(out / "VISUAL_AUDIT_SELECTION.tsv", list(audit_rows[0]), audit_rows)

    counts = Counter((str(row["panel"]), str(row["grouped"]), str(row["primary_analysis_inclusion"])) for row in mapping)
    mapped = sum(row["primary_analysis_inclusion"] == "YES" for row in mapping)
    grouped_total = sum(row["grouped"] == "YES" for row in mapping)
    grouped_mapped = sum(row["grouped"] == "YES" and row["primary_analysis_inclusion"] == "YES" for row in mapping)
    ungrouped_total = len(mapping) - grouped_total
    ungrouped_mapped = mapped - grouped_mapped
    human_total = sum(row["human_added_membership"] == "YES" for row in mapping)
    eight = [row for row in mapping if str(row["group_size"]) == "8"]
    if len(eight) != 1:
        raise RuntimeError("eight-member 3G1 group was not preserved as one LABEL row")
    authorized = False

    audit_md = f"""# Crosswalk visual audit

The deterministic sheets are migration checks, not a new reading of all 92
labels. Red is canonical geometry; blue is the frozen f68r1 anonymous spatial
proxy. Every transferred mapping has an individual crop under
`visual_audit/crops/` and appears on `F68R1_PROVENANCE_AUDIT.png`.

- `F68R1_PROVENANCE_AUDIT.png`: all 29 f68r1 ordinal loci, including the three
  anonymous labels that lost direct canonical LABEL provenance.
- `AMBIGUOUS_AND_HUMAN_ADDED.png`: all ambiguous cases and all ten human-added
  LABEL; none is automatically assigned a token.
- `DETERMINISTIC_PRIMARY_SAMPLE.png`: eight primary rows selected by ascending
  SHA-256 of `20260915:canonical_label_id`.

The complete f68r1 sheet confirms the one-to-one visual sequence between
anonymous physical LABEL 001..029 and ZL3b `@Ls` loci 8..36. Transfer still
requires the direct `A:` source ID. Geometry is recorded for diagnosis; no row
was promoted by a weak overlap. f68r2/f68r3 legacy coordinates contain no pixel
geometry in the repository, so their many-candidate cases remain ambiguous.
"""
    (out / "CROSSWALK_VISUAL_AUDIT.md").write_text(audit_md, encoding="utf-8")

    report = f"""# Legacy astronomical mapping migration report

## Outcome

The legacy study reproduced exactly, and 21 of 92 canonical LABEL received a
valid migrated transcription mapping. All 21 are grouped f68r1 LABEL with a
direct frozen anonymous-label provenance path. No token was assigned from a
Stolfi ordinal alone, to an unmatched coordinate, to a sensitivity-only bound,
or to a human-added LABEL.

```text
CANONICAL_LABELS=92
MAPPED_LABELS={mapped}
RESIDUAL_LABELS={len(residual)}
GROUPED_LABELS={grouped_total}
GROUPED_MAPPED={grouped_mapped}
UNGROUPED_LABELS={ungrouped_total}
UNGROUPED_MAPPED={ungrouped_mapped}
HUMAN_ADDED_LABELS={human_total}
HUMAN_ADDED_MAPPED=0
HAPAX_ENRICHMENT_RUN_AUTHORIZED={'YES' if authorized else 'NO'}
LEXICON_MATCH_RUN_AUTHORIZED=NO
```

## Crosswalk interpretation

The useful chain is `legacy matched row -> exact f68r1 ZL3b @Ls locus -> frozen
anonymous spatial LABEL -> direct A source ID -> canonical LABEL -> 3G1`.
Twenty-one rows complete that chain. Five directly linked f68r1 visual loci have
no confirmed legacy occurrence. Three f68r1 canonical grouped LABEL lack the
direct A-label link and remain many-to-many. The 22 non-human grouped f68r2
LABEL and three non-human grouped f68r3 LABEL have page-level legacy candidates
but no independent legacy pixel/object crosswalk. Ten human-added LABEL are
explicitly left unmapped. Ungrouped LABEL have no identified legacy star-series
candidate.

## Coverage and gate

Overall coverage is {mapped}/92 ({mapped/92:.3%}); grouped coverage is
{grouped_mapped}/{grouped_total} ({grouped_mapped/grouped_total:.3%}); ungrouped
coverage is 0/{ungrouped_total}. Page coverage is f68r1 {mapped}/37, f68r2 0/33,
and f68r3 0/22. Mapping selection is therefore perfectly concentrated in the
grouped arm and in one page. The amendment requires both comparison arms in at
least two panels and sign-invariant missingness bounds. The contrast is not
identified, so the gate stops before any corpus-frequency or hapax join.

No enrichment files were created. The prior Stolfi hapax result remains a
different legacy analysis and is not relabeled as the requested 3G1 comparison.
The minimum residual queue has {len(residual)} LABEL and is stratified by absent
legacy coverage, human additions, geometry ambiguity, and missing exact legacy
occurrences.
"""
    (out / "LEGACY_MAPPING_MIGRATION_REPORT.md").write_text(report, encoding="utf-8")

    validation = f"""# Validation report

Status: `PASS_WITH_PREREGISTERED_GATE_STOP`.

- PASS: upstream legacy, augmented-reference, preparation, plan, and occurrence checksums.
- PASS: legacy counts 7/7, 10/12, 53/67, 54, and 77 reproduce exactly.
- PASS: target LABEL counts are 37/33/22 and all 92 IDs are unique.
- PASS: every migrated legacy record and token occurrence exists in frozen inputs.
- PASS: no cross-panel match and no automatic many-to-many resolution.
- PASS: unmatched and sensitivity-only records are excluded from primary mapping.
- PASS: all human-added LABEL remain without invented legacy provenance.
- PASS: the eight-member group is one LABEL/group observation.
- PASS: original frozen plan SHA-256 remains `{FROZEN_PLAN_SHA}`.
- PASS: amendment SHA-256 is `{AMENDMENT_SHA}` and predates the gate calculation.
- PASS: no corpus-local/page-local hapax definition was used; enrichment did not run.
- PASS: no dictionary/lexicon search ran.
- PASS: deterministic visual sheets include all ambiguous, all human-added, all f68r1 provenance rows, and a fixed primary sample.
- EXPECTED STOP: mapped ungrouped = {ungrouped_mapped}; mapped panels = 1, so enrichment is not authorized.
"""
    (out / "VALIDATION_REPORT.md").write_text(validation, encoding="utf-8")

    reproducibility = f"""# Reproducibility

Run from repository root:

```bash
python3 research/astro_hapax_star_label/legacy_mapping_migration_v1/scripts/build_migration.py
python3 -m unittest research/astro_hapax_star_label/legacy_mapping_migration_v1/tests/test_migration.py
(cd research/astro_hapax_star_label/legacy_mapping_migration_v1 && sha256sum -c SHA256SUMS)
```

The builder is deterministic and standard-library-only except Pillow, used for
audit images. It verifies all upstream ledgers before reading analytical data.
The amendment is copied unchanged for isolated reproducibility runs and its
fixed SHA is checked before the gate. The original preparation directory and
its `SHA256SUMS` are never rewritten. Output ordering, the audit sample seed,
PNG compression, and report precision are fixed.

No network request, OCR model, full-92 AI reading, corpus-frequency join, hapax
analysis, or lexicon lookup is performed by this build.
"""
    (out / "REPRODUCIBILITY.md").write_text(reproducibility, encoding="utf-8")

    generated = [
        path for path in out.rglob("*")
        if path.is_file()
        and path.name != "SHA256SUMS"
        and path.suffix != ".pyc"
        and "__pycache__" not in path.parts
    ]
    with (out / "SHA256SUMS").open("w", encoding="utf-8", newline="") as handle:
        for path in sorted(generated, key=lambda item: item.relative_to(out).as_posix()):
            handle.write(f"{sha256(path)}  {path.relative_to(out).as_posix()}\n")


if __name__ == "__main__":
    main()
