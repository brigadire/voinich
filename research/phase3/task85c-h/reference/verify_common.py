#!/usr/bin/env python3
import csv, hashlib, json, os, subprocess, sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[4]; OUT=ROOT/"research/phase3/task85c-h"
ENV=dict(os.environ); ENV["GOCACHE"]="/tmp/task85ch121-gocache"
def rows(name):
 p=OUT/name
 with p.open(encoding="utf-8") as f:return list(csv.DictReader(f,delimiter="\t"))
def check(which):
 if which=="authority": assert all(x["status"]=="PASS" for x in rows("TASK85C_H_AUTHORITY_VALIDATION.tsv"))
 elif which=="registry":
  x=json.loads((OUT/"TASK85C_H_HANDLER_REGISTRY.json").read_text()); assert len(x["candidate_handlers"])==43 and len(x["stages"])==7 and len(x["generation_paths"])==26
 elif which=="traceability": assert all(x["status"]=="PASS" for x in rows("TASK85C_H_IMPLEMENTATION_TRACEABILITY.tsv"))
 elif which=="generation":
  x=json.loads((OUT/"TASK85C_H_OPEN_SCIENTIFIC_GOLDENS.json").read_text()); assert len(x["generation_routes"])==12 and len(rows("TASK85C_H_GENERATION_COVERAGE.tsv"))==26
 elif which=="status": assert len(rows("TASK85C_H_STATUS_REACHABILITY_COVERAGE.tsv"))==21
 elif which=="jobid": subprocess.check_call(["go","test","./internal/g1v2science","-run","TestE3JobIDGolden","-count=1"],cwd=ROOT,env=ENV)
 elif which in {"model","pm","f2","complexity","aggregation"}:
  fresh=json.loads(subprocess.check_output(["go","run","./internal/g1v2science/goldencmd","-phase3","research/phase3"],cwd=ROOT,env=ENV)); frozen=json.loads((OUT/"TASK85C_H_OPEN_SCIENTIFIC_GOLDENS.json").read_text()); assert fresh==frozen
 elif which=="evidence_only": subprocess.check_call(["go","test","./internal/g1v2science","-run","TestEvidenceOnlyReconstructionAndMutations","-count=1"],cwd=ROOT,env=ENV)
 elif which=="manifest":
  m=json.loads((OUT/"TASK85C_H_RESULTS_MANIFEST.json").read_text());
  for x in m["artifacts"]: assert hashlib.sha256((ROOT/x["path"]).read_bytes()).hexdigest()==x["sha256"]
 else: raise AssertionError(which)
 print(f"{which.upper()}=PASS")
if __name__=="__main__": check(sys.argv[1])
