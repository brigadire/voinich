# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 2105
- Testable candidates: 642
- Candidates with at least 2 physical blocks: 551
- Candidates with at least 3 physical blocks: 344
- Candidates crossing joint classes: 499
- Candidates crossing Currier classes: 499
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 335 |
| WEAK | 1491 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `d0aeb9696c8524fe29e6cccf27367ce3c117cdc96b76d0493850828b1931ee1d`
- Canonical token count: 43713
- Metadata-map SHA256: `d0aeb9696c8524fe29e6cccf27367ce3c117cdc96b76d0493850828b1931ee1d`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `783ec4eef2844575024fcde5677ead7dd4fc51dff9ac0f21c2fd99b242d9ee71` |
| `distance_context_pairs.yaml` | 43713 | `2961eff7677bf71df62b66c31f3247850943f63b96ee51f12d621290515e34d6` |
| `sequence_analysis.yaml` | 43713 | `7bcdc580932ff9ea25a9351de8f2ecc46ba189f247760a39248be4246f97d0b1` |
| `structural_reliability.yaml` | 43713 | `11f3ed36f2d43d53eed37d46c67343d6f7612012b23f37b42493a6bd3768a34d` |
| `structural_classes.yaml` | 0 | `2a3c049261e1268cfb3f05f1a9d7bbf0c532b0c63e86808c88f12ddc28ee4c04` |
| `soft_structural_space.yaml` | 0 | `2bd77a79a20f9a1007e1f3e56d3e99464c33961c35f1bc3181b081e19556aa46` |
| `soft_structural_pairs.tsv` | 0 | `aac915911b462e7de8dddf893e913b9af4880200d3b1164359ef7d2bcacbc904` |
| `begin_end_top.tsv` | 43713 | `45c921881c439f06b7afe1410668398fbcd871bdef82be4eebcb6995ae240972` |
| `distance_context_top.tsv` | 43713 | `76639f945057853e5d8502facba1fa3f2e189714dedf4ea92911228c35df04c8` |
| `structural_validation.yaml` | 43713 | `c16032589eb9a9bb16f9f57ff438255de05c675040eabce1d51330abeddc05df` |
| `structural_profile_stability.yaml` | 43713 | `3b03f71383cd145534d44f6cc8e530f615f97b047e3f48942bb9ec9bcbf14610` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
