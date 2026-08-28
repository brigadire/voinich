#!/usr/bin/env python3
"""Prove the frozen Task85c status/reachability inconsistency."""
from __future__ import annotations

import csv
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
T85 = ROOT.parent / "task85c"


def rows(name):
    with (T85 / name).open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


statuses = {r["status"]: r for r in rows("G1V2_STATUS_REGISTRY.tsv")}
reach = rows("G1V2_REACHABILITY_CONTRACT.tsv")
reach_statuses = {r["upstream_status"] for r in reach}

assert "SCIENTIFIC_FAILURE" in reach_statuses
assert "SCIENTIFIC_FAILURE" not in statuses

registered_failures = {"FIT_FAILURE", "NUMERICAL_FAILURE", "INDUCTION_CAP", "GENERATION_FAILURE", "PROTOCOL_VETO"}
assert registered_failures <= set(statuses)
assert registered_failures.isdisjoint(reach_statuses)

generic_generation = next(r for r in reach if r["upstream_stage"] == "GENERATION" and r["upstream_status"] == "SCIENTIFIC_FAILURE" and r["downstream_stage"] == "STRUCTURAL")
assert generic_generation["action"] == "RUN"
assert statuses["GENERATION_FAILURE"]["downstream"] == "structural NOT_REACHED"

fit_fail = [r for r in reach if r["upstream_stage"] == "FIT" and r["upstream_status"] == "FAIL"]
assert fit_fail
assert statuses["FAIL"]["legal_producer"] == "gate/verifier"

print("EXPANDED_DEFECT_REPRODUCED=E01_E02_E03_E04")

