# Cross-metadata token relation validation

## Global accounting

- Frozen candidates: 1861
- Testable candidates: 990
- Candidates with at least 2 physical blocks: 455
- Candidates with at least 3 physical blocks: 203
- Candidates crossing joint classes: 395
- Candidates crossing Currier classes: 390
- Candidates crossing hands: 359
- Unknown-metadata tokens excluded from primary evidence: 14453

## Classification counts

| Class | Count |
|---|---:|
| UNIVERSAL | 110 |
| CURRIER_SPECIFIC | 26 |
| HAND_SPECIFIC | 5 |
| BLOCK_SPECIFIC | 786 |
| WEAK | 934 |

## Previously strongest relations

Only relations actually present in the frozen inventory are shown.

| Relation | Family | Blocks | Joint classes | Transfer | q | Classification |
|---|---|---:|---:|---:|---:|---|
| `chedy -> qokeedy` | directional | 3 | 2 | 1.000 | 0.8971 | UNIVERSAL |
| `chol -> daiin` | directional | 4 | 1 | 0.750 | 0.4569 | WEAK |
| `daiin -> chol` | directional | 4 | 1 | 0.750 | 0.4559 | WEAK |
| `or -> aiin` | directional | 9 | 5 | 1.000 | 0.3071 | UNIVERSAL |
| `qokedy -> qokeedy` | directional | 1 | 1 | 0.000 | 0.9564 | BLOCK_SPECIFIC |
| `qokeedy -> chedy` | directional | 3 | 2 | 1.000 | 0.8971 | UNIVERSAL |
| `qokeedy -> qokedy` | directional | 1 | 1 | 0.000 | 0.9564 | BLOCK_SPECIFIC |
| `shedy -> chedy` | directional | 1 | 1 | 0.000 | 0.6594 | BLOCK_SPECIFIC |
| `sho -> daiin` | directional | 2 | 1 | 1.000 | 0.7745 | CURRIER_SPECIFIC |
| `chol / daiin` | distance-profile | 7 | 3 | 0.857 | 0.00455 | WEAK |
| `chedy / qokeedy` | structural | 2 | 2 | 0.500 | 0.9804 | WEAK |
| `chedy / shedy` | structural | 4 | 3 | 0.250 | 0.0529 | BLOCK_SPECIFIC |
| `chol / daiin` | structural | 7 | 3 | 0.143 | 0.9455 | BLOCK_SPECIFIC |
| `qokedy / qokeedy` | structural | 1 | 1 | 0.000 | 0.006312 | BLOCK_SPECIFIC |

## Interpretation

`UNIVERSAL` means a formally stable, cross-metadata transferable local relation under the pre-specified criteria. It does not imply grammar, cipher, natural language, operators, operands, or an algorithmic language. Pooled frequency is not independent replication; the primary unit here is a physical block.

## Reproducibility

- Corpus SHA256: `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`
- Canonical token count: 39026
- Metadata-map SHA256: `148745adbc889150ad1b59715bbfa75fa17e24b566694d94a0445d06393a7e68`
- Seed: 1
- Initial permutations: 1000
- Refinement permutations: 10000

| Frozen input | Stored token count | SHA256 |
|---|---:|---|
| `begin_end_candidates.yaml` | 39026 | `35a64e129cae70a8230cf0e8e431da1382ecc28f3eee856ab4e4db0abba08325` |
| `distance_context_pairs.yaml` | 39026 | `e04a176bd43c40f7e85c6152c0f684f17fa5ae4f864a17d6618573bd10fe0357` |
| `sequence_analysis.yaml` | 39026 | `0bb2dd7a0486bf27c5d7e9532401e95627a5f39d3d2a54d2fcfae3e5807f3c78` |
| `structural_reliability.yaml` | 39026 | `62549643efe9a2eb09c47720776f18d9559a2389d3d781233b2ed15ae75e7bf3` |
| `structural_classes.yaml` | 0 | `6a820553617ddc9ed85552596faf4528afc05ea87ab833bf1867eea48cba2fdb` |
| `soft_structural_space.yaml` | 0 | `0f398ddacfd15cfcea52c6c9dcfc043e6f56670d5ad65577158aa26369d1b4c8` |
| `soft_structural_pairs.tsv` | 0 | `8a8cc7676a3e624ee051b061a20960ddbd32fe8088f1205811f268999382f3ac` |
| `begin_end_top.tsv` | 39026 | `050b9eba94466a934706a7ebe0d7c845288a3a51e8e72fb4cc81ff00a60fee60` |
| `distance_context_top.tsv` | 39026 | `8acb681837066de798d6f28196c66f242c414ff0dc1451db6033721c3c5a7209` |
| `structural_validation.yaml` | 39026 | `a5ed5d4ccad955ccf737da609877701342f06ca9db210fa1c9171400e5e36e89` |
| `structural_profile_stability.yaml` | 39026 | `9b7010a6f809d2aa07faa323b143f594b90200f9d359f8009fdfbf70ef09f29c` |

Physical blocks are zero-based contiguous runs of one known Currier×hand state, exactly matching residual diagnostics. Primary directional eligibility requires both token counts ≥5 and ≥5 directional observations; primary structural support requires both counts ≥10, with ≥5 retained separately as descriptive evidence. Refinement is restricted to raw p<0.01, at least 3 eligible blocks, and at least 2 joint classes. Inputs made on the 38,887-token corpus contribute candidate identities and frozen settings only; every validation statistic is recomputed on the canonical corpus.
