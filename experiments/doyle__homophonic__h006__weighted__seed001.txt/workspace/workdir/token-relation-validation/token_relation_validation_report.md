# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1449
- Testable candidates: 272
- Candidates with at least 2 physical blocks: 249
- Candidates with at least 3 physical blocks: 206
- Candidates crossing joint classes: 238
- Candidates crossing Currier classes: 238
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 71 |
| WEAK | 1177 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `cdf533ee521a4e8ebb2dc64c3b862c33fc35de8452456a3e2daf58e5232a383c`
- Canonical token count: 43713
- Metadata-map SHA256: `cdf533ee521a4e8ebb2dc64c3b862c33fc35de8452456a3e2daf58e5232a383c`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `9cfaa19fc419c7500ce7b3281c01fc1f7f7740eb7c0cf11a26f66c470af3a79b` |
| `distance_context_pairs.yaml` | 43713 | `f864f54b93246aade93e9a72e7025be033a5e3e20c51a1d04a2d27fb03b51d26` |
| `sequence_analysis.yaml` | 43713 | `c102a7e896cc979595be6848d26e08b99204f5fa25dac525d04c5cdfacebf703` |
| `structural_reliability.yaml` | 43713 | `e3dbf4b8fbc56b04d7a7eecbd81c22b38973bdc3e012d8d9cd816f291ec60cca` |
| `structural_classes.yaml` | 0 | `71c7267571508775914c55626707c0b2cb15775c3d29ee982b6c79bcb05d399c` |
| `soft_structural_space.yaml` | 0 | `c41da007446dff3af721e8f0d35374d039bdc6e46051a9c1cec276c8f2d4a04e` |
| `soft_structural_pairs.tsv` | 0 | `fd789e5c3564b1bd7172c714077dece8fa2d59b6d0a36e2a0d517c45d4d76391` |
| `begin_end_top.tsv` | 43713 | `244e2b99de80209c197e8b5ef07002e2ff7ddb889b32eeccd9bad7f20913cfa6` |
| `distance_context_top.tsv` | 43713 | `afff8b57078d7a612df83df5499b3df5578638cfe4a60e51dd08656af7e5f711` |
| `structural_validation.yaml` | 43713 | `6b53ea1475bade7805c91309d6b045d9afd73a4d1175007cedd6885c7cae6586` |
| `structural_profile_stability.yaml` | 43713 | `a2f29379003c92422c29432ea51adbc4820012e3a9b7ff9a34702c777ce5edbe` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
