# Metric compatibility audit (Task45)

`experiment-compare` reads frozen artifacts only. It never invokes a pipeline
stage and it does not use experiment names or corpus labels as features.

| stage | artifact | raw metric | semantic meaning | comparable across corpus modes? | normalization | missing-value semantics |
|---|---|---|---|---|---|---|
| corpus | `sequence_analysis.yaml`, `transition_network_summary.yaml` | `token_count`, `unique_tokens` | token occurrences and vocabulary | yes, when the same tokenization contract is recorded | `eligible_tokens / token_count` | `MISSING_ARTIFACT` / `NOT_COMPUTED` |
| sequences | `replicated_local_structure_summary.tsv` | frozen, FDR-significant, robust cross-block sequence candidates | candidate/replication counts | yes for generic sequence rows | significant/frozen; replicated/significant | zero denominator is `NOT_COMPUTED` |
| relations | `relation_classification.tsv` | tested and generic-consistent relation rows | directional relation candidate counts | yes only for numeric rows; categorical labels are not compared | significant/tested | incompatible metadata labels excluded |
| higher order | `higher_order_validation.tsv` | candidate count, `HIGHER_ORDER_REPLICATED` count | higher-order sequence status counts | yes; statuses are this stage's generic structural statuses | replicated/candidate | absent artifact is `MISSING_ARTIFACT` |
| transition network | `transition_network_summary.yaml` | eligible tokens, tested/significant edges, backbone counts, profile counts, log-loss deltas | transition graph structure and predictive comparison | yes for numeric summary fields | backbone/significant and profile/eligible rates | zero denominator is `NOT_COMPUTED` |
| stages 23–27 metadata | stage TSV/YAML categorical columns | `Currier`, `Hand`, `GROUP_*`, metadata classes | different scientific meanings | **no** | none | `INCOMPATIBLE_METRIC` |

The v1 fingerprint contains only the normalized numeric dimensions produced by
the extractor. Raw counts remain available in `raw_metrics.tsv`; they are not
used as the principal cross-corpus fingerprint. No post-hoc feature selection,
classification, semantic interpretation, or document-class conclusion is
implemented.

## Formula registry

The machine-readable formula version is `1` and is emitted in the comparison
manifest and fingerprint YAML. The fixed formulas are:

```text
corpus.eligible_token_rate       = eligible_tokens / token_count
sequence.significant_rate        = significant_candidates / frozen_candidates
sequence.replication_rate        = replicated_candidates / significant_candidates
relation.significant_rate        = significant_relations / tested_relations
transition.preferred_rate        = preferred_significant / significant_preferred
transition.depleted_rate         = depleted_significant / significant_depleted
transition.backbone_retention    = strict_backbone / significant_backbone
transition.outgoing_profile_rate = replicated_outgoing_profiles / eligible_tokens
transition.incoming_profile_rate = replicated_incoming_profiles / eligible_tokens
higher_order.replication_rate    = higher_order_replicated / candidate_count
```

`0`, `NA`, `NOT_APPLICABLE`, `NOT_COMPUTED`, `MISSING_ARTIFACT`, and
`INCOMPATIBLE_METRIC` are distinct states. The current TSV representation
stores an empty value plus the explicit status column for non-numeric states.
