# Distributed Execution for Generic Stages 23-27 (Task44)

Task43 gave stages 23-27 (`token-relation-validate`,
`replicated-local-structure-audit`, `higher-order-sequence-validate`,
`positional-continuation-validate`, `transition-network-validate`) a
corpus-only `-generic-corpus` mode, but left all five single-process,
explicitly deferring distribution to a follow-up task. Task44 is that
follow-up: it puts each of these five stages on the exact Task28-35 mTLS
coordinator/worker executor already serving `conditional-regime-analyze`,
`structural-projection-analyze`, and `normalization-compare` - never a new
distributed framework - after first auditing each stage's expensive loop
for a provably independent unit of work. No scientific algorithm, null
model, RNG semantics, permutation count, threshold, window, fold, or
output schema changed anywhere in this task.

## 1. Dependency/parallelism audit

| stage | expensive loop | natural work unit | RNG dependency | reduction order | distributable |
|---|---|---|---|---|---|
| 23 token-relation-validate | 6 batteries: `direction(Refine)Permutations`, `sequence(Refine)Permutations`, `profile(Refine)Permutations` | 1 permutation replicate, per family | fresh seed per run: `c.Seed + run*1000003` - independent | commutative per-candidate-ID sum/exceed-count | **yes**, replicate-level |
| 24 replicated-local-structure-audit | distance-null loop | 1 permutation replicate | fresh seed per run (`c.Seed+run*104729+11`) | **positional**: a later jackknife reads `Distance[key][run]` by index - the reducer must restore run order before storing | **yes**, replicate-level, with order-restoration |
| 24 | shuffle-null / markov-null loops | 1 permutation replicate | fresh seed per run (`+23`/`+37`) | commutative sum/exceed-count | **yes**, replicate-level |
| 25 higher-order-sequence-validate | `runCMI` + LOBO/jackknife/context/position/structural-family, per candidate | **whole candidate** (never a permutation - `cmiWorkspace.permute` advances one sequential `*rand.Rand` stream per candidate; splitting it would change RNG semantics) | `candidateSeed(base, sequence)` independent per candidate; permutations *within* one candidate are not independently seeded | per-candidate result, order-independent | **yes, candidate-level only** |
| 26 positional-continuation-validate | 5 named batteries: `postest_line`, `postest_block`, `stratified_line`, `stratified_block`, `boundary` | **whole battery** (same shared-stream-RNG constraint as 25: each battery builds one workspace/`*rand.Rand` and loops internally) | `seedFor(c.Seed, name)` independent per battery | self-contained per battery | **yes, battery-level** |
| 27 transition-network-validate | primary + refinement permutation-null loops | 1 permutation replicate (global index, continued across the primary/refine boundary) | fresh seed per rep: `seed+rep*0x1f123bb5` | commutative per-edge-key exceed-counters | **yes**, replicate-level |

All five stages have a safe, provably independent unit; none received a
forced/fake distribution. Stages 25 and 26 get a *coarser* unit than
23/24/27 specifically because their existing permutation math shares one
sequential RNG stream within a candidate/battery - task44 forbids changing
that, so the unit of independence sits one level up (candidate/battery),
never the permutation itself. Every one of the five packages' existing
checkpoint structs already resumes at exactly this granularity once the
loop is routed through an executor indirection - no checkpoint schema
changed, only (for stage 25) the *resume resolution* from six
separately-tracked parts per candidate to one part-set per candidate (see
section 5).

## 2. RNG / reduction audit

Every stage's RNG formula was read, confirmed already independent per work
unit, and left untouched:

- **23**: `c.Seed + run*1000003` (all six families share this one formula
  and reuse the shared `runBattery` dispatcher).
- **24**: `c.Seed + run*104729 + {11,23,37}` for distance/shuffle/markov
  respectively.
- **25**: `candidateSeed(c.Seed, cand.Sequence)` (a SHA-256-derived mix,
  one per candidate - unchanged from pre-Task44 `run.go`), inside which
  `runCMI`'s own `rand.New(rand.NewSource(seed))` advances sequentially
  across that candidate's permutations exactly as before.
- **26**: `seedFor(c.Seed, name)` for each of the 5 battery names
  (`"line_position"`, `"block_position"`, `"stratified_line"`,
  `"stratified_block"`, `"boundary"` - note these are the *seedFor* name
  strings, distinct from the *battery job* names `postest_line`/
  `postest_block`/... used for `JobID.Combination`/checkpoint keys).
- **27**: `seed + rep*0x1f123bb5`, with the refine phase's global replicate
  index (`c.Permutations + n`) threaded through `runBattery`'s `resumeFrom`/
  `n` parameters unchanged, so the refine-phase seed continuation across
  the primary/refine boundary is exactly as before.

**Reduction.** Every stage's dispatcher (`runBattery` in 23/24/27,
`runCandidateBattery` in 25, `runBatteryDispatch` in 26 - three names for
one identical mechanism, retyped per package) collects results into
`pending map[int]Result` keyed by logical work-unit index and only calls
`onReady` once every index up to `next` has arrived, walking `next` upward
in strict ascending order regardless of which goroutine/process/remote
worker actually finished first. This single mechanism satisfies both
"commutative sums are safe regardless of completion order" (23, 24's
shuffle/markov, 27) and "positional results must be restored to canonical
order before reduction" (24's distance-null, whose `cp.Distance[key]`
append only ever happens inside `onReady`, which always fires in ascending
run order - proven directly by
`internal/replicatedlocalaudit/executor_test.go`'s
`TestDistanceOrderRestorationMatchesSequentialAppend`, which drives
out-of-order completion through a stub executor and confirms the sequential
append order survives).

## 3. Stop condition: distribution is justified

Task44's own charter is narrower than Task42's ">=50% of wall time"
threshold: it asks for a dependency audit proving *whether a safe unit of
independent work exists at all*, and forbids inventing one where it
doesn't. Section 1 shows all five stages have one; none was skipped or
forced. Sections 24-27's permutation-null loops are already the
overwhelmingly dominant cost at production permutation counts (1,000-
10,000 per battery/family, mirrored exactly by 23/24/27's `-permutations`
flags), matching the same profile shape `NORMALIZATION_COMPARE_DISTRIBUTION
_AUDIT.md` measured for `normalization-compare`'s `AnalyzeFile` calls;
25/26's candidate/battery counts are inherently small (a handful of frozen
candidates/named batteries), so their *justification* is "no other unit is
RNG-safe," not "a large fraction of wall time" - see section 7 for the
honest scaling consequence of that.

## 4. Architecture: reused, not reinvented

No new distributed framework was created. Each of the five stages gained a
**new job type** on the exact Task28-35 coordinator/worker/mTLS/JobID/
lease/retry/deterministic-collection infrastructure that already served
`part_a_*`/`part_b_global_correction`, `structural_projection_trial`, and
`normalization_compare_baseline`:

| stage | job type | `JobID.Combination` | `JobID.ReplicateIndex` |
|---|---|---|---|
| 23 token-relation-validate | `token_relation_permutation` | `family` (`direction`/`refine_direction`/`sequence`/...) | permutation run |
| 24 replicated-local-structure-audit | `replicated_local_null` | `phase` (`distance`/`shuffle`/`markov`) | permutation run |
| 25 higher-order-sequence-validate | `higher_order_candidate` | candidate's frozen sequence text | always 0 |
| 26 positional-continuation-validate | `positional_continuation_battery` | battery name (`postest_line`/...) | always 0 |
| 27 transition-network-validate | `transition_network_permutation` | phase (`primary`/`refine`) | permutation run |

Per stage, the same seven-part pattern was applied:

1. The target package gained an `executor.go`: an interface abstracting
   only *where* one work unit executes (`PermutationExecutor` in 23/24/27,
   `CandidateExecutor` in 25, `BatteryExecutor` in 26), a
   `defaultPermutationExecutor`/`defaultCandidateExecutor`/
   `defaultBatteryExecutor` wrapping the pre-existing computation
   byte-for-byte unchanged, and the `runBattery`/`runCandidateBattery`/
   `runBatteryDispatch` dispatcher described in section 2.
2. The target package gained a `distribution.go`: `LoadForDistribution(c
   Config)` (calling the exact same private loaders `RunAndWrite` already
   used - a pure extraction, never a reimplementation) and `Fingerprint(c
   Config) (string, error)`, delegating to each package's own pre-existing
   fingerprint formula so there is only ever one formula, never two.
3. `RunAndWrite`'s loop(s) were refactored to build the executor once and
   dispatch through it, with the original accumulation/checkpoint-write
   body copied verbatim into `onReady` - zero formula changes anywhere.
4. `internal/conditionalregime` gained one `<stage>_executor.go`
   (`<stage>Init`, `New<Stage>ProcessExecutor`, `New<Stage>RemoteExecutor`,
   a thin JSON-marshaling adapter over `jobExecutor.RunBlob`), one
   `new<Stage>RemotePool` in `remote.go`, a `<stage>Computer` type, and one
   new `case` in each of `RunWorker`'s process-mode switch and
   `runWorkerGeneration`'s remote-mode switch in `worker.go`.
5. `protocolMessage` gained a small field group per stage (`DiscoveryDir`/
   `Generic`/`RefinePermutations` for 23; `RelationDir` for 24; `AuditDir`
   for 25; `HigherOrderDir` for 26); pre-existing same-named/same-type
   fields (`Permutations`, `Seed`, `CorpusPath`, `TokenMetadataMap`) are
   freely reused across workloads, since only one workload's semantics
   ever apply per connection.
6. Each stage's `main.go` gained the identical flag block already used by
   `normalization-compare/main.go` (`-executor`, `-workers`,
   `-remote-listen`, `-tls-cert`, `-tls-key`, `-client-ca`,
   `-remote-deny-list`, `-remote-timeout`, `-remote-retries`,
   `-internal-worker`), wired to `New<Stage>{Process,Remote}Executor` and
   `conditionalregime.RunWorker`. None of the five needed a
   `-coordinator`/persistent-remote-worker mode of their own - a remote
   worker for any of these five job types is started with the *existing*
   `conditional-regime-analyze -coordinator ...` binary, since dispatch is
   driven entirely by the coordinator's handshake, not by which binary
   asked to be a worker.
7. `pipeline-orchestrate/stages.go`: each stage's entry gained
   `Executor: true` (one line), verified end-to-end via a manual `manifest
   -executor remote ...` check confirming the recorded `args` carry the
   remote/mTLS flags alongside `-generic-corpus` and never a stale
   `-token-metadata-map`.

### Directory-reconstruction generalization

`loadCandidates`-style loaders that hardcode specific filenames within one
directory (token-relation-validate's `-discovery-dir`,
replicated-local-structure-audit's `-relation-dir`/`-discovery-dir`,
higher-order-sequence-validate's `-audit-dir`/`-discovery-dir`,
positional-continuation-validate's `-higher-order-dir`) cannot simply be
handed a worker's hash-named cache directly. Task44 solved this once, in
`internal/conditionalregime/remote.go`, and reused it for every stage that
needed it: `reconstructNamedDir(cacheDir, inputs, prefix, dirName)` rebuilds
a directory of staged blobs under their *original* filenames (hardlink,
falling back to a copy across filesystems) from any subset of a worker's
cache keyed by a `"<prefix>:<name>"` convention. Stage 23's discovery
directory needed exactly one call; stage 24 needed two (`"relation:"` and
`"discovery:"` prefixes, since it depends on both its own discovery inputs
*and* stage 23's frozen output); 25 and 26 each needed one more
(`"audit:"`/`"higherorder:"`).

### Per-stage subtleties

- **23**: `directionScoresAll`/`buildDirectionEdges`/`sequenceScores`/
  `profilePermutationScores` all compute per-candidate scores independent
  of which other candidates share the call, so `ComputeReplicate` always
  scores the *full* family-typed candidate set rather than shipping
  per-job eligible-ID lists - this keeps the `JobID`/`jobExecutor.RunBlob`
  interface untouched, and is verified equivalent to the original
  eligibility-filtered computation by
  `TestComputeReplicateDirectionMatchesInlineComputation`.
- **27**: `ws.run` returns `map[EdgeKey]float64` (a struct key, which
  cannot be a JSON object key); `ComputeReplicate` stringifies it via
  `EdgeKey.String()` at the result-serialization boundary only - `ws.run`
  itself never changed.
- **24**: `distributionState` had to be exported to `DistributionState` -
  a worker-side adapter in `internal/conditionalregime` needs to name the
  type in a struct field, and an unexported type across a package boundary
  is a compile error, not merely a style nit (the same reason stage 27's
  `permWorkspace` became `PermWorkspace`, and stage 26's
  `higherOrderInputFiles`/`load.go`'s file-list vars became exported
  `*DirFiles` vars for the same staging reason).
- **25/26 - the JobID-collision constraint.** Because these two stages'
  `JobID.ReplicateIndex` is always 0, the same `JobID` can never be
  in-flight twice concurrently - `remotePool.runOutcome` explicitly
  rejects a duplicate in-flight `JobID` with "already in flight" (this is
  pre-existing Task33/34 code, unmodified). Sequential re-dispatch of an
  already-*completed* JobID is fine (it is removed from `pending` the
  moment its result arrives - exactly what the "same seed repeated"
  correctness test in section 6 relies on for every stage), but any
  "N workers, arbitrary completion order" or "interrupted + resumed"
  correctness test for stages 25/26 needs *multiple distinct* candidates/
  batteries dispatched concurrently, never N copies of one candidate's/
  battery's `JobID` the way 23/24/27's replicate-index-keyed tests could
  get away with.
- **25 - checkpoint resolution changed granularity, not schema.** Before
  Task44, `higherorderseq`'s checkpoint tracked six separate parts per
  candidate (`occ_cond`/`cmi`/`lobo`/`context_meta`/`jackknife`/
  `position_family`), each independently resumable. Since the distributed
  unit is now the *whole candidate*, `RunAndWrite` now treats a candidate
  as resumable only once every one of its six part-keys is marked done; a
  candidate that was only partially done under the old per-part
  granularity is simply recomputed whole on resume - idempotent, and never
  a different scientific value, just a smaller resume speedup than before
  in that one edge case. The checkpoint's JSON shape itself never changed.
- **26 - reordering, not restructuring.** The five distributable batteries
  (originally `runPart` calls at stages 4, 5, 7, 8 and 13 in `run.go`) were
  folded into one dispatch phase positioned where the first of them used
  to run. This changes only wall-clock order relative to the
  non-distributed parts in between (`aiin_control`, `model_lobo`,
  `cross_block`, `jackknife`, `line_vs_block`) - each of those was checked
  to depend only on occurrences/postest results already available, never
  on the stratified/boundary results that now compute slightly earlier
  than before. `classify_write`, the only part that reads everything,
  still runs last regardless.

## 5. Correctness

Every requirement below is an automated test, run as part of `go test
./...`:

| requirement | stage 23 | stage 24 | stage 25 | stage 26 | stage 27 |
|---|---|---|---|---|---|
| local vs remote, 1 worker | `TestTokenRelationRemoteMatchesLocalOracle` | `TestReplicatedLocalAuditRemoteMatchesLocalOracle` | `TestHigherOrderRemoteMatchesLocalOracle` | `TestPositionalContinuationRemoteMatchesLocalOracle` | `TestTransitionNetworkRemoteMatchesLocalOracle` |
| N workers, arbitrary completion order | `TestTokenRelationTwoRemoteWorkersMatchLocalInAnyCompletionOrder` | `TestReplicatedLocalAuditTwoRemoteWorkersMatchLocalInAnyCompletionOrder` | `TestHigherOrderTwoRemoteWorkersMatchLocalInAnyCompletionOrder` | `TestPositionalContinuationTwoRemoteWorkersMatchLocalInAnyCompletionOrder` | `TestTransitionNetworkTwoRemoteWorkersMatchLocalInAnyCompletionOrder` |
| retry after synthetic worker failure | `TestTokenRelationRemoteRetryAfterWorkerFailure` | `TestReplicatedLocalAuditRemoteRetryAfterWorkerFailure` | `TestHigherOrderRemoteRetryAfterWorkerFailure` | `TestPositionalContinuationRemoteRetryAfterWorkerFailure` | `TestTransitionNetworkRemoteRetryAfterWorkerFailure` |
| same seed repeated | `TestTokenRelationRemoteSameSeedRepeated` | `TestReplicatedLocalAuditRemoteSameSeedRepeated` | `TestHigherOrderRemoteSameSeedRepeated` | `TestPositionalContinuationRemoteSameSeedRepeated` | `TestTransitionNetworkRemoteSameSeedRepeated` |
| interrupted + resumed (two coordinator generations) | `TestTokenRelationRemoteInterruptedThenResumedMatchesOracle` | `TestReplicatedLocalAuditRemoteInterruptedThenResumedMatchesOracle` | `TestHigherOrderRemoteInterruptedThenResumedMatchesOracle` | `TestPositionalContinuationRemoteInterruptedThenResumedMatchesOracle` | `TestTransitionNetworkRemoteInterruptedThenResumedMatchesOracle` |

All are `reflect.DeepEqual` (or SHA256-of-JSON-equivalent) comparisons
against a local, non-distributed oracle computed via each package's own
`LoadForDistribution`/`ComputeReplicate`/`ComputeCandidate`/
`ComputeBattery` - never a tolerance/approximate comparison, per Task44's
explicit "never substitute tolerance-based equivalence for exact identity"
requirement. Every test file reuses `internal/conditionalregime`'s shared
remote-test helpers (`newRemotePKI`, `startRemoteWorker`,
`workerHTTPClient`, `leaseUntilAssigned`) - the same helpers
`normalization_remote_test.go` established.

Package-level unit tests (`executor_test.go` in each of the five target
packages) additionally cover: `ComputeReplicate`/`ComputeCandidate`/
`ComputeBattery` determinism per work unit, the dispatcher's canonical-
order restoration under a stub executor with artificial out-of-order
completion, error propagation from both a failing worker and a failing
`onReady`, and (24 specifically) `TestDistanceOrderRestorationMatchesSeque
ntialAppend`, which proves the positional-reduction requirement directly.

Full validation, run at the end of this task:

```
go build ./...
go vet ./...
go test ./...
go test -race ./...
git diff --check
pipeline-orchestrate verify -experiment-dir experiments/voynich-v1   # 340/340 checksums, unchanged
```

`experiments/voynich-v1` (the frozen Task36 baseline) was never touched by
this task and re-verifies at 340/340 - none of the code changes above run
against it, and none of the five stages' scientific output paths changed.

## 6. Scaling study

