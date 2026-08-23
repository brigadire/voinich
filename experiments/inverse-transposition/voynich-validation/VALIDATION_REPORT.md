# Voynich inverse-transposition validation

This is post-search validation. `structural-v2` search scores are ranking values only; they are not effect sizes or percentages. Raw effect is `candidate - original`.

## Fixed holdout

| corpus | transition | relation | sequence-2 | sequence-3 |
|---|---:|---:|---:|---:|
| original | 0.786086762964 | 0.00133208972659 | 0.140761922332 | 0.00461254612546 |
| w2 natural | 0.785121698756 | 0.000535093032071 | 0.095932109584 | 0.00156826568266 |
| delta | -0.000965064208103 | -0.000796996694518 | -0.0448298127479 | -0.0030442804428 |

## Discovery versus holdout

| metric | discovery delta | holdout delta | direction |
|---|---:|---:|---|
| transition | -0.00145707509549 | -0.000965064208103 | SAME_DIRECTION |
| relation | -0.00085826763698 | -0.000796996694518 | SAME_DIRECTION |
| sequence-2 | -0.0586861017293 | -0.0448298127479 | SAME_DIRECTION |
| sequence-3 | -0.00810097008482 | -0.0030442804428 | SAME_DIRECTION |

The complete candidate table is `discovery_effects.tsv`; `parameter_landscape.tsv` contains the same frozen candidates with raw metrics. The w2 score is a candidate-local min-max ranking score and is not interpreted as improvement.

## Controls and calibration

Doyle/T2/T4/T8 ranges are taken unchanged from `INVERSE_TRANSPOSITION_TASK54_1_REPORT.md`; no Voynich candidate participates. `fixed_calibration_effect_score` is post-hoc. The width-2 natural-text audit is in `width2_natural_controls.tsv` for Doyle and Longfellow.

## Null

Holdout nulls use the fixed budget and deterministic seed in `manifest.json`. Signed raw-delta percentiles are descriptive; the full distribution and fixed-calibration composite are in `null_distribution.tsv`.

## Classification

No pre-registered meaningful-effect threshold exists, so there is no significance claim. The split was not pre-registered and is documented as a limitation. No conclusion about decryption, decipherment, or plaintext recovery is made.
