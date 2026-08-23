# F01 Speculum — Formal Model (Task76 Block 1)

## Definitions

- Alphabet `A`, size `m` (23 for Latin23, 26 for Modern26), circularly
  ordered: `A.At(i) = A.Letters[i mod m]`.
- Device state `S = (o_0, ..., o_{n-1})`, one angular offset per ring,
  `o_i in {0..m-1} ∪ {MISSING}`.
- Fixed conventions `K_full = (radius r, order ∈ {inner_to_outer,
  outer_to_inner}, ring identity = the identity assignment, length known)`.
- `LetterAtRadius(o, k) = A.At(o + k)`. The marked reading radius is one
  particular value of `k`, namely `r`.

## E_K(M) = S

For `M = m_0 m_1 ... m_{L-1}`, `L <= n`:

```
for i in 0..L-1:
    ring = RingPos(i)              // i itself if inner_to_outer, else n-1-i
    idx  = A.IndexOf(m_i)
    o[ring] = (idx - r) mod m       // so that LetterAtRadius(o[ring], r) = m_i
for every ring not assigned above:
    o[ring] = filler(ring)          // deterministic PRNG, models "leftover" rotation
```

Implemented in `internal/speculumf01/model.go:Config.Encode`.

## D_K(S) = M-hat

Under `K_full` and an intact `S`:

```
for i in 0..L-1:
    ring = RingPos(i)
    m-hat_i = LetterAtRadius(o[ring], r)
```

Implemented as `Config.DecodeFull`. Every other decode path (ablation,
corruption) is a deliberately *weaker* decoder that returns the full
compatible candidate set rather than a single lucky answer — see
`EXPERIMENTAL_PROTOCOL.md`.

## Baseline identity check: D_K(E_K(M)) = M

Verified two ways:

1. **Unit tests** (`internal/speculumf01/model_test.go`):
   `TestBaselineRoundTrip` over 4 hand-picked words,
   `TestEncodeDeterministic` (repeated encoding of the same message with
   the same filler seed gives byte-identical state).
2. **Full pre-registered message set** (`BASELINE_RESULTS.tsv`): **24/24**
   messages (12 natural + 12 length-matched random controls) recover
   exactly under `K_full` and an intact state.

**The identity holds for every tested message with no exceptions.** This
is expected by construction — `Encode`/`DecodeFull` are inverse functions
of each other on the same convention, with no collision built into the
core mechanism itself (each ring position holds exactly one letter of the
alphabet at the marked radius; there is no many-to-one step in the E_K/D_K
pair). Task76 asks that, if the identity does *not* always hold, the
collision classes and their conditions be described. It always holds here
for `D_K(E_K(M))`; the two collision-like phenomena the experiments did
surface both live in the *ablation/corruption* layer, not in the baseline
identity itself, and are recorded here because they are genuine formal
findings about the model:

### Collision 1 — direction ambiguity collapses for palindromes (K6)

`ANNA`, read forward or reversed, is the same string, so under K6
(traversal direction unknown) `ANNA`'s compatible set has size **1**, not
2 like every other tested message (`ABLATION_RESULTS.tsv`,
`K6_traversal_direction_unknown` rows). This is a real, source-relevant
point: **any palindromic message is immune to the one ambiguity the
source itself admits it cannot resolve** (ring-order direction is
explicitly unspecified in the fragment). It does not weaken the general
K6 finding (mean compatible-set size 1.92 across the 12 natural
messages); it is a boundary case worth stating explicitly per task76's
instruction not to let such conditions pass unremarked.

### Collision 2 — edit-distance-based error-class inference is fooled by internal repetition

This is not a collision in `D_K(E_K(M))` but a measurement artifact
reported honestly rather than hidden: `internal/speculumf01/metrics.go`
classifies a corruption's error propagation as `local` /
`synchronization` / `global` / `cascading` from the decoded string alone
(edit distance + post-first-error positional agreement), *without* being
told which corruption scenario produced it. For `orientation_mark_loss`
(a true global, single-cause uniform shift, provably a fixed-point-free
permutation on distinct letters for any nonzero shift `delta != 0 mod
m`), the classifier correctly labels 22/24 messages as `global`, but
mislabels `CONSTANTINA` and `QKCICV` — both messages with strong internal
letter repetition — as `cascading`, because the repeated letters let a
cheaper non-substitution alignment collapse the raw Levenshtein distance
below the "every position differs" signature the classifier looks for
(`CORRUPTION_RESULTS.tsv`, rows for those two messages under
`orientation_mark_loss`). **Conclusion: the authoritative error-class
label is the corruption scenario that was actually applied (recorded in
`CORRUPTION_RESULTS.tsv`'s `scenario` column); the `error_class` column
is a useful but imperfect (~92%, 22/24) post-hoc heuristic, degraded
specifically by high-repetition messages.** This is exactly the kind of
"if it does not always hold, describe the collision class and its
condition" finding task76 Block 1 asks to surface rather than smooth
over.
