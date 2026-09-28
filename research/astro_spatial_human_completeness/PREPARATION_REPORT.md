# Human-assisted completeness preparation

Object package C1.0 ready for a future authorized reviewer pass, not started here.
Eight original canonical panels, 520 final human references (321 STAR/199 LABEL),
8 optional unresolved reference shapes and 222 preserved nonconfirmed provenance records
(206 REJECT, 8 UNCERTAIN, 8 NOT_REVIEWED). AI identities/history are excluded from CVAT.
All references use exact final human geometry, not source/candidate display boxes.

Separate object annotation XML/ZIP and image ZIP, navigation order/maps, neutral schemas,
strict reference-preserving validator and proposed-addition staging are prepared. New
objects receive deterministic canonical IDs at future ingestion, but are NOT_ACCEPTED.
No real human export imported, no annotations created by a reviewer, no new AI metrics.
Completion markers start NOT_STARTED; empty export never asserts no missed objects.

Attachment protocol 3 is a future gated workflow, distinct from frozen spatial 1/caption 2.
Dry run only f68r1/f68r2: 118 queued UNREVIEWED pairs, 2519 geometric filters
(not observations), 48 previous v2 pairs retained without re-review/conversion.
No non-applicable-panel Cartesian product. f68r3 limited scope cannot be reconstructed
from old whole-source-pair policy: map/catalog/options prepared, reviewer confirmation required.
Older f68v2/v2 decisions stay unchanged despite new NOT_APPLICABLE policy.

CVAT instance URL/version is unavailable (no local CVAT container/connector). Documented
UI lock/hide is optional convenience; export validator is authoritative protection.
CVAT_ROUNDTRIP_TEST denotes tested native XML copied-tree/two-decimal serialization,
not a claimed live-server test. Empty live-instance import/export preflight is required
before human review; see CVAT_SCHEMA.md and guides. No invented persistent lock field.

Validation and separate-directory byte reproduction: VALIDATION_REPORT.md. Input hashes
registered in INPUT_MANIFEST.tsv and rechecked after generation. File checksums include
only this new preparation package, never refreeze original inputs. All outputs are separate.

```text
HUMAN_OBJECT_COMPLETENESS_PACKAGE=READY
HUMAN_REVIEW_STARTED=NO
FROZEN_INPUTS_UNCHANGED=YES
REFERENCE_OBJECTS_PROTECTED=YES
AI_SOURCE_IDENTITY_HIDDEN=YES
ALL_PANELS_OBJECT_SCOPE=YES
ATTACHMENT_SCOPE_F68R1=READY
ATTACHMENT_SCOPE_F68R2=READY
ATTACHMENT_SCOPE_F68R3=AWAITING_REVIEWER_CONFIRMATION
NON_APPLICABLE_PANELS_EXCLUDED=YES
PRODUCTION_ATTACHMENT_PACKAGE_CREATED=NO
CVAT_ROUNDTRIP_TEST=PASS
RESULTS_REPRODUCIBLE=YES
```
