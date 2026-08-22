# Inverse Homophony Recovery — Design (Task57)

Written before any production search code, per task57 section 37. Covers
threat model, synthetic split, features, similarity estimators,
clustering/search method, anti-collapse constraints, baselines, validation
gate, Voynich protocol, and interpretation limits.

## 0. Scope decision (read this first)

Task57 is large enough that a literal reuse of every Stage 3/9/10/11/13/14/27
package for the *synthetic validation loop* is not tractable in-session: a
single full generic-corpus pipeline run over a Doyle-scale corpus takes
~37 minutes (observed: `experiments/doyle__homophonic__h004__uniform__seed001-v1`,
09:43:52→10:20:33), and Phase A needs dozens of corpus/baseline combinations.
Running the full pipeline dozens of times is a multi-day background job, not
a validation loop.

Resolution, fixed here before any validation numbers exist:

- **Phase A (synthetic, many repeated evaluations)** uses two genuinely
  reused production functions (`vocabularygrowth.Analyze`,
  `sequenceanalyze.AnalyzeLines`, both pure functions over token slices —
  see §7) plus one **self-contained lightweight analogue** of the
  transition/backbone estimand (§7.3), documented as an analogue, not a
  disguised duplicate of Stage 27's code. Stage 27 itself is Config/executor
  coupled and permutation-heavy — disproportionate to an inner validation
  loop that must run unattended, many times, deterministically.
- **Phase B (one preregistered Voynich run + matched nulls)** is small in
  run-count (one recovered partition + a fixed number of matched-null
  partitions), so it uses the real Stage 23-28 CLIs via
  `pipeline-orchestrate -generic-corpus`, satisfying task57 §25/§31 for the
  claim that actually gets published.

This is a scope decision made for engineering-feasibility reasons, fixed
before touching any synthetic corpus with the algorithm, and documented so
it cannot be quietly changed after seeing results.

## 1. Threat model

Ciphertext is an opaque token sequence. The recovery engine receives:

- the token sequence (or a token+natural-line-boundary structure, same as
  `genericsegmentation.ReadCorpus` already produces for every other stage);
- nothing else.

It must never receive: plaintext, the Task46/55 mapping/allocation TSV, the
source filename (which encodes H/model in dev/validation corpora), or any
metadata that reveals plaintext identity.

Families covered (task57 §2): `global-H + uniform`, `global-H +
triangular-v1 (weighted)`, `frequency-v1 + uniform`.

## 2. Opaque relabeling (anti-leakage safeguard, §5/§8)

Task46's cipher tokens are already opaque (`x000001` format, assigned by
`corpustransform.opaqueID` in **sorted plaintext-token order** — the
assignment order itself leaks nothing since recovery never sees plaintext).
But Voynich's canonical tokens are real EVA transliteration strings, and a
generic implementation must not assume its input already looks like
`x000001`. So the recovery engine always applies its own relabeling step,
uniformly, to any input (synthetic or Voynich):

    relabel(tokens) = deterministic map: sorted-distinct-token order -> x%06d

This makes graphemic similarity between relabeled tokens meaningless by
construction (§8): `x000001` vs `x000002` have no edit-distance content.
Graphemic features are therefore never computed post-relabel — they are
excluded from the feature set entirely (§7), not merely down-weighted.

The relabeling map itself is evaluator-only. The recovery engine's output
(`ciphertext_token -> latent_class`) is expressed in relabeled IDs; the
evaluator translates back to original tokens only when writing
human-facing artifacts (`voy_to_latent.tsv` etc.).

## 3. Features (§7)

Computed only from the relabeled ciphertext token stream + natural line
boundaries (`genericsegmentation.ReadCorpus`). Per distinct token:

1. **Frequency** — raw occurrence count.
2. **Predecessor / successor distribution** — counts of the immediately
   preceding / following relabeled token (distance 1, one side each).
3. **Distance-context profile** — counts of relabeled tokens at distances
   d∈{2,3,4,5} on either side (fixed, corpus-size-independent, decided here
   before any corpus is scored).
4. **Positional distribution** — histogram over 5 fixed buckets of
   `index_in_line / line_length` (natural line boundaries only, no IVTFF
   hand/Currier metadata).

Explicitly **not used**: graphemic similarity (forbidden post-relabel, §8);
local-regime profile, transition-network profile (candidates in §7, but
their production estimands are Config/executor-coupled and their marginal
signal over predecessor/successor + distance-context is small for the
window sizes at stake here — a scope choice, not a finding, recorded before
validation).

Each distribution is L1-normalized before comparison.

## 4. Similarity / merge evidence (§18)

For two relabeled cipher types a, b, four component scores, each
`1 - JensenShannonDivergence(dist_a, dist_b)` (bounded [0,1], symmetric,
handles disjoint support), for: predecessor, successor, distance-context
(concatenated d=2..5 profile), positional. Combined score is the
unweighted mean of the four components — equal weights fixed here, before
any discrimination number is computed, specifically to avoid the appearance
of hand-tuning weights to any corpus.

A pair is only scored (added as a graph edge) if it has at least
`minSupport = 5` combined predecessor+successor observations each — an
evidence floor, not a similarity threshold, to avoid noise from very rare
types dominating the graph.

## 5. Null test for merges — early stop gate (§19)

Before building any clustering search: for each *development* corpus with a
known oracle mapping, draw the true-homophone-pair score distribution and a
matched-degree false-pair (random cipher type pairs, excluded from being
true homophones) score distribution, and compute AUC (true pairs scored
higher). If AUC is not meaningfully above 0.5 on development corpora, STOP
— no clustering search is built, and Task57's result is the honest
insufficiency finding of §21. This diagnostic is `pair_discrimination.tsv`
+ the `diagnose` mode of the CLI, and it is run and inspected *before*
writing the clustering code.

## 6. Clustering / search (§17)

Simple, interpretable, deterministic:

1. Build the similarity graph (§4) over all cipher types meeting the
   support floor.
2. Sort edges by descending combined score.
3. Greedy union-find merge: process edges in that fixed order; merge the
   two endpoints' current classes iff:
   - score > τ (frozen threshold, chosen on development data only as the
     score maximizing Youden's J on the true/false pair diagnostic, §5);
   - the merged class's occurrence-fraction of the whole corpus would stay
     ≤ `maxClassFraction = 0.15` (frozen, anti-trivial-collapse, §15);
   - merging does not lower the Shannon entropy of the class-size
     distribution (weighted by occurrence count) below
     `minEntropyFraction = 0.5` of the entropy of the current NO_COLLAPSE
     partition (also frozen, §15).
4. Stop when the edge list is exhausted. No iterative re-scoring, no
   simulated annealing, no ML optimizer (§17).

τ, `maxClassFraction`, `minEntropyFraction`, `minSupport` are all fixed on
development corpora before validation corpora are scored (§16, §20), and
never adjusted afterward.

Every merge is logged with (a, b, score, component scores, evidence
support, running class-size / entropy state) — the audit trail required by
§17, written as `merge_audit.tsv`.

No plaintext vocabulary size, no oracle H, no allocation model is given to
this step (§16): unknown-H handling is entirely evidence/threshold driven,
not a compression-target parameter.

## 7. Evaluation

### 7.1 Class recovery (§10)

Computed from the predicted-partition × oracle-partition contingency table
(not O(n²) pairwise enumeration — the standard combinatorial identities for
pairwise precision/recall/F1, Adjusted Rand Index, and normalized mutual
information are all derivable from the contingency table in
O(#predicted classes × #oracle classes) time). Oracle partition access is
evaluator-only.

### 7.2 Structural recovery (§11/§12)

Three corpus states compared: plaintext P, ciphertext H(P), collapsed
R(H(P)). For each metric m: `Δcipher = m(H(P)) - m(P)`, `Δrecover =
m(R(H(P))) - m(P)`, and `recovery_fraction = 1 - |Δrecover|/|Δcipher|`
(only when `|Δcipher|` is not ≈0).

Families:

- **Vocabulary** — `vocabularygrowth.Analyze` (reused production function):
  type count, hapax fraction, dis-legomena fraction, Heaps β.
- **Sequence** — `sequenceanalyze.AnalyzeLines` (reused production
  function): repeated n-gram counts/fractions, extension structure.
- **Transition** — self-contained bigram-overrepresentation estimator
  (§0): for every observed bigram (a,b), a G-test statistic against the
  independence expectation `freq(a)*freq(b)/N`; "significant bigram
  fraction" = fraction of distinct bigrams exceeding the χ² critical value
  at a fixed α=0.01 (no multiple-comparison tuning beyond this fixed α).
  This is the primary analogue of Stage 27's backbone/significant-relation
  estimand for the validation loop; §25 uses the real Stage 27 for the one
  Voynich run.
- **Relation** — per task55's own finding ("relation family имеет optimum
  примерно около H≈4", i.e. non-monotonic in H), the forward effect is not
  stable, so per §13 this family is EXPLORATORY only: reported, never used
  as a gate criterion.

### 7.3 Directional predictions (§13, fixed before validation)

- Vocabulary: forward ↑ diversity ⇒ inverse predicts ↓ (V, hapax fraction
  decrease toward P).
- Transition: forward ↓ structure ⇒ inverse predicts ↑ (significant-bigram
  fraction increases toward P).
- Sequence: forward direction is corpus-dependent in prior work (task55:
  "sequence organization Voynich этими моделями не воспроизводится" is a
  level statement, not a direction one) — treated as **exploratory**
  unless the actual forward Doyle experiments (already-computed, existing
  `experiments/doyle__homophonic__*-v1` runs) show a stable sign, checked
  once during development and fixed here.
- Relation: exploratory (see above).

## 8. Baselines (§14)

- **NO_COLLAPSE** — identity partition.
- **FREQUENCY-ONLY** — deterministic `class = floor(log2(freq+1))`, no
  context features at all.
- **RANDOM_PARTITION** — uniform random assignment of cipher types into
  slots reproducing the recovered partition's exact class-size multiset,
  seeded deterministically (`subRand`-style derivation from a fixed purpose
  string + corpus id).
- **ORACLE** — the real Task46/55 mapping. Evaluator-only, synthetic-only
  (never available for Voynich).

## 9. Development / validation split (§4, fixed before final evaluation)

- **DEVELOPMENT** (used to fix τ, maxClassFraction, minEntropyFraction,
  minSupport, and to run the early-stop AUC check): Doyle H4 uniform, Doyle
  H4 triangular-v1.
- **VALIDATION** (scored once, after all thresholds are frozen): Doyle
  H6/H8 uniform, Doyle H6/H8 triangular-v1, Doyle frequency-v1 Hmax4/6/8
  uniform, Longfellow homophonic (unseen genre), Astafiev homophonic
  (unseen genre).

No validation corpus is read by any tuning step.

## 10. Validation gate (§20)

Pass requires all of:

1. class recovery (pairwise F1 and ARI) beats RANDOM_PARTITION with the
   same class-size distribution, on validation corpora;
2. structural recovery (vocabulary + transition families) beats both
   NO_COLLAPSE and RANDOM_PARTITION on validation corpora;
3. improvement holds on at least one non-Doyle validation corpus
   (Longfellow or Astafiev), not only on Doyle;
4. vocabulary and transition families show the predicted-direction
   `Δrecover` sign on validation corpora;
5. `recovered vocabulary size / NO_COLLAPSE vocabulary size` is not driven
   to a degenerate single-digit count (anti-trivial-collapse, §15).

No new numeric thresholds are invented post-hoc; "beats" means the
recovered value falls strictly on the plaintext side of the matched
RANDOM_PARTITION distribution (10 seeds), i.e. a relative criterion against
a matched null, per §20's fallback rule.

## 11. Voynich protocol (§23-§28) and interpretation limits (§29-30)

Exactly one preregistered run over `data_work/ZL3b-x7.canonical.txt`
(SHA256 recorded in the manifest), frozen method, no metadata. Outputs
`voy_to_latent.tsv`, `voynich_collapsed.txt`, `merge_audit.tsv`. Before/after
comparison uses the real Stage 23-28 CLIs via `pipeline-orchestrate
-generic-corpus` (§0), plus a matched-null-partition sweep with the same
class-count/class-size distribution. Task52 composite distance to Doyle is
reported only as a secondary diagnostic, never as evidence of success
(§26). Result is classified NEGATIVE / WEAK_MIXED / POSITIVE_MECHANISTIC
per §29, and no claim beyond §30's permitted phrasing is made regardless of
outcome. Phase B runs only if Phase A's gate passes, and no parameter
changes after Voynich exposure.
