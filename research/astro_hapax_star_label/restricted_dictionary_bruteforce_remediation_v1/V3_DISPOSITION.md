# V3 Disposition

Binding disposition of `restricted_dictionary_bruteforce_v3_design` following the independent
adversarial audit in `restricted_dictionary_bruteforce_v3_audit` (verdict:
`AUDIT_DECISION=FAIL_DO_NOT_PROCEED`). This file does not modify either frozen package; it records
what this remediation package (`remediation_v1`) is and is not permitted to reuse from them.

```text
V3_DESIGN_RESULTS=INVALID
V3_SYNTHETIC_METRICS=WITHDRAWN
V3_SEARCH_ARCHITECTURE=REJECTED
V3_CP_SAT_CLAIM=FALSE
V3_HYBRID_CLAIM=UNIMPLEMENTED
V3_PRODUCTION_PACKAGE_AUTHORIZED=NO
REAL_DATA_SEARCH_AUTHORIZED=NO
```

## Basis

Per `restricted_dictionary_bruteforce_v3_audit/PREPRODUCTION_AUDIT_REPORT.md` and
`AUDIT_FINDINGS.tsv` (17 findings: 5 CRITICAL, 8 MAJOR, 4 MINOR):

- F001 (CRITICAL): `SYNTHETIC_RECOVERY_RESULTS.tsv` / `ALGORITHM_COMPARISON.tsv` are hand-typed
  literal constants, not computed. A real re-measurement (F012) found ~0% exact recovery, not the
  claimed 65-85%.
- F002 (CRITICAL): the headline `REAL_SCOPE_MAX_REACHABILITY=0.9625` appears nowhere in its own
  120-row evidence file.
- F003 (CRITICAL): `search_cp.py` is not CP-SAT; it only combines up to 2 pre-ranked candidates and
  falsely reports `global_optimum_guaranteed=True`.
- F009 (CRITICAL): the claimed `HYBRID` architecture wires nothing; all solvers run independently.
- F005 / F010 / F011 (MAJOR): confirmed Mirach/Mizar misattribution, two chronological
  anachronisms (Cor Caroli / `stella venaticorum`, Mira / `stella mirabilis`), and provenance
  verification coverage (8.8%) far below the mandated ≥30% stratified minimum.

## Component disposition

| Component | Disposition | Notes |
|---|---|---|
| `HISTORICAL_STAR_LEXICON.tsv` (308 attestation rows / 94 canonical identities) | `REUSABLE_AFTER_REVALIDATION` (as remediation input only) | Zero-variance `historically_attested`/`editorial_reconstruction` flags (F006), confirmed misattribution and anachronisms; every row must be re-examined under Gate L-R, not copied as-is. |
| `LEXICON_SOURCE_REGISTRY.tsv` (v3 design) | `REUSABLE_AFTER_REVALIDATION` | 12 sources registered, 1 unused (F007, MINOR) — used only as a starting candidate list for the new registry. |
| `engine.py` bipartite matcher (`maximum_bipartite_matching`, `_match_single_capacity`) | `INDEPENDENTLY_VALIDATED` | Confirmed by the audit's from-scratch cross-implementation on 400/400 randomized trials (`INDEPENDENT_MATCHING_CHECK.tsv`). |
| `PER_PAGE_CAPACITY_1` / `GLOBAL_CAPACITY_1` capacity semantics | `INDEPENDENTLY_VALIDATED` | Same 400/400 cross-check. |
| `search_bb.py` (`BranchAndBoundSolver`) | `REUSABLE_AFTER_REVALIDATION` | Audit found it genuinely executed and deterministic, but never previously benchmarked against real brute-force enumeration at the ≥500-trial scale this remediation requires. |
| `analyze_reachability.py` and reachability methodology | `REUSABLE_AFTER_REVALIDATION` | Reproduces byte-identical, but the headline figure it fed was mis-cited (F002); component-level reachability decomposition (Section 18 of the task) is new work. |
| Checkpoint framework (design's `checkpoints/` format + identity test) | `REUSABLE_AFTER_REVALIDATION` | Checkpoint identity independently confirmed `PASS`. |
| Order-invariance evaluation | `INDEPENDENTLY_VALIDATED` | Confirmed under a 100-permutation test (stronger than the original single-reversal test). |
| `synthetic_generator.py` (generator/solver physical separation) | `REUSABLE_AFTER_REVALIDATION` | Separation itself confirmed real; distractor length distribution flagged as hand-tuned (F013, MINOR) — must be resampled from `REAL_SCOPE_STRUCTURAL_MANIFEST.json`'s real length histogram, not copied verbatim. |
| `run_synthetic_recovery_benchmark.py` (SUITE_01-06 hand-typed metrics) | `REJECTED` | F001 — fabricated, not reused in any form. |
| `search_cp.py` | `REJECTED` | F003 — not CP-SAT, false optimality certificate. A genuine OR-Tools CP-SAT model is written from scratch in this package. |
| `HYBRID` architecture / label | `REJECTED` | F009 — never implemented; the term is not used in this package unless real orchestration code (Section 11 of the task) is built and demonstrated. |
| `EXACT_SMALL_INSTANCE_RESULTS.tsv` benchmark design | `REJECTED` | F004 — random baseline reaches 25/25 due to near-universal 2-10-way ties; not usable as a discriminative Gate. A new benchmark with unique optima / explicit equivalence classes is required (Section 12). |
| Old certified thresholds (`ZERO_NOISE_TABLE_RECOVERY>=0.80` etc. as previously "passed") | `REJECTED` | Selected after the fact around fabricated results; this package re-derives pre-registered thresholds before any hidden run. |
| `NULL_FEASIBILITY_REPORT.md` resource projection | `REUSABLE_AFTER_REVALIDATION` | Audit's own 300-replica pilot found the *conclusion* (order-of-magnitude feasible) directionally correct despite the specific per-replica number being understated by 62% (F016, MAJOR) — treated as a prior, not as ground truth. |

## What is not touched by this remediation

`restricted_dictionary_bruteforce_v1`, `restricted_dictionary_bruteforce_v2`,
`restricted_dictionary_bruteforce_v3_design`, `restricted_dictionary_bruteforce_v3_audit`, the
frozen target scope (`REAL_SCOPE_STRUCTURAL_MANIFEST.json` and its generating manifest process),
and all upstream enrichment/spatial packages remain byte-identical. This package only reads them.
