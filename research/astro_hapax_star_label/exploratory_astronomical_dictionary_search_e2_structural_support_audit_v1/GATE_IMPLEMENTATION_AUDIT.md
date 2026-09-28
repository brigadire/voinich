# Gate implementation audit

The current monolithic remediated CP-SAT model has four relevant properties:

1. `assigned` is constrained to exactly `table_size`, so the legacy model requires exactly five active rules. It does not represent `table_size <= k`.
2. Support constraints are attached to every active `is_t[c,t]` rule. An active rule therefore needs two distinct EVA token types even when no final assignment uses that rule. This is stricter than the scientific definition, which is post-assignment and applies only to used rules.
3. The support literals are linked to assignment, matching, and the active rule, so the model does not count repeated occurrences as independent support. The distinct-EVA unit is correct.
4. The decomposed diagnostic computes support after the exact matching assignment and separately recomputes coverage after removing unsupported rules. It does not reject a table merely because an unused rule is unsupported.

The structural audit therefore finds no evidence that the distinct-EVA definition itself is impossible. It does identify the exact-five active-rule constraint and the monolithic pre-search support encoding as implementation barriers.
