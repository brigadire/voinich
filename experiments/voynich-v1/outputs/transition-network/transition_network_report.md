# Transition network validation

## Global accounting

Unique tokens: 8363; eligible tokens: 539; observed adjacent eligible edges: 9065; testable edges: 7836. Edges in >=2 blocks: 5654; in >=3 blocks: 4313. Cross-joint: 5316; cross-Currier: 5316; cross-hand: 5178.

FDR-significant preferred: 7; depleted: 0. Strict backbone preferred: 2; depleted: 0. Metadata-specific: 0; block-concentrated: 2; significant unstable: 3.

Replicated outgoing profiles: 0; incoming profiles: 1.

## Predictive validation

Mean held-out M0-M1 log-loss delta: -0.409849 (positive favors knowledge of A). Mean held-out M1-M2 delta: -0.350895; M2 wins in 0/32 blocks.

## Required questions

1. **Does A improve held-out prediction?** No convincing evidence under the pre-specified criteria.
2. **Do edge preferences/depletions reproduce?** Yes, by the pre-specified diagnostic.
3. **Is there a metadata-independent backbone?** Yes, by the pre-specified diagnostic.
4. **Are continuation constraints reproducible?** Yes, by the pre-specified diagnostic.
5. **Does second order materially improve?** No convincing evidence under the pre-specified criteria.

## Outcome

`EDGE_BACKBONE_ONLY`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.

## Sequence reconstruction diagnostic

Preferred-backbone edge coverage: 0.000220629; token coverage: 0.00742115; observed transition coverage: 0.000483598. A deterministic frequency-matched graph with the same number of edges covers 0.000483598 of transitions.
