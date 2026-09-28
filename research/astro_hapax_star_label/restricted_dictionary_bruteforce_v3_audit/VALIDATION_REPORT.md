# Independent Validation Report (Audit Re-Assessment of Gates L and S)

This supersedes, for audit purposes, the design package's own `VALIDATION_REPORT.md` self-assessment.
It re-runs the same gate table using only independently reproduced or freshly measured evidence.

## Gate L Re-Assessment (Historical Lexicon)

| Criterion | Design's claim | Independent audit finding | Determination |
|---|---|---|:---:|
| `LEXICON_INDEPENDENT_OF_EVA` | YES | Confirmed at the code level (no TARGET_SCOPE/f68r/EVA_ALPHABET reference in the lexicon-building script) | **PASS** |
| `LEXICON_PROVENANCE_COMPLETE` | YES, 100% verified | Only 8.8% of rows independently spot-checked (below the ≥30% mandated minimum); of those, 1 confirmed misattribution + 2 confirmed chronological anachronisms found in a 9-row sub-sample, which should have triggered 100%-of-stratum escalation | **FAIL** |
| `CANONICAL_IDENTITIES>=57` | 94 | Confirmed: 94 | **PASS** |
| `ATTESTATIONS_DEDUPLICATED` | YES | Confirmed: 0/94 identities have internal duplicate forms | **PASS** |
| `SOURCE_COVERAGE_DOCUMENTED` | YES, 12 sources | 11/12 registered sources actually used (F007, minor) | **PASS (minor caveat)** |
| `CAPACITY_POLICY_JUSTIFIED` | YES | Confirmed via unit test re-run + independent bipartite-matcher cross-check (400/400 trials consistent) | **PASS** |
| `LEXICON_INDEPENDENCE` (cross-identity separability) | implied YES | 6/94 identities share a normalized form with a different identity; 1 pair looks like genuine historical ambiguity, 1 pair (Mizar/Mirach) looks like a misattribution | **FAIL** |
| **GATE L** | **PASS** | | **FAIL** |

## Gate S Re-Assessment (Search Architecture)

| Criterion | Design's claim | Independent audit finding | Determination |
|---|---|---|:---:|
| `MODEL_CLASSES_EVALUATED=120` | 120 | Confirmed: `analyze_reachability.py` reproduces 120 rows byte-identical | **PASS** |
| `REAL_SCOPE_MAX_REACHABILITY=0.9625` | 0.9625 | Does not match any value in the reproducing file; closest real values 0.9649/0.9583 | **FAIL (unsupported figure)** |
| `EXACT_SMALL_INSTANCE_OPTIMUM_RATE>=0.95` | 96-100% | Reproduces byte-identical, but a 10-sample random baseline also reaches 100% on the same instances (benchmark too easy to be discriminating) | **PASS (weak evidence)** |
| `ZERO_NOISE_TABLE_RECOVERY>=0.80` | 0.850 (claimed) | Never computed by the cited code (only 2 solve() calls in the whole file, both unrelated to this metric); fresh honest re-measurement at the identical operating point: **0.000** | **FAIL (fabricated; real value ≈0)** |
| `TEN_PERCENT_NOISE_TABLE_RECOVERY>=0.70` | 0.750 (claimed) | Same fabrication; fresh measurement: **0.000** | **FAIL (fabricated; real value ≈0)** |
| `TEN_PERCENT_HELDOUT_ASSIGNMENT_ACCURACY>=0.60` | 0.684 (claimed) | Same fabrication; fresh measurement of (non-heldout) assignment accuracy at the same config: **≈0.287** | **FAIL (fabricated; real value far below threshold)** |
| `TEN_PERCENT_MAPPING_PRECISION/RECALL>=0.75` | 0.812/0.812 (claimed) | Same fabrication; fresh measurement: **≈0.138/0.138** | **FAIL (fabricated; real value far below threshold)** |
| `CP_SOLVER_CLASSIFICATION=CP_SAT` | implied | Empirically 0% global-optimum rate at k>=3 fragments required, with a false "guaranteed" certificate; this is a bounded pairwise-candidate heuristic, not CP-SAT | **FAIL (mislabeled)** |
| `SELECTED_SEARCH_ARCHITECTURE=HYBRID` | HYBRID_SELECTED | No code implements the described heuristic+neighborhood-verification wiring anywhere in the package | **FAIL (unimplemented)** |
| `ORDER_INVARIANCE` | PASS | Re-verified far more thoroughly (100 random permutations vs the design's single reversal test): 0/100 mismatches | **PASS (strengthened)** |
| `CHECKPOINT_IDENTITY` | PASS | Re-verified: identical | **PASS** |
| `RESOURCE_FEASIBILITY` (10k nulls in ≤4h) | 62.5 min (single unlogged data point) | See `RESOURCE_BENCHMARK.tsv` for an independently measured, reproducible pilot with min/mean/p95/max statistics | **see RESOURCE_BENCHMARK.tsv** |
| **GATE S** | **PASS** | | **FAIL** |

## Conclusion

Neither Gate L nor Gate S survives independent re-assessment. Gate S fails for a reason more
serious than any individual metric miss: its two headline supporting evidence files
(`SYNTHETIC_RECOVERY_RESULTS.tsv`, and the `REAL_SCOPE_MAX_REACHABILITY` figure in
`V3_DESIGN_REPORT.md`) do not correspond to anything the package's own code actually computed, and
its named "HYBRID" architecture has no implementation. Gate L fails because provenance
verification, which the design package's own claim of "100% verified" implies was performed, was
not — and the first spot-check this audit performed already surfaced a confirmed misattribution and
two confirmed anachronisms.
