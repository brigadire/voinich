#!/usr/bin/env python3
import hashlib,json,unicodedata
from pathlib import Path
R=Path(__file__).resolve().parents[1];P=R.parent;m=json.loads((R/"TASK85C_C_RESULTS_MANIFEST.json").read_text())
a={p.relative_to(R).as_posix():p for p in R.rglob("*") if p.is_file() and p.name!="TASK85C_C_RESULTS_MANIFEST.json" and "__pycache__" not in p.parts};l={x["path"]:x for x in m["artifacts"]}
assert set(a)==set(l)
for k,p in a.items():
 assert hashlib.sha256(p.read_bytes()).hexdigest()==l[k]["sha256"],k
 assert p.stat().st_size==l[k]["bytes"],k
def canon(x):
 def n(v):
  if isinstance(v,str):return unicodedata.normalize("NFC",v)
  if isinstance(v,list):return [n(y) for y in v]
  if isinstance(v,dict):return {unicodedata.normalize("NFC",k):n(v[k]) for k in sorted(v)}
  return v
 return (json.dumps(n(x),ensure_ascii=False,sort_keys=True,separators=(",",":"))+"\n").encode()
def root(prefix):
 return hashlib.sha256(canon([{"path":p.relative_to(R).as_posix(),"sha256":hashlib.sha256(p.read_bytes()).hexdigest()} for p in sorted((R/prefix).rglob("*")) if p.is_file()])).hexdigest()
assert hashlib.sha256(canon(m["artifacts"])).hexdigest()==m["artifact_root_sha256"]
assert root("schemas")==m["evidence_schema_root_v1_1_sha256"]
assert root("golden")==m["golden_suite_v1_1_root_sha256"]
assert hashlib.sha256((R/"G1V2_EXECUTABLE_CONTRACT_V1_1.md").read_bytes()).hexdigest()==m["executable_contract_v1_1_sha256"]
assert hashlib.sha256((R/"registries/G1V2_STATUS_REGISTRY_V2.tsv").read_bytes()).hexdigest()==m["status_registry_v1_1_sha256"]
assert hashlib.sha256((P/"task85c/G1V2_EXECUTABLE_CONTRACT.md").read_bytes()).hexdigest()==m["parent_v1_sha256"]
assert json.loads((P/"task85c/TASK85C_RESULTS_MANIFEST.json").read_text())["artifact_root_sha256"]==m["parent_task85c_root"]
assert json.loads((P/"task85c-a/TASK85C_A_RESULTS_MANIFEST.json").read_text())["artifact_root_sha256"]==m["task85c_a_root"]
assert hashlib.sha256((P/"task85c-b/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json").read_bytes()).hexdigest()==m["status_reachability_v2_sha256"]
assert json.loads((P/"task85c-b/TASK85C_B_RESULTS_MANIFEST.json").read_text())["artifact_root_sha256"]==m["task85c_b_root"]
assert hashlib.sha256((P/"task86c-v2-scientific-impl-r/TASK86C_V2_TASK85C_CONTRACT_DEFECT.md").read_bytes()).hexdigest()==m["scientific_impl_r_defect_sha256"]
assert m["task86c_v2_scientific_impl_r2_ready"]=="SUPPORTED"
print("TASK85C_C_RESULTS_MANIFEST=PASS")
