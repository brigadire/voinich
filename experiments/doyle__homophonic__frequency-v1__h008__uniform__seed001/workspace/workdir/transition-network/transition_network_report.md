# Transition network validation

## Global accounting

Unique tokens: 11528; eligible tokens: 667; observed adjacent eligible edges: 10160; testable edges: 3662. Edges in >=2 blocks: 2865; in >=3 blocks: 2420. Cross-joint: 2748; cross-Currier: 2748; cross-hand: 0.

FDR-significant preferred: 0; depleted: 0. Strict backbone preferred: 0; depleted: 0. Metadata-specific: 0; block-concentrated: 0; significant unstable: 0.

Replicated outgoing profiles: 1; incoming profiles: 0.

## Predictive validation

Mean held-out M0-M1 log-loss delta: 0.0493618 (positive favors knowledge of A). Mean held-out M1-M2 delta: -0.296474; M2 wins in 0/64 blocks.

## Required questions

1. **Does A improve held-out prediction?** Yes, by the pre-specified diagnostic.
2. **Do edge preferences/depletions reproduce?** No convincing evidence under the pre-specified criteria.
3. **Is there a metadata-independent backbone?** No convincing evidence under the pre-specified criteria.
4. **Are continuation constraints reproducible?** Yes, by the pre-specified diagnostic.
5. **Does second order materially improve?** No convincing evidence under the pre-specified criteria.

## Outcome

`PROFILE_BACKBONE`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.

## Sequence reconstruction diagnostic

Preferred-backbone edge coverage: 0; token coverage: 0; observed transition coverage: 0. A deterministic frequency-matched graph with the same number of edges covers 0 of transitions.
