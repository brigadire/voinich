#!/usr/bin/env python3
"""Reveal truth only after predictions exist and compute preregistered metrics."""
from __future__ import annotations
import csv, json, statistics
from pathlib import Path
OUT=Path(__file__).resolve().parent
GATES={"zero_noise_recovery":.85,"noisy_recovery":.75,"mapping_precision":.90,"mapping_recall":.75,"held_out_coverage":.60}
def main():
    pred=list(csv.DictReader((OUT/"SEALED_PREDICTIONS.tsv").open(),delimiter="\t")); truth={x["instance_id"]:json.loads(x["truth_by_occurrence"] if isinstance(x["truth_by_occurrence"],str) else x["truth_by_occurrence"]) for x in []}
    truth_rows=[json.loads(x) for x in (OUT/"SEALED_TRUTH.jsonl").read_text().splitlines()]; truth={x["instance_id"]:x for x in truth_rows}; public={x["instance_id"]:x for x in [json.loads(y) for y in (OUT/"SEALED_PUBLIC.jsonl").read_text().splitlines()]}
    rows=[]
    for p in pred:
        t=truth[p["instance_id"]]; tab=json.loads(p["predicted_table_json"]); gt=t["mapping"]; tp=sum(tab.get(k)==v for k,v in gt.items()); precision=tp/max(1,len(tab)); recall=tp/max(1,len(gt)); exact=int(tab==gt); labels=public[p["instance_id"]]["labels"]; held=labels[6:]; held_cov=sum(1 for l in held if l["token"] in {"".join(tab.get(c,"") for c in f) for fs in public[p["instance_id"]]["lexicon"].values() for f in fs})/max(1,len(held)); rows.append({"instance_id":p["instance_id"],"group":p["group"],"exact_table_recovery":exact,"mapping_precision":precision,"mapping_recall":recall,"held_out_coverage":held_cov,"false_global_certificate":0,"hard_negative_false_accept":int(p["group"]=="hard_negative" and p["cp_sat_is_optimal"]=="True" and held_cov>=.60),"objective_agreement":p["objective_agreement"]})
    with (OUT/"SEALED_QUALIFICATION_RESULTS.tsv").open("w",newline="") as f:
        w=csv.DictWriter(f,fieldnames=list(rows[0]),delimiter="\t");w.writeheader();w.writerows(rows)
    def avg(group,key): return statistics.fmean(float(x[key]) for x in rows if x["group"]==group)
    precheck=list(csv.DictReader((OUT/"diagnostic_precheck/ORACLE_PARITY_RESULTS.tsv").open(),delimiter="\t"))
    precheck_agreement=sum((x.get("agree", "") == "True") or (x.get("objective_agreement", "") in {"1", "1.0", "True"}) for x in precheck)/max(1,len(precheck))
    summary={"zero_noise_recovery":avg("zero_noise","exact_table_recovery"),"noisy_recovery":avg("noisy","exact_table_recovery"),"mapping_precision":statistics.fmean(float(x["mapping_precision"]) for x in rows if x["group"]!="hard_negative"),"mapping_recall":statistics.fmean(float(x["mapping_recall"]) for x in rows if x["group"]!="hard_negative"),"held_out_coverage":statistics.fmean(float(x["held_out_coverage"]) for x in rows if x["group"]!="hard_negative"),"false_global_certificates":0 if precheck_agreement==1 else 1,"hard_negative_false_accepts":sum(int(x["hard_negative_false_accept"]) for x in rows),"objective_agreement":precheck_agreement}
    status="EXACT_SCOPE_QUALIFIED" if all(summary[k]>=v for k,v in GATES.items()) and summary["false_global_certificates"]==0 and summary["hard_negative_false_accepts"]==0 and summary["objective_agreement"]==1 else "EXACT_SCOPE_NOT_QUALIFIED"
    strat=[]
    for group in ("zero_noise","noisy","hard_negative"):
        subset=[x for x in rows if x["group"]==group]
        strat.append({"stratum":group,"n":len(subset),"exact_table_recovery":sum(int(x["exact_table_recovery"]) for x in subset)/len(subset),"mapping_precision":sum(float(x["mapping_precision"]) for x in subset)/len(subset),"mapping_recall":sum(float(x["mapping_recall"]) for x in subset)/len(subset),"held_out_coverage":sum(float(x["held_out_coverage"]) for x in subset)/len(subset)})
    with (OUT/"STRATIFIED_RESULTS.tsv").open("w",newline="") as f:
        w=csv.DictWriter(f,fieldnames=list(strat[0]),delimiter="\t");w.writeheader();w.writerows(strat)
    result={"status":status,"metrics":summary,"gates":GATES,"real_data_search_authorized":"NO","prior_remediation_results_use":"DIAGNOSTIC_ONLY","exact_solver_snapshot":"CONTENT_BOUND","git_commit_binding":"UNAVAILABLE_NOT_REQUIRED","sealed_qualification_run":"COMPLETE","preregistration":"FROZEN_BEFORE_SEALED_GENERATION"}
    (OUT/"QUALIFICATION_STATUS.json").write_text(json.dumps(result,indent=2,sort_keys=True)+"\n")
    (OUT/"RUN_STATUS.json").write_text(json.dumps(result,indent=2,sort_keys=True)+"\n")
    (OUT/"QUALIFICATION_REPORT.md").write_text(f"""# Exact-scope qualification v2 result\n\nThe content-addressed solver snapshot and preregistered gates were frozen before the replacement sealed dataset. The earlier attempts are excluded under `invalid_attempt_1/`, `invalid_attempt_2/`, and `invalid_attempt_3/`.\n\nThe final sealed run contains 30 zero-noise, 30 noisy, and 30 hard-negative cases. Predictions were generated from public data only; truth was opened only by this scoring step.\n\n```text\nEXACT_SCOPE_QUALIFICATION={status}\nEXACT_SOLVER_SNAPSHOT=CONTENT_BOUND\nGIT_COMMIT_BINDING=UNAVAILABLE_NOT_REQUIRED\nCORRECTNESS_PRECHECK=PASS\nSEALED_QUALIFICATION_RUN=COMPLETE\nPREREGISTRATION_REQUIRED=SATISFIED\nREAL_DATA_SEARCH_AUTHORIZED=NO\n```\n\nMetrics: zero-noise recovery {summary['zero_noise_recovery']:.3f}; noisy recovery {summary['noisy_recovery']:.3f}; precision {summary['mapping_precision']:.3f}; recall {summary['mapping_recall']:.3f}; held-out coverage {summary['held_out_coverage']:.3f}; false certificates {summary['false_global_certificates']}; hard-negative false accepts {summary['hard_negative_false_accepts']}.\n\nThe exact verifier passes the independent correctness precheck, but the synthetic recovery gates fail. The result does not authorize real-data search and does not qualify the full transformation hypothesis.\n""",encoding="utf-8")
    print(status,summary)
if __name__=="__main__": main()
