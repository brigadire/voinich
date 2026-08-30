# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C01` representation_id=`LATIN-EXPANDED` checkpoint=`10000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 5.72469 | 5 |
| T | COMPARABLE | 18.4194 | 10 |
| S | COMPARABLE | 19.0476 | 20 |
| L | NOT_COMPARABLE |  | 0 |
| D | NOT_COMPARABLE |  | 0 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S02_TRANSITION_ZERO_DENSITY | S | 0.229372 |
| S01_TRANSITION_DENSITY | S | 0.229372 |
| G05_TRIGRAM_OCCUPANCY | G | 0.842771 |
| T04_HAPAX_RATIO | T | 1.55119 |
| S01_TRANSITION_DENSITY | S | 1.67598 |
| S02_TRANSITION_ZERO_DENSITY | S | 1.67598 |
| G04_BIGRAM_OCCUPANCY | G | 1.83836 |
| T06_SUFFIX_DIVERSITY | T | 2.45138 |
| G06_SYMBOL_CONDITIONAL_ENTROPY_NORM | G | 2.59512 |
| S03_TRANSITION_ENTROPY_NORM | S | 2.95693 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S01_TRANSITION_DENSITY | S | 65.3335 |
| S02_TRANSITION_ZERO_DENSITY | S | 65.3335 |
| T02_TOKEN_LENGTH_SD | T | 62.4055 |
| T08_EDIT_GRAPH_GIANT_SHARE | T | 51.0612 |
| S02_TRANSITION_ZERO_DENSITY | S | 48.6216 |
| S01_TRANSITION_DENSITY | S | 48.6216 |
| S02_TRANSITION_ZERO_DENSITY | S | 28.3966 |
| S01_TRANSITION_DENSITY | S | 28.3966 |
| T10_EDIT_DEGREE_FREQUENCY_SPEARMAN | T | 18.1305 |
| G08_MEAN_SYMBOLS_PER_TOKEN | G | 16.7123 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C01/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 19132 reaches checkpoint 10000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
