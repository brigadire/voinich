# Transcription alignment report — ZL3b-n.txt vs. IT2a-n.txt

Procedure: `TASK79C_DESIGN.md` §2 (exact folio-id / locus-id string
matching; positional token alignment within matched loci to the shared
prefix length; glyph-level identity comparison deferred because IT2a-n.txt
uses Basic Eva and ZL3b-n.txt uses extended Eva with high-ascii glyphs, so
a literal character comparison would conflate alphabet-encoding
differences with genuine transcription disagreement).

## Coverage

| Level | ZL (canonical) | IT (alternate) | Matched | Coverage of ZL | Coverage of IT |
|---|---:|---:|---:|---:|---:|
| Folios | 227 | 225 | 225 | 99.1% | 100.0% |
| Loci | 5,385 | 5,215 | 5,214 | 96.8% | 100.0% |
| Token positions (within matched loci, shared-prefix) | — | — | 36,537 | — | — |

Full per-level counts: `TRANSCRIPTION_ALIGNMENT.tsv`.

## Excluded folios

- Only in ZL, absent from IT entirely: `f116v` (the manuscript's marginal
  "Voynich/Michitonese" text on the last written page) and `fRos` (the
  large foldout "Rosettes" page). Both are manuscript regions with known
  transcription difficulty across the field; their absence from IT2a-n.txt
  is transcriber policy, not a Task79c exclusion.
- Only in IT, absent from ZL: none.

## Locus-level detail

- 171 loci present in ZL only (mostly sub-numbered fold-out loci, e.g.
  `f102r2.*`, `f102v1.*`, `f67v2.*`, plus every locus on the two
  ZL-only folios above).
- 1 locus present in IT only (`f89v2.26`).
- Of the 5,214 matched loci, 1,320 have a different token count between
  the two transcriptions (a segmentation/reading difference, e.g. one
  transcriber reading a run as one word where the other reads two). For
  these, only the shared-prefix token positions are compared
  positionally; the remaining 1,670 token positions (summed over both
  sides of the 1,320 mismatched loci) are recorded as unmatched, not
  interpolated or forced to agree, per §2's explicit prohibition on
  forcing incompatible conventions to coincide.

## What this does and does not establish

96.8%/100% locus coverage is high enough to recompute every CORE metric
that depends on IVTFF metadata (folio, section, locus type) without a
material change in sample size — the metric-level comparison in
`TRANSCRIPTION_STABILITY.tsv` confirms this (no CORE metric was
`NOT_TESTABLE`). It does **not** by itself establish glyph-level agreement:
two transcribers can agree on locus/word segmentation while disagreeing on
individual glyph readings (as the sampled `f1r.1`-`f1r.4` loci show, e.g.
`shor[cth:oto]res` in ZL vs. `cthres` in IT — a bracketed-uncertain-reading
notation difference, not a locus/token misalignment). Glyph-level
disagreement, where it exists, is absorbed into each transcription's own
independently computed metric value, which is exactly what
`TRANSCRIPTION_STABILITY.tsv`'s standardized-difference column measures.
