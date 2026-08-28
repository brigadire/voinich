# Task86C-v2 scientific contract ambiguity audit

## Classification

`SCIENTIFIC_CONTRACT_AMBIGUITY`

Task86C-v2-scientific-impl section 6 requires an immediate stop when Task85b
does not permit a unique implementation.  The following are independent
blocking ambiguities; any one is sufficient to prevent a production freeze.

## Blocking findings

| ID | Frozen source | Exact ambiguity | Why implementations can differ scientifically | Required frozen resolution |
|---|---|---|---|---|
| A01 | `G1V2_MODEL_REGISTRY.tsv`, M0 | M0 is described as IID over token glyphs with start/stop and “validation-selected smoothing”, but no probability equation, smoothing grid, validation objective, vocabulary/UNK rule, boundary count, generation rule, or complexity formula is given. | IID glyph MLE, additive smoothing, backoff, and the old Task85a token-categorical model yield different likelihoods and samples. | Freeze equations, grid, selection/ties, boundary/vocabulary rules, generation and complexity test vectors. |
| A02 | `G1V2_MODEL_REGISTRY.tsv`, M1 | Only orders 1..3 and “smoothing frozen before HELDOUT” are stated. Context definition, smoothing family/grid, boundary symbols, unseen-context recursion, validation selection and ties are absent. | Multiple standard Markov estimators satisfy the prose and produce different PM evidence. The older Task85a grid has orders 1..6 and therefore cannot be copied unchanged. | Freeze the full M1 estimator and candidate/selection contract. |
| A03 | `G1V2_MODEL_REGISTRY.tsv`, M2 | “variable-memory suffix-context” and validation-selected depth/pruning do not identify the induction algorithm, grid, escape estimator, pruning test, maximum memory, stopping rule, or ties. | PPM-C, PST and context-tree weighting are non-equivalent compliant readings. | Freeze one exact algorithm, finite grid, selection/stopping rules and golden fixtures. |
| A04 | `G1V2_MODEL_REGISTRY.tsv`, M3; `G1V1_G1V2_CHANGE_REGISTER.tsv` CH09 | The new “complete bounded exact path” and “preregistered approximate inducer” are not defined. No exact search, bound, approximate merge algorithm, normalized merge score, divergence convention, candidate order, operation accounting or class-combination rule exists. | Task85b explicitly prevents the older blue-fringe result from standing for the class; restoring it would violate CH09. Many exact/approximate DFA induction procedures fit the label. | Freeze both algorithms, bounds, candidate grid, route aggregation, complexity and failure semantics. |
| A05 | `G1V2_MODEL_REGISTRY.tsv`, M4 | State-count selection, EM restarts and convergence evidence are named, but topology, state-count grid, initialization, E/M equations, smoothing, restart count/seeds, likelihood objective, tolerance, iteration cap and tie selection are missing. | Distinct PFSA/HMM topologies and EM implementations have different fitted distributions and failures. The older Task85a M4 explicitly used no EM. | Freeze topology and complete numerical/selection protocol with deterministic vectors. |
| A06 | `G1V2_MODEL_REGISTRY.tsv`, M5; change register CH11 | “productive component backoff” replaces zero mass but its component inventory, backoff distribution/support, mixture weight, induction/search grid, unseen-form probability, generation and complexity cost are undefined. | Plausible productive backoffs assign different probability and change PM4/PM6, generation, and minimality. | Freeze the complete grammar/backoff equations, grids and complexity coding. |
| A07 | change register CH15; `PM6_NEGATIVE_PROTOCOL.md` | Task85b requires counter-based namespaced seeds but gives no global derivation, counter PRNG, byte encoding, domain registry or golden vectors. PM6 supplies a namespace string only. | Reusing Task85a's PCG stream, using a counter hash, or using another counter PRNG all satisfy different portions of the available text and yield different samples. | Freeze one RNG/seed specification for every fit, selection, generation, bootstrap, permutation and control domain, with vectors. |
| A08 | `PM_REDESIGN.md`; predictive registries | B1/B2 construction, M0's special self-baseline null, the development-null resampling unit, multiplicity family, quantile convention and most practical-effect floors are not executable definitions. | Different legal nulls and multiplicity corrections change gates. | Freeze baseline algorithms and a deterministic threshold-derivation schema/procedure. |
| A09 | `G1V2_METRIC_REGISTRY.tsv`, PM5 | “weighted adaptive-bin ECE with bins fixed on DEVELOPMENT” lacks the adaptive bin construction, number/minimum size of bins, prediction event/label definition, boundary ties and weights. | Common adaptive ECE definitions are not equivalent. | Freeze PM5 bins, event semantics and golden expected records. |
| A10 | `PM6_NEGATIVE_PROTOCOL.md` | Complement construction is defined, but the paired-bootstrap sampling unit/seed, percentile/quantile convention, label-permutation count/seed/ties and exact threshold statistic are not. | Bootstrap LCB and permutation q0.95 can differ at gate boundaries. | Freeze deterministic bootstrap and permutation algorithms and vectors. |
| A11 | `G1V2_STRUCTURAL_GATE_REGISTRY.tsv`; `STRUCTURAL_ADEQUACY_V2.md` | Structural thresholds refer to development MFC q95, practical-effect floors and dispersion q95 without numeric floors or a complete deterministic derivation. “All three scales” is not tied to exact generated token counts in the v2 contract. | Thresholds and generation populations can differ while satisfying the prose. | Freeze the structural threshold artifact recipe, floors, scale/count rule and dispersion estimator. |
| A12 | `G1V2_COMPLEXITY_CONTRACT.tsv` | “model bits” is not encoded for the changed v2 M0–M5 representations, especially M3 exact/approximate routes, M4 probability parameters and M5 productive backoff. | Different prefix codes and parameter precision change pairwise order and equivalence components. | Freeze canonical model serialization and exact bit-cost equations per v2 candidate. |
| A13 | `TASK86C_V2_CONTROL_DESIGN.md`; `G1V2_CONTROL_REGISTRY.tsv` | The contract says “at least two” independently authored generators per class but does not select generator implementations, parameters or seeds. Development controls are generic registry rows, not immutable corpora. | There is no unique blind synthetic population or independently computable expected manifest count. | Freeze exact independent generators, parameter cells, seeds, source hashes and development inputs. |
| A14 | `TASK86C_V2_CONTROL_DESIGN.md` | English, Latin and Sanskrit are named, but exact editions/files, licenses, preprocessing inputs, window seeds and occurrence-window selection are not identified. | Many corpora meet the language labels and yield different inputs and thresholds. | Freeze local immutable source artifacts and complete preprocessing/split vectors. |
| A15 | `STRUCTURAL_ADEQUACY_V2.md` line describing later execution | The reachability policy “may skip” generation after predictive FAIL; it does not freeze whether Task86C-v2 control jobs skip or execute each downstream cell. | Both DAGs conform to the prose but have different JobIDs, evidence and job counts. | Freeze a per-stage/per-control reachability table. |
| A16 | `DISTRIBUTED_EXECUTION_DESIGN.md`; control design | Candidate/job counts are estimates (“about 12 validation candidates per class”, “approximately 15,552”), not an exact candidate grid and DAG cardinality. | An authoritative manifest and independent completeness count cannot be derived. | Publish the exact v2 candidate registry and declarative DAG expansion rules. |
| A17 | evidence artifact registry | Required fields are described, but no complete machine-readable scientific schemas define canonical numeric encoding, status-specific required fields, model serialization, or aggregation records. | Evidence hashes and evidence-only validation can accept different byte representations/closures. | Freeze versioned JSON schemas and canonical examples for every handler output. |

## Conflict with silent v1 inheritance

The older `research/phase3/task85a/G1_EXECUTABLE_CONTRACT.json` cannot resolve
these items automatically.  Among other differences, its M0 is token
categorical rather than glyph IID, its M1 grid is orders 1..6 rather than 1..3,
its M3 is the superseded greedy JS merger, its M4 has no EM, and its M5 assigns
zero probability outside retained rules/exceptions rather than the new
productive backoff.  Task85b supplies no normative merge algorithm for
combining these incompatible contracts.

## Consequences

Because A01--A17 precede scientific execution, it is not possible to truthfully
produce any of the following: M0--M5 handlers, threshold artifact, blind input
closure, escrow, complete handler registry, exact run manifest, production
scientific executable hash, distributed scientific determinism result, or an
execution-ready operator handoff.

The existing `internal/g1v2` package remains the prep engineering executor; its
own package comment says it deliberately contains no fitting, generation,
metric, threshold-derivation, or corpus-reading code.  It is not promoted to a
scientific implementation.

## Firewalls

- No Stage C blind bulk computation was run.
- No Stage D confirmatory natural-language computation was run.
- No blind ground truth was created, opened, or used for tuning.
- No Voynich input was opened or evaluated by G1-v2.
- No historical Task86C-v2-prep provenance was rewritten.

