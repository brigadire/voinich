# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C06` representation_id=`MUSIC-R3` checkpoint=`5000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 15.3299 | 5 |
| T | COMPARABLE | 228.839 | 10 |
| S | COMPARABLE | 22.3461 | 20 |
| L | COMPARABLE | 21.5381 | 13 |
| D | COMPARABLE | 145.312 | 4 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| D_LINE_PROGRESSION | D | 0.423444 |
| G06_SYMBOL_CONDITIONAL_ENTROPY_NORM | G | 0.517709 |
| G05_TRIGRAM_OCCUPANCY | G | 0.615433 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 0.877711 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 1.04707 |
| G04_BIGRAM_OCCUPANCY | G | 1.73007 |
| S03_TRANSITION_ENTROPY_NORM | S | 2.23584 |
| L06_SAME_LINE_COOCCURRENCE_DENSITY | L | 3.10987 |
| L07_SAME_LINE_NONCOOCCURRENCE_DENSITY | L | 3.10987 |
| S03_TRANSITION_ENTROPY_NORM | S | 3.13154 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| T07_EDIT_GRAPH_DENSITY | T | 1225.02 |
| T05_PREFIX_DIVERSITY | T | 536.192 |
| D_LINE_COHERENCE | D | 488.693 |
| T06_SUFFIX_DIVERSITY | T | 286.397 |
| S02_TRANSITION_ZERO_DENSITY | S | 113.81 |
| S01_TRANSITION_DENSITY | S | 113.81 |
| T04_HAPAX_RATIO | T | 64.7938 |
| D_LINE_EXCLUSIVITY | D | 60.589 |
| T02_TOKEN_LENGTH_SD | T | 55.8038 |
| G08_MEAN_SYMBOLS_PER_TOKEN | G | 51.6168 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C06/MUSIC-R3/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 12475 reaches checkpoint 5000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
