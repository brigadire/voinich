#!/usr/bin/env python3
"""Small frozen-parent reproduction of H-SC01."""
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[3]
C = REPO / "research/phase3/task85c-c"
old = json.loads((C / "golden/schema-positive/cases.json").read_text())
fixture = next(x["instance"] for x in old if x["schema"] == "generation")
schema = json.loads((C / "schemas/generation.schema.json").read_text())
assert fixture["contract_version"] == "G1_V2_EXECUTABLE_CONTRACT_V1_1"
assert any(b["properties"]["contract_version"]["const"] == fixture["contract_version"] for b in schema["oneOf"])
changed = dict(fixture, contract_version="G1_V2_EXECUTABLE_CONTRACT_V1_2")
assert all(b["properties"]["contract_version"]["const"] != changed["contract_version"] for b in schema["oneOf"])
print("H_SC01_REPRODUCED=YES")
