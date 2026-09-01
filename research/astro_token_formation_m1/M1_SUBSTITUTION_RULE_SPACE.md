# M1 frozen global-substitution rule space

## Freeze boundary

This document, `M1_GRAPHEME_INVENTORY.tsv`, and `M1_SEARCH_CONFIG.yaml` were
frozen before the M1 production search. M1 reuses `ASTRO_TERM_CORPUS.tsv` and
`ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv` byte-for-byte. No corpus extension is used.
M0 remains an unchanged baseline.

## Segmentation

Source forms first receive one of the complete 128 M0 lexical/ending/
orthographic/vowel pipelines. Source strings are then segmented greedily with
the eight priority digraphs in the inventory; remaining Latin letters are
single units. EVA labels are segmented greedily with the ten repository-frozen
composites (`cth`, `ckh`, `cph`, `cfh`, `iin`, `ain`, `ch`, `sh`, `ee`, `in`);
remaining observed Astronomical EVA graphemes are single units. Segmentation
never changes during search or controls.

The five M0 abbreviation choices are preserved. Conceptually they are applied
after substitution by removing the output contribution of the final one/two
source units or retaining the contribution of the first three/four source
units. This is exactly computable by selecting those source units before
substitution; no unobserved target suffix is guessed.

## Global mapping constraints

Each retained source unit has exactly one output: one or two target EVA units.
Different source units may share an output. A repeated source unit must have
the same output everywhere. Empty outputs, arbitrary insertions, transposition,
positional mappings, object-conditioned mappings, and label exceptions are
forbidden.

For a term/label pair, every length-compatible 1-or-2-unit alignment is
enumerated. An alignment induces a partial substitution table and is rejected
if repeated source units conflict. Anonymous label↔object pairs can coexist
only if their tables agree exactly on every shared source unit and no term or
label is reused.

## Bounded deterministic optimisation

The full substitution-alphabet space is not exhaustively enumerable. For each
of the 640 M0 preprocessing/abbreviation pipelines,
`DETERMINISTIC_COMPATIBILITY_BEAM_V1` visits TRAIN labels in increasing
edge-count order. At every step it retains the 64 states with greatest matched
label count, then regularized score, then canonical signature. Four finalists
per pipeline enter the global comparison. Pipelines producing an identical
preprocessed corpus are deduplicated in favour of the least-complex
representative. This is a reproducible bounded search and makes no global
optimality claim.

The regularized TRAIN score is frozen as:

`coverage - .010*mappings - .005*one_to_two - .005*digraph_mappings -`
`.005*many_to_one_excess - .010*M0_complexity - .005*ambiguity_excess`

HELD_OUT is never part of optimisation. All mappings and rules are frozen
before it is scored.

## Complexity and ambiguity

`total_complexity` is the mapping count plus M0 complexity, one-to-two count,
digraph-mapping count, many-to-one excess, and any positional-rule count
(frozen at zero). `ambiguity_excess` is the number of compatible object choices
beyond one, summed over explained TRAIN labels and divided by TRAIN size.

For every reported model the output enumerates every compatible label/object
edge and marks whether it occurs in at least one maximum anonymous assignment.
Assignment multiplicity is counted exactly up to the frozen cap of one billion;
a capped result is explicitly marked.

## Candidate bands and null

Thresholds and penalties are exactly those in `M1_SEARCH_CONFIG.yaml`. A strong
candidate needs TRAIN ≥.70, HELD_OUT ≥.50, null advantage ≥.10, familywise
empirical p ≤.05, mapping size ≤18, and total complexity ≤25. Partial and
predictive gates are separately frozen in the config.

Each of the three controls has 100 replicates and reruns the same complete
bounded optimiser. The replicate statistic is the maximum regularized score
over M1, not the score of a mapping learned from the observed data. Random EVA
sets are class- and target-unit-length-matched occurrence samples; shuffled
terms preserve words and character multisets; pseudodictionaries preserve
class, form count, word shape, and empirical source-letter frequencies. Seed:
`20260901`. The p-value resolution limitation is `1/101 ≈ .0099`.

## Interpretation boundary

A positive result establishes only compact global grapheme-substitution
compatibility beyond these null searches. It does not identify a star, language,
plaintext, or decipherment. A negative result rejects only this frozen bounded
algorithm and representation, not all possible substitution systems.
