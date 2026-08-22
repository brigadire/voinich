# Task62 frozen design

Split: contiguous line blocks, 60% TRAIN / 20% VALIDATION / 20% TEST.
Representation: internal/evaglyph; token lengths are sampled empirically from TRAIN.
Models: IID, POSITION_IID, MARKOV_1, MARKOV_2, POSITION_MARKOV_1.
Smoothing: additive alpha=0.1 over the TRAIN glyph alphabet. No test data or
Task59/60/61 metric enters model selection. Selection rule: lowest validation
cross entropy, ties go to the simpler model in the listed order. Generation:
100 corpora, TEST token count, seeds 62000..62099. Metrics are validation only
after this design is frozen. Copy/mutate is a positive control with p=0.25,
fixed before generation; it is not an explanatory model.
