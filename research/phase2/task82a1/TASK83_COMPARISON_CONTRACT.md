# Task83 comparison contract (frozen target-blind)

## Allowed spaces

The metric IDs in `F2_COMMON_DIRECT.tsv` are the only permitted direct synthetic/target intersection. `F2_ASSEMBLER_PROJECTION.tsv` is an `ASSEMBLER_LINE_ONLY` diagnostic space and must never be compared numerically with target physical-line metrics. Hierarchy, locus, folio, recto/verso, section, hand, and Currier-dependent metrics are excluded. The intersection is fixed by semantics, not values.

## Missingness and eligibility

Missing values remain missing: no zero, mean, target, control, or model imputation. `PARTIALLY_COMPARABLE` requires all three direct CORE edit-family metrics in every compared cell; partial availability is `PROJECTION_COMPARABLE`; no direct CORE value is `NOT_COMPARABLE`. Supporting metrics cannot rescue missing CORE eligibility.

## Normalization and distance

For each available direct metric, use the frozen F2 standardized per-metric difference and its pre-target natural-language-control scale. A metric lacking its frozen normalization scale is excluded with an explicit reason, never rescaled from Task82a or target values. Average metric distances within family, then average families with equal family weight. Preserve replicate uncertainty and report cell-level distance intervals; do not pool LOCAL and GLOBAL cue policies.

Let `C` be available eligible direct-family weight divided by total eligible direct-family weight in `F2_COMMON_DIRECT.tsv`. Report both raw pairwise-available distance `D` and coverage-adjusted distance `D_adjusted = D / C`; if `C=0`, distance is unavailable. This preregistered penalty prevents missing difficult families from improving a score without imputing their values. Report metric, family, and CORE coverage beside every distance.

## Aggregation constraints

Mechanism-level summaries keep corpus, replicate, scale, and scaling-policy strata visible. Family aggregation is family-balanced. Instability is propagated as `UNSTABLE_FOR_CONFIRMATORY_COMPARISON`; it is not a basis for deleting a mechanism. No metric, threshold, normalization, null, family weight, coverage rule, or uncertainty rule may change after target access. No ranking or selection decision is authorized by this contract.
