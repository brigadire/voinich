# Human adjudication package validation

## Result

The package completed human calibration, H-High, H-Consensus-QC, and H-Medium under frozen STAR/LABEL protocol v1.2. Those reviewed layers remain unchanged. Reviewer-confirmed visual caption attachment protocol v2 supersedes the spatial relation task for active review: 99 main pairs and 26 special f68v2 pairs. The original spatial task is retained unchanged. No M3, lexical matching, semantic crosswalk, transcription, or astronomical interpretation was run.

Canonical JPEGs were byte-checked against both frozen input manifests before copying. Source counts are the required final counts: AI1 = 293 STAR / 175 LABEL; AI2 = 330 STAR / 206 LABEL. A is treated as positive evidence only; its absence outside the substantially annotated `f68r1` panel is not treated as disagreement.

Frozen pairwise match files are reconciled deterministically into physical candidates with at most one record per source. The canonical candidate envelope is the union of source boxes; the smaller display bbox is the arithmetic mean of available source coordinates and is explicitly editable. All original bboxes and source IDs remain in provenance TSVs. No candidate is automatically accepted.

## Generated queues

- 451 STAR candidates; 385 HIGH priority.
- 291 LABEL candidates; 271 HIGH priority.
- 675 later-stage LABEL ↔ OBJECT relation candidates.
- H-Cal covers exactly `f68r1`, `f68r3`, and `f68v2`.
- Consensus QC uses a deterministic hash-ranked sample: 34/173 STAR (19.65%) and 18/91 LABEL (19.78%).

Prepared H1 CVAT XML/ZIP files contain only neutral candidate IDs, geometry, decision fields, and human confidence. LABEL v1.1 uses rotated rectangles for local/radial text and ellipses for complete ring paths. Inclined runs were completed in the initial calibration; all three targeted ring-follow-up ellipses were accepted. The UI contains no source/model identity, support, informative priority, transcription, token status, semantic mapping, or astronomical identity. Priority and neutral support are available only when explicitly generating H2 after H1 freeze.

## Automated validation

`python3 -m unittest discover -s tests -v` passes all checks for stable IDs, complete one-time source-ID preservation, upstream frozen source checksums, canonical image hashes, UI blindness, relation referential integrity, calibration panels, and CVAT coordinate round-trip at ≤ 0.001 px. `SHA256SUMS` is verified from the package directory.

## Status

```text
HUMAN_ADJUDICATION_PACKAGE=READY

CANONICAL_IMAGES_VERIFIED=YES

STAR_CANDIDATES=451
LABEL_CANDIDATES=291
RELATION_CANDIDATES=675

HIGH_PRIORITY_STAR_CANDIDATES=385
HIGH_PRIORITY_LABEL_CANDIDATES=271

CALIBRATION_PANELS=f68r1,f68r3,f68v2

CVAT_IMPORT_READY=YES

CVAT_ROUNDTRIP_TEST=PASS

HUMAN_BLINDNESS_CHECK=PASS

PROTOCOL_FROZEN=YES

READY_FOR_HUMAN_CALIBRATION=YES

READY_FOR_HUMAN_PRODUCTION_ADJUDICATION=YES

H_HIGH_COMPLETE=YES

H_CONSENSUS_QC_COMPLETE=YES

H_MEDIUM_COMPLETE=YES

REVIEWED_STAR_LABEL_LAYERS_FROZEN=YES

READY_FOR_RELATION_PACKAGE_PREPARATION=YES

RELATION_CVAT_TASK_READY=YES

LABEL_STAR_RELATION_PAIRS=324

ACTIVE_RELATION_PROTOCOL_VERSION=2

ATTACHMENT_V2_MAIN_PAIRS=99

ATTACHMENT_V2_F68V2_PAIRS=26

REVIEWER_PANEL_UNASSIGNED=142

RIGHT_RULE_FILTERED_NOT_REVIEWED=57

RELATION_SOURCE_PAIRS_EXCLUDED=351
```

The STAR/LABEL freeze gate is satisfied for confirmed reviewed records only. Attachment v2 uses final endpoint geometry without source suggestions, supports multiple LABELs per STAR, and records panel-rule decisions separately from individual review. Queue filtering never fabricates negative observations. See `HUMAN_LABEL_STAR_ATTACHMENT_V2_GUIDE.md`, `HUMAN_LABEL_STAR_ATTACHMENT_PROTOCOL_V2.md`, and `HUMAN_LABEL_STAR_ATTACHMENT_V2_VALIDATION.md`.

Attachment-v2 ingestion is complete and frozen: main 99 decisions, f68v2 26 decisions, and 142 reviewer-rule UNASSIGNED records combine into 267 explicit decisions (53 VISUAL_LABEL_OF, 171 UNASSIGNED, 43 UNCERTAIN). The 57 rule-filtered pairs remain NOT_REVIEWED. The final raw UNREVIEWED pair was completed only through explicit reviewer confirmation. See `HUMAN_LABEL_STAR_ATTACHMENT_V2_INGEST_VALIDATION.md` and `HUMAN_LABEL_STAR_ATTACHMENT_V2_FREEZE_REPORT.md`.

H-High is complete at `385 STAR` and `271 LABEL`. Consensus-QC is complete at `34/34 STAR` and `18/18 LABEL`. The deduplicated H-Medium review added `24 STAR` and `6 LABEL`, all `MODIFY`. Cumulative unique coverage is `444/451 STAR` and `290/291 LABEL`. `7 STAR` and `1 LABEL` LOW-priority candidates remain explicitly `NOT_REVIEWED`; `5 STAR` and `3 LABEL` reviewed cases remain `UNCERTAIN`. See `HUMAN_PRODUCTION_FREEZE_REPORT.md`.
