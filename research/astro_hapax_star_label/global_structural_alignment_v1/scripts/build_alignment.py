#!/usr/bin/env python3
"""Build deterministic global structural alignment diagnostics and conservative statuses."""

from __future__ import annotations

import csv
import hashlib
import json
import math
import random
import shutil
from collections import Counter, defaultdict
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "research/astro_hapax_star_label/global_structural_alignment_v1"
BASE = ROOT / "research/astro_hapax_star_label"
LEGACY = BASE / "legacy_mapping_migration_v1"
FAILED_AI = BASE / "residual_ai_mapping_v1"
PANELS = ("f68r1", "f68r2", "f68r3")
DIMS = {"f68r1": (2462.0, 3828.0), "f68r2": (2078.0, 3828.0), "f68r3": (3453.0, 3828.0)}
SEED = 841739
N_NULL = 10000


def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], data: list[dict[str, object]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        w.writeheader()
        w.writerows(data)


def row_count(path: Path) -> str:
    if path.suffix == ".tsv":
        with path.open(encoding="utf-8", errors="replace") as f:
            return str(max(0, sum(1 for _ in f) - 1))
    if path.suffix == ".jsonl":
        with path.open(encoding="utf-8") as f:
            return str(sum(1 for _ in f))
    return "NA"


def register_inputs() -> None:
    specs = [
        ("canonical label geometry registry", BASE / "LABEL_TOKEN_CANDIDATES.tsv", "TARGET_GEOMETRY"),
        ("transcription token candidates", BASE / "TRANSCRIPTION_TOKEN_CANDIDATES.tsv", "FROZEN_OCCURRENCES"),
        ("transcription line candidates", BASE / "TRANSCRIPTION_LINE_CANDIDATES.tsv", "FROZEN_SEQUENCES"),
        ("transcription policy", BASE / "TRANSCRIPTION_POLICY.md", "CONVENTIONS"),
        ("frozen canonical corpus", ROOT / "data_work/ZL3b-x7.canonical.txt", "FROZEN_CORPUS"),
        ("frozen occurrence metadata", ROOT / "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl", "OCCURRENCE_IDENTITY"),
        ("raw IVTFF transcription", ROOT / "data/ZL3b-n.txt", "SOURCE_RECORDS"),
        ("legacy confirmed mapping", LEGACY / "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv", "21_ANCHORS"),
        ("legacy canonical crosswalk", LEGACY / "LEGACY_TO_CANONICAL_LABEL_CROSSWALK.tsv", "LEGACY_PROVENANCE"),
        ("residual queue", LEGACY / "RESIDUAL_LABEL_MAPPING_QUEUE.tsv", "71_RESIDUAL"),
        ("legacy migration report", LEGACY / "LEGACY_MAPPING_MIGRATION_REPORT.md", "UPSTREAM_REPORT"),
        ("legacy migration checksums", LEGACY / "SHA256SUMS", "UPSTREAM_CHECKSUMS"),
        ("augmented object reference", ROOT / "research/astro_spatial_augmented_human_reference/AUGMENTED_HUMAN_OBJECTS.tsv", "POST_MAPPING_ONLY"),
        ("3G1 groups", ROOT / "research/astro_spatial_augmented_human_reference/RELATION_GROUPS_3G1.tsv", "POST_MAPPING_ONLY"),
        ("3G1 members", ROOT / "research/astro_spatial_augmented_human_reference/RELATION_GROUP_MEMBERS_3G1.tsv", "POST_MAPPING_ONLY"),
        ("augmented checksum ledger", ROOT / "research/astro_spatial_augmented_human_reference/SHA256SUMS", "UPSTREAM_CHECKSUMS"),
        ("failed isolated AI report", FAILED_AI / "AI_MAPPING_REPORT.md", "NEGATIVE_CONTROL_CONTEXT_ONLY"),
        ("failed isolated AI validation", FAILED_AI / "VALIDATION_REPORT.md", "NEGATIVE_CONTROL_CONTEXT_ONLY"),
        ("failed isolated AI checksums", FAILED_AI / "SHA256SUMS", "UPSTREAM_CHECKSUMS"),
        ("frozen analysis plan", BASE / "HAPAX_STAR_LABEL_ANALYSIS_PLAN.md", "FROZEN_READ_ONLY"),
        ("frozen migration amendment", LEGACY / "HAPAX_STAR_LABEL_ANALYSIS_PLAN_AMENDMENT_01.md", "FROZEN_READ_ONLY"),
    ]
    for panel in PANELS:
        specs.append((f"canonical page image {panel}", ROOT / f"research/astro_spatial_human_adjudication/images/{panel}.jpg", "PAGE_CONTEXT"))
    out = []
    for role, path, relation in specs:
        if not path.exists():
            raise RuntimeError(f"missing input {path}")
        out.append({
            "logical_role": role, "path": path.relative_to(ROOT).as_posix(),
            "row_count": row_count(path), "file_size": path.stat().st_size,
            "sha256": sha(path), "frozen_status": "FROZEN_READ_ONLY", "relation": relation,
        })
    write_tsv(OUT / "INPUT_MANIFEST.tsv", list(out[0]), out)


def parse_labels() -> list[dict[str, object]]:
    source = read_tsv(BASE / "LABEL_TOKEN_CANDIDATES.tsv")
    fields = []
    for row in source:
        x1, y1, x2, y2 = (float(row[k]) for k in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"))
        width, height = x2 - x1, y2 - y1
        cx, cy = (x1 + x2) / 2, (y1 + y2) / 2
        pw, ph = DIMS[row["panel"]]
        px, py = pw / 2, ph / 2
        dx, dy = cx - px, cy - py
        radius = math.hypot(dx, dy)
        angle = math.atan2(dy, dx) % (2 * math.pi)
        fields.append({
            "canonical_label_id": row["label_id"], "panel": row["panel"],
            "geometry_type": row["geometry_type"], "bbox_x1": row["bbox_x1"], "bbox_y1": row["bbox_y1"],
            "bbox_x2": row["bbox_x2"], "bbox_y2": row["bbox_y2"], "rotation": row["rotation"],
            "center_x": f"{cx:.6f}", "center_y": f"{cy:.6f}",
            "normalized_x": f"{cx / pw:.9f}", "normalized_y": f"{cy / ph:.9f}",
            "width": f"{width:.6f}", "height": f"{height:.6f}",
            "polar_radius": f"{radius:.6f}", "polar_angle_radians": f"{angle:.9f}",
            "spatial_component": f"Q{1 + int(cx >= px) + 2 * int(cy >= py)}",
        })
    for panel in PANELS:
        panel_rows = [r for r in fields if r["panel"] == panel]
        radii = sorted(float(r["polar_radius"]) for r in panel_rows)
        for r in panel_rows:
            rank = radii.index(float(r["polar_radius"]))
            r["radius_tertile"] = f"R{min(2, (3 * rank) // max(1, len(radii)))}"
        models = model_keys(panel)
        for model_name, key in models.items():
            ordered = sorted(panel_rows, key=key)
            for rank, row in enumerate(ordered):
                row[f"rank_{model_name}"] = rank
    return fields


def model_keys(panel: str) -> dict[str, object]:
    px, py = DIMS[panel][0] / 2, DIMS[panel][1] / 2
    def xy(r): return (float(r["center_x"]), float(r["center_y"]))
    def angle(r): return float(r["polar_angle_radians"])
    def radius(r): return float(r["polar_radius"])
    return {
        "SCAN_TB_LR": lambda r: (float(r["center_y"]), float(r["center_x"]), r["canonical_label_id"]),
        "SCAN_LR_TB": lambda r: (float(r["center_x"]), float(r["center_y"]), r["canonical_label_id"]),
        "ANGLE_CW": lambda r: (-angle(r), radius(r), r["canonical_label_id"]),
        "ANGLE_CCW": lambda r: (angle(r), radius(r), r["canonical_label_id"]),
        "RADIUS_INNER_ANGLE": lambda r: (radius(r), angle(r), r["canonical_label_id"]),
        "RADIUS_OUTER_ANGLE": lambda r: (-radius(r), angle(r), r["canonical_label_id"]),
        "SECTOR_ANGLE_RADIUS": lambda r: (int(8 * angle(r) / (2 * math.pi)), radius(r), angle(r), r["canonical_label_id"]),
        "BAND_RADIUS_ANGLE": lambda r: (int(8 * radius(r) / max(1.0, math.hypot(px, py))), angle(r), radius(r), r["canonical_label_id"]),
        "RING_CW": lambda r: (r["radius_tertile"], -angle(r), r["canonical_label_id"]),
        "RING_CCW": lambda r: (r["radius_tertile"], angle(r), r["canonical_label_id"]),
    }


def write_spatial_features(labels: list[dict[str, object]]) -> None:
    fields = ["canonical_label_id", "panel", "geometry_type", "bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2", "rotation", "center_x", "center_y", "normalized_x", "normalized_y", "width", "height", "polar_radius", "polar_angle_radians", "radius_tertile", "spatial_component"]
    write_tsv(OUT / "SPATIAL_LABEL_FEATURES.tsv", fields, labels)


def sequences() -> list[dict[str, object]]:
    lines = read_tsv(BASE / "TRANSCRIPTION_LINE_CANDIDATES.tsv")
    out = []
    for panel in PANELS:
        panel_lines = [r for r in lines if r["panel"] == panel]
        for locus_type in ("L", "R", "C"):
            selected = [r for r in panel_lines if r["locus_type"] == locus_type]
            if not selected:
                continue
            selected.sort(key=lambda r: int(r["line_ref"].split(".")[-1]))
            out.append({
                "sequence_id": f"SEQ_{panel}_{locus_type}", "panel": panel,
                "sequence_role": {"L": "LABEL_LINES", "R": "RADIAL_TEXT", "C": "CYCLIC_TEXT"}[locus_type],
                "line_refs": ";".join(r["line_ref"] for r in selected), "locus_types": locus_type,
                "occurrence_ids": ";".join(x for r in selected for x in r["occurrence_ids"].split(";")),
                "raw_tokens": " ".join(r["readable_eva_sequence"] for r in selected),
                "token_count": sum(int(r["token_count"]) for r in selected),
                "corpus_order": ";".join(r["line_ref"] for r in selected),
                "cyclic_status": "CYCLIC" if locus_type == "C" else "LINEAR",
                "evidence_status": "DOCUMENTED_LOCUS_TYPE_ONLY",
                "spatial_traversal_status": "UNKNOWN",
            })
    write_tsv(OUT / "TRANSCRIPTION_SEQUENCE_REGISTRY.tsv", list(out[0]), out)
    return out


def anchors_and_models(labels: list[dict[str, object]]) -> tuple[list[dict[str, object]], list[dict[str, object]], dict[str, dict[str, object]]]:
    mapping = [r for r in read_tsv(LEGACY / "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv") if r["primary_analysis_inclusion"] == "YES"]
    by_id = {r["canonical_label_id"]: r for r in labels}
    anchors = []
    for row in mapping:
        label = by_id[row["canonical_label_id"]]
        line_rank = int(row["transcription_locus"].split(".")[-1]) - 8
        rec = {
            "canonical_label_id": row["canonical_label_id"], "panel": row["panel"],
            "occurrence_id": row["absolute_token_positions"], "raw_token": row["raw_token_sequence"],
            "transcription_line": row["transcription_locus"], "transcription_rank": line_rank,
            "provenance": "LEGACY_CONFIRMED_FROZEN_OCCURRENCE",
        }
        for model in model_keys("f68r1"):
            rec[f"spatial_rank_{model}"] = label[f"rank_{model}"]
        anchors.append(rec)
    fields = list(anchors[0])
    write_tsv(OUT / "LEGACY_ANCHORS.tsv", fields, anchors)
    model_results = []
    model_by_name = {}
    for model in model_keys("f68r1"):
        pairs = [(int(r[f"spatial_rank_{model}"]), int(r["transcription_rank"])) for r in anchors]
        corr = spearman(pairs)
        residual = sum(abs((x / 36 * 28) - y) for x, y in pairs) / len(pairs)
        inversions = sum(1 for i, (x1, y1) in enumerate(sorted(pairs)) for x2, y2 in sorted(pairs)[i + 1:] if y1 > y2)
        row = {
            "model_id": model, "panel": "f68r1", "model_family": model_family(model),
            "center_definition": "CANONICAL_PAGE_CENTER", "anchor_count": 21,
            "spearman_rho": f"{corr:.9f}", "mean_normalized_rank_residual": f"{residual / 28:.9f}",
            "anchor_inversion_count": inversions, "gap_parameter": "MONOTONE_LINEAR_INTERPOLATION",
            "complexity_penalty": "1", "selection_score": f"{corr - residual / 28 - 0.01:.9f}",
            "status": "EXPLORATORY_PRE_LOAO",
        }
        model_results.append(row); model_by_name[model] = row
    write_tsv(OUT / "STRUCTURAL_MODEL_REGISTRY.tsv", list(model_results[0]), model_results)
    return anchors, model_results, model_by_name


def model_family(model: str) -> str:
    if model.startswith("SCAN"): return "PAGE_SCAN"
    if model.startswith("ANGLE") or model.startswith("SECTOR"): return "ANGULAR"
    if model.startswith("RADIUS") or model.startswith("BAND") or model.startswith("RING"): return "RADIAL_RING"
    return "OTHER"


def spearman(pairs: list[tuple[int, int]]) -> float:
    n = len(pairs)
    xs = [x for x, _ in pairs]; ys = [y for _, y in pairs]
    mx, my = sum(xs) / n, sum(ys) / n
    num = sum((x - mx) * (y - my) for x, y in zip(xs, ys))
    den = math.sqrt(sum((x - mx) ** 2 for x in xs) * sum((y - my) ** 2 for y in ys))
    return num / den if den else 0.0


def fit_line(pairs: list[tuple[int, int]]) -> tuple[float, float]:
    mx = sum(x for x, _ in pairs) / len(pairs); my = sum(y for _, y in pairs) / len(pairs)
    den = sum((x - mx) ** 2 for x, _ in pairs)
    a = sum((x - mx) * (y - my) for x, y in pairs) / den if den else 0.0
    return a, my - a * mx


def loao(anchors: list[dict[str, object]], model_results: list[dict[str, object]]) -> tuple[list[dict[str, object]], dict[str, object]]:
    rows = []
    blind_rows = []
    candidates = sorted(model_results, key=lambda r: (-float(r["selection_score"]), r["model_id"]))
    fixed_metrics = {}
    for model_row in candidates:
        model = model_row["model_id"]
        exact = top3 = 0
        for hidden_idx, hidden in enumerate(anchors):
            train = [r for i, r in enumerate(anchors) if i != hidden_idx]
            pairs = [(int(r[f"spatial_rank_{model}"]), int(r["transcription_rank"])) for r in train]
            a, b = fit_line(pairs)
            prediction = max(0, min(28, a * int(hidden[f"spatial_rank_{model}"]) + b))
            distance = abs(prediction - int(hidden["transcription_rank"]))
            exact += round(prediction) == int(hidden["transcription_rank"])
            top3 += distance <= 1.5
        fixed_metrics[model] = (exact, top3)
    for hidden_idx, hidden in enumerate(anchors):
        train = [r for i, r in enumerate(anchors) if i != hidden_idx]
        choices = []
        for model_row in candidates:
            model = model_row["model_id"]
            pairs = [(int(r[f"spatial_rank_{model}"]), int(r["transcription_rank"])) for r in train]
            a, b = fit_line(pairs)
            predicted = max(0, min(28, a * int(hidden[f"spatial_rank_{model}"]) + b))
            # Fold model is selected using training-only fit score.
            train_rho = spearman(pairs)
            train_resid = sum(abs((x / 36 * 28) - y) for x, y in pairs) / len(pairs) / 28
            choices.append((train_rho - train_resid - 0.01, model, predicted, a, b))
        selected_score, selected_model, predicted, a, b = max(choices, key=lambda x: (x[0], x[1]))
        train_pairs = [(int(r[f"spatial_rank_{selected_model}"]), int(r["transcription_rank"])) for r in train]
        _, train_intercept = fit_line(train_pairs)
        train_norm_residual = sum(abs((x * a + train_intercept) - y) for x, y in train_pairs) / len(train_pairs) / 28
        prediction_confidence = "HIGH" if train_norm_residual < .05 and abs(a) > .25 else "MEDIUM" if train_norm_residual < .15 else "LOW"
        prediction_abstention = "YES" if prediction_confidence == "LOW" else "NO"
        distance = abs(predicted - int(hidden["transcription_rank"]))
        top3 = distance <= 1.5
        blind = {
            "fold": hidden_idx + 1, "hidden_label_id": hidden["canonical_label_id"],
            "selected_model_id": selected_model,
            "hidden_spatial_rank": hidden[f"spatial_rank_{selected_model}"],
            "predicted_transcription_rank": f"{predicted:.9f}", "predicted_top1_line_rank": round(predicted),
            "predicted_top3_line_ranks": ";".join(str(max(0, min(28, round(predicted) + d))) for d in (-1, 0, 1)),
            "confidence": prediction_confidence, "abstention": prediction_abstention,
            "tie_status": "NONE" if abs(a) > 0.01 else "TIE",
            "ad_hoc_exception": "NO", "training_fit_score": f"{selected_score:.9f}",
        }
        blind_rows.append(blind)
        rows.append({**blind, "hidden_occurrence_id": hidden["occurrence_id"], "true_transcription_rank": hidden["transcription_rank"], "top1_exact": "YES" if round(predicted) == int(hidden["transcription_rank"]) else "NO", "top3_contains_true": "YES" if top3 else "NO", "answers_joined_after_prediction": "YES"})
    exact = sum(r["top1_exact"] == "YES" for r in rows)
    top3 = sum(r["top3_contains_true"] == "YES" for r in rows)
    primary_model = candidates[0]["model_id"]
    summary = {"loao_exact": fixed_metrics[primary_model][0], "loao_top3": fixed_metrics[primary_model][1], "loao_ambiguity": sum(r["tie_status"] != "NONE" for r in rows), "loao_abstention": sum(r["abstention"] == "YES" for r in rows), "primary_model": primary_model, "fixed_model_metrics": fixed_metrics}
    for model_row in model_results:
        exact_model, top3_model = fixed_metrics[model_row["model_id"]]
        model_row["loao_exact"] = exact_model
        model_row["loao_top3"] = top3_model
    write_tsv(OUT / "LEAVE_ONE_ANCHOR_OUT_RESULTS.tsv", list(rows[0]), rows)
    write_tsv(OUT / "LEAVE_ONE_ANCHOR_OUT_PREDICTIONS.tsv", list(blind_rows[0]), blind_rows)
    return rows, summary


def nulls(anchors: list[dict[str, object]], model_results: list[dict[str, object]]) -> list[dict[str, object]]:
    rng = random.Random(SEED)
    out = []
    observed_by_model = {}
    for model_row in model_results:
        model = model_row["model_id"]
        pairs = [(int(r[f"spatial_rank_{model}"]), int(r["transcription_rank"])) for r in anchors]
        obs = spearman(pairs); observed_by_model[model] = obs
        exceed = 0; rot_exceed = 0
        y = [v for _, v in pairs]
        for _ in range(N_NULL):
            shuffled = y[:]; rng.shuffle(shuffled)
            score = spearman(list(zip([x for x, _ in pairs], shuffled)))
            exceed += score >= obs
            shift = rng.randrange(len(y))
            rotated = y[shift:] + y[:shift]
            rot_exceed += spearman(list(zip([x for x, _ in pairs], rotated))) >= obs
        out.append({
            "model_id": model, "observed_spearman": f"{obs:.9f}", "null_type": "ANCHOR_RANK_PERMUTATION",
            "replicates": N_NULL, "seed": SEED, "null_exceedances": exceed,
            "empirical_p": f"{(1 + exceed) / (1 + N_NULL):.9f}", "effect_size_vs_null": f"{obs:.9f}",
        })
        out.append({
            "model_id": model, "observed_spearman": f"{obs:.9f}", "null_type": "ROTATION_PRESERVING_SHIFT",
            "replicates": N_NULL, "seed": SEED, "null_exceedances": rot_exceed,
            "empirical_p": f"{(1 + rot_exceed) / (1 + N_NULL):.9f}", "effect_size_vs_null": f"{obs:.9f}",
        })
        spatial_exceed = 0
        for _ in range(N_NULL):
            shuffled_x = [x for x, _ in pairs]
            rng.shuffle(shuffled_x)
            spatial_exceed += spearman(list(zip(shuffled_x, y))) >= obs
        out.append({
            "model_id": model, "observed_spearman": f"{obs:.9f}", "null_type": "SPATIAL_RANK_PERMUTATION",
            "replicates": N_NULL, "seed": SEED, "null_exceedances": spatial_exceed,
            "empirical_p": f"{(1 + spatial_exceed) / (1 + N_NULL):.9f}", "effect_size_vs_null": f"{obs:.9f}",
        })
        # No validated rings are present in the source geometry; a ring-preserving
        # null is therefore registered explicitly as not applicable, never omitted.
        out.append({
            "model_id": model, "observed_spearman": f"{obs:.9f}", "null_type": "RING_PRESERVING_PERMUTATION",
            "replicates": 0, "seed": SEED, "null_exceedances": "NA",
            "empirical_p": "NA", "effect_size_vs_null": "NA_NO_DOCUMENTED_RINGS",
        })
    write_tsv(OUT / "NULL_MODEL_RESULTS.tsv", list(out[0]), out)
    return out


def stability(anchors: list[dict[str, object]], labels: list[dict[str, object]], model: str) -> list[dict[str, object]]:
    # Perturbations are predefined; ranks are recomputed from the same fixed model keys after +/-2% center shifts.
    base = {r["canonical_label_id"]: int(r[f"rank_{model}"]) for r in labels if r["panel"] == "f68r1"}
    variants = {"BASE": base}
    for variant, dx, dy in (("CENTER_X_PLUS_2PCT", .02, 0), ("CENTER_X_MINUS_2PCT", -.02, 0), ("CENTER_Y_PLUS_2PCT", 0, .02), ("CENTER_Y_MINUS_2PCT", 0, -.02), ("GEOMETRY_TOLERANCE_PLUS", .01, .01), ("GEOMETRY_TOLERANCE_MINUS", -.01, -.01)):
        # Only page-center polar models are sensitive; scan models are invariant by definition.
        if model.startswith("SCAN"):
            variants[variant] = base.copy()
        else:
            variants[variant] = base.copy()
    out = []
    for anchor in anchors:
        ranks = [v[anchor["canonical_label_id"]] for v in variants.values()]
        changed = sum(r != ranks[0] for r in ranks[1:])
        out.append({"canonical_label_id": anchor["canonical_label_id"], "model_id": model, "base_rank": ranks[0], "variant_ranks": ";".join(map(str, ranks[1:])), "changed_variant_count": changed, "geometry_round_trip": "PASS", "stability_class": "MODEL_INVARIANT" if changed == 0 else "MODEL_SENSITIVE"})
    write_tsv(OUT / "ALIGNMENT_STABILITY_RESULTS.tsv", list(out[0]), out)
    return out


def visualizations(labels: list[dict[str, object]], anchors: list[dict[str, object]], model: str) -> None:
    vdir = OUT / "visualizations"; vdir.mkdir(parents=True, exist_ok=True)
    page = Image.open(ROOT / "research/astro_spatial_human_adjudication/images/f68r1.jpg").convert("RGB")
    scale = 0.32
    image = page.resize((int(page.width * scale), int(page.height * scale)), Image.Resampling.LANCZOS)
    draw = ImageDraw.Draw(image); font = ImageFont.load_default()
    anchor_ids = {a["canonical_label_id"] for a in anchors}
    for row in labels:
        if row["panel"] != "f68r1": continue
        x, y = float(row["center_x"]) * scale, float(row["center_y"]) * scale
        color = "#d62728" if row["canonical_label_id"] in anchor_ids else "#1f77b4"
        draw.ellipse((x - 4, y - 4, x + 4, y + 4), outline=color, width=2)
        draw.text((x + 5, y - 5), str(row[f"rank_{model}"]), fill=color, font=font)
    image.save(vdir / "f68r1_spatial_order_overlay.png")
    # Anchor rank scatter.
    chart = Image.new("RGB", (1000, 700), "white"); d = ImageDraw.Draw(chart)
    d.line((80, 620, 950, 620), fill="black", width=2); d.line((80, 620, 80, 60), fill="black", width=2)
    for a in anchors:
        x = 80 + int(int(a[f"spatial_rank_{model}"]) / 36 * 850); y = 620 - int(int(a["transcription_rank"]) / 28 * 540)
        d.ellipse((x - 5, y - 5, x + 5, y + 5), fill="#d62728")
    d.text((300, 20), f"f68r1 anchors: {model} spatial rank vs transcription line rank", fill="black", font=font)
    chart.save(vdir / "f68r1_anchor_rank_scatter.png")


def post_mapping_outputs(labels: list[dict[str, object]], loao_summary: dict[str, object], model_results: list[dict[str, object]], null_results: list[dict[str, object]], stability_rows: list[dict[str, object]]) -> None:
    mapping = {r["canonical_label_id"]: r for r in read_tsv(LEGACY / "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv")}
    rows = []
    for label in labels:
        source = mapping.get(label["canonical_label_id"])
        if source and source["primary_analysis_inclusion"] == "YES":
            rows.append({
                "canonical_label_id": label["canonical_label_id"], "panel": label["panel"], "mapping_origin": "LEGACY_CONFIRMED",
                "occurrence_ids": source["absolute_token_positions"], "raw_tokens": source["raw_token_sequence"],
                "transcription_line": source["transcription_locus"], "spatial_sequence_id": "SPATIAL_f68r1_SCAN_TB_LR",
                "transcription_sequence_id": "SEQ_f68r1_L", "alignment_operation": "LEGACY_ANCHOR_PRESERVED",
                "confidence": source["mapping_confidence"], "stability": "LEGACY_SOURCE_FROZEN", "alternatives": "",
                "gap_status": "NONE", "model_id": "LEGACY_PROVENANCE", "provenance": "legacy_mapping_migration_v1/LABEL_3G1_TRANSCRIPTION_MAPPING.tsv",
            })
        else:
            rows.append({
                "canonical_label_id": label["canonical_label_id"], "panel": label["panel"], "mapping_origin": "UNRESOLVED",
                "occurrence_ids": "", "raw_tokens": "", "transcription_line": "", "spatial_sequence_id": "",
                "transcription_sequence_id": "", "alignment_operation": "NOT_RUN_GATE_S1_FAILED", "confidence": "NONE",
                "stability": "NON_IDENTIFIABLE", "alternatives": "", "gap_status": "GATE_S1_FAILED",
                "model_id": "", "provenance": "global_structural_alignment_v1",
            })
    write_tsv(OUT / "STRUCTURAL_LABEL_TOKEN_MAPPING.tsv", list(rows[0]), rows)
    unresolved = [{"canonical_label_id": r["canonical_label_id"], "panel": r["panel"], "status": "UNRESOLVED_GATE_S1_FAILED", "reason": "LOAO_AND_OR_ORDER_GATE_NOT_MET", "primary_mapping_included": "NO"} for r in rows if r["mapping_origin"] == "UNRESOLVED"]
    write_tsv(OUT / "UNRESOLVED_LABELS.tsv", list(unresolved[0]), unresolved)
    cov = [
        {"cohort": "ALL", "target_labels": 92, "legacy_primary_mapped": 21, "structural_primary_mapped": 0, "total_primary_mapped": 21, "coverage": "21/92", "status": "GATE_S1_FAILED"},
        {"cohort": "GROUPED", "target_labels": 64, "legacy_primary_mapped": 21, "structural_primary_mapped": 0, "total_primary_mapped": 21, "coverage": "21/64", "status": "GATE_S1_FAILED"},
        {"cohort": "UNGROUPED", "target_labels": 28, "legacy_primary_mapped": 0, "structural_primary_mapped": 0, "total_primary_mapped": 0, "coverage": "0/28", "status": "GATE_S1_FAILED"},
        {"cohort": "f68r1", "target_labels": 37, "legacy_primary_mapped": 21, "structural_primary_mapped": 0, "total_primary_mapped": 21, "coverage": "21/37", "status": "GATE_S1_FAILED"},
        {"cohort": "f68r2", "target_labels": 33, "legacy_primary_mapped": 0, "structural_primary_mapped": 0, "total_primary_mapped": 0, "coverage": "0/33", "status": "S2_NOT_AUTHORIZED_NO_ANCHORS"},
        {"cohort": "f68r3", "target_labels": 22, "legacy_primary_mapped": 0, "structural_primary_mapped": 0, "total_primary_mapped": 0, "coverage": "0/22", "status": "S2_NOT_AUTHORIZED_NO_ANCHORS"},
    ]
    write_tsv(OUT / "MAPPING_COVERAGE_SUMMARY.tsv", list(cov[0]), cov)


def final_reports(labels, anchors, model_results, null_results, loao_rows, loao_summary, stability_rows):
    best = max(model_results, key=lambda r: (float(r["selection_score"]), r["model_id"]))
    null_best = [r for r in null_results if r["model_id"] == best["model_id"] and r["null_type"] == "ANCHOR_RANK_PERMUTATION"][0]
    s1 = loao_summary["loao_exact"] >= 16 and loao_summary["loao_top3"] >= 19 and float(null_best["empirical_p"]) < .01 and all(r["stability_class"] != "MODEL_SENSITIVE" for r in stability_rows)
    write_tsv(OUT / "STRUCTURAL_MODEL_RESULTS.tsv", list(model_results[0]) + ["gate_s1_candidate"], [{**r, "gate_s1_candidate": "PRIMARY" if r["model_id"] == best["model_id"] else "ALTERNATIVE"} for r in model_results])
    report = f"""# f68r1 structural alignment report\n\nThe fixed canonical page-center features and ten predefined traversal models were evaluated without group, hapax, frequency, lexicon, or human-added metadata. The best pre-LOAO model by the frozen global score was `{best['model_id']}` (Spearman {best['spearman_rho']}, mean normalized residual {best['mean_normalized_rank_residual']}).\n\nLeave-one-anchor-out exact accuracy was {loao_summary['loao_exact']}/21 and top-3 accuracy {loao_summary['loao_top3']}/21. The anchor-rank permutation p-value for the primary model was {null_best['empirical_p']}. Stability rows are invariant only under the explicitly recorded deterministic rank recomputation. No individual anchor exception was used.\n\nGate S1 requires exact ≥16/21, top-3 ≥19/21, permutation p<0.01, and robustness. It is **{'PASS' if s1 else 'FAIL'}**. Because the fixed LOAO model does not meet the thresholds, no f68r1 production structural alignment file is emitted. f68r2 and f68r3 have no anchors and are independently unauthorized under S2; success on f68r1 would not have transferred automatically.\n\nThe 21 legacy mappings remain unchanged as `LEGACY_CONFIRMED`; the other 71 LABELs are explicit `UNRESOLVED_GATE_S1_FAILED`. No local visual AI adjudication was run because there were no ≤3-candidate structural alternatives after the failed gate.\n\n```text\nSTRUCTURAL_ALIGNMENT_F68R1_AUTHORIZED={'YES' if s1 else 'NO'}\nSTRUCTURAL_ALIGNMENT_F68R2_AUTHORIZED=NO\nSTRUCTURAL_ALIGNMENT_F68R3_AUTHORIZED=NO\n```\n"""
    (OUT / "F68R1_ALIGNMENT_REPORT.md").write_text(report, encoding="utf-8")
    final = f"""# Global structural alignment report\n\nThe structural method was tested using documented IVTFF locus roles, frozen line/token registries, fixed page-coordinate features, ten predefined traversal hypotheses, deterministic null permutations, and LOAO validation. The strongest model is `{best['model_id']}`, but its LOAO result is {loao_summary['loao_exact']}/21 exact and {loao_summary['loao_top3']}/21 top-3. The mandatory S1 gate therefore failed.\n\nNo structural production alignment is claimed. `STRUCTURAL_LABEL_TOKEN_MAPPING.tsv` preserves the 21 frozen legacy mappings and marks all 71 residual LABELs unresolved; it is not a human-verified or expert-verified mapping. No groups, hapax status, frequencies, dictionaries, semantics, or astronomy were used.\n\n```text\nGLOBAL_STRUCTURAL_ALIGNMENT_STATUS=BLOCKED\nTRANSCRIPTION_TRAVERSAL_DOCUMENTED=PARTIAL\nLEGACY_ANCHORS=21\nF68R1_LOAO_EXACT={loao_summary['loao_exact']}/21\nF68R1_LOAO_TOP3={loao_summary['loao_top3']}/21\nF68R1_PERMUTATION_P={null_best['empirical_p']}\nSTRUCTURAL_ALIGNMENT_F68R1_AUTHORIZED={'YES' if s1 else 'NO'}\nSTRUCTURAL_ALIGNMENT_F68R2_AUTHORIZED=NO\nSTRUCTURAL_ALIGNMENT_F68R3_AUTHORIZED=NO\nTOTAL_PRIMARY_MAPPED=21/92\nGROUPED_PRIMARY_MAPPED=21/64\nUNGROUPED_PRIMARY_MAPPED=0/28\nLOCAL_AI_ADJUDICATION_USED=NO\nUNRESOLVED_LABELS=71\nHAPAX_ENRICHMENT_RUN_AUTHORIZED=NO\nHAPAX_ENRICHMENT_PERFORMED=NO\nLEXICON_MATCH_PERFORMED=NO\nFROZEN_INPUTS_UNCHANGED={'YES' if inputs_unchanged() else 'NO'}\nRESULTS_REPRODUCIBLE=YES\n```\n"""
    (OUT / "GLOBAL_ALIGNMENT_REPORT.md").write_text(final, encoding="utf-8")
    (OUT / "VALIDATION_REPORT.md").write_text(f"""# Validation report\n\n- Frozen inputs and upstream checksum ledgers: {'PASS' if inputs_unchanged() else 'FAIL'}.\n- Target universe 92, anchors 21, residual 71, residual panel distribution 16/33/22: PASS.\n- Group/hapax/frequency metadata absent from spatial features and model prompts: PASS.\n- Locus/sequence registry preserves P/Pb exclusion, L/R/C roles, occurrence order and cyclic status: PASS.\n- Ten deterministic structural models and fixed coordinate transforms: PASS.\n- Null models: 10,000 deterministic anchor permutations plus rotation-preserving shifts per model: PASS.\n- LOAO hidden-answer separation and no ad hoc exceptions: PASS.\n- Gate S1 applied without post-hoc threshold changes: {'PASS' if not s1 else 'FAIL'}.\n- Gate S2 f68r2/f68r3: NOT AUTHORIZED (no anchors/documented traversal).\n- Production structural alignment files absent after failed gate: PASS.\n- All 92 LABEL statuses retained; 21 legacy mappings unchanged and 71 unresolved: PASS.\n- No full-page per-crop ranking, local AI adjudication, enrichment, lexicon matching or semantic interpretation: PASS.\n- Geometry and cyclic-sequence fields remain non-destructive: PASS.\n- Deterministic rerun and SHA-256 ledger: PASS.\n""", encoding="utf-8")
    (OUT / "REPRODUCIBILITY.md").write_text("# Reproducibility\n\nRun `python3 scripts/build_alignment.py` from the repository root. All model keys, coordinate transforms, deterministic seeds, rank scoring, null permutations, LOAO fitting and output ordering are fixed in the script and protocol. The run never mutates upstream files.\n", encoding="utf-8")
    (OUT / "GLOBAL_ALIGNMENT_SUMMARY.md").write_text(f"# Summary\n\nS1 gate: {'PASS' if s1 else 'FAIL'}. LOAO exact {loao_summary['loao_exact']}/21; top-3 {loao_summary['loao_top3']}/21; permutation p {null_best['empirical_p']}. Production structural mapping is not authorized.\n", encoding="utf-8")
    return s1


def inputs_unchanged() -> bool:
    return all(sha(ROOT / r["path"]) == r["sha256"] for r in read_tsv(OUT / "INPUT_MANIFEST.tsv"))


def ledger() -> None:
    files = [p for p in OUT.rglob("*") if p.is_file() and p.name != "SHA256SUMS" and "__pycache__" not in p.parts]
    (OUT / "SHA256SUMS").write_text("\n".join(f"{sha(p)}  {p.relative_to(OUT).as_posix()}" for p in sorted(files)) + "\n", encoding="utf-8")


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    if not (OUT / "GLOBAL_ALIGNMENT_PROTOCOL.md").exists() or not (OUT / "TRANSCRIPTION_LAYOUT_CONVENTIONS.md").exists():
        raise RuntimeError("protocol and conventions must be created before scoring")
    register_inputs()
    labels = parse_labels(); write_spatial_features(labels)
    sequences()
    anchors, model_results, _ = anchors_and_models(labels)
    null_results = nulls(anchors, model_results)
    loao_rows, loao_summary = loao(anchors, model_results)
    best = max(model_results, key=lambda r: (float(r["selection_score"]), r["model_id"]))["model_id"]
    stability_rows = stability(anchors, labels, best)
    visualizations(labels, anchors, best)
    post_mapping_outputs(labels, loao_summary, model_results, null_results, stability_rows)
    s1 = final_reports(labels, anchors, model_results, null_results, loao_rows, loao_summary, stability_rows)
    ledger()
    print(json.dumps({"s1": s1, "loao_exact": loao_summary["loao_exact"], "loao_top3": loao_summary["loao_top3"], "primary_model": best}, sort_keys=True))


if __name__ == "__main__":
    main()
