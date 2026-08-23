# Our pipeline metric inventory (Stages 1–28)

Implementation audit based on `pipeline-orchestrate/stages.go`, each stage's
`main.go` and its internal writers. Task56 changes no scientific code.

`N/A` is a non-meaningful estimator because required labels are absent;
`EMPTY` is a valid run with no eligible observations; `NOT_COMPUTED` is an
unavailable artifact/quantity. They are not numeric zero.

| Stage | command | purpose | metric / definition | null, test, or threshold | output artifact |
|---:|---|---|---|---|---|
| 1 | `dict-gen` | canonical corpus inventory | token count, type count, frequencies and contexts | descriptive | `workdir/dataset/dictionary.yaml` |
| 2 | `dict-analyze` | dictionary analysis | frequency spectrum, token lengths, first/last/neighbor summaries | descriptive frequency filters | `tokens_analysis.yaml` |
| 3 | `structural-analyze` | structural token relations | similarity, transition counts, predecessor/successor distributions, Shannon context entropy | min frequency/transition/context counts; reliability prior for ranking | `structural_analysis.yaml` |
| 4 | `sequence-analyze` | sequence survey | n-gram counts, repeated n-grams, context extensions, next-token entropy | min count/observation filters; descriptive | `sequence_analysis.yaml` |
| 5 | `begin-end-analyze` | boundary relations | token-pair distance/window and begin/end association profiles | 100 page/line boundary-preserving permutations; empirical upper tails | `begin_end_*`, report |
| 6 | `structural-normalize` | class projection | normalized structural-class sequence | deterministic transformation | `normalized*.txt`, `structural_classes.yaml` |
| 7 | `normalization-compare` | raw/class comparison | n-gram and sequence summaries plus raw-vs-class deltas | paired descriptive comparison | `normalization_comparison.yaml` |
| 8 | `structural-validate` | validate classes | class coverage, support and validation counts | frozen class/eligibility gates | `structural_validation.yaml` |
| 9 | `structural-profile-stability` | profile stability | fold/resample profile overlap/similarity | bootstrap/fold stability; empty support is `EMPTY` | `structural_profile_stability.yaml` |
| 10 | `structural-reliability` | relation reliability | entropy, `exp(H)` effective vocabulary, context diversity, bootstrap CI/probability | bootstrap; threshold is not a universal p-value | `structural_reliability.yaml` |
| 11 | `soft-structural-space` | uncertain relations | pair similarity and bootstrap probability above .70 | bootstrap threshold | `soft_structural_space.yaml`, TSV |
| 12 | `structural-graphemic` | two-coordinate comparison | Pearson/Spearman graphemic-vs-structural similarity | descriptive; no ordinary pair p-values | report/TSV |
| 13 | `structural-pair-decompose` | pair decomposition | Jaccard, JS similarity, entropy in nats, `exp(H)`, positional/context agreement | selected pairs and controls | pair TSV/YAML/report |
| 14 | `distance-context-analyze` | context-distance response | directional, boundary and position distance effects | predefined controls/permutation diagnostics | distance-context artifacts |
| 15 | `local-regime-analyze` | local regimes | window distributions, distances and contrasts | deterministic windows/control comparisons | local-regime artifacts |
| 16 | `property-trajectory-analyze` | property trajectory | structural/context properties by corpus position | descriptive; no semantic labels | trajectory artifacts |
| 17 | `structural-projection-analyze` | project hypotheses | trial/candidate projection scores and family summaries | frozen trials; checkpointed executor | projection artifacts |
| 18 | `global-regime-analyze` | discover regimes | window vectors, distances, clusters/change points | deterministic seed; discovery later audited | global-regime artifacts |
| 19 | `metadata-validate` | external-label validation | boundary/cluster agreement, tolerance curves, NMI/ARI-style associations | 10,000 metadata-aware null permutations; **generic N/A** | metadata report/map |
| 20 | `cluster-metadata-global` | global metadata test | cluster–Currier/hand association over frozen search space | 10,000 block-aware permutations/global correction; **generic N/A** | metadata-global report |
| 21 | `conditional-regime-analyze` | residual regimes | within-class/residual clustering and diagnostics | 1,000 block-aware permutations; eligibility gates; **generic N/A** | `conditional-regimes/` |
| 22 | `residual-diagnostic-analyze` | residual explanation | residual association with Currier/hand/joint labels | 1,000 block-aware permutations; **generic N/A** | `residual-diagnostics/` |
| 23 | `token-relation-validate` | replicate relations | candidate/tested/significant relation counts and classes | 1,000 initial + 10,000 refinement within-block permutations/FDR | `token-relation-validation/` |
| 24 | `replicated-local-structure-audit` | confirm local sequence | Markov/next-token profiles, conditional entropy, held-out replication | 1,000 null replicates and block/LOBO replication | `replicated-local-structure/` |
| 25 | `higher-order-sequence-validate` | higher-order dependence | CMI/conditional entropy, occurrence breadth, replicated status | 10,000 conditional-neighbor permutations; secondary/LOBO/jackknife | `higher-order-sequences/` |
| 26 | `positional-continuation-validate` | positional continuation | position-stratified continuation and entropy/MI contrasts | 10,000 positional/stratified permutations; LOBO comparisons | `positional-continuation/` |
| 27 | `transition-network-validate` | directed graph validation | eligible tokens, significant/backbone edges, profiles, entropy, held-out log-loss/model order | 1,000 + 10,000 edge permutations; metadata-transfer subtest **generic N/A** | `transition-network/` |
| 28 | `vocabulary-growth-analyze` | lexical productivity | `V(n)`, TTR, hapax/dis/tris, frequency-of-frequencies, `log V=log K+β log n`, new-type rate | shuffled-token null preserves multiset and destroys order; empirical p/effect | `vocabulary-growth/` |

## Cross-cutting implementation facts

* Task46 transposition is a pure token permutation: it preserves token count,
  types and frequencies and changes adjacency. Task54 tests inverse recovery.
* Task46 homophony is opaque, reversible and one-output-token-per-input-token.
  `global-H` fixes H; `triangular-v1` changes selection probabilities; Task55
  `frequency-v1` changes allocation by frequency rank, with selection still
  independent of plaintext frequency.
* Task49's shuffled null tests order-dependent growth, not vocabulary size.
* Task52 distinguishes `0`, `NA`, `NOT_APPLICABLE`, `NOT_COMPUTED`,
  `MISSING_ARTIFACT`, and `INCOMPATIBLE_METRIC`; absent checkpoints are not 0.
* Stages 19–22 require external IVTFF/Currier/hand truth. Stages 23–27 have
  generic modes using deterministic blocks; they do not reproduce the
  Currier/hand hypothesis. Stage 26 substitutes its generic target from Stage
  25 rather than using the Voynich literal tokens.

The inventory deliberately does not call metrics equivalent by name. For
example, glyph conditional entropy, token transition entropy, higher-order
CMI, and pair-context entropy are separate estimands.
