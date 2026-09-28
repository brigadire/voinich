# Full-page grouped attachment package

Status: READY_FOR_CVAT_IMPORT. Replaces all crop/pair packages.

PANELS=3
PANEL_IDS=f68r1;f68r2;f68r3
FULL_ORIGINAL_IMAGES=YES
ENDPOINTS=256
NEW_ENDPOINTS=17
FROZEN_V2_POSITIVE_EDGES=48
INITIAL_CONNECTED_GROUPS=46
F68R3_INITIAL_POSITIVE_EDGES=5
F68R3_INITIAL_CONNECTED_GROUPS=3
F68R3_FULL_PAGE_SCOPE=REVIEWER_CONFIRMED
GROUP_SEMANTICS=VISUAL_ASSOCIATION_HYPEREDGE
CARTESIAN_EDGES_INFERRED=NO
AUTOMATIC_NEW_RELATIONS=0
PREFLIGHT_VALIDATION=PASS

All endpoint geometry is the exact confirmed C1.0 object freeze geometry. Every endpoint
appears once on its full canonical page. Existing positive v2 edges initialize groups;
no v2 UNASSIGNED/UNCERTAIN decision is converted into a group. Reviewer grouping creates
a new version 3G1 result without rewriting v2.
