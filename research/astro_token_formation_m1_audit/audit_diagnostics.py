#!/usr/bin/env python3
"""AUDIT_ONLY diagnostics for the M1 methodological audit.

This script performs NO brute-force search expansion and makes NO changes to
any frozen M0/M1 artifact. It re-imports the frozen M1 search code as a
library and reruns bounded, already-defined computations (reproducibility
checks, beam-width sensitivity, alternative held-out semantics, per-class
breakdowns, capacity-ceiling arithmetic) purely for auditing purposes.

All rows emitted to M1_AUDIT_DIAGNOSTICS.tsv are tagged run_type=AUDIT_ONLY.
"""
from __future__ import annotations

import csv
import importlib.util
import json
import statistics
import sys
import time
from collections import Counter, defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "research/astro_token_formation_m1_audit"
M1_PATH = ROOT / "research/astro_token_formation_m1/main.py"

_spec = importlib.util.spec_from_file_location("astro_m1_for_audit", M1_PATH)
assert _spec and _spec.loader
m1 = importlib.util.module_from_spec(_spec)
sys.modules[_spec.name] = m1
_spec.loader.exec_module(m1)

diagnostic_rows: list[dict[str, object]] = []


def log(row: dict[str, object]) -> None:
    row = {"run_type": "AUDIT_ONLY", **row}
    diagnostic_rows.append(row)
    print(json.dumps(row, default=str)[:220])


def main() -> None:
    train = m1.load_labels("TRAIN")
    held = m1.load_labels("HELD_OUT")
    terms = m1.load_terms()
    assert len(train) == 20 and len(held) == 6

    # ------------------------------------------------------------------
    # 1. Reproducibility: rerun the frozen search at the frozen beam width
    #    and diff against the committed M1_TOKEN_FORMATION_MODELS.tsv.
    # ------------------------------------------------------------------
    t0 = time.time()
    m1.BEAM_WIDTH = 64
    observed = m1.run_search(train, terms, retain_details=True)
    elapsed = time.time() - t0
    frozen_models = list(csv.DictReader(open(OUT.parent / "astro_token_formation_m1/M1_TOKEN_FORMATION_MODELS.tsv"), delimiter="\t"))
    frozen_top = frozen_models[0]
    rerun_top = observed[0]
    reproduced = (
        abs(rerun_top["train_coverage"] - float(frozen_top["train_coverage"])) < 1e-9
        and rerun_top["mapping_size"] == int(frozen_top["mapping_size"])
        and rerun_top["pipeline"].rule_string == frozen_top["preprocessing_rules"]
    )
    log({
        "check": "REPRODUCIBILITY_BEAM64", "beam_width": 64, "seconds": round(elapsed, 2),
        "rerun_train_coverage": rerun_top["train_coverage"], "frozen_train_coverage": frozen_top["train_coverage"],
        "rerun_mapping_size": rerun_top["mapping_size"], "frozen_mapping_size": frozen_top["mapping_size"],
        "result": "MATCH" if reproduced else "MISMATCH",
    })

    # ------------------------------------------------------------------
    # 2. Beam-width sensitivity. Same pipelines/labels/corpus, only the
    #    beam width changes. Looks for any width that crosses 0.55 TRAIN
    #    coverage (i.e. beats the frozen result) or changes the winning
    #    rule family.
    # ------------------------------------------------------------------
    for width in (8, 16, 32, 64, 128, 256):
        m1.BEAM_WIDTH = width
        t0 = time.time()
        models = m1.run_search(train, terms, retain_details=True)
        elapsed = time.time() - t0
        best = models[0]
        top_rules = sorted({m["pipeline"].rule_string for m in models[:10]})
        log({
            "check": "BEAM_WIDTH_SWEEP", "beam_width": width, "seconds": round(elapsed, 2),
            "best_train_coverage": best["train_coverage"], "best_mapping_size": best["mapping_size"],
            "best_rule": best["pipeline"].rule_string, "n_distinct_rule_families_top10": len(top_rules),
        })
    m1.BEAM_WIDTH = 64

    # ------------------------------------------------------------------
    # 3. Label-visit-order sensitivity inside search_pipeline. The frozen
    #    order is (edge_count asc, label_id asc). Test reversed order and
    #    alphabetical-only order to see if the beam's greedy commitment
    #    order changes the outcome (multiple local optima).
    # ------------------------------------------------------------------
    orig_search_pipeline = m1.search_pipeline

    def search_pipeline_custom_order(labels, terms_, pipe, pre, order_fn):
        edges = m1.build_edges(labels, terms_, pre)
        order = order_fn(labels, edges)
        beam = [m1.State((), frozenset(), ())]
        for label_id in order:
            candidates = {}
            for state in beam:
                key = (state.mapping, state.used_objects)
                candidates.setdefault(key, state)
                for edge in edges[label_id]:
                    if edge.object_id in state.used_objects:
                        continue
                    merged = m1.compatible_merge(state.mapping, edge.constraint)
                    if merged is None:
                        continue
                    assignment = (edge.label, edge.object_id, edge.source_form, edge.source_units)
                    new = m1.State(merged, state.used_objects | {edge.object_id}, state.assignments + (assignment,))
                    key = (new.mapping, new.used_objects)
                    old = candidates.get(key)
                    if old is None or new.assignments < old.assignments:
                        candidates[key] = new
            beam = sorted(candidates.values(), key=lambda s: m1.state_rank(s, len(labels), pipe))[: m1.BEAM_WIDTH]
        return sorted(beam, key=lambda s: m1.state_rank(s, len(labels), pipe))[: m1.FINALISTS_PER_PIPELINE]

    def run_search_custom_order(labels, terms_, order_fn):
        representatives = {}
        for pipe in m1.all_pipelines():
            pre = m1.preprocess(terms_, pipe)
            key = m1.representation_key(pre)
            old = representatives.get(key)
            if old is None or (pipe.complexity, pipe.rule_string) < (old[0].complexity, old[0].rule_string):
                representatives[key] = (pipe, pre)
        raw = []
        for pipe, pre in sorted(representatives.values(), key=lambda x: x[0].pipeline_id):
            for state in search_pipeline_custom_order(labels, terms_, pipe, pre, order_fn):
                raw.append(m1.evaluate_state(labels, terms_, pipe, pre, state))
        raw.sort(key=lambda x: (-x["train_coverage"], -x["score"], x["mapping_size"], x["pipeline"].rule_string, x["mapping"]))
        return raw[:5]

    orderings = {
        "FROZEN_EDGECOUNT_ASC": lambda labels, edges: sorted((r["stolfi_coordinate"] for r in labels), key=lambda x: (len(edges[x]), x)),
        "EDGECOUNT_DESC": lambda labels, edges: sorted((r["stolfi_coordinate"] for r in labels), key=lambda x: (-len(edges[x]), x)),
        "COORDINATE_ASC": lambda labels, edges: sorted(r["stolfi_coordinate"] for r in labels),
        "COORDINATE_DESC": lambda labels, edges: sorted((r["stolfi_coordinate"] for r in labels), reverse=True),
    }
    for name, fn in orderings.items():
        t0 = time.time()
        top = run_search_custom_order(train, terms, fn)
        elapsed = time.time() - t0
        best = top[0]
        log({
            "check": "VISIT_ORDER_SWEEP", "order": name, "beam_width": 64, "seconds": round(elapsed, 2),
            "best_train_coverage": best["train_coverage"], "best_mapping_size": best["mapping_size"],
            "best_rule": best["pipeline"].rule_string,
        })

    # ------------------------------------------------------------------
    # 4. Dictionary capacity ceiling per class.
    # ------------------------------------------------------------------
    by_class_objects = Counter(str(t["object_class"]) for t in terms)
    all_selected = [r for r in csv.DictReader(open(ROOT / "research/astro_token_formation/ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv"), delimiter="\t") if r["selected"] == "1"]
    by_class_labels = Counter(r["object_class"] for r in all_selected)
    by_split_class = Counter((r["object_class"], r["split"]) for r in all_selected)
    for cls in ("STAR", "PLANET_MOON"):
        n_obj = by_class_objects[cls]
        n_lab = by_class_labels[cls]
        n_train = by_split_class[(cls, "TRAIN")]
        n_held = by_split_class[(cls, "HELD_OUT")]
        ceiling_joint = min(n_obj, n_lab)
        ceiling_train = min(n_obj, n_train)
        ceiling_held_if_train_uses_max = max(0, min(n_held, n_obj - min(n_obj, n_train)))
        log({
            "check": "DICTIONARY_CAPACITY_CEILING", "class": cls, "historical_objects": n_obj,
            "sampled_labels": n_lab, "train_labels": n_train, "held_labels": n_held,
            "joint_ceiling_pct": round(100 * ceiling_joint / n_lab, 1),
            "train_only_ceiling_pct": round(100 * ceiling_train / n_train, 1),
            "worst_case_held_ceiling_if_train_saturates": ceiling_held_if_train_uses_max,
            "zero_slack": n_obj == n_lab,
        })

    # ------------------------------------------------------------------
    # 5. Alternative HELD_OUT semantics: FORM/RULE-DISJOINT (object reuse
    #    allowed) vs the frozen OBJECT-DISJOINT rule, applied to the same
    #    frozen mappings recovered by the reproducibility rerun above.
    # ------------------------------------------------------------------
    m1.BEAM_WIDTH = 64
    observed = m1.run_search(train, terms, retain_details=True)
    for rank, model in enumerate(observed[:10], 1):
        used = set(model["train_match"].values())
        adj_obj_disjoint, _ = m1.adjacency(held, terms, model["pre"], model["mapping"], unavailable=used)
        match_obj_disjoint = m1.maximum_matching(adj_obj_disjoint)
        adj_form_disjoint, _ = m1.adjacency(held, terms, model["pre"], model["mapping"], unavailable=None)
        match_form_disjoint = m1.maximum_matching(adj_form_disjoint)
        log({
            "check": "HELDOUT_SEMANTICS_COMPARISON", "model_rank": rank,
            "pipeline": model["pipeline"].rule_string, "mapping_size": model["mapping_size"],
            "train_coverage": model["train_coverage"],
            "object_disjoint_heldout_coverage": len(match_obj_disjoint) / len(held),
            "form_rule_disjoint_heldout_coverage": len(match_form_disjoint) / len(held),
            "object_disjoint_matched": sorted(match_obj_disjoint),
            "form_rule_disjoint_matched": sorted(match_form_disjoint),
        })

    # ------------------------------------------------------------------
    # 6. Per-class TRAIN/HELD_OUT coverage breakdown for the top models
    #    (STAR evaluated independently from PLANET_MOON).
    # ------------------------------------------------------------------
    train_by_id = {r["stolfi_coordinate"]: r for r in train}
    held_by_id = {r["stolfi_coordinate"]: r for r in held}
    for rank, model in enumerate(observed[:10], 1):
        for cls in ("STAR", "PLANET_MOON"):
            train_ids = [r["stolfi_coordinate"] for r in train if r["object_class"] == cls]
            held_ids = [r["stolfi_coordinate"] for r in held if r["object_class"] == cls]
            t_matched = sum(1 for lid in train_ids if lid in model["train_match"])
            h_matched = sum(1 for lid in held_ids if lid in model["held_match"]) if "held_match" in model else None
            log({
                "check": "PER_CLASS_BREAKDOWN", "model_rank": rank, "pipeline": model["pipeline"].rule_string,
                "class": cls, "train_matched": t_matched, "train_total": len(train_ids),
                "train_coverage_class": round(t_matched / len(train_ids), 4) if train_ids else None,
                "held_total": len(held_ids),
            })

    # ------------------------------------------------------------------
    # 7. Split leakage checks (independent of freeze_split.py review).
    # ------------------------------------------------------------------
    tokens = [r["voynich_token"] for r in all_selected]
    dup_tokens = [t for t, c in Counter(tokens).items() if c > 1]
    log({
        "check": "SPLIT_LEAKAGE_TOKEN_DUPLICATES", "n_selected": len(all_selected),
        "n_distinct_tokens": len(set(tokens)), "duplicate_tokens": dup_tokens,
    })

    # ------------------------------------------------------------------
    # 8. Null-design cross-check: recompute the three controls' empirical
    #    max-coverage summary directly from the frozen M1_NULL_RESULTS.tsv
    #    (no new replicates run; this only re-derives the aggregate stats
    #    already implicit in the frozen file, for the audit report).
    # ------------------------------------------------------------------
    null_rows = list(csv.DictReader(open(ROOT / "research/astro_token_formation_m1/M1_NULL_RESULTS.tsv"), delimiter="\t"))
    by_control = defaultdict(list)
    for r in null_rows:
        by_control[r["control"]].append(float(r["max_train_coverage"]))
    for control, values in by_control.items():
        log({
            "check": "NULL_SUMMARY_CROSSCHECK", "control": control, "n": len(values),
            "mean": round(statistics.fmean(values), 4), "max": max(values), "min": min(values),
            "frac_ge_050": round(sum(v >= 0.50 for v in values) / len(values), 4),
        })

    fields = sorted({k for row in diagnostic_rows for k in row})
    with (OUT / "M1_AUDIT_DIAGNOSTICS.tsv").open("w", encoding="utf-8", newline="") as fh:
        writer = csv.DictWriter(fh, fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        writer.writeheader()
        for row in diagnostic_rows:
            writer.writerow({k: json.dumps(v) if isinstance(v, (list, dict)) else v for k, v in row.items()})
    print(f"\nWrote {len(diagnostic_rows)} diagnostic rows to {OUT / 'M1_AUDIT_DIAGNOSTICS.tsv'}")


if __name__ == "__main__":
    main()
