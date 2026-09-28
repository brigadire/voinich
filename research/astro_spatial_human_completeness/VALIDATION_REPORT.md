# Completeness preparation validation

Status PASS. Tests run: 69; failures=0, errors=0.
All tests use prepared/synthetic copied trees, not actual new reviewer annotations.
Full isolated regeneration: 49 files byte-identical, including native XML, deterministic
image/annotation ZIPs, final-reference TSVs, navigation/eligibility maps, protocols, manifests
and current-reference attachment dry run. Upstream + comparison input hashes rechecked after
generation. Only this preparation directory is frozen by SHA256SUMS.

PASS: final human geometry, unchanged IDs/provenance, blindness, optional UNCERTAIN layer,
hidden REJECT/NOT_REVIEWED, exact panel/image coordinates, explicit completion, deterministic
new IDs independent of CVAT shape IDs, no auto acceptance, four prior-overlap queue classes,
mutated/deleted/relabeled/duplicate references blocked, unset confidence forbidden.
CVAT's source=file serialization for unchanged imported references is accepted; additions
must still have an explicit HUMAN_ADDED class and source=manual.
Copied-tree roundtrip includes CVAT native ellipse/rotated-box/tag serialization rounded to
two decimals, tolerance 0.0051 on serialized fields. No canonical reference is replaced by
rounded export geometry. All panel markers start NOT_STARTED in actual annotation package.

PASS: f68r1/f68r2 eligibility, unconfirmed f68r3 excluded, non-applicable panels not iterated,
filtered pairs not observations, v2 retained without conversion, v3 version/spatial-field
validation, gated future production blocked without external object freeze, endpoints protected.
NOT_APPLICABLE_REGION becomes a reconciliation question, not a negative relation.
Optional blind v2-pair recheck preserves the frozen v2 result and creates independent
v3 decisions; grouped pair frames retain exact many-to-many endpoint semantics.
The reviewer-selected full-page 3G1 workflow validates three canonical pages, immutable
endpoint geometry and editable visual-association hyperedge groups without Cartesian expansion.

CVAT live instance/version was unavailable; no CVAT writes/review occurred. Roundtrip PASS
means native XML copied-tree compatibility, NOT a claimed live-server integration test.
Empty live import/export preflight remains necessary on reviewer instance (guides/schema).
Lock/hide are optional UI protections; strict reference comparison is mandatory fallback.

HUMAN_REVIEW_STARTED=NO
REAL_HUMAN_EXPORT_IMPORTED=NO
NEW_OBJECTS_AUTO_ACCEPTED=0
NEW_AI_RECALL_CALCULATED=NO
PRODUCTION_ATTACHMENT_PACKAGE_CREATED=NO
FROZEN_INPUTS_UNCHANGED=YES
RESULTS_REPRODUCIBLE=YES
