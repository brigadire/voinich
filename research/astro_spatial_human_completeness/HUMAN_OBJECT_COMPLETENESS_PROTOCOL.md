# Object completeness protocol C1.0

Status: prepared, human review NOT STARTED. This amendment adds human proposals and
panel completion to geometry rules of immutable STAR/LABEL v1.2; it does not revise
historical decisions. Scope: all eight canonical images in queue manifest.

## Reference and permissible actions

STAR_REFERENCE / LABEL_REFERENCE contain final human ACCEPT/MODIFY geometry only.
Attributes are neutral reference_id and object_protocol_version. IDs are stable blind
aliases; REFERENCE_OBJECTS.tsv maps them to unchanged historical canonical IDs. No
model/source identity, support, priority, confidence/history or semantic notes enters UI.
Do not move, resize, relabel, delete, copy or edit reference attributes.

UNCERTAIN_REFERENCE is an optional, separately hideable label layer (8 known unresolved
objects). Hide it normally, enable when checking possible duplication; do not re-review
or change it. It is not a confirmed object. REJECT and NOT_REVIEWED are hidden indices,
not visible ordinary reference shapes. Their absence is never a negative/new decision.

Create only STAR_HUMAN_ADDED or LABEL_HUMAN_ADDED. STAR uses unrotated box. Local,
inclined/radial physical text runs use CVAT rotated box; a continuous complete text ring
uses ellipse. Several words in one continuous line form a single LABEL. Ring ellipse
denotes a path, not the empty interior. No new geometry types, transcription or semantics.
Weak star-shaped marks and partially clipped edge stars are allowed; overlapping stars
remain separate objects. Star/decorative discrimination uses visible star form only.

New addition_status: PROPOSED_NEW; UNCERTAIN_NEW for doubtful new mark; POSSIBLE_DUPLICATE_NEW
for a duplicate/reconciliation question, never automatic acceptance. human_confidence
LOW/MEDIUM/HIGH must be explicitly set (default UNSET, no inheritance of old forced HIGH).
Panel completion counts proposals, not accepted objects. All additions remain NOT_ACCEPTED.

## Systematic review

Use the fixed 3×4 row-major navigation grid T01…T12 (NAVIGATION_TILES.tsv and maps).
Review STAR then LABEL; toggle corresponding reference labels; inspect boundaries,
circular regions and faint/damaged regions separately. Coordinates always remain those
of the original panel; maps/tiles are navigation only, never new annotation frames.
PANEL_COMPLETENESS is one image tag per original frame; its five checkboxes record the
completed STAR/LABEL/edges/circles/weak-contrast passes. It requires an explicit state:
COMPLETE_NO_NEW_OBJECTS, COMPLETE_WITH_NEW_OBJECTS, INCOMPLETE_TECHNICAL or
INCOMPLETE_REVIEWER_UNCERTAINTY. Initial NOT_STARTED is not completion. An unresolved new
proposal can remain UNCERTAIN_NEW after a completed scan; whole-panel inability to inspect
requires INCOMPLETE. No marker/all checks => no completion assertion.

## Protection and serialization amendment

Lock reference labels in the editor if available; hiding and locking are session/UI
controls, not trusted persistent XML authorization. The validator enforces every reference
ID, label, shape type, coordinate/rotation, attributes, occlusion/group and presence.
Any substantive edit/deletion blocks ingestion. Do not synthesize a `locked` XML attribute.
CLI also reconstructs reference/prior indices from original frozen human sources; editing
the prepared baseline TSV to conceal an export change is blocked independently of UI locks.

C1.0 serialization tolerance is 0.0051 px/degree on fields CVAT serializes at two decimals
(ellipse center/radii separately). This is an explicit amendment to earlier 0.001 roundtrip
tolerance, not permission to change geometry. Canonical reference records keep exact final
human values; rounded export references never replace them. New centers must be inside
canonical image; partial boundary extents are allowed. STAR rotation remains zero.

## Staging, IDs and reconciliation

New stable canonical ID HNEW_STAR/LABEL_<panel>_<16-hex> hashes class/panel/shape type and
coordinates/rotation rounded to two decimals; CVAT shape IDs and export order are ignored.
Exact same-signature duplicates block staging. Distinct proposals can be flagged mutually.
Store frame, geometry, reviewer, UTC timestamp, confidence/status, version/origin and raw
export hash. Origin HUMAN_COMPLETENESS_PASS. No script here accepts or freezes new objects.

Conservative same-class/panel overlap screen uses oriented/ellipse axis envelopes:
IoU≥0.25 OR smaller-envelope containment≥0.50 OR center distance≤0.25 smaller diagonal.
Flags: MATCHES_PRIOR_REJECT, MATCHES_PRIOR_UNCERTAIN, MATCHES_PRIOR_NOT_REVIEWED,
POSSIBLE_DUPLICATE_CONFIRMED, POSSIBLE_DUPLICATE_NEW. Ellipse envelope containment can
overflag text inside a ring; flags are review questions, not matches or negatives.
Every conflict requires explicit post-review reconciliation; no old outcome is reinterpreted.

External reviewer-authorized freeze must follow completed panel review and reconciliation.
Its contract is in OBJECT_FREEZE_CONTRACT.md; UNCERTAIN_NEW is not a normal relation
endpoint. No production attachment task before this independent freeze gate.
