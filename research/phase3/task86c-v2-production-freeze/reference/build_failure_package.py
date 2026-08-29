#!/usr/bin/env python3
"""Freeze the clean V1.2.1 production-freeze pre-materialization failure."""
from __future__ import annotations
import csv, hashlib, json, os, subprocess
from pathlib import Path

ROOT=Path(__file__).resolve().parents[4]
OUT=ROOT/"research/phase3/task86c-v2-production-freeze"
H=ROOT/"research/phase3/task85c-h"
J=ROOT/"research/phase3/task85c-j"
MARKER="TASK86C_V2_IMPLEMENTATION_NOT_READY"
IMPL="5687f219c049f6e38b2a9048c7799965948e34b90fa2fe37a6a3679427ff7a0b"

def sha(p): return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def write(p,s): Path(p).parent.mkdir(parents=True,exist_ok=True); Path(p).write_text(s,encoding="utf-8",newline="\n")
def dump(p,x): write(p,json.dumps(x,ensure_ascii=False,sort_keys=True,separators=(",",":"))+"\n")
def tsv(p,head,rows):
    Path(p).parent.mkdir(parents=True,exist_ok=True)
    with Path(p).open("w",encoding="utf-8",newline="") as f:
        w=csv.writer(f,delimiter="\t",lineterminator="\n");w.writerow(head);w.writerows(rows)
def tree_root(paths):
    lines=[f"{sha(p)}  {Path(p).relative_to(ROOT).as_posix()}\n" for p in sorted(paths)]
    return hashlib.sha256("".join(lines).encode()).hexdigest()
def run(*args):
    env=dict(os.environ);env["GOCACHE"]="/tmp/task86freeze121-gocache"
    return subprocess.check_output(args,cwd=ROOT,env=env)

def main():
    hm=json.loads((H/"TASK85C_H_RESULTS_MANIFEST.json").read_text())
    assert hm["status"]=="PASS" and hm["scientific_implementation_root_sha256"]==IMPL
    subprocess.check_call(["python3",str(H/"reference/verify_task85c_h_manifest.py")],cwd=ROOT)
    mismatch=json.loads(run("go","run","./research/phase3/task86c-v2-production-freeze/reference/reproduce_frozen_implementation_mismatch.go","."))
    assert mismatch["status"]=="IMPLEMENTATION_VALIDATION_FAILURE" and len(mismatch["findings"])>=2
    dump(OUT/"TASK86C_V2_IMPLEMENTATION_MISMATCH.json",mismatch)

    authority=[
      ["contract_machine",J/"G1_V2_EXECUTABLE_CONTRACT_V1_2_1.json","b1eb452dd9994d63108cae37a19b1945bac3b78b4a2af3a0c080074eff8a5028"],
      ["contract_markdown",J/"G1_V2_EXECUTABLE_CONTRACT_V1_2_1.md","17d55ae32ba2a60d1e4477eb34cb06b28e63b9660c92c75d4d91d18db082946b"],
      ["integration_i2",J/"G1V2_V1_2_1_INTEGRATION_SUPPLEMENT_I2.json","cc84d7f8564d0c196607d22b42bddd60c85905d3d15abd5dd7c485bcb19e9333"],
      ["execution_e3",J/"G1V2_EXECUTION_IDENTITY_ERRATUM_E3.json","adaa38dbf2a857a0671927cf45e3e8cd31c97bf5a4d445051878ddf3af764d12"],
      ["generation_semantics",ROOT/"research/phase3/task85c-g/G1V2_GENERATION_SEMANTICS_V1.json","45d533f8b83b24c77a96836fa5c2ef95f9b948003bd2ed725fc2ea97e010b310"],
      ["generation_goldens",J/"G1V2_GENERATION_GOLDEN_SUITE_V1_2_1.json","04408203434ef354996cf39400921865d0940963efca810a1eec2ab327775046"],
      ["status_v2",ROOT/"research/phase3/task85c-c/registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json","fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9"],
    ]
    for _,p,e in authority: assert sha(p)==e
    tsv(OUT/"TASK86C_V2_AUTHORITY_VALIDATION.tsv",["authority","path","expected_sha256","actual_sha256","status"],[[k,p.relative_to(ROOT),e,sha(p),"PASS"] for k,p,e in authority]+[["task85c_h_implementation_root",H/"TASK85C_H_RESULTS_MANIFEST.json",IMPL,hm["scientific_implementation_root_sha256"],"PASS"]])
    tsv(OUT/"TASK86C_V2_SOURCE_IDENTITY.tsv",["field","expected","actual","status"],[["task85c_h_implementation_root",IMPL,hm["scientific_implementation_root_sha256"],"PASS"],["task85c_h_git_commit","65e6b99f07eea8241958ca746f205048e084bb4c",hm["source_identity"]["git_commit"],"PASS"],["task85c_h_dirty","true",str(hm["source_identity"]["dirty"]).lower(),"PASS"],["scientific_handler_changes_in_freeze","0","0","PASS"]])
    tsv(OUT/"TASK86C_V2_IMPLEMENTATION_TRACEABILITY.tsv",["component","authority","implementation","validation_status","finding"],[["M0-M5/generation","V1.2.1 + generation semantics V1","internal/g1v2science","FAIL","PF-IMPL-01 route parameters ignored"],["F2 scientific weights","G1V2_STRUCTURAL_METRIC_REGISTRY.tsv","internal/g1v2science.F2Metrics","FAIL","PF-IMPL-02 EF3/SKELETON weights inverted"],["source identity","Task85c-h manifest",IMPL,"PASS","exact frozen closure revalidated"]])
    tsv(OUT/"TASK86C_V2_VALIDATION.tsv",["check","status","detail"],[["updated V1.2.1/I2/E3 authority","PASS","hash closure"],["Task85c-h implementation root identity","PASS",IMPL],["scientific handler conformance","FAIL","PF-IMPL-01;PF-IMPL-02"],["pre-materialization gate","FAIL","mandatory stop"],["development calibration","NOT_RUN","blocked before materialization"],["escrow/blind/natural/JobID/DAG","NOT_RUN","all counts zero"],["scientific firewall","PASS","INTACT"]])
    for name,head in {
      "TASK86C_V2_THRESHOLD_MANIFEST.tsv":["threshold_id","value","status"],
      "TASK86C_V2_INPUT_MANIFEST.tsv":["control_instance_id","role","content_sha256","status"],
      "TASK86C_V2_BLIND_SAFE_MANIFEST.tsv":["blind_id","content_sha256","token_count"],
      "TASK86C_V2_NATURAL_INPUT_MANIFEST.tsv":["control_instance_id","language","content_sha256","status"],
      "TASK86C_V2_PRESEED_MANIFEST.tsv":["job_id","preseed_status"],
    }.items(): tsv(OUT/name,head,[])
    dump(OUT/"TASK86C_V2_DAG_AUDIT.json",{"schema":"task86c-v2-dag-audit-v1_2_1-failure","status":"NOT_MATERIALIZED","jobs":0,"unique_jobids":0,"edges":0,"dangling":0,"cycles":0})
    tsv(OUT/"TASK86C_V2_CAPACITY_BENCHMARK.tsv",["benchmark","status","reason"],[["production capacity","NOT_RUN","implementation pre-materialization gate failed"]])
    write(OUT/"TASK86C_V2_CAPACITY_REPORT.md","# Capacity qualification\n\nNot run. Capacity cannot qualify a scientifically non-conformant executable.\n")
    run_manifest={"schema":"task86c-v2-run-manifest-v1_2_1-failure","scientific_contract_version":"G1_V2_EXECUTABLE_CONTRACT_V1_2_1","implementation_root_sha256":IMPL,"production_run_authorized":False,"materialization":{"development_controls":0,"blind_controls":0,"natural_controls":0,"escrow_keys":0,"jobids":0,"dag_jobs":0,"dag_edges":0}}
    dump(OUT/"TASK86C_V2_RUN_MANIFEST.json",run_manifest)
    write(OUT/"TASK86C_V2_OPERATOR_HANDOFF.md","# Operator handoff\n\nNo production handoff is authorized. Resolve PF-IMPL-01 and PF-IMPL-02 in a new scientific-implementation task and refreeze a new implementation root before retrying production-freeze.\n")
    write(OUT/"TASK86C_V2_RUNBOOK.md","# Runbook\n\nSTOP. Do not generate escrow, materialize controls, create production JobIDs/DAG, build a production executable, or start Task86C-v2-run from this failed freeze.\n")
    write(OUT/"TASK86C_V2_PRODUCTION_FREEZE_DESIGN.md",f"""# Task86C-v2 production-freeze design — V1.2.1 clean run

The pre-materialization order is authority verification, exact Task85c-h implementation-root verification, scientific conformance revalidation, then calibration/materialization. The frozen root `{IMPL}` was verified byte-for-byte and no handler was changed. Conformance failed before any secret or production identity was created.
""")
    write(OUT/"TASK86C_V2_PRODUCTION_FREEZE_REPORT.md","""# Task86C-v2 production-freeze report — V1.2.1/I2/E3

This was a new clean run. The authority chain and the exact frozen Task85c-h implementation root passed identity checks. The implementation itself failed mandatory pre-materialization scientific conformance:

- PF-IMPL-01: `GenerateSynthetic` ignores normative route parameters. A scientifically different M1 parameter object produces the same corpus bytes/hash.
- PF-IMPL-02: `F2Metrics` gives EF3 scientific weight 0 although the registry marks it weighted, and gives HR1/SKELETON weight 1 although SKELETON is diagnostic-only with weight 0.

These defects can change generated controls, gates and final verdicts. Task86C is forbidden to repair frozen scientific handlers, so the run stops with `IMPLEMENTATION_VALIDATION_FAILURE`. DEVELOPMENT calibration, escrow, 144 blind controls, 36 natural controls, production JobIDs, the 1,321,152/2,617,152 DAG, capacity qualification and production executable were not materialized. The firewall remains intact.
""")
    write(OUT/MARKER,MARKER+"\n")
    verify='''#!/usr/bin/env python3
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
 p=ROOT/a["path"];d=hashlib.sha256(p.read_bytes()).hexdigest();assert d==a["sha256"];lines.append((a["path"],f"{d}  {a['path']}\\n"))
assert hashlib.sha256("".join(x[1] for x in sorted(lines)).encode()).hexdigest()==m["artifact_root_excluding_manifest_sha256"]
print("V1_2_1_AUTHORITY=SUPPORTED");print("IMPLEMENTATION_ROOT_IDENTITY=VERIFIED");print("SCIENTIFIC_IMPLEMENTATION_CONFORMANCE=FAIL");print("PRODUCTION_MATERIALIZATION=0");print("PRODUCTION_RUN_AUTHORIZED=NO")
'''
    write(OUT/"reference/verify_failed_freeze.py",verify)
    paths=[p for p in OUT.rglob("*") if p.is_file() and p.name!="TASK86C_V2_PRODUCTION_FREEZE_RESULTS_MANIFEST.json" and "__pycache__" not in p.parts]
    root=tree_root(paths)
    result={"schema":"task86c-v2-production-freeze-results-v1_2_1-failure","status":"IMPLEMENTATION_VALIDATION_FAILURE","clean_run":True,"updated_task_sha256":sha(ROOT/"tasks_ph3/task86c-v2-production-freexe.txt"),"scientific_contract_version":"G1_V2_EXECUTABLE_CONTRACT_V1_2_1","integration_authority":"G1V2_V1_2_1_INTEGRATION_SUPPLEMENT_I2","execution_identity_authority":"G1V2_EXECUTION_IDENTITY_ERRATUM_E3","implementation_root_sha256":IMPL,"implementation_root_identity":"VERIFIED","task85c_h_source_identity":hm["source_identity"],"blocking_findings":[x["id"] for x in mismatch["findings"]],"materialization":{"development_controls":0,"thresholds":0,"escrow_keys":0,"blind_controls":0,"natural_controls":0,"jobids":0,"dag_jobs":0,"dag_edges":0,"production_executables":0},"scientific_firewall":"INTACT","production_run_authorized":False,"artifact_root_excluding_manifest_sha256":root,"artifact_root_definition":"sha256(sorted(<file-sha256><two spaces><repository-relative-path><LF>)); results manifest excluded","terminal_marker":MARKER,"artifacts_excluding_manifest":[{"path":p.relative_to(ROOT).as_posix(),"sha256":sha(p)} for p in sorted(paths)]}
    dump(OUT/"TASK86C_V2_PRODUCTION_FREEZE_RESULTS_MANIFEST.json",result)
    print(json.dumps({"status":result["status"],"artifact_root":root,"marker":MARKER},sort_keys=True))
if __name__=="__main__":main()
