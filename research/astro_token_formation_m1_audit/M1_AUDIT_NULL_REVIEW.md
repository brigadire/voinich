# M1 audit — null design review

Scope: `RANDOM_VOYNICH_SET`, `SHUFFLED_TERMS`, `PSEUDODICTIONARY` as implemented
in `research/astro_token_formation_m1/main.py` (`null_worker`, `alter_terms`,
`randomize_labels`, `random_pool`). No null replicate was rerun with different
parameters; the frozen `M1_NULL_RESULTS.tsv` (300 rows = 3 controls × 100
replicates) was re-aggregated for this review, and the reaggregation is
reproduced independently in `M1_AUDIT_DIAGNOSTICS.tsv` (`check =
NULL_SUMMARY_CROSSCHECK`).

## What each control models

| Control | Null hypothesis modeled | Length distribution | Repeated-token structure | Grapheme frequency | Historical morphology |
|---|---|---|---|---|---|
| `RANDOM_VOYNICH_SET` | The 20 TRAIN Voynich astronomical tokens carry no more Latin/Arabic-compatible structure than any other real Currier‑A occurrence of matching segmented length. | Preserved exactly (matched on `len(segment(token, TARGET_COMPOSITES))`). | Preserved — real corpus occurrences are drawn, so naturally repeated forms can recur exactly as they do in the real manuscript. | N/A (target side is untouched; only the label token is swapped). | N/A — this control randomizes the **Voynich** side, not the historical terms. |
| `SHUFFLED_TERMS` | The historical corpus's specific *letter order* (not its letter inventory or word count) is what drives any match; scrambling order should destroy true etymological/phonetic structure while keeping the raw material identical. | Word count and per-word length preserved exactly (only intra-word permutation). | N/A (terms, not labels, are altered). | Per-word character multiset preserved exactly (anagram of the real word). | Destroyed by construction — this is the point of the control. |
| `PSEUDODICTIONARY` | A "generic medieval-Latin/Arabic-shaped" vocabulary with the right aggregate letter statistics, but no real lexical identity, is enough to reach comparable coverage. | Word count and per-word length preserved exactly. | N/A | Preserved only at the **class level** (STAR vs. PLANET_MOON pooled letter-frequency table), not per word. | Destroyed — words are frequency-weighted random letter strings. |

All three are legitimate, mutually distinct null hypotheses and none is
redundant with another. `SHUFFLED_ASSIGNMENT` from M0 (permute the canonical
TRAIN pairing) is correctly treated as secondary in M0's own report and is not
re-run in M1 — reasonable, since with anonymous matching there is no true
assignment whose permutation would be informative.

## Correctness checks

1. **Same optimiser, not a fixed comparison.** `null_worker` calls
   `run_search(labels, terms, retain_details=False)`, which reruns the full
   640-pipeline / beam-64 bounded search for every replicate and returns only
   the best-of-search score/coverage (`search_max_train_coverage`). This is
   the correct way to account for the search's own multiple-comparisons
   freedom — a fixed "would this shuffled term match the frozen winning
   mapping" comparison would have been an easier, biased null. **No issue.**
2. **Determinism.** Each replicate's RNG is seeded by
   `sha256(f"{SEED}|{control}|{replicate}")`, so the 300 replicates are
   independently and reproducibly seeded. Confirmed by rerunning
   `null_worker` inputs conceptually via the reproducibility check in
   `M1_AUDIT_DIAGNOSTICS.tsv`.
3. **`RANDOM_VOYNICH_SET` pool realism.** The pool is drawn from
   `occurrence_metadata.jsonl`, filtered to `section == "A"` and tokens whose
   EVA-composite-expanded form is pure `[a-z]+`. This is a **real, observed**
   distribution of Currier-A tokens, not a synthetic uniform-random string —
   if anything this makes the null *harder to beat* than a naive random-string
   null would, because real Voynichese tokens already share the manuscript's
   own internal statistical regularities. No issue; if anything this control
   is conservative in the correct direction.
4. **Class-conditioning is missing from `RANDOM_VOYNICH_SET`.**
   `randomize_labels` draws replacement tokens by target length only; it does
   not condition on the label's `object_class` (STAR vs. PLANET_MOON), unlike
   `SHUFFLED_TERMS`/`PSEUDODICTIONARY`, which do use a class-specific
   alphabet. In this design the *token* side carries no class information for
   `RANDOM_VOYNICH_SET` — only the source-corpus side is class-restricted (via
   `term_class` in `build_edges`) — so this omission is inert: matching is
   already restricted by `label["object_class"]` regardless of how the label
   token was generated. **NO_ISSUE**, but the report does not spell this out,
   which is a documentation gap (`MINOR`).
5. **`PSEUDODICTIONARY` frequency table granularity.** Letters are pooled
   per class across *all* documented forms/variants (`arabic_form`,
   `latin_form`, `medieval_latinized_arabic_forms`,
   `documented_spelling_variants`), which mixes Arabic-transliteration letter
   statistics with Latin planet-name statistics only within, not across,
   `STAR`/`PLANET_MOON`. That is the right level: it matches the true
   generative process where STAR forms are Arabic-transliterated and
   PLANET_MOON forms are Latin. **NO_ISSUE.**
6. **Ceiling coincidence.** All three controls independently reach a **maximum
   observed coverage of exactly 0.500 (10/20)** over 100 replicates each,
   essentially matching the real-corpus result. Given the near-unconstrained
   nature of the one/two-EVA-unit global substitution table (a ~19-letter
   alphabet mapped almost freely to sequences drawn from ~20 EVA units, see
   `M1_AUDIT_MATCHING_REVIEW.md`), a coincidental 50% match on 20 short words
   is exactly what an unconstrained-cipher-fitting exercise predicts under
   any input, real or fake. This is not evidence the null is broken; it is
   corroborating evidence that the *search's own degrees of freedom*, not the
   historical corpus's genuine structure, explain the observed 10/20. This
   matches M1's own conclusion (`NULL_COMPATIBLE`, empirical p = 0.82) and the
   audit finds no reason to weaken that call — if anything the closeness of
   real vs. null strengthens it.
7. **Empirical p resolution.** `(1 + hits) / (n + 1)` over `n=100` gives a
   minimum resolvable p of `1/101 ≈ 0.0099`. The frozen bands require
   `p ≤ 0.05` for `STRONG_CANDIDATE`, which is above this resolution floor, so
   the granularity does not itself block a real signal from clearing the
   band. **NO_ISSUE.**

## Verdict on Section 8 (Null design)

No `CRITICAL` or `MAJOR` issue found. The three controls are well-motivated,
mutually non-redundant, correctly reuse the full bounded optimiser (avoiding
the classic "compare against a fixed answer" null-design bug), and their
result — a null ceiling equal to the observed result — is independently
explained by the mapping model's own degrees of freedom (see
`M1_AUDIT_MATCHING_REVIEW.md`, §"Why 50% is not surprising"). One
`MINOR` documentation gap noted (item 4).
