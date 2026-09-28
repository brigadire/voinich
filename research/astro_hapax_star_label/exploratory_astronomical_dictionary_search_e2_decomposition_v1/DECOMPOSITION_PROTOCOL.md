# E2 decomposition protocol

The frozen target scope, lexicon, transformation semantics, capacity policy, and distinct-EVA support definition are unchanged. D1 separates table search from matching: the master proposes canonical injective tables of size five, while the fixed-table oracle transforms all lexicon forms and solves the one-to-one identity/label assignment.

The realistic master is a deterministic beam search with an optimistic ordering proxy. It is therefore `HEURISTIC_ONLY`; it never emits a global optimum certificate. The fixed-table oracle is exhaustive on small instances and deterministic maximum matching on the realistic exploratory beam. Every accepted realistic incumbent is support-audited from its final assignment.

The real smoke test is deliberately not run in this package because the realistic gate is exploratory and has no exact global-optimum guarantee yet.
