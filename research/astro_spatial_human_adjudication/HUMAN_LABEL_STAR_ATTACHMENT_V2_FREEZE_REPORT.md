# Visual caption attachment v2 — completed freeze

- Status: FROZEN_COMPLETE_FOR_SELECTED_QUEUE
- Reviewer: R01
- Freeze date: 2026-09-14
- Relation protocol: 2; STAR/LABEL protocol v1.2 unchanged

## Results

- Main task: 99 decisions — 48 VISUAL_LABEL_OF, 21 UNASSIGNED, 30 UNCERTAIN.
- f68v2 task: 26 decisions — 5 VISUAL_LABEL_OF, 8 UNASSIGNED, 13 UNCERTAIN.
- Reviewer panel rules: 142 UNASSIGNED.
- Cumulative explicit decisions: 267 — 53 VISUAL_LABEL_OF, 171 UNASSIGNED, 43 UNCERTAIN; HIGH confidence throughout.
- Right-rule filtered: 57 pairs, still NOT_REVIEWED, not negative human judgments.
- Confirmed attachment graph: 53 links involving 51 LABELs and 51 STARs; two STARs have multiple assigned LABELs. No uniqueness constraint was imposed or link removed automatically.

HREL_68BD06396134 was completed as UNCERTAIN by explicit reviewer confirmation at 2026-09-14T10:59:59Z. The override is preserved separately, and the raw main ZIP still contains its original UNREVIEWED value. The f68v2 legacy auxiliary schema adaptation is recorded in decision_origin and did not reinterpret active spatial flags or alter caption decisions.

## Frozen outputs and scope

Canonical output: HUMAN_LABEL_OBJECT_RELATIONS.tsv. Identical immutable export: exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_FINAL_R01.tsv. Metrics: exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_METRICS.json. Every output has relation_protocol_version=2 and explicit decision provenance.

`scripts/freeze_attachment_v2.py` reproduces the snapshot byte-identically and refuses conflicting overwrites. Manifest/SHA256SUMS freeze output hashes. Unresolved outcomes and unreviewed exclusions are not promoted. The original 351 unconfirmed/non-STAR endpoint exclusions and original spatial-v1 package remain unchanged.

VISUAL_LABEL_OF means a visible caption-to-star attachment only. No text reading, lexical matching, star identification, astronomical interpretation, M3, or semantic crosswalk was performed.
