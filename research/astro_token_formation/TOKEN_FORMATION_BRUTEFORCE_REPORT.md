# Constrained astronomical token-formation search

## Result

`NO_MODEL`. The best of 640 frozen global pipelines explained
0/20 TRAIN labels (0.000000) and
0/6 HELD_OUT labels (0.000000).
It did not cross a candidate threshold. The ten rows in
`TOKEN_FORMATION_MODELS.tsv` are diagnostic null-compatible ties and are not
claimed as decipherments.

## Design

The spelling-blind SHA-256 sample contains 21 STAR and five PLANET_MOON
single-token labels. It was frozen as 20 TRAIN and six HELD_OUT labels before
the search. The corpus contains 21 Arabic star-name inscriptions attested on a
thirteenth-century astrolabe rete and seven medieval Latin planet names.
Question-marked Stolfi planet comments supplied only the morphological class;
their proposed identities were never used.

The search is anonymous one-to-one set matching within class because no
independent star identity exists for these manuscript labels. Consequently a
match would show only compatibility with a lexicon and global rule, not a
semantic identification. Exact transformed strings alone count; no edit
distance, image reading, per-label rule, or exception is available. Full rules
and preregistered bands are in `TOKEN_FORMATION_RULE_SPACE.md`.

## Negative controls

Each stochastic control used 1000 deterministic replicates (seed
20260901). The familywise statistic is the maximum over the full 640-rule grid in
each replicate. No control replicate is used to modify the corpus or rules.
For this run the best observed TRAIN coverage was 0.000000;
the best model's largest mean null coverage was 0.000000, for a
null advantage of 0.000000. See
`TOKEN_FORMATION_NULL_TEST.tsv` for control-specific tails and the frequency of
coverage at or above 0.70.

`SHUFFLED_ASSIGNMENT` is explicitly secondary: because label identities are
unknown, it tests stability of the canonical anonymous pairing rather than a
known semantic assignment. Random token sets, shuffled historical forms, and
matched pseudodictionaries are the substantive controls.

## Historical provenance

The star list is the National Museums Scotland catalogue article “A
1000-year-old star catcher,” live version accessed 2026-09-01, specifically its
table of 21 calculated pointers on the thirteenth-century rete of T.1959.62.
Only the Arabic forms in the `On rete` column were transcribed. Kunitzsch,
“An unknown Arabic source for star names,” *History of Oriental Astronomy*
(1987), DOI `10.1017/S0252921100105986`, supplies only the directly attested
1246 spelling `bedalgeuze`; doubtful or differently assigned names were not
imported. Walters MS W.73 (late twelfth century), fol. 2v, directly lists the
seven Latin heavenly bodies. The anonymous eighth-century pseudo-Bedan `Ordo
planetarum`, Migne PL 90, cols. 943D–946A, documents their inflected forms.

## Interpretation

Failure here rejects only this compact, historically constrained 640-model
space against this frozen corpus/sample. It is not proof that no historical
formation system exists. In particular, the test deliberately refuses an
arbitrary Latin/Arabic-to-EVA substitution alphabet, modern reconstructed
names, reverse strings, or post-hoc spellings. Unexplained labels remain in
`TOKEN_FORMATION_UNEXPLAINED_LABELS.tsv`; no translation is proposed for them.

## Reproduction

Run `python3 research/astro_token_formation/freeze_split.py`, then
`python3 research/astro_token_formation/main.py`. Corpus bytes remain governed
by `DATA.md`; the analysis uses the already frozen occurrence metadata and
Stolfi match audit.

## Final status

```text
ASTRO_TOKEN_FORMATION_SEARCH=NO_MODEL
MODELS_FOUND=0
MODELS_WITH_COVERAGE_GE_70=0
BEST_TRAIN_COVERAGE=0.000000
BEST_HELDOUT_COVERAGE=0.000000
BEST_MODEL_COMPLEXITY=0
BEST_MODEL_NULL_ADVANTAGE=0.000000
UNEXPLAINED_LABELS=26
```
