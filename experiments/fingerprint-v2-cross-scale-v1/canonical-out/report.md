# Fingerprint v2 lexical-paradigm report

- Version: `fingerprint-v2-lexical-paradigms-v1`
- Commit: `d06bed5b6e2464ebbf21f9dc54969d6ad8dec8cb`
- Seed/repetitions: `20260823` / `100`
- Primary corpus: `voynich-zl3b-eva`; tokens `39380`; types `8243`; SHA-256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`
- LP1 support Gini: `0.789799`; top-rule share: `0.00868989`
- EF1 edges/isolates/giant share: `41485` / `948` / `0.866917`
- EF2 clustering/triangles/3-paths/4-cycles: `0.28727` / `74101` / `773846` / `616670`
- EF3 Spearman(degree, log frequency): `0.639007`
- EF5 same-line/same-page/same-regime rate: `0` / n/a / n/a

## Verdicts

- **C_GRAMMAR_VALIDATION:** `PARTIALLY_SUPPORTED` — At least one positional or bigram total-variation diagnostic exceeded diagnostic_tolerance.
- **EDIT_NEIGHBORHOODS_EXCEED_GRAMMAR_NULL:** `NOT_SUPPORTED` — No EF graph statistic exceeds the C-GRAMMAR distribution after the declared EF FDR correction.
- **DIRECTIONAL_TRANSFORMATIONS_SUPPORTED:** `NOT_SUPPORTED` — LP2 C-GRAMMAR/C-LEN/C-FREQ support-concentration tests with lexical-family FDR.
- **CONTEXT_CONDITIONING_SUPPORTED:** `INCONCLUSIVE` — LP3 same-line family-occurrence locality under C-GLOBAL.
- **PARADIGM_PRODUCTIVITY_SUPPORTED:** `NOT_SUPPORTED` — LP2 concentration plus declared support threshold, with LP3 restricted to selected rules.
- **LEXICAL_PARADIGM_BLOCK_READY:** `INCONCLUSIVE` — LP1-LP4 and EF1-EF4 are computed with deterministic diagnostics and declared nulls.

## Task77 cross-scale verdicts

- **TASK75_RESULTS_REPRODUCED:** `PARTIALLY_SUPPORTED` — code re-derivation + determinism re-run (TestSeededPipelineIsDeterministic) + first real-corpus execution
- **EDIT_FAMILIES_STRUCTURALLY_STABLE:** `NOT_SUPPORTED` — consensus_status=INSUFFICIENT_SUPPORT across 5 testable perturbations
- **EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL:** `NOT_SUPPORTED` — EF1 giant-component share, EF2 clustering, EF3 |Spearman(degree, log-frequency)| vs C-GRAMMAR
- **TRANSFORMATION_MOTIFS_STABLE:** `INCONCLUSIVE` — components-vs-label-propagation ARI=0, NMI=0
- **FAMILY_LINE_POSITION_DEPENDENCE:** `INCONCLUSIVE` — P(edit family | line position) != P(edit family)
- **TRANSFORMATION_CONTEXT_DEPENDENCE:** `INCONCLUSIVE` — P(current family | previous-token family) != P(current family)
- **FAMILY_LOCUS_DEPENDENCE:** `NOT_SUPPORTED` — P(edit family | locus type) != P(edit family)
- **STRUCTURAL_DISTANCE_EDIT_DISTANCE_DEPENDENCE:** `NOT_SUPPORTED` — raw glyph edit distance between two vocabulary types correlates with the minimum distance (in lines) between their occurrences, beyond a frequency-bin-matched permutation
- **FAMILY_FOLIO_REGIME_DEPENDENCE:** `NOT_APPLICABLE` — cs4/family-currier: status=NOT_APPLICABLE p=0 q=0; cs4/family-section: status=NOT_APPLICABLE p=0 q=0
- **CROSS_SCALE_EFFECTS_SURVIVE_CONDITIONING:** `INCONCLUSIVE` — CS1 family/line-position effect re-tested within Currier A, Currier B and TEXT-locus strata (each against its own within-line null)
- **CROSS_SCALE_EFFECTS_GENERALIZE:** `INCONCLUSIVE` — out-of-fold log-loss improvement of family-label features over a folio-marginal baseline for predicting line-position class
- **EDIT_CROSS_SCALE_BLOCK_READY:** `PARTIALLY_SUPPORTED` — 5/8 cross-scale metrics computed (not NOT_APPLICABLE); consensus_status=INSUFFICIENT_SUPPORT

Raw null distributions, per-replicate grammar diagnostics, and input checksums are in `raw_results.json`. These values describe only the configured input corpus; this report does not identify it as a canonical Voynich run.
