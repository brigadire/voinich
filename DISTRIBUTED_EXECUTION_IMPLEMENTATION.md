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
