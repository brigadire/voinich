# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1409
- Testable candidates: 240
- Candidates with at least 2 physical blocks: 230
- Candidates with at least 3 physical blocks: 204
- Candidates crossing joint classes: 213
- Candidates crossing Currier classes: 213
- Candidates crossing hands: 0
- Unknown-metadata tokens excluded from primary evidence: 0

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 0 |
| CURRIER_SPECIFIC | 0 |
| HAND_SPECIFIC | 0 |
| BLOCK_SPECIFIC | 40 |
| WEAK | 1169 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy / shedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |
| `qokedy / qokeedy` | structural | 0 | 0 | 0.000 | 1 | WEAK |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `85b8e2a2ed84cbd42858479ece22956747770afcdb141dad6a25899ecbe56dce`
- Canonical token count: 43713
- Metadata-map SHA256: `85b8e2a2ed84cbd42858479ece22956747770afcdb141dad6a25899ecbe56dce`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 43713 | `52b327b1d1877054863d0d97794bd93688d040d3b37103dee22663449cc871f4` |
| `distance_context_pairs.yaml` | 43713 | `00e9fc23b56873ff1bf645f64178f7c346e789176cd053c0c1b4a44efecf0f44` |
| `sequence_analysis.yaml` | 43713 | `f82a8b80f665ac3ec0bad8eaca3478d3d943836d0514b94acfad4ef98986fc9b` |
| `structural_reliability.yaml` | 43713 | `16681db9c561177c85ce81cfe30ce0364771371d2ef0589673ce5be3312db175` |
| `structural_classes.yaml` | 0 | `b1382fd683e26f92455de4731a784966ce4f5e16bb20783457144b8af41799b4` |
| `soft_structural_space.yaml` | 0 | `406c22979e4aa317945202a52aaf0ce1d25bc7b8049da7fe9803a8013f290e99` |
| `soft_structural_pairs.tsv` | 0 | `e42e37d4f286e22916d3ad905273c114afd14bc6af540c56a4271ed59eb67499` |
| `begin_end_top.tsv` | 43713 | `d329ae10f0fc0ee4ef6eeb1b6da607e1a3b16a0949f87854fccb3be641c0109a` |
| `distance_context_top.tsv` | 43713 | `aa1c182d804447058231585cfe57f2ee8f734b9c0b3865baaf4d065585ad3802` |
| `structural_validation.yaml` | 43713 | `abe3a3619cece824769a4b6c00457d5ed29f6fd18d10bfa6ff2c11717fda62f2` |
| `structural_profile_stability.yaml` | 43713 | `e9764f024414f9103c685b26a7f93d2d508f100972a62c9dc5efb5c20deaf4ca` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
