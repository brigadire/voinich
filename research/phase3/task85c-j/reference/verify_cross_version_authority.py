#!/usr/bin/env python3
import hashlib,json
from pathlib import Path

out=Path(__file__).resolve().parent.parent
v="G1_V2_EXECUTABLE_CONTRACT_V1_2_1"
c=json.loads((out/"G1_V2_EXECUTABLE_CONTRACT_V1_2_1.json").read_text())
e=json.loads((out/"G1V2_EXECUTION_IDENTITY_ERRATUM_E3.json").read_text())
i=json.loads((out/"G1V2_V1_2_1_INTEGRATION_SUPPLEMENT_I2.json").read_text())
r=json.loads((out/"G1V2_V1_2_1_EVIDENCE_SCHEMA_REGISTRY.json").read_text())
j=json.loads((out/"G1V2_E3_JOBID_REGRESSION.json").read_text())
assert {c["contract_version"],e["scientific_contract_version"],e["jobid"]["scientific_identity_version"],i["scientific_contract_version"],r["scientific_contract_version"]}=={v}
assert j["different"] and j["dependency_field"]=="dependency_job_ids"
assert len(r["entries"])==15 and all(x["scientific_contract_version"]==v for x in r["entries"])
digest=lambda name:hashlib.sha256((out/name).read_bytes()).hexdigest()
assert i["scientific_contract_machine_sha256"]==digest("G1_V2_EXECUTABLE_CONTRACT_V1_2_1.json")
assert i["scientific_contract_markdown_sha256"]==digest("G1_V2_EXECUTABLE_CONTRACT_V1_2_1.md")
assert i["m0_semantics_sha256"]==digest("G1V2_M0_SEMANTICS.json")
assert i["execution_identity"]["machine_sha256"]==digest("G1V2_EXECUTION_IDENTITY_ERRATUM_E3.json")
assert i["evidence"]["registry_sha256"]==digest("G1V2_V1_2_1_EVIDENCE_SCHEMA_REGISTRY.json")
gold=json.loads((out/"G1V2_SCIENTIFIC_GOLDEN_ROOT_V1_2_1.json").read_text())
assert i["scientific_golden_root_sha256"]==gold["root_sha256"]
print("CROSS_VERSION_AUTHORITY=PASS")
