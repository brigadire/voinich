#!/usr/bin/env python3
"""Reproduce EI01 without generating production secrets or controls."""

import csv
import hashlib
import hmac
import json
from pathlib import Path

REPO = Path(__file__).resolve().parents[4]
T85 = REPO / "research/phase3/task85c"
T85CC = REPO / "research/phase3/task85c-c"

escrow = json.loads((T85 / "G1V2_BLIND_ESCROW_SCHEMA.json").read_text())
with (T85CC / "registries/G1V2_RNG_DOMAIN_REGISTRY.tsv").open(newline="") as f:
    domains = {r["domain_id"]: r for r in csv.DictReader(f, delimiter="\t")}
blind = domains["BLIND_ID"]

assert blind["namespace"] == "g1v2/blind/id"
assert blind["counter_fields"] == "generator_index,scale_index,replicate"
assert blind["output_interpretation"] == "digest"
assert "HMAC-SHA256" in escrow["creation"]
assert "random 32-byte escrow key" in escrow["creation"]
assert "truth_record_fields" not in escrow
assert "hmac_message" not in escrow
assert "domain_separator" not in escrow
assert "blind_id" in escrow["secret_fields"]

# Artificial diagnostic key/content only; this is not a production escrow key.
key = bytes(range(32))
base = {
    "class": "M0",
    "content_sha256": "00" * 32,
    "generator_id": "M0_GEN_A",
    "parameters": {"diagnostic": True},
    "seed": "11" * 32,
}
extended = dict(base, generator_index=0, replicate=0, scale=2000)

def cj(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode()

id_a = hmac.new(key, cj(base), hashlib.sha256).hexdigest()[:20]
id_b = hmac.new(key, cj(extended), hashlib.sha256).hexdigest()[:20]
assert id_a != id_b
print("EI01_EXECUTION_IDENTITY_DEFECT=REPRODUCED")
print("HMAC_TRUTH_FIELDS_ID=" + id_a)
print("HMAC_COUNTER_BOUND_ID=" + id_b)
print("PRODUCTION_SECRET_GENERATED=NO")
