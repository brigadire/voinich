# Task78 — Fontana device-family operational validation

## Outcome

Task78 validates one directly reproducible positional core for F08 Serpens and
bounded invariant/profile cores for F07, F10, F11 and F12. It does **not** find
one homogeneous Fontana machine family: the observed classes are an ordered
spatial store, cyclic selector, profile-dependent alignment store, indexed
cue store and event-triggered reminder. F02–F06 and F09 remain Tier C rather
than being completed with modern rules.

The evidence base and names come solely from task74. Task76 supplies the
experimental discipline and F01 comparison row. No Voynich data, Fingerprint
v2 metric, text-like-output objective or parameter tuning was used.

## Reproducibility and artifacts

`task78-analyze` writes 192 trials with complete JSON state per row, named
profiles and fixed seed 780823. Unit tests cover round trips, local/frame-loss
damage, cyclic periods, independent band transitions, lookup corruption and
clock/cue transitions. `TASK78_MODEL_SELECTION.tsv` records every task74
candidate; `EIHU_TABLE.tsv` prevents runnable H parameters from becoming
facts. The printable F08 SVG demonstrates traversal topology and explicitly
disclaims facsimile status.

Machine-readable packages use one schema with metadata, sources,
reconstruction, uncertainty, components, state schema, operations,
transitions, encoding/retrieval, knowledge, profiles, experiments, results,
tests and verdict. The package files are candidates for task80 input, not a
final frozen registry.

## Experimental findings

| model/profile | baseline | state/dynamics | ablation and damage conclusion |
|---|---:|---|---|
| F08 R0 | 6/6, CI 0.610–1.000 | fixed 12-hole H realization | unknown start/direction/boundary multiplies interpretations; all order-changing damage failed exact recovery; topology-preserving distortion was 6/6 |
| F07 R0 | 23/23, CI 0.857–1.000 | 23-state unit-step cycle | zero/step knowledge is essential; no literal message cycle validated |
| F10 R0 | 3/3, CI 0.439–1.000 | `23^7`; one-band cycle 23 | one-band damage is local in R0, cascading in coupled R1; therefore propagation is not historical invariant |
| F11 R0 | 6/6, CI 0.610–1.000 | partial index→cue map | loss/swap destroys exact lookup; unknown convention leaves all occupied indices compatible |
| F12 R0 | 12/12 transition/cue checks, CI 0.758–1.000 | 12-state H cycle | untrained formal decoder 0/3; wrong map 0/3; rich content is in learned association |

For F08 ablations, `observed` is merely the first deterministic enumeration,
so its occasional equality with the target is not blind-decoder accuracy. The
primary ablation measure is `compatible_interpretations`. Unknown semantic
association is reported as unbounded, not guessed. F12 empty-event ticks are
included in the 12 transition checks; only three ticks emit cues.

No human pilot was performed. Consequently human recall rate, false recall
rate after delay, latency and human-versus-formal differences are
`INCONCLUSIVE`, not silently replaced by algorithmic lookup.

## Per-model verdicts

### F08 Serpens

- `SOURCE_SUFFICIENCY`: `PARTIALLY_SUPPORTED`; `RECONSTRUCTION_CONFIDENCE`:
  `MEDIUM`.
- `OPERATIONAL_CYCLE_VALIDATED`, `BASELINE_RECOVERY`,
  `PRIOR_KNOWLEDGE_DEPENDENCE`, `STATE_DEPENDENCE`,
  `LITERAL_STORAGE_FUNCTION`: `SUPPORTED` for R0.
- `STATE_DAMAGE_ROBUSTNESS`: `NOT_SUPPORTED`; propagation `MIXED` (local,
  synchronization and global ambiguity by damage type).
- `INDEXED_RETRIEVAL_FUNCTION`: `NOT_APPLICABLE`; `MNEMONIC_CUE_FUNCTION`:
  `PARTIALLY_SUPPORTED` from purpose, not experimentally as human recall.
- `MODEL_READY_FOR_FREEZE`: `PARTIALLY_SUPPORTED`: invariant positional core
  is ready for task80 review, boundary/capacity/alphabet are not frozen.

### F07 Rota

Source and reconstruction are `MEDIUM`/`PARTIALLY_SUPPORTED` only for finite
cyclic selection. Baseline selector operation and state dependence are
supported; literal storage, damage robustness and freeze readiness are not.
Error propagation beyond a one-wheel shift is `INCONCLUSIVE`.

### F10 Cylindrus

Source and reconstruction are `MEDIUM`/`PARTIALLY_SUPPORTED`. R0 round trip is
supported as an H-bounded profile, not as a historical encoder. Literal
storage is `PARTIALLY_SUPPORTED`; robustness and freeze readiness are not.
Local versus cascade propagation is `MIXED` across R0/R1 and therefore cannot
be promoted beyond profile scope.

### F11 Arismetricum

Indexed retrieval and known-convention baseline are `SUPPORTED` for R0;
source sufficiency and full operational reconstruction are
`PARTIALLY_SUPPORTED`, confidence `MEDIUM`. Literal text storage and damage
robustness are not supported. Cue semantics, mapping and freeze readiness
remain unsupported.

### F12 Horalogius

Temporal transition, cue emission and prior-knowledge dependence are
supported for R0; exact mechanism and human mnemonic effectiveness are only
partially supported/inconclusive. Literal and indexed functions, robustness
and freeze readiness are not supported. A signal is not the remembered
message.

## Common verdicts

- `F08_OPERATIONAL_MODEL_READY`: `SUPPORTED` for R0 invariant core, `MEDIUM`
  reconstruction confidence; not a complete physical replica.
- `ADDITIONAL_FONTANA_MODELS_READY`: `PARTIALLY_SUPPORTED`; F11/F12 invariant
  cores and F07/F10 sensitivity profiles are suitable for task80 analysis,
  not final freezing.
- `COMMON_OPERATIONAL_CORE_SUPPORTED`: `PARTIALLY_SUPPORTED`; external state,
  user action and convention recur, but no shared encoder/decoder does.
- `DEVICE_FAMILY_HETEROGENEITY_SUPPORTED`: `SUPPORTED`, stable across all
  declared profiles.
- `FONTANA_VALIDATED_SET_READY_FOR_TASK80`: `SUPPORTED` with F08/F11/F12
  invariant cores and F07/F10 explicitly profile-scoped. It is **not** a
  `FONTANA_MODELS_FROZEN` declaration.

Historically unsupported properties include exact capacities and alphabets,
F07 disk multiplicity/step, F10 band coupling/route, F11 cue mapping, F12
drive/calibration and every human retention estimate. These limitations are
carried into the package uncertainty fields and `TASK80_OPEN_QUESTIONS.md`.
