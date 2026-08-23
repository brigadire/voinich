# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1479
- Testable candidates: 255
- Candidates with at least 2 physical blocks: 243
- Candidates with at least 3 physical blocks: 206
- Candidates crossing joint classes: 232
- Candidates crossing Currier classes: 232
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 56 |
| WEAK | 1224 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `4bc3c5b699cb6e433399c668b5cc49478f3c6d045128ca6c3a2bd0ce3a1423e5`
- Canonical token count: 43713
- Metadata-map SHA256: `4bc3c5b699cb6e433399c668b5cc49478f3c6d045128ca6c3a2bd0ce3a1423e5`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `030a68c1f73e2ac8a45f3017a4f2d7547007ce67bb39ab7c8b9036e45689381b` |
| `distance_context_pairs.yaml` | 43713 | `4adc32ebb2d52cadd8a20e028e6d0e0f17840e62e29cfd60708a9cabd2c155c3` |
| `sequence_analysis.yaml` | 43713 | `e81748ee97c1588be2d8dabb2541933e3aaf672fb51c6b2a9930fb261930f97a` |
| `structural_reliability.yaml` | 43713 | `8e1696b3a18c27c6d72f43bd4ce06b9890d9d074c88c8b9411ab9cc14cbc85cc` |
| `structural_classes.yaml` | 0 | `a6b3ae5254bd2fc69989a50a03dbe49771f21251d9db1c23040597c9428eedb7` |
| `soft_structural_space.yaml` | 0 | `f030e3cbe87ec211d83aac22db61f061eb24341ae360ea9354987a1c72c6b483` |
| `soft_structural_pairs.tsv` | 0 | `946a3ee7d9c77227af52b7cb1fb73ef16dc6bd002aeacb485bec46986896a67c` |
| `begin_end_top.tsv` | 43713 | `a95fbbfc8894f225e86ba8dd35932d79eb2e9b4a9d748e6a1e341123debe7f26` |
| `distance_context_top.tsv` | 43713 | `aa1c182d804447058231585cfe57f2ee8f734b9c0b3865baaf4d065585ad3802` |
| `structural_validation.yaml` | 43713 | `eef3f9a1863ed9e84b97909e59aa446c012724d62cb16d5bb47af6a1e0d5698b` |
| `structural_profile_stability.yaml` | 43713 | `18200a36b093ec61d4865baf78d1a3419e4de092e6ac352d124f8540a8895806` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
