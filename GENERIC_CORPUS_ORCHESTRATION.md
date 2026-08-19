# Generic corpus orchestration

This repo now distinguishes preparation from validation.

`codex_prepare` is the explicit preparation tool. It converts a raw text
document into a canonical generic corpus and writes a sibling prepare
manifest.

`pipeline-orchestrate` adds Stage 0 for generic runs only:

- `corpus-readiness-check`
- it validates the corpus in place;
- it never rewrites the corpus;
- it never calls `codex_prepare` as a transform;
- it stops the pipeline before Stage 1 if the corpus is not canonical.

Stage 0 uses the shared corpus-preparation validation logic. If a sibling
`*.prepare.json` exists, the pipeline records and rechecks its provenance.

The important operational rule is simple:

- prepare before you analyze;
- validate before you run scientific stages;
- never silently normalize inside the pipeline.

Cross-corpus comparisons still need a single, consistent line policy.
`preserve` keeps logical lines; `reflow` flattens the corpus into a token
stream. Pick one policy explicitly and keep it fixed for a comparison set.
