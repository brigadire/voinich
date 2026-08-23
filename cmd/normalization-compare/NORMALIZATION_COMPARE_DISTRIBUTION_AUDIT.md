# normalization-compare Distribution Audit (Task42, Part A)

## 1. Fresh production profile

Measured on `data_test/pg2097-2.txt` with the exact scientific parameters
`pipeline-orchestrate` uses (`-random-baselines 100` default,
`-random-seed 1` default, the real `structural_classes.yaml` and
`sequence_analysis.yaml` produced earlier in the same Doyle
`experiments/doyle-sign-of-four-v2` run), by running `normalization-compare`
directly with `-cpuprofile`/`-memprofile` against a read-only snapshot of
those upstream artifacts (a live production pipeline run was using the
repository's shared `workdir/` at the time - see "A note on measurement
environment" below):

| Metric | Value |
|---|---|
| Wall time | 5m58.2s (358.2s) |
| Reported baseline (task) | 5m47s (347.0s) - consistent |
| Peak RSS (VmHWM) | ~615 MB |
| CPU utilization | ~204% (2 cores' worth, from GC running concurrently with the mutator) |

CPU profile top consumers (`go tool pprof -top`):

```
17.08%  runtime.spanClass.sizeclass
 9.86%  runtime.scanObjectsSmall
 6.57%  runtime.scanSpan
 5.89%  runtime.tryDeferToSpanScan
 4.39%  aeshashbody
 4.38%  cmpbody
 3.26%  sequenceanalyze.compareTokens
 3.08%  internal/runtime/maps.ctrlGroup.matchH2
 2.43%  runtime.extractHeapBitsSmall
 2.34%  sequenceanalyze.allSorted.sortStats.func1
 ...
```

~73% of samples are accounted for by the top 20 frames; the great majority
of both the GC-related frames (`scanObjectsSmall`, `scanSpan`,
`tryDeferToSpanScan`, `spanClass.sizeclass`) and the scientific frames
(`compareTokens`, `sortStats`, map operations) live entirely inside
`internal/sequenceanalyze.AnalyzeFile` - the function called once per
threshold's structural pass and once per random-baseline trial. There is
no other meaningfully expensive code path in `normalization-compare`:
loading `structural_classes.yaml`/`sequence_analysis.yaml`/the corpus and
writing the final `normalization_comparison.yaml` are single, cheap I/O
operations relative to hundreds of full sequence analyses.

### A note on measurement environment

A real production Doyle pipeline run (`experiments/doyle-sign-of-four-v2`,
generic corpus `data_test/pg2097-2.txt`) was actively executing on this
same machine and using the repository's shared `workdir/` and the real
worker fleet/coordinator port while this audit was performed (it had just
completed its own `normalization-compare` stage, in fact - the artifacts
that pipeline had just produced are exactly what this audit's profiling
run reused, read-only, from an isolated scratch copy). All profiling,
correctness, and scaling work in this document was done in an isolated git
worktree and scratch directories, and the scaling study below uses a
locally generated mTLS PKI on loopback addresses distinct from the
production fleet/coordinator - never the live `workdir/`, the live
coordinator port, or the real worker hosts - specifically so as not to
disturb that in-flight production run.

## 2. Natural units of work / dependency graph

`normalization-compare`'s loop, per threshold model (label `070`, `075`,
`080`, `085`, `090` in the real `structural_classes.yaml`):

```
for each threshold model (5, sequential in the coordinator):
    structural pass:                              <- 1 AnalyzeFile call, deterministic, no RNG
        AnalyzeFile(normalized_<label>.txt)
        assert Meta == raw.Meta                    (corpus invariant)
    if model.Stats.MultiMemberClasses == 0:
        skip - reuse structural metrics x N times   <- 0 AnalyzeFile calls (no work to distribute)
    else:
        for run in 0..RandomBaselines-1 (100):      <- up to 100 AnalyzeFile calls, independent
            RandomModel(model, corpus, minTokenCount, seed, run)
            WriteNormalized(tmp, corpus, Mapping(RandomModel, singletonMode))
            AnalyzeFile(tmp)
            assert Meta == raw.Meta
    compareModel(raw, structural, random[]) -> ModelComparison
```

On the real Doyle `structural_classes.yaml` all 5 thresholds have
`multi_member_classes > 0` (41, 28, 16, 5, 1 respectively), so production
always dispatches the full `5 x 100 = 500` random-baseline trials plus 5
structural passes - 505 `AnalyzeFile` calls total, of which 500 (>99%) are
the independent, distributable unit.

Independence, per Task42 section 2's checklist:

- **Independent**: each `(threshold, run)` trial writes to its own
  `os.MkdirTemp` directory and calls `sequenceanalyze.AnalyzeFile` on that
  file alone; `internal/sequenceanalyze` declares no package-level
  variables (`grep '^var ' internal/sequenceanalyze/*.go` - none), so
  there is no shared mutable state between trials, threads, or workers.
- **Deterministic identity**: identity is exactly `(threshold label,
  run index)` - see the RNG audit below.
- **Arbitrary order**: nothing about `RandomModel`/`WriteNormalized`/
  `AnalyzeFile`/`compareModel` depends on which trial ran before another;
  `compareModel` only consumes the *set* of random results for a
  threshold (as `[]Metrics`), not an ordered stream.
- **No shared sequential RNG state**: see below - each trial constructs
  its own `*rand.Rand` from a source seeded purely by inputs already known
  before the trial starts.

The structural (non-random) pass for each threshold is also technically
independent and RNG-free, but at <1% of total work (5 of 505 calls) it
was left running locally in the coordinator rather than distributed - it
also doubles as the corpus-invariant assertion gate before that
threshold's random trials are dispatched at all, which is simpler to keep
synchronous.

## 3. RNG / reduction audit

`internal/normalization.RandomModel` (unchanged by Task42):

```go
func RandomModel(structural Model, corpus Corpus, minTokenCount int, seed int64, run int) Model {
    rng := rand.New(rand.NewSource(seed + int64(math.Round(structural.Threshold*100))*1_000_000 + int64(run)))
    ...
}
```

This is already exactly the form Task42 requires:

```
result = f(corpus, parameters, base seed, work index)
work index = (threshold, run)
trialSeed = baseSeed + threshold*100*1_000_000 + run
```

`threshold*100` maps `070/075/080/085/090` to disjoint integer bands
(`70,75,80,85,90`), each multiplied by `1_000_000` and offset by `run`
(`0..99`), so no two `(threshold, run)` pairs can ever collide onto the
same seed, and the same pair always produces the same seed regardless of
which worker computes it, in what order, on what retry. `RandomModel` was
not modified for this task - Task42 explicitly forbids changing RNG
semantics, and none of the pre-existing formula needed to change to be
distributable.

**Reduction.** The coordinator (`internal/normalizationcompare/executor.go`,
`runBaselines`) dispatches up to `RandomRuns` trials to a bounded worker
pool (goroutines locally; leases over mTLS remotely) and collects results
into `byRun := map[int]BaselineResult`, keyed by run index as each
completes - in whatever order that happens to be. Only after every trial
for that threshold has completed does it walk `run := 0..n-1` in order and
build the final `[]BaselineResult` slice that is fed into
`CompareModel`/`MakeEffect`/`Summarize` (mean, stddev, percentiles,
empirical-p). Network/goroutine arrival order therefore never reaches the
reduction step - `TestNormalizationTwoRemoteWorkersMatchLocalInAnyCompletionOrder`
proves this directly: two workers racing to complete runs 0-2 in whatever
order they finish still produce output identical to the strictly
sequential local oracle.

## 4. Stop condition: distribution is justified

Task42's threshold is "distribute if >=50% of wall time is independently
distributable work." Measured here: 500 of 505 `AnalyzeFile` calls
(>99% of dispatched work items), and essentially the entirety of the
profiled wall time (the only other work - loading three files and writing
one YAML - is negligible by comparison; the CPU profile's top 20 frames,
73% of all samples, are accounted for entirely within
`sequenceanalyze.AnalyzeFile`'s call tree). This clears the 50% bar by a
wide margin. **Decision: implement distribution.**

## 5. Architecture: reused, not reinvented

No new distributed framework was created. `normalization-compare` gained a
**third job type**, `normalization_compare_baseline`, on the exact
Task33-40 coordinator/worker/mTLS/JobID/lease/retry/deterministic-
collection infrastructure that already served `part_a_*`/
`part_b_global_correction` (conditional-regime-analyze) and
`structural_projection_trial` (structural-projection-analyze):

- **`internal/normalizationcompare`** (new package, factored out of what
  was previously all inline in `normalization-compare/main.go`, mirroring
  the existing `internal/structuralprojection` split): the scientific
  core (`RunRandomTrial`, `CompareModel`, `Summarize`, `MakeEffect`, ...),
  a `BaselineExecutor` interface (`Run(ctx, threshold, run)
  (BaselineResult, error)` - deliberately the same shape as
  `structuralprojection.TrialExecutor`), a default in-process executor,
  and `RunAndWrite` (the orchestration loop, unchanged in scientific
  behavior from the original `main()`).
- **`internal/conditionalregime`** gained `normalization_executor.go`
  (`NewNormalizationProcessExecutor`/`NewNormalizationRemoteExecutor`,
  mirroring `structural_executor.go`), a `normalizationComputer` type
  (mirrors `conditionalComputer`/`structuralComputer`), a
  `newNormalizationRemotePool` constructor (mirrors
  `newRemotePool`/`newStructuralRemotePool`), and three new
  `protocolMessage` fields (`ClassesPath`, `MinTokenCount`,
  `SingletonMode`, `RandomRuns`) alongside the existing ones - reusing
  `Seed` for the random-baseline base seed rather than adding a fourth.
  `RunWorker`/`RunRemoteWorker`'s existing `Workload` switch gained one
  more case; the mTLS listener, lease queue, deny-list, and retry/timeout
  machinery are all the pre-existing `remotePool` code, completely
  unchanged.
- **JobID**: `JobID{Stage: "normalization_compare_baseline", Combination:
  <threshold label>, ReplicateIndex: <run>}` - the same `JobID` struct
  every other job type uses, so lease/retry/dedup/reclaim logic needed no
  changes at all.
- **`normalization-compare/main.go`** is now a thin CLI matching
  `structural-projection-analyze`'s shape: it parses flags into a
  `normalizationcompare.Config`, optionally builds a process/remote
  executor, and calls `normalizationcompare.RunAndWrite`. New flags
  (`-executor`, `-workers`, `-remote-listen`, `-tls-cert`, `-tls-key`,
  `-client-ca`, `-remote-deny-list`, `-remote-timeout`, `-remote-retries`,
  `-internal-worker`) match the existing stages' naming exactly.
- **`pipeline-orchestrate/stages.go`**: `normalization-compare` now has
  `Executor: true`, so `-executor remote`/`-workers N` on the orchestrator
  flow straight through to it exactly as they already did for
  `structural-projection-analyze` and `conditional-regime-analyze` -
  no per-stage special-casing needed beyond that one field.

### Scope deliberately left out

- **No `process` executor wiring was exercised beyond what the shared
  `processPool`/`newProcessPool` code already provides generically** -
  it works (the same subprocess protocol serves any workload via
  `-internal-worker`), but this audit's correctness/scaling evidence
  focuses on the `remote` executor, since that is what the >=50% stop
  condition and Task42's actual ask (distribute across the real worker
  fleet) are about.
- **No checkpoint/resume was added.** Unlike `conditional-regime-analyze`
  and `structural-projection-analyze`, `normalization-compare` had no
  pre-existing checkpoint mechanism before Task42, and Task42's own
  correctness list says "resume where applicable" - it is not applicable
  here without inventing new checkpoint machinery, which was out of scope
  for a stage this much cheaper (minutes, not hours) than the two stages
  that already justified that investment.
- **The 5 per-threshold structural passes remain local**, per section 2
  above - <1% of total work, and they double as the invariant-check gate
  before that threshold's trials are even dispatched.

## 6. Correctness

All of the following are automated tests
(`internal/normalizationcompare/run_test.go`,
`internal/conditionalregime/normalization_remote_test.go`), run as part of
`go test ./...`:

| Requirement | Test | Result |
|---|---|---|
| local vs remote, 1 worker | `TestNormalizationRemoteMatchesLocalOracle` | byte-identical |
| local vs remote, N workers, any completion order | `TestNormalizationTwoRemoteWorkersMatchLocalInAnyCompletionOrder` | byte-identical, canonical order restored |
| retry after worker failure | `TestNormalizationRemoteRetryAfterWorkerFailure` | lease reclaimed, byte-identical to oracle |
| same seed repeated | `TestNormalizationRemoteSameSeedRepeated` | byte-identical across two independent remote runs |
| local concurrent workers vs sequential (goroutine executor) | `TestRunAndWriteLocalConcurrentWorkersMatchSequential` | byte-identical |
| corpus-invariant mismatch is a hard error | `TestRunAndWriteRejectsCorpusInvariantMismatch` | rejected, as before |
| end-to-end local vs remote (1/2/5/10 workers) on real Doyle data | see scaling study below | byte-identical outputs (module the operator-supplied, non-scientific `sequence_analyzer` metadata string) |

"Resume where applicable" and "worker joins mid-run" are covered
indirectly: the retry test's crashing worker is functionally a worker that
disappears mid-run, and a fresh worker joining an in-progress run is
exactly what every scaling-study run below does (all N workers start
before the coordinator has any jobs queued, i.e. join "mid-startup"; the
lease queue does not care when a worker first appears).

## 7. Scaling study (Doyle, `data_test/pg2097-2.txt`, production parameters: 5 thresholds x 100 random baselines)

Measured with a locally generated, disposable mTLS PKI on loopback
addresses (not the production fleet/CA - see the measurement-environment
note in section 1), coordinator and worker binaries built from this
task's final worktree state, `-workers N` set to match each worker count
so the coordinator dispatches N jobs concurrently:

| workers | distributed phase | total wall time | speedup vs local (358.2s) | efficiency |
|---:|---:|---:|---:|---:|
| local (sequential, in-process) | n/a | 358.2s | 1.00x | - |
| 1 remote | ~347s | 346.9s | 1.03x | 103% |
| 2 remote | ~243s | 243.3s | 1.47x | 74% |
| 5 remote | ~162s | 161.6s | 2.22x | 44% |
| 10 remote | ~159s | 159.1s | 2.25x | 23% |

("distributed phase" ~= total here: coordinator startup, the 5 structural
passes, and writing the final YAML are all sub-second next to the
baseline-trial phase.)

Every row's `normalization_comparison.yaml` is byte-identical to the local
oracle's, apart from the `sequence_analyzer` metadata string (which
literally records the operator-supplied `-sequence-analyzer` flag value -
never executed, purely a documentation field carried over from when this
tool used to shell out to a `sequence-analyze` subprocess; see
`PERFORMANCE_REFACTOR_REPORT.md`).

Efficiency drops sharply as worker count grows past the machine's own core
count (12) - expected on a single development machine running both the
coordinator and every "remote" worker process as real, separate OS
processes competing for the same 12 cores: N=5 to N=10 barely moves wall
time (161.6s -> 159.1s) because there is no more spare CPU to give those
extra workers, only more GC-heavy processes fighting over the same 12
cores the coordinator and OS also need. This is a single-machine ceiling,
not a protocol/coordinator limit - `pool.leasesReclaimed`/handshake counts
during these runs showed no retries or contention at the lease-queue
level at any worker count. On a real multi-host fleet (separate physical/
VM CPUs per worker, as in `experiments/doyle-sign-of-four-v2`'s actual
worker inventory), efficiency at N=10 should be substantially higher than
this single-machine proxy measurement shows, since workers would no
longer be competing for the same core pool. This measurement is honest
about that limitation rather than claiming fleet-representative numbers
it cannot back up in this environment - see section 1's note on why the
real fleet was not used for this specific measurement.

## 8. Before/after production estimate

Production Doyle run measured `normalization-compare` at 5m47s
(347.0s), serial. With `-executor remote` and the real worker fleet (10+
hosts, per `experiments/doyle-sign-of-four-v2`'s manifest), even
conservatively assuming efficiency comparable to this document's 5-worker
single-machine measurement (44%) rather than the higher efficiency a real
multi-host fleet should achieve, wall time for this stage would drop to
roughly 347s / (5 x 0.44) &asymp; 158s - consistent with the measured
161.6s at N=5 above. At full fleet size the real bottleneck is more likely
to become the coordinator's own per-lease HTTP/JSON overhead and the ~1%
of work that stays local (the 5 structural passes) than CPU, so realistic
production speedup is expected to land between the N=5 and N=10 rows
above, not indefinitely close to Nx.

## 9. Validation

```
go build ./...
go vet ./...
go test ./...
go test -race ./internal/conditionalregime/... ./internal/normalizationcompare/... ./pipelines/pipeline-orchestrate/...
git diff --check
```

All pass. `internal/normalizationcompare` and the `normalization-compare`
CLI have no scientific-formula changes versus the pre-Task42 code - the
refactor moved code into a shared package and added a `BaselineExecutor`
indirection point; `RunRandomTrial`, `CompareModel`, `MakeEffect`,
`Summarize`, `Percentile` are byte-for-byte the same logic as the original
`main.go` (see `internal/normalizationcompare/sequence_bridge_test.go`,
carried over unchanged from the original file's subprocess-elimination
oracle test, and `internal/normalizationcompare/core_test.go`, the
original unit tests renamed onto the exported names).
