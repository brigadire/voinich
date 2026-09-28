# Human calibration report

- Status: `COMPLETE — READY FOR H-HIGH`
- Protocol version: `v1.2`
- Protocol SHA-256: `9b33c65ac744041c765ca79d748988966059c1fdfbaa7c71704aedb712cdcfa6`
- Reviewer ID: `R01`
- Review dates (UTC): STAR `2026-09-03`; LABEL `2026-09-10`
- Calibration panels: `f68r1`, `f68r3`, `f68v2`
- STAR records reviewed: `207` (`54 ACCEPT`, `61 REJECT`, `88 MODIFY`, `4 UNCERTAIN`)
- LABEL records reviewed: `93` (`10 ACCEPT`, `20 REJECT`, `60 MODIFY`, `3 UNCERTAIN`) after three-record ring follow-up (`3 ACCEPT`)
- Human confidence: `HIGH` for every record by explicit reviewer direction on `2026-09-10`

## Ambiguities observed

- Faint star-like marks occurred and were treated as separate objects.
- Partial stars at panel edges occurred and were accepted.
- Overlapping stars were always treated as separate objects.
- STAR versus decoration was decided from visible star form.
- No reviewed STAR case established a bbox-margin rule for collision with adjacent text because no such collision was observed.
- All inclined/local text runs were processed in the initial LABEL calibration; a non-zero angle does not by itself make a LABEL radial or circular.
- A focused follow-up resolved all three complete-ring candidates as `ACCEPT`; each complete circular writing path is one LABEL.
- `SPLIT` and `MERGE` were not exercised, so their operational boundary was not calibrated.

## Operational clarifications

### STAR_OBJECT

1. A faint star-like mark is a separate STAR_OBJECT when its visible form is recognizably star-shaped.
2. A partial edge star is accepted as an object.
3. Overlapping stars are always separate objects.
4. Decorative strokes are distinguished from stars by visible star form, without astronomical interpretation.
5. The calibration set provides no empirical rule for STAR bbox overlap with neighbouring text; such a production case must be left UNCERTAIN pending targeted clarification.

### LABEL

1. Several words on one continuous line form one physical LABEL run.
2. A candidate may not be moved to a different text location. A candidate with no text at its location is rejected or left UNCERTAIN.
3. New LABEL candidates are not created in candidate adjudication.
4. Each complete continuous text ring is one LABEL; all three focused ring candidates were accepted.
5. Inclined and radial/local runs use the same physical-run rule and are calibrated. `SPLIT`/`MERGE` were not needed in calibration; production uses the definitions in protocol v1.2.

## Rules frozen for production

STAR and LABEL rules above are frozen in protocol v1.2. H-High may start. Calibration `UNCERTAIN` records remain unresolved and must not be promoted automatically. Relation adjudication remains gated until production STAR and LABEL layers are frozen.

## Calibration agreement metrics

These are calibration-subset diagnostics, not full-corpus estimates. `ACCEPT` and `MODIFY` both confirm candidate existence.

- `HUMAN_ACCEPT_RATE_AI_CONSENSUS=0.948052` (`n=77`)
- `HUMAN_ACCEPT_RATE_AI1_ONLY=0.033898` (`n=59`)
- `HUMAN_ACCEPT_RATE_AI2_ONLY=0.955882` (`n=68`)
- `HUMAN_REJECT_RATE_AI_CONSENSUS=0.038961` (`n=77`)
- `HUMAN_UNCERTAIN_RATE=0.019324` (`n=207`)
- `A_AI1_HUMAN_AGREEMENT=1.000000` on `f68r1` (`n=26`)
- `A_AI2_HUMAN_AGREEMENT=1.000000` on `f68r1` (`n=28`)
- `AI1_AI2_HUMAN_AGREEMENT=1.000000` on `f68r1` (`n=27`)

Machine-readable values are in `exports/HUMAN_CALIBRATION_AGREEMENT_METRICS.json`. AI agreement was used only for grouping and never for automatic promotion.

## Unresolved cases

### STAR

- `HOBJ_f68r3_12AE99CCE783`
- `HOBJ_f68r3_3A1ED95C8F74`
- `HOBJ_f68r3_AFE2C5FFA316`
- `HOBJ_f68v2_8070891630D4`

### LABEL

- `HLABEL_f68r3_19EDC31524DE`
- `HLABEL_f68r3_30626DCDB0AC`
- `HLABEL_f68v2_B1CEDCD5FC66`

## Protocol-change decision

- [x] Focused ring follow-up completed; protocol v1.2 is frozen for production.
- [ ] Substantive ontology change; issue a new protocol version and repeat all calibration.
