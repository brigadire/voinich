# Contract for a future reviewer-authorized object freeze

No freeze or object acceptance is performed by preparation/staging scripts. A future
authorized reconciliation step creates a separate immutable manifest with:

```json
{
  "status": "OBJECT_COMPLETENESS_REVIEWED_AND_FROZEN",
  "object_protocol_version": "C1.0",
  "reviewer_id": "R01",
  "freeze_timestamp": "<UTC ISO-8601 Z>",
  "prepared_reference_sha256": "<REFERENCE_OBJECTS.tsv SHA-256>",
  "confirmed_endpoints": {"path": "CONFIRMED_ENDPOINTS.tsv", "sha256": "<hash>"},
  "proposed_additions": {"path": "HUMAN_ADDITIONS_PROPOSED.tsv", "sha256": "<hash>"},
  "panel_completion": {"path": "PANEL_COMPLETION.tsv", "sha256": "<hash>"}
}
```

Paths relative to manifest directory, cannot escape it. confirmed_endpoints includes
every old REFERENCE_OBJECTS.tsv row unchanged (all REF_FIELDS exact) plus only explicitly
confirmed new proposals. New canonical IDs must belong to staged proposal table; geometry
equals proposal, reference_id=REF_<sha256(canonical_id)[:16].upper()>; layer corresponds
to STAR_REFERENCE/LABEL_REFERENCE. Additional new-endpoint columns:
post_review_decision=CONFIRMED_NEW, confirmation_reviewer_id, confirmation_timestamp,
reconciliation_status=EXPLICITLY_RECONCILED_NO_UNRESOLVED_CONFLICT. No UNCERTAIN_NEW ordinary
endpoints. All original input hashes still verified at generation/validation.

All eight completion rows have COMPLETE_* status consistent with proposal count and five
true review checks. Zero additions requires explicit COMPLETE_NO_NEW_OBJECTS, not empty
file inference. Reference-rounding from CVAT is not written back to baseline snapshot.
Prior REJECT/UNCERTAIN/NOT_REVIEWED history stays immutable; post-review reconciliation
cannot silently rewrite it. Geometry revision or new unresolved-to-confirmed review rules
require an explicit later amendment/revision, not bypassing this prepared gate.

This is a format contract, not a filled manifest. Production attachment ZIP is not
created during preparation. A manifest carrying a status string alone, without matched
checksums/markers/proposal memberships/manual confirmation, does not pass the gate.
