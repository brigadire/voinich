# Results report

The amendment keeps the old package unchanged and audits witness-search power.

`VALID_NULL_WITNESSES=0`, `REAL_CALIBRATION_WITNESSES=1`, and `REJECTED_NULL_CANDIDATES=1` are the required ledger counts. The one real calibration item is the historical S043 assignment; corrected global replay rejects it, so it is a fixture count rather than a valid witness. The one null candidate from the old runner is also rejected.

The corrected incremental search passed synthetic positive calibration at all tested budgets: 20/20 at 5, 30, 120, and 600 seconds. On real data, 20 independent 5-second searches rediscovered zero valid witnesses. The 30/120/600-second real sweeps were not run because the required real calibration gate failed.

This gives no estimate of the true `P(null >= 5)`. It shows only that the current 5-second search has not demonstrated adequate real-witness detection power. No new null and no E4 are authorized by this package.
