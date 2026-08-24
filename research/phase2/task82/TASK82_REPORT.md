# Task82 blind results report

All 672 frozen manifest jobs completed and are checksum-accounted. All Task81
V1.1 and Task80 bindings matched; there were no implementation/resource
failures, leakage failures, or silent exclusions. Two pre-freeze implementation
defects were caught by validation, corrected without changing Task81 semantics,
and all jobs were regenerated; see BUG_AUDIT.tsv. Deterministic regeneration
and aggregate-from-raw checks are covered by the runner verification tests.

## Answers to the preregistered questions

* Exact at R0: f01_speculum_core, f01_speculum_profile_latin23_r12, f08_serpens_core, f11_arismetricum_core, f12_horalogius_core, m_restricted_rotation_index, m_restricted_storage_associate, synthetic_ambiguous, synthetic_cue_based, synthetic_indexed_lookup, synthetic_literal_storage.
* Intrinsically non-recovering at R0: negative_randomized_convention, negative_randomized_cue_association, negative_randomized_index_mapping, synthetic_cyclic_state. Observable-only ambiguity occurs in:
  f01_speculum_core, f01_speculum_profile_latin23_r12, f08_serpens_core, negative_randomized_convention, negative_randomized_path, synthetic_literal_storage.
* Convention dependence: f01_speculum_core, f01_speculum_profile_latin23_r12, f08_serpens_core, f11_arismetricum_core, m_restricted_rotation_index, negative_randomized_path, synthetic_indexed_lookup, synthetic_literal_storage. Geometry/path dependence: f08_serpens_core, synthetic_literal_storage. Internal-memory
  dependence: f12_horalogius_core, m_restricted_storage_associate, synthetic_ambiguous, synthetic_cue_based. Context dependence: f11_arismetricum_core, f12_horalogius_core, m_restricted_rotation_index, m_restricted_storage_associate, synthetic_ambiguous, synthetic_cue_based, synthetic_indexed_lookup. No frozen runnable mechanism declares
  history as necessary; history removals are NOT_APPLICABLE.
* The randomized convention/path/association/index controls fail to recover the
  intended input under wrong knowledge, demonstrating specificity rather than
  mere extra bits. Full result classes and exact-value checks are retained.
* Collision groups are recorded without resolving them through hidden input.
  Cue surfaces are predominantly input-insensitive; literal surfaces retain
  bounded corpus differences. Replicate and corpus results, variance components,
  parameter identifiability, and all frozen ablation contrasts are in the TSVs.
* EM0--EM4 classes are defined in the frozen design and reported per mechanism.
  Information destruction is separated from knowledge-dependent inaccessibility
  by R0 versus carrier-removal/R6 scores. Declared-but-empirically-redundant
  carriers are explicitly retained in CARRIER_NECESSITY.tsv.
* Generic F2 extraction was not preregistered as an invocation in Task81, so no
  F2 portfolio was generated; Task83 may use the frozen observable documents
  but must not regenerate Task82 outputs.

The frozen manifest has only one parameter point per canonical mechanism, so
within-mechanism parameter effects are not identifiable; PARAMETER_SENSITIVITY
states that limitation. The condition-specific F01 seeds also prevent strict
paired R0--R6 state contrasts and are flagged in every affected raw job.

No Voynich data/reference vector, Task79/79c result, BDD/notation-control result,
or shorthand trajectory was read or compared. No ranking, winner selection,
copyist fitting, or historical-intention claim is made.

## Required report checklist

1. Yes, all 672 frozen jobs are COMPLETE.
2. No freeze/checksum mismatch occurred.
3. No unintended scientific failure occurred; negative and cue-only controls
   produced their preregistered result classes.
4. Two implementation defects were found before results freeze, audited, fixed,
   and followed by full regeneration; none remains unresolved.
5. The exact-R0 mechanisms are listed above and in MECHANISM_SUMMARY.tsv.
6. Intrinsically ambiguous/lossy mechanisms are listed above; ambiguity remains
   explicit rather than being resolved from hidden input.
7. Convention dependence is listed above and quantified in KD.
8. Geometry/path dependence is listed above and quantified in KD.
9. History is unused; every R4 removal is NOT_APPLICABLE.
10. InternalMemoryState dependence is listed above and quantified in KD.
11. Context dependence is listed above and quantified in KD.
12. Yes, frozen wrong-knowledge controls show that specific shared knowledge is
    required where declared.
13. Observable collision groups are retained in COLLISIONS.tsv.
14. Observable-only ambiguity rates/cardinalities are in RECOVERY_RESULTS.tsv
    and MECHANISM_SUMMARY.tsv.
15. Every mechanism has an operational EM0--EM4 class.
16. Corpus dependence and descriptive variance are measured.
17. Replicate stability is measured under the preregistered rule.
18. No within-mechanism parameter effect is identifiable because each canonical
    mechanism has one frozen point; this is reported rather than fitted.
19. All 13 full/ablation links declared by the frozen blind manifest have
    descriptive contrasts (12 named ablation forms in the Task81 freeze).
20. Generic and negative controls behave as frozen and are validated separately.
21. Declared redundant carriers are explicitly marked REDUNDANT.
22. R0 versus R6/removal separates destruction from inaccessible information.
23. All observable documents are frozen; raw F2 was not preregistered and is not
    claimed ready.
24. The Voynich firewall was preserved.
25. The notation-control firewall was preserved.
26. Yes, the blind observable/recovery portfolio is frozen for confirmatory
    Task83, subject to the explicitly reported absence of raw F2 vectors.

## Final verdicts

| Verdict | Result |
| --- | --- |
| BLIND_MANIFEST_COMPLETENESS | SUPPORTED |
| FREEZE_INTEGRITY | SUPPORTED |
| DETERMINISTIC_REPRODUCIBILITY | SUPPORTED |
| OBSERVABLE_DOCUMENT_INTEGRITY | SUPPORTED |
| RECOVERY_CONTRACT_INTEGRITY | SUPPORTED |
| INFORMATION_ACCOUNTING_VALID | PARTIAL |
| CONTROL_BEHAVIOR_VALID | SUPPORTED |
| KNOWLEDGE_DEPENDENCE_MEASURABLE | PARTIAL |
| CROSS_CORPUS_STABILITY_MEASURED | SUPPORTED |
| PLAINTEXT_DEPENDENCE_MEASURED | SUPPORTED |
| ABLATION_ANALYSIS_COMPLETE | SUPPORTED |
| RAW_F2_PORTFOLIO_READY | NOT_SUPPORTED |
| VOYNICH_FIREWALL_PRESERVED | SUPPORTED |
| NOTATION_CONTROL_FIREWALL_PRESERVED | SUPPORTED |

**Final Task82 verdict: TASK82_BLIND_RESULTS_FROZEN.**
