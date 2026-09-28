# Threshold-null generator remediation

The new runner processes one null dictionary and one operation profile at a time. It never constructs a 64-profile union graph, preserves the frozen 64-profile model-selection OR, and releases profile paths and CP-SAT objects between profiles.

The first full null replica did not complete in the available runtime. The bottleneck is scorer-consistent path generation/alignment before CP-SAT; no null verdict was assigned. Stage A was therefore not started.

The runner is available as:

```text
/home/brigadire/.venv/bin/python3 -u run_remediation.py
```

No p-value, `NULL_GE_5` frequency, or scientific interpretation is reported.
