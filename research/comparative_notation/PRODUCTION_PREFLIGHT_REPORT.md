# Production corpus subset preflight report

Scope: `PRODUCTION_CORPUS_SELECTION.json`, not comparative-run authorization.

| Gate | Status | Detail |
|---|---|---|
| Global protocol state | PASS | `GLOBAL_COMPARISON_PROTOCOL_FROZEN=true`. |
| Selection identity | PASS | Version, base Git revision, timestamp, and nine candidate-manifest hashes are explicit. |
| Included readiness | PASS | C01, C02, and C06 raw inputs, policies, profiles, provenance, USC, validation, and reproducibility are hash-bound. |
| Excluded/deferred decisions | PASS | C03/C04/C05/C07/C08 exclusions and C09 deferment are reasoned and hash-bound before analysis. |
| Panel diversity | PASS | BDD Latin and JLSDD polyphony are distinct notation/source families. |
| Frozen statistical applicability | PASS | All inputs support 5k/10k checkpoints; larger checkpoints are preregistered `NOT_COMPARABLE`; no unsupported within-class distribution is claimed. |
| Negative authorization | PASS | `PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false`; no comparative computation was run. |

Machine-readable validation is in `production_corpus/validation.json`. The
separate comparative-run preflight and authorization remain intentionally not
executed by this task.

```text
FULL_CANDIDATE_PANEL_READY=false
PRODUCTION_CORPUS_SUBSET_FROZEN=true
PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false
```
