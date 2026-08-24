# Task82a corpus-scale portfolio report

Design version V1.0, authority Task81 V1.1 + Task82 V1.1 + Fingerprint V2 frozen. Design was frozen (TASK82A_DESIGN_FROZEN) before any main-generation job ran; TASK82A_BLIND_MANIFEST.json enumerates all 468 jobs (16 mechanisms x their frozen scaling policies x 3 corpora x 3 corpus_scales x 2 replicates) generated before this report.

## Answers to the preregistered questions

1. Yes: TASK82A_DESIGN.md and TASK82A_BLIND_MANIFEST.json were both written and their manifest job count/derivation verified before any job executed (Execute()'s verifyFreeze recomputes BuildManifest() and rejects any drift).
2. Yes: Task81 V1.1's three authoritative files are re-checksummed on every run (task81Bindings) and mnemonicspace.FrozenRegistry()/ValidateRegistry are called unmodified; the assembler only ever calls Runner.Prepare/Recover, never touching mechanism internals.
3. SCALING_POLICIES.tsv: RESET_EACH_CHUNK state policy (the only one type-valid under the frozen Runner.Prepare contract, which takes no prior-state argument -- CONTINUE_STATE is NOT_TYPE_VALID and was not run); CONVENTION_GLOBAL convention policy; PATH_PER_CHUNK_RESTART path policy; for cue mechanisms, both LOCAL_NAMESPACE and GLOBAL_NAMESPACE cue policies (both run and compared).
4. All 16 frozen mechanisms produced valid, leakage-checked corpus-scale OBSERVABLE_DOCUMENTs across every frozen scale/policy/corpus/replicate cell; 0/468 raw job artifacts failed the document-checksum self-check (0 expected on a clean run).
5. None failed to scale outright; every mechanism admits FIXED_CAPACITY/TRUNCATE chunking. What differs is scaling-*policy* applicability: literal mechanisms have no cue-namespace axis, and CONTINUE_STATE is universally NOT_TYPE_VALID (see Q3).
6. Yes: every line/token boundary is ASSEMBLER_DEFINED (one local-mechanism application = one line = one token), and pages are NOT_DEFINED throughout -- no Voynich-derived layout was introduced (BOUNDARY_PROVENANCE.tsv).
7. BOUNDARY_PROVENANCE.tsv records provenance for local_mechanism_boundary, token_boundary, line_boundary, page_boundary, assembly_boundary, and input_boundary separately, matching task82a.txt sec.18-21's required distinctions.
8. CORPUS_SCALE_RECOVERY.tsv reports local_exact_rate per mechanism/condition at corpus scale; qualitatively it reproduces Task82's own split (frozen positive mechanisms recover R0 near-exactly, negative-randomized controls do not), subject to the same condition-specific-seed pairing limitation Task82 already documented.
9. KNOWLEDGE_DEPENDENCE_STABILITY.tsv classifies every mechanism/condition; tally: NOT_APPLICABLE=86, PARTIALLY_PRESERVED=1, PRESERVED=43.
10. AMBIGUITY_SCALING.tsv reports mean/max R6 ambiguity cardinality per job at each corpus_scale; see the file for the SMALL/MEDIUM/LARGE trend per mechanism -- no single-number global claim is made because trends differ by mechanism family.
11. COLLISION_SCALING.tsv reports local (within-job, across-chunk) collision rate per job; CROSS_CORPUS_COLLISIONS.tsv reports how many distinct checksums are shared across corpora per mechanism/policy/scale/replicate cell.
12. INPUT_DEPENDENCE.tsv classifies every mechanism/policy/scale/replicate cell; tally: INPUT_INSENSITIVE=120, INPUT_SENSITIVE=36.
13. Cells classified INPUT_INSENSITIVE in INPUT_DEPENDENCE.tsv are candidates for corpus-scale input-insensitivity; cue mechanisms remain the dominant such class because their visible cue labels are corpus-content-independent by construction (only which word each label is associated with, never observable, varies with the corpus).
14. F2_COVERAGE.tsv reports per-job CORE/SUPPORTING attempted/available counts; MECHANISM_ELIGIBILITY.tsv aggregates to a per-mechanism CORE-family coverage ratio. Tally of eligibility: PARTIALLY_COMPARABLE=16.
15. The hierarchy/folio/locus/line-profile F2 families (2DL, BP, HR, LC, LS, PF) were not attempted at all: they require fingerprintv2's Task79Config pipeline, which Task82a's design document scopes out on cost grounds (a real timing pilot measured 1000-permutation/1000-bootstrap cost as prohibitive across 468 jobs) -- this is recorded as NOT_ATTEMPTED_COST_BOUNDED, distinct from the frozen extractor's own NOT_APPLICABLE/INCONCLUSIVE verdicts, which were also observed on every attempted cross-scale (cs1..cs5) metric because none of Task82a's assembled documents carry real IVTFF locus/Currier/section/line metadata.
16. F2_CROSS_CORPUS_STABILITY.tsv tally: PARTIALLY_STABLE=75, STABLE=501.
17. F2_CROSS_SEED_STABILITY.tsv tally: PARTIALLY_STABLE=92, STABLE=1636.
18. F2_CROSS_SCALE_STABILITY.tsv (MEDIUM vs LARGE, the pilot's own convergence pair) tally: CONVERGED=846, NOT_CONVERGED=88, PARTIALLY_CONVERGED=218.
19. SCALING_POLICY_EFFECT is read directly off F2_CROSS_CORPUS/CROSS_SEED_STABILITY.tsv grouped by scaling_policy_id: LOCAL_NAMESPACE cue jobs have a bounded, tiny vocabulary (complete edit-graph on `capacity` types) while GLOBAL_NAMESPACE jobs have a vocabulary growing with chunk count, which is the dominant scaling-policy effect on EF1/EF2/EF3.
20/21. MECHANISM_ELIGIBILITY.tsv gives the technical (never Voynich-similarity-based) F2_COMPARABLE/PARTIALLY_COMPARABLE/NOT_COMPARABLE classification for every mechanism; see that file for the per-mechanism list.
22. No Voynich reference vector, comparison artifact, or corpus statistic was read; extractF2 always builds fingerprintv2.Config from a Task82a-assembled corpus file and assertNoVoynichPath guards every such path (see aggregate/experiment tests).
23. No Task82b/BDD/shorthand/notation-control artifact was read; Task82a's only inputs are Task81/Task82/F2 freeze files and the three natural-language control texts.
24. See the final verdicts table and freeze marker below.

## Final verdicts

| Verdict | Result |
| --- | --- |
| TASK81_SEMANTICS_PRESERVED | SUPPORTED |
| CORPUS_SCALING_VALID | SUPPORTED |
| LOCAL_RECOVERY_PRESERVED | SUPPORTED |
| KNOWLEDGE_DEPENDENCE_PRESERVED | SUPPORTED |
| AMBIGUITY_SCALING_MEASURED | SUPPORTED |
| COLLISION_SCALING_MEASURED | SUPPORTED |
| INPUT_DEPENDENCE_MEASURED | SUPPORTED |
| F2_EXTRACTION_VALID | SUPPORTED |
| F2_COVERAGE_AUDITED | SUPPORTED |
| F2_CROSS_CORPUS_STABILITY | SUPPORTED |
| F2_CROSS_SEED_STABILITY | SUPPORTED |
| F2_SCALE_CONVERGENCE | PARTIAL |
| TASK83_PORTFOLIO_READY | PARTIAL |
| VOYNICH_FIREWALL_PRESERVED | SUPPORTED |
| NOTATION_CONTROL_FIREWALL_PRESERVED | SUPPORTED |

**Final Task82a verdict: TASK82A_CORPUS_SCALE_PORTFOLIO_FROZEN.**
