# Production comparative run preflight report

Run: `CNS-PROD01-20260830`. Scope: preflight and authorization decision for the frozen production corpus subset C01, C02, C06. The comparative run itself is explicitly not executed by this task.

| Gate | Status | Detail |
|---|---|---|
| `corpus_subset_and_candidate_bundles` | PASS | PRODUCTION_CORPUS_SELECTION/MANIFEST/STATUS/SHA256SUMS and C01,C02,C06 bundles (provenance, policy, normalization, USC, validation, reproducibility, checksums) are valid |
| `global_freeze` | FAIL | global freeze contract is incomplete or inconsistent: GLOBAL_FREEZE_REPORT.md is absent from GLOBAL_FREEZE_MANIFEST.json; USC_SPEC.md is absent from GLOBAL_FREEZE_MANIFEST.json; CALIBRATION_PANEL_SPEC.md is absent from GLOBAL_FREEZE_MANIFEST.json; CALIBRATION_PANEL_REPORT.md is absent from GLOBAL_FREEZE_MANIFEST.json; VM_REFERENCE_V2_MANIFEST.json is absent from GLOBAL_FREEZE_MANIFEST.json; VM_REFERENCE_RECONCILIATION.md is absent from GLOBAL_FREEZE_MANIFEST.json |
| `representation_independence` | PASS | MUSIC-R1/R2/R3 are three representations of one candidate C06, never treated as independent candidate corpora |
| `statistical_protocol_applicable_at_n3` | PASS | every frozen-mandatory procedure applies at N=3; the only inapplicable procedure (within-class distribution) is frozen-conditional on >=3 independent corpora per class, never mandatory |
| `clean_git_revision` | PASS | git status --porcelain is empty; commit 743589101f3edaad743a57cc7ef5eb598ea0df5f |
| `go_test_including_A1_A10` | PASS | passed |
| `go_vet` | PASS | passed |
| `run_manifest_frozen` | PASS | PRODUCTION_RUN_MANIFEST.json is explicit for candidates=C01,C02,C06 with no runtime-default parameters |
| `deterministic_technical_prerun` | PASS | two independent technical pre-runs over C01,C02,C06 (loading, USC validation, structural traversal, seed schedule) produced byte-identical serialized output |

## Statistical protocol applicability at N=3 (task section 6)

| Procedure | Frozen source | Depends on panel N | Applicable at N=3 | Status |
|---|---|---|---|---|
| rarefaction (G/T/S/L/D per corpus/representation) | RAREFACTION_PROTOCOL.md section 1-4 | false | true | APPLICABLE |
| block bootstrap (per corpus/representation) | BOOTSTRAP_PROTOCOL.md section 1-2 | false | true | APPLICABLE |
| calibration scale construction | CALIBRATION_PANEL_SPEC.md; GLOBAL_FREEZE_MANIFEST.json | false | true | APPLICABLE |
| candidate-vs-VM metric comparison (d_G,d_T,d_S,d_L,d_D) | COMPARISON_PROTOCOL.md section 1-7 | false | true | APPLICABLE |
| paired notation delta (C01 LATIN-EXPANDED vs C02 LATIN-DIPLOMATIC) | PAIRED_NOTATION_PROTOCOL.md | true | true | APPLICABLE |
| within-class pair distances and variance (CLASS_SUMMARY / WITHIN_CLASS_DISTANCES) | COMPARISON_PROTOCOL.md line 28; COMPARATIVE_EXPERIMENT_SPEC.md line 35 | true | false | NOT_APPLICABLE_FOR_CURRENT_PANEL |
| cross-class ranking / PCA / UMAP / nearest-neighbour | COMPARISON_PROTOCOL.md line 30-31; COMPARATIVE_EXPERIMENT_SPEC.md line 37 | false | false | OUT_OF_SCOPE_REPOSITORY_LOCKED |

## Representation independence (task section 7)

MUSIC-R1, MUSIC-R2, MUSIC-R3 are three frozen representations of the single candidate C06 (C06_CORPUS_DECISION.md, REPRESENTATION_REGISTRY.md); they are never treated as three independent candidate corpora in candidate_order, the run manifest, or any statistic that assumes cross-candidate independence.

```text
GLOBAL_COMPARISON_PROTOCOL_FROZEN=true
PRODUCTION_CORPUS_SUBSET_FROZEN=true
PRODUCTION_CORPUS_INCLUDED=C01,C02,C06
PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false
PRODUCTION_COMPARATIVE_RUN_COMPLETED=false
PRODUCTION_COMPARATIVE_RUN_VALID=false
```

## Blockers

- global_freeze (global freeze contract is incomplete or inconsistent: GLOBAL_FREEZE_REPORT.md is absent from GLOBAL_FREEZE_MANIFEST.json; USC_SPEC.md is absent from GLOBAL_FREEZE_MANIFEST.json; CALIBRATION_PANEL_SPEC.md is absent from GLOBAL_FREEZE_MANIFEST.json; CALIBRATION_PANEL_REPORT.md is absent from GLOBAL_FREEZE_MANIFEST.json; VM_REFERENCE_V2_MANIFEST.json is absent from GLOBAL_FREEZE_MANIFEST.json; VM_REFERENCE_RECONCILIATION.md is absent from GLOBAL_FREEZE_MANIFEST.json)
