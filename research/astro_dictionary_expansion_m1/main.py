#!/usr/bin/env python3
"""Frozen D0/D1 comparison for the audited M1 astronomical-label search."""

from __future__ import annotations

import argparse
import csv
import hashlib
import importlib.util
import json
import os
import random
import statistics
import sys
from collections import Counter, defaultdict
from datetime import datetime, timezone
from multiprocessing import get_context
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "research/astro_dictionary_expansion_m1"
M1_PATH = ROOT / "research/astro_token_formation_m1/main.py"
_spec = importlib.util.spec_from_file_location("frozen_m1_for_d0d1", M1_PATH)
assert _spec and _spec.loader
m1 = importlib.util.module_from_spec(_spec)
sys.modules[_spec.name] = m1
_spec.loader.exec_module(m1)

SEED = m1.SEED
CONTROLS = ("RANDOM_VOYNICH_SET", "SHUFFLED_TERMS", "PSEUDODICTIONARY")
EVIDENCE_FIELDS = ["term_id", "concept", "object_class", "language_layer",
                   "normalized_form", "attested_form", "date", "source",
                   "source_location", "confidence"]
RUN_TERMS: list[dict[str, object]] = []


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as fh:
        return list(csv.DictReader(fh, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as fh:
        writer = csv.DictWriter(fh, fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        writer.writeheader()
        writer.writerows(rows)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def fmt(value: float) -> str:
    return f"{value:.6f}"


def percentile(values: list[float], q: float) -> float:
    values = sorted(values)
    if not values:
        return 0.0
    return values[min(len(values) - 1, max(0, int((len(values) - 1) * q + .999999999)))]


def load_d0() -> list[dict[str, object]]:
    return m1.load_terms()


def load_d1() -> list[dict[str, object]]:
    grouped: dict[tuple[str, str], set[str]] = defaultdict(set)
    for row in read_tsv(OUT / "ASTRO_TERM_CORPUS_EXPANDED.tsv"):
        if row["object_class"] not in {"STAR", "PLANET_MOON"}:
            continue
        grouped[(row["term_id"], row["object_class"])].add(row["normalized_form"])
    return [{"object_id": term_id, "object_class": object_class, "forms": sorted(forms)}
            for (term_id, object_class), forms in sorted(grouped.items())]


def per_class(model: dict[str, object], labels: list[dict[str, str]], key: str) -> dict[str, tuple[int, int, float]]:
    match = model[key]
    result = {}
    for object_class in ("STAR", "PLANET_MOON"):
        class_labels = [row for row in labels if row["object_class"] == object_class]
        matched = sum(row["stolfi_coordinate"] in match for row in class_labels)
        result[object_class] = (matched, len(class_labels), matched / len(class_labels))
    return result


def null_worker(task: tuple[str, int]) -> dict[str, object]:
    control, replicate = task
    rng = random.Random(m1.seed_for(control, replicate))
    labels = m1.load_labels("TRAIN")
    terms = RUN_TERMS
    if control == "RANDOM_VOYNICH_SET":
        labels = m1.randomize_labels(labels, rng)
    elif control == "SHUFFLED_TERMS":
        terms = m1.alter_terms(terms, rng, False)
    elif control == "PSEUDODICTIONARY":
        terms = m1.alter_terms(terms, rng, True)
    best = m1.run_search(labels, terms, retain_details=False)[0]
    breakdown = per_class(best, labels, "train_match")
    return {
        "control": control, "replicate": replicate,
        "seed": m1.seed_for(control, replicate), "max_score": best["score"],
        "max_train_coverage": best["search_max_train_coverage"],
        "selected_model_coverage": best["train_coverage"],
        "star_selected_coverage": breakdown["STAR"][2],
        "planet_moon_selected_coverage": breakdown["PLANET_MOON"][2],
        "mapping_size": best["mapping_size"], "rule": best["pipeline"].rule_string,
    }


def run_one(run: str, terms: list[dict[str, object]], replicates: int, workers: int) -> dict[str, object]:
    global RUN_TERMS
    RUN_TERMS = terms
    train, held = m1.load_labels("TRAIN"), m1.load_labels("HELD_OUT")
    observed = m1.run_search(train, terms, retain_details=True)
    tasks = [(control, replicate) for control in CONTROLS for replicate in range(replicates)]
    if workers == 1:
        null_rows = [null_worker(task) for task in tasks]
    else:
        with get_context("fork").Pool(workers) as pool:
            null_rows = list(pool.imap(null_worker, tasks, chunksize=1))
    by_control: dict[str, list[dict[str, object]]] = defaultdict(list)
    for row in null_rows:
        by_control[str(row["control"])].append(row)
    conservative_mean = max(statistics.fmean(float(row["max_train_coverage"]) for row in rows)
                            for rows in by_control.values())
    for model in observed:
        model.update(m1.compute_heldout(model, held, terms))
        model["null_advantage"] = model["train_coverage"] - conservative_mean
        model["empirical_p"] = max(
            (1 + sum(float(row["max_score"]) >= model["score"] for row in rows)) / (len(rows) + 1)
            for rows in by_control.values())
        model["candidate_band"] = m1.model_band(model)
    observed.sort(key=lambda model: (-model["train_coverage"], -model["score"], model["empirical_p"],
                                     -model["heldout_coverage"], model["mapping_size"], model["pipeline"].rule_string))
    for index, model in enumerate(observed, 1):
        model["model_id"] = f"{run}_M1_{index:03d}"
    best = observed[0]
    return {"run": run, "terms": terms, "train": train, "held": held,
            "models": observed, "best": best, "null_rows": null_rows,
            "by_control": by_control, "conservative_mean": conservative_mean}


def corpus_audit(rows: list[dict[str, str]], d0: list[dict[str, object]], d1: list[dict[str, object]]) -> None:
    class_concepts: dict[str, set[str]] = defaultdict(set)
    class_forms: Counter[str] = Counter()
    layers: Counter[str] = Counter()
    dates: Counter[str] = Counter()
    sources: Counter[str] = Counter()
    variants: Counter[str] = Counter()
    for row in rows:
        class_concepts[row["object_class"]].add(row["term_id"])
        class_forms[row["object_class"]] += 1
        layers[row["language_layer"]] += 1
        dates[row["date"]] += 1
        sources[row["source"]] += 1
        variants[row["term_id"]] += 1
    d0_counts = Counter(str(term["object_class"]) for term in d0)
    d1_counts = Counter(str(term["object_class"]) for term in d1)
    train = m1.load_labels("TRAIN")
    selected = train + m1.load_labels("HELD_OUT")
    label_counts = Counter(row["object_class"] for row in selected)
    ceilings = {cls: min(label_counts[cls], d1_counts[cls]) for cls in label_counts}
    theoretical = sum(ceilings.values()) / len(selected)
    lines = [
        "# Astronomical term corpus audit", "",
        "The expanded evidence table stores one attestation per row. `term_id` is the",
        "concept key used by one-to-one matching; multiple spellings under a key do not",
        "create additional dictionary capacity.", "", "## Concepts and forms", "",
        "| Class | D0 concepts | D1 concepts | Evidence rows |", "|---|---:|---:|---:|",
    ]
    for cls in sorted(class_concepts):
        lines.append(f"| {cls} | {d0_counts.get(cls, 0)} | {len(class_concepts[cls])} | {class_forms[cls]} |")
    lines += ["", "## Language layers", "", "| Layer | Forms |", "|---|---:|"]
    lines += [f"| {key} | {value} |" for key, value in sorted(layers.items())]
    lines += ["", "## Date distribution", "", "| Date | Forms |", "|---|---:|"]
    lines += [f"| {key} | {value} |" for key, value in sorted(dates.items())]
    lines += ["", "## Source distribution", "", "| Source | Forms |", "|---|---:|"]
    lines += [f"| {key} | {value} |" for key, value in sorted(sources.items())]
    multi = sum(value > 1 for value in variants.values())
    excess = sum(value - 1 for value in variants.values())
    lines += ["", "## Variant statistics", "",
              f"Concepts with more than one attestation: **{multi}**.",
              f"Attestation rows beyond one-per-concept: **{excess}**.",
              "Exact duplicate normalized forms across sources remain separate evidence rows",
              "but collapse to one form before M1 search.", "", "## Matching capacity", "", "```text",
              f"STAR_LABELS={label_counts['STAR']}", f"STAR_CONCEPTS={d1_counts['STAR']}",
              f"PLANET_LABELS={label_counts['PLANET_MOON']}",
              f"PLANET_CONCEPTS={d1_counts['PLANET_MOON']}",
              f"THEORETICAL_MAX_COVERAGE={theoretical:.6f}", "```", "",
              "ZODIAC evidence is audited but excluded from M1 because no frozen label has",
              "that class. The capacity calculation therefore uses only STAR and PLANET_MOON.", ""]
    (OUT / "ASTRO_TERM_CORPUS_AUDIT.md").write_text("\n".join(lines), encoding="utf-8")


def model_rows(result: dict[str, object]) -> list[dict[str, object]]:
    rows = []
    for model in result["models"]:
        train_class = per_class(model, result["train"], "train_match")
        held_class = per_class(model, result["held"], "held_match")
        rows.append({
            "model_id": model["model_id"], "candidate_band": model["candidate_band"],
            "train_coverage": fmt(model["train_coverage"]), "train_matched": len(model["train_match"]), "train_total": len(result["train"]),
            "star_train_coverage": fmt(train_class["STAR"][2]), "planet_moon_train_coverage": fmt(train_class["PLANET_MOON"][2]),
            "heldout_coverage": fmt(model["heldout_coverage"]), "heldout_matched": len(model["held_match"]), "heldout_total": len(result["held"]),
            "star_heldout_coverage": fmt(held_class["STAR"][2]), "planet_moon_heldout_coverage": fmt(held_class["PLANET_MOON"][2]),
            "pipeline_id": model["pipeline"].pipeline_id, "preprocessing_rules": model["pipeline"].rule_string,
            "substitution_table": ";".join(f"{source}->{'+'.join(target)}" for source, target in model["mapping"]),
            "mapping_size": model["mapping_size"], "total_complexity": model["total_complexity"],
            "regularized_score": fmt(model["score"]), "null_advantage": fmt(model["null_advantage"]),
            "empirical_p": fmt(model["empirical_p"]),
        })
    return rows


def write_d1_details(result: dict[str, object]) -> None:
    models = model_rows(result)
    write_tsv(OUT / "D1_TOKEN_FORMATION_MODELS.tsv", list(models[0]), models)
    assignments: list[dict[str, object]] = []
    held_rows: list[dict[str, object]] = []
    for model in result["models"]:
        for split, labels, match, adj, evidence in (
            ("TRAIN", result["train"], model["train_match"], model["adjacency"], model["evidence"]),
            ("HELD_OUT", result["held"], model["held_match"], model["held_adjacency"], model["held_evidence"]),
        ):
            for label in labels:
                lid = label["stolfi_coordinate"]
                for object_id in sorted(adj.get(lid, set())):
                    forms = sorted({item[0] for item in evidence[(lid, object_id)]})
                    assignments.append({"model_id": model["model_id"], "split": split, "label": lid,
                                        "object_class": label["object_class"], "token": label["voynich_token"],
                                        "object_id": object_id, "historical_forms": ";".join(forms),
                                        "canonical_assignment": "1" if match.get(lid) == object_id else "0"})
                if split == "HELD_OUT":
                    held_rows.append({"model_id": model["model_id"], "label": lid,
                                      "object_class": label["object_class"], "token": label["voynich_token"],
                                      "prediction": "MATCHED" if lid in match else "UNEXPLAINED",
                                      "canonical_object_id": match.get(lid, ""),
                                      "candidate_object_count": len(adj.get(lid, set()))})
    write_tsv(OUT / "D1_LABEL_TERM_ASSIGNMENTS.tsv",
              ["model_id", "split", "label", "object_class", "token", "object_id", "historical_forms", "canonical_assignment"], assignments)
    write_tsv(OUT / "D1_HELDOUT_RESULTS.tsv",
              ["model_id", "label", "object_class", "token", "prediction", "canonical_object_id", "candidate_object_count"], held_rows)
    null_rows = [{**row, "max_score": fmt(float(row["max_score"])),
                  "max_train_coverage": fmt(float(row["max_train_coverage"])),
                  "selected_model_coverage": fmt(float(row["selected_model_coverage"])),
                  "star_selected_coverage": fmt(float(row["star_selected_coverage"])),
                  "planet_moon_selected_coverage": fmt(float(row["planet_moon_selected_coverage"]))}
                 for row in sorted(result["null_rows"], key=lambda row: (str(row["control"]), int(row["replicate"])))]
    write_tsv(OUT / "D1_NULL_RESULTS.tsv", list(null_rows[0]), null_rows)


def source_for_form(rows: list[dict[str, str]], object_id: str, form: str) -> tuple[str, str]:
    candidates = [row for row in rows if row["term_id"] == object_id and row["normalized_form"] == form]
    if not candidates:
        return "", ""
    row = candidates[0]
    return row["source"], row["source_location"]


def write_comparison(d0: dict[str, object], d1: dict[str, object], evidence: list[dict[str, str]]) -> None:
    comparison = []
    for result in (d0, d1):
        best = result["best"]
        train_class = per_class(best, result["train"], "train_match")
        held_class = per_class(best, result["held"], "held_match")
        null_values = [float(row["max_train_coverage"]) for row in result["null_rows"]]
        comparison.append({
            "run": result["run"], "concepts": len(result["terms"]),
            "forms": sum(len(term["forms"]) for term in result["terms"]),
            "best_train_coverage": fmt(best["train_coverage"]),
            "star_train_coverage": fmt(train_class["STAR"][2]),
            "planet_moon_train_coverage": fmt(train_class["PLANET_MOON"][2]),
            "best_heldout_coverage": fmt(best["heldout_coverage"]),
            "star_heldout_coverage": fmt(held_class["STAR"][2]),
            "planet_moon_heldout_coverage": fmt(held_class["PLANET_MOON"][2]),
            "models_coverage_ge_70": sum(model["train_coverage"] >= .70 for model in result["models"]),
            "mapping_size": best["mapping_size"], "complexity": best["total_complexity"],
            "empirical_p": fmt(best["empirical_p"]),
            "null_mean_maximum": fmt(statistics.fmean(null_values)),
            "null_p95": fmt(percentile(null_values, .95)), "null_maximum": fmt(max(null_values)),
            "real_null_advantage": fmt(best["null_advantage"]),
        })
    write_tsv(OUT / "DICTIONARY_D0_D1_COMPARISON.tsv", list(comparison[0]), comparison)

    d0_best, d1_best = d0["best"], d1["best"]
    newly = []
    for split, labels, d0_match, d1_match, d1_evidence in (
        ("TRAIN", d1["train"], d0_best["train_match"], d1_best["train_match"], d1_best["evidence"]),
        ("HELD_OUT", d1["held"], d0_best["held_match"], d1_best["held_match"], d1_best["held_evidence"]),
    ):
        by_id = {row["stolfi_coordinate"]: row for row in labels}
        for label, object_id in sorted(d1_match.items()):
            if label in d0_match:
                continue
            historical_form = sorted({item[0] for item in d1_evidence[(label, object_id)]})[0]
            source, location = source_for_form(evidence, object_id, historical_form)
            newly.append({"split": split, "label": label, "token": by_id[label]["voynich_token"],
                          "d1_matched_concept": object_id, "historical_form": historical_form,
                          "source": source, "source_location": location, "model_id": d1_best["model_id"]})
    write_tsv(OUT / "D1_NEWLY_EXPLAINED_LABELS.tsv",
              ["split", "label", "token", "d1_matched_concept", "historical_form", "source", "source_location", "model_id"], newly)

    real_gain = d1_best["train_coverage"] - d0_best["train_coverage"]
    null_gain = d1["conservative_mean"] - d0["conservative_mean"]
    advantage_gain = d1_best["null_advantage"] - d0_best["null_advantage"]
    if advantage_gain >= .10 and d1_best["empirical_p"] <= .05 and d1_best["heldout_coverage"] > 0:
        status = "STRONG_SIGNAL"
    elif advantage_gain > .05 and real_gain > null_gain:
        status = "SIGNAL_IMPROVED"
    elif real_gain > 0 and advantage_gain <= .05:
        status = "NULL_COMPATIBLE_GAIN"
    else:
        status = "NO_EFFECT"
    report = f"""# D0/D1 astronomical dictionary comparison

## Frozen design

D0 and D1 were rerun with the audited M1 implementation, the same 20/6 split,
640 pipelines, source/target segmentation, substitution constraints, beam width
64, regularized score, thresholds, and search-level controls. Only the term
corpus differs. Every D1 null replicate receives the complete D1 concept/form
inventory, so dictionary size, variants, lengths, language mix, and class mix
are inherited by the randomized corpus. STAR and PLANET_MOON results are also
reported separately as required by the audit.

PLANET_MOON remains advisory: the five f67r2 labels are anonymous diagram
positions and may not denote five unique objects. Both object-disjoint and
form/rule-disjoint HELD_OUT semantics were equivalent in the prerequisite audit.

## Result

D0 matched {len(d0_best['train_match'])}/20 TRAIN and {len(d0_best['held_match'])}/6
HELD_OUT; D1 matched {len(d1_best['train_match'])}/20 TRAIN and
{len(d1_best['held_match'])}/6 HELD_OUT. The real TRAIN gain is {real_gain:.6f};
the conservative null-mean gain is {null_gain:.6f}; the change in real-null
advantage is {advantage_gain:.6f}. “Newly explained” means membership in the
top D1 model's canonical maximum matching but not the top D0 model's canonical
matching; anonymous assignments are not object identifications.

## Final status

```text
ASTRO_DICTIONARY_EXPANSION={status}

D0_TRAIN_COVERAGE={d0_best['train_coverage']:.6f}
D1_TRAIN_COVERAGE={d1_best['train_coverage']:.6f}

D0_HELDOUT_COVERAGE={d0_best['heldout_coverage']:.6f}
D1_HELDOUT_COVERAGE={d1_best['heldout_coverage']:.6f}

D0_NULL_ADVANTAGE={d0_best['null_advantage']:.6f}
D1_NULL_ADVANTAGE={d1_best['null_advantage']:.6f}

D0_EMPIRICAL_P={d0_best['empirical_p']:.6f}
D1_EMPIRICAL_P={d1_best['empirical_p']:.6f}

D1_MODELS_WITH_COVERAGE_GE_70={sum(model['train_coverage'] >= .70 for model in d1['models'])}
NEWLY_EXPLAINED_LABELS={len(newly)}
```

The result tests lexical coverage only. It neither translates labels nor
identifies astronomical objects without independent object-level evidence.
"""
    (OUT / "D1_TOKEN_FORMATION_REPORT.md").write_text(report, encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--workers", type=int, default=min(4, os.cpu_count() or 1))
    parser.add_argument("--null-replicates", type=int, default=m1.NULL_REPLICATES)
    parser.add_argument("--observed-only", action="store_true")
    args = parser.parse_args()
    evidence = read_tsv(OUT / "ASTRO_TERM_CORPUS_EXPANDED.tsv")
    d0_terms, d1_terms = load_d0(), load_d1()
    corpus_audit(evidence, d0_terms, d1_terms)
    if args.observed_only:
        for run, terms in (("D0", d0_terms), ("D1", d1_terms)):
            best = m1.run_search(m1.load_labels("TRAIN"), terms, retain_details=True)[0]
            print(run, best["train_coverage"], best["score"], best["mapping_size"], best["pipeline"].rule_string)
        return
    d0 = run_one("D0", d0_terms, args.null_replicates, args.workers)
    d1 = run_one("D1", d1_terms, args.null_replicates, args.workers)
    write_d1_details(d1)
    write_comparison(d0, d1, evidence)
    artifacts = ["ASTRO_TERM_CORPUS_EXPANDED.tsv", "ASTRO_TERM_CORPUS_EXPANSION_SOURCES.md",
                 "ASTRO_TERM_CORPUS_AUDIT.md", "DICTIONARY_D0_D1_COMPARISON.tsv",
                 "D1_TOKEN_FORMATION_MODELS.tsv", "D1_LABEL_TERM_ASSIGNMENTS.tsv",
                 "D1_HELDOUT_RESULTS.tsv", "D1_NULL_RESULTS.tsv",
                 "D1_NEWLY_EXPLAINED_LABELS.tsv", "D1_TOKEN_FORMATION_REPORT.md"]
    manifest = {
        "experiment": "astro-dictionary-expansion-m1-v2", "generated_utc": datetime.now(timezone.utc).replace(microsecond=0).isoformat(),
        "seed": SEED, "null_replicates_per_control": args.null_replicates,
        "dictionary_expansion_authorized": True, "audit_verdict": "VALID_WITH_LIMITATIONS",
        "frozen_protocol": {"pipelines": 640, "beam_width": m1.BEAM_WIDTH,
                            "finalists_per_pipeline": m1.FINALISTS_PER_PIPELINE,
                            "retained_models": m1.RETAINED, "controls": list(CONTROLS)},
        "input_sha256": {
            "research/astro_token_formation/ASTRO_TERM_CORPUS.tsv": sha(ROOT / "research/astro_token_formation/ASTRO_TERM_CORPUS.tsv"),
            "research/astro_token_formation/ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv": sha(ROOT / "research/astro_token_formation/ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv"),
            "research/astro_token_formation_m1/main.py": sha(M1_PATH),
            "research/astro_token_formation_m1_audit/M1_METHODOLOGICAL_AUDIT.md": sha(ROOT / "research/astro_token_formation_m1_audit/M1_METHODOLOGICAL_AUDIT.md"),
            "research/astro_dictionary_expansion_m1/main.py": sha(Path(__file__)),
        },
        "artifact_sha256": {name: sha(OUT / name) for name in artifacts},
    }
    (OUT / "manifest.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    checksum_names = artifacts + ["main.py", "manifest.json"]
    (OUT / "SHA256SUMS").write_text("".join(f"{sha(OUT / name)}  {name}\n" for name in sorted(checksum_names)), encoding="utf-8")


if __name__ == "__main__":
    main()
