# Exploratory positional-numeral report

Decision class: **LOCAL_NUMERIC_SIGNAL**

The experiment met the mechanical threshold for its decision class after the
same mapping optimization was applied to matched controls. This classification
is about surface regularity and is not evidence that token values are numbers.

## Required answers

1. Inventory: 25 admitted lowercase ASCII transcription symbols; 39380 raw tokens,
   8244 unique raw tokens, 175 excluded tokens. The literal transcription inventory
   was chosen as the simplest deterministic primary representation; it is not a
   claim about physical atomic glyphs.
2. Base: **B=25**.
3. Baseline regularity: score 0.323774 (sequential 0.202162, difference 0.679962,
   document 0.089199). It is descriptive, not a historical mapping.
4. Mapping search: best score 0.323774, an improvement of 0.000000. Best mapping is in
   NUMERIC_MAPPING_RESULTS.tsv.
5. Registered optimized-control comparisons significant after BH-FDR: SEQUENTIAL vs C1_WITHIN_TOKEN_GLYPH_SHUFFLE (q=0.04065, z=7.601); DOCUMENT vs C1_WITHIN_TOKEN_GLYPH_SHUFFLE (q=0.04065, z=11.211); AGGREGATE vs C1_WITHIN_TOKEN_GLYPH_SHUFFLE (q=0.04065, z=7.868); SEQUENTIAL vs C2_TOKEN_SHUFFLE_WITHIN_LINE (q=0.04065, z=19.511); DOCUMENT vs C2_TOKEN_SHUFFLE_WITHIN_LINE (q=0.04065, z=17.771); AGGREGATE vs C2_TOKEN_SHUFFLE_WITHIN_LINE (q=0.04065, z=24.088); SEQUENTIAL vs C3_GLYPH_BIGRAM_MARKOV (q=0.04065, z=17.699); DOCUMENT vs C3_GLYPH_BIGRAM_MARKOV (q=0.04065, z=17.355); AGGREGATE vs C3_GLYPH_BIGRAM_MARKOV (q=0.04065, z=13.489).
6. Metric families: see NUMERIC_PRIMARY_RESULTS.tsv; edit substitution
   consistency is 1.000000 and is counterevidence-neutral because it is forced by
   positional arithmetic, not independent support.
7. Document structure: folio/section/locus/line-position summaries are in
   NUMERIC_DOCUMENT_RESULTS.tsv; interpretation is 2D-LITE.
8. Independent IT2a: NOT_COMPARABLE; B=22, baseline 0.322396, optimized 0.322396.
9. Conclusion: **LOCAL_NUMERIC_SIGNAL**. This result does not prove the absence of numbers or of
   other numerical notations.
10. Strongest counterevidence: optimization did not improve the baseline;
    DIFFERENCE showed no excess against any control; EDIT consistency was 1 in
    VM and controls because it is algebraically imposed; and IT2a was not
    comparable under the literal-inventory rule. Further limitations: glyph
    identity is transcription-dependent; the optimizer is heuristic; p-value
    resolution is limited by 40 replicates per control; C3 preserves bigrams
    approximately rather than exactly; ordinary text also yields formal
    positional patterns (baseline 0.278354, optimized 0.287935); and no geometric coordinates or independent
    optimized-null distribution for IT2a are used.

Input SHA256: f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692; IVTFF SHA256: bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc; tokens analyzed: 39205; physical lines: 5385;
transcription: Zandbergen-Landini ZL3b canonical x7 aligned to IVTFF.

No hand-selected meaningful numbers were searched. There are no confirmatory
post-hoc observations.
