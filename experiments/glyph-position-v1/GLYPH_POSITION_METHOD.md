# Task59 — Glyph Positional Specialization: Method

Written before results are interpreted, per task59 section 3. Describes exactly
what `independent/glyph-position-analyze` computes.

## Glyph definition

Voynich tokens are parsed with the same EVA composite-collapsing rule as
Task58 (`internal/evaglyph.CollapseEVA`, shared code, not a re-implementation):
the token is lowercased, then the composites `cth,ckh,cph,cfh,iin,ain` (longest
first) and `ch,sh,ee,in` are greedily collapsed into single atomic uppercase
symbols (`C,K,P,F,N,A,H,S,E,I`); every remaining character is its own glyph.
This is a fixed, mechanical rule — not a semantic judgement about "real"
Voynich glyphs — applied identically everywhere the corpus is read.

Natural-language controls (Doyle, Longfellow, Astafiev) use
`internal/evaglyph.NaturalGlyphs`: lowercase, Unicode letters/digits only,
punctuation and whitespace stripped. Each surviving character is one glyph.
This is a coarser, standard definition (glyph = character); it does not need
to match Voynich's composite rule, since natural-language controls are a
comparison baseline for magnitude, not a claim of representational parity.

## Position classes

For an n-glyph token, glyph index i (0-based):

- n == 1: `SINGLETON`
- i == 0: `INITIAL`
- i == n-1: `FINAL`
- otherwise: `MEDIAL`

SINGLETON is never folded into INITIAL/FINAL/MEDIAL; it is reported as its own
count and excluded from the 3-way INITIAL/MEDIAL/FINAL dominant-share/entropy
calculation (which is over exactly the categories that apply to that token).

## Per-glyph statistics

For every distinct glyph g: `N, N_initial, N_medial, N_final, N_singleton`;
`dominant_position` and `dominant_position_share = max(P(initial|g),
P(medial|g), P(final|g))`; Shannon entropy `H(position|g)` and
`H_norm(g) = H(position|g) / log(K)` where K is the number of applicable
categories (3, or 4 if the glyph occurs as a singleton at all).

## Null models

- **Within-token shuffle** (primary null, section 11): for each token
  occurrence, permute its glyphs among themselves. This preserves token
  boundaries, token lengths, the multiset of glyphs per occurrence, and every
  glyph's global frequency; it destroys positional ordering only. 1000
  permutations, seed `20260822`.
- **Global glyph shuffle** (section 12, stronger null): redistribute all glyph
  occurrences across all token slots corpus-wide, preserving the token-length
  sequence and global glyph frequencies but not per-token co-occurrence.

Per-glyph z-score and empirical p-value (`(#null >= observed + 1)/(n+1)`) are
computed against the within-token null; q-values via Benjamini-Hochberg FDR
(section 14).

## Frequency stratification

`FREQUENCY_STRATIFICATION.tsv` reports, at N-thresholds `5,10,30,100,300`: how
many glyphs clear the threshold, and among those, how many reach
dominant-share `>=0.90`, `>=0.95`, and `=1.0` (sections 7-9). The corpus-level
comparison table additionally reports `HighFreqSpecialists` (N>=100 AND
share>=0.95) and `Exclusions` (N>=100 AND some position count is exactly 0).

## Synthetic homophony controls (sections 17-22)

Two independent mechanisms, each a real seeded `math/rand.Rand`, not a
formula of the occurrence's own position — this is the point of the negative
control, and the property an earlier draft of this tool got wrong (a
position-index-derived homophone id silently reintroduced the position
signal the control is supposed to rule out; fixed before these results were
produced):

- **Position-independent (negative control), H in {2,4,8}**: every plaintext
  glyph occurrence independently draws `k = r.Intn(H)` and is relabeled
  `glyph_k`. The draw does not depend on the occurrence's within-token index,
  so under the null hypothesis of section 18 the homophones should roughly
  inherit the plaintext glyph's own positional distribution rather than
  create new specialization.
- **Position-dependent (positive control), H=4**: each occurrence still draws
  `k = r.Intn(H)`, but the label is `glyph_{position-class}_k` — a distinct
  homophone-symbol pool per INITIAL/MEDIAL/FINAL/SINGLETON class. By
  construction every concrete synthetic symbol occurs at exactly one position
  class, so this must be detected as maximal specialization (dominant share
  = 1.0 for every such symbol); it validates that the analyzer can detect
  artificially created positional classes at all, exactly as section 21
  requires. It is not a Voynich model and not a decipherment candidate.
- **Structured-token positive control (section 22)**: synthetic tokens of the
  fixed shape `Prefix Core Core Suffix` drawn from three disjoint alphabets
  (`P0-P2`, `C0-C6`, `F0-F2`), independent of any natural corpus. Also
  expected to show maximal specialization.

## Glyph-edge comparison

`TASK58_EDGE_COMPARISON.tsv` reports `I(last(T_i); first(T_i+1))` using
`internal/evaglyph.MI`, the same discrete-MI estimator Task58 uses for its
glyph-edge estimand (section 23-24) — plugged in, not re-derived, so the two
experiments' edge numbers are directly comparable.

## Per-corpus detail files

`controls/`, `homophony/`, and `positive-controls/` each hold one
`<Corpus>_GLYPH_POSITION.tsv` per corpus in that category, in the same
per-glyph format as the top-level `VOYNICH_GLYPH_POSITION.tsv`, computed at
50 within-token-shuffle permutations (`comparison_control_permutations` in
the manifest) rather than the primary 1000 used for Voynich itself.
`POSITIONAL_SPECIALIZATION_COMPARISON.tsv` is the corpus-level rollup of the
same runs.

## Provenance

`manifest.json` records the corpus path/SHA256/token and type counts, git
commit and dirty state, the parser/position-class/normalization rules above,
the permutation counts and seed, and which controls were run.

## Interpretation limit

Positional specialization in natural language is expected (morphology,
orthography, final letter forms) and is not, by itself, evidence of a cipher.
This document fixes what is measured and how; `REPORT.md` and
`POSITIONAL_SPECIALIZATION_COMPARISON.tsv` separate the observation from any
mechanistic or historical interpretation, per section 34.
