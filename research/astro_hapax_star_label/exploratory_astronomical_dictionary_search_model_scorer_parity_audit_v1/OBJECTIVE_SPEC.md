# Objective and parity specification

The CP-SAT model objective is `M = sum(assign[identity,label])`, the number of selected identity-label edges. The scorer reports `matched = |assignment|`, `complexity = table_size + deletion_cost`, and `fitness = 100*matched - complexity` for a valid table. Parity must first establish `M == matched`, then compare selected edges, identity sets, raw coverage, and distinct EVA types. The distinct-EVA support gate is a post-hoc scientific acceptance gate and is not currently encoded in the CP-SAT model.
