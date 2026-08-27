# Structural adequacy v2

The discriminating G1 battery remains the seven Task85 metrics in EDIT and LEXICAL_PARADIGM. Their computations and family thresholds are frozen in `G1V2_STRUCTURAL_GATE_REGISTRY.tsv`. Layout/hierarchy measures whose value is copied by the borrowed skeleton remain recorded diagnostics with zero gate weight; a construction-guaranteed match is not independent evidence.

For each generation scale and replicate, every metric record contains generated value, target/control value, signed and absolute distance, baseline, threshold ID/value, applicability, availability, finite flag, and PASS/FAIL/NOT_ASSESSABLE. A generation not run is `NOT_REACHED` with upstream job/gate/reason. It is never serialized as FAIL.

Metric PASS means a finite applicable distance is at most its frozen development-derived bound; equality passes. Metric FAIL requires a finite applicable distance above the bound. Everything else is NOT_ASSESSABLE. EDIT passes with at least three assessable members and three passes; LEXICAL_PARADIGM passes with at least two assessable members and two passes. A valid family shortfall is FAIL; insufficient assessable members is NOT_ASSESSABLE.

Each required scale uses four generation replicates. First take the median metric value, while retaining every replicate; a scale family passes under the family rule, fails on valid shortfall, otherwise is NOT_ASSESSABLE. Structural PASS requires both families PASS at all three scales. Any valid family FAIL yields STRUCTURAL_FAIL only when the remaining required evidence has no integrity/missingness defect; otherwise STRUCTURAL_NOT_ASSESSABLE. Replicate instability above the frozen development q95 dispersion bound makes the scale NOT_ASSESSABLE.

During development/control validation, structural diagnostics run even after predictive failure so the evidence path can be audited. In later target execution, a frozen reachability policy may skip generation after a valid predictive FAIL; each planned downstream cell must still exist as NOT_REACHED with dependency and reason.
