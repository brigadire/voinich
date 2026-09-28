# Prospective three-way comparison plan, revision 1

This file is fixed before outcome aggregation. The analysis consumes existing frozen
datasets only and never launches another annotation pass. All outputs live here.
Seed: **20260914**. Bootstrap replicates: **2000**. STAR/LABEL human protocol: **1.2**;
spatial relations: **1**; caption attachment: **2**, strictly separate.

## Inputs, gate and provenance

Discover inputs from the four existing package checksum ledgers and actual manifests.
Register their paths, byte sizes and SHA-256, including original images, primary source
tables, pairwise matches, candidate provenance, phase queues/decisions, reports and raw
exports. Verify every ledger entry before aggregation. Validate all IDs, panel IDs,
source cardinality, source boxes, review codes, phase duplicates, endpoint integrity
and published counts. A critical inconsistency produces BLOCKER_REPORT.md, not a fix.
Reconstruct the deterministic constrained frozen reconciliation in memory solely to
verify its provenance; do not replace it or perform new geometric matching.

## Disjoint cohorts

Unit = reconciled physical candidate. ACCEPT and MODIFY with the expected final class
are confirmed; REJECT is rejected. UNCERTAIN is a separate unresolved cohort.
Missing review is NOT_REVIEWED, never a negative. SPLIT/MERGE or contradictory class
decisions require a blocker unless an unambiguous mapping exists in frozen data.
Primary strict = ACCEPT/MODIFY/REJECT, production panels only. Calibration panels
are f68r1, f68r3, f68v2 regardless of later H-High membership. Combined and calibration
are secondary. Phase membership is preserved independently; deduplicate identical
records, with calibration > High > Consensus-QC > Medium precedence for stage labels.
Other diagram classes with no human adjudication are not human negatives.

Support groups = both AI, AI1-only, AI2-only, neither AI; A-positive is an additional
stratum, not a negative comparator. No A-absence penalty, even on f68r1 in this analysis.

## Candidate and pairwise metrics

Report all numerator/denominator pairs. Conditional confirmation/rejection/ACCEPT/
MODIFY use strict reviewed denominators; uncertainty uses all reviewed denominators.
Additionally report confirmation among all reviewed (with uncertainty unresolved).
Conditional coverage uses confirmed candidate union, not page-level recall/precision.
Pairwise AI agreement uses existing match ledgers: counts, per-source matched fractions,
matched-pair Jaccard = matches/(AI1+AI2-matches), class agreement and raw envelope IoU.
These ledgers required equal object classes: class agreement is selection-conditioned,
not an independent classification test. No artificial true negatives or missing labels.

Independent confidence fields on matched AI pairs permit an exploratory confidence
confusion matrix/raw agreement/Cohen kappa. Human confidence was explicitly forced HIGH:
no valid independent three-rater confidence kappa/alpha. Human confirmation is not an
independently sampled three-source categorical feature; no synthetic three-rater kappa.

## Geometry and conventions

Keep raw source boxes/orientations, candidate display and union boxes, and final human
shapes. Use STAR AABB IoU, Euclidean and panel-diagonal-normalized center distance,
absolute relative width/height error. LABEL CVAT BOX uses convex polygon IoU and axial
angle error (0/180 equivalent); ELLIPSE uses a 256-vertex filled-ellipse envelope proxy,
center, axes and axial orientation errors. Ellipse denotes a ring *path*, but no ring
thickness/path segmentation is recorded: filled-envelope IoU is NOT ink/ring overlap.

AI LABEL files provide bbox and orientation but no explicit oriented side-length schema.
Do not claim exact recovered rotated geometry. Primary derived proxy: for nonzero axial
orientation take longer bbox side as local text length and shorter as thickness, then
rotate at frozen angle; for zero angle retain bbox dimensions. Also calculate raw bbox
envelope overlap and raw-dimensions-rotated proxy sensitivity. For human ring candidates
the AI bbox becomes a filled ellipse-envelope proxy, explicitly not an AI ring prediction.
Candidate-to-human correction uses actual CVAT geometry, not the raw source proxy.
No new matching, so no matching threshold sensitivity; exploratory error flags use IoU
0.25/0.50/0.75, center/shape-diagonal 0.50, angle 30 degrees, side ratio 0.5/2.

Summaries: n, median, Q1/Q3/IQR and percentile 95% panel-clustered bootstrap CI by class,
panel, source, human unchanged/modified and geometry type. Resample panels with all their
observations, retaining within-panel dependence. For a one-panel cell use candidate
bootstrap and label its inability to support across-panel inference. Few panel clusters
and candidate-derived gold limit all inferential claims.

## Hypothesis families

Confirmatory descriptive contrasts on production, separately STAR/LABEL:
both vs AI1-only, both vs AI2-only, AI1-only vs AI2-only. Outcomes: confirmation/rejection
on strict reviewed, uncertainty on all reviewed, MODIFY on confirmed. Effect = risk
difference, 95% panel-clustered bootstrap CI; two-sided Fisher exact p supplementary
(object-level independence assumption). Holm correction within class categorical family
(12 tests). Small/empty groups are insufficient, not evidence of equality.
Continuous family per class: candidate-to-human 1-IoU on confirmed for three support
contrasts, plus matched-candidate paired AI1-minus-AI2 human IoU. Effect = median difference
(paired median difference for AI comparison), cluster-bootstrap CI and two-sided
bootstrap sign-tail p, Holm within each class continuous family. Mark geometric proxy
inference exploratory. Do not interpret a p-value alone.

Proportion CI: Wilson 95% (identified as conditional candidate-level interval), plus
panel-clustered bootstrap for headline confirmation and coverage. Bootstrap keys are
stable SHA-256-derived seeds; results independent of call order. Null/NA for zero n.

## Sensitivity and selection

Policies: primary production strict; uncertainty negative lower bound; uncertainty positive
upper bound; calibration only; combined strict; production High only; production High +
deduplicated Medium; production without A-positive candidates; production A-positive only;
geometry raw-envelope/raw-dimensions-rotated alternatives and overlap flag thresholds.
NOT_REVIEWED never enters any accuracy denominator. These are policy scenarios, not edits.

Case selection is deterministic, by neutral ID tie-break: for each class select one
confirmed low-correction case, one MODIFY maximal correction, one REJECT with both AI
support (else any REJECT), one UNCERTAIN, and one confirmed maximal paired AI IoU
discrepancy; deduplicate selections. Show canonical crop and all available AI/human
shapes; rejected/unresolved human shape is retained but never called confirmed geometry.
No semantic source notes are used for assigning morphology. Error taxonomy is geometric
proxy flags plus explicitly recorded review flags/confidence, not visual causal inference.

## Attachment v2

Verify 267 explicit decisions (53 positive/171 unassigned/43 uncertain), 57 filtered,
351 endpoint-excluded original pairs; 53 links/51 LABEL/51 STAR. Audit original AI
relation schema/type universes: absent independently compatible v2 predictions => human
descriptive graph only. Report origins, panels, degrees, components and review coverage
267/324 eligible source pairs and 267/675 original pairs; individual review 125/324 and
125/125 selected review tasks. 142 panel-rule negatives are not individual pair review.

## Reproducibility and completion

Deterministic tables, graphics with companion source TSVs, input manifest, tests,
source-preservation checks, twice-run byte comparison, checksums and self-contained
technical/short reports. No transcription, lexical/semantic matching, M3 or astronomical
identification. Completion requires all gates, not merely successful script execution.
