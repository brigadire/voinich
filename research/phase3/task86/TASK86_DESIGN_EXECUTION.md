# Task86 design execution

**Status:** `TASK86_EXPERIMENT_BLOCKED`

## Scope actually executed

Only the mandatory pre-fitting audit was executed. It verified the authoritative
corpus hashes, the Task85 split hashes, both Task85 freeze sentinels, and the
presence and checksums of every Task85 artifact named by Task86 section 1.

No corpus partition was parsed into tokens, no message-free population was
generated, no Voynich model was fitted, and no VALIDATION or HELDOUT statistic
was computed. Reading the corpus byte streams for the mandatory SHA-256 check
did not expose partition membership or token values to a selection procedure.

## Blocking contract gaps

The frozen Task85 contract does not determine a unique Task86 experiment:

1. `GRAMMAR_MODEL_REGISTRY.tsv` gives no finite numeric smoothing grid for M0
   or M1. M1 freezes only order `n in {1..6}`; it leaves the smoothing method
   and constant open.
2. M2 has no finite max-depth grid, back-off penalty grid, escape-probability
   grid, or uniquely specified context-tree/PPM algorithm.
3. M3/M4 have no merging-threshold grid, max-state grid, uniquely specified
   state-merging algorithm, or M4 probability-estimation grid. M3 also has no
   frozen generative tie-break, so the registry's own failure rule makes it
   `NON_GENERATIVE` for generation validation.
4. M5 has no slot-count grid, minimum-support grid, inventory-cap grid,
   component-mining algorithm, or unambiguous component-boundary rule.
5. MFC0-MFC2 name generator families but do not freeze concrete generator
   configurations, population sizes, calibration partitions, or the statistic
   used to turn their results into a "null spread".
6. The convergence rule requires a "pre-registered tolerance", but no numeric
   tolerance, replicate checkpoints, or complete scale grid is frozen.
7. The cross-transcription statuses require a pre-registered tolerance, but no
   such tolerance is present.
8. The failure registry refers to frozen thresholds for numerical instability
   and complexity-growth slope, but neither threshold exists in the named
   validation or complexity contract. The parameter-to-data ratio needed for
   `MEMORIZATION_DOMINATED` is likewise not specified.
9. The negative-token rule does not specify a deterministic sampler or a
   definition of glyph-frequency class. Therefore PM6 cannot be regenerated
   independently from the contract.
10. PM5 specifies a bucketed calibration error but freezes neither buckets nor
    the exact error functional.

These choices affect model selection, failure disposition, predictive adequacy,
structural adequacy, and minimality. Supplying them in Task86 would be a
substantive post-freeze design change, expressly prohibited by Task86 sections
1, 5, 6, 27, 29, and 39.

## Stop point

Execution stopped before the first MFC generation and before the first Voynich
fit. Consequently `G1_CALIBRATION_FROZEN` and
`GRAMMAR_MODEL_SELECTION_FROZEN` were not created, and HELDOUT remained closed
to evaluation. The required result tables are intentionally absent: empty or
fabricated rows would falsely represent an executed experiment.

