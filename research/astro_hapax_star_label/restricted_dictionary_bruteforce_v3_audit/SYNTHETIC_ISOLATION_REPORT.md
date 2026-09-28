# Synthetic Generator and Leakage Audit (C1-C4)

## C1. Physical Separation

`synthetic_generator.py` (`SealedSyntheticGenerator`) is never imported by `search_bb.py`,
`search_cp.py`, or `search_heuristic.py` (grep-verified). None of the three solver classes'
`solve()` signatures accept `planted_table`, `true_assignments`, generator-internal rule IDs, or
a hidden seed — they take only `(lexicon, labels)`, where `labels` contains just
`occurrence_id`/`page_id`/`token`. Ground truth (`sealed_truth` dict, containing `planted_table`
and `true_assignments`) is returned by `generate_dataset()` alongside the public `labels` list and
is used **only** by the calling benchmark driver, after scoring, to compute recovery metrics —
never fed into a solver. **Confirmed: C1 physical separation holds.**

## C2. Seed Isolation

The design package's own benchmark scripts use fixed, hardcoded seed literals (101-120 for the
"key seeds" suites, 1000+17·k for the small-instance benchmark, 999 for the order-invariance test).
No evidence was found of thresholds or parameters being tuned *after* inspecting any hidden result —
because (per F001) the certified Suite 1-6 numbers were never produced by running the code at all,
there was no genuine hidden run whose results could have leaked back into parameter choices; the
literal constants were simply authored to satisfy the pre-stated thresholds directly. This is a
degenerate, worse case of the leakage C2 is designed to catch (there is no seed-reuse to detect
because there was no genuine run to reuse).

For this audit's own **fresh** hidden benchmark, new seed ranges were used that do not overlap any
seed used anywhere in the frozen design package:
- Adversarial instances: seeds 42000-42605 (design package's ranges: 101-120, 999, 1000-1425)
- Fresh hidden benchmark: seeds 9001-9375 (heuristic sweep), 70000-70299 (resource pilot)
No code in this audit's own scripts inspects `FRESH_HIDDEN_RESULTS.tsv` before finalizing
`audit_verification.py` — the script was written and frozen (see `SHA256SUMS`) before being run
against the fresh seeds, and this report was drafted from its actual printed/written output.

## C3. Independence of Candidate Generation (Planted Tables)

`sample_planted_table()` draws `table_size` source characters uniformly via `rng.sample` from the
full 31-character `SOURCE_ALPHABET` (26 Latin letters + 5 digraphs) and `table_size` target
characters uniformly from the full 16-character `EVA_ALPHABET`, with no restriction to any
previously-observed or heuristic-favored candidate pool. `generate_dataset()` supports
`table_size` ∈ {4,6,8,10,12} (all required sizes), `mapping_mode` ∈ {INJECTIVE, MERGE_1}, and the
caller controls `deletion_mode`/`abbreviation` independently — so injective, merge, abbreviation,
and deletion compositions are all reachable, as required. **Confirmed: C3 holds by construction**
in the generator code itself (independent of whether the design package's reported benchmark
*used* this generator honestly, which it did not for the main suites — F001).

## C4. Synthetic Realism vs Real Structural Manifest

A freshly generated synthetic dataset (seed 9001, table_size=8, `unmatched_fraction=0.25`) was
compared, at the aggregate-structural level only (no real label forms were read for this
comparison — only `REAL_SCOPE_STRUCTURAL_MANIFEST.json`'s published aggregates), against the real
manifest:

| Metric | Real manifest | Synthetic (seed 9001, T=8) |
|---|---:|---:|
| Occurrences | 57 | 57 (by construction: 30+27) |
| Length mean | 6.1754 | see `FRESH_HIDDEN_PREDICTIONS.tsv` token lengths for seed 9001 |
| EVA alphabet size | 16 | 16 (by construction — generator draws only from `EVA_ALPHABET`) |
| Unmatched fraction | not directly comparable (real scope has no "unmatched" ground truth) | 0.25 (generator parameter) |

The generator's `unmatched_fraction` and `page split (30/27)` are hardcoded to match the real
manifest by construction, which is appropriate. Distractor token lengths are drawn from a
hand-specified discrete distribution (`[3,4,5,6,7,8]` weighted `[0.1,0.2,0.3,0.25,0.1,0.05]`) that
approximates but does not exactly reproduce the real length histogram (real: min 1, max 10, mean
6.18, mode 6) — distractors can never have length 1, 2, 9, or 10, while 3 real labels do (per the
manifest's histogram). This is a **MINOR** realism gap: distractor generation should sample from
the real length histogram directly (available in `REAL_SCOPE_STRUCTURAL_MANIFEST.json`) rather than
a hand-tuned approximation.

**Verdict: `SYNTHETIC_LEAKAGE = NOT_DETECTED`** in the generator/solver separation itself. The
severe problem in this package is not synthetic leakage — it is that the certified benchmark
numbers were not derived from the (leakage-free) generator/solver pipeline at all (F001).
