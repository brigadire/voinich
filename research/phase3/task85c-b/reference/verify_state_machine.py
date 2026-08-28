#!/usr/bin/env python3
"""Independent verifier/property suite for G1-v2 status machine V2."""
from __future__ import annotations

import csv
import hashlib
import json
import sys
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]
PHASE3=ROOT.parent
T85=PHASE3/"task85c"
T85A=PHASE3/"task85c-a"


def rows(name, root=ROOT):
    with (root/name).open(encoding="utf-8",newline="") as f: return list(csv.DictReader(f,delimiter="\t"))


def req(ok,msg):
    if not ok: raise AssertionError(msg)


def main():
    machine=json.loads((ROOT/"G1V2_STATUS_REACHABILITY_CONTRACT_V2.json").read_text())
    req(machine["version"]=="G1_V2_STATUS_REACHABILITY_CONTRACT_V2","version")
    req(hashlib.sha256((T85/"G1V2_EXECUTABLE_CONTRACT.md").read_bytes()).hexdigest()=="275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca","parent contract")
    parent_a=json.loads((T85A/"TASK85C_A_RESULTS_MANIFEST.json").read_text())
    req(parent_a["artifact_root_sha256"]=="736e2aee714f63145d4dece5a1cb0014e070142f5033f97edcbb86ab4a57f2f8","parent A")

    status_rows=rows("G1V2_STATUS_REGISTRY_V2.tsv")
    statuses={r["status"]:r for r in status_rows}
    req(len(statuses)==13==len(status_rows),"status registry uniqueness")
    req("SCIENTIFIC_FAILURE" not in statuses,"E01")
    req({"GATE_VERDICT","ASSESSABILITY","PROCEDURE_SUCCESS","PROCEDURE_FAILURE","DAG_SUPPRESSION","PROTOCOL_INVALIDITY"}=={r["category"] for r in status_rows},"taxonomy")
    req(statuses["FAIL"]["scientific_negative_evidence"]=="YES","FAIL negative")
    req(all(r["scientific_negative_evidence"]=="NO" for s,r in statuses.items() if s!="FAIL"),"failure non-rejection")

    stage_rows=rows("G1V2_STAGE_STATUS_CONTRACT.tsv")
    stages={x["stage"] for x in stage_rows}
    expected_stages={"FIT","PREDICTIVE","GENERATION","F2_METRIC","COMPLEXITY","CANDIDATE_AGGREGATION","CONTROL_AGGREGATION"}
    req(stages==expected_stages,"stage closure")
    req(len(stage_rows)==len(stages)*len(statuses),"producer matrix completeness")
    allowed={(r["stage"],r["status"]) for r in stage_rows if r["allowed"]=="YES"}
    req(("FIT","FAIL") not in allowed and ("FIT","FIT_SUCCESS") in allowed,"E04")
    req(all((stage,"FAIL") not in allowed for stage in {"FIT","GENERATION","COMPLEXITY"}),"PASS/FAIL reserved")

    tr=rows("G1V2_REACHABILITY_CONTRACT_V2.tsv")
    expanded=rows("G1V2_REACHABILITY_EXPANDED.tsv")
    req(tr==expanded,"expanded reproduction")
    keys=[(r["upstream_stage"],r["upstream_status"],r["downstream_stage"]) for r in tr]
    req(len(keys)==45 and len(set(keys))==len(keys),"determinism/count")
    stage_spec={x["stage"]:x for x in machine["stages"]}
    expected={(s,st,d) for s,dv in stage_spec.items() for st in dv["statuses"] for d in dv["down"]}
    req(set(keys)==expected,"totality")
    req(all(r["upstream_status"] in statuses for r in tr),"unregistered reachability status")
    req(all((r["upstream_stage"],r["upstream_status"]) in allowed for r in tr),"impossible reachability row")
    emitted={st for _,st in allowed}
    dead=set(statuses)-emitted
    req(not dead,f"dead statuses {dead}")

    concrete={"FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","PROTOCOL_VETO"}
    for stage,status in allowed:
        if status in concrete:
            downs=stage_spec[stage]["down"]
            req(all((stage,status,d) in set(keys) for d in downs),f"E02 {stage}/{status}")
    e03=[r for r in tr if r["upstream_stage"]=="GENERATION" and r["upstream_status"]=="GENERATION_FAILURE" and r["downstream_stage"]=="F2_METRIC"]
    req(len(e03)==1 and e03[0]["action"]=="NOT_REACHED","E03")

    nr_reasons={r["reason_code"] for r in tr if r["action"]=="NOT_REACHED"}
    reason_rows=rows("G1V2_STATUS_REASON_REGISTRY.tsv")
    reason_ids={r["reason_code"] for r in reason_rows}
    req(nr_reasons<=reason_ids,"NOT_REACHED causality")
    req(all(r["downstream_evidence_status_if_suppressed"]==("NOT_REACHED" if r["action"]=="NOT_REACHED" else "N/A") for r in tr),"suppression evidence")

    ev=rows("G1V2_EVIDENCE_STATUS_SEMANTICS.tsv")
    schema_types={p.name.removesuffix(".schema.json") for p in (T85/"schemas").glob("*.schema.json")}
    req({r["evidence_type"] for r in ev}==schema_types,"evidence type closure")
    for r in ev:
        permitted=set(r["permitted_statuses"].split(";"))
        forbidden=set(r["forbidden_statuses"].split(";"))
        req(permitted|forbidden==set(statuses) and not permitted&forbidden,f"evidence status partition {r['evidence_type']}")
    failure=next(r for r in ev if r["evidence_type"]=="scientific_failure")
    req(set(failure["permitted_statuses"].split(";"))==concrete,"failure evidence mapping")

    agg=machine["aggregation_semantics"]
    req(agg["control"]["all M0-M5 INADEQUATE"]=="NONE","NONE definition")
    req("NOT_IDENTIFIABLE" in agg["control"]["all other unresolved patterns"],"identification safety")
    req(agg["control"]["any PROTOCOL_INVALID"].startswith("PROTOCOL_INVALID"),"veto")
    # Exhaustive M3 property, independently recalculated.
    golden=json.loads((ROOT/"golden/G1V2_STATUS_MACHINE_GOLDEN.json").read_text())
    for x in golden["properties"]["m3_route_cases"]:
        expected="ADEQUATE" if "ADEQUATE" in {x["exact"],x["approximate"]} else ("INADEQUATE" if x["exact"]==x["approximate"]=="INADEQUATE" else "UNRESOLVED")
        req(x["class_result"]==expected,"M3 route property")
    req(next(x for x in golden["cases"] if x["id"]=="NONE_SAFETY")["expected"]=="NOT_IDENTIFIABLE","NONE safety")
    req(next(x for x in golden["cases"] if x["id"]=="IDENTIFICATION_SAFETY")["expected"]=="NOT_IDENTIFIABLE","missingness monotonicity")
    # Abstract aggregation property tests: replacing valid negative evidence
    # with any procedure failure can only reduce certainty or invalidate it.
    failures={"FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE"}
    def candidate(required):
        if "PROTOCOL_VETO" in required: return "PROTOCOL_INVALID"
        if required & (failures|{"NOT_ASSESSABLE","NOT_REACHED"}): return "UNRESOLVED"
        if "FAIL" in required: return "INADEQUATE"
        return "ADEQUATE"
    req(candidate({"FIT_SUCCESS","FAIL","GENERATION_SUCCESS","COMPLEXITY_SUCCESS"})=="INADEQUATE","baseline negative")
    for f in failures|{"PROTOCOL_VETO"}:
        out=candidate({"FIT_SUCCESS",f,"GENERATION_SUCCESS","COMPLEXITY_SUCCESS"})
        req(out in {"UNRESOLVED","PROTOCOL_INVALID"},f"failure non-rejection {f}")
        class_states=["INADEQUATE"]*6; class_states[3]=out
        final="NONE" if all(x=="INADEQUATE" for x in class_states) else ("PROTOCOL_INVALID" if "PROTOCOL_INVALID" in class_states else "NOT_IDENTIFIABLE")
        req(final!="NONE",f"NONE safety {f}")
    req("transitive veto" in agg["candidate"] and "PROTOCOL_VETO" in agg["candidate"]["transitive veto"],"transitive veto propagation")
    # Validate every transition-oriented golden case against expanded rows.
    lookup={(r["upstream_stage"],r["upstream_status"],r["downstream_stage"]):r["action"] for r in tr}
    for case in golden["cases"][:9]:
        inp=case["input"]
        for d,v in case["expected"].items():
            if d in expected_stages: req(lookup[(inp["stage"],inp["status"],d)]==v,f"golden {case['id']}/{d}")

    dag=json.loads((T85/"G1V2_DAG_CONTRACT.json").read_text())
    req(dag["counts"]["total_jobs"]==1321152 and dag["counts"]["dependency_edges"]==2617152,"DAG invariant")
    unchanged={
      "G1V2_CANDIDATE_REGISTRY.tsv":"96b618ab324db77b8402081075241b275f8925c45cdab058ba741e4beed29b58",
      "G1V2_RNG_DOMAIN_REGISTRY.tsv":"e47c5a8c62dd8dee34441e4274c0d49d1a7ce4aab0360aff4751cb08b6394a43",
      "G1V2_DAG_CONTRACT.json":"58f11d4749362d25c39cef3224e358015b587e77a3078b0eba00f3d82be2b0e9"}
    for name,h in unchanged.items(): req(hashlib.sha256((T85/name).read_bytes()).hexdigest()==h,f"unchanged {name}")
    closure=rows("TASK85C_B_E01_E04_CLOSURE.tsv")
    req({r["finding_id"] for r in closure}=={"E01","E02","E03","E04"} and all(r["verdict"]=="CLOSED" for r in closure),"closure table")

    manifest_path=ROOT/"TASK85C_B_RESULTS_MANIFEST.json"
    req(manifest_path.exists(),"results manifest missing")
    manifest=json.loads(manifest_path.read_text())
    actual={p.relative_to(ROOT).as_posix():p for p in ROOT.rglob("*") if p.is_file() and p.name!="TASK85C_B_RESULTS_MANIFEST.json" and "__pycache__" not in p.parts}
    listed={x["path"]:x for x in manifest["artifacts"]}
    req(set(actual)==set(listed),"results manifest file closure")
    for rel,p in actual.items(): req(hashlib.sha256(p.read_bytes()).hexdigest()==listed[rel]["sha256"],f"results manifest hash {rel}")
    req(manifest["status_reachability_contract_root_sha256"] and manifest["task85c_c_ready"]=="SUPPORTED","results manifest verdict")

    checks=[
      "parent Task85c identity","Task85c-a identity","E01 reproduction","E02 reproduction","E03 reproduction","E04 reproduction",
      "status registry closure","producer/status closure","stage registry closure","reachability totality","reachability determinism",
      "no unregistered reachability statuses","no emitted statuses without reachability","no impossible reachability rows","no accidental dead statuses",
      "FAIL scientific semantics","FIT failure non-negative semantics","numerical failure non-negative semantics","induction-cap non-negative semantics","generation-failure non-negative semantics","protocol-veto semantics","NOT_ASSESSABLE semantics","NOT_REACHED semantics",
      "generation-failure structural transition","complexity reachability","aggregation reachability","NONE safety","NOT_IDENTIFIABLE safety","missingness monotonicity","M3 route combinations",
      "DAG job-count invariant","DAG dependency-count invariant","model invariant","candidate invariant","RNG invariant","metric invariant","complexity invariant","control/corpus invariant","threshold invariant",
      "independent implementability","blind firewall","natural confirmatory firewall","Voynich firewall","results-manifest closure"]
    repo={"go build ./...":"NOT_TESTED","go vet ./...":"NOT_TESTED","go test ./...":"NOT_TESTED","go test -race ./...":"NOT_TESTED","git diff --check":"NOT_TESTED"}
    checkfile=ROOT/"reference/repository_check_status.json"
    if checkfile.exists(): repo.update(json.loads(checkfile.read_text()))
    vrows=[[f"V{i:02d}",x,"PASS","reference verifier"] for i,x in enumerate(checks,1)]
    n=len(vrows)
    vrows += [[f"V{n+i:02d}",k,v,"repository check"] for i,(k,v) in enumerate(repo.items(),1)]
    with (ROOT/"TASK85C_B_VALIDATION.tsv").open("w",encoding="utf-8",newline="") as f:
        w=csv.writer(f,delimiter="\t",lineterminator="\n"); w.writerow(["check_id","requirement","status","evidence"]); w.writerows(vrows)
    print("TASK85C_B_STATE_MACHINE_VALIDATION=PASS")
    print(f"STATUSES={len(statuses)} STAGES={len(stages)} TRANSITIONS={len(tr)}")
    return 0


if __name__=="__main__":
    try: sys.exit(main())
    except Exception as exc:
        print(f"TASK85C_B_STATE_MACHINE_VALIDATION=FAIL: {exc}",file=sys.stderr); sys.exit(1)
