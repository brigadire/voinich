# Task77 null-model registry

Status: **preregistered, implemented**. This document mirrors
`nullModelRegistry()` in `internal/fingerprintv2/nullmodels.go` (the
machine-readable copy is emitted as `cross_scale.null_registry` in every
`fingerprint.json`); the two must not drift, and this file exists so the
registry is readable without parsing JSON. Task77 §7 requires, for every
null: what it preserves, what it destroys, which hypothesis it tests, what
confounders remain, and why it is neither too weak nor too restrictive.

| ID | Name | Preserves | Destroys | Tests | Remaining confounders | Why this null |
|---|---|---|---|---|---|---|
| N1 | Global token shuffle | corpus-wide token multiset (frequency), length distribution | essentially all sequence/line/locus/folio/regime structure | nothing on its own — a weak baseline only | none removed beyond the global marginal | Used only where no more targeted null exists (CS6's line-length-preserving variant), never as the sole control for a cross-scale claim |
| N2 | Within-line shuffle | each line's token multiset, line length, folio/section composition | within-line order, position-within-line, local adjacency | CS1 (family x line position), CS2 (transformation x local context), CS8's per-stratum conditioning | folio/regime-level composition (deliberately: the hypothesis is about *within-line* position/adjacency, not between-line composition) | Because every permutation is drawn from the same physical line, a between-line or between-folio confound cannot manufacture a spurious CS1/CS2 result; this is verified directly by `TestCS1ConfoundedByRegime` |
| N3 | Within-position-bucket shuffle | the marginal class distribution at each normalized line-position bucket; the exact sequence of line lengths | which specific tokens/families occupy a line's slots; true line composition | CS6 (family composition x line/folio structure), realized here as a global shuffle that happens to preserve line-length boundaries by construction (see report) | position-class base rates are held fixed by construction | Targets composition/diversity *given* position, not the position marginal itself |
| N4 | Within-folio shuffle | folio vocabulary and frequency composition | line/local organization within the folio | CS3 (family x locus type) | regime-level composition (folios are mostly regime-homogeneous) | Isolates a within-folio locus-type effect from cross-folio composition differences |
| N5 | Within-regime shuffle | Currier/section partition sizes and per-regime vocabulary composition | within-regime local/line/folio placement | CS5 (local context x larger regime), the regime-stratum component of CS4 | does not itself test whether the regime label matters, only whether structure exists once regime is fixed | Direct within-partition analogue of N2 at the regime scale; implemented here as a folio-level label permutation, which additionally answers "does regime *identity* matter" (CS4) since folio-level (not occurrence-level) shuffling is the correct unit for a folio-constant attribute |
| N6 | Frequency/length-matched random reassignment | token length and (frequency-matched variant) log2-frequency bin of both endpoints | which specific same-length/frequency pair is edit-adjacent, or which structural distance a given edit distance is paired with | LP2 C-LEN/C-FREQ (inherited from task75), CS7's frequency control | residual within-bin frequency variation (bins are coarse, log2) | Directly controls the frequency confound task77 §8 names for CS7 ("high-frequency tokens by definition occur closer together") |
| N7 | C-GRAMMAR (structure-preserving / frequency-aware) | token count, exact length distribution, alphabet, positional/endpoint/bigram glyph profiles within tolerance | any lexical/paradigm structure beyond bounded token-formation constraints | EF4 (inherited), `EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL` | fairness of the C-GRAMMAR construction itself is a modeling choice (`FINGERPRINT_V2_CONTROLS.md`) | The only null that asks whether cross-scale/edit-family structure could arise purely from bounded token formation |
| N8 | Family-label permutation | family size distribution and each token's own frequency (tokens keep their identity; only which family number they carry moves) | which specific tokens belong to which family | Robustness check on CS1 (`cs1/family-line-position/n8`, reported as `additional_nulls`, not part of the primary FDR family) | graph topology within a family is not re-randomized, only membership labels | Tests whether specific families carry information beyond size/frequency alone |

## Notes on scope

- N9/N10-style "distance-two" or "multilayer" nulls were considered for
  the deferred graph representations (see `edit_graph_validation.
  graph_representations`) and explicitly not built, since task77 warns
  against combining edge-type definitions without preserving their type;
  see `TASK77_REPORT.md` §2.
- Every null above is implemented as an *empirical* permutation
  distribution (`repetitions` replicates, seeded), not a parametric
  approximation, consistent with task75's existing convention
  (`internal/fingerprintv2/stats.go:nullTest`).
- CS1's role test, CS3, CS4 (Currier/Section) and CS6 all reuse the same
  generic `nmiPermutationTest`/`permuteWithinGroups` machinery
  (`internal/fingerprintv2/nullmodels.go`) with only the grouping key
  differing (line id -> N2, folio id -> N4, constant/global -> N1); this
  is a deliberate implementation choice to keep the null constructions
  auditable from one place rather than five near-duplicate loops.
