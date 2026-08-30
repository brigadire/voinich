# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C01` representation_id=`LATIN-EXPANDED` checkpoint=`5000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 4.43147 | 5 |
| T | COMPARABLE | 13.3077 | 10 |
| S | COMPARABLE | 29.0533 | 20 |
| L | NOT_COMPARABLE |  | 0 |
| D | NOT_COMPARABLE |  | 0 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S02_TRANSITION_ZERO_DENSITY | S | 0.133448 |
| S01_TRANSITION_DENSITY | S | 0.133448 |
| S01_TRANSITION_DENSITY | S | 0.47742 |
| S02_TRANSITION_ZERO_DENSITY | S | 0.47742 |
| G05_TRIGRAM_OCCUPANCY | G | 0.857658 |
| G04_BIGRAM_OCCUPANCY | G | 1.40714 |
| T04_HAPAX_RATIO | T | 1.46489 |
| T06_SUFFIX_DIVERSITY | T | 1.56136 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 1.98057 |
| S03_TRANSITION_ENTROPY_NORM | S | 2.6599 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S01_TRANSITION_DENSITY | S | 112.793 |
| S02_TRANSITION_ZERO_DENSITY | S | 112.793 |
| S02_TRANSITION_ZERO_DENSITY | S | 86.5572 |
| S01_TRANSITION_DENSITY | S | 86.5572 |
| S02_TRANSITION_ZERO_DENSITY | S | 49.6941 |
| S01_TRANSITION_DENSITY | S | 49.6941 |
| T08_EDIT_GRAPH_GIANT_SHARE | T | 36.595 |
| T02_TOKEN_LENGTH_SD | T | 33.9195 |
| T10_EDIT_DEGREE_FREQUENCY_SPEARMAN | T | 26.6005 |
| S03_TRANSITION_ENTROPY_NORM | S | 14.3642 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C01/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 19132 reaches checkpoint 5000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
