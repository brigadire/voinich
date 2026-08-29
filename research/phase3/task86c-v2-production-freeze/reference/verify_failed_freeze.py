#!/usr/bin/env python3
import csv,hashlib,json,os,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[4];OUT=ROOT/"research/phase3/task86c-v2-production-freeze"
m=json.loads((OUT/"TASK86C_V2_PRODUCTION_FREEZE_RESULTS_MANIFEST.json").read_text()); assert m["status"]=="IMPLEMENTATION_VALIDATION_FAILURE"
assert m["implementation_root_sha256"]=="5687f219c049f6e38b2a9048c7799965948e34b90fa2fe37a6a3679427ff7a0b"
x=json.loads((OUT/"TASK86C_V2_IMPLEMENTATION_MISMATCH.json").read_text()); assert len(x["findings"])>=2 and x["materialization_allowed"] is False
assert sum(m["materialization"].values())==0 and m["production_run_authorized"] is False
markers=[p.name for p in OUT.iterdir() if p.is_file() and (p.name.endswith("RUN_FROZEN") or "NOT_READY" in p.name or "DEFECT" in p.name or p.name.endswith("FREEZE_INVALID"))];assert markers==["TASK86C_V2_IMPLEMENTATION_NOT_READY"],markers
lines=[]
for a in m["artifacts_excluding_manifest"]:
 p=ROOT/a["path"];d=hashlib.sha256(p.read_bytes()).hexdigest();assert d==a["sha256"];lines.append((a["path"],f"{d}  {a['path']}\n"))
assert hashlib.sha256("".join(x[1] for x in sorted(lines)).encode()).hexdigest()==m["artifact_root_excluding_manifest_sha256"]
print("V1_2_1_AUTHORITY=SUPPORTED");print("IMPLEMENTATION_ROOT_IDENTITY=VERIFIED");print("SCIENTIFIC_IMPLEMENTATION_CONFORMANCE=FAIL");print("PRODUCTION_MATERIALIZATION=0");print("PRODUCTION_RUN_AUTHORIZED=NO")
