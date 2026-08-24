# Task83 confirmatory comparison design

Status: **FROZEN BEFORE TARGET OPENING**. This file fixes the implementation
of the target-blind contract in `task82a1/TASK83_COMPARISON_CONTRACT.md`.
Task83 is evaluation only; it will not regenerate, select, or tune any model.

## Inputs and immutable spaces

The primary space is exactly `F2_COMMON_DIRECT.tsv`. CORE inference uses its
three edit-family CORE metrics; the five SUPPORTING metrics are reported as
supporting diagnostics and never rescue CORE failure. The assembler list is
reported only as `PROJECTION_EVIDENCE`: synthetic assembler-line values are
not numerically compared with Voynich physical-line values. All other frozen
CORE metrics remain in coverage/residual accounting as `NOT_MODELLED`,
`CONTRADICTED`, `NOT_APPLICABLE`, or `EXPLAINED`; none is silently discarded.
SX is contextual because Voynich has no paired expansion. AX3--AX6 are
descriptive only because AX validation failed, and cannot alter any verdict.

Both ZL3b (Zandbergen--Landini) and IT2a (Takahashi) are mandatory targets.
Neither may be selected based on fit. Doyle, Longfellow, and Astafiev are the
only natural references. Frozen Task82/82a/82a.1 mechanism strata and frozen
Task82b shorthand/extraction strata and matched nulls are retained in full,
including failed, unstable, and degenerate rows.

## Normalization, endpoint distance, and coverage

For metric *m*, comparisons use the frozen natural-control location and scale
already exported by Task79c discriminative validation. A missing or zero scale
makes that metric explicitly unavailable; it is never re-estimated from a
target or model. The standardized absolute residual is
`r_m = |V_m-X_m|/s_m`. Metric residuals are averaged within frozen family and
families are averaged equally. Coverage is `C = available eligible direct
family weight / total eligible direct family weight`; the two reported
distances are `D` and `D_adjusted=D/C`. No imputation is permitted. Replicate,
corpus, scale, policy, and seed strata remain visible; summaries report median
and empirical 2.5/97.5 percentiles. LOCAL and GLOBAL cue policies are never
pooled.

Endpoint compatibility is not synonymous with closest. A class is endpoint
compatible only when both transcriptions fall within the frozen/control-based
95% empirical envelope of that class (leave-one-out natural distances for the
natural class; replicate distribution for a frozen mechanism class), with
adequate direct CORE coverage. Empirical percentiles and sample counts are
always reported. With fewer than 20 independent units, a tail result is
reported as low-resolution and cannot alone establish statistical support.

## Trajectories and nulls

The preregistered natural centroid `N` is the equal-corpus centroid of Doyle,
Longfellow, and Astafiev. `DeltaV=V-N`. Frozen class deltas are standardized by
the same per-metric natural scales. On jointly available non-degenerate direct
metrics report cosine similarity, sign agreement, Euclidean norm ratio
`||DeltaClass||/||DeltaV||`, and overshoot (`ratio>1.5` with positive cosine).
Metric directions are `TARGET_ALIGNED`, `TARGET_OPPOSED`, `NEUTRAL`,
`UNSTABLE`, or `NOT_APPLICABLE`; zero target or class displacement is neutral.
The class trajectory is the frozen stratum distribution, never a fitted
combination. Fontana trajectory is unavailable unless a before/after pair is
already explicit in an upstream frozen artifact.

Shorthand is tested first by its real abbreviated-minus-expanded trajectory
against RANDOM_DELETION_MATCHED, FREQUENCY_MATCHED_DELETION, and
POSITION_MATCHED. Pair-defined line-position degeneracies are excluded.
Extraction is tested by operator-minus-carrier against
RANDOM_SUBSEQUENCE_MATCHED, POSITION_STRATIFIED_RANDOM, and PERIODIC_PHASE.
FIRST/LAST results are labelled line-collapse-confounded unless their frozen
matched null isolates a positional-specific advantage. Null separation
requires a two-sided frozen `p<=0.05`, consistent direction on both
transcriptions, and advantage over every applicable matched-null family. No
multiple correlated metric is counted as an independent confirmation.

## Evidence and decision gates

Evidence levels and allowed verdict vocabulary are exactly those in Task83
sections 50--58. LEVEL 2/S3/A3 requires multi-family evidence where the space
contains multiple usable families, transcription stability, and null
separation. LEVEL 3 additionally requires mechanism-specific trajectory or
behaviour and worse competing classes. LEVEL 4 additionally requires frozen
knowledge-dependent recovery, corpus-scale persistence, and no comparable
simpler explanation. Task83 cannot award LEVEL 5, S4, or A4.

If no class passes its support gate, `BEST_SUPPORTED_CLASS=NONE_SUPPORTED`;
if evidence resolution or coverage prevents the gates from being evaluated,
it is `INCONCLUSIVE`; equivalent supported classes produce
`MULTIPLE_COMPATIBLE`. Pairwise advantage is meaningful only when the 95%
empirical distance intervals do not overlap and the favored class passes its
own support gate. Equifinality is present when two or more supported classes
have overlapping intervals and no null-separated pairwise advantage.

The final frozen marker is issued only after checksum re-verification,
complete output reproduction, and the validation suite. Any scientific
semantic change after the sentinel invalidates the experiment. A symmetric
implementation-only bug fix must be documented and all affected hypotheses
recomputed.
