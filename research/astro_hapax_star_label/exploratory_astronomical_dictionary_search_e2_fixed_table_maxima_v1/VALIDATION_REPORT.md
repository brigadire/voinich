# Validation report

For each fixed table, all compatible paths whose rule set is a subset of the table were enumerated. Exact branch-and-bound then enforced distinct labels, distinct canonical identities, and distinct-EVA support of every active rule. Unused parent-table rules were absent from the assignment's active support set.

All 90 mode/table runs completed with status `OPTIMAL`; no timeout or implementation gate failure occurred. The cross-page mode returned zero feasible assignments for every table, including assignments of any size.

Every retained assignment is represented in `OPTIMAL_ASSIGNMENTS.tsv` and independently preserves the graph's scorer-produced encoded token.
