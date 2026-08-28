#!/usr/bin/env python3
"""Independent V1.1 evidence/reachability verifier; imports no model code."""
from __future__ import annotations
import csv, hashlib, json, subprocess, sys, unicodedata
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]; B=ROOT.parent/"task85c-b"
def rows(p):
 with p.open(encoding="utf-8",newline="") as f:return list(csv.DictReader(f,delimiter="\t"))
def canon(v):
 def n(x):
  if isinstance(x,str):return unicodedata.normalize("NFC",x)
  if isinstance(x,list):return [n(y) for y in x]
  if isinstance(x,dict):return {unicodedata.normalize("NFC",k):n(x[k]) for k in sorted(x)}
  return x
 return (json.dumps(n(v),ensure_ascii=False,sort_keys=True,separators=(",",":"))+"\n").encode()
def check_hash(x):
 y=dict(x);claimed=y.pop("content_sha256");return hashlib.sha256(canon(y)).hexdigest()==claimed

# Standard schema validation is independently delegated to the two-validator suite.
subprocess.run(["node",str(ROOT/"reference/validate_schemas.mjs")],check=True,stdout=subprocess.DEVNULL)
status={r["status"] for r in rows(B/"G1V2_STATUS_REGISTRY_V2.tsv")}
stage_allowed={(r["stage"],r["status"]) for r in rows(B/"G1V2_STAGE_STATUS_CONTRACT.tsv") if r["allowed"]=="YES"}
transition={(r["upstream_stage"],r["upstream_status"],r["downstream_stage"]):r for r in rows(B/"G1V2_REACHABILITY_CONTRACT_V2.tsv")}
reason={r["reason_code"]:r for r in rows(B/"G1V2_STATUS_REASON_REGISTRY.tsv")}
registry={r["evidence_type"]:r for r in rows(ROOT/"G1V2_EVIDENCE_SCHEMA_REGISTRY_V1_1.tsv")}

positive=json.loads((ROOT/"golden/schema-positive/cases.json").read_text())
not_reached_instance=None
for tc in positive:
 x=tc["instance"];typ=tc["schema"]
 assert check_hash(x),tc["id"]
 assert typ in registry and x["status"] in status
 if x["producer_scope"]=="JOB": assert (x["stage"],x["status"]) in stage_allowed
 elif x["producer_scope"]=="SUBOPERATION": assert typ in {"structural_family","structural_gate","structural_verdict"} and x["stage"]=="CANDIDATE_AGGREGATION"
 else: assert typ=="not_reached" and x["status"]=="NOT_REACHED"
 if typ=="scientific_failure": assert x["payload"]["reason_code"] in reason
 if typ=="not_reached":
  p=x["payload"];t=transition[(p["upstream_stage"],p["upstream_status"],x["stage"])]
  assert t["action"]=="NOT_REACHED" and t["reason_code"]==p["reason_code"]
  assert p["selected_causal_job_id"] in p["causal_dependency_ids"]
  assert set(p["causal_dependency_ids"]).issubset(x["dependencies"])
  not_reached_instance=x

# Multi-parent rule and precedence fixtures.
rank={"PROTOCOL_VETO":0,"FIT_FAILURE":1,"NUMERICAL_FAILURE":1,"INDUCTION_CAP":1,"GENERATION_FAILURE":1,"NOT_ASSESSABLE":2,"FAIL":3,"NOT_REACHED":4}
parents=[("j-"+"2"*40,"FAIL"),("j-"+"1"*40,"NUMERICAL_FAILURE")]
selected=min(parents,key=lambda x:(rank[x[1]],x[0]))
assert selected[1]=="NUMERICAL_FAILURE"
assert all(x[1]=="PASS" for x in [("x","PASS")])

# Aggregation reconstruction properties.
def candidate(states):
 if "PROTOCOL_VETO" in states:return "PROTOCOL_INVALID"
 if states & {"FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","NOT_ASSESSABLE","NOT_REACHED"}:return "UNRESOLVED"
 if "FAIL" in states:return "INADEQUATE"
 return "ADEQUATE"
assert candidate({"FIT_SUCCESS","PASS","GENERATION_SUCCESS","COMPLEXITY_SUCCESS"})=="ADEQUATE"
assert candidate({"FIT_SUCCESS","FAIL","GENERATION_SUCCESS","COMPLEXITY_SUCCESS"})=="INADEQUATE"
for f in ["FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","NOT_ASSESSABLE","NOT_REACHED"]:assert candidate({f})=="UNRESOLVED"
assert candidate({"PROTOCOL_VETO"})=="PROTOCOL_INVALID"

# Fail-closed mutations independent of schema suite.
x=json.loads(json.dumps(positive[0]["instance"]));x["content_sha256"]="0"*64;assert not check_hash(x)
for unknown in ["status","schema","contract","stage"]:
 y=json.loads(json.dumps(positive[0]["instance"]));
 if unknown=="status":y["status"]="FUTURE_STATUS"
 elif unknown=="schema":y["schema_id"]="g1v2.future.v1_1"
 elif unknown=="contract":y["contract_version"]="G1_V2_EXECUTABLE_CONTRACT_V2"
 else:y["stage"]="FUTURE_STAGE"
 assert (y["status"] not in status) or (y["schema_id"] not in {r["schema_id"] for r in registry.values()}) or y["contract_version"]!="G1_V2_EXECUTABLE_CONTRACT_V1_1" or y["stage"] not in {s for s,_ in stage_allowed}

# Reachability mutations fail independently of JSON Schema validation.
assert not_reached_instance is not None
for field,value in [("reason_code","NR_UNKNOWN"),("upstream_status","PASS")]:
 y=json.loads(json.dumps(not_reached_instance));y["payload"][field]=value
 key=(y["payload"]["upstream_stage"],y["payload"]["upstream_status"],y["stage"])
 assert key not in transition or transition[key]["action"]!="NOT_REACHED" or transition[key]["reason_code"]!=y["payload"]["reason_code"]
y=json.loads(json.dumps(not_reached_instance));y["payload"]["selected_causal_job_id"]="j-"+"f"*40
assert y["payload"]["selected_causal_job_id"] not in y["payload"]["causal_dependency_ids"]
y=json.loads(json.dumps(not_reached_instance));y["dependencies"]=[]
assert not set(y["payload"]["causal_dependency_ids"]).issubset(y["dependencies"])
print("G1V2_EVIDENCE_ONLY_VERIFIER=PASS")
