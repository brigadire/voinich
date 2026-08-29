#!/usr/bin/env python3
"""Deterministic V1.2 evidence-only verifier and standard-validator launcher."""
import hashlib
import json
import subprocess
import sys
import unicodedata
from pathlib import Path

HERE = Path(__file__).resolve().parent
OUT = HERE.parent
V12 = "G1_V2_EXECUTABLE_CONTRACT_V1_2"

def norm(x):
    if isinstance(x,str): return unicodedata.normalize("NFC",x)
    if isinstance(x,list): return [norm(v) for v in x]
    if isinstance(x,dict): return {k:norm(x[k]) for k in sorted(x)}
    return x

def canon(x): return (json.dumps(norm(x),ensure_ascii=False,sort_keys=True,separators=(",",":"))+"\n").encode()

i1=json.loads((OUT/"G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.json").read_text())
registry=json.loads((OUT/"G1V2_V1_2_EVIDENCE_SCHEMA_REGISTRY.json").read_text())
root=json.loads((OUT/"G1V2_V1_2_EVIDENCE_SCHEMA_ROOT.json").read_text())
assert i1["evidence"]["scientific_contract_version"]==registry["scientific_contract_version"]==V12
assert i1["evidence"]["schema_root_sha256"]==root["root_sha256"]
assert registry["selection"].startswith("select exactly") and registry["unknown_version_disposition"]=="FAIL_CLOSED"
assert len(registry["entries"])==15 and len({x["evidence_type"] for x in registry["entries"]})==15
items=[]
for entry in registry["entries"]:
    digest=hashlib.sha256((OUT/entry["schema_path"]).read_bytes()).hexdigest()
    assert digest==entry["schema_sha256"]
    items.append({"path":entry["schema_path"],"sha256":digest})
assert hashlib.sha256(canon(items)).hexdigest()==root["root_sha256"]
def select_family(version, supplied_root):
    if version!=V12 or supplied_root!=root["root_sha256"]: raise ValueError("FAIL_CLOSED")
    return registry
assert select_family(V12,root["root_sha256"]) is registry
for version,supplied_root in [("G1_V2_EXECUTABLE_CONTRACT_UNKNOWN",root["root_sha256"]),(V12,"0"*64)]:
    try: select_family(version,supplied_root)
    except ValueError as exc: assert str(exc)=="FAIL_CLOSED"
    else: raise AssertionError("invalid evidence authority selected")
fixtures=json.loads((OUT/"fixtures/G1V2_V1_2_EVIDENCE_POSITIVE_FIXTURES.json").read_text())
for tc in fixtures:
    x=tc["instance"]
    assert x["contract_version"]==V12
    bare=dict(x); expected=bare.pop("content_sha256")
    assert hashlib.sha256(canon(bare)).hexdigest()==expected,tc["id"]
subprocess.run(["node",str(HERE/"verify_evidence_schemas_v1_2.mjs")],check=True)
print("EVIDENCE_ONLY_VERIFIER_V1_2=PASS")
print("UNKNOWN_VERSION=FAIL_CLOSED")
print("WRONG_SCHEMA_ROOT=FAIL_CLOSED")
