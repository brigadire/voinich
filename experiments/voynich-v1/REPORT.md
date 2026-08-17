# Voynich Baseline - Experiment Report

ExperimentID: `b958a06f6b6404974ea32e54c3ba2b295fc5f57ba87e90caab8fd5dd25313b75`

Git commit: `61d6c206e6dcfab6b721fc1172125818867e9eed`

Created: 2026-08-16T19:06:26Z

Platform: linux/amd64, Go go1.26.4-X:nodwarf5, host `adelie`, 12 CPUs

IVTFF source: `data/ZL3b-n.txt` (sha256 `bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc`)

Frozen corpus: `data_work/ZL3b-x7.txt` (sha256 `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`)

Executor: `remote` - conditional-regime-analyze used -executor remote with 10 authenticated remote worker(s) over Task34 mTLS

Workers:

- remote: `cognition`
- remote: `soa-tech-gitlab02`
- remote: `soa-tech-gitlab03`
- remote: `soa-tech-gitrunner-hw`
- remote: `soa-tech-gitrunner-hw2`
- remote: `soa-tech-sentry2`
- remote: `tech-services-gitlab`
- remote: `tech-services-logger`
- remote: `tech-services-monitoring`
- remote: `tech-services-nexus`

## Stage results

| # | Stage | Status | Wall time | User CPU | Sys CPU | Max RSS |
|---|---|---|---|---|---|---|
| 1 | dict-gen | completed | 0.7s | 1.3s | 0.3s | 653252 KB |
| 2 | dict-analyze | completed | 1.5s | 2.3s | 0.4s | 832288 KB |
| 3 | structural-analyze | completed | 1.8s | 2.1s | 0.1s | 183096 KB |
| 4 | sequence-analyze | completed | 0.7s | 1.1s | 0.1s | 234936 KB |
| 5 | begin-end-analyze | completed | 58.0s | 59.6s | 0.7s | 843016 KB |
| 6 | structural-normalize | completed | 1.2s | 1.9s | 0.4s | 957872 KB |
| 7 | normalization-compare | completed | 140.5s | 279.7s | 7.8s | 372880 KB |
| 8 | structural-validate | completed | 31.4s | 57.1s | 2.0s | 1311048 KB |
| 9 | structural-profile-stability | completed | 31.2s | 35.2s | 1.3s | 2541624 KB |
| 10 | structural-reliability | completed | 23.0s | 25.3s | 0.3s | 348820 KB |
| 11 | soft-structural-space | completed | 2.8s | 3.3s | 0.2s | 236500 KB |
| 12 | structural-graphemic | completed | 1.0s | 1.1s | 0.1s | 158504 KB |
| 13 | structural-pair-decompose | completed | 3.7s | 7.5s | 0.6s | 644324 KB |
| 14 | distance-context-analyze | completed | 30.6s | 42.4s | 0.2s | 387940 KB |
| 15 | local-regime-analyze | completed | 45.3s | 50.6s | 1.1s | 1880352 KB |
| 16 | property-trajectory-analyze | completed | 12.6s | 13.5s | 0.5s | 679848 KB |
| 17 | structural-projection-analyze | completed | 12811.0s | 14550.8s | 442.1s | 14623792 KB |
| 18 | global-regime-analyze | completed | 116.4s | 117.1s | 0.3s | 170452 KB |
| 19 | metadata-validate | completed | 491.1s | 497.4s | 2.0s | 1880724 KB |
| 20 | cluster-metadata-global | completed | 117.0s | 118.7s | 1.1s | 1380020 KB |
| 21 | conditional-regime-analyze | completed | 2243.8s | 2267.2s | 89.2s | 1813380 KB |
| 22 | residual-diagnostic-analyze | completed | 63.0s | 87.4s | 0.7s | 99616 KB |
| 23 | token-relation-validate | completed | 54.5s | 63.2s | 0.8s | 640088 KB |
| 24 | replicated-local-structure-audit | completed | 135.1s | 159.7s | 1.7s | 1961884 KB |
| 25 | higher-order-sequence-validate | completed | 1.3s | 1.3s | 0.0s | 46956 KB |
| 26 | positional-continuation-validate | completed | 0.5s | 0.6s | 0.1s | 47580 KB |
| 27 | transition-network-validate | completed | 95.0s | 104.7s | 1.5s | 291776 KB |

**Total wall time (sum of stages): 4h35m15s**

Full manifest: `manifest.json`. Per-file checksums: `checksums.sha256`. Per-stage logs: `logs/`.
