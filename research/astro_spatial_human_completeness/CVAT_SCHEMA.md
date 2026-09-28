# CVAT schemas and reference protection

CVAT_OBJECT_COMPLETENESS_SCHEMA.json has 6 classes: STAR_REFERENCE (rectangle),
LABEL_REFERENCE (any: box/ellipse only), optional UNCERTAIN_REFERENCE (any: box/ellipse
according to historical geometry), STAR_HUMAN_ADDED (unrotated rectangle),
LABEL_HUMAN_ADDED (rotated box/ellipse), PANEL_COMPLETENESS (tag).
Importer rejects other actual geometries, despite `any` allowing UI choices.

Every attribute `values` is a NON-EMPTY array, including text [`UNSET`] or version value.
Default values are in declared values; real free text reference IDs are not select enums.
This avoids previous “attribute values must be a non-empty array” / wrong-label schema
errors. Names in JSON, XML and validator are identical. No numeric CVAT attribute IDs
are persisted: they belong to the destination task, not the canonical schema.

Reference attributes only reference_id + object_protocol_version, immutable spec fields.
No human decision/confidence history in references. New status/confidence are separate.
Attachment schema has STAR_REFERENCE, LABEL_REFERENCE and PAIR_ATTACHMENT tag with
neutral pair ID, version 3, decision and confidence; no legacy spatial checkboxes.

Format uses native [CVAT for images 1.1](https://docs.cvat.ai/docs/manual/advanced/formats/format-cvat/),
including attributes/tags/boxes/ellipses. Runtime instance version was not available here;
local Docker contains no CVAT server and no configured CVAT connector/URL was provided.
Validation therefore uses copied-tree serialization roundtrip, not a claimed live server test.

The documented [Objects sidebar](https://docs.cvat.ai/docs/annotation/annotation-editor/objects-sidebar/)
offers label-level lock/hide and per-shape lock. Apply those UI controls during review if
present. Native XML does not promise persistent per-reference access control; no fabricated
locked XML field is added. Exact exported reference comparison is the authoritative fallback.
Runtime preflight: identify version (scripts/check_cvat_version.py), import/export without
review, run object validator --allow-incomplete. Any incompatibility must be resolved in a
new prepared revision before human review; never modify frozen references to make import pass.
