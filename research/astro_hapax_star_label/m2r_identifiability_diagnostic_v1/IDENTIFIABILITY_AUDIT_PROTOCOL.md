# IDENTIFIABILITY AUDIT PROTOCOL: M2R IDENTIFIABILITY DIAGNOSTIC v1
**Document ID:** `M2R-IDENTIFIABILITY-AUDIT-PROTOCOL-V1`  
**Status:** `LOCKED_PRE_EXPERIMENT`  
**Date:** 2026-09-16  
**Package:** `research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1`  
**Rule:** Frozen prior to diagnostic execution. No post-hoc modification allowed.

---

## 1. Context and Objective

The M2R-v2 engine successfully satisfied all 25 unit and regression tests, yet suffered a catastrophic failure during independent hidden validation:
* Latent-rule precision: 29.72% (gate >= 80%)
* Latent-rule recall: 10.00% (gate >= 75%)
* Train assignment accuracy: 2.21% (gate >= 75%)
* Held-out coverage: 0.00% (gate >= 70%)
* Null control execution: 0/8 completed (aborted due to memory accounting failure)
* Order invariance and checkpoint-resume identity: unconfirmed due to incomplete runs.

Crucially, M2R-v2 demonstrated positive normalized compression gain despite selecting almost completely spurious rules and assignments. This indicates that the system optimizes for compact, but entirely ungrounded, artifactual structure.

The objective of this diagnostic audit is to definitively determine the mathematical and structural causes of this failure, and to establish whether developing an M2R-v3 engine for the real restricted STAR LABEL scope (57 star labels across f68r1/f68r2) is scientifically justified.

This audit:
1. Does NOT modify `m2r_real_engine_v1` or `m2r_real_engine_v2`.
2. Does NOT access real Voynich manuscripts, star label transcriptions, or enrichment dictionaries.
3. Uses synthetic datasets and disclosed M2R-v2 hidden datasets exclusively for scientific diagnosis.
4. Operates strictly under this pre-declared protocol.

---

## 2. Testable Hypotheses

The audit formalizes and evaluates six specific hypotheses explaining the failure of M2R:

* **H1: Candidate Generation Bottleneck (`CANDIDATE_GENERATION`)**  
  True latent rules are absent from the candidate pool. The candidate generation mechanism fails due to:
  - Strict minimum independent support requirements ($N \ge 3$) when latent rules occur infrequently;
  - Aggressive pool size truncation (pool cap = 48) combined with heuristic "potential saving" sorting that favors high-frequency spurious character n-grams over true morphological rules;
  - Positional role assignment mismatches or segmentation boundaries.

* **H2: Search Space Optimization Failure (`SEARCH_FAILURE`)**  
  True latent rules are present in the candidate pool, and the true model achieves a competitive objective score, but the bounded beam search fails to reach the global optimum due to:
  - Severe search depth limits (8 iterations, adding at most 1 rule per iteration, capping model size at 8 rules when true models contain 15 rules);
  - Narrow beam width (beam = 3, max states = 96) leading to greedy premature convergence;
  - Elimination of necessary intermediate rules that do not provide immediate incremental gain.

* **H3: Minimum Description Length Scoring Failure (`SCORING_FAILURE`)**  
  The bounded search optimizer reaches the global optimum of the scoring function (or close to it), but the objective function itself prefers spurious or partial models over the ground truth:
  - Heavy fixed model complexity penalty ($10 + L(\text{src}) + L(\text{tgt})$ bits per rule) penalizes complete 15-rule models relative to compact 2-4 rule approximations;
  - Alignment gap and literal exception costs fail to sufficiently penalize incorrect pairings;
  - Spurious rules combining frequent unigrams yield higher normalized compression gain than the true latent rules.

* **H4: Anonymous Assignment Non-Identifiability (`ASSIGNMENT_AMBIGUITY`)**  
  Under the anonymous setting (unpaired bags of terms and labels), the bipartite matching problem lacks sufficient signal to identify the ground-truth pairing:
  - Inherent permutation symmetries (e.g., Cartesian combinations of affixes) allow multiple distinct bijections to produce identical surface statistics and identical MDL scores;
  - The objective landscape over assignments is flat or multimodal, leading to near-zero margins between correct and arbitrary pairings;
  - Hungarian matching arbitrarily selects among degenerate optima based on alphabetical tie-breaking.

* **H5: Sample Size and Hapax Starvation (`DATA_STARVATION`)**  
  The anonymous morphology induction problem is statistically under-determined at small sample sizes (such as the 57 labels of f68r1/f68r2) characterized by high hapax legomena rates and limited combinatorial overlap:
  - Identifiability requires $N \gg 57$ distinct co-occurring pairs to distinguish true morphological units from chance statistical regularities;
  - At $N \le 57$, the probability of observing every morphological component $\ge 3$ times is low, leaving rules unsupported.

* **H6: Model Class Misspecification (`MODEL_MISSPECIFICATION`)**  
  The mathematical assumptions of the M2R model class (strict monotonic prefix/medial/suffix roles without infixes, circumfixes, deletions, or context conditioning) cannot accurately represent the true data-generating process.

---

## 3. Oracle Decomposition Design

To isolate each subsystem, all diagnostic synthetic datasets are evaluated under six standardized oracle access modes:

| Mode | Assignment | Rules | Candidate Pool | Diagnostic Purpose |
| :--- | :--- | :--- | :--- | :--- |
| **A. Normal** | Unknown | Unknown | Standard (48) | Baseline inference: reproduces full unguided M2R task. |
| **B. Oracle Assignment** | True (Fixed) | Unknown | Standard (48) | Evaluates rule induction in the absence of assignment ambiguity. |
| **C. Oracle Rules** | Unknown | True (Fixed) | Oracle Rules | Evaluates bipartite assignment identifiability when rules are perfectly known. |
| **D. Oracle Candidate Pool** | Unknown | Unknown | True Rules Forced In | Evaluates whether search succeeds when candidate generation is perfect. |
| **E. Oracle Assignment + Pool** | True (Fixed) | Unknown | True Rules Forced In | Evaluates search and scoring without assignment ambiguity or candidate loss. |
| **F. Full Oracle Ceiling** | True (Fixed) | True (Fixed) | Oracle Rules | Measures theoretical upper bound on MDL compression gain and coverage. |

All six modes apply identical MDL description length definitions and alignment mechanics.

---

## 4. Diagnostic Metric Framework

The audit measures the following formal metrics across all experiments:

### 4.1. Candidate Generation Metrics
* `TRUE_RULE_CANDIDATE_RECALL`: Fraction of true latent rules present in the generated candidate pool.
* `EQUIVALENCE_AWARE_CANDIDATE_RECALL`: Fraction of true rules represented either exactly or by functionally identical sub-units.
* `ROLE_RECALL_{WHOLE, INITIAL, MEDIAL, FINAL}`: Candidate recall stratified by morphological role.
* `SUPPORT_RECALL_{1, 2, 3+}`: Candidate recall stratified by true observed support in the dataset.
* `LENGTH_RECALL_{1, 2, 3, 4+}`: Candidate recall stratified by unit character length.
* `POOL_RECALL_UPPER_BOUND`: Theoretical maximum rule recall achievable given candidate pool truncation.

### 4.2. Rule and Assignment Recovery Metrics
* `LATENT_RULE_PRECISION`: $|R_{\text{inferred}} \cap R_{\text{true}}| / |R_{\text{inferred}}|$.
* `LATENT_RULE_RECALL`: $|R_{\text{inferred}} \cap R_{\text{true}}| / |R_{\text{true}}|$.
* `ROLE_AGNOSTIC_PRECISION / RECALL`: Match rates ignoring positional role tags.
* `STRICT_ASSIGNMENT_ACCURACY`: Fraction of terms assigned to their exact ground-truth label.
* `EQUIVALENCE_AWARE_ASSIGNMENT_ACCURACY`: Assignment accuracy after quotienting by automorphism/permutation symmetries.
* `TOP_K_ASSIGNMENT_ACCURACY`: Ground-truth assignment containment within top-$k$ alternative optima.
* `ASSIGNMENT_MARGIN`: Difference in objective value between the optimal assignment and the next best distinct assignment.
* `ASSIGNMENT_ENTROPY`: Shannon entropy of candidate pairing probabilities.

### 4.3. Search and Scoring Metrics
* `BOUNDED_SEARCH_EXACT_OPTIMUM_RATE`: Proportion of small instances where bounded search matches the exact exhaustive global optimum.
* `EXHAUSTIVE_TRUTH_IS_GLOBAL_OPTIMUM`: Whether the ground-truth model achieves the global minimum description length among all admissible models.
* `MDL_GAP_TRUTH_VS_OPTIMUM`: $\text{MDL}(M_{\text{truth}}) - \text{MDL}(M_{\text{global\_opt}})$.
* `FALSE_OPTIMUM_COMPONENTS`: Decomposition of description length differences (Rule Code vs Assignment Code vs Alignments vs Gaps).
* `EQUIVALENT_OPTIMA_COUNT`: Number of distinct rule/assignment configurations achieving the exact optimal score.

### 4.4. Generalization and Resource Metrics
* `TRAIN_COVERAGE` / `HELDOUT_COVERAGE`: Fraction of tokens explained without literal gaps.
* `TRAIN_GAIN` / `HELDOUT_GAIN`: Normalized compression gain relative to uncompressed baseline.
* `PEAK_MEMORY_RSS_MB`: Process-isolated resident set size measured directly without parent inheritance artifact.
* `SEARCH_RUNTIME_SECONDS`: Wall-clock execution time per instance.
* `CHECKPOINT_RESUME_IDENTITY`: Byte-exact and semantic identity of continuous vs interrupted runs.
* `ORDER_INVARIANCE`: Deterministic identity under input shuffling, reverse ordering, and ID relabeling.

---

## 5. Factorial Scaling Study Design

The scaling investigation systematically varies the following orthogonal experimental factors across $\ge 20$ independent random seeds per configuration:

1. **Dataset Size ($N$):** $N \in \{6, 9, 12, 20, 29, 57, 100\}$.
2. **Morphological Complexity ($|R_{\text{true}}|$):** 4 rules (2x2), 6 rules (3x3), 9 rules (3x3x3), 15 rules (5x5x5).
3. **Minimum Support Threshold:** 1, 2, and 3 distinct examples.
4. **Hapax Legomena Proportion:** 0% (dense Cartesian product), 50%, 80%, 100% (all words appear exactly once).
5. **Surface Noise / Distractor Rate:** 0%, 10%, 20% unmatchable terms/labels.
6. **Alphabet Redundancy & Unit Lengths:** Lengths 1 to 4 code points.

### Real Scope Benchmark
A dedicated benchmark specifically simulates the restricted real scope:
- $N = 57$ pairs (matching the 57 star labels of f68r1/f68r2);
- High hapax rate ($>90\%$ unique occurrences);
- Alphabet inventory size matching Voynich star label transcripts;
- 2-page partition structure.

---

## 6. Null Controls and Resource Accounting

Six deterministic null controls must complete without execution failure:
1. `shuffled_assignments`: Terms and labels identical to positive corpus, but pairings randomly permuted.
2. `shuffled_labels`: Target labels replaced by character-level shuffled sequences.
3. `shuffled_terms`: Source terms replaced by character-level shuffled sequences.
4. `morphology_destroying`: Random character strings preserving length distribution only.
5. `frequency_preserving`: Unigram-preserving permutations destroying higher-order structure.
6. `length_preserving`: Bigram/length-matched synthetic nulls.

### Technical Memory Optimization Policy
To prevent the Linux glibc/kernel `ru_maxrss` parent-inheritance defect observed in M2R-v2:
* Worker RSS must be measured via process-isolated `/proc/self/status` `VmRSS` and Python `tracemalloc`, not lifetime parent peak;
* Parent runner must use streaming JSON generation and explicit garbage collection between runs;
* The 512 MiB per-task limit must be strictly enforced on true worker consumption.

---

## 7. Pre-declared Success and Stopping Criteria

The following thresholds are fixed and cannot be modified after running experiments:

| Metric Criterion | Target Threshold | Interpretation if Failed |
| :--- | :--- | :--- |
| `TRUE_RULE_CANDIDATE_RECALL` | $\ge 0.95$ | Candidate generation is a primary bottleneck. |
| `ORACLE_ASSIGNMENT_RULE_RECALL` | $\ge 0.80$ | Rules cannot be recovered even when pairings are known. |
| `ORACLE_RULE_ASSIGNMENT_ACCURACY` | $\ge 0.80$ | Assignments are non-identifiable even with perfect rules. |
| `BOUNDED_SEARCH_EXACT_OPTIMUM_RATE` | $\ge 0.95$ | Bounded search optimizer fails to find scoring optima. |
| `EXHAUSTIVE_TRUTH_IS_GLOBAL_OPTIMUM` | $\ge 0.90$ | Scoring function is misspecified / prefers false structures. |
| `NULL_CONTROLS_COMPLETE` | $100\%$ complete | Resource accounting or execution failure. |
| `POSITIVE_MEDIAN_MINUS_NULL_P95` | $\ge 0.05$ | System cannot separate signal from null noise. |
| `ORDER_INVARIANCE` | PASS (100%) | Optimizer exhibits non-deterministic tie-breaking. |
| `CHECKPOINT_IDENTITY` | PASS (100%) | State serialization is lossy or non-deterministic. |

---

## 8. Failure Classification and Decision Rules

At completion, the audit must select exactly ONE primary failure mode and ONE decision recommendation:

### Primary Failure Mode Classification
1. `CANDIDATE_GENERATION`: True rules absent from candidate pool (`CANDIDATE_RECALL < 0.95`), but rules recover when forced in.
2. `SEARCH`: True rules in pool, true model is global optimum, but bounded search fails (`BOUNDED_SEARCH_RATE < 0.95`).
3. `SCORING`: Bounded search finds optimum, but optimum is NOT the truth (`EXHAUSTIVE_TRUTH_IS_GLOBAL_OPTIMUM < 0.90`).
4. `ASSIGNMENT`: Rules recoverable under oracle assignment, but anonymous assignment fails due to symmetry/ambiguity.
5. `NON_IDENTIFIABILITY`: Multiple distinct solutions produce identical surface data and identical scores.
6. `MODEL_CLASS`: Truth cannot be represented within M2R's role/composition model.
7. `MIXED`: Multiple compounding failure points.

### M2R-v3 Final Recommendation
* `PROCEED_CANDIDATE_GENERATION`: Feasible if and only if only candidate generation failed.
* `PROCEED_ASSIGNMENT`: Feasible if and only if only assignment resolution failed.
* `PROCEED_SEARCH`: Feasible if and only if only search budget failed.
* `PROCEED_SCORING`: Feasible if and only if only objective formulation failed.
* `PROCEED_MODEL_CLASS`: Feasible if and only if model architecture was inadequate.
* `DO_NOT_PROCEED`: Selected if the problem is mathematically non-identifiable in the anonymous setting, or if feasibility strictly requires data volume or repetition properties absent from the 57 real STAR LABELs.
* `INCONCLUSIVE`: Selected only if experiments abort or yield statistically contradictory results.

### Real Scope Feasibility Constraint
If positive identifiability requires data volume or repetition frequency substantially exceeding the real 57 star labels:
$$\text{REAL\_SCOPE\_FEASIBILITY} = \text{NOT\_SUPPORTED}$$
$$\text{REAL\_DATA\_SEARCH\_AUTHORIZED} = \text{NO}$$
No execution on real data may be authorized under any circumstances.
