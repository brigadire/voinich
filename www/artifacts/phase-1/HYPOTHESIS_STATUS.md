# Status of broad hypotheses after Phase I

These hypotheses are model classes, not mutually exhaustive implementations. “Supports” means compatibility or relative plausibility under a stated test; it is not posterior probability. The claim-level literature matrix is [HYPOTHESIS_EVIDENCE_MATRIX.tsv](../literature/HYPOTHESIS_EVIDENCE_MATRIX.tsv).

## H_L — natural or artificial/formal symbolic language

**Supporting observations.** Non-random co-occurrence, edge coupling, boundary/position specialization, repeatable sequences, local regimes and manuscript heterogeneity are compatible with an organized symbolic language. Published Currier/section effects and co-occurrence results point in the same broad direction.

**Contradictory or tension-producing observations.** If visible EVA tokens are treated as ordinary prose words, corrected adjacent-token order is very weak; conditional glyph entropy is atypically low under stated representations; near-edit adjacency and local form families are unusually strong relative to natural controls. These are weak/narrow contradictions to the *ordinary visible word-token* version, not to artificial language, abbreviatory writing, altered segmentation or encoded linguistic units.

**Non-discriminating observations.** Positional glyphs, Zipf-like marginals, Currier partitions, page clustering and repetition can arise in languages, encodings and structured generators.

**Missing tests.** No validated semantics, translation, external referent prediction, compositional generalization or independent unit mapping. Status: **VIABLE, NOT SELECTED**.

## H_C — complex transformation of meaningful plaintext

**Supporting observations.** Greshko's reversible Naibbe forward model shows that structured verbose homophonic transformation can reproduce selected Voynich-like statistics. Task66 preserves source dependence in 33/36 strong/partial comparisons while constrained mechanisms improve compatibility. Task67 shows recoverability can depend on codebook/state and differ from surface statistics.

**Contradictory or tension-producing observations.** The selected simple inverse transposition worsened all four holdout targets. Simple position-independent homophony did not jointly reproduce positional specialization or near-repeat geometry. No inverse homophony result exists because its synthetic gate failed. No validated Voynich plaintext recovery exists.

**Non-discriminating observations.** Low entropy, local repetition, positional structure and regimes can be produced by both encodings and message-free systems. Synthetic reversibility does not imply historical use.

**Missing tests.** No comprehensive historically specified transform matches the full held-out fingerprint and recovers known or Voynich plaintext under declared external knowledge. Transcription/segmentation damage and key/table availability are unknown. Status: **VIABLE, SIMPLE VARIANTS DISFAVORED, NOT SELECTED**.

## H_G — structured generation without a transmitted message

**Supporting observations.** Copy/modify and constrained formation mechanisms can reproduce selected repetition/form properties; visible-token order is weak; Task66 finds constrained formation more useful than added randomness. Literature demonstrates that selected “language-like” marginals are not language-specific.

**Contradictory or tension-producing observations.** No tested message-free generator covers the whole v1 fingerprint on held-out data. Some sequential, boundary and regime structure persists after matched controls. Task66's compatible candidates generally preserve input dependence, so improvement does not require message erasure.

**Non-discriminating observations.** Repetition, edit families, low entropy, Currier-like variation and page clustering are all compatible with tuned generation and with transformed or formal meaningful systems.

**Missing tests.** No preregistered full-fingerprint, three-class held-out comparison; no operational test establishes absence of transmitted information. Status: **VIABLE, NOT SELECTED**.

## Why Phase I does not decide

Each broad class contains mechanisms capable of producing some measured properties. Most published and local controls compare only one or two classes, and statistical fit is not semantic evidence. Task67 further shows that surface compatibility does not identify reversibility or information retention. The defensible Phase I status is therefore `PARTIALLY_DISTINGUISHABLE`: narrow mechanisms have changed status, but H_L, H_C and H_G remain open.
