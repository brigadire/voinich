# Task58 report

## Status

The independent analyzer and the exact primary estimands are implemented.
Canonical Voynich and the available natural controls were analyzed. Existing
global-uniform and frequency-v1 transformed controls were analyzed at token
level. Their synthetic glyph-edge values are deliberately `NOT_APPLICABLE`:
the `xNNNNNN` labels are opaque token IDs, not a glyph representation.

The full `comparison.tsv` is a complete-series smoke/batch run with 10
shuffles (the implementation default and frozen protocol remain 100). The
key Doyle and canonical-Voynich rows are also recomputed with 100 shuffles in
`KEY_100_RESULTS.tsv`; this is the primary-parameter check.

## Key 100-shuffle results

| corpus | token raw MI | shuffle mean | corrected | share | edge corrected | edge share |
|---|---:|---:|---:|---:|---:|---:|
| Doyle | 3.274235 | 2.510842 | 0.763393 | 9.376% | 0.072081 | 1.766% |
| canonical Voynich | 3.053200 | 2.962465 | 0.090736 | 1.099% | 0.216302 | 6.271% |

This directionally reproduces the published headline: canonical Voynich has
much weaker capped adjacent-token order than Doyle while its last-to-first
glyph coupling is higher. Because the source transcription, line selection,
and sample size differ from R&T's 32,747-token matched subset, this is
`METHOD_REPLICATION_DIFFERENT_TRANSCRIPTION`, not exact numerical replication.

## Controls and dose response

In the available Doyle global-uniform series, corrected token share is:

`H2 6.825% -> H4 4.830% -> H6 4.028% -> H8 3.325%`.

Thus increasing homophonic load moves token-order share monotonically toward
Voynich's 1.10% in this series, but does not reach it. The same conclusion is
descriptive, not an optimization claim. Frequency-v1 values at Hmax4/6/8 are
5.483%, 4.423%, and 3.596%; allocation changes the magnitude but preserves
the broad downward trajectory. Longfellow and Astafiev rows are included for
genre controls. The requested triangular-v1 files do not exist among the
frozen prepared artifacts, so no new models were created and that cell is
explicitly unavailable.

Glyph-edge coupling is not compared for synthetic labels. Consequently this
experiment cannot claim that homophony reproduces the independent R&T
edge-glyph property; it only tests token-order response for those controls.

## Relation to existing stages

Task25/27 outputs remain scientifically separate. No Stage metric was used as
an objective, and no aligned re-estimation was silently substituted for the
R&T estimand. A follow-up may join this table to a preregistered Stage25/27
export, but the present result already establishes the required separation.

No decryption, historical-authorship, or “Voynich is homophonically
encrypted” conclusion follows.
