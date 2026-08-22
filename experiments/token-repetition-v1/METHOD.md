# Task60 — Voynich Exact Runs, Near-Repetitions, and Illustration Labels: Method

Written before any number below was computed, per task60's repeated instruction
not to choose thresholds/parameters post-hoc. Describes exactly what
`independent/token-repetition-analyze` computes and why.

## 0. Scope decisions

Task60 is a large independent study (50 sections). To keep it tractable while
staying faithful to every required estimand, the following engineering
choices are fixed here, before any corpus is scored:

- **Adjacency** always means the immediate pair `(w_i, w_i+1)` — no wider
  window — and, matching Task58's own "adjacent" definition (its MI
  estimator skips any pair whose line IDs differ), a pair that crosses a
  natural line boundary is never counted as adjacent. This keeps "adjacent"
  meaning the same thing across Task58 and Task60 rather than defining it a
  second, incompatible way.
- **Edit-family graph** (section 23, explicitly exploratory) uses a
  deletion-signature index (each token type of length L generates its L
  possible length-(L-1) deletions; two types sharing a deletion-signature
  bucket, or an exact string match at length-difference 0/1, are edit-distance
  candidates whose true Levenshtein distance is then verified exactly) — this
  avoids the forbidden O(V²) all-pairs scan (section 48) at the cost of
  missing edit-distance-1 pairs that do NOT share any single-deletion
  signature, which cannot happen for true distance-1 pairs (a substitution
  pair of equal length always shares neither's own deletions directly, so
  substitutions are instead found by checking, for tokens of equal length,
  Hamming distance 1 directly within same-length buckets — both mechanisms
  are combined; see `families.go`).
- **BEGIN/MIDDLE/END** operation position (section 21) is the changed/
  inserted/deleted glyph's 0-indexed position divided by the length of the
  longer of the two tokens in the pair, bucketed at thirds
  (`<1/3`,`[1/3,2/3)`,`>=2/3`).
- **Frequency/length-matched null** (section 20): for each observed
  edit-distance-1 adjacent pair `(a,b)`, draw `matchedNullDraws` (frozen
  below) independent random pairs `(a',b')` from the corpus vocabulary such
  that `len(a')=len(a)`, `len(b')=len(b)`, and `a'`/`b'` are drawn from tokens
  whose frequency rank is within `rankTolerance` of `a`/`b`'s own rank (falling
  back to the nearest available rank window if too few candidates exist at
  that length) — then measure what fraction of matched draws are themselves
  edit-distance-1, giving a length/frequency-controlled expected rate to
  compare the true adjacent-pair rate against.
- **Null models**: `nullPermutations=1000` for the exact-repetition global
  and within-line shuffle nulls, and the same count
  (`nearNullPermutations=1000`) for the near-repetition (P(d<=1)) global and
  within-line shuffle nulls, matching Task59's precedent; `matchedNullDraws=200`
  per observed edit-distance-1 pair for the more expensive length/frequency-
  matched null (section 20). The edit-family adjacency-enrichment check
  (section 24) uses the closed-form bigram-independence expectation
  `freq(a)*freq(b)/(N-1)` summed over the edit-distance-1 graph's edges
  (the same independence formula Task57's transition estimator uses) rather
  than a separate resampling procedure - a deterministic, exact expectation,
  not an approximation needing draws.
- **Chains** (section 26): maximal runs of glyph types `A -> A' -> A'' -> ...`
  in the edit-family graph where each consecutive pair has distance <=1,
  reported for chain length >= 3, using the graph already built for section 23
  (not a separate search).
- **Homophonic dose-response** (sections 13-14) reuses the token-level
  ciphertext corpora already on disk under `data_test/transformed/` (Task46/55)
  for the *exact*-repetition/run family, since exact-token-equality is
  well-defined on opaque `xNNNNNN` labels. It does NOT attempt near-repetition
  on those corpora (section 27: `NOT_APPLICABLE_OPAQUE_TOKENS`, enforced in
  code, not just by convention).
- **Glyph-level homophony for near-repetition** (section 28) reuses Task59's
  fixed, shared homophone-assignment mechanism
  (`internal/evaglyph.RandomHomophony`: every plaintext-glyph occurrence
  independently draws `k = r.Intn(H)` via a seeded PRNG, entirely independent
  of token-internal position) applied directly to a natural corpus's own
  glyph representation (Doyle), producing real synthetic glyph strings — not
  Task46/55's opaque token IDs — so glyph-level edit distance stays
  meaningful. H in {2,4,8}.
- **Label matched-size subsampling** (sections 30-31): `labelSubsamples=500`
  random samples (without replacement within each draw) of the running
  canonical corpus, each the same token count as the label corpus, to build a
  null distribution for V, hapax/type, adjacent-repeat rate, and
  edit-distance-1 adjacency rate.
- **Seed**: `20260823` (base). A single `math/rand` source is created from
  it once, at the start of the run, and threaded through every permutation/
  sampling procedure in this document's order (exact-repetition nulls per
  corpus, near-repetition nulls per corpus, matched-null draws, glyph-level
  homophony controls, label subsampling) - deterministic given the fixed
  code path and call order (verified by re-running end-to-end and diffing
  every output byte-for-byte), not by re-seeding a fresh source per
  procedure.

## 1. Input

Running text: `data_work/ZL3b-x7.canonical.txt`, read with
`internal/genericsegmentation.ReadCorpus` (same reader every other generic
stage uses) for tokens + natural line boundaries + SHA256.

Labels: parsed from the raw IVTFF source `data/ZL3b-n.txt` with
`internal/metadatavalidation.ParseIVTFF`, filtering `Locus.Type == "L"`
(labels, IVTFF locus-type letter `L`); each locus's `AlignmentText` is
whitespace-split into label tokens (the same dot/comma/gap-as-word-break
convention already used to build the canonical corpus). No metadata field
beyond locus type is read (no Currier/hand — task60 section 23 explicitly
forbids using that as a recovery hint; here it is simply never touched).

**Caveat, disclosed rather than silently worked around**: `data_work/ZL3b-x7.canonical.txt`
was built via `ivtt -x7 data/ZL3b-n.txt data_work/ZL3b-x7.txt` (see README.md),
and `-x7` does not itself restrict which locus types are kept. A direct check
confirms at least some label-locus text is present verbatim in the canonical
corpus. The label-vs-running-text comparison below is therefore not
guaranteed to be against a fully disjoint stream; labels may be a (small)
subset of the running-text token pool rather than wholly external to it. This
is reported as a limitation (task60 section 46), not corrected by rebuilding
the canonical corpus (out of scope; section 2 specifies using the existing
canonical corpus as-is).

## 2. Natural-language and homophonic controls

Doyle, Longfellow, Astafiev (already-prepared UTF-8 corpora, same files
Task58/59 use) plus the existing Task46/55 homophonic series on disk. No new
transformation model is created (task60 section 12/38).

## 3. Statistical outputs

See `REPORT.md` section list for the full breakdown; each TSV under this
directory corresponds 1:1 to a task60 section (see file docstring/manifest
for the mapping). Class recovery / semantic-identity claims are out of scope
throughout — this is a structural-statistics study, not a decryption attempt.
