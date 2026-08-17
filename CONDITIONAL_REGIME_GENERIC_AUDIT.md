# Conditional-regime genericity audit

## Final verdict

`FULLY_METADATA_DEPENDENT`

All exported scientific results of `conditional-regime-analyze` depend constitutively on aligned Currier and/or Davis-hand labels. The package contains corpus-generic mechanics (token ingestion, window vectors, distances, clustering, label-comparison statistics, and change-point primitives), but stage 21 never exposes those mechanics as a scientifically meaningful corpus-only analysis. They are always applied to metadata-selected classes, metadata-defined contiguous blocks, metadata-conditioned residuals, or null models whose exchangeability units are those blocks.

Consequently, a generic pipeline for this stage is `NOT_APPLICABLE`. Removing the metadata input, replacing it with one synthetic class, or treating corpus lines as surrogate blocks would change the hypotheses, samples, null distributions, RNG consumption, multiple-testing families, and outputs. It would not be a generic execution of the existing experiment.

## Scope and classification rule

The audit traces the executable entry point, `internal/conditionalregime`, its writers and reports, its local/process/remote workers, and the task-19 scientific specification that stage 21 implements.

The classifications used below are:

- `CORPUS_GENERIC`: remains scientifically well-defined, with the same hypothesis, using only the token sequence. A reusable helper is marked generic only as a support primitive; that does not make a stage output generic.
- `METADATA_DEPENDENT`: metadata defines the analyzed population, strata, training baseline, physical exchangeability blocks, statistic, or null model. Such a component normally uses corpus tokens as well.
- `MIXED`: an aggregate or comparison combines more than one metadata-derived result or an external reference. It does **not** mean that a corpus-only subset exists.

The distinction is scientific rather than structural: carrying labels in a common struct is insufficient evidence of dependence, while changing the sample or null randomization when labels disappear is sufficient evidence.

## Actual data flow

```text
corpus text --------------------------> flattened token sequence + corpus hash
                                             |
token_metadata_map.tsv -> Currier/hand ------+-> aligned labels
                                             |      |
                                             |      +-> joint, Currier-only, hand-only
                                             |          contiguous blocks + eligibility
                                             |                  |
                                             |                  +-> Part A within-class analyses
                                             |                  |      +-> block-preserving Null A
                                             |                  |
                                             |                  +-> Part B class-conditioned residuals
                                             |                         +-> global Null A
                                             |                         +-> Part C conditional boundaries
                                             |                         +-> within-class/block transition Null B
                                             |
cluster_metadata_global_summary.tsv --------+-> original-vs-residual NMI comparison only
```

`token_metadata_map.tsv` is mandatory and is length-checked against the flattened corpus. Runtime parsing reads only `token_position`, `currier`, and `hand`; empty, `?`, and `null` labels become unknown. A block is a maximal contiguous run of one active class, and an unknown position splits it. The primary scheme is the Currier × hand joint class, with Currier-only and hand-only sensitivity schemes.

Neither `metadata_validation_report.md` nor raw IVTFF is read by stage 21. Folio and line identifiers are not read either. IVTFF/folio/line information is therefore an **upstream provenance and validation dependency**, not a direct runtime field dependency. The stage needs the validated, position-aligned labels and their induced contiguous block boundaries.

The optional hard-coded input `workdir/metadata-validation/cluster_metadata_global_summary.tsv` supplies only the task-18 original/global Currier and hand NMI baselines. If absent, execution continues with zero baselines, degrading that comparison but not enabling a generic mode.

## Component and output audit

| component | output | corpus input | metadata input | classification | generic feasibility | reasoning |
|---|---|---|---|---|---|---|
| Corpus ingestion | flattened tokens; corpus SHA-256 (internal) | Full corpus token stream | None | `CORPUS_GENERIC` | Support-only | Reading and hashing tokens is generic, but it is not a standalone stage-21 scientific result. |
| Generic numerical primitives | windows, distance matrices, cluster labels, silhouette/ARI/NMI/change statistics (internal) | Token windows or derived vectors | None intrinsically | `CORPUS_GENERIC` | Reusable library code only | The algorithms accept generic numeric data; every stage-21 caller supplies metadata-selected windows or metadata-conditioned residuals. |
| Metadata ingestion and alignment | Currier/hand arrays; metadata SHA-256 (internal) | Corpus length for alignment check | `token_metadata_map.tsv`: position, Currier, hand | `METADATA_DEPENDENT` | None | A missing or misaligned map aborts execution. |
| Class/block construction and eligibility | `conditional_class_inventory.tsv` | Token count/positions | Currier, hand, joint labels; unknown gaps; contiguous runs | `METADATA_DEPENDENT` | None | Classes, block counts, exclusion reasons, token support, and eligible analysis universe are defined by metadata. |
| Part A within-class sweep | `within_class_regimes.tsv` | Token frequencies in windows | Scheme, class membership, class blocks | `METADATA_DEPENDENT` | None | Windows are built only inside eligible metadata blocks and never cross their boundaries; K/scale winners are class-specific. |
| Part A stability | stability rows in `within_class_stability.tsv` | Class-window vectors | Metadata blocks and class identity | `METADATA_DEPENDENT` | None | Cross-validation leaves out physical metadata blocks, or uses contiguous folds within a single metadata block. |
| Part A cross-block recurrence | recurrence rows in `within_class_stability.tsv` | Cluster labels for class windows | Distinct metadata blocks and class identity | `METADATA_DEPENDENT` | None | The statistic explicitly asks whether a regime recurs across blocks belonging to the same Currier/hand class. |
| Part A Null A and refinement | `within_class_permutations.yaml` | Tokens and observed silhouette statistic | Class blocks as independent shuffle domains | `METADATA_DEPENDENT` | None | Tokens are shuffled independently within metadata blocks. Removing labels changes the exchangeability assumption, p-values, refinement set, and correction family. |
| Part A visualization | `within_class_stability_by_scale.svg` | Part A stability values | Class/scheme labels and eligible blocks | `METADATA_DEPENDENT` | None | It visualizes a metadata-conditioned result rather than an independently generic statistic. |
| Part B residual construction and sweep | `residual_cluster_assignments.tsv`; `residual_cluster_summary.tsv` | Window token-frequency vectors | Joint classes, class blocks, class-specific training means and variances | `METADATA_DEPENDENT` | None | Each residual is defined as deviation from its Currier × hand training baseline; raw and standardized representations are both conditioned. |
| Part B metadata association/reference comparison | `residual_metadata_association.tsv` | Residual cluster assignments | Residual window Currier/hand labels; optional task-18 global association summary | `MIXED` | No corpus-only subset | NMI/ARI is explicitly measured against Currier and hand, then compared with an external metadata-association baseline. |
| Part B global correction | `residual_permutations.yaml` | Tokens and maximum residual silhouette | Eligible joint classes and their physical blocks | `METADATA_DEPENDENT` | None | Each null replicate shuffles within every class block, rebuilds conditioned residuals, and takes a global maximum over the frozen search. |
| Part B visualizations | `residual_cluster_stability.svg`; `original_vs_residual_currier_nmi.svg`; `original_vs_residual_hand_nmi.svg`; `residual_regime_metadata_entropy.svg` | Residual clustering results | Joint classes; Currier/hand association; optional task-18 baseline | `MIXED` | No corpus-only subset | All plotted quantities originate in class-conditioned residuals; several additionally visualize direct metadata association. |
| Part C conditional boundaries | `conditional_stable_boundaries.tsv` | Tokens and generic change-detector calculations | Eligible joint class and physical block boundaries | `METADATA_DEPENDENT` | None | Detection is restarted inside each Currier × hand block. Support and recurrence count distinct metadata classes/blocks. |
| Part C residual transitions and Null B | `residual_transition_matrix.tsv` | Ordered residual-cluster labels | Residual class and block identity | `METADATA_DEPENDENT` | None | Observed transitions never cross class/block boundaries; the null permutes labels separately within each class/block sequence. |
| Part C candidate synthesis | `residual_regime_candidates.tsv` | Residual cluster sizes/stability and transition evidence | Currier/hand/joint class counts, block counts, entropy, conditional-boundary support | `METADATA_DEPENDENT` | None | Candidate ranking and interpretation combine only metadata-conditioned evidence. |
| Part C visualization | `residual_transition_enrichment.svg` | Transition enrichment values | Metadata-bounded residual sequences | `METADATA_DEPENDENT` | None | The plotted transition null and observations are metadata-bounded. |
| Aggregate machine-readable result | `conditional_regime_analysis.yaml` | Corpus hash, parameters, all Parts A-C results | Metadata hash, schemes, classes, blocks, all conditioned results | `MIXED` | No generic branch | This is a composite of the metadata-dependent analyses and reference comparison. |
| Aggregate narrative | `conditional_regime_report.md` | All computed results | All metadata-conditioned components and metadata-independence comparison | `MIXED` | No generic branch | The report's scientific question is whether structure survives control for Currier and hand. |

## Part A: within-class structure

Part A does not merely annotate an otherwise global result. For each joint, Currier-only, and hand-only class, it constructs windows within that class's contiguous blocks, performs its own scale/K sweep, validates stability by metadata-block folds, and measures recurrence across those blocks.

Null A preserves each block's length and unigram composition by shuffling tokens independently inside the block. A corpus-only shuffle would allow exchange across Currier/hand strata or require a new arbitrary block definition. Either choice tests a different null hypothesis. Creating a single synthetic class does not preserve existing behavior: it changes the window population, block folds, observed statistics, permutation seeds consumed by the job set, p-value correction family, and final rows.

Classification: `METADATA_DEPENDENT` for every exported Part A result.

## Part B: residual structure after metadata control

The defining operation is metadata conditioning. For each window, the code subtracts a mean estimated from training windows of the same joint Currier × hand class; standardized mode additionally uses the corresponding training variance. It then pools these residuals for clustering. Even before the explicit NMI/ARI comparison, the numeric input to clustering depends on metadata labels and metadata-block folds.

The global permutation correction repeats the same conditioned construction after within-block shuffles. The residual association output then measures remaining association with Currier and hand and optionally compares it with the frozen task-18 global baseline.

Without metadata there is no unchanged meaning for “residual after Currier × hand.” Subtracting one global mean would convert the experiment into ordinary global clustering and would largely duplicate the global-regime analysis rather than expose a hidden generic subset of stage 21.

Classification: core residual discovery and its null are `METADATA_DEPENDENT`; the external original-vs-residual metadata comparison and aggregate visualizations are `MIXED`, with no corpus-only scientific component.

## Part C: conditional localization and transitions

The boundary detector is algorithmically generic but its stage-21 unit of analysis is not. It is invoked independently inside eligible joint-class physical blocks, cannot emit a boundary across such a block, and summarizes recurrence by metadata class/block support.

Transition counts operate on ordered Part B residual labels only when adjacent windows share both class and block. Null B permutes labels inside each class/block sequence. Removing metadata would change which adjacencies exist and the permutation strata, so it changes both the observed matrix and its reference distribution.

Classification: `METADATA_DEPENDENT` for all exported Part C outputs.

## Corpus-only feasibility and reproducibility

There is no substantial subset that can run without synthetic metadata while preserving the current scientific questions:

1. Empty/unknown labels produce no eligible classes and therefore no analysis, not a generic subset.
2. A single synthetic label turns a conditional experiment into a global one and invents a block model.
3. Corpus lines cannot substitute for blocks: stage 21 does not ingest line identifiers, and selecting lines as exchangeability units would introduce a new scientific assumption.
4. Running the reusable window/clustering or change-point primitives over the entire corpus would be a new output contract and overlap the existing global-regime analysis.
5. A metadata-optional branch would alter loop cardinalities, job IDs, seed derivation/consumption, null distributions, multiple-testing correction, checkpoints, reports, and serialized output. It cannot meet the requirement that Voynich outputs remain byte-identical without maintaining two scientifically different pipelines.

No refactor is recommended. The minimal and clearest architecture is the existing conceptual separation: corpus-global analysis belongs in the global-regime stage, while `conditional-regime-analyze` remains explicitly annotation-conditioned. If another corpus supplies real, aligned categorical annotations, the engine may be reusable as an **annotated-corpus conditional analysis**; that is broader than Voynich-specific nomenclature but is still not corpus-generic.

Because this audit changes no implementation, current Voynich behavior and bytes are unaffected.

## Distributed-workload assessment

The distributed executor handles only:

- `part_a_significance` and `part_a_refinement`, whose job identity and computation include scheme, class, scale, method, and class blocks;
- `part_b_global_correction`, which reconstructs residual maxima from eligible joint classes and their blocks.

Worker initialization requires both corpus and metadata-map inputs, verifies both fingerprints, builds the same class/block inventory, and rejects incompatible work. Part C is local. Thus the amount of existing distributed workload applicable to a corpus-only subset is **zero**. Genericizing only low-level math helpers would not remove any current remote job, and adding a whole-corpus path would create new work for a different experiment rather than reuse a valid subset.

## Verification

The audit was checked against the current implementation and its local, subprocess, and remote protocol tests:

```text
GOCACHE=/tmp/voinich-go-build go test ./internal/conditionalregime ./conditional-regime-analyze ./pipeline-orchestrate
```

Result: all tests passed. The test was run with loopback access because remote protocol tests create a local listener; no code changes were made for this audit.

## Required disposition

Final verdict: `FULLY_METADATA_DEPENDENT`

Generic pipeline: `NOT_APPLICABLE`
