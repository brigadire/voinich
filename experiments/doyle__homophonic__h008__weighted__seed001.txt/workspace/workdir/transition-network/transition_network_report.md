# Transition network validation

## Global accounting

Unique tokens: 11776; eligible tokens: 660; observed adjacent eligible edges: 9938; testable edges: 3478. Edges in >=2 blocks: 2812; in >=3 blocks: 2298. Cross-joint: 2726; cross-Currier: 2726; cross-hand: 0.

FDR-significant preferred: 8; depleted: 0. Strict backbone preferred: 4; depleted: 0. Metadata-specific: 0; block-concentrated: 2; significant unstable: 2.

Replicated outgoing profiles: 1; incoming profiles: 0.

## Predictive validation

Mean held-out M0-M1 log-loss delta: 0.154087 (positive favors knowledge of A). Mean held-out M1-M2 delta: -0.332138; M2 wins in 0/64 blocks.

## Required questions

1. **Does A improve held-out prediction?** Yes, by the pre-specified diagnostic.
2. **Do edge preferences/depletions reproduce?** Yes, by the pre-specified diagnostic.
3. **Is there a metadata-independent backbone?** Yes, by the pre-specified diagnostic.
4. **Are continuation constraints reproducible?** Yes, by the pre-specified diagnostic.
5. **Does second order materially improve?** No convincing evidence under the pre-specified criteria.

## Outcome

`PROFILE_BACKBONE`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.

## Sequence reconstruction diagnostic

Preferred-backbone edge coverage: 0.000402495; token coverage: 0.0106061; observed transition coverage: 0.00095291. A deterministic frequency-matched graph with the same number of edges covers 0.00095291 of transitions.
