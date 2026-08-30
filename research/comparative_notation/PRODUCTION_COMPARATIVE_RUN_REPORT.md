# Production comparative run report

This report is technical only: it records what was computed and whether it validated. It draws no conclusion about the Voynich Manuscript's origin, mechanism, or historical identity; that is explicitly deferred to a separate task.

- run_id: `CNS-PROD01-RUN-20260830T182652Z`
- authorization reference: `research/comparative_notation/PRODUCTION_RUN_AUTHORIZATION.json` (sha256 `588bbc0a6300c7d7137410660595f5df66760de5366f0638842b9171e733164c`)
- git revision: `2fdb30885374759632133e2b458966acfb6cdca9`
- frozen corpus selection: C01, C02, C06 (sha256 `d802859a84760f1ec550a13f44a7da2aeecf95564cd5671fa0645aa752d56b17`)
- C06 representation set: MUSIC-R1, MUSIC-R2, MUSIC-R3 (three representations of one candidate, not independent candidates)
- measurements executed (raw metric rows, all candidates/representations): 394
- rarefaction draws executed: 81800
- bootstrap rows executed: 394

## Validation summary

# Post-run validation report

| Check | Status |
|---|---|
| `input_integrity` | PASS |
| `coverage` | PASS |
| `rarefaction` | PASS |
| `bootstrap` | PASS |
| `calibration` | PASS |
| `vm_reference` | PASS |
| `numeric_integrity` | PASS |
| `aggregation` | PASS |

Candidates present: `C01,C02,C06`

Representations seen: `C01/LATIN-EXPANDED,C02/LATIN-DIPLOMATIC,C06/MUSIC-R1,C06/MUSIC-R2,C06/MUSIC-R3`

## Reproducibility summary

# Reproducibility report

Second pass run id: `CNS-PROD01-RUN-20260830T182652Z-REPRO`, directory: `(temporary, outside the repository; removed after comparison)`.

Scientific files compared byte-for-byte (metadata/run-id/timestamp files excluded): 84.

Result: **PASS**

## Excluded / not-applicable operations

- C03-C05, C07-C09: excluded/deferred per the frozen production corpus selection; not revisited.
- Within-class pair distances and variance (CLASS_SUMMARY/WITHIN_CLASS_DISTANCES): NOT_APPLICABLE_FOR_CURRENT_PANEL — every class has exactly one independent corpus.
- Cross-class ranking, PCA, UMAP, nearest-neighbour analysis: OUT_OF_SCOPE_REPOSITORY_LOCKED.
- Checkpoints 20000 and 39380: NOT_COMPARABLE for every candidate in this subset (all observed sizes are below both).
- STRUCTURALLY_CLOSE_ON/STRUCTURALLY_DISTANT_ON verdicts: left PENDING — no frozen numeric threshold exists for this classification.

## Result bundle

`research/comparative_notation/production_runs/CNS-PROD01-RUN-20260830T182652Z`

Checksums: `research/comparative_notation/production_runs/CNS-PROD01-RUN-20260830T182652Z/SHA256SUMS`

## Final validity status

```text
PRODUCTION_COMPARATIVE_RUN_COMPLETED=true
PRODUCTION_COMPARATIVE_RUN_VALID=true
```

PRODUCTION_COMPARATIVE_RUN_VALID=true
