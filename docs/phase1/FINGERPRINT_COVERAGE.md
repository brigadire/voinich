# Fingerprint coverage audit

Coverage grades describe how much of a direction Phase I measured with explicit controls and reproducible artifacts. They do not rate how “Voynich-like” a model is.

| direction | grade | evidence | principal gap |
|---|---|---|---|
| marginal distributions | WELL_COVERED | Stage2/3/28 inventories, lengths, ranks, vocabulary growth; natural and transformed controls | transcription sensitivity and cross-corpus unit alignment |
| joint distributions | PARTIALLY_COVERED | adjacent token MI, edge MI, glyph bigrams, transition networks | higher-dimensional joint estimators and common representations |
| cross-scale dependencies | PARTIALLY_COVERED | token/glyph, adjacency/line/page/manuscript comparisons; Tasks58–65 | one integrated hierarchical model and uncertainty propagation |
| glyph structure | WELL_COVERED | positional specialization, edge coupling, conditional entropy, shuffles | visual allography, paleography and image-grounded glyph uncertainty |
| token formation | WELL_COVERED | edit networks, position Markov models, held-out Task62, Task66 ablations | broader grammar classes and independent transcription replication |
| lexical paradigms | WEAKLY_COVERED | edit-distance families and operation positions | no demonstrated paradigm semantics, direction or ancestry |
| sequence structure | PARTIALLY_COVERED | MI, frozen sequence replication, Task63 matching, higher-order audits | sparse n≥3 support; long-range predictive tests |
| repetition | WELL_COVERED | exact/near runs, line/global/matched nulls, homophony dose response | copying direction and image/scribal evidence |
| line structure | WELL_COVERED | membership nulls, line/shifted/fixed scale comparison, start/end profiles | physical geometry and uncertain line segmentation |
| page/2-D structure | WEAKLY_COVERED | page-matched controls and page-boundary/persistence summaries | coordinates, diagrams, labels, columns and spatial neighborhoods |
| manuscript/regime | PARTIALLY_COVERED | blind change profiles, metadata conditioning, lag/topology and split replication | causal identity of regimes and external codicology |
| hierarchy | WEAKLY_COVERED | token→line→page→manuscript comparisons exist separately | no single nested generative/predictive hierarchy |
| compression/algorithmic measures | NOT_COVERED | entropy is probabilistic, not algorithmic complexity | compression baselines, MDL and algorithmic regularity were not tested |
| predictive structure | PARTIALLY_COVERED | held-out formation/transition models, LOBO and mechanism heldout | end-to-end blinded prediction and semantic targets |
| transformation families | PARTIALLY_COVERED | simple transposition, several homophony controls, finite M0–M11 grid | compound, historical and position-dependent systems are broad |
| information/recoverability | WEAKLY_COVERED for Voynich; PARTIALLY_COVERED synthetically | Task67 known-plaintext recovery/error battery | no Voynich plaintext oracle, key/table model or manuscript loss estimate |
| semantics | NOT_COVERED | no eligible validated decipherment in literature audit | blinded semantic association and external validation |

The most consequential blind spots are image-grounded 2-D layout, semantics, explicit transcription uncertainty, lexical-paradigm identity, and end-to-end discrimination of H_L/H_C/H_G. Listing these gaps does not prescribe a future study.
