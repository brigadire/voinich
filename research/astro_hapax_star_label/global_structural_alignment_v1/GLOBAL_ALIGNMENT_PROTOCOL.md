# Global structural alignment protocol v1.0

Status: `FROZEN_BEFORE_ANCHOR_SCORING`.

## Scope and blindness

The protocol maps physical canonical LABEL geometry to exact frozen ZL3b
occurrences using page-level sequence constraints. It never modifies spatial
annotations, groups, corpus, legacy mappings, or the failed isolated-crop AI
run. Features and model prompts contain no grouped/ungrouped, group ID, STAR
count, human-added, hapax, frequency, dictionary, lexicon, semantic, or
astronomical fields. 3G1 metadata can be joined only after mapping is frozen.

## Coordinate frame and spatial sequences

Coordinates are the canonical page-pixel frame. Panel centers are fixed at half
the canonical image width and height: f68r1=(1231,1914), f68r2=(1039,1914),
f68r3=(1726.5,1914). Each LABEL uses bbox centre, normalized x/y, width,
height, rotation, radius and angle relative to this fixed centre. No centre is
optimized on anchors. Radius tertiles and 2×2 local quadrants are deterministic
descriptors only.

For each panel, test these predefined orders: top-to-bottom/left-to-right
scan, left-to-right/top-to-bottom scan, clockwise angle, counterclockwise
angle, inner-to-outer radius with angular tie-break, outer-to-inner radius,
angular sector with radial suborder, radial band with angular suborder, and
their ring-partition variants when at least six LABELs have a non-degenerate
radius. Clockwise/counterclockwise and cyclic rotations are alternatives, not
post-hoc choices. A documented traversal is included only if source evidence
exists; otherwise its status is `UNKNOWN`.

## Transcription sequences and alignment

Only `L`, `C`, and `R` label loci are eligible for structural label sequences;
`P` and `Pb` are registered as prose/page text and never mixed into the label
sequence. Singleton `L` loci are one occurrence; multi-token loci retain their
ordered occurrence IDs. `C` sequences remain cyclic with all rotations and
direction uncertainty; `R` sequences remain radial sequences. A monotone
sequence alignment may skip a physical LABEL, skip a transcription occurrence,
or mark an ambiguous/gap operation. No LABEL is assigned merely because a
candidate is first or unique.

The primary f68r1 anchor sequence is the 21 frozen legacy-confirmed mappings.
For a model, anchors are ordered by model rank and compared with their frozen
transcription line rank (f68r1.8 @Ls is rank 0). Scores are Spearman rank
correlation, normalized absolute rank residual, monotone inversion count, gap
count, and model complexity. Model choice is global and receives a fixed
complexity penalty; no individual-anchor exception is allowed.

## Nulls and leave-one-anchor-out

Use deterministic seed `restore-astronomical-row-v1`, 10,000 permutations per
model, preserving the observed transcription rank multiset. Nulls include
random anchor-rank permutations, random spatial-rank permutations, and
rotation-preserving shifts. Empirical p=(1+number of null scores ≥ observed)/
(1+N). LOAO hides each of 21 anchors, fits only the fixed model ordering to the
other 20, predicts the hidden transcription rank by monotone linear
interpolation/extrapolation, and records top-1/top-3, tie, gap and abstention.
The answer is joined only after the prediction is frozen.

## Gates and acceptance

Gate S1 is fixed at LOAO exact ≥16/21, LOAO top-3 ≥19/21, permutation p<0.01,
robustness to the predefined centre/geometry perturbations, and no candidate
presentation-order dependence. Gate S2 is evaluated independently for f68r2
and f68r3 and requires documented/structurally unique traversal, invariant
direction/rotation, limited gaps, stable alternatives, local visual support,
and negative-control abstention. A page without anchors cannot inherit f68r1
accuracy.

Only `MODEL_INVARIANT` and predefined stable `MODEL_STABLE` results can enter a
primary mapping. `MODEL_SENSITIVE`, `NON_IDENTIFIABLE`, `STRUCTURAL_TOP2`,
`STRUCTURAL_AMBIGUOUS`, gaps and abstentions remain unresolved. Local visual AI
is permitted only after global alignment, for at most three structural
alternatives, with two independent context-free passes; it is not run if S1
fails.

Hapax enrichment, frequency computation, lexicon search, semantic reading and
expert verification are prohibited. Thresholds cannot change after LOAO/null
results are observed.

