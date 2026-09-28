# Reproducibility

```text
/home/brigadire/.venv/bin/python3 oracle_validation.py
/home/brigadire/.venv/bin/python3 synthetic_convergence.py
/home/brigadire/.venv/bin/python3 run_seed.py 1 30
/home/brigadire/.venv/bin/python3 run_seed.py 2 30
/home/brigadire/.venv/bin/python3 run_seed.py 3 30
/home/brigadire/.venv/bin/python3 aggregate.py
```

The frozen target scope, lexicon, scorer, and structural potential-support index are referenced by content hashes in `SNAPSHOT_MANIFEST.json`.
