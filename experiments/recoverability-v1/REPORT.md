# Task67 report

## Scope and verdict

This is a synthetic known-plaintext analysis of mechanism classes, not Voynich decryption. All primary damage results now come from executed encode -> corrupt/represent/segment -> decode jobs. ERROR_RECOVERABILITY.tsv contains 102,900 raw stochastic rows; the single-error, reset, transcription, and segmentation tables retain positions, seeds, operation counts, and decoder outcomes.

## Required answers

1. **Task66 candidates preserving a message:** M0, M1, and the synthetic H2 representation recover clean glyphs exactly with their declared knowledge. G, M9, M10, and M11 recover only part of the message.
2. **Statistical dependence only:** constrained G/M9/M10/M11 retain positive paired-unit information but their actual decoded token rates are much lower; dependence is not unique recovery.
3. **Voynich-compatible and highly recoverable:** the tested constrained set has a fingerprint/recovery trade-off. G occupies the best tested compromise, but no constrained candidate has control-level unique recovery.
4. **Formation-information trade-off:** yes in this frozen set; high Task66 family compatibility coexists with lower exact token recovery.
5. **Loss from formation itself:** G is many-to-one even without state or corruption; PREIMAGE_MULTIPLICITY and the TEST-codebook oracle measure that loss.
6. **Unique constrained inverse:** none of the frozen constrained Task66 representatives has one. Reversible M0/M1/M2 controls show that the decoder can recover a unique inverse when the encoder preserves it.
7. **Ambiguous with exact key/state:** G, M9, M10, and M11; state knowledge cannot recreate distinctions discarded by the form map.
8. **Natural-language redundancy:** TRAIN+VALIDATION local priors improve choices for observed forms; wrong-language rows quantify the dependence. They do not eliminate intrinsic collisions.
9. **Glyph substitutions:** empirical 100-replicate curves are in RECOVERABILITY_CURVES.tsv; raw outcomes and exact seeds are in ERROR_RECOVERABILITY.tsv.
10. **Insertions/deletions:** glyph edits are generally local in single-error rows, while token-count-changing boundary edits can shift all later positional units.
11. **One-error severity:** ERROR_PROPAGATION.tsv reports actual incremental damaged units at four positions, including end-censored L_sync=-2 rows.
12. **Catastrophic state/position desynchronization:** observed for no-reset token-count-changing errors where no three-unit correct run returns before block end.
13. **Resynchronization length:** L_sync is measured, not inferred; -1 means catastrophic within the block and -2 means the block ended before a three-unit run was observable.
14. **Periodic resets:** token, line-sized, and fixed-N checkpoints localize the same requested boundary deletion; page-sized resets may be too sparse for a 128-word block.
15. **Boundary danger:** deletion, insertion, and +1 shifts are scored independently in SEGMENTATION_DAMAGE.tsv.
16. **Boundary reconstruction:** ciphertext-only dynamic programming produces measured precision, recall, F1, and downstream plaintext recovery; it is not given plaintext or oracle boundaries.
17. **Glyph conflation:** actual random class pairs are merged and decoded. Damage varies by pair/fraction/replicate rather than copying a clean score.
18. **Raw reversible to represented irreversible:** yes as a tested possibility; many-to-one conflation damages M0/M1 recovery despite a reversible raw encoding.
19. **Silent errors:** ERROR_DETECTABILITY.tsv distinguishes valid-form wrong decodes from invalid detectable forms.
20. **Dense valid-form risk:** valid-to-valid errors occur in the frozen form dictionary and are separately counted as undetectable/silent; this is a coding diagnostic, not a Voynich error-rate claim.
21. **Similar fingerprint, different recovery:** G, M9, M10, and M11 overlap strongly on the Task66 compatibility axis but differ substantially in Task67 recovery.
22. **Upper-right region:** no tested constrained candidate combines control-level recovery with maximal compatibility. G and M9 define the tested compromise frontier.
23. **Most encoding-like candidate:** G among constrained candidates, because it retains the highest measured clean recovery; M0/M1/M2 are reversible controls rather than Voynich-compatible claims.
24. **Most generator-like candidate:** the lowest-recovery stateful form representatives and the shuffled-input control; the latter demonstrates preserved grammar with message identity destroyed.
25. **Clean recoverable becoming practically undecodable:** supported for token-count-changing damage without sufficiently frequent checkpoints; reset rows show localization under the allowed robustness variants.
26. **Most destructive damage:** boundary merge/split and cascaded corruption are more globally disruptive than isolated within-token glyph substitutions in this positional decoder.
27. **Where information is lost:** encoding collisions (G/M9/M10/M11), synchronization after token-count errors, transcription conflation, and segmentation are separated from key secrecy and from clean decoder ambiguity.
28. **Original message but insufficient surviving representation:** SUPPORTED_AS_POSSIBILITY for tested synthetic mechanisms only.

## Classification

- `M0_IDENTITY`: `MATHEMATICALLY_REVERSIBLE`; sanity upper bound.
- `M1_MONOALPHABETIC`: `MATHEMATICALLY_REVERSIBLE`; bijective substitution; key required.
- `M2_HOMOPHONY_H2`: `AMBIGUOUS_BUT_DECODABLE`; Task59-style stochastic homophony.
- `G_FORM_MEDIUM`: `INTRINSICALLY_LOSSY`; constrained formation representative.
- `M9_GROUP_FORM_FIXED`: `INTRINSICALLY_LOSSY`; generated-boundary constrained representative.
- `M10_STATEFUL_FORM_K2`: `PRACTICALLY_FRAGILE`; Task66 Pareto representative.
- `M11_MIXED_FORM_K2`: `PRACTICALLY_FRAGILE`; Task66 Pareto representative.

The tested mechanisms support the possibility that an originally recoverable synthetic encoding may become practically unrecoverable after copying/transcription or segmentation damage: SUPPORTED_AS_POSSIBILITY. This is not a claim about the Voynich manuscript, EVA, historical error rates, or the cause of its undeciphered status.

Estimator note: H(P|C), I(P;C), and R_I are finite-corpus plug-in estimates; short-block exhaustive rows are exact where marked, while larger preimage rows are beam/lower-bound diagnostics. They are not the sole recoverability criterion.
