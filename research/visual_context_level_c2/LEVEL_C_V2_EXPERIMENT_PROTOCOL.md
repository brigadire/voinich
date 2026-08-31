# Level-C v2 protocol (frozen proposal; production not run)

Inputs are the frozen 1.0.0 visual table, schema/protocol/manifest, the existing page-level fingerprint table, broad visual taxonomy, and frozen metadata. No descriptor or textual metric may be added, removed, or redefined.

Primary test: a missing-aware, symmetric normalized visual-vector/text-fingerprint distance statistic. For each page pair within a broad section, compute Euclidean distance over the intersection of observed visual dimensions, divided by the square root of the number of observed dimensions; pairs with fewer than five shared observed descriptors are ineligible. The page-level text distance is the fixed z-scaled Euclidean distance over the ten frozen fingerprint metrics. The primary statistic is the within-section Mantel-style correlation between these two distance matrices, with physical-leaf blocks preserved. This uses partial observations without imputing missing descriptors.

Secondary tests are descriptor-wise partial-observation associations, fixed schema-family summaries, section-specific effects, fully-annotated-only sensitivity, Herbal-only/non-Herbal/all-section diagnostics, and two negative controls. Benjamini–Hochberg correction is applied within the descriptor-wise family.

Confounder strategy: restricted exchangeability blocks are section × compatible quire where at least two physical-leaf blocks exist; otherwise section × physical-leaf blocks are retained and the result is labelled residual-confounded. The saturated Currier/hand/quire/position/size model is not forced when rank or residual degrees of freedom fail.

Production null: exact enumeration for strata with at most eight exchangeable blocks; otherwise 10,000 seeded permutations (`seed=20260831`). The v2 protocol is frozen but production execution is explicitly prohibited in this task.

Decision classes are `DETECTED_ROBUST`, `DETECTED_WITH_LIMITATIONS`, `NOT_DETECTED`, `INCONCLUSIVE`, and `NOT_IDENTIFIABLE`. A positive result with unresolved quire/hand/Currier separation can never be `DETECTED_ROBUST`.
