# Task82a corpus-scale scaling design

**Design version:** V1.0. **Authority:** Task81 V1.1 mnemonic mechanism
space, Task82 V1.1 frozen bounded-adapter results, Fingerprint V2 frozen
(`fingerprint-v2-page-hierarchy-v1` / `fingerprint-v2-task79c-closure-v1`).
**Frozen before corpus-scale generation:** yes. Target-blind: no Voynich
data, Voynich reference vector, Voynich statistic, or notation-control
(Task82b/BDD) artifact is read by anything in this design or its runner.

## Scope boundary

Task82a is a scaling layer (`internal/task82a.assembleJob`, called
`CorpusScaleAssembler` per task82a.txt sec.7) over the unmodified Task81
V1.1 mechanisms. It never changes a primitive, operation, carrier,
retrieval rule, observable-serialization contract, historical status, or
frozen parameter value. Every chunk calls
`(mnemonicspace.Runner{}).Prepare`/`Recover` exactly as Task82's own
`runJob` does; the assembler only decides *how many times* and *with what
per-chunk input* those calls happen.

## Eligible mechanisms and local-unit capacity (sec.8-10)

All 16 frozen Task81 V1.1 mechanisms are eligible. Local-unit capacity is
read from each mechanism's own frozen `ParameterSet` field where one
exists; where none exists (F11, rotation-index have no `Capacity`/`Period`
field), capacity is fixed at 4 as a preregistered, target-independent
generic scaling rule matching the sibling F12/storage-associate capacity
already frozen in the same registry -- never derived from Voynich token,
line, or page statistics. The full table (`MechanismScales` in
`internal/task82a/types.go`):

| mechanism_id | parameter_set_id | surface | capacity | source |
| --- | --- | --- | --- | --- |
| f01_speculum_core | f01-core-bounded | LITERAL | 12 | F01Parameters.NumRings |
| f01_speculum_profile_latin23_r12 | f01-profile-latin23-r12 | LITERAL | 12 | F01Parameters.NumRings |
| f08_serpens_core | f08-centre-edge-small | LITERAL | 10 | F08Parameters.Capacity |
| synthetic_literal_storage | f08-centre-edge-small | LITERAL | 10 | shared parameter set |
| negative_randomized_convention | f01-core-bounded | LITERAL | 12 | shared parameter set |
| negative_randomized_path | f08-centre-edge-small | LITERAL | 10 | shared parameter set |
| f11_arismetricum_core | f11-core-default | CUE | 4 | GENERIC_SCALING_POLICY |
| f12_horalogius_core | f12-cycle-default | CUE | 4 | F12Parameters.Period |
| m_restricted_rotation_index | rotation-index-small | CUE | 4 | GENERIC_SCALING_POLICY |
| m_restricted_storage_associate | storage-associate-small | CUE | 4 | StorageAssociateParameters.Capacity |
| synthetic_cyclic_state | f12-cycle-default | CUE | 4 | shared parameter set |
| synthetic_indexed_lookup | f11-core-default | CUE | 4 | GENERIC_SCALING_POLICY |
| synthetic_cue_based | f12-cycle-default | CUE | 4 | shared parameter set |
| synthetic_ambiguous | storage-associate-small | CUE | 4 | shared parameter set |
| negative_randomized_cue_association | f12-cycle-default | CUE | 4 | shared parameter set |
| negative_randomized_index_mapping | f11-core-default | CUE | 4 | GENERIC_SCALING_POLICY |

## Chunking (sec.9-10)

FIXED_CAPACITY, TRUNCATE: the long input is split strictly into
capacity-sized chunks; any incomplete final remainder is dropped. TRUNCATE
was chosen (over PAD/PARTIAL_APPLICATION) because it is the one rule that
is uniformly valid for every mechanism's `Encode` contract without
per-mechanism special-casing (F01 requires <=NumRings letters, F08 accepts
1..Capacity with holes -- TRUNCATE simply never exercises the partial
path). Chunk sizes come only from mechanism capacity or the GENERIC
capacity-4 rule above -- never from Voynich token/line/page lengths.

## State policy (sec.11-12)

`mnemonicspace.Runner.Prepare(spec, params, input, seed)` takes no
prior-state argument: every call is a fresh, deterministic construction
from its own chunk input. **RESET_EACH_CHUNK is therefore the only state
policy that is type-valid** under the frozen mechanism contract;
CONTINUE_STATE is NOT_TYPE_VALID and is not run. Simulating continuation
would require changing what a frozen mechanism's `Prepare` accepts, which
sec.6 forbids. This is a fact about the frozen software contract, not a
historical claim about whether Fontana devices reset between blocks
(sec.12); it is recorded as GENERIC_SCALING_POLICY in
`SCALING_POLICIES.tsv`.

## Cue namespace (sec.15-16)

For the 10 cue-addressable mechanisms, both frozen variants are run and
compared:

- **CUE_RESET_LOCAL_V1**: every chunk reuses cue labels `C0..C(capacity-1)`
  (identical to Task82's own `C0..C3` convention, extended per chunk).
- **CUE_RESET_GLOBAL_V1**: chunk `i`'s local index `j` gets the globally
  unique label `C(i*capacity+j)`, never repeated across the document.

Only the *label string* differs between the two; the mechanism's own
Index/Position/Tick addressing is always local (0..capacity-1) within a
chunk, and the cue itself is never turned into plaintext or given lexical
meaning (sec.16). Literal mechanisms have no cue-namespace axis (one
policy, `LITERAL_RESET_V1`).

A real effect of RESET_EACH_CHUNK interacting with LOCAL_NAMESPACE was
observed directly in early pilot runs: F12 with `CUE_RESET_LOCAL_V1`
degenerates to a single repeated emitted cue across every chunk (fresh
state + identical local input every time -> identical output every time).
This is reported as a genuine corpus-scale finding (an extreme case of
input-insensitivity / repetition), not treated as a bug.

## Convention and path policy (sec.13-14)

Both a convention-reuse axis and a path-reuse axis are formally
preregistered; only one variant of each is included in the main manifest,
for computational-cost reasons decided before generation (`SCALING_POLICIES.tsv`
marks the excluded variants `included=false` with their reason, never
silently dropped):

- **CONVENTION_GLOBAL** (included): recovery-sampling tests reuse one
  deterministic convention/association scheme (same shape as Task82's
  `runJob`) for every sampled chunk. CONVENTION_PER_BLOCK (excluded) would
  regenerate a distinct-but-still-content-independent convention per block;
  it does not change any conclusion this design can support and was
  dropped for job-count cost.
- **PATH_PER_CHUNK_RESTART** (included): `GeometryKnowledge.Path` is
  rebuilt fresh per chunk (`0..capacity-1`), consistent with
  RESET_EACH_CHUNK. PATH_REUSED_GLOBAL (excluded) would contradict the
  state-reset policy and was not run.

## Cue semantics and document assembly (sec.16-22)

Cues remain opaque strings; `InternalMemoryState`/`ConventionKnowledge`
stay separate carriers, never serialized into the observable document. The
assembled `OBSERVABLE_DOCUMENT` is built purely from each chunk's own
`Document.Symbols` -- no plaintext, block ID, hidden state, convention,
path, association, or recovery metadata is included.

Boundary provenance (`BOUNDARY_PROVENANCE.tsv`):

- `local_mechanism_boundary` = one `Runner.Prepare` call = GENERATED_BY_MECHANISM.
- `token_boundary` / `line_boundary` = ASSEMBLER_DEFINED, one line per
  chunk. For **literal** mechanisms a line holds exactly one token: the
  chunk's symbols joined with no separator (so distinct "words" arise
  naturally as chunk content varies -- this is the corpus-scale analogue
  of a language's growing vocabulary). For **cue** mechanisms a line holds
  one token per emitted cue (so LOCAL_NAMESPACE gives a small, bounded
  vocabulary and GLOBAL_NAMESPACE gives one growing with chunk count).
  Collapsing an entire cue chunk into one joined token was tried first and
  rejected: it trivially degenerates LOCAL_NAMESPACE into a single
  repeated word regardless of content, which is an assembler artifact, not
  a property of the mechanism.
- `page_boundary` = NOT_DEFINED always. No synthetic page/folio/section
  structure is introduced anywhere in Task82a.
- `assembly_boundary` (the job's chunk count / corpus_scale) is distinct
  from `local_mechanism_boundary` and `input_boundary`, and is never
  treated as a historical line or page.

## Corpus scale (sec.23-25): the scale convergence pilot

A target-blind pilot (`RunScaleConvergencePilot`, Doyle only, literal
Latin23 stream only) measured plug-in symbol entropy at doubling chunk
counts (16, 32, ..., 4096) with a fixed, preregistered convergence
criterion: entropy changes by <=0.01 bits when the chunk count doubles.
The pilot converged at **256 chunks** (entropy 4.1209 bits vs. 4.1187 at
128 chunks, delta 0.0022; every checkpoint from 256 up stayed within the
same 0.01-bit band). A second, independently target-blind ceiling applies
on top of entropy convergence: a real timing pilot
(`f2_timing_test.go::TestF2Timing`, run on synthetic non-Voynich token
streams before any manifest job) measured frozen F2 extraction cost
directly against vocabulary size for the GLOBAL-cue/literal cost profile:
0.44s / 3.3s / 29.4s / 240s at 64 / 256 / 1024 / 4096 tokens. Scaling past
the entropy convergence point would only multiply job cost across ~470
main-manifest jobs without adding scientific signal the pilot had not
already shown converging, so **256 chunks doubles as both the entropy
convergence point and the computational-cost ceiling** (sec.23:
"estimator convergence" and "computational cost", not Voynich statistics).

Frozen scales (`CorpusScales`): **SMALL=64, MEDIUM=128, LARGE=256** chunks
-- the two pre-convergence doublings plus the convergence point itself, so
the approach to stability is visible without scaling past it.

## Pilot firewall (sec.25)

The pilot changed only the chunk-count checkpoints and read only Doyle's
literal stream. It did not touch mechanisms, parameters, state/cue/path
policy, serialization, or any F2 definition.

## Input corpora and adapters (sec.26-28)

Same three natural-language controls and provenance rule as Task82
(`data_test/pg2097-2.txt`, `data_test/pg30795-mod.txt`,
`data_test/astafiev-1000-culinar-receipts-utf8.txt`; SHA-256, byte size,
Unicode-letter count recorded, source bytes never copied into results).
The literal/cue-item mapping rule is mechanically identical to Task82's
bounded adapter (Unicode letters -> `upper(codepoint mod 23)` over the
frozen Latin23 alphabet; whitespace-delimited word tokens ->
`"I"+SHA-256(lower(token))[:16 hex]`); only the *length* read from each
corpus differs, and that length is matched across all three corpora
(`requiredLengths()`): the largest frozen `corpus_scale` (256 chunks) times
the largest literal capacity (12) for the literal stream, and 256 times
the cue capacity (4) for the item stream, applied identically regardless
of corpus. No corpus is allowed to produce more chunks than another.

## Seeds and replicates (sec.29-30)

`MasterSeed = 82024001` (Task81/82's own `81024001`/`80024001`-style
numbering convention). Per-job seed is the first 64 bits of
`SHA-256(MasterSeed|mechanism_id|scaling_policy_id|corpus_id|corpus_scale|mechanism_version|replicate)`.
Two replicates per cell, matching Task82's own frozen replicate count.
Per-chunk seeds are `SHA-256(job_seed|chunk_index)` -- deterministic,
content-independent. Job identity
(`mechanism_id, mechanism_version, scaling_policy_id, corpus_id,
corpus_scale, seed, replicate, ablation/control status`) hashes to a
16-byte hex job ID.

## Recovery sampling (sec.35-37)

Local recovery replication (all seven frozen R0-R6 conditions, same
candidate/environment/negative-control-corruption construction as Task82's
`runJob`) runs on a deterministic, preregistered stride sample of at most
8 chunk indices per job (always including chunk 0), fixed before any job
ran (`sampleIndices`). DOCUMENT_RECOVERY is reported per job/condition as
the mean of LOCAL_RECOVERY over the sampled chunks -- explicitly not
assumed equal to concatenation-implies-exact-document-recovery (sec.36).

## F2 extraction scope (sec.46-52)

F2 runs strictly as `FEATURE_EXTRACTION_ONLY`: every
`fingerprintv2.Config` Task82a builds sets only `Primary` (the assembled,
non-Voynich corpus file, `GlyphMode: "natural"`, no `IVTFFPath`), leaves
`Controls` empty, and is rejected outright by `assertNoVoynichPath` if the
path contains "voynich", "zl3b", "it2a", or an IVTFF-shaped path fragment.
Two attempted metric groups, chosen because they are computable from the
frozen extractor's base pass without fingerprint v2's `Task79Config`
pipeline:

- `EF1/EF2/EF3` (edit-family) and `LP1/LP4` (lexical paradigm): computed
  by `fingerprintv2.Run`'s base `Metrics` struct from vocabulary alone, no
  manuscript hierarchy needed.
- `cs1..cs7` (task77 cross-scale): computed by the default
  `CrossScaleConfig` from token adjacency/line index; empirically, `cs1`
  and `cs2` come back `INCONCLUSIVE` (too few family-bearing pairs at this
  vocabulary scale) and `cs3/cs4/cs5` come back `NOT_APPLICABLE`
  (genuinely require `ivtff_path` locus/Currier/section metadata Task82a
  never has), while `cs6/cs7` do get computed with a real observed
  statistic and a `SUPPORTED`/`NOT_SUPPORTED` significance verdict against
  their registered null model (`Status` here is a significance verdict on
  an available result, not an availability flag -- confirmed by reading
  `internal/fingerprintv2/crossscale_pipeline.go`'s post-FDR
  classification loop).

**Deliberately not attempted** (`NotAttemptedFamilies`): the
hierarchy/folio/locus/line-profile families gated behind
`fingerprintv2.Task79Config` (`2DL`, `BP`, `HR`, `LC`, `LS`, `PF`). This is
a cost-driven, target-blind scope decision made before generation, marked
`NOT_ATTEMPTED_COST_BOUNDED` in every report/coverage artifact -- distinct
from the frozen extractor's own `NOT_APPLICABLE`/`INCONCLUSIVE` results,
which are a property of the data, not of what Task82a chose to run.
`Task79Config`'s permutation/bootstrap cost (1000 permutations x 1000
bootstrap replicates, designed for real IVTFF-scale manuscripts) is
incompatible with running it across ~470 corpus-scale/seed/policy/scale
cells in this design's cost budget, and its `correction_layer` field
exists specifically for Voynich IVTFF metadata corrections that have no
meaning for a synthetic, non-manuscript corpus.

`Repetitions=30` (reduced from the canonical Task79 control's 1000) is
used for every F2 extraction: this affects only null-distribution
precision (wider confidence intervals on `cs6`/`cs7`'s significance
verdicts), never a metric definition, weight, normalization, or inclusion
decision (sec.52). The reduction is target-blind and cost-driven, measured
and fixed by the timing pilot before any manifest job ran.

## F2 coverage / comparison eligibility gate (sec.49-50)

No numeric CORE-coverage threshold exists in `F2_METRIC_REGISTRY_FINAL.tsv`
or the F2 freeze documents (confirmed by reading them); Task82a therefore
defines its own target-blind threshold, fixed here before any Voynich
vector is opened by Task83: group the 13 CORE metrics by family (`2DL, BP,
EF, HR, LC, LS, PF` -- 7 families). For a mechanism,

- **F2_COMPARABLE** if CORE-family coverage ratio (families with >=1
  available CORE metric / 7) >= 0.5;
- **PARTIALLY_COMPARABLE** if 0 < ratio < 0.5;
- **NOT_COMPARABLE** if ratio == 0.

Because only the `EF` family is attempted at all (`LP`/`cs*` are
SUPPORTING, not CORE), the maximum achievable ratio in this design is
1/7 ~ 0.14, so **no mechanism can reach F2_COMPARABLE under this
portfolio** -- every mechanism with a non-degenerate edit graph lands at
PARTIALLY_COMPARABLE, and NOT_COMPARABLE only for mechanisms whose
observable vocabulary collapses to <2 types (LOCAL_NAMESPACE-degenerate
jobs; see the state-policy note above). This ceiling is a property of the
scope decision above, not of the mechanisms, and is stated plainly in the
report rather than treated as a surprising negative result.

## No F2 weight tuning (sec.52)

Family weights, normalization, metric inclusion, distance definition, and
Pareto definition are never touched. Task82a only ever reads
`fingerprintv2.Run`'s output; it does not call any weighting/aggregation
code from `internal/fingerprintv2`.

## Generic and negative controls (sec.58-59)

All five generic controls (`synthetic_literal_storage`,
`synthetic_cyclic_state`, `synthetic_indexed_lookup`,
`synthetic_cue_based`, `synthetic_ambiguous`) and all four negative
controls (`negative_randomized_convention`, `negative_randomized_path`,
`negative_randomized_cue_association`, `negative_randomized_index_mapping`)
are included in the main manifest at full scale/policy/corpus/replicate
coverage, identically to the 7 positive/M-restricted mechanisms -- no
subsetting was needed since assembly cost is negligible (only F2
extraction cost is scale-sensitive, and control mechanisms are no more
expensive to assemble than their positive counterparts).

## Job identity and manifest (sec.30-31)

`ManifestJob = {job_id, mechanism_id, mechanism_version, parameter_set_id,
scaling_policy_id, input_corpus_id, corpus_scale, chunks, seed, replicate,
ablation_control_status}`. `BuildManifest()` enumerates mechanism x its
frozen scaling policies (1 for literal, 2 for cue) x 3 corpora x 3
corpus_scales x 2 replicates = **468 jobs**, deterministically sorted by
job ID. `TASK82A_BLIND_MANIFEST.json` is written once
(`-gen-manifest`) and re-verified byte-for-byte (via a fresh
`BuildManifest()` regeneration, not just a checksum) before every
subsequent run; any drift is a hard `FREEZE_MISMATCH`.

## Bug policy (sec.67)

Same as Task82: an implementation bug that changes only code while frozen
semantics are unchanged is documented, versioned, and its jobs
regenerated; a bug that would require changing this design after freeze
forces `TASK82A_SCALING_NOT_READY` and a new design version.
