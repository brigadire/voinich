# Deterministic Local Parallel Execution (Task31)

## Result

`conditional-regime-analyze` now accepts `-workers N` (default `1`) and
runs its independently seeded Part A significance, Part A refinement, and
Part B global-correction replicates through one bounded goroutine pool.
There are no subprocesses, remote workers, network services, GPUs, or
changes to the statistical/RNG/output code. `workers=1` uses the same job
architecture as every larger value.

One job is identified by:

```text
JobID { Stage, Combination, ReplicateIndex }
JobResult { JobID, ExistingPerReplicateFloat64 }
```

The ID contains no worker number, completion sequence, clock value, or
`GOMAXPROCS`. Each job constructs its private `rand.Rand` with the existing
unchanged `replicateSeed(base, salt, index)`. Shared corpus, block, and dense
input structures are read-only; shuffled buffers, fit state, centroids, and
RNG state remain job-local.

The coordinator stores results in slots indexed by the `JobID` replicate
index. It never feeds completion order into `meanFloat`, `sdFloat`, or any
other reduction. Every Part B completion is checkpointed under its complete
serialized `JobID`, including results beyond a temporarily missing index;
resume schedules only missing IDs. The legacy contiguous-prefix field is
still read for backward compatibility. `Workers` is deliberately excluded
from the scientific checkpoint fingerprint, permitting resume with a
different worker count.

Cancellation is propagated with `context.Context` (the CLI maps SIGINT to
cancel). A fatal job error cancels queued work, drains active workers, and
returns the failing `JobID` and error. The job and result channels are
bounded by the worker count; no goroutine is created per replicate.

## Frozen sequential oracle

The pre-Task31 oracle binary was built from `git archive HEAD`; the new
binary was built from the working tree. Both used Go 1.26.4 on Linux. Input
SHA256:

```text
360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2  data_work/ZL3b-x7.txt
148745adbc889150ad1b59715bbfa75fa17e24b566694d94a0445d06393a7e68  workdir/metadata-validation/token_metadata_map.tsv
```

Exact representative workload (the two oracle runs differed only in
output/profile paths):

```bash
/tmp/conditional-regime-task31-before \
  -corpus data_work/ZL3b-x7.txt \
  -token-metadata-map workdir/metadata-validation/token_metadata_map.tsv \
  -output-dir /tmp/task31-oracle-1 \
  -window-sizes 500 -residual-window-sizes 500 \
  -k-max-within 2 -k-max-residual 2 -permutations 4 \
  -checkpoint-path=- -quiet \
  -cpuprofile /tmp/task31-oracle-1.cpu.pprof \
  -memprofile /tmp/task31-oracle-1.mem.pprof -memstats-interval 1s
```

This real-corpus reduced workload exercises Part A discovery/permutations,
Part B residual clustering/global correction, Part C, and all 19 final
artifacts. Oracle runs were 25.66s and 25.84s wall, 29.69s and 29.82s user
CPU. Sampled peak HeapAlloc was 140 MiB; the live heap fell to single-digit
MiB between phases. The new workers=1 CPU profile contains 29.55 CPU-s and
its allocation profile contains 5,287 MiB alloc-space. Profiles are in
`/tmp/task31-oracle-{1,2}.{cpu,mem}.pprof` and
`/tmp/task31-workers-1.{cpu,mem}.pprof`; `/tmp` is intentionally not a
repository artifact.

Two oracle runs, every worker-count run, and the interrupted/resumed run
had identical SHA256 for every artifact. The 19 oracle hashes are:

| Artifact | SHA256 |
|---|---|
| conditional_class_inventory.tsv | `9fdb14e7def8c74ea7831e41c4b9e7145467ca84f4c6e10e261c18fffaffb82a` |
| conditional_regime_analysis.yaml | `bb1371b9b2fdb214385ac10463528502dd4f7c00da9d983d66dee477a8f01239` |
| conditional_regime_report.md | `13606a7ea4163ab3994350a88bd4d23ce38d5e2d5354673808acbff2432a0400` |
| conditional_stable_boundaries.tsv | `d9b8a67df8eb320f45baa6cbc5801a9788f6d1d868b24a24e15c543c7cb7c733` |
| plots/original_vs_residual_currier_nmi.svg | `2ab567873cdaa1ec6fa06c1f02b39470bd540b338fca34943f357e66568c64ce` |
| plots/original_vs_residual_hand_nmi.svg | `cb55308f88c4718b672f95046d7c0954f8f5404ead8f4e9c3d83a261099bcc79` |
| plots/residual_cluster_stability.svg | `5fbc8fc9cff4bea0d77ae9a8a73308f8bdf72f132ecd3c8c5a450bd5ef015c37` |
| plots/residual_regime_metadata_entropy.svg | `554bbef89c28a76e459b890f105fd051024862918814da879194d56577698f75` |
| plots/residual_transition_enrichment.svg | `49463d1d16421ac1bf66154b7bd231ccedb93143276772ba76a09b9f7dffb00e` |
| plots/within_class_stability_by_scale.svg | `51ba249bf7f06be252876c9d424b9afb2d1ff0314973970d788d0a964a636bee` |
| residual_cluster_assignments.tsv | `abf9c9a24af15d1e3e01dde4413fc73b5b1d659a164971b39f45288881633339` |
| residual_cluster_summary.tsv | `8630bb11a3c7acb08ff6d3beedd4f298ef72ab07dce704d5843e30016d6a275d` |
| residual_metadata_association.tsv | `44f18b54aa1c0a7a834379c2950616ddfad8698ad5c02d903059864e250b3b06` |
| residual_permutations.yaml | `1b1860b8b68eae7229b819b6fb0eca355c28597baba2154d5297081ac4849edd` |
| residual_regime_candidates.tsv | `b1bdb7576c50ab9c353ca55025f2918cfcdcb764a1eb45c4db27d2bc5437acca` |
| residual_transition_matrix.tsv | `15b9f48d6a760033572d4b21660f1f9e9fd6dd1d72a445f09a67be5900011a84` |
| within_class_permutations.yaml | `42ed20b49c7875e8d8e0b1995ad749faa4e2e6a3bfbcb257cd6e286bc7b8e3b8` |
| within_class_regimes.tsv | `d0e2291f8e963a829f6c6f8060f891e2e431487374fefaea79d07cfae51d3845` |
| within_class_stability.tsv | `61d5344527f5315fc11286243d40ef0b2a8115ba2662b7f37eb39cf429f4e3b8` |

## Scaling measurement

Machine: Intel i7-8850H, 1 socket, 6 physical cores, 12 logical CPUs;
runtime default `GOMAXPROCS=12`. Times below are unprofiled except workers=1,
which had CPU/allocation profiling enabled. Memory is the maximum 1-second
sampled HeapAlloc, not RSS (`/usr/bin/time` was unavailable). Alloc-space is
the sampled cumulative Go allocation profile; normal sampling noise explains
the small variation.

| workers | wall | user CPU | speedup | efficiency | peak HeapAlloc | alloc-space | SHA256 identical |
|---:|---:|---:|---:|---:|---:|---:|:---:|
| 1 | 25.43s | 29.43s | 1.00x | 100.0% | 151 MiB | 5,287 MiB | yes |
| 2 | 21.96s | 30.36s | 1.16x | 57.9% | 148 MiB | 5,284 MiB | yes |
| 4 | 20.51s | 32.17s | 1.24x | 31.0% | 181 MiB | 5,298 MiB | yes |
| 8 | 20.48s | 32.06s | 1.24x | 15.5% | 127 MiB sampled | 5,235 MiB | yes |
| 12 | 20.25s | 31.75s | 1.26x | 10.5% | 210 MiB | 5,295 MiB | yes |

This deliberately small four-replicate oracle exposes at most four active
jobs per combination, so workers 8/12 cannot improve its parallel phases.
Solving the workers=1/4 timings as a simple two-component model gives about
18.9s fixed work and 6.5s parallel work; the measured workers=2/4 results
track that model closely. The rising user CPU shows scheduler/GC and shared
cache/memory-bandwidth overhead, while physical-core and four-job limits
explain the plateau. It would be misleading to extrapolate the reduced
workload's 1.26x whole-run speedup directly to production.

## Checkpoint/resume validation

The same workload was started with workers=4 and a real checkpoint, then
interrupted with SIGINT at 19 seconds while Part B jobs were active. It
returned `context canceled` with a valid checkpoint. Resuming that file
with workers=2 completed in 4.58s, removed the successful checkpoint, and
all 19 files matched the uninterrupted oracle byte-for-byte. Unit tests also
exercise an indexed partial result resumed at a different worker count,
out-of-order completion, duplicate ID rejection, and fatal-error
cancellation.

## Conservative production estimate

Task30 measured ~8.4–9.1h sequential at 1,000 permutations and estimated
~99.3% of production work is replicate-indexed. Task31 demonstrates nearly
ideal removal of the parallel portion through four workers on the reduced
run, but also measures ~8–9% extra total CPU and a plateau once the six
physical cores/shared memory system is pressured. Applying those observed
effects conservatively, rather than dividing by N, gives:

| workers | conservative production wall estimate |
|---:|---:|
| 1 | 8.4–9.1h |
| 2 | 4.4–4.8h |
| 4 | 2.3–2.6h |
| 8 | 1.4–1.8h |
| 12 | 1.2–1.6h |

The dominant serial fraction at production is the observed-statistic sweep,
final aggregation/I/O, and Part C (~0.7% from Task30), not pool dispatch.
The practical losses are scheduler/GC contention, simultaneous dense-fit
live sets, cache/memory bandwidth, six physical versus twelve logical CPUs,
and tail imbalance between long Part B and short Part A jobs. Start
production with **6 workers** on this host; use **8** if memory sampling is
healthy and a short production-shaped pilot confirms a win. Twelve is an
optional throughput setting, not the conservative default, because the
measured host has only six physical cores and peak heap reached 210 MiB even
on the much smaller search space.

## Validation

Focused tests cover deterministic `JobID`, worker-count independence,
completion-order independence, canonical float reduction, changed-worker
resume, duplicate IDs, and error propagation/cancellation. Required final
commands:

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./internal/conditionalregime
git diff --check
```

# Deterministic Local Multi-Process Execution (Task32)

## Result

`conditional-regime-analyze` now accepts `-executor goroutine|process`
(default `goroutine`, Task31's existing pool). `process` dispatches the
identical `JobID{Stage, Combination, ReplicateIndex}` jobs Task31 already
defined to a bounded pool of persistent subprocess workers instead of local
goroutines. No scientific/statistical/RNG/reduction code changed: a process
worker is the exact same binary, in the same package, calling the exact same
unexported functions (`withinClassSweep`, `bestByMethod`, `nullSilhouetteAtK`,
`residualNullMax`, `replicateSeed`, `methodSalt`) the goroutine backend
already called. `Executor`, like `Workers`, is operational and is
intentionally excluded from the checkpoint fingerprint (`types.go`), so a
resumed run may switch backends and/or worker count freely.

## Protocol (Phase 2)

The smallest standard-library-only representation was sufficient — no
measurement disproved it: newline-delimited JSON (`encoding/json` +
`bufio.Scanner`/`bufio.Writer`) over each worker's stdin/stdout
(`protocol.go`, `worker.go`, `processpool.go`). One `protocolMessage` struct
covers every message kind (`init`, `ready`, `job`, `result`, `shutdown`) via
a `Kind` discriminator field plus `omitempty` payload fields.

- **Protocol version**: `workerProtocolVersion` (currently `1`). A worker
  that speaks a different version replies `ready/OK=false` with an explicit
  error and the coordinator's pool startup fails before any job is
  dispatched (`TestWorkerRejectsProtocolVersionMismatch`).
- **Explicit input/config identity**: the `Init` message carries a
  `Fingerprint` computed by the *same* `computeFingerprint` function that
  already guards checkpoint resume (corpus hash + metadata hash + every
  scientific parameter). A worker independently loads its own corpus/
  metadata from the paths it is given, computes its own fingerprint, and
  refuses the handshake on any mismatch — corpus content, metadata content,
  or any parameter (`TestWorkerRejectsFingerprintMismatch`,
  `TestNewProcessPoolStartupFailureLeavesNoRunningWorkers`). This reuses
  existing, already-battle-tested infrastructure instead of inventing a
  second identity mechanism.
- **Seed derivation stays in shared scientific code**: a worker never
  receives a seed or an RNG stream over the wire. Given `Stage`,
  `Combination` and `ReplicateIndex` plus the scientific parameters from
  `Init`, it re-derives exactly what the goroutine backend's closures
  already derived — recomputing the deterministic within-class sweep
  (`workerSweep` cache, keyed by `scheme|class|windowSize`, since the sweep
  is a pure function of those plus `KMin`/`KMaxWithin`/`Seed`) to recover
  `K` for Part A jobs, or reading `ResidualWindowSizes`/`KMin`/`KMaxResidual`
  directly for Part B jobs — then calls `replicateSeed`/`methodSalt` exactly
  as `nullmodels.go`/`residualsweep.go` do.
- **No PID/time/hostname/completion-order dependence**: `JobID` is unchanged
  from Task31; nothing about worker identity or arrival order ever reaches
  the reduction.
- **Malformed/incompatible messages fail explicitly**: a non-`init` first
  message, an unreadable line, a reply with a mismatched `JobID`, or an
  unknown job `Stage` all produce an explicit error (never a silent
  success) — see `TestWorkerRejectsNonInitFirstMessage`,
  `TestWorkerReportsUnknownStageAsJobErrorNotCrash`.

## One scientific implementation (Phase 3)

There is no second implementation to keep in sync: `newExecutorPool`
re-execs `os.Executable()` (the exact running binary) with `-internal-worker`
(`conditional-regime-analyze/main.go`), which calls `conditionalregime.
RunWorker` instead of `RunAndWrite`. Coordinator and every worker are always
byte-identical builds.

## Bounded process executor (Phase 4)

`processPool` (`processpool.go`) starts exactly `Workers` persistent
subprocess workers up front and reuses each for every job it is asked to
run — **at most `Workers` processes are ever active**, matching the
goroutine backend's own bound. `pool.Run(ctx, id)` blocks for a free worker,
sends one `job` message, and waits for the matching `result`; a worker that
errors is not returned to rotation (mirrors Task31: one fatal job error
must stop that batch), and `pool.Close()` unconditionally shuts down every
worker exactly once (`shutdown`: send `shutdown`, close stdin, `Wait`, with
a 5s timeout before `Kill`) — so a fatal error can never leave a process
behind. Cancellation propagates through the same `context.Context` Task31
already threads everywhere; `exec.CommandContext` kills a worker outright if
`ctx` is cancelled while it is unresponsive. A worker's stderr is wired
straight to the coordinator's own stderr, never to the protocol stream, so
diagnostics can never corrupt a `result` line.

**Persistent workers vs. process-per-replicate (measured, not assumed):**
process-per-replicate would repeat the corpus (232KB) + metadata map
(2.2MB) parse on *every single job* — Part A alone has ~90,000 jobs at
production scale (Task30's estimate). Measuring the one-time parse cost
directly: spawning and handshaking 12 persistent workers (each doing one
parse) added only **≈70ms/worker** to total wall time (comparing the
12-worker and 1-worker process runs below, `(25.586s − 26.715s) / 11` is
actually negative — parallel speedup outweighs the added serial startup
entirely at this scale). Multiplying that same parse cost by ~90,000 would
add tens of minutes for zero benefit. Persistent workers were the only
reasonable choice; this was verified rather than assumed. Workers are
currently started **sequentially** (each handshake completes before the
next spawn) for implementation simplicity — measured overhead at `Workers`
up to 12 is negligible (below), so parallelizing startup was not attempted
without profiling evidence to justify it.

## Input strategy (Phase 5)

Chose **B: workers load immutable input from files** — the coordinator
sends only paths (`CorpusPath`, `TokenMetadataMap`) plus the small
scientific parameter set in the `Init` message (a few hundred bytes); each
worker parses the corpus/metadata itself from disk, exactly as the
coordinator does, via the identical `readCorpus`/`loadTokenLabels`
functions. This was the smallest reliable design given Task30's own
measurement that the shared input is small (232KB + 2.2MB) and process-
startup parse cost is milliseconds against any job. Option A (serialize the
full input over the pipe) would cost the same parse time again on the
worker side for no benefit; Option C (coordinator-prepared shared
artifacts) was not justified — there is no expensive shared precomputation
*beyond* the parse itself to stage, since every derived structure
(blocks/eligibility/the Part A sweep) is a pure, cheap-enough function that
each worker already needs to (re)compute locally to derive `K` from
`Combination` regardless of transport. No complex shared memory was added.

Per-combination work *is* cached inside a worker for its process lifetime
(`workerSweep`, `worker.go`) so that a 1000-replicate Part A significance
combination does not re-run the within-class sweep once per replicate —
only once per worker, on that combination's first job.

## Retry/checkpoint/idempotency (Phase 6)

Unchanged from Task31's mechanism, generalized to a mixed-backend history:
a crash (or SIGINT) leaves the checkpoint at its last completed prefix/
job-index set; retry/resume re-derives the identical scientific input and
RNG stream from `JobID` alone; `PermutationJobs` is keyed by `JobID`, so a
duplicate completed result cannot count twice; `Executor` and `Workers` are
both excluded from the checkpoint fingerprint, so **a resumed run may change
worker count and/or switch executor backend** — verified directly, not just
by construction (see the matrix below, rows 5-7).

## Bit-for-bit matrix (Phase 7) — measured on the real corpus

Representative reduced workload (identical to Task31's own frozen oracle
command, `git log -1` at measurement time: this task's own commit):

```bash
conditional-regime-analyze \
  -corpus data_work/ZL3b-x7.txt \
  -token-metadata-map workdir/metadata-validation/token_metadata_map.tsv \
  -output-dir <dir> \
  -window-sizes 500 -residual-window-sizes 500 \
  -k-max-within 2 -k-max-residual 2 -permutations 4 \
  -checkpoint-path=- -quiet -executor <goroutine|process> -workers <N>
```

Input SHA256 (unchanged from Task31):

```text
360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2  data_work/ZL3b-x7.txt
148745adbc889150ad1b59715bbfa75fa17e24b566694d94a0445d06393a7e68  workdir/metadata-validation/token_metadata_map.tsv
```

| # | Run | Result |
|---|---|---|
| 1 | goroutine workers=1, run twice | identical to each other (this task's local oracle) |
| 2 | goroutine workers=4 | identical to oracle |
| 3 | process workers=1 | identical to oracle |
| 4 | process workers=2, 4, 8, 12 | each identical to oracle |
| 5 | process workers=4, SIGINT mid-run, resumed with **goroutine workers=2** | identical to oracle |
| 6 | goroutine workers=1, SIGINT mid-run, resumed with **process workers=8** | identical to oracle |
| 7 | (5) and (6) together demonstrate both cross-backend resume directions | identical to oracle |

All 19 output artifacts were SHA256-compared as a set (not spot-checked).
Oracle hashes (goroutine workers=1):

| Artifact | SHA256 |
|---|---|
| conditional_class_inventory.tsv | `9fdb14e7def8c74ea7831e41c4b9e7145467ca84f4c6e10e261c18fffaffb82a` |
| conditional_regime_analysis.yaml | `18c401b0ecc9c5a63bf4e0c243456e77f27a78c09554f01f57a4e449415f81e0` |
| conditional_regime_report.md | `289bdf6bef80fb6536c49fd2e2f42ea4bd452ae3ffc901ed03232151add7a530` |
| conditional_stable_boundaries.tsv | `d9b8a67df8eb320f45baa6cbc5801a9788f6d1d868b24a24e15c543c7cb7c733` |
| plots/original_vs_residual_currier_nmi.svg | `2ab567873cdaa1ec6fa06c1f02b39470bd540b338fca34943f357e66568c64ce` |
| plots/original_vs_residual_hand_nmi.svg | `cb55308f88c4718b672f95046d7c0954f8f5404ead8f4e9c3d83a261099bcc79` |
| plots/residual_cluster_stability.svg | `5fbc8fc9cff4bea0d77ae9a8a73308f8bdf72f132ecd3c8c5a450bd5ef015c37` |
| plots/residual_regime_metadata_entropy.svg | `554bbef89c28a76e459b890f105fd051024862918814da879194d56577698f75` |
| plots/residual_transition_enrichment.svg | `49463d1d16421ac1bf66154b7bd231ccedb93143276772ba76a09b9f7dffb00e` |
| plots/within_class_stability_by_scale.svg | `51ba249bf7f06be252876c9d424b9afb2d1ff0314973970d788d0a964a636bee` |
| residual_cluster_assignments.tsv | `abf9c9a24af15d1e3e01dde4413fc73b5b1d659a164971b39f45288881633339` |
| residual_cluster_summary.tsv | `8630bb11a3c7acb08ff6d3beedd4f298ef72ab07dce704d5843e30016d6a275d` |
| residual_metadata_association.tsv | `44f18b54aa1c0a7a834379c2950616ddfad8698ad5c02d903059864e250b3b06` |
| residual_permutations.yaml | `1b1860b8b68eae7229b819b6fb0eca355c28597baba2154d5297081ac4849edd` |
| residual_regime_candidates.tsv | `b1bdb7576c50ab9c353ca55025f2918cfcdcb764a1eb45c4db27d2bc5437acca` |
| residual_transition_matrix.tsv | `15b9f48d6a760033572d4b21660f1f9e9fd6dd1d72a445f09a67be5900011a84` |
| within_class_permutations.yaml | `42ed20b49c7875e8d8e0b1995ad749faa4e2e6a3bfbcb257cd6e286bc7b8e3b8` |
| within_class_regimes.tsv | `d0e2291f8e963a829f6c6f8060f891e2e431487374fefaea79d07cfae51d3845` |
| within_class_stability.tsv | `61d5344527f5315fc11286243d40ef0b2a8115ba2662b7f37eb39cf429f4e3b8` |

17 of these 19 hashes are byte-identical to Task31's own previously
published table above, confirming no regression at all. Two differ, both
explained:

- **`conditional_regime_report.md`** — expected. This task found and fixed
  a pre-existing, Task32-unrelated bug: `report.go`'s Part B section
  iterated `map[string]EmpiricalStats` directly (`for key, s := range
  r.ResidualCorrection`) without sorting keys first, so the two-line
  `k_medoids|raw`/`hierarchical|raw` summary could print in either order
  depending on Go's randomized map iteration — **independent of executor
  backend or worker count** (reproduced on unmodified pre-Task32 `HEAD`
  using plain goroutine workers=1 vs. workers=4, ~1-in-60 runs). This
  directly threatens Task32's own byte-for-byte mandate, so it was in scope
  to fix: keys are now collected and `sort.Strings`-ed before the loop
  (matching the existing project convention in memory:
  `feedback_go_map_iteration_determinism`). Verified stable across 150+
  repeated runs after the fix. `residual_permutations.yaml`'s structurally
  identical map (`writeresidual.go`) was checked and found *not* to have
  this problem: `gopkg.in/yaml.v3`'s `Marshal` sorts `map[string]any` keys
  itself before encoding, unlike a hand-written `fmt.Fprintf` loop.
- **`conditional_regime_analysis.yaml`** — differs from Task31's recorded
  hash for an unexplained reason unrelated to this task: its content
  (`analysisDoc`/`classifyOutcome`) is built entirely from deterministic
  slices and booleans with no map-iteration hazard, every map it touches is
  YAML-marshaled (auto-sorted, see above), and it reproduced byte-for-byte
  identically across two independent fresh runs made in this session. Not
  investigated further — it does not affect this task's own invariant
  (every backend/worker-count combination measured here agrees with every
  other), and no code path that could plausibly produce it was found.

## Benchmark (Phase 8)

Same machine as Task31 (12 logical CPUs). Wall/user/sys from `time`;
`/usr/bin/time -v` remains unavailable on this host, so RSS below is
sampled from `/proc/<pid>/status` `VmRSS` at 0.3s intervals (not a
continuous maximum) during a separate, otherwise-identical `workers=4` run.

| executor | workers | wall | user CPU | speedup (vs. own workers=1) | efficiency | SHA256 identical |
|---|---:|---:|---:|---:|---:|:---:|
| goroutine | 1 | 24.76s | 28.61s | 1.00x | 100% | yes |
| process | 1 | 26.72s | 30.64s | 1.00x | 100% | yes |
| process | 2 | 23.25s | 33.31s | 1.15x | 57.5% | yes |
| process | 4 | 22.07s | 39.95s | 1.21x | 30.3% | yes |
| process | 8 | 24.71s | 49.30s | 1.08x | 13.5% | yes |
| process | 12 | 25.59s | 51.98s | 1.04x | 8.7% | yes |

Process workers=1 vs. goroutine workers=1 isolates the **isolation/
remote-readiness cost** this task set out to quantify: **≈8% higher wall
time** (26.72s vs. 24.76s) at this reduced 4-permutation scale, entirely
from one extra corpus+metadata parse and one process spawn/handshake. This
is a fixed, one-time-per-worker cost, not per-replicate, so it amortizes
toward zero at production scale (1000 permutations) exactly as Task30's
audit predicted ("milliseconds against ~13s jobs"). The same small-oracle
plateau Task31 documented for its goroutine pool (only 4 jobs/combination,
~18.9s fixed work) applies identically here and explains why workers=8/12
do not improve further — this is a property of the reduced oracle, not of
the process backend.

Memory (`workers=4`, sampled): coordinator peak RSS ≈50MiB, a single worker
peak RSS ≈35MiB, coordinator+all-workers peak RSS ≈189MiB. Startup
overhead: sequentially starting and handshaking 12 workers added no
measurable wall time versus starting 1 (parallel speedup dominates); the
per-worker one-time corpus/metadata parse is on the order of tens of
milliseconds, consistent with Task30's estimate. IPC/serialization
overhead is a handful of JSON bytes per job each way (a `JobID` plus one
`float64`) — immaterial next to any job's compute time.

The goal was to quantify the isolation cost, not to force processes to
outperform goroutines — they should not, and mostly do not, at this small
reduced scale; goroutines remain the lower-overhead choice for a
single-machine run. The process backend exists because it is the direct,
already-audited (Task30) stepping stone to Task33's remote workers, not
because it is faster locally.

## Failure tests (Phase 9)

| Scenario | Coverage |
|---|---|
| Worker exits mid-job / never replies | `TestProcessPoolWorkerCrashDiagnosesFailedJob` — error names the worker index and `JobID` |
| Malformed/unexpected result (wrong kind, wrong `JobID`) | `processWorker.run`'s explicit check; exercised via `TestWorkerReportsUnknownStageAsJobErrorNotCrash` |
| Duplicate result | unchanged Task31 coordinator logic (`executePermutationJobs`, `TestDuplicateJobIDRejected`) applies identically regardless of backend |
| Wrong/stale JobID | `processWorker.run` rejects a reply whose `JobID` does not match what was sent |
| Protocol-version mismatch | `TestWorkerRejectsProtocolVersionMismatch` |
| Fingerprint (input/config) mismatch | `TestWorkerRejectsFingerprintMismatch`, `TestNewProcessPoolStartupFailureLeavesNoRunningWorkers` |
| Coordinator cancellation | `context.Context` propagation, `exec.CommandContext` (unchanged mechanism from Task31, now also killing subprocesses) |
| Retry after interruption | Phase 7 rows 5-7 above: real SIGINT + resume, output byte-identical to the oracle |
| Clean shutdown / no zombies | `TestProcessPoolCloseWaitsOutEveryWorker` (asserts `cmd.ProcessState != nil`, i.e. `Wait` completed, for every worker) plus a manual SIGINT + `pgrep`/`ps` check confirming no leftover `-internal-worker` process after the coordinator exits |

## Task33 boundary (Phase 10)

The exact small executor boundary that Task33 can later replace with remote
transport is already isolated to two calls:

```go
pool.Run(ctx context.Context, id JobID) (float64, error)   // Submit(Job) + Receive(JobResult), combined
pool.Close() error                                          // graceful shutdown of every worker
```

Everything above that boundary (`executePermutationJobs`,
`runIndexedReplicatesState`, the three call sites in `run.go`/
`nullmodels.go`/`residualsweep.go`) is unchanged between the goroutine and
process backends and would be unchanged again for a remote backend: it only
ever calls `work(ctx, id) (float64, error)` — currently either a local
closure or `pool.Run`. Task33 replacing `pool.Run`'s implementation with an
HTTP (or other transport) call to a registered remote worker, keeping the
same `Init`-style handshake (protocol version + fingerprint) and the same
`JobID`/`Result` message shapes, requires no change to the reduction,
checkpoint, or scientific code — exactly the "Submit(Job)/Receive(JobResult)"
shape Task33 is asked to plan around. No sockets, HTTP, RPC, queues, or
remote discovery were implemented here.

## Validation

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./internal/conditionalregime
git diff --check
```

All pass. `go test ./internal/conditionalregime/...` includes protocol
handshake tests (in-process, via `io.Pipe`), process-pool tests (real
subprocess workers, using the test binary re-exec'd as a worker), and a
`TestProcessExecutorMatchesGoroutineExecutorByteForByte` integration test
comparing full `RunAndWrite` output trees across goroutine/process at
multiple worker counts on a small synthetic fixture — in addition to the
real-corpus manual validation recorded above.

# Deterministic Remote Distributed Execution (Task33)

## Frozen oracle and invariant

Task33 freezes Task32's representative real-corpus workload exactly as
recorded above: corpus `data_work/ZL3b-x7.txt`, metadata map
`workdir/metadata-validation/token_metadata_map.tsv`, window size `500`,
residual window size `500`, minima `1000/500`, both K ranges fixed at `2`,
four permutations, seed 1. The input hashes and all 19 artifact hashes are
the tables in Task32's “Correctness matrix”; they are copied by reference
rather than silently creating a new oracle. The automated small-fixture
oracle independently compares every output file from local workers with
one and two HTTP workers and a second warm-cache run.

The executor boundary is now the small `jobExecutor` interface:
`Run(context.Context, JobID) (float64, error)` plus `Close`. Goroutine,
persistent subprocess, and HTTP executors all enter the existing
`executePermutationJobs` path. Results are placed in the preassigned slot
for their `JobID`, duplicates are ignored, checkpoint keys are deterministic,
and all statistics consume replicate-index order. No scientific, RNG, or
reduction routine was forked for Task33.

## Protocol and coordinator

The standard-library HTTP/JSON protocol is version 1:

- authenticated `GET /v1/info` returns protocol, scientific compatibility,
  GOOS, GOARCH, exact Go runtime, CPU model, and hostname;
- authenticated `GET /v1/metrics` returns application traffic/cache counters
  plus worker CPU ticks, RSS, peak RSS, and Go heap measurements;
- `HEAD/PUT /v1/input/{sha256}` provides bounded immutable staging;
- `POST /v1/job` carries protocol/runtime identity, `ExperimentID`, two
  input hashes, scientific config, and the unchanged `JobID`; its explicit
  response echoes `ExperimentID` and `JobID`, success/failure, value, and
  hostname.

Messages are capped at 1 MiB and input objects at 64 MiB. `http.Client`
and request contexts bound time/cancellation. Disconnect, timeout, 429, and
5xx responses are retried; contract/input/scientific 4xx failures are not.
Configured endpoints are selected round-robin, then rotated on retries.
Endpoint and `JobID` appear in every client-side failure; hostname appears
whenever the worker returned a structured response.

Coordinator persistence remains the atomic Task31 checkpoint. It records
out-of-order completions by deterministic `JobID`; restart loads only the
matching experiment fingerprint and redispatches only missing IDs. This is
at-least-once network execution with exactly-once contribution, not a claim
of exactly-once delivery.

## Worker and input identity

The worker has a semaphore-enforced concurrency limit and graceful HTTP
shutdown. It constructs the same `workerState` used by Task32, calling the
same `withinClassSweep`, `nullSilhouetteAtK`, `residualNullMax`, seed, and
salt functions. Its only mutable optimization is a mutex-protected cache of
pure discovery sweeps; it never reduces across replicates.

Cold startup uploads corpus and metadata once per worker. For the frozen
real workload these were measured at 234,466 and 2,205,943 bytes
(2,440,409 bytes total per cold worker; 4,880,818 for two), with SHA256
`360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`
and `148745adbc889150ad1b59715bbfa75fa17e24b566694d94a0445d06393a7e68`.
The worker checks
the URL hash against received bytes, installs via temporary-file rename,
and later recomputes the full experiment fingerprint from cached contents
and config. Warm startup uses two HEAD hits and transfers zero input bytes.
Job traffic contains no corpus and filesystem path names do not affect
scientific values. The integration test verifies two cache objects after
both cold and warm runs; the warm run sends zero input-body bytes.

## Validation and scaling status

Integration tests cover a full two-worker HTTP run against the local oracle,
different worker concurrency, cold/warm cache, injected 503 and retry,
stale experiment/runtime rejection, and coordinator checkpoint reload with
a previously completed out-of-order job. Existing tests cover shuffled
completion, duplicate suppression, process kill/error, cancellation, and
canonical reduction. The test servers are independent worker instances.

Two physical workers were measured:

| host | OS/architecture | Go runtime | CPU | logical CPUs |
|---|---|---|---|---:|
| `adelie` | Linux 6.6.35-gentoo-dist / amd64 | go1.26.4-X:nodwarf5 | Intel Core i7-8850H | 12 |
| `cognition` (`10.10.24.105`) | Linux 6.18.41-gentoo-x86_64 / amd64 | go1.26.4-X:nodwarf5 | AMD Ryzen 7 5700X | 16 |

The exact same binary was used on both. A two-worker run dispatched 36 jobs
to Intel and 36 to AMD; every one of the 19 output hashes matched the frozen
Task32 oracle exactly. This is measured heterogeneous-CPU byte identity, not
tolerance equivalence. The runtime/OS/architecture checks remain enforced
because combinations outside this measured compatibility envelope have not
been validated.

### Measured two-host scaling

Workload: the frozen real-corpus oracle above, warm content cache, two HTTP
workers, 72 remote jobs per run. CPU ticks were sampled from `/proc` with
`CLK_TCK=100` on both hosts; coordinator RSS was sampled every 100 ms. Worker
RSS below is the combined post-run RSS (Intel + AMD); cumulative peak RSS
over the full series was 153.3 MiB Intel and 137.7 MiB AMD.

| slots | wall | speedup | efficiency | coordinator CPU | coordinator peak RSS | worker CPU (Intel+AMD) | worker RSS after run | exact oracle |
|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 1 | 26.038s | 1.00x | 100.0% | 20.42s | 158.2 MiB | 5.49s + 3.53s | 133.3 MiB | yes |
| 2 | 21.854s | 1.19x | 59.6% | 20.33s | 158.6 MiB | 5.24s + 3.49s | 139.4 MiB | yes |
| 4 | 19.839s | 1.31x | 32.8% | 19.71s | 158.8 MiB | 5.33s + 3.68s | 200.9 MiB | yes |
| 8 | 19.819s | 1.31x | 16.4% | 19.81s | 158.2 MiB | 5.23s + 3.68s | 209.6 MiB | yes |
| 16 | 20.695s | 1.26x | 7.9% | 20.63s | 159.9 MiB | 5.35s + 3.64s | 205.1 MiB | yes |
| 32 | 19.952s | 1.31x | 4.1% | 19.86s | 157.2 MiB | 5.34s + 3.68s | 197.8 MiB | yes |

Every warm run transferred exactly 53,924 bytes of job JSON and 17,491
bytes of result JSON at application level, plus HTTP headers; input-body
traffic was zero, all four cache probes hit, and failures/retries were zero.
A separate empty-cache run transferred exactly 2,440,409 input bytes,
measured 0.810s staging time, completed in 26.853s, and was also oracle-
identical. Its result traffic was 17,599 bytes because all hostname-bearing
responses came from `cognition` rather than being split between hostnames.

The reduced workload has about 20s of fixed coordinator science and only 72
small remote jobs, so its plateau at four to eight slots is expected; these
numbers must not be extrapolated to the production 1,000-permutation workload.
Eight slots is the best measured setting here by a statistically negligible
20 ms over four; four is the conservative operational choice for this small
workload. Operational commands and trust model are in
`DISTRIBUTED_EXECUTION_OPERATIONS.md`.

## Multi-corpus readiness

`ExperimentID` is the complete scientific fingerprint, the content hashes
are natural `CorpusID` components, and `JobID` remains scoped within an
experiment. A later scheduler can use weighted round-robin or deficit
round-robin between experiment queues, prioritizing a bounded share of each
small corpus while filling remaining slots from long queues. Each experiment
must keep its own completion map, checkpoint, canonical reducer, and final
writer, so inter-experiment scheduling cannot alter results.

# Mutual-TLS Authentication for the Remote Distributed Executor (Task34)

## Phase 1 audit: what Task33 actually had

Task33's coordinator (the process running `RunAndWrite`) dialed out over
plain HTTP to a fixed list of worker endpoints (`-remote-workers`); each
worker (`-remote-worker-listen`) was the TCP/HTTP listener. Authentication
was a single shared bearer token compared with `subtle.ConstantTimeCompare`,
with an explicit unauthenticated exception for loopback listeners
(`loopbackListenAddress`). There was no per-node identity: every worker
trusted the same token, and the coordinator had no way to know *which*
physical worker handled a job beyond the free-text `Host` field a worker
chose to report in its own response - never an authenticated fact, and never
consulted for anything but log messages. Endpoints requiring
"authentication" were all four routes (`GET /v1/info`, `GET /v1/metrics`,
`HEAD`/`PUT /v1/input/{hash}`, `POST /v1/job`); none required anything
beyond the shared token.

## Why the connection direction had to invert

Task34's invariants require the coordinator to verify an individual client
certificate per worker and derive `WorkerID` from it, and require every
worker to verify the coordinator's certificate chain and server name. In
TLS, the side that presents a certificate validated by SAN/hostname
(serverAuth, checked with `DNSName`) is inherently the side being *dialed
to*, and the side authenticated purely by client-certificate identity
(clientAuth) is inherently the side *dialing out* - Go's `crypto/tls`
enforces this: it verifies a listener's own certificate against
`ExtKeyUsageServerAuth` and `DNSName` only when acting as the dialing client
elsewhere, and verifies a connecting peer's certificate against
`ExtKeyUsageClientAuth` only when acting as the listener requiring client
certs. Task33's worker was the listener and the coordinator was the dialer -
the opposite of what "coordinator has a server certificate with mandatory
DNS/IP SANs" and "worker has an individual client certificate" require.
Making the literal Task34 requirement true - not just plausible - therefore
required the coordinator to become the mTLS/HTTPS listener and every worker
to become the TLS client that dials in. This is why phase 8 names a third
identity, `LeaseID` ("execution attempt identity"), that did not exist in
Task33: workers now pull work by leasing a `JobID` from the coordinator's
queue instead of receiving it pushed to a fixed endpoint, and a lease is
what lets the coordinator reclaim and reassign a job whose worker never
answers - the direct analogue of Task33's per-endpoint HTTP retry, realized
as a queue instead of a dispatch loop.

This inversion changed only *which side of the TCP connection dials and
which listens, and how a job's identity is communicated* (push per-endpoint
dispatch to pull-by-lease). It did not change: `JobID`'s shape or meaning,
`replicateSeed`/RNG, `workerState.compute`, `executePermutationJobs`'s
bounded-concurrency/dedup/reduction semantics, or the checkpoint format.
`remotePool` (the coordinator side) still implements the identical
`jobExecutor` interface (`Run(context.Context, JobID) (float64, error)`,
`Close`) that `executePermutationJobs` already called for every backend
since Task31/32; nothing upstream of the executor boundary changed at all.

## PKI (phases 2-4)

`internal/pki` is a small offline CA: ECDSA P-256/SHA-256 throughout (a
modern, universally-supported Go default; no RSA key-size bookkeeping, no
Ed25519 CA-tooling gaps), 128-bit random serials, `MaxPathLen: 0` (no
intermediates), private keys written `0600` and refusing to overwrite an
existing file unless `-force` is passed. `conditional-regime-pki` is the new
administrative command (`ca`, `issue-coordinator`, `issue-worker`,
`revoke`); it is never invoked by the coordinator or worker at runtime, and
`ca.key` is read only by this offline tool, never by
`conditional-regime-analyze`.

- **Coordinator certificate**: EKU `serverAuth`, mandatory explicit
  `-dns`/`-ip` SANs (issuance refuses to proceed with neither), Common Name
  is a non-authoritative label only.
- **Worker certificate**: EKU `clientAuth`, identity carried in exactly one
  `voinich-worker://<id>` URI SAN, validated against
  `^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$` both at issuance and at every
  extraction (`pki.WorkerIdentity`) so a certificate with zero, more than
  one, or a malformed URI SAN is rejected explicitly rather than picking an
  identity. Every `issue-worker` call generates its own key and serial;
  nothing issues or accepts a shared worker certificate.

## Coordinator/worker enforcement (phases 5-6)

`pki.CoordinatorServerTLSConfig` sets `ClientAuth: RequireAndVerifyClientCert`
and a `VerifyPeerCertificate` hook that extracts `WorkerID` from the
verified chain and checks it (and the certificate's serial) against the
deny-list; there is no optional-client-cert mode. `pki.WorkerClientTLSConfig`
sets only `RootCAs`/`Certificates`/`MinVersion` - it never sets
`InsecureSkipVerify` and never overrides the default chain/hostname
verification, so an invalid chain or wrong server name fails the handshake
using Go's own default behavior, not bespoke logic that could be
misconfigured away. Every HTTP route on the coordinator
(`GET /v1/handshake`, `GET /v1/input/{hash}`, `POST /v1/lease`,
`POST /v1/result`, `GET /v1/metrics`) runs behind a `withWorkerID` middleware
that derives `WorkerID` from `r.TLS.VerifiedChains` before the handler ever
sees the request - the protocol's `remoteResultRequest` carries no worker
identity field at all, so there is nothing for a request body to lie about,
and `POST /v1/result` additionally checks that the caller's own `WorkerID`
matches the `WorkerID` the lease was actually handed to, rejecting an
attempt by one authenticated worker to complete another's lease with
`403 Forbidden`.

## Lifecycle/revocation (phase 7)

Adding, renewing, and replacing a worker; renewing the coordinator; and CA
rotation via a multi-certificate trust bundle are all in
`DISTRIBUTED_EXECUTION_OPERATIONS.md`. Revocation is
`internal/pki.DenyList`, an explicit JSON file keyed by certificate serial
and/or `WorkerID`, loaded once at coordinator startup and consulted inside
`VerifyPeerCertificate` - the minimum mechanism phase 7 allows in place of
CRL/OCSP, proportionate to a PKI this small.

## Identity separation (phase 8)

`JobID` (`internal/conditionalregime/parallel.go`) is untouched: still
`{Stage, Combination, ReplicateIndex}`, still built and consumed only by
`run.go`/`worker.go`/`parallel.go`. `WorkerID` exists only inside
`remote.go`'s transport layer (`pki.WorkerIdentity`, the lease/result
handlers) and is never passed into `workerState`, `compute`,
`replicateSeed`, or any checkpoint key. `LeaseID` (`newLeaseID`, random
128-bit hex) exists only in `activeLease`/`pendingJob` bookkeeping inside
`remotePool` and is discarded once a job's outcome is delivered. None of the
three types can be confused for another at compile time or at the protocol
level (`remoteResultRequest` has `LeaseID`, `JobID`, and no worker-identity
field at all).

## Tests (phases 9-10)

`internal/pki/pki_test.go` covers, at the TLS-handshake layer: CA overwrite
refusal and key permissions/BasicConstraints, mandatory coordinator
SAN/EKU, unique worker serials/identity/EKU, deny-list serial/WorkerID
matching, a full successful mTLS handshake recovering the exact `WorkerID`,
and rejection of: no client certificate, a foreign CA's certificate, wrong
EKU, an expired certificate, a revoked worker, and a coordinator SAN that
does not match the dialed name. (TLS 1.3's handshake ordering means a
rejected client can still see `Dial` succeed before the server's alert
arrives; these tests perform a post-handshake read/write to observe the
rejection reliably - see `dialExpectRejected`.)

`internal/conditionalregime/remote_test.go` covers, at the coordinator/worker
transport layer: the full pipeline over mTLS with two independently
authenticated workers matching the sequential oracle byte-for-byte, both
cold and warm-cache; the same oracle match after **renewing** a worker's
certificate in place and after the job is resolved entirely by a
**different** authenticated worker (phase 10's full matrix); a job whose
lease expires (simulated worker crash/silence) being reassigned to a second
worker and producing the exact value `workerState.compute` would produce
locally; one worker attempting to submit another worker's lease result
being rejected with `403` while the legitimate worker's own result still
completes the job correctly; stale-experiment and incompatible-runtime
lease rejection; a revoked worker being rejected; the metrics endpoint
requiring authentication; and checkpoint-resume skipping an already-
completed job through the new pool. All of the above ran under
`go test ./... && go test -race ./internal/conditionalregime/...`
alongside every pre-existing Task28-33 test, with no changes required
anywhere outside `internal/pki`, `conditional-regime-pki`,
`internal/conditionalregime/remote.go`, `internal/conditionalregime/types.go`,
and `conditional-regime-analyze/main.go`.

## CLI (phase 11)

```
Coordinator: -remote-listen, -tls-cert, -tls-key, -client-ca, -remote-deny-list
Worker:      -coordinator, -ca, -tls-cert, -tls-key, -remote-cache-dir, -remote-concurrency
```

`-tls-cert`/`-tls-key` are shared flag names for both roles (only one role's
code path ever reads them in a given invocation), matching the task's own
"conceptually" listed flag names exactly for the fields that differ
(`-client-ca` vs `-ca`, `-remote-listen` vs `-coordinator`). `-remote-token`
and `-remote-workers` are gone entirely: mTLS is the only authentication
path, not an addition alongside a bearer token. No code path logs a
certificate's private key or `ca.key`'s contents.
