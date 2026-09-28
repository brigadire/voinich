# Cross-domain experiment protocol v2

Preparation only; real cross-domain search is explicitly not executed.

The 77 v1 candidates are audited line-by-line in `BOTANICAL_LINE_AUDIT.tsv`; because the v1 evidence contains no concrete line locator, every row is `SOURCE_DESCRIPTION_ONLY` and excluded. `COMMON_IDENTITIES=min(I,94)` and `COMMON_FORMS` are computed from verified rows only. Here both are zero, so no primary panel exists.

When a later curation run supplies verified rows, panels must be built by the frozen minimum-cost matching specification in `MATCHED_PANEL_PARAMETERS.json`, with fixed seed and lexicographic tie-breaking. Sensitivity panels may be added only before search. The 3/27 astronomy result remains historical baseline; matched astronomy search is a future staged action.

Null execution is staged: if botanical maximum is below 4, run no nulls; if it is at least 4, run all 99 frozen seeds without changing the replicate count.
