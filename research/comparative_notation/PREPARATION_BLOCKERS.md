# Preparation blockers and STOP decisions

These items require a substantive or provenance decision capable of changing
the future scientific result. The preparation pipeline therefore fails closed
instead of choosing the convenient option.

| ID | Scope | Blocking ambiguity / missing fact | Required resolution |
|---|---|---|---|
| B01 | all comparison | No independent calibration panel and frozen scale `s_i` exists for every scalar metric. Estimating it from VM/candidate pairs would be post hoc. | Preregister controls, estimator, versioned scale TSV, and uncertainty policy before any VM distance. |
| B02 | VM reference | Several generic entropy, distribution, curve, preferred/depleted, and hierarchy metrics are not in the existing frozen VM catalog. | Compute them once with the generic analyzer after registry review, freeze the new VM artifact, never per candidate. |
| B03 | rarefaction | “Sample N tokens” does not determine whether samples are contiguous windows, documents/lines, or unordered tokens; these choices change sequence/line metrics. | Preregister sampling unit, boundary preservation, replicate allocation, CI estimator, and treatment of short documents. Checkpoints remain frozen at 5k/10k/20k/39380. |
| B04 | distributions | Scalar core and JS/Wasserstein/curve distance functions exist, but the versioned distribution/CI serialization and bootstrap output contract is not yet frozen. | Freeze bins/support alignment, bootstrap unit, CI level, and TSV schemas before production. |
| B05 | C03 | No open scholarly machine-readable historical shorthand running-text corpus with adequate provenance was identified. | Select/acquire a qualifying tradition-specific corpus. BDD cannot stand for shorthand in general. |
| B06 | C08/C09 | No concrete scholarly historical numerical/table transcription has been selected. Domain and layout sampling can materially affect structure. | Select sources and freeze inclusion rules; synthetic fixtures cannot replace historical corpora. |
| B07 | C01/C02/C04/C05/C07 | BDD project terms, DECODE record rights, Fontana scholarly transcription rights, and ECOLM full-text reuse terms require source-level verification. | Record exact licenses/permissions and immutable versions in SOURCE_PROVENANCE.json. |
| B08 | C04 | DECODE records still require mechanism/key stratification and alignment-quality screening. | Freeze monoalphabetic, homophonic, and nomenclator inclusion lists separately. |
| B09 | C06 | MEI work selection, validation schema version, event tie-break across voices, treatment of ligatures/rests, and historically meaningful system boundaries can alter all representations. | Preregister rules on a reviewed MEI subset; run MUSIC-R1/R2/R3 without post-hoc selection. |
| B10 | C07 | ECOLM tradition labels and TabCode conventions must be verified per source; rhythmic-sign tokenization remains source-convention dependent. | Freeze separate French/Italian/German adapter profiles and inspect TAB-R1/TAB-R2 fixtures against real source fragments. |
| B11 | C10 | First/second-order transition parameters and C-GRAMMAR production rules are not yet frozen. | Publish generator parameter files before generating calibration corpora. |
| B12 | class inference | Within-class placement is meaningful only with at least three independent corpora and a frozen pairwise normalization. | Acquire sufficient independent corpora; otherwise report individual corpus results only. |

No blocker was resolved by inspecting or optimizing a candidate-to-VM result;
no such result was computed.

## Closure: B01-B04 (Comparative Notation Study — Global Freeze Completion)

B01-B04 are closed as of this task. B05-B12 are unchanged above; no
candidate source/licensing/representation decision was made or implied.

| ID | status | closed_by | artifact | artifact_sha256 | date |
|---|---|---|---|---|---|
| B03 | CLOSED | Global Freeze Completion task | `RAREFACTION_PROTOCOL.md` | `d235ffecfcf609d5e89b3e4e5fc28e89c65e115d931f276dacf75491517b9672` | 2026-08-30 |
| B03 | CLOSED | Global Freeze Completion task | `RAREFACTION_SCHEMA.md` | `868a68a1628ab8cdd375c1596e938a9fbcd03c841b650ec632e976f6d78550d2` | 2026-08-30 |
| B04 | CLOSED | Global Freeze Completion task | `DISTRIBUTION_OUTPUT_CONTRACT.md` | `8b83ee83ab2d648b0bc03bf12a600668e3391198a1b1d26b44129297444df4bd` | 2026-08-30 |
| B04 | CLOSED | Global Freeze Completion task | `BOOTSTRAP_PROTOCOL.md` | `707e2facccc535dd22b52135b76084a86a5e12b50dd6dccfc8dd2d4a6c50dd81` | 2026-08-30 |
| B01 | CLOSED | Global Freeze Completion task | `CALIBRATION_PANEL_SPEC.md` | `a063149116f2f59d135ded5710fe4399344c05ef0ca6f936b857cb9c7ac2b7c7` | 2026-08-30 |
| B01 | CLOSED | Global Freeze Completion task | `CALIBRATION_SCALES.tsv` | `30539a09f287b8b3902ecc875538b3c07b52bcd4659c44e7d853cd354c7453f7` | 2026-08-30 |
| B01 | CLOSED | Global Freeze Completion task | `CALIBRATION_PANEL_REPORT.md` | `161791b40dc146db6d10fd234f2726854486d2d6adf7e5023f96ee673e5f59cc` | 2026-08-30 |
| B02 | CLOSED | Global Freeze Completion task | `VM_REFERENCE_V2.fingerprint.json` | `90d2254a7c9ab25c3c1a3167d3f0f4f6afc69689d3f3477e46486484fd938f42` | 2026-08-30 |
| B02 | CLOSED | Global Freeze Completion task | `VM_REFERENCE_RECONCILIATION.md` | `44f72d9d4ba432a16eb5794e4b0696c5b20e05651954982299fdf74cb9ecd617` | 2026-08-30 |

`GLOBAL_COMPARISON_PROTOCOL_FROZEN=true`. This does **not** imply
`PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=true` — see
`GLOBAL_FREEZE_REPORT.md` and section 47 of the task: each of C01-C09 still
requires its own listed B05-B12 blockers closed before any production run,
and B12 blocks only class-level inference, not an individual corpus report.
