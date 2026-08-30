# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C06` representation_id=`MUSIC-R1` checkpoint=`5000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 4.26034 | 5 |
| T | COMPARABLE | 72.3235 | 10 |
| S | COMPARABLE | 8.49816 | 20 |
| L | COMPARABLE | 19.2851 | 13 |
| D | COMPARABLE | 120.326 | 4 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| G06_SYMBOL_CONDITIONAL_ENTROPY_NORM | G | 0.215681 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 0.346083 |
| G05_TRIGRAM_OCCUPANCY | G | 0.471318 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 0.637865 |
| D_LINE_PROGRESSION | D | 0.907921 |
| S03_TRANSITION_ENTROPY_NORM | S | 1.37407 |
| G04_BIGRAM_OCCUPANCY | G | 1.92441 |
| S03_TRANSITION_ENTROPY_NORM | S | 2.00183 |
| L06_SAME_LINE_COOCCURRENCE_DENSITY | L | 2.56078 |
| L07_SAME_LINE_NONCOOCCURRENCE_DENSITY | L | 2.56078 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| D_LINE_COHERENCE | D | 396.879 |
| T07_EDIT_GRAPH_DENSITY | T | 346.042 |
| T06_SUFFIX_DIVERSITY | T | 215.964 |
| L02_LINE_SYMBOL_COUNT_MEAN | L | 128.726 |
| T04_HAPAX_RATIO | T | 66.6219 |
| D_LINE_EXCLUSIVITY | D | 61.6424 |
| S02_TRANSITION_ZERO_DENSITY | S | 42.8265 |
| S01_TRANSITION_DENSITY | S | 42.8265 |
| T02_TOKEN_LENGTH_SD | T | 38.0666 |
| L07_SAME_LINE_NONCOOCCURRENCE_DENSITY | L | 29.1176 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C06/MUSIC-R1/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 13117 reaches checkpoint 5000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
