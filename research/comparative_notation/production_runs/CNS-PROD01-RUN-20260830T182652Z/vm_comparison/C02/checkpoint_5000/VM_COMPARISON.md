# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C02` representation_id=`LATIN-DIPLOMATIC` checkpoint=`5000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 2.43527 | 5 |
| T | COMPARABLE | 13.8154 | 10 |
| S | COMPARABLE | 29.7564 | 20 |
| L | NOT_COMPARABLE |  | 0 |
| D | NOT_COMPARABLE |  | 0 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S02_TRANSITION_ZERO_DENSITY | S | 0.179757 |
| S01_TRANSITION_DENSITY | S | 0.179757 |
| G05_TRIGRAM_OCCUPANCY | G | 0.345838 |
| S01_TRANSITION_DENSITY | S | 0.493463 |
| S02_TRANSITION_ZERO_DENSITY | S | 0.493463 |
| G04_BIGRAM_OCCUPANCY | G | 0.961348 |
| T04_HAPAX_RATIO | T | 1.2264 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 2.01556 |
| G06_SYMBOL_CONDITIONAL_ENTROPY_NORM | G | 2.03759 |
| S03_TRANSITION_ENTROPY_NORM | S | 2.72908 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S01_TRANSITION_DENSITY | S | 115.312 |
| S02_TRANSITION_ZERO_DENSITY | S | 115.312 |
| S01_TRANSITION_DENSITY | S | 88.8549 |
| S02_TRANSITION_ZERO_DENSITY | S | 88.8549 |
| S02_TRANSITION_ZERO_DENSITY | S | 51.4994 |
| S01_TRANSITION_DENSITY | S | 51.4994 |
| T08_EDIT_GRAPH_GIANT_SHARE | T | 35.7704 |
| T02_TOKEN_LENGTH_SD | T | 29.8073 |
| T10_EDIT_DEGREE_FREQUENCY_SPEARMAN | T | 23.5839 |
| S03_TRANSITION_ENTROPY_NORM | S | 14.4795 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C02/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 19121 reaches checkpoint 5000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
