# Pre-Production Audit Report
## `restricted_dictionary_bruteforce_v3_design` — Independent Adversarial Audit

## 1. Mandate and Summary Verdict

This audit was tasked with attempting to **refute** the claimed `LEXICON_GATE=PASS` and
`SEARCH_ARCHITECTURE_GATE=PASS` of the v3 design package before it is frozen into a production
package. It succeeded in doing so on multiple independent grounds, several of which are severe
enough on their own to be disqualifying:

1. **The package's central quantitative evidence for Gate S was never computed.**
   `SYNTHETIC_RECOVERY_RESULTS.tsv` and `ALGORITHM_COMPARISON.tsv` — the files supporting
   `ZERO_NOISE_TABLE_RECOVERY=0.850`, `TEN_PERCENT_NOISE_TABLE_RECOVERY=0.750`,
   `TEN_PERCENT_HELDOUT_ASSIGNMENT_ACCURACY=0.684`, and `TEN_PERCENT_MAPPING_PRECISION/RECALL=0.812`
   — are hand-typed literal constants. The script that supposedly produces them
   (`run_synthetic_recovery_benchmark.py`) calls a real solver exactly twice in the entire file,
   both times for an unrelated order-invariance/checkpoint test. **(Finding F001, CRITICAL.)**
2. **A freshly and honestly executed re-measurement at the identical operating point (real
   solver, real 94-identity lexicon, new sealed seeds) found recovery rates near zero**, not 65-85%:
   exact table recovery 0.0% at every table size (4,6,8,10,12) and every noise level (0%,10%,25%)
   tested; mapping precision/recall ≈ 9-17% (claimed 81-87%); assignment accuracy ≈ 27-29% (claimed
   82-95%). **(Finding F012, CRITICAL — see `FRESH_HIDDEN_RESULTS.tsv`.)**
3. **The headline reachability figure is unsupported by its own cited evidence file.**
   `REAL_SCOPE_MAX_REACHABILITY=0.9625` appears nowhere in the 120-row `REACHABILITY_ANALYSIS.tsv`,
   which independently reproduces byte-identical from its generating script. **(Finding F002, CRITICAL.)**
4. **`search_cp.py` is not a CP-SAT solver and its completeness claim is false.** It only ever
   combines up to 2 pre-ranked candidate partial mappings; on adversarial instances requiring 3+
   independent fragments, it finds the true optimum in **0 of 5 trials at k=3, k=4, k=5, and k=6**,
   while still reporting `global_optimum_guaranteed=True`. **(Finding F003, CRITICAL.)**
5. **The claimed "HYBRID" search architecture does not exist in code.** No script wires the
   heuristic's output into a restricted-neighborhood exact-verification step; all three solvers
   run fully independently everywhere they are called. **(Finding F009, CRITICAL.)**
6. **Historical-lexicon provenance verification was incomplete, and the small sample that was
   checked already surfaced a confirmed misattribution and two confirmed chronological
   anachronisms** (a 1435-1445 manuscript "attesting" a constellation invented in 1687, and a star
   name framing that predates the star's documented recognition by ~150 years).
   **(Findings F005, F010, F011 — MAJOR.)**

None of these findings required searching the real 57 Voynich star labels; all were established
through code reading, independent re-execution inside an isolated sandbox, a from-scratch
cross-check implementation, freshly seeded synthetic data, and open secondary-scholarship
verification of historical claims.

## 2. What Held Up

Not everything in the design package failed re-audit:
- `analyze_reachability.py`, `run_small_instance_benchmark.py`, `engine.py`'s bipartite matcher,
  the design's own unit tests, and `search_bb.py`'s branch-and-bound are all genuinely executed,
  deterministic, and reproduce byte-identical on independent re-run.
- The synthetic-generator/solver physical separation (no truth leakage into search) is real and
  intact, confirmed by grep and by this audit's own fresh sealed-benchmark protocol.
- Order invariance held up even under a much more thorough test (100 random permutations vs. the
  original single reversal test) and checkpoint identity is genuine.
- Most of the historical lexicon's stratified spot-check sample (17/27, ~63%) looked etymologically
  plausible and consistent with independent scholarship.
- Capacity-policy semantics (`PER_PAGE_CAPACITY_1` vs `GLOBAL_CAPACITY_1`) are implemented and
  tested correctly, confirmed by an independent from-scratch bipartite-matching implementation
  agreeing on 400/400 randomized trials.
- The resource-feasibility *conclusion* (a full null pipeline is executable in roughly an hour on
  16 cores) survives independent re-measurement even though its specific supporting number does
  not: a real 300-replica timed pilot measured a 62% higher per-replica cost (1.05s vs. the claimed
  0.65s) but still extrapolates to the same order of magnitude (65.9-76.8 min vs. claimed 62.5 min
  for 60,000 replicas on 16 cores). This is the one headline claim in the package that turned out to
  be directionally correct despite being under-substantiated (Finding F016, MAJOR, not CRITICAL).

The failures found are not "everything is broken" — they are concentrated precisely in the
package's headline validation claims (the two gates), which is the worst possible place for them
to occur just before a production-freeze decision.

## 3. Findings Summary

See `AUDIT_FINDINGS.tsv` for the full, structured list (17 findings: 5 CRITICAL, 8 MAJOR, 4 MINOR)
with evidence, reproducing commands, affected metrics, required fixes, and whether a new hidden
benchmark is required for each.

## 4. Recommended Path Forward

1. Do not freeze this design package into a production package as currently written.
2. Fix F001 (delete the fabricated benchmark suites; genuinely execute the solvers), F002 (correct
   or re-derive the headline reachability figure), F003 (rename/fix or replace `search_cp.py`), and
   F009 (implement or remove the HYBRID architecture claim).
3. Complete the historical-lexicon provenance escalation triggered by F011 (100% verification of
   the `MEDIUM`-confidence stratum, plus a genuine ≥30% stratified sample of the rest) and resolve
   F005/F010's specific flagged rows.
4. Re-run the full-scale versions of the fresh sealed benchmark, adversarial exact benchmark, and
   resource pilot that this audit executed at reduced scale (see `AUDIT_PROTOCOL.md` §4 and
   `FRESH_HIDDEN_PROTOCOL.md` §5 for the exact scope reductions taken and why).
5. Only after all of the above, commission a new independent audit pass with fresh sealed seeds
   (per Section F's requirement that any post-fix hidden benchmark use new seeds not previously
   disclosed to whoever fixes the code).

## 5. Mandatory Final Status Block

```text
V3_DESIGN_AUDIT=COMPLETE
UPSTREAM_INTEGRITY=PASS
REAL_DATA_REMAIN_SEALED=YES
LEXICON_ROWS_AUDITED=27
LEXICON_PROVENANCE_VALID_RATE=0.630
LEXICON_SCOPE=PARTIAL_BUT_USABLE
LEXICON_INDEPENDENCE=FAIL
SYNTHETIC_LEAKAGE=NOT_DETECTED
SCORING_PARITY=PASS
MATCHING_PARITY=PASS
BRANCH_AND_BOUND_EXACTNESS=PASS
CP_SOLVER_CLASSIFICATION=OTHER
HYBRID_VERIFICATION=INVALID
ADVERSARIAL_GLOBAL_OPTIMUM_RATE_BRANCH_AND_BOUND=1.000
ADVERSARIAL_GLOBAL_OPTIMUM_RATE_CP_SAT=0.200
ADVERSARIAL_GLOBAL_OPTIMUM_RATE_DIRECTIONAL_HEURISTIC=1.000
FRESH_ZERO_NOISE_TABLE_RECOVERY=0.000
FRESH_TEN_PERCENT_NOISE_TABLE_RECOVERY=0.000
FRESH_TEN_PERCENT_MAPPING_PRECISION=0.138
FRESH_TEN_PERCENT_MAPPING_RECALL=0.138
FRESH_TEN_PERCENT_HELDOUT_ASSIGNMENT_ACCURACY=0.286
WORST_REQUIRED_MODE_STATUS=FAIL
REAL_SCOPE_REACHABILITY_RECALCULATED=0.9649
NULL_MODEL_SELECTION_PARITY=FAIL
NULL_RESOURCE_FEASIBILITY=PASS
ORDER_INVARIANCE=PASS
CHECKPOINT_IDENTITY=PASS
CRITICAL_FINDINGS=5
MAJOR_FINDINGS=8
MINOR_FINDINGS=4
AUDIT_DECISION=FAIL_DO_NOT_PROCEED
V3_PRODUCTION_PACKAGE_AUTHORIZED=NO
REAL_DATA_SEARCH_AUTHORIZED=NO
```

`REAL_SCOPE_REACHABILITY_RECALCULATED=0.9649` is the closest genuinely-computed value in
`REACHABILITY_RECALCULATION.tsv` to the design's claimed 0.9625 (see F002); it is **not** an
endorsement of the design's number, which remains unsupported by its own evidence file.
`ADVERSARIAL_GLOBAL_OPTIMUM_RATE` and `NULL_RESOURCE_FEASIBILITY` are filled in from
`ADVERSARIAL_EXACT_RESULTS.tsv` and `RESOURCE_BENCHMARK.tsv` respectively (see those files for the
per-algorithm breakdown; CP_SAT's rate at k≥3 is 0%, which is the operative number for
`CP_SOLVER_CLASSIFICATION=OTHER` above).
