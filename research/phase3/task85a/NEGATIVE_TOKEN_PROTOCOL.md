# Deterministic negative-token protocol

## Population and matching

The positive multiset is every HELDOUT TOKEN occurrence. Produce exactly one
negative per positive, preserving order and duplicates by occurrence index.
The source pool is DEVELOPMENT TOKEN occurrences only. A candidate source must
have the positive's glyph length. For M5 it must also have the same number of
components under that candidate's frozen segmentation; this extra match is not
applied to M0–M4.

Glyph frequencies are DEVELOPMENT occurrence counts divided by total
DEVELOPMENT glyph occurrences. Sort glyphs by `(frequency, code point)` and
assign weighted cumulative-frequency quartiles: class 0 `[0,.25)`, class 1
`[.25,.50)`, class 2 `[.50,.75)`, class 3 `[.75,1]`; all occurrences of a tied
glyph remain together in the class containing the tie block's first cumulative
position. An unseen glyph has class 0. This is `GLYPH_FREQUENCY_CLASS`.

## Sampler

Derive a seed with corpus id `NEGATIVE`, the candidate id, HELDOUT partition,
scale `1`, and positive occurrence index as replicate index. Seeded-shuffle the
eligible source occurrences. For each source in that order, seeded-shuffle
positions, replace the first position whose class contains an alternative
DEVELOPMENT glyph, and try alternative glyphs in seeded order within the same
class. Thus length, per-position glyph-frequency class, and M5 component count
where applicable are matched while the DEVELOPMENT edit-neighborhood is
preserved.

Reject a proposal equal to any observed TOKEN type in DEVELOPMENT, VALIDATION,
or HELDOUT, equal to a prior negative, empty, containing a reserved symbol, or
failing the M5 component-count match. Negatives are unique. If one-position
mutation exhausts, enumerate two-position mutations in lexicographic position
pair and seeded within-class glyph order. If that exhausts, record
`NEGATIVE_EXHAUSTED`; PM6 is unavailable for that candidate/transcription and
PredictiveAdequacy fails. No uniform-noise fallback is allowed.
