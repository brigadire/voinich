# G1-v2 predictive measurement contract

G1-v2 measures whether a fitted class predicts held-out token formation, not whether an implementation merely terminates. The normative metric definitions are in `G1V2_METRIC_REGISTRY.tsv`; the gate registry is executable policy.

PM1 is retained as joint surprisal and audit evidence, but it is not an independent gate: for a fixed heldout set it is exactly PM2 multiplied by the scored glyph count. PM2 is the required normalized predictive gate. PM4 remains useful evidence about novel forms, but a split without enough unseen occurrences cannot reject a model. PM5 remains required after closing the v1 partial-calibration loophole: all predictions, bin assignment, target frequency, weight, and contribution are stored. PM6 is required when identifiable and follows the separately frozen complement protocol.

Thresholds are not copied from G1-v1. Before confirmatory manifest freeze, development controls produce a named null distribution using the same estimator, split sizes, and multiplicity family. The frozen threshold is the conservative one-sided 0.95 null quantile plus the registry's practical-effect floor. The derivation code, inputs, values, and resulting threshold hash enter the config closure. Confirmatory values can never update it.

For each candidate × transcription × split, a PM record contains the candidate value, B1 value, every applicable B2 value, effect orientation, threshold ID/value, availability reason, finite flag, and a separate gate outcome for each baseline. Equality at a strict effect threshold is FAIL. A malformed probability, missing record, or nonfinite estimator is NOT_ASSESSABLE, not evidence of predictive inadequacy.

The complete required-gate truth table is:

| PM2 | PM5 | PM6 | Predictive verdict |
|---|---|---|---|
| PASS | PASS | PASS | PASS |
| FAIL | PASS/FAIL | PASS/FAIL | FAIL |
| PASS | FAIL | PASS/FAIL | FAIL |
| PASS | PASS | FAIL | FAIL |
| no FAIL and any NOT_ASSESSABLE | any | any | NOT_ASSESSABLE |

When a valid FAIL and an unrelated NOT_ASSESSABLE coexist, the candidate has observed inadequate evidence but cannot receive a complete sufficiency verdict: record the metric FAIL and overall `PREDICTIVE_NOT_ASSESSABLE`. This conservative rule prevents missing evidence from being hidden by a convenient failure. PM4 cannot change the verdict, but its complete record is mandatory.

`NONE` is available only if every candidate in every class was computationally completed and received a valid inadequacy verdict. Any unresolved candidate capable of changing the minimum makes the experiment `NOT_IDENTIFIABLE`.
