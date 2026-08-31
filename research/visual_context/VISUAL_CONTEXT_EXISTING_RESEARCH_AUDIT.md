# Existing research audit for visual context

This audit classifies evidence against the visual-context question, not against the broader Voynich hypotheses. Detailed machine-readable rows are in `VISUAL_CONTEXT_EXISTING_EVIDENCE.tsv`.

## Coverage by requested family

| Requested family | Principal artifacts | Granularity | Classification | Reuse decision |
|---|---|---|---|---|
| Phase I structural fingerprint | `docs/phase1/PHASE1_RESEARCH_REPORT.md`, `experiments/voynich-v1/outputs/` | corpus, token windows, line, page | INDIRECTLY_RELEVANT | interpretation and provenance |
| Token frequency / position | Task59, Task79 line profiles and occurrence metadata | token, line | INDIRECTLY_RELEVANT | aggregate frozen Task79 line values |
| Begin/end | baseline `begin_end_*`, Task79 boundary metrics | corpus/candidate, line | INDIRECTLY_RELEVANT | Task79 line profiles only |
| Sequence / n-gram | baseline `sequence_analysis*`, higher-order and transition-network outputs | corpus/block/candidate | INDIRECTLY_RELEVANT | not safely page-aggregatable except frozen line entropy/repetition |
| Distance context | baseline `distance_context_*` | corpus/candidate | INDIRECTLY_RELEVANT | no page-keyed raw profile; do not relabel |
| Structural normalization/validation | baseline stages 6-10 | corpus/token family | CONTROL_VARIABLE | establishes reliability, not section effect |
| Profile stability/reliability | baseline `structural_profile_stability.yaml`, `structural_reliability.yaml` | corpus/fold | CONTROL_VARIABLE | supports metric interpretation |
| Metadata validation | baseline metadata/cluster/conditional/residual outputs | window/boundary | CONTROL_VARIABLE | reuse Currier/hand/quire results |
| Page/section analysis | Task64/65 and Task79 | line/page/section | DIRECTLY_TESTS_VISUAL_CONTEXT | primary frozen evidence |
| Phase II | fingerprint-v2 Task73/79/79c/83b and Task83 synthesis | corpus/line/page/metadata | DIRECT + CONTROL | reuse Task79; note later CS4 was N/A in Task83b because IVTFF was not bound there |
| Comparative Notation Study | VM reference and metric registry | corpus | NOT_RELEVANT to section contrast | definitions/provenance only |
| Structure catalog | `research/structure_catalog/*` | token/line/folio/section | INDIRECTLY_RELEVANT | section rule inventory only; token-level tests risk pseudoreplication |

## What is already known

Pages and folios are measurably heterogeneous. Task64 found a page-local effect; Task65 reproduced it, mapped it by section, and found between-section distance greater than within-section distance in seven of eight `$I` classes. Task79 found section-associated line-length variance and token-length/regime association. Currier, hand and section are themselves strongly associated and manuscript regions are contiguous. Existing results therefore directly support a descriptive section fingerprint but do not isolate visual context.

Quire and folio boundaries have been tested as metadata boundaries, line and paragraph position have extensive direct tests, and Currier/hand have both global and conditional/residual analyses. No existing analysis uses page-specific image content. No available artifact contains image bounding boxes or object-count descriptors.

## Fingerprint dimensions and aggregation validity

| Dimension | Existing page-safe values | Valid aggregation | Caveat |
|---|---|---|---|
| Symbol/token | token/character counts, mean token length, vocabulary/diversity, token entropy | token-weighted or recomputed sufficient-statistic aggregation | diversity is size-sensitive |
| Sequence | exact/near adjacent repetition, transition entropy | pool transition counts/rates, otherwise weighted descriptive mean | sparse pages; no claim of independent transitions |
| Line/block | token count, character count, line count, line-length CV, boundary token lengths | page summary from frozen line profiles | labels/radial loci have different layout |
| Context | Task64 page profile and local-effect estimates | reuse Task65 outputs directly | most distance-context candidates lack page-keyed raw data |
| Document metadata | page, leaf, side, order, quire, `$I`, Currier, hand | direct join from frozen occurrence metadata | missing Currier values and strong aliasing |

The aggregator must retain support counts and missingness. It must not average p-values, treat tokens/lines as independent page observations, or introduce a cross-family composite.

## Existing direct results

Task65 `METADATA_EFFECTS.tsv` reports, by section, `BetweenMeanDistance > WithinMeanDistance` for Astronomical, Biological, Cosmological, Herbal, Pharmaceutical, Stars and Zodiac. Text has 1.7788 versus 2.2316 and fails the prediction. This is distributional evidence across pages; it is not unanimity across families.

Task79's primary registry reports:

- `HR1_SECTION_VARIANCE_SHARE = 0.125259`, permutation `p=0.000999`, BH `q=0.001306`, supported;
- `LC5_IVTFF_I_NMI = 0.008452`, `p=0.000999`, `q=0.001306`, supported;
- `HR6_CURRIER_SECTION_NMI = 0.648621`, `p=0.000999`, `q=0.001306`; audit shows its operational field is source `$C`, not Currier A/B (`$L`), so it is a general metadata-confounding warning rather than a valid A/B estimate;
- `PF2_FOLIO_COHERENCE = 0.173040`, `p=0.000999`, `q=0.001306`, supported.

Later Task83b artifacts mark `cs4/family-section` and `cs5/local-adjacency-x-regime` `NOT_APPLICABLE` because that execution did not bind IVTFF metadata. That absence is not negative evidence and does not erase the valid Task65/79 results.

## Limits

Task65's section rows use a fixed five-component profile, while Task79 section metrics emphasize line length and coarse token-length/regime classes. They do not cover every Phase I metric family at page scale. One transcription is aligned. Small classes, multi-panel foldouts, page-size imbalance, and near-aliasing with quire/hand reduce effective sample size. All stronger conclusions require either estimable overlap among confounders or new independently sourced visual descriptors.
