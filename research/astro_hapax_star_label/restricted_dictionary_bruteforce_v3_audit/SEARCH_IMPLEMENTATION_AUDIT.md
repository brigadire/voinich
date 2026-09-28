# Search Implementation Audit (E1-E5)

## E1. Branch-and-Bound

`search_bb.py` implements a genuine branch-and-bound over the table space with an admissible
relaxed-matching upper bound (`_compute_upper_bound`) and a `map/skip` branching rule per source
character. Independent verification:

- **Re-run in sandbox**: `EXACT_SMALL_INSTANCE_RESULTS.tsv` reproduces byte-identical (mod
  runtime/memory columns) on independent re-execution — genuine, not fabricated.
- **Adversarial k-fragment test** (`ADVERSARIAL_EXACT_RESULTS.tsv`): BRANCH_AND_BOUND reaches the
  true global optimum in **100%** of trials for k=2..6, matching brute-force ground truth for k≤4
  and its own independently-corroborated result for k=5,6 (see methodology note in that file).
- **Timeout/certificate discipline is correct**: `global_optimum_guaranteed` is set to
  `not self.timed_out`, and the branch function does check the wall clock and bail out cleanly.
  This is the one algorithm in the package whose completeness claim is actually earned by its code.
- MAJOR (carried from F004): the small-instance benchmark's own difficulty is inadequate to prove
  B&B outperforms chance at this scale — a 10-sample `RANDOM_BASELINE` also reaches the global
  optimum on 25/25 instances, because every instance has 2-10 tied co-optimal tables. This is a
  benchmark-design weakness, not a B&B implementation defect.

**Verdict: BRANCH_AND_BOUND_EXACTNESS = PASS** (genuinely exact within its explored branches and
correctly reports `TIMEOUT` vs certified optimum), subject to the small-instance benchmark
adequacy caveat above.

## E2. CP Solver — `search_cp.py`

This is the most severe implementation-level finding of the audit (see `AUDIT_FINDINGS.tsv` F003).

`search_cp.py` is **not** a CP-SAT or complete constraint-propagation solver. Reading the code:
1. It extracts, per (lexicon word, label) pair, all partial mappings consistent with a
   `DROP_UNMAPPED` subsequence alignment (`_extract_consistent_mappings`).
2. It ranks these partial mappings by how many (word,label) pairs agree on them.
3. It evaluates each ranked candidate **individually**.
4. It then tries **only pairwise combinations** of the top 50 ranked candidates
   (`for i ... for j in range(i+1, ...)`, `search_cp.py` lines ~174-213) — i.e. tables built from at
   most **2** pre-ranked fragments.
5. It unconditionally returns `"global_optimum_guaranteed": not self.timed_out` — claiming a
   completeness certificate whenever the wall clock hasn't expired, regardless of whether the
   search space it explored was remotely close to exhaustive.

There is no arc-consistency propagation across all `table_size` positions simultaneously, no
backtracking search tree, and no mechanism to combine 3 or more independently-informative
fragments. For any planted table whose maximum-fitness reconstruction requires ≥3 fragments not
already present as a single frequent partial mapping, this solver is **structurally incapable** of
finding it, no matter the time budget.

**Independent empirical confirmation** (`ADVERSARIAL_EXACT_RESULTS.tsv`, constructed so the unique
optimum requires exactly k independent single-character fragments):

| k (fragments required) | CP_SAT global-optimum rate |
|---|---:|
| 2 | 100% |
| 3 | **0%** |
| 4 | **0%** |
| 5 | **0%** |
| 6 | **0%** |

At k≥3 — well within the table sizes (8-12) the design report recommends for production — CP_SAT
finds the true optimum **zero times out of five trials**, while `global_optimum_guaranteed` was
still reported `True` (not timed out) on every one of those failing runs. This is a **false
completeness certificate**, not merely a slow/underpowered search.

Separately, the candidate-generation phase (`pair_mappings` construction, step 1 above) has **no
wall-clock check at all** — the `time_limit_sec` guard is only checked inside the two enumeration
loops that follow. On the real 94-identity/307-attestation lexicon at table_size=8 this phase
itself already dominates runtime; on pathological word/label combinations it has no safety bound.

**Verdict: `CP_SOLVER_CLASSIFICATION = CONSTRAINT_PROPAGATION`** is too generous — this is closer to
a **bounded pairwise-candidate heuristic**; it should not be named `CP_SAT`, and its
`global_optimum_guaranteed` field must not be trusted as a completeness certificate at any table
size ≥3 fragments.

## E3. Directional Heuristic

- **Order invariance**: independently re-tested across **100** random label-order permutations
  (not just the single reversal the design package tested). All 100 permutations produced the
  identical best fitness and identical best table as the unpermuted run
  (`ORDER_INVARIANCE_100.tsv`, 0 mismatches). **PASS**, and more thoroughly verified than the
  original design package's single-permutation test.
- Seed handling, stopping rule (time budget + fixed `local_search_steps`), and mutation operators
  (`swap_tgt`/`change_tgt`/`change_src`) are deterministic given `seed`, and correctly reject
  mapping-mode-invalid mutations before scoring.
- The heuristic is honestly *not* claimed to be exact (`"global_optimum_guaranteed": False`
  hardcoded) — this is the one place in the codebase where the completeness claim matches the code.
- On the adversarial k-fragment instances, the heuristic actually reached the true optimum 100% of
  the time for k=2..6 — better than CP_SAT on this specific adversarial construction, though this
  should not be read as a general result (the design's own fabricated Suite 1/2 numbers are the
  ones that were supposed to establish real-scale heuristic performance; see `AUDIT_FINDINGS.tsv`
  F001 and the real replacement numbers in `FRESH_HIDDEN_RESULTS.tsv`).

## E4. Meaning of "HYBRID"

`SEARCH_ARCHITECTURE_DECISION.md` describes the certificate engine as being "deployed on restricted
candidate neighborhoods... to verify that incumbent solutions cannot be improved within the local
topological ball" around the heuristic's output. **No code in the package implements this.** Every
benchmark script instantiates `BranchAndBoundSolver`, `CPSATSolver`, and `DirectionalHeuristicSolver`
as three fully independent, whole-table solvers; none of them ever receives another solver's
incumbent as an input or search-space restriction (grep-verified: no function anywhere accepts a
heuristic table and narrows a subsequent B&B/CP call around it). The three `GLOBAL_OPTIMUM_CERTIFIED
/ LOCAL_NEIGHBORHOOD_OPTIMUM / NO_IMPROVEMENT_FOUND / VERIFICATION_TIMEOUT` statuses the task
specification requires this design to define are **not defined anywhere** in the package, and there
is no `HYBRID` class, function, or orchestration script at all.

**Verdict: `HYBRID_VERIFICATION = INVALID`** — "HYBRID" as described in prose does not exist as
code; `SELECTED_SEARCH_ARCHITECTURE=HYBRID` is unimplemented.

## E5. Adversarial Exact Benchmark

See `ADVERSARIAL_EXACT_RESULTS.tsv` for the full k=2..6 × 5-trial matrix (30 rows). Summary:
Branch-and-bound and the heuristic both recover the score-optimal fitness on every trial;
CP_SAT fails categorically for k≥3. Critically, "reached the global-optimum **fitness**" and
"recovered the **true planted table**" are tracked as separate columns throughout this audit's own
outputs (`reached_global_optimum` vs a table-identity check) precisely because
`EXACT_SMALL_INSTANCE_RESULTS.tsv` already showed these two things diverge whenever ties exist
(F004) — this audit does not repeat the original report's conflation of the two.
