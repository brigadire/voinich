# Fingerprint V2.1 PRNG determinism contract

## Normative contract

Fingerprint V2.1 retains the frozen `math/rand` pseudo-random algorithm and
all published base seeds.  A computation's stream is identified by the tuple

`(producer version, corpus ID, metric/null ID, base seed, replicate order)`.

The current pipeline is sequential: there is no worker pool and no shared RNG
between goroutines.  Each top-level stochastic test receives an RNG derived
from its frozen base seed by the existing producer.  Replicates run in the
ascending integer order `0..n-1`.  Within a replicate, independent map-backed
groups are traversed in ascending integer-key or lexicographic string-key
order.  Aggregates use that same stable order.

Consequently no PRNG draw is assigned according to map insertion/runtime
order, filesystem enumeration, goroutine scheduling, `GOMAXPROCS`, wall-clock
time, PID, hostname, or absolute workspace path.  No implicit or time-based
seed is permitted in an authoritative computation.

## Frozen streams

- Main Fingerprint V2 analysis: config seed `20260824`; existing deterministic
  per-test offsets remain unchanged.
- Task79 permutations/bootstrap: 1000/1000 replicates from the same frozen
  config seed and existing derivation.
- PF4 leaf-paired null: seed `20260824`, 1000 permutations.
- HR3/HR5 fold assignment: seed `40260824`, five folio-block folds.

The unchanged producer derivation assigns primary corpus seed `S`; control
`i` receives `S + (i+1)*10000000`. Grammar replicates use
`S + (mode_index+1)*1000000 + replicate`. LP random-pair, attachment, LP3,
EF2, and EF3 streams use the existing `+2000000+replicate`, `+3000000`,
`+4000000`, `+5000000`, and `+6000000` namespaces. Edit-graph validation,
cross-scale, and Task79 use `+8000000`, `+9000000`, and `+10000000`.
Cross-scale tests retain their explicit `+100` through `+700` substream
offsets; Task79 tests retain their metric-indexed prime offsets declared in
`task79.go`. Thus metric/null identity, corpus position, and replicate index
determine every stream without execution-order input.

Changing the PRNG algorithm, seed derivation, null population, number of
replicates, estimator, or threshold is a scientific-definition change and is
outside Task83b.  Sorting the immutable group identifiers before consuming the
same draws is an implementation-determinism fix.

## Verification

`TestCS7MapInsertionOrderDoesNotAffectStatisticOrPRNG` constructs equivalent
maps in opposite insertion orders and requires exact statistic/null identity.
`TestFingerprintDeterminismAcrossProcessRestartAndGOMAXPROCS` self-executes in
fresh OS processes at `GOMAXPROCS=1,2,4` and requires byte-identical JSON.
The full clean RUN_A/RUN_B/RUN_C reconstruction is the release gate.
