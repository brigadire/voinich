#!/usr/bin/env python3
"""Build the fail-closed Task86C-v2-scientific-impl-R audit closure."""
from __future__ import annotations

import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
T85 = ROOT.parent / "task85c"


def tsv(name, header, rows):
    with (ROOT / name).open("w", encoding="utf-8", newline="") as f:
        w = csv.writer(f, delimiter="\t", lineterminator="\n")
        w.writerow(header)
        w.writerows(rows)


blocked = "BLOCKED_CONTRACT_DEFECT_A17"
requirements = [
 "Task85c hash closure", "Task85c golden reference", "canonicalization", "numerical profile", "RNG", "M0", "M1", "M2",
 "M3 exact", "M3 approximate", "M3 aggregation", "M4", "M5", "selection", "B1", "B2",
 "PM1", "PM2", "PM4", "PM5", "PM6", "predictive aggregation", "generation", "F2 integration",
 "structural thresholds", "structural aggregation", "complexity", "minimality", "NONE",
 "NOT_IDENTIFIABLE", "status semantics", "reachability", "evidence schemas", "evidence hashing",
 "evidence-only verification", "synthetic generator independence", "blind construction", "escrow isolation",
 "natural preprocessing", "input closure", "threshold freeze", "DAG exact counts", "JobID",
 "manifest blindness", "golden suite", "local E2E", "distributed E2E", "retry determinism",
 "coordinator restart", "duplicate handling", "conflict handling", "capacity qualification",
 "production executable freeze", "operator handoff", "blind firewall", "natural confirmatory firewall",
 "Voynich firewall"]
passed = {"Task85c hash closure", "Task85c golden reference", "blind firewall", "natural confirmatory firewall", "Voynich firewall"}
tsv("TASK86C_V2_SCIENTIFIC_IMPL_R_VALIDATION.tsv", ["check_id", "requirement", "status", "evidence", "notes"],
    [[f"V{i:02d}", x, "PASS" if x in passed else ("FAIL" if x in {"status semantics", "evidence schemas"} else "NOT_TESTED"),
      "TASK86C_V2_TASK85C_CONTRACT_DEFECT.md" if x not in passed else ("Task85c hashes/validators" if "Task85c" in x else "no prohibited execution performed"),
      "mandatory stop at A17" if x not in passed else ""] for i, x in enumerate(requirements, 1)])

trace_rows=[]
for x in requirements:
    trace_rows.append([x, "research/phase3/task85c", "NOT_IMPLEMENTED", "reproduce_contract_defect.py" if x in {"status semantics", "evidence schemas"} else "NOT_RUN", "Task85c suite" if x == "golden suite" else "NONE", "Task85c schemas" if "evidence" in x else "N/A", "CONTRACT_DEFECT" if x in {"status semantics", "evidence schemas"} else blocked, "stop rule section 5"])
tsv("G1V2_SCIENTIFIC_IMPLEMENTATION_TRACEABILITY.tsv",
    ["contract_requirement", "normative_source", "implementation", "tests", "golden_vectors", "evidence_schema", "status", "notes"], trace_rows)

stages=["FIT","PREDICTIVE","GENERATION","F2_METRIC","COMPLEXITY","CANDIDATE_AGGREGATION","CONTROL_AGGREGATION"]
tsv("G1V2_SCIENTIFIC_HANDLER_REGISTRY.tsv",
    ["handler_id", "stage", "contract_operation", "implementation", "input_schema", "output_schema", "RNG_domains", "deterministic_boundary", "tests", "status"],
    [[f"BLOCKED_{s}",s,"frozen operation","NOT_IMPLEMENTED","Task85c","Task85c","NONE","NOT_ESTABLISHED","NOT_RUN",blocked] for s in stages])

with (T85 / "G1V2_CANDIDATE_REGISTRY.tsv").open(encoding="utf-8", newline="") as f:
    candidates=list(csv.DictReader(f, delimiter="\t"))
tsv("G1V2_MODEL_IMPLEMENTATION_REGISTRY.tsv",
    ["candidate_id", "model_class", "route", "frozen_hyperparameters", "production_implementation", "tests", "status", "blocker"],
    [[r["candidate_id"],r["model_class"],r["route"],r["hyperparameters"],"NOT_IMPLEMENTED","NOT_RUN",blocked,"A17 status/schema contradiction"] for r in candidates])

tsv("TASK86C_V2_INPUT_MANIFEST.tsv", ["logical_id", "role", "source", "bytes", "sha256", "preprocessing_identity", "control_category", "content_addressed_location", "status"],
    [["NONE","NOT_PRODUCED","contract stop",0,"NONE","NONE","NONE","NONE",blocked]])
tsv("TASK86C_V2_PRESEED_MANIFEST.tsv", ["sha256", "bytes", "source_content_addressed_path", "destination_cache_identity", "status"],
    [["NONE",0,"NONE","NONE",blocked]])
tsv("TASK86C_V2_RUN_MANIFEST_AUDIT.tsv", ["check", "expected", "observed", "status", "evidence"], [
    ["contract hash","275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca","same","PASS","sha256sum"],
    ["candidate closure",43,43,"PASS","Task85c registry"],
    ["control/threshold/input/blind closure","complete","not constructed","NOT_TESTED","mandatory stop"],
    ["JobIDs/DAG counts/dependencies","1321152/2617152","not materialized","NOT_TESTED","mandatory stop"],
    ["handlers/schemas/reachability","valid","A17 contradiction","FAIL","contract defect report"],
    ["expected terminal cells",8256,"not materialized","NOT_TESTED","mandatory stop"]])

tsv("TASK86C_V2_CAPACITY_BENCHMARK.tsv", ["handler_class", "count", "median_wall_s", "p90_wall_s", "p99_wall_s", "cpu_s", "peak_rss_bytes", "evidence_bytes", "temporary_bytes", "cache_behavior", "status"],
    [["NONE",0,"NA","NA","NA","NA","NA","NA","NA","NOT_MEASURED",blocked]])

(ROOT / "TASK86C_V2_CAPACITY_MODEL.md").write_text("""# Task86C-v2 capacity model\n\nNo capacity projection is scientifically valid: the mandatory A17 contract-defect stop occurred before production handlers existed. `TASK86C_V2_CAPACITY_QUALIFIED = NOT_SUPPORTED`. This is not evidence that the frozen DAG is computationally impossible.\n""", encoding="utf-8")
(ROOT / "TASK86C_V2_CLUSTER_REQUIREMENTS.md").write_text("""# Task86C-v2 cluster requirements\n\nProduction requirements were not derived. No worker slots, RAM, evidence storage, transient storage, coordinator resources, or network projection may be frozen until the contract is repaired and real handlers are benchmarked.\n""", encoding="utf-8")
(ROOT / "TASK86C_V2_OPERATOR_HANDOFF.md").write_text("""# Task86C-v2 operator handoff — blocked\n\nDo not start Task86C-v2. There is no production executable, input closure, threshold artifact, preseed closure, or production run manifest. Resolve the frozen A17 defect documented in `TASK86C_V2_TASK85C_CONTRACT_DEFECT.md`, issue a new contract identity, and rerun scientific-impl-R. The historical prep executable is not a substitute.\n""", encoding="utf-8")

# The failure manifest binds every locally produced audit artifact, not absent
# production objects. It excludes itself to avoid a circular hash.
artifacts=[]
for p in sorted(ROOT.rglob("*")):
    if p.is_file() and p.name != "TASK86C_V2_SCIENTIFIC_IMPL_R_RESULTS_MANIFEST.json" and "__pycache__" not in p.parts:
        artifacts.append({"path":p.relative_to(ROOT).as_posix(),"bytes":p.stat().st_size,"sha256":hashlib.sha256(p.read_bytes()).hexdigest()})
manifest={
 "schema":"task86c-v2-scientific-impl-r-results-v1",
 "contract_version":"G1_V2_EXECUTABLE_CONTRACT_V1",
 "task85c_artifact_root_sha256":"273913473e3e37d6a776c79b0eb214753a90e9dbaf5d78e186dcb65a0c32c351",
 "contract_sha256":"275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca",
 "golden_suite_root_sha256":"b7443a962a82dd5c0cd67b71e24d8acea73fc9be4863fca4078bc53e468c7e51",
 "evidence_schema_root_sha256":"8462ceb7f34efce1674528af7e69bdbf6855cd4938494d6d5034247245235ed0",
 "production_executable_sha256":None,"threshold_root_sha256":None,"input_root_sha256":None,
 "blind_corpus_root_sha256":None,"escrow_sha256":None,"natural_corpus_root_sha256":None,
 "production_run_manifest_sha256":None,"preseed_manifest_sha256":None,
 "absence_reason":"CONTRACT_DEFECT_A17_STATUS_SCHEMA_CONTRADICTION",
 "artifacts":artifacts,
 "artifact_root_sha256":hashlib.sha256((json.dumps(artifacts,sort_keys=True,separators=(",",":"))+"\n").encode()).hexdigest(),
 "task86c_v2_scientific_execution_ready":"NOT_SUPPORTED",
 "terminal_marker":"TASK86C_V2_TASK85C_CONTRACT_DEFECT"}
(ROOT / "TASK86C_V2_SCIENTIFIC_IMPL_R_RESULTS_MANIFEST.json").write_text(json.dumps(manifest,sort_keys=True,separators=(",",":"))+"\n",encoding="utf-8")
print("FAILURE_CLOSURE_BUILT=TASK86C_V2_TASK85C_CONTRACT_DEFECT")
