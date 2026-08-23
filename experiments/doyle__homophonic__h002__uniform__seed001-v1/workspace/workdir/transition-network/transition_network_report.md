# Transition network validation

## Global accounting

Unique tokens: 7286; eligible tokens: 560; observed adjacent eligible edges: 11346; testable edges: 5935. Edges in >=2 blocks: 4976; in >=3 blocks: 4458. Cross-joint: 4855; cross-Currier: 4855; cross-hand: 0.

FDR-significant preferred: 163; depleted: 0. Strict backbone preferred: 116; depleted: 0. Metadata-specific: 0; block-concentrated: 4; significant unstable: 43.

Replicated outgoing profiles: 5; incoming profiles: 7.

## Predictive validation

Mean held-out M0-M1 log-loss delta: 0.559605 (positive favors knowledge of A). Mean held-out M1-M2 delta: -0.845238; M2 wins in 0/64 blocks.

## Required questions

1. **Does A improve held-out prediction?** Yes, by the pre-specified diagnostic.
2. **Do edge preferences/depletions reproduce?** Yes, by the pre-specified diagnostic.
3. **Is there a metadata-independent backbone?** Yes, by the pre-specified diagnostic.
4. **Are continuation constraints reproducible?** Yes, by the pre-specified diagnostic.
5. **Does second order materially improve?** No convincing evidence under the pre-specified criteria.

## Outcome

`PROFILE_BACKBONE`. This is evidence only about reproducible transition constraints; it does not imply language, grammar, semantics, or decipherment.

## Sequence reconstruction diagnostic

Preferred-backbone edge coverage: 0.0102239; token coverage: 0.230357; observed transition coverage: 0.0350761. A deterministic frequency-matched graph with the same number of edges covers 0.0350761 of transitions.
