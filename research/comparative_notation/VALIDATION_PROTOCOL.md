# Validation protocol

No full corpus may run until its 20–100-unit fixture passes. Each adapter keeps
the source fragment, manually reviewed `expected.usc.jsonl`, and independently
generated `generated.usc.jsonl`. Tests compare exact bytes and verify symbol
and token counts, physical-line/page boundaries, hierarchy, stable IDs,
ordering, and missing-level NULL semantics.

Generic metamorphic requirements are:

- M1 bijective symbol rename: every label-invariant G/T/S/L/D value unchanged;
- M2 document duplication: relative frequency/density metrics unchanged;
- M3 token-order shuffle: G/T unchanged and at least one S metric changes;
- M4 within-token shuffle: length distribution unchanged and adjacency changes;
- M5 line-label/order shuffle: within-line metrics unchanged and document
  progression changes when defined;
- M6 page-label/order shuffle: lower grammar unchanged and document progression
  changes when defined.

Numerical equality uses absolute tolerance `1e-12`. A metamorphic test that
cannot exercise its declared effect because the fixture is degenerate fails
the fixture design rather than passing vacuously.

Missing data statuses are `COMPARABLE`, `PARTIALLY_COMPARABLE`, and
`NOT_COMPARABLE`. Structural-level metadata is never imputed. The run manifest
must record source hash, adapter, representation, analyzer, metric registry,
parameters, seed, command, UTC timestamp, and output hashes.
