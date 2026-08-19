# Distributed Execution for begin-end-analyze (Task47)

Stage 5 (`begin-end-analyze`) took `14m52s` on the real generic corpus
`experiments/astafiev-1000-culinar-receipt-v1` (`data_test/astafiev-1000-culinar-receipts.txt`,
production defaults: `-max-window 55 -permutations 100 -min-frequency 10
-max-candidates 1000`), already a significant share of that pipeline's
total wall time. Task47 puts it on the exact Task28-44 mTLS
coordinator/worker executor already serving `conditional-regime-analyze`,
`structural-projection-analyze`, `normalization-compare`, and the five
Task44 generic stages - never a new distributed framework - after first
profiling the stage's actual hotspot on the real corpus. No scientific
formula, RNG seed formula, permutation count, threshold, or output schema
changed anywhere in this task; `begin-end-analyze`'s core logic was
extracted verbatim into `internal/beginendanalyze` as a pure refactor (see
section 3).

## 1. Profile / parallelism audit

CPU and heap profiles were captured with `-cpuprofile`/`-memprofile`
(wired into `begin-end-analyze` for this task, reusing the existing
`internal/profiling` package - no bespoke profiling code) against the real
Astafiev corpus at production parameters. Wall time: `15m38.7s` (profiling
overhead included; the un-profiled baseline measured later was `14m39.9s`,
consistent with the recorded `14m52s`).

| operation | CPU share (cumulative) | natural work unit | independent? | RNG dependency? | reduction dependency? | distributable? |
|---|---:|---|---|---|---|---|
| `directedDistance` (called twice per pair: line + page scope) | 65.01% | candidate pair | yes | no | none (pure per-pair result) | **yes** |
| `pageBalance` (called once per pair) | 29.42% | candidate pair | yes | no | aggregated afterward by `calibratePageBalance` (cross-candidate, cheap, stays local/sequential) | **yes**, bundled into the same per-pair batch as `directedDistance` |
| `runPermutations` (100 reps; `enumerateWindowHits` x2/rep) | 0.89% (`enumerateWindowHits`) + negligible `shuffledCorpus` allocation | permutation replicate | yes per-rep, but one sequential `*rand.Rand` stream (`rand.New(rand.NewSource(seed))`) is advanced across all 100 reps, never reseeded per rep | **yes - shared sequential stream** | commutative sum/exceed accumulation (`accumulateMoments`) | **not distributed** - see section 2; negligible cost (<1%) makes this moot |
| `nestingCounts` (up to `2*max-candidates` = 2000 candidates, post-sort) | not in top 25 profiler nodes (<0.5%) | candidate (post-sort subset) | yes | no | none | not worth distributing - negligible cost |
| `calibratePageBalance` / `sortCandidates` / final split | negligible (absorbed into GC/runtime overhead in the profile) | N/A (cross-candidate reduction) | N/A | no | inherently sequential aggregation over all candidates | stays local/sequential |

`directedDistance` + `pageBalance`, both called directly from the
`(ai,bi)` candidate-pair double loop, account for **94.43%** of wall time
on the real production workload. The permutation/RNG loop that this
stage's `-permutations` flag controls is, at the production default of
100, under 1% of the cost - the opposite of what stages 23/24/27 in
`DISTRIBUTED_GENERIC_STAGES_AUDIT.md` found, where the permutation-null
loop *was* the dominant cost. This is the cleanest possible case for
Task47's RNG-safety requirement: the expensive, distributable unit
(candidate pairs) has **zero** RNG dependency at all, so the RNG-bearing
part of the computation never needs to move off the coordinator - see
section 2.

`pageBalance`'s allocation profile (`alloc_space`) shows it responsible for
81.5% of all allocation (`diffs := make([]float64, len(pages))`, one
`len(pages)`-sized slice per candidate pair, called ~1.3M times) - a real
allocation hotspot, but distributing the same computation across processes
naturally distributes this allocation pressure too; no separate fix was
needed or attempted (task47 is about distribution, not algorithmic
optimization of the existing formulas).

## 2. Natural work unit and RNG audit

**Work unit: a batch of flat candidate-pair indexes.** The candidate space
is `k*k` (`k` = eligible token count after `-min-frequency`/`-include-unclear`
filtering; the diagonal `ai==bi` is always skipped, exactly as the
pre-Task47 nested loop's `if ai == bi { continue }` did), addressed by
`idx = ai*k + bi`. Task47 section 2 explicitly prefers deterministic
batches of candidate-ID ranges over one job per pair when the pair count is
large; at production scale (`k=1140`, `1,298,460` pairs), one-job-per-pair
would be over a million remote jobs - far too fine-grained. A batch is
`ComputeBatch(ws, batchIndex, batchSize)`: given a `Workspace` (built once,
read-only thereafter) it computes every pair in
`[batchIndex*batchSize, min((batchIndex+1)*batchSize, k*k))`, skipping the
diagonal, using the exact same per-pair body (`Workspace.candidateAt`) the
pre-Task47 inline loop ran.

**RNG: preserved exactly, never touched.** `runPermutations`'s single
sequential `*rand.Rand` stream (seeded once from `-random-seed`, advanced
across all 100 reps via `rng.Shuffle`) is completely unaffected by
distribution: `LoadForDistribution` runs it to completion, sequentially, on
whichever process calls it (the coordinator, and independently, once each,
every worker process reconstructing its own `Workspace`) - correct and
byte-identical every time because it is a pure function of
`(corpus, dictionary, parameters)`, never split or parallelized. The
distributable unit, `ComputeBatch`, only *reads* the resulting frozen
`lineMoments`/`pageMoments` (via `significance()`); it never draws from the
RNG. This is Task47 section 4's explicitly preferred option -
"preserve the existing RNG stream and prove work units are already
independent" - realized in the strongest possible form: the distributed
unit has no RNG dependency whatsoever, so there is no seed-formula
decision to make and nothing to document as a limitation.

**Reduction.** `runCandidateBatches` (the dispatcher `RunAndWrite`/
`collectCandidates` builds once per run) collects `BatchResult`s into a
`pending map[int]BatchResult` keyed by batch index and only appends a
batch's candidates to the running slice once every batch index up to
`next` has arrived, walking `next` upward in strict ascending order
regardless of which goroutine/process/remote worker actually finished
first - the identical mechanism `runBattery`/`runBatteryDispatch` use for
every other Task42-44 distributed stage. Because batches partition the
flat pair-index space contiguously, this reconstructs *exactly* the
ascending-pair-index order the pre-Task47 single-threaded double loop
produced natively, so every later step - `calibratePageBalance`,
`sortCandidates`, `nestingCounts`, the final local/main split - runs over
an identical candidate slice regardless of worker count, batch size, or
completion order. `TestComputeBatchPartitioningMatchesSingleBatch` and
`TestAnalysisMatchesAcrossExecutorWorkerCounts` verify this directly.

## 3. Architecture: reused, not reinvented

`begin-end-analyze` was the first Task28-44 distributed stage whose entire
scientific implementation lived in a `package main` CLI directory rather
than an `internal/*` library package. Task47 extracted it, unchanged, into
`internal/beginendanalyze` (`types.go`, `corpus.go`, `analyze.go`,
`reports.go` - byte-for-byte the same formulas, only reorganized around a
`Workspace`/`Config` split so the executor indirection has something to
call), following the exact same per-stage pattern Task44 established:

1. `internal/beginendanalyze/executor.go`: `CandidateBatchExecutor`
   interface, `defaultCandidateBatchExecutor` wrapping the unchanged
   per-pair computation, and `runCandidateBatches`, the dispatcher
   described in section 2.
2. `internal/beginendanalyze/distribution.go`: `Fingerprint(c Config)`
   hashing the corpus/dictionary file contents plus every scientific
   parameter (`LoadForDistribution` itself lives in `analyze.go`, since it
   is also the top-level preamble `RunAndWrite` calls - a pure extraction,
   not a new loader).
3. `internal/conditionalregime` gained one `begin_end_executor.go`
   (`beginEndInit`, `NewBeginEndProcessExecutor`, `NewBeginEndRemoteExecutor`,
   a JSON-marshaling adapter over `jobExecutor.RunBlob`), one
   `newBeginEndRemotePool` in `remote.go` (stages the corpus *and* the
   dictionary by content hash - this workload has no metadata-map/Generic
   mode, the dictionary is always required), a `beginEndComputer` type, and
   one new `case "begin_end_candidate_batch"` in both `RunWorker`'s
   process-mode switch and `runWorkerGeneration`'s remote-mode switch in
   `worker.go`.
4. `protocolMessage` gained a small field group (`DictionaryPath`,
   `MaxWindow`, `PermutationMode`, `IncludeUnclear`, `MaxCandidates`,
   `CandidateBatchSize`); `Permutations`, `Seed`, `MinTokenCount` are
   reused verbatim, per the file's own stated convention.
5. `begin-end-analyze/main.go` gained the identical flag block
   `normalization-compare`/`positional-continuation-validate` already use
   (`-executor`, `-workers`, `-remote-listen`, `-tls-cert`, `-tls-key`,
   `-client-ca`, `-remote-deny-list`, `-remote-timeout`, `-remote-retries`,
   `-internal-worker`), plus one new stage-specific flag,
   `-candidate-batch-size` (default `beginendanalyze.DefaultCandidateBatchSize
   = 2048`, justified in section 6). No `-coordinator`/persistent-worker
   mode of its own was needed: a remote worker for this job type is the
   *existing* `conditional-regime-analyze -coordinator ...` binary, exactly
   as for every Task44 stage - dispatch is driven entirely by the
   coordinator's handshake `Workload` field, never by which binary asked to
   be a worker.
6. `pipeline-orchestrate/stages.go`: the `begin-end-analyze` entry gained
   `Executor: true` (one line); a dedicated test,
   `TestBeginEndAnalyzeManifestIncludesRemoteExecutorArgs`
   (`pipeline-orchestrate/begin_end_manifest_test.go`), builds a real
   manifest with `-executor remote` and asserts the stage's actual
   generated `Args` contain every remote/mTLS flag *alongside* (never
   instead of) `-dictionary`/`-output-dir` - Task47 section 10's explicit
   requirement that `Executor: true` alone is not sufficient evidence.
   `begin-end-analyze` has no checkpoint mechanism (before or after this
   task) and none was added - like `normalization-compare`, this is a
   purely `Executor: true` stage; section 9's "coordinator restart/resume"
   requirement is satisfied the same way `normalization-compare`'s is,
   by two independent coordinator generations reproducing the oracle (see
   section 5's interrupted/resumed test), not by a resumable on-disk
   checkpoint.

No new distributed framework, no copied code from another package's
executor - every mechanism above is the existing Task28-44 machinery
gaining one more workload type.

## 4. Two real bugs the real-corpus study caught (neither existed before this task)

Every unit test and small-fixture remote test (ASCII-only tokens, at most a
few hundred candidate pairs) passed before the real-corpus scaling study
ran. Both bugs below are real, were caught only by running the actual
production workload, were root-caused to exact mechanisms, fixed, and then
locked in with regression tests using small-but-realistic fixtures so they
cannot silently reappear.

### 4a. Non-UTF-8 token corruption over the wire

The Astafiev (and Voynich) corpora contain tokens that are not valid UTF-8
byte sequences. Go's `encoding/json` silently replaces invalid UTF-8 in a
`string` with U+FFFD when marshaling - harmless for every prior workload's
wire payload (`float64`s, small numeric/edge-key maps), but
`begin_end_candidate_batch` is the first workload whose result payload is
verbatim corpus token text (`Candidate.BeginCandidate`/`EndCandidate`).
`-executor remote` (and `-executor process`, which shares the same
`jobExecutor.RunBlob` JSON path) at real corpus scale produced output that
differed from the local/goroutine oracle in **1088 of 1100** top-ranked
candidate pairs - the ranking/scores were computed correctly throughout
(confirmed by the handful of pairs whose *values* still matched exactly);
only the token *text* was corrupted in transit, which cascaded into a
different top-N selection since many distinct corrupted tokens collided.

Fixed with a transport-only wire type,
`wireBeginEndBatchResult`/`wireCandidate`
(`internal/conditionalregime/begin_end_executor.go`): the two token fields
are carried as `[]byte` (which `encoding/json` base64-encodes and decodes
exactly, regardless of UTF-8 validity) instead of `string`, with the
embedded `Candidate`'s own two string fields zeroed before marshal and
restored from the raw bytes after unmarshal. `internal/beginendanalyze`'s
own `Candidate`/`BatchResult` types - used for YAML output and the
goroutine backend's in-process path, where no serialization ever happens -
are completely unchanged; this is a pure fix at the JSON transport
boundary. `TestBeginEndWireRoundTripPreservesNonUTF8Tokens` locks this in
with a small deliberately-non-UTF-8 fixture (and documents, in the test
itself, that a naive `json.Marshal` round trip does corrupt the value -
proving the test would have caught the real bug).

### 4b. A 1 MiB message-size cap far too small for this workload's payload

`internal/conditionalregime/remote.go`'s `maxRemoteMessageBytes` (bounding
one lease/result/handshake JSON message, distinct from
`maxRemoteInputBytes`, which bounds a staged *input file*) was `1 << 20`
(1 MiB) - correct for every workload before this one, whose results are
scalars or small maps. A `begin_end_candidate_batch` result at the
production default batch size (2048 pairs) marshals to roughly **5.5 MB**
on the real Astafiev corpus (windows/histogram tables make each candidate
~2.7 KB on average). `decodeJSONBody` silently truncated the read at the
cap and returned `"request exceeds size limit"` without draining the rest
of the request body; the worker's still-in-flight POST then blocked on TCP
backpressure (the coordinator had stopped reading) until the full
`-remote-timeout` elapsed - repeated once per retry, with **no error ever
logged**, since the failure never reached the application layer on the
worker side. Symptom: `-candidate-batch-size 2048 -executor remote` hung
for `remote-timeout * (1+retries)` (20 minutes, at this task's
`-remote-timeout 5m -remote-retries 3`) and then failed outright with "no
worker returned a result after 4 attempt(s)".

Fixed by raising `maxRemoteMessageBytes` to `32 << 20` (32 MiB) - comfortable
headroom over every batch size this task's granularity study exercises
(up to 8192 pairs, ~22 MB), while remaining a real, explicit bound rather
than unbounded, following the same reasoning the pre-existing
`maxRemoteInputBytes` comment already documents for an analogous
worker-input-size incident. `TestBeginEndRemoteHandlesPayloadOverOldOneMiBCap`
locks this in with a small synthetic corpus (60 tokens, several pages)
whose single whole-pair-space batch marshals to ~9 MB; the test was
verified to fail with the exact same "no worker returned a result" symptom
against the reverted 1 MiB constant before being left passing against the
fix.

Both bugs are now covered by dedicated regression tests using deliberately
small-but-realistic fixtures (non-ASCII bytes; enough distinct tokens to
cross the old size cap) specifically so neither requires a 15-minute real
corpus run to catch in `go test ./...` again.

## 5. Correctness oracle

All six required comparisons were run; every output file
(`begin_end_candidates.yaml`, `begin_end_top.tsv`, `begin_end_report.md`)
was SHA256-identical in every case - both against each other and against
`experiments/astafiev-1000-culinar-receipt-v1`'s existing frozen
production output (proving the extraction into `internal/beginendanalyze`
changed nothing):

| # | configuration | corpus | SHA256 match |
|---|---|---|---|
| A | original/local reference (`experiments/astafiev-1000-culinar-receipt-v1`'s existing frozen output, produced by the pre-Task47 `package main` implementation) | real Astafiev, production params | baseline |
| B | new goroutine (`-executor goroutine -workers 1`) | real Astafiev, production params | **identical to A** |
| C | process (`-executor process`) | small fixture (`internal/beginendanalyze` executor tests exercise the same dispatcher goroutine-mode already proven identical to A; process mode shares `jobExecutor`/wire code with remote, covered by D/E) | identical |
| D | remote, 1 worker | real Astafiev, production params (`remote-w1`, batch 2048) | **identical to A** |
| E | remote, multiple workers (2, 5, 10) | real Astafiev, production params | **identical to A**, all three |
| F | interrupted remote + resume (two independent coordinator generations, first half then second half of the batch space) | small fixture | identical (`TestBeginEndRemoteInterruptedThenResumedMatchesOracle`) |

Additional automated coverage, run as part of `go test ./...`:

| requirement | test |
|---|---|
| local vs remote, 1 worker | `TestBeginEndRemoteMatchesLocalOracle` |
| N workers, arbitrary completion order | `TestBeginEndTwoRemoteWorkersMatchLocalInAnyCompletionOrder` |
| retry after synthetic worker failure | `TestBeginEndRemoteRetryAfterWorkerFailure` |
| same seed repeated | `TestBeginEndRemoteSameSeedRepeated` |
| interrupted + resumed (two coordinator generations) | `TestBeginEndRemoteInterruptedThenResumedMatchesOracle` |
| non-UTF-8 token wire round trip | `TestBeginEndWireRoundTripPreservesNonUTF8Tokens` |
| payload over the old 1 MiB cap | `TestBeginEndRemoteHandlesPayloadOverOldOneMiBCap` |
| batch-size/worker-count invariance (in-process) | `TestAnalysisMatchesAcrossExecutorWorkerCounts`, `TestComputeBatchPartitioningMatchesSingleBatch` |
| dispatcher order restoration / error propagation | `TestRunCandidateBatchesRestoresCanonicalOrderRegardlessOfCompletionOrder`, `TestRunCandidateBatchesPropagatesWorkerError`, `TestRunCandidateBatchesPropagatesOnReadyError` |

All are `reflect.DeepEqual` or SHA256 comparisons against a local,
non-distributed oracle - never a tolerance/approximate comparison.

## 6. Granularity study

Measured on the real Astafiev corpus at production parameters, 5 remote
workers, disposable loopback mTLS PKI (`conditional-regime-pki`), varying
only `-candidate-batch-size`:

| batch size | batches | wall time | vs local baseline (879.9s) | notes |
|---:|---:|---:|---:|---|
| 128 | 10,145 | 668.0s | 1.32x | many small requests; per-request overhead dominates |
| 512 | 2,537 | 534.3s | 1.65x | |
| **2048 (chosen default)** | **635** | **532.0s** | **1.65x** | effectively tied with 512, fewer requests |
| 8192 | 159 | 579.0s | 1.52x | regresses: fewer, chunkier batches produce a stronger straggler effect near the end of the run |

All four produced SHA256-identical output. `2048` was chosen as
`beginendanalyze.DefaultCandidateBatchSize`: it matches 512's wall time
with fewer, larger requests (more comfortable margin under the 32 MiB
message cap from section 4b) and avoids 8192's straggler regression.

## 7. Scaling study

Same real Astafiev corpus, production parameters, `-candidate-batch-size
2048`, single 12-core development machine, disposable loopback mTLS PKI -
`-workers N` matched to the number of connected `conditional-regime-analyze
-coordinator ...` worker processes (the existing shared worker binary,
unmodified, now also serving `begin_end_candidate_batch` per section 3):

| configuration | wall time | speedup vs local | parallel efficiency | coordinator CPU / RSS | worker CPU (each) / RSS (each) |
|---|---:|---:|---:|---|---|
| local (`-executor goroutine -workers 1`) | 879.9s (14m39.9s) | 1.00x | 100% | n/a (single process) | n/a |
| remote, 1 worker | 1123.6s (18m43.3s) | **0.78x (slower than local)** | n/a | 125.5s / 5.1 GB | 980.5s / 312 MB |
| remote, 2 workers | 697.5s (11m37.5s) | 1.26x | 63% | 138.0s / 4.9 GB | ~622.8s / ~300 MB |
| remote, 5 workers | 532.0s (8m52.0s) | 1.65x | 33% | 154.6s / 4.7 GB | ~500.0s / ~310 MB |
| remote, 10 workers | 576.4s (9m36.4s) | 1.53x (**worse than 5 workers**) | 15% | 205.7s / 4.8 GB | ~527.7s / ~305 MB |

All four remote rows produced SHA256-identical output to the local
baseline and to the frozen production experiment. Jobs completed: 635 (one
per batch) in every configuration except `-workers 1`, where the same 635
batches were leased sequentially to the single worker. Average job
duration at 5 workers: `532.0s * 5 / 635 ≈ 4.2s` of wall-clock lease time
per batch (compute + marshal + network + unmarshal combined).

**Coordinator memory is a real, sizable resource cost**: 4.7-5.1 GB RSS in
every remote configuration, because the coordinator accumulates all
`1,298,460` `Candidate` structs (each carrying a `Histogram` map and a
`Windows` slice) in memory before running `calibratePageBalance`/
`sortCandidates`/`nestingCounts` - identical to what the pre-Task47
in-process implementation already held in memory for the same computation,
just now also holding the JSON-decode intermediate garbage from
1,298,460/2048 ≈ 635 incoming results. This is not a regression introduced
by distribution; it is the same peak memory the original single-process
implementation always needed, observed for the first time at production
scale because Task47's profiling run was the first to capture it deliberately.

### Honest bottleneck: this stage is transport-bound once distributed, not compute-bound

Section 1 measured the distributed unit as 94.43% of total CPU time when
run in a single process (no serialization ever happens in-process). Once
distributed, each worker's own CPU time is dominated not by
`Workspace.candidateAt` but by **JSON-marshaling and TLS-encrypting its
own share of a large, richly-detailed per-candidate payload**: at 5
workers, each worker consumed ~500 CPU-seconds to compute its
`1,298,460 / 5 ≈ 259,692`-pair share - roughly **3x** the ~176 CPU-seconds
that share would cost in the local, no-serialization baseline
(`879.9s / 5`). Each candidate's `Histogram map[int]int` (up to 55 distinct
keys) and `Windows []WindowResult` (10 entries) make its wire
representation (~2.7 KB average on the real corpus, per section 4b)
disproportionately large relative to the ~0.68ms of actual computation
(`879.9s / 1,298,460` pairs) it took to produce - unlike every other
Task42-44 distributed stage, whose per-job wire payload (a `float64` or a
small edge-key map) is negligible next to its compute cost. This is why
speedup plateaus at 1.65x around 5 workers rather than approaching linear
scaling, and why 10 workers is *worse* than 5: beyond 5 workers, the
single coordinator process (and the shared 12-core machine hosting every
worker and the coordinator simultaneously) becomes the limiting resource,
not the candidate computation itself - the same single-machine measurement
ceiling `NORMALIZATION_COMPARE_DISTRIBUTION_AUDIT.md` section 7 and
`DISTRIBUTED_GENERIC_STAGES_AUDIT.md` section 6 already documented for
other stages, compounded here by a genuinely larger per-job payload. On a
real multi-host fleet, where each worker's marshal/TLS overhead runs on
its own dedicated hardware rather than sharing 12 cores with the
coordinator and every other worker, efficiency at 5-10 workers should be
substantially better than this single-machine proxy shows - but this
document reports the real, measured, honest number rather than an
extrapolated one, per Task47 section 14's explicit instruction not to hide
a weak result behind an assumed one.

**Recommendation: 5 workers** for this stage at this batch size on
hardware similar to this single-machine measurement. Going beyond that
provided no further benefit here; a real fleet deployment should be
re-measured on its own hardware before assuming a higher worker count
helps; this document does not claim it will.

## 8. Validation

```
go build ./...                                                      # clean
go vet ./...                                                        # clean
go test ./...                                                       # 33/33 tested packages ok
go test -race ./internal/beginendanalyze/... ./internal/conditionalregime/... ./pipeline-orchestrate/...
                                                                     # ok, no data races in any Task47 code
git diff --check                                                    # no whitespace errors
```

One pre-existing data race, unrelated to this task, was found while
running `go test -race ./...` across the whole repository: a Task44 race
in `internal/positionalcontinuation.buildBoundaryDistanceRows` /
`LoadForDistribution` (a write in one worker-generation's `LoadForDistribution`
racing a read in a different generation's `ComputeBattery`, both reachable
from `TestPositionalContinuationTwoRemoteWorkersMatchLocalInAnyCompletionOrder`).
It was confirmed to reproduce identically on a clean worktree checked out
at the commit immediately before this task began (`f6bf4bf`), with none of
Task47's changes present - it is not introduced or touched by this task,
and fixing an unrelated stage's pre-existing race is out of this task's
scope (Task47 section 18: "не распределять другие stages в рамках Task47").
It is noted here for visibility, not fixed here.

`experiments/astafiev-1000-culinar-receipt-v1`'s existing frozen output was
never modified by this task - every measurement in sections 5-7 wrote to a
scratch directory outside the repository and compared its SHA256 against
that frozen output, never overwriting it.
