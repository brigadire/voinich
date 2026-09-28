# Results report

The audit found that the old checkpoint format did not preserve raw solver responses. It preserved status, incumbent, and a numeric bound, but not objective direction metadata, response protobuf, solution info, branches, conflicts, timing, or stop reason. The old 24 values `UNKNOWN + bound=0.0` therefore could not be accepted as certificates.

A known-positive OR-Tools sentinel model demonstrated that `UNKNOWN`, objective 0.0, and best bound 0.0 can occur under zero or tiny search budgets although the true optimum is 1. The values were therefore classified `BOUND_METADATA_MISSING`.

All 24 affected frozen graphs were independently tested with the exact decision query `coverage >= 4`. Results: 24 `INFEASIBLE`, 0 `FEASIBLE`, 0 `UNKNOWN`, 0 implementation failures. Every query included a graph hash, model fingerprint, solver metadata, and independent scorer replay.

The global E3 maximum is now certified as 3. The result is technical and exploratory only; `NULL_RUN_REQUIRED=NO` because the frozen trigger is 4.
