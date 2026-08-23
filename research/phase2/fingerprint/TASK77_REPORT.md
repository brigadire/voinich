# Task77 report — edit-graph validation and cross-scale dependencies

Status: **implementation complete; canonical run against the real corpus
completed**. Unlike Task75, this task did have real corpus bytes available
locally (`data/ZL3b-n.txt` + `data_work/ZL3b-x7.canonical.txt`, gitignored
per `DATA.md`, present on this machine), so every number below is a real
run against the actual manuscript (39,380 tokens, 8,243 types, strict
IVTFF alignment), seed `20260823`, 100 grammar/null repetitions per
corpus, plus one natural-language control (Conan Doyle, `data_test/
pg2097-2.txt`, 43,713 tokens). Config: `experiments/fingerprint-v2-
cross-scale-v1/canonical.yaml`; full output: `experiments/fingerprint-v2-
cross-scale-v1/canonical-out/`.

## Zero: how to read this report

Everything here is deterministic given the config and seed
(`TestSeededPipelineIsDeterministic` still passes end to end). Every
number quoted below is read directly from `canonical-out/fingerprint.json`
or `raw_results.json`, not summarized from memory. Where a value is
`INCONCLUSIVE` or `NOT_APPLICABLE`, the reason is a specific, checkable
condition (a sample-size floor, a missing metadata field, or an empty
upstream family list), never a silent gap.

## 1. Stage-1 audit

See [TASK75_AUDIT.md](TASK75_AUDIT.md) for the full item-by-item table.
Headline: `TASK75_RESULTS_REPRODUCED = PARTIALLY_SUPPORTED`. Every LP1-LP4/
EF1-EF4 formula and null construction was confirmed correct; four
infrastructure defects were found and fixed because this was the first
time this pipeline had ever run against real corpus bytes:

1. **IVTFF alignment normalization was silently wrong.** Apostrophe,
   `?`, `@` and `;` were treated as ordinary content; the real canonical
   corpus treats all four as word breaks. Fixed with exact byte-level
   evidence recorded in code and tests.
2. **`k`-core decomposition had a peeling-order bug** (new task77 code,
   caught by `TestKCoreOnTriangleWithPendant` before it ever reached a
   real run).
3. **O(vocabulary × corpus) map rebuilds** in LP3/LP4 turned a single
   `analyzeBare` call from 0.67s into 72.7s on the real corpus. Fixed by
   hoisting `glyphByToken(c)` out of per-vocabulary-token and per-edge
   loops.
4. **O(attempts × edges) edge-list rebuild** inside `degreePreservingSwap`
   (EF2/EF3's null control) rebuilt and sorted the *entire* edge list
   after every accepted swap; with `graph_swaps × edge_count` attempts
   per replicate on a ~41k-edge real graph, this did not finish in
   bounded time. Fixed to update the two affected slots in place.

Without fixes 3-4, no canonical run in this section would have been
computationally possible at all.

## 2. Edit-graph validation (stage 2)

### 2.1 Graph representations

| Representation | Status |
|---|---|
| undirected unweighted | implemented (EF1-EF4's own graph) |
| directed transformation | implemented (LP1's rule census) |
| frequency-weighted | implemented, as a diagnostic (`frequency_weighted_hub_share`) |
| context-weighted (same-line co-occurrence) | implemented, as a diagnostic (`context_weighted_hub_share`) |
| distance-weighted | **deferred** — every edge here is edit-distance exactly one by construction; a distance-two+ census is out of task77's scope |
| multilayer | **deferred** — combining layers needs an arbitrary cross-layer weight, which task77 §2.1 explicitly warns against |

### 2.2/2.3/2.4 Transitive-merge audit, stability battery, consensus

This is where the single most important structural finding of the whole
task shows up. **LP2's productivity test did not reach significance on
the real corpus** (`productivity_state: NOT_SUPPORTED`, see §3 below),
which means the *LP2-gated* productive-rule graph that Stage 2 audits is
**empty** (0 nodes). Every transitive-merge diagnostic on that graph is
therefore a vacuous zero (`edit_graph_validation.transitive_merge`:
0 families, 0 hub-removal drop, 0 path-restriction components), not a
sign of broken infrastructure — this is the pipeline correctly refusing
to build a family catalog it has no statistical warrant for
(`consensus_status: INSUFFICIENT_SUPPORT`).

The stability battery makes this precise and surfaces a genuinely
interesting **exploratory** contrast:

| Perturbation | ARI vs. baseline (LP2-gated, empty) | Status |
|---|---|---|
| `min_rule_support` 2 | 0 | UNSTABLE |
| `min_rule_support` 4 | 0 | UNSTABLE |
| rare-token cutoff (freq≥2) | 0 | UNSTABLE |
| folio-half A vs. folio-half B (support-threshold only, **not** LP2-gated) | **0.719** (NMI 0.599) | **GLOBAL** |
| preprocessing profile (EVA vs. natural) | n/a | NOT_TESTABLE (disjoint alphabets) |
| transcription variant | n/a | NOT_TESTABLE (only one transcription available) |
| community-detection seed (label propagation) | 0 | UNSTABLE |

The "UNSTABLE" rows compare a non-empty support-threshold-only family set
against the empty LP2-gated baseline, which is mechanically bound to be
`ARI=0`; that comparison is not very informative by itself. The
**folio-half comparison is the real signal**: it compares two
*independent* support-threshold-only family sets (folio-half A rebuilt
from scratch vs. folio-half B rebuilt from scratch, neither LP2-gated) and
finds them **structurally stable** (ARI 0.72). Put together with LP1's
raw counts — 2,294 of 4,240 directed rule types have support ≥3 out of
39,380 tokens — this says something specific and non-obvious:

> **A large, internally-reproducible network of edit-distance-one rules
> exists in the real manuscript and replicates across independent halves
> of it, but this network's *aggregate concentration statistic* (LP1's
> rule-support Gini) is statistically indistinguishable from what a
> length/position/bigram-matched C-GRAMMAR generator produces on its
> own.** Reproducible local structure and "exceeds bounded-grammar chance"
> are two different claims, and only the first one holds here.

This is flagged as an **exploratory finding** (`edit_graph_validation`'s
stability battery was preregistered as a robustness check, but this
specific support-threshold-vs-LP2-gate contrast was not itself a
preregistered hypothesis) requiring a dedicated confirmatory follow-up:
a direct comparison of support-threshold-only family stability against a
C-GRAMMAR null (not just against the empty LP2-gated baseline).

`TRANSFORMATION_MOTIFS_STABLE = INCONCLUSIVE` for the same reason
(nothing to compare on an empty graph).

## 3. EF5 and the core grammar-boundedness result

EF5 (same-line/page/regime family concentration) reuses LP3's locality
computation; with zero LP2-gated families it is `NOT_APPLICABLE`
(`same_line_rate: 0`, `same_page_rate`/`same_regime_rate`: unavailable).

The more important result is **`EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL =
NOT_SUPPORTED`**, and it is worth stating precisely because it directly
answers task73's founding question ("is the giant edit family just a
consequence of a limited token grammar?") for the first time with real
data:

| EF statistic | Observed (real corpus) | C-GRAMMAR null mean (structure-preserving, 100 reps) | Effect size (SD) | One-sided p |
|---|---|---|---|---|
| EF1 giant-component share | 0.8669 | 0.9122 | **−13.1** | 1.0 |
| EF2 global clustering | 0.2873 | 0.3202 | **−11.0** | 1.0 |
| EF3 \|Spearman(degree, log-freq)\| | 0.6390 | 0.6657 | **−4.6** | 1.0 |

Only `structure-preserving` C-GRAMMAR validated (see next paragraph);
`frequency-aware` failed its positional/bigram diagnostics in all 100
replicates and is excluded from inference, exactly as task75's own
validation gate requires. Every one of the three EF statistics is *below*
its C-GRAMMAR expectation, by large, one-sided margins — not marginally
insignificant, but consistently on the wrong side of the null for a
"paradigm exceeds chance" claim. **The real manuscript's edit-neighborhood
graph is, if anything, slightly less densely connected and less clustered
than a length/position/bigram-matched random-token generator alone would
produce.** `EF4.verdict = CONSISTENT_WITH_GRAMMAR_BOUND`.

C-GRAMMAR mode adequacy on the real corpus: `structure-preserving`
preserved every exact marginal (token count, length distribution,
alphabet) in all 100 replicates and stayed within the positional/endpoint/
bigram tolerance in all 100 (fully validated). `frequency-aware`
preserved the exact marginals but breached the tolerance in all 100
replicates (never validated) — so `C_GRAMMAR_VALIDATION` is reported
`PARTIALLY_SUPPORTED` at the mode-summary level, while inference above
correctly used only the validating mode.

For contrast, LP2 also runs weaker C-LEN/C-FREQ controls (random
same-length / same-length-and-frequency type pairings): the real corpus's
support-Gini (0.7898) clearly **exceeds** both (null means 0.238/0.288,
p≈0.0099, q≈0.0116) — so the edit-neighborhood structure is far from a
naive random pairing of tokens by length/frequency alone. It is *only*
against the stronger, position/bigram-matched C-GRAMMAR null that the
result flips to "not exceeded." That contrast is itself informative: it
shows the C-GRAMMAR control is doing real work beyond the simpler C-LEN/
C-FREQ controls, not that the pipeline is insensitive.

## 4-8. Cross-scale metrics (CS1-CS8)

See [TASK77_CROSSSCALE_HYPOTHESES.md](TASK77_CROSSSCALE_HYPOTHESES.md) for
the preregistered matrix. Results on the real corpus:

| ID | Status | Observed | p / q | N |
|---|---|---|---|---|
| CS1 (family × line position) | `INCONCLUSIVE` | — | — | 0 |
| CS2 (transformation × context) | `INCONCLUSIVE` | — | — | 0 |
| CS3 (family × locus type) | `NOT_SUPPORTED` | 0 | 1.0 / 1.0 | 39,380 |
| CS4 (family × Currier) | `NOT_APPLICABLE` | — | — | 0 |
| CS4 (family × Section) | `NOT_APPLICABLE` | — | — | 0 |
| CS5 (local adjacency × regime) | `NOT_APPLICABLE` | — | — | 0 |
| CS6 (family diversity × line length) | `NOT_SUPPORTED` | 0 | 1.0 / 1.0 | 5,385 lines |
| CS7 (edit distance × structural distance) | `NOT_SUPPORTED` | 0.0343 | 0.129 / 0.386 | 2,000 sampled pairs |

CS1/CS2/CS4/CS5 are `INCONCLUSIVE`/`NOT_APPLICABLE` for the identical
reason as §2: they are defined over family-bearing occurrences, and there
are zero LP2-gated families this run. This is not a defect in these
metrics — it is the direct, correct downstream consequence of §3's
result, and it means task77's family-conditioned cross-scale hypotheses
were **not actually tested** on this corpus at this threshold; they are
reported `INCONCLUSIVE`/`NOT_APPLICABLE`, not misreported as
`NOT_SUPPORTED`.

CS3 and CS6, which don't depend on family membership having non-trivial
variation, ran to a real, substantive `NOT_SUPPORTED`: locus type (label
vs. text vs. special) shows no detectable association with the
(degenerate, all-`NONE`) family variable, and line-length does not
correlate with family-composition entropy beyond a line-length-preserving
global shuffle.

CS7, independent of family gating (it samples 2,000 vocabulary-type
pairs directly), is the one cross-scale test that measures something new
regardless of §3: raw glyph edit distance and minimum inter-occurrence
line distance show a small positive frequency-bin-controlled correlation
(0.0343) that does **not** clear the declared BH-FDR threshold
(q=0.386) — a real, negative result, not an artifact of missing data.

Held-out validation (`groupedKFoldLogLoss`, grouped by folio) is attached
to CS1; with zero family-bearing occurrences it correctly reports "no
held-out result attached" rather than a fabricated number.
`CROSS_SCALE_EFFECTS_GENERALIZE = INCONCLUSIVE` and
`CROSS_SCALE_EFFECTS_SURVIVE_CONDITIONING = INCONCLUSIVE` for the same
root cause.

**Methodological note on the confounding guardrail.** Task77's central
ban — "two separately significant marginal results do not establish
`X ⊥̸ Y`" — is exercised directly by `TestCS1ConfoundedByRegime`
(`internal/fingerprintv2/crossscale_test.go`): a synthetic corpus where
every regime uses a single constant family (so each regime *alone* has
zero family entropy, hence provably zero within-regime association) still
produces a nonzero **pooled** family/position association purely from
regime composition. The test asserts the pooled effect is nonzero while
each conditioned stratum is exactly zero, confirming the CS8 conditioning
mechanism (and CS1's N2 within-line null, which is inherently
folio/regime-blind by construction) behaves as designed. `TestCS1
PositiveControl` and `TestCS1NegativeControl` cover the maximal-effect and
exact-independence cases the same way.

## 9. Stability and replication

Covered inline in §2 (the perturbation battery) since task77 folds this
into the same battery for the edit-family block; no additional per-CS
stability sweep was run this session beyond what's already reported
(`cs1/family-line-position.partition_stability`, empty this run for the
reason above).

## 10. Redundancy analysis

Computed from the 200 grammar replicates (100 × 2 modes) of the primary
corpus as a same-corpus, same-generative-process sample (`redundancy.go`).
Real correlations found:

- `ef1_giant_component_share` vs. `ef1_isolate_share`: **r = −0.993**
- `ef1_giant_component_share` vs. `ef3_spearman_degree_log_freq`: **r = −0.967**
- `ef1_giant_component_share` vs. `lp4_prefix_nmi` / `lp4_suffix_nmi`: **r = −0.963 / −0.946**
- `ef1_giant_component_share` vs. `ef2_global_clustering`: **r = −0.864**

Classification: `ef1_giant_component_share`, `ef2_global_clustering` and
`lp1_support_gini` are kept `CORE`; `ef1_isolate_share`,
`ef3_spearman_degree_log_freq`, `lp4_prefix_nmi` and `lp4_suffix_nmi` are
`REDUNDANT` with `ef1_giant_component_share` at this corpus's C-GRAMMAR
generative process (|r|≥0.9), retained as diagnostics rather than dropped
per task77's own instruction not to drop metrics for correlation alone.
This redundancy is a property of *how these six statistics co-vary across
one random-graph-generation process*, not a claim that they would be
redundant across genuinely different corpora or mechanisms — a caveat
worth preregistering for any future redundancy pass across multiple
corpora.

## 11. Confirmatory vs. exploratory

**Confirmatory** (preregistered in `TASK77_CROSSSCALE_HYPOTHESES.md`):
CS1-CS8 and EF5 as tabulated in §3-4; all eight came back with a
documented status (three real `NOT_SUPPORTED`, one real negative CS7
result, the rest `INCONCLUSIVE`/`NOT_APPLICABLE` due to the empty LP2
family gate). `EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL = NOT_SUPPORTED` is
the headline confirmatory result of this entire report.

**Exploratory** (`cross_scale.exploratory_findings`, not the basis for any
`SUPPORTED` verdict):

1. Support-threshold-only edit-rule families are structurally stable
   across independent folio halves (ARI 0.72) even though the LP2-gated
   family set is empty — see §2. **Recommended confirmatory follow-up:**
   test support-threshold family stability directly against a C-GRAMMAR
   null (not just against the empty LP2 baseline), and consider whether
   LP2's all-or-nothing significance gate is the right threshold for
   downstream family-conditioned cross-scale work, or whether a graded/
   continuous family-membership weight would preserve more signal.
2. Connected-components-vs-label-propagation and hub-removal diagnostics
   are vacuous this run (empty graph); re-run once a follow-up relaxes or
   removes the LP2 gate for Stage-2 purposes specifically.
3. The locus-type `SPECIAL` bucket (comments/running/radial text) is
   pooled for statistical power in CS3; a per-code confirmatory follow-up
   needs more data than most individual special-locus codes currently
   provide.

## 12. Final verdicts

| Verdict | Value | Basis (one line) |
|---|---|---|
| `TASK75_RESULTS_REPRODUCED` | `PARTIALLY_SUPPORTED` | All formulas correct; four infrastructure defects found+fixed on first real run |
| `EDIT_FAMILIES_STRUCTURALLY_STABLE` | `NOT_SUPPORTED` | LP2-gated family graph is empty this run |
| `EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL` | **`NOT_SUPPORTED`** | EF1/EF2/EF3 all *below* C-GRAMMAR null, large one-sided effects |
| `TRANSFORMATION_MOTIFS_STABLE` | `INCONCLUSIVE` | Nothing to compare on an empty graph |
| `FAMILY_LINE_POSITION_DEPENDENCE` (CS1) | `INCONCLUSIVE` | 0 family-bearing occurrences |
| `TRANSFORMATION_CONTEXT_DEPENDENCE` (CS2) | `INCONCLUSIVE` | 0 family-bearing adjacent pairs |
| `FAMILY_LOCUS_DEPENDENCE` (CS3) | `NOT_SUPPORTED` | Real test, real null result |
| `STRUCTURAL_DISTANCE_EDIT_DISTANCE_DEPENDENCE` (CS7) | `NOT_SUPPORTED` | Real test, q=0.386 |
| `FAMILY_FOLIO_REGIME_DEPENDENCE` (CS4/EF5) | `NOT_APPLICABLE` | 0 family-bearing occurrences |
| `CROSS_SCALE_EFFECTS_SURVIVE_CONDITIONING` (CS8) | `INCONCLUSIVE` | Nothing to condition |
| `CROSS_SCALE_EFFECTS_GENERALIZE` | `INCONCLUSIVE` | No held-out result attached |
| `EDIT_CROSS_SCALE_BLOCK_READY` | `PARTIALLY_SUPPORTED` | Block runs end-to-end on real data with real nulls; not a content conclusion |

## What is and is not ready to freeze into Fingerprint v2

**Ready:** EF4/EF5's grammar-boundedness machinery (validated end-to-end
against real data for the first time — this is the strongest, cleanest
result in this report and directly answers task73 §8); CS3, CS6, CS7 as
genuine, real, non-family-gated cross-scale tests; the full null-model
registry (N1-N8); the redundancy/classification machinery.

**Not ready:** any family-conditioned metric (CS1, CS2, CS4, CS5, CS8,
Stage-2's structural diagnostics) as currently gated, since LP2's
significance requirement empties the family set on the real corpus. This
is not a bug to patch reflexively — it may be the *correct* answer (the
giant component may simply not be a "paradigm" by this operationalization)
— but it means task77's family-conditioned cross-scale block needs either
(a) a follow-up confirming this null result is stable across
`min_rule_support`/`alpha` choices, or (b) a deliberately different,
explicitly graded family-membership definition (e.g. the support-
threshold-only definition from the stability battery, evaluated in its
own right against its own null) before any of CS1/CS2/CS4/CS5/CS8 can be
frozen as evidence either way.

## Recommendations for the next stage

1. Re-run this exact config with `min_rule_support` at 1 and 2, and
   report whether `DIRECTIONAL_TRANSFORMATIONS_SUPPORTED` and the
   family-conditioned CS block remain `NOT_SUPPORTED`/`INCONCLUSIVE` or
   flip — directly following up finding (1) in §11.
2. If (1) still comes back `NOT_SUPPORTED`, treat
   `EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL = NOT_SUPPORTED` as a stable,
   confirmed Fingerprint v2 feature and move the family-conditioned CS
   block to `DEFERRED` (not `DROPPED`) pending a redesigned family
   definition, rather than leaving it perpetually `INCONCLUSIVE`.
3. Add a second, independent real transcription if one becomes available
   under the repository's data discipline, to finally test the
   `transcription_variant` stability row (`NOT_TESTABLE` in every run so
   far, including this one).
4. Extend the redundancy analysis across multiple corpora (not just one
   corpus's own grammar replicates) before treating the CORE/REDUNDANT
   classification in §10 as final — see that section's caveat.
