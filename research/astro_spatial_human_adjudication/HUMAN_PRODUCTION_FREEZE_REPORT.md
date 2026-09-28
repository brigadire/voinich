# Human STAR/LABEL production freeze

- Status: `REVIEWED STAR/LABEL LAYERS FROZEN — READY FOR RELATION PACKAGE PREPARATION`
- Protocol: frozen `v1.2` (unchanged)
- Reviewer: `R01`
- Freeze date: `2026-09-13`
- Completed phases: H-Cal, H-High, H-Consensus-QC, conditional H-Medium
- Confidence: `HIGH` throughout, by explicit reviewer direction

## Medium ingestion

- `task_12_annotations_2026_09_13_10_53_04_cvat for images 1.1.zip`: `24/24 STAR`, all `MODIFY`.
- `task_13_annotations_2026_09_13_11_00_30_cvat for images 1.1.zip`: `6/6 LABEL`, all `MODIFY`.
- No missing, unexpected, duplicate, or moved-to-another-panel IDs. All supplied shapes are retained. Substantive geometry edits are authorized by `MODIFY`.
- Raw ZIP files remain unchanged. Direct TSV imports preserve the reviewed geometry and apply the reviewer-wide confidence override.

## Frozen cumulative outputs

- `HUMAN_STAR_ADJUDICATION.tsv`: `444/451` reviewed STAR candidates — `115 ACCEPT`, `206 MODIFY`, `118 REJECT`, `5 UNCERTAIN`.
- `HUMAN_LABEL_ADJUDICATION.tsv`: `290/291` reviewed LABEL candidates — `36 ACCEPT`, `163 MODIFY`, `88 REJECT`, `3 UNCERTAIN`.
- `HUMAN_PRODUCTION_UNREVIEWED.tsv`: `7 STAR` and `1 LABEL`, all LOW-priority and `NOT_REVIEWED`.

This freezes the reviewed subset, not a complete manual census of the manuscript. LOW-priority candidates not selected by the staged workflow remain unreviewed; source consensus is not automatically accepted. UNCERTAIN remains unresolved. Calibration and source records are preserved, and duplicate calibration/H-High records are merged only when identical.

`scripts/freeze_reviewed_layers.py` reproduces the canonical snapshots. `manifest.json` records their SHA-256 digests. A later correction must create a new revision, not overwrite this completed snapshot.

## Relation boundary

The next step is preparation of a separate blind spatial RELATION task using only confirmed final STAR/LABEL records (`ACCEPT` or `MODIFY`, with appropriate final class). Rejected, unresolved, and unreviewed candidates are not confirmed endpoints. The original 675-record source relation queue must first be reconciled with the final human layers; it is not itself a reviewed or ready-to-import relation layer.

The three accepted edge STAR records on `f67v1` denote visible marks on a captured adjacent-page strip. Acceptance confirms visible existence in the canonical scan, not manuscript-page ownership. Do not infer a relation to main-page LABELs merely because these records share the scan panel name.
