#!/usr/bin/env python3
"""Repository closure validator for Task86C-v2-prep.

Contains no fitting, generation, model execution, or corpus access.
"""
from __future__ import annotations
import csv, hashlib, json, subprocess, sys
from pathlib import Path

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[2]
REQUIRED={
"TASK86C_V2_PREP_DESIGN.md","PHASE1_EXECUTOR_INTEGRATION.md","G1V2_EXECUTOR_ADAPTER.md",
"EXECUTION_COMPONENT_MANIFEST.json","WORKER_COMPATIBILITY_CONTRACT.md","ENGINEERING_FIXTURE_MANIFEST.json",
"EVIDENCE_VERIFIER_VALIDATION.tsv","NEGATIVE_VERIFIER_TESTS.tsv","DISTRIBUTED_FAILURE_TESTS.tsv",
"RESUME_TEST_RESULTS.tsv","DUPLICATE_EXECUTION_RESULTS.tsv","CONFLICT_QUARANTINE_RESULTS.tsv",
"CROSS_NODE_DETERMINISM.tsv","SCALABILITY_BENCHMARK_MANIFEST.json","SCALABILITY_RESULTS.tsv",
"JOB_RUNTIME_PROFILE.tsv","RESOURCE_PROFILE.tsv","CACHE_NETWORK_PROFILE.tsv","TASK86C_V2_CAPACITY_ESTIMATE.md",
"TASK86C_V2_CLUSTER_RECOMMENDATION.md","TASK86C_V2_PREP_VALIDATION.tsv","TASK86C_V2_SCIENTIFIC_HANDOFF.md",
"TASK86C_V2_PREP_REPORT.md","TASK86C_V2_PREP_RESULTS_MANIFEST.json","ENGINEERING_EVIDENCE_BUNDLE.tar.gz",
"ENGINEERING_RESULT_GRAPH.tsv","validate.py"}
def digest(p:Path)->str:return hashlib.sha256(p.read_bytes()).hexdigest()
def rows(p:Path):
    with p.open(encoding="utf-8",newline="") as f:return list(csv.DictReader(f,delimiter="\t"))
def main():
    missing=sorted(REQUIRED-{p.name for p in HERE.iterdir() if p.is_file()})
    if missing:raise ValueError("missing artifacts: "+", ".join(missing))
    subprocess.run([sys.executable,str(ROOT/"research/phase3/task85b/validate.py")],check=True)
    component=json.loads((HERE/"EXECUTION_COMPONENT_MANIFEST.json").read_text())
    for rel,want in component["source_hashes"].items():
        if digest(ROOT/rel)!=want:raise ValueError("source hash mismatch: "+rel)
    for name,want in component["contract_hashes"].items():
        if digest(ROOT/"research/phase3/task85b"/name)!=want:raise ValueError("registry hash mismatch: "+name)
    fixture=json.loads((HERE/"ENGINEERING_FIXTURE_MANIFEST.json").read_text())
    bench=json.loads((HERE/"SCALABILITY_BENCHMARK_MANIFEST.json").read_text())
    if fixture!=bench or len(fixture.get("jobs",[]))!=193:raise ValueError("benchmark is not the frozen fixture")
    ids=[j["job_id"] for j in fixture["jobs"]]
    if len(ids)!=len(set(ids)):raise ValueError("duplicate JobID")
    graph=rows(HERE/"ENGINEERING_RESULT_GRAPH.tsv")
    if len(graph)!=193 or {r["job_id"] for r in graph}!=set(ids):raise ValueError("evidence graph closure mismatch")
    val={r["verdict"]:r["value"] for r in rows(HERE/"TASK86C_V2_PREP_VALIDATION.tsv")}
    if len(val)!=18 or any(v!="SUPPORTED" for v in val.values()):raise ValueError("mandatory verdict not supported")
    scale={int(r["workers"]):r for r in rows(HERE/"SCALABILITY_RESULTS.tsv")}
    four=scale[4]
    if not(float(four["parallel_efficiency"])>=.60 and float(four["worker_utilization"])>=.70 and float(four["overhead_upper_bound"])<=.10 and float(four["straggler_fraction"])<=.25):raise ValueError("scalability threshold failure")
    if any(r["result"]!="PASS" for r in rows(HERE/"NEGATIVE_VERIFIER_TESTS.tsv")):raise ValueError("negative test failure")
    report=(HERE/"TASK86C_V2_PREP_REPORT.md").read_text()
    if report.count("TASK86C_V2_COMPUTE_READY_FROZEN.")!=1:raise ValueError("terminal marker mismatch")
    results=json.loads((HERE/"TASK86C_V2_PREP_RESULTS_MANIFEST.json").read_text())
    if results.get("terminal_marker")!="TASK86C_V2_COMPUTE_READY_FROZEN":raise ValueError("results terminal marker mismatch")
    for name,want in results["artifacts"].items():
        if digest(HERE/name)!=want:raise ValueError("results hash mismatch: "+name)
    print("TASK86C_V2_PREP_VALIDATION_PASS")
if __name__=="__main__":main()
