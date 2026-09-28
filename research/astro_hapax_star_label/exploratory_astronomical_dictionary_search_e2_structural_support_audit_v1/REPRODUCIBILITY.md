# Reproducibility

Run from this directory:

```text
/home/brigadire/.venv/bin/python3 audit.py
```

The audit enumerates all local DROP_UNMAPPED alignments for the frozen 57 labels and 299-row lexicon through k=5, builds the potential-support and co-occurrence indices, computes union-graph matching upper bounds, samples a bounded set of exact fixed-table assignments, and reconstructs the prior 3/57 diagnostic assignments.
