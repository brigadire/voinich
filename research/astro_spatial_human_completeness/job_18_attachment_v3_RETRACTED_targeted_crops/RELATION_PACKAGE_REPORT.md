# Attachment-v3 grouped targeted review package

Status: READY_FOR_CVAT_IMPORT. This package replaces and retracts the earlier 200-frame
overexpanded package. It contains exactly 57 UNREVIEWED pair frames on eligible panels:
29 f68r1 and 28 f68r2.

48 frames are the exact frozen attachment-v2 candidate pairs, included for a fresh blind
v3 recheck. Their old outcomes are neither displayed nor overwritten. The remaining 9
frames provide nearest-counterpart coverage for every one of the 9 new eligible f68r2
endpoints (7 new STAR, 2 new LABEL). Eight use the nearest right-side/Y-band candidate.
One STAR has no directional candidate and receives one explicitly marked nearest-any-side
fallback; this remains an UNREVIEWED question, not an asserted relation.

No newly generated old-old candidates remain. In particular HC_REL_1B9BFEF87B01B50B,
HC_REL_0A3D68A55420C4D6 and HC_REL_0A5ECC81332131BC are absent.

Each frame contains exactly one STAR/LABEL pair. Both endpoint copies share
`group_id=frame+1`; use CVAT Appearance -> Color By -> Group. Separate frames preserve
independent many-to-many edges. Choose VISUAL_LABEL_OF / UNASSIGNED / UNCERTAIN and
LOW/MEDIUM/HIGH for every frame. No decision is prefilled.

Create a new task using ../CVAT_ATTACHMENT_COMPLETENESS_SCHEMA.json, upload IMAGES.zip,
then import ANNOTATIONS.zip as CVAT for images 1.1. Do not continue the retracted CVAT task.

f68r3 remains excluded pending a formally confirmed limited eligibility rule.
All other panels are not applicable for LABEL-to-STAR review.

COPIED_TREE_VALIDATION=PASS
OBJECT_ENDPOINTS=557
CONFIRMED_NEW_OBJECTS=37
QUEUED_PAIRS=57
EXACT_PRIOR_V2_PAIRS=48
NEW_ENDPOINT_TARGETED_PAIRS=9
AUTOMATIC_RELATION_DECISIONS=0
