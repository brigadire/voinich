# Validation report

- Read-only import of the prior package: PASS, 64/64 checkpoints valid.
- Raw-bound provenance audit: PASS; missing response metadata explicitly recorded.
- Sentinel/default behavior: PASS; known optimum 1 produced UNKNOWN/0.0 under zero and tiny budgets.
- Direct decision queries: PASS, 24/24 `INFEASIBLE`.
- Independent replay: PASS for all direct-query results; no accepted real witness ≥4.
- Resume behavior: PASS; rerunning the audit skipped all 24 completed profiles and reproduced the final status.
- Atomic result updates: PASS; per-profile TSV updates use temporary files followed by `os.replace`.
- Null controls: not run and not authorized.
