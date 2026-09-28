#!/usr/bin/env python3
"""Compute requested post-review metrics; never promotes consensus automatically."""
from __future__ import annotations

import argparse
import csv
import json
from pathlib import Path

PKG = Path(__file__).resolve().parents[1]


def load(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as handle:
        return list(csv.DictReader(handle, delimiter="\t"))


def rate(rows, decisions):
    accepted = {decisions} if isinstance(decisions, str) else set(decisions)
    return sum(r["human_decision"] in accepted for r in rows) / len(rows) if rows else None


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--human-stars", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=PKG / "exports/HUMAN_AGREEMENT_METRICS.json")
    args = parser.parse_args()
    candidates = {r["candidate_id"]: r for r in load(PKG / "HUMAN_CANDIDATE_OBJECTS.tsv") if r["candidate_type"] == "STAR_OBJECT"}
    joined = [(candidates[r["candidate_id"]], r) for r in load(args.human_stars) if r["candidate_id"] in candidates]
    groups = {
        "AI_CONSENSUS": [h for c, h in joined if c["support_ai1"] == c["support_ai2"] == "YES"],
        "AI1_ONLY": [h for c, h in joined if c["agreement_type"] == "SUPPORT_AI1_ONLY"],
        "AI2_ONLY": [h for c, h in joined if c["agreement_type"] == "SUPPORT_AI2_ONLY"],
    }
    f68 = [(c, h) for c, h in joined if c["panel"] == "f68r1"]
    result = {
        # MODIFY confirms existence but corrects class/geometry, so it is a
        # positive existence outcome for accept-rate metrics.
        "HUMAN_ACCEPT_RATE_AI_CONSENSUS": rate(groups["AI_CONSENSUS"], {"ACCEPT", "MODIFY"}),
        "HUMAN_ACCEPT_RATE_AI1_ONLY": rate(groups["AI1_ONLY"], {"ACCEPT", "MODIFY"}),
        "HUMAN_ACCEPT_RATE_AI2_ONLY": rate(groups["AI2_ONLY"], {"ACCEPT", "MODIFY"}),
        "HUMAN_REJECT_RATE_AI_CONSENSUS": rate(groups["AI_CONSENSUS"], "REJECT"),
        "HUMAN_UNCERTAIN_RATE": rate([h for _, h in joined], "UNCERTAIN"),
        "A_AI1_HUMAN_AGREEMENT": rate([h for c, h in f68 if c["support_a"] == c["support_ai1"] == "YES"], {"ACCEPT", "MODIFY"}),
        "A_AI2_HUMAN_AGREEMENT": rate([h for c, h in f68 if c["support_a"] == c["support_ai2"] == "YES"], {"ACCEPT", "MODIFY"}),
        "AI1_AI2_HUMAN_AGREEMENT": rate([h for c, h in f68 if c["support_ai1"] == c["support_ai2"] == "YES"], {"ACCEPT", "MODIFY"}),
        "denominators": {
            "AI_CONSENSUS": len(groups["AI_CONSENSUS"]), "AI1_ONLY": len(groups["AI1_ONLY"]), "AI2_ONLY": len(groups["AI2_ONLY"]),
            "ALL_REVIEWED": len(joined),
            "F68R1_A_AI1": sum(c["support_a"] == c["support_ai1"] == "YES" for c, _ in f68),
            "F68R1_A_AI2": sum(c["support_a"] == c["support_ai2"] == "YES" for c, _ in f68),
            "F68R1_AI1_AI2": sum(c["support_ai1"] == c["support_ai2"] == "YES" for c, _ in f68),
        },
        "reviewed_records": len(joined), "note": "ACCEPT and MODIFY are positive existence outcomes; null means no reviewed denominator; consensus is never auto-accepted",
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
