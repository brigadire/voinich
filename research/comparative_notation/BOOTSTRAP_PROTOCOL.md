# Bootstrap protocol v1 (B04)

Executable implementation: `internal/notation/bootstrap.go`
(`RunBootstrap`). CLI: `notation-corpus bootstrap`.

## 1. Bootstrap unit

A block bootstrap over the same `structuralBlock` partition used by
rarefaction (`internal/notation/rarefaction.go`): `bootstrapDraw` resamples
whole blocks **with replacement**, drawing exactly as many blocks as the
source has. This is the frozen primary policy of section 18 collapsed to
one implementation, because (as documented in `RAREFACTION_PROTOCOL.md`)
the deepest source-observed hierarchy unit already equals the physical line
whenever lines exist, and equals the whole document when they do not — the
same unit G/T/S/D rarefaction uses. L-specific metrics inherit
`NOT_COMPARABLE` from `Analyze()` itself when physical lines are not
observed, exactly as elsewhere; bootstrap does not need a second unit
definition because it never overrides a metric's own comparability.

Each drawn occurrence of a block gets a disambiguated hierarchy key
(`disambiguateDeepestLevel`, appending `-boot<i>` to the deepest observed
level) so that drawing the same block twice in one bootstrap sample never
produces a duplicate `token_id` or a non-contiguous `token_index` — the
source order and content inside the block are otherwise untouched.

Bootstrap estimates the sampling uncertainty of the corpus's own point
estimate; unlike rarefaction it is evaluated once, at the corpus's actual
observed size, not once per checkpoint (`BOOTSTRAP_RESULTS.tsv` has no
checkpoint column).

## 2. Replicates: B=200 (reduced from the proposed 1,000, benchmarked)

A runtime benchmark on the frozen VM source (39,380 tokens, `internal/notation`
package, single `go run` timing) measured one full `Analyze()` pass at
≈4.85s. At `B=1,000` this alone is ≈81 CPU-minutes for one
corpus/representation, before accounting for smaller candidates or repeated
runs; this was judged impractical for a single preparation task (task
section 19 explicit STOP-and-document escape hatch).

`BootstrapReplicates = 200` is the frozen production value instead: ≈16
CPU-minutes for the VM's full-size bootstrap, and 200 replicates is a
standard lower bound quoted for a valid 95% percentile bootstrap CI (Efron
& Tibshirani). This value is applied uniformly to VM and to every future
candidate corpus — it is not reduced per candidate.

## 3. Confidence interval

95% CI (`BootstrapCILevel = 0.95`), percentile estimator: `ci_low`/`ci_high`
are the 2.5th/97.5th percentile of the `B` bootstrap draws (linear
interpolation, `percentile()`), chosen before any VM or candidate value was
computed.

## 4. Output schema — `BOOTSTRAP_RESULTS.tsv`

One row per metric (`notation.WriteBootstrapTSV`):

| column | meaning |
|---|---|
| `corpus_id`, `representation_id`, `family`, `metric_id`, `regime` | identity |
| `estimate` | the metric's own point value at the corpus's full observed size |
| `bootstrap_mean`, `bootstrap_sd` | across the `B` resampled draws |
| `ci_level` | `0.95` |
| `ci_low`, `ci_high` | percentile-bootstrap 95% CI |
| `n_valid` | number of bootstrap draws in which the metric was comparable |

## 5. Validation

- **D4 determinism**: `TestD4BootstrapDeterminism` runs `RunBootstrap` twice
  with the identical seed schedule and requires every row to compare equal.
  This required fixing a genuine pre-existing non-determinism defect in
  `ShuffleTokenOrder` (it consumed a shared `math/rand` source in raw Go map
  iteration order, which Go randomizes per process) — the fix sorts group
  keys before shuffling each group; see the commit history for
  `internal/notation/transform.go`. It also required making every
  map-iteration-order-dependent float64 accumulation in `analyze.go`
  (conditional entropy, edit-graph clustering, group coherence/variance
  share) sum over sorted keys instead, for the same reason.
- Bootstrap draws are validated (`Validate`) before analysis, so a
  malformed resample fails loudly rather than silently corrupting a metric.

`B04 = CLOSED` together with `DISTRIBUTION_OUTPUT_CONTRACT.md`.
