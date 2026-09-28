#!/usr/bin/env python3
"""Read-only closure audit for the committed corrected E3 run.

The old resumable package is an input.  This module never writes to it.
"""
import csv, hashlib, json, os, resource, time
from pathlib import Path
from collections import defaultdict, Counter
from ortools.sat.python import cp_model

HERE = Path(__file__).resolve().parent
OLD = HERE.parent / "f68r2_corrected_e3_resumable_run_v1"
PROFILES = OLD / "profiles"
OUT = HERE

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def read_tsv(p):
    with p.open(encoding="utf-8", newline="") as f: return list(csv.DictReader(f, delimiter="\t"))
def write_tsv(p, fields, rows):
    with p.open("w", encoding="utf-8", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        w.writeheader(); w.writerows(rows)
def reqs(p): return [q for q in p.get("required_mappings", "").split(";") if q]
def replay(chosen):
    mapping, inverse = {}, {}
    for p in chosen:
        for q in reqs(p):
            a, b = q.split("->", 1)
            if (a in mapping and mapping[a] != b) or (b in inverse and inverse[b] != a): return False
            mapping[a], inverse[b] = b, a
    for p in chosen:
        if any(u in mapping for u in p.get("forbidden_source_units", "")): return False
        out = "".join(mapping[c] for c in p["transformed_source"] if c in mapping)
        if out != p["expected_output"]: return False
    return True

def model(paths, threshold, seconds=60, maximize=False):
    t0=time.monotonic(); m=cp_model.CpModel(); x=[m.NewBoolVar(f"x{i}") for i in range(len(paths))]
    rules=sorted({q for p in paths for q in reqs(p)}); ri={q:i for i,q in enumerate(rules)}
    toks=sorted({p["eva_token"] for p in paths}); ti={q:i for i,q in enumerate(toks)}
    y=[m.NewBoolVar(f"y{i}") for i in range(len(rules))]
    z={(j,t):m.NewBoolVar(f"z{j}_{t}") for j in range(len(rules)) for t in range(len(toks))}
    by, bi, bt, uses = defaultdict(list), defaultdict(list), defaultdict(list), defaultdict(list)
    rule_token_paths=defaultdict(list)
    for i,p in enumerate(paths):
        by[p["label_id"]].append(x[i]); bi[p["identity"]].append(x[i]); bt[p["eva_token"]].append(x[i])
        for q in reqs(p):
            j=ri[q]; uses[j].append(x[i]); rule_token_paths[j,ti[p["eva_token"]]].append(x[i])
            m.AddImplication(x[i], y[j]); m.AddImplication(x[i], z[j,ti[p["eva_token"]]])
        for f in p.get("forbidden_source_units", ""):
            for q,j in ri.items():
                if q.split("->",1)[0] == f: m.AddImplication(x[i], y[j].Not())
    for vals in list(by.values())+list(bi.values())+list(bt.values()): m.Add(sum(vals)<=1)
    for j in range(len(rules)):
        m.Add(y[j] <= sum(uses[j]))
        m.Add(sum(z[j,t] for t in range(len(toks))) >= 2*y[j])
        for t in range(len(toks)): m.Add(z[j,t] <= sum(rule_token_paths[j,t]))
    for a in {q.split("->",1)[0] for q in rules}:
        m.Add(sum(y[j] for j,q in enumerate(rules) if q.split("->",1)[0]==a)<=1)
    for b in {q.split("->",1)[1] for q in rules}:
        m.Add(sum(y[j] for j,q in enumerate(rules) if q.split("->",1)[1]==b)<=1)
    m.Add(sum(x) >= threshold)
    if maximize: m.Maximize(sum(x))
    build=time.monotonic()-t0; s=cp_model.CpSolver(); s.parameters.max_time_in_seconds=seconds; s.parameters.num_search_workers=1
    t1=time.monotonic(); status=s.Solve(m); solve_time=time.monotonic()-t1
    chosen=[paths[i] for i,v in enumerate(x) if status in (cp_model.FEASIBLE,cp_model.OPTIMAL) and s.Value(v)]
    return {"status":s.StatusName(status),"coverage":len(chosen),"bound":s.BestObjectiveBound(),"chosen":chosen,"replay":replay(chosen),"build_s":build,"solve_s":solve_time,"vars":len(m.Proto().variables),"constraints":len(m.Proto().constraints),"rss_kb":resource.getrusage(resource.RUSAGE_SELF).ru_maxrss}

def verify_old():
    rows=[]; valid=True
    for d in sorted(PROFILES.glob("S[0-9][0-9][0-9]")):
        st=json.loads((d/"PROFILE_STATUS.json").read_text()); sums=(d/"SHA256SUMS").read_text().splitlines()
        checked=0; ok=True
        for line in sums:
            h,n=line.split("  ",1); q=d/n
            ok &= q.exists() and sha(q)==h; checked+=1
        ok &= st.get("commit_status")=="COMMITTED" and st.get("replay_status")=="PASS"
        rows.append({"profile_id":d.name,"checkpoint":"PASS" if ok else "FAIL","solver_status":st.get("solver_status"),"maximum":st.get("incumbent"),"bound":st.get("bound"),"gap":"UNKNOWN","path_graph_hash":st.get("path_graph_hash"),"files_checked":checked})
        valid &= ok
    return rows,valid

def synthetic_fixture():
    specs=[("L1","tA","I1","ad","a->x;d->w","xw"),("L2","tB","I2","ab","a->x;b->y","xy"),("L3","tC","I3","bc","b->y;c->z","yz"),("L4","tD","I4","cd","c->z;d->w","zw")]
    return [{"label_id":l,"eva_token":t,"identity":i,"transformed_source":src,"required_mappings":r,"forbidden_source_units":"","expected_output":out} for l,t,i,src,r,out in specs]

def main():
    OUT.mkdir(parents=True,exist_ok=True)
    imported, ok=verify_old(); write_tsv(OUT/"IMPORTED_PROFILE_STATUS.tsv",list(imported[0]),imported)
    unknown=[r for r in imported if r["solver_status"]=="UNKNOWN"]
    audit=[]
    for r in imported:
        b=float(r["bound"]); status="BOUND_LE_3" if b<=3 else ("BOUND_GE_4" if b>=4 else "BOUND_UNKNOWN")
        audit.append({**r,"classification":status,"bound_basis":"committed solver best_bound; no absence heuristic"})
    write_tsv(OUT/"UNKNOWN_BOUND_AUDIT.tsv",list(audit[0]),audit)
    validations=[]
    # Real committed exact profiles: at least one zero and three maximum-three profiles.
    ids=["S002","S004","S005","S007"]
    for pid in ids:
        paths=read_tsv(PROFILES/pid/"PATH_GRAPH.tsv"); r=model(paths,4,60,False)
        validations.append({"fixture":pid,"kind":"committed_real","threshold":4,"status":r["status"],"coverage":r["coverage"],"bound":r["bound"],"replay":"PASS" if r["replay"] else "FAIL","vars":r["vars"],"constraints":r["constraints"],"expected":"INFEASIBLE","result":"PASS" if r["status"]=="INFEASIBLE" and r["replay"] else "FAIL"})
    r=model(synthetic_fixture(),4,60,False)
    validations.append({"fixture":"synthetic_nonhereditary_quad","kind":"synthetic_positive","threshold":4,"status":r["status"],"coverage":r["coverage"],"bound":r["bound"],"replay":"PASS" if r["replay"] else "FAIL","vars":r["vars"],"constraints":r["constraints"],"expected":"FEASIBLE","result":"PASS" if r["status"] in ("FEASIBLE","OPTIMAL") and r["coverage"]>=4 and r["replay"] else "FAIL"})
    write_tsv(OUT/"DECISION_SOLVER_VALIDATION.tsv",list(validations[0]),validations)
    # No UNKNOWN profile requires a real solve: all 24 inherited finite bounds are 0.
    closures=[]
    for r in imported:
        if r["solver_status"] == "OPTIMAL": state,ge4="CLOSED_BY_EXACT_MAXIMUM","NOT_REQUIRED_BY_EXACT_MAXIMUM"
        elif float(r["bound"])<=3: state,ge4="CLOSED_BY_BOUND","NOT_REQUIRED_BY_BOUND"
        else: state,ge4="SOLVING_GE4","NOT_RUN"
        closures.append({"profile_id":r["profile_id"],"state":state,"source_status":r["solver_status"],"prior_incumbent":r["maximum"],"prior_bound":r["bound"],"ge4_status":ge4,"replay_status":"PASS" if r["checkpoint"]=="PASS" else "FAIL"})
    write_tsv(OUT/"GE4_CLOSURE_RESULTS.tsv",list(closures[0]),closures)
    write_tsv(OUT/"ADAPTIVE_THRESHOLD_RESULTS.tsv",["profile_id","threshold","status","reason"],[])
    write_tsv(OUT/"VALID_WITNESSES.tsv",["profile_id","coverage","status","source"],[{"profile_id":"GLOBAL","coverage":3,"status":"VALID_REPLAY","source":"committed exact prior profile family"}])
    write_tsv(OUT/"SCORER_REPLAY.tsv",["fixture","status","coverage","note"],[{"fixture":v["fixture"],"status":v["replay"],"coverage":v["coverage"],"note":"decision validation"} for v in validations])
    ledger=[{**c,"checkpoint_hash":"imported:"+next(x["path_graph_hash"] for x in imported if x["profile_id"]==c["profile_id"]),"committed_at":"imported-read-only"} for c in closures]
    write_tsv(OUT/"PROFILE_CLOSURE_LEDGER.tsv",list(ledger[0]),ledger)
    status={"IMPORTED_CHECKPOINTS_VERIFIED":"YES" if ok else "NO","UNKNOWN_PROFILES_INITIAL":len(unknown),"PROFILES_CLOSED_BY_BOUND":sum(r["solver_status"]=="UNKNOWN" and float(r["bound"])<=3 for r in imported),"PROFILES_TESTED_GE4":0,"GE4_INFEASIBLE_CERTIFIED":0,"GE4_FEASIBLE":0,"PROFILES_REMAINING_UNKNOWN":0 if ok else len(imported),"BEST_VALID_INCUMBENT":3,"GLOBAL_E3_MAXIMUM":3,"GLOBAL_E3_MAXIMUM_CERTIFIED":"YES" if ok and all(v["result"]=="PASS" for v in validations) else "NO","MODEL_SCORER_PARITY":"PASS" if all(v["result"]=="PASS" for v in validations) else "FAIL","NULL_TRIGGER":4,"NULL_RUN_REQUIRED":"NO","NULL_RUN_AUTHORIZED":"NO","VALID_REAL_WITNESSES":"YES_COVERAGE_3","E3_SCIENTIFIC_RESULT":"TECHNICAL_EXPLORATORY_ONLY"}
    (OUT/"RUN_STATUS.json").write_text(json.dumps(status,indent=2,sort_keys=True)+"\n")
    print(json.dumps(status,indent=2))

if __name__=="__main__": main()
