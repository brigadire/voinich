# Grammar validation contract

**Design version:** task85-v1. **Authority:** task85 sections 12-16, 20-27, 30-32, 35-36, 40, 50-56. **Frozen before grammar fitting:** yes. **Target-blind:** yes; no Voynich data or fit statistic is read to set any threshold below.

## 1. Partition roles (task85 sections 12-16)

`GRAMMAR_CORPUS_SPLIT.tsv` / `GRAMMAR_CORPUS_SPLIT_MANIFEST.json` freeze a hierarchical, FOLIO-level (physical-leaf), (Currier, section)-stratified partition into DEVELOPMENT (model construction, debugging, exploratory diagnostics), VALIDATION (model-class selection, hyperparameter selection, complexity calibration), and HELDOUT (final confirmatory evaluation only). HELDOUT is not opened for any purpose before `GRAMMAR_MODEL_SELECTION_FROZEN` (section 6 below). Random token splits are not used anywhere in Line A; the split unit is always the physical leaf, so recto/verso and multi-part foldout sides are never separated across partitions.

## 2. Cross-transcription protocol (task85 section 11)

Every VALIDATION-stage and HELDOUT-stage step in this contract is executed identically on ZL3b and IT2a; neither transcription is treated as the "real" fit and the other as a check. A primary Line A conclusion (a G_min class, a sufficiency/minimality verdict, a residual finding) requires:

- `TRANSCRIPTION_STABLE`: the quantitative conclusion, not merely its direction, agrees within a pre-registered tolerance across ZL3b and IT2a;
- `DIRECTION_STABLE`: the qualitative direction of the conclusion agrees, but magnitudes differ beyond the tolerance;
- `TRANSCRIPTION_SENSITIVE`: the conclusion does not survive transcription substitution and is reported as such, never silently dropped.

The frozen split's FOLIO-level partition assignment is IDENTICAL across ZL3b and IT2a (assigned from the union of leaf identifiers, ZL3b-authoritative when both carry a leaf); only the realized TOKEN/LINE counts per leaf differ between transcriptions (`GRAMMAR_CORPUS_SPLIT.tsv` columns `zl3b_*` / `it2a_*`).

## 3. Negative-token / unseen-form protocol (task85 sections 35-36)

`NEGATIVE_TOKEN_CONTROLS` are generated target-blind (constructed before any VM sequence statistic beyond length and GLYPH-frequency class is read) by resampling GLYPH sequences matched to the HELDOUT TOKEN under test on:

- length (mandatory);
- glyph-frequency class and COMPONENT inventory, where the model class under test defines one (best-effort; not mandatory where a model class has no COMPONENT alphabet).

Negative controls are drawn from a distribution that itself respects the frozen edit-graph/GLYPH-frequency structure of DEVELOPMENT (never uniform random GLYPH noise, which task85 section 36 explicitly forbids as trivializing the discrimination task). `PM6_NEGATIVE_DISCRIMINATION_AUC` (`GRAMMAR_METRIC_REGISTRY.tsv`) is the frozen scoring metric for this protocol.

## 4. Family-level structural gates (task85 sections 24-26)

Structural/generative evidence is aggregated at three levels before any verdict: metric, family (`GRAMMAR_F2_APPLICABILITY.tsv` family column), and structural scale (G1/G2/G3, via `min_g_level`). A single family passing does not constitute multi-family support (task85 section 25; PHASE3_GOALS section 25 "coverage is not fit").

The frozen family-level success gate for a candidate G at level G-k (k in {1,2,3}) is:

> A majority of that family's `GENERATION_APPLICABLE` and NOT `skeleton_only=TRUE` metrics (per `GRAMMAR_F2_APPLICABILITY.tsv`) show G's synthetic-corpus value within the family's own null-model spread of the matched HELDOUT-partition value, for every family with two or more such metrics; a single-metric family is reported but never used alone to grant or deny the gate.

`skeleton_only=TRUE` metrics (`HR1_FOLIO_VARIANCE_SHARE`, `HR1_SECTION_VARIANCE_SHARE`, `HR1_LOCUS_VARIANCE_SHARE`, `LS1_LINE_LENGTH_CV`, `PF5_WITHIN_FOLIO_PROGRESSION`) are reported for description only and never counted toward this gate, because under the borrowed-structural-skeleton generation protocol (section 8 below) every model class reproduces them by construction.

## 5. Overfitting control (task85 section 32)

Train (DEVELOPMENT), VALIDATION, and HELDOUT performance (PM1-PM6) and Complexity(G) are reported together for every candidate, never train-only. `MEMORIZATION_DOMINATED` (`GRAMMAR_FAILURE_REGISTRY.tsv`) is declared when the DEVELOPMENT-to-HELDOUT PM2 gap exceeds a threshold frozen at Task86 discovery time from the message-free calibration battery's (`MFC0`-`MFC3`) own train/heldout gap distribution — never chosen after seeing VM's own gap.

## 6. Sufficiency and minimality (task85 sections 30-31)

```
PredictiveAdequacy(G):
    PM1/PM2 improve over B1 (and, where applicable, B2) by more than the
    calibration-battery (MFC0-MFC3) null spread, on HELDOUT, on BOTH
    transcriptions at least DIRECTION_STABLE.

StructuralAdequacy(G):
    G's frozen family-level gate (section 4 above) is met for every
    GENERATION_APPLICABLE family at G's own level (G1 for an M0-M5 model,
    G1+G2 for M6, G1+G2+G3 for M7).

GRAMMAR_MINIMAL:
    G_min = argmin Complexity(G) subject to
            PredictiveAdequacy(G) AND StructuralAdequacy(G)
    over the frozen candidate set in GRAMMAR_MODEL_REGISTRY.tsv (status
    PRIMARY_CANDIDATE only).

GRAMMAR_SUFFICIENT (a candidate, not necessarily G_min, is SUFFICIENT if,
in addition to PredictiveAdequacy and StructuralAdequacy):
    - stable across seeds (section 7 below);
    - stable across the held-out partitions considered in section 40's
      ablation contract, i.e. not merely stable under one arbitrary
      resample.
    High likelihood alone is never sufficient (task85 section 31).
```

## 7. Seed contract, stability, and pilots (task85 sections 50-54)

- **Determinism**: identical (input, model, parameters, seed, code) yields byte-identical output across process restart, map order, worker count, and GOMAXPROCS (task85 section 50, inherited from Task83b). Every accumulation over a Go map is preceded by a sort of its keys before any float64 accumulation (project-wide convention; see `internal/evaglyph.entropy`, `task85-analyze`'s `dominant`/`sortedKeys`).
- **Seed derivation**: a stochastic model's PRNG stream is seeded by a pure function of `(model_id, hyperparameter point, transcription id, partition id, replicate index)`, never of job order, worker assignment, filesystem order, or map iteration. Seed cherry-picking is forbidden: every replicate's seed is enumerated in advance (replicate index 0..R-1) and every replicate's result is reported, none discarded post hoc.
- **Pilots**: target-aware pilots (runtime estimation, numerical stability, implementation validation, rough hyperparameter bounds) are permitted only on DEVELOPMENT and never read VALIDATION or HELDOUT. Any design change motivated by a pilot must land before this contract's freeze, not after.
- **Power/stability**: the number of seeds/replicates, folds, and generated-corpus scale for Task86+ generation validation are chosen from a stability/convergence diagnostic (e.g. the standard error of PM1/PM2 and of each `GENERATION_APPLICABLE` F2 statistic across an increasing replicate count, on DEVELOPMENT only) — stopping once additional replicates change the reported statistic by less than a pre-registered tolerance — never by the number of replicates needed to reach a desired p-value.

## 8. Generated corpus scale (task85 section 27)

Generation validation uses matched-size synthetic corpora (same TOKEN count as the partition being compared against) as the primary scale, plus at least one additional scale (e.g. 0.5x and 2x) reported for convergence diagnostics. Task82a's cost-scoping experience motivates checking convergence across scale, but its specific scale grid is not inherited automatically (task85 section 27); Task86+ selects its own grid from the section 7 stability diagnostic.

## 9. Multiple testing (task85 section 55)

Because generation validation runs many family/metric/transcription/scale combinations, the frozen policy is a family-level gate (section 4), not a per-metric significance threshold: no single metric's p-value is interpreted in isolation, and no correction-vs-no-correction choice is made after inspecting results. Where a single-metric confirmatory claim is unavoidable (e.g. the sole G2-level `cs2` metric), it is explicitly labeled SECONDARY (section 10) unless independently corroborated by a PRIMARY predictive gain.

## 10. Primary vs secondary evidence (task85 section 56)

PRIMARY: HELDOUT predictive performance (PM1-PM6); complexity-adjusted adequacy (section 6); the frozen family-level structural gates (section 4) on `GENERATION_APPLICABLE`, non-`skeleton_only` metrics; cross-transcription robustness (section 2).

SECONDARY: exploratory DEVELOPMENT diagnostics; individual TOKEN examples; visualization; SUPPORTING (non-CORE) F2 metrics used alone; `skeleton_only` metrics; `NOT_APPLICABLE` (VOYNICH_ONLY_CONTEXT) metrics; qualitative grammar inspection. No Line A verdict depends on SECONDARY evidence alone.

## 11. Generation protocol note (structural skeleton)

Grammar sampling for structural (G3-level) generative validation always fills TOKEN values onto a fixed, real, template-borrowed STRUCTURAL_STATE skeleton (the LINE/FOLIO/LOCUS/section layout of the partition being compared against), never a skeleton invented or generated by the grammar itself. Constructing a generated *manuscript object* (a skeleton with its own provenance) is Line B's (Task90+'s) responsibility, per the PHASE3_GOALS Line A/Line B firewall (section 32 there); Line A's G3 only ever asks "given this real layout, which TOKEN goes here." This is why `skeleton_only` F2 metrics (section 4 above) cannot discriminate model classes under Line A's generation protocol, and is stated once here as the single source of truth the rest of this contract and `GRAMMAR_F2_APPLICABILITY.tsv` refer back to.
