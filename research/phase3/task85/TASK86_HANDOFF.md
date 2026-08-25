# Task86 handoff

This file contains only the frozen Task85 Line A grammar experiment design as it applies to Task86's scope. It does not include, reference, or imply any Voynich comparison result, any grammar fit, or any HELDOUT content.

Frozen design root: `research/phase3/task85/`

Task86 is authorized to investigate **G1 — TOKEN FORMATION only** (`GRAMMAR_UNIT_REGISTRY.tsv`, `TASK85_DESIGN.md` section 5). Task86 does not redefine any part of this design; if a genuine gap is found, it is escalated, not patched in place (task85 section 59).

## Authoritative corpus

- ZL3b: `data_work/ZL3b-x7.canonical.txt` (sha256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`) + `data/ZL3b-n.txt` (sha256 `bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc`)
- IT2a: `data_work/IT2a-x7.canonical.txt` (sha256 `10286ee7e11ad974e9d0f884e3b0df1b588745a4b77ad428a638a5ff63946a8b`) + `data/IT2a-n.txt` (sha256 `7f27a8b0feed8f6de0a99900df6bf912dd1d295c38e5f830bac8b41c3f536fb5`)
- Load via `internal/genericsegmentation.ReadCorpus` + `internal/metadatavalidation.ParseIVTFF`/`Align` (the pattern in `research/phase3/task85-analyze/main.go`), not a naive line-index join — IT2a's raw locus count and physical LINE count differ (5215 vs 5207) because of same-line continuation loci.

## Allowed partitions

Read `GRAMMAR_CORPUS_SPLIT.tsv` and `GRAMMAR_CORPUS_SPLIT_MANIFEST.json`. Task86 discovery/fitting/hyperparameter search may read **DEVELOPMENT and VALIDATION only**. **HELDOUT must not be opened** until Task86 issues its own `GRAMMAR_MODEL_SELECTION_FROZEN` sentinel (task85 section 48; `TASK85_DESIGN.md` section 15) recording git commit, the checksums of every file listed below, the selected model/hyperparameters, generation settings, and the seed contract. After that sentinel, no further model-class or hyperparameter change is permitted for that run.

## Unit definition

`GRAMMAR_UNIT_REGISTRY.tsv`. G1 sees only GLYPH, TOKEN, COMPONENT, and TOKEN-internal POSITION — never CONTEXT or STRUCTURAL_STATE (no neighboring TOKENs, LINE position, PAGE, FOLIO, or SECTION), except the minimum information needed to identify the TOKEN boundary itself.

## Allowed G1 model classes

`GRAMMAR_MODEL_REGISTRY.tsv` rows M0-M5 only (status `PRIMARY_CANDIDATE`, level G1): IID/token-frequency (M0), character/glyph n-gram (M1), variable-order Markov (M2), deterministic finite-state (M3), probabilistic finite-state (M4), component/rule formation grammar (M5). `NEURAL_AUX_UPPERBOUND` may be run only as an auxiliary reference (task85 section 19), never for `G_min` selection.

## Hyperparameter ranges

Per model class, from `GRAMMAR_MODEL_REGISTRY.tsv`'s `hyperparameter_space` column: M1/M2 order/depth and smoothing constants; M3/M4 merging-significance threshold and max states; M5 slot count and minimum rule support. All hyperparameter selection happens on VALIDATION only (`GRAMMAR_VALIDATION_CONTRACT.md` section 1).

## Baselines

`GRAMMAR_BASELINE_REGISTRY.tsv` rows B0-B4 (B5 and the natural-language protocol are optional at G1's scope; the `MFC0`-`MFC2` message-free calibration battery restricted to G1-level generators, e.g. M0/M1/M3, must be run and reported **before** any VM DEVELOPMENT data is fit, per task85 section 22).

## Metrics

Predictive: `GRAMMAR_METRIC_REGISTRY.tsv` PM0-PM6, scored at the GLYPH level for G1-only models (TOKEN-level PM1-PM6 do not apply until a G1 model's TOKEN-formation probability is marginalized consistently; document which unit is scored). Structural: `GRAMMAR_F2_APPLICABILITY.tsv` rows with `min_g_level=G1` (`EF1_GIANT_COMPONENT_SHARE`, `EF1_ISOLATE_SHARE`, `EF2_GLOBAL_CLUSTERING`, `EF3_DEGREE_FREQUENCY_SPEARMAN`, `LP1_RULE_SUPPORT_GINI`, `LP4_PREFIX_ATTACHMENT_NMI`, `LP4_SUFFIX_ATTACHMENT_NMI`) are all `GENERATION_APPLICABLE` and non-`skeleton_only` — a genuinely discriminating battery available at G1 alone.

## Complexity accounting

`GRAMMAR_COMPLEXITY_CONTRACT.md` in full: `Complexity(G) = StructureCost(G) + LexiconCost(G) + ExceptionCost(G)`, coding assumptions fixed there, never re-derived per model.

## Generation protocol

`G + seed_i -> synthetic vocabulary_i` (G1 has no LINE/FOLIO skeleton to borrow; it generates a TOKEN population directly). Scale and replicate count from the stability diagnostic in `GRAMMAR_VALIDATION_CONTRACT.md` section 7, never from Voynich fit.

## Seeds

`GRAMMAR_VALIDATION_CONTRACT.md` section 7: seed is a pure function of `(model_id, hyperparameter point, transcription id, partition id, replicate index)`; every replicate enumerated and reported in advance, none discarded.

## Validation rules

`GRAMMAR_VALIDATION_CONTRACT.md` in full, in particular sections 3-6 (negative-token protocol, family-level gate restricted to G1-applicable families, overfitting control, `GRAMMAR_SUFFICIENT`/`GRAMMAR_MINIMAL` at G1 scope) and `GRAMMAR_ABLATION_REGISTRY.tsv` row A5 (remove lexical memory — applicable to M0 at G1 scope).

## Model-selection rule

`G_min` (restricted to G1 candidates) `= argmin Complexity(G)` subject to `PredictiveAdequacy(G) AND StructuralAdequacy(G)` (`GRAMMAR_VALIDATION_CONTRACT.md` section 6), evaluated identically on ZL3b and IT2a, cross-transcription status reported per `GRAMMAR_VALIDATION_CONTRACT.md` section 2. A model that fails is recorded under its `GRAMMAR_FAILURE_REGISTRY.tsv` class, never deleted from the job ledger.

## Freeze condition

Task86 issues `TOKEN_GRAMMAR_FROZEN` (per `tasks_ph3/02-PAHSE3-CONVERSATION.txt`'s Line A roadmap) only after: the `GRAMMAR_MODEL_SELECTION_FROZEN` sentinel above is recorded, HELDOUT confirmatory evaluation is complete and reported for every candidate (not only the winner), and cross-transcription status is reported for the selected G1 model.

## Known gap flagged forward (not Task86's to fix)

The frozen Fingerprint V2 CORE/SUPPORTING battery has almost no G2-natural metric (`GRAMMAR_F2_APPLICABILITY.tsv`: only `cs2/prev-family-current-family` has `min_g_level=G2`). Task87 may need to pre-register new G2-specific structural metrics rather than relying on Fingerprint V2 alone for its own family-level gate.
