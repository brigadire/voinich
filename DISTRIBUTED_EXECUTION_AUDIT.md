# Distributed Execution Feasibility Audit (Task30)

> **Task31 implementation update (2026-08-16):** the recommended
> replicate-index architecture is now implemented as a bounded in-process
> goroutine pool for `conditional-regime-analyze`. Measured results,
> byte-identical hashes, checkpoint/resume evidence, and the revised
> conservative production estimate are in
> `DISTRIBUTED_EXECUTION_IMPLEMENTATION.md`. The Task30 projections below
> remain the pre-implementation model; where they differ, Task31's measured
> local-host results supersede them.
>
> **Task32 implementation update (2026-08-16):** this audit's own recommended
> "Option B (local multi-process) as the first step" (Section 10) is now
> implemented: `-executor process` runs the identical `JobID`/`JobResult`
> jobs through a bounded pool of persistent subprocess workers instead of
> goroutines, reusing the same scientific implementation and the same
> checkpoint/resume mechanism. All 19 artifacts were measured byte-identical
> to the goroutine oracle at every tested worker count, including both
> directions of cross-backend interrupt/resume. Full protocol, measurements,
> and the discovered/fixed pre-existing `report.go` map-iteration
> nondeterminism bug are in `DISTRIBUTED_EXECUTION_IMPLEMENTATION.md`.

Scope: `tasks/task30.txt`. This is an **audit and design document only** — no
scientific/statistical algorithm, RNG algorithm/seed derivation, or output
format was changed to produce it, and no worker pool or distributed
dependency was implemented. Its job is to determine whether — and how —
production execution can be distributed across workers/machines while
staying **bit-for-bit identical** to the current single-process
implementation, and to hand Task31 a concrete architecture.

Verification performed after the audit (no source changed):
`go build ./...`, `go vet ./...`, `go test ./... -count=1`,
`go test -race ./internal/conditionalregime -count=1`, `git diff --check` —
all pass. `git log -1` at audit time: `7f70fb5` ("Optimize conditional
regime distance with dense vectors", the Task29 dense-rewrite commit) — the
profiling below is against the current, already-optimized HEAD.

---

## 1. New performance profile (post-Task29)

### Workload and machine

Same development machine and corpus as Task28/29: Linux 6.6.35, Go
toolchain, 12 logical CPUs, `data_work/ZL3b-x7.txt` (39,026 tokens, 4
eligible classes), `workdir/metadata-validation/token_metadata_map.tsv`.
Combined read-only shared input is small: corpus 232KB, metadata map 2.2MB.

Reproducible command (current HEAD, `k-max-residual 15`, reduced
`-permutations 3` — large enough to see per-replicate cost twice and cross-
check against Task29's `-permutations 1` measurement, small enough to run in
minutes rather than hours per this task's own instruction not to run the
~9h full production workload just to produce a profile):

```bash
go build -buildvcs=false -o /tmp/conditional-regime-analyze-task30 ./conditional-regime-analyze

time /tmp/conditional-regime-analyze-task30 \
  -corpus data_work/ZL3b-x7.txt \
  -token-metadata-map workdir/metadata-validation/token_metadata_map.tsv \
  -output-dir /tmp/task30-output \
  -permutations 3 -k-max-residual 15 -checkpoint-path=- -quiet \
  -cpuprofile profiles/conditionalregime.task30.cpu.pprof \
  -memprofile profiles/conditionalregime.task30.mem.pprof \
  -memstats-interval 5s
```

Profiles saved: `profiles/conditionalregime.task30.cpu.pprof`,
`profiles/conditionalregime.task30.mem.pprof` (this task's fresh
measurement); `profiles/conditionalregime.task29.cpu.dense.pprof` and
`.mem.dense.pprof` (Task29's `-permutations 1` measurement, reused here as a
second, independent data point at a different replicate count).

### Headline numbers (measured)

| Metric | Value |
|---|---:|
| wall time (`-permutations 3`) | 4m48.537s |
| user CPU time | 5m15.563s |
| total CPU profile samples | 318.78s (110.5% — concurrent GC threads) |
| sampled peak `HeapAlloc` | 1,169 MiB |
| output tokens / eligible classes | 39,026 / 4 |
| output equivalence | not diffed against a baseline in this task (no code changed since Task29; Task29 already proved byte-identical before/after its own change) |

### Top CPU hotspots (flat)

| Rank | Function | Flat % | Cum % |
|---|---|---:|---:|
| 1 | `conditionalregime.euclideanDistance` | 16.04% | 16.04% |
| 2 | `runtime.mapaccess1_faststr` | 11.70% | 32.09% |
| 3 | `internal/runtime/maps.ctrlGroup.matchH2` | 8.79% | 8.79% |
| 4 | `cmpbody` (string compare) | 7.36% | 7.36% |
| 5 | `aeshashbody` (string hash) | 6.95% | 6.95% |
| 6 | `globalregime.jsDistanceSorted` | 6.00% | 35.07% |
| 7 | `runtime.scanObject` (GC) | 4.87% | 8.28% |
| 8 | `slices.partitionOrdered[string]` | 2.99% | 5.46% |

**Interpretation:** `euclideanDistance` (Task29's dense rewrite) is now the
single largest flat consumer, exactly as Task29 predicted, but the combined
map-lookup/hash/compare machinery (`mapaccess1_faststr` + `ctrlGroup.matchH2`
+ `cmpbody` + `aeshashbody`, 34.8% flat) is now **larger** than the dense
distance kernel itself — and none of it lives inside
`conditionalregime.euclideanDistance` any more (that function is
allocation-free and map-free per Task29). It comes from
`globalregime.jsDistanceSorted`/Part A's `stabilityForClass`
(`map[string]float64` profile lookups), the one hot path Task29 explicitly
did not touch (scope: conditional-regime's own residual distance loop only).
This confirms Task29's own "Phase 5" prediction that
`globalregime.jsDistanceSorted`/`stabilityForClass` would become the new
dominant *fixed-cost* consumer once the residual path was fixed — measured
here, not merely inferred.

### Allocation hotspots (`alloc_space`)

| Rank | Function | Share |
|---|---|---:|
| 1 | `globalregime.jsDistance` | 12.99GB / 33.70% |
| 2 | `globalregime.normalize` | 7.96GB / 20.65% |
| 3 | `globalregime.sortedProfileKeys` | 4.35GB / 11.29% |
| 4 | `conditionalregime.subtractProfile` | 3.95GB / 10.25% |
| 5 | `conditionalregime.standardize` | 3.86GB / 10.00% |
| 6 | `conditionalregime.denseResidualVectors` | 2.75GB / 7.14% |

Total sampled: 38.56GB alloc-space at this reduced scale. The top three are
all `globalregime`'s still-map-based profile machinery (Part A, out of
Task29's scope) — a real, already-flagged, but *not yet implemented* next
target (see Task29's own "Phase 5" answer #5), unrelated to Task30's own
job and not touched here either.

### Share of time by pipeline stage/loop (this workload)

| Component | Cum CPU | Share | Scales with `-permutations`? |
|---|---:|---:|---|
| `stabilityForClass` (Part A, per class×window, fixed observed stat) | 74.94s | 23.51% | No |
| `withinClassSweep` (Part A, per class×window, fixed observed stat) | 67.79s | 21.27% | No |
| `residualGlobalCorrection` (Part B, the permutation/replicate loop Task29 optimized) | 79.22s | 24.85% | **Yes** — 2 methods × `-permutations` |
| `withinClassSignificance` (Part A, its *own* separate permutation loop) | 11.44s | 3.59% | **Yes** — ~90 class×window×method combos × `-permutations` |
| everything else (Part C boundaries/transitions, I/O, GC, etc.) | ~85s | ~27% | No (Part C's own `residualTransitionMatrix` permutation loop measured negligible at this corpus scale — did not appear above the 1.59s node-display threshold) |

**A genuine, measurement-driven correction to Task29's estimate:** Task29's
`~7.3h` production estimate was derived by scaling only
`residualGlobalCorrection` (Part B) and treating everything else as a
"conservative fixed allowance." Fresh profiling shows `withinClassSignificance`
(Part A's own block-aware permutation test, `nullmodels.go:107-124`) **also**
scales with `-permutations` and was not separately counted. It uses the
exact same independently-seeded-per-replicate pattern
(`replicateSeed(seed, salt, i)`, `nullmodels.go:118`), so it is just as
parallelizable as Part B's loop — but it is real, additive wall-clock time
at production scale, not "fixed."

### Revised production-scale estimate (`-permutations 1000`)

Per-replicate marginal costs, cross-checked from two independent
measurements (Task29's `-permutations 1` dense run and this task's
`-permutations 3` run):

- `residualGlobalCorrection`: 26.86s / 2 replicates = 13.43s (Task29) vs.
  79.22s / 6 replicates = 13.20s (Task30) — consistent, **~13.3s/method-replicate**.
- `withinClassSignificance`: 3.97s / (90 combos × 1 replicate) = 44.1ms
  (Task29) vs. 11.44s / (90 combos × 3 replicates) = 42.4ms (Task30) —
  consistent, **~43ms/(combo×replicate)**.

| Component | Production count | Estimated time |
|---|---|---:|
| `residualGlobalCorrection` (Part B) | 2 methods × 1000 | ~7.39h |
| `withinClassSignificance` (Part A) | ~90 combos × 1000 | ~1.05h |
| `refineTopCandidates` (fixed, ≤5 top candidates × 10,000 replicates, independent of `-permutations`) | ≤5 × 10,000 | ≤0.58h |
| Fixed one-time setup (Part A discovery sweep, `stabilityForClass`, cross-block, Part C boundaries) | n/a | ~0.06h |
| **Total (revised)** | | **~8.4–9.1h** (vs. Task29's published ~7.3h) |

This is an upward revision driven purely by more complete measurement, not a
regression — no code changed between Task29 and this audit. It matters for
Section 7's Amdahl model below: **the previously-uncounted portion is just
as parallelizable as the counted portion** (same independent-seed pattern),
so the correction changes the total baseline time but does not change the
qualitative distributed-execution story.

### Whole-pipeline share of wall time

Mining `PERFORMANCE_AUDIT.md`/`PERFORMANCE_REFACTOR_REPORT.md`'s own
already-measured, most-recent post-optimization wall times for every other
CLI in the pipeline (not re-run here — no reason to, nothing there changed):

| Stage | Latest measured wall time | Production replicate count |
|---|---|---|
| `conditional-regime-analyze` | ~8.4–9.1h (revised above) | 1000 × 2 methods (+ ~90×1000 Part A) |
| `structural-projection-analyze` | 2h59m–3h25m | 200 trials |
| `token-relation-validate` | ~20min (contended) | 1000 primary + 10000 refine × 1861 candidates |
| `structural-profile-stability` | 39.0s | 200 bootstrap |
| `structural-reliability` | 22.8s | 200 bootstrap |
| `replicated-local-structure-audit` | 27.68s (reduced scale; production 1000) | 1000 |
| `metadata-validate` | 11.064s (reduced `-permutations 200`; production default 10000) | 10000 |
| `local-regime-analyze` | 45.3s | shuffle-null loops |
| `property-trajectory-analyze` | 12.44s | ~40-80 pair fits |
| `positional-continuation-validate` | 518ms | 10000 |
| `higher-order-sequence-validate` | 1.256s | 10000 primary + 1000 secondary |
| everything else (structural-analyze, soft-structural-space, normalization-compare, structural-validate, sequence-analyze, dict-analyze, graphemic, pair-decompose, distance-context, cluster-metadata-global, begin-end, transition-network) | seconds to a few minutes each | small or none |

**`conditional-regime-analyze` and `structural-projection-analyze` together
account for roughly 93–98% of total pipeline wall time** (~8.4–9.1h + ~3h out
of an ~11.5–12.5h estimated total). Every other stage combined is on the
order of 20–30 minutes. Any distributed-execution effort must target these
two stages; parallelizing the rest would not materially change end-to-end
pipeline time even if fully eliminated. **Task31 should therefore scope
itself to `conditional-regime-analyze` first** (this audit's primary focus
throughout), with `structural-projection-analyze` as an explicit, structurally
similar secondary target (Section 10).

---

## 2. Natural units of parallelism

### `conditional-regime-analyze` (primary target)

The package has **zero concurrency today** (confirmed: no goroutines,
channels, or `sync.*` anywhere in the codebase) and **zero package-level
mutable state** in `internal/conditionalregime` — every function is pure with
respect to its parameters plus a passed-in `*rand.Rand`. This is a clean
starting point for distribution.

| Level | Independent? | Read-only shared input | Per-job private state | Dependencies | Input/output size | Uniformity | Verdict |
|---|---|---|---|---|---|---|---|
| **Method** (`k_medoids` / `hierarchical`) | Yes | same | separate correction target | none | trivial | 2 buckets, uneven cost (K sweep differs) | too coarse alone (only 2-way split) |
| **Scale** (window size, Part B: 50/100/200/500/1000) | Partially — `residualSweepProgress` iterates scales sequentially per call, reusing one `prep` per scale, but scales are logically independent | tokens, blocks | one `prep`/scale | none | small | uneven (window count varies by scale) | usable but each scale's own cost is small next to the correction loop — low leverage |
| **K** (2..15) | Yes, per prep | `prep` (shared, already built) | one `fitResidualClustering` call | none | trivial | fairly uniform | cheap, not the bottleneck — not worth its own distribution tier |
| **Permutation/replicate index `i`** | **Yes — by design** (`replicateSeed`, `seeding.go`) | tokens (232KB), classes/blocks (KB), scales/K range (bytes), base seed | shuffled corpus copy + dense vectors (per replicate, transient) | **none** | job in: a few bytes (method, salt, index); job out: 1 `float64` | **highly uniform** — measured ~13.3s/replicate (Part B) and ~43ms/(combo×replicate) (Part A) across two independent measurements | **the natural unit** |

**Recommended unit of work: `(stage, method-or-combo, replicate-index)`.**
At production scale this yields `2×1000 = 2000` jobs for Part B's
`residualGlobalCorrection` plus `~90×1000 = 90,000` (small, ~43ms each) jobs
for Part A's `withinClassSignificance`, plus a fixed `≤5×10,000` for
refinement — all using the identical index-derived-seed pattern, so all are
mergeable into one worker-pool design (Section 10).

### Survey of other packages (natural unit differs per package)

Not every package's replicate loop is independently seeded at its own
finest grain — this matters for what Task31 can safely split without
touching RNG semantics (forbidden by this task):

| Package | Natural distributable unit | Why |
|---|---|---|
| `transitionnetwork` | replicate | reference pattern; `seed + rep*0x1f123bb5` per replicate |
| `structuralprojection` | **trial** (200, not sub-trial) | each of 200 trials independently seeded; **within** one trial, `GenericSmoothing`/`RandomizeProjection` consume one shared RNG stream sequentially across every bin of every token — by design (confirmed in code comments), not splittable further |
| `globalregime` | K value | `seed + k*7919` per K |
| `replicatedlocalaudit` | run | `seed + run*104729 + 11`, already checkpoint-resumable at `run` granularity — same shape as conditionalregime |
| `normalization` | (threshold, run) | independently seeded pair |
| `higherorderseq` | **candidate**, not permutation replicate | one shared RNG per candidate, consumed sequentially across that candidate's whole permutation loop |
| `positionalcontinuation` | **named test**, not permutation replicate | same shared-stream-per-test shape |
| `structuralreliability`/`profilestability` bootstraps | **bootstrap run as currently coded is a shared stream across the whole `runs` loop** — needs confirmation of caller-side seed variation before treating individual bootstrap replicates as independent | flagged, not fully resolved by this audit |
| `clustermetadataglobal` | **kind only** (`currier`/`hand`, 2-way) | one shared `*rand.Rand` per kind advanced across the whole `permutations` loop — **already documented in `PERFORMANCE_AUDIT.md` as provably non-splittable** (`rand.Perm` is self-referential Fisher-Yates; sharing/splitting draws across replicates would change which values are computed for a given seed — a real methodology change, correctly never attempted) |

None of these narrower-than-ideal units matter for the pipeline's actual
wall-clock bottom line: per Section 1, everything outside
`conditional-regime-analyze`/`structural-projection-analyze` is a rounding
error against the ~11.5h+ total.

---

## 3. RNG audit (critical section)

### `conditional-regime-analyze` lifecycle — fully traced

1. **Seed creation:** `replicateSeed(base, salt, index) = base*seedStride + salt*104_729 + index` (`seeding.go:11-13`), `seedStride = 1_000_000_007`. `methodSalt` maps `"hierarchical"→2`, else `1`.
2. **Consumption order:** for a given replicate `i`, `rand.New(rand.NewSource(replicateSeed(seed, salt, i)))` (`residualsweep.go:185`, `nullmodels.go:118`, `nullmodels.go:154`) creates one fresh `*rand.Rand`, consumed **entirely within that one function call** — one `rng.Shuffle` over the corpus blocks (`shuffleAllClassesCompact`), then one `rng.Int63()` per `(scale, K)` combination visited in the fixed `scales`/`K` loop order. Nothing escapes the call; nothing from a prior or later replicate is touched.
3. **Order dependence:** **none.** The seed is a pure function of `(base, salt, index)` — not of how many draws any other replicate consumed, not of wall-clock order, not of which replicates ran before it. This is proven in production today by `checkpoint.go`: a crash-killed run resumes at replicate `len(null)` without replaying replicates `0..len(null)-1`'s draws (`residualsweep.go:184`: `for i := len(null); i < permutations; i++`).
4. **Shared/global RNG state:** none anywhere in the codebase. `grep`-confirmed zero bare `math/rand` package-level calls and zero package-level mutable `var` in `internal/conditionalregime`.
5. **Independent computability:** yes — replicate `i` can be computed by any worker, at any time, in any order, given only `(tokens, classes, blocksByClass, scales, kMin, kMax, method, standardized, base seed, salt, i)` — all of which are either the tiny shared corpus/config or the job's own index.
6. **Deterministic seed/state retrieval for a specific job:** trivial — `replicateSeed` is pure; a coordinator computes any job's seed without running any other job.

**Conclusion for the dominant stage: distributed execution is possible with
zero RNG algorithm or seed-derivation changes.** The existing
checkpoint/resume design is already, in effect, a single-process proof that
per-replicate independence holds.

### Other packages (see Section 2's table for which unit is safe)

The survey found two categories:

- **Independently seeded at their own dominant loop's granularity**
  (`transitionnetwork`, `structuralprojection`@trial, `globalregime`@K,
  `replicatedlocalaudit`@run, `normalization`@(threshold,run)) — directly
  distributable with zero RNG changes, same reasoning as above.
- **One shared `*rand.Rand` advanced sequentially across an entire loop**
  (`clustermetadataglobal`@replicate, `higherorderseq`@permutation-within-
  candidate, `positionalcontinuation`@permutation-within-test,
  `structuralreliability`/`profilestability`@bootstrap-run) — these cannot
  be split *below* their current natural unit (candidate, named test,
  kind) without changing which random values a given seed produces, which
  this task explicitly forbids. They remain distributable *at* their
  current unit.

No package-level (Go `var`) mutable state is shared across replicate
iterations anywhere surveyed except `structuralprojection`'s scratch buffers
(`normalizeKeysScratch`, `metricsFloatSeenScratch`, `metricsFloatKeysScratch`
— `core.go:22,152-153`), which are **not RNG state** but are relevant to
Section 6/8 (goroutine-safety).

**Bottom line: yes, distributed execution can be done without changing any
RNG algorithm or seed derivation, for the pipeline's actual bottleneck
(`conditional-regime-analyze`, ~65-75% of total pipeline time) and for most
other stages at their existing natural granularity.**

---

## 4. Reduction/aggregation audit

### `conditional-regime-analyze`

`residualGlobalCorrection` builds `null []float64` by appending replicate
results **strictly in replicate-index order** (`residualsweep.go:184-186`:
`for i := len(null); i < permutations; i++ { ...; null = append(null, ...) }`),
then calls `buildEmpiricalStats(observed, null)` (`stats.go:102-114`), which
splits into two shapes:

| Function | Shape | Order-sensitive? |
|---|---|---|
| `meanFloat`, `sdFloat` | plain sequential `for _, v := range x { s += v }` | **Yes** — non-associative fp addition; gathering in a different order changes the bit pattern |
| `percentileOf` | sorts a copy first | No |
| `maxFloat` | running max | No |
| `exceedances`, `empiricalP` | count-based | No |

**This is the concrete mechanism by which a naive distributed gather would
break bit-for-bit identity:** if workers return results in completion order
rather than submission/index order, and the reducer appends them as they
arrive, `NullMean`/`NullSD` (and anything derived from `EffectSize`, which
divides by `NullSD`) would differ in their last bits from the current
single-process run — even though the *set* of values is identical.

**The fix is architectural, not numerical:** tag every job's result with its
replicate index; the reducer buffers all results and **reassembles them
into index order before calling `meanFloat`/`sdFloat`** — reproducing the
exact summation order the current single-process loop already uses. This
requires no tolerance relaxation and no change to the statistics themselves.

### Other packages

The identical pattern repeats everywhere surveyed — every permutation-null
array is built by a sequential `for i := 0; i < N; i++` loop feeding a
`mean`/`sd`-style order-sensitive reducer plus an order-independent
percentile/count reducer, with **zero exceptions found**. One package
(`tokenrelationvalidation`) is already *stricter* than conditionalregime:
its `variance`/`distribution` helpers sort a copy before computing anything,
making them order-independent by construction — a strictly safer pattern
worth noting for Task31 but not something conditionalregime needs to adopt
(matching its own current, already-verified-correct summation order is what
bit-for-bit identity requires, not switching to a "safer" implementation
that would itself change the current single-process output).

**No leftover unsorted-map float-accumulation instances were found** in
either audit pass — the task27/28 class of same-seed nondeterminism bugs
(memory: [[feedback_go_map_iteration_determinism]]) appears fully closed
project-wide, confirmed independently of this task's own scope.

**Conclusion: every package's reduction step is solvable the same way** —
tag by index, reassemble in index order, sum. This is a solved problem, not
a numerical-tolerance concession.

---

## 5. Bit-for-bit reproducibility

**Verdict: achievable, without weakening to numerical/statistical
equivalence, for `conditional-regime-analyze` (and, on the same reasoning,
for every other package surveyed at its own natural unit) — provided the
distributed architecture follows four rules:**

1. Never change `replicateSeed`/equivalent seed derivation.
2. Compute every job fully self-contained (already true today).
3. **Reassemble all per-job outputs into their original single-process
   index order before any order-sensitive floating-point reduction**
   (mean/sd); leave order-independent statistics (percentile/max/exceedance)
   alone — gather order genuinely does not matter for these.
4. Never use a hierarchical/tree/pairwise-combine reduction (idiomatic in
   many distributed frameworks) in place of the flat, index-ordered linear
   sum the current code performs — a tree reduction is mathematically
   equivalent but **numerically different** (fp addition is not
   associative), and would silently break bit-for-bit identity even with a
   perfectly correct RNG.

**Where this could break, and why it wouldn't be fundamental:**

- **Wrong reduction shape** (tree-reduce instead of reassemble-then-linear-
  sum): a real risk if Task31's implementation reaches for a standard
  MapReduce-style combiner without reading this section — an
  implementation-discipline risk, not an architectural blocker.
- **Cross-machine floating-point determinism:** Go's `float64` arithmetic is
  IEEE-754 and deterministic given identical operation order on any
  conforming target (no x87 80-bit extended-precision hazard on modern
  amd64/arm64 Go). Not a known risk here, but Task31 should still verify by
  running one replicate on each candidate worker machine type and diffing
  output before trusting this across heterogeneous hardware — cheap
  insurance, not evidence of a real problem.
- **Non-conditionalregime packages with shared-RNG-stream loops**
  (Section 3): bit-for-bit reproducible, but only at their current coarser
  unit — going finer would require an RNG semantics change, which is out of
  scope and correctly not attempted.

**No fundamental floating-point or RNG obstacle to bit-for-bit distributed
reproducibility was found for the pipeline's actual bottleneck stage.**

---

## 6. Worker resource model

### Per-replicate job (`conditional-regime-analyze`, Part B, the dominant unit)

| Resource | Estimate | Basis |
|---|---|---|
| CPU | ~13.3s/replicate, 100% CPU-bound, no I/O | measured, two independent runs |
| Peak RAM (one job's own working set) | tens of MB, not GB | shuffled corpus copy (≤232KB) + dense vectors (≤200 samples × feature-count `float64`s, feature-count bounded by residual-window token types) + one distance matrix (≤200×200 `float64`s ≈ 320KB) |
| Read-only shared input | tokens (232KB) + class/block structures (KB) + scale/K params (bytes) | broadcast once, reused by every job/worker |
| Mutable/private data | shuffled corpus buffer, dense vectors, distance matrix, cluster labels | fresh per job, discarded after |
| Job input size | a few bytes (method tag + salt + replicate index) | seed is derived, not transmitted |
| Job output size | 1 `float64` (8 bytes) | the null-max silhouette for that replicate |

The ~1.17GB sampled process-wide peak (Section 1) reflects **many
replicates' garbage awaiting GC in one long-running process**, not one
job's live set — a distributed worker computing one job at a time (or a few
concurrently) will run at a small fraction of that.

### Goroutine-pool safety (Option A, Section 8)

- **`conditionalregime`: safe as-is.** Zero package-level mutable state
  found; every function is pure w.r.t. its parameters + passed-in `*rand.Rand`.
- **`structuralprojection`: NOT safe as-is.** `core.go` has three
  package-level mutable scratch buffers (`normalizeKeysScratch`,
  `metricsFloatSeenScratch`, `metricsFloatKeysScratch`) deliberately
  introduced in Task28 to eliminate per-call allocation. These are reused
  across calls and would race under concurrent goroutines. Using Option A
  for this package would require converting these to worker-local buffers
  (e.g. `sync.Pool` or one buffer per goroutine) — a real, small, identified
  code change, out of scope for Task30, in scope for whichever future task
  applies Option A to this specific package.

### Multi-process / multi-machine (Options B, C)

Process isolation sidesteps the goroutine-safety concern entirely for both
packages — this is one reason Option B is recommended as the first step
(Section 10). Process-startup cost (reparsing the 232KB corpus + 2.2MB
metadata map per process) is milliseconds against a ~13s job — immaterial.
Network/serialization cost for a distributed job (Option C) is a handful of
bytes each way — equally immaterial against ~13s of compute.

### Can Task29's dense structures be shared without copying?

The dense residual vectors Task29 introduced are **not** a large
precomputed structure sitting outside the replicate loop — they are rebuilt
fresh (cheaply — the corpus itself is shuffled per replicate, so the
downstream residual windows genuinely differ per replicate) inside each
job. The only large, genuinely shared, read-only object is the `tokens
[]string` slice itself, which Go goroutines already share by reference at
zero copy cost (never mutated after load); for cross-process/cross-machine
workers it is a one-time 232KB broadcast, not a per-job cost.

---

## 7. Scaling model (Amdahl's law)

Using the revised production baseline from Section 1 (~8.7h midpoint of the
8.4–9.1h range) and the measured split between parallelizable
replicate-indexed work (`residualGlobalCorrection` + `withinClassSignificance`
+ `refineTopCandidates`, ≈99.3% of total at reduced scale) and fixed,
one-time setup (Part A's observed-statistic sweep + Part C, ≈0.7%):

**p (parallel fraction) = 0.993, s (serial fraction) = 0.007.**
Speedup(N) = 1 / (s + p/N).

| Workers (N) | Theoretical speedup | Realistic wall time | Parallel efficiency | CPU-hours | Estimated RAM |
|---:|---:|---:|---:|---:|---:|
| 1 | 1.00× | ~8.7h | 100% | ~8.7 | ~1.2GB (single process, measured) |
| 2 | 1.99× | ~4.4h | 99.3% | ~8.8 | ~0.5–0.6GB |
| 4 | 3.92× | ~2.2h | 97.9% | ~8.8 | ~1.0–1.2GB |
| 8 | 7.63× | ~1.14h (68min) | 95.3% | ~8.9 | ~2.0–2.4GB |
| 16 | 14.48× | ~0.60h (36min) | 90.5% | ~9.0 | ~4.0–4.8GB |
| 32 | 26.29× | ~0.33h (20min) | 82.2% | ~9.2 | ~8.0–9.6GB |

CPU-hours stay essentially flat (total *work* is unchanged by
parallelizing it — only wall time drops); the small upward drift with N is
coordination/dispatch overhead, not real algorithmic cost. RAM estimated at
~250–300MB/worker (Section 6) × N, rounding up for Go runtime overhead;
this is comfortably within a single modern server's memory even at N=32.

**Diminishing returns:** this workload's very high parallel fraction
(~99.3%) keeps Amdahl's curve favorable well past N=16, so the practical
ceiling is set by **available hardware and coordination overhead**, not by
the algorithm's own serial fraction. This development machine has 12
logical CPUs — N=16/32 require either a larger single machine or genuinely
distributed workers across ≥2 hosts (Option C, Section 8). With 2,000
Part-B jobs (+ ~90,000 small Part-A jobs) to hand out, per-job dispatch
overhead stays negligible relative to job size (~13.3s / ~43ms respectively)
up to at least a few hundred workers — tail effects from "more workers than
jobs" are not a practical concern at this pipeline's scale. **The realistic
point of diminishing returns is around 12–16 workers** (this machine's core
count, and the point where efficiency crosses ~90%); scaling to 32 is still
a genuine ~2× win over 16 but starts trading efficiency for absolute time
more aggressively, and requires ≥2 machines.

---

## 8. Architectural options

**No external infrastructure (Kubernetes, Redis, Kafka, etc.) is justified**
— job payloads are a handful of bytes each way, and the corpus/config
broadcast is a one-time few-MB transfer; a from-scratch minimal coordinator
comfortably covers this workload's actual needs.

| Dimension | A. Local goroutine pool | B. Local multi-process | C. Distributed workers (coordinator + queue + reducer) |
|---|---|---|---|
| Implementation complexity | Low (channel of job indices, worker goroutines) | Low-Medium (spawn CLI/worker-mode processes, collect stdout/results) | Medium (coordinator, job dispatch/timeout, result collection, resumable state) |
| Reproducibility | Full, if reducer reassembles by index (Section 4) | Same | Same |
| Memory efficiency | Best (shared read-only `tokens` by reference, zero copy) | Good (small per-process corpus copy, ~232KB × N processes — negligible) | Good (same, plus network transfer of the same small payload) |
| Fault isolation | Weak — a worker goroutine panic can take down the whole process | Strong — a crashed worker process doesn't affect others or the coordinator | Strongest — a crashed machine doesn't affect any other |
| Scheduling | Trivial (buffered channel) | Simple (local process pool / `xargs -P`-style) | Needs an explicit job queue + worker registration |
| Retry semantics | Simple — recompute job `i`, same seed ⇒ same result | Same | Same, plus needs a dispatch timeout + redispatch policy |
| Network/storage requirement | None | None (local processes) | Minimal — few KB/s aggregate; a small persisted job-status file (checkpoint.go-style) |
| Suitable for many corpora at once | Yes, within one machine's core budget | Yes, within one machine's core budget | Yes — this is the option that scales past one machine |
| Code changes required first | **Yes for `structuralprojection`** (scratch-buffer goroutine-safety, Section 6); none for `conditionalregime` | None for either package (process isolation sidesteps the concern) | None for either package |

---

## 9. Fault/retry semantics

| Requirement | Design |
|---|---|
| Job identifier | `(stage, method-or-combo, replicate-index)` tuple — already implicit in today's checkpoint granularity; Task31 should make it an explicit `JobID` |
| Deterministic retry | Guaranteed today by `replicateSeed`'s pure-function derivation — rerunning job `i` always recomputes the bit-identical result; no change needed |
| Duplicate execution | Harmless by construction (idempotent, deterministic) — a coordinator may safely double-dispatch a slow/suspect job; the reducer just needs to dedupe by index, keeping one result per index |
| Partial failure / worker crash | Already modeled at single-process granularity by `checkpoint.go` (`ResidualCorrectionNull` accumulates completed replicates; resume continues from `len(null)`) — Task31's coordinator is a direct multi-worker generalization: track completed job-indices, redispatch only the missing ones |
| Coordinator crash | Coordinator state is exactly "which job-indices are done, and their results" — persist it the same way `checkpoint.go` already does (atomic write-then-rename JSON), and a restarted coordinator resumes exactly where it left off |
| Checkpoint/resume | Direct generalization of the existing, already-battle-tested per-replicate checkpoint mechanism — no new design concept needed, only a wider (multi-worker) scope |
| Atomic publication of results | Mirror `checkpoint.go`'s write-then-rename pattern for final reassembled output files, so a coordinator crash mid-write never corrupts a completed run's output |
| "Repeated run of one job gives the same result" | Guaranteed by `replicateSeed`'s determinism — already satisfied, no change needed |

---

## 10. Итоговая рекомендация — Recommended Task31 Architecture

**Unit of work:** one `(stage, method-or-combo, replicate-index)` job. For
`conditional-regime-analyze` this is, at production scale: 2,000 jobs for
Part B's `residualGlobalCorrection`, ~90,000 small jobs for Part A's
`withinClassSignificance`, and a fixed ≤50,000 for `refineTopCandidates` —
all using the identical `replicateSeed`-derived pattern, so all fit one
worker-pool design. `structural-projection-analyze`'s 200 independently
seeded trials are a structurally identical secondary target once its
scratch-buffer goroutine-safety issue is resolved (or sidestepped via
Option B/C, which requires no code change there).

**Where RNG state lives:** nowhere shared. Each job derives its own
`*rand.Rand` locally from `(base seed, salt, index)` via the existing
`replicateSeed` function (or its per-package equivalents). No seed material
needs to be transmitted beyond the tiny `(method, index)` pair that already
identifies the job.

**What a worker receives:** the small read-only shared broadcast (`tokens`,
class/block structures, scale/K parameters, base seed — ~250KB total,
sent once per worker process/machine, not per job) plus the job's own
`(stage, method-or-combo, index)` triple.

**What a worker returns:** one `float64` result tagged with its replicate
index (plus the job triple, for the coordinator's bookkeeping).

**Deterministic reduction:** the coordinator buffers all `(index, value)`
results, **reassembles them into original index order**, then applies the
existing `buildEmpiricalStats` exactly as today — `meanFloat`/`sdFloat` over
the reassembled ordered slice (order-sensitive, must match), and
`percentileOf`/`maxFloat`/`exceedances` unaffected by gather order. Never
use a tree/pairwise-combine reduction.

**Recommended option for local parallel execution: B (local multi-process)
as the first step.** It sidesteps `structuralprojection`'s scratch-buffer
goroutine-safety issue for free (process isolation), gets fault isolation
without extra design work, and process-startup overhead (milliseconds) is
immaterial against ~13s jobs. Option A (goroutine pool) is a reasonable
*later* refinement specifically for `conditionalregime` (which has no
goroutine-safety issue today) if process-spawn overhead ever becomes
material — but it is not required to hit the efficiency numbers in Section 7.

**Extending to multiple machines:** Option C, built as a thin
generalization of B — replace "spawn a local process" with "dispatch to a
registered remote worker" using the same job-descriptor/result-tuple
protocol and the same checkpoint-style persisted job-status file. No new
infrastructure (no Kubernetes/Redis/Kafka) is justified given the tiny
per-job payload sizes (Section 8).

**Expected speedup:** per Section 7's table, ~8.7h → ~36min at 16 workers,
~68min at 8 workers, ~20min at 32 workers (requires ≥2 machines given this
development machine's 12 logical CPUs) — all at ≥82% parallel efficiency.

**Practical worker count:** 8–16 is the sweet spot for one reasonably sized
machine or two, at 90–95%+ efficiency; scaling to 32 is a genuine further
win (~20min total) but needs ≥2 machines and starts trading efficiency for
absolute time more aggressively.

**Can bit-for-bit compatibility be guaranteed?** **Yes**, for the
architecture above, provided Task31's implementation follows the four rules
in Section 5 exactly (unchanged seed derivation, self-contained jobs,
reassemble-by-index before any ordered sum, no tree-reduction). No
fundamental floating-point or RNG obstacle was found for this pipeline's
actual bottleneck stage or for any other package surveyed at its own
natural unit.

## Task33 implementation update (2026-08-16)

The proposed remote step is now implemented for `conditional-regime-analyze`
with standard-library HTTP, the unchanged Task31/32 `JobID`/`JobResult`,
content-addressed input staging, compatibility checks, bounded concurrency,
transport retry, and the existing atomic checkpoint/canonical reducer.
Duplicate and stale results cannot contribute. The trust boundary is
explicitly private/VPN/SSH-tunnel plus bearer token; the service is not safe
for unauthenticated public exposure. Subsequent two-machine validation used
Intel i7-8850H and AMD Ryzen 7 5700X workers with identical linux/amd64 Go
runtime: all 19 files matched exactly at 1/2/4/8/16/32 slots. Compatibility
checks remain enforced rather than weakening byte identity to tolerance.
Implementation evidence and limits are in
`DISTRIBUTED_EXECUTION_IMPLEMENTATION.md`; commands are in
`DISTRIBUTED_EXECUTION_OPERATIONS.md`.
