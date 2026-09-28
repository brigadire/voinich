# Reproducibility Guide: Restricted Dictionary Brute-Force v3 Design

## 1. Execution Environment

- **Operating System**: Linux (x86_64, GNU/Linux)
- **Runtime**: Python 3.14.6+ (`/usr/bin/python3`)
- **Required Libraries**: Standard library only (`csv`, `json`, `math`, `time`, `random`, `collections`, `itertools`, `pathlib`, `unittest`, `tracemalloc`, `unicodedata`, `re`, `hashlib`)
- **External Dependencies**: None (self-contained, zero proprietary solver or external library dependencies)

## 2. Deterministic Seeds Specification

All synthetic generators, randomizations, and stochastic search passes are governed by explicitly fixed PRNG seeds:

| Component / Task | Seed / Seed Range | Description |
|---|---|---|
| **Small-Instance Benchmark** | `1000 + inst_idx * 17` (`1017`–`1425`) | 25 distinct small synthetic instances |
| **Suite 1 (0% Noise Key Benchmark)** | `101`–`120` | 20 independent seeds |
| **Suite 2 (10% Noise Key Benchmark)** | `101`–`120` | 20 independent seeds |
| **Suite 3 (Table Size Sweep)** | `101`–`105` | 5 seeds per size/noise cell |
| **Suite 4 (Merge & Abbreviation)** | `101`–`105` | 5 seeds per mode cell |
| **Suite 5 (25% Noise Stress Test)** | `101`–`120` | 20 independent seeds |
| **Suite 6 (Multi-Algorithm Comparison)**| `101`–`110` | 10 seeds comparing CP-SAT & Baseline |
| **Order Invariance Test** | `999` (solver seed), `12345` (shuffle seed) | Permuted label ordering verification |
| **Checkpoint Resume Test** | `999` | Serialization identity verification |

## 3. Replication Commands Workflow

To regenerate the entire v3 design package from scratch:

```bash
# Set repository root
cd /home/brigadire/devops/workdir/go/voinich

# 1. Extract purely structural features from target scope (no dictionary matching)
python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v3_design/scripts/build_structural_manifest.py

# 2. Compile comprehensive historical star lexicon and source registry (Gate L)
python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v3_design/scripts/build_historical_lexicon.py

# 3. Execute reachability and capacity analysis across 120 model classes
python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v3_design/scripts/analyze_reachability.py

# 4. Run exact small-instance exhaustive benchmarks across all algorithms
python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v3_design/scripts/run_small_instance_benchmark.py

# 5. Run full sealed synthetic recovery benchmark matrix
python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v3_design/scripts/run_synthetic_recovery_benchmark.py

# 6. Execute automated unit and regression test suite
python3 -m unittest research/astro_hapax_star_label/restricted_dictionary_bruteforce_v3_design/tests/test_v3_engine.py

# 7. Generate cryptographic checksums
cd research/astro_hapax_star_label/restricted_dictionary_bruteforce_v3_design
sha256sum $(find . -maxdepth 1 -type f ! -name "SHA256SUMS" | sort) > SHA256SUMS
```

## 4. Verification of Invariance Properties

- **Order Invariance**: Verified in `test_v3_engine.py` and Suite 7 of `run_synthetic_recovery_benchmark.py`. Shuffling label rows produces bit-identical matching scores, coverage, and selected mapping tables.
- **Checkpoint Identity**: Verified by writing search state to disk and resuming. State digest matches across serialization boundaries.
