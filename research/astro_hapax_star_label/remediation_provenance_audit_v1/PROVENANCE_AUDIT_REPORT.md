# Remediation provenance audit

## Finding

The CP-SAT implementation was found at `restricted_dictionary_bruteforce_remediation_v1/scripts/solver_cpsat.py`. It is a genuine OR-Tools model and imports successfully from the package-specific `/home/brigadire/.venv` environment. Fresh smoke checks reproduced 10/10 CP-SAT/oracle agreements and 20/20 exact-oracle agreements.

The previous remediation results therefore have an identifiable implementation path, but their provenance chain is not currently clean enough to treat the scientific result as verified. The package has zero git-tracked files in this checkout. Its general `SHA256SUMS` has a stale `REPRODUCIBILITY.md` entry, and `FROZEN_CODE_SHA256SUMS.txt` has a stale `SEALED_BENCHMARK_PROTOCOL.md` entry. These are concrete integrity failures, not evidence that the solver source is absent.

The raw hidden predictions and results are present and listed in the package ledger. Their linkage is only partial because the raw rows do not embed the solver source hash, OR-Tools version, command line, or execution log; those facts are supplied by surrounding manifests/protocol text.

```text
CP_SAT_CODE_FOUND=YES
CP_SAT_RAW_RESULT_LINK=PARTIAL
REMEDIATION_PROVENANCE=PARTIALLY_VERIFIED
PRIOR_REMEDIATION_RESULTS_TRUST_STATUS=UNVERIFIED_PENDING_REPAIR
EXACT_SCOPE_SCIENTIFIC_RESULT=NOT_EVALUATED
REAL_DATA_SEARCH_AUTHORIZED=NO
```

This audit does not repair or re-sign the old package, and it does not run any real STAR LABEL search. A new exact-scope qualification must use either a repaired, commit-bound remediation snapshot or a new solver version with fresh sealed seeds.
