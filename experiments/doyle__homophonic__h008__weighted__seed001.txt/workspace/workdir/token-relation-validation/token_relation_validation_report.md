# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1388
- Testable candidates: 233
- Candidates with at least 2 physical blocks: 224
- Candidates with at least 3 physical blocks: 205
- Candidates crossing joint classes: 217
- Candidates crossing Currier classes: 217
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 32 |
| WEAK | 1156 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `d3b3b297e1bf7f7cdd59ffbc2fd38c440eeac461b8db59bbcdd373f211d05057`
- Canonical token count: 43713
- Metadata-map SHA256: `d3b3b297e1bf7f7cdd59ffbc2fd38c440eeac461b8db59bbcdd373f211d05057`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `6821721246e0e60499ddb0689a48acdcdaecc1e3f32446305a7bd3c409f9fb58` |
| `distance_context_pairs.yaml` | 43713 | `6c8c0413322e3b002d1a730583ae10696f7a17a294adfb91354d8de289b67d26` |
| `sequence_analysis.yaml` | 43713 | `70f903107d82c7701571aba5b7c16c6a914e73c5d5fcb61ed1bd06fa90277712` |
| `structural_reliability.yaml` | 43713 | `4df12dd78299cbcd125f5b07f880511aaf28e5bdc51e7d75d3a663a322953568` |
| `structural_classes.yaml` | 0 | `d7cbcbe58e791dc90479d2095cf302242fee010fca3bbbce79cca42d3b718296` |
| `soft_structural_space.yaml` | 0 | `a28cfe856105ff18954b6adb21190e29147cb981e482cfc2d9ca756a433a3bb9` |
| `soft_structural_pairs.tsv` | 0 | `e15f187c2031dd6b4dd6f946ae42c8c26016cf2d709b39e1df5308489af9040c` |
| `begin_end_top.tsv` | 43713 | `972b99ec2a35f359e5d68ee04fbceea0caca505b491896c1f2f514f608417024` |
| `distance_context_top.tsv` | 43713 | `aa1c182d804447058231585cfe57f2ee8f734b9c0b3865baaf4d065585ad3802` |
| `structural_validation.yaml` | 43713 | `87e20df571f85b12b10fc7355fef4a2be6719fe44c13638fe813f4ffbdc04763` |
| `structural_profile_stability.yaml` | 43713 | `25e263d06837c0fe18b0ecd5e16251b4f6f73de72293d0700c8ec20e7774e4d6` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
