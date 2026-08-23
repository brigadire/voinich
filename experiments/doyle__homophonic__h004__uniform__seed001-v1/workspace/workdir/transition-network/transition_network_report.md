# Transition network validation

## Global accounting

Unique tokens: 9599; eligible tokens: 630; observed adjacent eligible edges: 11435; testable edges: 4656. Edges in >=2 blocks: 3936; in >=3 blocks: 3679. Cross-joint: 3816; cross-Currier: 3816; cross-hand: 0.

FDR-significant preferred: 37; depleted: 0. Strict backbone preferred: 17; depleted: 0. Metadata-specific: 0; block-concentrated: 2; significant unstable: 18.

Replicated outgoing profiles: 5; incoming profiles: 1.

## Predictive validation

Mean held-out M0-M1 log-loss delta: 0.276432 (positive favors knowledge of A). Mean held-out M1-M2 delta: -0.508997; M2 wins in 0/64 blocks.

## Required questions

1. **Does A improve held-out prediction?** Yes, by the pre-specified diagnostic.
2. **Do edge preferences/depletions reproduce?** Yes, by the pre-specified diagnostic.
3. **Is there a metadata-independent backbone?** Yes, by the pre-specified diagnostic.
4. **Are continuation constraints reproducible?** Yes, by the pre-specified diagnostic.
5. **Does second order materially improve?** No convincing evidence under the pre-specified criteria.

## Outcome

`PROFILE_BACKBONE`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.

## Sequence reconstruction diagnostic

Preferred-backbone edge coverage: 0.00148666; token coverage: 0.0460317; observed transition coverage: 0.0042934. A deterministic frequency-matched graph with the same number of edges covers 0.0042934 of transitions.
