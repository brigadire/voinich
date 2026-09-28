# Metric interpretation

`score` is the integer selection statistic `100 * matched - complexity`.
`coverage` is matched occurrences divided by the number of evaluated occurrences.
For directional rows this is held-out coverage; training fit is not substituted.
`total_assignment_cost` counts unmatched occurrences. Mean/median pair score are
NA for an empty assignment, not fabricated zero-quality pairs.

The `term_collisions` column in SYSTEM_RESULTS.tsv and endpoint files counts
excess candidate identities at the evaluated LABEL tokens. It is a target-adjacency
ambiguity count. DICTIONARY_COLLISIONS.tsv separately reports all transformed
dictionary collisions, including outputs never observed in a LABEL: number of
colliding output forms, and excess canonical identities across those outputs.
This additional diagnostic changes neither the score nor selection.

All equal edges have pair score 1 and tied rank 1. A local margin of 1 means only
one identity has an equality edge at that token; it does not imply a uniquely
optimal global assignment. The assignment margin forbids the edge and solves
maximum matching again. Candidate publication requires the stricter global margin.
Empty pair tables mean that no pairs meet the represented condition, not a failed
export. A zero-score selected system is only the deterministic tie-breaking output.

The null comparison unit is an entire selected system/dataset/split. Null replicas
reuse a frozen source pool; 10,000 algorithmic draws do not create 10,000 independent
historical corpora. The circular-text control has 29 observed occurrences and uses
sampling with replacement; its source-uniqueness statistic is reported explicitly.
Invariance controls have exact p=1 on a singleton quotient space and are diagnostic.
