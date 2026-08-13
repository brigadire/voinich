# Structural pair decomposition

Structural similarity is reproduced unchanged from the existing pair dataset. All statements below are formal corpus descriptions; no token meaning is inferred. Context similarities and differences use full distributions, while tables are display-limited. Entropy uses natural logarithms and effective vocabulary is `exp(entropy)`.

## `or` / `s`

Structural similarity: 0.7594; reliability: 0.9336; normalized graphemic distance: 1.0000; counts: 388/350.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9477 | 0.9874 |
| Left context | 0.4181 | 0.9051 |
| Right context | 0.9123 | 0.9084 |

- Primary component: positional agreement (0.948).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.912.
- Largest right-context difference: ar is more frequent for s (absolute probability difference 0.041).

Position summaries (A/B): line-start 0.0825/0.0771, line-end 0.0464/0.1257, mean 5.129/5.823, median 4.000/5.000. Position JS similarity: 0.9477.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ol | 0.0197 | 0.0279 | -0.0082 |
| daiin | 0.0225 | 0.0186 | +0.0039 |
| chol | 0.0112 | 0.0124 | -0.0011 |
| cheey | 0.0112 | 0.0093 | +0.0019 |
| dal | 0.0112 | 0.0093 | +0.0019 |
| aiin | 0.0084 | 0.0093 | -0.0009 |
| chor | 0.0084 | 0.0093 | -0.0009 |
| l | 0.0084 | 0.0279 | -0.0194 |
| qokeey | 0.0084 | 0.0093 | -0.0009 |
| chear | 0.0084 | 0.0062 | +0.0022 |
| dar | 0.0084 | 0.0062 | +0.0022 |
| sheey | 0.0084 | 0.0062 | +0.0022 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0337 | 0.0031 | +0.0306 |
| ar | 0.0281 | 0.0031 | +0.0250 |
| chy | 0.0000 | 0.0217 | -0.0217 |
| l | 0.0084 | 0.0279 | -0.0194 |
| al | 0.0028 | 0.0186 | -0.0158 |
| or | 0.0169 | 0.0031 | +0.0138 |
| cho | 0.0000 | 0.0124 | -0.0124 |
| qokaiin | 0.0112 | 0.0000 | +0.0112 |
| chl | 0.0028 | 0.0124 | -0.0096 |
| cthy | 0.0028 | 0.0124 | -0.0096 |
| sh | 0.0000 | 0.0093 | -0.0093 |
| dair | 0.0084 | 0.0000 | +0.0084 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.1541 | 0.1928 | -0.0388 |
| ar | 0.0243 | 0.0654 | -0.0410 |
| ain | 0.0324 | 0.0229 | +0.0096 |
| air | 0.0189 | 0.0294 | -0.0105 |
| al | 0.0189 | 0.0261 | -0.0072 |
| chol | 0.0189 | 0.0261 | -0.0072 |
| or | 0.0162 | 0.0392 | -0.0230 |
| y | 0.0135 | 0.0327 | -0.0192 |
| cheey | 0.0189 | 0.0131 | +0.0058 |
| chey | 0.0189 | 0.0131 | +0.0058 |
| ol | 0.0270 | 0.0131 | +0.0140 |
| aiiin | 0.0243 | 0.0065 | +0.0178 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ar | 0.0243 | 0.0654 | -0.0410 |
| aiin | 0.1541 | 0.1928 | -0.0388 |
| shedy | 0.0243 | 0.0000 | +0.0243 |
| or | 0.0162 | 0.0392 | -0.0230 |
| y | 0.0135 | 0.0327 | -0.0192 |
| aiiin | 0.0243 | 0.0065 | +0.0178 |
| ol | 0.0270 | 0.0131 | +0.0140 |
| o | 0.0027 | 0.0163 | -0.0136 |
| cheol | 0.0054 | 0.0163 | -0.0109 |
| chedy | 0.0108 | 0.0000 | +0.0108 |
| okain | 0.0108 | 0.0000 | +0.0108 |
| air | 0.0189 | 0.0294 | -0.0105 |

Context diagnostics: predecessor Jaccard 0.1091, JS 0.3020, entropy A/B 5.278/5.319, effective vocabulary A/B 196.03/204.16; successor Jaccard 0.1848, JS 0.5777, entropy A/B 4.579/4.103, effective vocabulary A/B 97.43/60.52.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `chedy`, `d`, `dain`, `o`, `chear`, `dar`, `sheey`, `aiin`, `chckhy`, `chor`, `qokeey`, `y`; right `cheos`, `dain`, `oiiin`, `am`, `chy`, `chor`, `cheol`, `cheey`, `chey`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `otaiin`, `okain`, `chdy`, `okedy`, `oteedy`, `cheor`, `shy`, `qotedy`, `chody`, `otol`, `oky`; right `qokeedy`, `qokeey`, `qokedy`, `qokain`, `qokaiin`, `qokal`, `okeey`, `otedy`, `qokar`, `otar`, `qoky`, `okar`.

## `chol` / `daiin`

Structural similarity: 0.7206; reliability: 0.9336; normalized graphemic distance: 1.0000; counts: 395/847.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9070 | 0.9874 |
| Left context | 0.7033 | 0.9051 |
| Right context | 0.5515 | 0.9084 |

- Primary component: positional agreement (0.907).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.703.
- Largest right-context difference: daiin is more frequent for chol (absolute probability difference 0.069).

Position summaries (A/B): line-start 0.0481/0.1889, line-end 0.0127/0.1547, mean 3.947/4.253, median 3.000/4.000. Position JS similarity: 0.9070.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chol | 0.0638 | 0.0480 | +0.0158 |
| daiin | 0.0186 | 0.0160 | +0.0026 |
| chor | 0.0213 | 0.0131 | +0.0082 |
| shol | 0.0106 | 0.0131 | -0.0025 |
| qokol | 0.0106 | 0.0102 | +0.0004 |
| cthol | 0.0106 | 0.0087 | +0.0019 |
| chedy | 0.0080 | 0.0146 | -0.0066 |
| dy | 0.0080 | 0.0131 | -0.0051 |
| otol | 0.0186 | 0.0073 | +0.0113 |
| cheol | 0.0160 | 0.0058 | +0.0101 |
| chey | 0.0053 | 0.0116 | -0.0063 |
| chody | 0.0053 | 0.0058 | -0.0005 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0213 | 0.0015 | +0.0198 |
| chol | 0.0638 | 0.0480 | +0.0158 |
| or | 0.0186 | 0.0029 | +0.0157 |
| chy | 0.0000 | 0.0131 | -0.0131 |
| qokeey | 0.0027 | 0.0146 | -0.0119 |
| otol | 0.0186 | 0.0073 | +0.0113 |
| qoky | 0.0053 | 0.0160 | -0.0107 |
| qokaiin | 0.0106 | 0.0000 | +0.0106 |
| dor | 0.0133 | 0.0029 | +0.0104 |
| cheol | 0.0160 | 0.0058 | +0.0101 |
| choky | 0.0106 | 0.0015 | +0.0092 |
| sho | 0.0106 | 0.0015 | +0.0092 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chey | 0.0154 | 0.0182 | -0.0028 |
| daiin | 0.0846 | 0.0154 | +0.0693 |
| cthy | 0.0128 | 0.0154 | -0.0025 |
| dain | 0.0128 | 0.0126 | +0.0003 |
| or | 0.0103 | 0.0112 | -0.0009 |
| chol | 0.0615 | 0.0098 | +0.0518 |
| chor | 0.0179 | 0.0098 | +0.0082 |
| cthol | 0.0154 | 0.0098 | +0.0056 |
| ol | 0.0231 | 0.0098 | +0.0133 |
| chy | 0.0205 | 0.0084 | +0.0121 |
| dy | 0.0205 | 0.0084 | +0.0121 |
| s | 0.0103 | 0.0084 | +0.0019 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0846 | 0.0154 | +0.0693 |
| chol | 0.0615 | 0.0098 | +0.0518 |
| shey | 0.0000 | 0.0140 | -0.0140 |
| ol | 0.0231 | 0.0098 | +0.0133 |
| chy | 0.0205 | 0.0084 | +0.0121 |
| dy | 0.0205 | 0.0084 | +0.0121 |
| chckhy | 0.0026 | 0.0126 | -0.0100 |
| dol | 0.0154 | 0.0056 | +0.0098 |
| shol | 0.0179 | 0.0084 | +0.0096 |
| chcthy | 0.0000 | 0.0084 | -0.0084 |
| chor | 0.0179 | 0.0098 | +0.0082 |
| dal | 0.0077 | 0.0154 | -0.0077 |

Context diagnostics: predecessor Jaccard 0.1782, JS 0.4580, entropy A/B 5.199/5.613, effective vocabulary A/B 181.14/273.86; successor Jaccard 0.1503, JS 0.4373, entropy A/B 4.932/5.711, effective vocabulary A/B 138.66/302.18.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `cheky`, `kchor`, `keol`, `kol`, `okol`, `otchol`, `qokeol`, `sheey`, `chody`, `shey`, `cthor`, `kchol`; right `ar`, `cheky`, `dair`, `o`, `okal`, `okeol`, `oky`, `otaiin`, `qodaiin`, `cheey`, `choky`, `sho`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qol`, `okain`, `okedy`, `oteedy`, `char`, `air`, `cheody`, `qoteedy`, `sar`, `qotar`, `shckhy`, `keedy`; right `qokaiin`, `r`, `lchedy`, `ain`, `qokey`, `qotedy`, `saiin`, `qoteedy`, `qotar`, `qotal`, `opchedy`, `qokchdy`.

## `r` / `s`

Structural similarity: 0.7010; reliability: 0.9336; normalized graphemic distance: 1.0000; counts: 168/350.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9146 | 0.9874 |
| Left context | 0.3598 | 0.9051 |
| Right context | 0.8287 | 0.9084 |

- Primary component: positional agreement (0.915).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.829.
- Largest right-context difference: ain is more frequent for r (absolute probability difference 0.054).

Position summaries (A/B): line-start 0.0417/0.0771, line-end 0.0655/0.1257, mean 7.673/5.823, median 5.000/5.000. Position JS similarity: 0.9146.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| l | 0.0311 | 0.0279 | +0.0032 |
| ol | 0.0186 | 0.0279 | -0.0092 |
| al | 0.0186 | 0.0186 | +0.0001 |
| cho | 0.0248 | 0.0124 | +0.0125 |
| cheey | 0.0124 | 0.0093 | +0.0031 |
| y | 0.0124 | 0.0093 | +0.0031 |
| chy | 0.0062 | 0.0217 | -0.0155 |
| qokeey | 0.0062 | 0.0093 | -0.0031 |
| chedy | 0.0062 | 0.0062 | +0.0000 |
| d | 0.0311 | 0.0062 | +0.0249 |
| o | 0.0435 | 0.0062 | +0.0373 |
| qokchdy | 0.0062 | 0.0062 | +0.0000 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| o | 0.0435 | 0.0062 | +0.0373 |
| cheo | 0.0311 | 0.0000 | +0.0311 |
| d | 0.0311 | 0.0062 | +0.0249 |
| a | 0.0248 | 0.0000 | +0.0248 |
| t | 0.0248 | 0.0000 | +0.0248 |
| keo | 0.0186 | 0.0000 | +0.0186 |
| lo | 0.0186 | 0.0000 | +0.0186 |
| okeo | 0.0186 | 0.0000 | +0.0186 |
| daiin | 0.0000 | 0.0186 | -0.0186 |
| chey | 0.0186 | 0.0031 | +0.0155 |
| chy | 0.0062 | 0.0217 | -0.0155 |
| cho | 0.0248 | 0.0124 | +0.0125 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.1592 | 0.1928 | -0.0336 |
| ar | 0.0382 | 0.0654 | -0.0271 |
| al | 0.0318 | 0.0261 | +0.0057 |
| ain | 0.0764 | 0.0229 | +0.0536 |
| cheey | 0.0191 | 0.0131 | +0.0060 |
| chey | 0.0255 | 0.0131 | +0.0124 |
| ol | 0.0510 | 0.0131 | +0.0379 |
| chy | 0.0127 | 0.0131 | -0.0003 |
| or | 0.0127 | 0.0392 | -0.0265 |
| aiiin | 0.0127 | 0.0065 | +0.0062 |
| shey | 0.0191 | 0.0065 | +0.0126 |
| shor | 0.0127 | 0.0065 | +0.0062 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ain | 0.0764 | 0.0229 | +0.0536 |
| ol | 0.0510 | 0.0131 | +0.0379 |
| aiin | 0.1592 | 0.1928 | -0.0336 |
| y | 0.0000 | 0.0327 | -0.0327 |
| air | 0.0000 | 0.0294 | -0.0294 |
| ar | 0.0382 | 0.0654 | -0.0271 |
| or | 0.0127 | 0.0392 | -0.0265 |
| chol | 0.0000 | 0.0261 | -0.0261 |
| @170; | 0.0255 | 0.0000 | +0.0255 |
| v | 0.0255 | 0.0000 | +0.0255 |
| char | 0.0191 | 0.0000 | +0.0191 |
| a | 0.0127 | 0.0000 | +0.0127 |

Context diagnostics: predecessor Jaccard 0.0948, JS 0.2511, entropy A/B 4.543/5.319, effective vocabulary A/B 93.99/204.16; successor Jaccard 0.1414, JS 0.5036, entropy A/B 3.880/4.103, effective vocabulary A/B 48.41/60.52.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `cheey`, `y`, `al`; right `aiiin`, `shor`, `chy`, `cheey`, `shey`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `qokain`, `qokaiin`, `okeey`, `dy`, `r`, `otedy`, `otar`, `otaiin`, `okar`, `otal`, `dair`; right `qokeedy`, `qokeey`, `qokedy`, `qokain`, `qokaiin`, `qokal`, `r`, `otedy`, `qokar`, `chdy`, `okal`, `otar`.

## `ol` / `y`

Structural similarity: 0.6787; reliability: 0.9336; normalized graphemic distance: 1.0000; counts: 557/304.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.8897 | 0.9874 |
| Left context | 0.5400 | 0.9051 |
| Right context | 0.6063 | 0.9084 |

- Primary component: positional agreement (0.890).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.606.
- Largest left-context difference: s is more frequent for y (absolute probability difference 0.036).

Position summaries (A/B): line-start 0.0557/0.2434, line-end 0.0754/0.1546, mean 5.210/5.618, median 4.000/4.000. Position JS similarity: 0.8897.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0190 | 0.0261 | -0.0071 |
| or | 0.0190 | 0.0217 | -0.0027 |
| ar | 0.0171 | 0.0304 | -0.0133 |
| qokaiin | 0.0152 | 0.0174 | -0.0022 |
| daiin | 0.0133 | 0.0304 | -0.0171 |
| okar | 0.0114 | 0.0174 | -0.0060 |
| dol | 0.0095 | 0.0130 | -0.0035 |
| okain | 0.0095 | 0.0174 | -0.0079 |
| ain | 0.0114 | 0.0087 | +0.0027 |
| chol | 0.0171 | 0.0087 | +0.0084 |
| dar | 0.0152 | 0.0087 | +0.0065 |
| ol | 0.0152 | 0.0087 | +0.0065 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0076 | 0.0435 | -0.0359 |
| qokain | 0.0247 | 0.0000 | +0.0247 |
| @171; | 0.0000 | 0.0174 | -0.0174 |
| d | 0.0000 | 0.0174 | -0.0174 |
| daiin | 0.0133 | 0.0304 | -0.0171 |
| r | 0.0152 | 0.0000 | +0.0152 |
| ar | 0.0171 | 0.0304 | -0.0133 |
| qokeed | 0.0000 | 0.0130 | -0.0130 |
| qokedy | 0.0114 | 0.0000 | +0.0114 |
| shedy | 0.0114 | 0.0000 | +0.0114 |
| cheor | 0.0095 | 0.0000 | +0.0095 |
| chey | 0.0095 | 0.0000 | +0.0095 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0466 | 0.0233 | +0.0233 |
| cheey | 0.0214 | 0.0272 | -0.0059 |
| kaiin | 0.0194 | 0.0195 | -0.0000 |
| chey | 0.0175 | 0.0195 | -0.0020 |
| daiin | 0.0175 | 0.0311 | -0.0137 |
| cheol | 0.0155 | 0.0195 | -0.0039 |
| chedy | 0.0408 | 0.0117 | +0.0291 |
| s | 0.0175 | 0.0117 | +0.0058 |
| shey | 0.0136 | 0.0117 | +0.0019 |
| aiin | 0.0272 | 0.0078 | +0.0194 |
| chor | 0.0097 | 0.0078 | +0.0019 |
| chy | 0.0097 | 0.0078 | +0.0019 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0408 | 0.0117 | +0.0291 |
| shedy | 0.0466 | 0.0233 | +0.0233 |
| aiin | 0.0272 | 0.0078 | +0.0194 |
| sheey | 0.0214 | 0.0039 | +0.0175 |
| taiin | 0.0000 | 0.0156 | -0.0156 |
| daiin | 0.0175 | 0.0311 | -0.0137 |
| dy | 0.0039 | 0.0156 | -0.0117 |
| c | 0.0000 | 0.0117 | -0.0117 |
| cheeo | 0.0000 | 0.0117 | -0.0117 |
| kal | 0.0000 | 0.0117 | -0.0117 |
| ky | 0.0000 | 0.0117 | -0.0117 |
| tchy | 0.0000 | 0.0117 | -0.0117 |

Context diagnostics: predecessor Jaccard 0.1471, JS 0.4066, entropy A/B 5.389/4.857, effective vocabulary A/B 219.09/128.70; successor Jaccard 0.1605, JS 0.4433, entropy A/B 5.093/5.035, effective vocabulary A/B 162.90/153.63.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `al`, `l`, `okaiin`, `taiin`, `qokeey`, `ain`, `dal`, `dol`, `okal`, `dar`, `ol`, `chol`; right `chedar`, `chol`, `dain`, `kor`, `r`, `raiin`, `chor`, `chy`, `kedy`, `kar`, `or`, `shey`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `shol`, `chdy`, `okedy`, `qokol`, `oty`, `cheody`, `cthy`, `qoteedy`, `qokchy`, `odaiin`, `qotchy`; right `qokedy`, `qokain`, `okar`, `oty`, `cthy`, `otain`, `shy`, `am`, `qotedy`, `qoty`, `shor`, `qotaiin`.

## `dar` / `ol`

Structural similarity: 0.6742; reliability: 0.9336; normalized graphemic distance: 1.0000; counts: 323/557.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9417 | 0.9874 |
| Left context | 0.4971 | 0.9051 |
| Right context | 0.5838 | 0.9084 |

- Primary component: positional agreement (0.942).
- Similarity is concentrated: the next component, successor-distribution overlap, is 0.584.
- Largest right-context difference: shedy is more frequent for ol (absolute probability difference 0.028).

Position summaries (A/B): line-start 0.1300/0.0557, line-end 0.1455/0.0754, mean 6.211/5.210, median 5.000/4.000. Position JS similarity: 0.9417.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokain | 0.0178 | 0.0247 | -0.0069 |
| daiin | 0.0249 | 0.0133 | +0.0116 |
| dar | 0.0107 | 0.0152 | -0.0045 |
| ol | 0.0107 | 0.0152 | -0.0045 |
| qokedy | 0.0107 | 0.0114 | -0.0007 |
| chedy | 0.0107 | 0.0095 | +0.0012 |
| chey | 0.0107 | 0.0095 | +0.0012 |
| qokal | 0.0249 | 0.0095 | +0.0154 |
| al | 0.0178 | 0.0076 | +0.0102 |
| aiin | 0.0071 | 0.0190 | -0.0119 |
| chol | 0.0071 | 0.0171 | -0.0100 |
| dain | 0.0071 | 0.0076 | -0.0005 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| or | 0.0000 | 0.0190 | -0.0190 |
| qokal | 0.0249 | 0.0095 | +0.0154 |
| r | 0.0000 | 0.0152 | -0.0152 |
| olky | 0.0142 | 0.0000 | +0.0142 |
| ar | 0.0036 | 0.0171 | -0.0136 |
| dal | 0.0178 | 0.0057 | +0.0121 |
| aiin | 0.0071 | 0.0190 | -0.0119 |
| qokaiin | 0.0036 | 0.0152 | -0.0117 |
| daiin | 0.0249 | 0.0133 | +0.0116 |
| ain | 0.0000 | 0.0114 | -0.0114 |
| oty | 0.0107 | 0.0000 | +0.0107 |
| qokeedy | 0.0107 | 0.0000 | +0.0107 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0254 | 0.0272 | -0.0018 |
| chedy | 0.0254 | 0.0408 | -0.0154 |
| shedy | 0.0181 | 0.0466 | -0.0285 |
| chey | 0.0181 | 0.0175 | +0.0006 |
| ol | 0.0290 | 0.0155 | +0.0135 |
| shey | 0.0217 | 0.0136 | +0.0081 |
| or | 0.0109 | 0.0136 | -0.0027 |
| ar | 0.0362 | 0.0097 | +0.0265 |
| chor | 0.0109 | 0.0097 | +0.0012 |
| al | 0.0217 | 0.0078 | +0.0140 |
| chdy | 0.0109 | 0.0078 | +0.0031 |
| okaiin | 0.0072 | 0.0117 | -0.0044 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0181 | 0.0466 | -0.0285 |
| ar | 0.0362 | 0.0097 | +0.0265 |
| oty | 0.0181 | 0.0000 | +0.0181 |
| cheey | 0.0036 | 0.0214 | -0.0177 |
| kaiin | 0.0036 | 0.0194 | -0.0158 |
| cheol | 0.0000 | 0.0155 | -0.0155 |
| chedy | 0.0254 | 0.0408 | -0.0154 |
| sheey | 0.0072 | 0.0214 | -0.0141 |
| al | 0.0217 | 0.0078 | +0.0140 |
| daiin | 0.0036 | 0.0175 | -0.0139 |
| ol | 0.0290 | 0.0155 | +0.0135 |
| kedy | 0.0000 | 0.0117 | -0.0117 |

Context diagnostics: predecessor Jaccard 0.1474, JS 0.3798, entropy A/B 5.174/5.389, effective vocabulary A/B 176.54/219.09; successor Jaccard 0.1623, JS 0.4338, entropy A/B 5.002/5.093, effective vocabulary A/B 148.65/162.90.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `sheey`, `tedy`, `dain`, `otaiin`, `shey`, `dol`, `qokeey`, `chedy`, `cheol`, `chey`, `okal`, `qokedy`; right `chckhy`, `dy`, `okedy`, `y`, `ain`, `chdy`, `chor`, `dar`, `okaiin`, `or`, `s`, `chey`.

Shared unobserved high-frequency contexts (descriptive absence only): left `chckhy`, `shy`, `char`, `air`, `cheody`, `cheo`, `dor`, `odaiin`, `qotchy`, `otey`, `raiin`, `cheedy`; right `qokedy`, `qokain`, `okar`, `qol`, `cthy`, `qokey`, `otain`, `shy`, `qotedy`, `qotaiin`, `saiin`, `okeol`.

## `ar` / `ol`

Structural similarity: 0.6657; reliability: 0.9336; normalized graphemic distance: 1.0000; counts: 402/557.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9384 | 0.9874 |
| Left context | 0.5764 | 0.9051 |
| Right context | 0.4824 | 0.9084 |

- Primary component: positional agreement (0.938).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.576.
- Largest right-context difference: aiin is more frequent for ar (absolute probability difference 0.046).

Position summaries (A/B): line-start 0.0100/0.0557, line-end 0.0796/0.0754, mean 6.647/5.210, median 6.000/4.000. Position JS similarity: 0.9384.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| or | 0.0226 | 0.0190 | +0.0036 |
| ar | 0.0201 | 0.0171 | +0.0030 |
| dar | 0.0251 | 0.0152 | +0.0099 |
| qokain | 0.0151 | 0.0247 | -0.0096 |
| r | 0.0151 | 0.0152 | -0.0001 |
| ol | 0.0126 | 0.0152 | -0.0026 |
| ain | 0.0151 | 0.0114 | +0.0037 |
| okar | 0.0302 | 0.0114 | +0.0187 |
| al | 0.0226 | 0.0076 | +0.0150 |
| dain | 0.0101 | 0.0076 | +0.0024 |
| s | 0.0503 | 0.0076 | +0.0426 |
| daiin | 0.0075 | 0.0133 | -0.0058 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0503 | 0.0076 | +0.0426 |
| otar | 0.0302 | 0.0057 | +0.0244 |
| okar | 0.0302 | 0.0114 | +0.0187 |
| char | 0.0151 | 0.0000 | +0.0151 |
| al | 0.0226 | 0.0076 | +0.0150 |
| aiin | 0.0050 | 0.0190 | -0.0140 |
| chol | 0.0050 | 0.0171 | -0.0121 |
| otain | 0.0176 | 0.0057 | +0.0119 |
| dar | 0.0251 | 0.0152 | +0.0099 |
| qokain | 0.0151 | 0.0247 | -0.0096 |
| qokal | 0.0000 | 0.0095 | -0.0095 |
| qokedy | 0.0025 | 0.0114 | -0.0089 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0730 | 0.0272 | +0.0458 |
| ol | 0.0243 | 0.0155 | +0.0088 |
| or | 0.0270 | 0.0136 | +0.0134 |
| cheey | 0.0108 | 0.0214 | -0.0105 |
| chey | 0.0108 | 0.0175 | -0.0067 |
| ar | 0.0216 | 0.0097 | +0.0119 |
| chedy | 0.0081 | 0.0408 | -0.0327 |
| daiin | 0.0081 | 0.0175 | -0.0094 |
| shedy | 0.0081 | 0.0466 | -0.0385 |
| shey | 0.0081 | 0.0136 | -0.0055 |
| al | 0.0459 | 0.0078 | +0.0382 |
| cheor | 0.0081 | 0.0058 | +0.0023 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0730 | 0.0272 | +0.0458 |
| shedy | 0.0081 | 0.0466 | -0.0385 |
| al | 0.0459 | 0.0078 | +0.0382 |
| chedy | 0.0081 | 0.0408 | -0.0327 |
| am | 0.0189 | 0.0000 | +0.0189 |
| kaiin | 0.0027 | 0.0194 | -0.0167 |
| aiiin | 0.0162 | 0.0000 | +0.0162 |
| sheey | 0.0054 | 0.0214 | -0.0160 |
| ain | 0.0189 | 0.0039 | +0.0150 |
| air | 0.0189 | 0.0039 | +0.0150 |
| y | 0.0189 | 0.0039 | +0.0150 |
| s | 0.0027 | 0.0175 | -0.0148 |

Context diagnostics: predecessor Jaccard 0.1585, JS 0.4527, entropy A/B 5.112/5.389, effective vocabulary A/B 165.93/219.09; successor Jaccard 0.1546, JS 0.4126, entropy A/B 4.974/5.093, effective vocabulary A/B 144.66/162.90.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `chees`, `dal`, `okal`, `saiin`, `ches`, `kar`, `lor`, `okaiin`, `okain`, `qokar`, `sain`, `chor`; right `chcthy`, `chedar`, `o`, `okedy`, `okol`, `cheor`, `chor`, `keey`, `okaiin`, `shey`, `chey`, `daiin`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `chdy`, `okedy`, `qokol`, `shy`, `chody`, `cthy`, `cheo`, `qoteedy`, `qokchy`, `dor`, `odaiin`; right `qokedy`, `qokain`, `otal`, `qol`, `oty`, `cthy`, `qokey`, `otain`, `qotedy`, `qoty`, `shor`, `saiin`.

## `chor` / `daiin`

Structural similarity: 0.6630; reliability: 0.9336; normalized graphemic distance: 1.0000; counts: 211/847.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.8917 | 0.9874 |
| Left context | 0.5748 | 0.9051 |
| Right context | 0.5225 | 0.9084 |

- Primary component: positional agreement (0.892).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.575.
- Largest right-context difference: chol is more frequent for chor (absolute probability difference 0.029).

Position summaries (A/B): line-start 0.0569/0.1889, line-end 0.0332/0.1547, mean 3.318/4.253, median 2.000/4.000. Position JS similarity: 0.8917.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chol | 0.0352 | 0.0480 | -0.0129 |
| daiin | 0.0352 | 0.0160 | +0.0192 |
| chor | 0.0352 | 0.0131 | +0.0221 |
| ol | 0.0251 | 0.0131 | +0.0120 |
| chy | 0.0101 | 0.0131 | -0.0031 |
| cthy | 0.0101 | 0.0102 | -0.0001 |
| qokol | 0.0101 | 0.0102 | -0.0001 |
| y | 0.0101 | 0.0116 | -0.0016 |
| dal | 0.0151 | 0.0087 | +0.0063 |
| shor | 0.0101 | 0.0073 | +0.0028 |
| cheol | 0.0101 | 0.0058 | +0.0042 |
| chckhy | 0.0050 | 0.0073 | -0.0023 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| or | 0.0251 | 0.0029 | +0.0222 |
| chor | 0.0352 | 0.0131 | +0.0221 |
| daiin | 0.0352 | 0.0160 | +0.0192 |
| dar | 0.0151 | 0.0015 | +0.0136 |
| shol | 0.0000 | 0.0131 | -0.0131 |
| chol | 0.0352 | 0.0480 | -0.0129 |
| ol | 0.0251 | 0.0131 | +0.0120 |
| cheey | 0.0000 | 0.0116 | -0.0116 |
| qoky | 0.0050 | 0.0160 | -0.0110 |
| cheor | 0.0101 | 0.0000 | +0.0101 |
| qokchor | 0.0101 | 0.0000 | +0.0101 |
| qot | 0.0101 | 0.0000 | +0.0101 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| cthy | 0.0196 | 0.0154 | +0.0042 |
| daiin | 0.0441 | 0.0154 | +0.0288 |
| cthor | 0.0147 | 0.0112 | +0.0035 |
| or | 0.0147 | 0.0112 | +0.0035 |
| chckhy | 0.0098 | 0.0126 | -0.0028 |
| chey | 0.0098 | 0.0182 | -0.0084 |
| chol | 0.0392 | 0.0098 | +0.0294 |
| chor | 0.0343 | 0.0098 | +0.0245 |
| cthol | 0.0147 | 0.0098 | +0.0049 |
| ol | 0.0098 | 0.0098 | +0.0000 |
| chy | 0.0245 | 0.0084 | +0.0161 |
| dy | 0.0098 | 0.0084 | +0.0014 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chol | 0.0392 | 0.0098 | +0.0294 |
| daiin | 0.0441 | 0.0154 | +0.0288 |
| chor | 0.0343 | 0.0098 | +0.0245 |
| aiin | 0.0245 | 0.0028 | +0.0217 |
| cheky | 0.0196 | 0.0028 | +0.0168 |
| chy | 0.0245 | 0.0084 | +0.0161 |
| ar | 0.0196 | 0.0042 | +0.0154 |
| kar | 0.0147 | 0.0000 | +0.0147 |
| shey | 0.0000 | 0.0140 | -0.0140 |
| sheey | 0.0147 | 0.0014 | +0.0133 |
| dal | 0.0049 | 0.0154 | -0.0105 |
| chal | 0.0098 | 0.0000 | +0.0098 |

Context diagnostics: predecessor Jaccard 0.1242, JS 0.3757, entropy A/B 4.835/5.613, effective vocabulary A/B 125.79/273.86; successor Jaccard 0.1175, JS 0.3730, entropy A/B 4.734/5.711, effective vocabulary A/B 113.74/302.18.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `ar`, `cheol`, `oky`, `shor`, `cthy`, `qokol`, `y`, `chy`, `dal`; right `al`, `dy`, `ol`, `chckhy`, `cthol`, `cthor`, `or`, `s`, `chey`, `ar`, `cheky`, `cthy`.

Shared unobserved high-frequency contexts (descriptive absence only): left `shedy`, `okal`, `okar`, `qol`, `okain`, `okedy`, `lchedy`, `oteedy`, `ain`, `air`, `cheody`, `qoteedy`; right `qokeedy`, `qokain`, `qokaiin`, `l`, `r`, `lchedy`, `ain`, `qokey`, `qotedy`, `air`, `kaiin`, `saiin`.

## `chey` / `ol`

Structural similarity: 0.6528; reliability: 0.9336; normalized graphemic distance: 1.0000; counts: 346/557.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9501 | 0.9874 |
| Left context | 0.6934 | 0.9051 |
| Right context | 0.3149 | 0.9084 |

- Primary component: positional agreement (0.950).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.693.
- Largest right-context difference: shedy is more frequent for ol (absolute probability difference 0.047).

Position summaries (A/B): line-start 0.0173/0.0557, line-end 0.0549/0.0754, mean 4.870/5.210, median 3.000/4.000. Position JS similarity: 0.9501.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0294 | 0.0190 | +0.0104 |
| or | 0.0206 | 0.0190 | +0.0016 |
| qokain | 0.0176 | 0.0247 | -0.0071 |
| chol | 0.0176 | 0.0171 | +0.0005 |
| ol | 0.0265 | 0.0152 | +0.0113 |
| qokaiin | 0.0176 | 0.0152 | +0.0024 |
| dar | 0.0147 | 0.0152 | -0.0005 |
| daiin | 0.0382 | 0.0133 | +0.0249 |
| ar | 0.0118 | 0.0171 | -0.0053 |
| r | 0.0118 | 0.0152 | -0.0034 |
| shedy | 0.0118 | 0.0114 | +0.0004 |
| cheor | 0.0147 | 0.0095 | +0.0052 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0382 | 0.0133 | +0.0249 |
| dain | 0.0235 | 0.0076 | +0.0159 |
| okar | 0.0000 | 0.0114 | -0.0114 |
| ol | 0.0265 | 0.0152 | +0.0113 |
| okain | 0.0206 | 0.0095 | +0.0111 |
| y | 0.0147 | 0.0038 | +0.0109 |
| aiin | 0.0294 | 0.0190 | +0.0104 |
| sol | 0.0118 | 0.0019 | +0.0099 |
| dol | 0.0000 | 0.0095 | -0.0095 |
| t | 0.0088 | 0.0000 | +0.0088 |
| qokedy | 0.0029 | 0.0114 | -0.0085 |
| kaiin | 0.0118 | 0.0038 | +0.0080 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0245 | 0.0175 | +0.0070 |
| ol | 0.0153 | 0.0155 | -0.0002 |
| keey | 0.0183 | 0.0117 | +0.0067 |
| okaiin | 0.0153 | 0.0117 | +0.0036 |
| kain | 0.0183 | 0.0097 | +0.0086 |
| keedy | 0.0092 | 0.0117 | -0.0025 |
| dain | 0.0092 | 0.0078 | +0.0014 |
| dal | 0.0122 | 0.0078 | +0.0045 |
| chey | 0.0061 | 0.0175 | -0.0114 |
| or | 0.0061 | 0.0136 | -0.0075 |
| dar | 0.0092 | 0.0058 | +0.0033 |
| l | 0.0122 | 0.0058 | +0.0064 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0000 | 0.0466 | -0.0466 |
| chedy | 0.0000 | 0.0408 | -0.0408 |
| qokain | 0.0245 | 0.0000 | +0.0245 |
| aiin | 0.0031 | 0.0272 | -0.0241 |
| cheey | 0.0000 | 0.0214 | -0.0214 |
| sheey | 0.0000 | 0.0214 | -0.0214 |
| qokeey | 0.0214 | 0.0019 | +0.0195 |
| qokaiin | 0.0183 | 0.0019 | +0.0164 |
| qokeedy | 0.0183 | 0.0019 | +0.0164 |
| kaiin | 0.0031 | 0.0194 | -0.0164 |
| cheol | 0.0000 | 0.0155 | -0.0155 |
| qol | 0.0153 | 0.0000 | +0.0153 |

Context diagnostics: predecessor Jaccard 0.1820, JS 0.4978, entropy A/B 4.999/5.389, effective vocabulary A/B 148.32/219.09; successor Jaccard 0.1569, JS 0.3412, entropy A/B 5.145/5.093, effective vocabulary A/B 171.51/162.90.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `chor`, `dal`, `kar`, `lkaiin`, `okal`, `sheor`, `okaiin`, `otain`, `chey`, `qokal`, `qokar`, `qokeey`; right `chol`, `dain`, `dar`, `kor`, `qokar`, `r`, `keedy`, `dal`, `l`, `lchedy`, `qoky`, `or`.

Shared unobserved high-frequency contexts (descriptive absence only): left `o`, `chdy`, `okedy`, `oty`, `char`, `air`, `cheody`, `cheo`, `qoteedy`, `qokchy`, `dor`, `qotchy`; right `oty`, `shy`, `am`, `shor`, `okeol`, `okey`, `odaiin`, `d`, `qotain`, `qotal`, `cthol`, `qokchy`.

## `lchedy` / `qokar`

Structural similarity: 0.6606; reliability: 0.9053; normalized graphemic distance: 1.0000; counts: 116/156.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9743 | 0.9808 |
| Left context | 0.5694 | 0.8652 |
| Right context | 0.4381 | 0.8699 |

- Primary component: positional agreement (0.974).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.569.
- Largest right-context difference: chedy is more frequent for lchedy (absolute probability difference 0.034).

Position summaries (A/B): line-start 0.0517/0.0449, line-end 0.1466/0.0256, mean 4.543/4.904, median 4.000/4.000. Position JS similarity: 0.9743.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0545 | 0.0671 | -0.0126 |
| qokeedy | 0.0455 | 0.0403 | +0.0052 |
| chdy | 0.0273 | 0.0201 | +0.0071 |
| chey | 0.0364 | 0.0201 | +0.0162 |
| dal | 0.0273 | 0.0134 | +0.0138 |
| okeey | 0.0182 | 0.0134 | +0.0048 |
| ol | 0.0273 | 0.0134 | +0.0138 |
| qokedy | 0.0273 | 0.0134 | +0.0138 |
| qokeey | 0.0364 | 0.0134 | +0.0229 |
| cheedy | 0.0091 | 0.0134 | -0.0043 |
| dy | 0.0091 | 0.0134 | -0.0043 |
| qoty | 0.0091 | 0.0134 | -0.0043 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| okedy | 0.0273 | 0.0000 | +0.0273 |
| oteedy | 0.0273 | 0.0000 | +0.0273 |
| shedy | 0.0000 | 0.0268 | -0.0268 |
| qokeey | 0.0364 | 0.0134 | +0.0229 |
| qokaiin | 0.0000 | 0.0201 | -0.0201 |
| al | 0.0182 | 0.0000 | +0.0182 |
| cheey | 0.0182 | 0.0000 | +0.0182 |
| lkedy | 0.0182 | 0.0000 | +0.0182 |
| qotchedy | 0.0182 | 0.0000 | +0.0182 |
| shey | 0.0091 | 0.0268 | -0.0178 |
| chey | 0.0364 | 0.0201 | +0.0162 |
| dal | 0.0273 | 0.0134 | +0.0138 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0404 | 0.0724 | -0.0320 |
| chckhy | 0.0303 | 0.0263 | +0.0040 |
| chedy | 0.0606 | 0.0263 | +0.0343 |
| qokain | 0.0202 | 0.0197 | +0.0005 |
| qokey | 0.0303 | 0.0132 | +0.0171 |
| chey | 0.0101 | 0.0132 | -0.0031 |
| okaiin | 0.0101 | 0.0132 | -0.0031 |
| okar | 0.0101 | 0.0395 | -0.0294 |
| ol | 0.0101 | 0.0329 | -0.0228 |
| qokal | 0.0101 | 0.0197 | -0.0096 |
| chol | 0.0101 | 0.0066 | +0.0035 |
| lchedy | 0.0101 | 0.0066 | +0.0035 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0606 | 0.0263 | +0.0343 |
| shedy | 0.0404 | 0.0724 | -0.0320 |
| qokaiin | 0.0303 | 0.0000 | +0.0303 |
| qokedy | 0.0303 | 0.0000 | +0.0303 |
| qokeedy | 0.0303 | 0.0000 | +0.0303 |
| okar | 0.0101 | 0.0395 | -0.0294 |
| qokeey | 0.0303 | 0.0066 | +0.0237 |
| ol | 0.0101 | 0.0329 | -0.0228 |
| lar | 0.0202 | 0.0000 | +0.0202 |
| lkaiin | 0.0202 | 0.0000 | +0.0202 |
| lkchedy | 0.0202 | 0.0000 | +0.0202 |
| ar | 0.0000 | 0.0197 | -0.0197 |

Context diagnostics: predecessor Jaccard 0.1131, JS 0.3458, entropy A/B 4.186/4.506, effective vocabulary A/B 65.76/90.58; successor Jaccard 0.1169, JS 0.3023, entropy A/B 4.161/4.365, effective vocabulary A/B 64.12/78.66.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `okeey`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `aiin`, `chol`, `ar`, `or`, `s`, `dar`, `qokain`, `chor`, `dain`, `l`, `r`, `chy`; right `daiin`, `aiin`, `dain`, `o`, `qokar`, `sheey`, `okain`, `oteey`, `qol`, `okedy`, `cthy`, `dol`.

## `lchedy` / `qokain`

Structural similarity: 0.6529; reliability: 0.9072; normalized graphemic distance: 1.0000; counts: 116/273.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9819 | 0.9813 |
| Left context | 0.5185 | 0.8679 |
| Right context | 0.4583 | 0.8725 |

- Primary component: positional agreement (0.982).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.519.
- Largest left-context difference: shey is more frequent for qokain (absolute probability difference 0.051).

Position summaries (A/B): line-start 0.0517/0.0879, line-end 0.1466/0.0366, mean 4.543/4.201, median 4.000/4.000. Position JS similarity: 0.9819.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0545 | 0.0723 | -0.0177 |
| chey | 0.0364 | 0.0321 | +0.0042 |
| qokedy | 0.0273 | 0.0241 | +0.0032 |
| qokeedy | 0.0455 | 0.0201 | +0.0254 |
| qokeey | 0.0364 | 0.0161 | +0.0203 |
| chdy | 0.0273 | 0.0120 | +0.0152 |
| chcthy | 0.0091 | 0.0161 | -0.0070 |
| keey | 0.0091 | 0.0120 | -0.0030 |
| qokar | 0.0091 | 0.0120 | -0.0030 |
| shey | 0.0091 | 0.0602 | -0.0512 |
| chckhy | 0.0091 | 0.0080 | +0.0011 |
| cheey | 0.0182 | 0.0080 | +0.0101 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shey | 0.0091 | 0.0602 | -0.0512 |
| shedy | 0.0000 | 0.0442 | -0.0442 |
| dal | 0.0273 | 0.0000 | +0.0273 |
| okedy | 0.0273 | 0.0000 | +0.0273 |
| ol | 0.0273 | 0.0000 | +0.0273 |
| qokeedy | 0.0455 | 0.0201 | +0.0254 |
| sheey | 0.0000 | 0.0241 | -0.0241 |
| oteedy | 0.0273 | 0.0040 | +0.0233 |
| qokeey | 0.0364 | 0.0161 | +0.0203 |
| al | 0.0182 | 0.0000 | +0.0182 |
| lkedy | 0.0182 | 0.0000 | +0.0182 |
| qotchedy | 0.0182 | 0.0000 | +0.0182 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0606 | 0.0494 | +0.0112 |
| shedy | 0.0404 | 0.0342 | +0.0062 |
| chckhy | 0.0303 | 0.0380 | -0.0077 |
| qokain | 0.0202 | 0.0114 | +0.0088 |
| checthy | 0.0101 | 0.0152 | -0.0051 |
| cheedy | 0.0101 | 0.0152 | -0.0051 |
| chey | 0.0101 | 0.0228 | -0.0127 |
| dar | 0.0101 | 0.0190 | -0.0089 |
| okaiin | 0.0101 | 0.0152 | -0.0051 |
| ol | 0.0101 | 0.0494 | -0.0393 |
| okar | 0.0101 | 0.0076 | +0.0025 |
| olshey | 0.0101 | 0.0076 | +0.0025 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ol | 0.0101 | 0.0494 | -0.0393 |
| qokaiin | 0.0303 | 0.0000 | +0.0303 |
| qokeedy | 0.0303 | 0.0000 | +0.0303 |
| qokey | 0.0303 | 0.0000 | +0.0303 |
| qokedy | 0.0303 | 0.0038 | +0.0265 |
| qokeey | 0.0303 | 0.0038 | +0.0265 |
| ar | 0.0000 | 0.0228 | -0.0228 |
| shey | 0.0000 | 0.0228 | -0.0228 |
| lar | 0.0202 | 0.0000 | +0.0202 |
| lkchedy | 0.0202 | 0.0000 | +0.0202 |
| okain | 0.0000 | 0.0190 | -0.0190 |
| lkaiin | 0.0202 | 0.0038 | +0.0164 |

Context diagnostics: predecessor Jaccard 0.1518, JS 0.3817, entropy A/B 4.186/4.578, effective vocabulary A/B 65.76/97.28; successor Jaccard 0.0964, JS 0.2925, entropy A/B 4.161/4.610, effective vocabulary A/B 64.12/100.43.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `cheey`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `daiin`, `ar`, `or`, `s`, `dar`, `qokaiin`, `chor`, `dain`, `o`, `l`, `r`, `chy`; right `aiin`, `y`, `shol`, `dain`, `chy`, `qokar`, `otaiin`, `chdy`, `oteey`, `qol`, `ain`, `okeedy`.

## `lchedy` / `qol`

Structural similarity: 0.6523; reliability: 0.8963; normalized graphemic distance: 1.0000; counts: 116/139.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9666 | 0.9787 |
| Left context | 0.5220 | 0.8525 |
| Right context | 0.4683 | 0.8576 |

- Primary component: positional agreement (0.967).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.522.
- Largest left-context difference: shedy is more frequent for qol (absolute probability difference 0.066).

Position summaries (A/B): line-start 0.0517/0.1223, line-end 0.1466/0.0647, mean 4.543/3.863, median 4.000/4.000. Position JS similarity: 0.9666.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0545 | 0.1066 | -0.0520 |
| chey | 0.0364 | 0.0410 | -0.0046 |
| qokeedy | 0.0455 | 0.0328 | +0.0127 |
| qokedy | 0.0273 | 0.0246 | +0.0027 |
| cheey | 0.0182 | 0.0246 | -0.0064 |
| dal | 0.0273 | 0.0164 | +0.0109 |
| okeey | 0.0182 | 0.0164 | +0.0018 |
| oteey | 0.0091 | 0.0246 | -0.0155 |
| qoky | 0.0091 | 0.0164 | -0.0073 |
| sheedy | 0.0091 | 0.0246 | -0.0155 |
| shey | 0.0091 | 0.0164 | -0.0073 |
| y | 0.0091 | 0.0164 | -0.0073 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0000 | 0.0656 | -0.0656 |
| chedy | 0.0545 | 0.1066 | -0.0520 |
| qokeey | 0.0364 | 0.0000 | +0.0364 |
| qol | 0.0000 | 0.0328 | -0.0328 |
| chdy | 0.0273 | 0.0000 | +0.0273 |
| okedy | 0.0273 | 0.0000 | +0.0273 |
| ol | 0.0273 | 0.0000 | +0.0273 |
| oteedy | 0.0273 | 0.0000 | +0.0273 |
| sheey | 0.0000 | 0.0246 | -0.0246 |
| al | 0.0182 | 0.0000 | +0.0182 |
| lkedy | 0.0182 | 0.0000 | +0.0182 |
| qotchedy | 0.0182 | 0.0000 | +0.0182 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0606 | 0.1154 | -0.0548 |
| shedy | 0.0404 | 0.0769 | -0.0365 |
| cheedy | 0.0101 | 0.0308 | -0.0207 |
| chey | 0.0101 | 0.0385 | -0.0284 |
| okaiin | 0.0101 | 0.0154 | -0.0053 |
| ol | 0.0101 | 0.0308 | -0.0207 |
| l | 0.0101 | 0.0077 | +0.0024 |
| okeey | 0.0101 | 0.0077 | +0.0024 |
| qokal | 0.0101 | 0.0077 | +0.0024 |
| r | 0.0101 | 0.0077 | +0.0024 |
| raiin | 0.0101 | 0.0077 | +0.0024 |
| rchedy | 0.0101 | 0.0077 | +0.0024 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0606 | 0.1154 | -0.0548 |
| sheedy | 0.0000 | 0.0462 | -0.0462 |
| shedy | 0.0404 | 0.0769 | -0.0365 |
| cheey | 0.0000 | 0.0308 | -0.0308 |
| qol | 0.0000 | 0.0308 | -0.0308 |
| chckhy | 0.0303 | 0.0000 | +0.0303 |
| qokaiin | 0.0303 | 0.0000 | +0.0303 |
| qokedy | 0.0303 | 0.0000 | +0.0303 |
| qokeedy | 0.0303 | 0.0000 | +0.0303 |
| qokeey | 0.0303 | 0.0000 | +0.0303 |
| qokey | 0.0303 | 0.0000 | +0.0303 |
| chey | 0.0101 | 0.0385 | -0.0284 |

Context diagnostics: predecessor Jaccard 0.1504, JS 0.3696, entropy A/B 4.186/4.012, effective vocabulary A/B 65.76/55.23; successor Jaccard 0.0882, JS 0.2521, entropy A/B 4.161/3.927, effective vocabulary A/B 64.12/50.74.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `okeey`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `aiin`, `ar`, `or`, `s`, `dar`, `qokain`, `qokaiin`, `chor`, `okaiin`, `dain`, `shol`, `l`; right `daiin`, `ar`, `al`, `dal`, `chor`, `dain`, `o`, `qokar`, `otaiin`, `okal`, `otal`, `okain`.

## `okaiin` / `ol`

Structural similarity: 0.6837; reliability: 0.9336; normalized graphemic distance: 0.8333; counts: 215/557.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9554 | 0.9874 |
| Left context | 0.6033 | 0.9051 |
| Right context | 0.4925 | 0.9084 |

- Primary component: positional agreement (0.955).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.603.
- Largest right-context difference: otaiin is more frequent for okaiin (absolute probability difference 0.034).

Position summaries (A/B): line-start 0.0884/0.0557, line-end 0.0837/0.0754, mean 4.237/5.210, median 4.000/4.000. Position JS similarity: 0.9554.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokain | 0.0204 | 0.0247 | -0.0043 |
| aiin | 0.0408 | 0.0190 | +0.0218 |
| or | 0.0153 | 0.0190 | -0.0037 |
| ol | 0.0306 | 0.0152 | +0.0154 |
| daiin | 0.0408 | 0.0133 | +0.0275 |
| ain | 0.0102 | 0.0114 | -0.0012 |
| ar | 0.0102 | 0.0171 | -0.0069 |
| chol | 0.0102 | 0.0171 | -0.0069 |
| dar | 0.0102 | 0.0152 | -0.0050 |
| chey | 0.0255 | 0.0095 | +0.0160 |
| okain | 0.0102 | 0.0095 | +0.0007 |
| qokar | 0.0102 | 0.0095 | +0.0007 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0408 | 0.0133 | +0.0275 |
| okaiin | 0.0306 | 0.0076 | +0.0230 |
| aiin | 0.0408 | 0.0190 | +0.0218 |
| chey | 0.0255 | 0.0095 | +0.0160 |
| ol | 0.0306 | 0.0152 | +0.0154 |
| chckhy | 0.0153 | 0.0000 | +0.0153 |
| air | 0.0102 | 0.0000 | +0.0102 |
| cheody | 0.0102 | 0.0000 | +0.0102 |
| okeor | 0.0102 | 0.0000 | +0.0102 |
| otey | 0.0102 | 0.0000 | +0.0102 |
| qokeeody | 0.0102 | 0.0000 | +0.0102 |
| qokaiin | 0.0051 | 0.0152 | -0.0101 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0305 | 0.0466 | -0.0161 |
| chedy | 0.0203 | 0.0408 | -0.0205 |
| ol | 0.0203 | 0.0155 | +0.0048 |
| cheey | 0.0152 | 0.0214 | -0.0061 |
| daiin | 0.0152 | 0.0175 | -0.0022 |
| okaiin | 0.0305 | 0.0117 | +0.0188 |
| chey | 0.0102 | 0.0175 | -0.0073 |
| or | 0.0102 | 0.0136 | -0.0034 |
| sheey | 0.0102 | 0.0214 | -0.0112 |
| ar | 0.0102 | 0.0097 | +0.0004 |
| al | 0.0152 | 0.0078 | +0.0075 |
| chckhy | 0.0305 | 0.0058 | +0.0246 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| otaiin | 0.0355 | 0.0019 | +0.0336 |
| aiin | 0.0000 | 0.0272 | -0.0272 |
| chckhy | 0.0305 | 0.0058 | +0.0246 |
| chedy | 0.0203 | 0.0408 | -0.0205 |
| kaiin | 0.0000 | 0.0194 | -0.0194 |
| okaiin | 0.0305 | 0.0117 | +0.0188 |
| shedy | 0.0305 | 0.0466 | -0.0161 |
| cheol | 0.0000 | 0.0155 | -0.0155 |
| cthy | 0.0152 | 0.0000 | +0.0152 |
| shckhy | 0.0152 | 0.0000 | +0.0152 |
| cheody | 0.0152 | 0.0019 | +0.0133 |
| okal | 0.0152 | 0.0019 | +0.0133 |

Context diagnostics: predecessor Jaccard 0.1540, JS 0.4398, entropy A/B 4.761/5.389, effective vocabulary A/B 116.90/219.09; successor Jaccard 0.1487, JS 0.3792, entropy A/B 4.738/5.093, effective vocabulary A/B 114.20/162.90.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `dain`, `okain`, `okal`, `okol`, `otal`, `qokar`, `qol`, `saiin`, `shey`, `taiin`, `ain`, `dar`; right `ar`, `oteedy`, `y`, `or`, `al`, `chol`, `chey`, `daiin`.

Shared unobserved high-frequency contexts (descriptive absence only): left `o`, `shol`, `oty`, `shy`, `cthy`, `cheo`, `qokchy`, `dor`, `qotchy`, `shckhy`, `raiin`, `cheedy`; right `qokedy`, `otal`, `qol`, `oty`, `okeedy`, `qokey`, `am`, `qotedy`, `qoty`, `shor`, `qotaiin`, `saiin`.

## `okain` / `ol`

Structural similarity: 0.6815; reliability: 0.9228; normalized graphemic distance: 0.8000; counts: 140/557.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9489 | 0.9849 |
| Left context | 0.5726 | 0.8898 |
| Right context | 0.5229 | 0.8937 |

- Primary component: positional agreement (0.949).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.573.
- Largest right-context difference: chey is more frequent for okain (absolute probability difference 0.038).

Position summaries (A/B): line-start 0.0857/0.0557, line-end 0.0929/0.0754, mean 4.607/5.210, median 4.000/4.000. Position JS similarity: 0.9489.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokain | 0.0391 | 0.0247 | +0.0143 |
| aiin | 0.0469 | 0.0190 | +0.0279 |
| or | 0.0312 | 0.0190 | +0.0122 |
| ar | 0.0312 | 0.0171 | +0.0141 |
| qokaiin | 0.0391 | 0.0152 | +0.0239 |
| qokedy | 0.0312 | 0.0114 | +0.0198 |
| shedy | 0.0156 | 0.0114 | +0.0042 |
| chedy | 0.0156 | 0.0095 | +0.0061 |
| chey | 0.0156 | 0.0095 | +0.0061 |
| qokeey | 0.0312 | 0.0095 | +0.0217 |
| chol | 0.0078 | 0.0171 | -0.0093 |
| okain | 0.0078 | 0.0095 | -0.0017 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0469 | 0.0190 | +0.0279 |
| qokaiin | 0.0391 | 0.0152 | +0.0239 |
| qokeey | 0.0312 | 0.0095 | +0.0217 |
| qokedy | 0.0312 | 0.0114 | +0.0198 |
| char | 0.0156 | 0.0000 | +0.0156 |
| dar | 0.0000 | 0.0152 | -0.0152 |
| r | 0.0000 | 0.0152 | -0.0152 |
| qokain | 0.0391 | 0.0247 | +0.0143 |
| ar | 0.0312 | 0.0171 | +0.0141 |
| daiin | 0.0000 | 0.0133 | -0.0133 |
| or | 0.0312 | 0.0190 | +0.0122 |
| ain | 0.0000 | 0.0114 | -0.0114 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0394 | 0.0408 | -0.0014 |
| shedy | 0.0236 | 0.0466 | -0.0230 |
| chey | 0.0551 | 0.0175 | +0.0376 |
| cheey | 0.0157 | 0.0214 | -0.0056 |
| ol | 0.0394 | 0.0155 | +0.0238 |
| shey | 0.0157 | 0.0136 | +0.0022 |
| okaiin | 0.0157 | 0.0117 | +0.0041 |
| ar | 0.0236 | 0.0097 | +0.0139 |
| aiin | 0.0079 | 0.0272 | -0.0193 |
| cheedy | 0.0079 | 0.0117 | -0.0038 |
| cheol | 0.0079 | 0.0155 | -0.0077 |
| sheey | 0.0079 | 0.0214 | -0.0135 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chey | 0.0551 | 0.0175 | +0.0376 |
| y | 0.0315 | 0.0039 | +0.0276 |
| ol | 0.0394 | 0.0155 | +0.0238 |
| qokain | 0.0236 | 0.0000 | +0.0236 |
| shedy | 0.0236 | 0.0466 | -0.0230 |
| kaiin | 0.0000 | 0.0194 | -0.0194 |
| aiin | 0.0079 | 0.0272 | -0.0193 |
| daiin | 0.0000 | 0.0175 | -0.0175 |
| s | 0.0000 | 0.0175 | -0.0175 |
| chear | 0.0157 | 0.0000 | +0.0157 |
| okar | 0.0157 | 0.0000 | +0.0157 |
| ar | 0.0236 | 0.0097 | +0.0139 |

Context diagnostics: predecessor Jaccard 0.0964, JS 0.3530, entropy A/B 4.415/5.389, effective vocabulary A/B 82.67/219.09; successor Jaccard 0.1364, JS 0.3894, entropy A/B 4.369/5.093, effective vocabulary A/B 78.93/162.90.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `chedy`, `chey`, `otain`, `shedy`; right `chcthy`, `chdy`, `okaiin`, `shey`.

Shared unobserved high-frequency contexts (descriptive absence only): left `o`, `shol`, `okedy`, `qokol`, `oty`, `shy`, `chody`, `air`, `cheody`, `cthy`, `cheo`, `qoteedy`; right `qokedy`, `otal`, `qol`, `oty`, `okeedy`, `cthy`, `qokey`, `otain`, `shy`, `qotedy`, `qoty`, `shor`.

## `qokaiin` / `qol`

Structural similarity: 0.7386; reliability: 0.9222; normalized graphemic distance: 0.7143; counts: 264/139.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9781 | 0.9847 |
| Left context | 0.6838 | 0.8890 |
| Right context | 0.5538 | 0.8929 |

- Primary component: positional agreement (0.978).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.684.
- Largest right-context difference: chedy is more frequent for qol (absolute probability difference 0.080).

Position summaries (A/B): line-start 0.1098/0.1223, line-end 0.0455/0.0647, mean 3.924/3.863, median 4.000/4.000. Position JS similarity: 0.9781.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0723 | 0.0656 | +0.0068 |
| chedy | 0.0426 | 0.1066 | -0.0640 |
| chey | 0.0255 | 0.0410 | -0.0155 |
| qokeedy | 0.0213 | 0.0328 | -0.0115 |
| sheedy | 0.0170 | 0.0246 | -0.0076 |
| sheey | 0.0170 | 0.0246 | -0.0076 |
| shey | 0.0340 | 0.0164 | +0.0176 |
| cheey | 0.0128 | 0.0246 | -0.0118 |
| oteey | 0.0085 | 0.0246 | -0.0161 |
| qokedy | 0.0085 | 0.0246 | -0.0161 |
| olchedy | 0.0085 | 0.0082 | +0.0003 |
| cheedy | 0.0043 | 0.0082 | -0.0039 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0426 | 0.1066 | -0.0640 |
| qol | 0.0000 | 0.0328 | -0.0328 |
| shey | 0.0340 | 0.0164 | +0.0176 |
| qokeey | 0.0170 | 0.0000 | +0.0170 |
| otedy | 0.0000 | 0.0164 | -0.0164 |
| qokey | 0.0000 | 0.0164 | -0.0164 |
| qoky | 0.0000 | 0.0164 | -0.0164 |
| sshey | 0.0000 | 0.0164 | -0.0164 |
| oteey | 0.0085 | 0.0246 | -0.0161 |
| qokedy | 0.0085 | 0.0246 | -0.0161 |
| chey | 0.0255 | 0.0410 | -0.0155 |
| aiiin | 0.0128 | 0.0000 | +0.0128 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0357 | 0.1154 | -0.0797 |
| shedy | 0.0357 | 0.0769 | -0.0412 |
| ol | 0.0317 | 0.0308 | +0.0010 |
| chey | 0.0238 | 0.0385 | -0.0147 |
| shey | 0.0159 | 0.0231 | -0.0072 |
| cheey | 0.0079 | 0.0308 | -0.0228 |
| cheol | 0.0079 | 0.0154 | -0.0074 |
| otain | 0.0079 | 0.0154 | -0.0074 |
| chcthy | 0.0079 | 0.0077 | +0.0002 |
| chdy | 0.0119 | 0.0077 | +0.0042 |
| dy | 0.0079 | 0.0077 | +0.0002 |
| oly | 0.0079 | 0.0077 | +0.0002 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0357 | 0.1154 | -0.0797 |
| sheedy | 0.0000 | 0.0462 | -0.0462 |
| shedy | 0.0357 | 0.0769 | -0.0412 |
| qol | 0.0000 | 0.0308 | -0.0308 |
| cheedy | 0.0040 | 0.0308 | -0.0268 |
| sheey | 0.0000 | 0.0231 | -0.0231 |
| cheey | 0.0079 | 0.0308 | -0.0228 |
| okain | 0.0198 | 0.0000 | +0.0198 |
| okal | 0.0198 | 0.0000 | +0.0198 |
| shckhy | 0.0198 | 0.0000 | +0.0198 |
| checkhy | 0.0159 | 0.0000 | +0.0159 |
| chol | 0.0159 | 0.0000 | +0.0159 |

Context diagnostics: predecessor Jaccard 0.1179, JS 0.3939, entropy A/B 4.793/4.012, effective vocabulary A/B 120.63/55.23; successor Jaccard 0.1053, JS 0.3362, entropy A/B 4.814/3.927, effective vocabulary A/B 123.28/50.74.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left ``; right `cheol`, `otain`.

Shared unobserved high-frequency contexts (descriptive absence only): left `or`, `s`, `qokain`, `al`, `chor`, `dain`, `l`, `cheol`, `r`, `qokar`, `otar`, `otaiin`; right `daiin`, `s`, `qokeedy`, `qokeey`, `qokain`, `o`, `qoky`, `oteey`, `okedy`, `lchedy`, `cthy`, `dol`.

## `daiin` / `dol`

Structural similarity: 0.6674; reliability: 0.9020; normalized graphemic distance: 0.8000; counts: 847/109.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9437 | 0.9801 |
| Left context | 0.6111 | 0.8605 |
| Right context | 0.4474 | 0.8654 |

- Primary component: positional agreement (0.944).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.611.
- Largest right-context difference: ol is more frequent for dol (absolute probability difference 0.039).

Position summaries (A/B): line-start 0.1889/0.1284, line-end 0.1547/0.0550, mean 4.253/5.018, median 4.000/4.000. Position JS similarity: 0.9437.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chol | 0.0480 | 0.0632 | -0.0151 |
| daiin | 0.0160 | 0.0421 | -0.0261 |
| qokeey | 0.0146 | 0.0211 | -0.0065 |
| dy | 0.0131 | 0.0211 | -0.0080 |
| ol | 0.0131 | 0.0316 | -0.0185 |
| cheey | 0.0116 | 0.0105 | +0.0011 |
| qoky | 0.0160 | 0.0105 | +0.0055 |
| shol | 0.0131 | 0.0105 | +0.0026 |
| y | 0.0116 | 0.0105 | +0.0011 |
| cthol | 0.0087 | 0.0105 | -0.0018 |
| dal | 0.0087 | 0.0211 | -0.0123 |
| otol | 0.0073 | 0.0105 | -0.0032 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokal | 0.0044 | 0.0316 | -0.0272 |
| daiin | 0.0160 | 0.0421 | -0.0261 |
| ol | 0.0131 | 0.0316 | -0.0185 |
| qokchy | 0.0044 | 0.0211 | -0.0167 |
| qotol | 0.0044 | 0.0211 | -0.0167 |
| chol | 0.0480 | 0.0632 | -0.0151 |
| chedy | 0.0146 | 0.0000 | +0.0146 |
| shey | 0.0073 | 0.0211 | -0.0138 |
| chor | 0.0131 | 0.0000 | +0.0131 |
| chy | 0.0131 | 0.0000 | +0.0131 |
| dal | 0.0087 | 0.0211 | -0.0123 |
| chey | 0.0116 | 0.0000 | +0.0116 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0154 | 0.0194 | -0.0041 |
| shey | 0.0140 | 0.0194 | -0.0055 |
| dain | 0.0126 | 0.0194 | -0.0068 |
| or | 0.0112 | 0.0291 | -0.0180 |
| dar | 0.0098 | 0.0194 | -0.0096 |
| ol | 0.0098 | 0.0485 | -0.0388 |
| y | 0.0098 | 0.0291 | -0.0193 |
| chckhy | 0.0126 | 0.0097 | +0.0029 |
| chedy | 0.0140 | 0.0097 | +0.0043 |
| dal | 0.0154 | 0.0097 | +0.0057 |
| chy | 0.0084 | 0.0097 | -0.0013 |
| dy | 0.0084 | 0.0097 | -0.0013 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ol | 0.0098 | 0.0485 | -0.0388 |
| shedy | 0.0070 | 0.0388 | -0.0319 |
| y | 0.0098 | 0.0291 | -0.0193 |
| chey | 0.0182 | 0.0000 | +0.0182 |
| checthy | 0.0014 | 0.0194 | -0.0180 |
| oty | 0.0014 | 0.0194 | -0.0180 |
| sheey | 0.0014 | 0.0194 | -0.0180 |
| or | 0.0112 | 0.0291 | -0.0180 |
| cthy | 0.0154 | 0.0000 | +0.0154 |
| dair | 0.0042 | 0.0194 | -0.0152 |
| cheol | 0.0056 | 0.0194 | -0.0138 |
| cthor | 0.0112 | 0.0000 | +0.0112 |

Context diagnostics: predecessor Jaccard 0.0909, JS 0.3593, entropy A/B 5.613/4.225, effective vocabulary A/B 273.86/68.40; successor Jaccard 0.0891, JS 0.3234, entropy A/B 5.711/4.304, effective vocabulary A/B 302.18/74.01.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left ``; right `cheol`, `daiin`, `dain`, `dair`, `dar`, `shey`, `shol`.

Shared unobserved high-frequency contexts (descriptive absence only): left `shedy`, `qokaiin`, `qokar`, `okal`, `okar`, `qol`, `okain`, `lchedy`, `oteedy`, `ain`, `cheor`, `char`; right `qokeedy`, `qokain`, `qokaiin`, `l`, `r`, `okain`, `lchedy`, `ain`, `qokey`, `qotedy`, `air`, `kaiin`.

## `qokain` / `qol`

Structural similarity: 0.7666; reliability: 0.9222; normalized graphemic distance: 0.6667; counts: 273/139.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9612 | 0.9847 |
| Left context | 0.7088 | 0.8890 |
| Right context | 0.6298 | 0.8929 |

- Primary component: positional agreement (0.961).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.709.
- Largest right-context difference: chedy is more frequent for qol (absolute probability difference 0.066).

Position summaries (A/B): line-start 0.0879/0.1223, line-end 0.0366/0.0647, mean 4.201/3.863, median 4.000/4.000. Position JS similarity: 0.9612.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0723 | 0.1066 | -0.0343 |
| shedy | 0.0442 | 0.0656 | -0.0214 |
| chey | 0.0321 | 0.0410 | -0.0089 |
| qokedy | 0.0241 | 0.0246 | -0.0005 |
| sheey | 0.0241 | 0.0246 | -0.0005 |
| qokeedy | 0.0201 | 0.0328 | -0.0127 |
| shey | 0.0602 | 0.0164 | +0.0438 |
| otedy | 0.0120 | 0.0164 | -0.0043 |
| checthy | 0.0080 | 0.0082 | -0.0002 |
| cheey | 0.0080 | 0.0246 | -0.0166 |
| oteey | 0.0080 | 0.0246 | -0.0166 |
| qo | 0.0080 | 0.0082 | -0.0002 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shey | 0.0602 | 0.0164 | +0.0438 |
| chedy | 0.0723 | 0.1066 | -0.0343 |
| qol | 0.0000 | 0.0328 | -0.0328 |
| shedy | 0.0442 | 0.0656 | -0.0214 |
| sheedy | 0.0040 | 0.0246 | -0.0206 |
| cheey | 0.0080 | 0.0246 | -0.0166 |
| oteey | 0.0080 | 0.0246 | -0.0166 |
| dal | 0.0000 | 0.0164 | -0.0164 |
| qoky | 0.0000 | 0.0164 | -0.0164 |
| sshey | 0.0000 | 0.0164 | -0.0164 |
| y | 0.0000 | 0.0164 | -0.0164 |
| chcthy | 0.0161 | 0.0000 | +0.0161 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0494 | 0.1154 | -0.0660 |
| shedy | 0.0342 | 0.0769 | -0.0427 |
| ol | 0.0494 | 0.0308 | +0.0187 |
| chey | 0.0228 | 0.0385 | -0.0156 |
| shey | 0.0228 | 0.0231 | -0.0003 |
| cheedy | 0.0152 | 0.0308 | -0.0156 |
| okaiin | 0.0152 | 0.0154 | -0.0002 |
| otar | 0.0152 | 0.0154 | -0.0002 |
| sheey | 0.0152 | 0.0231 | -0.0079 |
| cheey | 0.0114 | 0.0308 | -0.0194 |
| cheol | 0.0114 | 0.0154 | -0.0040 |
| otain | 0.0114 | 0.0154 | -0.0040 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0494 | 0.1154 | -0.0660 |
| shedy | 0.0342 | 0.0769 | -0.0427 |
| sheedy | 0.0076 | 0.0462 | -0.0385 |
| chckhy | 0.0380 | 0.0000 | +0.0380 |
| qol | 0.0000 | 0.0308 | -0.0308 |
| ar | 0.0228 | 0.0000 | +0.0228 |
| cheey | 0.0114 | 0.0308 | -0.0194 |
| dar | 0.0190 | 0.0000 | +0.0190 |
| okain | 0.0190 | 0.0000 | +0.0190 |
| ol | 0.0494 | 0.0308 | +0.0187 |
| chey | 0.0228 | 0.0385 | -0.0156 |
| cheedy | 0.0152 | 0.0308 | -0.0156 |

Context diagnostics: predecessor Jaccard 0.1302, JS 0.4209, entropy A/B 4.578/4.012, effective vocabulary A/B 97.28/55.23; successor Jaccard 0.1134, JS 0.4043, entropy A/B 4.610/3.927, effective vocabulary A/B 100.43/50.74.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `otedy`, `qokey`; right `cheol`, `okaiin`, `otain`, `otar`.

Shared unobserved high-frequency contexts (descriptive absence only): left `ol`, `ar`, `or`, `s`, `dar`, `qokaiin`, `al`, `chor`, `dain`, `l`, `cheol`, `r`; right `chol`, `s`, `qokeedy`, `qokaiin`, `dain`, `qokar`, `otaiin`, `oteey`, `lchedy`, `okeedy`, `cthy`, `dol`.

## `aiin` / `ar`

Structural similarity: 0.6710; reliability: 0.9336; normalized graphemic distance: 0.7500; counts: 504/402.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9580 | 0.9874 |
| Left context | 0.6441 | 0.9051 |
| Right context | 0.4109 | 0.9084 |

- Primary component: positional agreement (0.958).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.644.
- Largest left-context difference: or is more frequent for aiin (absolute probability difference 0.090).

Position summaries (A/B): line-start 0.0000/0.0100, line-end 0.0813/0.0796, mean 6.583/6.647, median 6.000/6.000. Position JS similarity: 0.9580.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.1171 | 0.0503 | +0.0668 |
| or | 0.1131 | 0.0226 | +0.0905 |
| ar | 0.0536 | 0.0201 | +0.0335 |
| r | 0.0496 | 0.0151 | +0.0345 |
| dar | 0.0139 | 0.0251 | -0.0112 |
| ol | 0.0278 | 0.0126 | +0.0152 |
| char | 0.0099 | 0.0151 | -0.0052 |
| chor | 0.0099 | 0.0101 | -0.0001 |
| otar | 0.0099 | 0.0302 | -0.0202 |
| ches | 0.0139 | 0.0075 | +0.0064 |
| lor | 0.0099 | 0.0075 | +0.0024 |
| o | 0.0139 | 0.0075 | +0.0064 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| or | 0.1131 | 0.0226 | +0.0905 |
| s | 0.1171 | 0.0503 | +0.0668 |
| r | 0.0496 | 0.0151 | +0.0345 |
| ar | 0.0536 | 0.0201 | +0.0335 |
| okar | 0.0000 | 0.0302 | -0.0302 |
| al | 0.0020 | 0.0226 | -0.0206 |
| otar | 0.0099 | 0.0302 | -0.0202 |
| otain | 0.0000 | 0.0176 | -0.0176 |
| d | 0.0159 | 0.0000 | +0.0159 |
| ol | 0.0278 | 0.0126 | +0.0152 |
| qokain | 0.0000 | 0.0151 | -0.0151 |
| ain | 0.0020 | 0.0151 | -0.0131 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ol | 0.0216 | 0.0243 | -0.0027 |
| al | 0.0194 | 0.0459 | -0.0265 |
| y | 0.0130 | 0.0189 | -0.0060 |
| chey | 0.0216 | 0.0108 | +0.0108 |
| okain | 0.0130 | 0.0108 | +0.0021 |
| am | 0.0108 | 0.0189 | -0.0081 |
| cheey | 0.0086 | 0.0108 | -0.0022 |
| otar | 0.0194 | 0.0081 | +0.0113 |
| oteey | 0.0086 | 0.0081 | +0.0005 |
| shedy | 0.0086 | 0.0081 | +0.0005 |
| chedy | 0.0065 | 0.0081 | -0.0016 |
| or | 0.0065 | 0.0270 | -0.0205 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0000 | 0.0730 | -0.0730 |
| al | 0.0194 | 0.0459 | -0.0265 |
| okal | 0.0216 | 0.0000 | +0.0216 |
| or | 0.0065 | 0.0270 | -0.0205 |
| ain | 0.0000 | 0.0189 | -0.0189 |
| air | 0.0000 | 0.0189 | -0.0189 |
| ar | 0.0043 | 0.0216 | -0.0173 |
| aiiin | 0.0000 | 0.0162 | -0.0162 |
| okaiin | 0.0173 | 0.0054 | +0.0119 |
| otar | 0.0194 | 0.0081 | +0.0113 |
| d | 0.0108 | 0.0000 | +0.0108 |
| chey | 0.0216 | 0.0108 | +0.0108 |

Context diagnostics: predecessor Jaccard 0.1604, JS 0.4445, entropy A/B 4.485/5.112, effective vocabulary A/B 88.70/165.93; successor Jaccard 0.1382, JS 0.3640, entropy A/B 5.437/4.974, effective vocabulary A/B 229.86/144.66.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `lkar`, `los`, `ytar`, `daiin`, `dair`, `tar`, `lor`, `sor`, `chor`, `ches`, `o`, `char`; right `chcthy`, `cheky`, `chor`, `okol`, `shol`, `yteey`, `otedy`, `sheey`, `chedy`, `oky`, `shey`, `o`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `qokal`, `cheey`, `okeey`, `dy`, `chy`, `chdy`, `okedy`, `lchedy`, `okeedy`, `oteedy`, `qokol`; right `qokedy`, `qokal`, `dain`, `qokar`, `qoky`, `qol`, `lchedy`, `sheol`, `dol`, `qokey`, `sho`, `dam`.

## `chol` / `cthy`

Structural similarity: 0.6832; reliability: 0.8972; normalized graphemic distance: 0.7500; counts: 395/103.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9271 | 0.9791 |
| Left context | 0.4803 | 0.8538 |
| Right context | 0.6424 | 0.8589 |

- Primary component: positional agreement (0.927).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.642.
- Largest left-context difference: daiin is more frequent for cthy (absolute probability difference 0.088).

Position summaries (A/B): line-start 0.0481/0.0000, line-end 0.0127/0.2524, mean 3.947/4.330, median 3.000/4.000. Position JS similarity: 0.9271.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chol | 0.0638 | 0.0485 | +0.0153 |
| chor | 0.0213 | 0.0388 | -0.0176 |
| daiin | 0.0186 | 0.1068 | -0.0882 |
| sho | 0.0106 | 0.0291 | -0.0185 |
| shol | 0.0106 | 0.0291 | -0.0185 |
| choky | 0.0106 | 0.0097 | +0.0009 |
| cthol | 0.0106 | 0.0097 | +0.0009 |
| aiin | 0.0080 | 0.0097 | -0.0017 |
| cthor | 0.0080 | 0.0097 | -0.0017 |
| okaiin | 0.0080 | 0.0291 | -0.0211 |
| chey | 0.0053 | 0.0097 | -0.0044 |
| chody | 0.0053 | 0.0097 | -0.0044 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0186 | 0.1068 | -0.0882 |
| chy | 0.0000 | 0.0388 | -0.0388 |
| shor | 0.0000 | 0.0388 | -0.0388 |
| s | 0.0213 | 0.0000 | +0.0213 |
| okaiin | 0.0080 | 0.0291 | -0.0211 |
| chodaiin | 0.0000 | 0.0194 | -0.0194 |
| dchy | 0.0000 | 0.0194 | -0.0194 |
| shaiin | 0.0000 | 0.0194 | -0.0194 |
| yodaiin | 0.0000 | 0.0194 | -0.0194 |
| or | 0.0186 | 0.0000 | +0.0186 |
| otol | 0.0186 | 0.0000 | +0.0186 |
| sho | 0.0106 | 0.0291 | -0.0185 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0846 | 0.0909 | -0.0063 |
| chol | 0.0615 | 0.0260 | +0.0356 |
| dy | 0.0205 | 0.0390 | -0.0184 |
| chor | 0.0179 | 0.0260 | -0.0080 |
| chey | 0.0154 | 0.0130 | +0.0024 |
| chy | 0.0205 | 0.0130 | +0.0075 |
| cthy | 0.0128 | 0.0130 | -0.0002 |
| dain | 0.0128 | 0.0260 | -0.0132 |
| or | 0.0103 | 0.0130 | -0.0027 |
| s | 0.0103 | 0.0519 | -0.0417 |
| sho | 0.0077 | 0.0130 | -0.0053 |
| chaiin | 0.0051 | 0.0130 | -0.0079 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0103 | 0.0519 | -0.0417 |
| chol | 0.0615 | 0.0260 | +0.0356 |
| ol | 0.0231 | 0.0000 | +0.0231 |
| dy | 0.0205 | 0.0390 | -0.0184 |
| shol | 0.0179 | 0.0000 | +0.0179 |
| cthol | 0.0154 | 0.0000 | +0.0154 |
| dol | 0.0154 | 0.0000 | +0.0154 |
| dain | 0.0128 | 0.0260 | -0.0132 |
| ?ain | 0.0000 | 0.0130 | -0.0130 |
| chakal | 0.0000 | 0.0130 | -0.0130 |
| chckhey | 0.0000 | 0.0130 | -0.0130 |
| che'eky | 0.0000 | 0.0130 | -0.0130 |

Context diagnostics: predecessor Jaccard 0.1183, JS 0.3385, entropy A/B 5.199/3.949, effective vocabulary A/B 181.14/51.87; successor Jaccard 0.0830, JS 0.3521, entropy A/B 4.932/3.998, effective vocabulary A/B 138.66/54.49.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `sheey`, `shey`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `ar`, `qokeedy`, `qokain`, `o`, `r`, `otedy`, `otar`, `oteey`, `qol`, `chckhy`, `okain`, `okedy`; right `shey`, `qokedy`, `qokal`, `r`, `okeey`, `qokar`, `okar`, `oteey`, `okedy`, `lchedy`, `ain`, `okeedy`.

## `lchedy` / `qokeey`

Structural similarity: 0.7038; reliability: 0.9072; normalized graphemic distance: 0.6667; counts: 116/306.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9574 | 0.9813 |
| Left context | 0.6246 | 0.8679 |
| Right context | 0.5294 | 0.8725 |

- Primary component: positional agreement (0.957).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.625.
- Largest left-context difference: shedy is more frequent for qokeey (absolute probability difference 0.047).

Position summaries (A/B): line-start 0.0517/0.1046, line-end 0.1466/0.0359, mean 4.543/4.075, median 4.000/4.000. Position JS similarity: 0.9574.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0545 | 0.0693 | -0.0148 |
| qokeey | 0.0364 | 0.0438 | -0.0074 |
| qokeedy | 0.0455 | 0.0328 | +0.0126 |
| chey | 0.0364 | 0.0255 | +0.0108 |
| cheey | 0.0182 | 0.0182 | -0.0001 |
| okeey | 0.0182 | 0.0328 | -0.0147 |
| cheol | 0.0091 | 0.0109 | -0.0019 |
| keey | 0.0091 | 0.0146 | -0.0055 |
| lchedy | 0.0091 | 0.0109 | -0.0019 |
| sheedy | 0.0091 | 0.0109 | -0.0019 |
| shey | 0.0091 | 0.0182 | -0.0092 |
| chdy | 0.0273 | 0.0073 | +0.0200 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0000 | 0.0474 | -0.0474 |
| dal | 0.0273 | 0.0000 | +0.0273 |
| ol | 0.0273 | 0.0036 | +0.0236 |
| chdy | 0.0273 | 0.0073 | +0.0200 |
| okedy | 0.0273 | 0.0073 | +0.0200 |
| oteedy | 0.0273 | 0.0073 | +0.0200 |
| qokedy | 0.0273 | 0.0073 | +0.0200 |
| lkedy | 0.0182 | 0.0000 | +0.0182 |
| chedy | 0.0545 | 0.0693 | -0.0148 |
| okeey | 0.0182 | 0.0328 | -0.0147 |
| qokey | 0.0000 | 0.0146 | -0.0146 |
| al | 0.0182 | 0.0036 | +0.0145 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokedy | 0.0303 | 0.0305 | -0.0002 |
| qokeedy | 0.0303 | 0.0305 | -0.0002 |
| qokeey | 0.0303 | 0.0407 | -0.0104 |
| chedy | 0.0606 | 0.0169 | +0.0437 |
| qokaiin | 0.0303 | 0.0136 | +0.0167 |
| qokain | 0.0202 | 0.0136 | +0.0066 |
| qoky | 0.0202 | 0.0136 | +0.0066 |
| chckhy | 0.0303 | 0.0102 | +0.0201 |
| qokey | 0.0303 | 0.0102 | +0.0201 |
| chey | 0.0101 | 0.0102 | -0.0001 |
| l | 0.0101 | 0.0102 | -0.0001 |
| lchedy | 0.0101 | 0.0136 | -0.0035 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0606 | 0.0169 | +0.0437 |
| daiin | 0.0000 | 0.0339 | -0.0339 |
| shedy | 0.0404 | 0.0068 | +0.0336 |
| lar | 0.0202 | 0.0000 | +0.0202 |
| lkchedy | 0.0202 | 0.0000 | +0.0202 |
| chckhy | 0.0303 | 0.0102 | +0.0201 |
| qokey | 0.0303 | 0.0102 | +0.0201 |
| lkaiin | 0.0202 | 0.0034 | +0.0168 |
| qokaiin | 0.0303 | 0.0136 | +0.0167 |
| okeey | 0.0101 | 0.0237 | -0.0136 |
| okain | 0.0000 | 0.0136 | -0.0136 |
| qotedy | 0.0000 | 0.0136 | -0.0136 |

Context diagnostics: predecessor Jaccard 0.1814, JS 0.4494, entropy A/B 4.186/4.700, effective vocabulary A/B 65.76/109.92; successor Jaccard 0.1304, JS 0.3854, entropy A/B 4.161/4.952, effective vocabulary A/B 64.12/141.40.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `cheey`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `chol`, `ar`, `or`, `s`, `qokaiin`, `chor`, `l`, `r`, `chy`, `otedy`, `otar`, `otaiin`; right `dy`, `al`, `cheey`, `chy`, `chdy`, `otar`, `qol`, `ain`, `cthy`, `sho`, `shy`, `oky`.

## `qokedy` / `qol`

Structural similarity: 0.6862; reliability: 0.9222; normalized graphemic distance: 0.6667; counts: 276/139.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9586 | 0.9847 |
| Left context | 0.6707 | 0.8890 |
| Right context | 0.4292 | 0.8929 |

- Primary component: positional agreement (0.959).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.671.
- Largest right-context difference: chedy is more frequent for qol (absolute probability difference 0.074).

Position summaries (A/B): line-start 0.0725/0.1223, line-end 0.0362/0.0647, mean 3.986/3.863, median 4.000/4.000. Position JS similarity: 0.9586.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0820 | 0.0656 | +0.0165 |
| chedy | 0.0547 | 0.1066 | -0.0519 |
| qokeedy | 0.0469 | 0.0328 | +0.0141 |
| qokedy | 0.0625 | 0.0246 | +0.0379 |
| sheedy | 0.0195 | 0.0246 | -0.0051 |
| otedy | 0.0234 | 0.0164 | +0.0070 |
| shey | 0.0156 | 0.0164 | -0.0008 |
| qokey | 0.0117 | 0.0164 | -0.0047 |
| daiin | 0.0078 | 0.0082 | -0.0004 |
| dy | 0.0078 | 0.0082 | -0.0004 |
| sheey | 0.0078 | 0.0246 | -0.0168 |
| ychedy | 0.0078 | 0.0082 | -0.0004 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0547 | 0.1066 | -0.0519 |
| qokedy | 0.0625 | 0.0246 | +0.0379 |
| chey | 0.0039 | 0.0410 | -0.0371 |
| qokeey | 0.0352 | 0.0000 | +0.0352 |
| qol | 0.0000 | 0.0328 | -0.0328 |
| okedy | 0.0234 | 0.0000 | +0.0234 |
| cheey | 0.0039 | 0.0246 | -0.0207 |
| oteey | 0.0039 | 0.0246 | -0.0207 |
| sheey | 0.0078 | 0.0246 | -0.0168 |
| shedy | 0.0820 | 0.0656 | +0.0165 |
| y | 0.0000 | 0.0164 | -0.0164 |
| qokeedy | 0.0469 | 0.0328 | +0.0141 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0414 | 0.1154 | -0.0740 |
| shedy | 0.0338 | 0.0769 | -0.0431 |
| ol | 0.0226 | 0.0308 | -0.0082 |
| qol | 0.0113 | 0.0308 | -0.0195 |
| dy | 0.0150 | 0.0077 | +0.0073 |
| otedy | 0.0263 | 0.0077 | +0.0186 |
| oteedy | 0.0113 | 0.0077 | +0.0036 |
| qokal | 0.0188 | 0.0077 | +0.0111 |
| chdy | 0.0075 | 0.0077 | -0.0002 |
| lchey | 0.0075 | 0.0077 | -0.0002 |
| or | 0.0075 | 0.0077 | -0.0002 |
| otar | 0.0075 | 0.0154 | -0.0079 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0414 | 0.1154 | -0.0740 |
| qokedy | 0.0602 | 0.0000 | +0.0602 |
| qokeedy | 0.0564 | 0.0000 | +0.0564 |
| shedy | 0.0338 | 0.0769 | -0.0431 |
| sheedy | 0.0038 | 0.0462 | -0.0424 |
| chey | 0.0038 | 0.0385 | -0.0347 |
| cheedy | 0.0000 | 0.0308 | -0.0308 |
| cheey | 0.0000 | 0.0308 | -0.0308 |
| dal | 0.0263 | 0.0000 | +0.0263 |
| qokain | 0.0226 | 0.0000 | +0.0226 |
| qol | 0.0113 | 0.0308 | -0.0195 |
| sheey | 0.0038 | 0.0231 | -0.0193 |

Context diagnostics: predecessor Jaccard 0.1443, JS 0.4361, entropy A/B 4.499/4.012, effective vocabulary A/B 89.96/55.23; successor Jaccard 0.1263, JS 0.3246, entropy A/B 4.606/3.927, effective vocabulary A/B 100.07/50.74.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `qokey`, `shey`; right `otar`.

Shared unobserved high-frequency contexts (descriptive absence only): left `ol`, `aiin`, `ar`, `or`, `s`, `dar`, `chor`, `okaiin`, `dain`, `shol`, `l`, `cheol`; right `chor`, `o`, `oteey`, `oty`, `okeedy`, `cthy`, `sho`, `shy`, `oky`, `cheor`, `cheody`, `dam`.

## `chedy` / `qokeey`

Structural similarity: 0.6726; reliability: 0.9336; normalized graphemic distance: 0.6667; counts: 504/306.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9276 | 0.9874 |
| Left context | 0.3801 | 0.9051 |
| Right context | 0.7102 | 0.9084 |

- Primary component: positional agreement (0.928).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.710.
- Largest left-context difference: chedy is more frequent for qokeey (absolute probability difference 0.049).

Position summaries (A/B): line-start 0.0119/0.1046, line-end 0.0694/0.0359, mean 5.393/4.075, median 5.000/4.000. Position JS similarity: 0.9276.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0201 | 0.0693 | -0.0493 |
| shedy | 0.0161 | 0.0474 | -0.0314 |
| qokeedy | 0.0141 | 0.0328 | -0.0188 |
| daiin | 0.0201 | 0.0109 | +0.0091 |
| lchedy | 0.0120 | 0.0109 | +0.0011 |
| qokeey | 0.0100 | 0.0438 | -0.0338 |
| okeedy | 0.0080 | 0.0109 | -0.0029 |
| dar | 0.0141 | 0.0073 | +0.0068 |
| okedy | 0.0080 | 0.0073 | +0.0007 |
| qokedy | 0.0221 | 0.0073 | +0.0148 |
| aiin | 0.0060 | 0.0073 | -0.0013 |
| cheol | 0.0040 | 0.0109 | -0.0069 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0201 | 0.0693 | -0.0493 |
| ol | 0.0422 | 0.0036 | +0.0385 |
| qokeey | 0.0100 | 0.0438 | -0.0338 |
| shedy | 0.0161 | 0.0474 | -0.0314 |
| qol | 0.0301 | 0.0000 | +0.0301 |
| okeey | 0.0040 | 0.0328 | -0.0288 |
| chey | 0.0000 | 0.0255 | -0.0255 |
| qokain | 0.0261 | 0.0036 | +0.0225 |
| l | 0.0221 | 0.0000 | +0.0221 |
| qokeedy | 0.0141 | 0.0328 | -0.0188 |
| qokal | 0.0221 | 0.0036 | +0.0184 |
| cheey | 0.0000 | 0.0182 | -0.0182 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokeey | 0.0405 | 0.0407 | -0.0002 |
| qokedy | 0.0299 | 0.0305 | -0.0007 |
| daiin | 0.0213 | 0.0339 | -0.0126 |
| qokeedy | 0.0192 | 0.0305 | -0.0113 |
| chedy | 0.0213 | 0.0169 | +0.0044 |
| qokaiin | 0.0213 | 0.0136 | +0.0078 |
| qokain | 0.0384 | 0.0136 | +0.0248 |
| lchedy | 0.0128 | 0.0136 | -0.0008 |
| okeey | 0.0107 | 0.0237 | -0.0131 |
| ol | 0.0107 | 0.0169 | -0.0063 |
| qokal | 0.0128 | 0.0102 | +0.0026 |
| qokey | 0.0128 | 0.0102 | +0.0026 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qol | 0.0277 | 0.0000 | +0.0277 |
| qokain | 0.0384 | 0.0136 | +0.0248 |
| qokar | 0.0213 | 0.0068 | +0.0145 |
| okeey | 0.0107 | 0.0237 | -0.0131 |
| daiin | 0.0213 | 0.0339 | -0.0126 |
| qokeedy | 0.0192 | 0.0305 | -0.0113 |
| lsheey | 0.0000 | 0.0102 | -0.0102 |
| oteey | 0.0000 | 0.0102 | -0.0102 |
| qotain | 0.0128 | 0.0034 | +0.0094 |
| okain | 0.0043 | 0.0136 | -0.0093 |
| dy | 0.0085 | 0.0000 | +0.0085 |
| okar | 0.0085 | 0.0000 | +0.0085 |

Context diagnostics: predecessor Jaccard 0.1354, JS 0.3364, entropy A/B 5.070/4.700, effective vocabulary A/B 159.23/109.92; successor Jaccard 0.2065, JS 0.5174, entropy A/B 5.127/4.952, effective vocabulary A/B 168.53/141.40.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `aiin`, `oteey`, `shol`, `okedy`, `cheol`, `okeedy`, `sheedy`, `lchedy`, `dar`; right `dar`, `lchey`, `otey`, `oty`, `raiin`, `rol`, `shedy`, `okeedy`, `chckhy`, `l`, `or`, `s`.

Shared unobserved high-frequency contexts (descriptive absence only): left `s`, `chy`, `otaiin`, `sho`, `oty`, `cheor`, `char`, `cho`, `cheo`, `kaiin`, `oky`, `chcthy`; right `al`, `cheey`, `ain`, `cthy`, `sho`, `cheor`, `air`, `dair`, `kaiin`, `shor`, `cho`, `cheo`.

## `dal` / `ol`

Structural similarity: 0.6706; reliability: 0.9336; normalized graphemic distance: 0.6667; counts: 242/557.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9383 | 0.9874 |
| Left context | 0.4911 | 0.9051 |
| Right context | 0.5823 | 0.9084 |

- Primary component: positional agreement (0.938).
- Similarity is concentrated: the next component, successor-distribution overlap, is 0.582.
- Largest left-context difference: daiin is more frequent for dal (absolute probability difference 0.034).

Position summaries (A/B): line-start 0.0331/0.0557, line-end 0.2025/0.0754, mean 6.707/5.210, median 5.000/4.000. Position JS similarity: 0.9383.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0171 | 0.0190 | -0.0019 |
| qokain | 0.0171 | 0.0247 | -0.0076 |
| ol | 0.0171 | 0.0152 | +0.0019 |
| daiin | 0.0470 | 0.0133 | +0.0337 |
| chol | 0.0128 | 0.0171 | -0.0043 |
| qokedy | 0.0299 | 0.0114 | +0.0185 |
| shedy | 0.0128 | 0.0114 | +0.0014 |
| chedy | 0.0171 | 0.0095 | +0.0076 |
| chey | 0.0171 | 0.0095 | +0.0076 |
| qokal | 0.0214 | 0.0095 | +0.0119 |
| al | 0.0085 | 0.0076 | +0.0009 |
| cheol | 0.0128 | 0.0057 | +0.0071 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0470 | 0.0133 | +0.0337 |
| or | 0.0000 | 0.0190 | -0.0190 |
| qokedy | 0.0299 | 0.0114 | +0.0185 |
| ar | 0.0000 | 0.0171 | -0.0171 |
| r | 0.0000 | 0.0152 | -0.0152 |
| qokal | 0.0214 | 0.0095 | +0.0119 |
| ain | 0.0000 | 0.0114 | -0.0114 |
| okar | 0.0000 | 0.0114 | -0.0114 |
| okal | 0.0171 | 0.0057 | +0.0114 |
| dar | 0.0043 | 0.0152 | -0.0109 |
| qokaiin | 0.0043 | 0.0152 | -0.0109 |
| cheey | 0.0128 | 0.0019 | +0.0109 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0415 | 0.0466 | -0.0052 |
| chedy | 0.0207 | 0.0408 | -0.0201 |
| daiin | 0.0311 | 0.0175 | +0.0136 |
| s | 0.0155 | 0.0175 | -0.0019 |
| ol | 0.0155 | 0.0155 | +0.0000 |
| or | 0.0207 | 0.0136 | +0.0071 |
| chey | 0.0104 | 0.0175 | -0.0071 |
| ar | 0.0104 | 0.0097 | +0.0007 |
| chor | 0.0155 | 0.0097 | +0.0058 |
| chy | 0.0104 | 0.0097 | +0.0007 |
| al | 0.0155 | 0.0078 | +0.0078 |
| chdy | 0.0155 | 0.0078 | +0.0078 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0052 | 0.0272 | -0.0220 |
| cheey | 0.0000 | 0.0214 | -0.0214 |
| sheey | 0.0000 | 0.0214 | -0.0214 |
| dar | 0.0259 | 0.0058 | +0.0201 |
| chedy | 0.0207 | 0.0408 | -0.0201 |
| dy | 0.0207 | 0.0039 | +0.0168 |
| kaiin | 0.0052 | 0.0194 | -0.0142 |
| daiin | 0.0311 | 0.0175 | +0.0136 |
| dair | 0.0155 | 0.0019 | +0.0136 |
| y | 0.0155 | 0.0039 | +0.0117 |
| cheedy | 0.0000 | 0.0117 | -0.0117 |
| kedy | 0.0000 | 0.0117 | -0.0117 |

Context diagnostics: predecessor Jaccard 0.1206, JS 0.3472, entropy A/B 4.969/5.389, effective vocabulary A/B 143.82/219.09; successor Jaccard 0.1416, JS 0.4093, entropy A/B 4.756/5.093, effective vocabulary A/B 116.25/162.90.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `al`, `otar`, `saiin`, `cheol`, `shedy`, `chedy`, `chey`, `okal`, `ol`, `chol`, `aiin`; right `ar`, `chol`, `chy`, `dol`, `qokar`, `sheedy`, `al`, `chdy`, `chor`, `lchedy`, `ol`, `y`.

Shared unobserved high-frequency contexts (descriptive absence only): left `o`, `shy`, `char`, `air`, `cheody`, `cthy`, `qoteedy`, `dor`, `odaiin`, `qotchy`, `shckhy`, `raiin`; right `qokain`, `otal`, `oty`, `cthy`, `qokey`, `otain`, `am`, `qotedy`, `qoty`, `shor`, `qotaiin`, `saiin`.

## `ain` / `ar`

Structural similarity: 0.6606; reliability: 0.9020; normalized graphemic distance: 0.6667; counts: 109/402.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9255 | 0.9801 |
| Left context | 0.5507 | 0.8605 |
| Right context | 0.5057 | 0.8654 |

- Primary component: positional agreement (0.926).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.551.
- Largest left-context difference: r is more frequent for ain (absolute probability difference 0.096).

Position summaries (A/B): line-start 0.0092/0.0100, line-end 0.1101/0.0796, mean 5.165/6.647, median 5.000/6.000. Position JS similarity: 0.9255.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0648 | 0.0503 | +0.0146 |
| dar | 0.0278 | 0.0251 | +0.0027 |
| or | 0.1111 | 0.0226 | +0.0885 |
| ar | 0.0648 | 0.0201 | +0.0447 |
| otar | 0.0185 | 0.0302 | -0.0116 |
| r | 0.1111 | 0.0151 | +0.0960 |
| ol | 0.0185 | 0.0126 | +0.0060 |
| ain | 0.0093 | 0.0151 | -0.0058 |
| okar | 0.0093 | 0.0302 | -0.0209 |
| otain | 0.0093 | 0.0176 | -0.0083 |
| ches | 0.0093 | 0.0075 | +0.0017 |
| dair | 0.0093 | 0.0075 | +0.0017 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| r | 0.1111 | 0.0151 | +0.0960 |
| or | 0.1111 | 0.0226 | +0.0885 |
| sar | 0.0556 | 0.0025 | +0.0530 |
| ar | 0.0648 | 0.0201 | +0.0447 |
| air | 0.0278 | 0.0050 | +0.0228 |
| al | 0.0000 | 0.0226 | -0.0226 |
| okar | 0.0093 | 0.0302 | -0.0209 |
| os | 0.0185 | 0.0000 | +0.0185 |
| lr | 0.0185 | 0.0025 | +0.0160 |
| rar | 0.0185 | 0.0025 | +0.0160 |
| char | 0.0000 | 0.0151 | -0.0151 |
| qokain | 0.0000 | 0.0151 | -0.0151 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| al | 0.0309 | 0.0459 | -0.0150 |
| ol | 0.0619 | 0.0243 | +0.0375 |
| ar | 0.0619 | 0.0216 | +0.0402 |
| or | 0.0206 | 0.0270 | -0.0064 |
| y | 0.0206 | 0.0189 | +0.0017 |
| chey | 0.0309 | 0.0108 | +0.0201 |
| aiiin | 0.0103 | 0.0162 | -0.0059 |
| aiin | 0.0103 | 0.0730 | -0.0627 |
| ain | 0.0103 | 0.0189 | -0.0086 |
| air | 0.0103 | 0.0189 | -0.0086 |
| aly | 0.0103 | 0.0108 | -0.0005 |
| am | 0.0103 | 0.0189 | -0.0086 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0103 | 0.0730 | -0.0627 |
| ar | 0.0619 | 0.0216 | +0.0402 |
| ol | 0.0619 | 0.0243 | +0.0375 |
| chl | 0.0206 | 0.0000 | +0.0206 |
| okan | 0.0206 | 0.0000 | +0.0206 |
| chey | 0.0309 | 0.0108 | +0.0201 |
| cheol | 0.0206 | 0.0027 | +0.0179 |
| okeey | 0.0206 | 0.0027 | +0.0179 |
| o | 0.0206 | 0.0054 | +0.0152 |
| okaiin | 0.0206 | 0.0054 | +0.0152 |
| al | 0.0309 | 0.0459 | -0.0150 |
| shedy | 0.0206 | 0.0081 | +0.0125 |

Context diagnostics: predecessor Jaccard 0.0952, JS 0.3810, entropy A/B 3.627/5.112, effective vocabulary A/B 37.61/165.93; successor Jaccard 0.1179, JS 0.3956, entropy A/B 4.142/4.974, effective vocabulary A/B 62.94/144.66.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `ol`, `sain`, `sor`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `y`, `qokal`, `cheey`, `okeey`, `dy`, `chy`, `sheey`, `chdy`, `qoky`, `okedy`, `lchedy`; right `qokeey`, `dy`, `qokedy`, `qokain`, `dal`, `qokal`, `qokar`, `okal`, `qoky`, `chckhy`, `oty`, `lchedy`.

## `ain` / `al`

Structural similarity: 0.6589; reliability: 0.9020; normalized graphemic distance: 0.6667; counts: 109/257.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9142 | 0.9801 |
| Left context | 0.6244 | 0.8605 |
| Right context | 0.4383 | 0.8654 |

- Primary component: positional agreement (0.914).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.624.
- Largest left-context difference: r is more frequent for ain (absolute probability difference 0.092).

Position summaries (A/B): line-start 0.0092/0.0039, line-end 0.1101/0.1128, mean 5.165/7.588, median 5.000/6.000. Position JS similarity: 0.9142.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ar | 0.0648 | 0.0664 | -0.0016 |
| s | 0.0648 | 0.0312 | +0.0336 |
| or | 0.1111 | 0.0273 | +0.0838 |
| sar | 0.0556 | 0.0273 | +0.0282 |
| dar | 0.0278 | 0.0234 | +0.0043 |
| r | 0.1111 | 0.0195 | +0.0916 |
| otar | 0.0185 | 0.0195 | -0.0010 |
| ol | 0.0185 | 0.0156 | +0.0029 |
| ain | 0.0093 | 0.0117 | -0.0025 |
| cheo | 0.0093 | 0.0117 | -0.0025 |
| dair | 0.0093 | 0.0234 | -0.0142 |
| okar | 0.0093 | 0.0117 | -0.0025 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| r | 0.1111 | 0.0195 | +0.0916 |
| or | 0.1111 | 0.0273 | +0.0838 |
| aiin | 0.0000 | 0.0352 | -0.0352 |
| s | 0.0648 | 0.0312 | +0.0336 |
| sar | 0.0556 | 0.0273 | +0.0282 |
| air | 0.0278 | 0.0078 | +0.0200 |
| daiin | 0.0000 | 0.0195 | -0.0195 |
| rar | 0.0185 | 0.0000 | +0.0185 |
| sor | 0.0185 | 0.0000 | +0.0185 |
| otaiin | 0.0000 | 0.0156 | -0.0156 |
| dair | 0.0093 | 0.0234 | -0.0142 |
| aiir | 0.0000 | 0.0117 | -0.0117 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ar | 0.0619 | 0.0395 | +0.0224 |
| ol | 0.0619 | 0.0175 | +0.0443 |
| al | 0.0309 | 0.0132 | +0.0178 |
| o | 0.0206 | 0.0132 | +0.0075 |
| chedy | 0.0103 | 0.0263 | -0.0160 |
| ches | 0.0103 | 0.0132 | -0.0028 |
| r | 0.0103 | 0.0132 | -0.0028 |
| chol | 0.0103 | 0.0088 | +0.0015 |
| okeey | 0.0206 | 0.0088 | +0.0118 |
| y | 0.0206 | 0.0088 | +0.0118 |
| aiiin | 0.0103 | 0.0044 | +0.0059 |
| aiin | 0.0103 | 0.0044 | +0.0059 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ol | 0.0619 | 0.0175 | +0.0443 |
| chey | 0.0309 | 0.0044 | +0.0265 |
| s | 0.0000 | 0.0263 | -0.0263 |
| ar | 0.0619 | 0.0395 | +0.0224 |
| dar | 0.0000 | 0.0219 | -0.0219 |
| chl | 0.0206 | 0.0000 | +0.0206 |
| okaiin | 0.0206 | 0.0000 | +0.0206 |
| okan | 0.0206 | 0.0000 | +0.0206 |
| shey | 0.0206 | 0.0000 | +0.0206 |
| al | 0.0309 | 0.0132 | +0.0178 |
| keedy | 0.0000 | 0.0175 | -0.0175 |
| cheol | 0.0206 | 0.0044 | +0.0162 |

Context diagnostics: predecessor Jaccard 0.1183, JS 0.4370, entropy A/B 3.627/4.681, effective vocabulary A/B 37.61/107.87; successor Jaccard 0.1009, JS 0.2754, entropy A/B 4.142/4.952, effective vocabulary A/B 62.94/141.49.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `lr`, `ol`, `os`, `sain`, `otar`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `chedy`, `shedy`, `qokeedy`, `qokeey`, `y`, `shey`, `qokal`, `shol`, `cheey`, `okeey`, `dy`, `cheol`; right `qokeedy`, `qokain`, `qokaiin`, `chor`, `shol`, `qokar`, `okal`, `qoky`, `chckhy`, `okar`, `okain`, `okedy`.

## `qokar` / `qol`

Structural similarity: 0.7106; reliability: 0.9202; normalized graphemic distance: 0.6000; counts: 156/139.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9469 | 0.9842 |
| Left context | 0.6409 | 0.8862 |
| Right context | 0.5438 | 0.8902 |

- Primary component: positional agreement (0.947).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.641.
- Largest right-context difference: chedy is more frequent for qol (absolute probability difference 0.089).

Position summaries (A/B): line-start 0.0449/0.1223, line-end 0.0256/0.0647, mean 4.904/3.863, median 4.000/4.000. Position JS similarity: 0.9469.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0671 | 0.1066 | -0.0394 |
| qokeedy | 0.0403 | 0.0328 | +0.0075 |
| shedy | 0.0268 | 0.0656 | -0.0387 |
| chey | 0.0201 | 0.0410 | -0.0208 |
| shey | 0.0268 | 0.0164 | +0.0105 |
| dal | 0.0134 | 0.0164 | -0.0030 |
| okeey | 0.0134 | 0.0164 | -0.0030 |
| qokedy | 0.0134 | 0.0246 | -0.0112 |
| sheedy | 0.0134 | 0.0246 | -0.0112 |
| cheedy | 0.0134 | 0.0082 | +0.0052 |
| daiin | 0.0134 | 0.0082 | +0.0052 |
| dy | 0.0134 | 0.0082 | +0.0052 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0671 | 0.1066 | -0.0394 |
| shedy | 0.0268 | 0.0656 | -0.0387 |
| qol | 0.0000 | 0.0328 | -0.0328 |
| cheey | 0.0000 | 0.0246 | -0.0246 |
| oteey | 0.0000 | 0.0246 | -0.0246 |
| sheey | 0.0000 | 0.0246 | -0.0246 |
| chey | 0.0201 | 0.0410 | -0.0208 |
| chdy | 0.0201 | 0.0000 | +0.0201 |
| qokaiin | 0.0201 | 0.0000 | +0.0201 |
| qoky | 0.0000 | 0.0164 | -0.0164 |
| sshey | 0.0000 | 0.0164 | -0.0164 |
| cheodain | 0.0134 | 0.0000 | +0.0134 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0724 | 0.0769 | -0.0046 |
| ol | 0.0329 | 0.0308 | +0.0021 |
| chedy | 0.0263 | 0.1154 | -0.0891 |
| shey | 0.0197 | 0.0231 | -0.0033 |
| chey | 0.0132 | 0.0385 | -0.0253 |
| okaiin | 0.0132 | 0.0154 | -0.0022 |
| otar | 0.0132 | 0.0154 | -0.0022 |
| sheedy | 0.0132 | 0.0462 | -0.0330 |
| chcthy | 0.0132 | 0.0077 | +0.0055 |
| chl | 0.0132 | 0.0077 | +0.0055 |
| chy | 0.0132 | 0.0077 | +0.0055 |
| or | 0.0197 | 0.0077 | +0.0120 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0263 | 0.1154 | -0.0891 |
| okar | 0.0395 | 0.0000 | +0.0395 |
| sheedy | 0.0132 | 0.0462 | -0.0330 |
| cheedy | 0.0000 | 0.0308 | -0.0308 |
| qol | 0.0000 | 0.0308 | -0.0308 |
| chckhy | 0.0263 | 0.0000 | +0.0263 |
| chey | 0.0132 | 0.0385 | -0.0253 |
| cheey | 0.0066 | 0.0308 | -0.0242 |
| sheey | 0.0000 | 0.0231 | -0.0231 |
| ar | 0.0197 | 0.0000 | +0.0197 |
| ary | 0.0197 | 0.0000 | +0.0197 |
| checkhy | 0.0197 | 0.0000 | +0.0197 |

Context diagnostics: predecessor Jaccard 0.1018, JS 0.3498, entropy A/B 4.506/4.012, effective vocabulary A/B 90.58/55.23; successor Jaccard 0.1701, JS 0.4008, entropy A/B 4.365/3.927, effective vocabulary A/B 78.66/50.74.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `dal`, `okeey`; right `okaiin`, `otar`.

Shared unobserved high-frequency contexts (descriptive absence only): left `aiin`, `ar`, `or`, `s`, `dar`, `qokain`, `al`, `chor`, `dain`, `l`, `r`, `chy`; right `daiin`, `dar`, `qokeedy`, `qokedy`, `qokaiin`, `dain`, `o`, `qokar`, `okain`, `oteey`, `okedy`, `cthy`.

## `qol` / `qotain`

Structural similarity: 0.6690; reliability: 0.8379; normalized graphemic distance: 0.6667; counts: 139/60.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9045 | 0.9628 |
| Left context | 0.6187 | 0.7716 |
| Right context | 0.4838 | 0.7793 |

- Primary component: positional agreement (0.904).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.619.
- Largest right-context difference: shedy is more frequent for qol (absolute probability difference 0.077).

Position summaries (A/B): line-start 0.1223/0.0500, line-end 0.0647/0.0667, mean 3.863/5.050, median 4.000/5.000. Position JS similarity: 0.9045.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.1066 | 0.1053 | +0.0013 |
| shedy | 0.0656 | 0.0526 | +0.0129 |
| qokeedy | 0.0328 | 0.0526 | -0.0198 |
| cheey | 0.0246 | 0.0351 | -0.0105 |
| sheey | 0.0246 | 0.0351 | -0.0105 |
| sheedy | 0.0246 | 0.0175 | +0.0070 |
| shey | 0.0164 | 0.0175 | -0.0012 |
| chol | 0.0082 | 0.0175 | -0.0093 |
| qokshedy | 0.0082 | 0.0175 | -0.0093 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chey | 0.0410 | 0.0000 | +0.0410 |
| chody | 0.0000 | 0.0351 | -0.0351 |
| otar | 0.0000 | 0.0351 | -0.0351 |
| qoteey | 0.0000 | 0.0351 | -0.0351 |
| qol | 0.0328 | 0.0000 | +0.0328 |
| oteey | 0.0246 | 0.0000 | +0.0246 |
| qokedy | 0.0246 | 0.0000 | +0.0246 |
| qokeedy | 0.0328 | 0.0526 | -0.0198 |
| chchy | 0.0000 | 0.0175 | -0.0175 |
| chckhy | 0.0000 | 0.0175 | -0.0175 |
| cheed | 0.0000 | 0.0175 | -0.0175 |
| kalkal | 0.0000 | 0.0175 | -0.0175 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.1154 | 0.1250 | -0.0096 |
| chey | 0.0385 | 0.0179 | +0.0206 |
| ol | 0.0308 | 0.0179 | +0.0129 |
| cheol | 0.0154 | 0.0179 | -0.0025 |
| okaiin | 0.0154 | 0.0179 | -0.0025 |
| otar | 0.0154 | 0.0179 | -0.0025 |
| chcthy | 0.0077 | 0.0179 | -0.0102 |
| oteedy | 0.0077 | 0.0714 | -0.0637 |
| sheckhy | 0.0077 | 0.0179 | -0.0102 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0769 | 0.0000 | +0.0769 |
| oteedy | 0.0077 | 0.0714 | -0.0637 |
| shcthy | 0.0000 | 0.0536 | -0.0536 |
| sheedy | 0.0462 | 0.0000 | +0.0462 |
| ar | 0.0000 | 0.0357 | -0.0357 |
| okedy | 0.0000 | 0.0357 | -0.0357 |
| qokain | 0.0000 | 0.0357 | -0.0357 |
| cheedy | 0.0308 | 0.0000 | +0.0308 |
| cheey | 0.0308 | 0.0000 | +0.0308 |
| qol | 0.0308 | 0.0000 | +0.0308 |
| sheey | 0.0231 | 0.0000 | +0.0231 |
| shey | 0.0231 | 0.0000 | +0.0231 |

Context diagnostics: predecessor Jaccard 0.0826, JS 0.3247, entropy A/B 4.012/3.617, effective vocabulary A/B 55.23/37.23; successor Jaccard 0.0841, JS 0.2589, entropy A/B 3.927/3.550, effective vocabulary A/B 50.74/34.81.

Shared unobserved high-frequency contexts (descriptive absence only): left `ol`, `aiin`, `ar`, `or`, `s`, `dar`, `qokaiin`, `al`, `chor`, `okaiin`, `dain`, `shol`; right `daiin`, `chol`, `s`, `dar`, `qokeedy`, `qokeey`, `al`, `qokedy`, `qokaiin`, `dal`, `chor`, `dain`.

## `okar` / `otain`

Structural similarity: 0.6619; reliability: 0.8802; normalized graphemic distance: 0.6000; counts: 140/95.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9357 | 0.9750 |
| Left context | 0.4265 | 0.8298 |
| Right context | 0.6235 | 0.8357 |

- Primary component: positional agreement (0.936).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.623.
- Largest right-context difference: chdy is more frequent for okar (absolute probability difference 0.038).

Position summaries (A/B): line-start 0.0786/0.0211, line-end 0.0571/0.0947, mean 5.579/5.474, median 5.000/5.000. Position JS similarity: 0.9357.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0233 | 0.0323 | -0.0090 |
| otain | 0.0233 | 0.0323 | -0.0090 |
| shedy | 0.0233 | 0.0430 | -0.0198 |
| daiin | 0.0310 | 0.0215 | +0.0095 |
| qokaiin | 0.0233 | 0.0215 | +0.0018 |
| okaiin | 0.0155 | 0.0215 | -0.0060 |
| qokain | 0.0155 | 0.0323 | -0.0168 |
| chedy | 0.0310 | 0.0108 | +0.0203 |
| qokal | 0.0155 | 0.0108 | +0.0048 |
| qokar | 0.0465 | 0.0108 | +0.0358 |
| qokedy | 0.0233 | 0.0108 | +0.0125 |
| chedal | 0.0078 | 0.0108 | -0.0030 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokar | 0.0465 | 0.0108 | +0.0358 |
| otaiin | 0.0078 | 0.0430 | -0.0353 |
| oteey | 0.0000 | 0.0323 | -0.0323 |
| chckhy | 0.0000 | 0.0215 | -0.0215 |
| l | 0.0000 | 0.0215 | -0.0215 |
| qol | 0.0000 | 0.0215 | -0.0215 |
| shey | 0.0000 | 0.0215 | -0.0215 |
| chedy | 0.0310 | 0.0108 | +0.0203 |
| shedy | 0.0233 | 0.0430 | -0.0198 |
| qokain | 0.0155 | 0.0323 | -0.0168 |
| chdy | 0.0155 | 0.0000 | +0.0155 |
| okain | 0.0155 | 0.0000 | +0.0155 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ar | 0.0909 | 0.0814 | +0.0095 |
| ol | 0.0455 | 0.0349 | +0.0106 |
| shedy | 0.0303 | 0.0465 | -0.0162 |
| otar | 0.0379 | 0.0233 | +0.0146 |
| al | 0.0227 | 0.0233 | -0.0005 |
| okedy | 0.0303 | 0.0116 | +0.0187 |
| shey | 0.0152 | 0.0116 | +0.0035 |
| y | 0.0303 | 0.0116 | +0.0187 |
| ain | 0.0076 | 0.0116 | -0.0041 |
| char | 0.0076 | 0.0116 | -0.0041 |
| chcthy | 0.0076 | 0.0116 | -0.0041 |
| chedy | 0.0076 | 0.0116 | -0.0041 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chdy | 0.0379 | 0.0000 | +0.0379 |
| chey | 0.0000 | 0.0349 | -0.0349 |
| otain | 0.0000 | 0.0349 | -0.0349 |
| okar | 0.0076 | 0.0349 | -0.0273 |
| otal | 0.0076 | 0.0349 | -0.0273 |
| chckhy | 0.0000 | 0.0233 | -0.0233 |
| okedy | 0.0303 | 0.0116 | +0.0187 |
| y | 0.0303 | 0.0116 | +0.0187 |
| shedy | 0.0303 | 0.0465 | -0.0162 |
| okain | 0.0076 | 0.0233 | -0.0157 |
| okol | 0.0152 | 0.0000 | +0.0152 |
| oky | 0.0152 | 0.0000 | +0.0152 |

Context diagnostics: predecessor Jaccard 0.1234, JS 0.3177, entropy A/B 4.488/4.167, effective vocabulary A/B 88.90/64.54; successor Jaccard 0.1642, JS 0.4073, entropy A/B 4.261/3.975, effective vocabulary A/B 70.85/53.27.

Shared unobserved high-frequency contexts (descriptive absence only): left `ol`, `chol`, `or`, `qokeedy`, `dar`, `y`, `al`, `chor`, `o`, `dy`, `cheol`, `r`; right `aiin`, `s`, `qokeedy`, `qokeey`, `qokain`, `dal`, `chor`, `cheey`, `qokal`, `l`, `shol`, `cheol`.

## Negative controls

Controls match unordered log-counts, normalized graphemic distance, and reliability, while favoring structural similarity near the full-corpus median. They are decomposed with exactly the target metrics.

| Target | Control | Structural | Reliability | Distance | Match cost |
|---|---|---:|---:|---:|---:|
| or/s | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 1.8591 |
| or/s | am/qokeey | 0.2792 | 0.8817 | 1.0000 | 1.9138 |
| or/s | ar/qotchy | 0.2761 | 0.8511 | 1.0000 | 2.0656 |
| chol/daiin | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 2.5861 |
| chol/daiin | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 2.7590 |
| chol/daiin | aiin/sor | 0.2553 | 0.8318 | 1.0000 | 2.8129 |
| r/s | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 1.0254 |
| r/s | am/qokeey | 0.2792 | 0.8817 | 1.0000 | 1.0802 |
| r/s | am/qokedy | 0.2926 | 0.8817 | 1.0000 | 1.3163 |
| ol/y | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 1.9065 |
| ol/y | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 2.0794 |
| ol/y | aiin/sor | 0.2553 | 0.8318 | 1.0000 | 2.1333 |
| dar/ol | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 1.9669 |
| dar/ol | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 2.1398 |
| dar/ol | aiin/sor | 0.2553 | 0.8318 | 1.0000 | 2.1937 |
| ar/ol | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 2.1851 |
| ar/ol | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 2.3580 |
| ar/ol | aiin/sor | 0.2553 | 0.8318 | 1.0000 | 2.4119 |
| chor/daiin | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 1.9613 |
| chor/daiin | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 2.1342 |
| chor/daiin | aiin/sor | 0.2553 | 0.8318 | 1.0000 | 2.1881 |
| chey/ol | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 2.0355 |
| chey/ol | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 2.2084 |
| chey/ol | aiin/sor | 0.2553 | 0.8318 | 1.0000 | 2.2623 |
| lchedy/qokar | cthy/saiin | 0.2946 | 0.8794 | 1.0000 | 0.7060 |
| lchedy/qokar | chy/sar | 0.2867 | 0.8765 | 1.0000 | 0.7099 |
| lchedy/qokar | chdy/sar | 0.2846 | 0.8697 | 1.0000 | 0.7246 |
| lchedy/qokain | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 0.5847 |
| lchedy/qokain | am/qokeey | 0.2792 | 0.8817 | 1.0000 | 0.6395 |
| lchedy/qokain | am/qokedy | 0.2926 | 0.8817 | 1.0000 | 0.6700 |
| lchedy/qol | cthy/saiin | 0.2946 | 0.8794 | 1.0000 | 0.5733 |
| lchedy/qol | chdy/sar | 0.2846 | 0.8697 | 1.0000 | 0.6896 |
| lchedy/qol | am/cthy | 0.2691 | 0.8481 | 1.0000 | 0.7671 |
| okaiin/ol | sar/shedy | 0.2977 | 0.8765 | 0.8000 | 1.7641 |
| okaiin/ol | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 1.8948 |
| okaiin/ol | shedy/sor | 0.2564 | 0.8318 | 0.8000 | 1.9955 |
| okain/ol | sar/shedy | 0.2977 | 0.8765 | 0.8000 | 1.2492 |
| okain/ol | shedy/sor | 0.2564 | 0.8318 | 0.8000 | 1.4807 |
| okain/ol | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 1.5133 |
| qokaiin/qol | sar/shey | 0.3025 | 0.8765 | 0.7500 | 1.1275 |
| qokaiin/qol | al/sol | 0.2925 | 0.8586 | 0.6667 | 1.2828 |
| qokaiin/qol | r/shor | 0.3027 | 0.8912 | 0.7500 | 1.3628 |
| daiin/dol | sar/shedy | 0.2977 | 0.8765 | 0.8000 | 1.3779 |
| daiin/dol | shedy/sor | 0.2564 | 0.8318 | 0.8000 | 1.6094 |
| daiin/dol | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 1.6420 |
| qokain/qol | sar/shey | 0.3025 | 0.8765 | 0.7500 | 1.2188 |
| qokain/qol | al/sol | 0.2925 | 0.8586 | 0.6667 | 1.2209 |
| qokain/qol | al/tol | 0.2649 | 0.8192 | 0.6667 | 1.4146 |
| aiin/ar | sar/shedy | 0.2977 | 0.8765 | 0.8000 | 2.3213 |
| aiin/ar | shedy/sor | 0.2564 | 0.8318 | 0.8000 | 2.5527 |
| aiin/ar | sar/shey | 0.3025 | 0.8765 | 0.7500 | 2.7437 |
| chol/cthy | sar/shedy | 0.2977 | 0.8765 | 0.8000 | 0.8341 |
| chol/cthy | shedy/sor | 0.2564 | 0.8318 | 0.8000 | 1.0655 |
| chol/cthy | sar/shey | 0.3025 | 0.8765 | 0.7500 | 1.0732 |
| lchedy/qokeey | sar/shey | 0.3025 | 0.8765 | 0.7500 | 1.1231 |
| lchedy/qokeey | al/sol | 0.2925 | 0.8586 | 0.6667 | 1.1253 |
| lchedy/qokeey | am/qokeedy | 0.2490 | 0.8817 | 1.0000 | 1.1377 |
| qokedy/qol | sar/shey | 0.3025 | 0.8765 | 0.7500 | 1.2297 |
| qokedy/qol | al/sol | 0.2925 | 0.8586 | 0.6667 | 1.2318 |
| qokedy/qol | otam/qokain | 0.2689 | 0.8257 | 0.6667 | 1.3286 |
| chedy/qokeey | sar/shedy | 0.2977 | 0.8765 | 0.8000 | 2.2158 |
| chedy/qokeey | shedy/sor | 0.2564 | 0.8318 | 0.8000 | 2.4473 |
| chedy/qokeey | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 2.4799 |
| dal/ol | sar/shedy | 0.2977 | 0.8765 | 0.8000 | 2.0819 |
| dal/ol | shedy/sor | 0.2564 | 0.8318 | 0.8000 | 2.3133 |
| dal/ol | aiin/qotchy | 0.2678 | 0.8511 | 1.0000 | 2.3459 |
| ain/ar | sar/shedy | 0.2977 | 0.8765 | 0.8000 | 1.0489 |
| ain/ar | shedy/sor | 0.2564 | 0.8318 | 0.8000 | 1.2803 |
| ain/ar | sar/shey | 0.3025 | 0.8765 | 0.7500 | 1.3231 |
| ain/al | sar/shey | 0.3025 | 0.8765 | 0.7500 | 0.9680 |
| ain/al | otam/qokain | 0.2689 | 0.8257 | 0.6667 | 1.0964 |
| ain/al | r/shor | 0.3027 | 0.8912 | 0.7500 | 1.1498 |
| qokar/qol | r/shor | 0.3027 | 0.8912 | 0.7500 | 1.2112 |
| qokar/qol | sar/shol | 0.2958 | 0.8765 | 0.7500 | 1.4411 |
| qokar/qol | sar/sheey | 0.3033 | 0.8707 | 0.8000 | 1.5051 |
| qol/qotain | sheey/sor | 0.2675 | 0.8265 | 0.8000 | 0.5729 |
| qol/qotain | shy/sor | 0.2632 | 0.7963 | 0.6667 | 0.6173 |
| qol/qotain | chckhy/pchedy | 0.2642 | 0.7864 | 0.6667 | 0.6385 |
| okar/otain | r/shor | 0.3027 | 0.8912 | 0.7500 | 0.9263 |
| okar/otain | am/sar | 0.2751 | 0.8289 | 0.6667 | 1.0136 |
| okar/otain | dam/saiin | 0.3053 | 0.8662 | 0.8000 | 1.0393 |

## Family decomposition

A family is a connected component; only listed edges define direct structural-distant links. Complete matrices, including non-edge pairs, are in `family_decomposition.yaml`.

### Family 1

Tokens: `aiin`, `ain`, `al`, `ar`, `chey`, `dal`, `dar`, `okaiin`, `okain`, `ol`, `y`. Structural medoid: `ol`. Peripheral token(s): `dal`.

Edges:

- `aiin` / `ar`: similarity 0.6710, reliability 0.9336, distance 0.7500
- `ain` / `al`: similarity 0.6589, reliability 0.9020, distance 0.6667
- `ain` / `ar`: similarity 0.6606, reliability 0.9020, distance 0.6667
- `ar` / `ol`: similarity 0.6657, reliability 0.9336, distance 1.0000
- `chey` / `ol`: similarity 0.6528, reliability 0.9336, distance 1.0000
- `dal` / `ol`: similarity 0.6706, reliability 0.9336, distance 0.6667
- `dar` / `ol`: similarity 0.6742, reliability 0.9336, distance 1.0000
- `okaiin` / `ol`: similarity 0.6837, reliability 0.9336, distance 0.8333
- `okain` / `ol`: similarity 0.6815, reliability 0.9228, distance 0.8000
- `ol` / `y`: similarity 0.6787, reliability 0.9336, distance 1.0000

### Family 2

Tokens: `chedy`, `lchedy`, `qokaiin`, `qokain`, `qokar`, `qokedy`, `qokeey`, `qol`, `qotain`. Structural medoid: `qokain`. Peripheral token(s): `chedy`.

Edges:

- `chedy` / `qokeey`: similarity 0.6726, reliability 0.9336, distance 0.6667
- `lchedy` / `qokain`: similarity 0.6529, reliability 0.9072, distance 1.0000
- `lchedy` / `qokar`: similarity 0.6606, reliability 0.9053, distance 1.0000
- `lchedy` / `qokeey`: similarity 0.7038, reliability 0.9072, distance 0.6667
- `lchedy` / `qol`: similarity 0.6523, reliability 0.8963, distance 1.0000
- `qokaiin` / `qol`: similarity 0.7386, reliability 0.9222, distance 0.7143
- `qokain` / `qol`: similarity 0.7666, reliability 0.9222, distance 0.6667
- `qokar` / `qol`: similarity 0.7106, reliability 0.9202, distance 0.6000
- `qokedy` / `qol`: similarity 0.6862, reliability 0.9222, distance 0.6667
- `qol` / `qotain`: similarity 0.6690, reliability 0.8379, distance 0.6667

### Family 3

Tokens: `chol`, `chor`, `cthy`, `daiin`, `dol`. Structural medoid: `chol`. Peripheral token(s): `dol`.

Edges:

- `chol` / `cthy`: similarity 0.6832, reliability 0.8972, distance 0.7500
- `chol` / `daiin`: similarity 0.7206, reliability 0.9336, distance 1.0000
- `chor` / `daiin`: similarity 0.6630, reliability 0.9336, distance 1.0000
- `daiin` / `dol`: similarity 0.6674, reliability 0.9020, distance 0.8000

### Family 4

Tokens: `okar`, `otain`. Structural medoid: `okar`. Peripheral token(s): `okar`, `otain`.

Edges:

- `okar` / `otain`: similarity 0.6619, reliability 0.8802, distance 0.6000

### Family 5

Tokens: `or`, `r`, `s`. Structural medoid: `s`. Peripheral token(s): `r`.

Edges:

- `or` / `s`: similarity 0.7594, reliability 0.9336, distance 1.0000
- `r` / `s`: similarity 0.7010, reliability 0.9336, distance 1.0000

## Limits

Observed absence is not proof of a prohibition. Context observations at line boundaries have no neighbor and therefore context totals can be below token counts. Pair rows are statistically dependent because tokens recur across pairs. Control matching is descriptive and does not make pairs independent.
