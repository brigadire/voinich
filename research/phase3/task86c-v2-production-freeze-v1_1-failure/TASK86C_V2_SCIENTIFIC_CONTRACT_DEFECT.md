# PF-SC01 — empty-token generation is scientifically undefined

## Classification

`SCIENTIFIC_CONTRACT_DEFECT`. Resolving this changes generated controls, RNG consumption, DEVELOPMENT calibration, structural/predictive thresholds, confirmatory inputs, and potentially final verdicts.

## Frozen contradiction

The V1.1 machine contract requires every corpus token to be nonempty. M0 generation and the frozen `M0_GEN_A` control generator sample `<EOS>` with positive probability at the initial token position, but no artifact defines what happens when EOS is the first outcome. M1 and M4 also permit initial EOS under their frozen rows.

For the actual frozen M0 control namespace and first counter tuple `(generator_index=0, scale_index=0, replicate=0, draw=0)`, G1V2-RNG-1 produces:

- digest `edb14d737bdf70ce62ff1a28cfbdcc6f05a1fc904a34ff493da64264578f6794`;
- `u53 = 0.92848667210989588`;
- M0_GEN_A outcome `<EOS>` because its EOS interval is `[0.8,1)`.

Thus the defect is reached by the first token of a required frozen control, not merely by a theoretical edge case.

## Scientifically non-equivalent repairs

At least two nonempty-token implementations are natural but yield different frozen corpus bytes:

1. reject initial EOS and consume successive RNG draws; draw 1 is `0.14656015785300869`, producing first glyph `a`;
2. condition the first event on non-EOS and reuse draw 0 against the renormalized glyph CDF; it produces first glyph `d`.

Accepting the initial EOS instead violates the machine corpus representation. Skipping the empty token, retrying a whole token, consuming another counter namespace, or emitting a replacement glyph introduce still more behaviors. No retry/counter/failure rule selects one.

## Required repair

Publish a scientific contract revision that freezes minimum token length, exact initial-EOS handling for every applicable model/control generator, RNG counter consumption, retry/failure semantics, and matching goldens. E1 is execution-identity-only and cannot supply this scientific choice.

Production calibration, escrow creation, blind/natural materialization, executable freeze, and run authorization must restart only after that repair.
