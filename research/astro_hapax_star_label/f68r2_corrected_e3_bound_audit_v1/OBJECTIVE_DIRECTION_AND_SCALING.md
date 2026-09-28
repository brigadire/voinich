# Objective direction and scaling

The old source model (`f68r2_corrected_e3_search_v1/e3_search.py`) calls `m.Maximize(sum(x))`, where each selected path variable has coefficient 1. Therefore, if a genuine CP-SAT response were available, objective and coverage would have conversion `coverage = objective_value`, and an upper bound would convert as `coverage_upper = best_objective_bound`; scaling is 1 and offset 0.

The committed `PROFILE_STATUS.json` files preserve only `solver_status`, `incumbent`, and `bound`. They do not preserve objective value, response protobuf, solution info, branches, conflicts, presolve/search timing, or stop reason. Consequently the 24 numeric `0.0` values cannot be independently classified as certified upper bounds. They are classified `BOUND_METADATA_MISSING`.

The sentinel tests exercise known-positive models under zero and tiny budgets. Their output is the authoritative test of whether `UNKNOWN` plus a zero-looking bound can arise without an infeasibility certificate.
