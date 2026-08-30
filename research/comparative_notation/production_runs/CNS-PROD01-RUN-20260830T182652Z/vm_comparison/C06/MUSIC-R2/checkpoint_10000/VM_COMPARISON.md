# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C06` representation_id=`MUSIC-R2` checkpoint=`10000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 23.572 | 5 |
| T | COMPARABLE | 3639.76 | 10 |
| S | COMPARABLE | 114.294 | 20 |
| L | COMPARABLE | 55.334 | 13 |
| D | COMPARABLE | 320.818 | 4 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S03_TRANSITION_ENTROPY_NORM | S | 0.48292 |
| G05_TRIGRAM_OCCUPANCY | G | 0.604751 |
| S03_TRANSITION_ENTROPY_NORM | S | 0.641171 |
| D_LINE_PROGRESSION | D | 1.14646 |
| G07_HIGHER_ORDER_ENTROPY_REDUCTION | G | 3.16048 |
| S02_TRANSITION_ZERO_DENSITY | S | 3.34421 |
| S01_TRANSITION_DENSITY | S | 3.34421 |
| G04_BIGRAM_OCCUPANCY | G | 3.66497 |
| G06_SYMBOL_CONDITIONAL_ENTROPY_NORM | G | 4.38435 |
| L04_POSITION_PROGRESSION | L | 7.44631 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| T07_EDIT_GRAPH_DENSITY | T | 21485.2 |
| T06_SUFFIX_DIVERSITY | T | 7442.91 |
| T05_PREFIX_DIVERSITY | T | 7059.98 |
| D_LINE_COHERENCE | D | 1157.53 |
| S01_TRANSITION_DENSITY | S | 749.227 |
| S02_TRANSITION_ZERO_DENSITY | S | 749.227 |
| S02_TRANSITION_ZERO_DENSITY | S | 168.598 |
| S01_TRANSITION_DENSITY | S | 168.598 |
| L03_BOUNDARY_SPECIALIZATION | L | 142.514 |
| L07_SAME_LINE_NONCOOCCURRENCE_DENSITY | L | 140.044 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C06/MUSIC-R2/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 11785 reaches checkpoint 10000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
