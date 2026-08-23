# Transition network validation

## Global accounting

Unique tokens: 9288; eligible tokens: 622; observed adjacent eligible edges: 11351; testable edges: 5091. Edges in >=2 blocks: 4126; in >=3 blocks: 3692. Cross-joint: 3973; cross-Currier: 3973; cross-hand: 0.

FDR-significant preferred: 88; depleted: 0. Strict backbone preferred: 54; depleted: 0. Metadata-specific: 0; block-concentrated: 4; significant unstable: 30.

Replicated outgoing profiles: 3; incoming profiles: 4.

## Predictive validation

Mean held-out M0-M1 log-loss delta: 0.400885 (positive favors knowledge of A). Mean held-out M1-M2 delta: -0.60109; M2 wins in 0/64 blocks.

## Required questions

1. **Does A improve held-out prediction?** Yes, by the pre-specified diagnostic.
2. **Do edge preferences/depletions reproduce?** Yes, by the pre-specified diagnostic.
3. **Is there a metadata-independent backbone?** Yes, by the pre-specified diagnostic.
4. **Are continuation constraints reproducible?** Yes, by the pre-specified diagnostic.
5. **Does second order materially improve?** No convincing evidence under the pre-specified criteria.

## Outcome

`PROFILE_BACKBONE`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.

## Sequence reconstruction diagnostic

Preferred-backbone edge coverage: 0.00475729; token coverage: 0.120579; observed transition coverage: 0.0138825. A deterministic frequency-matched graph with the same number of edges covers 0.0138825 of transitions.
