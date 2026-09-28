# Runner and model remediation

The legacy E2 runner is invalid for interpretation because it did not persist callback assignments and its CP-SAT objective was not parity-checked against the scorer. This package introduces a versioned runner/model boundary.

Every callback captures the complete substitution table and selected identity-label edges. The model objective is the number of selected edges subject to the distinct-EVA support gate: each active substitution rule must have at least two distinct EVA token types. The scorer independently recomputes assignments, raw coverage, distinct identity count, distinct EVA type count, complexity, and objective components.

Real E2 is not evaluated by this remediation package. The package must pass the small exhaustive tests before a new real-data run is created.
