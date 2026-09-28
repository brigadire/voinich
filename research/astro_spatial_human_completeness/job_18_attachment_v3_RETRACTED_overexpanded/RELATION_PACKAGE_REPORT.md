# Attachment-v3 grouped review package

Status: READY_FOR_CVAT_IMPORT. Production queue generated from the validated C1.0
object freeze. It contains 200 explicit UNREVIEWED pair frames: 75 f68r1 and 125 f68r2.
There are 166 old-old pairs, 32 old-new pairs and 2 new-new pairs. All 9 confirmed new
relation-eligible f68r2 endpoints occur in at least one pair. One otherwise uncovered new
STAR received a single nearest-endpoint fallback candidate; this is a review question,
not an asserted relation. Filtered pairs are not negatives.

The 48 frozen attachment-v2 pairs are included for a blind attachment-v3 recheck. Their
old outcomes are not present in the UI and are never overwritten. All other v3 candidates
also start UNREVIEWED; confidence starts UNSET. No automatic relation decisions exist.

Each frame has exactly one STAR/LABEL pair. Both shapes share `group_id=frame+1`.
In CVAT use Appearance -> Color By -> Group. Keeping a separate frame per edge permits
the same canonical endpoint to participate in multiple independent edges (many-to-many).

Create a new CVAT task with ../CVAT_ATTACHMENT_COMPLETENESS_SCHEMA.json, upload IMAGES.zip,
then import ANNOTATIONS.zip as CVAT for images 1.1. For every frame select
VISUAL_LABEL_OF / UNASSIGNED / UNCERTAIN (or NOT_APPLICABLE_REGION only for a genuine
scope error) and set LOW/MEDIUM/HIGH. Export the entire task in CVAT for images 1.1.

f68r3 is absent: its limited eligibility scope still awaits explicit reviewer confirmation.
All other panels remain PANEL_NOT_APPLICABLE_FOR_LABEL_STAR_ATTACHMENT.

COPIED_TREE_VALIDATION=PASS
OBJECT_ENDPOINTS=557
CONFIRMED_NEW_OBJECTS=37
QUEUED_PAIRS=200
GROUPED_PAIR_FRAMES=200
PRIOR_V2_PAIRS_BLIND_RECHECK=48
F68R3_INCLUDED=NO
AUTOMATIC_RELATION_DECISIONS=0
