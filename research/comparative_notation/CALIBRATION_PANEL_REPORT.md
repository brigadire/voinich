# Calibration panel report v1 (B01)

Production run: `notation-corpus calibrate --replicates 40` at every frozen
checkpoint (5,000 / 10,000 / 20,000 / 39,380), base seed `20260830`.
Generators and estimator are frozen in `CALIBRATION_PANEL_SPEC.md` and
`CALIBRATION_GENERATORS/*.json`. Output: `CALIBRATION_SCALES.tsv` (268
rows), diagnostic: `CALIBRATION_DIAGNOSTICS.json`.

## C1 — Non-zero coverage

219 / 268 scalar strata (81.7%) received a usable (`OK`) scale; 49 (18.3%)
are `DEGENERATE`. Coverage by checkpoint:

| checkpoint | OK | DEGENERATE |
|---|---|---|
| 5,000 | 52 | 15 |
| 10,000 | 52 | 15 |
| 20,000 | 52 | 15 |
| 39,380 | 63 | 4 |

Coverage improves at larger checkpoints because two of the degenerate
families (`S06_PREFERRED_TRANSITION_FRACTION`,
`S07_DEPLETED_TRANSITION_FRACTION`, all five support regimes) resolve to
`OK` once there is enough transition data for `CAL-MARKOV2` and
`CAL-CGRAMMAR` to produce a non-degenerate pooled spread; at 39,380 tokens
only 4 strata remain degenerate at every checkpoint (see C3 below for why).

## C2 — Generator dominance

`CALIBRATION_DIAGNOSTICS.json` records, per checkpoint, the number of
scalar-metric observations contributed by each generator
(`generator_observation_counts`). Every one of the eight generators
contributes **exactly 2,680** observations at every checkpoint (40
replicates × 67 comparable scalar strata) — the panel is balanced by
construction, so no generator can dominate a pooled MAD/IQR estimate by
volume. (67 is the count of scalar strata that are comparable *for that
generator's own structure*; each generator contributes the same count
because none of the eight ever produces an incomparable structural draw for
this synthetic alphabet/length design.)

## C3 — Stability and the four persistently degenerate strata

`LeaveOneGeneratorFamilyOut` recomputes the scale table after excluding one
generator's 40 replicates; `leave_one_out_stratum_counts` in the
diagnostics file shows the same 67-stratum coverage survives removing any
single generator, at every checkpoint — no generator is a single point of
failure for the *existence* of a scale.

Four strata remain `DEGENERATE` even at the full 39,380-token checkpoint,
and the cause is structural, not a code defect:

- `G02_INITIAL_RESTRICTION_DENSITY`, `G03_FINAL_RESTRICTION_DENSITY`,
  `T11_POSITIONAL_RESTRICTION_DENSITY` — seven of the eight generators use
  the full 24-symbol alphabet at every token position and converge to
  `0` restriction density; only `CAL-CGRAMMAR` deliberately restricts by
  design and converges to a different constant. The pooled sample is
  bimodal with the majority mass exactly at `0`, so both `MAD` and `IQR`
  land on `0` — this is the intended behavior of `EstimateScale` (no
  epsilon is invented for a majority-constant pooled sample), not a
  panel-composition bug.
- `L01_LINE_TOKEN_COUNT_MEAN` — every from-scratch generator emits exactly
  8 tokens per line by construction (`genTokens`'s line-flush rule), so the
  pooled line-length distribution has zero spread everywhere it is not a
  boundary effect from the corpus's own final partial line. This is a
  known limitation of the from-scratch generator design (see
  `CALIBRATION_PANEL_SPEC.md` §2): the panel does not vary physical-line
  length independently of the from-scratch generators' fixed 8-token rule.
  `L01` therefore correctly reports `DEGENERATE` rather than a fabricated
  scale, and any future candidate/VM comparison on `L01` must report
  `metric_scale_status = DEGENERATE` per B01 section 29, not silently skip
  scaling.

This is reported as a diagnostic (per the task's C3 instruction) and was
**not** used to pick a different estimator or to add an ad hoc epsilon; the
frozen `1.4826*MAD` → `IQR/1.349` → `DEGENERATE` fallback chain
(`CALIBRATION_PANEL_SPEC.md` §4) is applied uniformly.

## C4 — Reproducibility

`TestC4CalibrationReproducibility` (`internal/notation/calibration_test.go`)
runs the full panel-generation-and-scale pipeline twice with the identical
seed schedule and asserts a byte-identical `CALIBRATION_SCALES.tsv`
serialization; it passes deterministically (checked across five repeated
process invocations during this task, see `BOOTSTRAP_PROTOCOL.md` §5 for
the determinism defect this required fixing).

## C5 — No candidate access

`RunCalibrationPanel` takes no corpus, file path, or candidate argument;
`TestC5NoCandidateOrVMDataInGenerators` asserts no generated record ever
carries the VM corpus id. The production `calibrate` CLI invocation above
took no `--input`/`--candidate`/VM-path flag.

## B01 decision

Per `PREPARATION_BLOCKERS.md`: `B01 = CLOSED`.
