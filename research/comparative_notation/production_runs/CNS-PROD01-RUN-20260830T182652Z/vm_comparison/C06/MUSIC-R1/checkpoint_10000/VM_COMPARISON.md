# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C06` representation_id=`MUSIC-R1` checkpoint=`10000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 6.1029 | 5 |
| T | COMPARABLE | 132.404 | 10 |
| S | COMPARABLE | 8.59901 | 20 |
| L | COMPARABLE | 27.0739 | 13 |
| D | COMPARABLE | 143.382 | 4 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| G06_SYMBOL_CONDITIONAL_ENTROPY_NORM | G | 0.1959 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 0.432415 |
| G05_TRIGRAM_OCCUPANCY | G | 0.463137 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 0.99137 |
| S03_TRANSITION_ENTROPY_NORM | S | 1.52751 |
| L07_SAME_LINE_NONCOOCCURRENCE_DENSITY | L | 1.82031 |
| L06_SAME_LINE_COOCCURRENCE_DENSITY | L | 1.82031 |
| D_LINE_PROGRESSION | D | 2.45816 |
| G04_BIGRAM_OCCUPANCY | G | 2.51415 |
| S03_TRANSITION_ENTROPY_NORM | S | 2.54736 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| T07_EDIT_GRAPH_DENSITY | T | 772.232 |
| D_LINE_COHERENCE | D | 471.079 |
| T06_SUFFIX_DIVERSITY | T | 339.068 |
| L02_LINE_SYMBOL_COUNT_MEAN | L | 197.449 |
| T04_HAPAX_RATIO | T | 70.5469 |
| T02_TOKEN_LENGTH_SD | T | 70.0354 |
| D_LINE_EXCLUSIVITY | D | 65.7198 |
| L06_SAME_LINE_COOCCURRENCE_DENSITY | L | 34.4206 |
| L07_SAME_LINE_NONCOOCCURRENCE_DENSITY | L | 34.4206 |
| D_LINE_VARIANCE_SHARE | D | 34.2697 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C06/MUSIC-R1/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 13117 reaches checkpoint 10000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
