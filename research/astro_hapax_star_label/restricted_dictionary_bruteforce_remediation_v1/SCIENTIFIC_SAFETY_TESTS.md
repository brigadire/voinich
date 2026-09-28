# Scientific Safety Tests

`scripts/safety_tests.py` implements the automated prohibitions clean_room.md Section 20 requires,
as real runnable assertions, not prose promises. Run with:
```
python3 scripts/safety_tests.py
```

## Checks

| Check | What it verifies |
|---|---|
| `NO_TRUTH_LEAKAGE_IN_SOLVERS` | none of `oracle_bb.py`/`solver_cpsat.py`/`solver_heuristic.py`/`hybrid.py`'s actual code (docstrings/comments excluded) references `true_identity`, `held_out_labels`, or `TRUTH_SEALED` — a solver cannot see ground truth during optimization even in principle. |
| `HIDDEN_SOLVER_NEVER_OPENS_TRUTH` | `run_hidden_solver.py`'s code never references `TRUTH_SEALED` or `true_identity` — enforces `SEALED_BENCHMARK_PROTOCOL.md` step 8 (predictions exist before truth is read). |
| `REVEAL_SCRIPT_IS_SEPARATE_FROM_SOLVER` | `reveal_and_score.py` imports no solver class — it can only read predictions + truth, never re-run or influence solving after truth is known. |
| `CPSAT_OPTIMAL_ONLY_FROM_SOLVER_STATUS` | `is_optimal` is derived from OR-Tools' own `CpSolverStatus.OPTIMAL`, never asserted independently — the exact defect that made v3's `search_cp.py` false (F003) cannot recur here. |
| `BB_NEVER_CERTIFIES_ON_TIMEOUT` | `global_optimum_certified` is structurally `not timed_out` in `oracle_bb.py` — cannot be `True` on a timed-out run. |
| `NO_REAL_LABEL_DATA_ACCESS` | no script references the real-label directories (`restricted_m2_star_labels_v1`, `restricted_hapax_enrichment_v1`, `restricted_m0_m1_rerun_v1`), `TARGET_SCOPE.tsv`, or `EVA_LABEL`. (Synthetic page names `f68r1`/`f68r2` in `generator.py` are a disclosed naming convention for synthetic pages, not real content access — see `SYNTHETIC_GENERATOR_SPEC.md`.) |
| `REPRODUCIBLE::<script>` | re-running `run_scoring_parity.py`, `discriminative_benchmark.py`, and `reachability.py` byte-for-byte reproduces their committed TSVs — no hand-typed numbers, ever, anywhere a generator script exists for that output. |
| `GATE_DISCIPLINE::REMEDIATION_REPORT.md` | the final report's mandatory status block always contains `REAL_DATA_SEARCH_AUTHORIZED=NO`, regardless of every other gate's outcome. |

## Pre-registered invariance/identity gates

`scripts/invariance_tests.py` (run with `python3 scripts/invariance_tests.py`) implements the
three structural gates `SEALED_BENCHMARK_PROTOCOL.md` pre-registers:

| Gate | Method | Result |
|---|---|---|
| `ORDER_INVARIANCE` | 50 randomized trials: shuffle label order and reverse lexicon dict insertion order, confirm identical `matched`/`fitness`/`coverage` | 50/50 PASS |
| `ALPHABET_RENAMING_INVARIANCE` | 30 randomized trials: apply a consistent random permutation to the target alphabet and every label token, confirm the branch-and-bound oracle's certified optimum fitness is unchanged | 30/30 PASS |
| `CHECKPOINT_IDENTITY` | the checkpoint written mid-run at replica 100 of the null-feasibility pilot (`checkpoints/NULL_CHECKPOINT_00100.json`) is compared row-for-row against the corresponding first 100 rows of the completed `NULL_FEASIBILITY_RUNS.tsv` | exact match, PASS |

## A real critical bug this suite's execution caught (not a unit test — a live production hang)

`hybrid.py`'s `NEIGHBORHOOD` verifier fallback (used when an instance falls outside
`solver_cpsat.py`'s scope) originally generated the ENTIRE table space and filtered by Hamming
distance afterward. At `table_size=8` over the 16-symbol EVA alphabet that is `16**8` ~= 4.3
billion candidates before any filtering. This was not caught by any of the fast validation suites
(`discriminative_benchmark.py`'s cases are all `table_size<=6`) — it surfaced only when the actual
pre-registered dev grid hit a `table_size=8` cell using `SELECTIVE_VOWEL_DROP`/`SUSPENSION_1`
(outside CP-SAT's scope), and ran for **4.5 hours of CPU time** before being noticed and killed.
Fixed by constructing candidates directly from the incumbent (retarget `<=radius` of its existing
keys) instead of generate-then-filter; verified 0.11-0.13s on the previously-hanging cells
afterward. See `VALIDATION_REPORT.md` for the full account. This is the most severe defect this
package's own development process caught in itself.

## A real bug this suite caught

The first version of `discriminative_benchmark.py` derived its `RANDOM_BASELINE`, `HEURISTIC`, and
`HYBRID` per-case master seeds from Python's built-in `hash(case_id)` / `case_id.__hash__()`.
Python salts string hashing per-process (`PYTHONHASHSEED`) as a security hardening measure — it is
**not** a stable function across runs. Running `REPRODUCIBLE::discriminative_benchmark.py` twice in
a row produced different `RANDOM_BASELINE_SINGLE_DRAW_MEAN_OF_30` fitness values on 8 of 10 cases,
which is what surfaced this. Every seed derivation in `discriminative_benchmark.py` was replaced
with `stable_seed()` (SHA-256-based) before any result in this package was reported as final. This
is the same category of determinism defect this project has hit before in unrelated contexts (see
memory: Go map iteration order affecting float accumulation) — a reminder that "looks
deterministic" and "is deterministic across processes" are different claims, and only the second
one is checked here mechanically rather than assumed.

## What this suite does not cover

`safety_tests.py`'s `REPRODUCIBLE` check does not re-run the sealed dev/hidden grid or the 1000-run
null feasibility pilot (each takes tens of minutes; re-running them on every safety-test invocation
would be prohibitive). Their internal determinism is instead verified once, directly, via repeated
calls with an identical seed (see `SEALED_BENCHMARK_PROTOCOL.md` and `NULL_FEASIBILITY_RESULTS.tsv`
methodology notes) rather than via a full-grid re-run inside this script.

## Result

9/10 checks pass unconditionally; `GATE_DISCIPLINE::REMEDIATION_REPORT.md` passes once
`REMEDIATION_REPORT.md` exists with its mandatory final status block (verified as the very last
step of building this package, after which `safety_tests.py` was re-run to confirm 10/10).
