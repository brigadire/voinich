# Rarefaction protocol v1 (B03)

Executable implementation: `internal/notation/rarefaction.go`
(`Rarefy`, `RunRarefaction`). CLI: `notation-corpus rarefy`.

## 1. Sampling unit

The boundary-preserving sampling unit is a **structural block**: the
contiguous run of USC records sharing the same value of `lineKey`
(`document\x1fsection\x1fpage\x1flocus\x1fphysical_line`).

- If physical lines are source-observed, a block is exactly one physical
  line.
- If they are not, every deeper level is uniformly unobserved (validated at
  normalization time), so `lineKey` collapses to the shallowest
  source-observed level actually present — the deepest observed sequence
  unit required by section 6 of the task, with no separate code path.

No sampling procedure ever splits a block, and `sameSequenceUnit` (used by
every S/D metric and by the accumulation curves) compares the full five-part
hierarchy tuple, so two blocks that were not already adjacent in the source
can never produce a fabricated token transition when concatenated in a
different relative order.

### Shared draw across families

A single boundary-preserving draw is used for G, T, S, and D at each
checkpoint/replicate (`FamilyGroupStructural`); L uses its own draw
(`FamilyGroupLine`) because L requires source-observed physical lines
specifically and must never fall back to a coarser unit — if physical lines
are not observed, L is `NOT_COMPARABLE` at every checkpoint and is never
drawn. This is a documented reduction of the "one seed per metric_family"
schedule in the task text to two seed groups; it does not use one
destructive sampling procedure for every family; it uses one
boundary-preserving procedure whose unit already satisfies G/T/S/D and a
narrower one for L, which is what the task's own family rules resolve to
whenever physical lines exist (the common case for every corpus this study
considers).

## 2. Target size semantics and overshoot

`Rarefy(records, requestedN, seed)`:

1. Enumerates blocks, shuffles their order with a `math/rand` source seeded
   by `SeedFor`.
2. Walks the shuffled order, accumulating whole blocks.
3. When adding the next block would reach or exceed `requestedN`, it
   compares `|actual_with_block - requestedN|` against
   `|actual_without_block - requestedN|` and keeps whichever is closer.
4. **Tie-break**: on an exact tie, or when `actual_without_block == 0`
   (a single block already meets or exceeds the checkpoint), the block is
   included. This is deterministic and applied identically to VM,
   calibration controls, and any future candidate (R6).
5. If the corpus's total token count is below `requestedN`, `Rarefy` returns
   an error and the checkpoint is `NOT_COMPARABLE` (never truncated).

`requested_N`, `actual_N`, and the resulting deviation are all retained in
`RAREFACTION.tsv`.

## 3. Replicates

`R = 100` per checkpoint/corpus/representation/family-group
(`RarefactionReplicates`). A runtime benchmark on the frozen VM source
(39,380 tokens) measured a single full `Analyze()` pass at
5k/10k/20k/39380 tokens at approximately 0.13s/0.39s/1.4s/4.9s
(≈6.8s summed per replicate); at R=100 with two family-group draws per
replicate this is ≈23 CPU-minutes for one corpus/representation, which is
tractable as a one-time freeze cost and was not reduced.

## 4. Seed schedule

`SeedFor(base_seed, corpus_id, representation_id, family_group, checkpoint,
replicate_index)` (`internal/notation/seed.go`) derives a seed by SHA-256
hashing the tuple and taking the first 8 bytes as a non-negative int64.
`family_group` is `STRUCTURAL` or `LINE` (see above). `base_seed = 20260830`
is frozen (`notation.BaseSeed`). No runtime-random seed is used anywhere in
this pipeline.

## 5. Accumulation curves

`A2`/`A3`/`AT` are recomputed for every rarefaction replicate directly over
the drawn (already boundary-preserving) record set — never by re-slicing a
prefix of it — via `AccumulationCounts`, and stored as `family="CURVE"` rows
alongside G/T/S/L/D. The existing raw (non-rarefied) prefix curve in
`Analyze()` is retained unchanged as the raw-density estimate; both are
explicitly kept per COMPARATIVE_EXPERIMENT_SPEC.md ("Raw density and
rarefied estimates are both retained").

## 6. Validation (R1-R6)

All six pass in `internal/notation/rarefaction_test.go`:

- **R1 determinism**: identical `(records, seed)` → byte-identical draw.
- **R2 boundary preservation**: every sampled line is either fully included
  or fully excluded.
- **R3 no synthetic transitions**: no drawn record pair from different
  blocks has consecutive `token_index` — every adjacency in the draw came
  from a real source adjacency.
- **R4 size accounting**: `actual_N == len(records)` in the draw, and stays
  within one block's size of `requested_N`.
- **R5 order preservation**: token order within every retained block is
  unchanged.
- **R6 VM/candidate symmetry**: `RunRarefaction` is exercised on two
  differently-shaped synthetic corpora with the identical code path used for
  VM (`vm-reference`/`rarefy` CLI) and for calibration; there is no
  corpus-specific branch anywhere in `rarefaction.go`.

`B03 = CLOSED` (see `PREPARATION_BLOCKERS.md`).
