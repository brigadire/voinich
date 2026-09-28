# ANNOTATOR_B_AI Instructions

## Goal
Perform an independent visual annotation of astronomical marks, text labels, and structural elements across the 8 canonical astronomical panels:
f67r1, f67r2, f67v1, f68r1, f68r2, f68r3, f68v1, f68v2.

## Blindness Constraints
Annotator B operates strictly under blind conditions:
- No access to existing Annotator A annotations.
- No transcription or translation information (e.g., ZL3b, EVA, Stolfi records).
- No external dictionary, frequency table, or proposed celestial identifications.
- Annotations record visible surface marks only.

## Annotation Tasks
1. Identify all primary diagram objects (`STAR_OBJECT`, `CIRCLE`, `RADIAL_LINE`, `SECTOR`, `CENTRAL_OBJECT`, `MOON_OR_DISC_OBJECT`, `TEXT_ARC`, `OTHER_DIAGRAM_OBJECT`).
2. Identify all anonymous physical text labels (`B_LABEL_0001`, `B_LABEL_0002`, ...).
3. Assign visible relations between labels and objects (`NEAREST_OBJECT`, `ADJACENT_TO`, `INSIDE_OBJECT`, `ON_OBJECT`, `BETWEEN_OBJECTS`, `SECTOR_LABEL`, `RING_LABEL`, `UNASSIGNED`).
