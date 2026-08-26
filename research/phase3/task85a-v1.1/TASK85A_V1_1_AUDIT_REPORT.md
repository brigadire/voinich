# Task85a-v1.1 post-freeze implementation-resolution audit

## Outcome

`TASK86R_CONFIRMATORY_INTEGRITY = PARTIAL`.

The primary negative conclusion survives, but Task86R was not a wholly clean
execution of the frozen procedure. R1 and R2 complete unspecified scientific
choices; R3, R4, and R6 change frozen/inherited procedures; and three additional
scientific changes were found. Therefore `TOKEN_GRAMMAR_FROZEN` is historically
preserved but must be read with this qualification layer.

Required verdicts:

```
R1_RNG = SCIENTIFIC_COMPLETION
R2_SELECTION = SCIENTIFIC_COMPLETION
R3_B2 = SCIENTIFIC_CHANGE
R4_STATE_MERGING = SCIENTIFIC_CHANGE
R5_GLYPH_ALIAS = EQUIVALENT_IMPLEMENTATION
R6_CALIBRATION = SCIENTIFIC_CHANGE
UNDISCLOSED_SCIENTIFIC_DOF = DETECTED
PM6_NEGATIVE_EXHAUSTION_ROBUST = SUPPORTED
M3_M4_FAILURE_EVIDENCE = CONFIRMATORY
M5_PRODUCTIVE_COVERAGE_GAP_ROBUST = SUPPORTED
STRUCTURAL_FAILURE_ROBUST = PARTIAL
G1_NONE_VERDICT_ROBUST = SUPPORTED
TOKEN_FORMATION_DEPTH_ROBUST = SUPPORTED
TASK87_NOT_READY_ROBUST = SUPPORTED
TASK86R_CONFIRMATORY_INTEGRITY = PARTIAL
CORRECTIVE_ACTION = REPORT_CORRECTION_ONLY
```

## Answers to the audit questions

1. The six disclosed resolutions are exactly R1-R6 in the registry. They were
   documented only after freeze; target-blindness does not make them
   implementation-only.
2. Task85a-v1.1 explicitly permits deterministic execution but does not license
   unspecified RNG stream/warm-up, PM2 selection, fixed B2 order, blue-fringe,
   or a population-as-seed-replicate reinterpretation.
3. None of R1-R6 meets the strict PURE burden. R5 is controlled-regression
   equivalent.
4. R5 required and passed equivalence tests. R1 and R4 fail equivalence because
   controlled counterfactuals differ.
5. Scientific completions exist in R1 and R2.
6. Scientific changes exist in R3, R4, R6, JS normalization, the negative
   attempt cap, and partial-calibration acceptance.
7. Undisclosed DOF were detected: 20,000 negative attempts; a nonnormalized JS
   denominator; and omission of failed/nonfinite calibration inputs while
   declaring calibration valid. Winner-only VALIDATION retention is also an
   auditability defect.
8. RNG implementations are not output-equivalent: omitting the unspecified
   warm-up changes the stream. Historical vectors are frozen in this audit.
9. VALIDATION was preregistered as the selection partition; PM2 argmin was not
   preregistered as the within-class statistic.
10. B2 was not preregistered as fixed order 2 with PM2-selected alpha; inherited
    text instead says order is selected on VALIDATION.
11. Exhaustive and blue-fringe are distinct. The synthetic counterexample has
    different states and accepted language.
12. R4 can change success/failure in general because pair examinations and caps
    differ. On the actual DEVELOPMENT inputs, the normalized exhaustive
    procedure independently reaches the same operation-cap failure.
13. M3/M4 failure therefore remains confirmatory; Task86R's claimed algorithmic
    equivalence does not.
14. Glyph aliases are bijective, reversible, deterministic, collision-free,
    order-preserving, EVA-safe, and F2-preserving within 1e-12.
15. The 4032 workload and one generation/job are fixed. Reusing independent
    populations as a seed-replicate axis is not the frozen replicate operation.
16. PM6 exhaustion is robust despite the undocumented search cap. The frozen
    unique-negative requirement creates a pigeonhole impossibility: ZL3b has
    231 HELDOUT singleton occurrences but at most 15 unseen singleton glyph
    types; IT2a has 112 versus at most 6. Frequency-class and M5 restrictions
    only reduce those upper bounds.
17. M5's coverage gap is robust: the executed VALIDATION winner remains +Inf
    (no candidate with a finite PM2 displaced it) and HELDOUT PM1 is +Inf. This
    is the frozen zero-probability unseen behavior, not RNG or alias behavior.
18. The observed zero structural passes are not fully confirmatory: R1 changes
    generated populations and R2/R6/U3 can change candidates/thresholds. Hence
    `STRUCTURAL_FAILURE_ROBUST = PARTIAL`.
19. PredictiveAdequacy failure is robust because PM6 is required and is
    candidate-independently unconstructible on both transcriptions.
20. Consequently `G1_MINIMAL_CLASS=NONE` remains robust.
21. The frozen no-adequate-candidate mapping preserves
    `TOKEN_FORMATION_DEPTH=NOT_IDENTIFIABLE`.
22. `TASK87_READY=NOT_SUPPORTED` remains robust because there is no adequate
    frozen G1 input.
23. The historical marker is not fully confirmatory as a procedure claim even
    though its primary NONE value survives.
24. Qualification is required; exactly one qualified integrity marker is added.
25. Minimum corrective action is `REPORT_CORRECTION_ONLY`: preserve NONE and
    NOT_READY, retract full-integrity/equivalence and unconditional structural
    interpretations. A future rerun would first require a new target-blind
    contract version, but no rerun is necessary to support the primary result.

## Provenance and validation

Commit `0e60737` exists, is an ancestor of HEAD, and Task86R source/artifacts are
byte-identical between that commit and HEAD. Task85 and Task85a manifests and
their listed hashes pass `validate_contract.py`. Verified marker hashes include
Task85 freeze `a7895c...`, Task85a freeze `98acd9...`, calibration freeze
`68f75c...`, selection freeze `345d86...`, execution ledger `46be17...`, and
historical token marker `e65c68...`. All historical artifacts and markers are
unchanged.

No HELDOUT value was used for tuning or selection. HELDOUT was read only for
the required immutable-evidence regression and robustness proofs.

