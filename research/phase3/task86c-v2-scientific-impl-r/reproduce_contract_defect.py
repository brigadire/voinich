#!/usr/bin/env python3
"""Reproduce Task85c A17 status/schema contradiction without dependencies."""
from __future__ import annotations

import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
T85 = ROOT.parent / "task85c"


def canonical_without_hash(obj):
    value = dict(obj)
    value.pop("content_sha256")
    return (json.dumps(value, ensure_ascii=False, sort_keys=True,
                       separators=(",", ":")) + "\n").encode()


def schema_accepts_reproducer(schema, obj):
    # The reproducer exercises only const/enum/required/if-then keywords used
    # by the frozen schemas.  This is deliberately not a replacement schema
    # implementation.
    props = schema["properties"]
    if obj["schema_id"] != props["schema_id"]["const"]:
        return False
    if obj["contract_version"] != props["contract_version"]["const"]:
        return False
    if obj["status"] not in props["status"]["enum"]:
        return False
    if not set(schema["required"]) <= set(obj):
        return False
    for clause in schema["allOf"]:
        condition = clause["if"]["properties"]["status"]
        matched = obj["status"] == condition.get("const") or obj["status"] in condition.get("enum", [])
        if matched:
            required = clause["then"]["properties"]["payload"].get("required", [])
            if not set(required) <= set(obj["payload"]):
                return False
    return True


with (T85 / "G1V2_STATUS_REGISTRY.tsv").open(encoding="utf-8", newline="") as f:
    status = {r["status"]: r for r in csv.DictReader(f, delimiter="\t")}

cases = [
    ("not_reached", "not_reached_with_pass.json", "PASS", "DAG materializer"),
    ("fit", "fit_with_fail.json", "FAIL", "fit"),
]
for schema_name, fixture_name, forbidden_status, legal_producer in cases:
    schema = json.loads((T85 / "schemas" / f"{schema_name}.schema.json").read_text())
    obj = json.loads((ROOT / "contract-defect" / fixture_name).read_text())
    assert hashlib.sha256(canonical_without_hash(obj)).hexdigest() == obj["content_sha256"]
    assert schema_accepts_reproducer(schema, obj)
    assert status[forbidden_status]["legal_producer"] != legal_producer

print("CONTRACT_DEFECT_REPRODUCED=A17_STATUS_SCHEMA_CONTRADICTION")
