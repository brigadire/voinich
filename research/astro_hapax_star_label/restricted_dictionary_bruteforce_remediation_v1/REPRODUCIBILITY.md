# Reproducibility

All commands below are run from this package's root (`restricted_dictionary_bruteforce_remediation_v1/`)
using the Python environment at `/home/brigadire/.venv/bin/python3` (has `numpy`, `scipy`,
`networkx`, and OR-Tools `ortools` 9.15 installed — none of these were present in the base system
Python; OR-Tools was installed specifically for this package to have a genuine CP-SAT solver rather
than repeat v3_design's fabricated `search_cp.py`, see `V3_DISPOSITION.md`).

## Fast checks (seconds)

```
python3 scripts/run_scoring_parity.py 500 SCORING_PARITY_RESULTS.tsv        # 500/500, 0 mismatches
python3 scripts/run_exact_validation.py 500 EXACT_SOLVER_VALIDATION.tsv     # 500/500, 0 mismatches
python3 scripts/discriminative_benchmark.py DISCRIMINATIVE_BENCHMARK_RESULTS.tsv
python3 scripts/reachability.py REACHABILITY_REMEDIATION.tsv
python3 scripts/safety_tests.py                                            # 10/10 after REMEDIATION_REPORT.md exists
```

## Slower checks (minutes)

```
python3 scripts/run_cpsat_vs_oracle.py 150 CPSAT_ORACLE_PARITY.tsv         # ~3 min, 0/150 mismatches
```

## Sealed benchmark (tens of minutes; see SEALED_BENCHMARK_PROTOCOL.md for the full ordered steps)

```
sha256sum scripts/scorer.py scripts/oracle_bb.py scripts/solver_cpsat.py scripts/solver_heuristic.py \
  scripts/hybrid.py scripts/generator.py scripts/sealed_harness.py SEALED_BENCHMARK_PROTOCOL.md \
  > FROZEN_CODE_SHA256SUMS.txt                      # must match the committed file byte-for-byte
python3 scripts/run_dev_grid.py DEV_GRID_RESULTS.tsv              # dev-only, never mixed with hidden
python3 scripts/generate_hidden.py ../hidden                      # fresh disjoint seed range
python3 scripts/run_hidden_solver.py ../hidden/PUBLIC_SEALED.jsonl ../HIDDEN_PREDICTIONS.tsv
python3 scripts/reveal_and_score.py ../hidden/TRUTH_SEALED.jsonl ../HIDDEN_PREDICTIONS.tsv ../HIDDEN_RESULTS.tsv
python3 scripts/summarize_results.py HIDDEN_RESULTS.tsv
```

## Null feasibility (tens of minutes)

```
python3 scripts/null_feasibility.py NULL_FEASIBILITY_RUNS.tsv checkpoints
python3 scripts/aggregate_null_feasibility.py NULL_FEASIBILITY_RUNS.tsv NULL_FEASIBILITY_RESULTS.tsv 60000
python3 scripts/null_feasibility_60s_pilot.py NULL_FEASIBILITY_60S_PILOT.tsv
python3 scripts/aggregate_null_feasibility.py NULL_FEASIBILITY_60S_PILOT.tsv NULL_FEASIBILITY_60S_RESULTS.tsv 60000
```

## Budget-scaling audit (~15-20 minutes)

```
python3 scripts/budget_scaling_audit.py BUDGET_SCALING_AUDIT_RESULTS.tsv
```
Run in response to a methodological review of the sealed benchmark's first-pass
`FAILURE_DIAGNOSIS.md` — see `BUDGET_SCALING_AUDIT.md` and the amendment log in
`SEALED_BENCHMARK_PROTOCOL.md` for why this exists and what it found (a real generator bug and a
config-drift bug, on top of the resource-budget question it was designed to test).

## Determinism

- `HeuristicSolver`/`HybridSolver` are deterministic given `master_seed`: repeated calls with the
  same seed and inputs produce byte-identical tables and fitness (verified interactively; not
  re-checked by `safety_tests.py`'s automated suite since it would require re-running the full
  sealed grid twice).
- All seed derivations use either an explicit integer or `discriminative_benchmark.py`'s
  `stable_seed()` (SHA-256-based). Python's built-in `hash()`/`__hash__()` on strings is never used
  for a seed anywhere in this package after the fix documented in `SCIENTIFIC_SAFETY_TESTS.md` —
  it is per-process-salted (`PYTHONHASHSEED`) and produced a real, caught non-reproducibility bug
  during development.
- Map/dict/set iteration order is never relied upon for numeric results: `scorer.py` and
  `independent_scorer.py` both sort candidate keys/labels before any accumulation or matching.

## Environment notes

- `scripts/solver_cpsat.py` requires OR-Tools; if unavailable, `CP_SAT` rows report
  `OUT_OF_SCOPE`/absent rather than falling back to a fake result — no script silently substitutes
  a different algorithm under the `CP_SAT` label.
- All wall-clock figures in `NULL_FEASIBILITY_RESULTS.tsv` and this package's resource claims were
  measured on this machine (single process, `num_workers=1` for CP-SAT) and are not portable to
  other hardware without re-measurement — `RESOURCE_BENCHMARK.tsv`-style unlogged single-data-point
  claims (v3_design's Finding F016) are exactly what the timing methodology here is designed not to
  repeat: every reported time is a distribution over >=150 (parity) or >=1000 (null feasibility)
  measured replicas, with an explicit CI.
