# Task67 report

## Scope and verdict

This is a synthetic known-plaintext analysis of mechanism classes, not Voynich decryption. The tested Task66 representatives show that constrained formation can produce a Voynich-like structural target while retaining only a fraction of plaintext information. M0/M1 are clean reversible controls; M2 is key/ambiguity limited; G/M9/M10/M11 are lossy or fragile under this implementation.

## Required answers

1. Message preservation is high only for M0/M1; Task66 plaintext dependence alone does not imply recoverability.
2. Constrained formation is intrinsically many-to-one in G/M9/M10/M11; exact key/state does not recreate discarded distinctions.
3. Stateful candidates are particularly vulnerable to one boundary/glyph error: without checkpoints the tested model can remain desynchronized; resets localize the damage (RESET_REQUIRED_FOR_ROBUSTNESS).
4. Boundary operations, conflation, and insertion/deletion are reported separately. A reversible raw representation can become REPRESENTATION_INDUCED_INFORMATION_LOSS after many-to-one transcription.
5. Dense valid-form spaces permit silent valid-to-valid substitutions; therefore detected errors and wrong-but-confident outputs are not interchangeable.
6. Language redundancy is an external prior and is evaluated without TEST leakage; it cannot restore distinctions removed by a many-to-one encoder.
7. The fingerprint/information frontier contains both control and form representatives. Statistical compatibility therefore does not determine decryptability.

## Classification

- `M0_IDENTITY`: `MATHEMATICALLY_REVERSIBLE`; sanity upper bound.
- `M1_MONOALPHABETIC`: `MATHEMATICALLY_REVERSIBLE`; bijective substitution; key required.
- `M2_HOMOPHONY_H2`: `AMBIGUOUS_BUT_DECODABLE`; Task59-style stochastic homophony.
- `G_FORM_MEDIUM`: `INTRINSICALLY_LOSSY`; constrained formation representative.
- `M9_GROUP_FORM_FIXED`: `INTRINSICALLY_LOSSY`; generated-boundary constrained representative.
- `M10_STATEFUL_FORM_K2`: `PRACTICALLY_FRAGILE`; Task66 Pareto representative.
- `M11_MIXED_FORM_K2`: `PRACTICALLY_FRAGILE`; Task66 Pareto representative.

The tested mechanisms support the possibility that an originally recoverable synthetic encoding may become practically unrecoverable after copying/transcription damage: SUPPORTED_AS_POSSIBILITY. This is not a claim about the Voynich manuscript or the cause of its undeciphered status.

Estimator note: H(P|C), I(P;C), and R_I are finite-corpus plug-in estimates; short-block exhaustive rows are exact where marked, while larger preimage rows are beam/lower-bound diagnostics. They are not the sole recoverability criterion.
