# Phase I research report

## Executive finding

Phase I establishes that the observed Voynich transcription is a strongly structured symbolic system. Its glyph placement, token formation, edit geometry, repetition and local organization impose substantial joint constraints. Several narrow explanations—tested simple inverse transpositions, simple position-independent random homophony, and mechanism-grid variants lacking constrained formation—perform poorly. The evidence does **not** decide whether the structure carries encoded meaningful input, implements an artificial/formal symbolic language, or was generated without a transmitted message.

The evidence line in this report is deliberately scientific rather than chronological:

> manuscript structure → simple-transform checks → glyph/token structure → repetition/edit geometry → sequence organization → local/regime structure → mechanism space → recoverability → literature comparison → remaining unknowns

The claim-level audit is [PHASE1_CLAIMS.tsv](PHASE1_CLAIMS.tsv), the result entry point is [RESULT_INDEX.tsv](RESULT_INDEX.tsv), and the compact current measurement profile is [VOYNICH_FINGERPRINT_V1.md](VOYNICH_FINGERPRINT_V1.md).

## 1. Questions, scope and non-claims

Phase I asked: what reproducible structure exists in the chosen transcription; which simple transformations can or cannot reproduce or expose it; how glyphs form tokens and tokens relate locally; whether organization changes across line, page and manuscript scales; which components of a bounded synthetic mechanism space are necessary; and how statistical compatibility relates to information retention and recovery.

This synthesis uses only repository reports, frozen designs, stored artifacts, implementation where artifact semantics were ambiguous, the provenance audit, and the Task68/69 literature audit. It performs no metric recomputation, threshold selection or new inference.

“Glyph”, “token”, “line”, “page” and “regime” denote transcription or analytical units. They do not assert letters, words, sentences, topics, plaintext blocks or keys. Phase I did not decipher the manuscript, validate a plaintext, identify a language, infer intrinsic information loss in the manuscript, or prove a historical production mechanism.

## 2. Corpus, transcription and evidence discipline

The baseline pipeline used `data_work/ZL3b-x7.txt`, SHA-256 `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`, with 39,026 tokens reported by the later baseline analyses. Tasks59–67 use their canonical parsed corpus with SHA-256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692` and 39,380 tokens in Task59. These are not silently pooled: differences in parser, representation, boundaries and sample are part of each result's limitations.

The main natural controls are English Doyle and Longfellow and Russian Astafiev. Experiment-specific controls additionally include shuffles preserving selected invariants, known transformations, structured positive controls, copy/mutate generators, stationary/drift/state simulations, and held-out partitions. A null answers only the question defined by what it destroys and preserves.

Every strong statement below separates an observed measurement from its interpretation. “Incompatible” and “disfavored” refer to a frozen tested mechanism and metric family, never to every cipher or generator in the same broad class.

## 3. Baseline manuscript structure

### Baseline Stage1–28 pipeline

**Question.** What glyph, token, sequence, boundary, graph, metadata and distributional structure is present before mechanism-specific tests?

**Dataset.** Canonical `ZL3b-x7` corpus above, with IVTFF source `data/ZL3b-n.txt` and stage-specific metadata.

**Method.** A frozen multistage pipeline covering inventory and token distributions, n-grams and repetition, boundary effects, edit and graphemic families, sequence networks, conditional dependence, vocabulary growth, blind distributional change and metadata validation.

**Controls/nulls.** Stage-specific token and block permutations, held-out/leave-one-block-out checks, boundary controls, Currier/hand validation and natural controls. There is no valid omnibus p-value for the whole pipeline.

**Result.** All 27 configured analytical stages completed. The durable baseline is multiscale: constrained token forms and edit neighborhoods; non-random but uneven sequence recurrence; boundary/position effects; and distributional heterogeneity across the manuscript. Of 104 frozen sequence candidates, 51 replicated above a frequency null; 28 frozen distance-profile candidates produced 21 FDR-significant but zero robust cross-block profiles. Blind global change profiles found stable multiscale discontinuities, while metadata-conditioned residual clustering was weak/inconclusive.

**Interpretation.** The manuscript cannot be summarized by marginal frequencies alone. Its organization changes with unit and scale, and metadata explains only part of that structure.

**What this does not establish.** Pipeline completion does not make every exploratory candidate confirmatory. Clusters are not topics, repeated sequences are not grammar rules, and graph families are not morphemes.

**Artifacts.** [baseline report](../../experiments/voynich-v1/REPORT.md), [replicated local structure](../../experiments/voynich-v1/outputs/replicated-local-structure/replicated_local_structure_report.md), [global distributional report](../../experiments/voynich-v1/outputs/global_distributional_report.md), [conditional regimes](../../experiments/voynich-v1/outputs/conditional-regimes/conditional_regime_report.md).

**Implementation.** `runner/`, `cmd/`, and reusable `internal/` packages; the detailed mapping is in [METHOD_INDEX.md](../methods/METHOD_INDEX.md).

**Commit/provenance.** Baseline report records commit `61d6c206…`, a completed remote run and the corpus hash. Task70b preserved and audited the repository evidence boundary.

## 4. Checks of simple transformations and inverse methods

### Tested inverse transposition

**Question.** Can a simple inverse transposition restore structural order in Voynich on holdout data?

**Dataset.** Voynich logical lines split first 80% discovery / last 20% holdout; known-transposed Doyle for calibration.

**Method.** The frozen `structural-v2` objective ranked widths/orders without lexical oracle. Synthetic validation first showed top-3/exact recovery in 3/3 known transforms.

**Controls/nulls.** Doyle T2/T4/T8 ranges, Doyle/Longfellow width-2 natural audit, and 30 fixed-seed full-token permutations.

**Result.** The selected `w2 natural` inverse reduced all four holdout metrics: transition −0.000965, relation −0.000797, sequence-2 −0.044830 and sequence-3 −0.003044. Discovery and holdout directions agreed.

**Interpretation.** The selected tested inverse did not restore the targeted structural organization; this is a negative result for that search family.

**What this does not establish.** The 80/20 split and meaningful-effect threshold were not preregistered. The result does not exclude other widths, layouts, transposition families, compound systems or plaintext recovery.

**Artifacts.** [method validation](../../research/phase1/inverse-transposition-search/INVERSE_TRANSPOSITION_TASK54_1_REPORT.md), [Voynich validation](../../experiments/inverse-transposition/voynich-validation/VALIDATION_REPORT.md).

**Implementation.** `research/phase1/inverse-transposition-search`, `research/phase1/voynich-validation`, `internal/inversetransposition`.

**Commit/provenance.** Voynich manifest records `df5eb2410fe5fc57c0981e699c7392bc2ceb71be` and fixed seeds/hashes.

### Simple homophony and inverse homophony

**Question.** Does random homophonic expansion reproduce the observed profile, and can an inverse method recover known homophonic structure reliably enough to analyze Voynich?

**Dataset.** Doyle global-H/weighted/frequency transformations for forward controls; 11 synthetic corpora for inverse validation. Voynich was reserved for inverse Phase B.

**Method.** Forward dose-response comparisons and a separately frozen inverse class/structure recovery gate.

**Controls/nulls.** H2/H4/H6/H8 variants, untransformed prose, random baselines and held-out synthetic pairs.

**Result.** Forward homophony moved corrected token-order dependence toward Voynich but failed joint positional and repetition targets. The inverse method recovered classes above random in 9/9 but recovered the planted structure in only 3/9; its gate failed, so Voynich Phase B was `NOT_RUN_BY_DESIGN`.

**Interpretation.** Simple random homophony is not a sufficient joint mechanism, and the available inverse method cannot support a Voynich inference.

**What this does not establish.** Neither result excludes position-dependent homophony, structured code tables, verbosity/respacing systems such as Naibbe, or compound encodings. A failed inverse method is not evidence that homophony is present or absent.

**Artifacts.** [Task58 forward comparison](../../experiments/rozanova-temerev-v1/REPORT.md), [inverse synthetic validation](../../experiments/inverse-homophony/synthetic-validation/SYNTHETIC_VALIDATION_REPORT.md), and the original Task55 directories indexed by the provenance audit.

**Implementation.** `research/phase1/corpus-transform`, `research/phase1/inverse-homophony`, and their reusable internal packages.

**Commit/provenance.** Inverse manifest records `9737c991…` and `git_dirty:true`; Task70b verified all 11 input hashes. Ten Task55 result directories and 1,698 declared hashes are preserved; Phase B has no artifacts because the frozen gate prohibited it.

## 5. Glyph and token structure

### Token-order MI and glyph-edge coupling (Task58)

**Question.** How much adjacent dependence remains after correcting for within-line token frequencies, and is token-edge coupling different?

**Dataset.** Voynich plus Doyle, Longfellow and Astafiev, using the experiment's declared parsing.

**Method.** Adjacent token-identity MI and last-to-first glyph-edge MI minus the mean of 100 within-line shuffles.

**Controls/nulls.** Natural prose and forward homophony series.

**Result.** Voynich token MI was 3.053200 raw, 2.962465 shuffle mean, 0.090736 corrected, or 1.099% share. Edge corrected MI was 0.216302 bits, 6.271% share. Doyle token corrected share was 9.376% and edge share 1.766%.

**Interpretation.** Visible token identities have weak average order dependence, while boundaries retain glyph-level coupling. This combination argues against equating visible tokens with ordinary prose words without qualification.

**What this does not establish.** It does not show absence of a message, identify segmentation, or prove encryption.

**Artifacts.** [report](../../experiments/rozanova-temerev-v1/REPORT.md), [authoritative 100-shuffle rows](../../experiments/rozanova-temerev-v1/KEY_100_RESULTS.tsv).

**Implementation.** `research/phase1/rozanova-temerev`, `internal/evaglyph`.

**Commit/provenance.** Current artifacts are at `ba88b2de…`. The initial map-order summation was nondeterministic in low-order digits; sorted-key accumulation corrected it without changing the qualitative conclusion.

### Positional glyph specialization (Task59)

**Question.** Are frequent glyphs restricted to token positions, and does simple position-independent homophony reproduce this?

**Dataset.** Canonical 39,380-token, 45-glyph Voynich parse plus three natural controls.

**Method.** Initial/medial/final occurrence shares with a frozen near-strict threshold, within-token shuffles and multiple-testing correction.

**Controls/nulls.** Doyle H2/H4/H8 corrected random homophony, natural texts, position-dependent and structured-token positive controls.

**Result.** Six frequent glyphs (`q`, `N`, `e`, `E`, `i`, `m`) were near-strict specialists with shares 0.9505–0.9912. Natural controls produced 0–1; every simple homophony control produced zero.

**Interpretation.** Token-internal formation is strongly constrained; simple position-independent homophony is insufficient for this property.

**What this does not establish.** Positional specialization also occurs in morphology, orthography and scribal systems. It does not identify glyph values or exclude structured/position-dependent encoding.

**Artifacts.** [report](../../experiments/glyph-position-v1/REPORT.md), [comparison](../../experiments/glyph-position-v1/POSITIONAL_SPECIALIZATION_COMPARISON.tsv).

**Implementation.** `research/phase1/glyph-position-analyze`, `internal/evaglyph`.

**Commit/provenance.** The original negative-control generator leaked position and is invalid. Commit `ba88b2de…` replaced it with seeded occurrence-wise draws and regression tests. The legacy manifest records `a6904f0…`; the corrected completed tree is `ba88b2de…`.

## 6. Repetition and edit geometry

### Exact/near repetition (Task60)

**Question.** Are adjacent exact and near repeats enriched after controlling vocabulary, length and line structure, and where do edits occur?

**Dataset.** Same canonical corpus family as Tasks59–67, with natural controls and illustration-label subset.

**Method.** Exact adjacency, run counts, Levenshtein distance ≤1, operation/position decomposition and label-specific tabulation.

**Controls/nulls.** Global token shuffle, line-preserving shuffle, frequency/length-matched pairs, natural prose and corrected H2/H4/H8 homophony.

**Result.** Exact repetition was 0.009119 versus global 0.003127 (`z=21.574`) but line-null 0.008923 (`z=0.464`). Near repetition was 0.058656 versus global 0.025631, line-null 0.050938 (`z=7.705`) and matched 0.034936. Insertions/deletions concentrated at token beginnings. Homophony moved near repetition monotonically away from Voynich. Label analysis had only 170 within-label pairs and remained low-power.

**Interpretation.** Line/local organization explains much exact repetition but not all near-form enrichment. Edit geometry is structured, and simple random homophony is insufficient.

**What this does not establish.** Near neighbors are not demonstrated morphological paradigms, copying ancestry or semantic relations.

**Artifacts.** [report](../../experiments/token-repetition-v1/REPORT.md), `EXACT_ADJACENT_REPETITION.tsv`, `NULL_NEAR_REPETITION.tsv`, `EDIT_OPERATION_POSITION.tsv` and `HOMOPHONY_RUN_DOSE_RESPONSE.tsv` in that directory.

**Implementation.** `research/phase1/token-repetition-analyze` and shared glyph/edit packages.

**Commit/provenance.** Manifest records `ba88b2de…`, corpus hash and seed `20260823`.

## 7. Sequence organization and character predictability

### Character entropy (Task61)

**Question.** Is the glyph stream unusually predictable, and how sensitive is the estimate to boundaries?

**Dataset.** Shared-EVA continuous stream and explicitly re-represented boundary/within-token streams, plus natural controls.

**Method.** Shannon plug-in `h2 = H(X_i|X_{i-1})`; higher orders reported only with coverage.

**Controls/nulls.** Glyph shuffle, within-token glyph shuffle, global token shuffle, within-line token shuffle and corrected homophony series.

**Result.** Voynich continuous `h1=4.0058`, `h2=2.7726` bits versus Doyle `h2=3.5978`; boundary-symbol `h2=2.4458` and within-token `h2=2.5026`. Continuous `h4` coverage was only 0.116.

**Interpretation.** Low conditional entropy is reproducible in direction but is partly a property of representation and boundary policy.

**What this does not establish.** Entropy alone cannot distinguish language, cipher or generation; sparse high-order estimates are not evidence.

**Artifacts.** [report](../../experiments/character-entropy-v1/REPORT.md), `ENTROPY_BY_ORDER.tsv`, `ENTROPY_BOUNDARY_MODES.tsv`.

**Implementation.** `research/phase1/character-entropy-analyze`, `internal/characterentropy`.

**Commit/provenance.** Reconstructed verified artifact commit `1692874b…`; legacy dirty status is missing.

### Token formation (Task62)

**Question.** Can a local token-internal model reproduce form geometry without sequence rules?

**Dataset.** Contiguous physical line blocks split 60% train, 20% validation, 20% test.

**Method.** Frozen candidate formation models selected by validation cross entropy only.

**Controls/nulls.** Length/frequency, positional transitions, novel types, copy/mutate, structured-token and natural-language controls.

**Result.** `POSITION_MARKOV_1` won; test cross entropy was 2.360985 bits/glyph. It made novel and distance-1 types but generated near-repeat ≈0.0234 versus TEST 0.0395 and missed some entropy/position targets.

**Interpretation.** `LOCAL_FORMATION: PARTIAL`; `SEQUENCE: SEPARATE_SEQUENCE_RULE_REQUIRED`.

**What this does not establish.** The model is not a reconstruction of morphology, cipher grammar or production history.

**Artifacts.** [report](../../experiments/token-formation-v1/REPORT.md), [frozen design](../../experiments/token-formation-v1/TOKEN_FORMATION_DESIGN.md), `MODEL_SELECTION.tsv`, `GENERATIVE_VALIDATION.tsv`.

**Implementation.** `research/phase1/token-formation-analyze`, `internal/tokenformation`.

**Commit/provenance.** Reconstructed verified commit `9e21b01e…`; `DESIGN_FROZEN` is present.

### Token transitions and repeated sequences (Task63 plus baseline confirmation)

**Question.** Does local form dependence survive matching, and do frozen repeated sequences support higher-order organization?

**Dataset.** Canonical corpus with line/block partitions and held-out folds.

**Method.** Adjacent versus nonadjacent same-line matching; frozen local transition generator; separately, FDR/frequency-null replication and conditional continuation tests for pre-existing sequence candidates.

**Controls/nulls.** Length-pair matching, global/within-line shuffles, frequency and first-order Markov controls, leave-one-block-out prediction and jackknife.

**Result.** Task63 found ≈0.05855 adjacent near rate versus ≈0.05663 matched control: small `FORM_DEPENDENCE_ONLY/PARTIAL`. Baseline confirmation retained 51/104 sequences above a frequency null. Of three frozen n≥3 candidates, two lacked block support and one was position-dependent; a deeper cross-block audit classified the latter `NO_POSITIONAL_STRUCTURE` despite two nominal tests.

**Interpretation.** Sequence organization is real but modest after matching. Evidence for general higher-order rules is weak and candidate-specific.

**What this does not establish.** No token transition is a deciphered rule; nominal p-values with sparse candidates do not establish grammar.

**Artifacts.** [Task63 report](../../experiments/token-transition-v1/REPORT.md), [replication report](../../experiments/voynich-v1/outputs/replicated-local-structure/replicated_local_structure_report.md), [higher-order report](../../experiments/voynich-v1/outputs/higher-order-sequences/higher_order_sequence_report.md), [positional continuation](../../experiments/voynich-v1/outputs/positional-continuation/positional_continuation_report.md).

**Implementation.** `research/phase1/token-transition-analyze`, `internal/tokentransition`, and baseline sequence stages.

**Commit/provenance.** Task63 reconstructed verified at `6cf53a6d…`; frozen analysis/model markers are present.

## 8. Local, line and manuscript regimes

### Line/local scale (Task64)

**Question.** Is physical line membership the privileged scale of near-form similarity?

**Dataset.** Canonical corpus with page and line structure; held-out discovery/replication reporting.

**Method.** Length-matched same-line nonadjacent pairs, line-membership destruction, shifted line blocks, fixed windows and page comparisons.

**Controls/nulls.** Different-line same-page and different-page matches, within-line order shuffle, global and within-page line-membership shuffles.

**Result.** Same-line nonadjacent rate 0.057824 versus same-page control 0.049905; delta +0.007939, bootstrap 95% CI [0.004600, 0.012813]. But `SHIFTED_LINE_OFFSET1` effect 0.019273 exceeded physical `LINE` 0.016459. Classification: `BROADER_LOCAL_REGIME`. Matching regime did not remove Task63's adjacency residual.

**Interpretation.** Structure is local and line-sized, but physical line boundaries are not uniquely privileged.

**What this does not establish.** A local regime need not be a sentence, topic, key or plaintext unit; line filling/copying remain alternatives.

**Artifacts.** [report](../../experiments/line-regime-v1/REPORT.md), `LINE_PAIR_SIMILARITY.tsv`, `REGIME_SCALE_COMPARISON.tsv`.

**Implementation.** `research/phase1/line-regime-analyze`, `internal/lineregime`.

**Commit/provenance.** Manifest commit `6cf53a6d…`, seed `64000`, two frozen markers.

### Local-regime topology (Task65)

**Question.** Does local organization look stationary, smoothly drifting, discretely recurrent, or mixed; how much is metadata-associated?

**Dataset.** Canonical corpus, metadata partitions and independently repeated Task64 split.

**Method.** Empirical lag decay, clustering/stability, transition/dwell diagnostics, metadata-conditioned effects, change points and distant recurrence.

**Controls/nulls.** Stationary, smooth, discrete and mixed simulations; natural corpora; Task62 stationary G-only; corrected shuffled-position recurrence null.

**Result.** `LOCAL_STRUCTURE CONFIRMED`, `TOPOLOGY MIXED_DRIFT_AND_STATES`, `METADATA PARTIALLY_METADATA_EXPLAINED`. Same-page lag-1 excess similarity 0.0609062 persisted. The Task64 discovery/replication difference (0.030069436 versus 0.003132198) reproduced as true regional heterogeneity. For distant recurrence the statistic is a **distance**, so smaller means a closer—and therefore more recurrence-like—match. The observed nearest-distant mean was 0.046108, not smaller than the shuffled-position null 0.002320 (`null − observed = −0.043788`); the corrected test therefore did **not** support recurrence beyond chance.

**Interpretation.** Organization changes locally and unevenly; some state-like clustering coexists with drift. There is no supported recurrent-discrete-regime claim from the corrected test.

**What this does not establish.** Clusters are not semantic topics, scribal intentions or cryptographic keys; `K=5` is a selected descriptive resolution.

**Artifacts.** [report](../../experiments/local-regime-topology-v1/REPORT.md), `LOCAL_REGIME_DECAY.tsv`, `TASK64_SPLIT_DIAGNOSIS.tsv`, `DISTANT_REGIME_RECURRENCE.tsv`.

**Implementation.** `research/phase1/local-regime-topology-analyze`, `internal/localregimetopology`.

**Commit/provenance.** Manifest commit `8e67ee09…`, seed `65000`, frozen design.

## 9. Mechanism-space constraints (Task66)

**Question.** Which operations in a bounded synthetic architecture grid are necessary for multiaxis Voynich-fingerprint compatibility, and do results hold out of sample?

**Dataset.** Three natural source corpora, 17 frozen Voynich targets from Tasks58–65, and frozen screening/development/held-out partitions.

**Method.** M0–M11 grid, family-normalized coverage, Pareto comparison, component ablations, held-out confirmation and input-dependence tests.

**Controls/nulls.** Identity/monoalphabetic/homophony baselines, stationary and boundary nulls, ablations and corpus transfer.

**Result.** M1 made almost no progress. M2 homophony improved token order but worsened positional, repetition, transition and entropy axes. `CONSTRAINED_FORMATION REQUIRED` (ablation delta 4.19); memory, slow state, macro and generated boundaries were `NOT_REQUIRED`; homophony and stochastic output were `DISFAVORED`. Input dependence was strong/partial in 33/36 comparisons. M10 K2/K4 held out; other candidates were inconclusive. Several fingerprint axes stayed below threshold.

**Interpretation.** Within this grid, the minimal compatible core is constrained formation plus dependence on input. Extra randomness is not a substitute for structure.

**What this does not establish.** “Required” means required inside this grid and score definition. It neither identifies the Voynich mechanism nor proves the input was meaningful plaintext.

**Artifacts.** [report](../../experiments/mechanism-space-v1/REPORT.md), [design](../../experiments/mechanism-space-v1/TASK66_DESIGN.md), `ARCHITECTURE_ABLATION.tsv`, `HELDOUT_RESULTS.tsv`, `PLAINTEXT_SENSITIVITY.tsv`.

**Implementation.** `research/phase1/mechanism-space-analyze`, `internal/mechanismspace`.

**Commit/provenance.** Reconstructed verified at `d485fe1c…`; target hashes and grid are recorded. Per-job seeds/config hashes are deterministic in code but **not recorded in result rows**. The frozen Pareto artifact also retained M0 because of a historical output bug; later code excludes it, but the artifact was not rewritten.

## 10. Recoverability and information loss (Task67)

**Question.** For selected synthetic mechanisms, how do fingerprint compatibility, reversibility, exact recovery, ambiguity, information retention and error propagation relate?

**Dataset.** Known plaintext from Doyle, Longfellow and Astafiev, encoded with frozen Task66 representatives M0, M1, M2, G/M3, M9, M10 and M11.

**Method.** Actual encode→decode runs, preimage/ambiguity estimates, clean recovery, decoder-knowledge levels, single glyph/boundary/segmentation corruption, reset/resynchronization and joint frontier analysis.

**Controls/nulls.** Identity and bijection, declared-codebook homophony, shuffled input, short blocks, state-oracle variants and three source corpora.

**Result.** M0/M1 were reversible and exactly recovered. M2 was ambiguous in mechanism class but exactly decodable with its declared codebook. G/M9 were intrinsically many-to-one in the tested implementation; state knowledge could not undo collisions. M10/M11 were practically fragile. Boundary-count changes and split/merge errors caused the strongest desynchronization; resets localized damage. Constrained candidates showed a fingerprint/recovery trade-off and none combined control-level unique recovery with broad compatibility.

**Interpretation.** Mathematical reversibility, exact decoding with external knowledge, retained information, preimage ambiguity and robustness are distinct axes. Statistical resemblance alone does not determine recovery.

**What this does not establish.** These are synthetic known-plaintext results. They do not show that Voynich is lossy, unrecoverable, keyed, codebook-dependent or damaged, and they do not estimate the manuscript's preimage count.

**Artifacts.** [report](../../experiments/recoverability-v1/REPORT.md), `FINAL_CLASSIFICATION.tsv`, `CLEAN_RECOVERABILITY.tsv`, `AMBIGUITY_GROWTH.tsv`, `ERROR_PROPAGATION.tsv`, `FINGERPRINT_INFORMATION_FRONTIER.tsv`.

**Implementation.** `research/phase1/recoverability-analyze`, `internal/recoverability`.

**Commit/provenance.** Initial commit `553ef02…` contained proxy/scaffold outputs and is superseded. Commit `0e8b3da…` replaced them with 102,900-row measured stochastic artifacts and is current. Frozen candidate/decoder hashes link back to Task66; top-level corpus SHA and dirty status remain legacy-missing.

## 11. Literature comparison

The [literature synthesis](../literature/LITERATURE_SYNTHESIS.md) and [Task58–67 crosswalk](../literature/TASK58_67_LITERATURE_CROSSWALK.tsv) separate agreement from analogy.

- Currier-class heterogeneity, positional restrictions, weak corrected token order, glyph-edge coupling, low conditional entropy and local repetition agree wholly or partly with prior observations. Replication strength varies because transcriptions and estimands differ.
- Task58 is an independent directional replication of Rozanova–Temerev with a different sample. Task61 is a representation-sensitive partial replication of Lindemann–Bowern.
- Task59–60 add matched controls to known positional/repetition descriptions. Tasks62–65 use held-out and scale-conditioned tests for which no direct methodological match was found.
- Greshko's Naibbe is evidence that a specified reversible, verbose, structured homophonic cipher can reproduce selected Voynich-like statistics; it is not an inverse recovery and is not equivalent to simple global-H.
- Timm–Schinner's copy/modify model shows selected properties are not language-specific; it does not select message-free generation over every message-bearing model.
- Tasks66–67 are methodologically distinct extensions. `NO_DIRECT_PRECEDENT_FOUND` is a search-audit statement, not a priority claim.

Literature and Phase I share the same central limitation: prose-only, cipher-only or generator-only comparisons rarely discriminate all three broad hypotheses.

## 12. Integrated fingerprint and disfavored narrow explanations

The current fingerprint combines: strong within-token positional constraints; weak mean token-identity order but stronger boundary-glyph coupling; enriched near-edit adjacency with initial edit bias; low representation-dependent glyph conditional entropy; small residual local transition dependence; broader local/line-sized and manuscript heterogeneity; and incomplete synthetic mechanism coverage. Full values and coverage cautions are in [VOYNICH_FINGERPRINT_V1.md](VOYNICH_FINGERPRINT_V1.md).

Phase I disfavors only these scoped explanations:

1. The selected simple inverse-transposition family as a way to restore the four frozen structural metrics.
2. Simple position-independent random homophony as a joint explanation of positional specialization and near-repeat geometry.
3. The tested inverse-homophony procedure as a basis for Voynich analysis; its synthetic gate failed.
4. In Task66's finite grid, monoalphabetic remapping alone, homophony as the main improvement, stochastic output, and variants without constrained formation.
5. A recurrent discrete-regime reading of Task65's distant-neighbor statistic after its corrected multiple-opportunity null.

These findings do not reject transposition in general, complex homophonic systems, substitution families as a whole, encryption, language, or structured generation.

## 13. Remaining viable model classes

**H_G — structured/message-free generation.** Supported by the fact that constrained generators and copy/modify mechanisms can reproduce selected formal properties, and by weak visible-token order. It remains incomplete because no tested message-free generator matches the whole frozen fingerprint on holdout, and no semantic absence test exists.

**H_L — natural/artificial symbolic language.** Organized co-occurrence, boundaries, local dependence and regimes are compatible with symbolic language. Ordinary visible word-token language is strained by weak token-order information, low conditional entropy and edit/repetition geometry, but artificial/formal systems, abbreviation and nonstandard segmentation remain viable.

**H_C — complex transformation of meaningful plaintext.** Forward systems such as Naibbe show compatibility is possible, and Task66 shows input dependence can survive while synthetic output becomes more Voynich-like. Simple tested transforms fail jointly, no validated inverse plaintext exists, and recoverability may depend on external state/table and transcription integrity.

The detailed balance is in [HYPOTHESIS_STATUS.md](HYPOTHESIS_STATUS.md). No class is selected.

## 14. Corrections, limitations and unknowns

Correction history is evidence, not a footnote: Task58's nondeterministic summation was corrected; Task59's original homophony control was invalidated and rerun; Task67's proxy artifacts were replaced by measured results. Task66 lacks frozen row-level seeds/config hashes. These facts are recorded in [PHASE1_CLAIMS.tsv](PHASE1_CLAIMS.tsv) and the [provenance audit](../audit/PROVENANCE_AUDIT.tsv).

Coverage remains uneven. Marginals, glyph position, token formation, edit geometry and local 1-D structure are comparatively strong. Lexical paradigms, page layout/2-D relations, full hierarchy, algorithmic/compression measures, semantic prediction and end-to-end three-class comparison are weak or absent; see [FINGERPRINT_COVERAGE.md](FINGERPRINT_COVERAGE.md).

What remains unknown:

- whether transcription glyphs, spaces and tokens correspond to any historical production units;
- whether there is recoverable meaningful input, and what external key/table/state would be required;
- whether local regimes reflect content, scribal behavior, layout, copying, material organization or mixtures;
- whether near-edit families are morphological, cryptographic, generative or scribal;
- whether a single mechanism can match the complete fingerprint without target leakage and retain recoverability;
- whether any semantic association generalizes under blinded held-out testing;
- how conclusions change across independent transcriptions and explicit uncertainty models.

These are open questions, not a Phase II design.

## 15. Phase I conclusion

The observed Voynich text is a strongly structured symbolic system at glyph, token, edit, sequence and local/manuscript scales. Phase I narrows the mechanism space: structure is not captured by marginal distributions, simple random homophony, the selected inverse transposition, or unconstrained randomness, and constrained token formation is central in the tested synthetic grid. At the same time, statistical compatibility and recoverability are separable, and no experiment establishes the manuscript's information loss or plaintext.

Therefore Phase I does not conclusively distinguish complex encoded meaningful input, an artificial/formal symbolic language, and structured message-free generation. That unresolved distinction is the principal scientific result at the boundary of current evidence.
