# LABEL–STAR visual attachment v2 validation

Status: READY FOR ATTACHMENT V2 REVIEW. Reviewer confirmed the changed question on 2026-09-14. No STAR/LABEL ontology, geometry, or original decisions were changed.

The 324 confirmed-endpoint source pairs partition without overlap or loss:

- Main task: 99 pairs (f68r1: 29, f68r2: 19, f68r3: 51), plus 3 original-page context frames. Image data ZIP: 207,641,137 bytes (approximately 208 MB).
- Special f68v2 task: 26 pairs, plus 1 context frame. Image data ZIP: 52,969,507 bytes (approximately 53 MB). Initial state UNREVIEWED, not automatic UNCERTAIN.
- Explicit reviewer-panel-rule decisions: 142 UNASSIGNED, with R01, UTC timestamp, protocol version 2, and REVIEWER_PANEL_RULE provenance.
- Right-rule filtered: 57 RULE_FILTERED_NOT_REVIEWED; no negative human decision fabricated.

Main f68r1/f68r2 pairs pass both right-center and oriented-envelope vertical-overlap screening. f68r3 has no directional screening. Screening never produces automatic ASSIGNED. The original package's 351 excluded unconfirmed/non-STAR endpoint pairs remain excluded.

The v2 interface has a single relation_decision control and no spatial checkboxes. ASSIGNED exports VISUAL_LABEL_OF, not ADJACENT_TO or NEAREST_OBJECT. Multiple LABELs may independently attach to one STAR; no one-to-one constraint or forced LABEL merge is applied.

All reused crop/context files preserve version-1 bytes and original final endpoint coordinates. The old 324-pair task remains intact. New package input/file hashes and v2 protocol hash are recorded in relations_v2/RELATIONS_V2_MANIFEST.json and the main manifest.

Automated validation covers exact partition, reviewer-decision provenance, directional screening, byte-identical image reuse, valid schema/defaults, distinct UNREVIEWED, ASSIGNED/UNASSIGNED/UNCERTAIN import, many-to-one preservation, immutable endpoint geometry/panel/IDs, rejection of v1 checkbox data, and complete CLI import for both subsets. Live import into the user's CVAT instance has not been performed.
