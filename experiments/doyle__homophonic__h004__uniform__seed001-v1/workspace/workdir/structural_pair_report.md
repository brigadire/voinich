# Structural pair decomposition

Structural similarity is reproduced unchanged from the existing pair dataset. All statements below are formal corpus descriptions; no token meaning is inferred. Context similarities and differences use full distributions, while tables are display-limited. Entropy uses natural logarithms and effective vocabulary is `exp(entropy)`.

## `x009968` / `x020559`

Structural similarity: 0.6854; reliability: 0.8927; normalized graphemic distance: 0.7143; counts: 133/154.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9666 | 0.9901 |
| Left context | 0.6800 | 0.8368 |
| Right context | 0.4096 | 0.8511 |

- Primary component: positional agreement (0.967).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.680.
- Largest left-context difference: x010010 is more frequent for x009968 (absolute probability difference 0.046).

Position summaries (A/B): line-start 0.0526/0.0390, line-end 0.0602/0.0584, mean 5.729/6.026, median 6.000/6.000. Position JS similarity: 0.9666.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.1111 | 0.0743 | +0.0368 |
| x010009 | 0.0556 | 0.0338 | +0.0218 |
| x010010 | 0.0794 | 0.0338 | +0.0456 |
| x008704 | 0.0317 | 0.0270 | +0.0047 |
| x018976 | 0.0238 | 0.0338 | -0.0100 |
| x010011 | 0.0397 | 0.0203 | +0.0194 |
| x018978 | 0.0317 | 0.0203 | +0.0115 |
| x018979 | 0.0317 | 0.0203 | +0.0115 |
| x008706 | 0.0159 | 0.0338 | -0.0179 |
| x019053 | 0.0159 | 0.0135 | +0.0024 |
| x018925 | 0.0079 | 0.0270 | -0.0191 |
| x020828 | 0.0079 | 0.0135 | -0.0056 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010010 | 0.0794 | 0.0338 | +0.0456 |
| x009250 | 0.0000 | 0.0405 | -0.0405 |
| x010008 | 0.1111 | 0.0743 | +0.0368 |
| x009251 | 0.0000 | 0.0338 | -0.0338 |
| x020767 | 0.0317 | 0.0000 | +0.0317 |
| x018926 | 0.0238 | 0.0000 | +0.0238 |
| x020829 | 0.0238 | 0.0000 | +0.0238 |
| x010009 | 0.0556 | 0.0338 | +0.0218 |
| x008707 | 0.0000 | 0.0203 | -0.0203 |
| x009249 | 0.0000 | 0.0203 | -0.0203 |
| x018977 | 0.0000 | 0.0203 | -0.0203 |
| x010011 | 0.0397 | 0.0203 | +0.0194 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000055 | 0.0240 | 0.0345 | -0.0105 |
| x000053 | 0.0240 | 0.0207 | +0.0033 |
| x000054 | 0.0240 | 0.0207 | +0.0033 |
| x018928 | 0.0240 | 0.0207 | +0.0033 |
| x000052 | 0.0160 | 0.0138 | +0.0022 |
| x018931 | 0.0240 | 0.0138 | +0.0102 |
| x009437 | 0.0080 | 0.0138 | -0.0058 |
| x012464 | 0.0080 | 0.0138 | -0.0058 |
| x012467 | 0.0080 | 0.0345 | -0.0265 |
| x014628 | 0.0080 | 0.0138 | -0.0058 |
| x018929 | 0.0080 | 0.0207 | -0.0127 |
| x019270 | 0.0080 | 0.0207 | -0.0127 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x012467 | 0.0080 | 0.0345 | -0.0265 |
| x005177 | 0.0160 | 0.0000 | +0.0160 |
| x010011 | 0.0160 | 0.0000 | +0.0160 |
| x010167 | 0.0160 | 0.0000 | +0.0160 |
| x012045 | 0.0160 | 0.0000 | +0.0160 |
| x012495 | 0.0160 | 0.0000 | +0.0160 |
| x012779 | 0.0160 | 0.0000 | +0.0160 |
| x018977 | 0.0160 | 0.0000 | +0.0160 |
| x019269 | 0.0160 | 0.0000 | +0.0160 |
| x000696 | 0.0000 | 0.0138 | -0.0138 |
| x003286 | 0.0000 | 0.0138 | -0.0138 |
| x006261 | 0.0000 | 0.0138 | -0.0138 |

Context diagnostics: predecessor Jaccard 0.1338, JS 0.4327, entropy A/B 3.890/4.210, effective vocabulary A/B 48.90/67.34; successor Jaccard 0.1224, JS 0.2955, entropy A/B 4.563/4.656, effective vocabulary A/B 95.91/105.27.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `x019053`; right `x000052`.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018929`, `x018928`, `x018931`, `x018930`, `x019270`, `x000054`, `x019269`, `x000055`, `x012698`, `x012697`, `x012699`, `x019268`; right `x000714`, `x009249`, `x000712`, `x000715`, `x009251`, `x009248`, `x012699`, `x000713`, `x019268`, `x008707`, `x010010`, `x009438`.

## `x009971` / `x020558`

Structural similarity: 0.6997; reliability: 0.8719; normalized graphemic distance: 0.7143; counts: 110/149.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9347 | 0.9869 |
| Left context | 0.7083 | 0.8032 |
| Right context | 0.4561 | 0.8255 |

- Primary component: positional agreement (0.935).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.708.
- Largest left-context difference: x018927 is more frequent for x009971 (absolute probability difference 0.049).

Position summaries (A/B): line-start 0.0273/0.0336, line-end 0.0545/0.0805, mean 6.182/6.349, median 6.000/6.000. Position JS similarity: 0.9347.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010010 | 0.0935 | 0.0625 | +0.0310 |
| x010011 | 0.0654 | 0.0625 | +0.0029 |
| x010008 | 0.0654 | 0.0417 | +0.0238 |
| x010009 | 0.0374 | 0.0556 | -0.0182 |
| x008705 | 0.0280 | 0.0417 | -0.0136 |
| x018976 | 0.0561 | 0.0278 | +0.0283 |
| x018979 | 0.0467 | 0.0278 | +0.0190 |
| x008704 | 0.0187 | 0.0278 | -0.0091 |
| x018977 | 0.0187 | 0.0556 | -0.0369 |
| x018925 | 0.0374 | 0.0139 | +0.0235 |
| x018978 | 0.0187 | 0.0139 | +0.0048 |
| x008707 | 0.0093 | 0.0139 | -0.0045 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018927 | 0.0561 | 0.0069 | +0.0491 |
| x018977 | 0.0187 | 0.0556 | -0.0369 |
| x008706 | 0.0000 | 0.0347 | -0.0347 |
| x010010 | 0.0935 | 0.0625 | +0.0310 |
| x018976 | 0.0561 | 0.0278 | +0.0283 |
| x000715 | 0.0000 | 0.0278 | -0.0278 |
| x009248 | 0.0000 | 0.0278 | -0.0278 |
| x009250 | 0.0000 | 0.0278 | -0.0278 |
| x010008 | 0.0654 | 0.0417 | +0.0238 |
| x018925 | 0.0374 | 0.0139 | +0.0235 |
| x018924 | 0.0280 | 0.0069 | +0.0211 |
| x009249 | 0.0000 | 0.0208 | -0.0208 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000054 | 0.0577 | 0.0511 | +0.0066 |
| x000055 | 0.0481 | 0.0219 | +0.0262 |
| x018930 | 0.0288 | 0.0219 | +0.0069 |
| x000053 | 0.0192 | 0.0292 | -0.0100 |
| x012403 | 0.0192 | 0.0365 | -0.0173 |
| x000052 | 0.0096 | 0.0146 | -0.0050 |
| x018929 | 0.0096 | 0.0146 | -0.0050 |
| x000544 | 0.0096 | 0.0073 | +0.0023 |
| x000698 | 0.0096 | 0.0073 | +0.0023 |
| x001055 | 0.0192 | 0.0073 | +0.0119 |
| x011932 | 0.0096 | 0.0073 | +0.0023 |
| x011933 | 0.0096 | 0.0073 | +0.0023 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018925 | 0.0385 | 0.0000 | +0.0385 |
| x012401 | 0.0288 | 0.0000 | +0.0288 |
| x000055 | 0.0481 | 0.0219 | +0.0262 |
| x018928 | 0.0000 | 0.0219 | -0.0219 |
| x019268 | 0.0000 | 0.0219 | -0.0219 |
| x018931 | 0.0288 | 0.0073 | +0.0215 |
| x000545 | 0.0192 | 0.0000 | +0.0192 |
| x012464 | 0.0192 | 0.0000 | +0.0192 |
| x012699 | 0.0192 | 0.0000 | +0.0192 |
| x012776 | 0.0192 | 0.0000 | +0.0192 |
| x012403 | 0.0192 | 0.0365 | -0.0173 |
| x006450 | 0.0000 | 0.0146 | -0.0146 |

Context diagnostics: predecessor Jaccard 0.1802, JS 0.5225, entropy A/B 3.684/3.953, effective vocabulary A/B 39.80/52.10; successor Jaccard 0.1084, JS 0.2859, entropy A/B 4.209/4.524, effective vocabulary A/B 67.26/92.20.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `x018978`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018929`, `x018928`, `x018931`, `x018930`, `x019270`, `x000714`, `x000054`, `x019269`, `x000055`, `x012698`, `x012697`, `x012699`; right `x000714`, `x000712`, `x009251`, `x019269`, `x009248`, `x012697`, `x000713`, `x019271`, `x009439`, `x008707`, `x010008`, `x009436`.

## `x009970` / `x020559`

Structural similarity: 0.6709; reliability: 0.8822; normalized graphemic distance: 0.7143; counts: 119/154.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9417 | 0.9885 |
| Left context | 0.6265 | 0.8199 |
| Right context | 0.4446 | 0.8382 |

- Primary component: positional agreement (0.942).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.626.
- Largest right-context difference: x018931 is more frequent for x009970 (absolute probability difference 0.049).

Position summaries (A/B): line-start 0.0672/0.0390, line-end 0.0672/0.0584, mean 5.462/6.026, median 5.000/6.000. Position JS similarity: 0.9417.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.0541 | 0.0743 | -0.0203 |
| x010009 | 0.0721 | 0.0338 | +0.0383 |
| x010010 | 0.0811 | 0.0338 | +0.0473 |
| x018976 | 0.0450 | 0.0338 | +0.0113 |
| x010011 | 0.0360 | 0.0203 | +0.0158 |
| x018977 | 0.0270 | 0.0203 | +0.0068 |
| x018978 | 0.0360 | 0.0203 | +0.0158 |
| x018979 | 0.0270 | 0.0203 | +0.0068 |
| x016548 | 0.0180 | 0.0135 | +0.0045 |
| x019053 | 0.0360 | 0.0135 | +0.0225 |
| x020831 | 0.0270 | 0.0135 | +0.0135 |
| x008706 | 0.0090 | 0.0338 | -0.0248 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010010 | 0.0811 | 0.0338 | +0.0473 |
| x009250 | 0.0000 | 0.0405 | -0.0405 |
| x010009 | 0.0721 | 0.0338 | +0.0383 |
| x019055 | 0.0360 | 0.0000 | +0.0360 |
| x009251 | 0.0000 | 0.0338 | -0.0338 |
| x008704 | 0.0000 | 0.0270 | -0.0270 |
| x018926 | 0.0270 | 0.0000 | +0.0270 |
| x008706 | 0.0090 | 0.0338 | -0.0248 |
| x019053 | 0.0360 | 0.0135 | +0.0225 |
| x009249 | 0.0000 | 0.0203 | -0.0203 |
| x010008 | 0.0541 | 0.0743 | -0.0203 |
| x018925 | 0.0090 | 0.0270 | -0.0180 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x012467 | 0.0270 | 0.0345 | -0.0075 |
| x000053 | 0.0360 | 0.0207 | +0.0153 |
| x018928 | 0.0270 | 0.0207 | +0.0063 |
| x000054 | 0.0180 | 0.0207 | -0.0027 |
| x000055 | 0.0180 | 0.0345 | -0.0165 |
| x018929 | 0.0180 | 0.0207 | -0.0027 |
| x000052 | 0.0270 | 0.0138 | +0.0132 |
| x012464 | 0.0180 | 0.0138 | +0.0042 |
| x018931 | 0.0631 | 0.0138 | +0.0493 |
| x014628 | 0.0090 | 0.0138 | -0.0048 |
| x018925 | 0.0090 | 0.0138 | -0.0048 |
| x000699 | 0.0180 | 0.0069 | +0.0111 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018931 | 0.0631 | 0.0138 | +0.0493 |
| x012400 | 0.0270 | 0.0000 | +0.0270 |
| x019270 | 0.0000 | 0.0207 | -0.0207 |
| x018930 | 0.0270 | 0.0069 | +0.0201 |
| x012494 | 0.0180 | 0.0000 | +0.0180 |
| x020720 | 0.0180 | 0.0000 | +0.0180 |
| x000055 | 0.0180 | 0.0345 | -0.0165 |
| x000053 | 0.0360 | 0.0207 | +0.0153 |
| x000696 | 0.0000 | 0.0138 | -0.0138 |
| x003286 | 0.0000 | 0.0138 | -0.0138 |
| x006261 | 0.0000 | 0.0138 | -0.0138 |
| x009437 | 0.0000 | 0.0138 | -0.0138 |

Context diagnostics: predecessor Jaccard 0.1567, JS 0.4557, entropy A/B 3.881/4.210, effective vocabulary A/B 48.48/67.34; successor Jaccard 0.1105, JS 0.3071, entropy A/B 4.288/4.656, effective vocabulary A/B 72.86/105.27.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `x016548`; right `x012464`.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018929`, `x018928`, `x018931`, `x018930`, `x019270`, `x000054`, `x019269`, `x000055`, `x012698`, `x012697`, `x012699`, `x019268`; right `x000714`, `x009249`, `x000712`, `x009250`, `x000715`, `x009251`, `x019269`, `x009248`, `x012699`, `x000713`, `x019268`, `x008707`.

## `x009971` / `x020556`

Structural similarity: 0.6846; reliability: 0.8646; normalized graphemic distance: 0.7143; counts: 110/137.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9467 | 0.9857 |
| Left context | 0.6873 | 0.7915 |
| Right context | 0.4198 | 0.8164 |

- Primary component: positional agreement (0.947).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.687.
- Largest left-context difference: x018927 is more frequent for x009971 (absolute probability difference 0.056).

Position summaries (A/B): line-start 0.0273/0.0292, line-end 0.0545/0.0584, mean 6.182/5.942, median 6.000/6.000. Position JS similarity: 0.9467.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.0654 | 0.0602 | +0.0053 |
| x010010 | 0.0935 | 0.0526 | +0.0408 |
| x010011 | 0.0654 | 0.0526 | +0.0128 |
| x018976 | 0.0561 | 0.0376 | +0.0185 |
| x010009 | 0.0374 | 0.0526 | -0.0152 |
| x018979 | 0.0467 | 0.0301 | +0.0167 |
| x008705 | 0.0280 | 0.0376 | -0.0096 |
| x008704 | 0.0187 | 0.0301 | -0.0114 |
| x018977 | 0.0187 | 0.0226 | -0.0039 |
| x018978 | 0.0187 | 0.0150 | +0.0037 |
| x008707 | 0.0093 | 0.0526 | -0.0433 |
| x002557 | 0.0093 | 0.0075 | +0.0018 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018927 | 0.0561 | 0.0000 | +0.0561 |
| x008707 | 0.0093 | 0.0526 | -0.0433 |
| x010010 | 0.0935 | 0.0526 | +0.0408 |
| x009249 | 0.0000 | 0.0376 | -0.0376 |
| x018925 | 0.0374 | 0.0075 | +0.0299 |
| x018924 | 0.0280 | 0.0000 | +0.0280 |
| x009248 | 0.0000 | 0.0226 | -0.0226 |
| x009250 | 0.0000 | 0.0226 | -0.0226 |
| x008845 | 0.0187 | 0.0000 | +0.0187 |
| x009131 | 0.0187 | 0.0000 | +0.0187 |
| x018976 | 0.0561 | 0.0376 | +0.0185 |
| x018979 | 0.0467 | 0.0301 | +0.0167 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000055 | 0.0481 | 0.0465 | +0.0016 |
| x000054 | 0.0577 | 0.0388 | +0.0189 |
| x000053 | 0.0192 | 0.0388 | -0.0195 |
| x018931 | 0.0288 | 0.0155 | +0.0133 |
| x001055 | 0.0192 | 0.0078 | +0.0115 |
| x009250 | 0.0096 | 0.0078 | +0.0019 |
| x012401 | 0.0288 | 0.0078 | +0.0211 |
| x012464 | 0.0192 | 0.0078 | +0.0115 |
| x012492 | 0.0096 | 0.0078 | +0.0019 |
| x012776 | 0.0192 | 0.0078 | +0.0115 |
| x018929 | 0.0096 | 0.0078 | +0.0019 |
| x018930 | 0.0288 | 0.0078 | +0.0211 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018925 | 0.0385 | 0.0000 | +0.0385 |
| x008706 | 0.0000 | 0.0233 | -0.0233 |
| x009437 | 0.0000 | 0.0233 | -0.0233 |
| x018928 | 0.0000 | 0.0233 | -0.0233 |
| x012401 | 0.0288 | 0.0078 | +0.0211 |
| x018930 | 0.0288 | 0.0078 | +0.0211 |
| x000053 | 0.0192 | 0.0388 | -0.0195 |
| x000545 | 0.0192 | 0.0000 | +0.0192 |
| x012121 | 0.0192 | 0.0000 | +0.0192 |
| x012403 | 0.0192 | 0.0000 | +0.0192 |
| x012699 | 0.0192 | 0.0000 | +0.0192 |
| x000054 | 0.0577 | 0.0388 | +0.0189 |

Context diagnostics: predecessor Jaccard 0.1316, JS 0.4714, entropy A/B 3.684/3.964, effective vocabulary A/B 39.80/52.68; successor Jaccard 0.0719, JS 0.2296, entropy A/B 4.209/4.478, effective vocabulary A/B 67.26/88.09.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `x018978`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018929`, `x018928`, `x018931`, `x018930`, `x019270`, `x000712`, `x000714`, `x000054`, `x000715`, `x000713`, `x019269`, `x000055`; right `x000714`, `x009249`, `x019270`, `x000712`, `x009251`, `x019269`, `x012698`, `x009248`, `x012697`, `x000713`, `x019268`, `x008707`.

## `x009969` / `x020558`

Structural similarity: 0.6810; reliability: 0.8618; normalized graphemic distance: 0.7143; counts: 99/149.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9601 | 0.9854 |
| Left context | 0.6571 | 0.7867 |
| Right context | 0.4258 | 0.8131 |

- Primary component: positional agreement (0.960).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.657.
- Largest left-context difference: x010010 is more frequent for x009969 (absolute probability difference 0.042).

Position summaries (A/B): line-start 0.0303/0.0336, line-end 0.0808/0.0805, mean 5.717/6.349, median 5.000/6.000. Position JS similarity: 0.9601.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010010 | 0.1042 | 0.0625 | +0.0417 |
| x010008 | 0.0417 | 0.0417 | +0.0000 |
| x010009 | 0.0417 | 0.0556 | -0.0139 |
| x010011 | 0.0312 | 0.0625 | -0.0312 |
| x018979 | 0.0312 | 0.0278 | +0.0035 |
| x008706 | 0.0208 | 0.0347 | -0.0139 |
| x018977 | 0.0208 | 0.0556 | -0.0347 |
| x018978 | 0.0312 | 0.0139 | +0.0174 |
| x008704 | 0.0104 | 0.0278 | -0.0174 |
| x008705 | 0.0104 | 0.0417 | -0.0312 |
| x008707 | 0.0104 | 0.0139 | -0.0035 |
| x009250 | 0.0104 | 0.0278 | -0.0174 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010010 | 0.1042 | 0.0625 | +0.0417 |
| x019055 | 0.0417 | 0.0000 | +0.0417 |
| x018977 | 0.0208 | 0.0556 | -0.0347 |
| x008705 | 0.0104 | 0.0417 | -0.0312 |
| x010011 | 0.0312 | 0.0625 | -0.0312 |
| x000715 | 0.0000 | 0.0278 | -0.0278 |
| x009248 | 0.0000 | 0.0278 | -0.0278 |
| x018924 | 0.0312 | 0.0069 | +0.0243 |
| x009249 | 0.0000 | 0.0208 | -0.0208 |
| x009251 | 0.0000 | 0.0208 | -0.0208 |
| x019052 | 0.0208 | 0.0000 | +0.0208 |
| x008704 | 0.0104 | 0.0278 | -0.0174 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000053 | 0.0330 | 0.0292 | +0.0038 |
| x000055 | 0.0220 | 0.0219 | +0.0001 |
| x018930 | 0.0440 | 0.0219 | +0.0221 |
| x000052 | 0.0440 | 0.0146 | +0.0294 |
| x018929 | 0.0330 | 0.0146 | +0.0184 |
| x000054 | 0.0110 | 0.0511 | -0.0401 |
| x010009 | 0.0110 | 0.0146 | -0.0036 |
| x010010 | 0.0110 | 0.0146 | -0.0036 |
| x012403 | 0.0110 | 0.0365 | -0.0255 |
| x012773 | 0.0110 | 0.0146 | -0.0036 |
| x018927 | 0.0110 | 0.0146 | -0.0036 |
| x018928 | 0.0110 | 0.0219 | -0.0109 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000054 | 0.0110 | 0.0511 | -0.0401 |
| x000052 | 0.0440 | 0.0146 | +0.0294 |
| x012403 | 0.0110 | 0.0365 | -0.0255 |
| x018930 | 0.0440 | 0.0219 | +0.0221 |
| x012504 | 0.0220 | 0.0000 | +0.0220 |
| x019268 | 0.0000 | 0.0219 | -0.0219 |
| x018929 | 0.0330 | 0.0146 | +0.0184 |
| x012121 | 0.0220 | 0.0073 | +0.0147 |
| x018931 | 0.0220 | 0.0073 | +0.0147 |
| x006450 | 0.0000 | 0.0146 | -0.0146 |
| x008129 | 0.0000 | 0.0146 | -0.0146 |
| x009437 | 0.0000 | 0.0146 | -0.0146 |

Context diagnostics: predecessor Jaccard 0.1864, JS 0.4894, entropy A/B 3.942/3.953, effective vocabulary A/B 51.51/52.10; successor Jaccard 0.1296, JS 0.3226, entropy A/B 4.256/4.524, effective vocabulary A/B 70.50/92.20.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018929`, `x018928`, `x018931`, `x018930`, `x019270`, `x000714`, `x000054`, `x019269`, `x000055`, `x012698`, `x012697`, `x012699`; right `x000714`, `x000712`, `x009250`, `x000715`, `x009251`, `x019269`, `x009248`, `x012699`, `x000713`, `x019271`, `x009439`, `x009436`.

## `x009970` / `x020556`

Structural similarity: 0.6654; reliability: 0.8719; normalized graphemic distance: 0.7143; counts: 119/137.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9373 | 0.9869 |
| Left context | 0.6592 | 0.8034 |
| Right context | 0.3997 | 0.8255 |

- Primary component: positional agreement (0.937).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.659.
- Largest right-context difference: x018931 is more frequent for x009970 (absolute probability difference 0.048).

Position summaries (A/B): line-start 0.0672/0.0292, line-end 0.0672/0.0584, mean 5.462/5.942, median 5.000/6.000. Position JS similarity: 0.9373.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.0541 | 0.0602 | -0.0061 |
| x010009 | 0.0721 | 0.0526 | +0.0194 |
| x010010 | 0.0811 | 0.0526 | +0.0284 |
| x018976 | 0.0450 | 0.0376 | +0.0075 |
| x010011 | 0.0360 | 0.0526 | -0.0166 |
| x018979 | 0.0270 | 0.0301 | -0.0030 |
| x018977 | 0.0270 | 0.0226 | +0.0045 |
| x018978 | 0.0360 | 0.0150 | +0.0210 |
| x008707 | 0.0090 | 0.0526 | -0.0436 |
| x000546 | 0.0090 | 0.0075 | +0.0015 |
| x002778 | 0.0090 | 0.0075 | +0.0015 |
| x008706 | 0.0090 | 0.0075 | +0.0015 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x008707 | 0.0090 | 0.0526 | -0.0436 |
| x008705 | 0.0000 | 0.0376 | -0.0376 |
| x009249 | 0.0000 | 0.0376 | -0.0376 |
| x019053 | 0.0360 | 0.0000 | +0.0360 |
| x019055 | 0.0360 | 0.0000 | +0.0360 |
| x008704 | 0.0000 | 0.0301 | -0.0301 |
| x010010 | 0.0811 | 0.0526 | +0.0284 |
| x009248 | 0.0000 | 0.0226 | -0.0226 |
| x009250 | 0.0000 | 0.0226 | -0.0226 |
| x018978 | 0.0360 | 0.0150 | +0.0210 |
| x018926 | 0.0270 | 0.0075 | +0.0195 |
| x020831 | 0.0270 | 0.0075 | +0.0195 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000053 | 0.0360 | 0.0388 | -0.0027 |
| x018928 | 0.0270 | 0.0233 | +0.0038 |
| x000054 | 0.0180 | 0.0388 | -0.0207 |
| x000055 | 0.0180 | 0.0465 | -0.0285 |
| x018931 | 0.0631 | 0.0155 | +0.0476 |
| x012465 | 0.0090 | 0.0155 | -0.0065 |
| x001786 | 0.0090 | 0.0078 | +0.0013 |
| x009439 | 0.0090 | 0.0078 | +0.0013 |
| x010008 | 0.0090 | 0.0078 | +0.0013 |
| x010167 | 0.0090 | 0.0078 | +0.0013 |
| x012044 | 0.0090 | 0.0078 | +0.0013 |
| x012400 | 0.0270 | 0.0078 | +0.0193 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018931 | 0.0631 | 0.0155 | +0.0476 |
| x000055 | 0.0180 | 0.0465 | -0.0285 |
| x000052 | 0.0270 | 0.0000 | +0.0270 |
| x012467 | 0.0270 | 0.0000 | +0.0270 |
| x008706 | 0.0000 | 0.0233 | -0.0233 |
| x009437 | 0.0000 | 0.0233 | -0.0233 |
| x000054 | 0.0180 | 0.0388 | -0.0207 |
| x012400 | 0.0270 | 0.0078 | +0.0193 |
| x018930 | 0.0270 | 0.0078 | +0.0193 |
| x000699 | 0.0180 | 0.0000 | +0.0180 |
| x012494 | 0.0180 | 0.0000 | +0.0180 |
| x020720 | 0.0180 | 0.0000 | +0.0180 |

Context diagnostics: predecessor Jaccard 0.1681, JS 0.4641, entropy A/B 3.881/3.964, effective vocabulary A/B 48.48/52.68; successor Jaccard 0.1280, JS 0.2995, entropy A/B 4.288/4.478, effective vocabulary A/B 72.86/88.09.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018929`, `x018928`, `x018931`, `x018930`, `x019270`, `x000712`, `x000714`, `x000054`, `x000715`, `x000713`, `x019269`, `x000055`; right `x000714`, `x009249`, `x019270`, `x000712`, `x000715`, `x009251`, `x019269`, `x009248`, `x012699`, `x000713`, `x019268`, `x012696`.

## `x009971` / `x020559`

Structural similarity: 0.6555; reliability: 0.8747; normalized graphemic distance: 0.7143; counts: 110/154.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9359 | 0.9874 |
| Left context | 0.6109 | 0.8078 |
| Right context | 0.4198 | 0.8291 |

- Primary component: positional agreement (0.936).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.611.
- Largest left-context difference: x010010 is more frequent for x009971 (absolute probability difference 0.060).

Position summaries (A/B): line-start 0.0273/0.0390, line-end 0.0545/0.0584, mean 6.182/6.026, median 6.000/6.000. Position JS similarity: 0.9359.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.0654 | 0.0743 | -0.0089 |
| x010009 | 0.0374 | 0.0338 | +0.0036 |
| x010010 | 0.0935 | 0.0338 | +0.0597 |
| x018976 | 0.0561 | 0.0338 | +0.0223 |
| x018925 | 0.0374 | 0.0270 | +0.0104 |
| x010011 | 0.0654 | 0.0203 | +0.0452 |
| x018979 | 0.0467 | 0.0203 | +0.0265 |
| x008704 | 0.0187 | 0.0270 | -0.0083 |
| x018977 | 0.0187 | 0.0203 | -0.0016 |
| x018978 | 0.0187 | 0.0203 | -0.0016 |
| x008707 | 0.0093 | 0.0203 | -0.0109 |
| x019053 | 0.0093 | 0.0135 | -0.0042 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010010 | 0.0935 | 0.0338 | +0.0597 |
| x018927 | 0.0561 | 0.0000 | +0.0561 |
| x010011 | 0.0654 | 0.0203 | +0.0452 |
| x009250 | 0.0000 | 0.0405 | -0.0405 |
| x008706 | 0.0000 | 0.0338 | -0.0338 |
| x009251 | 0.0000 | 0.0338 | -0.0338 |
| x018979 | 0.0467 | 0.0203 | +0.0265 |
| x018976 | 0.0561 | 0.0338 | +0.0223 |
| x008705 | 0.0280 | 0.0068 | +0.0213 |
| x018924 | 0.0280 | 0.0068 | +0.0213 |
| x009249 | 0.0000 | 0.0203 | -0.0203 |
| x008845 | 0.0187 | 0.0000 | +0.0187 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000055 | 0.0481 | 0.0345 | +0.0136 |
| x000054 | 0.0577 | 0.0207 | +0.0370 |
| x000053 | 0.0192 | 0.0207 | -0.0015 |
| x012464 | 0.0192 | 0.0138 | +0.0054 |
| x018925 | 0.0385 | 0.0138 | +0.0247 |
| x018931 | 0.0288 | 0.0138 | +0.0151 |
| x000052 | 0.0096 | 0.0138 | -0.0042 |
| x012467 | 0.0096 | 0.0345 | -0.0249 |
| x018929 | 0.0096 | 0.0207 | -0.0111 |
| x000698 | 0.0096 | 0.0069 | +0.0027 |
| x001169 | 0.0096 | 0.0069 | +0.0027 |
| x001171 | 0.0096 | 0.0069 | +0.0027 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000054 | 0.0577 | 0.0207 | +0.0370 |
| x012467 | 0.0096 | 0.0345 | -0.0249 |
| x018925 | 0.0385 | 0.0138 | +0.0247 |
| x012401 | 0.0288 | 0.0069 | +0.0219 |
| x018930 | 0.0288 | 0.0069 | +0.0219 |
| x018928 | 0.0000 | 0.0207 | -0.0207 |
| x019270 | 0.0000 | 0.0207 | -0.0207 |
| x000545 | 0.0192 | 0.0000 | +0.0192 |
| x001055 | 0.0192 | 0.0000 | +0.0192 |
| x012121 | 0.0192 | 0.0000 | +0.0192 |
| x012699 | 0.0192 | 0.0000 | +0.0192 |
| x012776 | 0.0192 | 0.0000 | +0.0192 |

Context diagnostics: predecessor Jaccard 0.1417, JS 0.4415, entropy A/B 3.684/4.210, effective vocabulary A/B 39.80/67.34; successor Jaccard 0.1017, JS 0.2769, entropy A/B 4.209/4.656, effective vocabulary A/B 67.26/105.27.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left ``; right `x012464`.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018929`, `x018928`, `x018931`, `x018930`, `x019270`, `x000054`, `x019269`, `x000055`, `x012698`, `x012697`, `x012699`, `x019268`; right `x000714`, `x009249`, `x000712`, `x009251`, `x019269`, `x012698`, `x009248`, `x012697`, `x000713`, `x019268`, `x009439`, `x008707`.

## `x009970` / `x020558`

Structural similarity: 0.6512; reliability: 0.8793; normalized graphemic distance: 0.7143; counts: 119/149.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9105 | 0.9881 |
| Left context | 0.6507 | 0.8153 |
| Right context | 0.3923 | 0.8347 |

- Primary component: positional agreement (0.911).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.651.
- Largest right-context difference: x018931 is more frequent for x009970 (absolute probability difference 0.056).

Position summaries (A/B): line-start 0.0672/0.0336, line-end 0.0672/0.0805, mean 5.462/6.349, median 5.000/6.000. Position JS similarity: 0.9105.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010010 | 0.0811 | 0.0625 | +0.0186 |
| x010009 | 0.0721 | 0.0556 | +0.0165 |
| x010008 | 0.0541 | 0.0417 | +0.0124 |
| x010011 | 0.0360 | 0.0625 | -0.0265 |
| x018976 | 0.0450 | 0.0278 | +0.0173 |
| x018977 | 0.0270 | 0.0556 | -0.0285 |
| x018979 | 0.0270 | 0.0278 | -0.0008 |
| x018978 | 0.0360 | 0.0139 | +0.0221 |
| x008706 | 0.0090 | 0.0347 | -0.0257 |
| x008707 | 0.0090 | 0.0139 | -0.0049 |
| x018925 | 0.0090 | 0.0139 | -0.0049 |
| x002536 | 0.0090 | 0.0069 | +0.0021 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x008705 | 0.0000 | 0.0417 | -0.0417 |
| x019053 | 0.0360 | 0.0000 | +0.0360 |
| x019055 | 0.0360 | 0.0000 | +0.0360 |
| x018977 | 0.0270 | 0.0556 | -0.0285 |
| x000715 | 0.0000 | 0.0278 | -0.0278 |
| x008704 | 0.0000 | 0.0278 | -0.0278 |
| x009248 | 0.0000 | 0.0278 | -0.0278 |
| x009250 | 0.0000 | 0.0278 | -0.0278 |
| x018926 | 0.0270 | 0.0000 | +0.0270 |
| x020831 | 0.0270 | 0.0000 | +0.0270 |
| x010011 | 0.0360 | 0.0625 | -0.0265 |
| x008706 | 0.0090 | 0.0347 | -0.0257 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000053 | 0.0360 | 0.0292 | +0.0068 |
| x018928 | 0.0270 | 0.0219 | +0.0051 |
| x018930 | 0.0270 | 0.0219 | +0.0051 |
| x000054 | 0.0180 | 0.0511 | -0.0331 |
| x000055 | 0.0180 | 0.0219 | -0.0039 |
| x000052 | 0.0270 | 0.0146 | +0.0124 |
| x018929 | 0.0180 | 0.0146 | +0.0034 |
| x012403 | 0.0090 | 0.0365 | -0.0275 |
| x000544 | 0.0090 | 0.0073 | +0.0017 |
| x000699 | 0.0180 | 0.0073 | +0.0107 |
| x001168 | 0.0090 | 0.0073 | +0.0017 |
| x010165 | 0.0090 | 0.0073 | +0.0017 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018931 | 0.0631 | 0.0073 | +0.0558 |
| x000054 | 0.0180 | 0.0511 | -0.0331 |
| x012403 | 0.0090 | 0.0365 | -0.0275 |
| x012400 | 0.0270 | 0.0000 | +0.0270 |
| x012467 | 0.0270 | 0.0000 | +0.0270 |
| x019268 | 0.0000 | 0.0219 | -0.0219 |
| x012464 | 0.0180 | 0.0000 | +0.0180 |
| x012494 | 0.0180 | 0.0000 | +0.0180 |
| x006450 | 0.0000 | 0.0146 | -0.0146 |
| x008129 | 0.0000 | 0.0146 | -0.0146 |
| x009437 | 0.0000 | 0.0146 | -0.0146 |
| x010009 | 0.0000 | 0.0146 | -0.0146 |

Context diagnostics: predecessor Jaccard 0.1371, JS 0.4347, entropy A/B 3.881/3.953, effective vocabulary A/B 48.48/52.10; successor Jaccard 0.1243, JS 0.3039, entropy A/B 4.288/4.524, effective vocabulary A/B 72.86/92.20.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left ``; right `x018929`.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018929`, `x018928`, `x018931`, `x018930`, `x019270`, `x000714`, `x000054`, `x019269`, `x000055`, `x012698`, `x012697`, `x012699`; right `x000714`, `x000712`, `x009250`, `x000715`, `x009251`, `x019269`, `x009248`, `x012699`, `x000713`, `x019271`, `x008707`, `x009436`.

## Negative controls

Controls match unordered log-counts, normalized graphemic distance, and reliability, while favoring structural similarity near the full-corpus median. They are decomposed with exactly the target metrics.

| Target | Control | Structural | Reliability | Distance | Match cost |
|---|---|---:|---:|---:|---:|
| x009968/x020559 | x008945/x009971 | 0.3020 | 0.8695 | 0.4286 | 1.3458 |
| x009968/x020559 | x008688/x012121 | 0.2998 | 0.8237 | 0.7143 | 1.3481 |
| x009968/x020559 | x008706/x015853 | 0.2901 | 0.8096 | 0.7143 | 1.3696 |
| x009971/x020558 | x008688/x012121 | 0.2998 | 0.8237 | 0.7143 | 1.0854 |
| x009971/x020558 | x008706/x015853 | 0.2901 | 0.8096 | 0.7143 | 1.1070 |
| x009971/x020558 | x015853/x021400 | 0.2800 | 0.8000 | 0.7143 | 1.1465 |
| x009970/x020559 | x008945/x009971 | 0.3020 | 0.8695 | 0.4286 | 1.2146 |
| x009970/x020559 | x008688/x012121 | 0.2998 | 0.8237 | 0.7143 | 1.2169 |
| x009970/x020559 | x008706/x015853 | 0.2901 | 0.8096 | 0.7143 | 1.2384 |
| x009971/x020556 | x008688/x012121 | 0.2998 | 0.8237 | 0.7143 | 0.9874 |
| x009971/x020556 | x015853/x021400 | 0.2800 | 0.8000 | 0.7143 | 1.0485 |
| x009971/x020556 | x015855/x021400 | 0.2887 | 0.8050 | 0.7143 | 1.0769 |
| x009969/x020558 | x008688/x012121 | 0.2998 | 0.8237 | 0.7143 | 0.9608 |
| x009969/x020558 | x008706/x015853 | 0.2901 | 0.8096 | 0.7143 | 0.9823 |
| x009969/x020558 | x015853/x021400 | 0.2800 | 0.8000 | 0.7143 | 1.0219 |
| x009970/x020556 | x008688/x012121 | 0.2998 | 0.8237 | 0.7143 | 1.0801 |
| x009970/x020556 | x015853/x021400 | 0.2800 | 0.8000 | 0.7143 | 1.1412 |
| x009970/x020556 | x015855/x021400 | 0.2887 | 0.8050 | 0.7143 | 1.1696 |
| x009971/x020559 | x008688/x012121 | 0.2998 | 0.8237 | 0.7143 | 1.1239 |
| x009971/x020559 | x008706/x015853 | 0.2901 | 0.8096 | 0.7143 | 1.1455 |
| x009971/x020559 | x008707/x015853 | 0.2848 | 0.8157 | 0.7143 | 1.1622 |
| x009970/x020558 | x008945/x009971 | 0.3020 | 0.8695 | 0.4286 | 1.1760 |
| x009970/x020558 | x008688/x012121 | 0.2998 | 0.8237 | 0.7143 | 1.1783 |
| x009970/x020558 | x008706/x015853 | 0.2901 | 0.8096 | 0.7143 | 1.1998 |

## Family decomposition

A family is a connected component; only listed edges define direct structural-distant links. Complete matrices, including non-edge pairs, are in `family_decomposition.yaml`.

### Family 1

Tokens: `x009968`, `x009969`, `x009970`, `x009971`, `x020556`, `x020558`, `x020559`. Structural medoid: `x009968`. Peripheral token(s): `x009969`.

Edges:

- `x009968` / `x020559`: similarity 0.6854, reliability 0.8927, distance 0.7143
- `x009969` / `x020558`: similarity 0.6810, reliability 0.8618, distance 0.7143
- `x009970` / `x020556`: similarity 0.6654, reliability 0.8719, distance 0.7143
- `x009970` / `x020558`: similarity 0.6512, reliability 0.8793, distance 0.7143
- `x009970` / `x020559`: similarity 0.6709, reliability 0.8822, distance 0.7143
- `x009971` / `x020556`: similarity 0.6846, reliability 0.8646, distance 0.7143
- `x009971` / `x020558`: similarity 0.6997, reliability 0.8719, distance 0.7143
- `x009971` / `x020559`: similarity 0.6555, reliability 0.8747, distance 0.7143

## Limits

Observed absence is not proof of a prohibition. Context observations at line boundaries have no neighbor and therefore context totals can be below token counts. Pair rows are statistically dependent because tokens recur across pairs. Control matching is descriptive and does not make pairs independent.
