#!/usr/bin/env python3
"""Reproduce the matching-selection bias audit for Stolfi astro labels."""

from __future__ import annotations

import csv
import hashlib
import importlib.util
import json
import math
import platform
import re
import sys
from collections import Counter, defaultdict
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "research/stolfi_matching_bias"
MATCHES = ROOT / "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv"
LABEL_AUDIT = ROOT / "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_AUDIT.md"
ENRICHMENT = ROOT / "research/stolfi_label_hapax_enrichment/STOLFI_ASTRO_LABEL_HAPAX_ENRICHMENT.tsv"
BY_PANEL_HAPAX = ROOT / "research/stolfi_label_hapax_enrichment/STOLFI_ASTRO_LABEL_HAPAX_BY_PANEL.tsv"
ENRICHMENT_IMPL = ROOT / "research/stolfi_label_hapax_enrichment/main.py"

AUDIT_TSV = OUT / "STOLFI_MATCHING_BIAS_AUDIT.tsv"
BY_PANEL_FAMILY = OUT / "STOLFI_MATCHING_BIAS_BY_PANEL_FAMILY.tsv"
SENSITIVITY = OUT / "STOLFI_MATCHING_BIAS_SENSITIVITY.tsv"
REPORT = OUT / "STOLFI_MATCHING_BIAS_REPORT.md"
MANIFEST = OUT / "STOLFI_MATCHING_BIAS_MANIFEST.json"
CHECKSUMS = OUT / "STOLFI_MATCHING_BIAS_SHA256SUMS"

PERMUTATIONS = 10_000
SEED = 20260901


def load_enrichment_module():
    spec = importlib.util.spec_from_file_location("stolfi_enrichment", ENRICHMENT_IMPL)
    if spec is None or spec.loader is None:
        raise RuntimeError("cannot load enrichment implementation")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def fmt(value: float) -> str:
    return f"{value:.9f}"


def write_tsv(path: Path, header: list[str], rows: list[list[object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.writer(handle, delimiter="\t", lineterminator="\n")
        writer.writerow(header)
        writer.writerows(rows)


def family(row: dict[str, str]) -> str:
    if row["object_type"] == "star":
        return "STAR"
    if row["object_type"] == "planet?":
        return "PLANET_MOON"
    if row["panel"] == "f67r1" and row["stolfi_group"] == "S":
        return "CIRCLE_SECTOR"
    if row["panel"] == "f68v2" and row["stolfi_group"] == "R":
        return "RADIAL_TEXT"
    if row["panel"] == "f68v2" and row["stolfi_group"] in {"X", "Y"}:
        return "SECTOR_LABEL"
    if row["panel"] == "f68v2" and row["stolfi_group"] == "Z":
        return "OUTER_TITLE"
    return "OTHER_DIAGRAM_LABEL"


def label_tokens(eva: str) -> list[str]:
    return [token for token in re.split(r"[\s.,-]+", eva.strip()) if token]


def glyph_length(eva: str) -> int:
    return len(re.sub(r"[\s.,*?-]", "", eva))


def cramers_v(categories: list[str], statuses: list[int]) -> float:
    table = defaultdict(lambda: [0, 0])
    for category, status in zip(categories, statuses):
        table[category][status] += 1
    totals = [sum(row) for row in table.values()]
    columns = [sum(row[index] for row in table.values()) for index in (0, 1)]
    n = sum(totals)
    chi2 = 0.0
    for total, row in zip(totals, table.values()):
        for index in (0, 1):
            expected = total * columns[index] / n
            if expected:
                chi2 += (row[index] - expected) ** 2 / expected
    return math.sqrt(chi2 / n) if n else 0.0


def category_summary(values: list[str], statuses: list[int], wanted: int) -> str:
    counts = Counter(value for value, status in zip(values, statuses) if status == wanted)
    return ";".join(f"{key}={counts[key]}" for key in sorted(counts))


def mean_difference(values: list[float], statuses: list[int]) -> float:
    matched = [value for value, status in zip(values, statuses) if status]
    unmatched = [value for value, status in zip(values, statuses) if not status]
    return sum(matched) / len(matched) - sum(unmatched) / len(unmatched)


def standardized_mean_difference(values: list[float], statuses: list[int]) -> float:
    matched = [value for value, status in zip(values, statuses) if status]
    unmatched = [value for value, status in zip(values, statuses) if not status]
    mean_m = sum(matched) / len(matched)
    mean_u = sum(unmatched) / len(unmatched)
    variance_m = sum((value - mean_m) ** 2 for value in matched) / (len(matched) - 1)
    variance_u = sum((value - mean_u) ** 2 for value in unmatched) / (len(unmatched) - 1)
    pooled = math.sqrt(
        ((len(matched) - 1) * variance_m + (len(unmatched) - 1) * variance_u)
        / (len(matched) + len(unmatched) - 2)
    )
    return (mean_m - mean_u) / pooled if pooled else 0.0


def permutation_p(name: str, observed: float, statistic, values, statuses, enrichment) -> float:
    rng = enrichment.SplitMix64(enrichment.stream_seed(f"BIAS:{name}"))
    extreme = 0
    for _ in range(PERMUTATIONS):
        shuffled = statuses.copy()
        for index in range(len(shuffled) - 1, 0, -1):
            other = rng.randbelow(index + 1)
            shuffled[index], shuffled[other] = shuffled[other], shuffled[index]
        if abs(statistic(values, shuffled)) >= abs(observed) - 1e-15:
            extreme += 1
    return (1 + extreme) / (PERMUTATIONS + 1)


def main() -> None:
    enrichment = load_enrichment_module()
    occurrences, by_panel, labels, families, hapax = enrichment.load_inputs()
    with MATCHES.open(encoding="utf-8", newline="") as handle:
        records = list(csv.DictReader(handle, delimiter="\t"))
    if Counter(row["match_status"] for row in records) != Counter({"MATCHED": 130, "UNMATCHED": 61}):
        raise RuntimeError("unexpected record status counts")

    grouped = defaultdict(list)
    for row in records:
        grouped[row["stolfi_coordinate"]].append(row)
    coordinates = []
    for coordinate, variants in sorted(grouped.items()):
        statuses = {row["match_status"] for row in variants}
        if len(statuses) != 1:
            raise RuntimeError(f"mixed status at {coordinate}")
        token_lengths = [len(label_tokens(row["stolfi_eva"])) for row in variants]
        glyph_lengths = [glyph_length(row["stolfi_eva"]) for row in variants]
        coordinates.append({
            "coordinate": coordinate,
            "panel": variants[0]["panel"],
            "family": family(variants[0]),
            "group": variants[0]["stolfi_group"],
            "series": f"{variants[0]['panel']}.{variants[0]['stolfi_group']}",
            "matched": variants[0]["match_status"] == "MATCHED",
            "token_length_mean": sum(token_lengths) / len(token_lengths),
            "potential_tokens": max(token_lengths),
            "glyph_length_mean": sum(glyph_lengths) / len(glyph_lengths),
            "wildcard": any("*" in row["stolfi_eva"] for row in variants),
            "variants": len(variants),
            "exact_anchor": any(int(row["raw_lexical_candidate_count"]) > 0 for row in variants),
            "bridge": any(bool(row["validated_ivtff_locus_type"]) for row in variants),
            "validated_type": next((row["validated_ivtff_locus_type"] for row in variants
                                    if row["validated_ivtff_locus_type"]), "NONE"),
            "mapped_type": next((row["zl3b_locus_type"] for row in variants if row["zl3b_locus_type"]), "NONE"),
        })
    if len(coordinates) != 143 or sum(row["matched"] for row in coordinates) != 89:
        raise RuntimeError("unexpected physical-coordinate counts")

    audit_rows = []
    statuses = [int(row["matched"]) for row in coordinates]
    matched_n = sum(statuses)
    unmatched_n = len(statuses) - matched_n

    categorical = {
        "PANEL": [row["panel"] for row in coordinates],
        "LABEL_FAMILY": [row["family"] for row in coordinates],
        "STOLFI_GROUP": [row["group"] for row in coordinates],
        "PANEL_SERIES": [row["series"] for row in coordinates],
        "WILDCARD_PRESENT": [str(row["wildcard"]) for row in coordinates],
        "EXACT_LEXICAL_ANCHOR": [str(row["exact_anchor"]) for row in coordinates],
        "COORDINATE_BRIDGE_AVAILABLE": [str(row["bridge"]) for row in coordinates],
        "VALIDATED_IVTFF_LOCUS_TYPE": [row["validated_type"] for row in coordinates],
    }
    for name, values in categorical.items():
        effect = cramers_v(values, statuses)
        p_value = permutation_p(name, effect, cramers_v, values, statuses, enrichment)
        audit_rows.append([
            "PHYSICAL_COORDINATE", name, matched_n, unmatched_n, "CRAMERS_V", fmt(effect), fmt(p_value),
            "STRONG" if effect >= 0.3 else "MODERATE" if effect >= 0.2 else "WEAK",
            "pre-hapax feature; permutation test with fixed matched count",
            category_summary(values, statuses, 1), category_summary(values, statuses, 0),
        ])

    continuous = {
        "LABEL_LENGTH_GLYPHS": [row["glyph_length_mean"] for row in coordinates],
        "LABEL_LENGTH_TOKENS": [row["token_length_mean"] for row in coordinates],
        "TRANSCRIBER_VARIANTS": [float(row["variants"]) for row in coordinates],
    }
    for name, values in continuous.items():
        raw = mean_difference(values, statuses)
        smd = standardized_mean_difference(values, statuses)
        p_value = permutation_p(name, raw, mean_difference, values, statuses, enrichment)
        audit_rows.append([
            "PHYSICAL_COORDINATE", name, matched_n, unmatched_n, "STANDARDIZED_MEAN_DIFFERENCE",
            fmt(smd), fmt(p_value),
            "STRONG" if abs(smd) >= 0.5 else "MODERATE" if abs(smd) >= 0.3 else "WEAK",
            f"matched-minus-unmatched raw mean difference={fmt(raw)}",
            f"mean={fmt(sum(value for value, status in zip(values, statuses) if status) / matched_n)}",
            f"mean={fmt(sum(value for value, status in zip(values, statuses) if not status) / unmatched_n)}",
        ])

    # Record-level composition is descriptive because transcriber variants at
    # one physical coordinate are not independent observations.
    record_status = [int(row["match_status"] == "MATCHED") for row in records]
    for name, values in {
        "PANEL": [row["panel"] for row in records],
        "LABEL_FAMILY": [family(row) for row in records],
        "EXACT_LEXICAL_ANCHOR": [str(int(row["raw_lexical_candidate_count"]) > 0) for row in records],
    }.items():
        audit_rows.append([
            "SOURCE_RECORD", name, sum(record_status), len(record_status) - sum(record_status),
            "CRAMERS_V_DESCRIPTIVE", fmt(cramers_v(values, record_status)), "NA", "DESCRIPTIVE_ONLY",
            "variants at a coordinate are dependent; no record-level p-value",
            category_summary(values, record_status, 1), category_summary(values, record_status, 0),
        ])

    # Actual mapped locus type is a consequence of matching, not an independent
    # predictor; retain it explicitly so the requested feature is not hidden.
    mapped_values = [row["mapped_type"] for row in coordinates]
    audit_rows.append([
        "PHYSICAL_COORDINATE", "MAPPED_LOCUS_TYPE_USED", matched_n, unmatched_n,
        "CRAMERS_V_DESCRIPTIVE", fmt(cramers_v(mapped_values, statuses)), "NA", "STRUCTURAL",
        "post-mapping consequence: unmatched coordinates necessarily have NONE",
        category_summary(mapped_values, statuses, 1), category_summary(mapped_values, statuses, 0),
    ])
    write_tsv(
        AUDIT_TSV,
        ["unit", "feature", "matched_n", "unmatched_n", "statistic", "effect_size", "permutation_p",
         "effect_band", "notes", "matched_summary", "unmatched_summary"],
        audit_rows,
    )

    panel_background = {
        panel: len(set(by_panel[panel]) & hapax) / len(by_panel[panel])
        for panel in enrichment.PANELS
    }
    cell_rows = []
    cells = defaultdict(list)
    for coordinate in coordinates:
        cells[(coordinate["panel"], coordinate["family"], coordinate["group"])].append(coordinate)
    record_cells = defaultdict(list)
    for row in records:
        record_cells[(row["panel"], family(row), row["stolfi_group"])].append(row)
    for key in sorted(cells):
        coords = cells[key]
        recs = record_cells[key]
        mc = sum(row["matched"] for row in coords)
        mr = sum(row["match_status"] == "MATCHED" for row in recs)
        matched_lengths = [row["glyph_length_mean"] for row in coords if row["matched"]]
        unmatched_lengths = [row["glyph_length_mean"] for row in coords if not row["matched"]]
        matched_token_lengths = [row["token_length_mean"] for row in coords if row["matched"]]
        unmatched_token_lengths = [row["token_length_mean"] for row in coords if not row["matched"]]
        matched_variants = [row["variants"] for row in coords if row["matched"]]
        unmatched_variants = [row["variants"] for row in coords if not row["matched"]]
        unmatched_potential = sum(row["potential_tokens"] for row in coords if not row["matched"])
        cell_rows.append([
            *key, len(recs), mr, len(recs) - mr, fmt(mr / len(recs)),
            len(coords), mc, len(coords) - mc, fmt(mc / len(coords)),
            sum(row["wildcard"] for row in coords), sum(row["exact_anchor"] for row in coords),
            fmt(sum(matched_lengths) / len(matched_lengths)) if matched_lengths else "NA",
            fmt(sum(unmatched_lengths) / len(unmatched_lengths)) if unmatched_lengths else "NA",
            fmt(sum(matched_token_lengths) / len(matched_token_lengths)) if matched_token_lengths else "NA",
            fmt(sum(unmatched_token_lengths) / len(unmatched_token_lengths)) if unmatched_token_lengths else "NA",
            fmt(sum(matched_variants) / len(matched_variants)) if matched_variants else "NA",
            fmt(sum(unmatched_variants) / len(unmatched_variants)) if unmatched_variants else "NA",
            ",".join(sorted({row["validated_type"] for row in coords})),
            fmt(panel_background[key[0]]), unmatched_potential,
            fmt(unmatched_potential * panel_background[key[0]]),
        ])
    write_tsv(
        BY_PANEL_FAMILY,
        ["panel", "family", "stolfi_group", "records", "matched_records", "unmatched_records",
         "record_match_rate", "physical_coordinates", "matched_coordinates", "unmatched_coordinates",
         "physical_coordinate_coverage", "coordinates_with_wildcard", "coordinates_with_exact_anchor",
         "matched_mean_glyph_length", "unmatched_mean_glyph_length", "matched_mean_token_length",
         "unmatched_mean_token_length", "matched_mean_transcriber_variants",
         "unmatched_mean_transcriber_variants", "validated_ivtff_locus_types",
         "panel_background_hapax_fraction", "unmatched_potential_token_occurrences",
         "background_expected_unmatched_hapax"],
        cell_rows,
    )

    # Subset permutation tests.
    series = defaultdict(list)
    for coordinate in coordinates:
        series[coordinate["series"]].append(coordinate)
    high_coverage_series = {
        name for name, rows in series.items()
        if sum(row["matched"] for row in rows) / len(rows) >= 0.8
    }
    high_positions = set()
    for row in records:
        if row["match_status"] == "MATCHED" and f"{row['panel']}.{row['stolfi_group']}" in high_coverage_series:
            high_positions.update(int(x) for x in row["absolute_token_positions"].split(","))
    high_result = enrichment.run_null("BIAS:HIGH_COVERAGE_SERIES", high_positions, enrichment.PANELS,
                                      by_panel, hapax)
    without_f68v2 = {position for position in labels if occurrences[position]["folio"] != "f68v2"}
    without_result = enrichment.run_null(
        "BIAS:WITHOUT_F68V2", without_f68v2,
        tuple(panel for panel in enrichment.PANELS if panel != "f68v2"), by_panel, hapax,
    )
    star_result = enrichment.run_null("BIAS:STAR", families["STAR"], enrichment.PANELS, by_panel, hapax)

    # Cross-check the independently generated panel table before sensitivity
    # calculations, rather than silently trusting either implementation.
    with BY_PANEL_HAPAX.open(encoding="utf-8", newline="") as handle:
        for row in csv.DictReader(handle, delimiter="\t"):
            reported = float(row["panel_background_hapax_fraction"])
            if abs(reported - panel_background[row["panel"]]) > 5e-10:
                raise RuntimeError(f"panel background mismatch for {row['panel']}")
    unmatched = [row for row in coordinates if not row["matched"]]
    add_by_panel = Counter()
    add_by_family = Counter()
    for row in unmatched:
        add_by_panel[row["panel"]] += row["potential_tokens"]
        add_by_family[row["family"]] += row["potential_tokens"]
    potential = sum(add_by_panel.values())
    observed_hapax = len(labels & hapax)
    observed_total = len(labels)

    def expanded_null(additions: Counter) -> float:
        sizes = Counter({panel: len(labels & set(by_panel[panel])) for panel in enrichment.PANELS})
        sizes.update(additions)
        expected = sum(sizes[panel] * panel_background[panel] for panel in enrichment.PANELS)
        return expected / sum(sizes.values())

    background_imputed = sum(
        add_by_panel[panel] * panel_background[panel] for panel in enrichment.PANELS
    )
    expanded_expected = expanded_null(add_by_panel)
    sensitivity_rows = []

    def sensitivity_row(name, added_hapax, note):
        total = observed_total + potential
        fraction = (observed_hapax + added_hapax) / total
        difference = fraction - expanded_expected
        ratio = fraction / expanded_expected
        sensitivity_rows.append([
            "UNMATCHED_BOUND", name, potential, fmt(added_hapax), total, fmt(fraction),
            fmt(expanded_expected), fmt(difference), fmt(ratio),
            "POSITIVE" if difference > 0 else "NON_POSITIVE", note,
        ])

    sensitivity_row("PESSIMISTIC_ALL_NON_HAPAX", 0.0,
                    "all potential unmatched token occurrences assigned non-hapax")
    sensitivity_row("OPTIMISTIC_ALL_HAPAX", float(potential),
                    "all potential unmatched token occurrences assigned hapax")
    sensitivity_row("PANEL_FAMILY_BACKGROUND_IMPUTATION", background_imputed,
                    "within each panel/family cell, use that panel's full-corpus hapax rate; no label outcome used")

    for name, result, note in (
        ("HIGH_COVERAGE_SERIES", high_result,
         "panel.group series with >=80% physical-coordinate coverage: " + ",".join(sorted(high_coverage_series))),
        ("WITHOUT_F68V2", without_result, "confirmed labels excluding f68v2"),
        ("STAR_SUBSET", star_result, "confirmed STAR label token occurrences"),
    ):
        sensitivity_rows.append([
            "CONFIRMED_SUBSET", name, 0, result["observed_count"], result["total"],
            fmt(result["observed"]), fmt(result["null_mean"]), fmt(result["difference"]),
            fmt(result["ratio"]), "POSITIVE" if result["difference"] > 0 else "NON_POSITIVE",
            f"p_upper={fmt(result['p'])}; {note}",
        ])
    write_tsv(
        SENSITIVITY,
        ["analysis_type", "scenario", "potential_added_token_occurrences", "hapax_count_or_expectation",
         "total_occurrences", "hapax_fraction", "conditioned_background_fraction", "difference",
         "enrichment_ratio", "direction", "notes"],
        sensitivity_rows,
    )

    feature_effect = {row[1]: float(row[5]) for row in audit_rows if row[0] == "PHYSICAL_COORDINATE"}
    strong_composition = (
        feature_effect["PANEL"] >= 0.3
        or feature_effect["LABEL_FAMILY"] >= 0.3
        or abs(feature_effect["LABEL_LENGTH_GLYPHS"]) >= 0.5
    )
    realistic_positive = sensitivity_rows[2][9] == "POSITIVE"
    subsets_positive = all(result["difference"] > 0 for result in (high_result, without_result, star_result))
    decision = (
        "UNLIKELY_TO_EXPLAIN_EFFECT"
        if not strong_composition and realistic_positive and subsets_positive
        else "COULD_EXPLAIN_EFFECT"
    )

    worst = sensitivity_rows[0]
    report = f"""# Matching-selection bias audit

## Decision

`MATCHING_SELECTION_BIAS={decision}`.

The earlier enrichment is robust within the mapped data, but mapping is not composition-neutral. At the physical-coordinate level there are 89 matched and 54 unmatched labels. Panel and family composition, coordinate-bridge availability, and exact lexical anchors differ between the two groups; the entire `f68v2.X/Y` sector series is unmatched. This violates the conditions required for `UNLIKELY_TO_EXPLAIN_EFFECT`.

The audit does **not** show that selection bias caused the enrichment. The high-coverage, no-f68v2, and star-only confirmed subsets all retain positive enrichment, and panel/family-background imputation retains the direction. However, the status of the unmapped labels is unobserved, and the formal pessimistic bound reverses the effect. The defensible conclusion is therefore that matching incompleteness *could* explain some or all of the observed enrichment.

## Units and pre-hapax features

- Primary selection unit: 143 physical `panel.group.number` coordinates (89 matched, 54 unmatched).
- Secondary descriptive unit: 191 Stolfi source records (130 matched, 61 unmatched). Record variants at one coordinate are not treated as independent.
- Family definitions are frozen from Stolfi fields: `STAR`, `PLANET_MOON`, `CIRCLE_SECTOR`, `RADIAL_TEXT`, `SECTOR_LABEL`, `OUTER_TITLE`, and `OTHER_DIAGRAM_LABEL`.
- Glyph and token lengths come only from the Stolfi EVA strings. `*` is recorded as a wildcard and excluded from known-glyph length.
- Exact-anchor availability is `raw_lexical_candidate_count > 0`; bridge/type fields are taken from the completed matching audit before any hapax status is joined.
- Actual mapped locus type is reported descriptively only: `NONE` for an unmatched coordinate is a consequence of matching and cannot serve as an independent predictor.

Association strength uses Cramer's V for categorical features and pooled standardized mean difference for numeric features. Coordinate-level p-values are two-sided, fixed-matched-count permutation tests with {PERMUTATIONS:,} draws. These diagnose selection structure; they are not causal tests.

## Composition findings

| feature | effect measure | effect | permutation p | band |
|---|---|---:|---:|---|
"""
    for row in audit_rows:
        if row[0] == "PHYSICAL_COORDINATE" and row[1] != "MAPPED_LOCUS_TYPE_USED":
            report += f"| {row[1]} | {row[4]} | {row[5]} | {row[6]} | {row[7]} |\n"
    report += f"""

The detailed panel/family table shows the substantive imbalance. `f68v2` has 6/28 mapped coordinates; `SECTOR_LABEL` is 0/15 and `OUTER_TITLE` is 0/1. In contrast, planet/moon is 7/7, circle-sector is 10/12, and star labels are 53/67 overall. Thus unmatched coordinates are concentrated in specific panels and families, not missing uniformly.

Matched coordinates also have shorter Stolfi readings by known-glyph length and more transcriber variants on average. Wildcard presence is weakly associated with matching, while exact lexical anchoring is nearly deterministic because it is a core matching requirement. These are properties of selection into the confirmed set, not evidence about hapax status by themselves.

## Sensitivity bounds

Unmatched coordinates are not inserted into the corpus. For bounds only, each of the 54 coordinates contributes the maximum Stolfi token count seen among its transcriber variants, for {potential} potential token occurrences. This is an accounting device, not a reconstructed label corpus.

| scenario | added potential occurrences | resulting/expected hapax fraction | conditioned background | difference | direction |
|---|---:|---:|---:|---:|---|
| pessimistic: all non-hapax | {potential} | {sensitivity_rows[0][5]} | {sensitivity_rows[0][6]} | {sensitivity_rows[0][7]} | {sensitivity_rows[0][9]} |
| optimistic: all hapax | {potential} | {sensitivity_rows[1][5]} | {sensitivity_rows[1][6]} | {sensitivity_rows[1][7]} | {sensitivity_rows[1][9]} |
| panel/family background imputation | {potential} | {sensitivity_rows[2][5]} | {sensitivity_rows[2][6]} | {sensitivity_rows[2][7]} | {sensitivity_rows[2][9]} |

The conditioned imputation uses only each panel's observed 901-token background hapax rate, applied within the actual unmatched panel/family cells. It does not use the desired label result. It preserves the positive direction but necessarily shrinks it. The pessimistic bound is not a realistic estimate; it establishes that the missing data are numerous enough, in principle, to reverse the result.

## Confirmed-label subset checks

| subset | labels | observed | null mean | ratio | difference | p (upper) |
|---|---:|---:|---:|---:|---:|---:|
| high-coverage series | {high_result['total']} | {fmt(high_result['observed'])} | {fmt(high_result['null_mean'])} | {fmt(high_result['ratio'])} | {fmt(high_result['difference'])} | {fmt(high_result['p'])} |
| without f68v2 | {without_result['total']} | {fmt(without_result['observed'])} | {fmt(without_result['null_mean'])} | {fmt(without_result['ratio'])} | {fmt(without_result['difference'])} | {fmt(without_result['p'])} |
| star family | {star_result['total']} | {fmt(star_result['observed'])} | {fmt(star_result['null_mean'])} | {fmt(star_result['ratio'])} | {fmt(star_result['difference'])} | {fmt(star_result['p'])} |

The high-coverage definition is frozen mechanically at `panel.group >= 80%` physical-coordinate coverage; qualifying series are `{', '.join(sorted(high_coverage_series))}`. These checks show that f68v2 alone is not responsible for the positive result, but they cannot recover outcomes for excluded or unmapped labels.

## Scope

This audit asks only whether mapping incompleteness can explain the previously detected enrichment. It neither invalidates the positive-label test nor upgrades unmatched records to labels or non-labels. No image interpretation, semantic matching, or desired-result imputation is used.

## Final status

```text
MATCHING_SELECTION_BIAS={decision}
MATCHED_RECORDS=130
UNMATCHED_RECORDS=61
HIGH_COVERAGE_SUBSET_EFFECT=ratio={fmt(high_result['ratio'])};difference={fmt(high_result['difference'])};p={fmt(high_result['p'])}
WITHOUT_F68V2_EFFECT=ratio={fmt(without_result['ratio'])};difference={fmt(without_result['difference'])};p={fmt(without_result['p'])}
STAR_SUBSET_EFFECT=ratio={fmt(star_result['ratio'])};difference={fmt(star_result['difference'])};p={fmt(star_result['p'])}
WORST_CASE_EFFECT=ratio={worst[8]};difference={worst[7]};direction={worst[9]}
```
"""
    REPORT.write_text(report, encoding="utf-8")

    result_files = [AUDIT_TSV, BY_PANEL_FAMILY, SENSITIVITY, REPORT]
    input_files = [MATCHES, LABEL_AUDIT, ENRICHMENT, BY_PANEL_HAPAX, enrichment.METADATA]
    implementation_files = [Path(__file__).resolve(), ENRICHMENT_IMPL]
    manifest = {
        "schema_version": 1,
        "task": "matching-selection-bias",
        "generated_date": "2026-09-01",
        "panels": list(enrichment.PANELS),
        "records": len(records),
        "matched_records": 130,
        "unmatched_records": 61,
        "physical_coordinates": len(coordinates),
        "matched_physical_coordinates": 89,
        "unmatched_physical_coordinates": 54,
        "permutations_per_test": PERMUTATIONS,
        "base_seed": SEED,
        "decision": decision,
        "python": platform.python_version(),
        "inputs": {str(path.relative_to(ROOT)): sha256(path) for path in input_files},
        "implementation": {str(path.relative_to(ROOT)): sha256(path) for path in implementation_files},
        "outputs": {str(path.relative_to(ROOT)): sha256(path) for path in result_files},
    }
    MANIFEST.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    checksum_paths = [*input_files, *implementation_files, *result_files, MANIFEST]
    CHECKSUMS.write_text(
        "".join(f"{sha256(path)}  {path.relative_to(ROOT)}\n" for path in checksum_paths),
        encoding="utf-8",
    )

    print(f"MATCHING_SELECTION_BIAS={decision}")
    print("MATCHED_RECORDS=130")
    print("UNMATCHED_RECORDS=61")
    print(f"HIGH_COVERAGE_SUBSET_EFFECT=ratio={fmt(high_result['ratio'])};difference={fmt(high_result['difference'])};p={fmt(high_result['p'])}")
    print(f"WITHOUT_F68V2_EFFECT=ratio={fmt(without_result['ratio'])};difference={fmt(without_result['difference'])};p={fmt(without_result['p'])}")
    print(f"STAR_SUBSET_EFFECT=ratio={fmt(star_result['ratio'])};difference={fmt(star_result['difference'])};p={fmt(star_result['p'])}")
    print(f"WORST_CASE_EFFECT=ratio={worst[8]};difference={worst[7]};direction={worst[9]}")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"error: {error}", file=sys.stderr)
        raise
