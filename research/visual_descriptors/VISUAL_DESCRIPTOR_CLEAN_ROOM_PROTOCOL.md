# Visual descriptor clean-room protocol

Version: 1.0.0-rc. This protocol governs the replacement clean passes; the contaminated 0.9.0 draft is audit-only.

## Isolation boundary

Each clean annotator is started with no inherited conversation history and an explicit path allowlist. Inputs are staged under `/tmp/vd-cleanroom/common`; the repository is not an annotation input. The allowed set is:

1. Yale JPEG derivatives identified by canvas OID;
2. `page_identification.tsv` containing only page ID, physical leaf, IIIF locator, canvas label and dimensions;
3. `image_index.tsv` containing canvas OID, neutral Yale label, local derivative path and dimensions;
4. the schema and annotation protocol;
5. the locked 24-ID pilot list, or a later predeclared full/audit ID list;
6. the adjudicated panel crop registry after it is locked.

No `$I` class, Currier, hand, quire, token/line count, transcription, fingerprint, textual statistic, Level-B report, first-pass value, or outcome is staged. Annotators are forbidden to traverse the repository or another annotator's output directory. Separate output directories and zero-history worker contexts prevent cross-pass leakage. Each pass records hashes, randomized order, completion, lock status, and a no-other-input attestation before comparison.

The isolation is enforced by capability minimization and an explicit path contract rather than by deleting repository research artifacts. A pass is certifiable only if its manifest attests compliance and its output has a stable hash before comparison.

## Lock and comparison

Pilot membership is immutable. Each annotator randomizes independently and writes `page_id` plus frozen-order descriptors. The coordinator checks only structure and allowed values before both hashes are locked; values are not displayed or compared until both manifests say `locked=true`. Reliability and disagreement review then operate only on the two locked pass files, schema, protocol, and images.

The contaminated pilot draft is retained with a name that explicitly marks it unusable. It cannot be copied, majority-voted, or adjudicated into final measurement.

## Revision and final annotation

Exactly one schema revision is permitted after the initial clean pilot comparison. Both clean annotators then repeat all 24 units in newly randomized order without seeing the initial comparison. The 1.0.0 schema/protocol are frozen only after the complete rerun passes the fixed gate.

The full-corpus annotator receives only the frozen inputs and crop registry. If only one full pass is feasible, an independently randomized audit sample is selected and hashed before the full values are compared. Raw values and rule-based adjudications remain immutable.

## Prohibited operations

No textual association, joining to textual tables, classifier, permutation, correlation, descriptor selection by predictive behavior, or Level-C analysis is permitted inside or outside the clean room during this task.
