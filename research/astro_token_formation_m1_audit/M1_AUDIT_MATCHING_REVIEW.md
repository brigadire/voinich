# M1 audit — anonymous matching, capacity, and complexity review

All numbers below are reproduced independently by
`audit_diagnostics.py` (`M1_AUDIT_DIAGNOSTICS.tsv`) and
`audit_per_class_null.py` (`M1_AUDIT_PERCLASS_NULL.tsv`), both tagged
`run_type=AUDIT_ONLY`. No corpus, rule, or frozen M1 output was changed.

## 1. Dictionary capacity ceiling (Section 2)

| Class | Historical objects | Sampled labels | TRAIN | HELD_OUT | Joint ceiling | Worst-case HELD_OUT ceiling if TRAIN saturates | Zero slack? |
|---|---|---|---|---|---|---|---|
| STAR | 21 | 21 | 16 | 5 | 100% | 5/5 (100%) | Yes (21 objects = 21 labels) |
| PLANET_MOON | 7 | 5 | 4 | 1 | 100% | 1/1 (100%) | No (2 spare objects) |

**Finding: capacity is never the bottleneck, even in the worst case.**
STAR's 21-object dictionary exactly equals its 21-label sample (by design —
`TOKEN_FORMATION_RULE_SPACE.md` documents this explicitly: "The 21-star cap
equals the size of the artifact-based star lexicon. It prevents dictionary
capacity alone from making 70% coverage impossible"). Because a matching by
construction never uses more objects than there are TRAIN labels, at least
`21 − 16 = 5` STAR objects always remain for the 5 HELD_OUT labels regardless
of *which* 16 TRAIN objects are used, and PLANET_MOON has 2 objects of slack
beyond even that. **`DICTIONARY_CAPACITY_CEILING = 100%` for both classes,
jointly and object-disjoint.** The audit's a-priori concern that
object-disjoint HELD_OUT could "artificially" starve HELD_OUT of feasible
objects (Section 3, last paragraph) is **not supported empirically**: in
every retained model the reason a HELD_OUT label goes unmatched is
`NO_COMPLETE_MAPPED_TERM` (zero compatible candidate objects under that
model's substitution table), never a matching conflict with an object TRAIN
had claimed. `M1_UNEXPLAINED_LABELS.tsv` has 432 rows and every one carries
this reason; not one is `ONE_TO_ONE_ASSIGNMENT_CONFLICT`. `NO_ISSUE` (Section
2/3), verified rather than assumed.

## 2. Anonymous matching correctness (Section 3)

* One-object↔one-label and no duplicate-object-via-spelling-variant: verified
  by code reading (`preprocess()` collapses variant forms into a per-object
  set before matching; `maximum_matching` is a standard augmenting-path
  bipartite matcher keyed by `object_id`, not by form).
* HELD_OUT correctly excludes only the **canonical TRAIN matching's** objects
  (`used = set(model["train_match"].values())`), not every object touched by
  any candidate edge — this is the right scope for "used," and per item 1
  above it never binds in practice.
* `M1_001` (the top-ranked model) has **assignment multiplicity 1** and
  **ambiguity_excess 0.0** — every one of its 10 matched TRAIN labels has
  exactly one compatible object, not several competing ones. This is smaller
  than a naive prior would predict given a 19-entry near-free substitution
  table (see §4 below): the 1-or-2-EVA-unit length constraint plus per-object
  form/length filtering in `align_constraints` prunes most spurious
  candidates before matching even runs. Across the top 25 retained models,
  `ambiguity_excess` never exceeds 0.05 (≤1 extra candidate spread over 20
  labels). **NO_ISSUE** — the "artificial gain from interchangeable objects"
  risk flagged in Section 3 is real in principle but empirically negligible
  here.

## 3. STAR vs. PLANET_MOON evaluated independently (Section 1 / Section 3, "repeating categories")

The frozen report states a single pooled figure, `10/20 = 0.50`. Splitting the
top 10 retained models by class:

| Model rank | STAR (of 16) | PLANET_MOON (of 4) |
|---|---|---|
| 1–6 (best band, TRAIN 0.50) | 7/16 = **43.75%** | 3/4 = **75.00%** |
| 7–10 (TRAIN 0.45) | 6/16 = 37.5% | 3/4 = **75.00%** |

**This is the most significant finding of the audit (`MAJOR`).** The pooled
"50%" headline hides that the two classes are not evaluated against
comparable evidence: STAR carries 16 TRAIN items and PLANET_MOON only 4, so a
single pooled number cannot show whether either class actually departs from
chance. A **class-conditioned** null was therefore run for this audit — 30
replicates per control, per-class coverage recorded — since the frozen
`M1_NULL_RESULTS.tsv` only ever recorded the pooled 20-label statistic
(`M1_AUDIT_PERCLASS_NULL.tsv`):

| Class | Observed | Null mean (Pseudo / Random / Shuffled) | Null max | Fraction of null replicates ≥ observed |
|---|---|---|---|---|
| STAR (7/16) | 43.75% | 40.2% / 35.6% / 39.8% | 56.25% (all three) | 0.40 / 0.20 / 0.37 |
| PLANET_MOON (3/4) | 75.00% | 55.0% / 67.5% / 51.7% | 75% (all three) | 0.37 / **0.73** / 0.23 |

**Correcting an error in this audit's own first-pass reading:** an earlier
draft of this review compared the STAR-only observed figure against the
*pooled* (STAR+PLANET_MOON) null mean and called it "below null." That
comparison was wrong — the correct STAR-*specific* null (above) has a lower
mean (36–40%) than STAR's observed 43.75%. Read correctly, STAR sits
*mildly above* its own null's mean, at roughly the 60th–80th percentile
(pointwise p ≈ 0.20–0.40 across the three controls) — not a below-chance
result, but not a significant one either (nowhere near the ≤0.05/≤0.10 gates).

PLANET_MOON's 75% is unambiguously null-compatible once compared against the
right (class-conditioned, n=4) null: under `RANDOM_VOYNICH_SET` alone, 73% of
null replicates *already meet or exceed* the observed 75% — i.e. the observed
value sits at roughly the null's own median-to-below-median range for that
control, not in its tail. Two considerations further bound how much weight
the raw 75% figure can bear even before this null comparison:

1. **n = 4.** With only 4 TRAIN PLANET_MOON labels, the discrete outcomes are
   0%, 25%, 50%, 75%, 100% — hitting 75% by chance is unremarkable on its
   face, and the class-conditioned null above confirms this directly rather
   than leaving it as an inference from sample size alone.
2. **Object-identity risk specific to this sub-class.** The five sampled
   PLANET_MOON labels all come from folio f67r2's "moon" diagram, where
   Stolfi/Zandbergen's own transcription comments read `"moon 1, dark
   coloured; the Sun?"`, `"moon 3 …; Venus?"`, `"moon 4 …; Mercury?"`, `"moon
   6 …; Saturn?"`, `"moon 9 …; Jupiter?"`, `"moon 11 …; Mars?"` — i.e. a
   sequence of numbered diagram positions with the annotator's own tentative,
   question-marked planetary guesses. M0's report already declines to use
   these specific identity guesses ("Question-marked Stolfi planet comments
   supplied only the morphological class; their proposed identities were
   never used") — a correct and important caution — but the *object-disjoint,
   one-label-per-unique-planet* matching assumption inherits the same
   uncertainty: if this diagram in fact encodes a repeating/sequential
   category (moon phases, lunar mansions, or a generic list) rather than
   seven distinct named planets, the very idea of "matching a label to a
   unique historical planet object" may not be the right model for this
   sub-class at all, independent of any transformation rule. This is exactly
   the case the audit brief calls out: "не делает ли запрет повторного
   использования historical object... искусственно жёстким, если labels
   потенциально могут обозначать повторяющиеся категории."

**Recommendation:** report STAR and PLANET_MOON coverage, HELD_OUT, and null
comparisons **separately** in any future run, and treat the PLANET_MOON class
as a distinct, lower-confidence stream whose "identity" assumption itself
needs a dedicated methodological note — not as a class that should be pooled
with STAR to produce one headline percentage. This does not change the
overall verdict (STAR alone is already null-compatible; PLANET_MOON alone is
too small to certify either way) but it does mean the frozen `10/20 = 0.50`
headline is easy to over-read as "half the astronomical vocabulary matches" —
it is better read as "STAR does not beat chance; PLANET_MOON is too small a
sample, on a possibly-non-unique-object diagram, to say anything."

## 4. Why ~50% TRAIN coverage is not surprising under this substitution model (Section 5 / 7)

M1_001's substitution table has 19 entries, 18 of which map a single Latin
letter to a **freely chosen two-EVA-unit sequence** (drawn from an output
alphabet of ~20 target units, i.e. up to ~400 possible 2-unit outputs per
source letter, subject only to global consistency). This is very close to an
**unconstrained monoalphabetic-with-digraph-output substitution cipher** with
~19 free parameters, fitted greedily to just 20 short target strings (mean
length ≈ 7 EVA units). A model class this permissive is expected to
coincidentally satisfy roughly half of 20 independent short-string
constraints, which is exactly what all three independent null controls also
achieve (max coverage 0.500 in every one of 300 replicates — see
`M1_AUDIT_NULL_REVIEW.md`). The audit finds the `total_complexity=46`,
`mapping_size=19` penalty terms in the frozen scoring formula (§`M1_SEARCH_
CONFIG.yaml`) do *discourage* this outcome relative to smaller mappings
(`score` is lower for M1_001 than for less-complete but far smaller mappings
elsewhere in the retained list), but the penalty is not steep enough to stop
the beam search from settling near this ceiling — nor should it be, since
`NULL_COMPATIBLE` classification is exactly the intended outcome here and the
frozen bands correctly reject it (`OVERFIT`/`NULL_COMPATIBLE`, not
`STRONG_CANDIDATE`). **NO_ISSUE for the verdict; `MODERATE` note for
transparency** — the report text could say more explicitly that ~50%
coverage is close to the generic capacity of this model class regardless of
input, which the audit's beam-width/order sweep and null replicates now
demonstrate directly rather than leaving as an inference from the p-value
alone.

## 5. Search algorithm stability (Section 6)

Beam width was swept at 8/16/32/64/128/256 (4× the frozen width) on the
unmodified TRAIN sample and corpus. **Every width reproduces TRAIN coverage
= 0.500 with mapping_size = 19** and the same winning rule family
(`TAIL;STRIP_LATIN;ARABIC_LATIN;CONTRACT_INTERNAL;PREFIX_3` from width 16
up; width 8 finds an equally-scoring `IDENTITY` variant). Label visit order
was independently varied (frozen edge-count-ascending vs. descending vs.
coordinate-ascending vs. descending): all four converge to TRAIN coverage
0.500 as well, but through **different mapping sizes (19, 20, 22, 23) and
different specific tables** — confirming genuine **multiple local optima at
the same coverage ceiling**, not a single fragile solution, and confirming
that width 64 is not silently missing a materially better compact mapping.
`NO_ISSUE` for Section 6 — the frozen beam width is adequate for the
conclusion actually drawn (`NO_MODEL`/`NULL_COMPATIBLE`); a negative result
this robust to 4× the search budget and to visit-order perturbation is a
stronger negative than a single beam-64 run alone would justify claiming.

## 6. Provenance/manifest hygiene (new finding, not in the original ten sections but relevant to Section 6/repro trust)

`M1_MANIFEST.json`'s recorded `input_sha256` for
`research/astro_token_formation_m1/main.py` (`68a3102f…`) does **not** match
either the currently committed file or its working-tree copy (`5eb7a764…`,
identical to `git show HEAD`). This means the manifest's own chain-of-custody
claim about "which exact script produced these artifacts" cannot be verified
from the manifest alone. Independent reproduction in this audit
(`REPRODUCIBILITY_BEAM64` in `M1_AUDIT_DIAGNOSTICS.tsv`) reruns the
*currently committed* script and reproduces `M1_001` exactly
(TRAIN 10/20, mapping_size 19, identical rule string), so the discrepancy
appears to be a stale hash captured before a late, behavior-preserving edit
rather than a silent result-affecting change — but this cannot be proven
without the pre-edit script. `MINOR` finding: regenerate
`M1_MANIFEST.json`/`M1_SHA256SUMS` from the current `main.py` (or note in the
manifest that a post-freeze edit occurred) before treating M1 as fully
provenance-clean for any downstream (e.g. dictionary-expansion) comparison.
