# HYBRID semantics

`scripts/hybrid.py:HybridSolver` is the only place in this package the term `HYBRID` is used,
because it is the only place real orchestration code wires a heuristic's output into a verifier
and classifies the result. This directly replaces v3_design's `SEARCH_ARCHITECTURE_DECISION.md`,
which used the same word with no such code behind it (Finding F009, CRITICAL — see
`COMPONENT_REUSE_REGISTRY.tsv`, `HYBRID_LABEL: REJECTED`).

## Pipeline

1. `HeuristicSolver` (multi-start simulated annealing, 8 starts x 400 iterations, deterministic
   per `master_seed`) produces an **incumbent** table and fitness.
2. A **verifier** explores a formally-defined space and the outcome is classified into exactly one
   of the 5 categories `clean_room.md` Section 11 requires. Which verifier runs depends on whether
   the instance falls inside `solver_cpsat.py`'s documented scope (`DROP_UNMAPPED` deletion,
   `NONE` abbreviation, single-character source alphabet):

   **In scope -> `EXACT_FULL_SPACE` verifier** (genuine CP-SAT over the *entire* table space, not
   a neighborhood):
   | CP-SAT status | Condition | Classification |
   |---|---|---|
   | `OPTIMAL` | — | `GLOBAL_OPTIMUM_CERTIFIED` |
   | `FEASIBLE` | `best_objective_bound == best_seen_fitness` | `GLOBAL_OPTIMUM_CERTIFIED` (the dual bound itself proves optimality even though the solver's own status flag says `FEASIBLE`, not `OPTIMAL` — this is the exact `FEASIBLE`-vs-`OPTIMAL` distinction Section 10 of the task requires never be confused) |
   | `FEASIBLE` | `best_objective_bound > best_seen_fitness` | `BOUNDED_GAP_CERTIFIED`, gap reported |
   | anything else (no incumbent found before the time limit) | — | `VERIFICATION_TIMEOUT` |

   **Out of scope -> `NEIGHBORHOOD(radius)` verifier** (exhaustive brute-force enumeration of every
   valid table within Hamming distance `radius` of the incumbent — tractable because `radius` is
   small, unlike the full table space):
   | Outcome | Classification |
   |---|---|
   | No table within `radius` beats the incumbent | `LOCAL_NEIGHBORHOOD_OPTIMUM` |
   | A better table is found | incumbent is replaced and the neighborhood is re-searched, up to `neighborhood_max_rounds` times; if it still hasn't converged to a local optimum by then, the (honest, weak) result is `NO_IMPROVEMENT_FOUND` rather than a fabricated stronger claim |

   This is the concrete case the task singles out: *"если verifier проверяет только локальную
   окрестность, это не глобальный сертификат"* — `LOCAL_NEIGHBORHOOD_OPTIMUM` is never reported as,
   or upgraded to, `GLOBAL_OPTIMUM_CERTIFIED`.

## Why two verifier modes, not one

`solver_cpsat.py` only exactly encodes `DROP_UNMAPPED`/`NONE`/single-character-alphabet instances
(see that file's docstring for why). Configurations outside that scope (e.g.
`SELECTIVE_VOWEL_DROP`, or a source alphabet with digraphs) have no exact full-space verifier
available in this package, so `HybridSolver` is honest about only being able to offer a weaker,
disclosed neighborhood check for them — it does not silently widen CP-SAT's scope or claim a
certificate class stronger than what was actually verified.

## Empirical check

`run_cpsat_vs_oracle.py` cross-validates the `EXACT_FULL_SPACE` path itself (CP-SAT vs the
independent branch-and-bound oracle) — see `SOLVER_REGISTRY.tsv` and `CPSAT_ORACLE_PARITY.tsv`.
