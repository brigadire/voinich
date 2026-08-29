#!/usr/bin/env python3
"""Reproduce Task85c-h's V1.2 evidence/JobID authority contradiction."""
import hashlib, json
from pathlib import Path

HERE=Path(__file__).resolve().parent
OUT=HERE.parent
REPO=HERE.parents[3]
V12=REPO/"research/phase3/task85c-g"
V11=REPO/"research/phase3/task85c-c"
E1=REPO/"research/phase3/task85c-e/G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json"

pins={
 V12/"G1_V2_EXECUTABLE_CONTRACT_V1_2.md":"ec60bb23e55ce157fe954b5cafc63d22ab70ecec390822cb63f9ae273142c639",
 V12/"G1_V2_EXECUTABLE_CONTRACT_V1_2.json":"29e39e0c25dc8033f784480fdc537e3ede9eeb69baa0607c9f249d796d6b42dc",
 E1:"dbfb9a4a7101eed7006f751b9c4631b5f0286c3792f9777cc833c5dcfa42a3d3",
 V11/"registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json":"fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9"}
for path,want in pins.items(): assert hashlib.sha256(path.read_bytes()).hexdigest()==want,path

contract=json.loads((V12/"G1_V2_EXECUTABLE_CONTRACT_V1_2.json").read_text())
e1=json.loads(E1.read_text())
assert contract["contract_version"]=="G1_V2_EXECUTABLE_CONTRACT_V1_2"
assert e1["contract_version"]=="G1_V2_EXECUTABLE_CONTRACT_V1_1"
assert e1["jobid"]["scientific_identity_version"]=="G1_V2_EXECUTABLE_CONTRACT_V1_1"
assert contract["evidence_contract"]["precedence"][0]=="G1V2_EXECUTABLE_CONTRACT_V1_1.json"

schema_files=sorted((V11/"schemas").glob("*.schema.json"))
assert len(schema_files)==15
for path in schema_files:
    schema=json.loads(path.read_text())
    versions={branch["properties"]["contract_version"]["const"] for branch in schema["oneOf"]}
    assert versions=={"G1_V2_EXECUTABLE_CONTRACT_V1_1"},path

positive=json.loads((V11/"golden/schema-positive/cases.json").read_text())
fixture=next(case["instance"] for case in positive if case["schema"]=="generation" and case["instance"]["status"]=="GENERATION_SUCCESS")
assert fixture["contract_version"]=="G1_V2_EXECUTABLE_CONTRACT_V1_1"
v12_fixture=dict(fixture); v12_fixture["contract_version"]="G1_V2_EXECUTABLE_CONTRACT_V1_2"
schema=json.loads((V11/"schemas/generation.schema.json").read_text())
assert all(v12_fixture["contract_version"]!=branch["properties"]["contract_version"]["const"] for branch in schema["oneOf"])

result={
 "finding_ids":["H-SC01-EVIDENCE-CONTRACT-VERSION","H-SC02-E1-JOBID-SCIENTIFIC-VERSION"],
 "v1_2_contract_version":contract["contract_version"],
 "e1_contract_version":e1["contract_version"],
 "e1_jobid_scientific_identity_version":e1["jobid"]["scientific_identity_version"],
 "evidence_schema_count":len(schema_files),
 "evidence_schema_allowed_contract_versions":["G1_V2_EXECUTABLE_CONTRACT_V1_1"],
 "v1_2_generation_fixture_schema_result":"REJECT",
 "v1_1_generation_fixture_schema_result":"ACCEPT_BY_FROZEN_POSITIVE_GOLDEN",
 "scientific_effect":"different evidence bytes/hashes and different JobIDs",
 "resolution_required":"scientific authority/schema refreeze; not an implementation choice"
}
(OUT/"TASK85C_H_AUTHORITY_CONFLICT_REPRODUCTION.json").write_text(json.dumps(result,sort_keys=True,separators=(",",":"))+"\n")
print("H_SC01=REPRODUCED")
print("H_SC02=REPRODUCED")
print("STOP=SCIENTIFIC_CONTRACT_DEFECT")
