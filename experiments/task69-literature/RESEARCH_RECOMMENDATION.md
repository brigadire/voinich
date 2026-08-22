# Recommendation: prioritize C — plaintext-class sensitivity

| Criterion | G: frozen multi-seed generator | L: grammar/family induction | C: fixed-encoder plaintext classes |
|---|---|---|---|
| Novelty from this search | High | High | **Very high** |
| Feasibility | High with existing Task66 mechanisms | Medium; segmentation/model-selection choices are substantial | High with existing Task66/67 encoders and known plaintext controls |
| Computational cost | Moderate | Moderate to high | Moderate |
| H_L/H_C/H_G discrimination | Moderate: a successful generator establishes non-specificity only | Moderate: grammar is also compatible with H_G unless semantic validation is added | **Highest available:** separates encoder-dominated, plaintext-conditioned, and input-insensitive architecture predictions |
| Dependence on transcription assumptions | High | High | Low for synthetic known-plaintext phase; Voynich comparison remains transcription-dependent |

**Recommendation: C.** Run a preregistered factorial experiment `F(P,E)` using fixed encoders and matched source classes (prose across languages, poetry, recipes, lists, procedural text, tables/notation, deliberately repetitive/low-entropy controls). For each pair, generate repeated encoder seeds, retain the full known plaintext, and measure the frozen Task58–67 fingerprint plus recovery/preimage measures. Use the same suites for input-shuffled and message-free generator controls.

This is more discriminating than G alone: even a generator that matches a fingerprint cannot establish absence of a message. It is less exposed than L to unresolved EVA segmentation because the first phase is entirely synthetic. It cannot identify the Voynich mechanism by itself, so its result should be used to rule in/out mechanism classes, not to select a plaintext.
