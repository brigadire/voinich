# Voynich Fingerprint v2 — Controls, Stability, Redundancy, Metadata Stratification

Status: **DESIGN, NOT FROZEN**. Covers parent task `tasks_ph2/task73.txt`
sections 16-20. task77 implements its own cross-scale null registry under
the N1-N8 naming task77 itself specifies (`FINGERPRINT_V2_NULL_REGISTRY.md`);
those are not renamed C-IDs from this table, though several are close
analogues (N2≈C-LINE, N4≈C-PAGE, N5≈C-REGIME, N7=C-GRAMMAR) — see that
document's own table for the exact correspondence and task77-specific
constructions (e.g. N5/N8 are folio-level and family-label permutations,
not literal within-cluster token shuffles). Builds on
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md), whose metric families
reference the control IDs defined here.

## 1. Control-family catalog (parent section 19)

Each control below states what it **preserves** and what it **destroys**,
per parent task section 20's control-leakage requirement. A control is
invalid for a given metric if it destroys more structural levels than the
one the metric targets, because that makes the result ambiguous between
explanations (the failure mode Task59's original homophony control had,
per `docs/phase1/PHASE1_CLAIMS.tsv` claim P1-C027, before it was
corrected).

| ID | Construction | Preserves | Destroys | Primary use |
|---|---|---|---|---|
| C-GLOBAL | shuffle tokens uniformly across the whole corpus | corpus-wide token frequencies | all order: adjacency, line membership, page membership, regime membership | baseline "no order at all" reference for MI/adjacency/repetition metrics |
| C-LINE | shuffle tokens within each physical line only | line membership, per-line token multiset, corpus-wide frequencies | within-line order; does not touch cross-line structure | isolates whether an effect requires order *within* a line (used by Task58/60/63 already) |
| C-PAGE | shuffle tokens/lines within each folio only | page (folio) membership, per-page multiset | line-level and within-line order; cross-page structure untouched | isolates whether an effect requires structure *within* a page but not necessarily within a line (Task64's scale-comparison family) |
| C-LEN | resample/re-pair tokens matched on token length | length distribution | frequency-driven and identity-driven structure beyond length | separates "this happens because these things share a length" from anything else |
| C-FREQ | resample/re-pair tokens matched on corpus frequency | frequency distribution | length- and identity-driven structure beyond frequency | separates frequency-driven opportunity from real structure (Task60's existing matched-null) |
| C-POS | shuffle preserving each unit's positional marginal (e.g. glyph position-within-token marginal) but destroying the joint arrangement | positional marginal distribution | joint position x identity structure | isolates whether a joint effect is more than "this position's marginal implies this" |
| C-REGIME | shuffle restricted to within a single Task65 regime-cluster at a time | regime-cluster membership, within-regime marginals | cross-regime order; any structure that depends on being able to cross a regime boundary | isolates whether an effect requires being inside one regime versus being able to compare across regimes |
| C-WITHINTOKEN | shuffle glyphs within each token's boundaries only | token boundaries, token length, per-token glyph multiset | within-token glyph order/position | isolates within-token positional structure (Task59/61's existing convention) |
| C-GRAMMAR | generate independent synthetic tokens matching the corpus's length distribution and G2 positional-glyph marginals, with no joint/lexical/paradigm structure beyond those two marginals | length distribution, positional-glyph marginals | any joint, paradigm, family, or higher-order structure | tests whether an edit-family/lexical-paradigm finding is "just" a consequence of a bounded alphabet and length cap (parent section 8's core question; used by EF1-EF4, LP1-LP2, CS1) |
| C-NAT | Doyle, Longfellow (English), Astafiev (Russian) natural-language corpora, prepared exactly as in Phase I | natural prose structure at every level | nothing of Voynich's own structure (independent corpus) | scale reference: "what does a value like this look like in ordinary prose" |
| C-PHASE1 | reuse Phase I's frozen synthetic mechanisms: homophony series (H2/H4/H6/H8), copy/mutate and structured-token generators, M0-M11 mechanism-grid candidates | whatever each frozen mechanism was designed to preserve (documented in its own Phase I report) | whatever that mechanism does not model | scale reference against specific named alternative mechanisms already validated in Phase I; never re-derived, only reused |
| C-POS-BOUNDARY | randomize regime-change-point or other boundary positions while preserving their count and the underlying per-position value sequence | number of boundaries, marginal value sequence | the specific alignment between boundary position and another labeled sequence (e.g. locus-type changes) | isolates boundary-alignment tests (LL2) from "boundaries exist" to "boundaries are where I claim they are" |

## 2. Control assignment is metric-specific, not global

[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) assigns one or more of
the IDs above to every metric family at the point of definition. This
document does not repeat that assignment table; it defines what each ID
means and how leakage is checked (section 3 below). A metric is not
considered controlled until its assigned null's preserved/destroyed row
above is confirmed to target exactly the property under test — this
check is part of each family's Definition of Done at implementation time
(tracked per family in
[FINGERPRINT_V2_IMPLEMENTATION.tsv](FINGERPRINT_V2_IMPLEMENTATION.tsv)).

## 3. Control leakage discipline (parent section 20)

A control leaks when it destroys more than one structural level at once,
making a rejection ambiguous between "the targeted property is real" and
"some other, unintentionally destroyed property explains the result."
Before any control is used for a metric in
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md), it must pass this
three-part check, recorded once per (metric family, control ID) pair at
implementation time:

1. **Target match.** The property the metric claims to measure is listed
   under the control's "Destroys" column above; if it is not, the
   control cannot support that claim.
2. **Nuisance preservation.** Every other property that the metric's
   interpretation depends on holding constant (declared explicitly in
   the metric's own "known limitations" field in the spec) is listed
   under "Preserves."
3. **No compound confound.** If a metric's own known limitations already
   flag a specific alternative explanation (for example, PG1's low
   label-stratum power, or LL2's "locus-type changes are a coarse
   cataloguer partition"), the assigned control must not itself be the
   source of that same confound.

The single Phase I precedent for a leakage failure is exactly the
corrected case in `docs/phase1/PHASE1_CLAIMS.tsv` (P1-C027): the original
Task59 negative-control generator leaked within-token position, which
meant a "position-independent homophony" claim was actually being
compared against a control that was not, in fact, position-independent.
Fingerprint v2's leakage check exists specifically to catch that class of
error before, not after, a metric ships.

**Known compound-risk pairs identified at design time**, to be resolved
explicitly during implementation rather than left implicit:

- C-GRAMMAR simultaneously destroys lexical-paradigm structure (LP) and
  edit-family graph structure (EF) because both are downstream of the
  same "no joint structure beyond two marginals" construction. This is
  intentional (EF4 explicitly compares LP2 and EF1-EF3 against the same
  C-GRAMMAR baseline) but must never be read as two independent lines of
  evidence — it is one null applied twice, and the two resulting
  verdicts are correlated, not independent confirmations.
- C-REGIME and C-PAGE both operate at a similar spatial grain (a Task65
  regime cluster typically spans less than one page). Any metric using
  both (e.g. EF5) must report whether its result changes when the two
  are swapped, since a real page-boundary effect and a real
  regime-boundary effect can otherwise be mistaken for each other.

## 4. Stability requirements (parent section 17)

Every candidate metric family in
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) must pass the following
minimum battery before being eligible for the frozen fingerprint. A
metric that fails is **not** automatically excluded from the design, but
it does not enter the frozen fingerprint on its current definition — its
status is recorded as `UNSTABLE` with the failing check named, and it is
either narrowed (e.g. reported only for large-enough strata) or deferred.

1. **Bootstrap.** Resample the metric's natural unit (token, type, pair,
   or family, matching the unit each metric's own "Uncertainty" field
   already declares) with replacement; report the resulting interval
   width relative to the point estimate.
2. **Block bootstrap.** Where a metric's estimator assumes independence
   across a sequence (sequence-, line-, and page-level metrics), resample
   contiguous folio-sized blocks instead of individual units, since
   adjacent units in this corpus are known to be dependent
   (`docs/phase1/PHASE1_RESEARCH_REPORT.md` section 8).
3. **Folio subsampling.** Recompute the metric on repeated random
   subsets of folios (e.g. 80% of folios, repeated) and report the
   spread across subsets, independent of the bootstrap interval, to
   catch instability driven by a small number of unusual folios.
4. **Section/hand sensitivity.** Recompute separately for Currier A vs.
   B and, where sample size allows, by hand; report whether the metric's
   qualitative conclusion (not just its point estimate) changes. This
   check is what feeds the global-vs-heterogeneity split in section 6
   below.
5. **Transcription sensitivity, where possible.** Phase I's own
   limitation is that only one transcription family (`ZL3b-x7` /
   `ZL3b-n`) is available in-repository
   (`docs/phase1/PHASE1_RESEARCH_REPORT.md` section 2). Fingerprint v2
   cannot manufacture a second transcription; this check is therefore
   marked `DATA_NOT_AVAILABLE` for any metric until or unless a second
   transcription is acquired under `DATA.md`'s existing external-data
   discipline, and that limitation must be carried into every affected
   metric's reported uncertainty rather than silently omitted.

**Instability criterion.** A metric family is classified `UNSTABLE` if
either (a) its bootstrap interval spans a range that would flip its
qualitative interpretation (e.g. from "exceeds C-GRAMMAR" to "consistent
with C-GRAMMAR"), or (b) its folio-subsample spread shows the same
flip across plausible folio subsets. Both conditions are evaluated on the
metric's own declared interpretation boundary, not on an arbitrary
significance threshold, so that the criterion matches what the metric is
actually being used to decide.

## 5. Redundancy analysis (parent section 18)

**Procedure.** Once a batch of metric families is implemented, compute
pairwise correlation (Spearman, since several candidate metrics are
ranks/shares rather than approximately normal quantities) across the same
set of comparison corpora (Voynich plus C-NAT and C-PHASE1 controls) for
every pair of scalar-valued metrics in the fingerprint. Two metrics are
flagged as **candidate-redundant** if their correlation across this
comparison set exceeds a declared threshold (recorded, not silently
tuned, at the time redundancy analysis is first run).

**Resolution rule.**

1. If two candidate-redundant metrics measure the same property at the
   same level (this should be rare by design — see
   [FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) section 12's
   duplicate-estimator avoidance), keep the more interpretable one (fewer
   free modeling choices, more directly tied to a single named
   structural property) and record the dropped metric with a pointer to
   its replacement.
2. If two candidate-redundant metrics measure genuinely different
   properties that happen to co-vary across the available comparison
   corpora (for example, two different cross-scale metrics that both
   move together because they share an underlying regime-heterogeneity
   driver), they are **not** merged; instead this co-variation is
   recorded as a finding in its own right, since it is itself
   informative about the manuscript's structure.
3. Redundant *families* (not individual metrics) — e.g. if every member
   of the page/2D group turns out to correlate strongly with every
   member of the local-regime group — are explicitly reported at the
   family level in
   [FINGERPRINT_V2_DISTANCE.md](FINGERPRINT_V2_DISTANCE.md)'s family-
   weighting section, since family-level redundancy is what most
   directly threatens the "no family dominates by dimension count"
   requirement (parent task section 23).

**Anti-overweighting rule.** No single measured property may be
represented by more than a declared maximum number of near-duplicate
scalar metrics in the frozen fingerprint; if a metric family's own
internal specification (e.g. a per-stratum table like PG1 or TF2)
produces many correlated numbers, those numbers are treated as one
family-level vector for weighting purposes (see
[FINGERPRINT_V2_DISTANCE.md](FINGERPRINT_V2_DISTANCE.md)), not as that
many independent fingerprint dimensions.

## 6. Metadata stratification: global vs. heterogeneity fingerprint (parent section 16)

Where metadata is reliable — Currier A/B, hand, quire (all first-class
fields on `internal/metadatavalidation.TokenMetadata` per the check
performed for
[FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv)), and
label/running-text status (locus type, likewise already parsed) —
fingerprint v2 reports two distinct objects rather than one:

- **Fingerprint global** — every metric family in
  [FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md), computed pooled over
  the whole corpus, exactly as Phase I's v1 fingerprint did.
- **Heterogeneity fingerprint** — for each metadata axis with confirmed
  reliable coverage (Currier A/B, hand, quire, locus type), the same
  metrics recomputed per stratum, reported as a *difference* structure
  (per-stratum value minus pooled value, with the section/hand
  sensitivity check from section 4 above supplying the uncertainty on
  that difference), not as a second flat fingerprint vector.

This split exists because Phase I already found the manuscript to be
heterogeneous in ways only partially explained by metadata
(`docs/phase1/PHASE1_CLAIMS.tsv` P1-C018, P1-C020): collapsing
heterogeneity into the global fingerprint would hide exactly the
structure section 16 asks fingerprint v2 to preserve, while treating
every stratum as a fully separate fingerprint would make the object
unusably large and would violate the "do not maximize metric count"
constraint (parent task section 3).

**Which axes qualify.** Per the coverage audit
([FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv)), Currier
A/B, hand and locus type are `CONFIRMED AVAILABLE` machine-readable
fields. Section (in the sense of a validated codicological grouping
beyond quire) is not independently confirmed beyond what quire and
Currier/hand already encode in this transcription, and is therefore
folded into the quire/Currier/hand axes rather than added as a fifth,
partially redundant stratification axis. Folio/page is used as the
*unit of resampling* for stability (section 4 above) and as the subject
of the page/2D metrics themselves, not as an additional stratification
axis of the heterogeneity fingerprint (stratifying by every individual
folio would produce one row per folio, which is a diagnostic tool for
LL1/PG3, not a fingerprint dimension).

**Interaction with cross-scale metrics.** CS7 (lexical structure x
metadata) and G4/CS6 (entropy x regime) already test some
metadata-conditioned joint structure directly; the heterogeneity
fingerprint above is the systematic, corpus-wide counterpart applied to
every metric family, not a replacement for those specific cross-scale
tests.
