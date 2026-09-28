# E3 frozen model semantics

The only pre-registered operation family is the 64-system grid copied from the historical transformation registry. It is fixed before looking at E3 results:

- article: `KEEP` or `DROP_AL` (remove a whole word exactly equal to `al`; retain the original words if this would empty the form);
- orthography: `IDENTITY`, `IJ_UV` (`j→i`, `v→u`), `ARABIC_LATIN` (`kh→h`, `gh→g`, `sh→s`, `th→t`, `dh→d`, then `j→i`, `w→u`), or `VELAR_COLLAPSE` (the preceding Arabic-Latin replacements, then `q→k`, `c→k`);
- vowels: `KEEP` or `CONTRACT_INTERNAL` (remove internal `a,e,i,o,u,y` after the first and before the last character when length exceeds two);
- abbreviation: `NONE`, `SUSPEND_1` (drop the last character only when length exceeds three), `PREFIX_4` (retain four leading characters), or `STRIP_LATIN` (remove the first matching suffix from the frozen Latin-ending list, only if at least three characters remain).

Operations run in that order after lower-case ASCII word normalization and word joining. The scorer is exact `DROP_UNMAPPED`: every source character is either mapped by the global injective table or dropped by this explicit scorer mode. No subsequence heuristic, edit distance, learned rule, or score-driven operation was added.

For each frozen operation profile, paths contain one LABEL occurrence, one canonical identity, one transformed source form, one injective mapping signature, and one exact scorer output. Equivalent paths can be collapsed only when those fields and the output agree; provenance is retained in the path record. Active mapping rules require at least two distinct EVA token types. One identity and one LABEL occurrence may be used at most once. Unmatched occurrences are allowed.

The objective first maximizes matched occurrences, then distinct EVA token types and identities, and finally penalizes mapping count, applied abbreviation, and operation complexity. Any result with coverage at least 6 must trigger a complete model-selection-aware null over the same operation grid and complexity selection.
