# Hapax spatial and page-wise report

## Part A — Astronomical spatial analysis

`ASTRO_HAPAX_SPATIAL_PATTERN=INCONCLUSIVE`. The eight requested Astronomical panels are present in the frozen panel registry, but no frozen artifact binds individual token/line coordinates to `INSIDE_DIAGRAM`, `RING_OR_RADIAL_TEXT`, `MARGINAL_COMMENT`, or `OUTER_TEXT_BLOCK`. Crop mappings alone cannot assign token zones. All zone cells are therefore `DATA_NOT_AVAILABLE`; no approximate visual reading or generated image was used.

## Part B — page prevalence

Global hapax are defined from the complete frozen occurrence corpus; section-local hapax are defined independently within each broad section. On the global definition, 0/227 pages have `hapax > non-hapax`; therefore 227/227 have `non-hapax >= hapax`, pattern `MINORITY`. With section-local hapax, 12/227 pages have `hapax > non-hapax`, also `MINORITY`; the conclusion changes in magnitude but not class.

The section summary gives the distribution by broad section and physical leaf. No page is treated as an independent inferential observation; counts are descriptive and physical-leaf IDs are retained. The requested page-size diagnostic is descriptive: Pearson r between total tokens and hapax fraction is −0.1983 (global) and −0.1500 (section-local). It is not used to reinterpret the result. The frozen token data contain no reliable spatial-zone labels for the Astronomical comparison.

```text
ASTRO_HAPAX_SPATIAL_PATTERN=INCONCLUSIVE
PAGES_WITH_HAPAX_GT_NONHAPAX=0/227 (global), 12/227 (section-local)
GLOBAL_HAPAX_PAGE_PATTERN=MINORITY
SECTION_LOCAL_HAPAX_PAGE_PATTERN=MINORITY
```
