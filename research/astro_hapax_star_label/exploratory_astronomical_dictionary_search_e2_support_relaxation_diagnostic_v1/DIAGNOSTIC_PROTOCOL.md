# Support relaxation diagnostic

The experiment keeps `k=5`, `GLOBAL_CAPACITY_1`, `BALANCED`, the frozen target scope, lexicon, and transformation semantics. Three deterministic seeds use identical 30-second budgets at each support threshold: 0, 1, and 2 distinct EVA token types.

Support is measured from the final exact fixed-table matching assignment. The diagnostic records raw coverage, support-valid coverage, counts of rules with 0, 1, or at least 2 distinct EVA types, violating rules, correspondence dependencies, and coverage after removing unsupported rules. It is exploratory only.
