#!/usr/bin/env python3
"""Fail-closed validator for the Task85c executable contract."""
from __future__ import annotations

import csv
import hashlib
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
EXPECTED_SCHEMAS = {
    "fit", "fitted_model", "predictive_metric", "predictive_gate",
    "predictive_verdict", "generation", "f2_metric", "structural_family",
    "structural_gate", "structural_verdict", "complexity", "minimality",
    "final_verdict", "not_reached", "scientific_failure",
}
EXPECTED_GOLDEN = {
    "RNG", "preprocessing", "M0 fit/score/generate", "M1 fit/score/generate",
    "M2 induction/score/generate", "M3 exact", "M3 approximate",
    "M4 EM/restart/generation", "M5 backoff/generation",
    "predictive baselines", "PM5", "PM6 ordinary", "PM6 saturated",
    "structural threshold derivation", "structural aggregation",
    "complexity M0-M5", "minimality", "NONE", "NOT_IDENTIFIABLE",
    "reachability", "canonical evidence", "JobID/DAG expansion",
}


def rows(name):
    with (ROOT / name).open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def require(ok, message):
    if not ok:
        raise AssertionError(message)


def main():
    closure = rows("TASK85C_AMBIGUITY_CLOSURE.tsv")
    require([r["ambiguity_id"] for r in closure] == [f"A{i:02d}" for i in range(1, 18)], "A01-A17 exact order")
    require(all(r["verdict"] == "CLOSED" and r["residual_ambiguity"] == "NONE" for r in closure), "ambiguities not closed")

    cand = rows("G1V2_CANDIDATE_REGISTRY.tsv")
    ids = [r["candidate_id"] for r in cand]
    require(len(cand) == 43 and len(ids) == len(set(ids)), "candidate registry must contain 43 unique rows")
    require(set(r["model_class"] for r in cand) == {f"M{i}" for i in range(6)}, "M0-M5 coverage")
    for r in cand:
        hp = json.loads(r["hyperparameters"])
        require(isinstance(hp, dict) and hp, f"empty hyperparameters {r['candidate_id']}")

    domains = rows("G1V2_RNG_DOMAIN_REGISTRY.tsv")
    require(len({r["domain_id"] for r in domains}) == len(domains), "duplicate RNG domain")
    require(len({r["namespace"] for r in domains}) == len(domains), "duplicate RNG namespace")
    required_domains = {"FIT", "SELECT", "M4_RESTART", "GENERATE", "PM_BOOTSTRAP", "PM_PERMUTATION", "PM6_COMPLEMENT", "STRUCT_BOOTSTRAP", "CONTROL_GENERATE", "CORPUS_WINDOW"}
    require(required_domains <= {r["domain_id"] for r in domains}, "missing RNG consumers")

    pred = rows("G1V2_PREDICTIVE_METRIC_REGISTRY.tsv")
    require({r["metric_id"] for r in pred} == {"PM1", "PM2", "PM4", "PM5", "PM6"}, "predictive metric set")
    require(all(all(r[k] for k in r) for r in pred), "incomplete predictive metric")
    structural = rows("G1V2_STRUCTURAL_METRIC_REGISTRY.tsv")
    require(len(structural) == 12 and sum(r["independent"] == "NO" for r in structural) == 5, "structural/skeleton registry")
    require(all(r["skeleton_policy"] == "DIAGNOSTIC_ONLY" for r in structural if r["independent"] == "NO"), "skeleton weight")
    complexity = rows("G1V2_COMPLEXITY_CONTRACT.tsv")
    require({r["model_class"] for r in complexity} == {f"M{i}" for i in range(6)}, "complexity M0-M5")

    controls = rows("G1V2_CONTROL_REGISTRY.tsv")
    require(len(controls) == 27, "12 dev generator rows + 12 blind generator rows + 3 natural rows")
    require(sum(r["role"] == "DEVELOPMENT" for r in controls) == 12, "development controls")
    require(sum(r["role"] == "BLIND_SYNTHETIC" for r in controls) == 12, "blind generators")
    generators = rows("G1V2_SYNTHETIC_GENERATOR_REGISTRY.tsv")
    require(len(generators) == 12, "two frozen generators per M0-M5")
    require(all(sum(r["model_class"] == f"M{i}" for r in generators) == 2 for i in range(6)), "generator pair coverage")
    require(len({r["generator_id"] for r in generators}) == 12 and all(len(r["spec_sha256"]) == 64 for r in generators), "generator identity")
    require({r["generator_id"] for r in generators} == {r["generator_or_source"] for r in controls if r["role"] in {"DEVELOPMENT", "BLIND_SYNTHETIC"}}, "control generator references")
    corpora = rows("G1V2_CORPUS_REGISTRY.tsv")
    require({r["language"] for r in corpora} == {"English", "Latin", "Sanskrit"}, "natural corpora")
    require(all(len(x) == 64 for r in corpora for x in r["source_sha256"].split(";")), "corpus hashes")

    reach = rows("G1V2_REACHABILITY_CONTRACT.tsv")
    pairs = {(r["upstream_stage"], r["upstream_status"], r["downstream_stage"]) for r in reach}
    expected = sum(len(n) * 4 for n in [["PREDICTIVE", "GENERATION", "COMPLEXITY"], ["GENERATION", "AGGREGATION"], ["STRUCTURAL", "AGGREGATION"], ["AGGREGATION"]])
    require(len(pairs) == expected == len(reach), "reachability table not total/unique")

    contract = json.loads((ROOT / "G1V2_EXECUTABLE_CONTRACT.json").read_text())
    dag = json.loads((ROOT / "G1V2_DAG_CONTRACT.json").read_text())
    require(contract["candidate_count"] == len(cand), "candidate count mismatch")
    c = dag["counts"]
    require(c["total_jobs"] == 192 + 8256 * 160 == 1321152, "job arithmetic")
    require(c["dependency_edges"] == 8256 * 316 + 192 * 43 == 2617152, "edge arithmetic")
    require(c["candidate_fits"] == 192 * 43 and c["structural_metric_jobs"] == 192 * 43 * 12 * 12, "DAG axes")
    forbidden = {"hostname", "worker", "coordinator", "lease", "retry", "execution_order", "wall_clock"}
    require(forbidden.isdisjoint(dag["job_id"]["payload_fields"]), "operational JobID field")

    schema_names = {p.name.removesuffix(".schema.json") for p in (ROOT / "schemas").glob("*.schema.json")}
    require(schema_names == EXPECTED_SCHEMAS, "evidence schema coverage")
    for name in EXPECTED_SCHEMAS:
        s = json.loads((ROOT / "schemas" / f"{name}.schema.json").read_text())
        require(s.get("additionalProperties") is False and "content_sha256" in s["required"], f"schema closure {name}")

    cov = rows("TASK85C_GOLDEN_COVERAGE.tsv")
    require({r["operation"] for r in cov} == EXPECTED_GOLDEN and all(r["status"] == "COVERED" for r in cov), "golden coverage")
    suite = json.loads((ROOT / "golden/G1V2_GOLDEN_SUITE.json").read_text())
    case_ops = {c["operation"] for c in suite["cases"]}
    for op in EXPECTED_GOLDEN:
        if op == "complexity M0-M5":
            require(all(f"complexity M{i}" in case_ops for i in range(6)), "complexity goldens")
        elif op == "JobID/DAG expansion":
            require({"JobID", "DAG expansion"} <= case_ops, "JobID/DAG goldens")
        else:
            require(op in case_ops or any(x.startswith(op.split()[0]) for x in case_ops), f"missing golden {op}")

    prose = (ROOT / "G1V2_EXECUTABLE_CONTRACT.md").read_text()
    require("Voynich artifact was read" not in prose, "firewall claim malformed")
    scan = "\n".join(p.read_text(errors="replace") for p in ROOT.rglob("*") if p.is_file() and p.suffix in {".md", ".tsv", ".json"})
    require("experiments/inverse-transposition/voynich-validation" not in scan, "Voynich dependency")

    manifest = json.loads((ROOT / "TASK85C_RESULTS_MANIFEST.json").read_text())
    listed = {x["path"]: x for x in manifest["artifacts"]}
    actual = {p.relative_to(ROOT).as_posix(): p for p in ROOT.rglob("*") if p.is_file() and p.name != "TASK85C_RESULTS_MANIFEST.json" and "__pycache__" not in p.parts}
    require(set(listed) == set(actual), "manifest transitive file set")
    for rel, p in actual.items():
        require(hashlib.sha256(p.read_bytes()).hexdigest() == listed[rel]["sha256"], f"manifest hash {rel}")

    print("TASK85C_VALIDATION=PASS")
    print("A01-A17=CLOSED")
    print(f"CANDIDATES={len(cand)} JOBS={c['total_jobs']} EDGES={c['dependency_edges']}")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception as exc:
        print(f"TASK85C_VALIDATION=FAIL: {exc}", file=sys.stderr)
        sys.exit(1)
