# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1394
- Testable candidates: 218
- Candidates with at least 2 physical blocks: 217
- Candidates with at least 3 physical blocks: 195
- Candidates crossing joint classes: 208
- Candidates crossing Currier classes: 208
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 23 |
| WEAK | 1176 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `8a27e48af6f49cb88d37ed4425e06d28d8b17c0652e299d15e677a7f3227ac3d`
- Canonical token count: 43713
- Metadata-map SHA256: `8a27e48af6f49cb88d37ed4425e06d28d8b17c0652e299d15e677a7f3227ac3d`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `c2f22bdb5e8e3d4c4ecaf38eb6b71fc315762664736f952032b22c1321c6f391` |
| `distance_context_pairs.yaml` | 43713 | `538a201a50d1ebd93dd29097d8c96d6a74672621895460f0bb03d646137740c2` |
| `sequence_analysis.yaml` | 43713 | `4f7462189f06922933dd884c55afe1694229650fd9543e08a771035196f6af05` |
| `structural_reliability.yaml` | 43713 | `d67a8abbb4b90ccf8662e54b835bd87d3fdf47e10b78b9e42bb05fe9655a8a52` |
| `structural_classes.yaml` | 0 | `b15257a8c5e771bb7c7fb2410b28d9084f0112380104c01765c224d535eeae2a` |
| `soft_structural_space.yaml` | 0 | `4be448dc2820ec97f0ee46c92940c74ab72f047831399064713816b420cb7433` |
| `soft_structural_pairs.tsv` | 0 | `81cd1d1203516dafa8fe215952237a019b4c7ff9849c11a8d297e56eda150e6f` |
| `begin_end_top.tsv` | 43713 | `b61a5d230f725225ca84672c4849ecadb4088251d9f629b5fac394a360665b37` |
| `distance_context_top.tsv` | 43713 | `aa1c182d804447058231585cfe57f2ee8f734b9c0b3865baaf4d065585ad3802` |
| `structural_validation.yaml` | 43713 | `cb3e3c45a7f3dd1557e9d2fa7ab9ae3a488da159fa80189abb16b4f808f052a1` |
| `structural_profile_stability.yaml` | 43713 | `aed11bb0e391fc9bf668106842bdaadd2166eb78b070a72490216d0425087e5a` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
