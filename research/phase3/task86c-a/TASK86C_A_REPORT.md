# Task86C-a — G1 failure decomposition and latent-model adequacy audit

## Outcome

Task86C's `NONE` means that none of 4032 job × model executions satisfied the conjunction of the frozen aggregate predictive, structural, complexity and failure gates. It does **not** establish that M0–M5 fail to describe the controls. The frozen results omitted the individual evidence needed to make that inference.

Integrity checks account for 672 jobs and their ledger hashes; frozen aggregate artifact hashes and the Task86C terminal marker were verified. Integrity issues: NONE. No corpus file was opened, no model was fitted/generated, and no Voynich path was accessed.

## Failure accounting

There were 4032 model executions. 3394 reached Stage D with a fitted model and 638 recorded `TRAINING_FAILED`, affecting 319 jobs. Fit-success counts are M0=672, M1=672, M2=672, M3=353, M4=353, M5=672. All 4032 aggregate predictive gates failed; 684 aggregate structural gates passed; no final model was sufficient. Retained independent failure events are COMPLEXITY_UNBOUNDED=56, NEGATIVE_EXHAUSTED=3054, NUMERICALLY_UNSTABLE=2839, TRAINING_FAILED=638. Training failure is computational evidence, not model-class inadequacy.

PM6 produced a valid score in 340/4032 executions. The remaining cells are unavailable (mostly explicitly `NEGATIVE_EXHAUSTED`) or not reached after training failure. Crucially, Task86C did not retain whether any valid PM6 value passed its threshold.

## PM-by-PM answers

For PM1, PM2, PM4 and PM5, values, finiteness and gate outcomes are `NOT_RECORDED` for every fitted execution; rates therefore are not estimable. For PM6, validity/availability is recorded, but value and pass/fail are not. Consequently `PREDICTIVE_ADEQUACY_WITHOUT_PM6`, CF1 and CF2 are `NOT_ASSESSABLE`, and the number of FAIL→PASS* changes is unknown.

`PM6_VETO_ROLE = INCONCLUSIVE`. It cannot be called the sole universal veto: PM6 gate passage is absent, while all predictive aggregates fail even in the 340 executions with a valid PM6 score. Nor can an independent universal PM be named because their individual outcomes were discarded. `PRE_PM6_PREDICTIVE_INFORMATION = NON_INFORMATIVE` in the frozen artifact layer.

## Synthetic controls

Expected-class fitting and aggregate paths are listed by generator, scale and replicate. M0/M1/M2/M5 have no recorded training failures; M3/M4 have substantial induction failures. However the correct models' PM1/2/4/5 values, PM6 gate and structural families were not frozen. Thus latent recovery rates, context/depth recovery, fitted-vs-generating M0 distribution, M3/M4 state/operation diagnostics and M5 coverage/rule diagnostics are not computable without new experiments.

`SYNTHETIC_LATENT_RECOVERY = NOT_ASSESSABLE`; observed historical recovery remains 0%. M0/M1/M2 recovery diagnostics are `PARTIAL` only in the narrow sense that fitting survival is known while adequacy is not. M3/M4/M5 limitations are `MIXED`: retained induction/protocol/generation failures coexist with missing measurement evidence. Lower/higher-class discrimination is likewise not assessable.

## Natural-language controls

English, Latin and Sanskrit each have a complete job × M0–M5 bookkeeping matrix. It establishes fitting survival and aggregate rejection, but cannot show which models passed PM1/2/4/5 or which structural family failed. Therefore:

- `ENGLISH_LATENT_MODEL_SET = NOT_ASSESSABLE`
- `LATIN_LATENT_MODEL_SET = NOT_ASSESSABLE`
- `SANSKRIT_LATENT_MODEL_SET = NOT_ASSESSABLE`
- `NATURAL_LANGUAGE_LATENT_ADEQUACY = NOT_ASSESSABLE`

The evidence supports terminal interpretation C: existing artifacts are insufficient. Repeating `English = NONE` would not prove that M0–M5 do not model English; the same applies to Latin and Sanskrit.

## Structural, scale and replicate evidence

Structural evaluation was executed independently of predictive passage for fitted models, so aggregate false is retained as `FAIL`, not `NOT_REACHED`; training failures are `NOT_REACHED`. Individual F2 values/families/thresholds are `NOT_RECORDED`. Hence one-family failures and closeness cannot be identified.

Scale and replicate tables report stable distributions of the defensible first failure (`TRAINING` or aggregate `PREDICTIVE_GATE`). They cannot reveal whether a specific PM improves with scale, whether PM6 worsens, or whether dependency depth recovers. No strong metric-specific scale conclusion is possible.

## Required verdicts and next task

The primary G1-v1 failure source is `UNRESOLVED`: measurement observability is certainly defective, but retained independent induction, negative-construction and generation failures prevent attribution solely to protocol. `TASK85B_READY = SUPPORTED` because concrete observability and known-correct-control requirements can be stated. `TASK86C_A_VALID = NOT_SUPPORTED` because the definition of done requires decompositions that the authoritative frozen source does not contain.

The first universal recorded rejection point is the aggregate `PREDICTIVE_GATE`; the responsible PM is not recorded. Exact/minimal recovery are zero because every aggregate predictive gate failed and sufficiency requires predictive passage, structural passage, multi-scale sufficiency and no failure class.

## Direct answers to the 32 report questions

1. `NONE` is failure of the conjunction, not proof of model-space inadequacy. 2. No job was absent before fitting began; 319 jobs contain at least one class training failure. 3. 3394 model executions fitted. 4–8. PM1/2/4/5 rates are `NOT_RECORDED`; PM6 validity is 340/4032, but PM6 pass rate is `NOT_RECORDED`. 9. PM6 role is inconclusive. 10. Training, negative construction, generation and complexity failures exist independently as named events, though their temporal relation to hidden PM gates is limited. 11–16. Expected M0–M5 paths are in the survival table; individual PM causes are missing. 17. Synthetic latent recovery count is not assessable. 18. All predictive aggregates failed. 19. Natural fit outcomes are in the matrix. 20. PM1/2/4/5 passage is not recorded. 21. Only aggregate structural failures are known. 22–24. All latent model sets are not assessable. 25. No. 26. Measurement observability is a major defect, but primary causation remains unresolved. 27. Only recorded aggregate/failure scale effects can be tabulated. 28. Their replicate agreement is tabulated. 29. Aggregate predictive gate. 30. M3/M4 training, widespread negative exhaustion, numeric generation instability and some complexity-unbounded events. 31. G1-v2 properties are enumerated. 32. Yes, for Task85b requirements, not for a new adequacy verdict.

TASK86C_A_EVIDENCE_INSUFFICIENT
