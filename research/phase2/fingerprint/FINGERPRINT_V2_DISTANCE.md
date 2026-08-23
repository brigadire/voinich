# Voynich Fingerprint v2 — Representation, Distance, Weighting, Target-Blind Discipline

Status: **DESIGN, NOT FROZEN**. Covers parent task `tasks_ph2/task73.txt`
sections 21-25 and 30. Builds on
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) and
[FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md).

## 1. Multivariate representation (parent section 21)

Fingerprint v2 is **not** a single flat vector. It is the hierarchical
object:

```
F2(corpus) = {
  F_glyph,     # G1-G4
  F_token,     # TF1-TF3
  F_lexical,   # LP1-LP4
  F_edit,      # EF1-EF5
  F_sequence,  # SQ1-SQ5
  F_local,     # LL1-LL3
  F_page,      # PG1-PG4
  F_hier,      # HR1
  F_cross,     # CS1, CS3, CS5, CS7 (new) + CS2/CS4/CS6/CS8/CS9/CS10 (cross-references)
  F_compress,  # CP1-CP3
}
```

each itself a small structured record (a metric family's own per-stratum
table, not a single scalar, where the spec defines it that way — e.g.
`F_page.PG1` is a table indexed by locus type, not one number). Plus, per
[FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 6, every
corpus additionally carries a paired **heterogeneity** object
`H2(corpus)` with the same shape restricted to the metric families
whose per-stratum breakdown is meaningful, keyed by Currier A/B, hand,
quire and locus type.

This mirrors `docs/phase1/VOYNICH_FINGERPRINT_V1.md`'s own explicit
warning that "the fingerprint is the conjunction, not any single row";
v2 keeps that discipline but makes the grouping structure (which rows
belong together as one family) explicit and machine-readable instead of
implicit in prose.

**Why not flatten.** A flat vector implies every entry is commensurable
on the same footing, which is exactly what section 23 (family weighting)
warns against — a family with many correlated internal dimensions (e.g.
PG1's per-locus-type table) would silently dominate a flat Euclidean
comparison. Keeping the hierarchical shape means every downstream
comparison method below must explicitly decide how to aggregate within
and across families, rather than doing so by accident of vector length.

## 2. Distance / comparison methods (parent section 22)

No single distance is designated as *the* fingerprint distance. At least
the following are specified, to be applied together and reported
together; disagreement between them is itself informative, following the
same logic `research/phase1/mechanism-space-analyze/pareto.go` already
uses for multi-axis comparison (frozen Task66 Pareto comparison across
fingerprint targets, referenced per parent task section 22's pointer to
prior experience — "Task52" in the parent task's own numbering refers to
this class of multi-target comparison problem, which Task66 is the
in-repository instance of).

1. **Standardized per-family distance.** Within each family
   (`F_glyph`, `F_lexical`, ...), standardize each scalar/vector entry
   by its own bootstrap standard error (from
   [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 4),
   then take a family-internal Euclidean or Mahalanobis distance. This
   answers "how different are two corpora on this one family," on a
   scale that accounts for each metric's own estimation uncertainty
   rather than raw units.
2. **Family-balanced distance.** Combine the per-family standardized
   distances from (1) into one number by averaging **family-level**
   distances (one number per family) rather than metric-level distances,
   so a family with 50 correlated sub-entries and a family with 2
   independent ones contribute equally at the family level. This is the
   direct implementation of the anti-overweighting requirement in
   section 3 below.
3. **Distributional distance.** For metric families whose natural output
   is a distribution rather than a point (G1 n-grams, G2 position
   profile, SQ2 run-length distribution, EF1 degree distribution),
   compare corpora with a distributional distance appropriate to that
   family (e.g. Jensen-Shannon divergence or Earth Mover's Distance,
   declared per family at implementation time) instead of forcing the
   distribution into a scalar summary first and then Euclidean-comparing
   the summaries.
4. **Pairwise-available distance.** Not every corpus has every metric
   family available (e.g. C-NAT corpora will not have Currier/hand
   metadata, so `H2`'s stratified entries are undefined for them; some
   Phase I synthetic mechanisms may not support every cross-scale metric
   because they lack the conditioning variable, e.g. a mechanism with no
   notion of "line"). The pairwise-available distance restricts any
   comparison to the intersection of families both corpora actually
   have, and reports that intersection's coverage alongside the number,
   so a distance computed on 10% of families is never silently presented
   the same way as one computed on 90%.
5. **Common-core distance.** A declared minimal subset of families that
   every comparison corpus (Voynich, C-NAT, C-PHASE1, and any future
   generative/mnemonic candidate) is expected to support — at minimum
   `F_glyph`, `F_token`, `F_sequence` marginal-level entries, since every
   corpus has a glyph/token stream — used as the one comparison that is
   always available even when a candidate model cannot produce line/page
   structure at all (e.g. an unstructured message-free generator with no
   layout model).
6. **Pareto comparison.** As in Task66
   (`research/phase1/mechanism-space-analyze/pareto.go`,
   `experiments/mechanism-space-v1/`), report which comparison corpora
   are Pareto-dominated across the family-level distances from (2) rather
   than collapsing to one ranking number — a candidate that matches
   Voynich well on `F_lexical` but poorly on `F_page` is not
   automatically "better" or "worse" than one with the reverse profile
   until a stated purpose ranks those families, which fingerprint v2
   itself does not do (see section 4 below).

**No arbitrary single Euclidean distance.** Per the explicit parent
instruction, none of (1)-(6) is nominated as *the* answer; a future
comparison (e.g. against a Fontana-derived mnemonic model, Task74's
subject) must report at least the family-balanced distance (2) and the
Pareto comparison (6), and should report (3)-(5) wherever applicable,
rather than picking whichever single number is most favorable to a
preferred conclusion.

## 3. Family weighting (parent section 23)

Two safeguards, both already implied above, stated here explicitly as
the anti-overweighting policy:

1. **Family-level aggregation, not dimension-level.** Every distance in
   section 2 that combines information across families does so by first
   collapsing each family to one (or a small, declared, fixed number of)
   summary distance(s), never by concatenating every family's raw
   dimensions into one long vector before comparing. A family with 50
   correlated entries (e.g. a large per-stratum table) therefore
   contributes the same weight as a family with 2 independent entries.
2. **Redundancy-adjusted family weight, not equal-by-default weight.**
   Within a family whose internal entries are themselves highly
   correlated (as flagged by
   [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 5's
   redundancy analysis), the family-internal distance in section 2(1)
   uses a covariance-aware (Mahalanobis-style) combination rather than
   a naive average, so that near-duplicate internal entries do not
   inflate that one family's effective weight relative to others even
   after family-level aggregation.

Neither safeguard assigns families a substantive weight based on how
"important" a family is believed to be for any origin hypothesis — that
would violate section 4/5 below. Weighting here is purely a
dimensionality/redundancy correction, decided by the internal statistical
structure of each family, not by expected discriminating power for any
named model class.

## 4. Discrimination target (parent section 24)

Fingerprint v2's family-balanced representation and distance toolkit
above are built to compare, without presupposing which is closest to
Voynich a priori:

- Voynich itself (global and heterogeneity objects).
- natural language corpora (C-NAT: Doyle, Longfellow, Astafiev, and any
  future addition under `DATA.md`'s discipline).
- transformed plaintext (C-PHASE1's homophony and other Task54-58 forward
  transforms, and any future transform).
- formal/artificial symbolic systems (any future constructed-language or
  formal-grammar corpus, not yet in-repository).
- message-free generators (C-PHASE1's M0-M11 mechanism-grid candidates
  and copy/mutate generators, and any future generator).
- mnemonic/external-memory models (the subject of `tasks_ph2/task74.txt`,
  not yet built at the time this document is written).

No metric family, control, weighting rule, or distance method above was
chosen because it favors one of these classes. This is checkable
directly: every family's Justification field in
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) names a structural
property and a redundancy/estimability/stability argument, never a named
model class as the reason for inclusion.

## 5. Target-blind future use and freeze discipline (parent section 25)

Once fingerprint v2 reaches the frozen state defined in
[FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md)'s freeze gate, the
following are binding for every subsequent comparison, most importantly
the future Fontana/mnemonic-model comparison that `tasks_ph2/task74.txt`
is building toward:

1. **No metric addition on model fit.** A metric family is never added to
   the frozen fingerprint because a candidate model (Fontana-derived or
   otherwise) reproduces it well. New metric families may still be
   proposed for a future fingerprint **version** (v3), but not folded
   into the frozen v2 comparison retroactively.
2. **No metric removal on model failure.** A metric family already in
   the frozen fingerprint is never dropped because a candidate model
   fails to reproduce it. A documented estimator bug (the same standard
   Phase I already applied — see the corrected P1-C026/P1-C027/P1-C028
   cases in `docs/phase1/PHASE1_CLAIMS.tsv`) is the only valid reason to
   revise a frozen metric, and any such revision must be logged the same
   way Phase I logged its corrections.
3. **No weight changes after seeing results.** The family weighting in
   section 3 is fixed at freeze time, before any candidate-model
   comparison is run, and is not retuned after a comparison's results are
   known, even if a different weighting would have produced a more
   striking or a more convenient outcome.
4. **Distance-method reporting is fixed, not cherry-picked.** Every
   future comparison reports the full toolkit from section 2 that
   applies to the corpora being compared (per the pairwise-availability
   rule), not a single distance selected after seeing which one looks
   best for a preferred conclusion.

These rules are a precondition for freeze
([FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) section on the freeze
gate), not merely aspirational: the freeze gate treats "no
target-fitting occurred during design" (section 6 below) and "the above
four rules are written down before any Fontana-derived corpus exists" as
two separate, both-required conditions.

## 6. Fontana independence statement (parent section 30)

This document, [FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) and
[FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) were written
without reference to any Fontana-machine design detail. The only
Fontana-related input used anywhere in Task73 is the general framing
already stated in the parent task itself (section 30): fingerprint v2
should measure hierarchy, locality, and cross-scale structure. That
framing is satisfied by sections 9-14 of
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) (sequence, local/line,
page/2D, hierarchy, cross-scale) for reasons independent of Fontana:
those are exactly the levels
[FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv) grades
`PARTIALLY_COVERED` or `WEAKLY_COVERED` in Phase I, which is sufficient
justification on its own.

No metric family above references, was tuned against, or was selected
because of any specific mechanism from Secretum de thesauro
experimentorum ymaginationis hominum or any other Fontana work.
`tasks_ph2/task74.txt` (the Fontana source study) had not produced any
output that this document consulted at the time Task73 was written; the
sequencing itself (task73 frozen before any Fontana-derived corpus
exists) is the intended enforcement mechanism for section 25's rules
above, not merely a scheduling coincidence.
