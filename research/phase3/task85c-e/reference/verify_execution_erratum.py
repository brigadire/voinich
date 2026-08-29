#!/usr/bin/env python3
"""Fail-closed structural/invariant validator for E1."""

import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REPO = ROOT.parents[2]
e = json.loads((ROOT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json").read_text())

assert e["schema"] == "g1v2.execution-identity-erratum.e1"
assert e["erratum_id"] == "G1V2_EXECUTION_IDENTITY_ERRATUM_E1"
assert e["contract_version"] == "G1_V2_EXECUTABLE_CONTRACT_V1_1"
assert e["blind_id"] == {
    "algorithm": "HMAC-SHA256",
    "collision": "abort if one blind_id maps to distinct canonical truth-record bytes; no suffixing or regeneration",
    "domain_ascii": "G1V2-BLIND-ID",
    "encoding": "first 20 lowercase hexadecimal characters of full HMAC digest",
    "key": "exactly 32 bytes from an OS cryptographic random source; independent of G1V2-RNG-1",
    "message": "ASCII domain || NUL || u64be(canonical_truth_record_byte_length) || canonical_truth_record_bytes",
}
assert e["truth_record"]["additional_fields"] is False
assert "blind_id" in e["truth_record"]["forbidden_fields"]
assert e["jobid"]["dependency_field"] == "dependency_job_ids"
assert e["jobid"]["scientific_identity_version"] == "G1_V2_EXECUTABLE_CONTRACT_V1_1"
assert e["run_materialization"]["literal_blind_ids_are_scientific_variables"] is False
assert e["run_materialization"]["literal_jobids_are_scientific_variables"] is False

expected = {
    "research/phase3/task85c-c/G1V2_EXECUTABLE_CONTRACT_V1_1.md": "5c3cd272c1dbae9bfe1d7a100155faf102e86d34660da239e1cb31704ad470b0",
    "research/phase3/task85c-c/registries/G1V2_CANDIDATE_REGISTRY.tsv": "96b618ab324db77b8402081075241b275f8925c45cdab058ba741e4beed29b58",
    "research/phase3/task85c-c/registries/G1V2_RNG_DOMAIN_REGISTRY.tsv": "e47c5a8c62dd8dee34441e4274c0d49d1a7ce4aab0360aff4751cb08b6394a43",
    "research/phase3/task85c-c/registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json": "fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9",
}
for rel, digest in expected.items():
    assert hashlib.sha256((REPO / rel).read_bytes()).hexdigest() == digest, rel
print("EXECUTION_IDENTITY_ERRATUM_E1=PASS")
print("UNEXPECTED_SCIENTIFIC_CHANGE=0")
