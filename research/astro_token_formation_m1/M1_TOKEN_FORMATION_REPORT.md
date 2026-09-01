# M1 global grapheme-substitution search

## Result

`NO_MODEL`. The best bounded-search model matched 10/20
TRAIN labels (0.500000) and 0/6
HELD_OUT labels (0.000000). Its table has
19 source mappings and total complexity
46.

## Search and controls

M1 retained the frozen M0 sample, corpus, 640 preprocessing pipelines, and
anonymous within-class one-to-one matching. The only new capacity is a single
global source-unit→one/two-EVA-unit table. The deterministic beam width is 64;
therefore this is a bounded optimisation result, not a proof of the global
optimum.

Each control used 100 independently seeded replicates and reran the complete
bounded optimiser:

- `PSEUDODICTIONARY`: mean max coverage 0.446000; p95 0.500000; max 0.500000
- `RANDOM_VOYNICH_SET`: mean max coverage 0.432500; p95 0.500000; max 0.500000
- `SHUFFLED_TERMS`: mean max coverage 0.461500; p95 0.500000; max 0.500000

The conservative null advantage is 0.038500; empirical p is
0.821782. Assignment edges and multiplicity are reported
separately because maximum matching does not identify astronomical objects.

## Interpretation

A positive band means only that a compact global grapheme table is more
compatible with this anonymous vocabulary than the frozen controls. It is not
a translation, language identification, or star identification. Conversely a
negative band rejects only this inventory and bounded beam search. No mapping,
spelling, threshold, or HELD_OUT choice was added after TRAIN inspection.

## Final status

```text
ASTRO_TOKEN_FORMATION_M1=NO_MODEL
MODELS_FOUND=0
MODELS_WITH_TRAIN_COVERAGE_GE_70=0
BEST_TRAIN_COVERAGE=0.500000
BEST_HELDOUT_COVERAGE=0.000000
BEST_MAPPING_SIZE=19
BEST_MODEL_COMPLEXITY=46
BEST_NULL_ADVANTAGE=0.038500
BEST_EMPIRICAL_P=0.821782
UNEXPLAINED_LABELS=16
```
