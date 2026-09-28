# Reproducibility

Run:

```text
/home/brigadire/.venv/bin/python3 matching_oracle_tests.py
/home/brigadire/.venv/bin/python3 profile_monolith.py
/home/brigadire/.venv/bin/python3 integration_decomposition.py
/home/brigadire/.venv/bin/python3 run_decomposition.py
```

The package records its copied scorer in `snapshot/scorer.py`, generated profiles, test outputs, checkpoint, configuration, and SHA-256 manifest. Input scope and lexicon hashes are retained in the parent frozen packages; no real-data-dependent pruning is used.
