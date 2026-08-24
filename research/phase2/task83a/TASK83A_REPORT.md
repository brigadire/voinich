# Task83a report

## Outcome

The IT corpus anomaly itself is a top-level manifest typo: documented and historical preprocessing deterministically produces `10286ee7…`, and authoritative Task79c IT JSON artifacts independently embed that same hash. However, the full historical analyzer did not numerically reproduce all seeded Monte Carlo fields because old statistical code consumes PRNG draws through unsorted Go-map traversals. This second, independent integrity defect means the strict manifest-only gate fails.

Therefore:

```
ROOT_CAUSE = MULTIPLE_PROVENANCE_ERRORS
SCIENTIFIC_RECOMPUTATION_REQUIRED = YES
TASK83R_READY = NOT_SUPPORTED
```

No corrected authoritative freeze is issued. The required manifest-named file is an explicitly non-authoritative, machine-verifiable provenance audit graph; it must not be treated as a refreeze. Exactly one outcome marker is issued: `FINGERPRINT_V2_PROVENANCE_UNRESOLVED`.

## Required questions

1. **Origin of `3fb953…`:** first and only known occurrence is manual metadata in the old freeze, introduced by Task79c commit `6f185579…`. Generating bytes/command are not recoverable: `PARTIAL`.
2. **Origin of `10286e…`:** raw IT2a through the documented Go-native x7-compatible extractor and canonical preparer. The sidecar, files, clean runs, and Task79c artifacts agree: `RESOLVED`.
3. **Hash reproduced from raw:** `10286ee7e11ad974e9d0f884e3b0df1b588745a4b77ad428a638a5ff63946a8b`.
4. **Deterministic preparation:** yes; two independent current runs plus a Task79c-source run are byte-identical.
5. **Corpus actually used by Task79c:** `10286e…`.
6. **Independent of the old manifest:** yes; `fingerprint.json`, `raw_results.json`, and the run freeze manifest each embed `10286e…`; the sidecar and regeneration agree.
7. **Matches intended semantics:** yes, point by point in `IT2A_PREPROCESSING_CONTRACT.md`.
8. **Difference between prepared corpora:** unknowable at byte level because no `3fb953…` bytes/object were found. The available `10286e…` corpus is 229,490 bytes, 5,207 lines, 37,945 tokens.
9. **Only byte-level differences?:** cannot be classified for an unavailable variant.
10. **Do variant differences change F2?:** not computable without prohibited hash-target fitting. No evidence shows a second corpus was ever used scientifically.
11. **Analogous ZL3b problem:** no; raw, manifest, regenerated prepared bytes, and embedded hashes agree.
12. **Other input problems:** all other checksum-bearing old-manifest inputs match. All eight control files also match their embedded artifact hashes.
13. **Normative input links:** corpus/config bindings are supported; aggregate parentage is recorded transitively. Numeric reproducibility of two historical analyzer artifacts is only partial.
14. **Task79c reproduction from raw:** attempted successfully end-to-end for IT and preparation for ZL; PF4/hierarchy and full distance/Pareto were regenerated.
15. **Byte-identical outputs:** prepared corpora, PF4/hierarchy, and full distance/Pareto yes. Full IT fingerprint no (paths plus numeric fields differ).
16. **Root cause:** `MULTIPLE_PROVENANCE_ERRORS`: old-manifest IT checksum typo plus independently discovered seeded-analysis nondeterminism.
17. **Manifest-only?:** the original checksum anomaly is, but Task83a as a whole fails the manifest-only gate.
18. **Scientific recomputation required?:** yes, after deterministic ordering is fixed, for all affected F2/null results and dependent aggregates.
19. **F2 values changed in diagnostic regeneration?:** yes, several at floating-point/Monte Carlo level; see effect audit.
20. **CORE statuses changed?:** no.
21. **Task79c verdicts changed?:** no; canonicalized verdict files are identical.
22. **Fingerprint scientific semantics changed?:** no new fingerprint was issued; current labels/definitions remain historical, but authority is unresolved.
23. **New marker:** `FINGERPRINT_V2_PROVENANCE_UNRESOLVED`.
24. **Old freeze preserved?:** yes, unchanged, along with Task83 invalidation.
25. **Can the verifier detect the Task83 class?:** yes; it verifies file hashes, parent hashes, embedded corpus hashes, graph completeness/cycles, and fails both required negative tests.
26. **Ready for Task83r?:** no.
27. **Quarantined Task83 comparisons used?:** **NO**.
28. **Fontana/shorthand/extraction model selection performed?:** **NO**.

## Final verdicts

```
IT2A_RAW_PROVENANCE = SUPPORTED
IT2A_PREPROCESSING_REPRODUCIBILITY = SUPPORTED
IT2A_HASH_3FB953_ORIGIN = PARTIAL
IT2A_HASH_10286E_ORIGIN = RESOLVED
TASK79C_IT2A_INPUT_BINDING = SUPPORTED
ZL3B_PROVENANCE_INTEGRITY = SUPPORTED
FULL_F2_INPUT_INTEGRITY = PARTIAL
FULL_F2_ARTIFACT_INTEGRITY = PARTIAL
TASK79C_REPRODUCIBILITY = PARTIAL
ROOT_CAUSE = MULTIPLE_PROVENANCE_ERRORS
SCIENTIFIC_RECOMPUTATION_REQUIRED = YES
F2_VALUES_CHANGED = YES
F2_CORE_STATUS_CHANGED = NO
TASK79C_VERDICTS_CHANGED = NO
FINGERPRINT_V2_SCIENTIFIC_SEMANTICS_CHANGED = NO
TRANSITIVE_PROVENANCE_VERIFIER = SUPPORTED
ORIGINAL_TASK83_FAILURE_DETECTABLE = SUPPORTED
TASK83R_READY = NOT_SUPPORTED
```

Task83 remains formally `TASK83_EXPERIMENT_INVALID`.

## Validation

`go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`, `git diff --check`, JSON parsing, results-manifest self-check, transitive graph audit, and the verifier's two required negative tests pass. The verifier's normal (authoritative) mode deliberately exits non-zero on this unresolved audit manifest; `-allow-non-authoritative` verifies its checksum/binding graph without opening Task83r.
