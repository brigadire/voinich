# REPRODUCIBILITY GUIDE: M2R IDENTIFIABILITY DIAGNOSTIC v1
**Package:** `research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1`  
**Date:** 2026-09-16  
**Status:** `DETERMINISTIC_REPRODUCIBLE`

---

## 1. System Requirements and Environment

The diagnostic suite is designed to be 100% reproducible with zero third-party packages:
* **Operating System:** Linux (x86_64 or aarch64). Required for process-level memory instrumentation via `/proc/self/status` `VmRSS`.
* **Python Runtime:** Python 3.10+ (tested on Python 3.12/3.14).
* **Dependencies:** Python Standard Library only (`json`, `csv`, `hashlib`, `itertools`, `math`, `pathlib`, `random`, `statistics`, `time`, `unicodedata`, `collections`, `dataclasses`, `functools`, `unittest`). No external dependencies.

---

## 2. Directory Structure and Modules

```text
research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/
├── IDENTIFIABILITY_AUDIT_PROTOCOL.md     # Pre-declared locked protocol
├── IDENTIFIABILITY_AUDIT_REPORT.md       # Comprehensive scientific audit report
├── M2R_V3_DECISION.md                   # Formal decision disposition on M2R-v3
├── VALIDATION_REPORT.md                 # Package integrity validation report
├── REPRODUCIBILITY.md                   # Step-by-step reproduction instructions
├── SHA256SUMS                           # Cryptographic ledger of all package files
├── audit_engine.py                      # Core MDL engine with memory-safe accounting
├── candidate_audit.py                   # Section 3: Candidate pool audit
├── oracle_audit.py                      # Section 2: Oracle decomposition modes A-F
├── exact_search_audit.py                # Sections 4 & 5: Exhaustive search & symmetries
├── assignment_identifiability_audit.py  # Section 6: Assignment identifiability audit
├── scaling_audit.py                     # Section 7: Factorial scaling study
├── null_resource_audit.py               # Section 8: Null controls & resource profile
├── run_audit.py                         # Unified master execution entry point
├── tests/                               # Automated regression and unit test suite
│   ├── test_engine_parity.py            # Verifies exact parity with M2R-v2 scoring
│   ├── test_candidate_generation.py     # Verifies candidate pool filtering logic
│   └── test_exact_search_exhaustion.py  # Verifies exhaustive search optimizer
└── *.tsv                                # 8 generated scientific data artifacts
```

---

## 3. How to Reproduce All Results

From the repository root directory (`voinich/`):

### 3.1. Execute the Automated Test Suite
To verify the engine parity, candidate mechanics, and exhaustive optimizer:
```bash
python3 -m unittest discover -s research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/tests -v
```

### 3.2. Run the Full Diagnostic Pipeline
To regenerate all 8 TSV datasets, verify integrity, update `SHA256SUMS`, and print the mandatory final status block:
```bash
python3 research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/run_audit.py
```
*Expected runtime:* $\approx 4$–6 minutes total on standard modern CPU hardware. Peak memory remains $< 150$ MiB throughout.

### 3.3. Execute Individual Diagnostic Modules
Each diagnostic module can also be invoked independently:

* **Candidate Pool Audit:**
  ```bash
  python3 -c "import sys; from pathlib import Path; sys.path.insert(0, 'research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1'); import candidate_audit; candidate_audit.run_candidate_audit(Path('research/astro_hapax_star_label/m2r_real_engine_v2/DATASET_REGISTRY.json'), Path('research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/CANDIDATE_POOL_AUDIT.tsv'))"
  ```
* **Oracle Decomposition (Modes A–F):**
  ```bash
  python3 -c "import sys; from pathlib import Path; sys.path.insert(0, 'research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1'); import oracle_audit; oracle_audit.run_oracle_decomposition(Path('research/astro_hapax_star_label/m2r_real_engine_v2/DATASET_REGISTRY.json'), Path('research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/ORACLE_EXPERIMENT_MATRIX.tsv'))"
  ```
* **Exact Exhaustive Search:**
  ```bash
  python3 -c "import sys; from pathlib import Path; sys.path.insert(0, 'research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1'); import exact_search_audit; exact_search_audit.run_exact_search_audit(Path('research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/EXACT_SEARCH_COMPARISON.tsv'), Path('research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/EQUIVALENCE_CLASS_ANALYSIS.tsv'))"
  ```
* **Assignment Identifiability:**
  ```bash
  python3 -c "import sys; from pathlib import Path; sys.path.insert(0, 'research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1'); import assignment_identifiability_audit; assignment_identifiability_audit.run_assignment_identifiability_audit(Path('research/astro_hapax_star_label/m2r_real_engine_v2/DATASET_REGISTRY.json'), Path('research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/ASSIGNMENT_IDENTIFIABILITY.tsv'))"
  ```
* **Scaling Study:**
  ```bash
  python3 -c "import sys; from pathlib import Path; sys.path.insert(0, 'research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1'); import scaling_audit; scaling_audit.run_scaling_study(Path('research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/SCALING_RESULTS.tsv'))"
  ```
* **Null Controls & Resource Accounting:**
  ```bash
  python3 -c "import sys; from pathlib import Path; sys.path.insert(0, 'research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1'); import null_resource_audit; null_resource_audit.run_null_and_resource_audit(Path('research/astro_hapax_star_label/m2r_real_engine_v2/DATASET_REGISTRY.json'), Path('research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/NULL_CONTROL_RESULTS.tsv'), Path('research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1/RESOURCE_PROFILE.tsv'))"
  ```

---

## 4. Cryptographic Verification

Verify all file digests against the ledger:
```bash
cd research/astro_hapax_star_label/m2r_identifiability_diagnostic_v1 && sha256sum -c SHA256SUMS
```
All files must return `OK`.
