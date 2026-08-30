# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C06` representation_id=`MUSIC-R3` checkpoint=`10000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 21.2931 | 5 |
| T | COMPARABLE | 445.471 | 10 |
| S | COMPARABLE | 19.778 | 20 |
| L | COMPARABLE | 22.8527 | 13 |
| D | COMPARABLE | 173.804 | 4 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| G06_SYMBOL_CONDITIONAL_ENTROPY_NORM | G | 0.470228 |
| G05_TRIGRAM_OCCUPANCY | G | 0.604751 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 1.09666 |
| D_LINE_PROGRESSION | D | 1.14646 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 1.62735 |
| G04_BIGRAM_OCCUPANCY | G | 2.26025 |
| S03_TRANSITION_ENTROPY_NORM | S | 2.48551 |
| S03_TRANSITION_ENTROPY_NORM | S | 3.98493 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 4.56335 |
| S03_TRANSITION_ENTROPY_NORM | S | 4.93609 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| T07_EDIT_GRAPH_DENSITY | T | 2733.76 |
| T05_PREFIX_DIVERSITY | T | 956.043 |
| D_LINE_COHERENCE | D | 580.058 |
| T06_SUFFIX_DIVERSITY | T | 449.649 |
| T02_TOKEN_LENGTH_SD | T | 102.668 |
| T01_MEAN_TOKEN_LENGTH | T | 79.1732 |
| G08_MEAN_SYMBOLS_PER_TOKEN | G | 79.1732 |
| T04_HAPAX_RATIO | T | 68.6111 |
| S02_TRANSITION_ZERO_DENSITY | S | 65.9229 |
| S01_TRANSITION_DENSITY | S | 65.9229 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C06/MUSIC-R3/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 12475 reaches checkpoint 10000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
