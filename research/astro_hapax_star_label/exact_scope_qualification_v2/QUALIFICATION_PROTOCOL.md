# Exact-scope qualification v2

This protocol is frozen before the new qualification seeds are generated. The earlier 100 solver cases are retained under `diagnostic_precheck/` and cannot contribute to qualification gates.

The exact scope is `INJECTIVE` and `MERGE_1`, with `DROP_UNMAPPED`, no abbreviation, single-character source units, and capacity policy `PER_PAGE_CAPACITY_1` or `GLOBAL_CAPACITY_1`. The CP-SAT budget is 60 seconds with one worker. No retries or post-result parameter changes are permitted.

The first three attempted sealed sets were invalidated and retained under `invalid_attempt_1/`, `invalid_attempt_2/`, and `invalid_attempt_3/`; none can be reused. The replacement qualification seeds are fresh and disjoint from all prior attempts: zero-noise root `20261401`, noisy root `20261402`, hard-negative root `20261403`. The fixed pilot scale is five identities, eight labels, six train and two held-out labels per case. Positive cases contain zero-noise and 20% noise strata; hard negatives are generated independently and cannot be used to tune thresholds.

Every qualification row contains solver, oracle, scorer, configuration, dependency, and execution metadata hashes. Metrics are unconditional over every generated case. Missing, timeout, infeasible, and uncertified outputs remain in denominators and are never converted to zero silently.

The only permitted result states are `EXACT_SCOPE_QUALIFIED`, `EXACT_SCOPE_NOT_QUALIFIED`, `QUALIFICATION_INVALID`, and `QUALIFICATION_INCONCLUSIVE_RESOURCE_LIMIT`. Real STAR LABEL data is outside this run and remains unauthorized.
