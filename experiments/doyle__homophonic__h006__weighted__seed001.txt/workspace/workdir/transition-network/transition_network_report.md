# Transition network validation

## Global accounting

Unique tokens: 10671; eligible tokens: 648; observed adjacent eligible edges: 10675; testable edges: 4096. Edges in >=2 blocks: 3262; in >=3 blocks: 3044. Cross-joint: 3189; cross-Currier: 3189; cross-hand: 0.

FDR-significant preferred: 27; depleted: 0. Strict backbone preferred: 15; depleted: 0. Metadata-specific: 0; block-concentrated: 2; significant unstable: 10.

Replicated outgoing profiles: 2; incoming profiles: 3.

## Predictive validation

Mean held-out M0-M1 log-loss delta: 0.280576 (positive favors knowledge of A). Mean held-out M1-M2 delta: -0.439676; M2 wins in 0/64 blocks.

## Required questions

1. **Does A improve held-out prediction?** Yes, by the pre-specified diagnostic.
2. **Do edge preferences/depletions reproduce?** Yes, by the pre-specified diagnostic.
3. **Is there a metadata-independent backbone?** Yes, by the pre-specified diagnostic.
4. **Are continuation constraints reproducible?** Yes, by the pre-specified diagnostic.
5. **Does second order materially improve?** No convincing evidence under the pre-specified criteria.

## Outcome

`PROFILE_BACKBONE`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.

## Sequence reconstruction diagnostic

Preferred-backbone edge coverage: 0.00140515; token coverage: 0.0385802; observed transition coverage: 0.00386607. A deterministic frequency-matched graph with the same number of edges covers 0.00386607 of transitions.
