#!/usr/bin/env python3
import csv, hashlib, json
from pathlib import Path
HERE=Path(__file__).resolve().parent; OUT=HERE.parent; REPO=HERE.parents[3]
repro=json.loads((OUT/"TASK85C_H_AUTHORITY_CONFLICT_REPRODUCTION.json").read_text())
assert repro["finding_ids"]==["H-SC01-EVIDENCE-CONTRACT-VERSION","H-SC02-E1-JOBID-SCIENTIFIC-VERSION"]
with (OUT/"TASK85C_H_VALIDATION.tsv").open(newline="") as f: values={r["check"]:r["verdict"] for r in csv.DictReader(f,delimiter="\t")}
assert values["EVIDENCE_VALIDATION"]=="FAIL" and values["JOBID_V1_2_E1"]=="FAIL"
assert values["PRODUCTION_FREEZE_RETRY_READY"]=="NOT_SUPPORTED"
manifest=json.loads((OUT/"TASK85C_H_RESULTS_MANIFEST.json").read_text()); lines=[]
for item in manifest["artifacts_excluding_manifest"]:
    path=OUT/item["path"]; digest=hashlib.sha256(path.read_bytes()).hexdigest(); assert digest==item["sha256"],item["path"]
    lines.append((item["path"],f"{digest}  {item['path']}\n"))
assert hashlib.sha256("".join(line for _,line in sorted(lines)).encode()).hexdigest()==manifest["artifact_root_excluding_manifest_sha256"]
markers=[p.name for p in OUT.iterdir() if p.is_file() and (p.name.startswith("TASK85C_H_") and ("READY" in p.name or "FAILED" in p.name or "DEFECT" in p.name or "VIOLATION" in p.name)) and not p.name.endswith(".md")]
assert markers==["TASK85C_H_SCIENTIFIC_CONTRACT_DEFECT"],markers
assert manifest["production_materialization"]=={"blind_controls":0,"dag_created":False,"escrow_key_created":False,"jobids":0,"natural_controls":0,"thresholds_frozen":False}
print("TASK85C_H_FAILURE_PACKAGE=PASS")
print("SCIENTIFIC_FIREWALL=INTACT")
