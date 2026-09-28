# Validation report

`PATH_GRAPH.tsv` was generated from the frozen scope and lexicon using the existing E2 alignment enumerator followed by an independent exact call to `encode_word`. The graph therefore contains only scorer-consistent paths.

Required status interpretation:

- `SUPPORT_VALID_SOLUTION_FOUND=NO`
- `NO_SOLUTION_CERTIFIED_WITHIN_TESTED_THRESHOLDS=NO`
- `SEARCH_INCONCLUSIVE_TIMEOUT=YES`
- `EXACT_PATH_REMEDIATION_FAILED=YES` for the prior witness remediation

These statuses mean that no trustworthy correspondence has been found yet, while the corrected compatibility graph is available for a properly bounded compact search.
