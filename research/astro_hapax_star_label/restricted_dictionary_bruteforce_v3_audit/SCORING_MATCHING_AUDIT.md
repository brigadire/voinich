# Scoring and Matching Audit (D1-D3)

## D1. Single Implementation Across Pipelines

`engine.py::evaluate_table` (normalization, `encode_word` transformation, `maximum_bipartite_matching`,
`compute_complexity`, capacity handling, tie-breaking via deterministic sort) is imported and called by:
- `run_small_instance_benchmark.py` (both the exhaustive ground truth **and** all three solvers)
- `search_bb.py`, `search_cp.py`, `search_heuristic.py` (all three production solvers)
- `synthetic_generator.py` only for label construction (`encode_word`), not for scoring

**No alternate/simplified scorer exists anywhere in the package.** The one place a *different*
code path is used to produce "results" is `run_synthetic_recovery_benchmark.py`'s fabricated
suites (see F001) — but that is not a simplified scorer, it is literal constant data with no
scorer involved at all, which is a strictly worse violation than D1 anticipates (a hidden
simplified scorer would at least have *computed something*).

**Verdict: `SCORING_PARITY = PASS`** for the code that does compute scores; the caveat is that a
large fraction of the design package's *reported* numbers were never produced by this scorer in
the first place (tracked separately under F001, not a D1 violation per se).

## D2. Assignment Correctness — Independent Bipartite Solver Cross-Check

`engine.maximum_bipartite_matching` uses a recursive Kuhn's-algorithm augmenting-path DFS. The
audit implemented a **from-scratch, independently-coded** Hopcroft-Karp-style BFS+DFS matcher
(`audit_verification.py::independent_max_matching`) and cross-checked matching **cardinality** and
**per-page/global capacity validity** against `engine.maximum_bipartite_matching` on 200 randomized
instances (3-14 labels, 3-14 identities, random sparse adjacency) × 2 capacity policies
(`PER_PAGE_CAPACITY_1`, `GLOBAL_CAPACITY_1`) = 400 trials (`INDEPENDENT_MATCHING_CHECK.tsv`).

**Result: 0 cardinality mismatches, 0 capacity-validity violations out of 400 trials.**

This confirms:
- maximum-cardinality matching is computed correctly under both capacity regimes,
- `PER_PAGE_CAPACITY_1` correctly solves each page as an independent sub-problem (verified via
  `test_capacity_policy_independence` in the design's own unit tests, and by the audit's
  per-page independent-matcher reconstruction),
- no duplicate-identity-per-page violations occur.

Held-out/train separation: `evaluate_table` has no concept of "held-out" internally — held-out
scoring in the design's synthetic recovery protocol is (or would be, once genuinely implemented)
the responsibility of the *caller*, which never leaks held-out labels into the table search itself
in any of the genuinely-executed code (`run_small_instance_benchmark.py`,
`audit_verification.py`'s fresh hidden benchmark). This is architecturally sound.

**Verdict: `MATCHING_PARITY = PASS`.**

## D3. Complexity Penalty

`compute_complexity` applies a uniform formula (`len(table)` + mode/deletion/abbreviation
surcharges) identical across all three solvers and the exhaustive benchmark — confirmed by direct
code reading (single function, no per-solver override). The penalty:
- scales with table size (encourages parsimony),
- adds a fixed surcharge for `MERGE_1`, `DROP_UNMAPPED`/`SELECTIVE_VOWEL_DROP`, and each
  abbreviation mode,
- is **not** sensitivity-tuned against hidden metrics anywhere in the package (it is a fixed
  formula, not a learned or swept parameter) — so the D3 requirement "hidden metrics must not be
  used to select the penalty" is trivially satisfied because no penalty selection process exists
  at all. This is acceptable but means the design package has not actually performed the "sensitivity
  analysis of penalty on development synthetic data" that Section D3 also calls for; no such
  analysis exists in the package (`grep -rn "sensitivity" scripts/*.py *.md` finds no penalty
  sweep). This is a gap, not a violation of the no-hidden-tuning rule.

**Verdict: no violation found, but the required penalty-sensitivity analysis is simply absent from
the design package and should be added before freeze (MINOR/MAJOR documentation gap, not a leakage
finding).**
