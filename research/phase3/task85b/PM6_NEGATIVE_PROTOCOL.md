# PM6-v2 finite-complement negative protocol

## Definition and constructibility

For corpus alphabet `A`, token length `l`, and observed type set `V_l`, the exact negative space is

`C_l = A^l \\ V_l`, with size `|C_l| = |A|^l - |V_l|`.

Alphabet, normalization, allowed lengths, and `V_l` are frozen from DEVELOPMENT plus the candidate-independent corpus split; no model or Voynich value participates. Big integers are used for `|A|^l`. A lexicographic rank/unrank routine maps integers to `A^l`; a complement select routine returns the r-th string not in sorted `V_l`. This proves every requested draw exists whenever `|C_l|>0` and avoids rejection-attempt caps.

For every eligible heldout positive occurrence of length `l`, draw one value uniformly from `C_l` **with replacement**, using seed namespace `PM6V2/<corpus>/<split>/<replicate>/<l>/<occurrence>`. Replacement removes the v1 requirement for as many unique negatives as positives. Duplicates remain occurrence-weighted. The negative cannot be an observed type by construction.

Lengths with `|C_l|=0` are saturated and explicitly unavailable. PM6 construction is available only when eligible strata cover at least 80% of heldout occurrences, contain at least 100 matched pairs, and include at least two lengths. Otherwise record `NEGATIVE_TEST_NOT_IDENTIFIABLE`. This is a measurement result, never `MODEL_FAIL`. The same rules apply without language- or target-specific branches to synthetic, English, Latin, Sanskrit, and any later Voynich corpus.

## Score and acceptance

Persist separately:

- `PM6_CONSTRUCTION_AVAILABLE`: complement and coverage requirements pass;
- `PM6_SCORE_VALID`: every positive/negative has a model score and stratified AUC is finite;
- `PM6_ACCEPTED`: the valid score passes the frozen threshold.

Within each length, compute occurrence-weighted AUC with half credit for ties. Aggregate length AUCs by frozen heldout positive mass over eligible lengths. Acceptance requires the 2,000 paired-bootstrap 95% lower bound to exceed 0.5 and the frozen development label-permutation q0.95 threshold. Construction failure or invalid score yields NOT_ASSESSABLE; construction plus valid score below threshold yields FAIL.

## Saturation preflight

Development fixtures cover empty vocabulary, fully saturated one-length vocabularies, mixed saturated/unsaturated lengths, alphabet size one, large `|A|^l`, duplicate draws, and Unicode alphabets. Before confirmation, the manifest builder writes per-length alphabet size, observed count, total space, complement size, requested draws, coverage, and a proof hash. Confirmation is forbidden unless every scheduled corpus either passes constructibility or was preregistered as an applicability test expected to return `NEGATIVE_TEST_NOT_IDENTIFIABLE`; known-correct recovery controls must pass.

This protocol was selected for mathematical coverage and fixed missingness semantics, not because it changes Task86R/Task86C outcomes. Those outcomes are not recomputed.
