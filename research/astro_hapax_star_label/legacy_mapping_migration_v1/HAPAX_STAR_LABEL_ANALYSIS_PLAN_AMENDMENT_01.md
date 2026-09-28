# Amendment 01: legacy astronomical mapping migration

Version: `HSL-1.0-AMENDMENT-01`. Date frozen: `2026-09-15`.

This amendment supplements, and does not replace, the frozen
`../HAPAX_STAR_LABEL_ANALYSIS_PLAN.md`. It corrects the preparation-stage premise
that no prior LABEL/transcription mapping existed. The Stolfi inventory mapping
is reused only where its row-level occurrence provenance can be carried to one
canonical augmented LABEL without a panel, geometry, or multiplicity conflict.
No review of all 92 LABEL and no lexicon search is authorized here.

## Cohort and crosswalk fixed before enrichment

The target universe is the 92 frozen canonical LABEL on `f68r1`, `f68r2`, and
`f68r3`. A primary migrated row must satisfy all of the following:

1. the legacy source row is `MATCHED` and names an exact frozen ZL3b locus and
   occurrence position;
2. panel identity is exact;
3. the physical LABEL is connected to the canonical LABEL by a unique frozen
   object/provenance path, or by a unique panel-constrained geometric match;
4. the frozen occurrence bytes and token sequence agree with the legacy row;
5. the result is one-to-one at LABEL level, except that multiple transcriber
   variants or multiple tokens belonging to the same physical run remain one
   LABEL observation.

For `f68r1` only, the frozen first-pass anonymous labels provide a spatial
bridge. `f68r1-label-anon-NNN` was created in the same physical-label sequence
as ZL3b `@Ls` loci `f68r1.(NNN+7)`; this ordinal bridge must be checked against
the complete 29-crop visual sheet. It is admissible only when the canonical
LABEL retains the exact `A:f68r1-label-anon-NNN` source annotation. A nearby
canonical box without that direct source ID is not auto-selected.

The optional geometric class is predeclared as: same panel, unique maximum
axis-aligned IoU >= 0.50, normalized centre distance <= 0.025 of the panel
diagonal, and a centre-distance margin >= 0.010 over the runner-up. Both
conditions and the uniqueness margin are mandatory. Geometry is diagnostic,
not a substitute for the f68r1 ordinal/provenance path. No current row is
promoted merely to meet coverage.

Allowed primary result classes are `EXACT_LEGACY_CROSSWALK`,
`UNIQUE_PROVENANCE_CROSSWALK`, and `UNIQUE_GEOMETRIC_CROSSWALK` under these
rules. Many-to-many, cross-panel, token-only, unmatched, sensitivity-only,
human-added-without-legacy-correspondence, and ambiguous rows are excluded.

## Missingness, selection, and bounds

Coverage is reported overall and separately by panel, grouped status,
human-added status, geometry/form, size, group size, and available legacy
family. Wilson 95% intervals are used for coverage proportions. The mapped and
unmapped composition audit uses only pre-hapax features. The prior 54 unmatched
coordinates and their 77 sensitivity-only potential occurrences remain bounds
inputs; they are not reconstructed LABEL or primary mappings.

If enrichment becomes eligible, label-level missingness bounds must include
all-zero, all-one, and adversarial grouped/ungrouped assignments without
inventing token counts. Multi-token LABEL remain single dependence blocks, and
the eight-member 3G1 hyperedge remains one group observation.

## Enrichment gate

`HAPAX_ENRICHMENT_RUN_AUTHORIZED=YES` requires all of the following:

- exact reproduction of the five legacy control counts;
- valid upstream checksums and 92 unique frozen canonical LABEL;
- every primary mapping passes the frozen rules above;
- at least one mapped grouped and one mapped ungrouped LABEL in at least two
  panels each, so the stated grouped-versus-ungrouped contrast is identified
  beyond a single panel;
- the sign of the label-level grouped-minus-ungrouped effect is invariant under
  the predeclared adversarial missingness bounds.

Otherwise `HAPAX_ENRICHMENT_RUN_AUTHORIZED=NO`; no enrichment artifact is
created. The gate is deliberately evaluated before corpus-wide hapax statuses
are joined. The primary unit, if authorized, remains the frozen plan's token
occurrence with LABEL dependence blocks and panel strata; label-level
`CONTAINS_HAPAX` is reported separately.

## Lexicon prohibition

Historical Latin/Arabic dictionary search is outside this migration.
`LEXICON_MATCH_RUN_AUTHORIZED=NO` unless a later, separately frozen decision
changes it after a provenance-complete targeted cohort exists.
