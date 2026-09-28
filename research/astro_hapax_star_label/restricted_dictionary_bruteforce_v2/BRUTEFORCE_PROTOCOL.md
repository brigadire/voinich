# Restricted dictionary brute-force v2 protocol

This protocol is frozen before production. The input target is copied byte-for-byte
from v1's verified `TARGET_SCOPE.tsv`: 57 occurrences, 30 on f68r1, 27 on f68r2,
and 26 corpus-wide hapax. The 31 canonical star identities and all attestations are
copied from v1's frozen lexicon. v2 may append only rows whose historical source,
date, language and stable canonical identity are recorded in `LEXICON_ADDITIONS.tsv`;
no generated spelling is admitted. A canonical identity has capacity one despite
multiple attestations.

## Global writing-system family

The source side is Unicode NFKD, ASCII transliteration, lower case, removal of
spaces/hyphens/editorial punctuation, followed by the frozen v1 orthographic grid.
The target side is the observed EVA alphabet in TARGET_SCOPE. One table is shared
by every term and every page. A table contains at most four source graphemes and
maps each to one EVA symbol; source graphemes are selected from a fixed registry
of single letters and `kh, gh, sh, th, dh` digraphs. Mappings are injective by
default and a `MERGE` family allows at most one collision. Unmapped characters are
deleted only by the single globally declared `DELETE` mode. No table is fitted to
one term, one page, one LABEL, or one attractive correspondence.

The finite systems are the Cartesian product of: v1's 64 orthographic systems;
`TABLE_SIZE` 0, 1, 2, 3, 4; `TABLE_MODE` INJECTIVE or one-merge; `DELETE` KEEP or
DROP_UNMAPPED; and `ABBREVIATION` NONE, DROP_FINAL, PREFIX_4. Table candidates are
enumerated deterministically from the fixed source/target alphabets and a fixed
seeded beam of 256 globally distinct tables per system. Beam retention is by
corpus-independent table complexity and lexicographic table ID, never by target
score. Thus the production grid is finite and frozen; a later wider beam is v3.
Complexity is table entries + merge + deletion + abbreviation cost and the hard
limit is 8. Singleton support is forbidden: every learned mapping must be used by
at least two distinct dictionary identities in training, then is tested unchanged.

The substitution table is learned only from the training page by exhaustive
matching of whole transformed strings under the frozen table candidates. The
held-out page receives no table update. Matching is maximum-cardinality bipartite
matching between canonical identity and LABEL occurrence, with one identity used
once and unmatched allowed. Selection score is `100*matched - complexity`; ties
are lowest system ID, then table ID. Hapax status is post-selection only.

## Validation

Both directions are mandatory: f68r1→f68r2 and f68r2→f68r1. Joint selection is
reported only as corpus-level evidence. Leave-one-out and 200 page-stratified
resamples repeat complete system/table selection. Pair candidates require at least
three supporting neighboring systems, assignment margin, 95% LOO survival, 90%
resample survival, and agreement in both directions. They are labelled only
`CANDIDATE_FORMAL_COMPATIBILITY`.

Synthetic recovery precedes interpretation. At each of 20 seeds and noise levels
0,.10,.25, inject one known global table, plus separately a fixed abbreviation,
plus their composition, into 57 labels split 30/27. Use realistic unmatched items.
Require recovery of the planted table family/equivalent table on ≥18/20 seeds,
≥80% joint and ≥60% held-out true pairs. Each synthetic dataset has 99 complete
length/endpoint, unigram and bigram nulls with full model selection. Any failure
sets `ENGINE_SENSITIVITY_INSUFFICIENT` and prevents a real-data `NO_SIGNAL` claim.

Every real null repeats the complete table/grid selection, not merely scoring a
chosen production system. Families are shuffled LABEL, shuffled identity names,
length/endpoint synthetic, unigram-preserving synthetic, bigram-preserving
synthetic, historical nonastronomical names, medieval Latin, Arabic-Latin,
other-section LABEL, circular text and introductory prose. Main families have
10,000 seeds; exact invariance controls are reported separately. The same train,
held-out, matching, beam, complexity, stopping and checkpoint rules apply.

The primary statistic is selected held-out coverage and complexity-penalized score.
For three endpoints and nine informative null families use 27 comparisons,
add-one empirical p, Bonferroni FWER and BH FDR. `CROSS_PAGE_SIGNAL` requires
corrected p<.05 in both directions, same table family, positive null advantage,
synthetic PASS, and resampling/LOO gates. Otherwise use `CORPUS_LEVEL_SIGNAL_ONLY`,
`NO_SIGNAL`, or `ENGINE_INCONCLUSIVE` exactly as the sensitivity and completeness
gates dictate. No stopping or grid expansion follows observed scores.

The full command journal, frozen inputs, checkpoint hashes and output SHA256 are
mandatory. A partial run is `BLOCKED`; no best system is published until all nulls,
synthetic controls and corrections finish.
