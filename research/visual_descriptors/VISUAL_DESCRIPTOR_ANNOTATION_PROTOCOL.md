# Blind visual annotation protocol

Status: frozen 1.0.0, `ANNOTATION_PROTOCOL_FROZEN=true`.

## Inputs and unit

Use only Yale IIIF object 2002046 and an identification table containing `page_id`, `physical_leaf_id`, canvas OID, source locator, crop locator, and schema version. The annotation unit is the exact IVTFF page/panel ID. `physical_leaf_id` strips recto/verso and panel suffixes (for example, `f67r1 -> f67`); `fRos` is grouped as `f85-f86-foldout`.

Broad IVTFF `$I` may exist in administrative metadata but must be hidden in the coding UI. Never show Currier, hand, quire, token/line counts, textual metrics, Level-B outputs, or future outcomes.

## Viewing

Start with the complete unit at fit-to-window. Zoom up to native pixels when object boundaries or damage are unclear. Do not enhance color selectively; global rotation for viewing is allowed and must not change the stored locator. Do not crop ordinary pages. For foldouts, use only a pre-adjudicated crop tied to the exact panel ID. A panel with no verified crop receives `NOT_VISIBLE` for descriptors and `PANEL_MAP_PENDING`, not a copied value from the whole spread.

Writing may be recognized only as a spatial mark class. Do not transcribe, count glyphs, identify words, compare spellings, or consult textual metadata. Titles and folio numbers added by the holding institution may be used only to identify the image.

## Coding order

Code composition, object counts, geometry, conditional figure interaction, text-image spatial relations, then complexity. Apply the exact JSON definitions. Prefer a coarse category or `UNCERTAIN` over false precision. Use `NOT_APPLICABLE`, `NOT_VISIBLE`, and `IMAGE_MISSING` only under their distinct rules.

## Damage and missingness

If the source canvas is absent, all descriptors are `IMAGE_MISSING`. If the canvas exists but the relevant unit is hidden, clipped, folded, abraded beyond recognition, or lacks a verified panel crop, code `NOT_VISIBLE`. Local damage that does not prevent a decision does not change the code; note it in adjudication records.

## Pilot and reliability gate

Select 24 units before coding, covering Herbal, circular diagrams/Zodiac, Biological, Pharmaceutical, Stars/Text, and complex panels. Obtain two independent blinded passes from different annotators/models, or from one annotator after a documented separation and fresh randomized order. Do not expose the other pass.

For nominal/binary fields report Krippendorff's alpha and Cohen's kappa where defined; for ordered fields report weighted kappa. Simple agreement is supplemental. Fixed gate:

- retain: alpha or weighted kappa >= 0.80 and agreement >= 0.85;
- clarify and re-pilot once: 0.60–0.799, or kappa undefined because of prevalence with agreement >= 0.90;
- remove/coarsen: below 0.60 or agreement below 0.75;
- otherwise adjudicate before the single permitted revision.

The single revision and complete 24-unit rerun are complete. RC descriptors that failed were removed under the fixed gate. No further definition, category, threshold, or descriptor change is permitted. Definitions cannot change until Level C ends.

## Review

Annotators first submit locked passes. A reviewer compares disagreements without any textual data, records rule-based adjudications, and may apply only the predeclared one-cycle revision. Page-specific exceptions may not redefine a descriptor. The Level-C analyst receives the frozen descriptor TSV only after validator success.
