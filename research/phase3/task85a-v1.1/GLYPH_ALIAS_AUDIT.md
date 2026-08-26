# Glyph/rune alias audit

`NewGlyphAlias` sorts the complete observed alphabet plus reserved symbols,
deduplicates it, and assigns distinct lowercase letter-class Unicode runes.
Encoding is one glyph to one rune, so token order and boundaries are preserved;
natural mode lowercases without changing these aliases and never invokes EVA
composite collapse.

Independent reverse maps found no collision and recovered every glyph exactly:
45/45 for ZL3b and 32/32 for IT2a. On 512 frozen HELDOUT tokens per
transcription, a second disjoint bijection reproduced all seven F2 metrics
within `5e-16`. On 512 HELDOUT tokens restricted to already lowercase
single-rune glyphs, direct natural representation and alias representation
matched exactly for all seven metrics. The justified acceptance tolerance is
`1e-12`, far above observed floating accumulation noise and below frozen
diagnostic tolerances.

`R5_GLYPH_ALIAS = EQUIVALENT_IMPLEMENTATION`.

