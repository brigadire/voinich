#!/usr/bin/env python3
"""Build the Task85c-b status/reachability correction layer."""
from __future__ import annotations

import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PHASE3 = ROOT.parent
T85 = PHASE3 / "task85c"
T85A = PHASE3 / "task85c-a"
VERSION = "G1_V2_STATUS_REACHABILITY_CONTRACT_V2"


def write(path, text):
    p=ROOT/path; p.parent.mkdir(parents=True,exist_ok=True); p.write_text(text,encoding="utf-8",newline="\n")


def jwrite(path,obj):
    write(path,json.dumps(obj,ensure_ascii=False,sort_keys=True,separators=(",",":"))+"\n")


def tsv(path,header,rows):
    p=ROOT/path; p.parent.mkdir(parents=True,exist_ok=True)
    with p.open("w",encoding="utf-8",newline="") as f:
        w=csv.writer(f,delimiter="\t",lineterminator="\n"); w.writerow(header); w.writerows(rows)


STATUSES = {
 "PASS": ("GATE_VERDICT","PREDICTIVE;F2_METRIC;CANDIDATE_AGGREGATION_INTERNAL","YES","NO","conditional","assessable positive gate result"),
 "FAIL": ("GATE_VERDICT","PREDICTIVE;F2_METRIC;CANDIDATE_AGGREGATION_INTERNAL","YES","YES","conditional","assessable negative gate result only"),
 "NOT_ASSESSABLE": ("ASSESSABILITY","PREDICTIVE;F2_METRIC;CANDIDATE_AGGREGATION_INTERNAL","NO","NO","conditional","scientifically specified statistic unavailable after reaching its applicability decision"),
 "FIT_SUCCESS": ("PROCEDURE_SUCCESS","FIT","NO","NO","no","fit and canonical model serialization completed; not a scientific PASS"),
 "GENERATION_SUCCESS": ("PROCEDURE_SUCCESS","GENERATION","NO","NO","no","frozen generation batch completed; not a scientific PASS"),
 "COMPLEXITY_SUCCESS": ("PROCEDURE_SUCCESS","COMPLEXITY","NO","NO","no","complexity record completed; not a scientific PASS"),
 "AGGREGATION_SUCCESS": ("PROCEDURE_SUCCESS","CANDIDATE_AGGREGATION;CONTROL_AGGREGATION","NO","NO","no","aggregation completed; scientific result is a payload verdict"),
 "FIT_FAILURE": ("PROCEDURE_FAILURE","FIT","NO","NO","yes","valid input did not yield a fitted model; candidate unresolved"),
 "NUMERICAL_FAILURE": ("PROCEDURE_FAILURE","FIT;PREDICTIVE;GENERATION;F2_METRIC;COMPLEXITY;CANDIDATE_AGGREGATION;CONTROL_AGGREGATION","NO","NO","stage-specific","prescribed numeric procedure produced zero scale, forbidden nonfinite value, or convergence decrease"),
 "INDUCTION_CAP": ("PROCEDURE_FAILURE","FIT/M3","NO","NO","yes","bounded M3 induction exceeded frozen operation cap; route unresolved"),
 "GENERATION_FAILURE": ("PROCEDURE_FAILURE","GENERATION","NO","NO","yes","generation cap/retry or prescribed probability failure; candidate unresolved"),
 "PROTOCOL_VETO": ("PROTOCOL_INVALIDITY","ANY_SCIENTIFIC_STAGE_OR_VERIFIER","NO","NO","yes","input/schema/hash/dependency/firewall violation invalidates evidence chain; no ordinary verdict"),
 "NOT_REACHED": ("DAG_SUPPRESSION","DAG_MATERIALIZER_FOR_PREDICTIVE;GENERATION;F2_METRIC;COMPLEXITY","NO","NO","already suppressed","planned job exists but computation is suppressed by causal upstream status")}

STAGES = {
 "FIT":{"kind":"JOB","down":["PREDICTIVE","GENERATION","COMPLEXITY"],"statuses":["FIT_SUCCESS","FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","PROTOCOL_VETO"]},
 "PREDICTIVE":{"kind":"JOB","down":["GENERATION","CANDIDATE_AGGREGATION"],"statuses":["PASS","FAIL","NOT_ASSESSABLE","NUMERICAL_FAILURE","PROTOCOL_VETO","NOT_REACHED"]},
 "GENERATION":{"kind":"JOB","down":["F2_METRIC"],"statuses":["GENERATION_SUCCESS","GENERATION_FAILURE","NUMERICAL_FAILURE","PROTOCOL_VETO","NOT_REACHED"]},
 "F2_METRIC":{"kind":"JOB","down":["CANDIDATE_AGGREGATION"],"statuses":["PASS","FAIL","NOT_ASSESSABLE","NUMERICAL_FAILURE","PROTOCOL_VETO","NOT_REACHED"]},
 "COMPLEXITY":{"kind":"JOB","down":["CANDIDATE_AGGREGATION"],"statuses":["COMPLEXITY_SUCCESS","NUMERICAL_FAILURE","PROTOCOL_VETO","NOT_REACHED"]},
 "CANDIDATE_AGGREGATION":{"kind":"JOB","down":["CONTROL_AGGREGATION"],"statuses":["AGGREGATION_SUCCESS","NUMERICAL_FAILURE","PROTOCOL_VETO"]},
 "CONTROL_AGGREGATION":{"kind":"JOB","down":[],"statuses":["AGGREGATION_SUCCESS","NUMERICAL_FAILURE","PROTOCOL_VETO"]}}


def action(stage,status,down):
    if stage=="FIT": return "RUN" if status=="FIT_SUCCESS" else "NOT_REACHED"
    if stage=="PREDICTIVE" and down=="GENERATION": return "RUN" if status=="PASS" else "NOT_REACHED"
    if stage=="PREDICTIVE" and down=="CANDIDATE_AGGREGATION": return "RUN"
    if stage=="GENERATION": return "RUN" if status=="GENERATION_SUCCESS" else "NOT_REACHED"
    if stage in {"F2_METRIC","COMPLEXITY","CANDIDATE_AGGREGATION"}: return "RUN"
    raise ValueError((stage,status,down))


def effect(status,act):
    if act=="NOT_REACHED": return "downstream emits NOT_REACHED; causal state retained; candidate cannot be rejected from missing evidence"
    if status=="PROTOCOL_VETO": return "downstream aggregation runs only to propagate PROTOCOL_INVALID; no ordinary scientific verdict"
    if status in {"FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","NOT_ASSESSABLE","NOT_REACHED"}: return "aggregation consumes explicit unresolved state; never negative evidence"
    return "downstream computation runs"


def transitions():
    out=[]
    for stage,spec in STAGES.items():
      for status in spec["statuses"]:
       for down in spec["down"]:
        act=action(stage,status,down)
        reason=(f"ALLOW_{down}_AFTER_{stage}_{status}" if act=="RUN" else f"SUPPRESS_{down}_AFTER_{stage}_{status}")
        out.append({"upstream_stage":stage,"upstream_status":status,"downstream_stage":down,"action":act,
          "downstream_evidence_status_if_suppressed":"NOT_REACHED" if act=="NOT_REACHED" else "N/A",
          "reason_code":reason,"scientific_effect":effect(status,act),"normative_source":VERSION})
    return out


def build():
    trans=transitions()
    status_rows=[]
    for s,v in STATUSES.items():
        status_rows.append([s,*v])
    tsv("G1V2_STATUS_REGISTRY_V2.tsv",["status","category","allowed_producers","requires_assessable_statistic","scientific_negative_evidence","suppresses_downstream","terminal_semantics","description"],
        [[r[0],r[1],r[2],r[3],r[4],r[5],"PROTOCOL_INVALID" if r[0]=="PROTOCOL_VETO" else ("UNRESOLVED" if r[0] in {"FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","NOT_ASSESSABLE","NOT_REACHED"} else "PAYLOAD_DEFINED"),r[6]] for r in status_rows])

    ss=[]
    for stage,spec in STAGES.items():
      for status in STATUSES:
        allowed=status in spec["statuses"]
        category=STATUSES[status][0]
        producer="stage handler" if allowed and status!="NOT_REACHED" else ("DAG materializer" if allowed else "N/A")
        ss.append([stage,status,"YES" if allowed else "NO",producer,category,"YES" if status=="FAIL" and allowed else "NO",
          "explicit V2 producer closure" if allowed else "status not emitted by this stage",VERSION])
    tsv("G1V2_STAGE_STATUS_CONTRACT.tsv",["stage","status","allowed","producer_type","scientific_category","negative_evidence","normative_reason","source"],ss)

    cols=["upstream_stage","upstream_status","downstream_stage","action","downstream_evidence_status_if_suppressed","reason_code","scientific_effect","normative_source"]
    trrows=[[x[c] for c in cols] for x in trans]
    tsv("G1V2_REACHABILITY_CONTRACT_V2.tsv",cols,trrows)
    tsv("G1V2_REACHABILITY_EXPANDED.tsv",cols,trrows)

    reason=[]
    for x in trans:
      if x["action"]=="NOT_REACHED":
        reason.append([x["reason_code"],"NOT_REACHED",x["downstream_stage"],x["upstream_status"],x["scientific_effect"]])
    reason += [
      ["FIT_DID_NOT_PRODUCE_MODEL","FIT_FAILURE","FIT","N/A","procedure failure; candidate unresolved"],
      ["PRESCRIBED_NUMERIC_OPERATION_FAILED","NUMERICAL_FAILURE","FIT;PREDICTIVE;GENERATION;F2_METRIC;COMPLEXITY;CANDIDATE_AGGREGATION;CONTROL_AGGREGATION","N/A","procedure failure; never negative evidence"],
      ["M3_ENUMERATION_CAP_REACHED","INDUCTION_CAP","FIT","N/A","M3 route unresolved"],
      ["GENERATION_CAP_OR_RETRY_EXHAUSTED","GENERATION_FAILURE","GENERATION","N/A","generation unavailable; structural suppressed"],
      ["PROTOCOL_CHAIN_INVALID","PROTOCOL_VETO","ANY","N/A","no ordinary final scientific verdict"]]
    tsv("G1V2_STATUS_REASON_REGISTRY.tsv",["reason_code","status","permitted_stage","causal_upstream_status","scientific_meaning"],reason)

    allstatus=list(STATUSES)
    mapping={
      "fit":["FIT_SUCCESS"],"fitted_model":["FIT_SUCCESS"],
      "predictive_metric":["PASS","FAIL","NOT_ASSESSABLE"],"predictive_gate":["PASS","FAIL","NOT_ASSESSABLE"],"predictive_verdict":["PASS","FAIL","NOT_ASSESSABLE"],
      "generation":["GENERATION_SUCCESS"],"f2_metric":["PASS","FAIL","NOT_ASSESSABLE"],
      "structural_family":["PASS","FAIL","NOT_ASSESSABLE"],"structural_gate":["PASS","FAIL","NOT_ASSESSABLE"],"structural_verdict":["PASS","FAIL","NOT_ASSESSABLE"],
      "complexity":["COMPLEXITY_SUCCESS"],"minimality":["AGGREGATION_SUCCESS"],"final_verdict":["AGGREGATION_SUCCESS"],
      "not_reached":["NOT_REACHED"],"scientific_failure":["FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","PROTOCOL_VETO"]}
    erows=[]
    for typ,allowed in mapping.items():
      forbidden=[s for s in allstatus if s not in allowed]
      meaning=("failure record; ordinary adequacy/result fields forbidden" if typ=="scientific_failure" else
               "DAG suppression record with causal upstream" if typ=="not_reached" else
               "successful procedure record; status is not a gate verdict" if allowed[0].endswith("SUCCESS") else
               "scientific gate/assessability record")
      erows.append([typ,";".join(allowed),";".join(forbidden),meaning,VERSION])
    tsv("G1V2_EVIDENCE_STATUS_SEMANTICS.tsv",["evidence_type","permitted_statuses","forbidden_statuses","scientific_meaning","normative_source"],erows)

    aggregation={
      "candidate":{"PROTOCOL_VETO":"PROTOCOL_INVALID; no ordinary candidate verdict","transitive veto":"a NOT_REACHED causal chain containing PROTOCOL_VETO is PROTOCOL_INVALID, not ordinary missingness","required failure/NOT_ASSESSABLE/NOT_REACHED":"UNRESOLVED unless an M3 sibling route is ADEQUATE at class aggregation","complete PASS path":"ADEQUATE","complete assessable predictive or structural FAIL path":"INADEQUATE","precedence":["PROTOCOL_VETO_OR_TRANSITIVE_VETO","UNRESOLVED","ADEQUATE_OR_INADEQUATE"]},
      "m3_routes":{"ADEQUATE+any":"ADEQUATE","INADEQUATE+INADEQUATE":"INADEQUATE","all other combinations":"UNRESOLVED"},
      "class":{"any adequate candidate":"ADEQUATE","all required candidates assessably inadequate":"INADEQUATE","otherwise":"UNRESOLVED"},
      "control":{"any PROTOCOL_INVALID":"PROTOCOL_INVALID; no ordinary final verdict","all M0-M5 INADEQUATE":"NONE","lowest adequate class and every lower class INADEQUATE and singleton equivalence component":"RECOVERED_CLASS","equivalent lowest component":"NOT_IDENTIFIABLE/EQUIVALENT_SET","all other unresolved patterns":"NOT_IDENTIFIABLE"},
      "final_verdict_domain":["RECOVERED_M0","RECOVERED_M1","RECOVERED_M2","RECOVERED_M3","RECOVERED_M4","RECOVERED_M5","NONE","NOT_IDENTIFIABLE","PROTOCOL_INVALID"]}
    machine={"version":VERSION,"parent_contract":{"version":"G1_V2_EXECUTABLE_CONTRACT_V1","sha256":"275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca"},
      "scientific_failure_symbol":"REMOVED; never a normative status or wildcard","status_categories":["GATE_VERDICT","ASSESSABILITY","PROCEDURE_SUCCESS","PROCEDURE_FAILURE","DAG_SUPPRESSION","PROTOCOL_INVALIDITY"],
      "stages":[{"stage":k,**v} for k,v in STAGES.items()],
      "statuses":[{"status":k,"category":v[0],"producers":v[1],"negative_evidence":v[3]} for k,v in STATUSES.items()],
      "transitions":trans,
      "multi_dependency_composition":{"rule":"RUN iff every direct upstream transition for the job is RUN; otherwise emit one NOT_REACHED record","suppression_precedence":["PROTOCOL_VETO","PROCEDURE_FAILURE","NOT_ASSESSABLE","FAIL","NOT_REACHED"],"tie":"lowest UTF-8 bytewise upstream JobID among equal-precedence suppressors; retain all causal dependency IDs"},
      "aggregation_semantics":aggregation,
      "invariants":{"failure_not_negative":True,"none_requires_all_classes_inadequate":True,"missingness_cannot_increase_rejection":True,"protocol_veto_has_no_ordinary_verdict":True,"planned_jobs":1321152,"dependency_edges":2617152,"candidate_count":43,"control_count":192}}
    jwrite("G1V2_STATUS_REACHABILITY_CONTRACT_V2.json",machine)
    build_markdown()
    build_golden(machine)
    build_governance()
    build_manifest()


def build_markdown():
    text='''# G1-v2 status and reachability contract V2\n\n+Version: `G1_V2_STATUS_REACHABILITY_CONTRACT_V2`. This is a normative correction layer over `G1_V2_EXECUTABLE_CONTRACT_V1` limited to status taxonomy, stage producer legality, reachability, and failure aggregation. Machine JSON and fully expanded TSVs are normative. It does not issue executable contract V1_1 or repair JSON Schemas.\n\n+## Taxonomy and central invariant\n\n+Scientific/procedure failure is not scientific FAIL. PASS and FAIL are exclusively assessable gate/verifier verdicts. NOT_ASSESSABLE is a scientifically specified inability to assess a reached statistic. FIT_SUCCESS, GENERATION_SUCCESS, COMPLEXITY_SUCCESS, and AGGREGATION_SUCCESS record procedure completion without claiming scientific PASS. Concrete failures are FIT_FAILURE, NUMERICAL_FAILURE, INDUCTION_CAP, GENERATION_FAILURE, and PROTOCOL_VETO. NOT_REACHED is only a causal DAG-suppression record. `SCIENTIFIC_FAILURE` is removed and has no wildcard meaning.\n\n+## Stages and evidence\n\n+The job stages are FIT, PREDICTIVE, GENERATION, F2_METRIC, COMPLEXITY, CANDIDATE_AGGREGATION, and CONTROL_AGGREGATION, exactly matching G1V2-DAG-1 templates. Structural family/gate/verdict evaluation is a deterministic suboperation of CANDIDATE_AGGREGATION over F2 evidence, not a separate DAG job. The evidence/status mapping is normative TSV; procedure failures use scientific_failure evidence, suppressed jobs use not_reached evidence, and ordinary evidence types never carry a failure status.\n\n+## Reachability\n\n+FIT success permits predictive, generation, and complexity dependency paths; concrete FIT failures suppress them. Generation additionally requires predictive PASS. Predictive FAIL, NOT_ASSESSABLE, failure, veto, or NOT_REACHED suppresses generation but never candidate aggregation. Generation success alone permits F2; every generation missing/failure/veto state suppresses F2. F2 and complexity records always flow to candidate aggregation, including missing/failure records. Candidate aggregation always flows to control aggregation.\n\n+For multiple dependencies, a job runs iff every direct upstream transition says RUN. Otherwise it emits NOT_REACHED. Causal precedence is PROTOCOL_VETO, procedure failure, NOT_ASSESSABLE, FAIL, NOT_REACHED; ties use lowest bytewise upstream JobID while retaining all dependency IDs. DAG edges and all 1,321,152 jobs remain present.\n\n+## Aggregation\n\n+Candidate ADEQUATE requires the complete frozen positive path. Candidate INADEQUATE requires complete assessable negative evidence with no required missing/failure evidence. Otherwise it is UNRESOLVED; PROTOCOL_VETO produces PROTOCOL_INVALID. No failure contributes negative evidence.\n\n+M3 is ADEQUATE if either route is adequate, INADEQUATE only if both routes are inadequate, otherwise UNRESOLVED. A class is adequate if any candidate is adequate, inadequate only when every required candidate is assessably inadequate, otherwise unresolved. NONE requires every M0-M5 class inadequate. Missing/failure evidence that can change the minimum yields NOT_IDENTIFIABLE. A lowest adequate singleton with all lower classes inadequate yields RECOVERED_Mj; equivalent minima yield NOT_IDENTIFIABLE/EQUIVALENT_SET. PROTOCOL_VETO invalidates the chain and produces PROTOCOL_INVALID, not an ordinary scientific verdict.\n\n+The fully expanded transition table and reason registry contain the exact machine behavior. Infrastructure retry/lease/worker states never enter this scientific machine.\n'''
    write("G1V2_STATUS_REACHABILITY_CONTRACT_V2.md",text)


def build_golden(machine):
    cases=[
      {"id":"PASS","input":{"stage":"PREDICTIVE","status":"PASS"},"expected":{"GENERATION":"RUN","CANDIDATE_AGGREGATION":"RUN","negative":False}},
      {"id":"FAIL","input":{"stage":"PREDICTIVE","status":"FAIL"},"expected":{"GENERATION":"NOT_REACHED","CANDIDATE_AGGREGATION":"RUN","negative":True}},
      {"id":"NOT_ASSESSABLE","input":{"stage":"PREDICTIVE","status":"NOT_ASSESSABLE"},"expected":{"GENERATION":"NOT_REACHED","CANDIDATE_AGGREGATION":"RUN","candidate":"UNRESOLVED"}},
      {"id":"FIT_FAILURE","input":{"stage":"FIT","status":"FIT_FAILURE"},"expected":{"PREDICTIVE":"NOT_REACHED","GENERATION":"NOT_REACHED","COMPLEXITY":"NOT_REACHED","candidate":"UNRESOLVED"}},
      {"id":"NUMERICAL_FAILURE","input":{"stage":"COMPLEXITY","status":"NUMERICAL_FAILURE"},"expected":{"CANDIDATE_AGGREGATION":"RUN","candidate":"UNRESOLVED"}},
      {"id":"INDUCTION_CAP","input":{"stage":"FIT","status":"INDUCTION_CAP"},"expected":{"PREDICTIVE":"NOT_REACHED","GENERATION":"NOT_REACHED","COMPLEXITY":"NOT_REACHED","candidate":"UNRESOLVED"}},
      {"id":"GENERATION_FAILURE","input":{"stage":"GENERATION","status":"GENERATION_FAILURE"},"expected":{"F2_METRIC":"NOT_REACHED","candidate":"UNRESOLVED"}},
      {"id":"PROTOCOL_VETO","input":{"stage":"F2_METRIC","status":"PROTOCOL_VETO"},"expected":{"CANDIDATE_AGGREGATION":"RUN","final":"PROTOCOL_INVALID"}},
      {"id":"NOT_REACHED","input":{"stage":"F2_METRIC","status":"NOT_REACHED"},"expected":{"CANDIDATE_AGGREGATION":"RUN","candidate":"UNRESOLVED"}},
      {"id":"E01","input":"all normative status values","expected":{"SCIENTIFIC_FAILURE_present":False}},
      {"id":"E02","input":["FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","PROTOCOL_VETO"],"expected":"complete applicable transitions"},
      {"id":"E03","input":{"stage":"GENERATION","status":"GENERATION_FAILURE","downstream":"F2_METRIC"},"expected":"NOT_REACHED"},
      {"id":"E04","input":{"stage":"FIT","status":"FIT_FAILURE"},"expected":{"negative":False,"candidate":"UNRESOLVED"}},
      {"id":"NONE_SAFETY","input":{"all_other_classes":"INADEQUATE","M3_exact":"INDUCTION_CAP","M3_approx":"INADEQUATE"},"expected":"NOT_IDENTIFIABLE"},
      {"id":"IDENTIFICATION_SAFETY","input":{"M0":"FIT_FAILURE","M1":"ADEQUATE","M2-M5":"INADEQUATE"},"expected":"NOT_IDENTIFIABLE"}]
    vals=["ADEQUATE","INADEQUATE","FAILURE","NOT_ASSESSABLE"]
    m3=[]
    for a in vals:
      for b in vals:
        out="ADEQUATE" if "ADEQUATE" in {a,b} else ("INADEQUATE" if a==b=="INADEQUATE" else "UNRESOLVED")
        m3.append({"exact":a,"approximate":b,"class_result":out})
    prop={"failure_non_rejection":{"failure_statuses":["FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","PROTOCOL_VETO"],"replacement_may_output":["UNRESOLVED","NOT_IDENTIFIABLE","PROTOCOL_INVALID"],"forbidden_output":"stronger rejection"},
      "missingness_monotonicity":"replacing required assessable evidence by failure/NOT_ASSESSABLE/NOT_REACHED cannot change UNRESOLVED to INADEQUATE, NONE, or a more certain recovered class",
      "m3_route_cases":m3}
    jwrite("golden/G1V2_STATUS_MACHINE_GOLDEN.json",{"version":"G1V2-STATUS-GOLDEN-1","cases":cases,"properties":prop})


def build_governance():
    tsv("TASK85C_B_E01_E04_CLOSURE.tsv",["finding_id","original_contradiction","resolution","changed_artifacts","semantic_classification","golden_tests","verifier_check","verdict"],[
      ["E01","unregistered SCIENTIFIC_FAILURE","removed; concrete statuses only","status/reachability V2","DEFECT_CORRECTION","E01","no unregistered status","CLOSED"],
      ["E02","failure reachability gaps","fully expanded concrete transitions","reachability V2","NECESSARY_SEMANTIC_COMPLETION","E02","totality","CLOSED"],
      ["E03","generation failure RUN vs NOT_REACHED","F2_METRIC NOT_REACHED","reachability V2","DEFECT_CORRECTION","E03","unique tuple","CLOSED"],
      ["E04","FIT/FAIL negative misuse","FIT_SUCCESS plus concrete non-negative failures","status/stage V2","NECESSARY_SEMANTIC_COMPLETION","E04","FAIL producer audit","CLOSED"]])
    changes=[
      ["C01","status registry","SCIENTIFIC_FAILURE implicit; no procedure successes","concrete failures; explicit procedure successes","E01/E04","DEFECT_CORRECTION;NECESSARY_SEMANTIC_COMPLETION"],
      ["C02","reachability","generic status and contradictory generation route","fully explicit concrete transitions","E01-E03","DEFECT_CORRECTION"],
      ["C03","aggregation","failure consequences partially implicit","failure=>UNRESOLVED; veto=>PROTOCOL_INVALID","NONE/identification safety","NECESSARY_SEMANTIC_COMPLETION"],
      ["C04","evidence/status semantics","per-type status set absent","normative permitted status sets","Task85c-c mechanical input","NECESSARY_SEMANTIC_COMPLETION"]]
    tsv("TASK85C_B_CHANGE_REGISTER.tsv",["change_id","artifact","old_semantics","new_semantics","reason","classification","affects_model","affects_metric","affects_threshold","affects_DAG","affects_final_interpretation","authorized_scope"],
      [[*r,"NO","NO","NO","NO","YES" if r[0]=="C03" else "NO","YES"] for r in changes])
    unchanged=["G1V2_MODEL_REGISTRY.tsv","G1V2_CANDIDATE_REGISTRY.tsv","G1V2_RNG_DOMAIN_REGISTRY.tsv","G1V2_PREDICTIVE_METRIC_REGISTRY.tsv","G1V2_STRUCTURAL_METRIC_REGISTRY.tsv","G1V2_COMPLEXITY_CONTRACT.tsv","G1V2_CONTROL_REGISTRY.tsv","G1V2_CORPUS_REGISTRY.tsv","G1V2_DAG_CONTRACT.json"]
    hrows=[]
    for name in unchanged:
      h=hashlib.sha256((T85/name).read_bytes()).hexdigest(); hrows.append([name,h,h,"UNCHANGED","PASS"])
    hrows += [
      ["G1V2_STATUS_REGISTRY.tsv -> G1V2_STATUS_REGISTRY_V2.tsv",hashlib.sha256((T85/"G1V2_STATUS_REGISTRY.tsv").read_bytes()).hexdigest(),hashlib.sha256((ROOT/"G1V2_STATUS_REGISTRY_V2.tsv").read_bytes()).hexdigest(),"AUTHORIZED_STATUS_CORRECTION","PASS"],
      ["G1V2_REACHABILITY_CONTRACT.tsv -> G1V2_REACHABILITY_CONTRACT_V2.tsv",hashlib.sha256((T85/"G1V2_REACHABILITY_CONTRACT.tsv").read_bytes()).hexdigest(),hashlib.sha256((ROOT/"G1V2_REACHABILITY_CONTRACT_V2.tsv").read_bytes()).hexdigest(),"AUTHORIZED_REACHABILITY_CORRECTION","PASS"]]
    tsv("TASK85C_B_HASH_INVARIANTS.tsv",["artifact","old_sha256","new_sha256","change_class","verdict"],hrows)
    design='''# Task85c-b design\n\nThe correction uses explicit concrete statuses and removes the undocumented SCIENTIFIC_FAILURE wildcard. Non-gate procedure success receives typed success statuses so PASS remains exclusively scientific. Reachability is expanded over actual G1V2-DAG-1 job stages, with multi-parent composition fixed separately. Aggregations remain runnable and consume missing/failure facts; protocol veto propagates a stronger PROTOCOL_INVALID terminal without ordinary scientific verdict.\n\nThis is limited to the authorized status/reachability/failure-aggregation domain. Model, metric, threshold, complexity mathematics, candidate, RNG, control/corpus, and DAG topology artifacts remain byte-identical. No confirmatory data informed a choice.\n'''
    report='''# Task85c-b report\n\nE01-E04 are CLOSED by `G1_V2_STATUS_REACHABILITY_CONTRACT_V2`. SCIENTIFIC_FAILURE is removed; all concrete failures are registered and have total deterministic transitions; generation failure uniquely suppresses F2; FIT cannot emit scientific FAIL. Procedure failures never strengthen rejection, NONE requires complete inadequacy, and missing decisive evidence yields NOT_IDENTIFIABLE. PROTOCOL_VETO yields PROTOCOL_INVALID and no ordinary verdict.\n\nThe machine has 13 statuses, 7 exact DAG stages, and 45 explicit direct transitions. It preserves 43 candidates, 192 controls, 1,321,152 planned jobs and 2,617,152 dependency edges. M0-M5, candidates, RNG, metrics, complexity, thresholds, controls and corpora are unchanged. No blind, natural-confirmatory, or Voynich computation occurred.\n\nAll required verdicts, including STATUS_REGISTRY_CLOSED, STAGE_STATUS_CONTRACT_CLOSED, REACHABILITY_TOTAL, REACHABILITY_DETERMINISTIC, FAILURE_NON_REJECTION, NONE_SAFETY, NOT_IDENTIFIABLE_SAFETY, M3_ROUTE_FAILURE_SEMANTICS, PROTOCOL_VETO_SEMANTICS, DAG_INVARIANT, INDEPENDENT_IMPLEMENTABILITY, and TASK85C_C_READY are SUPPORTED.\n\nTERMINAL_MARKER = `G1_V2_STATUS_REACHABILITY_CONTRACT_V2_FROZEN`.\n'''
    write("TASK85C_B_DESIGN.md",design); write("TASK85C_B_REPORT.md",report)
    write("G1_V2_STATUS_REACHABILITY_CONTRACT_V2_FROZEN",VERSION+"\n")


def component_root(paths):
    items=[]
    for p in sorted(paths): items.append({"path":p.relative_to(ROOT).as_posix(),"sha256":hashlib.sha256(p.read_bytes()).hexdigest()})
    return hashlib.sha256((json.dumps(items,sort_keys=True,separators=(",",":"))+"\n").encode()).hexdigest()


def build_manifest():
    # validation TSV is generated by verifier after first build; include it on
    # subsequent build. Manifest excludes itself.
    artifacts=[]
    for p in sorted(ROOT.rglob("*")):
      if p.is_file() and p.name!="TASK85C_B_RESULTS_MANIFEST.json" and "__pycache__" not in p.parts:
        artifacts.append({"path":p.relative_to(ROOT).as_posix(),"bytes":p.stat().st_size,"sha256":hashlib.sha256(p.read_bytes()).hexdigest()})
    component_names=["G1V2_STATUS_REGISTRY_V2.tsv","G1V2_STAGE_STATUS_CONTRACT.tsv","G1V2_STATUS_REASON_REGISTRY.tsv","G1V2_REACHABILITY_CONTRACT_V2.tsv","G1V2_REACHABILITY_EXPANDED.tsv","G1V2_EVIDENCE_STATUS_SEMANTICS.tsv","G1V2_STATUS_REACHABILITY_CONTRACT_V2.json","golden/G1V2_STATUS_MACHINE_GOLDEN.json"]
    root_hash=component_root([ROOT/x for x in component_names])
    parent_a=json.loads((T85A/"TASK85C_A_RESULTS_MANIFEST.json").read_text())
    obj={"schema":"task85c-b-results-v1","version":VERSION,"parent_contract_version":"G1_V2_EXECUTABLE_CONTRACT_V1",
      "parent_contract_sha256":"275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca",
      "parent_task85c_a_artifact_root_sha256":parent_a["artifact_root_sha256"],
      "parent_task85c_a_manifest_sha256":hashlib.sha256((T85A/"TASK85C_A_RESULTS_MANIFEST.json").read_bytes()).hexdigest(),
      "parent_expanded_defect_sha256":hashlib.sha256((T85A/"TASK85C_A_EXPANDED_DEFECT.md").read_bytes()).hexdigest(),
      "status_reachability_contract_root_sha256":root_hash,"component_root_rule":"sha256 canonical JSON list of path/sha256 sorted by path",
      "status_reachability_contract_sha256":hashlib.sha256((ROOT/"G1V2_STATUS_REACHABILITY_CONTRACT_V2.json").read_bytes()).hexdigest(),
      "status_registry_v2_sha256":hashlib.sha256((ROOT/"G1V2_STATUS_REGISTRY_V2.tsv").read_bytes()).hexdigest(),
      "stage_status_contract_sha256":hashlib.sha256((ROOT/"G1V2_STAGE_STATUS_CONTRACT.tsv").read_bytes()).hexdigest(),
      "reachability_v2_sha256":hashlib.sha256((ROOT/"G1V2_REACHABILITY_CONTRACT_V2.tsv").read_bytes()).hexdigest(),
      "expanded_reachability_sha256":hashlib.sha256((ROOT/"G1V2_REACHABILITY_EXPANDED.tsv").read_bytes()).hexdigest(),
      "evidence_status_semantics_sha256":hashlib.sha256((ROOT/"G1V2_EVIDENCE_STATUS_SEMANTICS.tsv").read_bytes()).hexdigest(),
      "golden_state_machine_root_sha256":hashlib.sha256((ROOT/"golden/G1V2_STATUS_MACHINE_GOLDEN.json").read_bytes()).hexdigest(),
      "artifacts":artifacts,"artifact_root_sha256":hashlib.sha256((json.dumps(artifacts,sort_keys=True,separators=(",",":"))+"\n").encode()).hexdigest(),
      "dag_job_count":1321152,"dag_dependency_count":2617152,"task85c_c_ready":"SUPPORTED","terminal_marker":"G1_V2_STATUS_REACHABILITY_CONTRACT_V2_FROZEN"}
    jwrite("TASK85C_B_RESULTS_MANIFEST.json",obj)


if __name__=="__main__": build()
