# Visual-context / external-memory experiment review

Date: 2026-08-30

## 1. Research question

Does the textual structural fingerprint vary with an independently assigned visual context? `H_VISUAL_SECTION` asks for systematic differences among the frozen IVTFF illustration/content classes. `H_VISUAL_LOCAL` asks whether image-only, page-specific descriptors explain additional within-section variation. Neither hypothesis is a decipherment claim or a causal claim.

## 2. Existing evidence

The repository already contains direct Level-A evidence. Task65 computed page profiles with the pre-existing Task64 profile distance. In `experiments/local-regime-topology-v1/METADATA_EFFECTS.tsv`, between-section distance exceeds within-section distance for seven of eight classes; Text is the exception. Task79 independently found a section share of line-length variance of 0.125259 (`p=0.000999`, FDR `q=0.001306`) and an `$I` association with token-length/regime class of NMI 0.008452 (`p=0.000999`, `q=0.001306`). These are section effects, not page-image effects.

Task79 also measured strong dependence among known metadata: its registry calls `HR6_CURRIER_SECTION_NMI=0.648621`, but audit of the frozen occurrence rows against the source shows that its operational “Currier” field is IVTFF `$C`, not Currier A/B (`$L`). It remains a confounding warning but must not be cited as an A/B estimate. Task65 correctly uses `$L` and reports overlapping descriptive variance shares for section, hand and Currier. Therefore none of the existing positive section results identifies a visual effect beyond Currier, hand, quire and manuscript position.

## 3. Existing tools

- `internal/metadatavalidation`: strict IVTFF/canonical-corpus alignment; page, folio, quire, hand, section, locus and position metadata.
- `internal/lineregime`: Task64 primary page-profile implementation and distance.
- `research/phase1/local-regime-topology-analyze`: Task65 metadata effects, conditioned decay, hierarchical variance and block-aware diagnostics.
- `cmd/metadata-validate`, `cmd/cluster-metadata-global`, `cmd/conditional-regime-analyze`, `cmd/residual-diagnostic-analyze`: Currier/hand/quire and residual tests.
- `cmd/fingerprint-v2-analyze`: page/hierarchy metrics and permutation nulls; its frozen Task79 output includes line profiles and occurrence metadata.
- Comparative Notation Study registries and VM reference outputs: useful for metric definitions and provenance, but corpus-level only and not direct page/section evidence.

## 4. Existing results directly reusable

- `experiments/fingerprint-v2-task79-v1/canonical-out/line_profiles.json`: line-level token, sequence, repetition, entropy and boundary summaries with folio/section/Currier/scribe.
- `experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl`: aligned page taxonomy and confounders for every token.
- Task65 `METADATA_EFFECTS.tsv`, `HIERARCHICAL_VARIANCE.tsv`, `LOCAL_EFFECT_BY_REGION.tsv`, `METADATA_CONDITIONED_DECAY.tsv` and `CHANGE_POINT_METADATA_OVERLAP.tsv`.
- Task79 `metric_registry.json`, `ivtff_metadata_audit.json` and `stability_matrix.json`.
- Baseline Stage 1-28 raw TSV/YAML outputs for interpretation and audit, but most are corpus-, window- or candidate-level and cannot validly be relabelled as page fingerprints.

The new analysis may aggregate the two frozen Task79 datasets. It must not rerun Task59-65 or construct a parallel fingerprint.

## 5. Known section/Currier/scribe effects

Task65 page profiles show section separation, but its one-factor variance shares overlap: MeanTokenLength Currier 0.279, hand 0.383, section 0.373; GiantFraction Currier 0.219, hand 0.273, section 0.337. Task79 confirms a strong `$C`/section association under a mislabeled field; Task65 supplies the valid Currier-A/B comparison from `$L`. Baseline metadata validation finds Currier association with blind regimes and weaker hand association; conditional residual structure beyond Currier/hand is weak/inconclusive. Quire boundaries track physical manuscript regions and are a necessary control.

## 6. Available visual metadata

The independent repository-local taxonomy is IVTFF `$I`: A Astronomical, B Biological, C Cosmological, H Herbal, P Pharmaceutical, S Stars, T Text, Z Zodiac. It is recorded per page/panel and was created independently of this experiment's textual metrics. All 227 page IDs aligned by Task79 have `$I`; no pixel coordinates, object counts, image area, plant counts, figure counts, vessel counts or other page-specific image descriptors are present.

## 7. Confounders

Controls available without external data are Currier language (`$L` in the source; normalized in Task79), scribal hand (`$H`), quire (`$Q`), documentary page/folio order, token count and line count. Recto/verso and multiple panels of one physical leaf create non-independent observations. Section is concentrated in contiguous quires and hands, so categorical quire control may make section coefficients rank-deficient. This is an identifiability result, not a reason to drop quire post hoc.

## 8. Which parts of H_VISUAL_SECTION are already tested

Directly tested: the Task64 five-dimensional page profile has smaller within-section than between-section distances for seven classes; Task79 finds section-associated line length and token-length/regime structure under its frozen nulls. Indirectly tested: folio coherence, recto/verso coherence, locus-type effects, local regime persistence and section-specific local-effect estimates. These establish Level A only.

## 9. Which parts remain untested

- A family-by-family page-level comparison using the already frozen Task79 line profiles.
- Incremental section information after simultaneous control for Currier, hand, quire, position and size.
- Conservative classification with physical-leaf/quire grouping.
- All of `H_VISUAL_LOCAL`: there are no page-specific image-only descriptors.

## 10. Minimal missing tooling

One deterministic research aggregator/joiner is needed. It will read frozen Task79 outputs, emit a page-level dataset, contingency tables and family-separated section diagnostics. It may diagnose model rank/aliasing and run fixed-seed group-preserving permutations. It will not become a baseline pipeline stage, compute a new composite similarity score or rerun underlying analyzers.

## 11. Proposed statistical design

Primary reuse result: Task65's preselected Euclidean distance on its Task64 page profile. Supplemental analysis: for each frozen Task79 scalar independently, report within-section and between-section absolute differences, their difference and ratio, with section labels permuted at physical-leaf blocks. The inferential support floor is frozen at at least 20 tokens and two non-empty lines per page; all pages remain in descriptive/taxonomy outputs. Report effect size before p-values and adjust the fixed metric-family tests with Benjamini-Hochberg.

Confounder analysis first audits contingency and design-matrix rank. Compare reduced (Currier + hand + quire + log token count + line count + page position) and full (+ section) linear models per metric only when section adds estimable columns. Permutations preserve physical-leaf grouping. A classifier is secondary and is omitted if leave-one-quire-out folds lack identifiable train/test classes or confounding makes its interpretation invalid.

## 12. Expected computational cost

Reading about 39,380 occurrence records and about 5,385 line profiles plus 1,000 lightweight grouped permutations per retained scalar is expected to take seconds to low minutes and under 1 GiB. No Stage 1-28 rerun or remote execution is justified.

## 13. Blocking data gaps

The blocking gap for Level C is the complete absence of frozen page-specific visual descriptors. Creating them would require a separately frozen image-only schema, scholarly metadata acquisition or blinded image annotation. The current task does not authorize inventing those labels from textual outcomes. Cross-transcription robustness is also unavailable locally.

## 14. Recommendation

**MODIFY / PROCEED WITH LEVEL-A REUSE AND IDENTIFIABILITY AUDIT.** Complete the lightweight page aggregation and simultaneous-confounder feasibility test. Do not claim Level B unless section is estimable beyond the frozen controls. Do not implement the within-section image permutation test; report it as `NOT_COMPUTED_DATA_UNAVAILABLE` until independent page-specific descriptors exist.
