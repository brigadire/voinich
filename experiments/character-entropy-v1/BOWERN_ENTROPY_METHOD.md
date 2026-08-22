# Bowern–Lindemann entropy method

The published estimand is Shannon conditional character entropy, also called second-order character entropy: `h2 = H(X_i | X_{i-1})`, measured in bits. It is the empirical plug-in estimator of `-sum p(x,y) log2 p(y|x)`, not entropy of pairs. Whitespace and punctuation are excluded in the literature representation; line breaks are not glyphs.

Task61 additionally reports a separate shared-EVA representation using `internal/evaglyph.CollapseEVA`, and natural controls use lowercase Unicode letters/digits. Continuous, token-boundary (`<WB>`), within-token-only, and line-reset modes are explicit. Primary values have no correction; sample/context counts and coverage expose sparse higher orders. Normalized entropy is secondary: h_k/log2(|G|).

Primary sources: Bowern & Lindemann (2021), *The Linguistics of the Voynich Manuscript*, Annual Review of Linguistics 7:285–308; Lindemann & Bowern (2020), *Character Entropy in Modern and Historical Texts: Comparison Metrics for an Undeciphered Manuscript*, arXiv:2010.14697. The latter is the directly relevant entropy comparison and documents transcription, script size, composites, and positional constraints. Bennett (1976) is historical background cited by those works.

Currier/section/hand/label and Task59/60 dependent rows are marked unavailable when the required metadata is absent; no heuristic classification or Task52 optimization is used.
