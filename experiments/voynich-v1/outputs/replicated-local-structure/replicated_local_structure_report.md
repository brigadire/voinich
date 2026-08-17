# Replicated local structure audit

Confirmatory audit of frozen candidates only. Corpus SHA256: `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`. No token pairs, n-grams, distances, thresholds, or classification rules were added or changed.

## Metric semantics

Distance profile similarity is the mean Jensen–Shannon similarity (`1 - JSD/ln(2)`) across non-empty left/right context distributions at frozen distances ±1…±20. The legacy `fraction_above_frozen_threshold` compares this [0,1] similarity with the discovery field `max_distance=20`; its zeros are retained and documented, not repaired retrospectively. Absolute LOBO similarity and null-relative standardized effect are reported separately.

## Summary

- distance-profile: 28 frozen, 21 FDR-significant, 0 robust cross-block, 0 artifact/invalid.
- sequence: 104 frozen, 51 FDR-significant, 51 robust cross-block, 0 artifact/invalid.

## Distance-profile distributions

Frozen median similarity: q<=0.05 primary set median 0.0768088 (n=21); q>0.05 descriptive set median 0.0487769 (n=7). This comparison performs no new selection or significance test.

## Main outcome

C. HIGHER-ORDER SEQUENCES REMAIN — reproducible higher-order sequential organization exists.

## UNIVERSAL sequence inventory

Length distribution: n=2: 101; n=3: 3;

Canonical occurrence distribution: min 4, median 8.0, mean 10.27, P90 13.0, max 59.

Diagnostic statuses:
- REPLICATED_ABOVE_FREQUENCY_NULL: 51
- REPLICATED_BUT_EXPECTED_FROM_FREQUENCY: 53

Higher-order sequences (n>=3) exceeding both the shuffle FDR criterion and nominal Markov p<=0.05: 2. Markov values are secondary and are NA where leakage-free same-class training is unavailable.

## Interpretation guardrails

Distance and sequence p-values are separate families. New diagnostic statuses do not replace the frozen UNIVERSAL/WEAK/BLOCK_SPECIFIC classifications. The audit establishes formal reproducibility only; it does not establish language, semantics, grammar, or decipherment.
