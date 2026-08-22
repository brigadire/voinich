# Task63 report

Phase A/B found a small but reproducible adjacent form effect after length-pair
matching: the observed within-line near rate is approximately 0.05855 versus
approximately 0.05663 for non-adjacent same-line controls. The effect is not
the raw Task60 vocabulary opportunity alone, but it is much smaller than the
raw adjacent-vs-independent contrast. Global and within-line shuffle nulls are
reported separately. Exact copies and d=1 transitions remain separate.

Near similarity decays with separation in DISTANCE_BY_SEPARATION.tsv, with
the strongest rates at separation 1–3 and a noisier lower tail afterwards.
Operation, position, directionality, line-boundary, chain and family tables
are descriptive; rare glyph transitions are not interpreted as rules.

The frozen transition model is deliberately minimal and uses a TRAIN-derived
local form-transition probability with reset to the frozen Task62 generator.
The current artifact is classified **FORM_DEPENDENCE_ONLY / PARTIAL**: a
residual adjacency effect is present, but the complete out-of-sample G+S
preservation comparison remains conservative and does not establish a unique
transition mechanism. This does not imply language, morphology, cipher
structure, derivation, or decipherment.
