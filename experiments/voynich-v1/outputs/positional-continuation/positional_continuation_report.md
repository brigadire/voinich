# Positional continuation validation: "s aiin" -> "chey"

Confirmatory deep-dive on one frozen higher-order-sequence-validate finding: does the continuation distribution after the fixed context `s aiin` depend on structural position, and if so what explains it? Context and target continuation are frozen inputs (task23 section 2); no new n-gram discovery happens here. Corpus SHA256: `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`.

## Counts (section 100)

- total `aiin` occurrences: 285
- total `s aiin` occurrences: 42
- total `s aiin X` occurrences (X present): 42
- total `s aiin chey` occurrences: 4
- physical blocks containing `s aiin`: 17
- joint (Currier x hand) classes: 8
- Currier classes: 7
- hands: 4

Previous experiment (documented context only, not recomputed as an input - task23 section 103): canonical `s aiin chey` occurrences were reported as approximately 4, across 3 eligible physical blocks, with block-position TVD=0.571 and line-position TVD=0.226, giving a POSITION_DEPENDENT status. This run's fresh recount above should be compared against that figure honestly rather than assumed to match.

## A. Is `s aiin` itself positionally specialized?

- line_position: total variation distance between P(position|s,aiin) and P(position|aiin) = 0.1383458646616541 (support: 42 s-aiin occurrences, 285 aiin occurrences).
- block_position_coarse: total variation distance between P(position|s,aiin) and P(position|aiin) = 0.012781954887218033 (support: 42 s-aiin occurrences, 285 aiin occurrences).

## B. Does position change the continuation distribution after `s aiin`?

- line_position: observed I(X;position)=1.5522171247991357 bits, permutation empirical p=0.04789521047895211 (10000 permutations, support=42 occurrences with X present).
- block_position_coarse: observed I(X;position)=1.2345619523656903 bits, permutation empirical p=0.3588641135886411 (10000 permutations).

## C. Is `chey` specifically enriched at some position?

- LINE_START: n=1, chey=1, P(chey|position)=1 vs P(chey|s,aiin)=0.09523809523809523, enrichment=10.5, p=0.503949605039496.
- LINE_EARLY: n=1, chey=0, P(chey|position)=0 vs P(chey|s,aiin)=0.09523809523809523, enrichment=0, p=1.
- LINE_MIDDLE: n=24, chey=3, P(chey|position)=0.125 vs P(chey|s,aiin)=0.09523809523809523, enrichment=1.3125, p=0.6597340265973403.
- LINE_LATE: n=8, chey=0, P(chey|position)=0 vs P(chey|s,aiin)=0.09523809523809523, enrichment=0, p=1.
- LINE_END: n=8, chey=0, P(chey|position)=0 vs P(chey|s,aiin)=0.09523809523809523, enrichment=0, p=1.

## D. Does the same positional effect exist for `aiin` generally?

- LINE_START: aiin n=34, P(chey|aiin,position)=0.08823529411764706 vs P(chey|s,aiin,position)=1, within-position enrichment E(position)=11.333333333333332.
- LINE_EARLY: aiin n=19, P(chey|aiin,position)=0 vs P(chey|s,aiin,position)=0, within-position enrichment E(position)=0.
- LINE_MIDDLE: aiin n=135, P(chey|aiin,position)=0.02962962962962963 vs P(chey|s,aiin,position)=0.125, within-position enrichment E(position)=4.21875.
- LINE_LATE: aiin n=46, P(chey|aiin,position)=0 vs P(chey|s,aiin,position)=0, within-position enrichment E(position)=0.
- LINE_END: aiin n=51, P(chey|aiin,position)=0 vs P(chey|s,aiin,position)=0, within-position enrichment E(position)=0.

## E. After controlling position, does predecessor `s` still matter?

Stratified permutation test (chey ⟂ s | aiin, position), predecessor identity shuffled within (block, position) strata: line_position empirical p=0.009099090090990901 (10000 permutations), block_position_coarse empirical p=0.0025997400259974 (10000 permutations).

## F. Does adding position improve held-out prediction? / G. Does adding predecessor after position improve it further?

Leave-one-physical-block-out, alpha=0.5 smoothing, 28 tested blocks: M2 beats M1 in 25 blocks, M1 beats M2 in 3 (mean delta_21=0.5863121840996592 bits, median=0.6249954251044031). M3 beats M2 in 19 blocks, M2 beats M3 in 9 (mean delta_32=0.10127013558265599 bits, median=0.11599682606926232).

## H. Does the result reproduce across physical blocks?

Eligible blocks (>=1 s-aiin occurrence): 17, positive-sign blocks: 3, negative-sign blocks: 0, neutral: 14, sign consistency=0.17647058823529413.

## Line vs block position (Part M)

Pearson r(normalized_line_position, normalized_block_position) = -0.2069250535556904. Source of the positional effect: **BOTH**.

## Surrounding context (Part O)

- chey: n=4, preceding entropy=2 bits, following entropy=2 bits, unique surrounding contexts=4.
- not_chey: n=38, preceding entropy=5.017535737070865 bits, following entropy=5.090032776601483 bits, unique surrounding contexts=38.

## Part R: two questions, not one p-value (sections 93-95)

**Q1: does position affect continuation after s aiin?** primary (line_position) empirical p=0.04789521047895211; significant at alpha=0.05.

**Q2: does s affect continuation after aiin, once position is controlled?** stratified predecessor empirical p=0.009099090090990901; significant at alpha=0.05.

## Diagnostic status

**NO_POSITIONAL_STRUCTURE**

Support: 17 eligible physical blocks, cross-block sign consistency=0.17647058823529413, M3-vs-M2 held-out win fraction=0.6785714285714286, single-block-sensitive=false, boundary-formula support=false. This audit performs no new sequence discovery; it tests only the frozen `s aiin` -> `chey` finding.
