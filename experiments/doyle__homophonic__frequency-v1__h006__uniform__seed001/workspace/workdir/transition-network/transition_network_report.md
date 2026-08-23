# Transition network validation

## Global accounting

Unique tokens: 10285; eligible tokens: 643; observed adjacent eligible edges: 11097; testable edges: 4603. Edges in >=2 blocks: 3859; in >=3 blocks: 3335. Cross-joint: 3636; cross-Currier: 3636; cross-hand: 0.

FDR-significant preferred: 10; depleted: 0. Strict backbone preferred: 4; depleted: 0. Metadata-specific: 0; block-concentrated: 3; significant unstable: 3.

Replicated outgoing profiles: 2; incoming profiles: 1.

## Predictive validation

Mean held-out M0-M1 log-loss delta: 0.169801 (positive favors knowledge of A). Mean held-out M1-M2 delta: -0.406898; M2 wins in 0/64 blocks.

## Required questions

1. **Does A improve held-out prediction?** Yes, by the pre-specified diagnostic.
2. **Do edge preferences/depletions reproduce?** Yes, by the pre-specified diagnostic.
3. **Is there a metadata-independent backbone?** Yes, by the pre-specified diagnostic.
4. **Are continuation constraints reproducible?** Yes, by the pre-specified diagnostic.
5. **Does second order materially improve?** No convincing evidence under the pre-specified criteria.

## Outcome

`PROFILE_BACKBONE`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.

## Sequence reconstruction diagnostic

Preferred-backbone edge coverage: 0.000360458; token coverage: 0.0124417; observed transition coverage: 0.0008726. A deterministic frequency-matched graph with the same number of edges covers 0.0008726 of transitions.
