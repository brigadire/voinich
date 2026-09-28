# Human calibration ingest validation

Status: **PASS — calibration complete**.

## Successfully ingested

- STAR: 207/207 candidate IDs preserved; 54 ACCEPT, 61 REJECT, 88 MODIFY, 4 UNCERTAIN.
- Initial LABEL: 93/93 candidate IDs preserved; ring follow-up replaced exactly three matching records.
- Final LABEL: 10 ACCEPT, 20 REJECT, 60 MODIFY, 3 UNCERTAIN.
- LABEL geometry: 89 rotated/ordinary boxes and 4 ring ellipses survived the CVAT export.
- No candidate was deleted and no unknown candidate ID was introduced.

## Human confidence override

On 2026-09-10 the reviewer explicitly directed that every calibration record receive `human_confidence=HIGH`. The importer applied this reviewer-directed override to all 207 STAR and 93 LABEL TSV records. Raw CVAT XML/ZIP files remain unchanged and retain their original `UNSET` values for provenance; the override is neither inferred from AI support nor silently applied.

## Ring follow-up

The focused CVAT export contains exactly three expected ellipse candidate IDs. All three received `ACCEPT` and `HIGH`; they replaced the corresponding initial `UNCERTAIN` rows in the final LABEL TSV. The reviewer's latest focused decision supersedes the earlier preliminary statement that only one ring appeared real.

The redundant LABEL `uncertain` checkbox was `true` on all raw CVAT records. Ingest now treats `human_decision` as authoritative and deterministically normalizes the flag to `true` for exactly the six UNCERTAIN decisions; this does not infer a human judgment.

The imported TSVs contain complete decisions, geometry, and confidence. `HUMAN_CALIBRATION_REPORT.md` is complete and protocol v1.2 is frozen. Production H-High may begin; unresolved records stay unresolved.
