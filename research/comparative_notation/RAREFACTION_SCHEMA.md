# Rarefaction output schemas v1 (B03)

## RAREFACTION.tsv (per corpus/representation, `*_RAREFACTION_V2_RAW.tsv` for VM)

One row per (metric, checkpoint, replicate).

| column | meaning |
|---|---|
| `corpus_id` | corpus identifier |
| `representation_id` | representation identifier |
| `family` | `G`\|`T`\|`S`\|`L`\|`D`\|`CURVE` |
| `metric_id` | metric identifier (or `A2_BIGRAM_TYPES`/`A3_TRIGRAM_TYPES`/`AT_TRANSITION_TYPES` for `CURVE`) |
| `regime` | support regime, empty when not support-stratified |
| `checkpoint_requested` | the frozen target token count |
| `checkpoint_actual` | tokens actually drawn (boundary-preserving) |
| `replicate` | replicate index, `-1` marks a `NOT_COMPARABLE` checkpoint (corpus below target) |
| `seed` | the exact `SeedFor(...)` seed used for this row's draw |
| `value` | metric value if comparable, empty otherwise |
| `comparable` | `true`/`false` |

## RAREFACTION_SUMMARY.tsv (per corpus/representation, `VM_RAREFACTION_V2.tsv` for VM)

One row per (metric, checkpoint), aggregating every comparable replicate.

| column | meaning |
|---|---|
| `corpus_id`, `representation_id`, `family`, `metric_id`, `regime` | as above |
| `checkpoint` | the frozen target token count |
| `mean`, `median`, `sd` | across `n_valid` comparable replicates |
| `ci_low`, `ci_high` | 2.5th/97.5th percentile of the replicate distribution (linear interpolation) |
| `n_valid` | number of comparable replicates (0 when the metric is not comparable at any replicate) |

Both schemas are produced by `notation.WriteRarefactionTSV` /
`notation.WriteRarefactionSummaryTSV`; round-trip and byte-identity are
exercised by `TestR1Determinism` and the `notation-corpus rarefy` CLI.
