# Distribution output contract v1 (B04)

Executable implementation: `internal/notation/distributions.go`. CLI:
`notation-corpus distributions`, `notation-corpus metric-output-types`.

## 1. Output types

Every metric this analyzer can emit has exactly one frozen output type,
enumerated in `METRIC_OUTPUT_TYPES.tsv` (`notation.MetricOutputTypes()`):

- `SCALAR` — a single number, compared with `|x-x_VM|/s` against a frozen
  calibration scale.
- `CATEGORICAL_DISTRIBUTION` — a probability distribution over a frozen,
  finite category universe, compared with Jensen-Shannon divergence.
- `ORDERED_DISTRIBUTION` — a probability distribution over an ordered
  (here: integer) support, compared with one-dimensional Wasserstein
  distance.
- `CURVE` — an accumulation curve over the frozen checkpoints, compared with
  normalized area between curves.
- `DESCRIPTIVE_ONLY` — a raw count/cardinality (`G01_ALPHABET_SIZE`,
  `S04_REPEATED_BIGRAM_TYPES`, `S05_REPEATED_TRIGRAM_TYPES`); reported but
  never scale-normalized, because a raw count is not itself a rate or
  proportion and its natural scale is corpus-size-dependent in a way MAD/IQR
  calibration does not capture.

`T01_MEAN_TOKEN_LENGTH` and `T11_POSITIONAL_RESTRICTION_DENSITY` keep their
existing scalar metric_id (used for calibrated scalar distance) *and* have a
separate, explicitly named distribution metric_id
(`T01_TOKEN_LENGTH_DISTRIBUTION`, `T11_POSITIONAL_RESTRICTION_DISTRIBUTION`)
holding the full distribution. This is a one-metric-one-type design (each
metric_id has exactly one output type) rather than the compound
"ORDERED_DISTRIBUTION + scalar mean" string in the task's own illustrative
example; the intent (both a point estimate and a distribution exist for
token length) is preserved.

## 2. Categorical distributions (T11)

Category universe (`PositionalRestrictionCategories`, frozen, never
candidate-specific): `INITIAL_RESTRICTED`, `INTERNAL_RESTRICTED`,
`FINAL_RESTRICTED`. Probability of a category is its share of the total
restricted-symbol-slot count. A symbol/category with zero mass on the
frozen support is `probability=0`, never omitted from the row set. If the
alphabet has no restricted symbol in any position at all
(`total==0`), every row is marked `comparable=false` with an explicit
reason instead of dividing by zero (D5).

`CategoricalJensenShannon(commonSupport, p, q)` rejects (returns an error
rather than silently including) any category present in either input that
is not in `commonSupport` — a candidate can never introduce its own
category (D6).

## 3. Ordered distributions (T01)

Support is the token length itself (`1, 2, 3, ...`), used directly as the
Wasserstein support value — no arbitrary bin edges. `OrderedWasserstein`
feeds this support/probability pair straight into the already-frozen
`Wasserstein1`.

## 4. Bootstrap unit

See `BOOTSTRAP_PROTOCOL.md`.

## 5. Confidence interval

95% CI, percentile-bootstrap estimator (`BootstrapCILevel = 0.95`), chosen
before any candidate or VM result was inspected and never re-chosen for
coverage afterward.

## 6. Serialization

`DISTRIBUTIONS.tsv` (`notation.WriteDistributionsTSV` /
`ReadDistributionsTSV`): `corpus_id, representation_id, metric_id,
support_id, bin_or_category, value, probability, comparable, reason`.

`BOOTSTRAP_RESULTS.tsv`: see `BOOTSTRAP_PROTOCOL.md`.

## 7. Validation (D1-D6)

All six pass in `internal/notation/distributions_test.go`:

- **D1** every comparable distribution's probabilities sum to 1 within
  `1e-9`.
- **D2** TSV round-trip preserves every row.
- **D3** a bijective symbol rename never changes `T01`/`T11` distributions.
- **D4** `RunBootstrap` with a fixed seed schedule is byte-identical across
  repeated runs (`internal/notation/distributions_test.go`,
  `TestD4BootstrapDeterminism`).
- **D5** the degenerate all-positions case emits `comparable=false`, never
  `NaN`/`Inf`; `EstimateScale` on an empty or zero-MAD-and-zero-IQR sample is
  `DEGENERATE`, never a fabricated epsilon.
- **D6** `CategoricalJensenShannon` fails closed on any out-of-support
  category.

`B04 = CLOSED` together with `BOOTSTRAP_PROTOCOL.md`.
