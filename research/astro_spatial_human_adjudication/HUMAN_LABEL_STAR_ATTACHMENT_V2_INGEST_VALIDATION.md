# Attachment v2 ingestion validation

Status: COMPLETE — reviewer confirmed UNCERTAIN for HREL_68BD06396134; both subsets are imported and the cumulative v2 snapshot is frozen. The raw-export findings below are retained as ingestion history.

## Main export

Source: `exports/job_16_annotations_2026_09_14_10_16_23_cvat for images 1.1.zip`.

- 102/102 expected frames, 99/99 expected pairs; all endpoint IDs and geometry match the frozen v2 task within real-CVAT decimal tolerance.
- Decisions: 48 ASSIGNED, 21 UNASSIGNED, 29 UNCERTAIN, 1 UNREVIEWED.
- Unreviewed pair: HREL_68BD06396134, panel f68r3, LABEL HLABEL_f68r3_F59D510F9867, STAR HOBJ_f68r3_C1D20B14679B.
- Initial ingestion was blocked for this one UNREVIEWED pair. Reviewer explicitly confirmed UNCERTAIN on 2026-09-14; `exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_REVIEWER_OVERRIDES_R01.tsv` records that confirmation with UTC timestamp. It is applied to a copied XML tree, never to raw ZIP bytes.
- Final main import: 99 decisions — 48 VISUAL_LABEL_OF, 21 UNASSIGNED, 30 UNCERTAIN. The corrected pair has decision_origin REVIEWER_POST_EXPORT_CONFIRMATION; all other main records preserve their original export decision/timestamp.

## f68v2 export

Source: `exports/job_17_annotations_2026_09_14_10_51_55_cvat for images 1.1.zip`.

- 27/27 frames, 26/26 expected pairs; all three endpoint IDs and geometry match the prepared f68v2 task.
- Decisions: 5 VISUAL_LABEL_OF, 8 UNASSIGNED, 13 UNCERTAIN; HIGH confidence throughout.
- The task retained auxiliary spatial-v1 schema fields: five false checkbox values and no panel attribute. These are not active spatial observations.
- Explicit compatibility mode `--allow-unused-legacy-schema` discarded only those false flags on a copied XML tree and derived the missing panel from the verified frozen frame mapping. Original IDs, geometry, caption decisions, confidence, and raw ZIP were not changed. Active/invalid spatial flags, altered IDs/geometry, or UNREVIEWED remain errors.
- Imported output: `exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_F68V2_R01.tsv`, protocol version 2, decision_origin PAIR_REVIEW_LEGACY_SCHEMA_ADAPTED.

The 142 previously recorded reviewer-panel-rule UNASSIGNED decisions and 57 RULE_FILTERED_NOT_REVIEWED pairs remain unchanged. Neither semantic interpretation nor astronomical identification was performed.

## Complete cumulative freeze

`HUMAN_LABEL_OBJECT_RELATIONS.tsv` and `exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_FINAL_R01.tsv` contain the same 267 version-2 decisions: 53 VISUAL_LABEL_OF, 171 UNASSIGNED, 43 UNCERTAIN. Confidence is HIGH throughout. Of those, 125 are individual pair decisions (including the one post-export confirmation), and 142 are explicit reviewer panel rules. The remaining 57 queue-filtered pairs retain NOT_REVIEWED status separately. No spatial-v1 observations are mixed into the caption-attachment snapshot.
