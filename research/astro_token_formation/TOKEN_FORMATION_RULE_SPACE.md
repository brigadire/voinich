# Frozen token-formation rule space

## Freeze and scope

This specification was frozen before the production search. The search uses
only independently matched, single-token physical labels from
`STOLFI_ASTRO_LABEL_MATCHES.tsv`. The deterministic analysis sample contains
21 `STAR` coordinates and five `PLANET_MOON` coordinates. Stars are the 21
lowest SHA-256 ranks under `ASTRO_FORMATION_SAMPLE_V1|coordinate`; planets are
the five eligible single-token `planet?` coordinates. Sampling is independent
of token spelling. Within class, the lowest ranks under
`ASTRO_FORMATION_SPLIT_V1|coordinate` allocate 16/21 stars and 4/5 planets to
TRAIN; the remainder are HELD_OUT.

The 21-star cap equals the size of the artifact-based star lexicon. It prevents
dictionary capacity alone from making 70% coverage impossible. It is not a
claim that the anonymous manuscript stars have the identities in the corpus.
The five planet-like Stolfi records are morphological objects only: question
marks in Stolfi's comments are not treated as semantic identifications.

## Representations and matching

Historical forms are lower-cased, Unicode diacritics are removed, `j` is
retained until the selected orthographic rule, and non-letters delimit words.
EVA is compared as the exact lower-case ZL3b token string. No edit distance,
wildcard, per-label exception, or image interpretation is allowed.

Because the physical stars have no independently known identities, evaluation
is explicitly **anonymous set matching** within object class. A deterministic
maximum bipartite matching assigns at most one corpus object to a label and at
most one label to an object. A corpus object's documented variants are
alternative attestations, not additional objects. HELD_OUT uses the frozen
TRAIN model and corpus unchanged; objects used by the canonical TRAIN matching
are unavailable to HELD_OUT.

## Finite rule grid

Every model is one global Cartesian-product tuple. Operations run in the order
shown.

1. `lexical_selector` (one):
   `FULL_JOIN`, `DROP_AL`, `HEAD`, `TAIL`.
   `DROP_AL` removes only a word exactly equal to Arabic article *al*; `HEAD`
   and `TAIL` select one word globally and model medieval truncation/headword
   practice.
2. `case_ending` (one): `KEEP`, `STRIP_LATIN`.
   `STRIP_LATIN` removes the first applicable final ending from the frozen list
   `ibus,orum,arum,ium,ius,ae,is,us,um,ii,am,em,as,es,os,i,o,a,e` when at least
   three characters remain.
3. `orthography` (one):
   `IDENTITY`, `IJ_UV` (`j→i`, `v→u`),
   `ARABIC_LATIN` (`kh→h`, `gh→g`, `sh→s`, `th→t`, `dh→d`, then `j→i`,
   `w→u`), `VELAR_COLLAPSE` (the preceding changes plus `q,c→k`).
4. `vowel_rule` (one):
   `KEEP`, `DELETE_FINAL`, `COLLAPSE_A` (all `a,e,i,o,u,y→a`),
   `CONTRACT_INTERNAL` (remove vowels except an initial or final vowel).
5. `abbreviation` (one):
   `NONE`, `SUSPEND_1`, `SUSPEND_2`, `PREFIX_3`, `PREFIX_4`.
   Suspensions delete one or two final characters if at least three remain;
   prefix modes retain exactly the first three/four characters when longer.

The grid therefore contains `4 × 2 × 4 × 4 × 5 = 640` models. All documented
forms of an object pass through the same tuple. Duplicate outputs collapse.
There are no learned character mappings, label-conditioned choices, target
affixes, transpositions, reversals, arbitrary insertions, or unconstrained
deletions.

## Frozen complexity and scoring

Complexity is the number of non-identity choices, with `HEAD`, `TAIL`,
`PREFIX_3`, and `PREFIX_4` costing two rather than one because they discard an
unbounded suffix/prefix context. Consistency is 1.0 for every model because the
tuple is global. Transformation quality is `HIGH` for complexity 0–1,
`MEDIUM` for 2–3, and `LOW` above 3.

The diagnostic score is:

`train_coverage + heldout_coverage + null_advantage + 0.25*consistency - 0.05*complexity`

Unexplained labels are exceptions reported as counts and rows; no additional
penalty makes 100% coverage mandatory.

## Frozen bands

Null advantage is observed TRAIN coverage minus the largest mean TRAIN
coverage among the applicable null controls.

* `STRONG_CANDIDATE`: TRAIN ≥ 0.70, HELD_OUT ≥ 0.50, null advantage ≥ 0.20,
  complexity ≤ 4.
* `PARTIAL_CANDIDATE`: TRAIN ≥ 0.40, HELD_OUT ≥ 0.30, null advantage ≥ 0.10,
  complexity ≤ 3, and not strong.
* `OVERFIT`: TRAIN ≥ 0.70 but a strong condition fails.
* `NULL_COMPATIBLE`: everything else.
* `PREDICTIVE_MODEL` final status additionally requires a strong candidate,
  familywise empirical null `p ≤ 0.05`, and HELD_OUT ≥ 0.50.

All threshold-passing strong, partial, and overfit models are retained. If none
passes, the ten best `NULL_COMPATIBLE` models are retained diagnostically and
do not count as `MODELS_FOUND`.

## Frozen negative controls

The seed is `20260901`; each stochastic control has 1,000 replicates.

1. `SHUFFLED_ASSIGNMENT`: preserve each model's canonical TRAIN term outputs
   but randomly permute its label assignments; score the fraction of canonical
   pairs preserved. This is secondary because no true object identities exist.
2. `RANDOM_VOYNICH_SET`: draw class- and length-matched token occurrences
   without replacement from the frozen Astronomical-section ZL3b occurrence
   pool. Repeated forms remain possible because distinct real occurrences may
   carry the same form.
3. `SHUFFLED_TERMS`: independently permute characters within every normalized
   historical source form, then apply the frozen tuple.
4. `PSEUDODICTIONARY`: generate class- and length-matched source forms from
   the empirical historical-letter distribution, then apply the frozen tuple.

Model selection never uses HELD_OUT spellings or any post-run corpus change.

## Historical source keys

* `NMS_RETE_13C`: National Museums Scotland, “A 1000-year-old star catcher,”
  live catalogue article accessed 2026-09-01, table of the 21 pointers and
  Arabic inscriptions on the thirteenth-century rete of object T.1959.62.
* `KUNITZSCH_1987`: Paul Kunitzsch, “An unknown Arabic source for star names,”
  *History of Oriental Astronomy*, IAU Colloquium 91 (1987), pp. 155–163,
  DOI `10.1017/S0252921100105986`; the corpus uses its attested 1246 spelling
  *bedalgeuze* and does not import the paper's doubtful assignments.
* `WALTERS_W73`: Walters Art Museum MS W.73, medieval cosmographical diagrams
  naming Luna, Mercurius, Venus, Sol, Mars, Iupiter, and Saturnus.
* `ORDO_PLANETARUM`: pseudo-Bede, *Ordo planetarum*, anonymous eighth-century
  text in Migne, *Patrologia Latina* 90, cols. 943D–946A (Corpus Corporum
  transcription mirrored by Latin Wikisource, accessed 2026-09-01).
