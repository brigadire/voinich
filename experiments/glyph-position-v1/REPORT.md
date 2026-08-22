# Task59 — glyph positional specialization

Observed corpus: `data_work/ZL3b-x7.canonical.txt`, SHA256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`, 39380 tokens, 45 glyph types.

The parser collapses the Task58 EVA composites longest-first and treats the resulting symbols as atomic. Singleton tokens are retained as a separate category; they are not folded into initial/final.

## 1. Observation

High-frequency (N >= 100), near-strict (share >= 0.95) specialists in Voynich: 6.

- `q`: N=5436, dominant=INITIAL, share=0.9912, Hnorm=0.0402, q=0.001405
- `N`: N=4317, dominant=FINAL, share=0.9856, Hnorm=0.0575, q=0.001405
- `e`: N=10654, dominant=MEDIAL, share=0.9838, Hnorm=0.0698, q=0.001405
- `E`: N=4888, dominant=MEDIAL, share=0.9765, Hnorm=0.0908, q=0.001405
- `i`: N=1607, dominant=MEDIAL, share=0.9676, Hnorm=0.1228, q=0.001405
- `m`: N=1071, dominant=FINAL, share=0.9505, Hnorm=0.1651, q=0.001405

This independently confirms the frequently-cited claim (task59 section 33): several high-frequency Voynich glyphs are strongly position-specialized, not only rare ones. The within-token shuffle null preserves token lengths, glyph multisets, boundaries and per-glyph frequency; only positional order is destroyed. See `GLYPH_POSITION_EXCLUSIONS.tsv` for glyphs excluded entirely from a position (e.g. observed=0 against an expectation in the hundreds/thousands) and `FREQUENCY_STRATIFICATION.tsv` for the full frequency-stratified breakdown, so this conclusion is not built on rare glyphs alone.

## 2. Mechanistic result: simple position-independent homophony

Voynich has 6 high-frequency near-strict specialists; unperturbed Doyle has 0; Longfellow has 1; Astafiev has 1 (`POSITIONAL_SPECIALIZATION_COMPARISON.tsv`).

- Doyle-H2-uniform (position-independent homophony, negative control): 0 high-frequency near-strict specialists.
- Doyle-H4-uniform (position-independent homophony, negative control): 0 high-frequency near-strict specialists.
- Doyle-H8-uniform (position-independent homophony, negative control): 0 high-frequency near-strict specialists.

Every position-independent homophonic control (H2, H4, H8) produces zero high-frequency near-strict specialists, despite each control's glyph vocabulary growing substantially (homophone splitting turns each plaintext glyph into H synthetic sub-types, most of them individually rare - this is exactly why the N>=100 floor matters, per section 20's rare-homophone safeguard). The position-dependent (`Doyle-H4-position-dependent`) and structured-token (`structured-positive`) positive controls both show maximal specialization (every synthetic symbol a strict specialist), confirming the analyzer does detect artificially created positional classes when they are actually present (section 21) - so the H2/H4/H8 controls' lack of high-frequency specialists is not a sensitivity failure of the tool.

**Classification (task59 section 29): INCOMPATIBLE_WITH_SIMPLE_HOMOPHONY.** Voynich's high-frequency positional specialization (6 glyphs) is not reproduced by applying simple position-independent homophony to a natural-language source (0 in every tested H) and exceeds what is observed in the natural-language controls themselves. Per section 29 this means only that *simple, position-independent* homophony is an insufficient mechanism for this specific property at these frequencies - it does not rule out position-dependent homophony, structured token encoding, natural-language morphology, or other cipher systems.

## 3. Relation to Task58 glyph-edge coupling

Task58 edge MI (`I(last(T_i); first(T_i+1))`, inter-token) is reported separately in `TASK58_EDGE_COMPARISON.tsv` because it measures a different property than the intra-token statistic here (section 24); the two are not averaged into one score.

## 4. Interpretation limits

Positional specialization in natural language is expected (morphology, orthography, final letter forms) and is not, by itself, evidence of a cipher. No claim is made here about language identity, decipherment, or a specific cipher mechanism; the classification above is scoped strictly to the simple position-independent homophony model tested.
