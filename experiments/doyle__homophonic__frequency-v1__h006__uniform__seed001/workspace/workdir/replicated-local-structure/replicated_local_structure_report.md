# Replicated local structure audit

Confirmatory audit of frozen candidates only. Corpus SHA256: `4bc3c5b699cb6e433399c668b5cc49478f3c6d045128ca6c3a2bd0ce3a1423e5`. No token pairs, n-grams, distances, thresholds, or classification rules were added or changed.

## Metric semantics

Distance profile similarity is the mean Jensen–Shannon similarity (`1 - JSD/ln(2)`) across non-empty left/right context distributions at frozen distances ±1…±20. The legacy `fraction_above_frozen_threshold` compares this [0,1] similarity with the discovery field `max_distance=20`; its zeros are retained and documented, not repaired retrospectively. Absolute LOBO similarity and null-relative standardized effect are reported separately.

## Summary

- distance-profile: 0 frozen, 0 FDR-significant, 0 robust cross-block, 0 artifact/invalid.
- sequence: 197 frozen, 191 FDR-significant, 0 robust cross-block, 197 artifact/invalid.

## Distance-profile distributions

Frozen median similarity: q<=0.05 primary set median 0 (n=0); q>0.05 descriptive set median 0 (n=0). This comparison performs no new selection or significance test.

## Main outcome

A. NOTHING REMAINS — no convincing transferable local structure remains under current tests.

## UNIVERSAL sequence inventory

Length distribution: n=2: 197;

Canonical occurrence distribution: min 4, median 5.0, mean 6.07, P90 9.0, max 14.

Diagnostic statuses:
- TRANSCRIPTION_AMBIGUOUS: 197

Higher-order sequences (n>=3) exceeding both the shuffle FDR criterion and nominal Markov p<=0.05: 0. Markov values are secondary and are NA where leakage-free same-class training is unavailable.

## Interpretation guardrails

Distance and sequence p-values are separate families. New diagnostic statuses do not replace the frozen UNIVERSAL/WEAK/BLOCK_SPECIFIC classifications. The audit establishes formal reproducibility only; it does not establish language, semantics, grammar, or decipherment.
