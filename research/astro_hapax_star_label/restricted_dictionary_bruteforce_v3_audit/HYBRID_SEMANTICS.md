# Formal Semantics of "HYBRID" (Required by Task Section E4)

The task requires this audit to formally define, for the claimed HYBRID architecture:
what the heuristic passes to the verifier, the size of the verified neighborhood, which
mappings may change, what bounds are computed, what counts as a confirmed optimum, and how
timeout is handled — and to use the four statuses `GLOBAL_OPTIMUM_CERTIFIED`,
`LOCAL_NEIGHBORHOOD_OPTIMUM`, `NO_IMPROVEMENT_FOUND`, `VERIFICATION_TIMEOUT` rather than
conflating local and global optimality.

## Finding

None of this exists in the current codebase. There is no function, class, or script anywhere in
`restricted_dictionary_bruteforce_v3_design/scripts/` that:
- accepts a `DirectionalHeuristicSolver` result as input,
- constructs a "restricted candidate neighborhood" (e.g. a bounded set of single-position table
  edits) around it,
- and hands that restricted neighborhood to `BranchAndBoundSolver` or `CPSATSolver` for exact
  verification.

`BranchAndBoundSolver.solve()` and `CPSATSolver.solve()` each search the **entire** table space
(bounded only by `table_size` and the two solvers' own internal candidate-generation heuristics),
completely independently of any heuristic run. This was confirmed by:
1. Reading every call site of the three solver classes (`run_small_instance_benchmark.py`,
   `run_synthetic_recovery_benchmark.py`) — each instantiates all three independently on the same
   `(lexicon, labels)` input, never chaining one solver's output into another's constructor or
   `solve()` call.
2. Grepping for `neighborhood`, `incumbent`, `restrict` across all of `scripts/` — the only hit is a
   comment inside `search_bb.py` describing its own internal empty-table starting incumbent, unrelated
   to any heuristic hand-off.

## Consequence

Because the wiring does not exist, none of the four required statuses
(`GLOBAL_OPTIMUM_CERTIFIED`/`LOCAL_NEIGHBORHOOD_OPTIMUM`/`NO_IMPROVEMENT_FOUND`/`VERIFICATION_TIMEOUT`)
can currently be produced by this package for any real run. `SEARCH_ARCHITECTURE_DECISION.md`'s
prose description of the HYBRID architecture is a specification for code that has not yet been
written, not a description of what the certified benchmarks in this design package actually ran.

Everywhere `V3_DESIGN_REPORT.md` / `VALIDATION_REPORT.md` state `SELECTED_SEARCH_ARCHITECTURE=HYBRID`
or `SEARCH_ARCHITECTURE_GATE=PASS` on the strength of a "Hybrid Architecture" row in
`SEARCH_ARCHITECTURE_DECISION.md`'s decision matrix, that row's numbers are in fact simply copied
from the (fabricated, see `AUDIT_FINDINGS.tsv` F001) `DIRECTIONAL_HEURISTIC`-only results — there was
never a genuinely-hybrid run to report on.

## Required Before Any Future Freeze

1. Implement the actual hybrid wiring: given a heuristic incumbent table `T`, define the verified
   neighborhood explicitly (e.g. "all tables differing from `T` in at most `d` source→target
   reassignments, `d` fixed and documented"), and pass that neighborhood definition into a real
   bounded-exact verifier (a genuinely complete CP/SAT solver or a correctly-scoped B&B call).
2. Emit one of the four required statuses per run, with `GLOBAL_OPTIMUM_CERTIFIED` reserved for
   the case where the verifier can prove no table in the *entire* space (not just the local
   neighborhood) improves on the incumbent.
3. Re-run the certified benchmarks (small-instance, synthetic recovery, adversarial) against this
   actual implementation before re-claiming `SEARCH_ARCHITECTURE_GATE=PASS`.

**`HYBRID_VERIFICATION = INVALID`** (no implementation exists to classify as GLOBAL or LOCAL).
