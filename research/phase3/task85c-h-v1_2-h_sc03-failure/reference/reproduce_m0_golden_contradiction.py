#!/usr/bin/env python3
"""Prove the frozen Task85c M0 goldens cannot share one probability row."""
from __future__ import annotations

import hashlib
import json
from decimal import Decimal
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
OLD = ROOT / "research/phase3/task85c/golden/G1V2_GOLDEN_SUITE.json"
V12 = ROOT / "research/phase3/task85c-g/G1V2_GENERATION_GOLDEN_SUITE_V1.json"
CONTRACT = ROOT / "research/phase3/task85c-g/G1_V2_EXECUTABLE_CONTRACT_V1_2.json"
OUT = ROOT / "research/phase3/task85c-h/TASK85C_H_AUTHORITY_CONFLICT_REPRODUCTION.json"


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    old = json.loads(OLD.read_text(encoding="utf-8"))
    v12 = json.loads(V12.read_text(encoding="utf-8"))
    contract = json.loads(CONTRACT.read_text(encoding="utf-8"))
    fit = next(x for x in old["cases"] if x["id"] == "M0-FIT")
    unseen = next(x for x in old["cases"] if x["id"] == "M0-UNSEEN")
    inherited_fit = next(x for x in v12["cases"] if x["id"] == "INHERITED-M0-FIT")
    inherited_unseen = next(x for x in v12["cases"] if x["id"] == "INHERITED-M0-UNSEEN")
    assert inherited_fit["expected"] == "INHERITED_UNCHANGED"
    assert inherited_fit["source_case"] == fit
    assert inherited_unseen["expected"] == "INHERITED_UNCHANGED"
    assert inherited_unseen["source_case"] == unseen

    probabilities = [Decimal(fit["expected"][key]) for key in ("p_a", "p_b", "p_eos")]
    stated_mass = sum(probabilities)
    remaining_mass = Decimal(1) - stated_mass
    requires_positive_unk = unseen["expected"]["positive_alpha"] == "positive"
    contradiction = stated_mass == Decimal(1) and remaining_mass == 0 and requires_positive_unk
    assert contradiction

    # The normative M0 formula independently yields denominator 5 + 1*4 = 9:
    # DEVELOPMENT outcomes are a, b, UNK and EOS; observed count total is 5.
    normative_denominator = Decimal(5) + Decimal(1) * Decimal(4)
    result = {
        "schema": "task85c-h-scientific-contract-defect-v1",
        "finding": "H-SC03-M0-UNK-PROBABILITY-CONTRADICTION",
        "contract_version": contract["contract_version"],
        "contract_machine_sha256": sha(CONTRACT),
        "task85c_golden_sha256": sha(OLD),
        "generation_golden_suite_sha256": sha(V12),
        "m0_fit_expected_denominator": fit["expected"]["denominator"],
        "m0_fit_stated_probability_mass": str(stated_mass),
        "m0_fit_remaining_probability_mass_for_unk": str(remaining_mass),
        "m0_unseen_positive_alpha_requirement": unseen["expected"]["positive_alpha"],
        "normative_v1_2_outcomes": ["a", "b", "<UNK>", "<EOS>"],
        "normative_v1_2_denominator": str(normative_denominator),
        "competing_interpretations": {
            "honor_M0_FIT": "p(UNK)=0 and denominator=8; violates M0-UNSEEN positive-alpha golden and V1.2 M0 equation",
            "honor_M0_UNSEEN_and_V1_2": "p(UNK)=1/9 and denominator=9; violates mandatory inherited M0-FIT golden"
        },
        "scientific_effects": ["fitted_model", "predictive_values", "generated_corpus", "complexity_and_final_verdict"],
        "contradiction": contradiction
    }
    OUT.write_text(json.dumps(result, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n", encoding="utf-8")
    print("PASS H-SC03-M0-UNK-PROBABILITY-CONTRADICTION")


if __name__ == "__main__":
    main()
