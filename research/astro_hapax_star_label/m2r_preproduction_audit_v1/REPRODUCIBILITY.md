# Reproducibility Guide: M2R Pre-Production Audit v1

## How to Run the Audit
To reproduce the complete audit suite and verify all metrics, execute:
```bash
python3 research/astro_hapax_star_label/m2r_preproduction_audit_v1/scripts/run_audit.py
```

To run the automated test suite:
```bash
python3 -m unittest discover -s research/astro_hapax_star_label/m2r_preproduction_audit_v1/tests -p 'test_*.py'
```

## Determinism
All random generators in the audit harness use fixed seeds:
- Replication: Seeds 11001, 22001
- Null Corpora: Seeds 30000..30799
- Hard Negatives: Seed 40001
- Out-of-Family: Seed 50001
- Assignment: Seed 60001
- Determinism: Seed 70001
- Size Normalization: Seed 8888
