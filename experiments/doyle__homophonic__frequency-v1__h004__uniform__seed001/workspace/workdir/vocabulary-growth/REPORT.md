# Vocabulary growth analysis

This is a language-agnostic descriptive corpus analysis. Heaps-law fit, hapax counts, and null effects do not establish language, cipher, generator, or semantic class.

## Fixed parameters

- tokens: 43713
- checkpoints: [100 200 500 1000 2000 4000 8000 16000 32000 43713]
- windows: [500 1000 2000]
- segments: [4 8]
- null permutations: 100
- seed: 1

## Heaps fit

`V(n) = K * n^beta`, K=3.096142844, beta=0.7530863229, R²=0.9970972539, fitting range=[100,43713], points=10.

## Outputs

- `vocabulary_growth.tsv`: observed trajectory and effective beta.
- `frequency_of_frequencies.tsv`: hapax/dis/tris-legomena trajectory.
- `new_type_rate.tsv`: fixed-window productivity.
- `vocabulary_growth_null.tsv`: deterministic shuffled-token null ensemble.
- `segment_vocabulary_growth.tsv`: positional segment analysis.

## Limitations

The observed trajectory depends on token order and the canonical tokenization contract. The shuffled null preserves the token multiset but destroys order. Segment analysis is positional only and uses no manuscript metadata. Corpus-size comparisons must use a common checkpoint; unavailable checkpoints are not zero.
