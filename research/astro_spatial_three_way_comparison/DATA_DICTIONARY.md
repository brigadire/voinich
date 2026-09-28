# Data dictionary

Unit of object analysis: frozen physical candidate, not pixel, source record, or page.
Primary provenance is HUMAN_CANDIDATE_OBJECTS/LABELS.tsv. Source IDs retain A/AI1/AI2
namespaces. Original source classes/boxes/orientations are kept, not overwritten.

| Frozen decision | Analytic status | Strict metric | Geometry target |
|---|---|---|---|
| ACCEPT | CONFIRMED_UNCHANGED | positive | final frozen shape |
| MODIFY | CONFIRMED_MODIFIED | positive detection, correction separately measured | final frozen shape |
| REJECT | REJECTED | negative proposed candidate | not used as confirmed geometry |
| UNCERTAIN | UNCERTAIN | excluded; separate lower/upper sensitivity bounds | excluded |
| missing human row | NOT_REVIEWED | excluded in every accuracy policy | excluded |

SPLIT/MERGE are allowed in the protocol but absent in the frozen final records.
All observed final classes equal the proposed STAR_OBJECT/LABEL class. A REJECT row
retains a proposed shape/class; it does not independently classify a different object.
Confidence HIGH is the reviewer-wide explicit override, not calibrated certainty.

## Input/output schemas

INPUT_MANIFEST: repository-relative path, upstream package/version, bytes, sha256, role,
frozen=YES. All four ledgers are verified; unused ancillary upstream records remain
registered for source-preservation audit, but their lexical/semantic contents are not
analysed or exported. No nonexistent phase-specific manifest is presumed:
REVIEW_PHASE_INPUTS registers actual CVAT XML queues, final phase tables, counts/hashes;
the global human manifest and BUILD_SUMMARY register the phase freeze.
Consensus-QC final phase TSVs contain only the additional 7 STAR/2 LABEL; full task
samples contain 34 STAR/18 LABEL. Already-reviewed identical observations cover the
remaining 27/16. Full QC sample membership is retained, not counted as new human rows.

ANALYSIS_COHORTS / THREE_WAY_OBJECT_RESULTS:
candidate_id; panel; object_type STAR/LABEL; scope CALIBRATION/PRODUCTION; support_pattern
BOTH_AI/AI1_ONLY/AI2_ONLY/NEITHER_AI; a_positive_support and support_ai1/2 YES/NO;
priority HIGH/MEDIUM/LOW and frozen priority_reason; review_stage (first applicable phase)
and review_phase_memberships (all phases); human_decision; analytic_status; geometry_type;
source_annotation_ids; human_protocol_version=1.2; primary_strict/combined_strict,
uncertain_cohort/not_reviewed_cohort as 0/1. a_absence_negative_evidence is always 0.
Candidate geometry is display bbox plus rotation and union envelope; human geometry
comes from canonical final row; ai1/2/a IDs, raw bbox, class, orientation, confidence
are original records. Detection support does not imply acceptable geometry.

THREE_WAY_GEOMETRY_RESULTS contains confirmed targets only. source AI1/AI2/CANDIDATE,
source_id and human_id preserve provenance. geometry_type stratifies AABB,
ROTATED_RECTANGLE, ELLIPSE_ENVELOPE_PROXY. representation is raw_AABB,
orientation_normalized_proxy, raw_dimensions_rotated_proxy, filled_ellipse_envelope_proxy
or actual_candidate_display. primary_representation=0 only for alternate rotated proxy.
source_raw_bbox/final_human_bbox and orientations are retained. IoU is polygon area
intersection/union (filled envelope for ellipse, NOT ring ink); center_distance_px is
Euclidean, panel_norm divides by panel diagonal and shape_norm by target shape diagonal.
width/height_relative_error = abs(source-target)/target; width/height_ratio = source/target;
axial_angle_error_deg is min modulo-180 angle separation (NA for STAR AABB).
raw_envelope_iou compares raw AI bbox to axis-aligned envelope of actual human shape.
Missing metrics use blank TSV / JSON null, never invented zero.

GEOMETRY_SUMMARY_METRICS: n, median, q1/q3/iqr, mean, ci95_low/high, ci_method,
panel_clusters, bootstrap_replicates by scope/source/shape/representation/grouping/stratum
and metric. JSON embeds identical rows. Single-panel bootstrap cannot infer panel variability.

AI1_AI2_PAIRWISE_RESULTS retains frozen match_status MATCHED/LEFT_ONLY/RIGHT_ONLY,
both original and physical IDs, frozen raw-box IoU, retained human outcome if in scope,
independent source confidence, conditioned class agreement and same_physical_candidate.
PAIRWISE_RECONCILIATION_AUDIT includes all object classes (unreviewed other classes are
not human negatives). Constrained nontransitive pairwise edges can remain unjoined;
the reproduced frozen reconciliation, not a new IoU match, remains the analytic unit.

JSON fractions always contain numerator/denominator/estimate/ci95/ci_method;
headline confirmation/coverage also include cluster_ci95, cluster_ci_method and valid
bootstrap count. Strict confirmation denominators exclude uncertainty; all-reviewed
uncertainty denominators retain it. modification_among_confirmed conditions on ACCEPT/MODIFY.
All rates are conditional on proposed/adjudicated candidate scope, not page precision/recall.

STATISTICAL_TESTS: family class_categorical or class_continuous; group sizes/events;
left-minus-right risk difference or median difference; CI and valid bootstrap count;
Fisher supplementary p (candidate independence) or exploratory bootstrap sign-tail p;
p_holm within predeclared family. Paired human IoU is median of per-candidate differences.
Small groups/degenerate intervals/few clusters do not establish equivalence.
Independent AI-confidence raw agreement and conditioned fixed-class agreement also
include explicit numerator/denominator/Wilson CI in JSON; these are not object validity.

SENSITIVITY_RESULTS policies never change raw decisions. UNCERTAIN positive/negative
are explicit mathematical bounds; not-reviewed remains excluded. Geometry sensitivity
changes representations/diagnostic flags only, not matching or confirmation.

Attachment tables use relation_protocol_version=2, never spatial-v1 categories.
VISUAL_LABEL_OF = positive visible caption attachment, UNASSIGNED = explicit negative
for that pair, UNCERTAIN = unresolved. Origins separate 125 individual decisions from
142 reviewer panel-rule decisions. 57 rule-filtered and 351 endpoint-excluded pairs
are NOT_REVIEWED, not negatives. Graph degrees include all confirmed endpoints;
degree zero means no positive link in the selected source-derived queue only.
Components include positive-edge nodes, not isolated confirmed nodes.
