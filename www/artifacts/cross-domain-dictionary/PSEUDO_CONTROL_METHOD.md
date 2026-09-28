# Pseudo-control method

The search null contains 99 target-blind pseudo-lexicons generated from the frozen matched panels using the previously frozen seeds. Each lexicon preserves the panel size, one form per identity, form-length distribution and aggregate character/repetition constraints while destroying order and real lexical identities. It is not a single global alphabet permutation.

Before search, each pseudo-lexicon was checked for:

- no global alphabet isomorphism to its source panel;
- zero exact overlap with real astronomical, botanical or historical-control forms;
- no duplicate forms within a lexicon;
- deterministic regeneration from its frozen seed.

Every pseudo lexicon was evaluated over all 64 E3 profiles with the same corrected-global scorer, target, capacity rule, search budget, stop conditions and UNKNOWN handling as the three real panels. The model-selection-aware null therefore uses the maximum over profiles for each pseudo lexicon. All 99 runs were valid.

The public `NULL_DISTRIBUTION.tsv` records the resulting maxima; the source remediation package records the identity-level revalidation. Caesar shifts are not used as the primary null.
