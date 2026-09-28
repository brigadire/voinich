# LABEL–STAR relations: CVAT setup and review

## Import

Create a **new standalone image task** (or a new project with the same schema). Do not reuse a STAR/LABEL task with the older labels.

1. Paste `relations/CVAT_LABEL_STAR_RELATION_SCHEMA.json` into **Raw labels**, click **Done**, and verify the exact names `LABEL_ENDPOINT` and `STAR_ENDPOINT`.
2. Upload **data**: `relations/LABEL_STAR_RELATIONS_IMAGES.zip`. Do not upload the original eight-image folder: this task uses individually named pair crops.
3. Create the task; import **annotations** from `relations/cvat/label_star_relations_h1.zip` using **CVAT for images 1.1**.
4. There are 324 pair frames and 7 additional `ZZ_CONTEXT_*.jpg` original-page overview frames. The overview frames are reference-only and need no annotation.

The task uses native CVAT boxes/rotated boxes, ellipses, and attributes. Format reference: [official CVAT image format](https://docs.cvat.ai/docs/manual/advanced/formats/format-cvat/).

## One frame = one proposed pair

- Cyan `LABEL_ENDPOINT`: the confirmed physical writing run or complete ring path.
- Orange `STAR_ENDPOINT`: the confirmed star-like mark.
- Select the orange STAR and change **attributes only**. Do not move, resize, rotate, delete, reclassify, or create shapes. Both geometries are frozen.
- Candidate IDs are immutable. Other visible stars/text provide context, not additional candidates to create.
- Images are lossless crops of decoded canonical JPEG pixels, at original resolution, translated by an integer offset only. No ink is generated, erased, rotated, or rescaled.
- If the crop is insufficient, inspect the corresponding `ZZ_CONTEXT_<panel>.jpg` overview. Read the panel from the immutable ID: `HOBJ_f68r2_...` means `f68r2`; the queue TSV also records the mapping. If the evidence remains insufficient, use `UNCERTAIN`.

## Attributes on STAR_ENDPOINT

Set `relation_decision` for every pair:

- `ASSIGNED`: a visible spatial relation exists. Select at least one of the checkboxes below; several may be true.
- `UNASSIGNED`: no defensible visible spatial relation for this proposed pair. Leave all checkboxes false.
- `UNCERTAIN`: the available image does not resolve the relationship. Leave all checkboxes false; this remains unresolved.
- `UNREVIEWED`: initial state only; it must not remain in the completed export.

Spatial checkboxes are operational descriptions, not semantic attachment or astronomical identity:

- `adjacent_to`: the text run/ring path is visibly beside this star.
- `nearest_object`: this star is the visibly nearest STAR to the writing unit in the observed layout. Check surrounding stars, including the overview when needed; proximity does not establish what the text means.
- `between_objects`: the writing unit is visibly between this star and another star. Describe the other visible mark in notes if needed; do not create a second endpoint.
- `on_object`: writing is visibly placed on the star's outline/body.
- `inside_object`: writing is visibly inside the star's body/boundary.

Do not infer relationships from bbox overlap alone. A ring ellipse denotes the writing path, **not the blank centre**, and a rotated LABEL box denotes the writing run, not all pixels inside its envelope. `nearest_object` and `adjacent_to` may coexist when both observations are clear.

`human_confidence` is initially `HIGH` per the reviewer's prior instruction; it describes certainty in the selected decision, including certainty that a case is unresolved. Notes must describe visible layout only. Do not transcribe text, name stars, or guess semantic identity. No source/model names, support counts, source relation suggestions, or priorities are exposed in H1.

## Export

Save annotations, export **CVAT for images 1.1**, and place the ZIP in `exports/`. Import with:

```bash
python3 scripts/import_label_star_relations.py --input exports/reviewed_relations.zip --output exports/HUMAN_LABEL_STAR_RELATIONS_R01.tsv --reviewer-id R01
```

The importer rejects missing/new shapes, changed endpoint IDs/geometry, unreviewed decisions, and inconsistent flags. Real CVAT's two-decimal serialization is tolerated up to 0.0051 pixel/degree without changing the frozen original endpoint geometry. Multiple checked spatial types are preserved as a semicolon-separated `relation_type`; UNASSIGNED and UNCERTAIN remain explicit.

## Scope

Only original source-proposed pairs whose two endpoints are confirmed by HUMAN (`ACCEPT`/`MODIFY`, final STAR class) are presented. Rejected, unresolved, unreviewed, and non-STAR endpoints are excluded with an audit reason. No new nearest-neighbour pairs are inferred. This is candidate relation adjudication, not exhaustive discovery of all manuscript links.
