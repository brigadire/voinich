# Structural pair decomposition

Structural similarity is reproduced unchanged from the existing pair dataset. All statements below are formal corpus descriptions; no token meaning is inferred. Context similarities and differences use full distributions, while tables are display-limited. Entropy uses natural logarithms and effective vocabulary is `exp(entropy)`.

## `x009968` / `x020557`

Structural similarity: 0.7192; reliability: 0.9136; normalized graphemic distance: 0.7143; counts: 188/174.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9736 | 0.9926 |
| Left context | 0.6777 | 0.8667 |
| Right context | 0.5063 | 0.8816 |

- Primary component: positional agreement (0.974).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.678.
- Largest left-context difference: x010008 is more frequent for x009968 (absolute probability difference 0.092).

Position summaries (A/B): line-start 0.0426/0.0517, line-end 0.0638/0.0632, mean 5.793/6.149, median 6.000/5.000. Position JS similarity: 0.9736.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010009 | 0.0944 | 0.1030 | -0.0086 |
| x010010 | 0.0333 | 0.0606 | -0.0273 |
| x010008 | 0.1222 | 0.0303 | +0.0919 |
| x018977 | 0.0278 | 0.0485 | -0.0207 |
| x008704 | 0.0222 | 0.0667 | -0.0444 |
| x008705 | 0.0222 | 0.0364 | -0.0141 |
| x018976 | 0.0222 | 0.0424 | -0.0202 |
| x018979 | 0.0222 | 0.0182 | +0.0040 |
| x020829 | 0.0222 | 0.0182 | +0.0040 |
| x018925 | 0.0167 | 0.0182 | -0.0015 |
| x010011 | 0.0167 | 0.0121 | +0.0045 |
| x018978 | 0.0222 | 0.0121 | +0.0101 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.1222 | 0.0303 | +0.0919 |
| x008704 | 0.0222 | 0.0667 | -0.0444 |
| x009248 | 0.0000 | 0.0424 | -0.0424 |
| x010010 | 0.0333 | 0.0606 | -0.0273 |
| x009249 | 0.0056 | 0.0303 | -0.0247 |
| x020766 | 0.0222 | 0.0000 | +0.0222 |
| x018977 | 0.0278 | 0.0485 | -0.0207 |
| x018976 | 0.0222 | 0.0424 | -0.0202 |
| x008845 | 0.0167 | 0.0000 | +0.0167 |
| x019054 | 0.0222 | 0.0061 | +0.0162 |
| x020828 | 0.0222 | 0.0061 | +0.0162 |
| x008705 | 0.0222 | 0.0364 | -0.0141 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000052 | 0.0341 | 0.0368 | -0.0027 |
| x018929 | 0.0284 | 0.0368 | -0.0084 |
| x018928 | 0.0398 | 0.0245 | +0.0152 |
| x000054 | 0.0341 | 0.0184 | +0.0157 |
| x000053 | 0.0170 | 0.0491 | -0.0320 |
| x012400 | 0.0114 | 0.0123 | -0.0009 |
| x017336 | 0.0114 | 0.0123 | -0.0009 |
| x018926 | 0.0114 | 0.0123 | -0.0009 |
| x019268 | 0.0114 | 0.0245 | -0.0132 |
| x012697 | 0.0114 | 0.0061 | +0.0052 |
| x018930 | 0.0170 | 0.0061 | +0.0109 |
| x019270 | 0.0114 | 0.0061 | +0.0052 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000053 | 0.0170 | 0.0491 | -0.0320 |
| x009437 | 0.0000 | 0.0245 | -0.0245 |
| x001052 | 0.0000 | 0.0184 | -0.0184 |
| x010010 | 0.0170 | 0.0000 | +0.0170 |
| x012778 | 0.0170 | 0.0000 | +0.0170 |
| x018931 | 0.0170 | 0.0000 | +0.0170 |
| x000054 | 0.0341 | 0.0184 | +0.0157 |
| x018928 | 0.0398 | 0.0245 | +0.0152 |
| x019268 | 0.0114 | 0.0245 | -0.0132 |
| x012403 | 0.0057 | 0.0184 | -0.0127 |
| x000055 | 0.0000 | 0.0123 | -0.0123 |
| x006449 | 0.0000 | 0.0123 | -0.0123 |

Context diagnostics: predecessor Jaccard 0.1656, JS 0.5161, entropy A/B 4.031/3.967, effective vocabulary A/B 56.32/52.84; successor Jaccard 0.1053, JS 0.3088, entropy A/B 4.729/4.595, effective vocabulary A/B 113.20/98.99.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `x010011`, `x018925`; right `x012400`, `x017336`, `x018926`.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018928`, `x018929`, `x018930`, `x019268`, `x012696`, `x000052`, `x019269`, `x012697`, `x000053`, `x009436`, `x000054`, `x012698`; right `x009248`, `x000712`, `x000713`, `x000714`, `x020556`, `x021400`, `x009968`, `x008944`, `x020557`, `x008945`, `x021052`, `x021401`.

## `x009969` / `x020557`

Structural similarity: 0.7146; reliability: 0.8992; normalized graphemic distance: 0.7143; counts: 135/174.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9555 | 0.9901 |
| Left context | 0.6504 | 0.8450 |
| Right context | 0.5380 | 0.8625 |

- Primary component: positional agreement (0.956).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.650.
- Largest left-context difference: x010008 is more frequent for x009969 (absolute probability difference 0.072).

Position summaries (A/B): line-start 0.0593/0.0517, line-end 0.0815/0.0632, mean 5.489/6.149, median 5.000/5.000. Position JS similarity: 0.9555.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010009 | 0.0866 | 0.1030 | -0.0164 |
| x010010 | 0.0394 | 0.0606 | -0.0212 |
| x018977 | 0.0394 | 0.0485 | -0.0091 |
| x018976 | 0.0315 | 0.0424 | -0.0109 |
| x010008 | 0.1024 | 0.0303 | +0.0721 |
| x018925 | 0.0315 | 0.0182 | +0.0133 |
| x010011 | 0.0157 | 0.0121 | +0.0036 |
| x018978 | 0.0394 | 0.0121 | +0.0272 |
| x000712 | 0.0079 | 0.0121 | -0.0042 |
| x008704 | 0.0079 | 0.0667 | -0.0588 |
| x008705 | 0.0079 | 0.0364 | -0.0285 |
| x008996 | 0.0079 | 0.0121 | -0.0042 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.1024 | 0.0303 | +0.0721 |
| x008704 | 0.0079 | 0.0667 | -0.0588 |
| x009248 | 0.0000 | 0.0424 | -0.0424 |
| x009249 | 0.0000 | 0.0303 | -0.0303 |
| x008705 | 0.0079 | 0.0364 | -0.0285 |
| x018978 | 0.0394 | 0.0121 | +0.0272 |
| x019052 | 0.0236 | 0.0000 | +0.0236 |
| x010010 | 0.0394 | 0.0606 | -0.0212 |
| x020829 | 0.0000 | 0.0182 | -0.0182 |
| x010009 | 0.0866 | 0.1030 | -0.0164 |
| x016548 | 0.0157 | 0.0000 | +0.0157 |
| x019053 | 0.0157 | 0.0000 | +0.0157 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000053 | 0.0565 | 0.0491 | +0.0074 |
| x000052 | 0.0403 | 0.0368 | +0.0035 |
| x018929 | 0.0323 | 0.0368 | -0.0046 |
| x018928 | 0.0242 | 0.0245 | -0.0003 |
| x012400 | 0.0323 | 0.0123 | +0.0200 |
| x000054 | 0.0081 | 0.0184 | -0.0103 |
| x000055 | 0.0081 | 0.0123 | -0.0042 |
| x006449 | 0.0081 | 0.0123 | -0.0042 |
| x011932 | 0.0081 | 0.0123 | -0.0042 |
| x018926 | 0.0081 | 0.0123 | -0.0042 |
| x000544 | 0.0081 | 0.0061 | +0.0019 |
| x000699 | 0.0161 | 0.0061 | +0.0100 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018930 | 0.0323 | 0.0061 | +0.0261 |
| x009437 | 0.0000 | 0.0245 | -0.0245 |
| x019268 | 0.0000 | 0.0245 | -0.0245 |
| x012464 | 0.0242 | 0.0000 | +0.0242 |
| x012400 | 0.0323 | 0.0123 | +0.0200 |
| x001052 | 0.0000 | 0.0184 | -0.0184 |
| x012403 | 0.0000 | 0.0184 | -0.0184 |
| x020720 | 0.0242 | 0.0061 | +0.0181 |
| x012401 | 0.0161 | 0.0000 | +0.0161 |
| x012493 | 0.0161 | 0.0000 | +0.0161 |
| x018931 | 0.0161 | 0.0000 | +0.0161 |
| x020316 | 0.0161 | 0.0000 | +0.0161 |

Context diagnostics: predecessor Jaccard 0.1389, JS 0.4486, entropy A/B 4.005/3.967, effective vocabulary A/B 54.88/52.84; successor Jaccard 0.1223, JS 0.3275, entropy A/B 4.353/4.595, effective vocabulary A/B 77.74/98.99.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `x010011`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018928`, `x018929`, `x018930`, `x019268`, `x012696`, `x000052`, `x000713`, `x019269`, `x012697`, `x000053`, `x009436`, `x000054`; right `x009248`, `x000712`, `x000713`, `x009436`, `x000714`, `x009250`, `x012698`, `x020556`, `x021400`, `x009968`, `x008944`, `x020557`.

## `x009969` / `x020558`

Structural similarity: 0.6983; reliability: 0.8765; normalized graphemic distance: 0.7143; counts: 135/122.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9667 | 0.9862 |
| Left context | 0.6801 | 0.8108 |
| Right context | 0.4481 | 0.8324 |

- Primary component: positional agreement (0.967).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.680.
- Largest left-context difference: x010009 is more frequent for x009969 (absolute probability difference 0.070).

Position summaries (A/B): line-start 0.0593/0.0328, line-end 0.0815/0.0492, mean 5.489/5.828, median 5.000/5.000. Position JS similarity: 0.9667.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.1024 | 0.1356 | -0.0332 |
| x018976 | 0.0315 | 0.0678 | -0.0363 |
| x010010 | 0.0394 | 0.0254 | +0.0139 |
| x018977 | 0.0394 | 0.0254 | +0.0139 |
| x010009 | 0.0866 | 0.0169 | +0.0697 |
| x018925 | 0.0315 | 0.0169 | +0.0145 |
| x018978 | 0.0394 | 0.0169 | +0.0224 |
| x016548 | 0.0157 | 0.0169 | -0.0012 |
| x019053 | 0.0157 | 0.0169 | -0.0012 |
| x020764 | 0.0157 | 0.0085 | +0.0073 |
| x020830 | 0.0157 | 0.0085 | +0.0073 |
| x000712 | 0.0079 | 0.0169 | -0.0091 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010009 | 0.0866 | 0.0169 | +0.0697 |
| x009249 | 0.0000 | 0.0424 | -0.0424 |
| x018976 | 0.0315 | 0.0678 | -0.0363 |
| x010008 | 0.1024 | 0.1356 | -0.0332 |
| x008705 | 0.0079 | 0.0339 | -0.0260 |
| x009250 | 0.0000 | 0.0254 | -0.0254 |
| x009251 | 0.0000 | 0.0254 | -0.0254 |
| x019052 | 0.0236 | 0.0000 | +0.0236 |
| x018978 | 0.0394 | 0.0169 | +0.0224 |
| x008704 | 0.0079 | 0.0254 | -0.0175 |
| x018924 | 0.0079 | 0.0254 | -0.0175 |
| x008707 | 0.0000 | 0.0169 | -0.0169 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000053 | 0.0565 | 0.0517 | +0.0047 |
| x018928 | 0.0242 | 0.0345 | -0.0103 |
| x000052 | 0.0403 | 0.0172 | +0.0231 |
| x018931 | 0.0161 | 0.0172 | -0.0011 |
| x012400 | 0.0323 | 0.0086 | +0.0236 |
| x012401 | 0.0161 | 0.0086 | +0.0075 |
| x018929 | 0.0323 | 0.0086 | +0.0236 |
| x000054 | 0.0081 | 0.0259 | -0.0178 |
| x001054 | 0.0081 | 0.0172 | -0.0092 |
| x001168 | 0.0081 | 0.0086 | -0.0006 |
| x006260 | 0.0081 | 0.0086 | -0.0006 |
| x008945 | 0.0081 | 0.0086 | -0.0006 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018930 | 0.0323 | 0.0000 | +0.0323 |
| x012466 | 0.0081 | 0.0345 | -0.0264 |
| x019269 | 0.0000 | 0.0259 | -0.0259 |
| x012464 | 0.0242 | 0.0000 | +0.0242 |
| x020720 | 0.0242 | 0.0000 | +0.0242 |
| x012400 | 0.0323 | 0.0086 | +0.0236 |
| x018929 | 0.0323 | 0.0086 | +0.0236 |
| x000052 | 0.0403 | 0.0172 | +0.0231 |
| x000054 | 0.0081 | 0.0259 | -0.0178 |
| x000697 | 0.0000 | 0.0172 | -0.0172 |
| x003286 | 0.0000 | 0.0172 | -0.0172 |
| x008484 | 0.0000 | 0.0172 | -0.0172 |

Context diagnostics: predecessor Jaccard 0.1496, JS 0.4484, entropy A/B 4.005/3.854, effective vocabulary A/B 54.88/47.16; successor Jaccard 0.1429, JS 0.3178, entropy A/B 4.353/4.401, effective vocabulary A/B 77.74/81.53.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `x016548`, `x019053`; right `x018931`.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018928`, `x018929`, `x018930`, `x019268`, `x012696`, `x000052`, `x000713`, `x019269`, `x012697`, `x000053`, `x009436`, `x000054`; right `x000712`, `x019268`, `x000713`, `x009249`, `x009436`, `x008704`, `x000714`, `x009250`, `x012698`, `x020556`, `x021400`, `x009968`.

## `x009436` / `x012697`

Structural similarity: 0.6602; reliability: 0.9136; normalized graphemic distance: 0.7143; counts: 279/351.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9886 | 0.9926 |
| Left context | 0.0848 | 0.8667 |
| Right context | 0.9071 | 0.8816 |

- Primary component: positional agreement (0.989).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.907.
- Largest right-context difference: x018928 is more frequent for x009436 (absolute probability difference 0.049).

Position summaries (A/B): line-start 0.0681/0.0513, line-end 0.0860/0.0912, mean 6.154/6.635, median 6.000/7.000. Position JS similarity: 0.9886.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x009968 | 0.0038 | 0.0060 | -0.0022 |
| x011477 | 0.0038 | 0.0060 | -0.0022 |
| x012773 | 0.0038 | 0.0150 | -0.0112 |
| x000712 | 0.0115 | 0.0030 | +0.0085 |
| x000826 | 0.0077 | 0.0030 | +0.0047 |
| x002616 | 0.0038 | 0.0030 | +0.0008 |
| x002779 | 0.0038 | 0.0030 | +0.0008 |
| x003641 | 0.0038 | 0.0030 | +0.0008 |
| x008916 | 0.0154 | 0.0030 | +0.0124 |
| x010760 | 0.0077 | 0.0030 | +0.0047 |
| x011261 | 0.0038 | 0.0030 | +0.0008 |
| x011263 | 0.0038 | 0.0030 | +0.0008 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x020556 | 0.0192 | 0.0000 | +0.0192 |
| x008916 | 0.0154 | 0.0030 | +0.0124 |
| x010010 | 0.0115 | 0.0000 | +0.0115 |
| x018948 | 0.0115 | 0.0000 | +0.0115 |
| x020732 | 0.0115 | 0.0000 | +0.0115 |
| x012773 | 0.0038 | 0.0150 | -0.0112 |
| x001490 | 0.0000 | 0.0090 | -0.0090 |
| x013734 | 0.0000 | 0.0090 | -0.0090 |
| x016802 | 0.0000 | 0.0090 | -0.0090 |
| x016826 | 0.0000 | 0.0090 | -0.0090 |
| x000712 | 0.0115 | 0.0030 | +0.0085 |
| x000082 | 0.0077 | 0.0000 | +0.0077 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018928 | 0.1373 | 0.0878 | +0.0495 |
| x018929 | 0.0941 | 0.0846 | +0.0095 |
| x018930 | 0.0588 | 0.0439 | +0.0149 |
| x000052 | 0.0392 | 0.0251 | +0.0141 |
| x018931 | 0.0471 | 0.0251 | +0.0220 |
| x008944 | 0.0353 | 0.0188 | +0.0165 |
| x008945 | 0.0196 | 0.0188 | +0.0008 |
| x012121 | 0.0196 | 0.0188 | +0.0008 |
| x000054 | 0.0157 | 0.0125 | +0.0031 |
| x000053 | 0.0118 | 0.0125 | -0.0008 |
| x000697 | 0.0275 | 0.0094 | +0.0180 |
| x010008 | 0.0078 | 0.0157 | -0.0078 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018928 | 0.1373 | 0.0878 | +0.0495 |
| x018931 | 0.0471 | 0.0251 | +0.0220 |
| x000697 | 0.0275 | 0.0094 | +0.0180 |
| x008944 | 0.0353 | 0.0188 | +0.0165 |
| x018930 | 0.0588 | 0.0439 | +0.0149 |
| x000052 | 0.0392 | 0.0251 | +0.0141 |
| x008946 | 0.0157 | 0.0031 | +0.0126 |
| x010010 | 0.0000 | 0.0125 | -0.0125 |
| x019054 | 0.0118 | 0.0000 | +0.0118 |
| x018929 | 0.0941 | 0.0846 | +0.0095 |
| x019053 | 0.0157 | 0.0063 | +0.0094 |
| x004205 | 0.0000 | 0.0094 | -0.0094 |

Context diagnostics: predecessor Jaccard 0.0329, JS 0.0731, entropy A/B 5.367/5.682, effective vocabulary A/B 214.23/293.52; successor Jaccard 0.1208, JS 0.5319, entropy A/B 4.001/4.674, effective vocabulary A/B 54.63/107.12.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left ``; right `x000696`, `x012880`, `x018924`, `x000053`, `x012120`, `x010008`, `x000054`, `x019053`, `x008945`, `x012121`.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018928`, `x018929`, `x009248`, `x018930`, `x019268`, `x012696`, `x000052`, `x009249`, `x000713`, `x019269`, `x012697`, `x000053`; right `x009248`, `x019268`, `x012696`, `x000713`, `x009249`, `x012697`, `x009436`, `x008704`, `x000714`, `x009250`, `x012698`, `x020556`.

## `x009970` / `x020556`

Structural similarity: 0.6880; reliability: 0.8682; normalized graphemic distance: 0.7143; counts: 95/213.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9470 | 0.9850 |
| Left context | 0.7191 | 0.7980 |
| Right context | 0.3979 | 0.8215 |

- Primary component: positional agreement (0.947).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.719.
- Largest right-context difference: x018925 is more frequent for x009970 (absolute probability difference 0.051).

Position summaries (A/B): line-start 0.0421/0.0657, line-end 0.0526/0.0516, mean 6.179/5.671, median 6.000/5.000. Position JS similarity: 0.9470.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.0879 | 0.0704 | +0.0176 |
| x010009 | 0.0769 | 0.0704 | +0.0066 |
| x018976 | 0.0769 | 0.0452 | +0.0317 |
| x008704 | 0.0440 | 0.0452 | -0.0013 |
| x010010 | 0.0549 | 0.0402 | +0.0147 |
| x008707 | 0.0220 | 0.0201 | +0.0019 |
| x018925 | 0.0440 | 0.0151 | +0.0289 |
| x018979 | 0.0220 | 0.0151 | +0.0069 |
| x008705 | 0.0110 | 0.0251 | -0.0141 |
| x010011 | 0.0110 | 0.0201 | -0.0091 |
| x018977 | 0.0110 | 0.0251 | -0.0141 |
| x018978 | 0.0110 | 0.0151 | -0.0041 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018924 | 0.0549 | 0.0050 | +0.0499 |
| x009248 | 0.0000 | 0.0452 | -0.0452 |
| x019052 | 0.0330 | 0.0000 | +0.0330 |
| x018976 | 0.0769 | 0.0452 | +0.0317 |
| x008706 | 0.0000 | 0.0302 | -0.0302 |
| x018925 | 0.0440 | 0.0151 | +0.0289 |
| x009249 | 0.0000 | 0.0251 | -0.0251 |
| x020828 | 0.0000 | 0.0251 | -0.0251 |
| x020765 | 0.0220 | 0.0000 | +0.0220 |
| x016548 | 0.0000 | 0.0201 | -0.0201 |
| x010008 | 0.0879 | 0.0704 | +0.0176 |
| x009250 | 0.0000 | 0.0151 | -0.0151 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000053 | 0.0444 | 0.0347 | +0.0098 |
| x000054 | 0.0333 | 0.0248 | +0.0086 |
| x000052 | 0.0444 | 0.0198 | +0.0246 |
| x000055 | 0.0111 | 0.0198 | -0.0087 |
| x010008 | 0.0111 | 0.0149 | -0.0037 |
| x018928 | 0.0111 | 0.0248 | -0.0136 |
| x000544 | 0.0222 | 0.0099 | +0.0123 |
| x000633 | 0.0111 | 0.0099 | +0.0012 |
| x012400 | 0.0111 | 0.0099 | +0.0012 |
| x017916 | 0.0111 | 0.0099 | +0.0012 |
| x018929 | 0.0333 | 0.0099 | +0.0234 |
| x001785 | 0.0111 | 0.0050 | +0.0062 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018925 | 0.0556 | 0.0050 | +0.0506 |
| x018931 | 0.0333 | 0.0000 | +0.0333 |
| x009436 | 0.0000 | 0.0248 | -0.0248 |
| x000052 | 0.0444 | 0.0198 | +0.0246 |
| x018929 | 0.0333 | 0.0099 | +0.0234 |
| x012121 | 0.0222 | 0.0000 | +0.0222 |
| x012467 | 0.0222 | 0.0000 | +0.0222 |
| x009437 | 0.0000 | 0.0149 | -0.0149 |
| x009438 | 0.0000 | 0.0149 | -0.0149 |
| x012464 | 0.0000 | 0.0149 | -0.0149 |
| x018928 | 0.0111 | 0.0248 | -0.0136 |
| x000544 | 0.0222 | 0.0099 | +0.0123 |

Context diagnostics: predecessor Jaccard 0.1250, JS 0.4652, entropy A/B 3.648/4.147, effective vocabulary A/B 38.40/63.23; successor Jaccard 0.0870, JS 0.2575, entropy A/B 4.131/4.891, effective vocabulary A/B 62.25/133.13.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018928`, `x018929`, `x018930`, `x019268`, `x012696`, `x000052`, `x000713`, `x019269`, `x012697`, `x000053`, `x009436`, `x000054`; right `x009248`, `x000712`, `x000713`, `x012697`, `x000714`, `x009250`, `x012698`, `x020556`, `x021400`, `x010009`, `x009968`, `x008944`.

## `x009970` / `x020558`

Structural similarity: 0.6553; reliability: 0.8466; normalized graphemic distance: 0.7143; counts: 95/122.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9095 | 0.9811 |
| Left context | 0.6852 | 0.7658 |
| Right context | 0.3713 | 0.7928 |

- Primary component: positional agreement (0.910).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.685.
- Largest left-context difference: x010009 is more frequent for x009970 (absolute probability difference 0.060).

Position summaries (A/B): line-start 0.0421/0.0328, line-end 0.0526/0.0492, mean 6.179/5.828, median 6.000/5.000. Position JS similarity: 0.9095.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.0879 | 0.1356 | -0.0477 |
| x018976 | 0.0769 | 0.0678 | +0.0091 |
| x008704 | 0.0440 | 0.0254 | +0.0185 |
| x010010 | 0.0549 | 0.0254 | +0.0295 |
| x018924 | 0.0549 | 0.0254 | +0.0295 |
| x008707 | 0.0220 | 0.0169 | +0.0050 |
| x010009 | 0.0769 | 0.0169 | +0.0600 |
| x018925 | 0.0440 | 0.0169 | +0.0270 |
| x018979 | 0.0220 | 0.0169 | +0.0050 |
| x008705 | 0.0110 | 0.0339 | -0.0229 |
| x018977 | 0.0110 | 0.0254 | -0.0144 |
| x018978 | 0.0110 | 0.0169 | -0.0060 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010009 | 0.0769 | 0.0169 | +0.0600 |
| x010008 | 0.0879 | 0.1356 | -0.0477 |
| x009249 | 0.0000 | 0.0424 | -0.0424 |
| x019052 | 0.0330 | 0.0000 | +0.0330 |
| x010010 | 0.0549 | 0.0254 | +0.0295 |
| x018924 | 0.0549 | 0.0254 | +0.0295 |
| x018925 | 0.0440 | 0.0169 | +0.0270 |
| x009250 | 0.0000 | 0.0254 | -0.0254 |
| x009251 | 0.0000 | 0.0254 | -0.0254 |
| x008705 | 0.0110 | 0.0339 | -0.0229 |
| x020765 | 0.0220 | 0.0000 | +0.0220 |
| x008704 | 0.0440 | 0.0254 | +0.0185 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000053 | 0.0444 | 0.0517 | -0.0073 |
| x000054 | 0.0333 | 0.0259 | +0.0075 |
| x000052 | 0.0444 | 0.0172 | +0.0272 |
| x018931 | 0.0333 | 0.0172 | +0.0161 |
| x014628 | 0.0111 | 0.0172 | -0.0061 |
| x018928 | 0.0111 | 0.0345 | -0.0234 |
| x010933 | 0.0111 | 0.0086 | +0.0025 |
| x011932 | 0.0111 | 0.0086 | +0.0025 |
| x012400 | 0.0111 | 0.0086 | +0.0025 |
| x012465 | 0.0111 | 0.0086 | +0.0025 |
| x018924 | 0.0111 | 0.0086 | +0.0025 |
| x018925 | 0.0556 | 0.0086 | +0.0469 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x018925 | 0.0556 | 0.0086 | +0.0469 |
| x012466 | 0.0000 | 0.0345 | -0.0345 |
| x000052 | 0.0444 | 0.0172 | +0.0272 |
| x019269 | 0.0000 | 0.0259 | -0.0259 |
| x018929 | 0.0333 | 0.0086 | +0.0247 |
| x018928 | 0.0111 | 0.0345 | -0.0234 |
| x000544 | 0.0222 | 0.0000 | +0.0222 |
| x012121 | 0.0222 | 0.0000 | +0.0222 |
| x012467 | 0.0222 | 0.0000 | +0.0222 |
| x000697 | 0.0000 | 0.0172 | -0.0172 |
| x001054 | 0.0000 | 0.0172 | -0.0172 |
| x003286 | 0.0000 | 0.0172 | -0.0172 |

Context diagnostics: predecessor Jaccard 0.1429, JS 0.4622, entropy A/B 3.648/3.854, effective vocabulary A/B 38.40/47.16; successor Jaccard 0.0867, JS 0.2413, entropy A/B 4.131/4.401, effective vocabulary A/B 62.25/81.53.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018928`, `x018929`, `x018930`, `x019268`, `x012696`, `x000052`, `x000713`, `x019269`, `x012697`, `x000053`, `x009436`, `x000054`; right `x000712`, `x018930`, `x019268`, `x000713`, `x009249`, `x012697`, `x009436`, `x000714`, `x009250`, `x012698`, `x020556`, `x021400`.

## `x009971` / `x020556`

Structural similarity: 0.6619; reliability: 0.7914; normalized graphemic distance: 0.7143; counts: 43/213.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9528 | 0.9625 |
| Left context | 0.6397 | 0.6898 |
| Right context | 0.3930 | 0.7219 |

- Primary component: positional agreement (0.953).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.640.
- Largest right-context difference: x000545 is more frequent for x009971 (absolute probability difference 0.049).

Position summaries (A/B): line-start 0.0233/0.0657, line-end 0.0465/0.0516, mean 5.605/5.671, median 5.000/5.000. Position JS similarity: 0.9528.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010008 | 0.0952 | 0.0704 | +0.0249 |
| x010009 | 0.1190 | 0.0704 | +0.0487 |
| x018976 | 0.0714 | 0.0452 | +0.0262 |
| x010010 | 0.0238 | 0.0402 | -0.0164 |
| x018977 | 0.0238 | 0.0251 | -0.0013 |
| x010011 | 0.0476 | 0.0201 | +0.0275 |
| x018925 | 0.0238 | 0.0151 | +0.0087 |
| x018978 | 0.0238 | 0.0151 | +0.0087 |
| x018979 | 0.0476 | 0.0151 | +0.0325 |
| x002556 | 0.0238 | 0.0050 | +0.0188 |
| x010217 | 0.0238 | 0.0050 | +0.0188 |
| x011261 | 0.0238 | 0.0050 | +0.0188 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x010009 | 0.1190 | 0.0704 | +0.0487 |
| x018926 | 0.0476 | 0.0000 | +0.0476 |
| x018927 | 0.0476 | 0.0000 | +0.0476 |
| x008704 | 0.0000 | 0.0452 | -0.0452 |
| x009248 | 0.0000 | 0.0452 | -0.0452 |
| x018979 | 0.0476 | 0.0151 | +0.0325 |
| x008706 | 0.0000 | 0.0302 | -0.0302 |
| x010011 | 0.0476 | 0.0201 | +0.0275 |
| x018976 | 0.0714 | 0.0452 | +0.0262 |
| x008705 | 0.0000 | 0.0251 | -0.0251 |
| x009249 | 0.0000 | 0.0251 | -0.0251 |
| x020828 | 0.0000 | 0.0251 | -0.0251 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000053 | 0.0732 | 0.0347 | +0.0385 |
| x000054 | 0.0244 | 0.0248 | -0.0004 |
| x018928 | 0.0244 | 0.0248 | -0.0004 |
| x000055 | 0.0244 | 0.0198 | +0.0046 |
| x012464 | 0.0488 | 0.0149 | +0.0339 |
| x000697 | 0.0244 | 0.0099 | +0.0145 |
| x012400 | 0.0244 | 0.0099 | +0.0145 |
| x012401 | 0.0244 | 0.0099 | +0.0145 |
| x018929 | 0.0244 | 0.0099 | +0.0145 |
| x018930 | 0.0244 | 0.0099 | +0.0145 |
| x001054 | 0.0244 | 0.0050 | +0.0194 |
| x008872 | 0.0244 | 0.0050 | +0.0194 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000545 | 0.0488 | 0.0000 | +0.0488 |
| x000053 | 0.0732 | 0.0347 | +0.0385 |
| x012464 | 0.0488 | 0.0149 | +0.0339 |
| x009436 | 0.0000 | 0.0248 | -0.0248 |
| x000714 | 0.0244 | 0.0000 | +0.0244 |
| x001168 | 0.0244 | 0.0000 | +0.0244 |
| x003445 | 0.0244 | 0.0000 | +0.0244 |
| x004756 | 0.0244 | 0.0000 | +0.0244 |
| x005933 | 0.0244 | 0.0000 | +0.0244 |
| x009125 | 0.0244 | 0.0000 | +0.0244 |
| x009250 | 0.0244 | 0.0000 | +0.0244 |
| x009712 | 0.0244 | 0.0000 | +0.0244 |

Context diagnostics: predecessor Jaccard 0.1217, JS 0.4229, entropy A/B 3.204/4.147, effective vocabulary A/B 24.62/63.23; successor Jaccard 0.0791, JS 0.2576, entropy A/B 3.566/4.891, effective vocabulary A/B 35.36/133.13.

Shared unobserved high-frequency contexts (descriptive absence only): left `x018928`, `x018929`, `x018930`, `x019268`, `x012696`, `x000052`, `x000713`, `x019269`, `x012697`, `x000053`, `x009436`, `x000054`; right `x009248`, `x000712`, `x012696`, `x000713`, `x012697`, `x008704`, `x018931`, `x020556`, `x021400`, `x010009`, `x009968`, `x008944`.

## Negative controls

Controls match unordered log-counts, normalized graphemic distance, and reliability, while favoring structural similarity near the full-corpus median. They are decomposed with exactly the target metrics.

| Target | Control | Structural | Reliability | Distance | Match cost |
|---|---|---:|---:|---:|---:|
| x009968/x020557 | x008944/x015852 | 0.2934 | 0.8643 | 0.7143 | 1.1234 |
| x009968/x020557 | x008944/x015853 | 0.2742 | 0.8216 | 0.7143 | 1.4609 |
| x009968/x020557 | x015852/x020628 | 0.2954 | 0.8461 | 0.7143 | 1.5584 |
| x009969/x020557 | x008944/x015852 | 0.2934 | 0.8643 | 0.7143 | 0.9194 |
| x009969/x020557 | x015852/x020628 | 0.2954 | 0.8461 | 0.7143 | 1.2004 |
| x009969/x020557 | x008944/x015853 | 0.2742 | 0.8216 | 0.7143 | 1.2568 |
| x009969/x020558 | x015852/x020628 | 0.2954 | 0.8461 | 0.7143 | 0.8024 |
| x009969/x020558 | x015852/x020629 | 0.2952 | 0.8380 | 0.7143 | 0.9148 |
| x009969/x020558 | x008706/x015852 | 0.2970 | 0.8387 | 0.7143 | 0.9235 |
| x009436/x012697 | x000053/x015852 | 0.3040 | 0.8643 | 0.5714 | 2.0001 |
| x009436/x012697 | x009249/x015853 | 0.2897 | 0.8216 | 0.7143 | 2.1622 |
| x009436/x012697 | x008944/x015852 | 0.2934 | 0.8643 | 0.7143 | 2.1943 |
| x009970/x020556 | x008944/x015852 | 0.2934 | 0.8643 | 0.7143 | 0.5353 |
| x009970/x020556 | x015852/x018931 | 0.2968 | 0.8643 | 0.5714 | 0.8128 |
| x009970/x020556 | x008944/x015853 | 0.2742 | 0.8216 | 0.7143 | 0.8728 |
| x009970/x020558 | x015852/x020628 | 0.2954 | 0.8461 | 0.7143 | 0.4740 |
| x009970/x020558 | x015852/x020629 | 0.2952 | 0.8380 | 0.7143 | 0.5068 |
| x009970/x020558 | x008706/x015852 | 0.2970 | 0.8387 | 0.7143 | 0.5155 |
| x009971/x020556 | x018931/x020105 | 0.2642 | 0.7864 | 0.7143 | 0.2070 |
| x009971/x020556 | x000714/x012123 | 0.2706 | 0.7812 | 0.7143 | 0.3084 |
| x009971/x020556 | x012698/x020720 | 0.2774 | 0.7838 | 0.7143 | 0.3462 |

## Family decomposition

A family is a connected component; only listed edges define direct structural-distant links. Complete matrices, including non-edge pairs, are in `family_decomposition.yaml`.

### Family 1

Tokens: `x009436`, `x012697`. Structural medoid: `x009436`. Peripheral token(s): `x009436`, `x012697`.

Edges:

- `x009436` / `x012697`: similarity 0.6602, reliability 0.9136, distance 0.7143

### Family 2

Tokens: `x009968`, `x009969`, `x009970`, `x009971`, `x020556`, `x020557`, `x020558`. Structural medoid: `x009969`. Peripheral token(s): `x009971`.

Edges:

- `x009968` / `x020557`: similarity 0.7192, reliability 0.9136, distance 0.7143
- `x009969` / `x020557`: similarity 0.7146, reliability 0.8992, distance 0.7143
- `x009969` / `x020558`: similarity 0.6983, reliability 0.8765, distance 0.7143
- `x009970` / `x020556`: similarity 0.6880, reliability 0.8682, distance 0.7143
- `x009970` / `x020558`: similarity 0.6553, reliability 0.8466, distance 0.7143
- `x009971` / `x020556`: similarity 0.6619, reliability 0.7914, distance 0.7143

## Limits

Observed absence is not proof of a prohibition. Context observations at line boundaries have no neighbor and therefore context totals can be below token counts. Pair rows are statistically dependent because tokens recur across pairs. Control matching is descriptive and does not make pairs independent.
