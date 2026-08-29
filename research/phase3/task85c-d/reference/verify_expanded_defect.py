#!/usr/bin/env python3
"""Verify that the complete blind JobID set is unavailable by contract."""

import csv
import json
from pathlib import Path

REPO = Path(__file__).resolve().parents[4]
T85 = REPO / "research/phase3/task85c"
T85CC = REPO / "research/phase3/task85c-c"

dag = json.loads((T85CC / "registries/G1V2_DAG_CONTRACT.json").read_text())
escrow = json.loads((T85 / "G1V2_BLIND_ESCROW_SCHEMA.json").read_text())
with (T85CC / "registries/G1V2_CONTROL_REGISTRY.tsv").open(newline="") as f:
    controls = list(csv.DictReader(f, delimiter="\t"))

blind_rows = [x for x in controls if x["role"] == "BLIND_SYNTHETIC"]
assert len(blind_rows) == 12
assert dag["control_instances"]["blind_synthetic"] == 144
assert "random 32-byte escrow key" in escrow["creation"]
assert "HMAC-SHA256" in escrow["creation"]
assert "blind_id" in escrow["secret_fields"]
assert all("blind_id" not in row for row in blind_rows)
assert "control_instance_id" in dag["job_id"]["payload_fields"]
print("D01_BLIND_ID_CLOSURE=REPRODUCED")
print("BLIND_GENERATOR_ROWS=12")
print("REQUIRED_BLIND_INSTANCE_IDS=144")
print("AVAILABLE_FROZEN_BLIND_INSTANCE_IDS=0")
