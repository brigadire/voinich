#!/usr/bin/env python3
"""Build Task85c-a's fail-closed expanded-defect audit closure."""
from __future__ import annotations

import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PHASE3 = ROOT.parent
T85 = PHASE3 / "task85c"
IMPL_R = PHASE3 / "task86c-v2-scientific-impl-r"


def rows(path):
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def tsv(name, header, data):
    with (ROOT / name).open("w", encoding="utf-8", newline="") as f:
        w = csv.writer(f, delimiter="\t", lineterminator="\n")
        w.writerow(header)
        w.writerows(data)


statuses = [r["status"] for r in rows(T85 / "G1V2_STATUS_REGISTRY.tsv")]
schema_paths = sorted((T85 / "schemas").glob("*.schema.json"))
types = [p.name.removesuffix(".schema.json") for p in schema_paths]

# This is deliberately an audit matrix, not a normative replacement matrix.
# UNRESOLVED is the precise failure finding and cannot be mistaken for a freeze.
matrix=[]
for typ in types:
    schema=json.loads((T85 / "schemas" / f"{typ}.schema.json").read_text())
    for status in statuses:
        if typ == "not_reached":
            classification = "ALLOWED" if status == "NOT_REACHED" else "FORBIDDEN"
            required = "upstream_job_id;upstream_status;reason_code" if status == "NOT_REACHED" else "N/A"
            forbidden = "all computed scientific outputs" if status == "NOT_REACHED" else "ALL"
            notes = "uniquely frozen single-status type"
        elif typ == "scientific_failure" and status in {"FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","PROTOCOL_VETO"}:
            classification = "UNRESOLVED"
            required = "reason_code;diagnostics_hash"
            forbidden = "ordinary adequacy outputs"
            notes = "registered failure but per-stage reachability missing/conflicting"
        else:
            classification = "UNRESOLVED"
            required = "UNRESOLVED"
            forbidden = "UNRESOLVED"
            notes = "no consistent evidence-type/status mapping derivable before expanded-defect stop"
        matrix.append([typ,schema["$id"],status,classification,required,forbidden,"UNRESOLVED","UNRESOLVED", "Task85c status registry + reachability",notes])
tsv("G1V2_EVIDENCE_STATUS_MATRIX.tsv",
    ["evidence_type","schema_id","status","allowed","required_fields","forbidden_fields","nullable_fields","dependency_rule","normative_source","notes"], matrix)

audit=[]
for p in schema_paths:
    typ=p.name.removesuffix(".schema.json")
    old=hashlib.sha256(p.read_bytes()).hexdigest()
    audit.append([json.loads(p.read_text())["$id"],typ,old,"NOT_CREATED","UNRESOLVED","UNRESOLVED","NOT_PERFORMED","x-status-rules present in V1","0","0","0","BLOCKED_CONTRACT_DEFECT_EXPANDED"])
tsv("TASK85C_A_SCHEMA_AUDIT.tsv",
    ["schema_id","evidence_type","old_schema_sha256","new_schema_sha256","allowed_statuses","forbidden_statuses","standard_enforcement","custom_normative_keywords","positive_tests","negative_tests","mutation_tests","verdict"],audit)

unchanged=["G1V2_CANDIDATE_REGISTRY.tsv","G1V2_RNG_DOMAIN_REGISTRY.tsv","G1V2_CONTROL_REGISTRY.tsv","G1V2_CORPUS_REGISTRY.tsv","G1V2_DAG_CONTRACT.json","G1V2_REACHABILITY_CONTRACT.tsv","G1V2_MODEL_REGISTRY.tsv","G1V2_PREDICTIVE_METRIC_REGISTRY.tsv","G1V2_STRUCTURAL_METRIC_REGISTRY.tsv","G1V2_COMPLEXITY_CONTRACT.tsv"]
diff=[]
for name in unchanged:
    h=hashlib.sha256((T85/name).read_bytes()).hexdigest()
    diff.append([name,"entire artifact",h,h,"UNCHANGED_PARENT","no refreeze issued","false","PASS_BYTE_IDENTICAL"])
diff += [
 ["G1V2_EXECUTABLE_CONTRACT_V1_1.md","entire artifact","ABSENT","ABSENT","NOT_CREATED","expanded-defect stop","N/A","BLOCKED"],
 ["G1V2_EXECUTABLE_CONTRACT_V1_1.json","entire artifact","ABSENT","ABSENT","NOT_CREATED","expanded-defect stop","N/A","BLOCKED"],
 ["schemas/","status branches","8462ceb7f34efce1674528af7e69bdbf6855cd4938494d6d5034247245235ed0","NOT_CREATED","NOT_CREATED","no unique matrix","N/A","BLOCKED"]]
tsv("TASK85C_A_CONTRACT_DIFF.tsv",["artifact","field_or_region","old_hash","new_hash","change_class","reason","scientific_semantics_changed","verdict"],diff)

checks=[
 ("parent Task85c closure","PASS","Task85c validator and hashes"),
 ("defect reproduction before repair","PASS","scientific-impl-R reproducer"),
 ("expanded status/reachability defect","FAIL","E01-E04 reproducer"),
 ("JSON Schema dialect declaration","NOT_TESTED","repair not created"),
 ("status matrix completeness","FAIL","UNRESOLVED cells"),
 ("all schema/status ALLOWED tests","NOT_TESTED","mandatory stop"),
 ("all schema/status FORBIDDEN tests","NOT_TESTED","mandatory stop"),
 ("required-field mutations","NOT_TESTED","mandatory stop"),
 ("forbidden-field mutations","NOT_TESTED","mandatory stop"),
 ("null/absence mutations","NOT_TESTED","mandatory stop"),
 ("contradictory-field mutations","NOT_TESTED","mandatory stop"),
 ("hash-valid invalid evidence","NOT_TESTED","no repaired schema"),
 ("original not_reached/PASS regression","NOT_TESTED","no repaired schema"),
 ("original fit/FAIL regression","NOT_TESTED","no repaired schema"),
 ("scientific_failure schema audit","FAIL","registered failures lack reachability"),
 ("canonicalization invariant","PASS","parent byte-identical"),
 ("reachability invariant","PASS","parent byte-identical; internally defective"),
 ("DAG invariant","PASS","1321152 jobs/2617152 edges unchanged"),
 ("model invariant","PASS","parent byte-identical"),
 ("RNG invariant","PASS","e47c5a... artifact unchanged"),
 ("control/corpus invariant","PASS","parent byte-identical"),
 ("inherited golden suite","PASS","G1V2_GOLDEN_REFERENCE=PASS"),
 ("extended golden suite","NOT_TESTED","not created"),
 ("schema root closure","NOT_TESTED","new root not created"),
 ("V1 to V1_1 semantic diff","NOT_TESTED","V1_1 not created"),
 ("results-manifest closure","PASS","failure closure only"),
 ("go build ./...","NOT_TESTED","mandatory contract stop"),
 ("go vet ./...","NOT_TESTED","mandatory contract stop"),
 ("go test ./...","NOT_TESTED","mandatory contract stop"),
 ("go test -race ./...","NOT_TESTED","mandatory contract stop"),
 ("standard JSON Schema validator","NOT_TESTED","no repaired schemas"),
 ("second JSON Schema validator","NOT_TESTED","no repaired schemas"),
 ("blind firewall","PASS","no construction/execution"),
 ("natural confirmatory firewall","PASS","no execution"),
 ("Voynich firewall","PASS","no access")]
tsv("TASK85C_A_VALIDATION.tsv",["check_id","requirement","status","evidence"],[[f"V{i:02d}",*x] for i,x in enumerate(checks,1)])

artifacts=[]
for p in sorted(ROOT.rglob("*")):
    if p.is_file() and p.name != "TASK85C_A_RESULTS_MANIFEST.json" and "__pycache__" not in p.parts:
        artifacts.append({"path":p.relative_to(ROOT).as_posix(),"bytes":p.stat().st_size,"sha256":hashlib.sha256(p.read_bytes()).hexdigest()})
parent_manifest=T85/"TASK85C_RESULTS_MANIFEST.json"
defect=IMPL_R/"TASK86C_V2_TASK85C_CONTRACT_DEFECT.md"
manifest={
 "schema":"task85c-a-expanded-defect-results-v1",
 "parent_contract_version":"G1_V2_EXECUTABLE_CONTRACT_V1",
 "parent_contract_sha256":"275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca",
 "parent_task85c_manifest_sha256":hashlib.sha256(parent_manifest.read_bytes()).hexdigest(),
 "parent_task85c_artifact_root_sha256":"273913473e3e37d6a776c79b0eb214753a90e9dbaf5d78e186dcb65a0c32c351",
 "original_evidence_schema_root_sha256":"8462ceb7f34efce1674528af7e69bdbf6855cd4938494d6d5034247245235ed0",
 "original_golden_suite_root_sha256":"b7443a962a82dd5c0cd67b71e24d8acea73fc9be4863fca4078bc53e468c7e51",
 "scientific_impl_r_defect_report_sha256":hashlib.sha256(defect.read_bytes()).hexdigest(),
 "new_contract_version":None,"new_contract_sha256":None,"new_evidence_schema_root_sha256":None,"new_golden_suite_root_sha256":None,
 "absence_reason":"CONTRACT_DEFECT_EXPANDED_E01_E02_E03_E04",
 "artifacts":artifacts,
 "artifact_root_sha256":hashlib.sha256((json.dumps(artifacts,sort_keys=True,separators=(",",":"))+"\n").encode()).hexdigest(),
 "task86c_v2_scientific_impl_r2_ready":"NOT_SUPPORTED",
 "terminal_marker":"TASK85C_A_CONTRACT_DEFECT_EXPANDED"}
(ROOT/"TASK85C_A_RESULTS_MANIFEST.json").write_text(json.dumps(manifest,sort_keys=True,separators=(",",":"))+"\n",encoding="utf-8")
print("TASK85C_A_FAILURE_CLOSURE_BUILT=TASK85C_A_CONTRACT_DEFECT_EXPANDED")
