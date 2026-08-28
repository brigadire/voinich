#!/usr/bin/env python3
"""Check enum, root, parent and V1→V1.1 cross-artifact consistency."""
import csv,hashlib,json,sys
from pathlib import Path
R=Path(__file__).resolve().parents[1];P=R.parent;V1=P/"task85c";A=P/"task85c-a";B=P/"task85c-b"
def rows(p):
 with p.open() as f:return list(csv.DictReader(f,delimiter="\t"))
def req(x,m):
 if not x:raise AssertionError(m)
def root(prefix):
 items=[{"path":p.relative_to(R).as_posix(),"sha256":hashlib.sha256(p.read_bytes()).hexdigest()} for p in sorted((R/prefix).rglob("*")) if p.is_file()]
 data=(json.dumps(items,ensure_ascii=False,sort_keys=True,separators=(",",":"))+"\n").encode()
 return hashlib.sha256(data).hexdigest()
c=json.loads((R/"G1V2_EXECUTABLE_CONTRACT_V1_1.json").read_text());old=json.loads((V1/"G1V2_EXECUTABLE_CONTRACT.json").read_text());sm=json.loads((B/"G1V2_STATUS_REACHABILITY_CONTRACT_V2.json").read_text())
req(hashlib.sha256((V1/"G1V2_EXECUTABLE_CONTRACT.md").read_bytes()).hexdigest()=="275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca","V1 identity")
req(json.loads((V1/"TASK85C_RESULTS_MANIFEST.json").read_text())["artifact_root_sha256"]=="273913473e3e37d6a776c79b0eb214753a90e9dbaf5d78e186dcb65a0c32c351","Task85c root")
req(json.loads((A/"TASK85C_A_RESULTS_MANIFEST.json").read_text())["artifact_root_sha256"]=="736e2aee714f63145d4dece5a1cb0014e070142f5033f97edcbb86ab4a57f2f8","Task85c-a root")
req(hashlib.sha256((B/"G1V2_STATUS_REACHABILITY_CONTRACT_V2.json").read_bytes()).hexdigest()=="fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9","V2 identity")
req(json.loads((B/"TASK85C_B_RESULTS_MANIFEST.json").read_text())["artifact_root_sha256"]=="387d9af30d74e95af2ddc0e02eed2cb20ab2c72ca99c6bb6d3a218fa375bd57d","Task85c-b root")
req(c["contract_version"]=="G1_V2_EXECUTABLE_CONTRACT_V1_1","version")
for k in ["numeric","data","selection","rng","candidate_count","models","metrics","thresholds","generation","dag","canonicalization","firewall","schemas"]:req(c[k]==old[k],f"drift {k}")
req(len(sm["statuses"])==13 and len(sm["stages"])==7 and len(sm["transitions"])==45,"V2 counts")
matrix=rows(R/"G1V2_EVIDENCE_STATUS_MATRIX_V1_1.tsv");req(len(matrix)==195,"matrix")
reg=rows(R/"G1V2_EVIDENCE_SCHEMA_REGISTRY_V1_1.tsv");req(len(reg)==15,"schemas")
statuses={x["status"] for x in sm["statuses"]};req(statuses=={x["status"] for x in matrix},"status agreement")
req(all(json.loads((R/x["schema_path"]).read_text())["$id"]==x["schema_id"] for x in reg),"schema IDs")
reason=rows(R/"G1V2_REASON_REGISTRY_V1_1.tsv");req(any(x["status"]=="NOT_ASSESSABLE" for x in reason),"NOT_ASSESSABLE reasons")
for name in ["G1V2_CANDIDATE_REGISTRY.tsv","G1V2_RNG_DOMAIN_REGISTRY.tsv","G1V2_DAG_CONTRACT.json","G1V2_STATUS_REACHABILITY_CONTRACT_V2.json","G1V2_REACHABILITY_EXPANDED.tsv"]:
 src=(V1/name) if (V1/name).exists() else (B/name)
 req(hashlib.sha256((R/"registries"/name).read_bytes()).hexdigest()==hashlib.sha256(src.read_bytes()).hexdigest(),f"integrated copy {name}")
req(hashlib.sha256((V1/"G1V2_CANDIDATE_REGISTRY.tsv").read_bytes()).hexdigest()=="96b618ab324db77b8402081075241b275f8925c45cdab058ba741e4beed29b58","candidate")
req(hashlib.sha256((V1/"G1V2_RNG_DOMAIN_REGISTRY.tsv").read_bytes()).hexdigest()=="e47c5a8c62dd8dee34441e4274c0d49d1a7ce4aab0360aff4751cb08b6394a43","rng")
req(hashlib.sha256((B/"G1V2_REACHABILITY_EXPANDED.tsv").read_bytes()).hexdigest()=="3cc63c59bcce2269e5df94ba628a5f6bd169199e871a30f65e2d289fb5198b5c","reach")
req(c["dag"]["total_jobs"]==1321152 and c["dag"]["dependency_edges"]==2617152,"DAG")
req(c["evidence_contract"]["schema_root_v1_1_sha256"]==root("schemas"),"schema root")
req(c["evidence_contract"]["golden_suite_v1_1_root_sha256"]==root("golden"),"golden root")
print("TASK85C_C_CROSS_ARTIFACT=PASS CONTRADICTIONS=0")
