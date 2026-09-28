# M2R-v2 model

The authoritative contract is M2R_V2_IMPLEMENTATION_CONTRACT.md. This independent
implementation imports none of the rejected v1 inference functions. Its interface
is two unordered surface bags with neutral IDs. Duplicate spellings are conservatively
collapsed: this can discard genuinely homographic independent examples, but cannot
inflate support. IDs cannot convey assignments. Contradictory uses of one ID fail.

`units` enumerates spans of 1–4 code points plus whole tokens; a content-derived
unit ID includes side, exact surface and positional role. Each unit records separate
occurrence and unique-example support with content IDs and source spans. Rules are
ordered triples [role, source, target]. Role preserving composition is a DAG path
from (0,0) to (source_length,target_length), with explicit unexplained-character gaps.

`validate_rules` forbids forward/inverse collisions; segmentation ambiguity is
represented by competing paths. `transform` returns UNIQUE, AMBIGUOUS or UNEXPLAINED.
`evaluate` records each selected derivation, assignment, all-pair local explanations,
unmatched items and independent support. The final model is admissible only if every
rule has >=3 distinct supporting pairs, terms and labels. Proposals below support
may be evaluated but cannot enter the admissible beam. Heldout applies frozen rules
and keeps training complexity; it never relaxes support thresholds or adds rules.

Limitations: conservative duplicate collapse; positional roles are boundary-based,
not inferred linguistic categories; no context-conditioned exceptions; no claim
that the beam recovers all latent units. Multiple spellings of a concept without
shared surface/ID cannot be recognized as aliases without external information.
