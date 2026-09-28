# Transcription layout conventions (frozen source audit)

Status: `INVESTIGATED_BEFORE_ALIGNMENT`.

The primary local sources are `data/ZL3b-n.txt`,
`research/astro_hapax_star_label/TRANSCRIPTION_POLICY.md`,
`TRANSCRIPTION_LINE_CANDIDATES.tsv`, `TRANSCRIPTION_TOKEN_CANDIDATES.tsv`,
and the upstream occurrence metadata registry. The source uses IVTFF locus
references such as `<f68r1.9,@Ls>`; the numeric suffix identifies a source locus
within the folio, not a proven spatial rank. `P` is prose/page-text material,
`L` is a one-label line in the target segment, `C` is a circular or cyclic
sequence, and `R` is radial text. `Pb` is a page-text sub-locus. This role
interpretation is documented by the literal locus tags in the IVTFF source and
the prepared line registry; it is not inferred from token frequency.

Occurrences within a line follow the frozen tokenization and the registry's
`index_in_line`; lines follow source order in the IVTFF file. Punctuation,
comments, alternatives, uncertainty markers and numeric literals are retained
according to `TRANSCRIPTION_POLICY.md`; no new glyph normalization is allowed.
The fixed display expansion is C→cth, K→ckh, P→cph, F→cfh, N→iin, A→ain,
H→ch, S→sh, E→ee, I→in. This is `DOCUMENTED`.

The target pages contain f68r1 `@Ls` singleton label loci 8–29 and 31–36,
one multi-token `@Ls` locus 30, and a `@Cc` cyclic locus 37. f68r2 contains
`@Ls` and `@Cc` loci; f68r3 contains `@R`, `@Ls`, and `@Cc` loci. The exact
counts and token IDs are registered in `TRANSCRIPTION_SEQUENCE_REGISTRY.tsv`.
The existence of these locus types is `DOCUMENTED`; their correspondence to
physical spatial traversal is `UNKNOWN` until tested below.

No source document establishes clockwise, counterclockwise, radial, or
inner/outer traversal for these pages. Therefore every such order is marked
`EXPLORATORY`, while page scan orders are also exploratory hypotheses. Line
number is never treated as spatial order without anchor validation. Ring starts
and direction are retained as alternatives, not chosen by semantics.

