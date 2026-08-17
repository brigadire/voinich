# Higher-order sequential dependence validation

Confirmatory test of whether the first token of a frozen n>=3 replicated sequence A B C carries information about the third token C beyond what the second token B alone predicts, i.e. P(C|A,B) vs P(C|B). Corpus SHA256: `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`. No new bigrams, trigrams or n-grams were discovered; every candidate below is read programmatically from the previous audit.

## Frozen inventory

- `ol s aiin` (secondary): shuffle FDR q=0.006493506493506494, Markov block p=0.2057942057942058, 5 canonical occurrences across 4 physical blocks and 3 joint classes.
- `or or aiin` (primary): shuffle FDR q=0.006493506493506494, Markov block p=0.003996003996003996, 4 canonical occurrences across 4 physical blocks and 3 joint classes.
- `s aiin chey` (primary): shuffle FDR q=0.006493506493506494, Markov block p=0.023976023976023976, 4 canonical occurrences across 3 physical blocks and 3 joint classes.

## Summary table

| sequence | family | occurrences | eligible blocks | joint classes | P(C\|B) pooled | P(C\|A,B) pooled | enrichment | conditional p | conditional q | CMI (bits) | LOBO M2 advantage | sign consistency | jackknife | status |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ol s aiin | secondary | 5 | 1 | 1 | 0.1935483870967742 | 0.6666666666666666 | 3.444444444444444 | 0.1028971028971029 | 0.1028971028971029 | 5.543432839023097 | 0.17 | 1 | SINGLE_BLOCK_SENSITIVE | INSUFFICIENT_SUPPORT |
| or or aiin | primary | 4 | 1 | 1 | 0.0410958904109589 | 0.3333333333333333 | 8.11111111111111 | 0.2191780821917808 | 0.2191780821917808 | 5.8882222374665965 | 0.25 | 1 | SINGLE_BLOCK_SENSITIVE | INSUFFICIENT_SUPPORT |
| s aiin chey | primary | 4 | 3 | 3 | 0.04878048780487805 | 0.23076923076923078 | 4.730769230769231 | 0.009799020097990201 | 0.019598040195980403 | 5.045210910126452 | 0.65 | 0.6666666666666666 | stable | POSITION_DEPENDENT |

## Per-candidate discussion

### `ol s aiin`

Question: does aiin depend on ol once s is already known?

Pooled across 1 eligible physical blocks (count(A,B)>=3, count(B)>=10): P(aiin|s)=0.1935483870967742, P(aiin|ol,s)=0.6666666666666666, enrichment=3.444444444444444. The conditional-neighbor permutation null around s (1000 permutations) gives empirical p=0.1028971028971029 (family-corrected q=0.1028971028971029). Leave-one-physical-block-out prediction favors the second-order model in 1/6 tested blocks (mean delta log loss=-0.9560909772827478 bits). Sign of the effect is consistent in 100% of eligible blocks, spanning 1 distinct joint metadata classes (cross-Currier=false, cross-hand=false). Jackknife removing each eligible block one at a time: enrichment 0-0 (median 0), CMI 5.4378409907394465-5.4378409907394465 bits (median 5.4378409907394465) - SINGLE_BLOCK_SENSITIVE - sign flips when at least one block is removed.

**Diagnostic status: INSUFFICIENT_SUPPORT.** Fewer than 3 eligible physical blocks are available for `ol s aiin`.

Context-substitution control (Part O): among 12 sufficiently frequent left contexts of s, `ol s` ranks 2 (percentile 91.7, frozen P(aiin|ol,s)=0.625 vs baseline P(aiin|s)=0.1810344827586207) - i.e. this asks whether `ol` is unusual among all `X s` contexts, not merely unusual relative to the whole corpus.

### `or or aiin`

Question: does aiin depend on or once or is already known?

Pooled across 1 eligible physical blocks (count(A,B)>=3, count(B)>=10): P(aiin|or)=0.0410958904109589, P(aiin|or,or)=0.3333333333333333, enrichment=8.11111111111111. The conditional-neighbor permutation null around or (10000 permutations) gives empirical p=0.2191780821917808 (family-corrected q=0.2191780821917808). Leave-one-physical-block-out prediction favors the second-order model in 1/4 tested blocks (mean delta log loss=-1.0302036153294951 bits). Sign of the effect is consistent in 100% of eligible blocks, spanning 1 distinct joint metadata classes (cross-Currier=false, cross-hand=false). Jackknife removing each eligible block one at a time: enrichment 0-0 (median 0), CMI 5.680984611235542-5.680984611235542 bits (median 5.680984611235542) - SINGLE_BLOCK_SENSITIVE - sign flips when at least one block is removed.

**Diagnostic status: INSUFFICIENT_SUPPORT.** Fewer than 3 eligible physical blocks are available for `or or aiin`.

Context-substitution control (Part O): among 17 sufficiently frequent left contexts of or, `or or` ranks 1 (percentile 100.0, frozen P(aiin|or,or)=0.6666666666666666 vs baseline P(aiin|or)=0.13680781758957655) - i.e. this asks whether `or` is unusual among all `X or` contexts, not merely unusual relative to the whole corpus.

### `s aiin chey`

Question: does chey depend on s once aiin is already known?

Pooled across 3 eligible physical blocks (count(A,B)>=3, count(B)>=10): P(chey|aiin)=0.04878048780487805, P(chey|s,aiin)=0.23076923076923078, enrichment=4.730769230769231. The conditional-neighbor permutation null around aiin (10000 permutations) gives empirical p=0.009799020097990201 (family-corrected q=0.019598040195980403). Leave-one-physical-block-out prediction favors the second-order model in 11/17 tested blocks (mean delta log loss=0.5219575631452648 bits). Sign of the effect is consistent in 67% of eligible blocks, spanning 3 distinct joint metadata classes (cross-Currier=true, cross-hand=true). Jackknife removing each eligible block one at a time: enrichment 2.944444444444444-6 (median 5.175), CMI 4.924733106407539-5.080272986698707 bits (median 5.049681404765716) - stable across all removals.

**Diagnostic status: POSITION_DEPENDENT.** The apparent effect concentrates near block or line boundaries (block-position TVD=0.571, line-position TVD=0.226) rather than holding as a general second-order transition.

Context-substitution control (Part O): among 13 sufficiently frequent left contexts of aiin, `s aiin` ranks 2 (percentile 92.3, frozen P(chey|s,aiin)=0.09523809523809523 vs baseline P(chey|aiin)=0.02456140350877193) - i.e. this asks whether `s` is unusual among all `X aiin` contexts, not merely unusual relative to the whole corpus.


## Interpretation guardrails

A HIGHER_ORDER_REPLICATED status means the sequence exhibits replicated higher-order conditional dependence, not that "ol s aiin, or or aiin, s aiin chey is a rule". This audit performs no new sequence discovery, tests only the frozen candidates listed above, and establishes nothing about natural language, grammar, operator/operand structure, or decipherment.
