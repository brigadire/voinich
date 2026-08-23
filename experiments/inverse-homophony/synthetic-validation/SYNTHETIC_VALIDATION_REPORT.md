# Task57 Synthetic Validation Report

Method version: `inverse-homophony-v1`
Git commit: `9737c991613eb4fda6f18aec999a13a34225ca0f` (dirty: true)

## Development threshold fit

Pooled development AUC: 0.7807 (4004 true pairs, 4004 false pairs)

Frozen tau: 0.2734, MinSupport: 5, MaxClassFraction: 0.15, MinEntropyFraction: 0.50

## Validation gate (task57 section 20)

**Overall: FAIL**

| Criterion | Pass | Detail |
|---|---|---|
| class_recovery_beats_random | true | 9/9 validation corpora: recovered F1/ARI > mean random-partition F1/ARI |
| structural_recovery_beats_baselines | false | 3/9 validation corpora: recovered vocab_size & significant_bigram_fraction closer to plaintext than NO_COLLAPSE and mean RANDOM_PARTITION |
| transfers_beyond_doyle | false | at least one non-Doyle validation corpus (Longfellow/Astafiev) passes criterion 2 |
| predicted_direction | false | 6/9 validation corpora: recovery_fraction>0 for vocab_size and significant_bigram_fraction |
| no_trivial_collapse | true | 9/9 validation corpora: recovered class count in (10, NO_COLLAPSE class count) |

## Pair discrimination (section 19)

| split | label | cipher types | true pairs | false pairs | AUC |
|---|---|---|---|---|---|
| development | doyle_h004_uniform | 9599 | 2174 | 2174 | 0.7728 |
| development | doyle_h004_weighted | 9288 | 1830 | 1830 | 0.7897 |
| validation | doyle_h006_uniform | 11074 | 3857 | 3857 | 0.7382 |
| validation | doyle_h008_uniform | 12250 | 5701 | 5701 | 0.7086 |
| validation | doyle_h006_weighted | 10671 | 3205 | 3205 | 0.7588 |
| validation | doyle_h008_weighted | 11776 | 4743 | 4743 | 0.7299 |
| validation | doyle_freq_v1_hmax004_uniform | 8456 | 1383 | 1383 | 0.7842 |
| validation | doyle_freq_v1_hmax006_uniform | 10285 | 2997 | 2997 | 0.7487 |
| validation | doyle_freq_v1_hmax008_uniform | 11528 | 4810 | 4810 | 0.7196 |
| validation | longfellow_h004_uniform | 9519 | 1572 | 1572 | 0.9198 |
| validation | astafiev_h004_uniform | 13365 | 4953 | 4953 | 0.8800 |

Full per-corpus/per-method rows: class_recovery.tsv, structural_recovery.tsv, baseline_comparison.tsv, null_distribution.tsv.

## Diagnosis

Mean validation pair-discrimination AUC is 0.7764 (well above the 0.5 chance level), yet mean recovered-partition pairwise precision is only 0.0028 (recall 0.2022). This combination - real above-chance separability between individual true/false pairs, but a clustering result dominated by false merges - is the signature of the merge-evidence features (local predecessor/successor/distance-context similarity) picking up ordinary distributional similarity between words that are not homophones (e.g. two different function words that share syntactic contexts), not only true homophones of one plaintext type. The clustering step cannot distinguish 'these two cipher types are homophones of the same plaintext unit' from 'these two cipher types are used in similar grammatical positions' using local context alone. Per task57 section 21, this is the honest outcome: current ciphertext-only features are insufficient for blind inverse-homophony recovery to the standard task57 section 20 requires, and no threshold/feature change was made after seeing this number.

Gate FAILED. Per task57 section 21, this is a valid result: current ciphertext-only features are insufficient for blind inverse-homophony recovery under this frozen method. Voynich was not analyzed.
