# Restricted STAR LABEL hapax enrichment

The frozen scope contains 57 STAR LABEL token occurrences: 30 on f68r1 and 27 on f68r2, with 26 corpus-wide hapax tokens (45.6%). Controls were constructed without using brute-force scores or dictionary matches: all frozen introductory P/Pb text on the two pages (68 tokens), both circular C inscriptions (29 tokens), section-A astronomical tokens outside the target (901 tokens), and a deterministic length-matched sample from non-section-A text (57 tokens).

| Control | Hapax rate | Risk ratio | One-sided exact pooled permutation p |
|---|---:|---:|---:|
| Intro prose | 19/68 = 27.9% | 1.63 | 0.0312 |
| Circular text | 12/29 = 41.4% | 1.10 | 0.4439 |
| Other astronomical tokens (target tuples removed) | 212/844 = 25.1% | 1.82 | 0.0010 |
| Length-matched other-section tokens | 11/57 = 19.3% | 2.36 | 0.0024 |

Holm-adjusted p-values are approximately 0.062 (intro), 0.444 (circular), 0.0039 (other astronomical), and 0.0071 (length-matched); BH gives approximately 0.041, 0.444, 0.0020, and 0.0048 respectively. Thus the robust signal is against background astronomical and length-matched other-section text, not against circular text. These are pooled token-level comparisons; the two pages and the circular inscriptions are not independent large samples, so the result is an enrichment signal rather than a semantic or decipherment claim.

The exact pooled permutation calculation uses the combined binary hapax labels and the fixed target/control sample sizes (100,000-replicate budget recorded for comparability; the exact hypergeometric tail is evaluated). No M0/M1 result is used in the test. M2R is therefore justified as a conditional follow-up hypothesis, not yet as an automatic next step: first verify the scope/control definitions and, preferably, repeat with label/line-block-aware inference.
