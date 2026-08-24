# Task82a.1 report

## Audit and extension

All 33 frozen metrics were audited: 8 `DIRECTLY_APPLICABLE`, 9 `ASSEMBLER_APPLICABLE`, 5 `NOT_APPLICABLE_STRUCTURE`, 11 `NOT_APPLICABLE_METADATA`, and 0 `ESTIMATOR_INCOMPATIBLE`. Six metrics (the observed 2D-lite, boundary, and line estimators) were absent only because Task82a bounded the full Task79 pipeline by cost; all six were recovered across 468 immutable jobs. Existing EF/LP/cross-scale results were imported rather than recalculated.

The generic extraction is regression-tested against task79-v1. No mathematical definition, null, normalization, bootstrap unit, threshold, direction, or classification changed. `2DL1` retains the frozen implementation behavior despite a prose mismatch in the registry; correcting that mismatch would require a new fingerprint version. No synthetic hierarchy was introduced.

## Coverage and stability

Maximum direct CORE metric coverage is 3/13 (one direct CORE family); assembler projection adds 4/13 CORE metrics (three families), for 7/13 mathematically available CORE metrics. Hierarchy, locus, and folio CORE families remain unavailable because folios, physical lines, sections, loci, recto/verso, and manuscript metadata do not exist. Direct and assembler coverage are never combined for target distance.

Cross-corpus tally: PARTIALLY_STABLE=150, STABLE=1938. Cross-seed tally: PARTIALLY_STABLE=92, STABLE=3040. MEDIUM/LARGE scale tally: CONVERGED=1710, NOT_CONVERGED=88, PARTIALLY_CONVERGED=290. Exact rows and cue LOCAL/GLOBAL effects are in the corresponding TSV files. Inconclusive imported metrics remain distinct from structural non-applicability; vocabulary-collapse cases are marked `NOT_APPLICABLE_DATA_DEGENERACY` in sample diagnostics and are not removed.

## Required answers

1. Audited: 33. 2. Direct: 8. 3. Assembler: 9. 4. Structure-unavailable: 5. 5. Metadata-unavailable: 11. 6. Estimator-incompatible: 0. 7. Previously cost-bounded only: 6. 8. Recovered: all 6. 9. Genericized dependency: task79-v1 ordered-group observed estimators. 10. Mathematical definition changed: no. 11. Direct CORE: 3/13. 12. Assembler CORE: 4/13. 13-14. Hierarchy/locus/folio remain unavailable for absent manuscript structure/metadata. 15-18. Corpus, seed, scale, and cue-policy results are frozen in the stability tables. 19. Data-degenerate cells are explicit. 20. `F2_COMMON_DIRECT.tsv` is frozen. 21. `F2_ASSEMBLER_PROJECTION.tsv` is separate. 22. Eligibility is per mechanism/policy. 23. Task83 contract is frozen. 24-25. Both firewalls were preserved. 26. Portfolio readiness is partial because full-manuscript comparison is structurally impossible.

## Final verdicts

| Verdict | Result |
| --- | --- |
| TASK82A_INPUT_FREEZE_INTEGRITY | SUPPORTED |
| F2_APPLICABILITY_AUDIT_COMPLETE | SUPPORTED |
| F2_COST_VS_STRUCTURE_SEPARATED | SUPPORTED |
| NO_SYNTHETIC_HIERARCHY_INTRODUCED | SUPPORTED |
| F2_GENERIC_EXTRACTION_VALID | SUPPORTED |
| F2_DIRECT_CORE_COVERAGE | 3/13 PARTIAL |
| F2_PROJECTION_CORE_COVERAGE | 4/13 PARTIAL |
| F2_CROSS_CORPUS_STABILITY | PARTIAL |
| F2_CROSS_SEED_STABILITY | PARTIAL |
| F2_SCALE_STABILITY | PARTIAL |
| TASK83_COMPARISON_CONTRACT_FROZEN | SUPPORTED |
| TASK83_PORTFOLIO_READY | PARTIAL |
| VOYNICH_FIREWALL_PRESERVED | SUPPORTED |
| NOTATION_CONTROL_FIREWALL_PRESERVED | SUPPORTED |

**TASK82A_F2_COVERAGE_EXTENDED_FROZEN**
