# Reproducibility

The package is content-addressed through `SHA256SUMS` and `SNAPSHOT_MANIFEST.json`. It includes the target scope, lexicon, exclusions, protocol, scoring, search stages, synthetic tests, and preflight results. The v2 CP-SAT snapshot is copied as an immutable implementation dependency. Each future raw row must carry solver, scorer, config, lexicon, target-scope, dependency, seed, stage, and execution metadata.
