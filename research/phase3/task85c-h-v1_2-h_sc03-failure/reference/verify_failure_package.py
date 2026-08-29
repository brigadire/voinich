#!/usr/bin/env python3
"""Verify the bounded Task85c-h scientific-contract-defect package."""
from __future__ import annotations

import hashlib
import json
import subprocess
from pathlib import Path

HERE = Path(__file__).resolve().parent
TASK = HERE.parent

subprocess.run(["python3", str(HERE / "reproduce_m0_golden_contradiction.py")], check=True)
proof = json.loads((TASK / "TASK85C_H_AUTHORITY_CONFLICT_REPRODUCTION.json").read_text())
assert proof["contradiction"] is True
assert proof["m0_fit_stated_probability_mass"] == "1.000"
assert proof["m0_fit_remaining_probability_mass_for_unk"] == "0.000"
assert proof["normative_v1_2_denominator"] == "9"
marker_files = [p for p in TASK.iterdir() if p.name == "TASK85C_H_SCIENTIFIC_CONTRACT_DEFECT"]
assert len(marker_files) == 1
manifest = json.loads((TASK / "TASK85C_H_RESULTS_MANIFEST.json").read_text())
for item in manifest["artifacts"]:
    path = TASK / item["path"]
    assert hashlib.sha256(path.read_bytes()).hexdigest() == item["sha256"]
print("PASS task85c-h failure package")
