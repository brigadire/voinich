# PF-SC01 reproduction

The defect is independently reproduced for `M0_GEN_A` with G1V2-RNG-1 root
`6f5a9c731de248b480c66b237ace215044689c5fa2f593e510b73dce18a49027`,
namespace `g1v2/control/generate`, and counters `(0,0,0,0)`.

The SHA-256 RNG digest is
`edb14d737bdf70ce62ff1a28cfbdcc6f05a1fc904a34ff493da64264578f6794`;
its U53 is exactly `0.92848667210989588`. Under the explicit registry order,
the original intervals are `a:[0,.28)`, `b:[.28,.50)`, `c:[.50,.68)`,
`d:[.68,.80)`, and `EOS:[.80,1)`. Thus EOS is selected while the token length
is zero, violating the nonempty-token invariant.

Two V1.1-compatible continuations differ:

- Rejection consumes draw 0, advances to counters `(0,0,0,1)`, obtains
  U53 `0.14656015785300869`, emits `a`, consumes two draws before the first
  glyph, and next uses draw index 2.
- Conditional renormalization reuses draw 0 on probabilities
  `(0.35,0.275,0.225,0.15)`, emits `d`, consumes one draw, and next uses draw
  index 1.

The preferred state-boundary rule therefore deterministically yields `d`, but
it remains a candidate because the all-path audit found separate generation
ambiguities that prevent a narrow V1.2 freeze.
