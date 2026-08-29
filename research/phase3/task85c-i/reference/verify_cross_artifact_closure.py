#!/usr/bin/env python3
"""Verify the single V1.2 integration invariant and immutable parent pins."""
import hashlib
import json
from pathlib import Path

HERE=Path(__file__).resolve().parent; OUT=HERE.parent; REPO=HERE.parents[3]
V12="G1_V2_EXECUTABLE_CONTRACT_V1_2"
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
g=REPO/"research/phase3/task85c-g"; c=REPO/"research/phase3/task85c-c"; e=REPO/"research/phase3/task85c-e"; h=REPO/"research/phase3/task85c-h"
pins={g/"G1_V2_EXECUTABLE_CONTRACT_V1_2.md":"ec60bb23e55ce157fe954b5cafc63d22ab70ecec390822cb63f9ae273142c639",g/"G1_V2_EXECUTABLE_CONTRACT_V1_2.json":"29e39e0c25dc8033f784480fdc537e3ede9eeb69baa0607c9f249d796d6b42dc",g/"G1V2_GENERATION_SEMANTICS_V1.json":"45d533f8b83b24c77a96836fa5c2ef95f9b948003bd2ed725fc2ea97e010b310",g/"G1V2_GENERATION_GOLDEN_SUITE_V1.json":"143954667073a2c10f1bd59ce98b9c93dd84b50632bb67ea80d0d92449480acb",e/"G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json":"dbfb9a4a7101eed7006f751b9c4631b5f0286c3792f9777cc833c5dcfa42a3d3",c/"registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json":"fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9",h/"TASK85C_H_RESULTS_MANIFEST.json":"c9c54c9b5c20dd746ace32ab3a3a9dc916a8ceae00710481e8226314db2ea795"}
for p,w in pins.items():assert sha(p)==w,p
i1=json.loads((OUT/"G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.json").read_text()); e2=json.loads((OUT/"G1V2_EXECUTION_IDENTITY_ERRATUM_E2.json").read_text()); reg=json.loads((OUT/"G1V2_V1_2_EVIDENCE_SCHEMA_REGISTRY.json").read_text())
values=[i1["scientific_contract"]["version"],i1["execution_identity"]["scientific_identity_version"],e2["scientific_contract_version"],e2["jobid"]["scientific_identity_version"],i1["evidence"]["scientific_contract_version"],reg["scientific_contract_version"]]
assert values==[V12]*len(values)
assert i1["scientific_semantic_change"]==0 and e2["scientific_boundary"]["scientific_design_changed"] is False
assert i1["status_reachability"]["sha256"]==pins[c/"registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json"]
print("CROSS_ARTIFACT_V1_2_CLOSURE=PASS")
print("SCIENTIFIC_SEMANTIC_CHANGE=0")
print("SCIENTIFIC_FIREWALL=INTACT")
