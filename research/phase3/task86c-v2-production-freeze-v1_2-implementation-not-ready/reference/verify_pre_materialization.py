#!/usr/bin/env python3
import csv, hashlib, json
from pathlib import Path

HERE=Path(__file__).resolve().parent
OUT=HERE.parent
REPO=HERE.parents[3]

expected={
 "research/phase3/task85c-g/G1_V2_EXECUTABLE_CONTRACT_V1_2.md":"ec60bb23e55ce157fe954b5cafc63d22ab70ecec390822cb63f9ae273142c639",
 "research/phase3/task85c-g/G1_V2_EXECUTABLE_CONTRACT_V1_2.json":"29e39e0c25dc8033f784480fdc537e3ede9eeb69baa0607c9f249d796d6b42dc",
 "research/phase3/task85c-g/G1V2_GENERATION_SEMANTICS_V1.json":"45d533f8b83b24c77a96836fa5c2ef95f9b948003bd2ed725fc2ea97e010b310",
 "research/phase3/task85c-g/G1V2_GENERATION_GOLDEN_SUITE_V1.json":"143954667073a2c10f1bd59ce98b9c93dd84b50632bb67ea80d0d92449480acb",
 "research/phase3/task85c-e/G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json":"dbfb9a4a7101eed7006f751b9c4631b5f0286c3792f9777cc833c5dcfa42a3d3",
 "research/phase3/task85c-c/registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json":"fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9"}
for name,digest in expected.items(): assert hashlib.sha256((REPO/name).read_bytes()).hexdigest()==digest,name
command=(REPO/"cmd/g1v2-executor/main.go").read_text()
types=(REPO/"internal/g1v2/types.go").read_text()
assert "contains no fitting, generation, metric, or gate logic" in command
assert "deliberately contains no model fitting" in types
assert '"STRUCTURAL": true, "AGGREGATION": true' in types
with (OUT/"TASK86C_V2_IMPLEMENTATION_TRACEABILITY.tsv").open(newline="") as f: trace=list(csv.DictReader(f,delimiter="\t"))
assert any(r["validation_status"]=="MISSING_HANDLER" for r in trace)
run=json.loads((OUT/"TASK86C_V2_RUN_MANIFEST.json").read_text())
assert run["production_run_authorized"] is False and sum(run["materialization"].values())==0
markers=[p.name for p in OUT.iterdir() if p.is_file() and (p.name.startswith("TASK86C_V2_") and ("NOT_READY" in p.name or "DEFECT" in p.name or "FAILURE" in p.name or "INVALID" in p.name or p.name.endswith("RUN_FROZEN")))]
assert markers==["TASK86C_V2_IMPLEMENTATION_NOT_READY"],markers
manifest=json.loads((OUT/"TASK86C_V2_PRODUCTION_FREEZE_RESULTS_MANIFEST.json").read_text())
root_lines=[]
for item in manifest["artifacts_excluding_manifest"]:
    path=OUT/item["path"]
    digest=hashlib.sha256(path.read_bytes()).hexdigest()
    assert digest==item["sha256"],item["path"]
    root_lines.append((path.relative_to(REPO).as_posix(),f"{digest}  {path.relative_to(REPO).as_posix()}\n"))
assert hashlib.sha256("".join(line for _,line in sorted(root_lines)).encode()).hexdigest()==manifest["artifact_root_excluding_manifest_sha256"]
print("V1_2_AUTHORITY=SUPPORTED")
print("V1_2_GENERATION_CLOSURE=SUPPORTED")
print("SCIENTIFIC_IMPLEMENTATION_COVERAGE=INCOMPLETE")
print("PRODUCTION_RUN_AUTHORIZED=NO")
