# Formal Pre-Production Audit Report: Frozen M2R Engine v1

## 1. Executive Summary & Audit Mandate
An independent pre-production audit of the frozen M2R morphology engine (`research/astro_hapax_star_label/m2r_real_engine_v1`) was conducted in accordance with `tasks_other/preproduction_audit.md`.
The objective was to determine whether a one-time production run on sealed real data (Voynich star labels and controls) could be authorized without modifying the engine, thresholds, or configurations.

**Audit Conclusion**: **FAIL**.
Production execution on real data is **STRICTLY PROHIBITED**.
Seven (7) critical `BLOCKER` issues and three (3) `MAJOR` architectural defects were identified.
Per protocol, the frozen engine must not be patched in-place. A new `M2R-v2` package must be designed with proper assignment inference, recurrent morpheme segmentation, and independent support tracking.

---

## 2. Gate-by-Gate Evaluation

### Gate P1: Integrity — FAIL
* The frozen source files are intact and unedited since freeze.
* Sealed real inputs in `restricted_hapax_enrichment_v1` were strictly isolated and never accessed.
* **Failure Cause**: `m2r_real_engine_v1` failed to provide an internal `SHA256SUMS` ledger at freeze time (CR-06).

### Gate P2: Inference Validity — FAIL
* The engine does not discover morphological units; `units(s)` is dead code and units are restricted to single graphemes (CR-08).
* The engine does not infer term-label assignments (CR-01). It requires inputs to be pre-paired 1:1 and character-aligned via `zip(t, l)`.
* Latent rule precision in `run_development.py` hardcodes the development target alphabet `'oker'` (CR-04).

### Gate P3: Adversarial Validity — FAIL
* **Support >= 3 Violation**: `search()` tests `k in (2, 3, 4)` and admits support=2 rules whenever they increase data fit (CR-02).
* In null testing, support=2 rules were admitted in 20.0% of null replicates.
* **Fictive Support Counting**: `Counter(pairs)` counts character occurrences across rows rather than unique concept instances. Duplicate tokens and internal character repetitions inflate support (CR-03).
* In hard-negative testing, `HN4` (train-only morphology) produces high train score and is blindly accepted by the engine because `engine.py` lacks a held-out validation gate.

### Gate P4: Operational Parity — FAIL
* **Size Normalization**: The score formula is unnormalized (Score = Fit - 2*Un - Comp). On small datasets (circular text: 29 tokens), complexity dominates, whereas on large datasets (intro prose: 68 tokens), spurious rules easily pass (CR-09).
* **Missing Orchestration**: No production command line, no lexicon loader, no checkpoint/resume, and no interpretation status assigner exist in the engine package.

### Gate P5: Reproducibility — PASS
* The synthetic validation run is 100% byte-reproducible from seeds 11001 and 22001.

---

## 3. Key Quantitative Findings
* **Synthetic Positive Replication**: Train coverage = 1.000, Held-out coverage = 1.000, Precision = 1.000, Recall = 1.000, Mean Support = 46.08.
* **Synthetic Null Benchmark**: 240 replicates across 8 null families.
  - Null Median Score: 102
  - Null P95 Score: 124
  - Null P99 Score: 130
  - Null Max Score: 137
  - False Positive Rate: 0.0000
  - Replicates with k=2 admitted: 48 / 240 (20.0%)
* **Hard Negatives**: 0 / 7 hard negatives accepted under predictive heldout criteria, but `HN4` is accepted by train engine score alone.
* **Code Review Findings**: Total 12 findings (7 BLOCKER, 3 MAJOR, 2 MINOR).

---

## 4. Final Mandatory Status Declaration
```text
M2R_PREPRODUCTION_AUDIT=FAIL
P1_INTEGRITY=FAIL
P2_INFERENCE_VALIDITY=FAIL
P3_ADVERSARIAL_VALIDITY=FAIL
P4_OPERATIONAL_PARITY=FAIL
P5_REPRODUCIBILITY=PASS
FROZEN_ENGINE_UNCHANGED=YES
SEALED_REAL_INPUT_ISOLATION=PASS
LATENT_RULE_LEAKAGE=DETECTED
SYNTHETIC_RESULT_REPRODUCED=YES
SYNTHETIC_NULL_P95=124
SYNTHETIC_NULL_MAX=137
FALSE_POSITIVE_ACCEPTED_PROFILE_RATE=0.0000
HARD_NEGATIVES_ACCEPTED=0/7
OUT_OF_FAMILY_POSITIVE_RECOVERY=0.8000
ASSIGNMENT_RECOVERY_VALID=NO
MINIMUM_SUPPORT_VALID=NO
ORDER_INVARIANT=NO
CHECKPOINT_RESUME_IDENTICAL=NO
PRODUCTION_PARITY=FAIL
NULL_FULL_SEARCH_PARITY=FAIL
LEXICON_PARITY=FAIL
SIZE_NORMALIZATION=FAIL
OPEN_BLOCKERS=7
REAL_DATA_SEARCH_PERFORMED=NO
REAL_DATA_SEARCH_AUTHORIZED=NO
RESULTS_REPRODUCIBLE=YES
```
