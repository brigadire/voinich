# Task64 frozen minimal regime model (REGIME_MODEL)

G+R is deliberately minimal (task64 section 44): R is a categorical
regime index (K<=4) fit only from TRAIN eligible lines, bucketed by
quantiles of each line's mean token length (a structural, pre-registered
criterion, not chosen from any Voynich fingerprint). Each regime carries
three independent categorical distributions - P(length|R), P(initial
glyph|R), P(final glyph|R) - estimated as raw TRAIN counts; middle glyphs
are drawn from one global (non-regime) distribution, matching the
"small fixed set of transition parameters" ceiling in section 44.

Generation samples R from the TRAIN empirical regime-proportion
distribution (never from TEST content), then draws every token in the
line independently given R - tokens are conditionally independent given
R by construction, which is the operational target of section 42's
T_i ⟂ T_j | R_line test. Line-length structure (number of lines, tokens
per line) is taken as given external structure, never generated content
(section 47).

Ablations (ABLATION.tsv) remove or randomize each component: REMOVE_R
(=G_ONLY baseline), RANDOMIZE_R (regime label uniform instead of TRAIN
proportions), PRESERVE_R_SHUFFLE_WITHIN_LINE (order destroyed after
generation - expected near no-op, since generation is already
order-exchangeable given R), and REMOVE_LENGTH/INITIAL/FINAL_COMPONENT
(that one distribution reverts to the global, non-regime version).
