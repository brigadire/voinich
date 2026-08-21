# Task52 comparison metric audit

The comparison layer reads frozen artifacts only. A file is not included in
the fingerprint merely because it exists; the metric must have the same
scientific semantics in generic and IVTFF corpora.

| family/stage | artifact candidates | comparable quantity | normalization/support | fingerprint decision |
|---|---|---|---|---|
| corpus/token statistics | `sequence_analysis.yaml`, transition summary | token count, vocabulary, eligible tokens | rates retain token count denominator | raw/support; eligible rate included |
| sequence/repetition | `replicated_local_structure_summary.tsv` | frozen, significant, replicated sequence candidates | significant/frozen and replicated/significant | rates included; counts support-only |
| token relations | `relation_classification.tsv` | tested/significant/replicated generic relation rows | tested/significant denominators | rates included; metadata labels excluded |
| higher order | `higher_order_validation.tsv` | candidates and `HIGHER_ORDER_REPLICATED` | replicated/candidate | rate included; statuses remain descriptive |
| transition network | `transition_network_summary.yaml` | significant and strict backbone counts, profile counts, log-loss | explicit retention denominators | corrected retention rates included |
| structural/profile/reliability | structural, profile stability, reliability, soft-space outputs | corpus-specific candidate/effect tables | semantics are heterogeneous or support-dependent | raw audit only; excluded from fixed v2 core |
| regimes/context/trajectory/projection | local/global/conditional/distance/property/projection outputs | often window-, metadata-, or candidate-specific | no single cross-corpus denominator established | excluded; metadata labels are incompatible |
| begin/end and normalization comparison | candidate and normalization reports | hypothesis-specific candidate lists | no stable generic common quantity in current contract | excluded from fingerprint |
| vocabulary growth (Task49) | `vocabulary-growth/*` | V(n), Heaps coefficients, hapax fractions, null effects, segment summaries | common checkpoints; V(n)/n only at fixed n | optional family; missing legacy artifacts remain missing |
| structural family cardinality (Task52/53) | structural family diagnostics | family count, largest-family token/edge counts, total family edges | raw support for interpreting structural explosion | diagnostic only; excluded from aggregate distance |

## Transition semantics

`transition.preferred_backbone_retention` is
`backbone_preferred / fdr_significant_preferred`; the depleted equivalent uses
the depleted counts. These are retention measures, not significant-edge rates.
The combined `transition.backbone_retention` uses total strict backbone divided
by total significant edges. No preferred/depleted significance rate is emitted
because the current artifact does not expose scientifically valid separate
tested-preferred and tested-depleted denominators.

## Redundancy and distance eligibility

Raw counts, denominators and support values remain in `raw_metrics.tsv` and
`metric_support.tsv`, but are excluded from the distance vector when they are
strongly corpus-size dependent or support-only. Fixed proportions, coefficients
and standardized effects have equal feature weight. Family-level distances use
the same rule within each semantic family. P-values and categorical metadata
are never distance features.

Every fingerprint feature is recorded in machine-readable
`feature_definitions` with family, formula/version and `use_for_distance`.
