# SCORER_SPEC — clean-room scorer

## Purpose

One scorer implementation (`scripts/scorer.py:Scorer`) is used by every synthetic dev run, every
sealed hidden run, the discriminative benchmark, and the null pipeline in this package. There is no
second, informally-diverging scoring code path anywhere in `remediation_v1` — this is the specific
defect class this document exists to prevent (COMPONENT_REUSE_REGISTRY.tsv's rejection of
`SYNTHETIC_RECOVERY_RESULTS` shows what happens when reported numbers and executed code diverge).

## Model

A candidate solution is a **global substitution table** `table: source_grapheme -> target_symbol`
(target alphabet drawn from `EVA_ALPHABET` in `REAL_SCOPE_STRUCTURAL_MANIFEST.json`, 16 symbols).
Scoring a table proceeds in four stages, all inside `Scorer.evaluate()`:

1. **Encoding** (`encode_word`): every lexicon form is run through the table left-to-right,
   longest-key-first (so multi-character source graphemes like `kh`, `sh` match before single
   letters). Unmapped source characters are handled per `deletion_mode`:
   - `NONE` — kept as-is (falls through untranslated).
   - `DROP_UNMAPPED` — deleted.
   - `SELECTIVE_VOWEL_DROP` — unmapped vowels deleted, unmapped consonants kept.
   `abbreviation` then optionally truncates the encoded string (`SUSPENSION_1`/`SUSPENSION_2` drop
   trailing characters, `PREFIX_4` keeps only the first 4).
2. **Index construction**: encoded forms are inverted into `encoded_string -> {identity_ids}`.
3. **Bipartite assignment** (`maximum_bipartite_matching`): each label occurrence can match any
   identity whose *some* form encodes to the label's token. Capacity is exactly 1 per identity,
   scoped either per page (`PER_PAGE_CAPACITY_1`) or globally (`GLOBAL_CAPACITY_1`). A label that
   matches no identity, or loses the augmenting-path competition for its only candidate identity,
   is **unmatched** — this is a first-class outcome, not an error; `unmatched_occurrence_ids` is
   always reported.
4. **Complexity and fitness**: `complexity = |table| + mode_penalties`; `fitness = 100*matched -
   complexity` if the table is a *valid* table under its declared `mapping_mode`, else `-1e9`
   (`is_valid_table` — see below).

Table validity: `INJECTIVE` requires every target symbol used at most once. `MERGE_1` is **bounded**
merge: at most one target symbol may be shared by exactly two source graphemes (never three or
more, never two separate merged pairs at once). This bound is enforced by `is_valid_table` and
independently re-derived (not copied) in `independent_scorer.is_valid_table_v2`.

## Train/held-out split

`Scorer.evaluate(table, lexicon, labels, held_out_labels=...)` applies the *same* table/index
(built and optimized only against `labels`, the train set) to a disjoint `held_out_labels` set, and
reports `held_out_assignment_accuracy` against each held-out label's `true_identity` (present only
on synthetic data with known ground truth). This is what makes "train-only optimization" and
"held-out assignment accuracy" measurable — a solver that is allowed to see held-out ground truth
during its own optimization loop is not calling this API correctly, and `SCIENTIFIC_SAFETY_TESTS.md`
includes a check for that.

## Reuse and provenance

`encode_word`, `compute_complexity`, `maximum_bipartite_matching`, `_match_single_capacity` are
reused verbatim from `restricted_dictionary_bruteforce_v3_design/scripts/engine.py`
(`COMPONENT_REUSE_REGISTRY.tsv`: `BIPARTITE_MATCHER`, `CAPACITY_POLICY`, both
`INDEPENDENTLY_VALIDATED`). `is_valid_table`, the train/held-out split, and the single-entrypoint
`Scorer` class wrapping everything are new in this package.

## Independent parity check

`scripts/independent_scorer.py` reimplements the entire pipeline from scratch using deliberately
different mechanisms: `re`-compiled alternation instead of the manual longest-prefix scan for
encoding, and `networkx`'s Hopcroft-Karp bipartite matcher instead of the manual Kuhn/DFS
augmenting-path search for assignment. `scripts/run_scoring_parity.py` generates 500 randomized
small instances (2-6 identities, 3-10 labels, 1-2 pages, random table sizes 0-7 including both
`INJECTIVE` and `MERGE_1` tables, all three deletion modes, all four abbreviation modes, both
capacity policies) and compares `matched`, `complexity`, `fitness`, and table validity between the
two implementations (exact `assignments` dicts are allowed to differ when multiple maximum matchings
exist — the parity criterion is on scores, not on which specific tie-broken assignment is returned).

**Result:** `SCORING_PARITY_RESULTS.tsv`, 500/500 trials agree, 0 mismatches
(`SCORING_PARITY=PASS`). Reproduce with:
```
python3 scripts/run_scoring_parity.py 500 SCORING_PARITY_RESULTS.tsv
```
