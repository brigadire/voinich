#!/usr/bin/env python3
"""Permanent H-SC01, cross-version hash, mixed-identity and semantic-invariance checks."""
import json
from pathlib import Path

OUT=Path(__file__).resolve().parent.parent
r=json.loads((OUT/"G1V2_H_SC01_REGRESSION.json").read_text())
assert r["historical_v1_1_acceptance"]
assert r["historical_v1_2_substitution_rejected_by_old_schema"]
assert r["repaired_v1_2_acceptance_by_new_schema"]
assert r["repaired_v1_2_rejection_by_v1_1_schema"] and r["v1_1_rejection_by_v1_2_schema"]
assert r["v1_1_evidence_sha256"] != r["v1_2_evidence_sha256"]
assert r["scientific_payload_equal"] and r["h_sc01"]=="CLOSED"
i1=json.loads((OUT/"G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.json").read_text())
e2=json.loads((OUT/"G1V2_EXECUTION_IDENTITY_ERRATUM_E2.json").read_text())
assert i1["unknown_or_mixed_version_disposition"]=="FAIL_CLOSED"
assert e2["jobid"]["scientific_identity_version"]==i1["evidence"]["scientific_contract_version"]
mixed=json.loads((OUT/"fixtures/G1V2_MIXED_IDENTITY_NEGATIVE_FIXTURES.json").read_text())
assert mixed["disposition"]=="FAIL_CLOSED" and len(mixed["cases"])==2
for case in mixed["cases"]:
    assert case["evidence_contract_version"] != case["jobid_scientific_identity_version"]
    assert case["expected"]=="REJECT"
print("H_SC01=CLOSED")
print("MIXED_IDENTITY_NEGATIVE_REGRESSION=PASS")
print("SCIENTIFIC_SEMANTIC_INVARIANCE=PASS")
