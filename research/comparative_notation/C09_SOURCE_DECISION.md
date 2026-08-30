# C09 source decision — historical formal/tabular notation

Decision date: 2026-08-30. First-run decision: `DEFERRED`;
`C09_PRODUCTION_READY=false`.

## Sources considered

| Source | Evidence | Decision |
|---|---|---|
| OPenn LJS 361, *Astronomical and astrological tables* | Stable manuscript images; Public Domain Mark; metadata CC BY 4.0 | Scientifically suitable physical source, but reject for this freeze: no scholarly machine-readable cell transcription or frozen cell/row extraction exists. |
| LOCOMAT Digital Tables Library entries | Stable discovery catalogue with medieval/early-modern tables | Reject as production input: catalogue links are not a versioned diplomatic cell corpus. |
| Synthetic table fixture | Deterministic and schema-valid | Reject: cannot replace a historical layout-bearing source. |

LJS 361 is the preferred acquisition target, but image availability alone
does not justify inventing a transcription. Manual transcription and review
must be a separate, preregistered source-production task.

## Derived transcription feasibility

A limited transcription is technically possible because the images and table
coordinates are stable and rights-cleared. It is not feasible within this
freeze: cell inclusion rules, headers/body roles, damaged-cell uncertainty,
and double-keyed independent validation must be frozen before transcription,
and the resulting corpus must reach at least the first 5,000-token rarefaction
checkpoint to support any size-matched result. No such independently validated
cell set exists. C09 is therefore deferred, not filled with OCR guesses or a
synthetic table.

Sources checked: `https://openn.library.upenn.edu/Data/0001/html/ljs361.html`;
`https://locomat.loria.fr/locomat/digital-tables.html` (accessed 2026-08-30).
