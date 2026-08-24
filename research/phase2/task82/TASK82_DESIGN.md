# Task82 blind experiment design

**Design version:** V1.0. **Authority:** Task81 V1.1. **Frozen before
evaluation:** yes. The design is target-blind: no Voynich or notation-control
data, feature vectors, results, or reference distributions may be opened by the
Task82 runner.

## Inputs and adapters

The three manifest corpus IDs resolve only to the repository's documented
natural-language controls: `data_test/pg2097-2.txt`,
`data_test/pg30795-mod.txt`, and
`data_test/astafiev-1000-culinar-receipts-utf8.txt`. Source bytes are never
copied into results. SHA-256, byte size, Unicode-letter count, and preprocessing
are recorded. The mechanism-compatible bounded input is fixed before results:

* literal mechanisms receive the first eight Unicode letters, case-folded and
  mapped mechanically by code-point modulo the frozen 23-symbol alphabet;
* cue-addressable mechanisms receive four item IDs (`I` plus the first 16 hex
  digits of SHA-256 of successive Unicode word tokens) at indices/positions
  0--3 and ticks 1--4; visible cues are corpus-independent `C0`--`C3`.

The bounded length is required by the frozen F01/F08/storage capacities. This
adapter adds no lexical or semantic knowledge. It intentionally permits a
direct test of input sensitivity: literal surfaces can preserve corpus-specific
symbols while cue surfaces expose the same mechanism-defined cues.

## Estimands fixed before aggregation

`recovery_score` is 1 for EXACT, the reported fraction for PARTIAL, `1/n` for
an ambiguity set of size n, and 0 for CUE_ONLY or NO_RECOVERY. NOT_APPLICABLE
is missing, never success or failure. `KD_carrier = score(R0)-score(Rx)`.
Carrier necessity is NECESSARY for a positive delta in every corpus/replicate,
CONDITIONALLY_NECESSARY for a positive delta in some, REDUNDANT for zero delta
throughout, and NOT_APPLICABLE when every removal result has that class.

Entropy values are deterministic empirical plug-in descriptors of the bounded
symbol streams, not population Shannon quantities. `H(M|X)` is reported only
as the log2 ambiguity-cardinality recovery proxy and is explicitly labelled as
such. Collision groups use observable checksum and require distinct input IDs.

Input sensitivity is INPUT_SENSITIVE when all three corpus output checksums
differ, PARTIALLY_INPUT_SENSITIVE when two differ, otherwise INPUT_INSENSITIVE.
Replicate stability is STABLE when all scalar metric ranges are <=0.01,
PARTIALLY_STABLE when <=0.10, otherwise UNSTABLE. Parameter sensitivity is the
observed range; there is no fitting or threshold search. Variance components
are unweighted descriptive between-group variances.

EM classes are applied in order: EM4 if R0 has no intended recovery; EM0 if R6
is exact; EM3 if R6 is ambiguous and knowledge reduces ambiguity; EM2 if an
internal-memory/context carrier raises recovery; otherwise EM1 when a specific
convention/geometry carrier raises recovery. Continuous values remain primary.

The manifest's condition-specific seeds are authoritative. Consequently F01
R0--R6 contrasts compare independently seeded states; this frozen-design
limitation is carried into all KD conclusions. No graded carrier degradation or
64-subset minimal-set search is added. Generic F2 extraction is not run because
Task81 froze compatibility but did not preregister a Task82 extractor invocation.

## Execution and gates

Jobs are immutable manifest rows. The runner supports local execution,
`-shard-index/-shard-count`, resume, `-verify-only`, and deterministic
merge/report. Raw JSON contains separate input summary, E/G/H/K/I/C summaries,
observable document, information trace, recovery result, metrics, warnings,
runtime, and checksums. It never serializes plaintext or hidden carrier values.

Freeze mismatch, job-ID mismatch, leakage, unresolved implementation/resource
failure, incomplete ledger, checksum failure, nondeterminism, or failed exact
storage/control invariants prevents a successful results marker. Scientific
failure remains a result. Exactly one final marker is emitted.
