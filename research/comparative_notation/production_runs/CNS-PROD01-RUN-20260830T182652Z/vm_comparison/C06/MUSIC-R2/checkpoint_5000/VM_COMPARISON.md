# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C06` representation_id=`MUSIC-R2` checkpoint=`5000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 16.0617 | 5 |
| T | COMPARABLE | 1866.49 | 10 |
| S | COMPARABLE | 165.671 | 20 |
| L | COMPARABLE | 67.8723 | 13 |
| D | COMPARABLE | 269.422 | 4 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S03_TRANSITION_ENTROPY_NORM | S | 0.253481 |
| D_LINE_PROGRESSION | D | 0.423444 |
| S03_TRANSITION_ENTROPY_NORM | S | 0.576765 |
| G05_TRIGRAM_OCCUPANCY | G | 0.615433 |
| G04_BIGRAM_OCCUPANCY | G | 2.80529 |
| G07_HIGHER_ORDER_ENTROPY_REDUCTION | G | 2.92465 |
| L06_SAME_LINE_COOCCURRENCE_DENSITY | L | 4.77151 |
| L07_SAME_LINE_NONCOOCCURRENCE_DENSITY | L | 4.77151 |
| G06_SYMBOL_CONDITIONAL_ENTROPY_NORM | G | 4.82706 |
| L04_POSITION_PROGRESSION | L | 5.19954 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| T07_EDIT_GRAPH_DENSITY | T | 9627.63 |
| T06_SUFFIX_DIVERSITY | T | 4740.63 |
| T05_PREFIX_DIVERSITY | T | 3959.55 |
| S02_TRANSITION_ZERO_DENSITY | S | 1293.48 |
| S01_TRANSITION_DENSITY | S | 1293.48 |
| D_LINE_COHERENCE | D | 975.203 |
| L06_SAME_LINE_COOCCURRENCE_DENSITY | L | 297.295 |
| L07_SAME_LINE_NONCOOCCURRENCE_DENSITY | L | 297.295 |
| S01_TRANSITION_DENSITY | S | 230.88 |
| S02_TRANSITION_ZERO_DENSITY | S | 230.88 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C06/MUSIC-R2/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 11785 reaches checkpoint 5000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
