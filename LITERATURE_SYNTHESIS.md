# Voynich manuscript literature audit: language, encoding, and non-message generation

## 1. Scope and method

This audit separates direct observations from interpretations at claim level in `CLAIMS.tsv`. It covers peer-reviewed papers, proceedings, preprints lacking a reviewed version, historical primary material, and official repositories; the search record and exclusions are in `SEARCH_LOG.md`. Abstract-only records are never used for a detailed controls audit. “Supports” in the matrix means support for a result under specified controls, never a vote for a hypothesis.

## 2. Firmly established observations

The most durable observations are (a) a reproducible Currier A/B statistical partition, (b) strong positional specialization of some glyphs, (c) weak average corrected adjacent-token dependence alongside non-zero glyph-edge coupling, (d) atypical low character conditional entropy under stated representations, and (e) local repetition/form-family and local-regime structure. These are properties of transcriptions and estimands, not demonstrated properties of plaintext words, language identity, or a cipher key.

## 3. Evidence compatible with language, and its limits

Frequency regularity, non-random co-occurrence, section association, and topic-model page clustering are compatible with organized linguistic content (Reddy–Knight; Montemurro–Zanette; Sterneck–Polish–Bowern). Montemurro–Zanette's shuffled baselines establish non-randomness, not semantics: no held-out semantic prediction, message-bearing/cipher comparison, or structured-generator control was supplied. Topic associations are similarly confounded by Currier class, hands, section, layout, and local adjacency.

Low corrected token-order information and atypical entropy are problematic for treating visible EVA glyphs, spaces, and tokens as ordinary letters, word spaces, and words. They do **not** reject language represented through abbreviation, a codebook, altered segmentation, or encoding.

## 4. Evidence compatible with encoded plaintext, and its limits

Reddy–Knight and Bowern–Gaskell show that familiar word-level statistics do not exclude at least some cipher families. Greshko's specified Naibbe cipher is stronger mechanistic evidence: a known plaintext can be transformed to reproduce selected Voynich-like properties. It remains a forward compatibility demonstration, not an inverse recovery of Voynich plaintext, and it does not cover all reported local/positional/entropy properties under a frozen full metric suite.

Trithemius supplies a historical analogue for the essential distinction `C` versus `(C,K)`: ciphertext may be opaque without an external table while remaining recoverable with it. It is not evidence that Voynich uses such a system. No searched primary work directly measures the proposed link between a Voynich-like fingerprint, many-to-one information loss, and recoverability. That result is therefore `NO_DIRECT_PRECEDENT_FOUND`, not evidence that loss occurred.

## 5. Evidence compatible with structured non-message generation, and its limits

Schinner and Timm–Schinner identify repetition and form regularities, and the latter gives a concrete self-citation/copy-modify generator. This is decisive against treating the reproduced selected statistics as language-specific. It is not decisive evidence that the manuscript contains no message: the generator was selected against a subset of target properties and does not test whether a message-bearing mechanism can also match them.

## 6. Transcription, copying, Currier, and local structure

Published results use heterogeneous transcriptions and unit definitions. Entropy values especially cannot be compared as identical without character inventory, boundary treatment, estimator, and corpus-size details. Task61 independently confirms this sensitivity. Currier A/B is statistically reproducible, but neither Currier nor later classifier work establishes “two languages.” Task65 further shows why hand/section/page metadata and within-page local decay must be conditioned separately.

## 7. Controls, replication, and contradictions

The controls audit is represented at claim level and the replication/contradiction matrices. The recurring weakness is a one-class comparison: prose-only comparisons cannot select between encoding and generation; a successful generator cannot select H_G; and a forward cipher cannot select H_C. The direct R&T estimands have the clearest independent reproduction here (Task58), while entropy is a partial representation-dependent replication (Task61). No full independent replication of the self-citation or Naibbe model on a preregistered, multiscale held-out fingerprint was found.

## 8. Information loss and recoverability

Existing literature distinguishes recoverability with a key/table from ciphertext-only recovery in historical systems, but does not supply a direct Voynich-specific experimental test of intrinsic loss, transcription conflation, segmentation damage, or a fingerprint/recovery trade-off. Task67 is a synthetic known-plaintext extension: it shows that in tested constrained mechanisms statistical dependence, recoverability, and robustness are distinct. It does not infer information loss in Voynich.

## 9. Comparison with Tasks58–67

`TASK58_67_LITERATURE_CROSSWALK.tsv` was prepared after the source/claim audit. Tasks58–61 independently reproduce directions already in the literature with stated differences. Tasks62–67 chiefly add controlled, held-out mechanism and recoverability tests for which no direct methodological precedent was found. Their strongest joint conclusion is narrow: ordinary visible word-token language and simple position-independent homophony are insufficient for particular observations; no broad hypothesis is thereby eliminated.

## 10. Conclusion

**PARTIALLY_DISTINGUISHABLE.** Existing literature distinguishes some *narrow models* but cannot distinguish the broad alternatives H_L (language), H_C (encoded message), and H_G (structured non-message generation). The evidence matrix contains many `NON_DISCRIMINATING` cells because the same features occur under multiple classes and controls are usually incomplete. The highest-value next tests are the blinded semantic-association, full held-out three-class model comparison, and known-plaintext recoverability experiments in `RESEARCH_GAPS.md`.
