# Task69 focused-search log

**Date:** 2026-08-23. This is a new targeted search extending, rather than treating as sufficient, Task68's catalog and search log. Evidence is limited to primary publications, primary preprints, historical primary sources, and official author/reproducibility repositories. Publisher abstracts and bibliographic records are labelled as such; blogs, forums, and claimed decipherments were used only for discovery or excluded.

## Track G: replicable message-free generation

Queries: `"Voynich generator"`, `"Voynich generating algorithm"`, `"pseudo-text generator Voynich"`, `"autocopying Voynich"`, `"self-citation Voynich"`, `"copy modify Voynich"`, `"Markov Voynich"`, `"random walk Voynich"`, `"hoax generation Voynich"`, author searches for `Torsten Timm` and `Andreas Schinner`.

Citation chain: Task68 `TIMM_SCHINNER2019` and `SCHINNER2017` records → Crossref/OpenAlex bibliographic records → Timm arXiv author records (`1407.6639`, `1601.07435`) → official `TorstenTimm/SelfCitationTextgenerator` source and Zenodo `10.5281/zenodo.2531632`. The implementation was inspected because it is the only located executable message-free generator. It has repeatable pseudo-RNG output but also a seed-selection mode and explicit Voynich-targeted controls.

**Negative result:** no located primary source froze a message-free generator, generated many predeclared independent seeds/corpora, and evaluated a broad held-out multiscale fingerprint against Voynich. This is `NO_DIRECT_PRECEDENT_FOUND`, not a claim that no unpublished work exists.

## Track L: bottom-up formal language

Queries: `"Voynich morphology"`, `"Voynich morphological segmentation"`, `"Voynich word grammar"`, `"Voynich formal grammar"`, `"Voynich finite state"`, `"Voynich paradigms"`, `"Voynich token families"`, `"Voynich grammar induction"`, `"Voynich productivity held-out"`, `"Voynich syntax"`.

Citation chain: Currier primary report → Stolfi's cited grammar resource → ACL Reddy–Knight → IVCM/CEUR 2022 primary papers (Bowern–Gaskell, Lindemann, Zattera) → Rozanova–Temerev unit-control preprint → Timm–Schinner generator countermodel. The search distinguishes a template or in-sample parser from a grammar that predicts unseen forms and is tested against generators.

**Negative result:** no located primary source jointly reports bottom-up formation grammar, functionally supported lexical paradigms, held-out productive form prediction, multi-token grammar, and generator controls.

## Track C: ciphertext fingerprint dependence on plaintext

Queries: `"ciphertext statistics plaintext dependence"`, `"plaintext statistics preserved by cipher"`, `"cipher entropy plaintext entropy"`, `"homophonic cipher plaintext statistics"`, `"same cipher different plaintext ciphertext statistics"`, `"Voynich plaintext dependence"`, `"Voynich source text cipher"`, `"Voynich codebook"`, `"contextual messages source coding side information"`.

Citation chain: Task68 Bowern–Gaskell and Greshko records → CEUR primary paper and official Naibbe repository/Zenodo artifact → Trithemius historical primary book → Shannon 1949 source-redundancy analysis → Slepian–Wolf 1973 conditional-information analogue.

**Negative result:** no located primary study applies a fully fixed Voynich-like encoder to matched prose, poetry, recipes, lists, notation, and deliberately low-entropy/contextual sources, then reports repeated-seed, predeclared full-fingerprint sensitivity. General mechanistic sources are explicitly not evidence about Voynich.
