# Task69: focused literature search

## G — replicable message-free generation

Timm and Schinner provide the strongest located source: a concrete, official, message-free self-citation implementation. It recursively copies and morphs earlier generated forms, has line/page state, uses configurable stochastic source selection, and can repeat a run under pseudo-RNG seed 19. Its source code makes target tuning explicit: source selection, mutation, follow rules, token-family floors, Currier-aware modes, and layout settings are selected with Voynich properties in view. The published paper reports selected similarity/network and Zipf-like results, while the software demonstrates executable generation rather than an independent evaluation.

The crucial missing experiment is not mere reproducibility of a single seed. No located study freezes all settings before evaluation, generates a predeclared multi-seed corpus series, reports replicate distributions over a broad independent fingerprint, holds metrics out of tuning, or runs a blind synthetic-versus-Voynich test. `G_PRECEDENT: PARTIAL` for precursor mechanisms and reproducible code; **`NO_DIRECT_PRECEDENT_FOUND`** for the requested frozen multi-seed broad-held-out experiment. The closest work is Timm–Schinner 2019 / the 2019 software release.

## L — bottom-up formal language

There is considerable descriptive form work. Currier reports positional and boundary constraints; Stolfi supplies a hand-built probabilistic crust-mantle-core word grammar with high in-sample coverage but explicitly permissive rules; Zattera reports a 12-slot word template; Lindemann quantifies a vocabulary-complexity proxy. These sources support that internal form restrictions exist. They do not establish that the restrictions are morphemes, lexical categories, or semantics.

No located source combines productive lexical paradigms, held-out novel-token prediction, a multi-token grammar, and matched generator controls. Timm–Schinner is especially important here because a recursive non-message generator can create related forms without morphemes. `L_PRECEDENT: PARTIAL` for formation descriptions; **`NONE_FOUND`** for the integrated validated bottom-up grammar in the task question.

## C — plaintext-fingerprint dependence

Bowern–Gaskell show that word-level metrics do not exclude several encoding transformations. Greshko's Naibbe model is the strongest Voynich-facing forward model: it encrypts known Latin and Italian text and reports selected Voynich-like outputs. Neither holds every encoder choice fixed while varying a balanced panel of plaintext classes and measuring a predeclared multiscale fingerprint. Their results therefore establish compatibility, not `ΔF_ciphertext / ΔF_plaintext`.

Shannon is the strongest general mechanistic precedent: with a fixed classical secrecy system, English redundancy and an independent equiprobable source yield different cryptanalytic/statistical consequences. Slepian–Wolf gives only a source-coding/side-information analogy. `C_PRECEDENT: PARTIAL` for directly applicable mechanism and Voynich-facing forward models; **`NO_DIRECT_PRECEDENT_FOUND`** for a systematic Voynich-like plaintext-class sensitivity experiment, especially for procedural, list, contextual, and deliberately low-entropy sources.

## Answers to the three central questions

1. **G:** No. A frozen, broadly evaluated multi-seed non-message generator statistically indistinguishable from Voynich on held-out multiscale metrics was not found.
2. **L:** No. Published templates and permissive word grammars exist, but no bottom-up system was found that also validates paradigms, predicts held-out forms, supplies sequence grammar, and defeats generator controls.
3. **C:** No. Known-plaintext forward models and general theory exist, but a systematic same-encoder plaintext-class sensitivity study for a Voynich-like fingerprint was not found.
