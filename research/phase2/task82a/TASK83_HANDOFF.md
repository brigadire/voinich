# Task83 handoff

This file contains only frozen Task82a portfolio paths, checksums, mechanism/scaling-policy IDs, F2 coverage, convergence status, and comparison eligibility. It does not include, reference, or imply any Voynich comparison result.

Frozen portfolio root: `research/phase2/task82a/`

Results manifest checksum: `c06d8c9e48abdcc31d373d4acb8701ca9335c51e1ce885b1ac6140050fb956f6`

Frozen artifact paths (see TASK82A_RESULTS_MANIFEST.json for individual checksums):

- `research/phase2/task82a/TASK82A_BLIND_MANIFEST.json`
- `research/phase2/task82a/TASK82A_COST_MODEL.tsv`
- `research/phase2/task82a/TASK82A_JOB_LEDGER.tsv`
- `research/phase2/task82a/SCALING_POLICIES.tsv`
- `research/phase2/task82a/BOUNDARY_PROVENANCE.tsv`
- `research/phase2/task82a/CORPUS_SCALE_TRANSFORMATION.tsv`
- `research/phase2/task82a/CORPUS_SCALE_RECOVERY.tsv`
- `research/phase2/task82a/KNOWLEDGE_DEPENDENCE_STABILITY.tsv`
- `research/phase2/task82a/COLLISION_SCALING.tsv`
- `research/phase2/task82a/AMBIGUITY_SCALING.tsv`
- `research/phase2/task82a/INPUT_DEPENDENCE.tsv`
- `research/phase2/task82a/F2_RAW_VECTORS.jsonl`
- `research/phase2/task82a/F2_COVERAGE.tsv`
- `research/phase2/task82a/F2_CROSS_CORPUS_STABILITY.tsv`
- `research/phase2/task82a/F2_CROSS_SEED_STABILITY.tsv`
- `research/phase2/task82a/F2_CROSS_SCALE_STABILITY.tsv`
- `research/phase2/task82a/MECHANISM_ELIGIBILITY.tsv`
- `research/phase2/task82a/TASK82A_REPORT.md`

## Mechanism eligibility (technical only, no Voynich similarity)

```
mechanism_id	core_families_available	core_family_coverage_ratio	artifact_valid	eligibility
f01_speculum_core	EF	0.142857	true	PARTIALLY_COMPARABLE
f01_speculum_profile_latin23_r12	EF	0.142857	true	PARTIALLY_COMPARABLE
f08_serpens_core	EF	0.142857	true	PARTIALLY_COMPARABLE
f11_arismetricum_core	EF	0.142857	true	PARTIALLY_COMPARABLE
f12_horalogius_core	EF	0.142857	true	PARTIALLY_COMPARABLE
m_restricted_rotation_index	EF	0.142857	true	PARTIALLY_COMPARABLE
m_restricted_storage_associate	EF	0.142857	true	PARTIALLY_COMPARABLE
negative_randomized_convention	EF	0.142857	true	PARTIALLY_COMPARABLE
negative_randomized_cue_association	EF	0.142857	true	PARTIALLY_COMPARABLE
negative_randomized_index_mapping	EF	0.142857	true	PARTIALLY_COMPARABLE
negative_randomized_path	EF	0.142857	true	PARTIALLY_COMPARABLE
synthetic_ambiguous	EF	0.142857	true	PARTIALLY_COMPARABLE
synthetic_cue_based	EF	0.142857	true	PARTIALLY_COMPARABLE
synthetic_cyclic_state	EF	0.142857	true	PARTIALLY_COMPARABLE
synthetic_indexed_lookup	EF	0.142857	true	PARTIALLY_COMPARABLE
synthetic_literal_storage	EF	0.142857	true	PARTIALLY_COMPARABLE
```

## Known limitations

- Only the edit-family (EF)/lexical-paradigm (LP) F2 families and task77's cross-scale (cs1-cs7) family were attempted; hierarchy-dependent families (2DL, BP, HR, LC, LS, PF) require fingerprintv2's Task79Config pipeline and were out of scope on cost grounds (see TASK82A_DESIGN.md).
- cs3/cs4/cs5 cross-scale metrics are structurally NOT_APPLICABLE for every job: Task82a's assembled documents never carry real IVTFF locus/Currier/section metadata.
- CONTINUE_STATE and CONVENTION_PER_BLOCK/PATH_REUSED_GLOBAL scaling policies were preregistered but not run; see SCALING_POLICIES.tsv for why.
