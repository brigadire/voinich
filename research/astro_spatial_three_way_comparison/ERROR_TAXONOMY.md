# Error taxonomy: geometric/frozen proxies only

Flags are reproducible measurements, not manually inferred visual causes. The original
human fields do not code rejection causes, contrast, occlusion, object grouping or
ink/path thickness. Therefore merging, splitting, decorative confusion and faintness
cannot be assigned as established causes. No SPLIT/MERGE was exercised. A class-constrained
AI match does not validate the class of an actual visible mark.

| Flag | Count (combined candidate union; overlaps allowed) | Meaning |
|---|---|---|
| EXCESS_CANDIDATE_EXTENT | 13 | Candidate/final width or height >2 among confirmed |
| FROZEN_MAJOR_BBOX_OR_GRANULARITY_CONFLICT | 178 | Frozen priority_reason; possible geometry/granularity issue, not proof of split/merge |
| HUMAN_REJECTED_CANDIDATE_NO_CAUSE_CODE | 206 | Reviewer rejected proposed candidate; no cause code |
| INSUFFICIENT_CANDIDATE_EXTENT | 29 | Candidate/final width or height <0.5 among confirmed |
| LARGE_CENTER_CORRECTION | 10 | Candidate center shift >0.5 target shape diagonal among confirmed |
| LARGE_ORIENTATION_CORRECTION | 39 | Axial angle error >30 degrees among confirmed |
| LOW_SHAPE_OVERLAP_AFTER_REVIEW | 57 | Candidate/final polygon IoU <0.25 among confirmed |
| LOW_SOURCE_CONFIDENCE_PROXY_NOT_OBSERVED_CONTRAST | 43 | LOW/AMBIGUOUS source confidence, not measured image contrast |
| RING_PATH_ENVELOPE_ONLY | 9 | Filled ellipse envelope; ink/ring thickness not available |
| SCAN_RIGHT_BOUNDARY_OWNERSHIP_NOT_INFERRED | 3 | f67v1 scan boundary warning, not manuscript ownership |
| UNRESOLVED_VISUAL_CANDIDATE | 8 | Human UNCERTAIN, not a negative |

Source: ERROR_FLAGS.tsv. Deterministic case selection: figures/CASE_STUDIES.tsv and prospective plan. Partial edge stars may be valid visible objects; f67v1 adjacent-page strip ownership is not inferred from acceptance.
