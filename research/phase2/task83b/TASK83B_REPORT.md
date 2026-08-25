# Task83b report

## Outcome

Fingerprint V2 was reconstructed from the authoritative ZL3b and IT2a raw inputs after systematically removing execution-order dependence from the F2 implementation. Three independent full reconstructions (`GOMAXPROCS=1`, `2`, and default/NCPU) produced 75 normative files each and all corresponding SHA-256 values are identical. The normal authoritative verifier exits 0.

The scientific definitions and all 13 CORE statuses are unchanged. Ten of 66 recorded metric values changed; most differences are floating-point reduction effects, while the intended deterministic reshuffling changes the null estimate for `LS3_BOUNDARY_LENGTH_ASYMMETRY` and related stochastic fields. One non-CORE decision changes: IT2a `HR1_LOCUS_VARIANCE_SHARE` moves from `NOT_SUPPORTED` to `SUPPORTED`. Cross-transcription classification, PF4, HR3/HR5, control ordering, Pareto membership, and the final Task79c structural conclusions do not change. This is therefore a successful scientific refreeze, not a claim that every historical number was precise.

```
FINGERPRINT_VERSION = FINGERPRINT_V2.1_DETERMINISTIC_SCIENTIFIC_REFREEZE
FINGERPRINT_V2_DETERMINISTIC_SCIENTIFIC_REFROZEN
TASK83R_READY = SUPPORTED
```

## Required questions

1. **Nondeterminism sources:** PRNG consumption through unordered maps; unordered floating-point reductions; ordering derived from map iteration; and map-backed dominance traversal. No time seed, shared RNG, unstable filesystem walk, or schedule-dependent serializer was retained.
2. **Present in Task79c:** the affected F2 statistical and aggregate paths were Task79/Task79c producer code; the defect registry identifies each function and artifact closure.
3. **Affected F2 metrics:** LS3, HR1, PF4, PF5, CS2, CS7 and graph-derived/reduction-dependent outputs; conservative recomputation covered every F2 branch and aggregate.
4. **Affected artifacts:** ZL3b, IT2a, controls, PF4/hierarchy, combined discriminative validation, and distance/Pareto artifacts.
5. **Semantics preserved:** yes. Every change is classified `IMPLEMENTATION_DETERMINISM_FIX`.
6. **Any F2 definition changed:** no; all 33 refrozen registry rows are machine-classified `IDENTICAL`.
7. **PRNG stream:** frozen algorithm and base seeds, with canonical ordering of immutable job identifiers before draws; independent tools retain their frozen seeds and replicate identifiers.
8. **Depends on map order:** no, covered by insertion-order regression tests.
9. **Depends on GOMAXPROCS:** no; runs at 1, 2, and default/NCPU are byte-identical.
10. **Depends on worker count:** no observable dependency; independent branches and stable aggregation give identical outputs under the execution variations used.
11. **Depends on process restart:** no; RUN_A/B/C are separate OS processes/workspaces.
12. **IT2a prepared corpus reproduced:** yes, SHA-256 `10286ee7e11ad974e9d0f884e3b0df1b588745a4b77ad428a638a5ff63946a8b` in every run.
13. **ZL3b prepared corpus reproduced:** yes, SHA-256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692` in every run.
14. **Numeric changes:** 10/66 metric rows differ; see `F2_DETERMINISTIC_EFFECT_AUDIT.tsv` for exact old/new values and deltas.
15. **Magnitude:** most observed-value changes are at floating-point rounding scale; the audit reports relative and standardized deltas, including material null-estimate changes where map order formerly redirected PRNG draws.
16. **Monte Carlo changes:** the exact old/new effects and p-values for all 66 rows are in `MONTE_CARLO_REFREEZE_AUDIT.tsv`.
17. **Statistical decisions:** one non-CORE decision changes, IT2a `HR1_LOCUS_VARIANCE_SHARE`; all other recorded decisions are unchanged.
18. **13 CORE statuses:** preserved.
19. **Cross-transcription stability:** preserved: 3 `STABLE`, 10 `DIRECTION_STABLE`, 0 `UNSTABLE`.
20. **PF4:** remains `SUPPORTED`, effect 8.5369860763733865 SD and p=0.000999000999000999.
21. **HR3/HR5:** remains `INCONCLUSIVE`.
22. **Control ordering:** unchanged.
23. **Pareto conclusions:** unchanged.
24. **Final Task79c verdicts:** unchanged at the structural/final level; the one non-CORE decision change is explicitly retained in the effect audits.
25. **RUN_A/B/C byte identity:** yes, all 75 normative/reconstruction-input rows.
26. **Authoritative verifier:** yes, exits 0 without `-allow-non-authoritative`.
27. **Transitive provenance:** complete, acyclic, checksum-valid, parent-bound, and verifier-validated.
28. **New version:** Fingerprint V2.1 deterministic scientific refreeze.
29. **F2 definitions changed:** no.
30. **Downstream freezes:** Task81, Task82, Task82a, Task82a.1, and Task82b are target-blind and classified `UNAFFECTED`.
31. **Task81/82 blind portfolios valid:** yes; they were neither recomputed nor retuned.
32. **Task82a.1 comparison contract valid:** yes; normalization is pre-target and contains no old target-derived constants.
33. **Task82b portfolio valid:** yes; it remains target-blind and unchanged.
34. **Quarantined Task83 results used:** **NO.**
35. **Fontana/shorthand/extraction model selection performed:** **NO.** Historical controls were used only in their frozen Task79c validation role.
36. **Ready for Task83r:** yes, subject to the quarantine and artifact restrictions in `TASK83R_HANDOFF.md`.

## Final verdicts

```
DETERMINISM_AUDIT_COMPLETE = SUPPORTED
HISTORICAL_NONDETERMINISM_REPRODUCED = SUPPORTED
DETERMINISM_FIX_COMPLETE = SUPPORTED
SCIENTIFIC_DEFINITIONS_PRESERVED = SUPPORTED
IT2A_DETERMINISTIC_REPRODUCIBILITY = SUPPORTED
ZL3B_DETERMINISTIC_REPRODUCIBILITY = SUPPORTED
MONTE_CARLO_DETERMINISM = SUPPORTED
PARALLELISM_INDEPENDENCE = SUPPORTED
PROCESS_RESTART_INDEPENDENCE = SUPPORTED
MULTIRUN_BYTE_IDENTITY = SUPPORTED
F2_NUMERIC_VALUES_CHANGED = YES
CORE_STATUS_CHANGED = NO
CROSS_TRANSCRIPTION_VERDICT_CHANGED = NO
PF4_VERDICT_CHANGED = NO
HIERARCHY_VERDICT_CHANGED = NO
CONTROL_ORDERING_CHANGED = NO
TASK79C_FINAL_VERDICTS_CHANGED = NO
STRUCTURAL_CONCLUSIONS_REPLICATED_UNDER_DETERMINISM = SUPPORTED
TRANSITIVE_PROVENANCE_COMPLETE = SUPPORTED
AUTHORITATIVE_VERIFIER = SUPPORTED
DOWNSTREAM_BLIND_PORTFOLIOS_VALID = SUPPORTED
TASK83_COMPARISON_CONTRACT_VALID = SUPPORTED
TASK83R_READY = SUPPORTED
```

The unresolved historical `3fb953…` value remains recorded only as `HISTORICAL_UNRESOLVED_METADATA / NON_AUTHORITATIVE`. The old manifests are unchanged, and Task83 remains `TASK83_EXPERIMENT_INVALID` and quarantined.

## Validation

The refreeze passed multi-process byte identity and the authoritative provenance verifier. Repository validation comprises `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`, `git diff --check`, manifest checksum/binding verification, and verifier negative tests. The final command results are recorded at completion, not inferred from the scientific runs.
