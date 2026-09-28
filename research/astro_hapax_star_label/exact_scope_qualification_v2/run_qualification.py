#!/usr/bin/env python3
"""Run CP-SAT on public sealed data only; never opens SEALED_TRUTH.jsonl."""
from __future__ import annotations
import csv, hashlib, importlib, json, platform, sys, time
from pathlib import Path
OUT=Path(__file__).resolve().parent; SNAP=OUT/"snapshot"; sys.path.insert(0,str(SNAP))
from solver_cpsat import CPSATSolver
import ortools

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
    public=[json.loads(x) for x in (OUT/"SEALED_PUBLIC.jsonl").read_text().splitlines()]
    solver_hash,oracle_hash,scorer_hash=sha(SNAP/"solver_cpsat.py"),sha(SNAP/"oracle_bb.py"),sha(SNAP/"scorer.py")
    config={"time_limit_seconds":60.0,"num_workers":1,"deletion_mode":"DROP_UNMAPPED","abbreviation":"NONE"}; config_hash=hashlib.sha256(json.dumps(config,sort_keys=True).encode()).hexdigest(); rows=[]
    for x in public:
        lex={k:set(v) for k,v in x["lexicon"].items()}; labels=x["labels"]; source=tuple(sorted(set(c for fs in lex.values() for f in fs for c in f))); target=tuple("acdefhi"); size=4
        train=labels[:6]
        t=time.monotonic(); cp=CPSATSolver(source,target,size,x["mode"],x["capacity"],time_limit_sec=60.0,num_workers=1).solve(lex,train)
        rows.append({"instance_id":x["instance_id"],"group":x["group"],"seed":x["seed"],"classification":cp["status"],"cp_sat_is_optimal":cp["is_optimal"],"cp_sat_fitness":cp["fitness"],"oracle_fitness":"NOT_RUN_IN_QUALIFICATION","objective_agreement":"NA","predicted_table_json":json.dumps(cp["table"],sort_keys=True),"runtime_seconds":round(time.monotonic()-t,6),"solver_sha256":solver_hash,"oracle_sha256":oracle_hash,"scorer_sha256":scorer_hash,"config_sha256":config_hash,"ortools_version":ortools.__version__,"python_version":platform.python_version()})
    with (OUT/"SEALED_PREDICTIONS.tsv").open("w",newline="") as f:
        w=csv.DictWriter(f,fieldnames=list(rows[0]),delimiter="\t"); w.writeheader(); w.writerows(rows)
    print("wrote",len(rows),"predictions; truth was not opened")
if __name__=="__main__": main()
