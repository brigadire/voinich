# Performance & Architecture Audit

Scope: all executable analysis CLIs at repo root and their `internal/*`
packages, per `tasks/task27.txt` Phase 1. This audit is descriptive only —
no algorithms, statistics, or output schemas were changed to produce it.

Corpus scale for context (`data_work/ZL3b-x7.txt`): ~39,026 tokens total,
~8,363 distinct token types, ~539 distinct tokens at frequency ≥ 10, ~205 at
frequency ≥ 30.

Reference optimization already landed: `internal/transitionnetwork/permworkspace.go`
(see `ce4c445`). Pattern used there — precompute a dense, vocab-indexed
"workspace" once, replace per-replicate `map[string]...` with `[]int32`/`[]float64`
slices, hoist everything invariant across replicates out of the loop, reuse
scratch buffers — is the template most rows below point back to.

## Methodology

Each internal package plus its CLI driver was read directly (no execution)
by an independent reviewer focused only on: hot-loop asymptotic cost,
allocation inside replicate loops, string-keyed maps in hot paths,
recomputed-but-invariant quantities, dense-integer feasibility, workspace/
buffer-reuse opportunities, and wasted/discarded computation. Reviewers were
explicitly instructed not to invent problems in genuinely efficient code.

## Summary table

| CLI/package | hot stage | estimated complexity | allocation risk | string-map risk | recomputed invariants | optimization opportunity | correctness risk | priority |
|---|---|---|---|---|---|---|---|---|
| conditional-regime-analyze / conditionalregime | `residualGlobalCorrection`→`residualNullMax`→`fitResidualClustering` (nullmodels.go, residualsweep.go, residual.go:222) | O(Permutations(1000)×scales(5)×K-range(14)×fitCap²(200²)) | High — `map[string]bool`/`vector`/`Profile` maps rebuilt per window every replicate | High — `vector`/`Profile` = `map[string]float64` throughout | Sample/distance matrix (residual.go:222-229) doesn't depend on K but is rebuilt once per K (14×) inside the (scale,replicate) loop — ~14x redundant | permworkspace-style dense vocab-index residual vectors; compute distance matrix once per (scale,replicate), reuse across K | Low-Medium — must preserve leave-one-block-out fold logic exactly; existing map-iteration-before-accumulation convention must carry over | **High** |
| residual-diagnostic-analyze / residualdiagnostic | block-permutation classification (run.go:105) + `blockAssociationRows` (diagnostics.go:313) | O(targets(3)×models(2)×Permutations(1000)) + O(Permutations) | Medium — `blockPermutation` allocates a fresh `map[string][]int` grouping every call | Medium — `sparse = map[string]float64` | Block→index grouping is identical across ~7000 permutation calls, rebuilt every time | Precompute block→indices grouping once, pass into `blockPermutation` | Low — grouping cache doesn't change RNG draw order | Medium |
| positional-continuation-validate / positionalcontinuation | `permutation.go:runPositionalTests`, `stratified.go:runStratifiedPredecessorTest` | O(K·N), K=10000 default, ×2 variables + jackknife per block | High — fresh `[]string`/`[]bool`, `map[string][]int` byBlock, 3 maps in `mutualInformationBits` **every one of 10000 iterations** | High — `countMap`/`mutualInformationBits`/`entropyBits` all keyed by raw token strings | Per-block grouping, block-ID sort, global entropy/marginal baselines recomputed every replicate despite block membership and `xs` never changing | permworkspace-style dense token index + per-block index lists; shuffle a preallocated index buffer, reuse scratch maps | Low — RNG draw order and shuffle-call sequence are seed-sensitive; must preserve exactly | **High** |
| higher-order-sequence-validate / higherorderseq | `cmi.go:runCMI`/`cmiBits`/`permuteWithinBlocks`; `jackknife.go:jackknifeRow` | Per candidate O(K·\|obs\|), K=10000 primary+1000 secondary; jackknife ≈O(B²) per candidate × ~100+ candidates | High — 3 fresh maps in `jointTable` **every permutation iteration**, ×10000×~100 candidates | High — `bNeighbor.left/right`, context counts all raw-string keyed | Block partition + sort recomputed every CMI replicate for a fixed candidate; left/right context counts rescanned twice per candidate | Dense TokenID V×V matrix (permworkspace pattern) + per-candidate reusable workspace holding block partition/marginals/scratch | Low — shuffle order per block is seed/reproducibility-sensitive, same hazard class as positionalcontinuation | **High** |
| token-relation-validate / tokenrelationvalidation | `profilePermutations`/`refineProfilePermutations`→`buildLocalProfiles`; `directionPermutations`→`directionScoresAll` | O(K·B·L·D), K=1000+10000(gated) | High — `map[string]Profile` + nested `map[string][2][]map[string]int` rebuilt per block per rep | High — every hot key is a raw token string | `directionScoresAll`'s `edges map[Pair][]directedRef` rebuilt every replicate though candidates/maxD never change | Dense vocab index; dense arrays keyed by (tokenIdx,side,distance); hoist `edges` map construction out of the loop | Medium — permuted content genuinely changes per rep (not wasted work), but float op order must stay reproducible to ULP tolerance | **High** |
| replicated-local-structure-audit / replicatedlocalaudit | `RunAndWrite` stage 5, `markovBlocks`→`buildMarkov` (markov.go) | O(K·H·N), K=1000 default, but per-(H) model is **replicate-invariant** | High — `buildMarkov` allocates nested `map[string]map[string]int` + sorts **fresh inside the K-loop** | High — nested `map[string][]string`/`map[string][]int` transition tables | `buildMarkov(train)` for a given held-out block is identical across all 1000 replicates — recomputed 1000× instead of once | Build each held-out block's Markov model **once** (H models total), cache, reuse across all K draws — removes ~99.9% of stage-5 work; jackknife `nullMean` resum (O(D·R²·P)→O(D·R·P)) is a second, smaller win | Low-Medium — cached models are byte-identical by construction (training-set membership doesn't depend on rep index); per-rep RNG draw order/count must stay identical | **High** |
| structural-reliability / structuralreliability | `runBootstrap` (pairs.go:110), `buildTokenMetrics` (tokenstability.go:29) | O(K·(N+P·log V)) bootstrap, K=200; O(T·F²·E²) neighbor search | High — `profilestability.BuildProfiles` rebuilt per bootstrap replicate | High — shared `Compare`/`cosine`/`positionJSD` sort map keys **on every call**, invoked millions of times | Eligibility/vocab structure invariant across bootstrap runs but profiles fully rebuilt each time; full O(E²) neighbor search redone per threshold | Dense TokenID workspace for Profile counts + precomputed sorted key index reused across `Compare` calls (shared fix benefits 3 CLIs, see below) | Low-Medium — must keep sorted iteration for byte-identical YAML | **High** |
| structural-projection-analyze / structuralprojection | random/smoothing control loop (analyze.go:143); `familyAnalysis`/`matchedGroup` (extended.go:70,189) | O(R·V) projection rebuild + O(R·V²) worst case, R=200; O(families·D·200), D=20 | High — `Projection` map rebuilt wholesale every trial; per-trial caches reallocated ×200 | High — nearly all state is string-keyed maps, `x+"\x1f"+y` concatenated keys | `matchedGroup` doesn't depend on distance `d` but is called fresh inside the `for d:=0..MaxDistance` loop — **20x redundant per family**; frequency `bins` rebuilt every trial despite being trial-invariant | Hoist `matchedGroup` out of the distance loop (single biggest win); precompute frequency bins once; dense-ID projection | Medium — RNG draw order must stay identical; **CLI has profiling flags wired but has never actually been profiled** (`profiles/` only has `transition.*`) | **High** |

**Update (post-profiling, see `PERFORMANCE_REFACTOR_REPORT.md` item 3):**
both hypotheses above were confirmed and hoisted, but profiling revealed the
row's own "dense-ID projection" opportunity is not optional polish — it is
now this analyzer's dominant cost. `GenericSmoothing` alone accounts for
~64% of all allocated bytes and ~35% of CPU time (both essentially
unaffected by the frequency-bins hoist), from an O(V²)-per-call pattern
where nearly the entire ~8,363-token vocabulary is copied and shuffled for
every single token processed. Four separate pre-existing same-seed
nondeterminism bugs (`RandomizeProjection`, `normalize`, `ProjectDistribution`,
`metricsFloat` — all map-iteration-order float/RNG dependencies) were also
found and fixed here, unrelated to performance but required before any
correctness comparison of this row's optimizations was possible. A full
production run (`-random-projections 200`) was not completed during this
task (killed after 19 minutes with memory still climbing) — treat this
analyzer's priority as higher than the table above implies.

**Second update (GenericSmoothing fixed, see `PERFORMANCE_REFACTOR_REPORT.md`,
"`GenericSmoothing` buffer-reuse optimization"):** dependency analysis
proved the O(V²) *touch* count is inherent (a single shared RNG stream is
consumed sequentially across every bin of every token, so skipping any
shuffle would desynchronize all subsequent tokens' randomness for a given
seed) but the *allocation* backing it was not — reusing preallocated
buffers instead of allocating fresh ones per token dropped allocated bytes
at the real vocabulary size by ~500x (2.4GB → 4.7MB per call) with no
dense-ID rewrite needed. `GenericSmoothing`'s own flat allocation is no
longer a pathological outlier; the remaining cost is `normalize`/
`metricsFloat`/`countsFloat`, already-flagged shared primitives used
throughout the package, not specific to this row. One further pre-existing
nondeterminism bug (`transitions()`'s unstable tie-break, unrelated to
`GenericSmoothing`) was found via the required old-vs-new comparison and
fixed. **Production-scale (`-random-projections 200`) run completed successfully
on two machines**: 2h59m-3h25m wall time, 14.82-14.90GB peak RSS (nearly
identical across machines — memory is now a stable, bounded property of
the workload, not a runaway). This is well within this project's own
established range for expensive stages (previously 3-16 hours) — versus
the pre-optimization state, which did not complete even after 19 minutes,
with RSS past 10GB and still climbing. All three stop-condition criteria
(reasonable runtime, bounded memory, `GenericSmoothing` no longer
pathological) are met — `structuralprojection` optimization stopped here
per this task's own instruction; task27 resumed at `metadatavalidation`.
See `PERFORMANCE_REFACTOR_REPORT.md` for full before/after profiles,
scaling benchmarks, and both machines' exact figures.
| structural-profile-stability / profilestability | `runBootstrap` (analysis.go:165) | O(K·(N+P·log V)), K=200 | Medium-High — same `BuildProfiles` allocation pattern (shared code) | High — same shared `Compare`/`cosine`/`positionJSD` cost | None beyond the shared bootstrap/Compare pattern; folds/neighbor tables already precomputed once and reused correctly | Same dense-ID `Compare`/`Profile` refactor as structuralreliability benefits this package for free (it's the canonical shared implementation) | Low — this is the implementation other packages defer to; needs the broadest test coverage before touching | Medium |
| cluster-metadata-global / clustermetadataglobal | `RunSearchSpace`→`computeGrid`→`fastMetrics` (compute.go:77,183) | O(P·kinds·scopes·methods·windowSizes·Ksweep) ≈ 12.6M combo evals, P=10000 | Low — fixed-size stack arrays, cumulative buffers already reused | Medium — `fs.Combos[comboKey{ws,method,k}]` map looked up ~12.6M times in the hottest loop | None found — already well-structured | Replace `fs.Combos` map with a flat slice indexed by precomputed method/window/K indices | Low — deterministic index mapping | Medium |
| global-regime-analyze / globalregime | `expandLabels`+`jsDistance` over full window set × K-sweep (cluster.go, core.go) | Several million `jsDistance` calls over `map[string]float64` window profiles, K=2..15 × 2 methods × ~7,800 windows | High — every window/centroid profile is a fresh `map[string]float64`; no reuse across 119 K-sweep values | High — `profile = map[string]float64` used throughout the hottest loop | Window distributions invariant across the K-sweep but re-iterated via map for every K/method combo instead of once | Dense corpus-wide TokenID vocabulary; window/centroid profiles as `[]float64` arrays | Medium — Go map iteration order is randomized, so current summation order already varies run-to-run; dense rewrite needs a tolerance oracle, not byte-identical, per [[feedback_go_map_iteration_determinism]] | **High** |
| local-regime-analyze / localregime | `buildOffsetCounts`/`offsetProfile` sweep; `matchedExpected`; `positions()` inside per-distance loops | `buildOffsetCounts`: O(occ·2·maxRadius); sweep ≈2100 range-merges; `positions()` O(N=39k) called once per distance **per pair×token×corpus-variant** (~5,600 redundant scans where ~280 would do) | High — up to 1000 fresh maps per token in `buildOffsetCounts`; new map on every `offsetProfile` call (~2100+) | High — `profile`/`offsetCounts` pervasive `map[string]...`/`map[int]map[string]int` | `positions(c,t)` doesn't depend on distance `d` but is recomputed inside the `d:=1..MaxDistance` loop | Hoist `positions()` out of the distance loop (trivial, safe — discarded `oi` confirms reordering is already safe); dense TokenID×offset cumulative-sum table for `buildOffsetCounts` | Medium — same map-order/float-accumulation caveat as globalregime | **High** |
| metadata-validate / metadatavalidation | `ValidateBoundaries` + `clusterPermutationSummary` | O(support(3)×kind(6)×tol(5)×K(10000)×N) dominated by `rng.Perm(N-1)` full O(N) shuffle | High — full-array permute+discard-to-count(≤10000×90 calls) | Medium — fresh contingency maps per K per permutation | `UniformBoundaries`/`CircularShiftBoundaries` redrawn independently per tolerance though the draw doesn't depend on tolerance — 5x redundant | Reorder loops to reuse one draw across all 5 tolerances; partial/reservoir sample of `count` items instead of full `rng.Perm(n-1)` | Low, if RNG draw order preserved | **High** |

**Update (post-profiling, see `PERFORMANCE_REFACTOR_REPORT.md` item 4):**
both "reorder loops" and "partial/reservoir sample" ideas above were
investigated and **rejected** — Go's `rand.Perm` is a self-referential
Fisher-Yates where any later step can still write an earlier position, so
there is no way to compute only the first `count` values without consuming
the full sequence of RNG draws, and sharing one draw across the 5
tolerances would change which specific values get computed for a given
seed (a real methodology change, not a hoist). What *was* safe: reusing a
preallocated, never-reset scratch buffer across all ~900,000
`UniformBoundaries` calls, since the algorithm provably overwrites every
position before it's ever read — cut total allocation by ~4.6x
(`math/rand.(*Rand).Perm`, 72% of allocated bytes before, no longer
appears in the after-profile at all) with zero algorithm change. A second,
smaller, confirmed-safe hoist (`clusterPermutationSummary`'s per-k string
conversion, invariant across the replicate loop) was also done. Along the
way, three more pre-existing same-seed nondeterminism bugs were found and
fixed in `AssociationMetrics`/`entropyCounts`/`conditionalEntropy` (all
summing over unsorted map keys) — significant because `AssociationMetrics`
is exported and also used by `internal/conditionalregime` and
`internal/residualdiagnostic`. Verified via 3 independent full-production
runs (`-permutations 10000`): SHA256-identical.
| distance-context-analyze / distancecontext | `baselines()`, `matchedCohesions` | O(T²×maxD) pairwise (T≈205,maxD=20→~400k calls); +O(maxD×200×\|family\|×T) | Medium | Medium-High — fresh `map[string]bool` key-union per call | Per-token/per-distance totals re-walked repeatedly across pairs sharing a token | Precompute per-token/per-distance totals once; merge over pre-sorted keys / shared dense ids | Low — no permutation loop, deterministic pairwise pass | Medium |
| structural-pair-decompose / pairdecomposition | `chooseControls`, `decompose` | O(targets(~50-90)×allPairs), single pass, no replicate loop | Low | Low | None material | Cosmetic only (shared key-union helper) | N/A | Low |
| property-trajectory-analyze / propertytrajectory | `fallbackMatched` (per selected pair, ~40); `buildTrajectoryCache` (×3) | `fallbackMatched`: O(T² log T) per call, T≈539 → ~145k pairs sorted, repeated ~40× (dominant pipeline cost); cache: O(N×maxD×28 properties) | High | Medium — `map[string][]float64` keyed by 28 property names | O(T²) candidate-pair pool + cost-sort is almost entirely target-independent yet regenerated/re-sorted from scratch per target; comparator recomputes `math.Log1p` on every comparison | Precompute one frequency-rank-sorted structure once, targeted lookup per target; Schwartzian-transform the sort; dense `[]float64` by property index | Low-Medium — must preserve RNG draw/shuffle order and map-iteration-independent accumulation | **High** |
| structural-validate / validation | `Run` (F folds × `BuildTrainModel` O(V³) + baselines); `runAblations` | O(F×(V³+RandomBaselines×lines×maxN)) | High — `AnalyzeSequences` allocates `map[string]*sequenceNGram` + rebuilds `tokenKey` strings per n-gram occurrence, ~500+ calls | High — `positionJSD` (training.go:129) allocates a fresh `map[int]bool`+slice **per pair inside an O(V²) loop** | Test-fold tokenization/positions recomputed from scratch on every fold and every random baseline despite fixed TEST set | Dense TokenID once per fold; integer-tuple keys instead of `tokenKey` string concat; hoist `positionJSD`'s position-union out of the O(V²) loop | Low if refactor preserves existing `sort.Strings(keys)` deterministic aggregation | **High** |
| structural-normalize / normalization | `buildModel` complete-link agglomeration (O(clusters³)) | O(thresholds(5)×V³) + O(V²) `pairKey` string builds per merge iteration | High — `pairKey` does string concat on every pair lookup inside the O(clusters²) inner loop, every merge step | High — `pairs map[string]PairMetrics` keyed by concatenated string | Best-pair search redone from scratch every merge iteration (classic O(n³) clustering) though pair metrics are precomputed once | `[2]int` TokenID keys instead of string concat (softstructural already does this with `[2]string`); consider incremental best-pair tracking | Low — same algorithm, different key representation | **High** |
| normalization-compare | per (threshold×RandomBaseline=100): `exec.Command` spawns a **subprocess** (`sequence-analyze`), writes a temp corpus, reads back YAML | O(thresholds×100) full process spawns + disk I/O, each re-tokenizing the entire corpus from scratch | N/A (cross-process) | N/A | Corpus tokenization/n-gram counting fully redone by a fresh OS process every single baseline — zero in-process reuse | Call `sequence-analyze`'s logic as an in-process library function instead of shelling out — eliminates temp-file I/O and process-spawn overhead entirely | Low — same algorithm, just removes the subprocess boundary | **High** |
| sequence-analyze | single-pass n-gram/context counting | O(lines×(maxN-minN)×line_len), genuinely single-pass | Low | Medium — `ngramKey` string-concat per n-gram occurrence | None internal — but re-run hundreds of times by callers above | Dense TokenID + integer-keyed n-gram maps would help modestly; real win is reducing caller call-count | Low | Low (standalone) |
| structural-analyze | `equivalenceRanking` O(V²) all-pairs via shared `profilestability.Compare` | O(V²), single pass, no replicate loop | Medium (shared `Compare` cost) | High (shared) | None — one-shot ranking | Same shared `Compare` fix as structuralreliability applies here for free | Low | Low/Medium |
| soft-structural-space / softstructural | `BuildAll` O(V²) pair generation; `rank()` sorts each token's neighbor list 3× | O(V²) pairs + O(V² log V) from triple-sorting in `rank()` | Medium — `append([]Pair(nil),...)` copy 3× per token | Low — `bootstrap map[[2]string]*float64` already uses array-ish keys | `rank`'s three near-identical sorts of the same slice instead of one sort + derived orderings | Single sort + partial selection instead of 3 full sorts; reuse scratch slice | Low | Medium |
| structural-graphemic / graphemic | single-pass TSV + BFS union-find | O(pairs)+O(V+E), single pass | Low | Low | None | Negligible; already linear | None | Low |
| dict-analyze | single-pass token/transition analysis | O(tokens+transitions) | Low | Low | None | None needed | None | Low |
| begin-end-analyze | `runPermutations`/`enumerateWindowHits`, 100 permutations default | O(Permutations×corpus_tokens) + O(k²) candidate scoring | Medium — new `[]string` + `map[string][]int` positions rebuilt per shuffle | Medium — `scopeSequence.positions` string-keyed | Token IDs are stable but sequences re-tokenized as strings each run | Dense `[]int` token IDs once; shuffle `[]int`, avoid `map[string][]int` | Low — pure perf refactor, seeding preserved | Medium |

*(`transition-network-validate` is excluded — already optimized per `ce4c445`, used here as the reference pattern.)*

## Cross-cutting finding: shared `profilestability.Compare` bottleneck

`internal/profilestability.Compare` / `cosine` / `positionJSD` sort map keys
on every call and are invoked from `structural-reliability`,
`structural-profile-stability`, and `structural-analyze` alike. Per the
README this is deliberately the single canonical similarity implementation —
**do not fork it**; fixing allocation/sorting there once (Phase 3, "reusable
computational primitives") benefits all three callers without touching the
similarity formula itself. This is the highest-leverage single fix in the
whole audit in terms of "packages touched per line changed," but the widest
blast radius, so it needs the broadest test coverage before any change lands.

## Priority ranking (expected total saved pipeline runtime)

**High, in the specific implementation patterns already proven safe by the
`transitionnetwork` reference (invariant-hoisting / cache-the-invariant-model
/ dense-workspace), roughly ordered by expected win size and how low-risk the
fix is:**

1. `replicatedlocalaudit` — cache `buildMarkov` per held-out block outside the
   1000-replicate loop (~99.9% of stage-5 work is currently wasted; lowest
   correctness risk in this batch, closest analogue to the reference case).
2. `normalization-compare` — stop shelling out to a subprocess per baseline;
   call `sequence-analyze` in-process (removes ~100s of process spawns + temp
   file I/O per run; near-zero correctness risk, pure plumbing change).
3. `structuralprojection` — hoist `matchedGroup` out of the per-distance loop
   (20x redundant per family) and precompute frequency bins once; this CLI's
   profiling flags exist but have never been exercised — profile it first.
4. `metadatavalidation` — reuse one permutation draw across the 5 tolerance
   values instead of 5 independent full `rng.Perm(N-1)` draws; consider
   partial/reservoir sampling instead of full-array shuffle+truncate.
5. **Done** — `conditionalregime`: hoisted the (scale,replicate) distance
   matrix out of the K-loop as audited, but profiling then found the actual
   dominant cost was `euclideanDistance` re-sorting its key union on every
   pairwise call (72% of CLI CPU time, 3.5TB cumulative allocation) — fixed
   by sorting each vector's keys once and merge-walking. ~4.15x faster
   end-to-end (`-permutations 10`: 3h08m→45m19s). Also fixed an unrelated,
   out-of-band same-seed nondeterminism bug in Part C's `boundarySignature`
   tie-breaking, found via the before/after output diff. Follow-up done:
   `withinclass.go`'s `fitClustering` (Part A) had the same per-K
   redundancy — hoisted identically (~3.1x faster on the isolated sweep,
   ~5min saved end-to-end at `-permutations 1`, all 12 output files
   SHA256-identical before/after). `globalregime.jsDistance`'s own
   sort-per-call cost remains open, folded into item 7. See
   `PERFORMANCE_REFACTOR_REPORT.md` item 5.
6. **Done** — `positionalcontinuation`, `higherorderseq`,
   `tokenrelationvalidation`: all three got the permworkspace-style
   dense-vocab-index rewrite. `positionalcontinuation`'s
   `runPositionalTests`/`runStratifiedPredecessorTest`: ~16x/~3.8x faster,
   ~196x/~137x less memory; full CLI 1.69s→518ms, all 19 outputs
   SHA256-identical. `higherorderseq`'s `runCMI`: ~21.7x faster, ~950x less
   memory (profiling confirmed the audit's own diagnosis exactly, unlike
   item 5); full CLI 3.98s→1.26s, all 18 outputs identical.
   `tokenrelationvalidation` (1861 candidates, largest and most
   heterogeneous of the three): `directionScoresAll`'s audited edges-hoist
   turned out NOT to be the dominant cost — profiling instead found
   `buildLocalProfiles`'s per-token map skeleton was 93% of all allocated
   objects in a full production run, fixed via a per-block-ID reusable
   skeleton cache; also fixed an out-of-band same-seed nondeterminism bug in
   `jsOverlap` (same map-iteration-order bug class as item 5). See
   `PERFORMANCE_REFACTOR_REPORT.md` item 6.
7. **Done** — `globalregime`, `localregime`: `globalregime`'s
   `distanceMatrix`/`expandLabels` got the sortedProfile merge-walk rewrite
   (~3.4x/~2.5x faster, ~71x/~42x less memory), which the audit's own
   nondeterminism caveat here turned out to be stale for (item 5 already
   fixed `jsDistance`'s determinism, so this rewrite hit bit-identical
   output, not just a tolerance oracle). `localregime` got the audited
   `positions()` hoist (~3.9x faster) plus four found-and-fixed
   nondeterminism bugs (`jsSimilarity`/`weightedOverlap`/`cosine`/
   `concentration`) — whose sort-per-call fix then caused a real,
   profiler-caught performance regression (1m11s→2m16s) before a
   sortedProfile caching pass (mirroring 7a) fixed both correctness and
   performance together, landing at 45.3s (~1.57x faster than the original,
   uncontended). See `PERFORMANCE_REFACTOR_REPORT.md` item 7.
8. **Done** — `validation` (structural-validate): profiling found
   `AnalyzeSequences` (not `positionJSD`) was the real dominant cost (44.6%
   of total CPU); fixed via a partial dense-key rewrite (n-grams only,
   contexts left on the original string key to preserve exact float-sum
   order for `ConditionalEntropy`). Caught a wrong assumption before
   shipping: `normalization.Mapping` assigns synthetic class IDs (e.g.
   "C0001") absent from the original corpus, so a statically-built
   vocabulary would have silently produced false collisions - fixed with a
   lazily-growing vocabulary instead. Full CLI SHA256-identical;
   ~1.09x faster (~1.82x on `AnalyzeSequences` in isolation).
   `normalization`'s audited `pairKey` concern was profiled and found **not
   actionable** at current corpus scale (YAML marshaling dominates a
   1.35s total run, not clustering) - flagged, not fixed. See
   `PERFORMANCE_REFACTOR_REPORT.md` item 8.
9. **Done** — `propertytrajectory`: profiling confirmed `fallbackMatched`'s
   O(T²) pool rebuild + `math.Log1p`-per-comparison sort was 67.5% of
   total CPU (38.4% in `log1p` alone) - fixed via a precomputed pool +
   cached log1p + Schwartzian-transform sort (~15.1x faster on
   `fallbackMatched`, ~2.7x end-to-end). Also found and fixed an
   out-of-band, pre-existing same-seed nondeterminism bug in the shared
   `entropy` primitive, confirmed by running the original binary twice and
   getting two different output hashes. See
   `PERFORMANCE_REFACTOR_REPORT.md` item 9.
10. **Done** — `structuralreliability` / `profilestability`: the shared
    `profilestability.Compare`/`cosine`/`positionJSD` cost (78.5%/69.4%/9.1%
    of total CPU respectively, with `sort.Strings` alone at 40.0%) was fixed
    by caching each profile's sorted context-map keys once (`SortedProfile`/
    `PrecomputeAll`/`CompareSorted`, `Compare` kept as a thin wrapper - no
    forked formula) and threading the precomputed workspace through every
    hot loop that reuses a profile across many `Compare` calls in
    `profilestability`, `structuralreliability`, `softstructural`, and
    `structural-analyze` alike. Full CLI SHA256-identical on all four:
    `structural-reliability` 68.6s→22.8s (~3.0x), `structural-profile-stability`
    90.9s→39.0s (~2.33x), `soft-structural-space` 4.26s→2.68s (~1.59x),
    `structural-analyze` 3.33s→1.73s (~1.93x). See
    `PERFORMANCE_REFACTOR_REPORT.md` item 10.

With item 10 done, every backlog item from this audit (1-10) is now closed.

## Task31: deterministic local parallel execution

`conditional-regime-analyze` now has a bounded `-workers N` goroutine pool
for its independently seeded replicate-index work. Results are restored to
canonical index order before existing floating-point reductions; RNG and
scientific outputs are unchanged. A real-corpus benchmark measured
25.43s/21.96s/20.51s/20.48s/20.25s at workers 1/2/4/8/12 respectively,
with all 19 artifacts SHA256-identical to the frozen pre-change oracle.
The reduced oracle has only four jobs per combination and ~18.9s fixed work,
so its plateau above four workers is expected and is not a production-scale
speedup forecast. See `DISTRIBUTED_EXECUTION_IMPLEMENTATION.md` for profiles,
memory/allocation measurements, resume evidence, and production estimates.

**Medium:** `residualdiagnostic` (cache one grouping map), `clustermetadataglobal`
(flat combo slice instead of struct-keyed map — already close to optimal),
`distancecontext`, `softstructural`, `begin-end-analyze`.

**Low (already efficient single-pass tools, do not touch):** `pairdecomposition`,
`sequence-analyze` (as a library, its own cost is fine — the problem is call
count from callers), `graphemic`, `dict-analyze`, `structural-analyze` (aside
from the shared `Compare` cost).

## Phase 4 note: profiling coverage

Only `transition-network-validate` and `structural-projection-analyze` have
`internal/profiling` wired in today. `structural-projection-analyze` has
never actually been profiled (`profiles/` contains only `transition.*.pprof`).
Wiring the same 3-line pattern (`profiling.RegisterFlags` before
`flag.Parse`, `profiling.Start`/deferred `sess.Stop()` around the run) into
the remaining CLIs, starting with the High-priority list above, is required
before Phase 4/5 work on each one and carries no correctness risk — it is
purely additive and a no-op when the flags are unset.

## task28: pre-production dominant-stage optimization

With this audit's own backlog (items 1-10) fully closed, task28 asked for
one more pre-production pass on the two stages known to dominate end-to-end
runtime. See `PERFORMANCE_REFACTOR_REPORT.md`'s "Pre-production
dominant-stage optimization" section for full detail.

**Phase 1 (`conditional-regime-analyze`) — verified already fixed, no
change made.** task28 restated this audit's own claim that
`fitResidualClustering` rebuilds its distance matrix per K; that was
already fixed by item 5 above (`prepareResidualFit`). Fresh profiling
(`-permutations 5`, real corpus) confirmed the fix holds
(`prepareResidualFit`/`residualDistanceMatrix` = 2.82% of CPU, not the ~14x
higher cost a per-K rebuild would cause) and surfaced two further,
explicitly-not-fixed bottlenecks: `expandResidualLabels`'s map-based
`euclideanDistance` is now the dominant cost (73% of CPU) — genuinely
K-dependent required work, not redundant, but map-lookup-bound; and
`stability.go`'s `heldOutSeparation` calls the unhoisted exported
`globalregime.JSDistance` in an O(heldOut×k) loop (real, but Part A-only,
so it doesn't scale with `-permutations` and stays low-impact at
production scale). Extrapolated production wall time
(`-permutations 1000`): **~41.9 hours**, which is the necessary cost of the
frozen search space, not implementation waste — hence the existing
per-replicate checkpoint/resume mechanism.

**Phase 2 (`structural-projection-analyze`) — fixed the smallest justified
piece; a larger dense-representation rewrite remains deferred.** Fresh
profiling of the current, already-item-3-optimized binary confirmed
`normalize`/`metricsFloat`/`countsFloat` (98%+ of allocation, ~85%+ of CPU
cumulative) as the dominant cost, exactly the functions item 3 flagged as a
deferred follow-up — but a full dense-TokenID rewrite of `Projection` was
rejected as unjustified by the dependency analysis (their hot-path inputs
are degree-bounded, not vocabulary-sized, so a dense whole-vocabulary array
per call would reintroduce an O(V²) cost, the same trap `GenericSmoothing`
already escaped). Instead, `normalize`/`metricsFloat`'s transient per-call
sort scratch was converted to a reused package-level buffer (the same
`clear()`-and-reuse pattern already validated for `GenericSmoothing` in
item 3) — safe because that scratch is never retained in the returned
value. Result: total allocation for a representative run dropped ~47%
(103.4GB to 54.6GB), `metricsFloat` became fully allocation-free (0 B/op at
every benchmarked size) and 17-35% faster in isolation; full-CLI output
confirmed byte-for-byte identical (`diff -rq`) before vs after. A genuinely
new diagnostic was added to make this possible: `internal/profiling` gained
an opt-in `-memstats-interval` live-heap sampler, because the existing
end-of-run `-memprofile` snapshot is empty by the time it's taken (all
trial-loop memory is long collected) and cannot show what a multi-hour
run's memory actually does mid-flight. At reduced scale that sampler showed
the heap spikes once early then plateaus at a small 170-270MB live set for
the rest of the run — most of the previously-reported 14-15GB is Go's
runtime holding onto address space from that one transient peak, not
genuinely live data. A partial (39m40s) run at the real production
configuration refined this picture further before being deliberately
stopped short of completion (see below): at production scale the heap
instead oscillates continuously between ~1-11.5GB rather than settling low,
consistent with the true peak driver being `familyAnalysis`'s *retained*
per-distance candidate cache (untouched by this fix) rather than the
transient scratch this fix removed — so the fix is expected to meaningfully
cut total allocation and modestly help wall time, but not to sharply lower
peak RSS. See `PERFORMANCE_REFACTOR_REPORT.md`'s Phase 2 STEP 10 for the
full reasoning and the explicit decision to stop the production run early
(cost of ~2 more hours vs. evidence already in hand) rather than complete
it.

## Task29: dense conditional-regime distance hot path

Task29 tested the specific follow-up identified above. The Task28 sparse
implementation represented every residual as a `map[string]float64` plus a
sorted key slice. `euclideanDistance` merge-walked the two sorted key slices,
but every visited dimension still required a string comparison and one or two
map lookups. The feature keys are stable for the complete lifetime of one
`residualFitPrep`: the same vectors are used for its one capped pairwise
matrix and for every K in the `2..15` sweep. Distance evaluations are
`S*(S-1)/2` for the matrix (`S <= 200`) plus up to `n*K` in each label
expansion; with all K non-empty, the expansion total is `119*n` distances per
prep. The old distance loop itself allocated nothing, but preparing sorted
keys and rebuilding sparse centroids accounted for 1.29 GiB alloc-space and
about 95,000 allocations in the representative run.

The implementation now constructs one lexicographically sorted
feature-to-index map per prep and converts each selected residual map once to
a `[]float64`. All vectors and centroids share that index. Conversion occurs
before matrix construction and the K loop; the distance loop is now a single
numeric slice pass with zero allocations. Lexicographic indexing preserves
the old summation order; dimensions absent on both sides merely add exact
zero. Fully dense storage was accepted because the observed live-heap result
remained safe: sampled peak `HeapAlloc` fell from 1,458 MiB to 1,019 MiB in
the paired run. Total allocation volume rose 3.0%, from 27.08 to 27.89 GiB,
which is the explicit space-for-time tradeoff.

The paired real-corpus workload (`39,026` tokens, four eligible classes,
production K range, one permutation) was deliberately 10 minutes before the
change and used identical inputs and flags after it. The existing Task28
five-permutation profile remains the larger corroborating baseline. Profiles
are stored as `profiles/conditionalregime.task29.{cpu,mem}.{before,dense}.pprof`.

| Metric | Task28 baseline (Task29 reproduction) | Task29 dense |
|---|---:|---:|
| benchmark wall time | 10m09.727s | 3m48.149s (2.67x) |
| total CPU samples | 632.57s | 250.70s (2.52x) |
| `expandResidualLabels` CPU | 387.79s / 61.30% | 30.73s / 12.26% (12.62x) |
| `euclideanDistance` CPU | 405.65s / 64.13% | 31.95s / 12.74% (12.70x) |
| allocations / allocated bytes | 18.295M / 27.08 GiB | 18.569M / 27.89 GiB |
| sampled peak `HeapAlloc` / `HeapInuse` | 1,458 / 1,465 MiB | 1,019 / 1,025 MiB |
| dense conversion CPU | n/a | 4.22s / 1.68% |
| estimated 1000-permutation runtime | ~41.9h | ~7.3h conservative |
| output equivalence | baseline | byte-identical, all 19 files |

The removed lookup/string machinery represented about 92% of the old
`euclideanDistance` cumulative time (`405.65s -> 31.95s`), so the Task28
map/hash hypothesis was correct. Applying the measured 5.82x speedup of the
permutation-scaling `residualGlobalCorrection` to Task28's 75.2 seconds per
method-replicate gives about 12.9 seconds; 2,000 method-replicates plus the
fixed allowance yields a conservative **~7.3 hours**, saving approximately
**34.6 hours** from the previous 41.9-hour estimate.

The post-change reduced run is dominated overall by
`globalregime.jsDistanceSorted` (108.08 CPU-s cumulative, 43.11%), mostly in
fixed Part A `stabilityForClass`; it does not scale with permutation count.
Inside the production-scaling `residualGlobalCorrection`, the dense
`euclideanDistance` remains the largest flat consumer (10.05 of 26.86 CPU-s,
37.4%), followed by map-based residual construction/
`meanAndVarianceProfiles` (7.58s cumulative, 28.2%). The kernel is now dense
and SIMD-friendly, so SIMD is worth a bounded investigation; GPU offload is
not yet justified because the vectors are modest and transfers/launches
would consume much of the remaining kernel budget. The single next target is
therefore vectorization of the dense distance loop (benchmark first; do not
alter its accumulation semantics). No such follow-up was implemented here.

## Task30: distributed execution feasibility audit

Task30 is a design/audit task, not a performance-optimization task — see
`DISTRIBUTED_EXECUTION_AUDIT.md` for the full result. It re-profiled current
HEAD (`7f70fb5`, this file's own Task29 dense rewrite) and found a
measurement-driven correction worth recording here: Part A's
`withinClassSignificance` (`nullmodels.go:107-124`) also scales with
`-permutations` via the same independently-seeded-per-replicate pattern as
Part B's `residualGlobalCorrection`, and was not separately counted in this
file's `~7.3h` production estimate above. The corrected estimate is
**~8.4-9.1h** — no code changed, this is purely a more complete measurement.
`globalregime.jsDistanceSorted`/`stabilityForClass`'s map-based cost, already
flagged above as the predicted next dominant consumer once the residual path
was fixed, is now confirmed as such by fresh profiling (35.07% cumulative CPU
at reduced scale) — still not fixed, still an open, explicitly-deferred
target, exactly as this file already documents.

## Task32: deterministic local process executor

`conditional-regime-analyze` now also has a bounded `-executor process`
backend: the same Task31 `JobID`/`JobResult` jobs, dispatched to persistent
subprocess workers instead of goroutines, with no scientific/RNG/reduction
change. Measured on the same reduced real-corpus oracle: process workers
1/2/4/8/12 took 26.72s/23.25s/22.07s/24.71s/25.59s wall, all 19 artifacts
SHA256-identical to the goroutine oracle, including both directions of a
real SIGINT-interrupt-then-resume across executor backends and worker
counts. Isolating the pure process-vs-goroutine cost at workers=1 (26.72s
vs. 24.76s) shows an ≈8% fixed overhead from one extra corpus/metadata
parse and process spawn — a one-time cost that amortizes toward zero at
production scale, exactly as this audit's Task30 section predicted
("process-startup cost... milliseconds against a ~13s job"). Validating
this task's own byte-for-byte requirement also surfaced and fixed one
small, pre-existing, Task32-unrelated bug: `report.go` printed its Part B
summary by iterating a `map[string]EmpiricalStats` without sorting keys
first, so the two-line summary could reorder across separate process runs
regardless of executor or worker count (reproduced on unmodified
pre-Task32 code). See `DISTRIBUTED_EXECUTION_IMPLEMENTATION.md` for the
full protocol, benchmark, memory, and failure-mode detail.

## Task33: remote executor

Task33 adds a thin HTTP transport around the same measured job boundary.
Cold traffic is exactly the two hashed input files per worker and warm
traffic stages no input; replicate requests/results are small bounded JSON.
Two physical hosts (Intel i7-8850H and AMD Ryzen 7 5700X, identical
linux/amd64 Go runtime) were subsequently measured. Frozen-oracle remote
wall times for 1/2/4/8/16/32 slots were
26.038/21.854/19.839/19.819/20.695/19.952s, with all 19 outputs exactly
SHA256-identical. Coordinator peak RSS stayed 157.2-159.9 MiB; combined
worker CPU stayed 8.73-9.02s per run; maximum worker peak RSS over the series
was 153.3 MiB (Intel) and 137.7 MiB (AMD). Each warm run transferred 53,924
bytes of job payload and 17,491 bytes of result payload, zero input bytes,
and had zero retries. The measured 0.810s cold stage transferred 2,440,409
bytes. Fixed coordinator work dominates this four-permutation workload, so
four to eight slots plateau and these speedups are not production-scale
extrapolations. Full tables are in `DISTRIBUTED_EXECUTION_IMPLEMENTATION.md`.

## Task42: normalization-compare distribution and persistent worker reconnect

A fresh profile of `normalization-compare` on the real Doyle corpus
(`data_test/pg2097-2.txt`, production `-random-baselines 100`) measured
358.2s wall time, ~615 MB peak RSS, and confirmed this row's original
in-process-vs-subprocess fix (above, and `PERFORMANCE_REFACTOR_REPORT.md`
item 2) left the remaining cost almost entirely inside
`sequenceanalyze.AnalyzeFile`, called once per threshold's structural pass
and up to 100 times per threshold for independent random-baseline trials -
505 calls total on real data, of which 500 (>99%) are order-independent
and already keyed by a pure function of `(base seed, threshold, run
index)` with no shared package-level state. That cleared the >=50%
distribution threshold by a wide margin, so this stage gained a third job
type (`normalization_compare_baseline`) on the existing Task33-40
coordinator/worker/mTLS infrastructure, exactly like Task40 did for
`structural_projection_trial`. Measured with a disposable local mTLS PKI
(the real Doyle production run and worker fleet were active on this
machine and deliberately left undisturbed): local/1/2/5/10 workers ran
358.2s/346.9s/243.3s/161.6s/159.1s, all outputs byte-identical; efficiency
falls off sharply past this development machine's own 12 cores (5->10
workers: 161.6s->159.1s), a single-machine ceiling rather than a
coordinator/protocol limit. Full profile, RNG/dependency audit, and
scaling table are in `NORMALIZATION_COMPARE_DISTRIBUTION_AUDIT.md`.

Separately, Task42 found that a worker's process lifetime had always been
implicitly scoped to one coordinator run (one handshake, one computer-
state build, then leased jobs against that one `ExperimentID` forever) -
functionally correct for a single production run, but it meant an
already-deployed worker fleet could not serve a second experiment, or a
restarted coordinator, without restarting every worker process. This is
an availability/operability fix, not a throughput one - no scientific
computation changed - so it is recorded here only briefly; see
`REMOTE_WORKER_LIFECYCLE.md` for the design (bounded-backoff-with-jitter
reconnect, per-generation state rebuild, permanent-vs-transient failure
classification) and the Ansible role changes it enabled.

## Task47: begin-end-analyze distribution

A fresh profile of `begin-end-analyze` on the real Astafiev corpus
(`data_test/astafiev-1000-culinar-receipts.txt`, production defaults)
measured 15m38.7s wall time (14m39.9s un-profiled), 4.7-5.1 GB coordinator
RSS once distributed (all of it pre-existing peak memory the single-process
implementation always needed - `1,298,460` candidate structs live in
memory at once - not a regression this task introduced), and found
`directedDistance`+`pageBalance` (both called once or twice per candidate
pair, from the stage's one `O(k^2)` double loop) account for 94.43% of
wall time - a candidate-pair batch, not a permutation replicate, is
therefore the distributed unit, with zero RNG dependency (the
`-permutations`-controlled loop this stage also has is under 1% of cost at
the production default). That is easily the largest independent-work
fraction any stage in this repository's distribution series has measured,
yet the resulting scaling study is this series' *weakest* speedup so far -
1.65x/33% efficiency at 5 workers, and 10 workers measured *worse* than 5 -
because this is the first stage whose per-job wire payload
(`~2.7 KB`/candidate, driven by nested histogram/window tables) is large
relative to its own compute cost, making it transport/serialization-bound
once distributed rather than compute-bound. Real-corpus testing also
caught two genuine wire-layer bugs no small-fixture test had reason to
find: `encoding/json` silently corrupts non-UTF-8 corpus token bytes on
marshal (this is the first workload whose result payload is verbatim token
text), and the pre-existing 1 MiB `maxRemoteMessageBytes` cap was far too
small for this workload's ~5.5 MB production batch payload, causing a
silent multi-minute hang rather than a loud error (raised to 32 MiB). Full
profile, RNG/dependency audit, both bug root-causes, and the granularity +
scaling studies are in `BEGIN_END_ANALYZE_DISTRIBUTED_AUDIT.md`.
