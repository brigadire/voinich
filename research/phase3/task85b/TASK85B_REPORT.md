# Task85b report — frozen G1-v2 design

## Outcome

Task85b freezes a scientifically identifiable-by-validation, evidence-complete, distributed G1-v2 contract. It does not claim that G1-v2 has passed blind controls and it does not open Voynich.

## Required questions

1. G1-v1 was non-identifiable because every known class returned NONE, while unavailable metrics, induction/generation failures, and aggregate-only evidence were allowed to occupy the rejection path.
2. Proven: aggregate failure, zero recovery, widespread negative exhaustion, M3/M4 training failures, numeric generation failures, and missing intermediate evidence. Unresolved: individual PM1/2/4/5 behavior, PM6 acceptance, exact F2 causes, latent adequacy, and whether PM6 alone vetoed results.
3. M0 IID, M1 fixed local dependence, M2 variable memory, M3 deterministic finite state, M4 probabilistic finite state, and M5 explicit component/rule grammar are retained.
4. M1 selection ambiguity is completed; M3/M4 class evidence is separated from inducer outcomes; M5 gains a development-frozen productive backoff. Every change is classified in the change register.
5. Separate fit/induction/computational statuses precede class evidence. An inducer cap or convergence failure yields MODEL_NOT_IDENTIFIABLE, never MODEL_INADEQUATE.
6. PM1, PM2, PM4, PM5, and PM6 are retained.
7. PM1 becomes diagnostic because it duplicates PM2 at fixed N; PM4 becomes supporting because unseen-set availability is split-dependent; PM2/PM5/PM6 remain required with new thresholds/missingness semantics.
8. PM6 uses exact finite-complement, length-matched sampling with replacement and separates construction, score validity, and acceptance.
9. `|A|^l-|V_l|` is computed exactly and complement rank/unrank constructs every draw whenever nonzero; replacement avoids unique-sample exhaustion.
10. Unavailable/nonfinite required metrics produce NOT_ASSESSABLE; they never silently fail a model.
11. PredictiveAdequacy is PASS only when PM2, PM5, and PM6 pass; valid failure is FAIL; any unresolved required evidence makes NOT_ASSESSABLE.
12. The seven EDIT/LEXICAL_PARADIGM G1 metrics remain discriminating.
13. Metric → family → scale → replicate rules are frozen; both families must pass every required scale.
14. Borrowed-skeleton metrics are persisted diagnostics with zero adequacy weight.
15. Minimality uses adequate candidates, description length, frozen equivalence bounds, and requires all lower ranks to be validly resolved.
16. EQUIVALENT_SET is returned when multiple adequate candidates/classes occupy the lowest observational-equivalence component.
17. NOT_IDENTIFIABLE is returned whenever missing, computational, protocol, or comparison evidence could change the minimum.
18. Inputs, fit diagnostics, models, every PM/baseline/threshold/gate, negative-space proof, generation batch, every F2/family/scale/replicate gate, complexity edge, hashes, and statuses are persisted.
19. Yes by design and prototype: the verifier consumes only frozen evidence; it imports no fitting/generation implementation.
20. Negative tests remove PM/F2 evidence, change threshold/hash, map unavailable to FAIL, substitute class failure for induction failure, and simulate a duplicate conflict.
21. Two independently authored new generators per M0–M5 class, three scales, and four replicates validate recovery; old Task86C populations are diagnostic only.
22. Corpus IDs are opaque and ground truth/parameters/seeds are escrowed until blinded analysis results freeze.
23. Exact ≥0.70, minimal-compatible ≥0.85, equivalence-compatible ≥0.90, NONE ≤0.05, NOT_IDENTIFIABLE ≤0.10, with class floors and under/over-complexity limits in the recovery registry.
24. Fixed English, Latin, and Sanskrit corpora test non-degenerate ordinary-token applicability, not linguistic grammar discovery.
25. Confirmatory observations occur after threshold/config freeze; changing thresholds invalidates the manifest rather than updating it.
26. Path allowlists, config-hash closure, access logs, and a separate task boundary prohibit Voynich reads or tuning until all four Task86C-v2 validations pass.
27. Phase I provides a deterministic mTLS pull/lease coordinator, persistent workers, content-addressed input caching, checkpoints, retries, and deterministic reduction.
28. Queue, PKI, worker lifecycle, cache, JobID/checkpoint patterns, deployment, and reduction are reused; a G1 DAG/evidence adapter is added.
29. No new scheduler is needed. Existing infrastructure is partial only because it lacks the G1-v2 evidence graph and scientific manifest adapter.
30. Jobs are fit candidate, predictive candidate, generation scale×replicate×batch, structural family, and aggregation DAG nodes.
31. Yes. Job identity excludes worker and every immutable bundle may be leased to any compatible authenticated worker.
32. Workers pull the next ready job after completion; runtime estimates affect priority only.
33. An atomic verified completion index keyed by JobID reconstructs readiness; restart or worker-count change schedules only missing/unverified jobs.
34. Scientific failures are immutable outcomes and not retried; infrastructure failures may retry with identical identity.
35. M0/M1, M2, M3, M4, and M5 representatives run on multiple nodes before bulk work; byte identity is preferred and any tolerance is frozen beforehand.
36. Identical duplicates count once and retain provenance; conflicting copies are all quarantined and block descendants/aggregation.
37. Each job bundle carries transitive dependency and artifact hashes; aggregation requires the complete verified graph.
38. Capacity design estimates 216 corpus instances, ~15,552 candidate fits, up to 5,184 generation batches, ~36,288 family-scale F2 evaluations, 330–380 CPU-hours, and 60–100 GB.
39. Conservative node-wall estimates are 30h/16h/9h/5.5h for 1/2/4/8 nodes; acceptance instead uses measured efficiency ≥0.60 at four workers, utilization ≥0.70, overhead ≤0.10, and straggler share ≤0.25.
40. Synthetic recovery, natural-language applicability, evidence reconstruction, cross-node determinism, distributed manifest/failure tests, and scalability must all pass.
41. Task86C-v2 implements the frozen adapter/verifier, runs Stage A/B open validation, freezes thresholds, then blind C/D analysis, freezes, unblinds, and mechanically evaluates recovery.
42. Yes. Task87 remains blocked until an interpretable G1-v2 Voynich result; NONE/NOT_IDENTIFIABLE requires an explicit residual-first decision.

## Verdicts

```
G1V2_MODEL_LADDER_DEFINED = SUPPORTED
G1V2_PREDICTIVE_PROTOCOL_DEFINED = SUPPORTED
G1V2_NEGATIVE_PROTOCOL_DEFINED = SUPPORTED
G1V2_STRUCTURAL_PROTOCOL_DEFINED = SUPPORTED
G1V2_MINIMALITY_PROTOCOL_DEFINED = SUPPORTED
G1V2_FAILURE_SEMANTICS_DEFINED = SUPPORTED
G1V2_EVIDENCE_CHAIN_COMPLETE = SUPPORTED
G1V2_DECISION_RECONSTRUCTIBLE = SUPPORTED
G1V2_CONTROL_VALIDATION_DEFINED = SUPPORTED
G1V2_VOYNICH_FIREWALL_DEFINED = SUPPORTED
PHASE1_DISTRIBUTED_REUSE = PARTIAL
G1V2_DISTRIBUTED_EXECUTION_DEFINED = SUPPORTED
G1V2_HORIZONTAL_SCALABILITY_DEFINED = SUPPORTED
TASK86C_V2_READY = SUPPORTED
TASK86V_READY = NOT_SUPPORTED
TASK87_READY = NOT_SUPPORTED
```

`TASK86C_V2_READY` means the next experiment is specified and may implement/run the frozen contract; it does not mean validation has passed. Cross-node G1-v2 execution remains a Task86C-v2 Stage-A acceptance item. The local design tests validate schemas, three-valued gates, decision regeneration, corruption rejection, resume semantics, and duplicate conflict behavior without executing scientific models.

G1_V2_EXPERIMENT_CONTRACT_FROZEN.
