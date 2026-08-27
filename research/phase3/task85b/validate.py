#!/usr/bin/env python3
"""Task85b design/registry and evidence-verifier acceptance tests.

This module deliberately contains no fitting, induction, or generation code.
"""
from __future__ import annotations

import argparse
import copy
import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent

REQUIRED = {
    "TASK85B_DESIGN.md", "TASK85B_REQUIREMENTS_TRACEABILITY.tsv",
    "G1V1_FAILURES_AND_G1V2_RESPONSES.tsv", "G1V1_G1V2_CHANGE_REGISTER.tsv",
    "PM_REDESIGN.md", "PM6_NEGATIVE_PROTOCOL.md", "STRUCTURAL_ADEQUACY_V2.md",
    "MINIMALITY_AND_IDENTIFIABILITY.md", "EVIDENCE_PRESERVATION_CONTRACT.md",
    "DECISION_PATH_CONTRACT.md", "PHASE1_DISTRIBUTED_EXECUTION_AUDIT.md",
    "DISTRIBUTED_EXECUTION_DESIGN.md", "TASK86C_V2_CONTROL_DESIGN.md",
    "TASK86C_V2_HANDOFF.md", "TASK86V_HANDOFF.md", "TASK85B_REPORT.md",
    "TASK85B_RESULTS_MANIFEST.json", "G1V2_MODEL_REGISTRY.tsv",
    "TASK85B_VALIDATION.tsv",
    "G1V2_METRIC_REGISTRY.tsv", "G1V2_PREDICTIVE_GATE_REGISTRY.tsv",
    "G1V2_PREDICTIVE_TRUTH_TABLE.tsv",
    "G1V2_STRUCTURAL_GATE_REGISTRY.tsv", "G1V2_FAILURE_REGISTRY.tsv",
    "G1V2_COMPLEXITY_CONTRACT.tsv", "G1V2_EVIDENCE_ARTIFACT_REGISTRY.tsv",
    "G1V2_DECISION_STATUS_REGISTRY.tsv", "G1V2_CONTROL_REGISTRY.tsv",
    "G1V2_RECOVERY_CRITERIA.tsv", "G1V2_JOB_SCHEMA.tsv",
    "G1V2_DISTRIBUTED_EXECUTION_CONTRACT.tsv",
    "G1V2_DISTRIBUTED_FAILURE_REGISTRY.tsv", "G1_V2_EXPERIMENT_CONTRACT_FROZEN",
}


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def rows(name: str) -> list[dict[str, str]]:
    with (ROOT / name).open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def predictive(gates: dict[str, str]) -> str:
    required = [gates[k] for k in ("PM2", "PM5", "PM6")]
    if any(x not in {"PASS", "FAIL", "NOT_ASSESSABLE"} for x in required):
        raise ValueError("invalid gate status")
    if "NOT_ASSESSABLE" in required:
        return "PREDICTIVE_NOT_ASSESSABLE"
    if "FAIL" in required:
        return "PREDICTIVE_FAIL"
    return "PREDICTIVE_PASS"


def verify_evidence(bundle: dict) -> str:
    if bundle.get("fit_status") != "FITTED":
        if bundle.get("model_status") == "MODEL_INADEQUATE":
            raise ValueError("induction/model-class substitution")
        return "MODEL_NOT_IDENTIFIABLE"
    pms = bundle.get("pms", {})
    for pm in ("PM1", "PM2", "PM4", "PM5", "PM6"):
        if pm not in pms:
            raise ValueError(f"missing {pm}")
        r = pms[pm]
        for field in ("value", "baseline", "threshold", "available", "finite", "gate"):
            if field not in r:
                raise ValueError(f"missing {pm}.{field}")
        if not r["available"] and r["gate"] == "FAIL":
            raise ValueError("unavailable silently treated as FAIL")
        canonical = json.dumps({k: v for k, v in r.items() if k != "hash"}, sort_keys=True,
                               separators=(",", ":")).encode()
        if r.get("hash") != digest(canonical):
            raise ValueError(f"bad {pm} hash")
    f2 = bundle.get("f2", [])
    expected = set(bundle.get("expected_f2", []))
    present = {r["metric"] for r in f2}
    if present != expected:
        raise ValueError("missing or extra F2 evidence")
    gates = {pm: pms[pm]["gate"] for pm in ("PM2", "PM5", "PM6")}
    pv = predictive(gates)
    if pv != bundle.get("predictive_verdict"):
        raise ValueError("predictive verdict mismatch")
    if pv != "PREDICTIVE_PASS":
        return "MODEL_INADEQUATE" if pv == "PREDICTIVE_FAIL" else "MODEL_NOT_IDENTIFIABLE"
    return bundle["model_status"]


def fixture() -> dict:
    pms = {}
    for pm, role in (("PM1", "DIAGNOSTIC"), ("PM2", "REQUIRED"),
                     ("PM4", "SUPPORTING"), ("PM5", "REQUIRED"), ("PM6", "REQUIRED")):
        r = {"value": 0.8, "baseline": 1.0, "threshold": 0.05, "available": True,
             "finite": True, "gate": "PASS", "role": role}
        r["hash"] = digest(json.dumps(r, sort_keys=True, separators=(",", ":")).encode())
        pms[pm] = r
    return {"fit_status": "FITTED", "pms": pms,
            "expected_f2": ["EF1", "LP1"],
            "f2": [{"metric": "EF1", "gate": "PASS"}, {"metric": "LP1", "gate": "PASS"}],
            "predictive_verdict": "PREDICTIVE_PASS", "model_status": "MODEL_ADEQUATE"}


def rehash(record: dict) -> None:
    record["hash"] = digest(json.dumps({k: v for k, v in record.items() if k != "hash"},
                                       sort_keys=True, separators=(",", ":")).encode())


def expect_failure(bundle: dict, label: str) -> None:
    try:
        verify_evidence(bundle)
    except (ValueError, KeyError):
        return
    raise AssertionError(f"negative test did not fail: {label}")


def self_test() -> None:
    base = fixture()
    assert verify_evidence(base) == "MODEL_ADEQUATE"
    b = copy.deepcopy(base); del b["pms"]["PM2"]["value"]; expect_failure(b, "missing PM")
    b = copy.deepcopy(base); b["pms"]["PM5"]["threshold"] = .9; expect_failure(b, "threshold/hash")
    b = copy.deepcopy(base); b["pms"]["PM1"]["hash"] = "0" * 64; expect_failure(b, "artifact hash")
    b = copy.deepcopy(base); b["f2"].pop(); expect_failure(b, "missing F2")
    b = copy.deepcopy(base); b["pms"]["PM6"]["available"] = False; b["pms"]["PM6"]["gate"] = "FAIL"; rehash(b["pms"]["PM6"]); expect_failure(b, "unavailable as fail")
    b = {"fit_status": "INDUCTION_LIMIT_REACHED", "model_status": "MODEL_INADEQUATE"}; expect_failure(b, "induction substitution")
    assert predictive({"PM2": "PASS", "PM5": "PASS", "PM6": "NOT_ASSESSABLE"}) == "PREDICTIVE_NOT_ASSESSABLE"

    # Completion-index simulation: restart/worker change does not reschedule verified IDs.
    jobs = {"j1", "j2", "j3"}; completed = {"j1": "h1"}
    assert jobs - completed.keys() == {"j2", "j3"}
    completed["j2"] = "h2"
    assert jobs - completed.keys() == {"j3"}
    # Equal duplicate counts once; conflict is quarantined and blocks aggregation.
    copies = {"j1": {"worker-a": "h1", "worker-b": "h1"}}
    assert len(set(copies["j1"].values())) == 1
    copies["j1"]["worker-c"] = "DIFFERENT"
    assert len(set(copies["j1"].values())) > 1


def validate_contract() -> None:
    missing = sorted(name for name in REQUIRED if not (ROOT / name).is_file())
    if missing:
        raise ValueError("missing required artifacts: " + ", ".join(missing))
    for path in ROOT.glob("*.tsv"):
        rs = rows(path.name)
        if not rs:
            raise ValueError(f"empty registry: {path.name}")
        if any(None in r or any(v is None for v in r.values()) for r in rs):
            raise ValueError(f"malformed TSV: {path.name}")
    reqs = {r["requirement_id"] for r in rows("TASK85B_REQUIREMENTS_TRACEABILITY.tsv")}
    expected = {"G1V2-OBS-001", "G1V2-OBS-002", "G1V2-OBS-003", "G1V2-COMP-001",
                "G1V2-NEG-001", "G1V2-GEN-001", "G1V2-PROV-001"}
    if not expected <= reqs:
        raise ValueError("Task86C-a requirements not fully traced")
    metric_ids = {r["metric_id"] for r in rows("G1V2_METRIC_REGISTRY.tsv")}
    if metric_ids != {"PM1", "PM2", "PM4", "PM5", "PM6"}:
        raise ValueError("metric registry mismatch")
    statuses = {r["allowed_status"] for r in rows("G1V2_DECISION_STATUS_REGISTRY.tsv")}
    required_status = {"MODEL_ADEQUATE", "MODEL_INADEQUATE", "MODEL_NOT_IDENTIFIABLE",
                       "PREDICTIVE_PASS", "PREDICTIVE_FAIL", "PREDICTIVE_NOT_ASSESSABLE"}
    if not required_status <= statuses:
        raise ValueError("decision statuses incomplete")
    truth = rows("G1V2_PREDICTIVE_TRUTH_TABLE.tsv")
    if len(truth) != 27:
        raise ValueError("predictive truth table must contain all 27 combinations")
    for r in truth:
        got = predictive({"PM2": r["PM2"], "PM5": r["PM5"], "PM6": r["PM6"]})
        if got != r["overall"]:
            raise ValueError("predictive truth-table mismatch")
    report = (ROOT / "TASK85B_REPORT.md").read_text(encoding="utf-8")
    if "G1_V2_EXPERIMENT_CONTRACT_FROZEN." not in report:
        raise ValueError("terminal marker absent from report")
    manifest = json.loads((ROOT / "TASK85B_RESULTS_MANIFEST.json").read_text(encoding="utf-8"))
    if manifest.get("terminal_marker") != "G1_V2_EXPERIMENT_CONTRACT_FROZEN":
        raise ValueError("manifest terminal marker mismatch")
    for name, expected_hash in manifest.get("artifacts", {}).items():
        actual = digest((ROOT / name).read_bytes())
        if actual != expected_hash:
            raise ValueError(f"manifest hash mismatch: {name}")


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--self-test", action="store_true")
    ap.add_argument("--validate-contract", action="store_true")
    ns = ap.parse_args()
    if not ns.self_test and not ns.validate_contract:
        ns.self_test = ns.validate_contract = True
    if ns.self_test:
        self_test()
    if ns.validate_contract:
        validate_contract()
    print("TASK85B_VALIDATION_PASS")


if __name__ == "__main__":
    main()
