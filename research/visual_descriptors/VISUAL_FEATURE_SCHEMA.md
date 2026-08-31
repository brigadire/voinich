# VM page-specific visual feature schema

Status: **frozen 1.0.0**. `VISUAL_FEATURE_SCHEMA_FROZEN=true`.

The normative machine-readable definition is `VISUAL_FEATURE_SCHEMA.json`. It contains 15 descriptors in six families: composition (2), object counts (5), geometry (3), figure interaction (2), text-image spatial relation (1), and visual complexity (2). Every retained descriptor passed the fixed final-pilot gate. Writing is treated only as an unread spatial texture; no sign is transcribed or interpreted.

Common states have distinct meanings: `UNCERTAIN` means the image is visible but the rule does not decide; `NOT_APPLICABLE` is allowed only where the descriptor's rule says its precondition is absent; `NOT_VISIBLE` means damage, occlusion, or an unresolved crop prevents inspection; `IMAGE_MISSING` means no source canvas was available. These states must never be collapsed during annotation.

## Inclusion rationale

Composition describes how page-local visual material partitions space. Coarse object counts preserve observable differences without false precision. Geometry captures organization with comparatively little semantic interpretation. Figure interaction is conditional on visible human-like forms. Text-image descriptors use position and baseline shape only. Density, complexity, and repetition provide anchored ordinal summaries. No descriptor was proposed or retained using textual outcomes.

## Freeze record

Two zero-history clean-room annotators repeated the complete locked 24-unit pilot after the one controlled revision. Six RC descriptors that did not pass were removed without changing any retained definition. The 15 remaining descriptors are frozen and cannot change until Level C is complete. Artifact hashes are recorded in `VISUAL_DESCRIPTOR_FREEZE_SHA256SUMS`.
