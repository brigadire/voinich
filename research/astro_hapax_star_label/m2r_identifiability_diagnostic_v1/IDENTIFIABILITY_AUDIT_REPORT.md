# M2R IDENTIFIABILITY DIAGNOSTIC AUDIT REPORT
**Package:** `research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1`  
**Date:** 2026-09-16  
**Status:** `AUDIT_COMPLETE`  
**Reference Protocol:** `IDENTIFIABILITY_AUDIT_PROTOCOL.md`

---

## Executive Summary

Following the hidden validation failure of M2R-v2 (where the model achieved positive compression gain despite a 10% rule recall and 2.21% assignment accuracy), an independent, reproducible identifiability audit was conducted. The audit performed an oracle decomposition (Modes A–F), a complete audit of the candidate generation pool, exact exhaustive search comparisons on small synthetic tasks, symmetry and automorphism equivalence analysis, assignment identifiability measurements, factorial scaling experiments over multiple independent seeds, and full execution of all null controls.

The audit establishes that the failure of M2R-v2 was not a simple implementation bug, but the result of **compounding structural failure modes** (`PRIMARY_FAILURE_MODE=MIXED`):
1. **Candidate Generation Bottleneck:** 51.85% of true latent rules never enter the candidate pool (`TRUE_RULE_CANDIDATE_RECALL=0.4815`) due to premature pruning (pool cap = 48) and support thresholding.
2. **Search Capacity Deficit:** The bounded beam search (width 3, 8 iterations, max 12 rules) can discover at most 8 rules, strictly preventing discovery of the 15-rule generating system. In exhaustive small tasks, bounded search matched the exact global optimum in only 37.50% of cases (`BOUNDED_SEARCH_EXACT_OPTIMUM_RATE=0.3750`).
3. **Scoring Misspecification in the Presence of Hapaxes:** Under the Minimum Description Length (MDL) objective, full 15-rule models pay a heavy rule complexity penalty ($10 + L(\text{src}) + L(\text{tgt})$ bits per rule). When hapax legomena or infrequent affixes are present, the objective prefers compact, spurious 2–4 rule approximations over the ground truth (`EXHAUSTIVE_TRUTH_IS_GLOBAL_OPTIMUM=MIXED`).
4. **Anonymous Permutation Non-Identifiability:** In an anonymous setting without anchor pairs, Cartesian combinations of affixes yield an orbit of symmetric bijections achieving identical surface bags and identical MDL scores (`EQUIVALENT_OPTIMA_MEDIAN=2.0`, up to 4–36 equivalent optima). Shuffled assignments achieve identical compression to true assignments.
5. **Real Scope Starvation:** In synthetic benchmarks specifically calibrating the real restricted scope of 57 Voynich star labels (f68r1/f68r2) with high hapax rates, rule recall collapsed to **1.70%** and strict assignment accuracy collapsed to **0.60%** across 20 independent seeds.

**Final Determination:**  
`REAL_SCOPE_FEASIBILITY=NOT_SUPPORTED`  
`M2R_V3_RECOMMENDATION=DO_NOT_PROCEED`  
`REAL_DATA_SEARCH_AUTHORIZED=NO`

---

## 1. Oracle Decomposition Analysis (Section 2)

All diagnostic synthetic datasets (including disclosed hidden positive datasets `d61001`–`d61006`, calibration `d51001`, and development `d41001`) were evaluated across the six standardized oracle modes defined in the protocol.

### Summary of Hidden Positive Corpus Macro-Averages

| Oracle Mode | Latent Rule Precision | Latent Rule Recall | Strict Assignment Accuracy | Train Compression Gain | Train Token Coverage |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **A. Normal (Unguided)** | 29.72% | 10.00% | 2.21% | 0.272 | 0.035 |
| **B. Oracle Assignment** | 100.00% | 28.89% | 100.00% | 0.230 | 0.027 |
| **C. Oracle Rules** | 100.00% | 100.00% | 100.00% | 0.782 | 0.951 |
| **D. Oracle Candidate Pool** | 100.00% | 26.67% | 18.80% | 0.229 | 0.019 |
| **E. Oracle Assignment + Pool** | 100.00% | 26.67% | 100.00% | 0.205 | 0.019 |
| **F. Full Oracle Ceiling** | 100.00% | 100.00% | 100.00% | 0.782 | 0.951 |

### Key Findings from Decomposition:
1. **Mode A reproduced the failure exactly:** The unguided M2R search reproduces the 29.72% precision, 10.00% recall, and 2.21% assignment accuracy observed during hidden validation.
2. **Mode C demonstrates assignment feasibility under perfect rules:** When true rules are provided, the Hungarian bipartite matching recovers **100.00%** of assignments with a 78.2% compression gain and 95.1% full-token coverage. Bipartite matching ambiguity is zero when rules are correct.
3. **Modes D and E reveal search and scoring ceilings:** Even when true rules are forcibly injected into the candidate pool (Mode D) or when both true rules and true pairings are provided (Mode E), bounded beam search recovers only 26.67% of rules (4 out of 15). The search optimizer stops adding rules because adding further rules incurs more rule complexity penalty than incremental literal savings under partial coverage.

*Data source:* [`ORACLE_EXPERIMENT_MATRIX.tsv`](file:///home/brigadire/devops/workdir/go/voinich/research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/ORACLE_EXPERIMENT_MATRIX.tsv)

---

## 2. Candidate Generation Audit (Section 3)

An exhaustive audit was performed across all 135 ground-truth latent rules in the target positive corpora.

* **Exact True-Rule Candidate Recall:** **0.4815** (65 / 135 true rules present in pool)
* **Equivalence-Aware Candidate Recall:** **0.7037** (95 / 135 true rules present or represented by sub-units)
* **Candidate Recall by Role:**
  - `INITIAL`: 42.22%
  - `MEDIAL`: 60.00%
  - `FINAL`: 42.22%
* **Candidate Recall by Observed Support:**
  - Support $\ge 3$: 48.87%
  - Support = 2: 0.00%
  - Support = 1: 0.00%
* **Upper Bound on Latent Rule Recall:** Conditioned on the candidate pool, latent-rule recall cannot exceed **48.15%**, which is below the mandatory 75% gate.

### Failure Breakdown in Candidate Generation:
1. **Support Thresholding:** In sample sizes of $N=29$ (e.g. `d61001`), rules appearing 1 or 2 times fail the $N \ge 3$ support filter and are discarded immediately.
2. **Pool Cap Truncation (48 Rules):** Across datasets, between 1,500 and 3,000 candidate unit pairs are formed. The heuristic sorting formula ($\text{potential} = \min(n_a, n_b) \cdot (\text{bits} - 2) - (\text{bits} + 10)$) ranks pairs purely on frequency and length. Frequent spurious substring co-occurrences push true rules down to ranks 100–1,600, causing them to be pruned before search begins.

*Data source:* [`CANDIDATE_POOL_AUDIT.tsv`](file:///home/brigadire/devops/workdir/go/voinich/research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/CANDIDATE_POOL_AUDIT.tsv)

---

## 3. Exact Search and Equivalence Class Audit (Sections 4 & 5)

Eight small synthetic instances were subjected to complete exhaustive optimization over all admissible rule subsets (up to 6 rules) and compared against bounded M2R-v2 search:

| Task ID | $N$ Pairs | True Rules | Bounded Gain | Global Optimum Gain | Bounded Matched Global? | Truth is Global Optimum? | Number of Global Optima | Classification |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :--- |
| `T1_cartesian_2x2` | 4 | 4 | 0.146 | 0.146 | **YES** | **YES** | 4 | `NON_IDENTIFIABLE` |
| `T2_cartesian_3x2` | 6 | 5 | 0.147 | 0.240 | **NO** | **YES** | 2 | `SEARCH_FAILURE` |
| `T3_cartesian_3x3` | 9 | 6 | 0.334 | 0.334 | **YES** | **YES** | 2 | `NON_IDENTIFIABLE` |
| `T4_broken_symmetry` | 8 | 6 | 0.198 | 0.284 | **NO** | **YES** | 4 | `SEARCH_FAILURE` |
| `T5_hapax_rich` | 7 | 6 | 0.163 | 0.241 | **NO** | **NO** | 2 | `SCORING_FAILURE` |
| `T6_three_role_composition` | 8 | 6 | 0.055 | 0.212 | **NO** | **YES** | 4 | `SEARCH_FAILURE` |
| `T7_distractor_noise` | 7 | 5 | 0.120 | 0.201 | **NO** | **YES** | 2 | `SEARCH_FAILURE` |
| `T8_shuffled_assignment_null`| 9 | 6 | 0.334 | 0.334 | **YES** | **YES** | 2 | `NON_IDENTIFIABLE` |

### Core Diagnostic Answers:
1. **Does bounded search reach the global optimum?**  
   **NO in 62.5% of tasks** (`BOUNDED_SEARCH_EXACT_OPTIMUM_RATE=0.3750`). In tasks `T2`, `T4`, `T5`, `T6`, and `T7`, bounded beam search converges prematurely to suboptimal rule sets.
2. **Is the true structure the global optimum of the MDL objective?**  
   **MIXED.** In dense Cartesian tasks (`T1`–`T4`, `T6`, `T7`), the true structure achieves the optimal score, but shares that exact score with multiple symmetric alternatives. In `T5` (hapax-rich), a compact 2-rule model achieves gain 0.241, beating the true 6-rule model's gain of 0.219 (`SCORING_FAILURE`).
3. **What false structures win and why?**  
   In hapax settings, partial compact models win because the rule complexity penalty ($10 + \text{lengths}$ bits per rule) exceeds the incremental savings from explaining single hapax tokens.
4. **Shuffled Assignment Invariance (`T8`):**  
   When terms and labels are permuted into a random bipartite matching, the global optimum gain is **0.334**, identical to the ground truth `T3` gain to 16 decimal places. This proves that in the anonymous setting, shuffled assignments preserve identical MDL compression, rendering assignment verification statistically impossible without external supervision.

*Data sources:* [`EXACT_SEARCH_COMPARISON.tsv`](file:///home/brigadire/devops/workdir/go/voinich/research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/EXACT_SEARCH_COMPARISON.tsv), [`EQUIVALENCE_CLASS_ANALYSIS.tsv`](file:///home/brigadire/devops/workdir/go/voinich/research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/EQUIVALENCE_CLASS_ANALYSIS.tsv)

---

## 4. Assignment Identifiability Audit (Section 6)

When true latent rules are known, Hungarian matching was evaluated on all target datasets:
* **Optimal Bipartite Accuracy:** **100.00%** (1.000 across all 9 evaluation datasets).
* **Mean Assignment Margin:** **76.5 to 119.2 bits** per pair between the correct match and the second-best candidate.
* **Mean Pairing Entropy:** **0.000 to 0.098 bits**.
* **Order Permutation Stability:** **PASS** (100% stable under reverse ordering).
* **Surface Noise Perturbation:** **PASS** (100% stable under small numerical weight perturbations).

**Conclusion on Assignment:** The synthetic data contains abundant statistical signal to identify assignments **if and only if the true rules are active**. When rules are unknown or only partially discovered, the margin collapses to zero.

*Data source:* [`ASSIGNMENT_IDENTIFIABILITY.tsv`](file:///home/brigadire/devops/workdir/go/voinich/research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/ASSIGNMENT_IDENTIFIABILITY.tsv)

---

## 5. Scaling Study and Real Scope Feasibility (Section 7)

A factorial scaling experiment was executed across 13 configurations and multiple independent random seeds (including 20 seeds for key configurations and the real restricted scope benchmark):

* **Sample Size Scaling ($N=6$ to $N=100$):**
  - $N=6$: Rule recall = 10.00%, Strict assignment = 0.00%
  - $N=12$: Rule recall = 11.11%, Strict assignment = 11.67%
  - $N=20$: Rule recall = 4.44%, Strict assignment = 0.00%
  - $N=29$: Rule recall = 4.72%, Strict assignment = 3.51%
  - $N=57$: Rule recall = 1.67%, Strict assignment = 0.58%
* **Support Threshold Scaling at $N=57$:** Relaxing support threshold from 3 to 1 or 2 did not improve rule recall (0.00%), because search quickly saturated with spurious unigram rules.
* **Hapax Rate Scaling at $N=57$:** As hapax proportion increases from 0% to 90%, candidate recall drops and assignment accuracy remains $< 1\%$.
* **Real Scope Benchmark (`REAL_SCOPE_57_STAR_LABELS`, 20 seeds):**
  - Simulated parameters: $N = 57$ pairs, 2-page partition, $>90\%$ hapax rate, 10% distractor noise, anonymous bags.
  - `mean_candidate_recall`: **0.287** (28.7%)
  - `mean_rule_precision`: **0.000** (0.0%)
  - `mean_rule_recall`: **0.017** (1.7%)
  - `mean_strict_assignment`: **0.006** (0.6% — fewer than 1 in 100 pairs recovered)
  - `mean_gain`: **+0.067** (System still achieves positive compression gain on spurious rules!).

**Definitive Scaling Conclusion:** The restricted real scope of 57 Voynich star labels is statistically incapable of supporting anonymous morphological rule induction. The positive compression gain observed in M2R-v2 is an artifact of finding spurious compact regularities in small character bags.

*Data source:* [`SCALING_RESULTS.tsv`](file:///home/brigadire/devops/workdir/go/voinich/research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/SCALING_RESULTS.tsv)

---

## 6. Null Controls and Resource Accounting (Section 8)

All eight synthetic null controls were executed to complete termination:
* `d71001` (independent_random): COMPLETE, Gain = 0.000, Accepted = NO
* `d71002` (length_matched): COMPLETE, Gain = -0.027, Accepted = NO
* `d71003` (unigram_preserving): COMPLETE, Gain = +0.022, Accepted = NO
* `d71004` (bigram_preserving): COMPLETE, Gain = +0.109, Accepted = NO
* `d71005` (shuffled_assignments): COMPLETE, Gain = +0.088, Accepted = NO
* `d71006` (repeated_fragment): COMPLETE, Gain = +0.048, Accepted = NO
* `d71007` (role_frequency): COMPLETE, Gain = +0.091, Accepted = NO
* `d71008` (size_matched): COMPLETE, Gain = -0.027, Accepted = NO

### Diagnosis of M2R-v2 OOM Failure:
M2R-v2 suffered from a Linux process accounting defect in `Budget.check()`. When parent orchestration processes accumulated memory across tests and runs, `os.fork()` caused Linux glibc/kernel task structures to inherit the parent's `mm->hiwater_rss`. When child processes executed, `getrusage(RUSAGE_SELF).ru_maxrss` reported the inherited parent peak ($>512$ MiB) rather than the child's actual resident memory ($\approx 10$–95 MiB), causing immediate `Limit('memory')` exceptions on iteration 0. In this diagnostic suite, measuring isolated process resident memory via `/proc/self/status` `VmRSS` enabled 100% completion across all null runs without exceeding memory limits.

### Determinism Verification:
* **Checkpoint-Resume Identity:** **PASS** (Bit-exact digest match).
* **Order Invariance:** **PASS** (Bit-exact digest match under reversed inputs, permuted inputs, and neutral ID renaming).

*Data sources:* [`NULL_CONTROL_RESULTS.tsv`](file:///home/brigadire/devops/workdir/go/voinich/research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/NULL_CONTROL_RESULTS.tsv), [`RESOURCE_PROFILE.tsv`](file:///home/brigadire/devops/workdir/go/voinich/research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/RESOURCE_PROFILE.tsv)

---

## Mandatory Final Status Block

```text
M2R_IDENTIFIABILITY_DIAGNOSTIC=COMPLETE
TRUE_RULE_CANDIDATE_RECALL=0.4815
EQUIVALENCE_AWARE_CANDIDATE_RECALL=0.7037
ORACLE_ASSIGNMENT_RULE_PRECISION=1.0000
ORACLE_ASSIGNMENT_RULE_RECALL=0.2889
ORACLE_RULE_ASSIGNMENT_ACCURACY=1.0000
EQUIVALENCE_AWARE_ASSIGNMENT_ACCURACY=0.5372
BOUNDED_SEARCH_EXACT_OPTIMUM_RATE=0.3750
EXHAUSTIVE_TRUTH_IS_GLOBAL_OPTIMUM=MIXED
EQUIVALENT_OPTIMA_MEDIAN=2.0
NULL_CONTROLS_COMPLETE=YES
CHECKPOINT_IDENTITY=YES
ORDER_INVARIANCE=PASS
PRIMARY_FAILURE_MODE=MIXED
REAL_SCOPE_FEASIBILITY=NOT_SUPPORTED
M2R_V3_RECOMMENDATION=DO_NOT_PROCEED
REAL_DATA_SEARCH_AUTHORIZED=NO
```
