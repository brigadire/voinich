# Structural fingerprint schema v2

Schema v2 supersedes the Task45 prototype v1. It is the first schema intended
for reproducible comparisons. A future change to the fixed feature set or
formula registry requires a new schema version.

Machine-readable outputs:

- `raw_metrics.tsv`: direct artifact values and support counts;
- `derived_metrics.tsv`: fixed formula, version, value and status;
- `normalized_metrics.tsv` / `structural_fingerprint.*`: fixed fingerprint
  features, values, and explicit missingness;
- `comparison_manifest.json`: hashes, verification, feature definitions,
  common-core dimensions and comparison identity.

Statuses are distinct: `VALUE`, `COMPLETED_EMPTY`, `NOT_APPLICABLE`,
`NOT_COMPUTED`, `MISSING_ARTIFACT`, and `FAILED`/`INVALID`. Numeric zero is a
real `VALUE` when the artifact reports zero.
