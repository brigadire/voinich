# Task60 Report

## 1. Exact repetition

See EXACT_ADJACENT_REPETITION.tsv, NULL_EXACT_REPETITION.tsv.

| Corpus | R2 | null mean (global) | z (global) | percentile (global) | null mean (lines) | z (lines) | percentile (lines) |
|---|---:|---:|---:|---:|---:|---:|---:|
| Voynich | 0.009119 | 0.003127 | 21.574 | 1.000 | 0.008923 | 0.464 | 0.700 |
| Doyle | 0.000574 | 0.009085 | -18.505 | 0.000 | 0.009309 | -19.209 | 0.000 |
| Longfellow | 0.001098 | 0.014427 | -23.330 | 0.000 | 0.009261 | -17.470 | 0.000 |
| Astafiev | 0.000233 | 0.005764 | -21.999 | 0.000 | 0.010246 | -28.395 | 0.000 |

## 2. Long runs

See EXACT_RUNS.tsv, RUN_DISTRIBUTION.tsv.

| Corpus | max run | runs>=3 | runs>=4 | runs>=5 |
|---|---:|---:|---:|---:|
| Voynich | 4 | 11 | 1 | 0 |
| Doyle | 2 | 0 | 0 | 0 |
| Longfellow | 3 | 5 | 0 | 0 |
| Astafiev | 2 | 0 | 0 | 0 |

## 3. Near repetition

See EDIT_DISTANCE_DISTRIBUTION.tsv, EDIT_DISTANCE_ONE.tsv, NULL_NEAR_REPETITION.tsv.

| Corpus | P(d<=1) | null mean (global) | z (global) | null mean (lines) | z (lines) | freq/length-matched null rate |
|---|---:|---:|---:|---:|---:|---:|
| Voynich | 0.058656 | 0.025631 | 42.312 | 0.050938 | 7.705 | 0.034936 |
| Doyle | 0.012087 | 0.025521 | -18.111 | 0.024929 | -18.379 | 0.012299 |
| Longfellow | 0.007616 | 0.027990 | -24.223 | 0.021795 | -20.354 | 0.007834 |
| Astafiev | 0.015685 | 0.031466 | -31.228 | 0.035385 | -37.824 | 0.014078 |

## 4. Edit operation structure

See EDIT_OPERATION_POSITION.tsv, SUBSTITUTION_MATRIX.tsv, EDIT_FAMILIES.tsv, NEAR_REPEAT_CHAINS.tsv. Total INSERTION vs DELETION counts in EDIT_OPERATION_POSITION.tsv directly give the directional short->long vs long->short adjacency asymmetry (task60 section 25): more INSERTION events than DELETION means the right member of an adjacent pair is more often the longer one.


## 5. Homophony controls

See HOMOPHONY_RUN_DOSE_RESPONSE.tsv, HOMOPHONY_THEORETICAL_VS_OBSERVED.tsv. Near-repetition on Task46/55 corpora is NOT_APPLICABLE_OPAQUE_TOKENS (task60 section 27); glyph-level H2/H4/H8 controls (task60 section 28, shared with Task59's fixed generator) are used instead.


**Near-repetition homophony classification (task60 section 42): INCOMPATIBLE_WITH_SIMPLE_RANDOM_HOMOPHONY.** Plaintext (Doyle, glyph-level) P(d<=1) = 0.012087; Voynich's own P(d<=1) = 0.058656 (already above the natural-language range). H=2: 0.005569. H=4: 0.003172. H=8: 0.001773. Increasing H monotonically *reduces* near-repetition below the plaintext baseline, moving away from - not toward - Voynich's enriched value; simple position-independent homophony is therefore not a sufficient mechanism for this specific property. This does not rule out position-dependent homophony, structured token formation, natural-language morphology, or other cipher systems (task60 section 42).

## 6. Illustration labels

See LABEL_CORPUS.tsv, LABEL_REPETITION.tsv, LABEL_RUNNING_COMPARISON.tsv, LABEL_VOCABULARY_OVERLAP.tsv.

Caveat (task60 section 10): most labels are one or two tokens long, leaving only 170 valid within-label adjacent pairs corpus-wide; the label AdjacentRepeatRate/EditLe1Rate estimates in LABEL_RUNNING_COMPARISON.tsv have correspondingly low statistical power.


## 7. Relation to Task58

See TASK58_COMPARISON.tsv.

| Corpus | Task58 token-order MI | Task58 token share | Task58 glyph-edge MI |
|---|---:|---:|---:|
| Voynich | 3.053200 | 0.011018 | 0.251730 |
| Doyle | 3.274235 | 0.094041 | 0.085627 |
| Longfellow | 3.480215 | 0.024153 | 0.159878 |
| Astafiev | 4.343707 | 0.186333 | 0.381453 |

## 8. Relation to Task59

See TASK59_COMPARISON.tsv.

| Corpus | Task59 high-freq specialists | Task59 weighted entropy |
|---|---:|---:|
| Voynich | 6 | 0.596409 |
| Doyle | 0 | 0.887697 |
| Longfellow | 1 | 0.834766 |
| Astafiev | 1 | 0.723373 |

Task60 section 37 asks whether edit-operation position concentrates at a token boundary alongside Task59's positional specialization. For Voynich, EDIT_OPERATION_POSITION.tsv shows INSERTION and DELETION concentrated at BEGIN (not END, and not evenly spread like SUBSTITUTION), i.e. near-repeat pairs differ by an extra/missing glyph disproportionately at the token's *start* - reported as an observation only; this is compatible with, but does not by itself prove, a relationship to Task59's own initial-position specialists.

## 9. Mechanistic interpretation (task60 section 36: critical combination)

Task58 token_share for Voynich: 0.011018 (weak-MI threshold check: true). Exact-repetition enrichment vs both nulls: true. Near-repetition (d<=1) enrichment vs both nulls: true.

All three legs hold: weak global adjacent-token order (Task58 token_share), enriched exact adjacency, and enriched edit-distance-1 adjacency - i.e. low average token-order dependency does not mean an absence of local sequential structure (task60 section 36).

## 10. Limitations

- Illustration labels may not be fully disjoint from the running-text corpus (see METHOD.md's caveat on ivtt -x7 not restricting locus type).
- The edit-family graph and chain extraction (section 23/26) use a deterministic greedy walk, not exhaustive path enumeration; this is exploratory, per the task's own framing.
- No claim is made here about language identity, decipherment, or a specific cipher mechanism (task60 sections 29/30/42/43/46).

## Notes

- Glyph-level homophony control (Doyle, natural characters as glyphs): plaintext P(d<=1)=0.012087.
- Glyph-level homophony control (Doyle, H=2, position-independent, shared with Task59): P(d<=1)=0.005569 (plaintext 0.012087).
- Glyph-level homophony control (Doyle, H=4, position-independent, shared with Task59): P(d<=1)=0.003172 (plaintext 0.012087).
- Glyph-level homophony control (Doyle, H=8, position-independent, shared with Task59): P(d<=1)=0.001773 (plaintext 0.012087).
