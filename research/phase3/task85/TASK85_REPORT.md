# Task85 report: Voynich Formal Grammar Experimental Design, Baselines and Validation Contract

This report answers task85 section 60's required questions and issues task85 section 61's verdicts. It presupposes `TASK85_DESIGN.md` and the registries/contracts alongside it; citations below point to the specific file and section that is authoritative for each answer.

## Answers to the preregistered questions

1. **What is operationally called "grammar"?** An explicit generative model over TOKEN sequences and STRUCTURAL_STATE context with a formal representation, finite description, deterministic-given-seed implementation, computable likelihood where the model class permits, a generative sampler, and a measured complexity (`TASK85_DESIGN.md` section 3). Never defined through a semantic assumption.

2. **What transcription units are used?** GLYPH (evaglyph.CollapseEVA atomic symbol), TOKEN (whitespace-delimited field of the frozen canonical corpus), COMPONENT (model-class-defined decomposition), LINE (physical-line join via `metadatavalidation.Align`), LOCUS, FOLIO (physical leaf), SECTION (`$I`), CURRIER (`$L`) — `GRAMMAR_UNIT_REGISTRY.tsv`.

3. **Why this unit representation?** It reuses the same GLYPH/TOKEN/metadata machinery already frozen and shared by Task58/59/60/65/77/82b/83b (`internal/evaglyph`, `internal/metadatavalidation`, `internal/genericsegmentation`, `internal/fingerprintv2`), so Line A's units are identical to the units Fingerprint V2 already measures — required for `GRAMMAR_F2_APPLICABILITY.tsv` to be meaningful at all — rather than inventing a second, incompatible tokenization.

4. **How are G1/G2/G3 divided?** By which observable variables are readable: G1 TOKEN-internal only; G2 adds bounded local TOKEN context, never manuscript hierarchy; G3 adds applicability-audited STRUCTURAL_STATE variables (`TASK85_DESIGN.md` section 5). G1⊂G2⊂G3 by construction in `GRAMMAR_MODEL_REGISTRY.tsv` (M6 backs off to a frozen G1; M7 is built on a frozen G1+G2).

5. **Which model classes are admitted?** M0-M7 (`GRAMMAR_MODEL_REGISTRY.tsv`), plus a non-primary neural auxiliary-upper-bound row. No class was chosen for expected VM fit.

6. **Why are these classes different enough for meaningful comparison?** Each adjacent pair in the M0-M7 ladder differs in exactly one structural capability (frequency-only -> fixed-order history -> variable-order history -> deterministic automaton -> probabilistic automaton -> explicit component/rule structure -> token-sequence context -> structural conditioning), so a pairwise comparison isolates one capability's contribution rather than conflating several (`TASK85_DESIGN.md` section 8).

7. **Which baselines are used?** B0-B5 (`GRAMMAR_BASELINE_REGISTRY.tsv`) plus the separate `MFC0`-`MFC3` message-free calibration battery, run before any VM data is fit.

8. **How is DEVELOPMENT/VALIDATION/HELDOUT structured?** A fixed, seed-free, (Currier,section)-stratified, physical-leaf-level split: 70/18/15 leaves, ≈68/19/13% of TOKENs respectively, byte-identical across two independent regenerations (`GRAMMAR_CORPUS_SPLIT.tsv`, `GRAMMAR_CORPUS_SPLIT_MANIFEST.json`, `TASK85_DESIGN.md` section 7).

9. **How is structural leakage prevented?** The split unit is the whole physical leaf (recto+verso+foldout sides always co-assigned), so no locus/line-level near-duplicate content can appear on both sides of a partition boundary; only leaf-to-leaf adjacency crosses boundaries (48/96 adjacent pairs), an accepted coarser granularity (`TASK85_DESIGN.md` section 7, "Leakage control").

10. **How is cross-transcription robustness checked?** The identical discovery/validation protocol runs on ZL3b and IT2a; primary conclusions are labeled `TRANSCRIPTION_STABLE`, `DIRECTION_STABLE`, or `TRANSCRIPTION_SENSITIVE` (`GRAMMAR_VALIDATION_CONTRACT.md` section 2). The split's leaf-level partition assignment is identical across both transcriptions by construction.

11. **Which predictive metrics are primary?** Held-out NLL, cross-entropy, unseen-TOKEN probability, predictive calibration, negative-discrimination AUC (PM1,PM2,PM4,PM5,PM6); perplexity (PM3) is primary only when the scored unit is held fixed across compared models; training likelihood (PM0) is never primary (`GRAMMAR_METRIC_REGISTRY.tsv`).

12. **Which F2 metrics/families are applicable?** All 33 frozen Fingerprint V2 CORE/SUPPORTING metrics were audited (`GRAMMAR_F2_APPLICABILITY.tsv`): 7 are G1-natural, 1 is G2-natural, 22 are G3-natural, 3 are `NOT_APPLICABLE` (VOYNICH_ONLY_CONTEXT). Of the 13 CORE metrics, 10 are `GENERATION_APPLICABLE`/discriminating and 3 are `skeleton_only` (excluded from generation-validation evidence — see question 16 below and `TASK85_DESIGN.md` section 10).

13. **How is complexity measured?** `Complexity(G) = StructureCost(G) + LexiconCost(G) + ExceptionCost(G)`, a two-part MDL-style code with coding assumptions fixed before fitting (`GRAMMAR_COMPLEXITY_CONTRACT.md`).

14. **How is the lexicon accounted for?** `LexiconCost(G)` charges every explicitly stored TOKEN/COMPONENT/rule-support table a Shannon code over its own DEVELOPMENT frequency; a model is never called compact for having a small `StructureCost` while hiding size in an unaccounted lexicon (`GRAMMAR_COMPLEXITY_CONTRACT.md` section 2).

15. **How are exceptions accounted for?** `ExceptionCost(G)` charges every hand-added special case its own `LexiconCost` plus a fixed 1-bit exception-flag overhead, in one shared `{RULE, LEXICAL_EXCEPTION, STRUCTURAL_EXCEPTION}` representation (`GRAMMAR_COMPLEXITY_CONTRACT.md` section 3).

16. **What does `GRAMMAR_SUFFICIENT` mean?** Predictive adequacy (HELDOUT improvement over baselines beyond the calibration-battery null spread, on both transcriptions at least direction-stable) AND structural adequacy (the frozen family-level gate met on `GENERATION_APPLICABLE`, non-`skeleton_only` metrics at the model's own level) AND seed/partition stability. High likelihood alone never suffices (`GRAMMAR_VALIDATION_CONTRACT.md` section 6).

17. **What does `GRAMMAR_MINIMAL` mean?** `G_min = argmin Complexity(G)` subject to `PredictiveAdequacy(G) AND StructuralAdequacy(G)`, over the frozen `PRIMARY_CANDIDATE` set only (`GRAMMAR_VALIDATION_CONTRACT.md` section 6) — not the fewest rules in the abstract, but the least complex model meeting both frozen gates.

18. **How will contextual gain be measured?** `CONTEXT_INFORMATION_GAIN` = the complexity-adjusted PM1/PM2 improvement of a frozen G2 (M6) over its own frozen G1 parent, on HELDOUT, via ablation A2 (remove local context) (`TASK85_DESIGN.md` section 5, `GRAMMAR_ABLATION_REGISTRY.tsv`).

19. **How will structural gain be measured?** `STRUCTURAL_INFORMATION_GAIN` = the complexity-adjusted PM1/PM2 improvement of a frozen G3 (M7) over its own frozen G2 parent, on HELDOUT, via ablation A4 (remove hierarchy) (`TASK85_DESIGN.md` section 5, `GRAMMAR_ABLATION_REGISTRY.tsv`).

20. **Which ablations are mandatory?** Remove token formation (A1), remove local context (A2), remove line position (A3), remove hierarchy (A4), remove lexical memory (A5), remove state (A6) — each mapped only to the model classes where it is meaningful (`GRAMMAR_ABLATION_REGISTRY.tsv`).

21. **How must a grammar generate synthetic corpora?** `G + seed_i -> synthetic corpus_i`; at G3 scope, TOKEN values are sampled onto a fixed, real, borrowed STRUCTURAL_STATE skeleton (the layout of whichever partition is being compared against), never a skeleton the grammar invents — that is Line B's job (`TASK85_DESIGN.md` sections 10 and 12, `GRAMMAR_VALIDATION_CONTRACT.md` section 11).

22. **How are seeds/scales/replicates chosen?** From a stability/convergence diagnostic on DEVELOPMENT (stop once additional replicates change the reported statistic by less than a pre-registered tolerance), never by desired significance; scale is matched-size plus at least one additional scale for convergence checking (`GRAMMAR_VALIDATION_CONTRACT.md` sections 7-8).

23. **How is seed cherry-picking prevented?** Every replicate index 0..R-1 is enumerated in advance from a pure function of `(model_id, hyperparameters, transcription, partition, replicate index)` and every result is reported; none may be discarded post hoc (`GRAMMAR_VALIDATION_CONTRACT.md` section 7).

24. **What will count as information residual?** `V = G_min + R`; `R` is the information in HELDOUT not predicted by frozen `G_min`, represented as a surprisal sequence, coding residual, prediction-error sequence, or sampled-choice residual (Task89 selects and freezes exactly one as primary); never called plaintext (`GRAMMAR_RESIDUAL_CONTRACT.md` sections 1-2).

25. **What residual nulls are planned?** The same residual-structure test statistic computed on the residual of a matched-scale message-free calibration corpus (`MFC0`-`MFC3`); a residual test without this matched null cannot be interpreted (`GRAMMAR_RESIDUAL_CONTRACT.md` section 4).

26. **How is grammar discovery separated from the message-free experiment?** `G_min` is frozen by Task86-88 using only predictive/structural adequacy on DEVELOPMENT/VALIDATION/HELDOUT; only afterward may Task89 generate message-free corpora from it, and `G_min` may never be selected or reselected by how well those samples match VM (`TASK85_DESIGN.md` section 14, `GRAMMAR_RESIDUAL_CONTRACT.md` section 5).

27. **How is the semantics firewall enforced?** No Task85-89 artifact is interpreted through a proposed translation, candidate plaintext, semantic label, presumed subject matter, or illustrated-page meaning; stated once centrally and referenced by every contract (`TASK85_DESIGN.md` section 14; `PHASE3_LINE_A_RESEARCH_QUESTIONS.md`).

28. **How is the Fontana/mechanism firewall enforced?** No grammar representation is chosen or excluded for resembling (or not resembling) a Fontana mechanism, and no Task81-83r mechanism-fit result is used to set a Line A class or threshold; both stated centrally in `TASK85_DESIGN.md` section 14.

29. **When does HELDOUT open?** Not until the task first needing it (starting with Task86) issues its own `GRAMMAR_MODEL_SELECTION_FROZEN` sentinel recording git commit, the checksums of every frozen Task85 artifact it used, its selected model/hyperparameters, generation settings, and seed contract (`TASK85_DESIGN.md` section 15, `TASK86_HANDOFF.md`).

30. **What invalidates the experiment after HELDOUT opens?** A scientific-definition change (redefining a metric, gate, or class after seeing HELDOUT) invalidates the confirmatory run; an implementation bug with unchanged scientific semantics is instead documented, fixed, regression-tested, and all affected models recomputed symmetrically (`TASK85_DESIGN.md` section 15).

31. **Is the contract ready for Task86?** Yes — see `TASK86_READY = SUPPORTED` below and `TASK86_HANDOFF.md`.

## Verdicts (task85 section 61)

```
GRAMMAR_UNIT_DEFINITION        = SUPPORTED
G1_G2_G3_SEPARATION            = SUPPORTED
HIERARCHICAL_SPLIT_VALID       = SUPPORTED
STRUCTURAL_LEAKAGE_CONTROL     = SUPPORTED
CROSS_TRANSCRIPTION_PROTOCOL   = SUPPORTED
MODEL_SPACE_PREDEFINED         = SUPPORTED
BASELINE_SPACE_PREDEFINED      = SUPPORTED
COMPLEXITY_ACCOUNTING          = SUPPORTED
LEXICON_COST_ACCOUNTED         = SUPPORTED
GRAMMAR_SUFFICIENCY_DEFINED    = SUPPORTED
GRAMMAR_MINIMALITY_DEFINED     = SUPPORTED
GENERATIVE_VALIDATION_DEFINED  = SUPPORTED
F2_VALIDATION_SPACE_DEFINED    = PARTIAL
RESIDUAL_CONTRACT_DEFINED      = PARTIAL
SEMANTICS_FIREWALL             = SUPPORTED
PHASE2_MECHANISM_FIREWALL      = SUPPORTED
DETERMINISM_CONTRACT           = SUPPORTED
TASK86_READY                   = SUPPORTED
```

`F2_VALIDATION_SPACE_DEFINED = PARTIAL`: fully defined and audited for G1 and G3 (29 of 33 frozen metrics have a definite, well-justified role); genuinely thin at G2 (only one frozen metric is G2-natural), which is a real gap in the *existing* Fingerprint V2 battery, not a gap in this audit — flagged forward to Task87 in `TASK86_HANDOFF.md`.

`RESIDUAL_CONTRACT_DEFINED = PARTIAL`: the residual's representation candidates, estimator requirements, and required test scales are fully fixed (`GRAMMAR_RESIDUAL_CONTRACT.md`), but per task85 section 42 the final estimator choice is explicitly deferred to (at latest) Task89 and is not, and should not be, frozen by Task85 itself.

## Success criterion (task85 section 62)

This design is written so that Task86 can be executed by a different researcher without a new substantive scientific-design decision: `TASK86_HANDOFF.md` fixes the corpus, partitions, unit definition, model classes, hyperparameter ranges, baselines, metrics, complexity accounting, generation protocol, seeds, validation rules, model-selection rule, and freeze condition for G1 specifically.
