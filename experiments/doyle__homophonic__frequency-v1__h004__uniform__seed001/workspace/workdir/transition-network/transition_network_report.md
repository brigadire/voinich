# Transition network validation

## Global accounting

Unique tokens: 8456; eligible tokens: 616; observed adjacent eligible edges: 11923; testable edges: 5505. Edges in >=2 blocks: 4521; in >=3 blocks: 3994. Cross-joint: 4338; cross-Currier: 4338; cross-hand: 0.

FDR-significant preferred: 101; depleted: 0. Strict backbone preferred: 71; depleted: 0. Metadata-specific: 0; block-concentrated: 1; significant unstable: 29.

Replicated outgoing profiles: 6; incoming profiles: 2.

## Predictive validation

Mean held-out M0-M1 log-loss delta: 0.384885 (positive favors knowledge of A). Mean held-out M1-M2 delta: -0.628988; M2 wins in 0/64 blocks.

## Required questions

1. **Does A improve held-out prediction?** Yes, by the pre-specified diagnostic.
2. **Do edge preferences/depletions reproduce?** Yes, by the pre-specified diagnostic.
3. **Is there a metadata-independent backbone?** Yes, by the pre-specified diagnostic.
4. **Are continuation constraints reproducible?** Yes, by the pre-specified diagnostic.
5. **Does second order materially improve?** No convincing evidence under the pre-specified criteria.

## Outcome

`PROFILE_BACKBONE`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.

## Sequence reconstruction diagnostic

Preferred-backbone edge coverage: 0.00595488; token coverage: 0.154221; observed transition coverage: 0.017177. A deterministic frequency-matched graph with the same number of edges covers 0.017177 of transitions.
