# Schema changelog

## 0.9.0-pilot — 2026-08-31

- Created a 23-descriptor image-only pilot schema.
- Chose coarse count bins and four distinct uncertainty/missing states.
- Excluded animal-like count, grid/table organization, symmetry, captions/labels, and semantic action categories from the minimal draft because their boundaries were not sufficiently stable during source review.
- Recorded Yale IIIF as the sole primary image source.
- Freeze withheld: no second independent pass and no adjudicated foldout-panel crop registry exist yet.

No textual result was used to add, remove, merge, or threshold a descriptor. Strict clean-room certification is withheld for the reason recorded in the final report.

## 1.0.0-rc — controlled revision

The revision used only the two locked clean-room pilot passes and the predeclared reliability gate.

- `COMP_ILLUSTRATED_AREA`: merged `MEDIUM`/`HIGH` boundary into `LOW` (through one half) versus `HIGH` (over one half).
- `COMP_ILLUSTRATION_REGIONS`: merged `2-3` and `4+` into `MULTIPLE`.
- `COMP_DOMINANT_LAYOUT`: merged `SINGLE_CENTRAL` and `DIAGRAM_CENTRIC` into geometry-only `CENTRAL_VISUAL` and clarified `MIXED`.
- `OBJ_STAR_LIKE`: coarsened the unreliable count bins to `0` versus `1+`.
- `GEO_REPEATED_MODULES`: coarsened amount to `NONE` versus `PRESENT`.
- `TXT_EMBEDDED_IMAGE`: clarified that the writing-zone center must lie within a visibly closed boundary.
- Removed `TXT_SURROUNDS_IMAGE` because the three-side rule remained viewpoint-sensitive.
- Removed redundant `VIS_MOTIF_REPETITION`; its disagreements duplicated `GEO_REPEATED_MODULES`.

No descriptor was added. This is the only permitted revision cycle.

## 1.0.0 — final freeze

The complete two-annotator rerun retained 15 descriptors. Six RC descriptors were excluded under the predeclared final rule, without another wording or threshold change:

- `COMP_ILLUSTRATED_AREA` (weighted κ 0.489; agreement 0.783);
- `COMP_DOMINANT_LAYOUT` (α 0.733; agreement 0.826);
- `COMP_IMAGE_INTERLEAVED` (α 0.400; agreement 0.696);
- `GEO_CONNECTED_NETWORK` (prevalence failure, α -0.047 despite agreement 0.870);
- `GEO_REPEATED_MODULES` (α 0.651; agreement 0.957);
- `TXT_EMBEDDED_IMAGE` (α 0.685; agreement 0.870).

All retained descriptors meet the fixed gate. Version 1.0.0 is frozen; hashes bind the schema, human-readable schema, protocol, and panel registry.
