# Generic stage applicability audit (Task43)

Stages 1–18 of `pipeline-orchestrate` were already generic-applicable before
this audit (see `pipeline-orchestrate/STAGE_AUDIT.md`). Stages 19–27 were
excluded from generic-corpus experiments with one blanket reason:
`NOT_APPLICABLE: requires IVTFF metadata`. That classification was too
coarse: several of these stages use IVTFF-derived Currier/hand metadata only
as a segmentation device, not as the scientific variable under test.

This document is the full per-stage audit task43 requires: for each stage,
what metadata is actually consumed (traced to the function/line that reads
the value, not inferred from CLI flag names), whether that dependency is
scientific (Class A), segmentation-only (Class B), or an accidental
implementation coupling (Class C), and — for every legitimately B/C stage —
the generic strategy actually implemented.

## Summary table

| stage | current reason (before) | actual metadata dependency | class | generic enabled | generic segmentation used | scientific comparability notes | distribution opportunity |
|---|---|---|---|---|---|---|---|
| 19 metadata-validate | requires IVTFF metadata | validates blind boundary/cluster discovery against real Folio/LineID/ParagraphID/Currier/Hand/Quire as *ground truth* | A | no | n/a | stays `NOT_APPLICABLE`; no generic corpus has an independent segmentation record to validate blind discovery against | n/a |
| 20 cluster-metadata-global | requires IVTFF metadata | whole-search-space permutation test of real Currier/Hand identity vs frozen clustering | A | no | n/a | stays `NOT_APPLICABLE`; the permuted variable *is* Currier/hand identity | n/a |
| 21 conditional-regime-analyze | requires IVTFF metadata | Currier×hand identity is the confound under test, not a partition device | A | no | n/a | stays `NOT_APPLICABLE`; conditioning on real scribal/language identity is the entire hypothesis | n/a |
| 22 residual-diagnostic-analyze | requires IVTFF metadata | entire output surface is Currier/hand/joint-NMI diagnostics on stage 21's residuals | A | no | n/a | stays `NOT_APPLICABLE`; nothing to diagnose without real metadata truth labels | n/a |
| 23 token-relation-validate | requires IVTFF metadata | core block-enrichment/sign-consistency/transfer-success math is block-partition-only (B/C); `Classify`/`ClassifyDetailed`'s UNIVERSAL/CURRIER_SPECIFIC/HAND_SPECIFIC labeling is Currier/hand-conditioned (A) | mixed B/C + A | **yes**, core stats only | generic deterministic blocks (`internal/genericsegmentation`) | generic mode reports `GROUP_CONSISTENT`/`GROUP_LIMITED` instead of Currier/hand-labeled classes; single-dimension MetadataTransfer rows labeled `"group"` | permutation nulls (`PermuteWithinBlocks`), LOBO transfer checks |
| 24 replicated-local-structure-audit | requires IVTFF metadata | distance/sequence null-model math is block-partition-only (B/C); `ROBUST_RELATIVE_REPLICATION`/`METADATA_LIMITED_REPLICATION` status requires real cross-Currier-and-hand transfer (A) | mixed B/C + A | **yes**, core stats only | generic deterministic blocks | generic mode emits `GROUP_REPLICATED`/`GROUP_LIMITED_REPLICATION`, dropped currier/hand report columns, `no_cross_group_transfer` failure condition | shuffle/markov null replicates |
| 25 higher-order-sequence-validate | requires IVTFF metadata | CMI/conditional/occurrence/LOBO/jackknife math only ever treats Block as an opaque partition (B); `DistinctJoint>=2` is a replication-breadth gate, not a hand/Currier test | B | **yes** (chain-gated on stage 24's generic output) | generic deterministic blocks | `DistinctJoint>=2` reinterpreted as "≥2 independent generic resampling folds"; `METADATA_LIMITED` renamed `GROUP_LIMITED`, no longer gated on the always-1 DistinctHand | primary/secondary permutation families, LOBO, jackknife |
| 26 positional-continuation-validate | requires IVTFF metadata | block/position mechanics are pure line/token-index arithmetic (B); **separately**, the stage hardcodes the literal Voynichese target `"s"`/`"aiin"`/`"chey"` — a lexical dependency, not a metadata one | B + separate lexical dependency | **yes** (chain-gated on stage 25's generic output) | generic deterministic blocks + deterministic top-ranked-candidate target substitution | frozen target triple substituted with stage 25's own best-ranked `HIGHER_ORDER_REPLICATED` candidate (no semantic/lexical judgment); documented hypothesis substitution below | positional/stratified permutation tests, LOBO model comparison |
| 27 transition-network-validate | requires IVTFF metadata | transition matrix/edge effects/entropy/permutation-null/held-out-prediction/model-order are block-partition-only (B/C); `computeGraphDiagnostics`'s `dim="currier"`/`"hand"` metadata-transfer rows are a genuine cross-Currier/cross-hand test (A) | mixed B/C + A | **yes**, core stats only | generic deterministic blocks | `dim="currier"`/`"hand"` skipped entirely in generic mode (real `NOT_APPLICABLE` for that sub-component, not a degenerate run); `dim="joint"` kept; `METADATA_SPECIFIC` renamed `GROUP_SPECIFIC` | edge permutation nulls, held-out block folds |

Six of nine stages (23–27, with 25/26 chain-gated on 24/25 respectively) are
now generic-enabled; four (19–22) remain `NOT_APPLICABLE` with a precise,
stage-specific scientific reason instead of the old blanket phrase.

## Phase 1/2 evidence detail

### 19 — metadata-validate (Class A)

**Field table**

| field | source | consumed at | scientific meaning | mathematically required | generic substitute possible |
|---|---|---|---|---|---|
| Folio/LineID/ParagraphID | `internal/metadatavalidation/parser.go` from IVTFF page/locus headers, via `Locus` | `validation.go:ExtractBoundaries` compares consecutive records' Folio/LineID/ParagraphID to emit `MetadataBoundary`s | ground-truth manuscript segmentation boundaries | yes — the reference the blind boundaries are validated against | no |
| Currier/Hand/Quire | `Locus.Variables["C"]/["H"]/["Q"]` (IVTFF `$C=`/`$H=`/`$Q=`) → `TokenMetadata` | `ExtractBoundaries`, `AnalyzeAssignments`, `clusterPermutationSummary` | ground-truth scribal-hand/Currier-language/quire assignment | yes — the dependent variable tested for association with blind clusters | no |

`RunAndWrite` loads frozen corpus-only discovery outputs and validates them
against these five metadata dimensions; there is no path where the stage
does anything except compare blind structure to real external ground truth.

**NOT_APPLICABLE reason:** validates blind boundary/cluster discovery
against real external manuscript ground truth (line, paragraph, folio,
Currier-language, scribal-hand, quire boundaries parsed from IVTFF); a
generic corpus has no independent external segmentation record to compare
against, so the analysis has no referent — not merely a missing input file.

### 20 — cluster-metadata-global (Class A)

**Field table**

| field | source | consumed at | scientific meaning | mathematically required | generic substitute possible |
|---|---|---|---|---|---|
| currier, hand columns | `load.go:loadTokenLabels` from `token_metadata_map.tsv` (stage 19's output) | `compute.go:RunSearchSpace`/`permuteKnownBlocks`: metadata sequence is permuted block-wise and cross-tabulated (NMI/ARI) against every frozen cluster assignment | Currier-language/scribal-hand identity | yes — the sole permuted/tested variable | no |

Package doc (`types.go`): "performs the confirmatory, whole-search-space
multiple-comparison correction for the association between frozen blind
distributional regimes and Currier/hand metadata... only the metadata
labels are permuted." The frozen clustering side is never recomputed.

**NOT_APPLICABLE reason:** performs the whole-search-space (5 window sizes ×
3 methods × 14 K values, ×2) multiple-comparison-corrected test of
association between real Currier-language/scribal-hand identity and blind
distributional clustering; Currier and hand are Voynich-manuscript-specific
external labels with no generic-corpus analogue — substituting any other
partition would test a different hypothesis, not this one.

### 21 — conditional-regime-analyze (Class A)

Package doc (`types.go`): "...after conditioning on Currier and Davis hand,
is there still reproducible distributional structure in the corpus?"
Currier×hand is not a segmentation convenience — it is the confound the
whole study exists to control for.

- Part A (`metadata.go:buildAllBlocks`/`buildBlocks`): builds physical
  blocks strictly from `currier[i]`/`hand[i]` labels, then runs
  global-regime-analyze's own clustering *inside* each metadata-homogeneous
  block and permutation-tests it (Null A shuffles within blocks, preserving
  the metadata partition exactly). The claim — "structure survives after
  conditioning on the known Currier/hand confound" — is meaningless without
  genuine Currier/hand identity.
- Part B (`nullmodels.go`/residual sweep): removes "the expected per-class
  signature" (the Currier/hand class mean) before clustering the residual,
  then checks the residual clusters' NMI *against the real Currier/hand
  labels*. Metadata identity is the dependent variable being tested for
  independence, not a device.
- Part C (`boundaries.go`): change points are confined to one Currier×hand
  block specifically so the detector can never cross a metadata transition.

**NOT_APPLICABLE reason:** tests whether distributional/regime structure
remains after conditioning on real Currier/hand scribal metadata; the class
identity is the experimental variable, not a partitioning device — no
generic corpus-only analogue tests the same hypothesis.

### 22 — residual-diagnostic-analyze (Class A)

Package doc: "explains metadata association in the frozen residual
clustering produced by conditional-regime-analyze." It hard-requires stage
21's Currier/hand-conditioned outputs (`residual_cluster_assignments.tsv`,
`residual_metadata_association.tsv`), and its entire output surface is
metadata-association diagnostics: `representationRow{Silhouette,
CurrierNMI, HandNMI, JointNMI, BlockNMI}`, cross-validated
`currier`/`hand`/`joint` classifiers scored against real truth labels, and a
decision summary reading "metadata survives leakage-safe whitening, so the
current conditioning model remains insufficient."

`BlockNMI`/physical-block fields exist alongside the metadata-NMI fields
only as an internal control (showing block structure alone isn't driving
the metadata NMI) — repurposing them for a generic version would silently
change the hypothesis from "is there Currier/hand signal" to "is there
arbitrary-block signal", a different, already-answered-elsewhere question.

**NOT_APPLICABLE reason:** diagnostic exists solely to explain/rule out
Currier/hand/joint association in stage 21's residual clusters; without real
metadata truth labels there is nothing to diagnose.

### 23 — token-relation-validate (mixed B/C + A)

**Field table**

| field | source | consumed where | role |
|---|---|---|---|
| currier, hand (raw) | `token_metadata_map.tsv` | `load.go:ExtractBlocks` | build contiguous same-Joint runs = physical Block |
| `Block.Currier/Hand/Joint` | `ExtractBlocks` | copied into `DirectionBlock`/`ProfileBlock`; counted into `CurrierClasses`/`Hands`/`JointClasses` | opaque partition id + eligibility-gate counts, never read as values in a formula |
| crossCurrier/crossHand/withinCurrier/withinHand | derived from `MetadataTransfer` rows and `CurrierClasses==1`/`Hands==1` | `Classify`/`ClassifyDetailed` (`metrics.go`) | explicit metadata-conditioned label (UNIVERSAL/CURRIER_SPECIFIC/HAND_SPECIFIC) |

Core math (`directionForBlock`, `profileForBlock`, `buildLocalProfiles`,
`mergeLocalProfiles`, `PermuteWithinBlocks`, BH FDR) reads only `Token.Text`
and treats Block as an opaque, pre-partitioned span — Class B/C. The
`Classify`/`ClassifyDetailed` labeling step is Class A.

**Generic strategy implemented:** `internal/tokenrelationvalidation` gains
`Config.Generic`. When set, `loadGenericMetadata` builds Tokens/Blocks from
`internal/genericsegmentation` instead of reading `-token-metadata-map`
(Currier = deterministic Group label, Hand = `genericsegmentation.Sentinel`,
never a fabricated value). `buildMetadataTransfers` computes only a single
`"group"` dimension in generic mode (never `"Currier"`/`"hand"`), and a new
`ClassifyGeneric` function reasons about that one dimension, emitting
`GROUP_CONSISTENT`/`GROUP_LIMITED`/`BLOCK_SPECIFIC`/`WEAK` — vocabulary that
never borrows "CURRIER_SPECIFIC"/"HAND_SPECIFIC"/"UNIVERSAL", since those
claim a real manuscript covariate a generic corpus does not have. The
underlying `RelationSummary` numeric fields (enrichment, sign-consistency,
transfer success, `JointClasses`/`EligibleBlocks` gates) are computed
identically to IVTFF mode — they were already group-only.

### 24 — replicated-local-structure-audit (mixed B/C + A)

"Replicate" as implemented: a `distanceCandidate`/`sequenceCandidate` tested
for whether its effect holds across independent physical blocks (contiguous
Currier×hand runs) via shuffle/markov null models — Class B/C. But
`WithinCurrier/WithinHand/CrossCurrier/CrossHand/WithinJoint/CrossJoint` are
genuinely surfaced in the report (`write.go`'s
`distance_profile_replication_status.tsv` and the `familySummary` rollup),
and `run.go`'s `Status` decision requires `CrossCurrier && CrossHand &&
CrossJoint` for the top tier (`ROBUST_RELATIVE_REPLICATION`) — Class A.

**Correctness issue found during implementation:** because Hand is a
constant sentinel in generic mode, `CrossHand` would always be structurally
false, making the ROBUST tier permanently unreachable and silently
mislabeling every generically-replicated finding as
`METADATA_LIMITED_REPLICATION` — a label that, read literally, wrongly
implies real hand-specificity. Fixed by branching the Status decision and
`FailedConditions` message on `Config.Generic`: generic mode uses `CrossJoint`
alone, emitting `GROUP_REPLICATED`/`GROUP_LIMITED_REPLICATION` and a
`no_cross_group_transfer` failure condition instead. The
`distance_profile_replication_status.tsv` report also drops the
currier/hand-labeled columns in generic mode (`within_group_transfer`/
`cross_group_transfer` instead), so a reader never mistakes a
structurally-constant placeholder for a real null-metadata result. A second
bug of the same kind: `loadInputs`'s sequence filter looked for stage 23's
`"UNIVERSAL"` classification literally, which never appears in stage 23's
generic-mode output (`"GROUP_CONSISTENT"` instead) — fixed by making the
wanted-classification string mode-aware.

### 25 — higher-order-sequence-validate (Class B, chain-gated)

Block construction (`load.go:loadCorpusAndBlocks`) is the only place
metadata identity string-matches anything. Every actual statistic —
`occurrences.go`, `conditional.go`, `cmi.go:runCMI`, `lobo.go`,
`jackknife.go` — operates exclusively on `Block.Tokens` (position/text),
never on `Block.Currier`/`Block.Hand` values, treating blocks as an opaque
partition. `permuteWithinBlocks` shuffles only within a block; LOBO/jackknife
leave out one block at a time by ID. `crossblock.go`'s
`DistinctJoint`/`CrossJoint` fields are used by `classify.go` as a
replication-breadth gate (`DistinctJoint>=2` required for
`HIGHER_ORDER_REPLICATED`), not an identity comparison.

**Chain dependency:** `loadFrozenCandidates` reads
`strict_replicated_sequences.tsv`/`sequence_null_validation.tsv` from
stage 24's own output — this stage cannot run generically unless stage 24
does too (it now can).

**Correctness issue found during implementation:** `classify.go`'s
`MetadataLimited` computation originally ORed in `DistinctHand<=1`, which is
always true in generic mode (Hand is a constant sentinel) — this would
trivially mislabel any candidate as metadata-limited regardless of its real
`DistinctJoint` count. Fixed: in generic mode, `MetadataLimited` depends on
`DistinctJoint` alone, and the resulting status is `GROUP_LIMITED` rather
than `METADATA_LIMITED`. `higher_order_cross_block.tsv`'s
`distinct_currier`/`distinct_hand`/`cross_currier`/`cross_hand` columns are
replaced with `distinct_group`/`cross_group` in generic mode.

### 26 — positional-continuation-validate (Class B + separate lexical dependency)

`Line`/`TokenIndexLine` are pure per-line token-index arithmetic — corpus/
line-relative, not manuscript-specific — confirmed by `load.go` reading
`line_id`/`token_index_in_line` columns generically the same way regardless
of corpus. Block boundaries are the only metadata entry point
(`loadCorpusAndBlocks`), used purely for permutation scope and LOBO folds
(`boundary.go:permuteIsCheyWithinBlocks`, `stratified.go`, `models.go`'s
leave-one-block-out) — never for identity comparisons.

**Separate finding:** `types.go` hardcodes `FrozenS="s"`, `FrozenAiin="aiin"`,
`FrozenChey="chey"` — literal Voynichese tokens a human selected after
inspecting stage 25's Voynich results. This has zero counterpart in a
generic corpus and is a target-selection dependency, not a metadata one.

**Generic strategy implemented (user-approved substitution):** `FrozenS`/
`FrozenAiin`/`FrozenChey`/`FrozenSAiin` became package-level vars (still
literal `"s"`/`"aiin"`/`"chey"`/`"s aiin"` by default). In generic mode,
`RunAndWrite` calls `resolveGenericTarget`, which reads stage 25's own
generic-mode `higher_order_validation.tsv`, filters
`final_status=="HIGHER_ORDER_REPLICATED"`, and picks the single row with the
lowest `conditional_fdr_q` (ties broken by sequence text) — a ranking lookup
already exposed by stage 25, with no semantic/lexical judgment. Every one of
task23's ~3,000 lines of position/entropy/MI/LOBO/jackknife math is
untouched; it only ever reads these three tokens by name.

### 27 — transition-network-validate (mixed B/C + A)

Per-component audit (`load.go`'s `Joint = Currier+"/"+Hand` block partition
is the sole metadata entry point):

| component | metadata role | class |
|---|---|---|
| transition matrix / edge counts | none (pure token-text bigram counts) | C |
| edge effects | `Block.Currier/Hand/Joint` copied into rows for reporting only | B |
| profiles / profile stability | `Joint` used only as a replication-breadth gate (`JointClasses`) | B |
| entropy | none | C |
| permutation null | shuffles within each block, blocks are the scope not identity | B |
| held-out prediction | leave-one-block-out; no cross-Currier/hand split exists | B |
| model order | same per-block LOBO loop | B/C |
| metadata transfer (`dim="currier"`/`"hand"`) | explicitly groups edge effects by raw Currier/hand label and compares named groups | **A** |
| metadata transfer (`dim="joint"`) | redundant with the block-group partition itself | B |
| `Replicated` flag | `JointClasses>=2` replication gate, not an identity test | B |

**Generic strategy implemented:** `Config.Generic` branches
`loadCorpusAndBlocks` to the shared generic segmentation, and
`computeGraphDiagnostics`'s `dim` loop is restricted to `{"joint"}` in
generic mode — `dim="currier"`/`"hand"` are skipped entirely (a real
`NOT_APPLICABLE` for that one sub-component, not a degenerate run of it
against constant-sentinel data). `analyze.go:classify`'s `JointClasses<2`
branch is relabeled `GROUP_SPECIFIC` instead of `METADATA_SPECIFIC` in
generic mode; the gate itself is unchanged since it already reasoned about
the joint/group partition alone.

## Phase 5 — hypothesis equivalence (Class B stages)

### 23 token-relation-validate

- **Original Voynich hypothesis:** does a token-pair relation (directional
  order, distance profile, structural similarity) hold consistently across
  independent Currier/hand-defined manuscript blocks, and does it transfer
  across real Currier/hand states?
- **Role of metadata:** Currier×hand defines the physical block partition
  used for permutation nulls and leave-one-block-out transfer; the
  classification layer additionally asks whether the relation is
  Currier/hand-specific.
- **Generic hypothesis:** does the same token-pair relation hold
  consistently across independent deterministic resampling blocks of a
  plain corpus, and does it transfer across the corpus's own generic
  resampling groups?
- **Why the statistical test remains meaningful:** the block-level
  enrichment/sign-consistency/transfer-success math never depended on real
  Currier/hand identity — only on having a genuine, non-overlapping
  partition of the corpus, which the generic segmentation supplies with the
  same statistical properties (deterministic, corpus-size-scaled, ≥2
  independent groups guaranteed).
- **Legitimate comparison:** whether relation-level statistics (enrichment,
  sign-consistency, transfer success) behave similarly in magnitude/shape
  between the Voynich corpus and a natural-language corpus of comparable
  scale.
- **Not legitimate:** treating a generic `GROUP_CONSISTENT` result as
  evidence about Currier-language or scribal-hand independence — the
  generic groups are statistical folds, not authorship/language classes.

### 24 replicated-local-structure-audit

- **Original hypothesis:** does a distance-profile or sequence-recurrence
  candidate replicate across independent Currier/hand-defined manuscript
  regions, including transfer across real Currier/hand states?
- **Role of metadata:** defines the block partition for null models; the
  status tier additionally requires cross-Currier-and-hand transfer.
- **Generic hypothesis:** does the same candidate replicate across
  independent deterministic resampling blocks/groups of a plain corpus?
- **Why it remains meaningful:** the markov/shuffle null-model math and
  distance-profile similarity computations never read block identity, only
  block membership.
- **Legitimate comparison:** replication rate/robustness of local structure
  across generic corpora of different genres/sizes vs. the Voynich corpus.
- **Not legitimate:** treating `GROUP_REPLICATED` as equivalent evidence to
  `ROBUST_RELATIVE_REPLICATION` — the latter specifically established
  independence from two distinct real covariates (language and scribe); the
  former only established independence from an arbitrary resampling fold.

### 25 higher-order-sequence-validate

- **Original hypothesis:** does a higher-order (second-order) sequence
  dependence A→B→C reproduce across independent physical blocks and across
  ≥2 distinct Currier×hand metadata classes, surviving jackknife and LOBO?
- **Role of metadata:** defines physical blocks (fine partition) and the
  coarser Joint grouping whose distinct-count gates the top confirmatory
  tier.
- **Generic hypothesis:** does the same A→B→C dependence (read from stage
  24's own frozen candidate list, never rediscovered) reproduce across
  independent generic blocks and across ≥2 distinct generic resampling
  groups?
- **Why it remains meaningful:** the CMI/conditional-probability/jackknife/
  LOBO math is entirely block-membership-based; nothing here needs
  metadata semantics, only a genuine multi-group partition.
- **Legitimate comparison:** whether the strength/robustness of the CMI
  effect and its cross-block replication rate is comparable between corpora.
- **Not legitimate:** treating generic `HIGHER_ORDER_REPLICATED` as
  "replicates across independent hands/languages" — it only established
  replication across independent statistical folds of one corpus, a weaker
  claim than the Voynich case's cross-metadata replication.

### 26 positional-continuation-validate

- **Original hypothesis:** does the positional concentration of the frozen
  Voynich finding "s aiin → chey" reflect a genuine position-conditioned
  continuation constraint (vs. a general property of "aiin", a boundary
  formula, or a single-block artifact)?
- **Role of metadata:** defines blocks/positions for the same permutation/
  LOBO scaffolding as stage 25; separately, the specific token triple tested
  is a frozen literal chosen by a human from the Voynich case.
- **Generic hypothesis:** does the same class of position-conditioned
  continuation constraint hold for the corpus's own top-ranked higher-order
  candidate (deterministically selected by stage 25's own ranking, not by
  semantic judgment)?
- **Why it remains meaningful:** every formula (position categories,
  mutual information, LOBO model comparison, jackknife) is generic over
  which three tokens are being tested; substituting the target changes only
  which sequence is examined, not the test.
- **Legitimate comparison:** whether position-dependence of a corpus's own
  strongest higher-order finding is comparable in strength/shape to the
  Voynich case's "s aiin → chey" finding.
- **Not legitimate:** treating a positive generic-mode result as directly
  confirming or refuting the specific Voynich "s aiin → chey" finding — they
  are different sequences by construction; the comparison is about whether
  the same *kind* of position effect exists, not about the same tokens.

### 27 transition-network-validate

- **Original hypothesis:** is the adjacent-transition backbone
  (transition matrix, edge effects, entropy reduction, held-out
  prediction) stable across independent Currier/hand-defined blocks, and
  does it also transfer across real Currier/hand states specifically?
- **Role of metadata:** defines block partition/permutation scope for the
  entire backbone analysis; the metadata-transfer sub-component additionally
  tests cross-Currier/cross-hand transfer specifically.
- **Generic hypothesis:** is the same backbone stable across independent
  deterministic resampling blocks of a plain corpus?
- **Why it remains meaningful:** transition-matrix/edge-effect/entropy/
  permutation-null/held-out-prediction/model-order math is entirely
  block-membership-based, never metadata-identity-based.
- **Legitimate comparison:** backbone stability/entropy-reduction magnitude
  across corpora of comparable scale.
- **Not legitimate:** the `dim="currier"`/`"hand"` metadata-transfer
  component has no generic equivalent at all and is not computed in generic
  mode — there is nothing to compare here between corpora.

## Phase 4 — generic segmentation specification

One shared package, `internal/genericsegmentation`, supplies every Class B/C
stage's block/group partition (task43 explicitly forbids a separate
segmentation per stage). Algorithm, in full:

1. `ReadCorpus` tokenizes a plain-text corpus exactly like every existing
   stage already does (scan lines, split each line on whitespace),
   additionally recording each token's natural line number.
2. `L` = number of distinct natural corpus lines. `L<2` → `ErrNotEnoughData`.
3. `targetBlocks = clamp(round(sqrt(L)), 8, 64)` — enough fine blocks for a
   permutation/leave-one-block-out null to be meaningful, without
   degenerating into one-line slivers for large corpora. Not derived from
   any Voynich fact.
4. Lines are grouped into contiguous fine blocks of `ceil(L/targetBlocks)`
   lines each; an undersized trailing remainder (less than half a full
   block) merges into the previous block. A block never splits a natural
   line. `fineBlockCount<2` → `ErrNotEnoughData`.
5. `K = min(4, fineBlockCount)` coarse resampling groups — "4" is a fixed
   design constant (ordinary k-fold practice), not derived from the corpus.
   Fine blocks are assigned round-robin (`block index mod K`), so adjacent
   fine blocks always land in different groups — this makes every
   downstream package's existing "maximal contiguous run of one Joint
   value" block-builder produce exactly one fine block per physical block,
   completely unchanged.

Properties: deterministic (pure function of line count and content),
language-agnostic (no lexical/semantic signal used), corpus-size-aware
(scales with `sqrt(L)`), no Voynich constants, no punctuation requirement,
reproducible, independent of worker count and execution order (single-pass,
no concurrency), and natural-line-preserving. The coarse `Group` label is
explicitly a statistical resampling fold — never presented as, or named
after, a fabricated hand/Currier/folio.

Each of the five Class B/C stages substitutes `Currier = Group`, `Hand =
genericsegmentation.Sentinel` (a fixed constant, never a fabricated
per-token identity, chosen because every existing `known()`/block-builder
gate would otherwise reject an empty Hand value), `Joint = Group + "/" +
Sentinel` — every existing per-package block-construction function
(`ExtractBlocks` and equivalents) runs completely unmodified on these
synthetic values.

## Phase 14 — distribution opportunities (backlog only, not implemented)

Task43 explicitly excludes new distribution work. Natural work units
identified for a future task:

- Stage 23/24/27: within-block permutation nulls (`PermuteWithinBlocks`,
  shuffle/markov replicates, edge permutation nulls) are per-candidate,
  independent, and already the kind of unit `structural-projection-analyze`
  and `conditional-regime-analyze` distribute today.
- Stage 25/26: primary/secondary permutation families and positional/
  stratified permutation tests are per-candidate and per-position-variable
  respectively.
- Stage 24/27: leave-one-block-out folds are naturally parallel across
  blocks.

None of this was implemented here; Task28-35's existing local/process/remote
executor infrastructure is the natural fit if a future task takes it on.

## What was deliberately not done

- No fake hand/Currier/folio metadata was fabricated anywhere.
- No scientific hypothesis was silently substituted for another — every
  substitution (stage 26's target triple; the Class-A-only vocabulary
  renames in 23/24/25/27) is documented above with what remains and does
  not remain a legitimate cross-corpus comparison.
- No formula, threshold, RNG definition, or normalization changed in any
  package; the only new definitions are the generic segmentation algorithm
  itself (documented above and in `internal/genericsegmentation`'s package
  doc) and the generic-mode-only classification vocabulary.
- Stages 19–22 were not given an artificial generic analogue merely to grow
  the enabled-stage count.
- `experiments/voynich-v1/` was not touched; every touched package's IVTFF
  code path is reached only when `-generic-corpus` is absent (default
  `false`), and `pipeline-orchestrate/generic_test.go`'s
  `TestVoynichExecutionPlanRegression` continues to assert the Voynich
  argument lists are byte-for-byte unchanged.
