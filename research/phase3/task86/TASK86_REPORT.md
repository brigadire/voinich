# Task86 — token-formation grammar discovery

## Terminal result

`TASK86_EXPERIMENT_BLOCKED`

The authoritative inputs and freeze markers reproduce exactly, but the frozen
Task85 contract is not executable without adding outcome-relevant scientific
choices. The complete gap list and stop point are recorded in
`TASK86_DESIGN_EXECUTION.md` and `CONTRACT_PREFLIGHT.tsv`.

## Required questions

1. **Frozen inputs reproduced?** Yes, at the byte/checksum and freeze-sentinel
   level.
2. **Calibration before Voynich fitting?** No calibration was run; execution
   stopped before both calibration and fitting.
3. **M0-M5 implemented?** Not under this contract. Implementing unspecified
   variants would silently enlarge or redefine the frozen model space.
4. **Failed models?** None were fitted. Contract incompleteness is not one of
   the frozen model failure classes and is therefore not mislabeled as a model
   failure.
5. **Selected hyperparameters?** None.
6. **HELDOUT closed before selection freeze?** Yes. No selection freeze was
   created and no HELDOUT evaluation occurred.
7. **Best unseen-data predictor?** Not evaluated.
8. **Negative-token discrimination?** Not evaluated; its sampler is one of the
   blocking omissions.
9. **Productivity versus memorization?** Not evaluated.
10. **G1 structural gates?** Not evaluated.
11. **Cross-transcription stability?** Not evaluated; its tolerance is absent.
12. **M0-to-M5 ladder?** Not evaluated.
13. **Complexity justification?** Not evaluated.
14. **G1 grammar sufficient?** Not identifiable in a blocked experiment.
15. **Unique G1_min?** Not identifiable.
16. **G1 minimal class?** Not identifiable.
17. **Token-formation depth?** `NOT_IDENTIFIABLE`.
18. **Explicit component/rule grammar required?** Inconclusive.
19. **Simpler Markov/finite-state model sufficient?** Inconclusive.
20. **Unexplained formation properties?** Not measured.
21. **Task87 handoff?** No frozen G1 exists; Task87 is not ready.

## Verdict disposition

The Task86 section 43 evidence verdicts are not issued as positive or negative
scientific findings because no valid experiment occurred. Their disposition is:

- `G1_CALIBRATION_VALID = NOT_EVALUATED_BLOCKED`
- `G1_MODEL_SPACE_EXECUTED = NOT_SUPPORTED`
- `HELDOUT_FIREWALL_PRESERVED = SUPPORTED`
- `G1_PREDICTIVE_STRUCTURE = NOT_EVALUATED_BLOCKED`
- `G1_NEGATIVE_DISCRIMINATION = NOT_EVALUATED_BLOCKED`
- `G1_PRODUCTIVE_FORMATION = NOT_EVALUATED_BLOCKED`
- `G1_STRUCTURAL_REPRODUCTION = NOT_EVALUATED_BLOCKED`
- `G1_CROSS_TRANSCRIPTION_STABILITY = NOT_EVALUATED_BLOCKED`
- `G1_GRAMMAR_SUFFICIENT = NOT_EVALUATED_BLOCKED`
- `G1_MINIMAL_CLASS = NONE`
- `TOKEN_FORMATION_DEPTH = NOT_IDENTIFIABLE`
- `EXPLICIT_RULE_GRAMMAR_REQUIRED = INCONCLUSIVE`
- `G1_UNEXPLAINED_STRUCTURE = INCONCLUSIVE`
- `TASK87_READY = NOT_SUPPORTED`

`NOT_EVALUATED_BLOCKED` is deliberately outside the evidentiary enum: mapping
an unrun test to `NOT_SUPPORTED` would manufacture a negative scientific result.

## Interpretation boundary

This result says nothing about Voynich token structure, language status,
meaning, or generation mechanism. It is solely a reproducibility finding about
the frozen experimental specification.

