# Task82a.1 frozen F2 coverage extension

This post-generation measurement extension consumes only immutable Task82a raw observable documents and frozen F2 definitions. It creates no pages, folios, sections, hands, loci, recto/verso pairs, or Currier-like metadata. Metrics computed on `ASSEMBLER_LINE` groups remain segregated from direct metrics and are never called physical-line measurements.

## Frozen computation plan

The plan is written before extended-vector generation. Existing Task82a EF/LP/cross-scale rows are imported verbatim. Six observed Task79 statistics are extracted through a generic ordered-group API proven bit-equivalent to the frozen implementation on regression fixtures. They need no permutation/bootstrap repetitions because this extension reports observed values, not new significance verdicts. In particular, `2DL1` preserves the frozen task79-v1 implementation's three-class boundary behavior; the registry prose/implementation mismatch is documented and is not corrected because doing so would create a new metric version.

## Preregistered stability rules

For cross-corpus and cross-seed ranges: `STABLE <= 0.01`, `PARTIALLY_STABLE <= 0.10`, otherwise `UNSTABLE`; fewer than two available observations is `INCONCLUSIVE`. Scale convergence applies to paired MEDIUM/LARGE observations with the same thresholds and names `CONVERGED`, `PARTIALLY_CONVERGED`, `NOT_CONVERGED`. Cue-policy effects use the same absolute-range bands and never aggregate LOCAL with GLOBAL. These descriptive thresholds were fixed before regenerated extended values were written.
