#!/usr/bin/env python3
"""Prepare deterministic blind calibration packages for residual mapping v1."""

from __future__ import annotations

import csv
import hashlib
import json
import shutil
from collections import Counter
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "research/astro_hapax_star_label/residual_ai_mapping_v1"
LEGACY = ROOT / "research/astro_hapax_star_label/legacy_mapping_migration_v1"
BASE = ROOT / "research/astro_hapax_star_label"
PROTOCOL = OUT / "RESIDUAL_AI_MAPPING_PROTOCOL.md"
PANELS = ("f68r1", "f68r2", "f68r3")
PASS_NAMES = ("visual_a", "alignment_b_base", "visual_c", "alignment_b_shuffled")
FORBIDDEN_FIELD_FRAGMENTS = (
    "hapax", "frequency", "group", "star_count", "human_added", "provenance",
    "legacy", "origin", "hypothesis", "dictionary", "lexicon",
)


def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def rows(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], data: list[dict[str, object]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter="\t", lineterminator="\n")
        w.writeheader()
        w.writerows(data)


def write_json(path: Path, value: object) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def row_count(path: Path) -> str:
    if path.suffix in {".tsv", ".csv"}:
        with path.open(encoding="utf-8", errors="replace") as f:
            return str(max(0, sum(1 for _ in f) - 1))
    if path.suffix == ".jsonl":
        with path.open(encoding="utf-8") as f:
            return str(sum(1 for _ in f))
    return "NA"


def register_inputs() -> None:
    specs = [
        ("legacy canonical crosswalk", LEGACY / "LEGACY_TO_CANONICAL_LABEL_CROSSWALK.tsv"),
        ("legacy confirmed mapping", LEGACY / "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv"),
        ("residual queue", LEGACY / "RESIDUAL_LABEL_MAPPING_QUEUE.tsv"),
        ("legacy coverage summary", LEGACY / "MAPPING_COVERAGE_SUMMARY.tsv"),
        ("legacy migration report", LEGACY / "LEGACY_MAPPING_MIGRATION_REPORT.md"),
        ("legacy migration checksums", LEGACY / "SHA256SUMS"),
        ("prepared label registry and crop bindings", BASE / "LABEL_TOKEN_CANDIDATES.tsv"),
        ("candidate occurrence table", BASE / "TRANSCRIPTION_TOKEN_CANDIDATES.tsv"),
        ("candidate line table", BASE / "TRANSCRIPTION_LINE_CANDIDATES.tsv"),
        ("transcription conventions", BASE / "TRANSCRIPTION_POLICY.md"),
        ("preparation manifest", BASE / "PREPARATION_MANIFEST.json"),
        ("preparation checksums", BASE / "SHA256SUMS"),
        ("frozen canonical corpus", ROOT / "data_work/ZL3b-x7.canonical.txt"),
        ("frozen occurrence metadata", ROOT / "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl"),
        ("frozen analysis plan", BASE / "HAPAX_STAR_LABEL_ANALYSIS_PLAN.md"),
        ("frozen migration amendment", LEGACY / "HAPAX_STAR_LABEL_ANALYSIS_PLAN_AMENDMENT_01.md"),
        ("canonical augmented reference", ROOT / "research/astro_spatial_augmented_human_reference/AUGMENTED_HUMAN_OBJECTS.tsv"),
        ("3G1 group table", ROOT / "research/astro_spatial_augmented_human_reference/RELATION_GROUPS_3G1.tsv"),
        ("3G1 membership table", ROOT / "research/astro_spatial_augmented_human_reference/RELATION_GROUP_MEMBERS_3G1.tsv"),
        ("augmented reference checksums", ROOT / "research/astro_spatial_augmented_human_reference/SHA256SUMS"),
    ]
    for panel in PANELS:
        specs.append((f"canonical page image {panel}", ROOT / f"research/astro_spatial_human_adjudication/images/{panel}.jpg"))
    data = []
    for role, path in specs:
        if not path.exists():
            raise RuntimeError(f"missing input: {path}")
        data.append({
            "logical_role": role,
            "path": path.relative_to(ROOT).as_posix(),
            "row_count": row_count(path),
            "file_size": path.stat().st_size,
            "sha256": sha(path),
            "status": "FROZEN_READ_ONLY",
        })
    write_tsv(OUT / "INPUT_MANIFEST.tsv", list(data[0]), data)


def make_split() -> tuple[list[dict[str, str]], dict[str, dict[str, str]], dict[str, dict[str, str]]]:
    mapping = rows(LEGACY / "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv")
    known = [r for r in mapping if r["primary_analysis_inclusion"] == "YES"]
    label_rows = {r["label_id"]: r for r in rows(BASE / "LABEL_TOKEN_CANDIDATES.tsv")}
    if len(known) != 21:
        raise RuntimeError(f"expected 21 known mappings, got {len(known)}")
    known.sort(key=lambda r: (
        (float(label_rows[r["canonical_label_id"]]["bbox_y1"]) + float(label_rows[r["canonical_label_id"]]["bbox_y2"])) / 2,
        (float(label_rows[r["canonical_label_id"]]["bbox_x1"]) + float(label_rows[r["canonical_label_id"]]["bbox_x2"])) / 2,
    ))
    split = []
    neutral_crosswalk = {}
    hidden_answers = []
    known_by_id = {r["canonical_label_id"]: r for r in known}
    for idx, r in enumerate(known, 1):
        neutral = f"CR_LABEL_{idx:04d}"
        neutral_crosswalk[r["canonical_label_id"]] = neutral
    for stratum in range(7):
        members = known[stratum * 3:stratum * 3 + 3]
        hidden = min(members, key=lambda r: hashlib.sha256(("residual-ai-mapping-v1|" + r["canonical_label_id"]).encode()).hexdigest())
        for r in members:
            label = label_rows[r["canonical_label_id"]]
            role = "EVALUATION_HIDDEN" if r is hidden else "CALIBRATION_VISIBLE"
            split.append({
                "canonical_label_id": r["canonical_label_id"],
                "neutral_label_id": neutral_crosswalk[r["canonical_label_id"]],
                "panel": r["panel"],
                "split_role": role,
                "spatial_stratum": stratum + 1,
                "geometry_type": label["geometry_type"],
                "token_count": r["token_count"],
                "token_length_readable": len(r["raw_token_sequence"]),
                "bbox_center_x": f"{(float(label['bbox_x1']) + float(label['bbox_x2'])) / 2:.6f}",
                "bbox_center_y": f"{(float(label['bbox_y1']) + float(label['bbox_y2'])) / 2:.6f}",
                "selection_rule": "ONE_MIN_SHA256_PER_VERTICAL_STRATUM" if role == "EVALUATION_HIDDEN" else "STRATUM_REMAINDER",
            })
            if role == "EVALUATION_HIDDEN":
                hidden_answers.append({
                    "neutral_label_id": neutral_crosswalk[r["canonical_label_id"]],
                    "canonical_label_id": r["canonical_label_id"],
                    "correct_candidate_id": f"ZC_{int(r['absolute_token_positions']):05d}",
                    "correct_occurrence_id": r["absolute_token_positions"],
                    "correct_raw_token": r["raw_token_sequence"],
                    "correct_normalized_token": r["normalized_token_sequence"],
                })
    split.sort(key=lambda r: r["neutral_label_id"])
    write_tsv(OUT / "CALIBRATION_SPLIT.tsv", list(split[0]), split)
    secret = OUT / "sealed_answers"
    write_tsv(secret / "NEUTRAL_LABEL_ID_CROSSWALK.tsv", ["neutral_label_id", "canonical_label_id"], [
        {"neutral_label_id": neutral, "canonical_label_id": canonical}
        for canonical, neutral in sorted(neutral_crosswalk.items(), key=lambda x: x[1])
    ])
    write_tsv(secret / "EVALUATION_HIDDEN_ANSWERS.tsv", list(hidden_answers[0]), hidden_answers)
    return split, label_rows, known_by_id


def candidate_rows(order_seed: str) -> list[dict[str, str]]:
    source = rows(BASE / "TRANSCRIPTION_TOKEN_CANDIDATES.tsv")
    clean = [{
        "candidate_id": f"ZC_{int(r['occurrence_id']):05d}",
        "panel": r["panel"],
        "line_ref": r["line_ref"],
        "locus_type": r["locus_type"],
        "index_in_line": r["index_in_line"],
        "canonical_token_key": r["canonical_token_key"],
        "readable_eva": r["readable_eva"],
        "label_text_status": r["label_text_status"],
        "missing_status": r["missing_status"],
    } for r in source]
    clean.sort(key=lambda r: hashlib.sha256((order_seed + "|" + r["candidate_id"]).encode()).hexdigest())
    for panel in PANELS:
        panel_rows = [r for r in clean if r["panel"] == panel]
        for rank, r in enumerate(panel_rows, 1):
            r["display_order"] = str(rank)
    clean.sort(key=lambda r: (PANELS.index(r["panel"]), int(r["display_order"])))
    return clean


def make_crop(source: Path, bbox: tuple[float, float, float, float], target: Path) -> None:
    image = Image.open(source).convert("RGB")
    x1, y1, x2, y2 = bbox
    pad_x = max(45, (x2 - x1) * .22)
    pad_y = max(45, (y2 - y1) * .65)
    box = (
        max(0, int(x1 - pad_x)), max(0, int(y1 - pad_y)),
        min(image.width, int(x2 + pad_x)), min(image.height, int(y2 + pad_y)),
    )
    crop = image.crop(box)
    scale = max(2, min(5, 900 // max(1, crop.width)))
    crop = crop.resize((crop.width * scale, crop.height * scale), Image.Resampling.LANCZOS)
    target.parent.mkdir(parents=True, exist_ok=True)
    crop.save(target, quality=96, subsampling=0)


def make_contact_sheet(package: Path, cases: list[dict[str, str]]) -> None:
    thumbs = []
    font = ImageFont.load_default()
    for case in cases:
        im = Image.open(package / case["crop_file"]).convert("RGB")
        im.thumbnail((700, 300), Image.Resampling.LANCZOS)
        canvas = Image.new("RGB", (720, 350), "white")
        canvas.paste(im, ((720 - im.width) // 2, 35 + (300 - im.height) // 2))
        ImageDraw.Draw(canvas).text((10, 10), case["neutral_label_id"], fill="black", font=font)
        thumbs.append(canvas)
    cols = 2
    rows_n = (len(thumbs) + cols - 1) // cols
    sheet = Image.new("RGB", (cols * 720, rows_n * 350), "#dddddd")
    for i, im in enumerate(thumbs):
        sheet.paste(im, ((i % cols) * 720, (i // cols) * 350))
    sheet.save(package / "CONTACT_SHEET.jpg", quality=95, subsampling=0)


def visual_prompt(pass_id: str) -> str:
    return f"""# {pass_id.upper()} blind visual transcription\n\nRead only files in this directory. Do not inspect any parent directory, repository file, sealed answer, prior mapping, or other agent output. This is an instruction-enforced clean room.\n\nFor every row in CASES.tsv, inspect its crop and page context. Describe visible glyph form only; do not infer meaning or use external Voynich knowledge. You have no candidate list. Preserve uncertainty and abstain when needed. ZL3b/EVA display conventions: C=cth, K=ckh, P=cph, F=cfh, N=iin, A=ain, H=ch, S=sh, E=ee, I=in.\n\nWrite one compact JSON object per line, in CASES.tsv order, with exactly these keys: neutral_label_id, outcome, visual_glyph_sequence, token_boundaries, direction, start_point, uncertain_glyph_positions, alternative_visual_readings, confidence, rationale. Allowed outcomes: SINGLE_TOKEN_READING, MULTI_TOKEN_READING, RING_CYCLIC_READING, AMBIGUOUS_READING, UNREADABLE. Confidence: HIGH, MEDIUM, LOW. Use JSON arrays for uncertain_glyph_positions and alternative_visual_readings. Never force a reading.\n"""


def alignment_prompt(order_name: str) -> str:
    return f"""# AI_ALIGNMENT_B blind candidate ranking ({order_name})\n\nRead only files in this directory. Do not inspect any parent directory, repository file, sealed answer, prior mapping, or other agent output. This is an instruction-enforced clean room.\n\nFor every row in CASES.tsv, inspect its crop and the complete same-panel candidate list in CANDIDATES.tsv. Rank by glyph-by-glyph visual alignment, never by display position, uniqueness, semantics, or expected yield. You do not have Visual A/C. Preserve uncertainty and choose NO_MATCH when no candidate adequately accounts for the image. ZL3b/EVA display conventions: C=cth, K=ckh, P=cph, F=cfh, N=iin, A=ain, H=ch, S=sh, E=ee, I=in.\n\nWrite one compact JSON object per line, in CASES.tsv order, with exactly these keys: neutral_label_id, outcome, ranked_candidate_ids, glyph_alignment, unmatched_visible_glyphs, unmatched_transcription_glyphs, boundary_agreement, direction_start_assessment, top1_confidence, top2_margin, alternatives, rationale. outcome is RANKED_MATCH or NO_MATCH. ranked_candidate_ids and alternatives are JSON arrays with at most three candidate IDs. top1_confidence is HIGH, MEDIUM, or LOW; top2_margin is LARGE, SMALL, TIE, or NOT_APPLICABLE. Never force a match.\n"""


def build_packages(split: list[dict[str, str]], label_rows: dict[str, dict[str, str]]) -> None:
    clean_root = OUT / "cleanroom"
    if clean_root.exists():
        shutil.rmtree(clean_root)
    case_rows = []
    canonical_by_neutral = {r["neutral_label_id"]: r["canonical_label_id"] for r in split}
    for r in split:
        source = label_rows[r["canonical_label_id"]]
        case_rows.append({
            "neutral_label_id": r["neutral_label_id"],
            "panel": r["panel"],
            "case_phase": r["split_role"],
            "geometry_type": source["geometry_type"],
            "bbox_x1": source["bbox_x1"], "bbox_y1": source["bbox_y1"],
            "bbox_x2": source["bbox_x2"], "bbox_y2": source["bbox_y2"],
            "rotation": source["rotation"],
            "crop_file": f"crops/{r['neutral_label_id']}.jpg",
            "page_file": f"pages/{r['panel']}.jpg",
        })
    for pass_name in PASS_NAMES:
        package = clean_root / pass_name
        (package / "crops").mkdir(parents=True, exist_ok=True)
        (package / "pages").mkdir(parents=True, exist_ok=True)
        for panel in sorted({r["panel"] for r in case_rows}):
            shutil.copyfile(ROOT / f"research/astro_spatial_human_adjudication/images/{panel}.jpg", package / f"pages/{panel}.jpg")
        for r in case_rows:
            source = label_rows[canonical_by_neutral[r["neutral_label_id"]]]
            bbox = tuple(float(source[k]) for k in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"))
            make_crop(ROOT / f"research/astro_spatial_human_adjudication/images/{r['panel']}.jpg", bbox, package / r["crop_file"])
        write_tsv(package / "CASES.tsv", list(case_rows[0]), case_rows)
        make_contact_sheet(package, case_rows)
        if pass_name == "visual_a":
            (package / "PROMPT.md").write_text(visual_prompt("AI_VISUAL_A"), encoding="utf-8")
        elif pass_name == "visual_c":
            (package / "PROMPT.md").write_text(visual_prompt("AI_VISUAL_C"), encoding="utf-8")
        else:
            seed = "candidate-base-v1" if pass_name.endswith("base") else "candidate-shuffle-v1"
            cand = candidate_rows(seed)
            write_tsv(package / "CANDIDATES.tsv", list(cand[0]), cand)
            (package / "PROMPT.md").write_text(alignment_prompt("base order" if pass_name.endswith("base") else "shuffled order"), encoding="utf-8")


def audit_and_manifest() -> None:
    records = []
    violations = []
    for package in sorted((OUT / "cleanroom").iterdir()):
        for path in sorted(p for p in package.rglob("*") if p.is_file()):
            rel = path.relative_to(OUT).as_posix()
            records.append({
                "package": package.name,
                "relative_path": rel,
                "file_size": path.stat().st_size,
                "sha256": sha(path),
            })
            if path.suffix.lower() in {".tsv", ".md", ".json", ".jsonl"}:
                text = path.read_text(encoding="utf-8").lower()
                header = text.splitlines()[0] if text.splitlines() else ""
                for term in FORBIDDEN_FIELD_FRAGMENTS:
                    if term in header:
                        violations.append(f"{rel}: forbidden field/header fragment {term}")
                if "correct_candidate" in text or "correct_occurrence" in text:
                    violations.append(f"{rel}: answer-bearing text")
    write_tsv(OUT / "CLEAN_ROOM_PACKAGE_MANIFEST.tsv", list(records[0]), records)
    counts = Counter(r["package"] for r in records)
    report = [
        "# Blindness audit", "", "Result: **PASS**" if not violations else "Result: **FAIL**", "",
        "The audit scanned all clean-room text headers and content for answer-bearing fields and forbidden metadata. Images are copied or deterministically cropped from canonical pages; no annotations, group marks, or answers are rendered into them.", "",
        "Excluded from packages: canonical-ID crosswalk, known answers, group membership and IDs, STAR counts, human-added status, provenance, hapax/frequency data, dictionaries, naming hypotheses, and prior AI outputs.", "",
        "Per-package files: " + ", ".join(f"`{k}`={v}" for k, v in sorted(counts.items())) + ".", "",
        "Agent filesystem isolation is instruction-enforced rather than OS-enforced; the prompts explicitly prohibit reading outside the package. Package content blindness is checksum-audited.", "",
    ]
    if violations:
        report += ["Violations:", ""] + [f"- {v}" for v in violations]
    (OUT / "BLINDNESS_AUDIT.md").write_text("\n".join(report) + "\n", encoding="utf-8")
    if violations:
        raise RuntimeError("clean-room blindness audit failed")


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    if not PROTOCOL.exists():
        raise RuntimeError("protocol must exist before preparation")
    protocol_sha = sha(PROTOCOL)
    freeze = OUT / "PROTOCOL_FREEZE.json"
    if freeze.exists():
        existing = json.loads(freeze.read_text(encoding="utf-8"))
        if existing["sha256"] != protocol_sha:
            raise RuntimeError("frozen protocol changed")
    else:
        write_json(freeze, {"path": PROTOCOL.name, "sha256": protocol_sha, "status": "FROZEN_BEFORE_AI_PASS", "version": "1.0"})
    register_inputs()
    split, label_rows, _ = make_split()
    build_packages(split, label_rows)
    audit_and_manifest()
    (OUT / "agent_outputs").mkdir(exist_ok=True)
    summary = {
        "status": "CALIBRATION_PACKAGES_READY",
        "protocol_sha256": protocol_sha,
        "known_mappings": len(split),
        "calibration_visible": sum(r["split_role"] == "CALIBRATION_VISIBLE" for r in split),
        "evaluation_hidden": sum(r["split_role"] == "EVALUATION_HIDDEN" for r in split),
        "candidate_counts": dict(Counter(r["panel"] for r in candidate_rows("candidate-base-v1"))),
    }
    write_json(OUT / "PRE_RUN_STATE.json", summary)
    print(json.dumps(summary, sort_keys=True))


if __name__ == "__main__":
    main()
