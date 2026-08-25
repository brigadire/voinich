# Executable G1 adequacy gates

## PredictiveAdequacy

Fit on DEVELOPMENT, select the candidate within a class on VALIDATION, freeze
selection, then evaluate HELDOUT. Required metrics are PM1, PM2, PM4, PM5, and
PM6; PM3 is reporting-only because M0 and formation models have different
native units. Compare against B1 and against B2 where B2 has the same scoring
unit. PM1, PM2, and PM5 must be lower; PM4 and PM6 must be higher. Improvement
must be strict and greater than the maximum applicable MFC nearest-rank 0.95
null threshold on each transcription. Exact equality fails. All five metrics
must pass, every value must be finite except a registered degeneracy, and each
cross-transcription effect must be at least `DIRECTION_STABLE`.

`PredictiveAdequacy(G)=PASS` exactly when the preceding conjunction is true;
otherwise it is `FAIL`. M0 equals B1 and therefore remains a measured floor but
cannot pass its own strict improvement gate.

## StructuralAdequacy

Use only the seven inherited G1 Fingerprint V2 metrics. At matched scale and
the deterministic stopping checkpoint, a metric passes when
`abs(median_generated - heldout) <= max_applicable_MFC_q95_distance`. Equality
passes. Degenerate/unavailable metrics fail; none may be silently removed.

The `edit family` members are EF1_GIANT_COMPONENT_SHARE,
EF1_ISOLATE_SHARE, EF2_GLOBAL_CLUSTERING, and
EF3_DEGREE_FREQUENCY_SPEARMAN; at least three must pass. The `lexical paradigm`
members are LP1_RULE_SUPPORT_GINI, LP4_PREFIX_ATTACHMENT_NMI, and
LP4_SUFFIX_ATTACHMENT_NMI; at least two must pass. Each family has at least two
members and both families must pass on both transcriptions with at least
direction stability. That conjunction is `StructuralAdequacy(G)=PASS`.

## Sufficiency, failure, and minimality

Sufficiency additionally requires convergence, no registered failure, all
enumerated seeds retained, and the same gate result at scales 0.5, 1.0, and 2.0.
Minimality is computed separately per transcription among candidates passing
both adequacy gates. Complexity is the inherited Task85 sum. Candidates are
equivalent when absolute difference is at most
`max(1 bit, 1e-6*min(complexity_a,complexity_b))`; within the set prefer lower
model rank M0..M5 and then candidate id. A final G1 requires the same selected
class in both transcriptions. Complexity values are never pooled across
transcriptions.
