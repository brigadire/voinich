# Validation report

The fixed-table oracle passed 100 deterministic synthetic tests with exhaustive agreement, repeated EVA occurrences, multiple attestations, unmatched labels, co-optimal alternatives, and infeasible support cases. The realistic synthetic D1 benchmark used 57 labels, 299 lexicon rows, 94 identities, 26 source units, and 16 target units; it built immediately, captured objective 35 with a support-valid complete assignment, and passed checkpoint resume and serialization checks.

The full monolithic equivalence gate is `PARTIAL`: fixed-table matching semantics agree on small cases, but a 100-instance end-to-end comparison against the real monolithic CP-SAT formulation is not claimed because the latter fails at real-scale model construction. No real smoke test was executed and no scientific result is reported.

The synthetic-positive convergence probe recorded best objective 5 then 7 across independent deterministic seeds, with monotone best-so-far improvement. The five-second real smoke completed build and search start but captured no feasible incumbent; this is a technical smoke outcome only.
