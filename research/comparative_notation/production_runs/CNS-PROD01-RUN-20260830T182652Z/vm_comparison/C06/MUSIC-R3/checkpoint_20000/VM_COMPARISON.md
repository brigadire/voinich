# CORPUS VS VM STRUCTURAL REPORT

candidate_id=`C06` representation_id=`MUSIC-R3` checkpoint=`20000` reachable=`false`

| Family | Comparable | Mean distance | Comparable metrics |
|---|---|---:|---:|
| G | NOT_COMPARABLE |  | 0 |
| T | NOT_COMPARABLE |  | 0 |
| S | NOT_COMPARABLE |  | 0 |
| L | NOT_COMPARABLE |  | 0 |
| D | NOT_COMPARABLE |  | 0 |

## Strongest similarities (smallest calibrated distance)

(no comparable metrics at this checkpoint)

## Strongest differences (largest calibrated distance)

(no comparable metrics at this checkpoint)

## Shared corpus rules / VM rules not reproduced / Candidate rules absent in VM

NOT_COMPUTED: the generic USC/Fingerprint schema retains only aggregate metric values, not symbol- or rule-level detail, so per-rule reproduction cannot be derived without new, out-of-scope code.

## Corpus-size sensitivity

See `rarefaction/C06/MUSIC-R3/RAREFACTION_SUMMARY.tsv` for this candidate/representation's own metric estimates at every frozen checkpoint.

## Comparability limitations

Corpus size 12475 is below checkpoint 20000; every metric is NOT_COMPARABLE at this checkpoint by frozen protocol rule.

## Result

`STRUCTURALLY_CLOSE_ON: PENDING`  
`STRUCTURALLY_DISTANT_ON: PENDING`  
`NOT_COMPARABLE_ON: PENDING`

No frozen numeric threshold for STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON exists anywhere in the frozen protocol (VM_COMPARISON_TEMPLATE.md itself marks this PENDING); inventing one here would be an undocumented post-hoc parameter, so it stays PENDING for a separate task. This report contains no interpretation of what these distances mean about the Voynich Manuscript.
