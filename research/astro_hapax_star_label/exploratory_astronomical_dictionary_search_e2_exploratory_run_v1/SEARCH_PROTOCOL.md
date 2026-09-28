# E2 exploratory search protocol

Six independent deterministic seeds were run with identical `k=5`, `GLOBAL_CAPACITY_1`, `BALANCED`, and 900-second budgets. Each seed used canonical rule ordering with seed-dependent deterministic beam tie-breaking. No E1 warm start was used because its support validity is zero.

Only tables with a complete exact fixed-table one-to-one matching and distinct-EVA support for every active rule could become incumbents. Each accepted incumbent would retain the full table, selected identity-label edges, support audit, unmatched labels, assignment hash, objective, and seed. The master is heuristic and no global optimum is certified.
