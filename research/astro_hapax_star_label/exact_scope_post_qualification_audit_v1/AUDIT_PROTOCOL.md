# Post-qualification audit protocol

The v2 qualification record is immutable input. This audit enumerates the exact-scope table space independently with the frozen scorer, records every co-optimal table count and hash, checks whether ground truth is optimal, and classifies each case as `GROUND_TRUTH_NOT_OPTIMAL`, `MULTIPLE_OPTIMA`, or `UNIQUE_CORRECT_OPTIMUM`.

No solver, objective, threshold, metric, seed, or v2 file is changed. The audit is diagnostic and cannot authorize real-data search.
