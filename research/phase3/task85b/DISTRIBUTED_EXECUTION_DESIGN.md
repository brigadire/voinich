# Distributed execution design

The scientific manifest is canonical and immutable. A manifest compiler expands it into a DAG of small jobs. The existing Phase-I mTLS pull queue leases ready nodes to any authenticated worker. Workers have no scientific gate logic: they validate the immutable bundle, run the named computation, and return a content-addressed bundle plus operational telemetry. A separate verifier checks evidence; only then can the deterministic aggregator run.

## Identity and granularity

`job_id = SHA256(canonical_json(experiment, protocol_version, corpus_id, model, candidate, scale, replicate, seed, stage, sorted input/dependency hashes, code_hash, config_hash, output_schema))`. Host, worker, lease, attempt, time, and completion order are excluded.

FIT is candidate × corpus/split; PREDICTIVE is fitted candidate × heldout split; GENERATION is candidate × scale × replicate × batch; STRUCTURAL is generation artifact × metric family (individual metrics remain in the bundle); AGGREGATION is corpus/branch after all declared dependencies. Generation batches target 5–20 minutes. M3/M4 induction is split by frozen restart/candidate where class semantics allow; one indivisible induction run may remain a long-tail job and is identified in telemetry. M5 rule search is split by validation candidate, never by changing its search budget. F2 extraction splits EDIT and LEXICAL_PARADIGM families.

Scientific reachability is a manifest decision. In validation all diagnostic stages execute after a fitted candidate even if a gate fails. Where later target policy skips a stage, a tiny reachability job emits every planned NOT_REACHED record. Executor timeout or worker loss can only retry, never create a scientific status.

## Resume, retry, and duplicates

Verified result hashes are atomically indexed by JobID. Restart reconstructs readiness from this index and never schedules a completed verified job. Infrastructure failures expire/reissue the identical bundle; scientific failures are published once and never retried merely to seek a better outcome. Deliberate duplicates are allowed. Equal canonical scientific hashes are counted once and retain all provenance; disagreement quarantines all copies and blocks descendants and aggregation.

Canonical scientific JSON uses sorted keys, UTF-8 NFC strings, integers where exact, and decimal IEEE-754 values rendered with the frozen shortest-roundtrip algorithm. Metrics designated numerically equivalent must declare absolute/relative tolerance in the frozen manifest before runs; otherwise byte identity is required. No tolerance may be inferred from observed disagreements.

## Telemetry and scaling

Per attempt store worker certificate identity, UTC start/end, wall/CPU seconds, peak RSS, retries, infrastructure status, lease history, and transfer bytes. This record is outside the scientific payload hash. Runtime prediction by class/scale/stage influences queue priority only.

The Stage-A representative benchmark runs the same manifest with 1, 2, and 4 workers (8 where available). At four workers require parallel efficiency ≥0.60, utilization ≥0.70 while enough ready work exists, scheduler/transfer overhead ≤10% CPU-equivalent, and straggler contribution ≤25% wall time. A lower result blocks distributed validation; it cannot change scientific limits or thresholds.

Failure tests interrupt worker and coordinator, change worker count, add a worker, retry a lease, submit identical/conflicting duplicates, corrupt bytes, omit an artifact, and use a wrong executable hash. Recovery must preserve JobID. Aggregation refuses any unresolved conflict, missing required job, hash/schema/dependency mismatch, wrong seed/config/code revision, incomplete evidence, or post-freeze configuration.

## Capacity model

Task86C used 90.1 CPU-hours for 672 corpus jobs/4032 model executions and about 12 hours on two nodes. The planned design has 144 blind synthetic corpus instances (two generators × six classes × three scales × four replicates), 36 natural instances, and 36 development/preflight instances: 216 corpus instances. With about 12 validation candidates per class this is approximately 15,552 candidate fits, up to 5,184 generation batches, and 36,288 family-scale F2 evaluations. Allowing 3.5× Task86C compute plus verifier/duplicate overhead gives 330–380 CPU-hours and 60–100 GB immutable storage.

With a worker node providing about 16 effective slots, conservative wall estimates are 30h, 16h, 9h, and 5.5h at 1, 2, 4, and 8 nodes. These are capacity estimates, not acceptance thresholds; Stage A replaces them with measured class/stage runtime models. M3/M4 induction, M5 search, large generation batches, and edit-graph extraction are expected stragglers.
