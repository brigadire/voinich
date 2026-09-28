# Global maximum derivation v2

The prior resumable package was treated as read-only. All 64 committed checkpoints were revalidated. Its 24 UNKNOWN checkpoints contain only a numeric `bound=0.0`; raw CP-SAT response data and objective metadata were not persisted.

The source model maximizes `sum(selected_path_variables)` with coefficient 1, so a genuine solver response would convert as `coverage = objective` and `coverage_upper = best_objective_bound`. That conversion alone does not establish provenance for the old numeric field.

The sentinel experiment on the same OR-Tools runtime used a known-positive one-variable model with optimum 1. Under zero or near-zero budgets it returned `UNKNOWN`, objective 0.0, and best bound 0.0. Thus the old UNKNOWN/0.0 combinations were conservatively classified as `BOUND_METADATA_MISSING`, not as certificates.

All 24 UNKNOWN profiles were then solved independently with the exact `coverage >= 4` decision model against their frozen committed path graphs. Every query returned `INFEASIBLE`; every result passed independent global replay. Therefore the 24 profiles are now directly closed. Together with the 40 previously exact profiles (maximum at most 3), this establishes:

`GLOBAL_E3_MAXIMUM=3`, `GLOBAL_E3_MAXIMUM_CERTIFIED=YES`.

No null run was performed or authorized. Coverage 3 remains a technical exploratory witness and is not a scientific correlation.
