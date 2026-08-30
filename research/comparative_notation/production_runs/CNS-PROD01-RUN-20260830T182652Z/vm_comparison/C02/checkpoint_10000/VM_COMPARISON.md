# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C02` representation_id=`LATIN-DIPLOMATIC` checkpoint=`10000` reachable=`true`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | COMPARABLE | 2.87253 | 5 |
| T | COMPARABLE | 19.3897 | 10 |
| S | COMPARABLE | 19.4783 | 20 |
| L | NOT_COMPARABLE |  | 0 |
| D | NOT_COMPARABLE |  | 0 |

## Strongest similarities (smallest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S02_TRANSITION_ZERO_DENSITY | S | 0.308969 |
| S01_TRANSITION_DENSITY | S | 0.308969 |
| G05_TRIGRAM_OCCUPANCY | G | 0.339835 |
| G04_BIGRAM_OCCUPANCY | G | 1.25595 |
| T04_HAPAX_RATIO | T | 1.29865 |
| S01_TRANSITION_DENSITY | S | 1.7323 |
| S02_TRANSITION_ZERO_DENSITY | S | 1.7323 |
| G06_SYMBOL_CONDITIONAL_ENTROPY_NORM | G | 1.85072 |
| S03_TRANSITION_ENTROPY_NORM | S | 3.03384 |
| S08_HIGHER_ORDER_PREDICTIVE_GAIN | S | 3.13258 |

## Strongest differences (largest calibrated distance)

| metric_id | family | distance |
|---|---|---:|
| S01_TRANSITION_DENSITY | S | 66.793 |
| S02_TRANSITION_ZERO_DENSITY | S | 66.793 |
| T02_TOKEN_LENGTH_SD | T | 54.8397 |
| S01_TRANSITION_DENSITY | S | 49.9123 |
| S02_TRANSITION_ZERO_DENSITY | S | 49.9123 |
| T08_EDIT_GRAPH_GIANT_SHARE | T | 49.9106 |
| S02_TRANSITION_ZERO_DENSITY | S | 29.4282 |
| S01_TRANSITION_DENSITY | S | 29.4282 |
| T05_PREFIX_DIVERSITY | T | 21.8926 |
| T06_SUFFIX_DIVERSITY | T | 19.0824 |

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C02/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 19121 reaches checkpoint 10000.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
