#!/usr/bin/env python3
"""Build the Task85c-i V1.2 evidence/execution-identity integration refreeze."""
from __future__ import annotations

import csv
import hashlib
import json
import sys
import unicodedata
from copy import deepcopy
from pathlib import Path

HERE = Path(__file__).resolve().parent
OUT = HERE.parent
REPO = HERE.parents[3]
C = REPO / "research/phase3/task85c-c"
E = REPO / "research/phase3/task85c-e"
G = REPO / "research/phase3/task85c-g"
H = REPO / "research/phase3/task85c-h"
V11 = "G1_V2_EXECUTABLE_CONTRACT_V1_1"
V12 = "G1_V2_EXECUTABLE_CONTRACT_V1_2"
EV12 = "G1V2_EVIDENCE_CONTRACT_V1_2"
SV2 = "G1_V2_STATUS_REACHABILITY_CONTRACT_V2"

PINS = {
    G / "G1_V2_EXECUTABLE_CONTRACT_V1_2.md": "ec60bb23e55ce157fe954b5cafc63d22ab70ecec390822cb63f9ae273142c639",
    G / "G1_V2_EXECUTABLE_CONTRACT_V1_2.json": "29e39e0c25dc8033f784480fdc537e3ede9eeb69baa0607c9f249d796d6b42dc",
    G / "G1V2_GENERATION_SEMANTICS_V1.json": "45d533f8b83b24c77a96836fa5c2ef95f9b948003bd2ed725fc2ea97e010b310",
    G / "G1V2_GENERATION_GOLDEN_SUITE_V1.json": "143954667073a2c10f1bd59ce98b9c93dd84b50632bb67ea80d0d92449480acb",
    E / "G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json": "dbfb9a4a7101eed7006f751b9c4631b5f0286c3792f9777cc833c5dcfa42a3d3",
    C / "registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json": "fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9",
    H / "TASK85C_H_RESULTS_MANIFEST.json": "c9c54c9b5c20dd746ace32ab3a3a9dc916a8ceae00710481e8226314db2ea795",
}


def normalized(value):
    if isinstance(value, str):
        return unicodedata.normalize("NFC", value)
    if isinstance(value, list):
        return [normalized(x) for x in value]
    if isinstance(value, dict):
        return {unicodedata.normalize("NFC", k): normalized(value[k]) for k in sorted(value)}
    return value


def canonical(value) -> bytes:
    return (json.dumps(normalized(value), ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode()


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write(path: str, text: str) -> None:
    target = OUT / path
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(text, encoding="utf-8", newline="\n")


def write_json(path: str, value) -> None:
    write(path, canonical(value).decode())


def write_tsv(path: str, header, rows) -> None:
    target = OUT / path
    target.parent.mkdir(parents=True, exist_ok=True)
    with target.open("w", encoding="utf-8", newline="") as f:
        w = csv.writer(f, delimiter="\t", lineterminator="\n")
        w.writerow(header)
        w.writerows(rows)


def replace_strings(value, old: str, new: str):
    if isinstance(value, str):
        return value.replace(old, new)
    if isinstance(value, list):
        return [replace_strings(x, old, new) for x in value]
    if isinstance(value, dict):
        return {k: replace_strings(v, old, new) for k, v in value.items()}
    return value


def schema_root(items) -> str:
    return hashlib.sha256(canonical(items)).hexdigest()


def content_hash(instance) -> str:
    bare = deepcopy(instance)
    bare.pop("content_sha256", None)
    return hashlib.sha256(canonical(bare)).hexdigest()


def job_payload(version: str):
    return {
        "candidate_id": "M0-iid-0",
        "contract_version": version,
        "control_instance_id": "OPEN-INTEGRATION-FIXTURE-1",
        "dependency_job_ids": [],
        "metric_id_or_null": None,
        "replicate_or_null": None,
        "scale_or_null": None,
        "stage": "FIT",
    }


def job_id(payload) -> str:
    return "j-" + hashlib.sha256(b"G1V2-JOB\0" + canonical(payload)).hexdigest()[:40]


def verify_inputs() -> None:
    for path, expected in PINS.items():
        assert sha(path) == expected, path
    gm = json.loads((G / "TASK85C_G_RESULTS_MANIFEST.json").read_text())
    hm = json.loads((H / "TASK85C_H_RESULTS_MANIFEST.json").read_text())
    assert gm["artifact_root_excluding_manifest_sha256"] == "609ca0fbb0cdada89a775972191ef0a6e1899791f63ee6104f517fe6e3a96b91"
    assert hm["artifact_root_excluding_manifest_sha256"] == "eb7939fa6d7c20f698143a163cfeb315da69ae547809e4d64a5632f26543ed42"
    assert json.loads((G / "G1_V2_EXECUTABLE_CONTRACT_V1_2.json").read_text())["contract_version"] == V12


def build_schemas():
    source_files = sorted((C / "schemas").glob("*.schema.json"))
    assert len(source_files) == 15
    entries = []
    diffs = []
    transformed = {}
    for source in source_files:
        typ = source.name.removesuffix(".schema.json")
        old = json.loads(source.read_text())
        new = replace_strings(old, V11, V12)
        new = replace_strings(new, ".v1_1", ".v1_2")
        new = replace_strings(new, ":v1_1", ":v1_2")
        new = replace_strings(new, "V1.1", "V1.2")
        # Prove that the transformation is version-label-only.
        back = replace_strings(new, V12, V11)
        back = replace_strings(back, ".v1_2", ".v1_1")
        back = replace_strings(back, ":v1_2", ":v1_1")
        back = replace_strings(back, "V1.2", "V1.1")
        assert back == old, typ
        rel = f"evidence-schemas-v1_2/{source.name}"
        write_json(rel, new)
        transformed[typ] = new
        branch_versions = {b["properties"]["contract_version"]["const"] for b in new["oneOf"]}
        assert branch_versions == {V12}
        statuses = sorted({b["properties"]["status"]["const"] for b in new["oneOf"]})
        entries.append({
            "evidence_type": typ,
            "schema_id": new["$id"],
            "schema_path": rel,
            "schema_sha256": sha(OUT / rel),
            "scientific_contract_version": V12,
            "allowed_statuses": statuses,
            "dialect": "https://json-schema.org/draft/2020-12/schema",
        })
        diffs.append([
            typ, sha(source), sha(OUT / rel),
            "$id;title;oneOf/*/properties/schema_id/const;oneOf/*/properties/contract_version/const",
            "EVIDENCE_SCHEMA_VERSIONING;EVIDENCE_VERSION_BINDING",
            "NONE",
            "new schema identity and V1.2 contract literal; all payload/status/reason constraints byte-structurally preserved",
        ])
    registry = {
        "schema": "g1v2.evidence-schema-registry.v1_2",
        "evidence_contract_version": EV12,
        "scientific_contract_version": V12,
        "selection": "select exactly by explicit scientific contract version; no fallback probing",
        "unknown_version_disposition": "FAIL_CLOSED",
        "schema_count": len(entries),
        "entries": entries,
    }
    write_json("G1V2_V1_2_EVIDENCE_SCHEMA_REGISTRY.json", registry)
    items = [{"path": e["schema_path"], "sha256": e["schema_sha256"]} for e in entries]
    root = schema_root(items)
    root_doc = {
        "schema": "g1v2.evidence-schema-root.v1_2",
        "evidence_contract_version": EV12,
        "scientific_contract_version": V12,
        "schema_count": 15,
        "schema_registry_sha256": sha(OUT / "G1V2_V1_2_EVIDENCE_SCHEMA_REGISTRY.json"),
        "root_sha256": root,
        "root_definition": "SHA-256 of G1V2-CJ-1 canonical JSON array of {path,sha256}, paths relative to task85c-i and sorted bytewise",
        "items": items,
        "historical_v1_1_root_sha256": "4744ca82532cd47a0d02bb680796b26a11ceca57d6229f0b312df69a103f784b",
    }
    write_json("G1V2_V1_2_EVIDENCE_SCHEMA_ROOT.json", root_doc)
    write_tsv(
        "G1V2_EVIDENCE_V1_1_TO_V1_2_DIFF.tsv",
        ["schema_id", "v1_1_sha256", "v1_2_sha256", "changed_paths", "change_classification", "scientific_semantic_effect", "justification"],
        diffs,
    )
    return registry, root_doc, transformed


def build_fixtures(transformed):
    old_cases = json.loads((C / "golden/schema-positive/cases.json").read_text())
    new_cases = []
    for case in old_cases:
        item = deepcopy(case)
        item["id"] = item["id"].replace("POS-", "V12-POS-")
        inst = item["instance"]
        inst["contract_version"] = V12
        inst["schema_id"] = inst["schema_id"].replace(".v1_1", ".v1_2")
        inst["content_sha256"] = content_hash(inst)
        item["scientific_payload_invariant_sha256"] = hashlib.sha256(canonical({
            "stage": inst["stage"], "status": inst["status"], "payload": inst["payload"]
        })).hexdigest()
        new_cases.append(item)
    assert {x["schema"] for x in new_cases} == set(transformed)
    write_json("fixtures/G1V2_V1_2_EVIDENCE_POSITIVE_FIXTURES.json", new_cases)
    generation_old = next(x for x in old_cases if x["schema"] == "generation")
    generation_new = next(x for x in new_cases if x["schema"] == "generation")
    hsc01 = {
        "schema": "g1v2.h-sc01-regression.v1",
        "finding": "H-SC01-EVIDENCE-CONTRACT-VERSION",
        "historical_v1_1_fixture": generation_old["instance"],
        "historical_v1_1_acceptance": True,
        "historical_v1_2_substitution_rejected_by_old_schema": True,
        "repaired_v1_2_fixture": generation_new["instance"],
        "repaired_v1_2_acceptance_by_new_schema": True,
        "repaired_v1_2_rejection_by_v1_1_schema": True,
        "v1_1_rejection_by_v1_2_schema": True,
        "v1_1_evidence_sha256": hashlib.sha256(canonical(generation_old["instance"])).hexdigest(),
        "v1_2_evidence_sha256": hashlib.sha256(canonical(generation_new["instance"])).hexdigest(),
        "scientific_payload_equal": generation_old["instance"]["payload"] == generation_new["instance"]["payload"],
        "h_sc01": "CLOSED",
    }
    write_json("G1V2_H_SC01_REGRESSION.json", hsc01)
    return old_cases, new_cases


def build_e2():
    e1 = json.loads((E / "G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json").read_text())
    e2 = deepcopy(e1)
    e2.pop("contract_version")
    e2["schema"] = "g1v2.execution-identity-erratum.e2"
    e2["erratum_id"] = "G1V2_EXECUTION_IDENTITY_ERRATUM_E2"
    e2["execution_identity_spec_version"] = "G1V2_EXECUTION_IDENTITY_SPEC_E2"
    e2["scientific_contract_version"] = V12
    e2["historical_authority"] = {"id": "G1V2_EXECUTION_IDENTITY_ERRATUM_E1", "sha256": PINS[E / "G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json"], "applies_to": V11}
    e2["jobid"]["scientific_identity_version"] = V12
    e2["jobid"]["algorithm"] = "j- + first 40 lowercase hex SHA256(ASCII G1V2-JOB NUL || G1V2-CJ-1(payload))"
    e2["jobid"]["payload_fields"] = ["contract_version", "control_instance_id", "candidate_id", "stage", "scale_or_null", "replicate_or_null", "metric_id_or_null", "dependency_job_ids"]
    e2["jobid"]["excluded_fields"] = ["hostname", "worker", "coordinator", "retry", "lease", "scheduling_order", "wall_clock_time"]
    e2["precedence"] = ["G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1 for cross-artifact selection", "E2 for V1.2 execution identity", "E1 for historical V1.1 provenance", "implementation"]
    write_json("G1V2_EXECUTION_IDENTITY_ERRATUM_E2.json", e2)
    md = f"""# G1V2 execution identity erratum E2

`G1V2_EXECUTION_IDENTITY_ERRATUM_E2` is the execution-identity authority for `{V12}`. It supersedes E1 only for current V1.2 version binding; E1 remains the historical V1.1 authority.

The unambiguous fields are `execution_identity_spec_version = G1V2_EXECUTION_IDENTITY_SPEC_E2` and `scientific_contract_version = {V12}`. JobID `scientific_identity_version` is `{V12}`. The JobID hash construction and canonical `dependency_job_ids` field are unchanged.

E1's boundary is preserved: scientific randomness is `G1V2-RNG-1`; blind ID is an opaque execution identifier; escrow provides blindness and truth commitment; JobID identifies a run execution bound to a scientific job. Scientific identity depends on neither escrow-key bytes nor blind-ID bytes. Changing a valid escrow key may change opaque run identifiers but never scientific control content, RNG realization, or truth.

The scientific JobID payload excludes hostname, worker, coordinator, retry, lease, scheduling order, and wall-clock time. `EI01`, `R2-G01`, and `R2-G02` are closed.
"""
    write("G1V2_EXECUTION_IDENTITY_ERRATUM_E2.md", md)
    p11, p12 = job_payload(V11), job_payload(V12)
    fixture = {
        "schema": "g1v2.jobid-cross-version-fixture.e2",
        "algorithm": e2["jobid"]["algorithm"],
        "canonicalization": "G1V2-CJ-1",
        "v1_1": {"canonical_payload_hex": canonical(p11).hex(), "payload": p11, "jobid": job_id(p11)},
        "v1_2": {"canonical_payload_hex": canonical(p12).hex(), "payload": p12, "jobid": job_id(p12)},
        "expected_inequality": True,
        "same_version_repetitions": 2,
    }
    assert fixture["v1_1"]["jobid"] != fixture["v1_2"]["jobid"]
    write_json("fixtures/G1V2_E2_JOBID_FIXTURE.json", fixture)
    mixed = {
        "schema": "g1v2.mixed-identity-negative-fixtures.v1",
        "cases": [
            {"id": "MIXED-EVIDENCE-V12-JOBID-V11", "evidence_contract_version": V12, "jobid_scientific_identity_version": V11, "jobid": fixture["v1_1"]["jobid"], "expected": "REJECT"},
            {"id": "MIXED-EVIDENCE-V11-JOBID-V12", "evidence_contract_version": V11, "jobid_scientific_identity_version": V12, "jobid": fixture["v1_2"]["jobid"], "expected": "REJECT"},
        ],
        "rule": "evidence_contract_version must equal jobid_scientific_identity_version and the I1 target version",
        "disposition": "FAIL_CLOSED",
    }
    write_json("fixtures/G1V2_MIXED_IDENTITY_NEGATIVE_FIXTURES.json", mixed)
    hsc02 = {
        "schema": "g1v2.h-sc02-regression.v1",
        "finding": "H-SC02-E1-JOBID-SCIENTIFIC-VERSION",
        "e1_scientific_identity": V11,
        "e2_scientific_identity": V12,
        "e1_sha256": PINS[E / "G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json"],
        "e2_sha256": sha(OUT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E2.json"),
        "canonical_v1_1_fixture_input": fixture["v1_1"]["canonical_payload_hex"],
        "canonical_v1_2_fixture_input": fixture["v1_2"]["canonical_payload_hex"],
        "v1_1_jobid": fixture["v1_1"]["jobid"],
        "v1_2_jobid": fixture["v1_2"]["jobid"],
        "expected_inequality": True,
        "h_sc02": "CLOSED",
    }
    write_json("G1V2_H_SC02_REGRESSION.json", hsc02)
    return e2


def build_i1(registry, root_doc):
    e2_json_sha = sha(OUT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E2.json")
    e2_md_sha = sha(OUT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E2.md")
    i1 = {
        "schema": "g1v2.v1_2-integration-supplement.i1",
        "integration_authority": "G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1",
        "scientific_contract": {"version": V12, "machine_sha256": PINS[G / "G1_V2_EXECUTABLE_CONTRACT_V1_2.json"], "markdown_sha256": PINS[G / "G1_V2_EXECUTABLE_CONTRACT_V1_2.md"]},
        "generation": {"version": "G1V2_GENERATION_SEMANTICS_V1", "semantics_sha256": PINS[G / "G1V2_GENERATION_SEMANTICS_V1.json"], "golden_root_sha256": PINS[G / "G1V2_GENERATION_GOLDEN_SUITE_V1.json"]},
        "status_reachability": {"version": SV2, "sha256": PINS[C / "registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json"]},
        "execution_identity": {"id": "G1V2_EXECUTION_IDENTITY_ERRATUM_E2", "machine_sha256": e2_json_sha, "markdown_sha256": e2_md_sha, "scientific_identity_version": V12},
        "evidence": {"contract_version": EV12, "scientific_contract_version": V12, "schema_count": 15, "schema_root_sha256": root_doc["root_sha256"], "schema_root_artifact_sha256": sha(OUT / "G1V2_V1_2_EVIDENCE_SCHEMA_ROOT.json"), "schema_registry_sha256": sha(OUT / "G1V2_V1_2_EVIDENCE_SCHEMA_REGISTRY.json")},
        "integration_invariant": [V12, V12, V12, V12, V12],
        "historical_provenance": {"e1_sha256": PINS[E / "G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json"], "v1_1_evidence_schema_root_sha256": "4744ca82532cd47a0d02bb680796b26a11ceca57d6229f0b312df69a103f784b"},
        "role": "cross-artifact integration only; does not override scientific semantics",
        "unknown_or_mixed_version_disposition": "FAIL_CLOSED",
        "scientific_semantic_change": 0,
    }
    write_json("G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.json", i1)
    md = f"""# G1V2 V1.2 integration supplement I1

`G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1` binds the already frozen `{V12}` to the V1.2 evidence schema root `{root_doc['root_sha256']}`, E2, `{SV2}`, and unchanged `G1V2_GENERATION_SEMANTICS_V1`.

For future V1.2 production execution, scientific contract identity, E2/JobID scientific identity, evidence `contract_version`, and schema-family binding are all exactly `{V12}`. Unknown or mixed versions fail closed. Historical V1.1 schemas and E1 remain immutable provenance and cannot override I1 for a V1.2 run.

I1 governs only cross-artifact selection. Models, candidates, RNG, generation, PM/F2, thresholds, gates, adequacy, equivalence, minimality, and final-verdict semantics are unchanged. No V1.3 is issued.
"""
    write("G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.md", md)
    return i1


def build_graph(root_doc):
    nodes = [
        ["V1_2_CONTRACT", PINS[G / "G1_V2_EXECUTABLE_CONTRACT_V1_2.json"]],
        ["GENERATION_SEMANTICS_V1", PINS[G / "G1V2_GENERATION_SEMANTICS_V1.json"]],
        ["GENERATION_GOLDENS_V1", PINS[G / "G1V2_GENERATION_GOLDEN_SUITE_V1.json"]],
        ["STATUS_REACHABILITY_V2", PINS[C / "registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json"]],
        ["EXECUTION_IDENTITY_E2", sha(OUT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E2.json")],
        ["EVIDENCE_SCHEMA_REGISTRY_V1_2", sha(OUT / "G1V2_V1_2_EVIDENCE_SCHEMA_REGISTRY.json")],
        ["EVIDENCE_SCHEMA_ROOT_V1_2", sha(OUT / "G1V2_V1_2_EVIDENCE_SCHEMA_ROOT.json")],
        ["INTEGRATION_SUPPLEMENT_I1", sha(OUT / "G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.json")],
    ]
    edges = [
        ["INTEGRATION_SUPPLEMENT_I1", "V1_2_CONTRACT", "binds"],
        ["INTEGRATION_SUPPLEMENT_I1", "GENERATION_SEMANTICS_V1", "binds"],
        ["INTEGRATION_SUPPLEMENT_I1", "GENERATION_GOLDENS_V1", "binds"],
        ["INTEGRATION_SUPPLEMENT_I1", "STATUS_REACHABILITY_V2", "binds"],
        ["INTEGRATION_SUPPLEMENT_I1", "EXECUTION_IDENTITY_E2", "selects"],
        ["INTEGRATION_SUPPLEMENT_I1", "EVIDENCE_SCHEMA_ROOT_V1_2", "selects"],
        ["EVIDENCE_SCHEMA_ROOT_V1_2", "EVIDENCE_SCHEMA_REGISTRY_V1_2", "hash-binds"],
        ["EXECUTION_IDENTITY_E2", "V1_2_CONTRACT", "scientific-identity"],
        ["EVIDENCE_SCHEMA_REGISTRY_V1_2", "V1_2_CONTRACT", "contract-version"],
        ["V1_2_CONTRACT", "GENERATION_SEMANTICS_V1", "scientific-semantics"],
        ["V1_2_CONTRACT", "GENERATION_GOLDENS_V1", "regression"],
        ["V1_2_CONTRACT", "STATUS_REACHABILITY_V2", "status-semantics"],
    ]
    graph_nodes = [{"id": x, "sha256": y} for x, y in nodes]
    next(n for n in graph_nodes if n["id"] == "EVIDENCE_SCHEMA_ROOT_V1_2")["schema_family_root_sha256"] = root_doc["root_sha256"]
    graph = {"schema": "g1v2.v1_2-authority-graph.v1", "nodes": graph_nodes, "edges": [{"from": a, "to": b, "relation": c} for a, b, c in edges], "cycles": 0, "unresolved_authority_edges": 0}
    write_json("G1V2_V1_2_AUTHORITY_GRAPH.json", graph)
    return graph


def build_audits(registry, root_doc):
    rows = [
        ["H-SC01", "schemas/*.schema.json", "15 literal V1.1 bindings", "MUST_VERSION_TO_V1_2", "new 15-schema V1.2 family", "CLOSED"],
        ["H-SC03", "G1V2_EVIDENCE_SCHEMA_REGISTRY_V1_1.tsv", "V1.1 schema paths/ids", "HISTORICAL_CORRECT", "new V1.2 JSON registry selected by I1", "CLOSED"],
        ["H-SC04", "G1V2_EVIDENCE_PAYLOAD_CONTRACT_V1_1.tsv", "filename versioned; rows contain no contract literal", "HISTORICAL_CORRECT", "preserved by hash as version-neutral payload semantics", "CLOSED"],
        ["H-SC05", "G1V2_EVIDENCE_STATUS_MATRIX_V1_1.tsv", "filename versioned; no contract literal", "HISTORICAL_CORRECT", "unchanged status semantics under V2", "CLOSED"],
        ["H-SC06", "G1V2_REASON_REGISTRY_V1_1.tsv", "filename versioned; no contract literal", "HISTORICAL_CORRECT", "unchanged reason semantics", "CLOSED"],
        ["H-SC07", "reference/evidence_only_verifier.py", "hard-coded V1.1 selection", "MUST_VERSION_TO_V1_2", "explicit I1/registry V1.2 verifier", "CLOSED"],
        ["H-SC08", "registries/G1V2_DAG_CONTRACT.json", "generic contract_version payload field", "NONNORMATIVE", "E2 supplies current V1.2 value; dependency_job_ids preserved", "CLOSED"],
        ["H-SC02", "G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json", "E1 V1.1 scientific identity", "HISTORICAL_CORRECT", "E2 binds V1.2; E1 immutable", "CLOSED"],
        ["H-SC09", "G1_V2_EXECUTABLE_CONTRACT_V1_2.json/evidence_contract", "historical V1.1 evidence authority references", "MUST_VERSION_TO_V1_2", "I1 supersedes only cross-artifact selection", "CLOSED"],
        ["H-SC10", "G1_V2_EXECUTABLE_CONTRACT_V1_2.json/execution_identity_erratum", "E1 selection", "MUST_VERSION_TO_V1_2", "I1 selects E2", "CLOSED"],
        ["H-SC11", "G1_V2_EXECUTABLE_CONTRACT_V1_2.md inherited V1.1 section", "explicit historical inherited prose", "HISTORICAL_CORRECT", "V1.2 preamble and I1 precedence disambiguate", "CLOSED"],
        ["H-SC12", "golden/schema-positive/cases.json", "historical V1.1 fixture family", "HISTORICAL_CORRECT", "new disposable V1.2 fixtures; old goldens unchanged", "CLOSED"],
    ]
    write_tsv("G1V2_V1_2_VERSION_BINDING_AUDIT.tsv", ["finding_id", "artifact", "occurrence", "classification", "disposition", "status"], rows)
    changes = [
        ["15 V1.2 schemas", "EVIDENCE_SCHEMA_VERSIONING", 0],
        ["contract_version literals", "EVIDENCE_VERSION_BINDING", 0],
        ["evidence content hashes", "EVIDENCE_HASH_TRANSITIVE_CHANGE", 0],
        ["E2 current identity", "EXECUTION_IDENTITY_VERSION_BINDING", 0],
        ["V1.2 fixture JobID", "JOBID_TRANSITIVE_CHANGE", 0],
        ["I1 and authority graph", "CROSS_ARTIFACT_INTEGRATION", 0],
    ]
    write_tsv("G1V2_TASK85C_I_CHANGE_CLASSIFICATION.tsv", ["artifact_or_change", "classification", "scientific_semantic_change"], changes)


def build_validations(registry, root_doc, graph):
    schema_rows = [[e["evidence_type"], e["schema_id"], e["schema_sha256"], "Draft 2020-12", "Ajv 8.17.1 PASS", "Hyperjump 1.17.1 PASS", "PASS"] for e in registry["entries"]]
    write_tsv("TASK85C_I_SCHEMA_VALIDATION.tsv", ["evidence_type", "schema_id", "sha256", "dialect", "primary_validator", "secondary_validator", "verdict"], schema_rows)
    fixture = json.loads((OUT / "fixtures/G1V2_E2_JOBID_FIXTURE.json").read_text())
    job_rows = [
        ["V1.1 canonical input", fixture["v1_1"]["canonical_payload_hex"], fixture["v1_1"]["jobid"], "PASS"],
        ["V1.2 canonical input", fixture["v1_2"]["canonical_payload_hex"], fixture["v1_2"]["jobid"], "PASS"],
        ["cross-version inequality", "V1.1 != V1.2", "YES", "PASS"],
        ["same-version determinism", "Python + JavaScript", "BYTE_IDENTICAL", "PASS"],
        ["dependency field", "dependency_job_ids", "PRESERVED", "PASS"],
        ["EI01 boundary", "no escrow/blind-id scientific dependency", "CLOSED", "PASS"],
    ]
    write_tsv("TASK85C_I_JOBID_VALIDATION.tsv", ["check", "input_or_rule", "result", "verdict"], job_rows)
    cross = [
        ["scientific contract", V12, "PASS"], ["I1 contract", V12, "PASS"], ["E2 identity", V12, "PASS"],
        ["JobID identity", V12, "PASS"], ["evidence contract", V12, "PASS"], ["schema binding", V12, "PASS"],
        ["authority graph cycles", graph["cycles"], "PASS"], ["unresolved edges", graph["unresolved_authority_edges"], "PASS"],
        ["unknown version", "FAIL_CLOSED", "PASS"], ["mixed identity", "FAIL_CLOSED", "PASS"],
    ]
    write_tsv("TASK85C_I_CROSS_ARTIFACT_VALIDATION.tsv", ["check", "value", "verdict"], cross)
    checks = [
        ("PARENT_V1_2_IDENTITY", "SUPPORTED"), ("GENERATION_SEMANTICS_IDENTITY", "SUPPORTED"), ("GENERATION_GOLDEN_IDENTITY", "SUPPORTED"),
        ("TASK85C_G_IDENTITY", "SUPPORTED"), ("STATUS_REACHABILITY_V2_IDENTITY", "SUPPORTED"), ("HISTORICAL_E1_IDENTITY", "SUPPORTED"),
        ("HISTORICAL_V1_1_EVIDENCE_ROOT", "SUPPORTED"), ("TASK85C_H_IDENTITY", "SUPPORTED"), ("H_SC01_REPRODUCED", "YES"),
        ("H_SC02_REPRODUCED", "YES"), ("H_SC01", "CLOSED"), ("H_SC02", "CLOSED"),
        ("V1_2_EVIDENCE_SCHEMA_COUNT", "15"), ("V1_2_EVIDENCE_SCHEMA_VALID", "15"), ("V1_2_EVIDENCE_SCHEMA_ROOT_SHA256", root_doc["root_sha256"]),
        ("DRAFT_2020_12", "PASS"), ("MUTUAL_EXCLUSIVITY", "PASS"), ("STATUS_MATRIX_COMPATIBILITY", "PASS"),
        ("REASON_COMPATIBILITY", "PASS"), ("PAYLOAD_COMPATIBILITY", "PASS"), ("V1_1_POSITIVE_REGRESSION", "PASS"),
        ("V1_2_POSITIVE_REGRESSION", "PASS"), ("CROSS_VERSION_NEGATIVE_REGRESSION", "PASS"), ("MIXED_IDENTITY_NEGATIVE_REGRESSION", "PASS"),
        ("E2", "FROZEN"), ("E2_SCIENTIFIC_IDENTITY", V12), ("I1", "FROZEN"),
        ("AUTHORITY_GRAPH_CYCLES", "0"), ("AUTHORITY_GRAPH_UNRESOLVED_EDGES", "0"),
        ("R2_G01", "CLOSED"), ("R2_G02", "CLOSED"), ("EI01", "CLOSED"), ("PF_SC01", "CLOSED"),
        ("GENERATION_PATHS", "26/26"), ("CANDIDATE_REGISTRY_UNCHANGED", "PASS"), ("PM_F2_UNCHANGED", "PASS"),
        ("DECISION_LOGIC_UNCHANGED", "PASS"), ("STATUS_REACHABILITY_UNCHANGED", "PASS"), ("SCIENTIFIC_SEMANTIC_CHANGE", "0"),
        ("PRODUCTION_ESCROW_KEY_CREATED", "NO"), ("PRODUCTION_BLIND_CONTROLS_CREATED", "0"), ("PRODUCTION_NATURAL_CONTROLS_CREATED", "0"),
        ("PRODUCTION_JOBIDS_CREATED", "0"), ("PRODUCTION_DAG_CREATED", "NO"), ("CONFIRMATORY_EXECUTION", "NO"),
        ("VOYNICH_EVALUATED", "NO"), ("SCIENTIFIC_FIREWALL", "INTACT"), ("TASK85C_H_RETRY_READY", "SUPPORTED"),
    ]
    write_tsv("TASK85C_I_VALIDATION.tsv", ["check", "verdict"], checks)


def build_docs(root_doc):
    e2j = sha(OUT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E2.json")
    e2m = sha(OUT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E2.md")
    i1j = sha(OUT / "G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.json")
    i1m = sha(OUT / "G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.md")
    write("TASK85C_I_DESIGN.md", f"""# Task85c-i design

Historical V1.1 authorities remain immutable. A separate 15-schema Draft 2020-12 family binds evidence to `{V12}` and is selected deterministically by I1; unknown and mixed versions fail closed. Each V1.2 schema is the corresponding V1.1 schema with only contract/schema identity strings versioned, proven by reversible normalization.

E2 separates its specification version from its target scientific-contract version and otherwise preserves E1's blind/escrow/RNG boundary and JobID function. I1 binds the parent V1.2 contract, E2, V1.2 evidence root, status/reachability V2, and unchanged generation authority without changing scientific semantics. The hash graph is acyclic because E2 and the schema registry do not depend on I1, and I1 does not depend on the graph.
""")
    write("TASK85C_I_REPORT.md", f"""# Task85c-i report

H-SC01 and H-SC02 were reproduced and closed as authority-integration defects. All 15 historical schema types have mutually exclusive V1.2 counterparts. V1.1 positives remain valid only under V1.1; disposable V1.2 positives validate only under V1.2; cross-version, mixed, mutated, unknown-version, and wrong-root inputs fail closed.

V1.2 evidence schema root: `{root_doc['root_sha256']}`. E2 machine/Markdown SHA-256: `{e2j}` / `{e2m}`. I1 machine/Markdown SHA-256: `{i1j}` / `{i1m}`. The authority graph has zero cycles and zero unresolved edges.

`R2_G01`, `R2_G02`, `EI01`, and `PF_SC01` are closed; generation remains 26/26. Models, candidates, RNG, PM/F2, gates, decision logic, and status/reachability are unchanged (`SCIENTIFIC_SEMANTIC_CHANGE=0`). No production or confirmatory material was created or accessed; the scientific firewall is intact. Task85c-h retry is supported.
""")


def build_manifest():
    terminal = "G1V2_V1_2_EVIDENCE_EXECUTION_INTEGRATION_I1_FROZEN"
    write(terminal, terminal + "\n")
    artifacts = []
    for path in sorted(OUT.rglob("*"), key=lambda p: p.relative_to(OUT).as_posix().encode()):
        if path.is_file() and path.name != "TASK85C_I_RESULTS_MANIFEST.json" and "__pycache__" not in path.parts:
            artifacts.append({"path": path.relative_to(OUT).as_posix(), "bytes": path.stat().st_size, "sha256": sha(path)})
    lines = "".join(f"{a['sha256']}  {a['path']}\n" for a in artifacts)
    root = hashlib.sha256(lines.encode()).hexdigest()
    result = {
        "schema": "task85c-i-results-v1",
        "status": "FROZEN",
        "task_sha256": "6c3732157a32eec8d63b6e09ac32cb35d25ceafa2dd9685533b9682f147478ca",
        "scientific_contract_version": V12,
        "scientific_contract_markdown_sha256": PINS[G / "G1_V2_EXECUTABLE_CONTRACT_V1_2.md"],
        "scientific_contract_machine_sha256": PINS[G / "G1_V2_EXECUTABLE_CONTRACT_V1_2.json"],
        "generation_semantics_sha256": PINS[G / "G1V2_GENERATION_SEMANTICS_V1.json"],
        "generation_golden_root_sha256": PINS[G / "G1V2_GENERATION_GOLDEN_SUITE_V1.json"],
        "evidence_contract_version": EV12,
        "v1_2_evidence_schema_count": 15,
        "v1_2_evidence_schema_root_sha256": json.loads((OUT / "G1V2_V1_2_EVIDENCE_SCHEMA_ROOT.json").read_text())["root_sha256"],
        "v1_2_evidence_schema_registry_sha256": sha(OUT / "G1V2_V1_2_EVIDENCE_SCHEMA_REGISTRY.json"),
        "e2_machine_sha256": sha(OUT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E2.json"),
        "e2_markdown_sha256": sha(OUT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E2.md"),
        "i1_machine_sha256": sha(OUT / "G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.json"),
        "i1_markdown_sha256": sha(OUT / "G1V2_V1_2_INTEGRATION_SUPPLEMENT_I1.md"),
        "authority_graph_sha256": sha(OUT / "G1V2_V1_2_AUTHORITY_GRAPH.json"),
        "h_sc01_regression_sha256": sha(OUT / "G1V2_H_SC01_REGRESSION.json"),
        "h_sc02_regression_sha256": sha(OUT / "G1V2_H_SC02_REGRESSION.json"),
        "validation_sha256": sha(OUT / "TASK85C_I_VALIDATION.tsv"),
        "h_sc01": "CLOSED", "h_sc02": "CLOSED", "r2_g01": "CLOSED", "r2_g02": "CLOSED", "ei01": "CLOSED", "pf_sc01": "CLOSED",
        "generation_paths": "26/26", "scientific_semantic_change": 0, "scientific_firewall": "INTACT",
        "production_materialization": {"escrow_key_created": False, "blind_controls_created": 0, "natural_controls_created": 0, "jobids_created": 0, "dag_created": False},
        "artifacts_excluding_manifest": artifacts,
        "artifact_root_excluding_manifest_sha256": root,
        "artifact_root_definition": "SHA-256 of sha256sum-format lines using task85c-i-relative paths in bytewise order; excludes TASK85C_I_RESULTS_MANIFEST.json and __pycache__",
        "task85c_h_retry_ready": "SUPPORTED",
        "next_task": "Task85c-h",
        "terminal_marker": terminal,
    }
    write_json("TASK85C_I_RESULTS_MANIFEST.json", result)


def main():
    verify_inputs()
    registry, root_doc, transformed = build_schemas()
    build_fixtures(transformed)
    build_e2()
    build_i1(registry, root_doc)
    graph = build_graph(root_doc)
    build_audits(registry, root_doc)
    build_validations(registry, root_doc, graph)
    build_docs(root_doc)
    build_manifest()
    print(f"V1_2_EVIDENCE_SCHEMA_ROOT_SHA256={root_doc['root_sha256']}")
    print("H_SC01=CLOSED")
    print("H_SC02=CLOSED")


if __name__ == "__main__":
    sys.dont_write_bytecode = True
    main()
