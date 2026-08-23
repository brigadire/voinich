# Contributing

Use GitHub issues for reproducible bug reports, including the command, Go
version, operating system, input checksum, seed, and the smallest relevant
output. Do not attach unpublished, private, or license-unclear corpora.

Methodological issues should identify the claim, experiment, artifact, and
assumption or control in question. New experiments must use a new manifest,
declared corpus provenance, recorded seeds, and a report that distinguishes
exploratory results from confirmatory results.

Pull requests should be narrow, tested, and preserve frozen artifacts unless
the change explicitly creates a superseding experiment. Ordinary refactoring
and scientific corrections are different changes: a scientific correction must
not silently rewrite an existing published artifact. Keep the old artifact,
publish a correction or superseding experiment with its reason and provenance,
and update the affected claim/result index.
