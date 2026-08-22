# Task64 report: line-level form regimes and local mixture structure

Classification: **BROADER_LOCAL_REGIME**

This report answers section 64's fourteen questions from the artifacts in this
directory. R is called a *regime* or *line-associated structure* throughout,
never a topic, sentence, semantic state, cipher key or grammatical context
(section 65); a PHYSICAL_LINE_REGIME finding does not imply the physical line
was a unit of the underlying plaintext, since scribal line-filling behaviour is
an equally available explanation (section 66). No claim is made here about
language, semantics, a cipher key or decipherment.

## 1-4. Are same-line tokens more similar, does it survive non-adjacency, is it stronger than the same-page control, and does it survive length matching?

Same-line non-adjacent DLE1 rate = 0.057824; different-line/same-page matched control = 0.049905. Delta_line = +0.007939 (bootstrap 95% CI [0.004600, 0.012813], 500 line resamples). The comparison already excludes adjacent pairs and is matched on both token lengths, so a positive Delta_line answers questions 1-4 together: yes to same-line similarity, yes it survives excluding adjacency, and the same-page matched control is the direct answer to
question 3's 'stronger than' contrast. See LINE_PAIR_SIMILARITY.tsv for the full exact/d=1/d<=1/mean/normalized breakdown across ADJACENT_SAME_LINE, NONADJACENT_SAME_LINE, SEP2/SEP3+_SAME_LINE, DIFFERENT_LINE_SAME_PAGE_MATCHED and DIFFERENT_PAGE_MATCHED.

## 5-6. Within-line shuffle and line-membership destruction

LINE_NULLS.tsv reports the same-line-non-adjacent rate, adjacent rate, matched control and Delta for REAL_LINES against six nulls. Because every within-line pair (all C(n,2) combinations) is already counted somewhere in the ADJACENT+NONADJACENT pool, an order-only within-line shuffle only relabels which n-1 pairs count as 'adjacent' each draw; the pooled non-adjacent rate should stay close to its unshuffled value (it is not pinned bit-for-bit, since the excluded adjacent subset changes), while it is genuinely diagnostic for the separation-specific (adjacency) rate. Line-membership shuffle and within-page line-membership shuffle both preserve line lengths and (respectively) the global/per-page token multiset while destroying actual line assignment: a Delta that collapses toward zero under these two nulls is the direct evidence that real line membership - not just line length or page vocabulary - carries the signal.

## 7. Does the physical line differ from shifted/fixed local blocks?

REGIME_SCALE_COMPARISON.tsv compares ADJACENCY, LINE, three SHIFTED_LINE offsets, three FIXED_WINDOW sizes and PAGE against one shared null (the global length-matched different-page rate), each with a page-level bootstrap CI/Z and separate discovery/replication point estimates. The classification above (**BROADER_LOCAL_REGIME**) is read directly off whether the LINE row's effect size dominates every shifted/window row.

## 8. Is there a separate page-level effect?

Yes in the sense that LINE_PAIR_SIMILARITY.tsv's DIFFERENT_LINE_SAME_PAGE_MATCHED row and DIFFERENT_PAGE_MATCHED row are reported separately and the PAGE row in REGIME_SCALE_COMPARISON.tsv gives the page scale its own effect size against the same null; comparing the two matched-control rows gives the same_line > same_page > global hierarchy from section 19 directly.

## 9. How long does the regime persist across neighboring lines?

REGIME_PERSISTENCE.tsv reports mean line-profile distance for k=1..10 lines apart (same page only) plus a PAGE_BOUNDARY row (last line of page P to first line of P+1); a persistence effect that is not confined to k=1, or a PAGE_BOUNDARY row that stands out from ordinary k=1 transitions, would indicate the regime is not unique to a single line, or that page boundaries partially reset it, respectively.

## 10. Is there first/last token specialization?

LINE_POSITION_EFFECTS.tsv reports FIRST/SECOND/INTERIOR/PENULTIMATE/LAST token length, top-initial/final-glyph share and giant-fraction; LINE_START_END.tsv repeats the first/interior/last contrast restricted to the single modal token length so the comparison is not confounded by length. Per section 30, this is not interpreted as sentence capitalization or syntactic marking.

## 11. Does an adjacency residual survive conditioning on line?

Test	AdjacentRate	MatchedNonAdjacentRate	Gap
TASK63_PUBLISHED	0.058548216	0.056634876	0.001913340
TASK64_REPLICATION_LENGTH_ONLY	0.059144733	0.054914713	0.004230019
TASK64_LENGTH_AND_REGIME_MATCHED	0.059144733	0.053915552	0.005229181

TASK64_REPLICATION_LENGTH_ONLY replicates Task63's own length-matched adjacent-vs-non-adjacent contrast on this corpus/parser; TASK64_LENGTH_AND_REGIME_MATCHED additionally requires the non-adjacent control to come from a line in the same regime bucket. Verdict: the residual gap did NOT shrink materially once the non-adjacent control was additionally matched to the adjacent pair's regime bucket (it went from the length-only gap to a slightly LARGER regime-matched gap) - line regime does not explain Task63's adjacency residual on this run.

## 12. Which scale best explains local form similarity?

- ADJACENCY: effect=0.018020 (null=0.041049, Z=5.987, discovery=0.025771, replication=0.010698)
- LINE: effect=0.016459 (null=0.041049, Z=1.594, discovery=0.030069, replication=0.003132)
- SHIFTED_LINE_OFFSET1: effect=0.019273 (null=0.041049, Z=1.561, discovery=0.029990, replication=0.003182)
- SHIFTED_LINE_OFFSET2: effect=0.017574 (null=0.041049, Z=1.316, discovery=0.029987, replication=0.002616)
- SHIFTED_LINE_OFFSET3: effect=0.014605 (null=0.041049, Z=1.466, discovery=0.028749, replication=0.002190)
- FIXED_WINDOW_5: effect=0.010467 (null=0.041049, Z=5.415, discovery=0.016758, replication=0.003308)
- FIXED_WINDOW_10: effect=0.008628 (null=0.041049, Z=3.765, discovery=0.015369, replication=0.000217)
- FIXED_WINDOW_20: effect=0.006240 (null=0.041049, Z=2.723, discovery=0.013270, replication=-0.001012)
- PAGE: effect=0.002915 (null=0.041049, Z=2.642, discovery=0.004721, replication=-0.008925)

Verdict: the largest non-adjacency effect size is **SHIFTED_LINE_OFFSET1** (0.019273); LINE itself is 0.016459 and ADJACENCY is 0.018020. Classification: **BROADER_LOCAL_REGIME** - the clustering is real but its best-supported scale is local-and-line-sized, not the physical line boundary specifically (at least one SHIFTED_LINE or FIXED_WINDOW row matches or exceeds LINE's effect size).

## 13. Can G+R reproduce the observed structure?

REGIME_GENERATIVE_VALIDATION.tsv compares G_ONLY, the frozen G_PLUS_R_FULL model and its ablations, the Task62 G-only corpus chunked into real Voynich line-length blocks, a copy/mutate positive control chunked the same way, and the real held-out TEST-fold Voynich statistics (VOYNICH_TEST_OBSERVED), all generated to TEST's own line-count/tokens-per-line structure (never its token identities) and averaged over 25 replicates per model. PRESERVATION.tsv juxtaposes the frozen G_PLUS_R_FULL corpus's giant-component fraction, order-2 glyph entropy and positional weighted entropy against the authoritative Task59-61 reference values (read live from those experiments' own artifacts), plus an adjacent-exact-repeat proxy for Task58's token-order dependence. ABLATION.tsv isolates which of R/length/initial-glyph/final-glyph drives whatever same-line effect G+R produces. Caveat: the frozen model draws middle glyphs from one global (non-regime) distribution rather than Task62's within-token POSITION_MARKOV_1 mechanism, so PRESERVATION.tsv's glyph-entropy and positional-entropy rows are expected to run higher than the authoritative reference - that gap reflects the deliberately minimal token-internal mechanism (section 44 caps it at length/initial/final categoricals), not a failure to reproduce line-level clustering, which is what GiantComponentFraction and the generative Delta rows target.

## 14. Is a separate Task63 S-process still needed?

Yes: the length-and-regime-matched gap is not materially smaller than the length-only gap (see question 11), so line/local regime does not substitute for Task63's weak sequential residual - a separate S-process (or none, given how small both gaps are) remains the open question exactly as Task63 itself left it (FORM_DEPENDENCE_ONLY / PARTIAL).

## Scope

Stages 1-28 were not touched; no Stage29 was added. Every distance/entropy computation reuses internal/tokentransition, internal/characterentropy and internal/evaglyph rather than redefining glyph or token distance. This report makes no claim about language, morphology, a specific cipher, or decipherment.
