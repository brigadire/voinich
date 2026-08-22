# Task62 report

Design was frozen before generation (`DESIGN_FROZEN`). The contiguous line-block split is 60% TRAIN / 20% VALIDATION / 20% TEST; no random token-level split is used. Model selection used validation cross entropy only.

`POSITION_MARKOV_1` was selected. Its held-out test cross entropy is 2.360985 bits/glyph (perplexity 5.137); test held-out likelihood is primary. The generated model produces novel types and a substantial d=1 component, but does not reproduce all observed h2/positional targets and does not reproduce adjacent near-repeat enrichment (generated mean about 0.0234 versus TEST 0.0395).

Classification: **LOCAL_FORMATION: PARTIAL**; **SEQUENCE: SEPARATE_SEQUENCE_RULE_REQUIRED**. Length/frequency controls, positional transitions, novel types, copy/mutate, structured-token controls, and natural-language controls are reported separately. These results do not imply language identity, morphology, cipher reconstruction, or decipherment.
