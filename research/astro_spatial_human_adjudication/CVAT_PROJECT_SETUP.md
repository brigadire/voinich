# CVAT project setup

Create two separate image-annotation projects/tasks. Upload the eight files from `images/` without renaming them.

**Before annotation import, open the created Task details and verify that its Labels section literally contains `STAR_OBJECT` (and `OTHER_OBJECT`) for STAR, or `LABEL` for LABEL.** Annotation import does not register missing task labels. If the task belongs to a project, edit/import the schema on the **project**, then create the task inside it; CVAT tasks inside projects inherit project labels and cannot use an independent task schema. For a standalone task, paste the schema in the task's Raw labels editor, press **Done**, and only then submit/create the task.

## STAR project

Import `CVAT_STAR_LABEL_SCHEMA.json` as the label definition (or reproduce it in the UI), then import one of `cvat/star_*_h1.xml` using **CVAT XML 1.1** (called **CVAT for images 1.1** in some UI versions). H1 is the default. Do not import LABEL rectangles into this project.

## LABEL project

Use `CVAT_TEXT_LABEL_SCHEMA.json`, verify that `LABEL` appears in Task details, then import `cvat/label_*_h1.xml` using the same native CVAT image/XML format. No transcription field is defined.

LABEL tasks require the current schema with `type: "any"`: local/radial runs are imported as rotated boxes and complete ring paths as ellipses. Do not use an older LABEL task created with rectangle-only schema. A LABEL task created before protocol v1.1 must be recreated; STAR tasks are unaffected.

The reviewer confirmed that all inclined/local runs were already processed. The separate three-record ring-only `cvat/label_calibration_followup_h1.xml` was completed with all three ellipse candidates accepted. Its human confidence was prefilled as `HIGH` per the reviewer's explicit blanket instruction.

H-High, Consensus-QC, and H-Medium are complete. The completed Medium queues contained `24 STAR` and `6 LABEL` candidates, all reviewed as `MODIFY`. Do not re-import these earlier queues for another review. The canonical reviewed STAR/LABEL snapshots are frozen at package root; rejected, unresolved, and NOT_REVIEWED candidates are not confirmed relation endpoints.

## Active LABEL–STAR visual attachment task v2

Reviewer confirmed the new question on 2026-09-14: "Does this physical LABEL visually function as the caption of this STAR?" This is not spatial adjacency/nearest-object scoring. Use `HUMAN_LABEL_STAR_ATTACHMENT_V2_GUIDE.md` and a **new** CVAT schema from `relations_v2/CVAT_LABEL_STAR_ATTACHMENT_SCHEMA.json` (no spatial checkbox attributes).

Main task: upload `relations_v2/main/IMAGES.zip` (99 pairs + 3 context frames), then import `relations_v2/main/ANNOTATIONS.zip` as CVAT for images 1.1. Separate f68v2 task: same schema, upload `relations_v2/f68v2/IMAGES.zip` (26 pairs + 1 context), import `relations_v2/f68v2/ANNOTATIONS.zip`. Edit only `relation_decision` on STAR_ENDPOINT; geometries and IDs remain frozen. Both tasks begin UNREVIEWED. Many LABELs can attach to one STAR.

142 other-panel pairs are recorded UNASSIGNED by the reviewer rule; 57 pairs are RULE_FILTERED_NOT_REVIEWED under the right-side rule. Import new exports with `scripts/import_label_star_attachment_v2.py`, selecting the correct subset. Do not continue the old 324-pair task for attachment v2.

## Archived spatial LABEL–STAR RELATION task v1

The original 324-pair spatial H1 package is retained unchanged under `relations/`; its guide is `HUMAN_LABEL_STAR_RELATION_GUIDE.md`. Its ADJACENT_TO/NEAREST_OBJECT/etc. observations are not attachment-v2 decisions. The original 675-record source relation queue remains provenance, not reviewed output.

If import reports `Label 'STAR_OBJECT' is not registered for this task`, stop and fix the Task/Project Labels section; retrying or changing the ZIP does not register the label. Label names are case-sensitive.

If saving reports `Trying to save an attribute ... invalid value`, the task was created from an older schema/XML with an empty attribute value. The current schemas declare explicit `default_value` values, and every H1 box carries neutral `priority=UNSET`. If review has not begun, recreate the task with the current schema and reimport the current XML. Updating files on disk does not retroactively replace attribute specifications already stored in an existing CVAT task.

Reviewers must not delete boxes. They set `decision`; only `MODIFY` permits geometry/class edits. Export as CVAT for images 1.1 and run `scripts/import_from_cvat.py`. H2 files are generated only after H1 freeze and expose neutral `candidate_support`/`priority` attributes.

Relation ingestion uses `scripts/import_label_star_relations.py`, not the STAR/LABEL importer. `HUMAN_CANDIDATE_RELATIONS.tsv` is the original source queue; `relations/LABEL_STAR_RELATION_QUEUE.tsv` is the filtered ready-to-review queue.

Example import command after review:

```bash
python3 scripts/import_from_cvat.py --task star --input reviewed.xml --output exports/HUMAN_STAR_ADJUDICATION.tsv --reviewer-id R01
```
