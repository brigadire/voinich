#!/usr/bin/env python3
"""Validate the E1 scientific/opaque identity boundary table."""

import csv
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
with (ROOT / "G1V2_BLIND_IDENTITY_BOUNDARY.tsv").open(newline="") as f:
    rows = list(csv.DictReader(f, delimiter="\t"))

allowed = {"SCIENTIFIC_NORMATIVE", "BLINDNESS_NORMATIVE", "INTEGRITY_NORMATIVE", "EXECUTION_PROVENANCE"}
assert len(rows) >= 10
assert all(r["classification"] in allowed and all(r.values()) for r in rows)
by_requirement = {r["requirement"]: r for r in rows}
assert by_requirement["G1V2-RNG-1 root/CONTROL_GENERATE counters"]["classification"] == "SCIENTIFIC_NORMATIVE"
assert by_requirement["blind_id literal bytes"]["classification"] == "EXECUTION_PROVENANCE"
assert by_requirement["JobID literal bytes"]["classification"] == "EXECUTION_PROVENANCE"
assert by_requirement["truth sealing and later authenticated opening"]["classification"] == "BLINDNESS_NORMATIVE"
assert by_requirement["HMAC algorithm/domain/framing"]["classification"] == "INTEGRITY_NORMATIVE"
print("BLIND_IDENTITY_BOUNDARY=PASS")
