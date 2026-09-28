# DECISION ON M2R-v3: FORMAL SCIENTIFIC DISPOSITION
**Document ID:** `M2R-V3-DECISION-DISPOSITION`  
**Package:** `research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1`  
**Date:** 2026-09-16  
**Final Status:** `DO_NOT_PROCEED`  
**Real Scope Feasibility:** `NOT_SUPPORTED`  
**Real Data Search Authorized:** `NO`

---

## 1. Formal Recommendation

Pursuant to the pre-declared evaluation protocol (`IDENTIFIABILITY_AUDIT_PROTOCOL.md`) and the empirical findings documented in `IDENTIFIABILITY_AUDIT_REPORT.md`, the diagnostic audit renders the following binding disposition:

$$\mathbf{M2R\_V3\_RECOMMENDATION = DO\_NOT\_PROCEED}$$
$$\mathbf{REAL\_SCOPE\_FEASIBILITY = NOT\_SUPPORTED}$$
$$\mathbf{REAL\_DATA\_SEARCH\_AUTHORIZED = NO}$$

Development of an M2R-v3 engine targeting the restricted STAR LABEL scope of Voynich folios f68r1/f68r2 is **scientifically unjustified and prohibited**.

---

## 2. Answers to Mandatory Completion Criteria

The audit provides definitive, mathematically proved answers to the six central scientific questions:

### 1. Does the model contain sufficient information to recover true latent rules?
**NO, under the anonymous setting.**  
While the data contains sufficient mutual information to recover assignments *when true rules are already known* (Mode C achieves 100% assignment accuracy and 78.2% compression gain), the reverse is not true. In an unsupervised, anonymous setting without paired term-label supervision, the surface data admits multiple mutually inconsistent rule sets that yield identical surface bags and identical compression scores.

### 2. Can specific term–label assignments be recovered, rather than merely plausible compression?
**NO.**  
In unguided search (Mode A), assignment accuracy is 2.21% on synthetic benchmarks and 0.60% on real-scope simulations. Shuffled-assignment null tests (`T8_shuffled_assignment_null` and `d71005`) prove that randomly permuted pairings produce the **exact same compression gain** as the ground truth. M2R discovers compact regularities in character n-gram frequencies, not grounded semantic or lexical correspondences.

### 3. At what stage is ground truth lost: candidate generation, assignment, search, or scoring?
**Truth is lost through compounding failures at ALL four stages (`PRIMARY_FAILURE_MODE=MIXED`):**
* *Candidate Generation:* 51.85% of true rules are discarded prior to search due to support thresholding and pool cap truncation (`TRUE_RULE_CANDIDATE_RECALL=0.4815`).
* *Search:* The bounded beam search fails to reach the global optimum in 62.5% of small instances (`BOUNDED_SEARCH_EXACT_OPTIMUM_RATE=0.3750`) and is structurally limited to at most 8 rules when true models require 15.
* *Scoring:* When hapax legomena are present, the MDL objective penalizes the 15 true rules ($10 + L(s) + L(t)$ bits per rule) and favors compact, false 2–3 rule models.
* *Assignment:* Symmetries in Cartesian affix combinations create degenerate equivalence classes where Hungarian matching makes arbitrary selections based on tie-breaking.

### 4. Is the true solution optimal according to the current objective?
**MIXED.**  
In dense Cartesian synthetic tasks, the ground truth is an optimum, but it shares that optimum with multiple false permutations (e.g., 4 to 36 equivalent optima). In hapax-rich tasks, the ground truth is **NOT** the global optimum: a compact, false model achieves higher normalized compression gain (`EXHAUSTIVE_TRUTH_IS_GLOBAL_OPTIMUM=MIXED`).

### 5. Is a data volume comparable to the 57 real STAR LABELs sufficient?
**ABSOLUTELY NOT.**  
Factorial scaling experiments over 20 independent random seeds simulating 57 star labels across two pages with realistic hapax frequencies demonstrated:
* Mean candidate recall: 28.70%
* Mean latent rule recall: 1.70%
* Mean strict assignment accuracy: 0.60% (less than 1 pair in 100 recovered).
The restricted scope of 57 hapax-heavy labels is statistically starved of combinatorial repetitions.

### 6. Is there a justified path to M2R-v3?
**NO.**  
No algorithmic refinement of candidate generation, beam search, or description length scoring can overcome the fundamental under-determination of anonymous dictionary matching at $N=57$. Any attempt to run an M2R variant on real data would output compact, spurious character alignments that appear mathematically valid (positive compression gain) but represent zero historical or linguistic truth.

---

## 3. Prohibitions and Final Mandate

1. No M2R-v3 package shall be created for unsupervised star label decipherment.
2. Real data from f68r1/f68r2, Voynich transcriptions, and enrichment dictionaries shall remain sealed and unauthorized for engine execution.
3. The M2R-v2 failure is permanently closed with the verdict `BLOCKED`.
