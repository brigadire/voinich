# Handoff for independent review

Read FINAL_STATUS.json and the validation report before using any artifacts. A BLOCKED
package is a rejected candidate and is not eligible for production pre-audit authorization.
CANDIDATE_SEAL.json only proves the candidate was fixed before hidden disclosure.

Verify SHA256SUMS and unchanged v1 hashes, then run regression tests with bytecode disabled.
Review the implementation contract and CR-01–CR-09 traceability against executable tests.
Surface inputs and separately sealed latent truth are distinct directories; the worker
must never read truth. Independently run surface-only inference before opening truth.
Review exact matching, candidate pruning, support provenance, role-sensitive derivations,
MDL arithmetic, search bounds, checkpoint identity and null parity. Pairwise explanations
for all edges are stored in each selected model. Inspect negative acceptance and positive
recovery together; rejecting everything is not a successful inference engine.

Existing sealed-real metadata is copied without token contents. Its original v1 manifest
is incomplete (no per-file registry or schema version); no missing metadata is invented.
No real input data, lexicons or real-data result files are distributed in this package.
Real-data search remains unauthorized regardless of synthetic outcomes.
