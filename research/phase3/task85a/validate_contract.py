#!/usr/bin/env python3
"""Target-blind Task85a contract validator and finite-grid materializer.

The only reads outside task85a are byte-level SHA-256 verification and JSON
provenance checks. This program contains no corpus parser or fitting entrypoint.
"""

from __future__ import annotations

import argparse
import csv
import hashlib
import itertools
import json
import math
import pathlib
import sys
import tempfile

ROOT = pathlib.Path(__file__).resolve().parents[3]
HERE = pathlib.Path(__file__).resolve().parent
CONTRACT = HERE / "G1_EXECUTABLE_CONTRACT.json"

PARENT_HASHES = {
    "GRAMMAR_ABLATION_REGISTRY.tsv": "efd327e7ea7191c5069158861cecc666075755d5546ad957fd9f321742cd8fe1",
    "GRAMMAR_BASELINE_REGISTRY.tsv": "584a00b2373ead6feb1744910d67cd0507b3395d77619969a43832344704e983",
    "GRAMMAR_COMPLEXITY_CONTRACT.md": "285026cf6f0b5a60c805d48d67a761f2104ad2df6870e5814140ed0a429a2558",
    "GRAMMAR_CORPUS_SPLIT.tsv": "639969bd91daaf362df49afa4fc12c3f5289cea4200cbb71d5f743d5c5bff551",
    "GRAMMAR_CORPUS_SPLIT_FROZEN": "f816d3ad9f1c702fe80a9b4314d06aad4b0ad150d3ba6f6b93c3a897a0e46145",
    "GRAMMAR_CORPUS_SPLIT_MANIFEST.json": "80b49086623f53968a0c925c3a6780a82d35b9c118747a9ab6f90fe7ce03719a",
    "GRAMMAR_EXPERIMENT_CONTRACT_FROZEN": "a7895c4e4c91bcacd215a71d26cfdc13bdac7013e374f28efd7dad832ac8d2c6",
    "GRAMMAR_F2_APPLICABILITY.tsv": "2816e863408cb8248df25e9d9d09e3c6b531e0581fed7705a15d628366edebfa",
    "GRAMMAR_FAILURE_REGISTRY.tsv": "8aca3beaac133fa133972a9436c2a852a7e393a0eda11981fb113ca46cb0e29d",
    "GRAMMAR_METRIC_REGISTRY.tsv": "3c66514c0bae75eab66a22e774133e2ec9c4143cdb0ed687389070a843ac034e",
    "GRAMMAR_MODEL_REGISTRY.tsv": "e3e744af5b02fa5078776a3f14f62d2655926b1b30d65a2d2c18c645c1b7d20f",
    "GRAMMAR_RESIDUAL_CONTRACT.md": "2a44094c4f93901a0f3c5063f41f5442b827a0db1787d385751b756a33a87e93",
    "GRAMMAR_UNIT_REGISTRY.tsv": "3e2618d7b3a3f68b5af46964a9acbf73134e58d4241dd3e6645e6d432e0e7b5c",
    "GRAMMAR_VALIDATION_CONTRACT.md": "8b286dbb0ceb16e293f0d00ed06506edd266c21a758f97dd42c2873e5568a47d",
    "PHASE3_LINE_A_RESEARCH_QUESTIONS.md": "ea666cdff65ae811259ab6d01890af8223b765ff5e61c69e9037d21d2e130844",
    "TASK85_DESIGN.md": "6e30280f31f6cd707bbc22af0f4866f746ba2cfc29a37ec9400c0b8ab4a14f15",
    "TASK85_REPORT.md": "4aa68c5ac68def15685bda6fe10858731fe693ef8dc6a76ce13dc232beceeaf6",
    "TASK85_RESULTS_MANIFEST.json": "0d894d9e12631f691bf94b9fa88260b04373c8c5470b1980848e157b59366021",
    "TASK86_HANDOFF.md": "ae1085bab0cc4dd6b2340545e006356601ab18c2132ad03598edf59e3aa3b316",
}
CORPUS_HASHES = {
    "data_work/ZL3b-x7.canonical.txt": "f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692",
    "data_work/IT2a-x7.canonical.txt": "10286ee7e11ad974e9d0f884e3b0df1b588745a4b77ad428a638a5ff63946a8b",
    "data/ZL3b-n.txt": "bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc",
    "data/IT2a-n.txt": "7f27a8b0feed8f6de0a99900df6bf912dd1d295c38e5f830bac8b41c3f536fb5",
}
TASK86_HASHES = {
    "CONTRACT_PREFLIGHT.tsv": "131dcb9224c7f60b56c518c5a72329cf85736b1fe9173318b6285be309e46c48",
    "TASK86_DESIGN_EXECUTION.md": "2334845867dcb4fca87913a2410c72f45419f6db68db955c2c08322a361699e7",
    "TASK86_EXPERIMENT_BLOCKED": "1a55251347040f9b2be889aaec17b1016a5a55c2974c87371a9bf4efba24651f",
    "TASK86_REPORT.md": "a8781f744decf2b0f7a8e163c2b6be5597e5709aca0be587bc696ff2576eb259",
    "TASK86_RESULTS_MANIFEST.json": "2a29deb88da8c3bf09084ea6b80a135fca195f50c09455b53dc112809d1ced80",
    "TASK87_HANDOFF.md": "a32ca632612c911d0f9ce40e2a70fc98bb7108b5192a64c32152d1c4f3d115ed",
}
REQUIRED = [
    "TASK85A_DESIGN.md", "TASK85A_TARGET_BLINDNESS_AUDIT.md",
    "CONTRACT_GAP_REGISTRY.tsv", "G1_ALGORITHM_REGISTRY.tsv",
    "G1_HYPERPARAMETER_GRID.tsv", "M2_HYPERPARAMETER_GRID.tsv",
    "M4_HYPERPARAMETER_GRID.tsv", "M5_HYPERPARAMETER_GRID.tsv",
    "G1_CALIBRATION_CONTRACT.md", "G1_STABILITY_CONTRACT.md",
    "G1_TRANSCRIPTION_STABILITY_CONTRACT.md", "G1_FAILURE_THRESHOLDS.tsv",
    "NEGATIVE_TOKEN_PROTOCOL.md", "PM5_CALIBRATION_SPEC.md",
    "PM6_DISCRIMINATION_SPEC.md", "G1_ADEQUACY_GATES.md",
    "G1_MODEL_LADDER_CONTRACT.md", "G1_SEED_CONTRACT.md",
    "G1_EXECUTABLE_CONTRACT.json", "TASK86_BLOCKER_REGRESSION.tsv",
    "TASK86R_HANDOFF.md", "TASK85A_REPORT.md",
    "TASK85A_RESULTS_MANIFEST.json",
    "PLACEHOLDER_AUDIT.tsv",
    "GRAMMAR_EXPERIMENT_CONTRACT_V1_1_MANIFEST.json",
    "GRAMMAR_EXPERIMENT_CONTRACT_V1.1_FROZEN",
    "validate_contract.py",
]


def sha256(path: pathlib.Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def load_contract(path: pathlib.Path = CONTRACT) -> dict:
    with path.open(encoding="utf-8") as stream:
        return json.load(stream, object_pairs_hook=dict)


def candidates(contract: dict) -> list[dict[str, str]]:
    rows: list[dict[str, str]] = []
    for model in [f"M{i}" for i in range(6)]:
        spec = contract["models"][model]
        dimensions = spec["parameters"]
        names = list(dimensions)
        for index, values in enumerate(itertools.product(*(dimensions[n] for n in names)), 1):
            params = dict(zip(names, values))
            rows.append({
                "model_class": model,
                "candidate_id": f"{model}-{index:03d}",
                "parameters_json": json.dumps(params, sort_keys=True, separators=(",", ":")),
                "deterministic": "TRUE",
                "generative": "TRUE" if spec["generative"] else "FALSE",
                "applicable_metrics": "PM0,PM1,PM2,PM4,PM5,PM6;PM3_WHERE_UNIT_MATCHED;G1_F2",
                "applicable_ablations": "A5" if model == "M0" else ("A6" if model in {"M3", "M4"} else "NONE"),
            })
    return rows


def write_tsv(path: pathlib.Path, rows: list[dict[str, str]], fields: list[str]) -> None:
    with path.open("w", encoding="utf-8", newline="") as stream:
        writer = csv.DictWriter(stream, fieldnames=fields, delimiter="\t", lineterminator="\n")
        writer.writeheader()
        writer.writerows(rows)


def algorithm_rows(contract: dict) -> list[dict[str, str]]:
    algorithms = []
    for model in [f"M{i}" for i in range(6)]:
        spec = contract["models"][model]
        algorithms.append({
            "model_class": model, "algorithm": spec["algorithm"],
            "implementation_version": "task85a-v1.1", "input_scope": "G1_TOKEN_INTERNAL_DEVELOPMENT",
            "topology_policy": spec.get("topology", spec.get("initial", "not_applicable")),
            "generation": spec["generation"], "unseen": spec.get("unseen", spec.get("unseen_scoring", "specified_in_algorithm")),
            "ordering": spec.get("tie_break", contract["ordering"]["candidate_enumeration"]),
            "resolution_basis": "simplest sufficient target-blind implementation",
        })
    return algorithms


def write_derived(contract: dict) -> None:
    rows = candidates(contract)
    fields = ["model_class", "candidate_id", "parameters_json", "deterministic", "generative", "applicable_metrics", "applicable_ablations"]
    write_tsv(HERE / "G1_HYPERPARAMETER_GRID.tsv", rows, fields)
    for model in ("M2", "M4", "M5"):
        write_tsv(HERE / f"{model}_HYPERPARAMETER_GRID.tsv", [r for r in rows if r["model_class"] == model], fields)
    algorithms = algorithm_rows(contract)
    write_tsv(HERE / "G1_ALGORITHM_REGISTRY.tsv", algorithms, list(algorithms[0]))


def table_rows(path: pathlib.Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as stream:
        return list(csv.DictReader(stream, delimiter="\t"))


def validate(contract_path: pathlib.Path = CONTRACT, require_all: bool = True) -> list[str]:
    errors: list[str] = []
    try:
        contract = load_contract(contract_path)
    except (OSError, json.JSONDecodeError) as exc:
        return [f"contract parse: {exc}"]
    if contract.get("schema") != "G1_EXECUTABLE_CONTRACT_V1_1":
        errors.append("wrong contract schema")
    models = contract.get("models", {})
    if set(models) != {f"M{i}" for i in range(6)}:
        errors.append("model space is not exactly M0-M5")
    for model, spec in models.items():
        if not spec.get("algorithm") or "parameters" not in spec or not spec.get("generation"):
            errors.append(f"{model}: incomplete algorithm")
            continue
        if not spec["parameters"]:
            errors.append(f"{model}: empty parameter map")
        for name, values in spec["parameters"].items():
            if not isinstance(values, list) or not values:
                errors.append(f"{model}.{name}: empty/non-finite dimension")
            elif any(not isinstance(v, (int, float)) or not math.isfinite(v) for v in values):
                errors.append(f"{model}.{name}: non-numeric/non-finite value")
    computed = candidates(contract)
    counts = {m: sum(r["model_class"] == m for r in computed) for m in models}
    counts["total"] = len(computed)
    if counts != contract.get("candidate_counts"):
        errors.append(f"candidate counts mismatch: {counts}")
    for section in ("seed", "calibration", "generation", "metrics", "thresholds", "gates", "failure"):
        if not contract.get(section):
            errors.append(f"missing/empty {section}")
    required_thresholds = {
        "probability_normalization_abs_error", "undefined_score_fraction",
        "run_variation_cv", "max_token_glyphs", "complexity_growth_point_slope",
        "complexity_growth_lower_ci", "memorization_complexity_ratio",
        "transcription_relative_effect", "minimality_abs_bits",
        "minimality_relative", "auc_neutral",
    }
    if set(contract.get("thresholds", {})) != required_thresholds:
        errors.append("threshold key set incomplete or enlarged")
    if contract.get("generation", {}).get("scales") != [0.5, 1.0, 2.0]:
        errors.append("scale grid mismatch")
    if contract.get("generation", {}).get("replicate_checkpoints") != [4, 8, 16, 32]:
        errors.append("replicate checkpoints mismatch")
    if contract.get("calibration", {}).get("population_count_per_generator") != 16:
        errors.append("calibration population count mismatch")
    if set(contract.get("metrics", {})) != {"PM1", "PM2", "PM3", "PM4", "PM5", "PM6", "structural"}:
        errors.append("metric functional set incomplete or enlarged")
    if set(contract.get("gates", {})) != {
        "predictive", "structural_metric", "edit_family",
        "lexical_paradigm_family", "structural_overall", "minimality",
        "representational_gain", "token_formation_depth", "explicit_rule_required",
    }:
        errors.append("gate set incomplete or enlarged")
    if set(contract.get("failure", {})) != {
        "NUMERICALLY_UNSTABLE", "COMPLEXITY_UNBOUNDED",
        "MEMORIZATION_DOMINATED", "TRAINING_FAILED", "NON_GENERATIVE",
        "HELDOUT_DEGENERATE", "STRUCTURAL_VALIDATION_FAILED",
    }:
        errors.append("failure rule set incomplete or enlarged")
    if set(contract.get("preflight", [])) != {
        "CONTRACT_COMPLETE", "ALL_GRIDS_FINITE", "ALL_THRESHOLDS_DEFINED",
        "ALL_GATES_EXECUTABLE", "NEGATIVE_PROTOCOL_COMPLETE", "PM5_COMPLETE",
        "PM6_COMPLETE", "SEED_CONTRACT_COMPLETE", "TASK86_BLOCKERS_RESOLVED",
    }:
        errors.append("preflight flags incomplete")
    if require_all:
        for name in REQUIRED:
            if not (HERE / name).is_file():
                errors.append(f"missing artifact: {name}")
        grid_path = HERE / "G1_HYPERPARAMETER_GRID.tsv"
        if grid_path.is_file() and table_rows(grid_path) != computed:
            errors.append("unified grid is not the deterministic JSON expansion")
        for model in ("M2", "M4", "M5"):
            path = HERE / f"{model}_HYPERPARAMETER_GRID.tsv"
            if path.is_file() and table_rows(path) != [r for r in computed if r["model_class"] == model]:
                errors.append(f"{model} grid mismatch")
        algorithm_path = HERE / "G1_ALGORITHM_REGISTRY.tsv"
        if algorithm_path.is_file() and table_rows(algorithm_path) != algorithm_rows(contract):
            errors.append("algorithm registry is not the deterministic JSON expansion")
        for name, expected in PARENT_HASHES.items():
            path = ROOT / "research/phase3/task85" / name
            if not path.is_file() or sha256(path) != expected:
                errors.append(f"Task85 immutable hash mismatch: {name}")
        for name, expected in TASK86_HASHES.items():
            path = ROOT / "research/phase3/task86" / name
            if not path.is_file() or sha256(path) != expected:
                errors.append(f"Task86 provenance hash mismatch: {name}")
        for rel, expected in CORPUS_HASHES.items():
            path = ROOT / rel
            if not path.is_file() or sha256(path) != expected:
                errors.append(f"authoritative corpus hash mismatch: {rel}")
        blocked = json.loads((ROOT / "research/phase3/task86/TASK86_RESULTS_MANIFEST.json").read_text(encoding="utf-8"))
        for key in ("heldout_evaluated", "voynich_models_fitted", "message_free_calibration_run", "model_selection_freeze_created"):
            if blocked.get(key) is not False:
                errors.append(f"target-blindness failure: Task86 {key}")
        for filename, key, expected_count in (
            ("CONTRACT_GAP_REGISTRY.tsv", "status", 10),
            ("TASK86_BLOCKER_REGRESSION.tsv", "regression_status", 10),
        ):
            path = HERE / filename
            if path.is_file():
                rows = table_rows(path)
                if len(rows) != expected_count or any(row.get(key) != "RESOLVED" for row in rows):
                    errors.append(f"{filename}: expected ten RESOLVED rows")
        audit_files = [p for p in HERE.iterdir() if p.suffix in {".md", ".tsv", ".json"} and "MANIFEST" not in p.name and p.name != "PLACEHOLDER_AUDIT.tsv"]
        forbidden = ("T" + "BD", "T" + "ODO", "stable " + "enough", "as " + "needed", "if " + "useful")
        for path in audit_files:
            text = path.read_text(encoding="utf-8").lower()
            for phrase in forbidden:
                if phrase.lower() in text:
                    errors.append(f"unresolved placeholder phrase in {path.name}: {phrase}")
        manifest = HERE / "GRAMMAR_EXPERIMENT_CONTRACT_V1_1_MANIFEST.json"
        if manifest.is_file():
            data = json.loads(manifest.read_text(encoding="utf-8"))
            actual = {p.name for p in HERE.iterdir() if p.is_file() and p.name != manifest.name}
            listed = set(data.get("task85a_artifacts", {}))
            if actual != listed:
                errors.append(f"V1.1 manifest inventory mismatch: missing={sorted(actual-listed)} extra={sorted(listed-actual)}")
            for rel, expected in data.get("task85a_artifacts", {}).items():
                path = HERE / rel
                if not path.is_file() or sha256(path) != expected:
                    errors.append(f"V1.1 manifest mismatch: {rel}")
    return errors


def self_test() -> list[str]:
    original = load_contract()
    failures: list[str] = []
    cases = []
    broken = json.loads(json.dumps(original)); broken["models"]["M2"]["parameters"]["max_depth"] = []; cases.append(broken)
    broken = json.loads(json.dumps(original)); del broken["thresholds"]; cases.append(broken)
    broken = json.loads(json.dumps(original)); broken["models"]["M4"]["generation"] = ""; cases.append(broken)
    for index, case in enumerate(cases):
        with tempfile.TemporaryDirectory() as dirname:
            path = pathlib.Path(dirname) / "contract.json"
            path.write_text(json.dumps(case), encoding="utf-8")
            if not validate(path, require_all=False):
                failures.append(f"negative self-test {index} was accepted")
    return failures


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--write-derived", action="store_true")
    parser.add_argument("--self-test", action="store_true")
    args = parser.parse_args()
    contract = load_contract()
    if args.write_derived:
        write_derived(contract)
    errors = validate()
    if args.self_test:
        errors.extend(self_test())
    if errors:
        for error in errors:
            print(f"FAIL\t{error}")
        return 1
    print(f"PASS\tCONTRACT_COMPLETE\tcandidates={len(candidates(contract))}")
    if args.self_test:
        print("PASS\tNEGATIVE_SELF_TESTS\t3")
    return 0


if __name__ == "__main__":
    sys.exit(main())
