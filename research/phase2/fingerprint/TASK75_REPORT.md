# Task75 report — lexical-paradigm Fingerprint v2 block

Status: **implementation complete; no canonical corpus result generated in
this repository**.

## Delivered protocol

`internal/fingerprintv2` and `cmd/fingerprint-v2-analyze` implement the
first deterministic Fingerprint v2 block. The CLI loads a primary generic
corpus and named generic controls with natural line boundaries. It can
strictly align a supplied frozen token stream with IVTFF metadata using the
existing metadata validator; alignment mismatch is an error. No Phase I
input or frozen artifact is modified.

LP1 enumerates global vocabulary edit-distance-one edges using Task60's
candidate-indexed edit graph and operation classifier. It counts directed,
type-level transformation rules and reports support concentration. LP2 tests
the rule-support Gini against both C-GRAMMAR modes and explicit C-LEN/C-FREQ
random re-pairing analogues, with a declared Benjamini-Hochberg family. LP3
uses only thresholded rules after a significant grammar-null LP2 result and
reports family component branching, exact depth, overlap, and available
line/page locality. LP4 uses a declared one-glyph prefix/suffix decomposition
and type-level normalized core/affix MI with both a within-length
affix-permutation null and validated C-GRAMMAR comparisons.

EF1 reports graph degree/components including isolates. EF2 reports
transitivity, triangles, 3-paths and 4-cycles, calibrated by degree-preserving
double-edge swaps. EF3 reports Spearman degree/log-frequency with a C-FREQ
label-permutation control. EF4 gives a grammar-bounded consolidated verdict.
Raw output includes all generated grammar diagnostics and null distributions.

## C-GRAMMAR

Both modes preserve token count and length distribution exactly, constrain
draws by exact length/position glyph profiles, initial/final positions and
first-order local transitions, and retain the observed alphabet. The
frequency-aware mode generates forms, de-duplicates them, then independently
assigns observed frequency ranks within length class before duplication. It
does not copy observed token forms or edit families deliberately.

Generator adequacy is not assumed. Every replicate records position, endpoint
and bigram total variation along with vocabulary, singleton/rare and
frequency-distribution diagnostics. A diagnostic-tolerance breach is surfaced
in `C_GRAMMAR_VALIDATION` and warnings; that grammar mode is retained for
audit but excluded from inferential tests. If no mode validates, every
grammar-dependent conclusion is `INCONCLUSIVE`.

## Formal verdict rules

- `C_GRAMMAR_VALIDATION` is `SUPPORTED` only if all exact marginals hold and
  every replicate's profile diagnostic is within configured tolerance.
- `DIRECTIONAL_TRANSFORMATIONS_SUPPORTED` and
  `PARADIGM_PRODUCTIVITY_SUPPORTED` require an FDR-significant LP2
  C-GRAMMAR comparison plus at least one rule above support threshold.
- `EDIT_NEIGHBORHOODS_EXCEED_GRAMMAR_NULL` requires both that directional
  result and EF4 `EXCEEDS_GRAMMAR_BOUND`; grammar-consistent EF4 is
  `NOT_SUPPORTED`.
- `CONTEXT_CONDITIONING_SUPPORTED` is at most `PARTIALLY_SUPPORTED` in this
  first block, from LP3 same-line locality. It is not a claim about all
  metadata covariates.
- `LEXICAL_PARADIGM_BLOCK_READY` is `PARTIALLY_SUPPORTED` when the pipeline
  and diagnostics complete; it becomes `INCONCLUSIVE` if C-GRAMMAR validation
  fails. This readiness label is infrastructural, not a content conclusion.

Valid values remain `SUPPORTED`, `PARTIALLY_SUPPORTED`, `NOT_SUPPORTED`, and
`INCONCLUSIVE`.

## Available and unavailable results

Unit tests use synthetic positive and negative productivity fixtures, C-GRAMMAR
determinism/preservation checks, and an end-to-end CLI fixture with strict
IVTFF alignment. They do **not** use or invent a canonical Voynich corpus
result. Corpus bytes are deliberately unavailable under the repository data
discipline, so a real canonical run requires a locally acquired, checksummed
source and documented configuration. Named controls are direct configured
inputs, not downloaded by this tool.

Current limitations are deliberate: no held-out folio productivity test, no
Currier/hand/fold stability battery, no C-PAGE locality null, no distance-two
rule census, and no transcription-sensitivity comparison. Graph components
are operational edit families, never asserted morphemes or paradigms by
themselves.

## Task77 dependencies

Task77 should consume the schema/artifacts defined in
[FINGERPRINT_V2_SCHEMA.md](FINGERPRINT_V2_SCHEMA.md), add folio/metadata
folds and transcription sensitivity when data become available, run canonical
and documented natural/synthetic controls, assess C-GRAMMAR sensitivity, and
only then consider freezing a broader v2 fingerprint. It should not treat
Task75 fixture output as research evidence.
