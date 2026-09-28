# Hapax–STAR/LABEL analysis plan (preregistered)

Version: `HSL-1.0`. This plan is frozen before any human-verified LABEL/token cohort, enrichment statistic, or new lexicon match is produced.

## Primary hypothesis and units

The directional primary hypothesis is `P(hapax | grouped LABEL token occurrence) > P(hapax | ungrouped LABEL token occurrence)` on f68r1, f68r2 and f68r3. The primary analytical unit is a verified transcription token occurrence, with LABEL as the dependence block and panel as the randomization/cluster stratum. A multi-token LABEL contributes its verified occurrences but is never treated as independent tokens for resampling. The eight-member 3G1 group contributes one LABEL/group observation, never seven pairs.

## Hapax and denominator

Primary `HAPAX` means an exact frozen normalized token key with frequency exactly one in the complete ZL3b occurrence corpus, not within these pages or within the Astronomical section. The authoritative prepared corpus is `data_work/ZL3b-x7.canonical.txt` (SHA-256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`); occurrence identities and metadata come from the byte-registered ZL3b occurrence JSONL. Metadata and comments are not tokens. Upstream token boundaries are retained. Composite EVA sequences (`cth/ckh/cph/cfh/iin/ain/ch/sh/ee/in`) remain frozen atomic keys; readable expansion is display-only. Apostrophe, `?`, `@NNN;`, digits and other retained literal transcription markers are not silently removed. Upstream selection of the first explicit IVTFF alternative is not reopened.

## Mapping and uncertainty

Every one of the 92 LABEL objects must receive exactly one explicit human outcome: `SINGLE_TOKEN`, `MULTI_TOKEN`, `RING_SEQUENCE`, `AMBIGUOUS`, `UNREADABLE`, `NO_TRANSCRIPTION_MATCH`, or `OUT_OF_TRANSCRIPTION_SCOPE`. Empty is incomplete, never NO_MATCH. Primary analysis includes only verified SINGLE_TOKEN/MULTI_TOKEN/RING_SEQUENCE mappings whose selected occurrence IDs exist byte-for-byte in the frozen registry; ambiguous, unreadable and unmatched records are excluded. Ring start and order may be `UNDETERMINED`; no order is invented. Human UI is blind to group membership, origin, hapax/frequency, prior lexicon results and per-token hypothesis metadata.

## Primary test and effect sizes

Report exact numerators/denominators, token-occurrence hapax risk difference, risk ratio and odds ratio with explicit zero-cell handling. Use a one-sided panel-stratified permutation that reallocates grouped status at LABEL-block level while preserving within-panel grouped counts and token bundles; enumerate when feasible, otherwise 100,000 deterministic draws with seed `20260915`. Report a two-sided 95% label-block bootstrap CI stratified by panel (100,000 replicates, seed `20260916`). Because there are only three panels, panel-aware inference is primary and uncertainty is described as weak-cluster pilot inference.

## Label- and group-level analyses

Separately report `LABEL_CONTAINS_HAPAX`, controlling descriptively for token count, token length, confidence and panel; use label-block stratified permutation. Each 3G1 group is one row with its single LABEL, group size, STAR count, initial/new/extended category and human-added membership. No Cartesian expansion is permitted.

## Secondary family and multiplicity

Secondary analyses are: panel-specific effects; new versus unchanged groups; human-added LABEL; the one eight-member group descriptively; thresholds frequency <=2 and <=3; token length; initial/final glyph sequences; and concentration by panel. Apply Benjamini–Hochberg at q=0.05 across this secondary family. The primary test is a single preregistered family and is not combined with secondary p-values. Controls requiring unavailable verified spatial mappings are reported unavailable, not synthesized.

## Sensitivity policies

Run (1) exact readable EVA form, (2) frozen composite normalized key, and (3) exclusion of tokens containing uncertain/unreadable markers or nonalphabetic atoms. An alternative normalization is allowed only if already registered here or in a frozen upstream specification; none may be invented after results. Compare grouped LABEL against same-page ungrouped LABEL primarily, then eligible astronomical and length-matched other-section controls only when reproducibly constructible.

## Lexicon gate and tiers

No lexicon stage runs before complete frozen human mapping and nonempty grouped-hapax subset. Matching rules must be frozen first: Tier 0 exact minimal normalization; Tier 1 only source-documented orthographic variants; Tier 2 only preregistered historically attested Arabic-to-Latin transliterations; Tier 3 constrained structural similarity with frozen threshold, length controls, pseudo-token controls and FDR. If primary enrichment is not supported, any permitted search is `EXPLORATORY`; otherwise it may be `HYPOTHESIS_FOLLOWUP`. Single similarities never establish etymology, a named star, or decipherment.

## Blocking and stopping

Stop before enrichment if any frozen checksum fails, fewer or more than 92 LABEL outcomes exist, a canonical ID/geometry/group membership changes, selected occurrences do not exist, UI blindness fails, or mapping is incomplete. Stop before lexicon matching if provenance or matching-plan freeze is absent. Do not alter hypotheses, hapax definition, thresholds, controls, or normalization after mapping/results inspection.
