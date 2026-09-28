# E2 real run v3 protocol

This is a new versioned run after the legacy runner was invalidated. It uses the content-bound remediated CP-SAT runner, k=5, `GLOBAL_CAPACITY_1`, `BALANCED`, and a 900-second single-worker budget.

The target scope and lexicon are the frozen E2 inputs. Distinct EVA token support is enforced inside the model: every active source-to-target rule requires at least two distinct EVA token types. The prior E1 table is used only as a solver hint after the model's support gate is applied; it is not accepted as a result unless it is feasible under that gate.

Every incumbent must contain its complete assignment. Its objective is checked against the captured assignment, each selected edge is re-encoded independently, identities and labels are checked for uniqueness, and support is recomputed from distinct EVA token types. The checkpoint is written atomically after each improvement.

The run is exploratory and produces no real-data scientific claim. The previous 900-second package remains immutable and invalid for interpretation.
