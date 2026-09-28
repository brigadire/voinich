#!/usr/bin/env python3
"""Generate M1_AUDIT_MANIFEST.json and SHA256SUMS for the audit deliverables.
This script touches only files inside research/astro_token_formation_m1_audit/;
it never writes to the frozen M0/M1 directories."""
from __future__ import annotations

import hashlib
import json
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "research/astro_token_formation_m1_audit"


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    artifacts = [
        "M1_METHODOLOGICAL_AUDIT.md",
        "M1_AUDIT_FINDINGS.tsv",
        "M1_AUDIT_DIAGNOSTICS.tsv",
        "M1_AUDIT_NULL_REVIEW.md",
        "M1_AUDIT_MATCHING_REVIEW.md",
        "M1_AUDIT_PERCLASS_NULL.tsv",
        "audit_diagnostics.py",
        "audit_per_class_null.py",
    ]
    inputs = [
        "research/astro_token_formation/ASTRO_TERM_CORPUS.tsv",
        "research/astro_token_formation/ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv",
        "research/astro_token_formation/TOKEN_FORMATION_RULE_SPACE.md",
        "research/astro_token_formation/TOKEN_FORMATION_MODELS.tsv",
        "research/astro_token_formation/TOKEN_FORMATION_MANIFEST.json",
        "research/astro_token_formation_m1/main.py",
        "research/astro_token_formation_m1/M1_SEARCH_CONFIG.yaml",
        "research/astro_token_formation_m1/M1_SUBSTITUTION_RULE_SPACE.md",
        "research/astro_token_formation_m1/M1_GRAPHEME_INVENTORY.tsv",
        "research/astro_token_formation_m1/M1_TOKEN_FORMATION_MODELS.tsv",
        "research/astro_token_formation_m1/M1_LABEL_TERM_ASSIGNMENTS.tsv",
        "research/astro_token_formation_m1/M1_HELDOUT_RESULTS.tsv",
        "research/astro_token_formation_m1/M1_NULL_RESULTS.tsv",
        "research/astro_token_formation_m1/M1_UNEXPLAINED_LABELS.tsv",
        "research/astro_token_formation_m1/M1_COMPARISON_WITH_M0.md",
        "research/astro_token_formation_m1/M1_TOKEN_FORMATION_REPORT.md",
        "research/astro_token_formation_m1/M1_MANIFEST.json",
        "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv",
    ]
    manifest = {
        "audit": "m1-methodological-audit-v1",
        "generated_utc": datetime.now(timezone.utc).replace(microsecond=0).isoformat(),
        "scope": "Independent methodological audit of research/astro_token_formation_m1 and its research/astro_token_formation (M0) foundation, per tasks_other/brutforce-astronomical-labels-M1-audit.md. No frozen M0/M1 artifact, corpus, rule, or threshold was modified.",
        "diagnostics_note": "All computation performed for this audit is tagged run_type=AUDIT_ONLY in M1_AUDIT_DIAGNOSTICS.tsv and M1_AUDIT_PERCLASS_NULL.tsv. No new expanded brute-force search was run; all reruns reuse the frozen M0/M1 corpus, rule grid, and (for null diagnostics) the same seed_for() function as the frozen production run.",
        "known_manifest_discrepancy": "research/astro_token_formation_m1/M1_MANIFEST.json records an input_sha256 for research/astro_token_formation_m1/main.py that does not match the currently committed file (see M1_AUDIT_FINDINGS.tsv F11); this audit's REPRODUCIBILITY_BEAM64 diagnostic independently reruns the current main.py and reproduces the frozen M1_001 result exactly.",
        "reviewed_input_sha256": {p: sha(ROOT / p) for p in inputs},
        "artifact_sha256": {p: sha(OUT / p) for p in artifacts},
    }
    (OUT / "M1_AUDIT_MANIFEST.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    checksum_files = artifacts + ["M1_AUDIT_MANIFEST.json"]
    (OUT / "SHA256SUMS").write_text(
        "".join(f"{sha(OUT / p)}  {p}\n" for p in sorted(checksum_files)), encoding="utf-8"
    )
    print("Wrote M1_AUDIT_MANIFEST.json and SHA256SUMS")


if __name__ == "__main__":
    main()
