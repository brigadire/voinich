# Rozanova–Temerev independent method record

This file freezes the protocol used by the independent analyzer in
`independent/rozanova-temerev/`. The primary source is Rozanova & Temerev,
“A Glyph Is Not a Letter, a Token Is Not a Word, a Space Is Not a Space”,
arXiv:2608.17096v1 (17 August 2026), together with the authors' public
reproducibility archive: <https://github.com/lrozanova/voynich-units>.
Task56 is navigation only; Task52 is not used.

## Estimands

For adjacent within-line token pairs `(T_i,T_{i+1})`, the token-order
estimand is plug-in mutual information

`I(X;Y) = H(X) + H(Y) - H(X,Y)`

where `X` and `Y` are token identities after the fixed top-2,000 vocabulary
cap. The 2,000 most frequent types are retained (frequency ties are resolved
lexicographically); every other type is `<other>`. Entropies and MI use
base-2 logarithms and empirical pair-position frequencies. The reported
primary value is `MI_corrected = MI_raw - mean(MI_shuffle)` and its share is
`MI_corrected / H(Y)`. This is the paper's “order share” statistic, not Stage
25 CMI and not Stage 27 transition statistics.

The glyph-edge estimand is the same plug-in MI for the pair
`(last glyph of T_i, first glyph of T_{i+1})`, with no token vocabulary cap.
Its normalized value is `MI_corrected / H(first glyph of T_{i+1})`. It is kept
separate from token identity order.

## Null and uncertainty protocol

The null is the paper's within-line shuffle: each line keeps its length and
token multiset, tokens are uniformly permuted independently within that line,
and no cross-line pair is created. The implementation uses 100 shuffles and
seed `20260816`, matching the authors' edge-order driver. It reports raw MI,
shuffle mean, sample SD, corrected MI, and normalized share. These are the
primary replication quantities; no bootstrap or extra permutation battery is
mixed into them.

## Representations and preprocessing

Input is UTF-8, one whitespace-tokenized line per source line. Empty lines are
ignored, source bytes are hashed with SHA-256, and no punctuation or case is
silently removed. Natural controls use Unicode runes as glyphs. Voynich uses
the R&T composite-collapsed EVA representation (longest-first `cth`, `ckh`,
`cph`, `cfh`, `iin`, `ain`, then `ch`, `sh`, `ee`, `in`); this is explicitly
different from Unicode-character tokenization. The canonical input is
`data_work/ZL3b-x7.canonical.txt`.

Opaque Task46/55 labels such as `x000712` have no defensible internal glyph
semantics. They are accepted for token-order MI, but glyph-edge is reported
`NOT_APPLICABLE_OPAQUE_TOKENS`; no post-hoc encoding is introduced.

Astafiev is analyzed only from the prepared UTF-8 file. The legacy source is
not passed to glyph analysis. Existing transformed files are inputs only;
Task58 creates no transformations.

## Replication classification

The R&T corpus uses a Zandbergen–Landini source and a matched 32,747-token,
3,950-line subset. Our canonical file has its own provenance and line
template, so its result is `METHOD_REPLICATION_DIFFERENT_TRANSCRIPTION`, not
`EXACT_CORPUS_REPLICATION`. The implementation is validated independently
with toy sequences and deterministic single-corpus/batch equivalence tests.

## Scope boundary

Phase A is the estimator and null validation. Phase B is the descriptive
application to Doyle, Longfellow, Astafiev, canonical Voynich, and existing
homophonic files. Stage 25/27 comparisons are descriptive only and are not
optimization targets. Agreement with a statistical profile is not a
decryption or historical homophony claim.
