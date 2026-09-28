# Reproducibility

Run from this directory:

```text
/home/brigadire/.venv/bin/python3 diagnostic_seed.py 1 30
/home/brigadire/.venv/bin/python3 diagnostic_seed.py 2 30
/home/brigadire/.venv/bin/python3 diagnostic_seed.py 3 30
/home/brigadire/.venv/bin/python3 aggregate_diagnostic.py
```

Each threshold run uses deterministic seed state and the same 30-second budget. Fixed-table matching is deterministic and assignments are retained in the per-table JSON records before aggregation.
