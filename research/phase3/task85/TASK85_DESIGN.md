# Task85 design: Voynich Formal Grammar Experimental Design, Baselines and Validation Contract

**Design version:** V1.0. **Authority:** tasks_ph3/task85.txt (all sections); tasks_ph3/00-PHASE3_GOALS.txt; Phase I report (`docs/phase1/PHASE1_RESEARCH_REPORT.md`); Phase II report (`docs/reports/phase2/PHASE2_REPORT.md`); `docs/reports/phase2/PHASE1_TO_PHASE2_CONCLUSION_MAP.tsv`; `docs/reports/phase2/PHASE2_HYPOTHESIS_STATUS.tsv`; `docs/reports/phase2/PHASE_III_OPEN_PROBLEMS.md`; Task83b's frozen corpora and provenance (`research/phase2/task83b/`); the frozen Fingerprint V2 registry (`research/phase2/fingerprint/F2_METRIC_REGISTRY_FINAL.tsv`). **Frozen before any grammar fitting:** yes. **Target-blind:** yes; no Voynich sequence statistic beyond what Task83b already froze (corpus content, provenance, metadata) is read to choose any threshold, model class, or split parameter in this design.

Task85 does not build, select, or fit a grammar. It freezes the experiment that Task86-89 will run.

## 1. Scientific context (task85 section 1)

Task85 takes as given, without reinterpreting, two authoritative predecessor results:

- **Phase I conclusion A** (`tasks_ph3/00-PHASE3_GOALS.txt` section 1; `PHASE1_TO_PHASE2_CONCLUSION_MAP.tsv` row P1P2-01): Voynichese token formation is strongly constrained, and this constrained formation is central and reproducible ("CONFIRMED... Fingerprint V2 preserves edit/token CORE structure across ZL3b and IT2a; all 13 CORE statuses remain stable"). The generating cause of this constraint remains unidentified.
- **Phase II result B** (`PHASE2_REPORT.md` sec.12; `PHASE2_HYPOTHESIS_STATUS.tsv`): `BEST_SUPPORTED_CLASS = INCONCLUSIVE`; `MECHANISM_IDENTIFICATION_FROM_F2 = NOT_IDENTIFIABLE`; the principal methodological limitation was a direct model/target coverage of 3/13 CORE F2 metrics (`EF1_GIANT_COMPONENT_SHARE`, `EF2_GLOBAL_CLUSTERING`, `EF3_DEGREE_FREQUENCY_SPEARMAN`), all one edit-related family.

Neither result is re-derived, re-tested, or reinterpreted here.

## 2. Central research question (task85 section 2)

> Can the observed structure of Voynichese be described by a compact formal generative system, capable of predicting previously unused parts of the corpus?

Not "what is the grammar of the Voynich language," and not "what does the text mean." See `PHASE3_LINE_A_RESEARCH_QUESTIONS.md` for the full scoped question set.

## 3. Operational definition of grammar (task85 section 3)

A grammar `G` is: an explicit generative model over observable TOKEN sequences and STRUCTURAL_STATE context, with (a) a formally specified representation, (b) a finite description, (c) a deterministic-given-seed implementation, (d) a computable likelihood/probability where the model class permits, and (e) a generative sampling procedure, and (f) a measurable complexity (`GRAMMAR_COMPLEXITY_CONTRACT.md`). `G` is never defined through a semantic assumption.

## 4. Terminological neutrality (task85 section 4)

Task85 uses only the neutral vocabulary GLYPH / TOKEN / COMPONENT / POSITION / CONTEXT / STRUCTURAL_STATE, defined without assuming glyph=letter, token=word, component=morpheme, line=sentence, page=discourse unit, or prefix/suffix=linguistic prefix/suffix. See `GRAMMAR_UNIT_REGISTRY.tsv` for the frozen definitions and their sources in the existing, already-frozen `internal/evaglyph` / `internal/metadatavalidation` machinery.

## 5. Three grammar levels and their observable variables (task85 sections 5-8)

| Level | Question | Observable variables allowed |
|---|---|---|
| G1 — TOKEN FORMATION | `P(token structure)` | GLYPH/COMPONENT combinations, TOKEN length, internal POSITION, formation constraints, edit-family relations, lexical paradigms, rare/unseen forms. Never: neighboring TOKENs, LINE position, PAGE, FOLIO, SECTION (except the minimum needed to identify the transcription unit itself, e.g. which GLYPHs belong to which TOKEN) |
| G2 — CONTEXTUAL SEQUENCE | `P(token_i \| local context, G1)` | Preceding TOKENs; following TOKENs only where the discovery protocol declares it legitimate; local transition history; repetition state; short-range TOKEN context. Never: manuscript hierarchy (LINE/PAGE/FOLIO/SECTION identity) — otherwise G3's incremental contribution over G2 cannot be measured (task85 section 7) |
| G3 — STRUCTURAL CONDITIONING | `P(token_i \| local context, structural state, G1, G2)` | Candidate STRUCTURAL_STATE variables audited in `GRAMMAR_F2_APPLICABILITY.tsv`: LINE position/boundary, LOCUS, FOLIO, recto/verso, SECTION, CURRIER, local regime. No variable is admitted merely because it improves fit (task85 section 8) |

### Nested model principle (task85 section 9)

`GRAMMAR_MODEL_REGISTRY.tsv` builds G1⊂G2⊂G3 explicitly: M6 (G2) is built on a frozen G1 as its back-off; M7 (G3) is built on a frozen G1+G2. `Δ(G1→G2)` and `Δ(G2→G3)` are measured as `CONTEXT_INFORMATION_GAIN` and `STRUCTURAL_INFORMATION_GAIN` (`TASK86_HANDOFF.md` forward note; `GRAMMAR_VALIDATION_CONTRACT.md`), the complexity-adjusted PM1/PM2 improvement of the nested model over its parent, on HELDOUT.

## 6. Corpus authority (task85 sections 10-11)

Authoritative corpora, Task83b-frozen, provenance-verified via `cmd/fingerprint-v2-verify`:

| Transcription | Raw IVTFF | Prepared canonical | TOKENs | Vocabulary types | LINEs | Leaves (FOLIOs) | GLYPH inventory |
|---|---|---|---|---|---|---|---|
| ZL3b | `data/ZL3b-n.txt` (sha256 `bf5b6d4a...`) | `data_work/ZL3b-x7.canonical.txt` (sha256 `f46f4190...`) | 39,380 | 8,243 | 5,385 | 103 | 45 |
| IT2a | `data/IT2a-n.txt` (sha256 `7f27a8b0...`) | `data_work/IT2a-x7.canonical.txt` (sha256 `10286ee7...`) | 37,945 | 8,069 | 5,207 | 102 | 32 |

Full checksums in `GRAMMAR_CORPUS_SPLIT_MANIFEST.json`'s `corpus_totals`. No older/unresolved corpus artifact is used. The two transcriptions' differing GLYPH-inventory sizes (45 vs. 32) is a transcription-notational difference, not evidence about Voynichese, and is never used to prefer one transcription's fitted grammar over the other.

### Transcription protocol (task85 section 11)

Model selection protocol (discovery on DEVELOPMENT, calibration on VALIDATION) is run identically on both transcriptions. A primary Line A conclusion requires `TRANSCRIPTION_STABLE` or explicit `DIRECTION_STABLE`/`TRANSCRIPTION_SENSITIVE` labeling (`GRAMMAR_VALIDATION_CONTRACT.md` section 2); neither transcription is a "discovery-only" corpus and the other a "replication-only" corpus — the identical protocol runs on both.

## 7. Data partitioning and structural split design (task85 sections 12-16)

Random token holdout is not used. The split unit is the physical **leaf** (FOLIO): recto/verso and multi-part foldout sides of the same leaf are always co-assigned, so no partition boundary crosses within one physical page. Leaves are stratified by (CURRIER language `$L`, SECTION `$I`) — using ZL3b's own metadata when a leaf exists there (its larger coverage), IT2a's otherwise — then, within each stratum, sorted by leaf number and assigned by a **fixed, seed-free positional rule**: index mod 5 in {0,1,2} -> DEVELOPMENT, 3 -> VALIDATION, 4 -> HELDOUT (60/20/20 target). This rule needs no PRNG and cannot be re-run into a different answer; it is implemented in `research/phase3/task85-analyze/main.go` and its output is `GRAMMAR_CORPUS_SPLIT.tsv` / `GRAMMAR_CORPUS_SPLIT_MANIFEST.json`.

Realized split (103 leaves total, `GRAMMAR_CORPUS_SPLIT_MANIFEST.json`):

| Partition | Leaves | ZL3b TOKENs | IT2a TOKENs |
|---|---|---|---|
| DEVELOPMENT | 70 | 26,831 (68.1%) | 25,720 (67.8%) |
| VALIDATION | 18 | 7,393 (18.8%) | 7,211 (19.0%) |
| HELDOUT | 15 | 5,156 (13.1%) | 5,014 (13.2%) |

Eight (Currier,section) strata have fewer than 5 leaves (`GRAMMAR_CORPUS_SPLIT_MANIFEST.json`'s `small_strata_lt_5_leaves_all_or_mostly_development`); by construction of the fixed rule these contribute disproportionately to DEVELOPMENT and little or nothing to VALIDATION/HELDOUT. This is a deliberate, documented trade-off (rare strata inform model construction rather than fragmenting HELDOUT into single-leaf slivers), not a hidden imbalance.

### Leakage control (task85 section 14)

Because the split unit is the whole physical leaf, near-identical local material spanning adjacent LOCI or LINEs can never land on both sides of a partition boundary — the finest granularity at which within-leaf repetition/near-duplication could leak is entirely contained in one partition. Of 96 numerically-adjacent leaf pairs, 48 cross a partition boundary (`GRAMMAR_CORPUS_SPLIT_MANIFEST.json`); this is an accepted, coarser (leaf-to-leaf, not locus-to-locus) boundary, consistent with "folio holdout" being one of task85 section 12's own listed granularities.

### Balance (task85 section 15)

Balance was checked, not engineered, by stratifying on (CURRIER, SECTION) before the fixed positional rule — never by permuting individual TOKENs. Per-partition, per-transcription SECTION composition, CURRIER composition, and mean/SD TOKEN length are recorded in `GRAMMAR_CORPUS_SPLIT_MANIFEST.json`'s `partition_balance`; mean TOKEN glyph-length is stable across partitions (ZL3b: 4.03/4.01/3.93 glyphs for DEVELOPMENT/VALIDATION/HELDOUT; IT2a: 4.10/4.09/4.02).

### Freeze (task85 section 16)

`GRAMMAR_CORPUS_SPLIT.tsv` and `GRAMMAR_CORPUS_SPLIT_MANIFEST.json` are frozen by the `GRAMMAR_CORPUS_SPLIT_FROZEN` marker in this directory. After freeze, the split is never modified in response to any Task86-89 result.

## 8. Model-class registry and justification (task85 sections 17-19)

`GRAMMAR_MODEL_REGISTRY.tsv` fixes M0-M7 plus an explicitly non-primary `NEURAL_AUX_UPPERBOUND` row. This is the standard formal-language-theory ladder (frequency-only -> fixed-order Markov -> variable-order Markov -> deterministic finite-state -> probabilistic finite-state -> explicit component/rule grammar -> token-sequence context -> structural conditioning), applied uniformly to GLYPH (G1), TOKEN-sequence (G2), and STRUCTURAL_STATE (G3) observables, with G1⊂G2⊂G3 nesting throughout (section 5 above). Each adjacent pair differs in exactly one structural capability, so a pairwise comparison isolates one capability at a time (task85 section 18's fairness requirement) rather than bundling several capability changes into one comparison. No class was chosen because it is expected to fit VM well (task85 section 17's prohibition); the set is the same ladder that would be applied to any comparably-structured symbol sequence.

**Model-class fairness** (task85 section 18): every row of `GRAMMAR_MODEL_REGISTRY.tsv` fixes inputs, accessible context, parameterization, training algorithm, hyperparameter space, generation algorithm, complexity measure, seed behavior, and failure conditions before any fitting. Two models are compared only when the complexity accounting (`GRAMMAR_COMPLEXITY_CONTRACT.md`) makes any asymmetric access to target information explicit (e.g. M7's larger STRUCTURAL_STATE input is charged its own state-partition description cost).

**Neural model policy** (task85 section 19): `NEURAL_AUX_UPPERBOUND` may be used only as an auxiliary predictive upper-bound reference, never as the primary grammar, unless a separate, prior contract defines an interpretable representation, comparable complexity measure, generation contract, and residual accounting for it — none is defined here, so it starts out-of-scope for `G_min` selection and for any PRIMARY evidence tier (`GRAMMAR_VALIDATION_CONTRACT.md` section 10).

## 9. Baselines, natural-language and message-free calibration controls (task85 sections 20-22)

`GRAMMAR_BASELINE_REGISTRY.tsv` fixes B0-B5 (VM-comparison baselines/controls, target-blind by construction) and `MFC0`-`MFC3` (a separate message-free calibration battery: fit the Task86+ pipeline on synthetic corpora from known-message-free generators, to calibrate what "grammar detected" means before ever touching VM). B5 reuses the repository's existing provenance-valid natural-text corpora (`internal/evaglyph.NaturalGlyphs`'s Doyle/Longfellow/Astafiev controls) rather than constructing new ones; if VALIDATION later shows these inadequate in scale, task85 section 21's scale-normalized protocol applies before any VM comparison, not an ad hoc new control.

## 10. Predictive and structural metrics (task85 sections 23-27)

**Predictive** (`GRAMMAR_METRIC_REGISTRY.tsv`): held-out negative log-likelihood, cross-entropy, perplexity where the scored unit is held fixed across compared models, unseen-TOKEN probability, predictive calibration, and a negative-discrimination AUC against matched negative controls. Training likelihood alone (`PM0`) is explicitly SECONDARY_DIAGNOSTIC_ONLY.

**Structural** (`GRAMMAR_F2_APPLICABILITY.tsv`): every metric in the frozen 33-row Fingerprint V2 registry (`research/phase2/fingerprint/F2_METRIC_REGISTRY_FINAL.tsv`) is audited for (a) which grammar level naturally exercises it, and (b) whether it can discriminate model classes under generation, or is "generation-vacuous" because it depends only on the borrowed STRUCTURAL_STATE skeleton (section 12 below) rather than on any TOKEN the grammar actually chooses.

**Key finding**: of the 13 CORE metrics, **10 discriminate** under generation (`2DL1`, `BP1`, `EF1`, `EF2`, `EF3`, `LC1`, `LC2`, `LS2`, `LS3`, `PF2` — spanning the 2D-LITE, boundary, edit, locus, line, and folio families) and **3 do not** (`HR1_FOLIO_VARIANCE_SHARE`, `HR1_SECTION_VARIANCE_SHARE`, `PF5_WITHIN_FOLIO_PROGRESSION` — all pure LINE-length/skeleton statistics that every model class reproduces identically by construction). This means Line A's own generation-validation battery already reaches genuine **multi-family** direct coverage (6 of the frozen non-trivial F2 families) on the 10 discriminating CORE metrics alone, addressing part of PHASE3_GOALS section 24's call to move beyond Phase II's single-family 3/13 — a design-time methodological observation about measurement, not a claim about which mechanism generated the VM (task85 section 47 firewall).

A second finding: the frozen F2 CORE/SUPPORTING battery has almost no G2-natural metric (only `cs2/prev-family-current-family`); Task87 (G2, the level this battery under-covers) may need new, pre-registered G2-specific structural metrics rather than relying on Fingerprint V2 alone. This is flagged forward in `TASK86_HANDOFF.md`.

**Family-level gates**: `GRAMMAR_VALIDATION_CONTRACT.md` section 4 fixes the frozen family-level success rule used by `GRAMMAR_SUFFICIENT`/`GRAMMAR_MINIMAL` before any Task86 fitting.

**Generative validation and scale** (task85 sections 26-27): every candidate admitted to final validation must generate independent samples, `G + seed_i -> synthetic corpus_i`, always onto a fixed, real, borrowed STRUCTURAL_STATE skeleton (never a skeleton the grammar invents — see section 12). Seed/replicate count and corpus scale are chosen by the stability/convergence diagnostic of `GRAMMAR_VALIDATION_CONTRACT.md` section 7, never by Voynich fit; Task82a's own scale-grid parameters are not inherited automatically (task85 section 27).

## 11. Complexity (task85 sections 28-29, 33-34)

`GRAMMAR_COMPLEXITY_CONTRACT.md` fixes the single primary measure, a two-part-code (MDL-style) `Complexity(G) = StructureCost(G) + LexiconCost(G) + ExceptionCost(G)`, with its coding assumptions fixed before any fitting: per-choice-point structure cost, a shared BIC-style per-real-parameter cost, a Shannon-coded lexicon cost, and a fixed 1-bit exception-flag overhead on top of an exception's own lexicon cost. Regular rules and lexical/structural exceptions share one representation; nothing is complexity-free for being labeled an exception.

## 12. Minimality, sufficiency, overfitting, unseen forms, ablations (task85 sections 30-32, 35-40)

`GRAMMAR_VALIDATION_CONTRACT.md` fixes: `GRAMMAR_MINIMAL` as the least-complex model meeting both frozen predictive and structural adequacy gates; `GRAMMAR_SUFFICIENT` as meeting those gates plus seed- and partition-stability; the overfitting-control reporting requirement (train/validation/heldout/complexity always together); the target-blind negative-token protocol; and the six required ablations (`GRAMMAR_ABLATION_REGISTRY.tsv`: remove token formation, remove local context, remove line position, remove hierarchy, remove lexical memory, remove state), each mapped only to the model classes where it is meaningful.

**Generation protocol / borrowed structural skeleton**: G3-level generation always fills TOKEN values onto the real, observed STRUCTURAL_STATE layout of whichever partition is being compared against; it never invents its own manuscript layout. Building a generated manuscript *object* with its own structural provenance is Line B's (Task90+'s) job, not Line A's — this is why some CORE F2 metrics are "skeleton-only" (section 10 above) and is stated once, centrally, in `GRAMMAR_VALIDATION_CONTRACT.md` section 11.

## 13. Information residual preparation (task85 sections 41-43)

`GRAMMAR_RESIDUAL_CONTRACT.md` fixes candidate residual representations, candidate `H(V|G)`/description-length estimators and their required properties, and the required residual-structure test scales — without computing any residual now and without ever calling it plaintext.

## 14. Message-free boundary and firewalls (task85 sections 44-47)

**Message-free boundary**: Task86-88 build and freeze `G_min`. Only after `G_min` is frozen may Task89 run `G_min + RNG -> message-free corpora`; `G_min` is never selected, or reselected, according to how well its message-free samples match the full VM target (task85 section 44).

**Semantics firewall** (section 45): no Task85-89 result is interpreted through a proposed translation, candidate plaintext, semantic label, presumed subject matter, or illustrated-page meaning. Line A is a structural experiment.

**Fontana/mechanism firewall** (section 46): no grammar representation is chosen, preferred, or excluded because it resembles or fails to resemble a Fontana external-memory mechanism; that comparison happens only after Line A freezes, and outside Task85-89.

**Phase II mechanism firewall** (section 47): Task81/82 mechanism fit, Task83r distances, "natural closest," Fontana endpoint values, shorthand distances, and extraction rankings are not used to choose any grammar class or threshold here. Phase II may motivate the open problems this design addresses (`PHASE3_GOALS`, `PHASE_III_OPEN_PROBLEMS.md`); it never sets a Line A target-fit parameter.

## 15. HELDOUT sentinel and bug policy (task85 sections 48-49)

Before any Task86-89 process first reads HELDOUT content for a selection decision, that task must create a `GRAMMAR_MODEL_SELECTION_FROZEN` sentinel recording: git commit; `GRAMMAR_CORPUS_SPLIT_MANIFEST.json` checksum; `GRAMMAR_MODEL_REGISTRY.tsv` checksum; `GRAMMAR_METRIC_REGISTRY.tsv` checksum; `GRAMMAR_COMPLEXITY_CONTRACT.md` checksum; the selected model/hyperparameters; generation settings; seed contract. After that sentinel, model selection is closed for that task. Post-opening bug policy follows Task83b's precedent (`research/phase2/task83b/PRNG_DETERMINISM_CONTRACT.md`'s normative-contract style): an implementation bug with unchanged scientific semantics is documented, fixed, regression-tested, and all affected models are recomputed symmetrically; a scientific-definition change invalidates the confirmatory run instead. This contract itself is never changed post hoc to fit a result.

## 16. Determinism (task85 section 50)

Inherited from Task83b: identical (input, model, parameters, seed, code) must yield a byte-identical result across process restart, map iteration order, worker count, and GOMAXPROCS. `research/phase3/task85-analyze/main.go` was run twice; `GRAMMAR_CORPUS_SPLIT.tsv` and `GRAMMAR_CORPUS_SPLIT_MANIFEST.json` were byte-identical both times (sha256 verified). Every map-keyed accumulation in this design's own tooling sorts keys before float64 accumulation or before a max/argmax selection (`dominant`, `sortedKeys` in `task85-analyze/main.go`), per the project-wide convention.

## 17. Seed contract (task85 section 51)

Fixed in `GRAMMAR_VALIDATION_CONTRACT.md` section 7: a stochastic model's seed is a pure function of `(model_id, hyperparameter point, transcription id, partition id, replicate index)`, never of job/worker/filesystem/map order; every replicate index is enumerated in advance and reported, none discarded (no seed cherry-picking).

## 18. Model failure and pilot policy (task85 sections 52-53)

`GRAMMAR_FAILURE_REGISTRY.tsv` fixes the seven required failure classes, each with a detection rule and a disposition rule that a failed fit is retained in the job ledger, never silently deleted. `GRAMMAR_VALIDATION_CONTRACT.md` section 7 fixes that target-aware pilots run only on DEVELOPMENT, and that any pilot-motivated design change lands before this freeze, not after.

## 19. Power/stability and multiple testing (task85 sections 54-55)

Fixed in `GRAMMAR_VALIDATION_CONTRACT.md` sections 7 and 9: fold/seed/scale counts come from a convergence diagnostic, never from chasing a p-value; generation validation is judged by a family-level gate rather than by per-metric significance, so no separate multiple-testing correction decision is deferred to after results are seen.

## 20. Primary vs secondary evidence (task85 section 56)

Fixed in `GRAMMAR_VALIDATION_CONTRACT.md` section 10.

## 21. Required registries and artifacts (task85 sections 57-58)

All produced in this directory; see `TASK85_RESULTS_MANIFEST.json` for the complete checksummed list.

## 22. Task86 handoff and required questions (task85 sections 59-60)

See `TASK86_HANDOFF.md` and `TASK85_REPORT.md`.

## 23. Verdicts, success criterion, non-goals, validation (task85 sections 61-64)

See `TASK85_REPORT.md` for the frozen verdict list, and section 64 validation results below.

### Validation performed (task85 section 64)

- All authoritative input paths and checksums verified against `research/phase2/task83b/` and `research/phase2/task83a/` provenance artifacts (sha256 recomputed directly by `task85-analyze`, matching the recorded provenance).
- Task83b's own `cmd/fingerprint-v2-verify` provenance verifier was not re-run here (it verifies Fingerprint V2's own refreeze manifest, unchanged by Task85); Task85 instead independently recomputed the sha256 of every corpus file it reads and recorded them in `GRAMMAR_CORPUS_SPLIT_MANIFEST.json`, so any drift from Task83b's recorded values would be visible by direct comparison.
- Corpus split determinism: two independent runs of `research/phase3/task85-analyze` produced byte-identical `GRAMMAR_CORPUS_SPLIT.tsv` and `GRAMMAR_CORPUS_SPLIT_MANIFEST.json` (sha256 verified).
- No split overlap: the split unit is the physical leaf; by construction one leaf belongs to exactly one partition, and recto/verso/foldout sides of a leaf are never split (`writeSplitTSV` iterates `allLeaves`, a deduplicated set).
- Registry consistency: `GRAMMAR_MODEL_REGISTRY.tsv`'s levels match `GRAMMAR_UNIT_REGISTRY.tsv`'s level tags; `GRAMMAR_ABLATION_REGISTRY.tsv`'s `applies_to_model_classes` only names ids present in `GRAMMAR_MODEL_REGISTRY.tsv`; `GRAMMAR_F2_APPLICABILITY.tsv`'s 33 metric rows match `research/phase2/fingerprint/F2_METRIC_REGISTRY_FINAL.tsv` exactly (verified by row count and metric_id set).
- Metric applicability consistency: every `NOT_APPLICABLE` row in `GRAMMAR_F2_APPLICABILITY.tsv` corresponds to a `comparison_eligibility=VOYNICH_ONLY_CONTEXT` row in the frozen F2 registry; no other row is marked `NOT_APPLICABLE`.
- Model -> allowed-input consistency and G1/G2/G3 access restrictions: enforced structurally by `GRAMMAR_MODEL_REGISTRY.tsv`'s `accessible_context` column and restated in section 5's table above; no G1 row's `accessible_context` mentions CONTEXT or STRUCTURAL_STATE, and no G2 row's mentions STRUCTURAL_STATE.
- `go build ./... && go vet ./... && go test ./... && go test -race ./...` and `git diff --check` results are recorded in `TASK85_RESULTS_MANIFEST.json`.

## 24. Non-goals (task85 section 63)

Task85 does not derive a Voynichese grammar, select a winning grammar, analyze HELDOUT, compute a final residual, generate a final message-free portfolio, compare any grammar to Fontana, repeat Task83r, or determine whether the VM is natural language, has meaning, or contains a translatable plaintext. None of this design's artifacts should be read as having done any of that.
