# Validation

- Verified frozen TARGET_SETS.tsv, D1 lexicon and corpus-metadata SHA256 before preparation.
- Reconstructed 57 unique occurrence IDs, 30/27 pages and 26/31 hapax strata from the full corpus.
- All frozen experiment inputs and executable bytes verified on run/resume and completion.
- 900 atomic hashed checkpoint shards: 90,000 primary and 30,000 independent-capacity datasets.
- Every primary family/end-point has exactly 10,000 rows; 27 multiplicity-corrected comparisons.
- Exact invariance checks passed for within-page order permutation and identity-group bijection.
- 57 complete LOO searches and 200 page-stratified subsamples.
- All upstream snapshot bytes unchanged. Pre-existing enrichment checksum discrepancies documented.
- Automated tests and independent replay checks: see VERIFICATION.json and TEST_RESULTS.txt.

Synthetic sensitivity is separately gated and must be read from SUMMARY.json;
completed computation is not evidence that all sensitivity or signal gates passed.
