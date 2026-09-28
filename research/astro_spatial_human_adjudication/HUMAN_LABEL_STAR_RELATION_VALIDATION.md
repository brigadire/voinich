# LABEL–STAR relation package validation

- Result: `PASS — READY FOR RELATION H1`
- Source queue: `675` frozen relation candidates.
- Selected: `324` source-proposed pairs with HUMAN-confirmed LABEL and final-class STAR endpoints (`ACCEPT` or `MODIFY`).
- Excluded: `351` pairs — `232 STAR_NOT_CONFIRMED`, `44 LABEL_NOT_CONFIRMED`, `75` with both endpoints unconfirmed. Non-STAR object types are not STAR endpoints.
- New pairs inferred beyond the source queue: `0`; source relations accepted automatically: `0`.
- Pair frames: `324`; unannotated full canonical-page context frames: `7`.
- Each pair frame has exactly one `LABEL_ENDPOINT` and one `STAR_ENDPOINT`, translated from the frozen final geometries by a recorded integer crop offset. Rotation and ellipse geometry are preserved.
- Pair PNGs equal decoded RGB pixel crops of the frozen canonical JPEGs, with no scale, rotation, drawn raster overlays, or lossy re-encoding. Context JPEGs preserve original bytes. Image data ZIP is approximately `630 MB`.
- H1 exposes no source/model names, support counts, priority, or source relation suggestions. Initial `UNREVIEWED` is distinct from legitimate `UNCERTAIN` or `UNASSIGNED`.
- The importer rejects altered endpoint geometry/IDs, changed frame dimensions, deleted/new/duplicated endpoints, incomplete decisions, and inconsistent relation flags. Real CVAT two-decimal serialization tolerance is `0.0051` pixel/degree; original canonical endpoint geometry remains frozen.
- Multiple spatial observations are preserved as a semicolon-separated relation_type. Every selected source pair receives one explicit human result; no inference from nearest-neighbour geometry is made.
- `relations/RELATIONS_MANIFEST.json` freezes source input digests, generated image/annotation file digests, crop pixel hashes, and Pillow version. The main package manifest also freezes this manifest and the relation guide.

Automated checks cover candidate filtering/provenance, endpoint translation, crop pixel identity, schema/default validity, H1 blindness, successful multi-type/UNASSIGNED/UNCERTAIN import, CVAT decimal rounding, and rejection of incomplete or corrupted review data. Live import into the user's CVAT instance has not been performed.
