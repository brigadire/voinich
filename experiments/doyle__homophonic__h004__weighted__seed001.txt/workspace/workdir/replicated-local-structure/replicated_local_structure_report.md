# Replicated local structure audit

Confirmatory audit of frozen candidates only. Corpus SHA256: `165319e63f30ae712e0cb47ffc414e333b7dd2f96ab82beedb4c70a2458b30a8`. No token pairs, n-grams, distances, thresholds, or classification rules were added or changed.

## Metric semantics

Distance profile similarity is the mean Jensen–Shannon similarity (`1 - JSD/ln(2)`) across non-empty left/right context distributions at frozen distances ±1…±20. The legacy `fraction_above_frozen_threshold` compares this [0,1] similarity with the discovery field `max_distance=20`; its zeros are retained and documented, not repaired retrospectively. Absolute LOBO similarity and null-relative standardized effect are reported separately.

## Summary

- distance-profile: 7 frozen, 1 FDR-significant, 0 robust cross-block, 0 artifact/invalid.
- sequence: 207 frozen, 196 FDR-significant, 0 robust cross-block, 207 artifact/invalid.

## Distance-profile distributions

Frozen median similarity: q<=0.05 primary set median 0.042874 (n=1); q>0.05 descriptive set median 0 (n=6). This comparison performs no new selection or significance test.

## Main outcome

A. NOTHING REMAINS — no convincing transferable local structure remains under current tests.

## UNIVERSAL sequence inventory

Length distribution: n=2: 200; n=3: 7;

Canonical occurrence distribution: min 3, median 8.0, mean 9.56, P90 15.0, max 39.

Diagnostic statuses:
- TRANSCRIPTION_AMBIGUOUS: 207

Higher-order sequences (n>=3) exceeding both the shuffle FDR criterion and nominal Markov p<=0.05: 0. Markov values are secondary and are NA where leakage-free same-class training is unavailable.

## Interpretation guardrails

Distance and sequence p-values are separate families. New diagnostic statuses do not replace the frozen UNIVERSAL/WEAK/BLOCK_SPECIFIC classifications. The audit establishes formal reproducibility only; it does not establish language, semantics, grammar, or decipherment.
