# Phase I method index

This is a routing document. Definitions are short and defer to frozen experiment designs and reports; similarly named metrics are not assumed equivalent across representations.

## Corpus parsing and EVA glyph handling

**Purpose.** Convert declared transcriptions into glyph, token, line and block units. **Definition.** Experiment-specific parsing with shared EVA composite handling where adopted. **Assumptions.** Transcription symbols and separators are observational units, not necessarily letters/words. **Controls.** Corpus hashes, parser tests, explicit boundary modes. **Implementation.** `internal/evaglyph` and baseline parsers. **Used by.** Tasks58–61 and later shared metric families.

## Marginal inventory, frequency and vocabulary growth

**Purpose.** Describe types, lengths, ranks, hapax and `V(N)`. **Definition.** Counts and curves on a declared token/glyph representation. **Assumptions.** Units are consistently parsed within a run. **Controls.** Natural/transformed corpora, matched checkpoints and token shuffles where applicable. **Implementation.** baseline Stages2/3/28 and related `internal/` packages. **Used by.** Baseline, Task55, Task66 fingerprint targets.

## Mutual information with shuffle correction

**Purpose.** Measure adjacent token or edge-glyph dependence beyond local marginals. **Definition.** Observed adjacent MI minus mean MI after within-line permutations; share normalizes by the declared entropy cap. **Assumptions.** The shuffle preserves line membership and token frequencies while destroying order. **Controls.** 100 shuffles in Task58, natural corpora and transformed controls. **Implementation.** `research/phase1/rozanova-temerev`, deterministic sorted-key estimator in `internal/evaglyph`. **Used by.** Task58 and Task66 targets.

## Positional specialization

**Purpose.** Detect glyphs concentrated at initial, medial or final token positions. **Definition.** Frozen frequency/share thresholds with within-token permutation inference and FDR. **Assumptions.** Coarse positions are meaningful analytical strata. **Controls.** Natural prose, corrected position-independent homophony, position-dependent and structured positive controls. **Implementation.** `research/phase1/glyph-position-analyze`, `internal/evaglyph`. **Used by.** Task59, Task66.

## Shannon conditional character entropy

**Purpose.** Quantify glyph-stream predictability. **Definition.** Plug-in `h_n = H(X_i | X_{i-n+1}^{i-1})`, with `h2` primary and explicit coverage. **Assumptions.** Inventory, stream construction and boundary policy define the estimand. **Controls.** Continuous/boundary-symbol/within-token modes, glyph and token shuffles, natural/homophony controls. **Implementation.** `research/phase1/character-entropy-analyze`, `internal/characterentropy`. **Used by.** Task61, Tasks64–67 via authoritative joins.

## Exact repetition, runs and edit distance

**Purpose.** Measure exact copies and local form neighborhoods. **Definition.** Adjacent equality, run length and Levenshtein distance/operation decomposition. **Assumptions.** Edit cost 1 is formal glyph similarity, not linguistic derivation. **Controls.** Global and line-preserving shuffles, frequency/length matching, natural corpora and homophony dose response. **Implementation.** `research/phase1/token-repetition-analyze` and shared edit utilities. **Used by.** Task60 and later fingerprint/model tests.

## Structural inverse-transposition objective

**Purpose.** Search narrow inverse permutations without a lexical oracle. **Definition.** Frozen candidate-local combination of transition, relation and sequence-2/3 structural metrics. **Assumptions.** A successful tested inverse raises the selected structure measures. **Controls.** Known T2/T4/T8 transforms, blind recovery, natural width-2 audit, held-out Voynich split and token-permutation null. **Implementation.** `research/phase1/inverse-transposition-search`, `research/phase1/voynich-validation`, `internal/inversetransposition`. **Used by.** Tasks54/54b.

## Inverse-homophony validation gate

**Purpose.** Determine whether latent homophone classes and planted structure can be recovered before any Voynich use. **Definition.** Frozen class recovery, pair discrimination and structural recovery thresholds. **Assumptions.** Synthetic transfer is a prerequisite, not evidence about Voynich. **Controls.** Development/validation splits, random baseline and 11 synthetic corpora. **Implementation.** `research/phase1/inverse-homophony` and `internal/inversehomophony`. **Used by.** Task57 Phase A; Phase B was not run.

## Held-out token-formation modeling

**Purpose.** Test whether local token grammar explains form geometry. **Definition.** Candidate generators selected on validation cross entropy and evaluated on contiguous held-out blocks. **Assumptions.** Token-internal model class is deliberately bounded. **Controls.** Length/frequency, copy/mutate, structured-token and natural controls; novel-token validation. **Implementation.** `research/phase1/token-formation-analyze`, `internal/tokenformation`. **Used by.** Task62 and as a control in Tasks64–66.

## Matched local-transition analysis

**Purpose.** Isolate adjacency-specific form similarity. **Definition.** Compare adjacent pairs with nonadjacent same-line controls matched on token-length pair; keep exact and distance-1 effects separate. **Assumptions.** Matching removes specified opportunity differences, not all confounding. **Controls.** Global/within-line shuffles and held-out generative preservation. **Implementation.** `research/phase1/token-transition-analyze`, `internal/tokentransition`. **Used by.** Task63, replicated/conditioned by Task64.

## Frozen-sequence replication and higher-order tests

**Purpose.** Confirm previously discovered sequences without rediscovery. **Definition.** Frequency-null FDR and cross-block replication; for n≥3, test `P(C|A,B)` against `P(C|B)` with conditional permutations and LOBO. **Assumptions.** Candidate inventory remains frozen; low counts constrain inference. **Controls.** Frequency and Markov nulls, block eligibility, jackknife and positional stratification. **Implementation.** baseline higher-order/replicated-local stages. **Used by.** Baseline sequence confirmation.

## Line/local scale comparison

**Purpose.** Test whether physical lines uniquely organize similar forms. **Definition.** Length-matched pair-rate effects for adjacency, lines, shifted line-sized blocks, fixed windows and pages. **Assumptions.** Linear token order approximates the tested local scale. **Controls.** Same-page/different-page matches and line-membership destruction. **Implementation.** `research/phase1/line-regime-analyze`, `internal/lineregime`. **Used by.** Task64.

## Local-regime topology

**Purpose.** Distinguish local decay, drift, state-like clustering, metadata association and recurrence. **Definition.** Empirical lag curves, stable cluster sweeps, dwell/transitions, change points and nearest-distant comparisons. **Assumptions.** Cluster labels are descriptive states. **Controls.** Stationary/smooth/discrete/mixed simulations, natural corpora, Task62 G-only and corrected multiple-opportunity recurrence null. **Implementation.** `research/phase1/local-regime-topology-analyze`, `internal/localregimetopology`. **Used by.** Task65.

## Mechanism-grid scoring and ablation

**Purpose.** Identify necessary components within a frozen synthetic grid. **Definition.** Family-normalized coverage of 17 frozen targets, validation/held-out scoring, Pareto comparison and component removal. **Assumptions.** Necessity is conditional on M0–M11, thresholds and source corpora. **Controls.** Identity, monoalphabetic, homophony, state/boundary nulls, ablations and corpus transfer. **Implementation.** `research/phase1/mechanism-space-analyze`, `internal/mechanismspace`. **Used by.** Task66. Per-job seed/config hash is derivable but absent from result rows.

## Recoverability, ambiguity and damage

**Purpose.** Separate mathematical reversibility, exact recovery, retained information, preimage ambiguity and robustness. **Definition.** Frozen known-plaintext encode/decode; decoder knowledge levels; blockwise ambiguity; single-error propagation and resynchronization. **Assumptions.** Results apply to implemented synthetic candidates and declared decoder knowledge. **Controls.** Identity/bijection/homophony, three plaintext corpora, shuffled input, short blocks, state knowledge and error families. **Implementation.** `research/phase1/recoverability-analyze`, `internal/recoverability`. **Used by.** Task67. Only measured `0e8b3da…` artifacts are current.

## Literature claim and control audit

**Purpose.** Compare estimands, evidence quality and hypothesis discrimination without vote counting. **Definition.** Source-level catalog, claim registry, control/replication/contradiction matrices and metric crosswalk. **Assumptions.** “No precedent found” is bounded by the logged search. **Controls.** Primary-source priority, explicit publication status and separation of exact/partial/no methodological match. **Implementation.** `docs/literature/`. **Used by.** Tasks68/69 and this synthesis.
