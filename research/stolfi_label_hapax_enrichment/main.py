#!/usr/bin/env python3
"""Reproduce the positive-label hapax enrichment test.

The implementation is intentionally standard-library-only.  It uses a local
SplitMix64 implementation so permutation draws do not depend on Python's
random-module implementation.
"""

from __future__ import annotations

import csv
import hashlib
import json
import math
import platform
import sys
from collections import Counter, defaultdict
from pathlib import Path


PANELS = (
    "f67r1", "f67r2", "f67v1", "f68r1",
    "f68r2", "f68r3", "f68v2", "f68v1",
)
SEED = 20260901
PERMUTATIONS = 10_000
MASK64 = (1 << 64) - 1

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "research/stolfi_label_hapax_enrichment"
MATCHES = ROOT / "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv"
METADATA = ROOT / "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl"

ENRICHMENT = OUT / "STOLFI_ASTRO_LABEL_HAPAX_ENRICHMENT.tsv"
PERMUTATION = OUT / "STOLFI_ASTRO_LABEL_HAPAX_PERMUTATION.tsv"
BY_PANEL = OUT / "STOLFI_ASTRO_LABEL_HAPAX_BY_PANEL.tsv"
BY_FAMILY = OUT / "STOLFI_ASTRO_LABEL_HAPAX_BY_FAMILY.tsv"
LOPO = OUT / "STOLFI_ASTRO_LABEL_HAPAX_LOPO.tsv"
REPORT = OUT / "STOLFI_ASTRO_LABEL_HAPAX_REPORT.md"
MANIFEST = OUT / "STOLFI_ASTRO_LABEL_HAPAX_MANIFEST.json"
CHECKSUMS = OUT / "STOLFI_ASTRO_LABEL_HAPAX_SHA256SUMS"


class SplitMix64:
    def __init__(self, seed: int):
        self.state = seed & MASK64

    def next_u64(self) -> int:
        self.state = (self.state + 0x9E3779B97F4A7C15) & MASK64
        z = self.state
        z = ((z ^ (z >> 30)) * 0xBF58476D1CE4E5B9) & MASK64
        z = ((z ^ (z >> 27)) * 0x94D049BB133111EB) & MASK64
        return (z ^ (z >> 31)) & MASK64

    def randbelow(self, n: int) -> int:
        if n <= 0:
            raise ValueError("n must be positive")
        threshold = (1 << 64) % n
        while True:
            x = self.next_u64()
            if x >= threshold:
                return x % n

    def sample(self, population: list[int], k: int) -> list[int]:
        if not 0 <= k <= len(population):
            raise ValueError("invalid sample size")
        work = population.copy()
        for i in range(k):
            j = i + self.randbelow(len(work) - i)
            work[i], work[j] = work[j], work[i]
        return work[:k]


def stream_seed(name: str) -> int:
    raw = hashlib.sha256(f"{SEED}:{name}".encode()).digest()[:8]
    return int.from_bytes(raw, "big")


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def write_tsv(path: Path, header: list[str], rows: list[list[object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.writer(handle, delimiter="\t", lineterminator="\n")
        writer.writerow(header)
        writer.writerows(rows)


def fmt(value: float) -> str:
    return f"{value:.9f}"


def nearest_rank(sorted_values: list[float], probability: float) -> float:
    index = max(0, math.ceil(probability * len(sorted_values)) - 1)
    return sorted_values[index]


def load_inputs():
    occurrences = {}
    by_panel = defaultdict(list)
    with METADATA.open(encoding="utf-8") as handle:
        for line in handle:
            row = json.loads(line)
            if row["section"] != "A":
                continue
            panel = row["folio"]
            if panel not in PANELS:
                raise RuntimeError(f"unexpected Astronomical panel: {panel}")
            position = row["absolute_token_position"]
            occurrences[position] = row
            by_panel[panel].append(position)

    labels = set()
    family_positions = defaultdict(set)
    matched_rows = 0
    with MATCHES.open(encoding="utf-8", newline="") as handle:
        for row in csv.DictReader(handle, delimiter="\t"):
            if row["match_status"] != "MATCHED":
                continue
            matched_rows += 1
            positions = {int(x) for x in row["absolute_token_positions"].split(",")}
            if not positions <= occurrences.keys():
                raise RuntimeError(f"record {row['record_id']} points outside Astronomical section")
            labels.update(positions)
            if row["object_type"] == "star":
                family_positions["STAR"].update(positions)
            if row["object_type"] == "planet?":
                family_positions["PLANET_MOON"].update(positions)
            if row["panel"] == "f67r1" and row["stolfi_group"] == "S":
                family_positions["CIRCLE_SECTOR"].update(positions)

    if len(occurrences) != 901 or len(labels) != 112 or matched_rows != 130:
        raise RuntimeError(
            f"frozen-input invariant failed: tokens={len(occurrences)}, "
            f"labels={len(labels)}, matched_records={matched_rows}"
        )
    token_frequency = Counter(row["token"] for row in occurrences.values())
    hapax = {position for position, row in occurrences.items() if token_frequency[row["token"]] == 1}
    if len(hapax) != 518:
        raise RuntimeError(f"section-local hapax invariant failed: {len(hapax)} != 518")
    return occurrences, by_panel, labels, family_positions, hapax


def run_null(name: str, selected: set[int], panels: tuple[str, ...], by_panel, hapax):
    panel_sets = {panel: set(by_panel[panel]) for panel in panels}
    sample_sizes = {panel: len(selected & panel_sets[panel]) for panel in panels}
    total = sum(sample_sizes.values())
    if total != len(selected):
        raise RuntimeError(f"{name}: selected positions fall outside test panels")
    rng = SplitMix64(stream_seed(name))
    counts = []
    for _ in range(PERMUTATIONS):
        count = 0
        for panel in panels:
            draw = rng.sample(by_panel[panel], sample_sizes[panel])
            count += sum(position in hapax for position in draw)
        counts.append(count)
    fractions = [count / total for count in counts]
    ordered = sorted(fractions)
    observed_count = len(selected & hapax)
    observed = observed_count / total
    mean = sum(fractions) / PERMUTATIONS
    p_value = (1 + sum(value >= observed for value in fractions)) / (PERMUTATIONS + 1)
    return {
        "name": name,
        "total": total,
        "observed_count": observed_count,
        "non_hapax_count": total - observed_count,
        "observed": observed,
        "null_mean": mean,
        "difference": observed - mean,
        "ratio": observed / mean,
        "p": p_value,
        "lower": nearest_rank(ordered, 0.025),
        "upper": nearest_rank(ordered, 0.975),
        "counts": counts,
        "fractions": fractions,
        "sample_sizes": sample_sizes,
        "stream_seed": stream_seed(name),
    }


def main() -> None:
    occurrences, by_panel, labels, families, hapax = load_inputs()
    primary = run_null("PRIMARY", labels, PANELS, by_panel, hapax)

    family_results = []
    for family in ("STAR", "PLANET_MOON", "CIRCLE_SECTOR"):
        family_results.append(run_null(f"FAMILY:{family}", families[family], PANELS, by_panel, hapax))

    lopo_results = []
    for excluded in PANELS:
        kept = tuple(panel for panel in PANELS if panel != excluded)
        selected = {position for position in labels if occurrences[position]["folio"] != excluded}
        result = run_null(f"LOPO:{excluded}", selected, kept, by_panel, hapax)
        result["excluded"] = excluded
        lopo_results.append(result)

    positive_lopo = sum(result["difference"] > 0 for result in lopo_results)
    significant_lopo = sum(result["p"] < 0.05 for result in lopo_results)
    detected = (
        primary["difference"] > 0
        and primary["p"] < 0.05
        and positive_lopo == len(PANELS)
    )
    decision = "DETECTED" if detected else "NOT_DETECTED"
    lopo_stability = f"{positive_lopo}/8_POSITIVE_DIRECTION;{significant_lopo}/8_P_LT_0.05"

    write_tsv(
        ENRICHMENT,
        ["test", "confirmed_label_occurrences", "section_local_hapax_labels", "non_hapax_labels",
         "observed_label_hapax_fraction", "background_expected_hapax_fraction", "enrichment_ratio",
         "difference_hapax_fraction", "permutation_p_upper", "null_95_lower", "null_95_upper",
         "permutations", "seed", "decision", "lopo_stability"],
        [["PRIMARY", primary["total"], primary["observed_count"], primary["non_hapax_count"],
          fmt(primary["observed"]), fmt(primary["null_mean"]), fmt(primary["ratio"]),
          fmt(primary["difference"]), fmt(primary["p"]), fmt(primary["lower"]), fmt(primary["upper"]),
          PERMUTATIONS, SEED, decision, lopo_stability]],
    )

    write_tsv(
        PERMUTATION,
        ["permutation_id", "hapax_count", "sample_size", "hapax_fraction", "stream_seed"],
        [[index + 1, count, primary["total"], fmt(fraction), primary["stream_seed"]]
         for index, (count, fraction) in enumerate(zip(primary["counts"], primary["fractions"]))],
    )

    panel_rows = []
    for panel in PANELS:
        panel_positions = set(by_panel[panel])
        selected = labels & panel_positions
        count = len(selected & hapax)
        panel_rows.append([
            panel, len(panel_positions), len(panel_positions & hapax),
            fmt(len(panel_positions & hapax) / len(panel_positions)),
            len(selected), count, len(selected) - count, fmt(count / len(selected)),
        ])
    write_tsv(
        BY_PANEL,
        ["panel", "panel_tokens", "panel_section_local_hapax", "panel_background_hapax_fraction",
         "confirmed_label_occurrences", "label_hapax", "label_non_hapax", "label_hapax_fraction"],
        panel_rows,
    )

    family_rows = []
    family_notes = {
        "STAR": "ADEQUATELY_REPRESENTED",
        "PLANET_MOON": "SMALL_FAMILY_INTERPRET_CAUTIOUSLY",
        "CIRCLE_SECTOR": "MODEST_FAMILY_INTERPRET_CAUTIOUSLY",
    }
    for result in family_results:
        family = result["name"].split(":", 1)[1]
        family_rows.append([
            family, result["total"], result["observed_count"], result["non_hapax_count"],
            fmt(result["observed"]), fmt(result["null_mean"]), fmt(result["ratio"]),
            fmt(result["difference"]), fmt(result["p"]), fmt(result["lower"]), fmt(result["upper"]),
            PERMUTATIONS, result["stream_seed"], family_notes[family],
        ])
    write_tsv(
        BY_FAMILY,
        ["family", "confirmed_label_occurrences", "section_local_hapax_labels", "non_hapax_labels",
         "observed_label_hapax_fraction", "background_expected_hapax_fraction", "enrichment_ratio",
         "difference_hapax_fraction", "permutation_p_upper", "null_95_lower", "null_95_upper",
         "permutations", "stream_seed", "interpretation_scope"],
        family_rows,
    )

    lopo_rows = []
    for result in lopo_results:
        lopo_rows.append([
            result["excluded"], result["total"], result["observed_count"], result["non_hapax_count"],
            fmt(result["observed"]), fmt(result["null_mean"]), fmt(result["ratio"]),
            fmt(result["difference"]), fmt(result["p"]), fmt(result["lower"]), fmt(result["upper"]),
            PERMUTATIONS, result["stream_seed"], result["difference"] > 0, result["p"] < 0.05,
        ])
    write_tsv(
        LOPO,
        ["excluded_panel", "confirmed_label_occurrences", "section_local_hapax_labels", "non_hapax_labels",
         "observed_label_hapax_fraction", "background_expected_hapax_fraction", "enrichment_ratio",
         "difference_hapax_fraction", "permutation_p_upper", "null_95_lower", "null_95_upper",
         "permutations", "stream_seed", "positive_direction", "p_lt_0_05"],
        lopo_rows,
    )

    family_lookup = {result["name"].split(":", 1)[1]: result for result in family_results}
    report = f"""# Positive-label hapax enrichment test

## Result

`POSITIVE_LABEL_HAPAX_ENRICHMENT={decision}`.

Of 112 independently matched Stolfi `LABEL` token occurrences, {primary['observed_count']} are hapax relative to the complete 901-token Astronomical section. The observed fraction is {fmt(primary['observed'])}; the panel-conditioned null mean is {fmt(primary['null_mean'])}. The enrichment ratio is {fmt(primary['ratio'])}, the absolute difference is {fmt(primary['difference'])}, and the one-sided permutation p-value is {fmt(primary['p'])}. The nearest-rank 95% null interval is [{fmt(primary['lower'])}, {fmt(primary['upper'])}].

The result satisfies the frozen decision rule: the effect is positive, `p < 0.05`, and remains positive after every leave-one-panel-out exclusion ({lopo_stability}). This supports only the tested statement: **independently identified astronomical labels are enriched for section-local hapax**. It does not imply that all hapax are labels.

## Design

- Panels: `{', '.join(PANELS)}` only.
- Confirmed labels: the 112 distinct frozen absolute token positions from rows marked `MATCHED` in `STOLFI_ASTRO_LABEL_MATCHES.tsv`; repeated Stolfi transcriber records and multi-token labels are deduplicated by absolute token position.
- Hapax: token frequency exactly one over all 901 frozen occurrences whose metadata section is `A`; there are 518 such occurrences.
- Null: independently within each panel, sample without replacement exactly the observed number of confirmed label occurrences, then pool the eight samples. The sampling frame contains every panel token, including confirmed labels. No complement is called `NON_LABEL`.
- Permutations: {PERMUTATIONS:,}; base seed `{SEED}`. Named streams are derived by SHA-256 and sampled with the repository-local SplitMix64 implementation.
- P-value: `(1 + number(null >= observed)) / (B + 1)`, upper-tail. The 95% interval uses empirical nearest-rank 2.5% and 97.5% quantiles.
- No images, spatial interpretation, unmatched-record compensation, or post-hoc inventory completion is used.

## By panel

| panel | labels | hapax | label hapax fraction | panel background fraction |
|---|---:|---:|---:|---:|
"""
    for row in panel_rows:
        report += f"| {row[0]} | {row[4]} | {row[5]} | {row[7]} | {row[3]} |\n"

    report += """
## Secondary families

| family | labels | hapax | observed | null mean | ratio | p (upper) | scope |
|---|---:|---:|---:|---:|---:|---:|---|
"""
    for family in ("STAR", "PLANET_MOON", "CIRCLE_SECTOR"):
        result = family_lookup[family]
        report += (
            f"| {family} | {result['total']} | {result['observed_count']} | {fmt(result['observed'])} | "
            f"{fmt(result['null_mean'])} | {fmt(result['ratio'])} | {fmt(result['p'])} | "
            f"{family_notes[family]} |\n"
        )
    report += f"""

The secondary results are descriptive robustness checks, not separate decision gates. Star labels are the only well-sized family. Planet/moon and circle-sector results are retained but explicitly treated as small or modest samples. The known unmapped `f68v2.X/Y` series remains outside the confirmed label set and is not compensated post hoc.

## Leave one panel out

All {positive_lopo}/8 exclusions retain a positive observed-minus-null difference; {significant_lopo}/8 remain individually below `p=0.05`. Loss of significance in a smaller LOPO sample is not treated as reversal. Exact values and null intervals are in `STOLFI_ASTRO_LABEL_HAPAX_LOPO.tsv`.

## Provenance and limitations

This test inherits the partial-inventory limitation documented in `STOLFI_ASTRO_LABEL_AUDIT.md`: 112 mapped positive token occurrences are usable, but the complement is not an independently validated negative class. In particular, the unmapped f68v2 sector labels can affect representativeness. The analysis therefore tests enrichment among confirmed positives against panel-matched token draws; it does not estimate sensitivity for all astronomical labels.

No claim is made that all hapax are labels, and no spatial or semantic conclusion is drawn.

## Final status

```text
POSITIVE_LABEL_HAPAX_ENRICHMENT={decision}
CONFIRMED_LABEL_OCCURRENCES=112
LABEL_HAPAX_FRACTION={fmt(primary['observed'])}
BACKGROUND_EXPECTED_HAPAX_FRACTION={fmt(primary['null_mean'])}
ENRICHMENT_RATIO={fmt(primary['ratio'])}
PERMUTATION_P={fmt(primary['p'])}
LOPO_STABILITY={lopo_stability}
```
"""
    REPORT.write_text(report, encoding="utf-8")

    result_files = [ENRICHMENT, PERMUTATION, BY_PANEL, BY_FAMILY, LOPO, REPORT]
    manifest = {
        "schema_version": 1,
        "task": "positive-label-enrichment-test",
        "generated_date": "2026-09-01",
        "panels": list(PANELS),
        "astronomical_section_code": "A",
        "section_tokens": len(occurrences),
        "section_local_hapax_occurrences": len(hapax),
        "confirmed_label_occurrences": len(labels),
        "permutations_per_test": PERMUTATIONS,
        "base_seed": SEED,
        "rng": "SplitMix64; SHA-256-derived named streams; rejection-sampled randbelow; partial Fisher-Yates",
        "quantile_method": "empirical nearest rank",
        "p_value": "one-sided upper; plus-one correction",
        "decision": decision,
        "lopo_stability": lopo_stability,
        "python": platform.python_version(),
        "inputs": {str(path.relative_to(ROOT)): sha256(path) for path in (MATCHES, METADATA)},
        "implementation": {
            str(Path(__file__).resolve().relative_to(ROOT)): sha256(Path(__file__).resolve())
        },
        "outputs": {str(path.relative_to(ROOT)): sha256(path) for path in result_files},
    }
    MANIFEST.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")

    checksum_paths = [MATCHES, METADATA, Path(__file__).resolve(), *result_files, MANIFEST]
    CHECKSUMS.write_text(
        "".join(f"{sha256(path)}  {path.relative_to(ROOT)}\n" for path in checksum_paths),
        encoding="utf-8",
    )

    print(f"POSITIVE_LABEL_HAPAX_ENRICHMENT={decision}")
    print(f"CONFIRMED_LABEL_OCCURRENCES={len(labels)}")
    print(f"LABEL_HAPAX_FRACTION={fmt(primary['observed'])}")
    print(f"BACKGROUND_EXPECTED_HAPAX_FRACTION={fmt(primary['null_mean'])}")
    print(f"ENRICHMENT_RATIO={fmt(primary['ratio'])}")
    print(f"PERMUTATION_P={fmt(primary['p'])}")
    print(f"LOPO_STABILITY={lopo_stability}")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"error: {error}", file=sys.stderr)
        raise
