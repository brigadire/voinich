#!/usr/bin/env python3
"""AUDIT_ONLY: per-class null coverage using the first 30 of the frozen 100
seeds per control (same seed_for() function as production M1, no new corpus,
no new rules). Purpose: contextualize the STAR-only vs PLANET_MOON-only TRAIN
coverage split found in the combined-statistic audit (M1_AUDIT_DIAGNOSTICS.tsv,
PER_CLASS_BREAKDOWN rows), since M1's own null replicates only recorded the
aggregate 20-label statistic.
"""
from __future__ import annotations

import csv
import importlib.util
import json
import random
import statistics
import sys
from collections import defaultdict
from multiprocessing import get_context
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "research/astro_token_formation_m1_audit"
M1_PATH = ROOT / "research/astro_token_formation_m1/main.py"

_spec = importlib.util.spec_from_file_location("astro_m1_for_audit2", M1_PATH)
assert _spec and _spec.loader
m1 = importlib.util.module_from_spec(_spec)
sys.modules[_spec.name] = m1
_spec.loader.exec_module(m1)

N_REPLICATES = 30


def worker(task: tuple[str, int]) -> dict[str, object]:
    control, replicate = task
    rng = random.Random(m1.seed_for(control, replicate))
    labels = m1.load_labels("TRAIN")
    terms = m1.load_terms()
    if control == "RANDOM_VOYNICH_SET":
        labels = m1.randomize_labels(labels, rng)
    elif control == "SHUFFLED_TERMS":
        terms = m1.alter_terms(terms, rng, False)
    elif control == "PSEUDODICTIONARY":
        terms = m1.alter_terms(terms, rng, True)
    models = m1.run_search(labels, terms, retain_details=True)
    best = max(models, key=lambda x: (x["score"], x["train_coverage"], -x["mapping_size"]))
    star_ids = [r["stolfi_coordinate"] for r in labels if r["object_class"] == "STAR"]
    planet_ids = [r["stolfi_coordinate"] for r in labels if r["object_class"] == "PLANET_MOON"]
    star_matched = sum(1 for lid in star_ids if lid in best["train_match"])
    planet_matched = sum(1 for lid in planet_ids if lid in best["train_match"])
    return {
        "control": control, "replicate": replicate, "seed": m1.seed_for(control, replicate),
        "overall_coverage": best["train_coverage"], "star_matched": star_matched, "star_total": len(star_ids),
        "planet_matched": planet_matched, "planet_total": len(planet_ids),
        "star_coverage": star_matched / len(star_ids) if star_ids else None,
        "planet_coverage": planet_matched / len(planet_ids) if planet_ids else None,
    }


def main() -> None:
    tasks = [(c, r) for c in ("RANDOM_VOYNICH_SET", "SHUFFLED_TERMS", "PSEUDODICTIONARY") for r in range(N_REPLICATES)]
    with get_context("fork").Pool(12) as pool:
        rows = list(pool.imap(worker, tasks, chunksize=1))
    with (OUT / "M1_AUDIT_PERCLASS_NULL.tsv").open("w", encoding="utf-8", newline="") as fh:
        fields = ["run_type"] + list(rows[0])
        w = csv.DictWriter(fh, fields, delimiter="\t", lineterminator="\n")
        w.writeheader()
        for r in sorted(rows, key=lambda x: (x["control"], x["replicate"])):
            w.writerow({"run_type": "AUDIT_ONLY", **r})
    by_control = defaultdict(list)
    for r in rows:
        by_control[r["control"]].append(r)
    print(f"n={N_REPLICATES} per control (subset of the frozen 100 seeds, seeds 0..{N_REPLICATES - 1})")
    for control, rs in sorted(by_control.items()):
        star_cov = [r["star_coverage"] for r in rs]
        planet_cov = [r["planet_coverage"] for r in rs]
        overall = [r["overall_coverage"] for r in rs]
        print(control,
              "overall_mean=", round(statistics.fmean(overall), 4), "overall_max=", max(overall),
              "| star_mean=", round(statistics.fmean(star_cov), 4), "star_max=", max(star_cov),
              "frac_star_ge_0.4375=", round(sum(x >= 0.4375 - 1e-9 for x in star_cov) / len(star_cov), 4),
              "| planet_mean=", round(statistics.fmean(planet_cov), 4), "planet_max=", max(planet_cov),
              "frac_planet_ge_0.75=", round(sum(x >= 0.75 - 1e-9 for x in planet_cov) / len(planet_cov), 4))


if __name__ == "__main__":
    main()
