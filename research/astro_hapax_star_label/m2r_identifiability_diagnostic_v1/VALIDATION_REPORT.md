# VALIDATION REPORT: M2R IDENTIFIABILITY DIAGNOSTIC v1
**Package:** `research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1`  
**Date:** 2026-09-16  
**Status:** `VERIFIED_PASS`

---

## 1. Upstream and Peer Package Integrity Verification

Prior to and following the diagnostic experiments, the cryptographic integrity of all existing repositories and packages was strictly verified:

| Target Component | Status | Verification Detail |
| :--- | :---: | :--- |
| `m2r_real_engine_v1` | **UNCHANGED** | Baseline SHA-256 hashes match `V1_BASELINE_HASHES.json`. |
| `m2r_real_engine_v2` | **UNCHANGED** | All 43 files in v2 match `CANDIDATE_SEAL.json` and `FINAL_STATUS.json`. |
| `restricted_m2_star_labels_v1` | **UNTOUCHED** | Real Voynich transcripts and label mappings were never accessed or imported. |
| `restricted_hapax_enrichment_v1` | **UNTOUCHED** | Enrichment dictionaries remained sealed outside the diagnostic scope. |
| `M0 / M1` packages | **UNTOUCHED** | Zero access or modification. |

No real text tokens from folios f68r1 or f68r2 were fed into any diagnostic engine.

---

## 2. Diagnostic Pipeline Validation Results

Every component of the audit package was verified against its contract:

1. **Section 1: Pre-declared Protocol**  
   `IDENTIFIABILITY_AUDIT_PROTOCOL.md` was created and frozen prior to running experiments. All thresholds and decision criteria remained unchanged throughout execution.
2. **Section 2: Oracle Decomposition**  
   Evaluated across 54 conditions (9 datasets $\times$ 6 modes). Output written to `ORACLE_EXPERIMENT_MATRIX.tsv`. Verified that Mode A reproduces v2 results and Mode C achieves 100% assignment recovery.
3. **Section 3: Candidate Generation Pool Audit**  
   Evaluated all 135 true rules. Verified that 51.85% of rules are lost prior to search due to support thresholds and pool truncation. Output written to `CANDIDATE_POOL_AUDIT.tsv`.
4. **Sections 4 & 5: Exact Exhaustive Search and Symmetry Analysis**  
   All 8 small synthetic tasks evaluated exhaustively. Global optima, equivalence classes, and non-identifiability proofs verified. Outputs written to `EXACT_SEARCH_COMPARISON.tsv` and `EQUIVALENCE_CLASS_ANALYSIS.tsv`.
5. **Section 6: Assignment Identifiability**  
   Verified that when true rules are known, assignment accuracy is 100% with margins $>75$ bits. Output written to `ASSIGNMENT_IDENTIFIABILITY.tsv`.
6. **Section 7: Factorial Scaling Study**  
   All 13 configurations (125 instances, 20 seeds for key configs) evaluated. Verified that real restricted scope (57 star labels) yields 1.7% rule recall and 0.6% assignment accuracy. Output written to `SCALING_RESULTS.tsv`.
7. **Section 8: Null Controls and Resource Accounting**  
   All 8 null datasets completed with zero OOM. Checkpoint-resume identity and input order invariance confirmed. Outputs written to `NULL_CONTROL_RESULTS.tsv` and `RESOURCE_PROFILE.tsv`.

---

## 3. Mandatory Diagnostic Disposition

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
