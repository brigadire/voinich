# Reproducibility Notes (Audit Package)

## Environment
- Python 3.14.6, single reference host, 12 logical cores.
- No third-party packages used beyond the Python standard library (no `scipy`/`networkx` available
  or required — the independent bipartite matcher in `scripts/audit_verification.py` is a
  from-scratch implementation for exactly this reason).

## How to Reproduce Every Number in This Audit

1. **Reachability recomputation** (`REACHABILITY_RECALCULATION.tsv`):
   ```
   cd repro_sandbox/scripts && python3 analyze_reachability.py
   diff ../REACHABILITY_ANALYSIS.tsv ../../restricted_dictionary_bruteforce_v3_design/REACHABILITY_ANALYSIS.tsv
   ```
   Reproduces byte-identical (confirmed during this audit).

2. **Small-instance benchmark** (cited in `SEARCH_IMPLEMENTATION_AUDIT.md`):
   ```
   cd repro_sandbox/scripts && python3 run_small_instance_benchmark.py
   diff <(cut -f1-11 ../EXACT_SMALL_INSTANCE_RESULTS.tsv) <(cut -f1-11 ../../restricted_dictionary_bruteforce_v3_design/EXACT_SMALL_INSTANCE_RESULTS.tsv)
   ```
   Reproduces identical on all substantive columns (fitness, optimum flags, precision/recall/F1,
   assignment accuracy, planted-table rank, optimal-tables-count); only `runtime_ms`/
   `peak_memory_kb` differ, as expected on different hardware.

3. **Unit tests**:
   ```
   cd repro_sandbox && python3 -m unittest discover -s tests -v
   ```
   5/5 pass.

4. **Full independent verification suite** (adversarial k-fragment test, 100-permutation order
   invariance, 400-trial independent bipartite matcher, fresh sealed hidden benchmark, resource
   pilot):
   ```
   python3 scripts/audit_verification.py
   ```
   This is deterministic given the fixed seed constants inside the script (all seeds are literal
   integers, no wall-clock or OS entropy is used for anything except timing/memory measurement
   columns). Re-running reproduces identical `reached_global_optimum`, `order_invariant`,
   `cardinality_match`, `exact_table_recovery`, `mapping_precision/recall`, and
   `assignment_accuracy` columns; only timing/memory columns and the final resource-benchmark
   wall-clock statistics vary run-to-run (expected, disclosed as such in `RESOURCE_BENCHMARK.tsv`).

## Non-Determinism Explicitly Isolated

- `runtime_ms` / `runtime_sec` / `peak_memory_kb` columns in every TSV in this package are
  hardware-dependent and are not claimed to be reproducible bit-for-bit; every other column is.
- `RESOURCE_BENCHMARK.tsv`'s extrapolations are explicitly labeled with the pilot size (`n_pilot_runs`)
  and reported with min/mean/p95/max, not a single point estimate, precisely because the design
  package's own `NULL_FEASIBILITY_REPORT.md` asserted a single unlogged "0.65s measured" data point
  with no pilot size, no variance, and no reproducing command (Finding F016).

## What This Audit Did Not Attempt to Reproduce

- The original `SYNTHETIC_RECOVERY_RESULTS.tsv` / `ALGORITHM_COMPARISON.tsv` cannot be "reproduced"
  because they were never produced by execution in the first place (F001) — there is nothing to
  re-run. `FRESH_HIDDEN_RESULTS.tsv` is offered as the honest replacement measurement at the same
  operating point, not a reproduction of the fabricated file.
