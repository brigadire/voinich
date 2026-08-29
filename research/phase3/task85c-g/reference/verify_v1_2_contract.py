#!/usr/bin/env python3
import csv,hashlib,json
from pathlib import Path
HERE=Path(__file__).resolve().parent; OUT=HERE.parent; REPO=HERE.parents[3]
assert hashlib.sha256((REPO/"research/phase3/task85c-c/G1V2_EXECUTABLE_CONTRACT_V1_1.md").read_bytes()).hexdigest()=="5c3cd272c1dbae9bfe1d7a100155faf102e86d34660da239e1cb31704ad470b0"
contract=json.loads((OUT/"G1_V2_EXECUTABLE_CONTRACT_V1_2.json").read_text()); sem=json.loads((OUT/"G1V2_GENERATION_SEMANTICS_V1.json").read_text())
assert contract["contract_version"]=="G1_V2_EXECUTABLE_CONTRACT_V1_2" and sem["version"]=="G1V2_GENERATION_SEMANTICS_V1"
assert contract["generation"]["semantics_sha256"]==hashlib.sha256((OUT/"G1V2_GENERATION_SEMANTICS_V1.json").read_bytes()).hexdigest()
with (OUT/"G1V2_GENERATION_PATH_INVENTORY.tsv").open(newline="") as f: paths=list(csv.DictReader(f,delimiter="\t"))
with (OUT/"G1V2_GENERATION_AMBIGUITY_REGISTRY.tsv").open(newline="") as f: ambiguities=list(csv.DictReader(f,delimiter="\t"))
assert len(paths)==26 and all(x["status"]=="RESOLVED" for x in paths)
assert len(ambiguities)==14 and all(x["status"]=="RESOLVED" for x in ambiguities)
manifest=json.loads((OUT/"TASK85C_G_RESULTS_MANIFEST.json").read_text()); lines=[]
for item in manifest["artifacts_excluding_manifest"]:
    p=OUT/item["path"]; digest=hashlib.sha256(p.read_bytes()).hexdigest(); assert digest==item["sha256"],item["path"]
    lines.append(f'{digest}  {item["path"]}\n')
assert hashlib.sha256("".join(lines).encode()).hexdigest()==manifest["artifact_root_excluding_manifest_sha256"]
assert manifest["generation_ambiguities_open"]==0 and manifest["v1_2_ready"]=="SUPPORTED"
markers=[p for p in OUT.iterdir() if p.is_file() and ("FROZEN" in p.name or p.name.startswith("TASK85C_G_") and ("FAILED" in p.name or "DEFECT" in p.name or "VIOLATION" in p.name))]
assert [p.name for p in markers]==["G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_V1_2_FROZEN"]
print("V1_2_CONTRACT_VALIDATION=PASS")
print("SECOND_IMPLEMENTER_UNSTATED_CHOICES=0")

