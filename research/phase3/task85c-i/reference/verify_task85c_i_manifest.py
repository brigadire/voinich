#!/usr/bin/env python3
"""Verify Task85c-i artifact hashes/root and exact terminal-marker policy."""
import hashlib
import json
from pathlib import Path

OUT=Path(__file__).resolve().parent.parent
manifest=json.loads((OUT/"TASK85C_I_RESULTS_MANIFEST.json").read_text())
lines=[]
for item in manifest["artifacts_excluding_manifest"]:
    p=OUT/item["path"]
    assert p.is_file(),item["path"]
    digest=hashlib.sha256(p.read_bytes()).hexdigest()
    assert digest==item["sha256"],item["path"]
    assert p.stat().st_size==item["bytes"]
    lines.append(f"{digest}  {item['path']}\n")
assert hashlib.sha256("".join(lines).encode()).hexdigest()==manifest["artifact_root_excluding_manifest_sha256"]
terminal="G1V2_V1_2_EVIDENCE_EXECUTION_INTEGRATION_I1_FROZEN"
known={terminal,"TASK85C_I_PARENT_DEFECT_NOT_REPRODUCED","TASK85C_I_SCIENTIFIC_SCOPE_EXPANDED","TASK85C_I_AUTHORITY_INTEGRATION_DEFECT_REMAINS","TASK85C_I_SCHEMA_VALIDATION_FAILED","TASK85C_I_EXECUTION_IDENTITY_VALIDATION_FAILED","TASK85C_I_CROSS_ARTIFACT_VALIDATION_FAILED","TASK85C_I_UNEXPECTED_SCIENTIFIC_CHANGE","TASK85C_I_SCIENTIFIC_FIREWALL_VIOLATION"}
markers=[p.name for p in OUT.iterdir() if p.is_file() and p.name in known]
assert markers==[terminal],markers
assert manifest["terminal_marker"]==terminal and manifest["task85c_h_retry_ready"]=="SUPPORTED"
assert manifest["production_materialization"]=={"escrow_key_created":False,"blind_controls_created":0,"natural_controls_created":0,"jobids_created":0,"dag_created":False}
print("TASK85C_I_RESULTS_MANIFEST=PASS")
print(f"TASK85C_I_ARTIFACT_ROOT_SHA256={manifest['artifact_root_excluding_manifest_sha256']}")
