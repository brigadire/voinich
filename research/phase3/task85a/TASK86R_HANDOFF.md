# Task86R executable handoff

## Mandatory preflight and inputs

Run `python3 research/phase3/task85a/validate_contract.py --self-test`. Before
any MFC generation, all nine reported preflight flags must be `SUPPORTED`:
`CONTRACT_COMPLETE`, `ALL_GRIDS_FINITE`, `ALL_THRESHOLDS_DEFINED`,
`ALL_GATES_EXECUTABLE`, `NEGATIVE_PROTOCOL_COMPLETE`, `PM5_COMPLETE`,
`PM6_COMPLETE`, `SEED_CONTRACT_COMPLETE`, and `TASK86_BLOCKERS_RESOLVED`.
Any other result stops Task86R before fitting.

Authoritative inputs are the four corpus paths and hashes in the version
manifest, every inherited Task85 artifact at its recorded hash, and the exact
Task85 split. Do not regenerate the split. G1 may read only TOKEN-internal
GLYPH, COMPONENT, and POSITION data; it may not read neighboring TOKENs or any
STRUCTURAL_STATE variable.

## Execution order

1. Validate hashes, grids, thresholds, gates, deterministic parsing, and the
   no-fitting preflight.
2. Generate all 16 populations for each fixed MFC0–MFC2 generator and run the
   full 84-candidate pipeline. Materialize every deterministic q0.95 threshold;
   freeze calibration and its checksum before any Voynich fit.
3. Fit all rows of `G1_HYPERPARAMETER_GRID.tsv` on DEVELOPMENT separately for
   ZL3b and IT2a. Report failed rows; do not delete or replace them.
4. Use VALIDATION only for frozen-grid candidate selection and failure checks.
   Freeze model selection, source revision, selected candidate ids, calibration
   hashes, and code/config hashes before opening HELDOUT.
5. Evaluate every retained candidate on HELDOUT. Generate at scales 0.5, 1.0,
   and 2.0 with checkpoints 4, 8, 16, 32 and the frozen convergence stop.
6. Apply PM5, the deterministic negative protocol and PM6, the seven inherited
   G1 F2 metrics, transcription rules, adequacy gates, minimality, ladder, token
   depth, and explicit-rule mapping exactly as specified.

## Normative map

- Algorithms and numeric values: `G1_EXECUTABLE_CONTRACT.json` and
  `G1_ALGORITHM_REGISTRY.tsv`.
- Candidate rows: unified grid and the M2/M4/M5 projections; Task86R may add no
  candidate.
- MFC generators, populations, calibrated quantities, and thresholds:
  `G1_CALIBRATION_CONTRACT.md`.
- Scales, replicate checkpoints, convergence, numerical and growth rules:
  `G1_STABILITY_CONTRACT.md` and `G1_FAILURE_THRESHOLDS.tsv`.
- Cross-transcription classification:
  `G1_TRANSCRIPTION_STABILITY_CONTRACT.md`.
- Negative controls and metrics: `NEGATIVE_TOKEN_PROTOCOL.md`,
  `PM5_CALIBRATION_SPEC.md`, and `PM6_DISCRIMINATION_SPEC.md`.
- Gate, tie, family, ladder, depth, and explicit-rule verdict logic:
  `G1_ADEQUACY_GATES.md` and `G1_MODEL_LADDER_CONTRACT.md`.
- Seeds and byte determinism: `G1_SEED_CONTRACT.md`.

## Firewalls and terminal mapping

HELDOUT remains closed until a byte-addressed selection freeze exists. No MFC
threshold, grid, algorithm, scale, replicate, sampler, metric, or gate changes
after results. Preserve the semantics and Fontana/mechanism firewalls. If
preflight, calibration, or freeze integrity fails, issue a methodological
blocked status, not a scientific Voynich verdict. Otherwise use the fixed
mapping for `TOKEN_FORMATION_DEPTH` and
`EXPLICIT_RULE_GRAMMAR_REQUIRED`; issue `TOKEN_GRAMMAR_FROZEN` only after all
confirmatory rows and transcription statuses are complete.
