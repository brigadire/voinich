# Validation report

Status: PASS. No frozen input was modified. All critical input gates passed before
aggregation; source IDs, panels, source boxes, one-record-per-source physical mapping,
phase deduplication, final decisions, attachment endpoints and published counts agree.

- Automatic tests: 30 run, failures=0, errors=0.
- Full second-run byte comparison: PASS, 49 regenerated files,
  including bootstrap tables/JSON, reports, source tables and PNG/SVG figures.
- Seed/plan lock: PASS; strict cohorts exclude UNCERTAIN and NOT_REVIEWED.
- Calibration/production and spatial-v1/caption-v2 separation: PASS.
- A-absence negative evidence: NONE.
- Denominators, raw/JSON aggregates and figure source links: PASS.
- Upstream checksums reverified after analysis: 673 registered inputs.
- Output SHA256SUMS: generated after this report; verified by sha256sum -c --quiet.

astro_spatial_annotation: 21 ledger entries; astro_spatial_annotation_ai_b: 12 ledger entries; astro_spatial_annotation_ai_b2: 20 ledger entries; astro_spatial_human_adjudication: 600 ledger entries;

The first validation may skip the output-ledger test because freezing follows validation.
The explicit post-freeze checksum command and subsequent test run verify the frozen output.
Single-panel intervals are candidate bootstrap, not between-panel confidence; few panel
clusters and geometry-proxy limitations remain statistical restrictions, not validation failures.
