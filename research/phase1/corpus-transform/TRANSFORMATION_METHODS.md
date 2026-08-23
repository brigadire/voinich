# Transformation Methods (task46)

This document defines, precisely and reproducibly, the two token-level
historical-cipher mechanisms `corpus-transform` implements, and the design
decisions that keep them deterministic and free of hidden scientific
parameters.

**These are mechanistic controls, not a historical claim.** Token-level
transposition and homophonic substitution over an existing word-token
corpus are experiment inputs for the existing generic scientific pipeline.
They do not assert that the Voynich manuscript was, historically, enciphered
this way. They exist so the pipeline's statistics can be exercised against
corpora with a known, deliberately introduced structural distortion, as a
mechanistic baseline/control - nothing here is tuned to, or compared
against, Voynich.

## 1. Input model

The input corpus is read as a flat sequence of whitespace-delimited tokens,
`T = [t0, t1, ..., tN-1]`, using Go's `strings.Fields` per natural line
(i.e. `bufio.Scanner` line splitting, then whitespace splitting each line -
the same convention `internal/genericsegmentation.ReadCorpus` already uses
for every generic-mode pipeline stage). No additional linguistic
tokenization, case folding, or punctuation handling is performed: the input
is assumed to already be a prepared canonical corpus (e.g. via
`codex_prepare`), though `corpus-transform` does not require it - any
whitespace-tokenized text file works, matching the task's own example of
transforming `data_test/pg2097-2.txt` directly. `N` (the input token count)
is always recorded and checked against the output.

## 2. Transposition

### 2.1 Definition

Given tokens `T[0..N-1]` and a width `W >= 1`, the token stream is laid out
row-major into an implicit grid: token `i` occupies row `i / W`, column
`i % W`. The cipher output is produced by reading the grid column-by-column,
top-to-bottom, in a chosen column order (`natural` or `keyed`, section 2.2).

Concretely, for each column `c` visited (in the chosen order):

```
for row = 0, 1, 2, ...:
    idx = row*W + c
    if idx >= N: stop this column
    emit T[idx]
```

For `W=4` and `T = A B C D E F G H I J K L` (`N=12`), natural order
`0,1,2,3` gives:

```
col0 -> A E I
col1 -> B F J
col2 -> C G K
col3 -> D H L
```

concatenated: `A E I B F J C G K D H L` - exactly the worked example in
task46 section 3.

This single rule handles every edge case without special-casing:

- **Incomplete final row**: a column whose only membership would be the
  missing final row is simply shorter by one entry; no padding token is
  introduced or removed. This is the `remainder_policy` recorded in every
  transposition manifest, verbatim:

  > short-columns: token i (0-based) belongs to row i/width, column
  > i%width; a column is read by increasing row as long as
  > row\*width+column < token_count; trailing columns whose only membership
  > would be the missing final row are simply shorter, no padding token is
  > introduced or removed

- **`width=1`**: every token is its own row with one column; reading
  produces the original order unchanged (identity).
- **`width > N`**: only row 0 exists; the natural columns `0..N-1` each
  have exactly one token and columns `>=N` are empty, so natural order
  again reproduces the original order (identity); keyed order permutes the
  existing `N` columns.
- **Empty / one-token corpus**: `N=0` produces empty output; `N=1`
  produces the single token unchanged for any width.

**Invariants** (checked by tests, section 5 below, and printed in the
sanity report): output token count equals input token count, and the
output token multiset equals the input token multiset - transposition only
reorders tokens, it never adds, drops, or alters one.

### 2.2 Column order: natural vs. keyed

- `-transposition-order natural`: columns are read `0, 1, ..., W-1`.
- `-transposition-order keyed`: columns are read in a permutation of
  `0..W-1` derived deterministically from `-seed` and `W` via a
  Fisher-Yates shuffle over a seeded PRNG substream (section 4). No textual
  password/key is used or needed, per task46 section 4.

### 2.3 Rounds

`-rounds N` (default `1`) reapplies the *same* transposition (same width,
same column order, derived once from seed+width) `N` times in sequence.
This is a documented simplification, not a claim of historical
double-transposition fidelity: real double/route transpositions sometimes
use a different key per pass, but reusing one deterministically-derived
permutation keeps the manifest unambiguous (`rounds` is always recorded
alongside `width`/`order`, so single- and multi-round runs are never
conflated).

### 2.4 Reversibility

`Untranspose` (used only by tests, task46 section 18) replays the identical
column/row enumeration order used to build the transposed stream and
scatters each transposed token back to the grid index that produced it,
proving `inverse(transform(T)) == T` for every width/order/rounds
combination, including the edge cases above.

## 3. Homophonic substitution

### 3.1 Definition

Let `V` be the sorted (never map-iteration-ordered) set of distinct tokens
in `T`. Each `t in V` is assigned `H` **opaque** cipher tokens
`h(t,0), ..., h(t,H-1)`, formatted `x%06d` from a single global counter that
increments once per assigned homophone, walking `V` in sorted order. This
guarantees, structurally (not by post-hoc checking):

- non-overlapping domains across distinct plaintext tokens (the counter
  never repeats, so no cipher ID is ever assigned twice);
- cipher tokens carry no plaintext identity (`x000001` has no relationship
  to the string it stands for, unlike a scheme such as `the_1`, which task46
  explicitly forbids).

Each occurrence of `t` in `T` (in original left-to-right order) is replaced
by one of its `H` homophones, drawn from a single deterministic PRNG stream
seeded from `-seed` (section 4), consumed exactly once per occurrence in
corpus order. Occurrence order - not vocabulary/map order - drives the
draws, so the result depends only on `-seed` and the corpus's own token
sequence.

### 3.2 Homophone count model (`-homophone-model`)

- `fixed` (default): every plaintext token gets the same `H`, from
  `-homophones` (`homophonic-global-v1`).
- `frequency`: rank-quantile allocation (`frequency-v1`), specified in
  [FREQUENCY_HOMOPHONE_MODEL.md](../../../docs/literature/FREQUENCY_HOMOPHONE_MODEL.md). Here
  `-homophones` means `Hmax`; selection remains independent.

### 3.3 Selection distribution (`-homophone-selection`)

- `uniform` (default): each of the `H` homophones has probability `1/H`.
- `weighted`: **"triangular-v1"**, a fixed, versioned, corpus-independent
  formula. Homophone index `k` (0-based, in opaque-ID assignment order,
  i.e. the order the IDs were minted in) gets weight:

  ```
  weight(k) = (H - k) / (H*(H+1)/2),   k = 0..H-1
  ```

  which sums to 1 by construction (the denominator is the sum of
  `H, H-1, ..., 1`). Weights strictly decrease with `k`; homophone 0 is
  always the most probable. This formula depends only on `H` - never on the
  plaintext token, the corpus, or Voynich statistics - and is versioned so
  a future scheme can be added without silently changing this one's
  behavior.

Occurrence selection draws a uniform `float64 in [0,1)` from the seeded
stream and picks the homophone whose cumulative-probability interval
contains the draw (a standard cumulative-distribution sample; the uniform
case is simply the special case where every interval has width `1/H`).

### 3.4 Mapping audit file

Every homophonic run writes `<output>.mapping.tsv`:

```
plaintext_token	cipher_token	probability
```

one row per `(plaintext token, homophone)` pair, in the same deterministic
order as `V` (sorted) x homophone-index. This file is for
reproducibility/audit only; **the scientific pipeline never receives it**.
Its SHA256 is recorded in the manifest as `mapping_sha256`.

Frequency runs additionally write `<output>.homophone_allocation.tsv` with
`plaintext_token`, `raw_frequency`, `frequency_rank`, `frequency_quantile`,
and `allocated_H`; its SHA256 is recorded as `allocation_sha256`.

### 3.5 Reversibility

`Decode` inverts `Encode` using the mapping's reverse lookup (built from the
same non-overlapping domain guarantee above), proving
`decode(transform(T)) == T` for every tested `H`/selection combination
(task46 section 18).

### 3.6 `H=1` special case

With `H=1` every plaintext token has exactly one homophone and every draw
selects it regardless of the PRNG output: this is an ordinary monoalphabetic
opaque-token substitution, exercised explicitly by tests.

## 4. Randomness and seed derivation

All stochastic choices - the keyed-transposition column permutation, and
homophonic occurrence selection - come from `math/rand/v2` `PCG` streams
seeded via:

```
digest = SHA256("corpus-transform/v1/<purpose>/seed=<seed>/extra=<extra>")
stream = rand.NewPCG(first8Bytes(digest), next8Bytes(digest))
```

`purpose` distinguishes independent streams (e.g.
`"transposition-keyed-column-order"` vs
`"homophonic-occurrence-selection"`) so that two features derived from the
same `-seed` never accidentally share PRNG state; `extra` carries a
method-specific integer (transposition width) so different widths get
independent-looking permutations. SHA256 here is used purely as a fixed,
deterministic hash function to spread a small integer seed into PRNG state
words - **it is not an entropy source**. No current time, process ID, map
iteration order, or `crypto/rand` is ever read (task46 section 8 forbids
all four); the same corpus SHA256 + method + parameters + seed + tool
version always yields byte-for-byte identical output, verified by
end-to-end determinism tests.

## 5. Output format and invariants

Output is a plain, whitespace-tokenized text corpus: no headers, comments,
method name, seed, or markers of any kind (task46 section 9) - it must be
directly acceptable to
`pipeline-orchestrate manifest -generic-corpus -corpus <transformed-corpus>`
with no special-casing.

## 6. Line boundaries: audit finding and policy

Task46 section 10 requires an audit of whether the existing generic
pipeline treats natural corpus line boundaries as scientifically
meaningful before choosing a line-wrapping policy. **The audit finding is
that they are**: `internal/genericsegmentation.ReadCorpus` + `.Build` -
used by every one of the five generic-capable pipeline stages
(`token-relation-validate`, `replicated-local-structure-audit`,
`higher-order-sequence-validate`, `positional-continuation-validate`,
`transition-network-validate`) - derives its block/fold partition purely
from the corpus's **natural line count** `L` (`targetBlocks =
clamp(round(sqrt(L)), 8, 64)`, blocks never split a line) and returns
`ErrNotEnoughData` outright when `L < 2` or the resulting fine-block count
is `< 2`.

Consequences:

- Collapsing a corpus to one physical line (e.g. the existing
  `corpusprep`/`codex_prepare` `reflow` policy, which joins the *entire*
  corpus into a single line) would set `L=1` and make **all five**
  generic-capable stages fail outright against a transposed/homophonic
  corpus.
- Wrapping at an arbitrary fixed token width would still change `L` (and
  therefore the block-partition granularity) relative to what the *same*
  plaintext, unwrapped, would produce - i.e. it would fabricate a new,
  corpus-independent line-count parameter that silently controls
  downstream statistical granularity. That is exactly the "hidden
  scientific parameter" task46 section 10 warns against.

**Decision**: `-line-policy preserve` is the default (this intentionally
overrides task46 section 10's example suggestion of `reflow`, because the
audit that section itself demands shows `reflow` is unsafe as a default
here - the section explicitly asks to "stop and document the problem
before choosing a policy" when lines are structural, which is what this
section does).

`preserve` is defined precisely as: **chunk the final, post-transform
token stream using the input corpus's own sequence of per-line token
counts, in order.**

- For **homophonic substitution**, token order never changes, so this
  reproduces the original line boundaries exactly - line `k` of the output
  is exactly line `k` of the input, substituted token-for-token. This is an
  unambiguous, faithful reconstruction.
- For **transposition**, token order is globally rearranged by
  construction (that is the entire point of the cipher), so `preserve`
  cannot mean "line k of the output is line k of the input" - no such
  correspondence exists once tokens are read out by column. What it *does*
  guarantee is that the output has the same natural-line **count** and the
  same per-line **length distribution** as the untransformed corpus, so
  `internal/genericsegmentation.Build`'s block-partition granularity is
  identical to what it would be on the plaintext control - the one
  property needed to avoid a hidden parameter. It is not, and must not be
  read as, a reconstruction of which original tokens were adjacent.

`-line-policy reflow` is available as an explicit, non-default opt-in: it
wraps the output at a fixed, corpus-content-independent width
(`ReflowTokensPerLine = 10`, a pure serialization constant - never derived
from a corpus, a cipher parameter, or Voynich statistics). Choosing it
changes the generic pipeline's block-partition granularity relative to the
plaintext control; this must be treated as a deliberate methodological
choice by whoever runs the pipeline against the result, not a neutral
formatting option.

## 7. Manifest and mapping-file schema

See `<output>.transform.json` (task46 section 11) and
`<output>.mapping.tsv` (task46 section 12) for the exact field/column
layout; both are produced by `internal/corpustransform`
(`manifest.go`/`run.go`) and covered by end-to-end determinism tests.

## 8. Limitations and backlog

Not implemented, deliberately, per task46 sections 7 and 22:

- **`-homophone-model frequency`**: task46 section 7 requires a formula to
  be defined, documented, proven deterministic, and free of any
  Voynich-derived tuning *before* implementation - and explicitly permits
  leaving it as backlog if that is nontrivial. No formula is sketched here
  on purpose, to avoid accidentally designing one that looks tuned;
  `-homophone-model frequency` is rejected with a clear error pointing back
  to this section. A future task should define the formula, e.g. some
  monotonic function of a token's raw corpus frequency and a target
  average homophone load, document it, and prove determinism, before any
  code is written.
- Character-level substitution, character-level homophonic cipher,
  character-level transposition, nomenclator schemes, null insertion,
  padding, combined substitution+transposition, homophonic+nulls,
  frequency-flattening homophonic systems, and historically specific
  cipher reconstructions are all out of scope for task46 (see task46
  section 22) and are not implemented here.
