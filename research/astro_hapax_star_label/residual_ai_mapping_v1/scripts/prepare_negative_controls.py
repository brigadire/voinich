#!/usr/bin/env python3
"""Create blind negative controls without exposing the omitted correct answers."""

from __future__ import annotations

import csv
import hashlib
import json
import shutil
from pathlib import Path


ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "research/astro_hapax_star_label/residual_ai_mapping_v1"
DEST = OUT / "cleanroom/negative_controls"


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter="\t", lineterminator="\n")
        w.writeheader()
        w.writerows(rows)


def main() -> None:
    hidden = read_tsv(OUT / "sealed_answers/EVALUATION_HIDDEN_ANSWERS.tsv")
    all_candidates = read_tsv(OUT / "cleanroom/alignment_b_shuffled/CANDIDATES.tsv")
    f68r1 = [r for r in all_candidates if r["panel"] == "f68r1"]
    f68r2 = [r for r in all_candidates if r["panel"] == "f68r2"]
    if DEST.exists():
        shutil.rmtree(DEST)
    (DEST / "crops").mkdir(parents=True)
    source = OUT / "cleanroom/visual_a"
    cases = []
    candidates = []
    for i, answer in enumerate(hidden):
        neutral = answer["neutral_label_id"]
        shutil.copyfile(source / f"crops/{neutral}.jpg", DEST / f"crops/{neutral}.jpg")
        controls = [
            ("OMIT", "SAME_PAGE_CORRECT_OMITTED", [r for r in f68r1 if r["candidate_id"] != answer["correct_candidate_id"]]),
            ("WRONG", "WRONG_PAGE_CANDIDATES", f68r2),
            ("UNREL", "VISUALLY_UNRELATED_RESTRICTED_SET", [
                r for r in f68r1
                if r["candidate_id"] in {
                    hidden[(i + 1) % len(hidden)]["correct_candidate_id"],
                    hidden[(i + 2) % len(hidden)]["correct_candidate_id"],
                    hidden[(i + 3) % len(hidden)]["correct_candidate_id"],
                }
            ]),
        ]
        for short, kind, pool in controls:
            control_id = f"NC_{i + 1:02d}_{short}"
            cases.append({
                "control_id": control_id,
                "neutral_label_id": neutral,
                "control_type": kind,
                "crop_file": f"crops/{neutral}.jpg",
                "candidate_count": len(pool),
            })
            ordered = sorted(pool, key=lambda r: hashlib.sha256(("negative-control-v1|" + control_id + "|" + r["candidate_id"]).encode()).hexdigest())
            for display_order, r in enumerate(ordered, 1):
                candidates.append({
                    "control_id": control_id,
                    "display_order": display_order,
                    "candidate_id": r["candidate_id"],
                    "panel": r["panel"],
                    "line_ref": r["line_ref"],
                    "canonical_token_key": r["canonical_token_key"],
                    "readable_eva": r["readable_eva"],
                })
    write_tsv(DEST / "CASES.tsv", list(cases[0]), cases)
    write_tsv(DEST / "CANDIDATES.tsv", list(candidates[0]), candidates)
    prompt = """# Blind negative-control alignment test\n\nRead only files in this directory. Do not inspect any parent directory, repository file, sealed answer, prior mapping, or other agent output. For every CASES.tsv row inspect the crop and only that control_id's candidate rows. The sets are deliberately adverse. Test whether visual evidence supports any candidate; do not assume a match exists. Candidate order is random. Do not use meaning.\n\nWrite exactly one compact JSON object per CASES row, in order, with keys: control_id, neutral_label_id, outcome, selected_candidate_id, confidence, abstained, rationale. outcome is RANKED_MATCH, AMBIGUOUS, or NO_MATCH. selected_candidate_id is an empty string on abstention. confidence is HIGH/MEDIUM/LOW. abstained is true for AMBIGUOUS and NO_MATCH. Never force a candidate.\n"""
    (DEST / "PROMPT.md").write_text(prompt, encoding="utf-8")
    print(json.dumps({"controls": len(cases), "candidate_rows": len(candidates), "hidden_answers_in_package": False}, sort_keys=True))


if __name__ == "__main__":
    main()
