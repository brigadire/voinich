# Structural support report

The compatibility graph contains 372 individual source-to-EVA rules with potential support from at least two distinct EVA token types. A table of size at most one is therefore structurally potential-support feasible, and the same is true for every k through five. This is a graph-level statement; it does not certify that one table simultaneously realizes support after one-to-one matching.

The union-graph matching upper bounds are:

| k | raw coverage upper bound | potential-support coverage upper bound | sampled realized support-valid coverage |
|---:|---:|---:|---:|
| 1 | 1 | 1 | 0 |
| 2 | 3 | 3 | 0 |
| 3 | 7 | 7 | 0 |
| 4 | 15 | 15 | 0 |
| 5 | 35 | 34 | 0 |

The first two columns are valid upper bounds from the union of all compatible identity-label edges; they are not claims that one table attains them. The realized column is a bounded deterministic sample, not an exact maximum.

The prior 3/57 tables were reconstructed with exact fixed-table matching. Their selected rules have realized support counts of zero or one distinct EVA type. Two of the three correspondences in the best table depend on the repeated `hy` token. They disappear when unsupported rules are removed.

Decision status: `SUPPORT_FEASIBLE`. Potentially supported rules exist in the current compatibility space, so `SUPPORT_GATE_INCOMPATIBLE_WITH_CURRENT_SPACE` is not justified. The exact realized support-valid frontier remains unresolved; no long optimization run or scientific claim is authorized by this audit.
