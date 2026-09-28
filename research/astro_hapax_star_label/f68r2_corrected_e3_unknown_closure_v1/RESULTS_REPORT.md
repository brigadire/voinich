# Results report

The new package imported and independently verified 64 committed checkpoints from the read-only prior package. The import found 40 exact maxima and 24 `UNKNOWN` profiles. All 24 `UNKNOWN` profiles have committed finite best bounds of 0.0, hence `BOUND_LE_3`; they were closed by bound without rerunning optimization. No `.tmp` directory or uncommitted result was used.

The independent decision solver was validated on committed profiles S002, S004, S005, and S007. Each returned `INFEASIBLE` for `coverage >= 4` with replay PASS. A synthetic four-path non-hereditary support fixture returned `OPTIMAL/FEASIBLE`, coverage 4, and replay PASS. This confirms that the closure does not rely on the invalid claim that a valid larger solution must contain a valid smaller subset.

The resulting exact family-level maximum is **3**. This is a technical exploratory result. `NULL_RUN_REQUIRED=NO` because the frozen trigger is 4; no scientific correlation or astronomical identification is claimed.
