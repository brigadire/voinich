# Voynich Fingerprint v2 — Specification

Status: **DESIGN, NOT FROZEN**. This document specifies candidate metric
families for Phase II. It does not run them, does not compute new numbers,
and does not select or exclude any origin hypothesis for the Voynich
Manuscript. See [FINGERPRINT_V2_GAPS.md](FINGERPRINT_V2_GAPS.md) for
priority and [TASK73_REPORT.md](TASK73_REPORT.md) for the synthesis.

This spec builds on the audit in
[FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv) and is
constrained by parent task `tasks_ph2/task73.txt`: it must not maximize
metric count, must separate marginal/conditional/joint/cross-scale
properties, and must not be designed with any Fontana/mnemonic machine in
mind (`tasks_ph2/task73.txt` sections 3, 4, 30).

## 0. Control-family legend

Every metric below names one or more control families by ID. Full
preserved/destroyed semantics, and the rule for choosing among them, are in
[FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md). IDs used here:

| ID | Short description |
|---|---|
| C-GLOBAL | global token shuffle across the whole corpus |
| C-LINE | within-line token shuffle |
| C-PAGE | within-page (cross-line, same folio) token/line shuffle |
| C-LEN | length-matched random resampling/pairing |
| C-FREQ | frequency-matched random resampling/pairing |
| C-POS | position-matched shuffle (preserves positional marginals) |
| C-REGIME | shuffle restricted to within Task65-style regime clusters |
| C-WITHINTOKEN | within-token glyph shuffle |
| C-GRAMMAR | grammar-bounded synthetic null: independent tokens matching length distribution and positional-glyph marginals only |
| C-NAT | natural-language controls (Doyle, Longfellow, Astafiev) |
| C-PHASE1 | Phase I generated/synthetic mechanisms (homophony series, copy/mutate, M0-M11, structured-token positive controls) |

## 1. Framework: marginal / conditional / joint / cross-scale

Every metric family below is tagged with exactly one primary type
(parent task section 4):

- **marginal** — describes `P(X)` for a single property at a single level.
- **conditional** — describes `P(X|Y)` for one property given one
  conditioning variable at the same or an adjacent level.
- **joint** — describes `P(X,Y)` for two properties without collapsing to
  a single conditional direction.
- **cross-scale** — a conditional or joint statistic where `X` and `Y`
  come from two different levels of the section-13 hierarchy
  (glyph -> token -> lexical family -> local regime -> page ->
  section/manuscript).

Fingerprint v2 explicitly weights conditional/joint/cross-scale families
over marginal ones: five of eleven level-groups below (lexical paradigms,
edit-family geometry, cross-scale, hierarchy, sequence) contribute no new
pure-marginal metric at all, because Phase I already covers marginals for
those levels well ([FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv)).

Each family below carries a **Justification** field that answers, per
parent task section 3: what property is measured; what models it could
distinguish; whether it is redundant with an existing Phase I metric or
another v2 family; whether it is statistically estimable at corpus size;
and what stability is expected across subsamples. A family that fails the
redundancy or estimability check is not included, and is instead recorded
in the "considered and excluded" note at the end of its section.

---

## 2. Glyph level

### G1 — Glyph n-gram distribution with explicit coverage

- **Type:** marginal (order 2), extended to conditional at G4.
- **Definition:** the empirical distribution of glyph bigrams and
  trigrams over a declared stream representation (continuous /
  boundary-symbol / within-token, matching Task61's three modes), reported
  together with the fraction of possible contexts with support above a
  declared minimum count (coverage), not as a bare entropy number.
- **Level:** glyph.
- **Estimator:** plug-in frequency counts; report both the distribution
  and Good-Turing-style unseen-mass estimate for the trigram tier where
  coverage is low.
- **Input:** parsed glyph stream, per representation mode, per corpus.
- **Normalization:** counts normalized by valid-context total; coverage
  reported per order per representation mode.
- **Null/control:** C-GLOBAL (glyph-shuffle analogue: shuffle glyphs
  within token boundaries preserved, i.e. C-WITHINTOKEN, plus C-NAT,
  C-PHASE1 homophony series for calibration).
- **Uncertainty:** bootstrap over folio blocks (see
  [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) stability
  section); trigram tier confidence intervals conditioned on coverage.
- **Missing-data behavior:** any context cell below the minimum-count
  threshold is reported as `INSUFFICIENT_SUPPORT`, not imputed or dropped
  silently.
- **Interpretation:** a direct, representation-explicit extension of
  Task61's h2; trigram tier only interpreted where coverage passes
  threshold.
- **Known limitations:** representation-sensitive like h2; trigram tier
  will likely stay low-coverage given corpus size (Task61 h4 coverage was
  only 0.116).
- **Justification:** measures short-range glyph predictability beyond a
  single scalar; distinguishes constrained-alphabet generators from
  higher-entropy ones; not redundant with Task61's h2 because it retains
  the full distribution and an explicit coverage report instead of one
  number; estimable at bigram order, marginal at trigram order; expected
  stable at bigram order under folio-block bootstrap, expected unstable at
  trigram order (this is reported, not hidden).

### G2 — Full internal-position profile

- **Type:** conditional — `P(glyph | position index, token length)`.
- **Definition:** replaces Task59's three-bucket (initial/medial/final)
  specialization test with the full curve of glyph frequency against
  position index normalized by token length (position index divided by
  length, binned), reported separately for each length class with enough
  tokens.
- **Level:** glyph / within-token boundary.
- **Estimator:** empirical conditional frequency per (length-class,
  normalized-position-bin) cell.
- **Input:** parsed tokens with per-glyph position and token length.
- **Normalization:** per length-class, positions normalized to `[0,1]`
  before binning so classes of different length are comparable.
- **Null/control:** C-WITHINTOKEN shuffle (destroys position, preserves
  token length and glyph multiset); C-NAT; C-PHASE1 position-independent
  and position-dependent homophony variants (Task59's existing controls,
  reused).
- **Uncertainty:** per-cell binomial/multinomial confidence interval;
  cells below minimum count marked `INSUFFICIENT_SUPPORT`.
- **Missing-data behavior:** length classes with too few tokens are
  excluded from the curve, not extrapolated.
- **Interpretation:** subsumes and extends Task59's near-strict
  specialist result (start/end asymmetry is a special case of this curve
  at the bin boundaries, so it is not specified as a separate family).
- **Known limitations:** binning choice affects resolution; still ties
  "position" to linear transcription order within a token, not to any
  claimed phonological or graphemic role.
- **Justification:** subsumes the "start/end asymmetry" and "internal
  position dependency" bullets from parent task section 5 in one
  estimator instead of two, avoiding a redundant pair of near-identical
  families; distinguishes coarse three-class specialization (already
  known) from finer within-position structure that a three-bucket test
  cannot see; estimable for the high-frequency glyphs already known to
  specialize, marginal for rare glyphs (reported via coverage, not
  hidden); expected to reproduce Task59's six specialists as a special
  case, which is itself a stability check.

### G3 — Glyph substitution neighborhood

- **Type:** joint — `P(glyph_out | glyph_in, aligned position)` over
  adjacent-in-corpus token pairs at edit distance 1.
- **Definition:** for every globally (not just text-adjacent) edit-
  distance-1 token pair sharing a length, tabulate which glyph at which
  aligned position is substituted for which other glyph, producing a
  glyph x glyph substitution-count matrix per position class.
- **Level:** glyph, feeding edit-family and lexical-paradigm levels.
- **Estimator:** count matrix plus row/column-normalized substitution
  probabilities; report matrix sparsity and effective rank.
- **Input:** the edit-distance-1 pair set already used for the lexical
  and edit-family levels (single shared computation, see redundancy note
  below).
- **Normalization:** row-normalized (given `glyph_in`, distribution over
  `glyph_out`).
- **Null/control:** C-GRAMMAR (grammar-bounded null: same length
  distribution and G2 positional marginals, independent draws) to test
  whether the substitution matrix is more concentrated than "any glyph
  can become any glyph with marginal-implied probability."
- **Uncertainty:** bootstrap over distinct token-pair instances (not over
  raw glyph counts, to avoid pseudo-replication from one frequent pair
  appearing many times).
- **Missing-data behavior:** position classes with fewer than a declared
  minimum number of pair instances are reported `INSUFFICIENT_SUPPORT`.
- **Interpretation:** a concentrated, position-specific substitution
  matrix is compatible with a small structured alphabet/cipher table or
  a constrained morphological system; a diffuse matrix close to the
  grammar-bounded null is compatible with accidental edit-distance
  proximity under a bounded alphabet.
- **Known limitations:** edit distance 1 is a formal, not a
  linguistic/cryptographic, definition of "substitution"; the same pair
  can be explained by more than one underlying process.
- **Justification:** operationalizes the "glyph substitution
  neighborhoods" bullet from parent task section 5 as a statistic that
  directly feeds the lexical-paradigm question (section 7) rather than
  standing alone; not redundant with LP1 (token-level transformation
  census) because it aggregates at the glyph-pair grain, which LP1 does
  not; estimable given Task60's existing pair volume; stability expected
  to depend on pair-instance count per position class, which will be
  reported rather than assumed.

### G4 — Conditional glyph entropy by length and by regime

- **Type:** conditional — `P(X_i|X_{i-1}, token length)` and
  `P(X_i|X_{i-1}, regime)`.
- **Definition:** Task61's `h2` recomputed within length-class strata and
  within Task65 regime-cluster strata, instead of pooled over the whole
  corpus.
- **Level:** glyph, cross-referenced as CS6 at the cross-scale level.
- **Estimator:** same plug-in conditional entropy as Task61, computed per
  stratum with declared minimum-count coverage per stratum.
- **Input:** parsed glyph stream plus token-length and regime-cluster
  labels.
- **Normalization:** none beyond the existing h2 definition; comparability
  across strata is via the shared estimator, not a rescaled index.
- **Null/control:** C-WITHINTOKEN, C-GLOBAL, C-REGIME, C-NAT, C-PHASE1
  (same families as Task61, applied per stratum).
- **Uncertainty:** per-stratum bootstrap; strata below minimum coverage
  marked `INSUFFICIENT_SUPPORT`.
- **Missing-data behavior:** small regimes/length classes excluded from
  the stratified report rather than pooled silently into a misleading
  average.
- **Interpretation:** if entropy is materially stable across strata,
  low entropy is a global property of the writing system; if it varies
  by regime or length, entropy is partly a local/contextual property,
  which changes what "low glyph entropy" as a fingerprint row means.
- **Known limitations:** stratification reduces sample size per cell and
  can widen intervals enough to mask real but small differences.
- **Justification:** this is the glyph-level half of the cross-scale
  "entropy x regime" example named explicitly in parent task section 14;
  it is not redundant with Task61's pooled h2 because pooling is exactly
  the property being tested against; estimable for the larger regimes
  and length classes, reported as insufficient for the rest; expected to
  be less stable than the pooled estimate by construction, which is the
  point of stratifying.

**Considered and not included:** a separate "prefix/suffix dependency at
glyph grain" family was considered and folded into the token-formation
level (TF1) instead, because Task62's zone-decomposed formation model is
the more natural estimator for prefix/core/suffix structure and a
glyph-only version would duplicate it without adding information.

---

## 3. Token formation

### TF1 — Zone-decomposed (prefix/core/suffix) formation model

- **Type:** conditional — `P(glyph sequence | zone, token length)`.
- **Definition:** extends Task62's `POSITION_MARKOV_1` by explicitly
  partitioning each token into three zones (prefix, core, suffix) using a
  frozen, declared boundary rule (e.g. fixed-length prefix/suffix windows
  or a data-driven zone boundary chosen on the training split only, never
  on test), and fitting/reporting formation statistics separately per
  zone instead of one pooled first-order model.
- **Level:** within-token / token formation.
- **Estimator:** per-zone position-conditioned glyph model, selected and
  validated exactly as Task62 (train/validation/test contiguous blocks,
  validation-only model selection).
- **Input:** parsed tokens, zone boundaries, contiguous block split.
- **Normalization:** cross-entropy in bits/glyph, reported per zone and
  pooled, so the pooled number remains comparable to Task62's baseline.
- **Null/control:** the same family used by Task62: length/frequency
  controls, C-PHASE1 copy/mutate and structured-token controls, C-NAT.
- **Uncertainty:** test-block bootstrap, as Task62.
- **Missing-data behavior:** zones with too few tokens at a given length
  class are reported `INSUFFICIENT_SUPPORT`, not merged into another
  zone.
- **Interpretation:** if per-zone models fit materially better than the
  pooled Task62 model, formation is zone-structured (compatible with
  affix-like or codebook-slot structure); if not, the pooled model
  already captured the available structure.
- **Known limitations:** the zone-boundary rule is a modeling choice; a
  wrong boundary understates zone structure without falsifying it.
- **Justification:** directly answers part of the parent section 6
  question ("prefix/core/suffix dependencies"); not redundant with
  Task62 because it changes the model class, not just re-reports it;
  estimable using the existing Task62 split; stability follows Task62's
  precedent (same estimator family, same split discipline).

### TF2 — Stratified held-out predictive likelihood

- **Type:** conditional, applied across several conditioning axes.
- **Definition:** the single formation-model held-out cross-entropy
  (Task62's headline number) reported disaggregated by conditioning
  stratum — token length, line position (from PG2), local regime
  (Task65 clusters), and lexical-paradigm family (once LP1-LP2 exist) —
  instead of one pooled scalar.
- **Level:** within-token, cross-referenced as CS3 (formation x line
  position) at the cross-scale level.
- **Estimator:** same as TF1/Task62, evaluated per stratum on the same
  frozen test blocks.
- **Input:** Task62 model outputs plus stratum labels.
- **Normalization:** bits/glyph per stratum, with per-stratum token
  count reported alongside.
- **Null/control:** same as TF1.
- **Uncertainty:** per-stratum bootstrap; small strata flagged.
- **Missing-data behavior:** strata without a paradigm-family label yet
  (before LP1-LP2 are implemented) are reported as `NOT_YET_AVAILABLE`,
  not silently pooled.
- **Interpretation:** directly answers the section-6 question "how
  global are token-formation rules, how much do they depend on
  context/regime?" — a flat cross-entropy across strata argues for
  global rules; systematic stratum differences argue for context
  dependence.
- **Known limitations:** disaggregation reduces per-cell sample size;
  requires the lexical-family labels from section 7 to be complete.
- **Justification:** consolidates four separate "conditioned formation"
  bullets from parent task section 6 (length-, position-, family-,
  regime-conditioned) into one estimator applied under different
  strata, instead of four redundant model-fitting exercises; estimable
  now for length/position/regime strata, deferred for family strata
  until section 7 output exists; expected stability is the whole point
  of the metric (it measures where stability breaks down).

### TF3 — Token novelty and near-novelty rate

- **Type:** marginal (novelty rate) plus conditional (rate | distance to
  nearest training type).
- **Definition:** on the same held-out test blocks as Task62/TF1, the
  fraction of test tokens absent from the training vocabulary (exact
  novelty), and the distribution of edit distance from each novel test
  token to its nearest training-vocabulary neighbor (near-novelty).
- **Level:** token / lexical boundary.
- **Estimator:** direct counting plus nearest-neighbor edit distance
  (reusing the edit-distance-1 pipeline from Task60/edit-family level at
  distance thresholds beyond 1 where needed).
- **Input:** train/test vocabulary sets from the frozen contiguous split.
- **Normalization:** rate per test-block size; distance distribution
  reported as a histogram, not a single mean.
- **Null/control:** C-PHASE1 copy/mutate and structured-token generators
  (Task62's own controls, which already report a generated-vs-observed
  near-repeat comparison); C-NAT for scale reference.
- **Uncertainty:** block bootstrap over held-out blocks.
- **Missing-data behavior:** none expected; novelty is well-defined for
  any nonempty test block.
- **Interpretation:** a system generating tokens from a small closed
  paradigm set would show low exact novelty and near-novelty
  concentrated at distance 1; a system with an open/generative token
  space would show higher exact novelty.
- **Known limitations:** sensitive to train/test split size and to
  corpus-wide vocabulary size, so cross-corpus comparison must control
  for those (see [FINGERPRINT_V2_DISTANCE.md](FINGERPRINT_V2_DISTANCE.md)
  common-core distance).
- **Justification:** operationalizes the "token novelty" bullet from
  parent task section 6; not redundant with Task62's aggregate
  cross-entropy because novelty is a discrete-event statistic, not an
  average log-likelihood, and behaves differently under model
  misspecification; estimable directly; stability checked via block
  bootstrap.

**Considered and not included:** a fully separate "family-conditioned
formation" family was not created; it is one of TF2's conditioning axes
once lexical-paradigm labels (section 4 below) exist, to avoid
duplicating the estimator.

---

## 4. Lexical paradigms

This is the largest explicitly named Phase I gap
([FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv), row
`lexical`). The four families below are ordered so that each depends on
the output of the previous one, and together they are designed to answer
the parent task's explicit requirement: separate accidental edit-distance
proximity from statistically productive transformation systems.

### LP1 — Transformation-pattern census

- **Type:** joint — `P(rule)` where a rule is `(position class, glyph_in,
  glyph_out)` or `(position class, operation type)` for insertions/
  deletions, counted across **distinct** token-pair instances.
- **Definition:** for every edit-distance-1 (and, separately, distance-2)
  pair of distinct vocabulary types anywhere in the corpus (not only
  text-adjacent pairs, unlike Task60), classify the edit into a rule as
  above, and tabulate rule support as the number of *distinct token-pair
  instances* exhibiting that rule (not raw occurrence count of the
  tokens, which would let one frequent pair dominate).
- **Level:** lexical.
- **Estimator:** direct enumeration and counting over the vocabulary
  graph; reuses the edit-distance computation already built for Task60
  and for G3/EF1 (single shared computation across three families, see
  redundancy note below each).
- **Input:** corpus vocabulary (types only, not token occurrences) plus
  token lengths and positions for classification.
- **Normalization:** rule support reported as a raw count and as a share
  of all distance-1/2 pairs.
- **Null/control:** C-GRAMMAR (grammar-bounded null using the same
  vocabulary size and length/positional-glyph marginals) to establish
  what rule-support concentration would arise from a bounded alphabet
  alone, with no productive paradigm.
- **Uncertainty:** bootstrap over vocabulary types (resample types with
  replacement, recompute rule support).
- **Missing-data behavior:** rules with support below a declared minimum
  are pooled into an explicit `RARE_RULE` bucket, not dropped silently.
- **Interpretation:** this census alone is descriptive; it becomes
  evidence for or against productivity only jointly with LP2.
- **Known limitations:** type-level (not token-level) counting avoids
  frequency domination but is sensitive to how many hapax types exist;
  distance-2 enumeration is combinatorially larger and may need a
  candidate-generation shortcut (e.g. only pairs sharing a declared
  minimum common substring) which must be declared, not left implicit.
- **Justification:** measures whether a small number of transformation
  rules account for a disproportionate share of near-duplicate pairs
  (the operational meaning of "paradigm" here); distinguishes an
  affix-substitution/codebook system (few high-support rules) from an
  unstructured near-repeat cloud (many low-support rules, matching
  C-GRAMMAR); not redundant with Task60 because Task60 measured pair
  *rates*, never rule *identity* or *support concentration*; estimable
  at distance 1 with existing pair volume, more marginal at distance 2;
  stability checked via type-resampling bootstrap.

### LP2 — Paradigm productivity test

- **Type:** joint, a hypothesis test over LP1's output.
- **Definition:** compare the observed concentration of rule support
  (e.g. Gini coefficient of the rule-support distribution, or the count
  of rules with support at or above a declared threshold `k`) against
  the same statistic computed under C-GRAMMAR and under C-LEN/C-FREQ
  matched random re-pairing of the same vocabulary.
- **Level:** lexical.
- **Estimator:** permutation test: recompute LP1's concentration
  statistic under many draws from each null family; report the
  percentile of the observed statistic.
- **Input:** LP1 output plus the same vocabulary/length/frequency data
  used to build its nulls.
- **Normalization:** concentration statistics are already
  scale-normalized (Gini in `[0,1]`; rule count above threshold reported
  alongside total distinct rules for context).
- **Null/control:** C-GRAMMAR (primary), C-LEN, C-FREQ.
- **Uncertainty:** permutation-derived p-value/percentile; multiple
  testing correction if more than one concentration statistic is
  reported as confirmatory.
- **Missing-data behavior:** if the vocabulary is too small for a
  meaningful null distribution at distance 2, report
  `INSUFFICIENT_SUPPORT` for that distance rather than a null with wide,
  uninformative bounds treated as a pass.
- **Interpretation:** a rejection of C-GRAMMAR (observed concentration
  higher than the null distribution supports) is evidence for productive
  transformation structure beyond what a bounded alphabet alone
  predicts; failure to reject means LP1's census, however visually
  striking, does not exceed the bounded-grammar expectation.
- **Known limitations:** a single global test can be significant while
  being driven by a handful of rules; LP3 examines structure within the
  productive rule set to address this.
- **Justification:** this is the operational answer to the parent task's
  explicit instruction to separate accidental edit-distance proximity
  from productive systems (section 7); not redundant with EF4 (grammar-
  boundedness at the graph-geometry level) because LP2 tests rule-
  identity concentration while EF4 tests graph-shape statistics — they
  can disagree and that disagreement is itself informative; estimable
  given LP1's output; stability follows from the permutation framework
  by construction.

### LP3 — Family branching, depth, overlap and locality

- **Type:** joint/conditional, computed on the subgraph of rules that
  pass LP2.
- **Definition:** restricting the edit-distance graph to the
  productive-rule subset from LP2, compute per connected component
  (family): branching factor (mean number of distinct-rule neighbors per
  token), depth (longest shortest-path chain within the family), overlap
  (fraction of tokens reachable by more than one distinct productive
  rule), and locality (whether family-member tokens co-occur in the same
  line/page more than a C-GLOBAL/C-PAGE null predicts).
- **Level:** lexical, cross-referenced at cross-scale (locality component
  overlaps with EF5).
- **Estimator:** standard graph statistics on the filtered subgraph;
  locality via the same co-occurrence null used for EF5 (shared control,
  not a duplicate one).
- **Input:** LP2's confirmed-productive rule set plus the full vocabulary
  graph and token position metadata.
- **Normalization:** branching/depth reported per family with family size
  as context (a large family will mechanically have higher branching);
  overlap and locality reported as rates.
- **Null/control:** C-GRAMMAR for branching/depth/overlap; C-GLOBAL and
  C-PAGE for locality.
- **Uncertainty:** bootstrap over families (resample families, not
  tokens, to avoid one large family dominating).
- **Missing-data behavior:** families below a minimum size are excluded
  from branching/depth summaries and reported separately as a count of
  small/singleton families.
- **Interpretation:** high branching and overlap with low depth is
  compatible with a grid-like (multi-slot codebook or paradigm-table)
  system; high depth with low branching is compatible with a chain-like
  copy/mutate process (Task62/Task66's copy/mutate control is the
  natural comparison point); strong locality suggests scribal/production
  proximity rather than a purely lexical/grammatical paradigm.
- **Known limitations:** graph statistics are sensitive to the distance
  threshold and to the LP2 significance threshold used to build the
  filtered subgraph.
- **Justification:** directly requested by parent task section 7
  (branching, depth, overlap, locality); not redundant with EF1-EF3
  because those describe the *raw* edit graph while LP3 describes only
  the *validated-productive* subgraph — the difference between them is
  itself diagnostic (see EF4); estimable once LP1-LP2 exist; stability
  via family-level bootstrap.

### LP4 — Core/affix joint attachment test

- **Type:** joint — `P(core, affix)` vs. the independence product
  `P(core) x P(affix)`.
- **Definition:** for a declared affix-boundary convention (shared with
  TF1's zone decomposition), test whether the empirical joint
  distribution of (core, prefix) and (core, suffix) pairs departs from
  what independent marginal attachment would predict, using a
  chi-square/mutual-information style statistic with a permutation null.
- **Level:** lexical, tightly coupled to TF1.
- **Estimator:** contingency-table MI or G-test on the (core, affix)
  table, restricted to cores/affixes with sufficient support.
- **Input:** TF1's zone decomposition applied to the full vocabulary.
- **Normalization:** MI normalized by marginal entropy, as in Task58's
  existing MI-share convention, for comparability with other MI-based
  rows in the fingerprint.
- **Null/control:** C-GRAMMAR and C-LEN (independent attachment given
  the same core/affix marginal frequencies and length distribution).
- **Uncertainty:** permutation-based, as LP2.
- **Missing-data behavior:** cores or affixes below minimum support
  excluded and counted, not imputed.
- **Interpretation:** strong joint dependence between specific cores and
  specific affixes is compatible with a productive
  affix-substitution/paradigm-table system (the commonly discussed
  "qo-/-dy/-aiin"-type structure in Voynichese literature, tested here
  statistically rather than assumed); independence is compatible with
  affixes attaching freely to any core.
- **Known limitations:** depends entirely on the TF1 zone-boundary
  convention; a poorly chosen boundary will understate real structure.
- **Justification:** gives a direct joint-distribution test for the
  specific "prefix/suffix substitution paradigms" bullet in parent task
  section 7, phrased as `P(X,Y)` rather than a marginal count, matching
  the parent task's own worked example in section 4; not redundant with
  LP1/LP2 because those operate on edit distance between whole tokens,
  while LP4 operates on a declared zone decomposition; estimable for
  high-support cores/affixes; stability tied to TF1's zone convention
  and checked jointly with it.

---

## 5. Edit-family geometry

### EF1 — Degree distribution and component-size distribution

- **Type:** marginal (on the graph).
- **Definition:** build the global edit-distance<=1 graph over the
  corpus vocabulary (shared computation with LP1/G3); report the degree
  distribution and the connected-component size distribution, including
  the size and share of the giant component.
- **Level:** edit-family.
- **Estimator:** direct graph construction and standard degree/component
  algorithms.
- **Input:** corpus vocabulary and pairwise edit distances (threshold 1,
  optionally 2 as a sensitivity check).
- **Normalization:** degree reported both raw and as a share of
  vocabulary size; component sizes reported as a rank-size distribution.
- **Null/control:** C-GRAMMAR (see EF4 for the full comparison).
- **Uncertainty:** bootstrap over vocabulary types.
- **Missing-data behavior:** none; well-defined for any vocabulary.
- **Interpretation:** by itself, descriptive; interpreted jointly with
  EF4.
- **Known limitations:** distance-1 threshold is a modeling choice;
  sensitivity to it must be reported (distance 2 as an explicit
  robustness check, not a silent alternative).
- **Justification:** the parent report explicitly warns that Phase I
  never characterized the giant edit family as a graph
  ("graph families are not morphemes" is a caution, not a measurement);
  this is a low-cost, well-estimable extension of data already collected
  for Task60; expected reasonably stable under type-resampling given the
  large vocabulary size.

### EF2 — Clustering coefficient and local motif census

- **Type:** marginal (on the graph).
- **Definition:** global and local (per-node) clustering coefficient,
  plus counts of small motifs (triangles, 3-paths, 4-cycles) in the
  edit-distance graph.
- **Level:** edit-family.
- **Estimator:** standard graph-theoretic statistics.
- **Input:** same graph as EF1.
- **Normalization:** clustering coefficient is already `[0,1]`; motif
  counts normalized against a degree-sequence-preserving random-graph
  null (configuration model), in addition to C-GRAMMAR.
- **Null/control:** configuration-model random graph (degree-sequence
  preserving, a graph-theoretic control distinct from but complementary
  to C-GRAMMAR) and C-GRAMMAR.
- **Uncertainty:** bootstrap and configuration-model resampling.
- **Missing-data behavior:** none.
- **Interpretation:** clustering/motif excess over the configuration-
  model null indicates structure beyond the degree sequence alone (e.g.
  shared paradigm membership creating triangles); excess over C-GRAMMAR
  indicates structure beyond the length/positional-glyph marginals.
- **Known limitations:** motif census on a large near-complete component
  can be computationally heavy; a declared sampling procedure should be
  used if exact enumeration is infeasible, and that procedure must be
  reported.
- **Justification:** directly requested (parent task section 8:
  clustering, local motifs); not redundant with EF1 because degree
  distribution alone does not determine clustering/motif structure;
  estimable with standard graph sampling techniques; stability checked
  via the two independent nulls.

### EF3 — Family centrality vs. token frequency

- **Type:** conditional — `P(degree | corpus frequency)`.
- **Definition:** correlation/regression of edit-graph degree (or a
  centrality measure such as eigenvector centrality) against corpus
  token frequency.
- **Level:** edit-family, cross-referencing the token level.
- **Estimator:** rank correlation (robust to the heavy-tailed frequency
  distribution) plus a regression with confidence bands.
- **Input:** EF1's graph plus token frequency counts.
- **Normalization:** frequency log-transformed before regression.
- **Null/control:** C-GRAMMAR (frequency and degree are independent by
  construction there, since it does not encode any frequency-dependent
  process) and C-FREQ.
- **Uncertainty:** bootstrap over vocabulary types.
- **Missing-data behavior:** none.
- **Interpretation:** a strong frequency-degree correlation is compatible
  with high-frequency tokens acting as paradigm "hubs" (e.g. common
  function-like tokens with many attested near-variants); a weak
  correlation argues against that specific picture.
- **Known limitations:** correlation does not identify direction or
  mechanism; frequent tokens have more opportunity to have near-variants
  simply by corpus size, which is exactly why C-FREQ is required.
- **Justification:** directly requested (parent task section 8: "token-
  frequency relation"); not redundant with EF1 (a joint statistic, not a
  marginal one); estimable directly; stability via bootstrap.

### EF4 — Grammar-boundedness diagnostic

- **Type:** joint, a comparison across EF1-EF3 (and LP1-LP2) against
  C-GRAMMAR, elevated to its own named diagnostic because it answers a
  specific question posed by the parent task.
- **Definition:** report, as one consolidated table, how close the
  observed EF1 (degree/component), EF2 (clustering/motif) and EF3
  (frequency-degree) statistics are to their C-GRAMMAR expectation,
  alongside LP2's productivity-test verdict.
- **Level:** edit-family / cross-scale (edit-family x token-formation
  marginals).
- **Estimator:** the individual EF1-EF3/LP2 estimators; EF4 itself adds
  no new computation, only a joint reporting table and a single
  qualitative verdict field (`CONSISTENT_WITH_GRAMMAR_BOUND`,
  `EXCEEDS_GRAMMAR_BOUND`, `MIXED`).
- **Input:** EF1-EF3 and LP1-LP2 outputs.
- **Normalization:** not applicable (a synthesis table).
- **Null/control:** inherited from EF1-EF3/LP1-LP2.
- **Uncertainty:** inherited.
- **Missing-data behavior:** if any contributing metric is
  `INSUFFICIENT_SUPPORT`, the verdict field reports `MIXED` with the
  reason, never a forced binary verdict.
- **Interpretation:** this is the direct answer to the parent task
  section 8 question, "is the giant family simply a consequence of a
  limited token grammar?"
- **Known limitations:** the verdict depends on the C-GRAMMAR
  construction being a fair representation of "bounded grammar, no
  productive paradigm," which is itself a modeling choice that must be
  documented in [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md).
- **Justification:** answers a question posed explicitly by name in the
  parent task rather than leaving it implicit across several separate
  tables; adds no redundant computation (pure synthesis); estimable
  whenever its inputs are; stability inherited from its inputs.

### EF5 — Regime and spatial concentration of family membership

- **Type:** cross-scale — edit-family membership x local regime / page /
  line.
- **Definition:** test whether tokens belonging to the same edit family
  (post LP2 filtering) co-occur within the same line, page or Task65
  regime cluster more than a C-GLOBAL/C-PAGE/C-REGIME null predicts.
- **Level:** cross-scale (edit-family x line/local, edit-family x
  regime).
- **Estimator:** same co-occurrence-rate statistic as LP3's locality
  component (shared computation, reported once and referenced by both
  families rather than duplicated).
- **Input:** LP2-filtered family membership, line/page metadata, Task65
  regime labels.
- **Normalization:** co-occurrence rate normalized by family size and by
  the corpus base rate of same-line/same-page/same-regime co-occurrence
  for arbitrary token pairs.
- **Null/control:** C-GLOBAL, C-PAGE, C-REGIME.
- **Uncertainty:** bootstrap over families.
- **Missing-data behavior:** small families excluded from the rate
  estimate and separately counted.
- **Interpretation:** strong concentration argues for a scribal/
  production-local process (copying, local generation) rather than a
  manuscript-wide lexical paradigm; weak concentration is more
  compatible with a global paradigm system.
- **Known limitations:** page/regime size differences must be controlled
  for, which the normalization step above attempts but cannot fully
  guarantee.
- **Justification:** directly requested (parent task section 8:
  "spatial/local concentration", "regime-conditioned graph structure")
  and simultaneously the edit-family half of the cross-scale "edit
  family x regime" example in section 14; identical to LP3's locality
  computation and therefore implemented once, not twice; estimable given
  LP2 output; stability via family-level bootstrap.

---

## 6. Sequence structure

### SQ1 — Lag-dependent conditional token entropy curve

- **Type:** conditional — `H(X_i | X_{i-k})` for `k=1..K`.
- **Definition:** extends Task58's single adjacent-MI number to a curve
  over lag `k`, each corrected the same way (observed minus mean of
  within-line shuffles), with an explicit per-lag coverage report.
- **Level:** sequence (adjacent and longer).
- **Estimator:** same corrected-MI estimator as Task58, applied at each
  lag.
- **Input:** token stream with line boundaries.
- **Normalization:** same capped-entropy share convention as Task58, per
  lag.
- **Null/control:** C-LINE (100+ shuffles, Task58's own convention);
  C-NAT; C-PHASE1 homophony series.
- **Uncertainty:** shuffle-distribution-based, as Task58.
- **Missing-data behavior:** lags exceeding a line's typical length lose
  coverage rapidly; report per-lag effective sample size and mark low-
  coverage lags explicitly rather than extrapolating a smooth curve.
- **Interpretation:** a rapidly decaying curve is consistent with only
  short-range dependence (matching Task58/63's existing weak/small
  findings); a slowly decaying or non-monotonic curve would be new
  evidence for longer-range structure that the single-lag Task58 number
  cannot show.
- **Known limitations:** within-line-only shuffles bound the maximum
  informative lag to within-line distances; a separate within-page
  version would be needed for cross-line lag, which is deferred to
  SQ5/PG-level cross-scale work rather than duplicated here.
- **Justification:** directly requested (parent task section 9,
  "conditional token entropy," "lag dependency"); not redundant with
  Task58 because it is a curve, not a point estimate, and a curve can
  show structure a single lag misses even when lag-1 alone looks weak;
  estimable at short lags with existing line lengths; stability expected
  to degrade with lag, reported via effective sample size.

### SQ2 — Run-length distribution with tail test

- **Type:** marginal/joint — full distribution rather than Task60's raw
  counts.
- **Definition:** the full distribution of exact-repeat and near-repeat
  (distance<=1) run lengths, with an explicit tail test (probability of
  a run at least as long as the observed maximum) against C-GLOBAL and
  C-LINE nulls.
- **Level:** sequence / repetition.
- **Estimator:** direct run enumeration (already computed for Task60);
  tail probability via the shuffle null's empirical run-length
  distribution.
- **Input:** the adjacency data already used by Task60.
- **Normalization:** run-length counts reported per corpus length for
  cross-corpus comparability.
- **Null/control:** C-GLOBAL, C-LINE (Task60's own families, reused).
- **Uncertainty:** null-derived percentile for the tail statistic;
  bootstrap for the rest of the distribution.
- **Missing-data behavior:** none; well-defined for any corpus.
- **Interpretation:** Task60 already reports "maximum run 4, 11 runs>=3,
  one>=4" as bare counts; this reframes them with an explicit null-based
  tail probability so the numbers are comparable across corpora of
  different size.
- **Known limitations:** tail statistics are inherently noisy for rare
  events (a handful of long runs).
- **Justification:** directly requested ("run structure," parent task
  section 9); minor extension of Task60's existing counts into a
  distribution with a proper null-referenced tail test, not a new data
  collection effort; estimable and low-risk; stability limited only by
  the rarity of the tail event itself, which is reported, not hidden.

### SQ3 — Transition persistence (second-order near-repeat dependence)

- **Type:** conditional — `P(near-repeat at t+2 | near-repeat at t)` vs.
  the memoryless base rate.
- **Definition:** tests whether an adjacent near-repeat event at position
  `t` changes the probability of another near-repeat event at `t+1` or
  `t+2`, beyond what the base near-repeat rate alone predicts.
- **Level:** sequence.
- **Estimator:** conditional rate comparison with a C-GLOBAL/C-LINE
  shuffle null.
- **Input:** the near-repeat event stream from Task60/SQ2.
- **Normalization:** reported as an odds ratio relative to the base
  rate.
- **Null/control:** C-GLOBAL, C-LINE.
- **Uncertainty:** bootstrap over events.
- **Missing-data behavior:** none.
- **Interpretation:** persistence above the memoryless base rate would be
  compatible with "bursty" local copying/elaboration episodes rather
  than independent per-position near-repeat events.
- **Known limitations:** near-repeat events are already spatially
  clustered by line/page structure (Task64/65), so this must be
  evaluated conditional on regime (a cross-scale extension, not
  duplicated here) to avoid mistaking known local-regime clustering for
  a genuinely new transition-persistence effect.
- **Justification:** directly requested ("transition persistence,"
  parent task section 9); genuinely new (Task60/63 never tested second-
  order dependence between repeat events); estimable given the existing
  event stream; stability is the open question this metric is designed
  to answer, so no strong prior is assumed.

### SQ4 — Family-to-family transition

- **Type:** cross-scale — sequence x lexical family (also referenced as
  CS2/CS8 in the cross-scale table).
- **Definition:** using LP1-LP2's productive-rule families as token
  labels, build a family-level first-order transition matrix and test
  its deviation from a family-frequency-only (independence) null.
- **Level:** cross-scale.
- **Estimator:** categorical transition-matrix MI, same convention as
  Task58's token-identity MI but at the family label grain.
- **Input:** token stream with family labels (LP1-LP2 output) and line
  boundaries.
- **Normalization:** MI share, same convention as Task58.
- **Null/control:** C-LINE, C-GLOBAL.
- **Uncertainty:** shuffle-distribution-based, as Task58.
- **Missing-data behavior:** tokens with no assigned family (most of the
  vocabulary, since paradigms are expected to be a minority structure)
  are pooled into an explicit `NO_FAMILY` category, not dropped.
- **Interpretation:** if families transition non-randomly (e.g. a
  family tends to follow specific other families), that is evidence
  for higher-order organization above the individual-token level that
  Task58's token-identity MI could not detect because it is diluted by
  the whole vocabulary.
- **Known limitations:** depends entirely on LP1-LP2 family quality;
  a small number of paradigm families will limit statistical power.
- **Justification:** this is one estimator serving two purposes named
  separately in the parent task (a "sequence structure" bullet in
  section 9, "transformation sequence," "family-to-family transition,"
  and a cross-scale example in section 14), so it is defined once here
  and referenced, not duplicated, at the cross-scale level; not
  redundant with Task58 because token identity and family identity are
  different (coarser, paradigm-aware) alphabets; estimable once LP1-LP2
  exist; stability depends on family count and size, to be reported.

### SQ5 — Systematic longer-range conditional dependence sweep

- **Type:** conditional, a systematic replacement for Task63's
  frozen-candidate-only approach.
- **Definition:** for `n=2..5`, systematically test `P(token_{i+n} |
  token_i, ..., token_{i+n-1})` against the corresponding lower-order
  marginal, over **all** sufficiently frequent token n-grams (not a
  frozen candidate list), with FDR correction and an explicit minimum-
  count/power threshold reported per candidate.
- **Level:** longer sequence.
- **Estimator:** conditional permutation test (as used in the baseline
  higher-order stages) generalized to a systematic sweep instead of a
  frozen inventory.
- **Input:** token stream, line/block partitions for leave-one-block-out.
- **Normalization:** FDR-controlled q-values; minimum count threshold
  declared up front, not tuned post hoc.
- **Null/control:** frequency null and first-order Markov null (as
  Task63's baseline confirmation), leave-one-block-out, jackknife.
- **Uncertainty:** LOBO/jackknife as in the existing baseline stages.
- **Missing-data behavior:** candidates below the minimum-count threshold
  are excluded from the systematic sweep and reported as a count of
  excluded candidates, so the sweep's coverage is auditable.
- **Interpretation:** replaces the earlier conclusion "evidence for
  general higher-order rules is weak and candidate-specific" (which was
  itself limited by using a frozen, previously-discovered candidate
  list) with a systematic, power-reported statement about what a full
  sweep does or does not find.
- **Known limitations:** a full systematic sweep at `n=5` is
  combinatorially large; the minimum-count threshold will exclude most
  candidates, and that exclusion must be reported, not treated as
  negative evidence.
- **Justification:** directly requested ("longer-range conditional
  dependence," parent task section 9); not redundant with the existing
  Task63/baseline frozen-candidate work because it removes the
  candidate-selection bias that Phase I's own report flags as a
  limitation; estimable at `n=2,3` given corpus size, expected weak
  power at `n=4,5` (reported, not concealed); stability via LOBO as
  already validated in the baseline pipeline.

---

## 7. Local / line structure

Parent task section 10 explicitly forbids collapsing this into one "line
effect." The three families below are designed to be estimated jointly
so that physical-line, line-sized-locality, sub-page-regime, page-
composition, drift and discrete-state components each get their own
number.

### LL1 — Joint variance decomposition of near-form similarity

- **Type:** joint, a single decomposition rather than Task64's separate
  pairwise comparisons.
- **Definition:** decompose the near-form-similarity effect (Task64's
  measured quantity) into additive components attributable to: physical
  line membership, line-sized shifted-window membership, page
  membership, and residual, using a single model (e.g. a fixed-effects
  or ANOVA-style decomposition on the matched-pair rate) instead of
  reporting each scale comparison as an independent test as Task64 did.
- **Level:** line/local.
- **Estimator:** linear/ANOVA-style decomposition on the same
  length-matched pair data Task64 already built.
- **Input:** Task64's matched-pair dataset.
- **Normalization:** components reported as shares of total explained
  variance, summing to a reported total plus residual.
- **Null/control:** C-LINE, C-PAGE, C-GLOBAL (Task64's own families).
- **Uncertainty:** bootstrap over pairs, stratified by page.
- **Missing-data behavior:** none beyond what Task64 already handles.
- **Interpretation:** directly operationalizes the parent task's
  requirement to distinguish physical-line effect from broader local
  regime from page composition, as shares of one decomposition instead
  of a set of separate binary comparisons whose relationship to each
  other is only narrative.
- **Known limitations:** additive decomposition assumes the components
  do not interact strongly; an interaction term should be checked and
  reported if the additive model fits poorly.
- **Justification:** consolidates Task64's already-strong pairwise
  results into the joint form the parent task asks for, without
  re-collecting data; estimable directly from existing artifacts;
  stability via page-stratified bootstrap.

### LL2 — Page composition and regime-boundary alignment

- **Type:** cross-scale — page layout x regime (also referenced as CS10).
- **Definition:** compute page composition statistics (lines per page,
  tokens per line distribution, paragraph-locus count per page) and test
  whether Task65 regime-cluster boundaries align with locus-type changes
  (paragraph-to-label, paragraph-to-radial/circular) more than a
  position-shuffled null predicts.
- **Level:** cross-scale (page/2D x regime).
- **Estimator:** boundary-alignment test (distance between nearest
  regime-change point and nearest locus-type-change point, compared to
  the same statistic under randomized regime-change positions).
- **Input:** Task65 regime-cluster labels, IVTFF locus-type sequence.
- **Normalization:** alignment distance normalized by average
  inter-boundary spacing.
- **Null/control:** position-shuffled regime-boundary null (a C-POS
  variant specific to boundary positions, documented in
  [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md)).
- **Uncertainty:** permutation-based.
- **Missing-data behavior:** pages with a single locus type have no
  internal boundary to test and are excluded, counted separately.
- **Interpretation:** alignment would suggest regimes are partly a
  layout/genre artifact (paragraph vs. label vs. circular text having
  different local statistics almost by definition); no alignment would
  argue regimes reflect something orthogonal to the catalogued layout
  categories.
- **Known limitations:** locus-type changes are themselves a coarse,
  cataloguer-assigned partition, not a physical measurement.
- **Justification:** directly requested (parent task section 10, "page
  composition"; section 14 cross-scale example list is explicitly
  non-exhaustive, and this fills one clear gap: regime x layout); not
  redundant with Task65 (which conditioned on Currier/hand metadata, not
  on locus-type or page composition); estimable using already-parsed
  locus-type sequence and existing regime labels; stability via
  permutation testing.

### LL3 — Quantitative drift-vs-state decomposition

- **Type:** conditional, turning Task65's qualitative
  `MIXED_DRIFT_AND_STATES` label into a quantitative, comparable number.
- **Definition:** using the same lag-decay and change-point machinery as
  Task65, report the fraction of total local-similarity variance
  attributable to smooth drift (captured by a smoothly varying baseline)
  versus discrete state-switches (captured by the change-point/cluster
  model), as one number per corpus, instead of a qualitative label only.
- **Level:** regime / local.
- **Estimator:** variance attributed to each component of Task65's
  existing model, reported as an explicit split rather than a
  categorical topology label.
- **Input:** Task65's existing lag/cluster/change-point outputs.
- **Normalization:** shares summing to 1 (drift share + state share +
  residual).
- **Null/control:** the same stationary/smooth/discrete/mixed synthetic
  controls Task65 already uses, since each control corresponds to a
  known ground-truth split that calibrates the decomposition.
- **Uncertainty:** bootstrap over folio blocks.
- **Missing-data behavior:** none beyond Task65's existing handling.
- **Interpretation:** makes the "mixed drift and states" finding
  comparable across corpora, sections, and — later — candidate
  generative models, rather than a label that cannot be compared
  quantitatively.
- **Known limitations:** the drift/state split depends on the smoothing
  bandwidth and change-point sensitivity already chosen in Task65;
  those choices must be declared and held fixed.
- **Justification:** directly requested (parent task section 10:
  "slow drift," "discrete states," treated as distinguishable
  quantities); reuses Task65's existing model rather than fitting a new
  one, avoiding redundant computation; estimable directly; stability
  calibrated against Task65's own synthetic ground-truth controls.

---

## 8. Page / 2D structure and illustration-related metrics

Grounded in the fact check performed for
[FINGERPRINT_V2_COVERAGE.tsv](FINGERPRINT_V2_COVERAGE.tsv): IVTFF header
variables (`$Q` quire, `$P` page-in-quire, `$F` foliation, `$B` bifolio,
`$I` illustration/content-type code, `$H` hand, `$L` Currier language,
`$X` text-block-position code) and locus-type letters (`P` paragraph,
`L` label, `R` radial/circular, `C` circular text) are machine-parseable
today via `internal/metadatavalidation.ParseIVTFF`, and `IndexInLine`/
`IndexInFolio` are already first-class fields on `TokenMetadata`. True
pixel/vector coordinates and image-region content are **not** available
and are explicitly out of scope (parent task section 12: no manual
subjective coding of image content).

### PG1 — Locus-type stratified fingerprint

- **Type:** conditional — recompute a small, declared subset of existing
  marginal/joint metrics (glyph conditional entropy, edit-adjacency
  rate, positional glyph shares) separately by locus type
  (paragraph/label/radial-circular) instead of pooled.
- **Level:** page/2D, operationalizing "label/running-text status"
  (parent task section 11) directly.
- **Estimator:** the existing Task59/60/61 estimators, re-run per
  stratum.
- **Input:** IVTFF locus-type per token (already parsed).
- **Normalization:** none beyond the underlying metrics' own convention.
- **Null/control:** the same control families as the underlying metric
  (C-WITHINTOKEN for entropy, C-LINE/C-GLOBAL for edit-adjacency), plus
  a locus-type-shuffle null to test whether the stratified difference
  itself exceeds chance.
- **Uncertainty:** per-stratum bootstrap; the label stratum is small
  (Task60 already reports only 170 label near-repeat pairs) and is
  expected to be low-power, which must be reported.
- **Missing-data behavior:** strata below minimum count are marked
  `INSUFFICIENT_SUPPORT`, most likely to apply to radial/circular text
  given its rarity.
- **Interpretation:** if labels and running text differ systematically,
  pooled fingerprint numbers conflate two distinct regimes; if they do
  not differ, pooling is justified for those metrics.
- **Known limitations:** low label/radial sample size limits power, as
  already flagged for Task60's label subset.
- **Justification:** reuses existing estimators purely as a
  stratification exercise (no new estimator, only new conditioning),
  so redundancy risk is by design; directly answers the parent task's
  named "label/running text" question; estimable now; stability
  explicitly expected to be poor for the rare strata, which will be
  reported rather than glossed over.

### PG2 — Intra-line token-ordinal position effect

- **Type:** conditional — property | `IndexInLine`.
- **Definition:** test whether a token's ordinal position within its
  line (already parsed, used here as an explicit proxy for horizontal
  position since true coordinates are unavailable) predicts token
  length, lexical-family membership (LP1-LP2), or edit-adjacency
  involvement, beyond what line-membership alone predicts.
- **Level:** page/2D.
- **Estimator:** conditional-effect regression/tabulation against
  `IndexInLine`, controlling for line length.
- **Input:** `TokenMetadata.IndexInLine` (already parsed).
- **Normalization:** `IndexInLine` normalized by line length (as in G2)
  so lines of different length are comparable.
- **Null/control:** C-LINE (shuffle tokens within a line, destroying
  ordinal position while preserving line membership and content).
- **Uncertainty:** bootstrap over lines.
- **Missing-data behavior:** none.
- **Interpretation:** an effect here would be new evidence that "line
  position" carries information beyond adjacency-to-neighbors (already
  covered by SQ1/Task63); no effect narrows what "local structure" can
  mean to purely relational (neighbor-based), not positional.
- **Known limitations:** this is transcription-order position, not
  physical millimeters on the page; it must never be reported as a
  spatial/geometric coordinate without this caveat.
- **Justification:** directly requested ("token horizontal position,"
  parent task section 11), implemented as the best available proxy
  given documented data limits rather than invented or estimated from
  images; not redundant with SQ1 (relational, not positional); estimable
  now; stability via line-stratified bootstrap.

### PG3 — Page-boundary effect

- **Type:** conditional — property | distance to nearest folio boundary.
- **Definition:** extends Task64's page-scale comparison into an
  explicit boundary-alignment test: do local-similarity or entropy
  statistics show a discontinuity precisely at folio boundaries, versus
  matched interior points at the same within-page distance.
- **Level:** page/2D.
- **Estimator:** boundary-vs-interior matched comparison, reusing
  Task64's matched-pair machinery with folio boundary as the boundary
  type instead of line boundary.
- **Input:** folio identifiers (already parsed) plus the metrics from
  Task64/65 already computed near line/page edges.
- **Normalization:** distance-to-boundary normalized by page length in
  lines.
- **Null/control:** C-PAGE, matched interior points (Task64's own
  convention, applied to a new boundary type).
- **Uncertainty:** bootstrap over boundaries.
- **Missing-data behavior:** none; every folio has exactly one starting
  boundary.
- **Interpretation:** a genuine page-boundary discontinuity (beyond
  what within-page distance alone predicts) would support page as a
  meaningful production/copying unit; its absence would argue current
  page-level findings (e.g. Task65's within-page persistence) are better
  explained by within-page distance than by the physical page edge
  itself.
- **Known limitations:** folio recto/verso and quire-gathering order
  effects (bifolio structure, `$B`) could confound a naive boundary test
  and should be reported as a covariate, not ignored.
- **Justification:** directly requested ("page boundary," parent task
  section 11); extends rather than duplicates Task64's existing
  matched-pair method; estimable now; stability via boundary-level
  bootstrap.

### PG4 — Content-type ($I) stratified fingerprint

- **Type:** conditional — core fingerprint metrics | `$I` catalogue code.
- **Definition:** cross-tabulate a declared subset of core marginal/
  joint metrics by the IVTFF `$I` header value (the cataloguer's
  content-type code per page, values observed in-repo: A, B, C, H, P, S,
  T, Z), strictly as a categorical covariate.
- **Level:** page/2D, satisfying parent task section 12 to the extent
  reliable data exists.
- **Estimator:** the same estimators as PG1, restratified by `$I`
  instead of locus type.
- **Input:** `$I` values, currently reachable via
  `Document.PageVariables` and not yet surfaced on `TokenMetadata`
  (implementation note carried into
  [FINGERPRINT_V2_IMPLEMENTATION.tsv](FINGERPRINT_V2_IMPLEMENTATION.tsv)).
- **Normalization:** none beyond the underlying metrics.
- **Null/control:** same as PG1, plus an `$I`-shuffle null.
- **Uncertainty:** per-stratum bootstrap.
- **Missing-data behavior:** pages without a recorded `$I` value (some
  header lines omit it, as observed directly in `data/ZL3b-n.txt`) are
  pooled into an explicit `UNCODED` stratum, never dropped silently.
- **Interpretation:** this measures association with the cataloguer's
  content-type code, not with any claim about what the image depicts;
  any difference found is a layout/genre-category association, and must
  be reported as such, never as an image-content finding.
- **Known limitations:** `$I` is an external cataloguing convention, not
  a measurement Phase I or Phase II performed; its granularity and
  criteria are outside this project's control.
- **Justification:** the only defensible way to satisfy parent task
  section 12 ("illustration-related metrics") without subjective manual
  image coding, per the parent task's own prohibition; not redundant
  with PG1 (a different, orthogonal covariate); estimable once `$I` is
  surfaced (a minor extension, not new data collection); stability
  expected to be uneven across the eight `$I` categories by sample size,
  to be reported per stratum.

**Explicitly NOT_AVAILABLE for section 12 beyond PG1/PG4:** pixel/vector
image coordinates, diagram or figure bounding boxes, fine-grained
spatial grouping of text around specific image elements, and any
semantic image-content label. No metric family is specified for these;
inventing one would require exactly the subjective coding the parent
task prohibits.

---

## 9. Hierarchical structure

### HR1 — Nested variance decomposition across the hierarchy

- **Type:** cross-scale, a method template rather than a single fixed
  statistic.
- **Definition:** for a short list of already-defined scalar properties
  (glyph conditional entropy from G4, edit-adjacency rate from Task60/
  EF-level, positional-glyph-share deviation from G2), decompose their
  variance across the nested hierarchy glyph -> token -> lexical family
  (LP1-LP2) -> local regime (Task65) -> page -> section/manuscript
  (Currier/hand/quire), using a hierarchical/mixed-model-style variance
  split (e.g. nested ANOVA or variance-components estimation) rather
  than the separate single-level tests Phase I already ran.
- **Level:** cross-scale / hierarchy.
- **Estimator:** variance-components estimation on each chosen base
  scalar, with the hierarchy levels as nested grouping factors.
- **Input:** the base scalar's own per-unit values plus every level's
  membership label (family, regime, page, section).
- **Normalization:** variance shares reported per level, summing to 1
  plus residual.
- **Null/control:** a fully-crossed-random-assignment null (randomly
  reassign units to hierarchy groups of the same sizes) to establish the
  variance-share baseline expected under no real hierarchical structure.
- **Uncertainty:** bootstrap over top-level units (pages/sections) to
  respect the nesting structure.
- **Missing-data behavior:** a base scalar unavailable at a given level
  (e.g. a page with too few tokens) is excluded from that level's
  estimate and the exclusion count is reported.
- **Interpretation:** directly answers the parent task's stated main
  question for this section — "at what level does the main structure
  arise?" — as a single, comparable variance-share table instead of a
  set of separately-run single-level significance tests whose relative
  magnitude cannot otherwise be compared.
- **Known limitations:** variance-components estimation assumes a
  reasonably balanced or at least well-behaved nesting structure; highly
  unbalanced group sizes (e.g. very few Currier-B sections) will widen
  uncertainty at that level, which must be reported rather than
  smoothed over.
- **Justification:** consolidates a hierarchy-wide question that would
  otherwise require re-deriving comparability across every separate
  single-level test in this document; is not redundant with any single
  family above because none of them attempt a joint multi-level split;
  estimable once the constituent scalars and labels exist (all defined
  above); stability the central subject of the metric, reported via
  bootstrap rather than assumed.

---

## 10. Cross-scale dependencies

Per parent task section 14, this is one of the two primary targets of
fingerprint v2 (with lexical paradigms). The table below lists every
cross-scale pair specified in this document, whether newly defined here
or already defined above and simply cross-referenced, to avoid
duplicating estimators.

| ID | Pair | Status |
|---|---|---|
| CS1 | glyph structure x lexical family | new, defined below |
| CS2 | token family x adjacency | = SQ4 |
| CS3 | token formation x line position | new, defined below |
| CS4 | edit family x local regime | = EF5 |
| CS5 | repetition x page position | new, defined below |
| CS6 | entropy x regime | = G4 |
| CS7 | lexical structure x metadata | new, defined below |
| CS8 | sequence structure x lexical family | = SQ4 |
| CS9 | hierarchy-wide variance attribution | = HR1 |
| CS10 | page/2D x regime (boundary alignment) | = LL2 |

### CS1 — Glyph structure x lexical family

- **Type:** cross-scale, joint.
- **Definition:** test whether G2's positional-glyph-share profile
  differs systematically across LP1-LP2's productive-rule families,
  versus a null where family assignment is shuffled.
- **Level:** cross-scale (glyph x lexical).
- **Estimator:** per-family G2 profile comparison (e.g. a
  distributional-distance statistic between per-family profiles and the
  pooled profile).
- **Input:** G2 output, LP1-LP2 family labels.
- **Normalization:** as G2.
- **Null/control:** family-label shuffle (a C-POS-family variant),
  C-GRAMMAR.
- **Uncertainty:** bootstrap over families.
- **Missing-data behavior:** small families excluded and counted.
- **Interpretation:** family-specific glyph-position profiles would
  argue paradigm families are themselves partly a glyph-position
  phenomenon (e.g. affixes reusing specialized glyphs), not an
  independent lexical-only structure.
- **Known limitations:** depends on LP1-LP2 family quality.
- **Justification:** fills a named example (parent task section 14);
  reuses G2 and LP1-LP2 estimators without modification, adding only
  the joint comparison; estimable once LP1-LP2 exist; stability via
  family-level bootstrap.

### CS3 — Token formation x line position

- **Type:** cross-scale, conditional. (Already introduced as one of
  TF2's conditioning axes; listed here only for completeness of the
  cross-scale table, not redefined.)

### CS5 — Repetition x page position

- **Type:** cross-scale, conditional.
- **Definition:** test whether near/exact repetition rate (Task60) varies
  systematically with PG2 (intra-line ordinal position) or PG3 (distance
  to page boundary), beyond what line/page membership alone (already
  tested by Task64) predicts.
- **Level:** cross-scale (repetition x page/2D).
- **Estimator:** conditional-rate regression against PG2/PG3 covariates.
- **Input:** Task60 near-repeat event stream, PG2/PG3 covariates.
- **Normalization:** rate normalized as in Task60.
- **Null/control:** C-LINE, C-PAGE (as Task60), plus the PG2/PG3
  covariate shuffles.
- **Uncertainty:** bootstrap over events.
- **Missing-data behavior:** none beyond PG2/PG3's own handling.
- **Interpretation:** a positional gradient in repetition (e.g. more
  repetition near line starts or page edges) would connect the known
  edit-operation BEGIN-bias (Task60) to a page-level spatial pattern
  rather than a purely token-internal one.
- **Known limitations:** correlated with known line/page effects already
  measured; must be reported as incremental over those, not instead of
  them.
- **Justification:** fills a named example (parent task section 14);
  reuses Task60/PG2/PG3 estimators; estimable now; stability via
  event-level bootstrap.

### CS7 — Lexical structure x metadata

- **Type:** cross-scale, joint.
- **Definition:** test whether LP1-LP2's productive-rule family
  distribution differs across Currier A/B, hand, or quire, versus a
  metadata-shuffle null.
- **Level:** cross-scale (lexical x manuscript-global metadata).
- **Estimator:** categorical association test (chi-square/MI) between
  family label and metadata label.
- **Input:** LP1-LP2 family labels, Currier/hand/quire metadata (already
  parsed).
- **Normalization:** MI share convention, as elsewhere in this document.
- **Null/control:** metadata-label shuffle, C-GRAMMAR.
- **Uncertainty:** permutation-based.
- **Missing-data behavior:** unlabeled metadata pooled into an explicit
  `UNKNOWN` category.
- **Interpretation:** directly informs section 16's global-vs-
  heterogeneity fingerprint split for the lexical level specifically:
  paradigm families tied to Currier/hand/quire would argue for a
  heterogeneity fingerprint at the lexical level; families spread evenly
  would argue lexical paradigms are closer to a manuscript-wide
  constant.
- **Known limitations:** depends on LP1-LP2 quality; Currier/hand/quire
  partitions are themselves uneven in size (Currier B and later hands
  are smaller), limiting power for rarer strata.
- **Justification:** fills a named example (parent task section 14) and
  feeds section 16 directly; reuses existing metadata parsing and
  LP1-LP2 output; estimable once LP1-LP2 exist; stability limited by
  the smaller metadata strata, to be reported.

---

## 11. Compression / predictability

Per parent task section 15, no raw compressor output is used as an
unexplained metric; each family below is tied to a specific structural
interpretation already established elsewhere in this document.

### CP1 — Normalized compression ratio as a higher-order-structure check

- **Type:** marginal, used only as a cross-check against G1/G4.
- **Definition:** compression ratio of the glyph stream (per Task61's
  three representation modes) under one fixed, declared general-purpose
  compressor and setting, reported alongside — never instead of — the
  bigram/trigram entropy estimates from G1/G4, specifically to see
  whether the compressor recovers redundancy the low-order plug-in
  entropy model misses.
- **Level:** glyph / sequence.
- **Estimator:** compressed-size / raw-size ratio, fixed compressor and
  settings declared once and reused for every corpus and control.
- **Input:** the same stream representations as Task61/G1.
- **Normalization:** ratio relative to the same compressor's ratio on
  C-NAT and C-PHASE1 controls (a compressor's baseline ratio is not
  comparable across corpora without this).
- **Null/control:** C-WITHINTOKEN, C-GLOBAL, C-NAT, C-PHASE1 (same
  families as Task61, applied to compression instead of plug-in
  entropy).
- **Uncertainty:** block bootstrap.
- **Missing-data behavior:** none.
- **Interpretation:** if the compression ratio tracks the bigram entropy
  estimate closely, the compressor is not finding structure beyond what
  is already measured; if compression finds substantially more
  redundancy than the bigram/trigram model, that gap is itself the
  finding — evidence of higher-order or longer-range structure not
  captured by G1/G4 — and must be reported as exactly that gap, not as
  an independent "low entropy" claim.
- **Known limitations:** general-purpose compressors are tuned for
  natural-language-like statistics and byte-level patterns, not
  necessarily for a small custom glyph alphabet; results must be
  interpreted relative to the same compressor's behavior on controls,
  never as an absolute complexity measure.
- **Justification:** directly requested ("compression ratio," parent
  task section 15) with the explicit non-magic-number constraint
  satisfied by tying it to G1/G4; not redundant with G1/G4 because it
  can catch structure they miss (that gap is the entire point);
  estimable trivially; stability via block bootstrap.

### CP2 — Cross-compression / relative predictive distance

- **Type:** conditional, an asymmetric cross-corpus measure.
- **Definition:** using a compression-based relative-distance measure
  (e.g. compress corpus A using a dictionary/model trained on corpus B,
  and vice versa) between Voynich and each natural-language/synthetic
  control, interpreted explicitly as "how well does B's sequence
  structure predict A," tying it to SQ1/SQ5's sequence-dependency
  findings rather than treated as a freestanding similarity score.
- **Level:** sequence / cross-corpus.
- **Estimator:** a declared, fixed cross-compression or normalized-
  compression-distance algorithm, applied symmetrically to every corpus
  pair in the comparison set.
- **Input:** token or glyph streams per corpus.
- **Normalization:** self-compression baseline per corpus (A predicting
  A) reported alongside cross-corpus values so asymmetry is
  interpretable.
- **Null/control:** C-NAT, C-PHASE1, C-GLOBAL (shuffled version of each
  corpus as a lower bound on achievable cross-prediction).
- **Uncertainty:** block bootstrap per corpus.
- **Missing-data behavior:** none.
- **Interpretation:** this becomes one of the entries in the corpus-
  comparison distance toolkit
  ([FINGERPRINT_V2_DISTANCE.md](FINGERPRINT_V2_DISTANCE.md)), not a
  standalone fingerprint row; its value is explicitly framed as
  "consistent with SQ1/SQ5 findings" or "in tension with them," never
  interpreted alone.
- **Known limitations:** compression-based distances can be sensitive to
  corpus length differences; must be reported alongside a length-
  matched control.
- **Justification:** directly requested ("cross-compression," parent
  task section 15); explicitly not used as a magic standalone number,
  satisfying the parent constraint; estimable with standard tooling;
  stability via block bootstrap and length-matched controls.

### CP3 — Consolidated held-out predictive likelihood report

- **Type:** conditional, a synthesis rather than a new estimator.
- **Definition:** report TF1/TF2's held-out formation cross-entropy and
  SQ1/SQ5's held-out sequence conditional-entropy together, on a common
  bits/token basis, as the fingerprint's single "predictive coding"
  entry, rather than introducing a new opaque predictive-likelihood
  metric.
- **Level:** cross-scale (token formation x sequence).
- **Estimator:** none beyond TF1/TF2/SQ1/SQ5's own estimators; this
  entry is a reporting convention.
- **Input:** TF1/TF2/SQ1/SQ5 outputs.
- **Normalization:** common bits/token basis.
- **Null/control:** inherited.
- **Uncertainty:** inherited.
- **Missing-data behavior:** inherited.
- **Interpretation:** directly operationalizes "predictive coding" and
  "held-out predictive likelihood" (parent task section 15) without a
  redundant new model.
- **Known limitations:** inherited from its constituent metrics.
- **Justification:** avoids introducing a fourth, redundant predictive-
  likelihood estimator when three already exist in this document;
  purely a synthesis, so redundancy/estimability/stability are
  inherited by construction.

---

## 12. Summary count and redundancy statement

37 metric families are specified above (4 glyph, 3 token-formation, 4
lexical-paradigm, 5 edit-family, 5 sequence, 3 local/line, 4 page/2D, 1
hierarchy template, 5 cross-scale-only, 3 compression). Of the 10
cross-scale pairs named in the section 10 table, 6 (CS2, CS4, CS6, CS8,
CS9, CS10) are explicitly cross-referenced to a family defined elsewhere
rather than redefined, and 4 (CS1, CS3, CS5, CS7) are new joint/
conditional estimators specific to the cross-scale pairing. No metric
family duplicates another family's estimator;
every case of apparent overlap (glyph/token-formation prefix-suffix
handling, edit-family/lexical-paradigm graph structure, sequence/
cross-scale family transitions, local-line/page-2D boundary effects) is
resolved explicitly in the relevant section's text by assigning the
estimator to exactly one family and cross-referencing it from the
other(s). Formal correlation-based redundancy checking across the
*implemented* metrics (as opposed to this design-time avoidance of
duplicate estimators) is a stability/redundancy-analysis activity
specified in
[FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 3, to be
run once metrics are implemented, not at design time.
