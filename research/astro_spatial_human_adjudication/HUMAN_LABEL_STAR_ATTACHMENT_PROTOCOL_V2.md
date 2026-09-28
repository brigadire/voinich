# LABEL–STAR attachment protocol v2 — frozen

Confirmed by reviewer R01 on 2026-09-14. This is a new relation protocol, not a reinterpretation of historical spatial-relation decisions. STAR/LABEL protocol v1.2 and its completed snapshots remain unchanged. The original 324-pair spatial task is retained as version 1.

## Question and decisions

Question: **Does this physical LABEL visibly function as the caption of this STAR in the manuscript layout?** Do not read/transcribe the text, identify a star, or infer astronomical/lexical meaning. Geometric adjacency or nearest distance alone is not enough to assert attachment.

- ASSIGNED: visually defensible caption-to-star attachment; output relation_type VISUAL_LABEL_OF.
- UNASSIGNED: no visually defensible attachment for this proposed pair.
- UNCERTAIN: attachment cannot be resolved from the available layout; unresolved.
- UNREVIEWED: interface initial state only; rejected by the completed-review importer.

One pair may have one attachment decision. Many LABELs may attach to one STAR; no uniqueness restriction, forced merge, or split is introduced. No new source pairs or endpoint objects are created. Endpoint geometry and IDs are frozen.

## Reviewer-confirmed panel rules

- f68r1 and f68r2: review only candidate pairs where LABEL is to the right of STAR and in its horizontal writing band.
- Operational coarse queue filter: LABEL center_x > STAR center_x AND the y-ranges of their oriented envelopes overlap (touching counts). This screens the queue only; it is not automatic confirmation or rejection of attachment.
- f68r3: review all confirmed-endpoint source pairs; no right-side filter.
- f67r1, f67v1, f68v1: UNASSIGNED by the reviewer's explicit panel-level instruction, recorded as REVIEWER_PANEL_RULE, not as individual pair inspection or AI consensus.
- f68v2: a separate special review, initially UNREVIEWED, not automatically UNCERTAIN. Two writing blocks may each attach to the same star. If unclear, choose UNCERTAIN; keep the existing physical LABEL units separate.

## Scope and audit

The original 324 confirmed-endpoint source pairs are partitioned into 99 main review pairs, 26 f68v2 special review pairs, 142 reviewer-rule UNASSIGNED decisions, and 57 pairs not queued by the right-side rule. The last 57 are RULE_FILTERED / NOT_REVIEWED, not negative human observations. Original exclusions for rejected, unresolved, unreviewed, and non-STAR endpoints remain in the version-1 package.

No original STAR/LABEL or relation export is overwritten. Confidence defaults HIGH per reviewer direction and can be explicitly changed in the interface. Notes describe visible layout only. Source/model identity, support count, source guesses, and informative priority remain hidden. Manifest digests freeze this protocol, source inputs, and generated tasks.
