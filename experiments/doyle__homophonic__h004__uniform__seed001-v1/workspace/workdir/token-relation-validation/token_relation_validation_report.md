# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1588
- Testable candidates: 327
- Candidates with at least 2 physical blocks: 291
- Candidates with at least 3 physical blocks: 209
- Candidates crossing joint classes: 274
- Candidates crossing Currier classes: 274
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 124 |
| WEAK | 1263 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `d44a2dabd08cf969d23a2ec4e91f719eb9da85f49f966184f31aa40f745dead9`
- Canonical token count: 43713
- Metadata-map SHA256: `d44a2dabd08cf969d23a2ec4e91f719eb9da85f49f966184f31aa40f745dead9`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `8890646ee0ae6173dd7bfad4438b7a9c2cf5de3d4e774cef4eebf76ebf4bfb5c` |
| `distance_context_pairs.yaml` | 43713 | `f49c68f9763e58bf90aefffde90d4de548aff32a5af5f29339cfdf47069bdf96` |
| `sequence_analysis.yaml` | 43713 | `30e61ba86d837a5fd8b3784a8ebed290c795a16ceb15199f99832a88c25dbeed` |
| `structural_reliability.yaml` | 43713 | `b2e63d7a16aef9e174602b296501f1256edd85c64bcced465b80ce675112407b` |
| `structural_classes.yaml` | 0 | `83e29c5aa7047b9e28e9d5edbad514ce5f1914e1012df4383f2507701f8c5fdf` |
| `soft_structural_space.yaml` | 0 | `0e9117d84b4598313d24230f5daf1cacac86955670c7a951551077bddb4961ef` |
| `soft_structural_pairs.tsv` | 0 | `efcdf75985ff69be994a08f1deea642878a27cf4ea69ea41ad32cfb95a49627f` |
| `begin_end_top.tsv` | 43713 | `63a1107999ae3a8c5aa2c0133353201598d2d385c86fb1d0ce40a4a49ad54e4e` |
| `distance_context_top.tsv` | 43713 | `d74d9782e80cb2e995a66ce8f8f212a6660db683d18045db94407dc7bee57879` |
| `structural_validation.yaml` | 43713 | `5aa13b370018cc799c8fcf66ee67ce4343108ac42c1b0bf70d5bddac4357adb8` |
| `structural_profile_stability.yaml` | 43713 | `6f884f6d6bfefae251ea873797bf978c1f99bf98a193945120e5dc3b6943441c` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
