# Task73 report — Voynich Fingerprint v2 Coverage and Design

This is a DESIGN + COVERAGE report. It specifies fingerprint v2; it does
not compute any new number, does not run any experiment, does not test
the Fontana hypothesis, does not build a mnemonic generator, does not
select a Voynich Manuscript origin mechanism, and does not decipher
anything. Every claim below about *Phase I* is sourced from
`docs/phase1/*` and `docs/methods/METHOD_INDEX.md`; every claim about
*current data availability* was checked directly against
`data/ZL3b-n.txt` and the relevant `internal/` packages during this task,
not assumed.

Outputs produced by Task73:

- [FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv) (task73a)
- [FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) (task73b, task73c, task73f)
- [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) (task73d)
- [FINGERPRINT_V2_DISTANCE.md](FINGERPRINT_V2_DISTANCE.md) (task73e)
- [FINGERPRINT_V2_IMPLEMENTATION.tsv](FINGERPRINT_V2_IMPLEMENTATION.tsv) (task73f)
- [FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) (task73f)
- this report (task73f)

## 1. How complete is Fingerprint v1?

Uneven, in a specific and now more precisely mapped way. Re-graded across
the 12 finer levels the parent task requires (rather than v1's own
coarser groupings), Phase I is strong (`WELL_COVERED`-equivalent) at
glyph, token marginals, repetition, and line/local scale; adequate but
partial (`PARTIALLY_COVERED`-equivalent) at within-token formation,
adjacent sequence, longer sequence, regime and manuscript-global; and
weak or absent at lexical paradigms, page/2D, and cross-scale — see
[FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv) for the full
row-by-row evidence and artifact citations. Confidence in the well-
covered rows is high; confidence in the weak rows is high about *what is
missing*, precisely because Phase I's own reports (`FINGERPRINT_COVERAGE.md`,
`PHASE1_RESEARCH_REPORT.md` section 14) already named these gaps — this
audit's contribution is making them into an operational, per-level table
rather than a prose list.

## 2. What major gaps remain?

In order of how directly the parent task names them: (1) lexical
paradigms — no existing metric distinguishes a productive transformation
system from accidental edit-distance proximity; (2) cross-scale
dependencies — every existing cross-level statement in the Phase I
report is a narrative conjunction of separately-computed results, never
a jointly estimated statistic; (3) page/2D structure — page has only
been used as a 1-D scale reference, never as a genuine two-dimensional
layout space, despite rich unused IVTFF metadata; (4) hierarchy — no
single model spans glyph through manuscript; (5) edit-family graph
geometry — the giant near-repeat family has never been characterized as
a graph, so the specific question "is it just a consequence of bounded
grammar" has never been tested; (6) compression/algorithmic measures —
entropy has been measured probabilistically but never cross-checked
against a compressor.

## 3. Which gaps are actually discriminating?

Lexical paradigms and cross-scale dependencies carry the highest
discrimination value: both bear directly on distinguishing H_L/H_C/H_G
(`docs/phase1/HYPOTHESIS_STATUS.md`) because a demonstrated productive
paradigm system would strengthen a codebook/morphological reading, while
its absence relative to a bounded-grammar null (EF4/LP2) would strengthen
a generative reading. Page/2D and hierarchy are also high-value because
they are true blind spots, not merely under-measured directions — no
prior Phase I result can currently be wrong or right about them. Glyph
refinement and pure marginal extensions are lower-discrimination because
Phase I already measured the relevant marginals well; further precision
there mostly sharpens existing conclusions rather than opening new ones.
See [FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) section 1 for the
full value/cost/risk table.

## 4. What new metric families are needed?

37 are specified in
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md): 4 glyph (G1-G4), 3
token-formation (TF1-TF3), 4 lexical-paradigm (LP1-LP4), 5 edit-family
(EF1-EF5), 5 sequence (SQ1-SQ5), 3 local/line (LL1-LL3), 4 page/2D
(PG1-PG4), 1 hierarchy template (HR1), 4 cross-scale-only estimators
(CS1, CS3, CS5, CS7 — CS3 aliases TF2), and 3 compression (CP1-CP3).
6 further cross-scale pairs (CS2, CS4, CS6, CS8, CS9, CS10) are
explicitly satisfied by cross-referencing an existing family rather than
adding a new estimator. This count was deliberately not maximized: every
family's Justification field in the spec states what would be lost by
omitting it, and several candidates explicitly named in the parent task
(a separate glyph-level prefix/suffix family, a separate
family-conditioned-formation family) were folded into an existing family
instead of added as new ones (spec sections 2 and 3, "Considered and not
included" notes).

## 5. Which existing metrics are redundant?

None of Phase I's *existing* metrics were found redundant with each
other during this audit — v1's rows each measure a distinct declared
property. The redundancy risk instead arises **within v2's new
specification**, where several requested bullets from the parent task
turned out to be the same estimator applied under different names:
start/end asymmetry is a special case of the full position curve (G2);
family-conditioned and regime-conditioned formation are conditioning
axes of one stratified estimator (TF2), not separate models; edit-family
locality (parent section 8) and lexical-family locality (parent section
7) are the same co-occurrence computation (EF5/LP3); and six of ten
named cross-scale examples reduce to families already defined elsewhere
(spec section 10's table). [FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md)
section 12 records this explicitly rather than silently building
duplicate estimators. A formal, correlation-based redundancy check across
*implemented* metrics is specified but not yet run
([FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 5),
since it requires implemented data.

## 6. Which 2D/page metrics are actually available?

More than v1's `WEAKLY_COVERED` grade suggested, once checked directly.
`data/ZL3b-n.txt` and `internal/metadatavalidation.ParseIVTFF` already
provide, per token: folio, quire, page-in-quire, foliation, bifolio,
hand, Currier language, illustration/content-type code (`$I`), text-
block-position code (`$X`), locus type (paragraph/label/radial/circular),
line ID, and ordinal position within the line and within the folio. Four
metric families (PG1-PG4) are specified to use exactly this data:
locus-type stratification (label vs. running text), intra-line ordinal
position as a horizontal-position proxy, page-boundary alignment, and
content-type-code stratification. What remains genuinely
`NOT_AVAILABLE`, confirmed rather than assumed, is anything requiring
pixel/vector coordinates or image-region content — no metric family is
proposed for these, matching the parent task's explicit prohibition on
manual subjective image coding (section 12).

## 7. What lexical-paradigm metrics are needed?

Four, ordered as a dependency chain in
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) section 4: LP1 censuses
transformation rules across all edit-distance-1/2 vocabulary pairs by
distinct-instance support (not raw occurrence, to avoid one frequent
pair dominating); LP2 is the load-bearing statistical test — it compares
LP1's rule-support concentration against a purpose-built grammar-bounded
null (C-GRAMMAR) to decide whether apparent paradigms exceed what a
small alphabet and length cap alone would produce; LP3 measures
branching/depth/overlap/locality only on the subgraph LP2 validates as
productive; LP4 tests core/affix joint attachment as an explicit
`P(core,affix)` vs. independence comparison, directly implementing the
parent task's own worked marginal-vs-joint example (section 4) for the
specific case of Voynichese affix-like chunks.

## 8. What cross-scale dependencies are needed?

Ten pairs are specified (spec section 10's table): four with new
estimators (CS1 glyph x lexical family, CS5 repetition x page position,
CS7 lexical structure x metadata, and CS3 which aliases TF2's line-
position conditioning), and six that alias families already defined for
other reasons (CS2/CS8 sequence x family = SQ4; CS4 edit-family x
regime = EF5; CS6 entropy x regime = G4; CS9 hierarchy-wide attribution =
HR1; CS10 page x regime boundary alignment = LL2). The parent task's own
example list (section 14) was explicitly extended, not merely
reproduced, and the extension (CS1, CS5, CS7, plus HR1 as a genuinely
hierarchy-wide cross-scale method) was chosen to close gaps the coverage
audit found, not to hit a target count.

## 9. How should v2 compare corpora?

Not with one arbitrary distance.
[FINGERPRINT_V2_DISTANCE.md](FINGERPRINT_V2_DISTANCE.md) specifies six
non-exclusive methods to be reported together: standardized per-family
distance, family-balanced distance (family-level, not dimension-level,
aggregation), distributional distance for families whose natural output
is a distribution, pairwise-available distance (restricted to families
both corpora actually support, with coverage reported), common-core
distance (the minimal always-available subset), and Pareto comparison
(reusing Task66's existing infrastructure in
`research/phase1/mechanism-space-analyze/pareto.go`). Fingerprint v2
itself is represented hierarchically (`F2(corpus)`, ten grouped
sub-objects) rather than as one flat vector, plus a paired heterogeneity
object (`H2(corpus)`) for the metadata axes confirmed reliable
(Currier A/B, hand, quire, locus type).

## 10. How is metric-family overweighting avoided?

Two rules, both in
[FINGERPRINT_V2_DISTANCE.md](FINGERPRINT_V2_DISTANCE.md) section 3: every
cross-family distance aggregates at the family level (one number, or a
small declared set, per family) before combining across families, so a
family with fifty correlated internal entries counts the same as a
family with two independent ones; and within a family whose internal
entries are flagged mutually correlated by the redundancy analysis, the
family-internal distance uses a covariance-aware combination instead of
a naive average, so near-duplicate entries cannot inflate that family's
effective weight even after family-level aggregation. Neither rule
assigns weight based on a family's presumed importance to any origin
hypothesis.

## 11. What controls are needed?

Eleven control families are cataloged in
[FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 1
(global/line/page shuffle; length-, frequency- and position-matching;
regime-conditioned shuffle; within-token shuffle; the new grammar-
bounded null C-GRAMMAR needed specifically for the lexical-paradigm and
edit-family productivity questions; natural-language and Phase I
generated-mechanism controls; and a boundary-position-randomization
control for alignment tests), each with an explicit preserved/destroyed
statement. Every metric family in the spec is assigned one or more of
these at the point of definition, and a three-part leakage check
(target match, nuisance preservation, no compound confound) is required
before any control is used in practice, modeled directly on the one
documented Phase I leakage failure (Task59's original position-leaking
control, `docs/phase1/PHASE1_CLAIMS.tsv` P1-C027).

## 12. Which metrics are expected to be unstable?

None are pre-judged unstable — stability is an empirical outcome, not a
design-time guess — but several are flagged in
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md) as likely low-power or
sample-limited and therefore worth watching first: G1's trigram tier
(Task61's own h4 coverage was only 0.116); PG1's label and radial/
circular strata (Task60 already found only 170 label pairs); SQ5's
sweep at n=4-5; and any Currier-B/rare-hand stratum in the heterogeneity
fingerprint, simply because those partitions are small. The minimum
stability battery that will determine actual status —bootstrap, block
bootstrap, folio subsampling, section/hand sensitivity, and a
transcription-sensitivity check currently blocked by data availability
— is defined in
[FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 4, with
an explicit criterion: a metric is `UNSTABLE` if its bootstrap interval
or folio-subsample spread could flip its qualitative interpretation.

## 13. What must be implemented before freeze?

Per [FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) section 3, at
minimum every family in the lexical-paradigms, edit-family-geometry,
cross-scale, and page/2D-stratification groups — the four the parent
task names as primary gaps — must be implemented, control-leakage-
checked, stability-checked, and included in a redundancy pass, and the
family-balanced distance plus Pareto comparison must run end-to-end at
least once against existing comparison corpora, with every metric
following its declared missing-data semantics. A recommended build order
(shared C-GRAMMAR/`$I` infrastructure first, then lexical paradigms,
then page/2D and edit-family in parallel, then cross-scale/local-line
consolidation, then the remaining groups, compression last) is given in
[FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) section 2, as a
recommendation for future tasks, not a Task73 commitment.

## 14. What does the freeze gate look like?

Seven jointly-required conditions
([FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) section 3): required
metrics (the four primary-gap groups, at minimum) implemented; controls
validated against the leakage check; stability checked with no metric
entering frozen status without a reported stability verdict; redundancy
analysis run once across all implemented metrics with its resolution
rule applied; the distance/Pareto toolkit exercised at least once
end-to-end; missing-data semantics verified as followed (no silent
imputation or pooling); and standard repository tests passing. Partial
freezing of only the easy groups (e.g. glyph and page/2D alone) is
explicitly rejected as a substitute for the full gate, because it would
freeze a fingerprint that does not yet address the two gaps the parent
task names as primary. Task73 defines this gate and does not satisfy it;
`FINGERPRINT_V2_FROZEN` is not set by this task.

## 15. What implementation tasks come next?

As scoped in [FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) section 2:
first, the small shared C-GRAMMAR null-generator and `$I`-field-surfacing
infrastructure that several higher-value groups depend on; then lexical
paradigms (LP1-LP4), since it is both the highest-named gap and a
dependency for four other groups; then page/2D stratification (PG1-PG4)
and edit-family geometry (EF1-EF5) in parallel, since neither blocks the
other; then local/line decomposition (LL1-LL3) and the newly-unblocked
cross-scale estimators (CS5 immediately, CS1/CS7 once lexical paradigms
land); then token-formation extension, hierarchy, and sequence
extension; and compression/predictability last, consistent with its
explicitly supplementary role. None of this sequencing was chosen with
any Fontana-machine detail in mind (see below).

## Definition-of-done confirmation (parent task section 33)

1. Phase I fingerprint audited — [FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv).
2. Coverage matrix created — same file, 12 levels.
3. Gaps classified — [FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) section 1.
4. Joint/cross-scale gaps separately defined — spec sections 1 and 10.
5. Lexical-paradigm gap designed — spec section 4 (LP1-LP4).
6. 2D/page gap designed — spec section 8 (PG1-PG4), grounded in a direct
   IVTFF/code check, not assumption.
7. Candidate metrics specified — spec sections 2-11, 37 families plus
   10 cross-scale pairs.
8. Null/control strategy defined — [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) sections 1, 3.
9. Stability requirements defined — controls document section 4.
10. Redundancy strategy defined — controls document section 5.
11. Multivariate representation defined — [FINGERPRINT_V2_DISTANCE.md](FINGERPRINT_V2_DISTANCE.md) section 1.
12. Distance strategy defined — distance document section 2.
13. Family weighting defined — distance document section 3.
14. Future model classes not used for target fitting — distance document
    sections 4-6 state the discrimination target without favoring a
    class and bind freeze-time non-target-fitting rules; this report's
    own design process consulted no Fontana-machine detail (below).
15. Implementation plan created — [FINGERPRINT_V2_IMPLEMENTATION.tsv](FINGERPRINT_V2_IMPLEMENTATION.tsv), covering all 37 families plus the 6 cross-scale aliases plus the 2 declined/NOT_AVAILABLE items.
16. Freeze gate defined — [FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) section 3, not applied.
17. TASK73_REPORT.md created — this document.
18. No Fontana/Voynich-mechanism conclusions made anywhere in the
    Task73 output set: every document above specifies measurement
    machinery and evaluation discipline; none states or implies a
    conclusion about what produced the Voynich Manuscript, whether it
    encodes meaningful language, or whether any Fontana-derived
    mechanism is compatible with it. `tasks_ph2/task74.txt` (the Fontana
    source study) had not produced any output this task consulted.
