# Human H-High review report

- Status: `COMPLETE — READY FOR H-CONSENSUS-QC`
- Protocol version: `v1.2`
- Reviewer ID: `R01`
- STAR source export: `exports/task_8_annotations_2026_09_12_10_59_03_cvat for images 1.1.zip`
- LABEL source export: `exports/task_9_annotations_2026_09_12_13_21_49_cvat for images 1.1.zip`
- Human confidence: `HIGH` for every final record by explicit reviewer direction

## Final H-High results

- STAR: `385` records (`112 ACCEPT`, `116 REJECT`, `154 MODIFY`, `3 UNCERTAIN`).
- LABEL: `271` records (`36 ACCEPT`, `87 REJECT`, `145 MODIFY`, `3 UNCERTAIN`).
- STAR provenance: `179` frozen calibration records and `206` new H-High reviews.
- LABEL provenance: `82` frozen calibration records and `189` new H-High reviews.
- Cumulative unique coverage including calibration-only records: `413/451 STAR` and `282/291 LABEL` candidates.

Repeated calibration candidates were not re-adjudicated. Their frozen calibration decisions replace the neutral or duplicate values in the full H-High CVAT exports.

## Explicit post-export clarifications

The raw CVAT ZIP files remain unchanged. `exports/HUMAN_HIGH_REVIEW_OVERRIDES_R01.tsv` records the reviewer's explicit clarifications:

- `HOBJ_f67r1_3C48B2BAAA37`: the resized rectangle is intentional; decision corrected from `ACCEPT` to `MODIFY` and reviewed geometry retained.
- `HLABEL_f68v1_3B985EFBA4D9`: the ellipse shift was accidental; decision remains `ACCEPT` and original candidate geometry is restored.
- `HOBJ_f67v1_7D4D96035479`, `HOBJ_f67v1_AC0352A0A7AB`, and `HOBJ_f67v1_D3A2C03DFB8C`: `ACCEPT`; each is a visible star-like mark on the adjacent-page strip captured at the right edge of the scan.

For new H-High records, non-`MODIFY` geometry is restored from the frozen candidate layer. This removes harmless CVAT decimal rounding and prevents accidental geometry edits from being promoted.

## Unresolved records

These frozen calibration decisions remain unresolved and are not promoted automatically:

### STAR

- `HOBJ_f68r3_12AE99CCE783`
- `HOBJ_f68r3_AFE2C5FFA316`
- `HOBJ_f68v2_8070891630D4`

### LABEL

- `HLABEL_f68r3_19EDC31524DE`
- `HLABEL_f68r3_30626DCDB0AC`
- `HLABEL_f68v2_B1CEDCD5FC66`

## Next phase

Of the deterministic Consensus-QC samples, the completed blind reviews already cover `27/34 STAR` and `16/18 LABEL` records. Only `7 STAR` and `2 LABEL` candidates remain for H-Consensus-QC. Relation adjudication remains gated until Consensus-QC is completed and the production STAR/LABEL layers are frozen.
