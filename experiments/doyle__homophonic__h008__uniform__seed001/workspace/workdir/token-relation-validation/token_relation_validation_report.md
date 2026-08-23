# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1384
- Testable candidates: 210
- Candidates with at least 2 physical blocks: 209
- Candidates with at least 3 physical blocks: 193
- Candidates crossing joint classes: 197
- Candidates crossing Currier classes: 197
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 17 |
| WEAK | 1174 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `b22cd99a1f450bdad831fe3af004ff2825ed36490983b39337afe6fee7254983`
- Canonical token count: 43713
- Metadata-map SHA256: `b22cd99a1f450bdad831fe3af004ff2825ed36490983b39337afe6fee7254983`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `8026168240c2d28927b996d86d0d3708155691ce50660ada15a12bf9d526788a` |
| `distance_context_pairs.yaml` | 43713 | `9e567792ff3023cff785ab6042008323b260a0ce5d8c65756b248dc9f07f4d15` |
| `sequence_analysis.yaml` | 43713 | `0874e3492eb9cb035cb927ec9a5801f4d77ecc08ecc6b2eb4b4e28bb07ea39b0` |
| `structural_reliability.yaml` | 43713 | `7af1fda96484757b20f3eb4ab58624db1d3c97b454c738e0027cab3263ff7208` |
| `structural_classes.yaml` | 0 | `89dc44dcc5408dc031d0a72321409e43eccfc5bf9e9193b6a0a4640ff8db593c` |
| `soft_structural_space.yaml` | 0 | `f602409666f55e144bb1fa345ae6666f0562dcecf189f1d646513e271380262d` |
| `soft_structural_pairs.tsv` | 0 | `bfda55bb5aa3e9e27a4cd100b8a2f62036ea09a7b2908aa3f226e88e94dcffe3` |
| `begin_end_top.tsv` | 43713 | `95f8e714d33c25de55023400ce7ad279a4cc120a3e4f89421a088ff58c7fd3e5` |
| `distance_context_top.tsv` | 43713 | `aa1c182d804447058231585cfe57f2ee8f734b9c0b3865baaf4d065585ad3802` |
| `structural_validation.yaml` | 43713 | `cfec66c04b449c0bf5e819932e91b1ce8a73840a0747a2b1571c514581a72a10` |
| `structural_profile_stability.yaml` | 43713 | `acd660f6829142520050322fa61bbccd76ac636f625641675715b8c80fef136d` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
