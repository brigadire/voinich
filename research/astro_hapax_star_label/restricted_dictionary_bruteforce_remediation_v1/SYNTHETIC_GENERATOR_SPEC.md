# Synthetic Generator Spec

`scripts/generator.py`. Physically separate from every solver module (no solver imports it, it
imports no solver) — the only connection point for the sealed benchmark is the file system
(`SEALED_BENCHMARK_PROTOCOL.md`'s dev/hidden pipeline).

## What it reads from the real scope

Only `restricted_dictionary_bruteforce_v3_design/REAL_SCOPE_STRUCTURAL_MANIFEST.json` — the
"previously authorized structural manifest" the task's prohibition list explicitly carves out. It
contains aggregate statistics only (`occurrences_total=57`, the 16-symbol EVA alphabet, a length
histogram, a distinct-symbol-count histogram, the 30/27 page split) and its own
`restrictions_enforced` block confirms no dictionary scores, assignments, or real-data search were
ever computed from it. No actual EVA label string is read by this generator or anywhere else in
this package.

## Fix to a known defect (F013)

v3_design's `synthetic_generator.py` used a hand-tuned discrete distractor-length distribution
(`[3-8]` weighted) that did not match the real histogram (which includes lengths 1, 2, 9, 10).
`generator.py`'s `_weighted_choice()` samples distractor and lexicon-form lengths directly from
`REAL_SCOPE_STRUCTURAL_MANIFEST.json`'s `length_distribution.histogram`, so the synthetic length
profile matches the real one by construction rather than by hand-tuning.

## Generation procedure (`generate(GeneratorConfig)`)

1. Draw a random valid `true_table` (respecting `mapping_mode`) from a disjoint 16-letter synthetic
   source alphabet onto the real EVA target alphabet.
2. Build `n_identities` lexicon entries, each with 1-2 forms whose lengths are drawn from the real
   length histogram.
3. Encode each identity's primary form under `true_table` to get its "true" label token.
4. Populate `n_occurrences` (default 57, matching the real scope) label slots:
   - `(1 - unmatched_rate)` fraction get a real identity's true-encoded token (assigned respecting
     `capacity_policy`), optionally corrupted by `noise_rate` (a single random symbol substitution
     — this deliberately caps achievable recovery at high noise, which is realistic, not a bug).
   - the remainder are **distractors**: random EVA-alphabet strings with lengths drawn from the same
     real histogram, `true_identity=None`.
5. Page assignment follows the real 30/27 (`f68r1`/`f68r2`) split by default.

## Physical separation (`split_public_private`)

Splits any generated instance into:
- **public**: lexicon + labels with `true_identity` stripped — everything a solver may see.
- **private**: `true_table` + the true-identity map + the generating seed — held out until reveal.

`SEALED_BENCHMARK_PROTOCOL.md` enforces this split as separate files handled by separate pipeline
stages, not just as separate dict keys inside one process.
