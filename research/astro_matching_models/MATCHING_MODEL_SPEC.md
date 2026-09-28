# Frozen matching-model specification

The historical D1 corpus, 20/6 split, 640 preprocessing pipelines, grapheme
segmentation, substitution layer, beam width 64, mapping complexity, and three
search-level controls are inherited unchanged.

## Applicable models

`ANONYMOUS_SET` is the D1 one-to-one maximum matching baseline, reproduced
from the frozen D1 artifacts and checksums.

`REUSE_ALLOWED` removes only the one-concept-per-label uniqueness constraint.
Each label still needs a complete transformed-form edge. Search score is:

`frozen_M1_regularized_score - 0.05 × reuse_excess`

where `reuse_excess = Σ max(0, uses(concept)-1)`. The value 0.05 is frozen
before the production run. HELD_OUT receives the TRAIN mapping unchanged and
does not exclude TRAIN-used concepts. Null replicates rerun all 640 pipelines.

## Not applicable

`ORDER_PRESERVING`, `CYCLIC_ORDER` (both orientation variants),
`GROUP_CONSTRAINED`, and `SEQUENCE_ALIGNMENT` are not run because the required
Voynich-side structure is absent from independently documented frozen data.
Accordingly ORDER_SHUFFLE_NULL and GROUP_SHUFFLE_NULL are also not applicable.
No arbitrary order, rotation, group, gap penalty, or orientation is introduced.
