# Deterministic attachment-v3 queue spec

For f68r1/f68r2 reuse the previously approved *screening* right-band rule, not an AI
relation proposal: label bbox center_x > star center_x, and rotated/ellipse envelope
y-ranges overlap (touching included). No nearest-neighbor assignment or semantic filter.
Distance is recorded, not used as a truth criterion; no arbitrary k=1 truncation.
f68r3, once explicitly limited by frozen mask/whitelist, uses envelope y-overlap inside
that scope; all other panels are not iterated. No non-applicable Cartesian products.

Existing explicit v2 endpoint-pair decisions are retained, not reviewed/converted; history
appears only in RETAINED_PRIOR_V2.tsv, never UI. Eligible endpoint pairs not passing the
window appear as FILTERED_PAIR_NOT_A_HUMAN_OBSERVATION, not negative or decision rows.
Absent source-proposed pairs passing the window may enter the completeness queue.

Canonical HC_REL_<hash> uses protocol 3 + LABEL ID + STAR ID; sort panel/pair ID.
Each queue row records inclusion_reason, center distance, exact endpoint geometries,
eligibility/version and initial UNREVIEWED. Crop encloses both oriented endpoint envelopes
+80 px context clipped to canonical image; crop does not rescale coordinates, only
subtracts an integer offset that validator adds back. No new geometry types.

Production generation requires a checksum-verified external object freeze with completed
8-panel markers, exact existing reference snapshot, staged proposal membership and
explicit reviewer-confirmed/reconciled new endpoints. Gate rejects proposal-only,
UNCERTAIN_NEW, changed baseline/foreign IDs. Schema and dry run exist now; production
generation is explicitly blocked until that contract is satisfied.
