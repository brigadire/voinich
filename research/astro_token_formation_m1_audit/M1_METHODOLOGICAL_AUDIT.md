# M1 methodological audit

## Scope and method

This is an independent audit of the frozen M1 global-substitution search
(`research/astro_token_formation_m1/`) and its M0 foundation
(`research/astro_token_formation/`), performed before any dictionary
expansion, per `tasks_other/brutforce-astronomical-labels-M1-audit.md`. No
frozen M0/M1 artifact, corpus, rule, or threshold was changed. All
computation performed for this audit is tagged `AUDIT_ONLY` and lives under
`research/astro_token_formation_m1_audit/`:

* `audit_diagnostics.py` → `M1_AUDIT_DIAGNOSTICS.tsv` (47 rows): reproducibility,
  beam-width sweep (8–256), label-visit-order sweep, dictionary-capacity
  arithmetic, HELD_OUT semantics comparison (object- vs. form/rule-disjoint),
  per-class TRAIN breakdown, split-leakage check, null-summary cross-check.
* `audit_per_class_null.py` → `M1_AUDIT_PERCLASS_NULL.tsv` (90 rows): a
  30-replicate-per-control subset of the frozen null seeds, rerun with a
  per-class (STAR vs. PLANET_MOON) coverage breakdown that the original
  `M1_NULL_RESULTS.tsv` did not record.
* `M1_AUDIT_NULL_REVIEW.md`, `M1_AUDIT_MATCHING_REVIEW.md`: narrative review
  for Sections 8 and 2/3/5/6/7 respectively.
* `M1_AUDIT_FINDINGS.tsv`: the full findings register (15 rows).

Every frozen M0 and M1 artifact was also independently hash-verified against
its own manifest; see Section 6/Provenance below for the one discrepancy
found.

## Section-by-section findings

### 1. Sampling and split — `NO_ISSUE` (leakage), `MAJOR` (class pooling)

The split is spelling-blind and leak-free (F01): both `sample_rank_sha256`
and `split_rank_sha256` are computed only from `stolfi_coordinate`, never
from `voynich_token` text, and an independent check of all 26 selected labels
found zero duplicate tokens between TRAIN and HELD_OUT.

However, STAR and PLANET_MOON are pooled into one TRAIN/HELD_OUT statistic
(F02), with very unequal sample sizes (16 vs. 4 TRAIN items) that cannot be
judged against the same null. A class-conditioned null rerun for this audit
(30 replicates/control; `M1_AUDIT_PERCLASS_NULL.tsv`) shows: **STAR (7/16 =
43.75%)** sits only mildly above its own class-specific null mean (36–40%
across the three controls; pointwise p ≈ 0.20–0.40 — not significant), and
**PLANET_MOON (3/4 = 75%)** is squarely null-compatible once compared to its
own n=4 null — under `RANDOM_VOYNICH_SET`, 73% of null replicates already
meet or exceed 75%. (An earlier pass of this audit mistakenly compared
STAR's figure to the *pooled* null mean and called it "below null" — that
comparison was wrong and is corrected here using the class-conditioned
null.) The audit also finds the specific PLANET_MOON object identity
assigned to the same label is unstable across otherwise-comparable rule
families (F03) — further evidence the 75% is a small-sample/model-freedom
coincidence rather than a stable correspondence. See "Per-class null" below
for the full table.

**Families should be evaluated independently**, as the audit brief
anticipated; this is the single most consequential audit finding, though it
does not overturn the pooled NULL_COMPATIBLE verdict (see "Overall verdict").

### 2. Historical-object capacity — `NO_ISSUE`

`DICTIONARY_CAPACITY_CEILING = 100%` for both classes, including the
worst case where TRAIN's matching uses the maximum number of distinct
objects it is allowed to (F04). STAR's dictionary size (21) was deliberately
set equal to its label sample size specifically so that capacity could never
be the reason 70% coverage is unreachable (documented in
`TOKEN_FORMATION_RULE_SPACE.md`), and the audit confirms this design goal is
met: at least `21 − 16 = 5` objects always remain for the 5 HELD_OUT STAR
labels regardless of which 16 TRAIN objects a matching happens to use.
PLANET_MOON has 2 objects of slack beyond even that. No label is
mathematically unreachable due to dictionary size.

### 3. Anonymous maximum matching — `NO_ISSUE` (mechanics), `MAJOR` (identity assumption for PLANET_MOON)

The matching implementation is correct by code reading and by direct testing
(F05, F06): one object per label and vice versa, spelling variants are
alternatives of the same object rather than extra objects, HELD_OUT excludes
only the objects the canonical TRAIN matching actually used, and every
single one of the 432 unexplained-label rows across the 25 retained models
fails for `NO_COMPLETE_MAPPED_TERM` — never a one-to-one assignment
conflict. Matching ambiguity is also empirically negligible: the top model
has `assignment_multiplicity=1` and `ambiguity_excess=0.0`.

The one substantive concern (repeated from Section 1, F03) is that the
PLANET_MOON sample's very premise — that each label denotes one distinct,
uniquely identifiable planet — rests on an annotator's own tentative,
question-marked guesses for a numbered diagram sequence ("moon 1… the Sun?",
"moon 3… Venus?", etc.), which M0's report already declines to use as
identities but which the *matching model itself* still assumes are unique
objects.

### 4. Grapheme segmentation — `NO_ISSUE`

The EVA target-composite inventory (`cth, ckh, cph, cfh, iin, ain, ch, sh,
ee, in`) is a repository-wide frozen convention that predates and is reused
outside M1 (`research/astro_label_cross_section/main.py` uses the identical
list) — it was not tuned to this label set (F07). Source/target
commensurability (1-or-2-unit outputs, digraphs as atomic source units) is
explicit in `M1_SEARCH_CONFIG.yaml` and `M1_GRAPHEME_INVENTORY.tsv` and was
not altered during or after the search.

### 5. M1 substitution model — `MODERATE`

The model correctly forbids zero-output mappings, context-sensitivity, and
per-label exceptions, and allows one-to-one/many-to-one/one-to-two mappings
as documented. One inconsistency was found (F08): M1's abbreviation step
truncates by **segmented-unit count**, while the identically-named M0 rule it
is nominally inherited from truncates by **raw character count** — these
differ whenever a source form contains a digraph near the truncation
boundary, yet M1 reuses M0's `Rule.complexity` cost unchanged for these
labels. Narrow impact, but the complexity accounting is not exactly
equivalent to what its name implies.

The mapping-size/complexity-46 concern the audit brief specifically flags
("не слишком ли много степеней свободы") is real and is addressed
quantitatively in Section 7 below.

### 6. Search algorithm — `NO_ISSUE`, plus one provenance `MINOR`

Beam width was swept 8→256 (4× the frozen width) and label visit order was
independently varied across four orderings. **Every configuration converges
to TRAIN coverage 0.500**, via several different specific mappings at the
same coverage (multiple local optima, not one fragile solution) — see
`M1_AUDIT_MATCHING_REVIEW.md` §5 for the full table (F10). This means the
frozen `NO_MODEL` verdict is robust to search-budget and tie-break choices
far beyond what the single frozen beam-64 run alone would justify claiming.

Separately (F11): `M1_MANIFEST.json`'s recorded hash for M1's own `main.py`
does not match the committed script (both the working tree and `git HEAD`
agree with each other, and both differ from the manifest). Independent
reproduction using the current script reproduces the frozen top model
exactly, so this looks like a stale hash from a late, behavior-preserving
edit — but the manifest cannot, by itself, prove that.

### 7. Complexity metric — `MODERATE`

`total_complexity=46` for the best model corresponds to 18 of 19 mapped
source letters emitting a freely chosen two-EVA-unit output drawn from ~20
possible target units (up to ~400 combinations per letter), constrained only
by internal self-consistency across the corpus. This is close to an
unconstrained substitution cipher fitted to 20 short strings, and this audit's
beam-width/order sweep plus the null cross-check (F09) show this is *exactly*
why the ceiling is ~50%: not just the observed model, but all 300 null
replicates across three structurally different controls also top out at
0.500. The frozen complexity penalty discourages this outcome relative to
smaller mappings in the score, but correctly does not prevent it from being
found — the frozen band logic then correctly classifies it as
`NULL_COMPATIBLE`, not `STRONG_CANDIDATE`. No defect in the verdict; the
report could state the "≈50% is close to this model class's generic
ceiling" point more directly rather than leaving it implicit in the p-value.

### 8. Null design — `NO_ISSUE`, one `MINOR`

Full review in `M1_AUDIT_NULL_REVIEW.md`. All three controls model distinct,
non-redundant null hypotheses (F12), each correctly reruns the complete
bounded 640-pipeline optimiser per replicate rather than comparing against a
fixed target (the single most common way a null design like this goes
wrong), and `RANDOM_VOYNICH_SET` conservatively draws from real observed
Currier-A occurrences rather than synthetic strings. One documentation gap
(F13): `RANDOM_VOYNICH_SET` does not class-condition its token draw, which is
inert because matching is separately class-restricted downstream, but this
is not stated explicitly in the frozen docs.

### 9. Held-out semantics — `NO_ISSUE` for this run, methodologically live for PLANET_MOON

Re-scoring the top 10 retained models' frozen mappings under a
FORM/RULE-DISJOINT semantics (objects **not** excluded after TRAIN use)
against the same OBJECT-DISJOINT semantics the frozen run used gives
**identical HELD_OUT results (0/6) for every one of the 10 models** (F14).
The frozen `HELD_OUT=0/6` is therefore not an artifact of which
disjointness rule was chosen for this run. The distinction remains live in
principle, specifically for PLANET_MOON (per Section 1/3), where the
one-object-per-label premise is itself uncertain — a future run evaluating
PLANET_MOON on its own should report both semantics rather than only the
object-disjoint default.

### 10. 70% threshold — `MODERATE`

TRAIN's 20-item sample gives 5%-wide discrete coverage steps (1/20), which
is reasonably fine relative to the 70% gate. HELD_OUT's 6-item sample gives
only 16.7%-wide steps, so the frozen `STRONG_CANDIDATE` HELD_OUT gate
(≥0.50) requires exactly 3/6 and cannot be approached gradually — one item
flips the gate by a full 16.7 points (F15). This is not a design error (n=6
is essentially fixed by how many eligible single-token labels exist at all)
but readers should not treat a HELD_OUT gate crossing as more precise
evidence than a 6-item binomial sample supports.

## Per-class null (additional diagnostic beyond the ten sections)

Because the frozen `M1_NULL_RESULTS.tsv` only recorded the pooled 20-label
statistic, this audit reran a 30-replicate-per-control subset of the same
frozen seeds (`seed_for(control, replicate)` for `replicate in 0..29`, same
seeding function as production) with a per-class breakdown
(`M1_AUDIT_PERCLASS_NULL.tsv`):

| Class | Observed | PSEUDODICTIONARY mean / max | RANDOM_VOYNICH_SET mean / max | SHUFFLED_TERMS mean / max | P(null ≥ observed) |
|---|---|---|---|---|---|
| STAR (7/16) | 43.75% | 40.2% / 56.25% | 35.6% / 43.75% | 39.8% / 56.25% | 0.40 / 0.20 / 0.37 |
| PLANET_MOON (3/4) | 75.00% | 55.0% / 75% | 67.5% / 75% | 51.7% / 75% | 0.37 / **0.73** / 0.23 |

Both classes are individually null-compatible: STAR sits mildly above its
own null's mean but nowhere near a significant tail (pointwise p 0.20–0.40),
and PLANET_MOON sits at or below the median of at least one of its own null
distributions (`RANDOM_VOYNICH_SET`, p ≈ 0.73). This directly corrects an
error in this audit's first-pass narrative, which had compared STAR's
per-class figure to the *pooled* null mean rather than a class-conditioned
one and mischaracterized it as "below null" — see `M1_AUDIT_MATCHING_
REVIEW.md` §3 for the corrected discussion. The corrected, class-conditioned
picture is if anything a cleaner confirmation of `NULL_COMPATIBLE` than the
pooled statistic alone: decomposed correctly, *neither* class shows a
significant departure from its own chance baseline.

## Overall verdict

```text
M1_METHODOLOGICAL_AUDIT=
VALID_WITH_LIMITATIONS

CRITICAL_FINDINGS=
0
MAJOR_FINDINGS=
2
MODERATE_FINDINGS=
3

M1_RESULT_STILL_INTERPRETABLE=
YES

DICTIONARY_EXPANSION_AUTHORIZED=
YES
```

**Rationale.** No finding invalidates M1's core inference
(`ASTRO_TOKEN_FORMATION_M1=NO_MODEL`, i.e. this bounded compact
grapheme-substitution search does not exceed its own null on this frozen
corpus/sample). If anything, the audit's independent diagnostics — beam-width
and visit-order stability up to 4× the frozen search budget, capacity ceiling
at 100% for both classes, matching-conflict analysis showing genuine
no-candidate failures rather than starvation artifacts, and a class-
conditioned null showing *neither* class individually departs significantly
from its own chance baseline (STAR pointwise p 0.20–0.40; PLANET_MOON
pointwise p up to 0.73) — make the negative conclusion *more* robust, not
less. The two `MAJOR` findings (class pooling that cannot be naively
decomposed without a matching per-class null, and the PLANET_MOON uniqueness
assumption) are about
how the result should be *reported and read*, not about a search-mechanics
defect that suppressed a true signal or inflated the null. They should be
fixed before or alongside any dictionary-expansion run (report STAR/
PLANET_MOON separately; flag PLANET_MOON's identity assumption explicitly),
but they do not make a small-vs-expanded-dictionary comparison invalid,
since (a) dictionary capacity was never the binding constraint (Section 2)
and (b) the search/null/matching mechanics that such a comparison would rely
on are independently verified sound in this audit.
