# Reproducible experiment comparison

Production mode requires every input directory to contain `FROZEN`, a valid
`checksums.sha256`, and a successful checksum verification. `-allow-unfrozen`
is an explicit development override; its report is marked
`NON_REPRODUCIBLE_DEVELOPMENT_COMPARISON` and must not be treated as a frozen
scientific snapshot.

Task52 emits schema version 2. The comparison manifest records experiment IDs,
manifest hashes, individual artifact hashes, frozen/verification status,
corpus hashes, comparison program commit, feature definitions, normalization
scope, common-core feature names, and a deterministic comparison identity.

Distances include pairwise-available metrics (with coverage), common-core
metrics (identical dimensions for all experiments), and semantic-family
distances. Cohort standardization is deterministic and explicitly exploratory
for small N. It uses equal feature weights and no post-hoc selection.

Task49 vocabulary-growth artifacts are optional. Legacy experiments without
them remain valid and receive `MISSING_ARTIFACT`; checkpoints beyond a corpus
length are `NOT_APPLICABLE`. Pairwise Task49 rows use the largest shared
defined checkpoint and never truncate an input corpus.

The output disclaimer is mandatory: similarity is not classification, causal
explanation, or proof that the nearest corpus has the same document type.
