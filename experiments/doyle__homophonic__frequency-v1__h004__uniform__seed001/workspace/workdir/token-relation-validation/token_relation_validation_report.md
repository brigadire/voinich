# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1768
- Testable candidates: 460
- Candidates with at least 2 physical blocks: 403
- Candidates with at least 3 physical blocks: 223
- Candidates crossing joint classes: 353
- Candidates crossing Currier classes: 353
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 246 |
| WEAK | 1312 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `9b0b4fba404ddbf574649b9f64c515f9f0ce35113e5d6c8401b01c39ec3d77c3`
- Canonical token count: 43713
- Metadata-map SHA256: `9b0b4fba404ddbf574649b9f64c515f9f0ce35113e5d6c8401b01c39ec3d77c3`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `b978153d2091783f439bceb8affd59ce4e5ff4be7867680cbc85124ba779d6c7` |
| `distance_context_pairs.yaml` | 43713 | `89c52b4c0ae638b8ab1f450fa39be9f2cbd51c25009d13bf72ae5cb18ceb9725` |
| `sequence_analysis.yaml` | 43713 | `f907ec7bd671d3cf7028fffeda716301e71fc5e9703413d8053d6ad2bde6cc3f` |
| `structural_reliability.yaml` | 43713 | `ab92949c338d019363387ac34179dd6795ef66bad6b835e54cc61b713e8331df` |
| `structural_classes.yaml` | 0 | `8ed403f5e0ecf74b4856c90fe7a7a9e63098ead0c97af69b2777d55da63517b6` |
| `soft_structural_space.yaml` | 0 | `062e2cbcbf1cb65f8de73013a1323028c551c11d853dee5b4d3c5824b3fc0f74` |
| `soft_structural_pairs.tsv` | 0 | `85ea60bcb51155ac46dabe3e5197b0c5186aaada5138331a2a51b5260c9fa73a` |
| `begin_end_top.tsv` | 43713 | `4a98254731752e1c24c0639a2747cf420f0f1dfbb4a8a490d883aaaa0084a8d9` |
| `distance_context_top.tsv` | 43713 | `aa1c182d804447058231585cfe57f2ee8f734b9c0b3865baaf4d065585ad3802` |
| `structural_validation.yaml` | 43713 | `a7109ed6e94da7de7f24c831131e1757f998c84c69670eaffe748c9d722f7f5a` |
| `structural_profile_stability.yaml` | 43713 | `fde783e1a8042161c9f0ae46b588dd44c49322c1873a7410d9dc61b906b91d38` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
