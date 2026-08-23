# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1705
- Testable candidates: 418
- Candidates with at least 2 physical blocks: 367
- Candidates with at least 3 physical blocks: 229
- Candidates crossing joint classes: 328
- Candidates crossing Currier classes: 328
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 206 |
| WEAK | 1291 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `165319e63f30ae712e0cb47ffc414e333b7dd2f96ab82beedb4c70a2458b30a8`
- Canonical token count: 43713
- Metadata-map SHA256: `165319e63f30ae712e0cb47ffc414e333b7dd2f96ab82beedb4c70a2458b30a8`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `68f605bf7b508a7740a852171c4b31bfd5c599263c76a813a3f969c27592683e` |
| `distance_context_pairs.yaml` | 43713 | `91c31f5fc7cc1d6d6f2f0f4a38a802b175f35e974bd5349c9eda71353436802d` |
| `sequence_analysis.yaml` | 43713 | `58743e7df3634aa511faa589cc31c76b03b8d8810d69562040be67656b335521` |
| `structural_reliability.yaml` | 43713 | `de522e5c5801ed3efc28fbbaf662a661d9fdc86a617e58076606ab01fc20b3fb` |
| `structural_classes.yaml` | 0 | `362a1aa3315151d95390470c5cee5f5053a5143f181290a7563e890b3f4cd6f0` |
| `soft_structural_space.yaml` | 0 | `8689724b6d453840172193a694ee59c02bd8cb7bfc5cbe2b3afd9f87e0a6f698` |
| `soft_structural_pairs.tsv` | 0 | `9d2f075ed71a2d1c2f4c70bd3e9647b8ee616bf186922f49591d93720687d149` |
| `begin_end_top.tsv` | 43713 | `49a1f10d8c6ab140ceb215e368365245f8c82c01bfd0dccb5125b65859336ef8` |
| `distance_context_top.tsv` | 43713 | `cef4494c4b4051b51395b9d20be9cd68e8efeea916e79a424d0c743bff0fe2f0` |
| `structural_validation.yaml` | 43713 | `f813cd408f4dc5d5bad6a0a093665ac81def1ceb285faec1e850b93f994fca2b` |
| `structural_profile_stability.yaml` | 43713 | `1adf894b6ada3caccf855de3c36021baad92e7125840343ff6cc38fdb59c4a1d` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
