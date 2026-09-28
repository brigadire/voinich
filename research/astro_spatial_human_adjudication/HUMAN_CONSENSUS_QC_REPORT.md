# Human Consensus-QC report

- Status: `COMPLETE — H-MEDIUM REQUIRED`
- Protocol version: `v1.2`
- Reviewer ID: `R01`
- New STAR records: `7` (`6 MODIFY`, `1 UNCERTAIN`)
- New LABEL records: `2` (`2 MODIFY`)
- Human confidence: `HIGH` for every final record by explicit reviewer direction

The new exports exactly match the deduplicated pending queues. No candidate IDs are missing, duplicated, or unexpected. All substantive geometry changes have decision `MODIFY`; the one `UNCERTAIN` STAR retains frozen candidate geometry.

## Complete deterministic sample

Previously completed blind calibration/H-High records and the new pending reviews together cover the full deterministic samples:

- STAR: `34/34`; `32 MODIFY`, `2 UNCERTAIN`, existence-positive rate `94.12%`, reject rate `0%`, modification rate `94.12%`.
- LABEL: `18/18`; `18 MODIFY`, existence-positive rate `100%`, reject rate `0%`, modification rate `100%`.

Consensus therefore showed no observed false positives in this sample, but the near-universal geometry correction rate demonstrates that additional human geometry review is necessary. Per the frozen workflow, H-Medium is activated before the STAR and LABEL production layers can be frozen.

## Next phase

After excluding calibration, H-High, and Consensus-QC records, H-Medium contains `24 STAR` and `6 LABEL` candidates. Relation adjudication remains gated.
