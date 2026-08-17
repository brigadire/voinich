# Structural pair decomposition

Structural similarity is reproduced unchanged from the existing pair dataset. All statements below are formal corpus descriptions; no token meaning is inferred. Context similarities and differences use full distributions, while tables are display-limited. Entropy uses natural logarithms and effective vocabulary is `exp(entropy)`.

## `or` / `s`

Structural similarity: 0.7588; reliability: 0.9332; normalized graphemic distance: 1.0000; counts: 390/350.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9474 | 0.9873 |
| Left context | 0.4171 | 0.9046 |
| Right context | 0.9118 | 0.9077 |

- Primary component: positional agreement (0.947).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.912.
- Largest right-context difference: ar is more frequent for s (absolute probability difference 0.041).

Position summaries (A/B): line-start 0.0821/0.0771, line-end 0.0462/0.1257, mean 5.146/5.823, median 4.000/5.000. Position JS similarity: 0.9474.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ol | 0.0196 | 0.0279 | -0.0083 |
| daiin | 0.0223 | 0.0186 | +0.0038 |
| chol | 0.0112 | 0.0124 | -0.0012 |
| cheey | 0.0112 | 0.0093 | +0.0019 |
| dal | 0.0112 | 0.0093 | +0.0019 |
| aiin | 0.0084 | 0.0093 | -0.0009 |
| chor | 0.0084 | 0.0093 | -0.0009 |
| l | 0.0084 | 0.0279 | -0.0195 |
| qokeey | 0.0084 | 0.0093 | -0.0009 |
| chear | 0.0084 | 0.0062 | +0.0022 |
| dar | 0.0084 | 0.0062 | +0.0022 |
| sheey | 0.0084 | 0.0062 | +0.0022 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0335 | 0.0031 | +0.0304 |
| ar | 0.0279 | 0.0031 | +0.0248 |
| chy | 0.0000 | 0.0217 | -0.0217 |
| l | 0.0084 | 0.0279 | -0.0195 |
| al | 0.0028 | 0.0186 | -0.0158 |
| or | 0.0168 | 0.0031 | +0.0137 |
| cho | 0.0000 | 0.0124 | -0.0124 |
| qokaiin | 0.0112 | 0.0000 | +0.0112 |
| chl | 0.0028 | 0.0124 | -0.0096 |
| cthy | 0.0028 | 0.0124 | -0.0096 |
| sh | 0.0000 | 0.0093 | -0.0093 |
| dair | 0.0084 | 0.0000 | +0.0084 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.1532 | 0.1928 | -0.0396 |
| ar | 0.0242 | 0.0654 | -0.0412 |
| ain | 0.0323 | 0.0229 | +0.0094 |
| al | 0.0215 | 0.0261 | -0.0046 |
| air | 0.0188 | 0.0294 | -0.0106 |
| chol | 0.0188 | 0.0261 | -0.0073 |
| or | 0.0161 | 0.0392 | -0.0231 |
| y | 0.0134 | 0.0327 | -0.0192 |
| cheey | 0.0215 | 0.0131 | +0.0084 |
| chey | 0.0188 | 0.0131 | +0.0057 |
| ol | 0.0269 | 0.0131 | +0.0138 |
| aiiin | 0.0242 | 0.0065 | +0.0177 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ar | 0.0242 | 0.0654 | -0.0412 |
| aiin | 0.1532 | 0.1928 | -0.0396 |
| shedy | 0.0242 | 0.0000 | +0.0242 |
| or | 0.0161 | 0.0392 | -0.0231 |
| y | 0.0134 | 0.0327 | -0.0192 |
| aiiin | 0.0242 | 0.0065 | +0.0177 |
| ol | 0.0269 | 0.0131 | +0.0138 |
| o | 0.0027 | 0.0163 | -0.0137 |
| cheol | 0.0054 | 0.0163 | -0.0110 |
| chedy | 0.0108 | 0.0000 | +0.0108 |
| okain | 0.0108 | 0.0000 | +0.0108 |
| air | 0.0188 | 0.0294 | -0.0106 |

Context diagnostics: predecessor Jaccard 0.1088, JS 0.3012, entropy A/B 5.283/5.319, effective vocabulary A/B 197.02/204.16; successor Jaccard 0.1848, JS 0.5786, entropy A/B 4.575/4.103, effective vocabulary A/B 97.07/60.52.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `chedy`, `d`, `dain`, `o`, `chear`, `dar`, `sheey`, `aiin`, `chckhy`, `chor`, `qokeey`, `y`; right `cheos`, `dain`, `oiiin`, `am`, `chy`, `chor`, `cheol`, `chey`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `otaiin`, `okain`, `chdy`, `okedy`, `oteedy`, `cheor`, `shy`, `qotedy`, `chody`, `otol`, `oky`; right `qokeedy`, `qokeey`, `qokedy`, `qokain`, `qokaiin`, `qokal`, `okeey`, `otedy`, `qokar`, `otar`, `qoky`, `okar`.

## `chol` / `daiin`

Structural similarity: 0.7202; reliability: 0.9332; normalized graphemic distance: 1.0000; counts: 395/848.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9067 | 0.9873 |
| Left context | 0.7033 | 0.9046 |
| Right context | 0.5508 | 0.9077 |

- Primary component: positional agreement (0.907).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.703.
- Largest right-context difference: daiin is more frequent for chol (absolute probability difference 0.069).

Position summaries (A/B): line-start 0.0481/0.1899, line-end 0.0127/0.1545, mean 3.947/4.248, median 3.000/4.000. Position JS similarity: 0.9067.

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
| chey | 0.0154 | 0.0181 | -0.0027 |
| daiin | 0.0846 | 0.0153 | +0.0693 |
| cthy | 0.0128 | 0.0153 | -0.0025 |
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
| daiin | 0.0846 | 0.0153 | +0.0693 |
| chol | 0.0615 | 0.0098 | +0.0518 |
| shey | 0.0000 | 0.0139 | -0.0139 |
| ol | 0.0231 | 0.0098 | +0.0133 |
| chy | 0.0205 | 0.0084 | +0.0121 |
| dy | 0.0205 | 0.0084 | +0.0121 |
| chckhy | 0.0026 | 0.0126 | -0.0100 |
| dol | 0.0154 | 0.0056 | +0.0098 |
| shol | 0.0179 | 0.0084 | +0.0096 |
| chcthy | 0.0000 | 0.0084 | -0.0084 |
| chor | 0.0179 | 0.0098 | +0.0082 |
| dal | 0.0077 | 0.0153 | -0.0076 |

Context diagnostics: predecessor Jaccard 0.1782, JS 0.4580, entropy A/B 5.199/5.613, effective vocabulary A/B 181.14/273.86; successor Jaccard 0.1503, JS 0.4370, entropy A/B 4.932/5.710, effective vocabulary A/B 138.66/302.02.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `cheky`, `kchor`, `keol`, `kol`, `okol`, `otchol`, `qokeol`, `sheey`, `chody`, `shey`, `cthor`, `kchol`; right `ar`, `cheky`, `dair`, `o`, `okal`, `okeol`, `oky`, `otaiin`, `qodaiin`, `cheey`, `choky`, `sho`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qol`, `okain`, `okedy`, `oteedy`, `char`, `air`, `cheody`, `qoteedy`, `sar`, `shckhy`, `qotar`, `raiin`; right `qokaiin`, `r`, `ain`, `lchedy`, `qokey`, `qotedy`, `saiin`, `qoteedy`, `qotar`, `qotal`, `opchedy`, `qokchdy`.

## `r` / `s`

Structural similarity: 0.6989; reliability: 0.9332; normalized graphemic distance: 1.0000; counts: 169/350.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9150 | 0.9873 |
| Left context | 0.3593 | 0.9046 |
| Right context | 0.8223 | 0.9077 |

- Primary component: positional agreement (0.915).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.822.
- Largest right-context difference: ain is more frequent for r (absolute probability difference 0.059).

Position summaries (A/B): line-start 0.0414/0.0771, line-end 0.0651/0.1257, mean 7.645/5.823, median 5.000/5.000. Position JS similarity: 0.9150.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| l | 0.0309 | 0.0279 | +0.0030 |
| al | 0.0185 | 0.0186 | -0.0001 |
| ol | 0.0185 | 0.0279 | -0.0093 |
| cho | 0.0247 | 0.0124 | +0.0123 |
| cheey | 0.0123 | 0.0093 | +0.0031 |
| y | 0.0123 | 0.0093 | +0.0031 |
| d | 0.0309 | 0.0062 | +0.0247 |
| o | 0.0432 | 0.0062 | +0.0370 |
| chedy | 0.0062 | 0.0062 | -0.0000 |
| chy | 0.0062 | 0.0217 | -0.0155 |
| qokchdy | 0.0062 | 0.0062 | -0.0000 |
| qokeey | 0.0062 | 0.0093 | -0.0031 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| o | 0.0432 | 0.0062 | +0.0370 |
| cheo | 0.0309 | 0.0000 | +0.0309 |
| a | 0.0247 | 0.0000 | +0.0247 |
| t | 0.0247 | 0.0000 | +0.0247 |
| d | 0.0309 | 0.0062 | +0.0247 |
| daiin | 0.0000 | 0.0186 | -0.0186 |
| keo | 0.0185 | 0.0000 | +0.0185 |
| lo | 0.0185 | 0.0000 | +0.0185 |
| okeo | 0.0185 | 0.0000 | +0.0185 |
| chy | 0.0062 | 0.0217 | -0.0155 |
| chey | 0.0185 | 0.0031 | +0.0154 |
| chl | 0.0000 | 0.0124 | -0.0124 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.1582 | 0.1928 | -0.0346 |
| ar | 0.0380 | 0.0654 | -0.0274 |
| al | 0.0316 | 0.0261 | +0.0055 |
| ain | 0.0823 | 0.0229 | +0.0594 |
| cheey | 0.0190 | 0.0131 | +0.0059 |
| chey | 0.0253 | 0.0131 | +0.0122 |
| ol | 0.0506 | 0.0131 | +0.0376 |
| chy | 0.0127 | 0.0131 | -0.0004 |
| or | 0.0127 | 0.0392 | -0.0266 |
| aiiin | 0.0127 | 0.0065 | +0.0061 |
| shey | 0.0190 | 0.0065 | +0.0125 |
| shor | 0.0127 | 0.0065 | +0.0061 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ain | 0.0823 | 0.0229 | +0.0594 |
| ol | 0.0506 | 0.0131 | +0.0376 |
| aiin | 0.1582 | 0.1928 | -0.0346 |
| y | 0.0000 | 0.0327 | -0.0327 |
| air | 0.0000 | 0.0294 | -0.0294 |
| ar | 0.0380 | 0.0654 | -0.0274 |
| or | 0.0127 | 0.0392 | -0.0266 |
| chol | 0.0000 | 0.0261 | -0.0261 |
| @170; | 0.0253 | 0.0000 | +0.0253 |
| v | 0.0253 | 0.0000 | +0.0253 |
| char | 0.0190 | 0.0000 | +0.0190 |
| a | 0.0127 | 0.0000 | +0.0127 |

Context diagnostics: predecessor Jaccard 0.0945, JS 0.2504, entropy A/B 4.553/5.319, effective vocabulary A/B 94.89/204.16; successor Jaccard 0.1414, JS 0.5032, entropy A/B 3.871/4.103, effective vocabulary A/B 48.00/60.52.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `cheey`, `y`, `al`; right `aiiin`, `shor`, `chy`, `cheey`, `shey`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `qokain`, `qokaiin`, `okeey`, `dy`, `r`, `otedy`, `otar`, `otaiin`, `okar`, `otal`, `dair`; right `qokeedy`, `qokeey`, `qokedy`, `qokain`, `qokaiin`, `qokal`, `r`, `otedy`, `qokar`, `okal`, `chdy`, `otar`.

## `ol` / `y`

Structural similarity: 0.6788; reliability: 0.9332; normalized graphemic distance: 1.0000; counts: 560/305.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.8900 | 0.9873 |
| Left context | 0.5430 | 0.9046 |
| Right context | 0.6034 | 0.9077 |

- Primary component: positional agreement (0.890).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.603.
- Largest left-context difference: s is more frequent for y (absolute probability difference 0.036).

Position summaries (A/B): line-start 0.0554/0.2426, line-end 0.0750/0.1541, mean 5.220/5.623, median 4.000/4.000. Position JS similarity: 0.8900.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0189 | 0.0260 | -0.0071 |
| or | 0.0189 | 0.0216 | -0.0027 |
| ar | 0.0170 | 0.0303 | -0.0133 |
| qokaiin | 0.0151 | 0.0173 | -0.0022 |
| daiin | 0.0132 | 0.0303 | -0.0171 |
| okar | 0.0113 | 0.0173 | -0.0060 |
| dol | 0.0095 | 0.0130 | -0.0035 |
| okain | 0.0095 | 0.0173 | -0.0079 |
| ain | 0.0113 | 0.0087 | +0.0027 |
| chol | 0.0170 | 0.0087 | +0.0084 |
| dar | 0.0151 | 0.0087 | +0.0065 |
| ol | 0.0151 | 0.0087 | +0.0065 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0076 | 0.0433 | -0.0357 |
| qokain | 0.0246 | 0.0000 | +0.0246 |
| @171; | 0.0000 | 0.0173 | -0.0173 |
| d | 0.0000 | 0.0173 | -0.0173 |
| daiin | 0.0132 | 0.0303 | -0.0171 |
| r | 0.0151 | 0.0000 | +0.0151 |
| ar | 0.0170 | 0.0303 | -0.0133 |
| qokeed | 0.0000 | 0.0130 | -0.0130 |
| qokedy | 0.0113 | 0.0000 | +0.0113 |
| shedy | 0.0113 | 0.0000 | +0.0113 |
| cheor | 0.0095 | 0.0000 | +0.0095 |
| chey | 0.0095 | 0.0000 | +0.0095 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0463 | 0.0233 | +0.0231 |
| cheey | 0.0212 | 0.0271 | -0.0059 |
| kaiin | 0.0193 | 0.0194 | -0.0001 |
| chey | 0.0174 | 0.0194 | -0.0020 |
| daiin | 0.0174 | 0.0310 | -0.0136 |
| cheol | 0.0154 | 0.0194 | -0.0039 |
| chedy | 0.0425 | 0.0116 | +0.0308 |
| s | 0.0174 | 0.0116 | +0.0057 |
| shey | 0.0135 | 0.0116 | +0.0019 |
| aiin | 0.0270 | 0.0078 | +0.0193 |
| chor | 0.0097 | 0.0078 | +0.0019 |
| chy | 0.0097 | 0.0078 | +0.0019 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0425 | 0.0116 | +0.0308 |
| shedy | 0.0463 | 0.0233 | +0.0231 |
| aiin | 0.0270 | 0.0078 | +0.0193 |
| sheey | 0.0212 | 0.0039 | +0.0174 |
| taiin | 0.0000 | 0.0155 | -0.0155 |
| daiin | 0.0174 | 0.0310 | -0.0136 |
| dy | 0.0039 | 0.0155 | -0.0116 |
| c | 0.0000 | 0.0116 | -0.0116 |
| cheeo | 0.0000 | 0.0116 | -0.0116 |
| kal | 0.0000 | 0.0116 | -0.0116 |
| ky | 0.0000 | 0.0116 | -0.0116 |
| tchy | 0.0000 | 0.0116 | -0.0116 |

Context diagnostics: predecessor Jaccard 0.1485, JS 0.4102, entropy A/B 5.400/4.864, effective vocabulary A/B 221.44/129.58; successor Jaccard 0.1593, JS 0.4415, entropy A/B 5.098/5.040, effective vocabulary A/B 163.65/154.53.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `al`, `l`, `okaiin`, `taiin`, `qokeey`, `ain`, `dal`, `dol`, `okal`, `dar`, `ol`, `chol`; right `chedar`, `chol`, `dain`, `kor`, `r`, `raiin`, `chor`, `chy`, `kedy`, `kar`, `or`, `shey`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `shol`, `chdy`, `okedy`, `qokol`, `oty`, `cheody`, `cthy`, `qoteedy`, `qokchy`, `odaiin`, `qotchy`; right `qokedy`, `qokain`, `okar`, `oty`, `cthy`, `otain`, `shy`, `am`, `qotedy`, `qoty`, `shor`, `qotaiin`.

## `dar` / `ol`

Structural similarity: 0.6746; reliability: 0.9332; normalized graphemic distance: 1.0000; counts: 323/560.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9420 | 0.9873 |
| Left context | 0.4976 | 0.9046 |
| Right context | 0.5842 | 0.9077 |

- Primary component: positional agreement (0.942).
- Similarity is concentrated: the next component, successor-distribution overlap, is 0.584.
- Largest right-context difference: shedy is more frequent for ol (absolute probability difference 0.028).

Position summaries (A/B): line-start 0.1300/0.0554, line-end 0.1455/0.0750, mean 6.211/5.220, median 5.000/4.000. Position JS similarity: 0.9420.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokain | 0.0178 | 0.0246 | -0.0068 |
| daiin | 0.0249 | 0.0132 | +0.0117 |
| dar | 0.0107 | 0.0151 | -0.0044 |
| ol | 0.0107 | 0.0151 | -0.0044 |
| qokedy | 0.0107 | 0.0113 | -0.0007 |
| chedy | 0.0107 | 0.0095 | +0.0012 |
| chey | 0.0107 | 0.0095 | +0.0012 |
| qokal | 0.0249 | 0.0095 | +0.0155 |
| al | 0.0178 | 0.0076 | +0.0102 |
| aiin | 0.0071 | 0.0189 | -0.0118 |
| chol | 0.0071 | 0.0170 | -0.0099 |
| dain | 0.0071 | 0.0076 | -0.0004 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| or | 0.0000 | 0.0189 | -0.0189 |
| qokal | 0.0249 | 0.0095 | +0.0155 |
| r | 0.0000 | 0.0151 | -0.0151 |
| olky | 0.0142 | 0.0000 | +0.0142 |
| ar | 0.0036 | 0.0170 | -0.0135 |
| dal | 0.0178 | 0.0057 | +0.0121 |
| aiin | 0.0071 | 0.0189 | -0.0118 |
| daiin | 0.0249 | 0.0132 | +0.0117 |
| qokaiin | 0.0036 | 0.0151 | -0.0116 |
| ain | 0.0000 | 0.0113 | -0.0113 |
| oty | 0.0107 | 0.0000 | +0.0107 |
| qokeedy | 0.0107 | 0.0000 | +0.0107 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0254 | 0.0270 | -0.0017 |
| chedy | 0.0254 | 0.0425 | -0.0171 |
| shedy | 0.0181 | 0.0463 | -0.0282 |
| chey | 0.0181 | 0.0174 | +0.0007 |
| ol | 0.0290 | 0.0154 | +0.0135 |
| shey | 0.0217 | 0.0135 | +0.0082 |
| or | 0.0109 | 0.0135 | -0.0026 |
| ar | 0.0362 | 0.0097 | +0.0266 |
| chor | 0.0109 | 0.0097 | +0.0012 |
| al | 0.0217 | 0.0077 | +0.0140 |
| chdy | 0.0109 | 0.0077 | +0.0031 |
| okaiin | 0.0072 | 0.0116 | -0.0043 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0181 | 0.0463 | -0.0282 |
| ar | 0.0362 | 0.0097 | +0.0266 |
| oty | 0.0181 | 0.0000 | +0.0181 |
| cheey | 0.0036 | 0.0212 | -0.0176 |
| chedy | 0.0254 | 0.0425 | -0.0171 |
| kaiin | 0.0036 | 0.0193 | -0.0157 |
| cheol | 0.0000 | 0.0154 | -0.0154 |
| al | 0.0217 | 0.0077 | +0.0140 |
| sheey | 0.0072 | 0.0212 | -0.0140 |
| daiin | 0.0036 | 0.0174 | -0.0138 |
| ol | 0.0290 | 0.0154 | +0.0135 |
| kedy | 0.0000 | 0.0116 | -0.0116 |

Context diagnostics: predecessor Jaccard 0.1490, JS 0.3812, entropy A/B 5.174/5.400, effective vocabulary A/B 176.54/221.44; successor Jaccard 0.1615, JS 0.4332, entropy A/B 5.002/5.098, effective vocabulary A/B 148.65/163.65.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `sheey`, `tedy`, `dain`, `otaiin`, `shey`, `dol`, `qokeey`, `chedy`, `cheol`, `chey`, `okal`, `qokedy`; right `chckhy`, `dy`, `okedy`, `y`, `ain`, `chdy`, `chor`, `dar`, `okaiin`, `or`, `s`, `chey`.

Shared unobserved high-frequency contexts (descriptive absence only): left `chckhy`, `shy`, `char`, `air`, `cheody`, `cheo`, `dor`, `odaiin`, `qotchy`, `otey`, `raiin`, `cheedy`; right `qokedy`, `qokain`, `okar`, `qol`, `cthy`, `qokey`, `otain`, `shy`, `qotedy`, `qotaiin`, `saiin`, `okeol`.

## `ar` / `ol`

Structural similarity: 0.6639; reliability: 0.9332; normalized graphemic distance: 1.0000; counts: 403/560.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9386 | 0.9873 |
| Left context | 0.5758 | 0.9046 |
| Right context | 0.4773 | 0.9077 |

- Primary component: positional agreement (0.939).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.576.
- Largest right-context difference: aiin is more frequent for ar (absolute probability difference 0.046).

Position summaries (A/B): line-start 0.0099/0.0554, line-end 0.0794/0.0750, mean 6.633/5.220, median 6.000/4.000. Position JS similarity: 0.9386.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| or | 0.0226 | 0.0189 | +0.0037 |
| ar | 0.0201 | 0.0170 | +0.0030 |
| dar | 0.0251 | 0.0151 | +0.0099 |
| qokain | 0.0150 | 0.0246 | -0.0095 |
| r | 0.0150 | 0.0151 | -0.0001 |
| ol | 0.0125 | 0.0151 | -0.0026 |
| ain | 0.0150 | 0.0113 | +0.0037 |
| okar | 0.0301 | 0.0113 | +0.0187 |
| al | 0.0226 | 0.0076 | +0.0150 |
| dain | 0.0100 | 0.0076 | +0.0025 |
| s | 0.0501 | 0.0076 | +0.0426 |
| daiin | 0.0075 | 0.0132 | -0.0057 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0501 | 0.0076 | +0.0426 |
| otar | 0.0301 | 0.0057 | +0.0244 |
| okar | 0.0301 | 0.0113 | +0.0187 |
| char | 0.0150 | 0.0000 | +0.0150 |
| al | 0.0226 | 0.0076 | +0.0150 |
| aiin | 0.0050 | 0.0189 | -0.0139 |
| chol | 0.0050 | 0.0170 | -0.0120 |
| otain | 0.0175 | 0.0057 | +0.0119 |
| dar | 0.0251 | 0.0151 | +0.0099 |
| qokain | 0.0150 | 0.0246 | -0.0095 |
| qokal | 0.0000 | 0.0095 | -0.0095 |
| qokedy | 0.0025 | 0.0113 | -0.0088 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0728 | 0.0270 | +0.0457 |
| ol | 0.0243 | 0.0154 | +0.0088 |
| or | 0.0270 | 0.0135 | +0.0134 |
| cheey | 0.0108 | 0.0212 | -0.0105 |
| chey | 0.0108 | 0.0174 | -0.0066 |
| ar | 0.0216 | 0.0097 | +0.0119 |
| chedy | 0.0081 | 0.0425 | -0.0344 |
| daiin | 0.0081 | 0.0174 | -0.0093 |
| shedy | 0.0081 | 0.0463 | -0.0382 |
| shey | 0.0081 | 0.0135 | -0.0054 |
| al | 0.0485 | 0.0077 | +0.0408 |
| cheor | 0.0081 | 0.0058 | +0.0023 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0728 | 0.0270 | +0.0457 |
| al | 0.0485 | 0.0077 | +0.0408 |
| shedy | 0.0081 | 0.0463 | -0.0382 |
| chedy | 0.0081 | 0.0425 | -0.0344 |
| am | 0.0189 | 0.0000 | +0.0189 |
| kaiin | 0.0027 | 0.0193 | -0.0166 |
| aiiin | 0.0162 | 0.0000 | +0.0162 |
| sheey | 0.0054 | 0.0212 | -0.0158 |
| ain | 0.0189 | 0.0039 | +0.0150 |
| air | 0.0189 | 0.0039 | +0.0150 |
| y | 0.0189 | 0.0039 | +0.0150 |
| s | 0.0027 | 0.0174 | -0.0147 |

Context diagnostics: predecessor Jaccard 0.1571, JS 0.4509, entropy A/B 5.116/5.400, effective vocabulary A/B 166.71/221.44; successor Jaccard 0.1538, JS 0.4114, entropy A/B 4.969/5.098, effective vocabulary A/B 143.92/163.65.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `chees`, `dal`, `okal`, `saiin`, `ches`, `kar`, `lor`, `okaiin`, `okain`, `qokar`, `sain`, `chor`; right `chcthy`, `chedar`, `o`, `okedy`, `okol`, `cheor`, `chor`, `keey`, `okaiin`, `shey`, `chey`, `daiin`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `chdy`, `okedy`, `qokol`, `shy`, `chody`, `cthy`, `cheo`, `qoteedy`, `qokchy`, `dor`, `odaiin`; right `qokedy`, `qokain`, `otal`, `qol`, `oty`, `cthy`, `qokey`, `otain`, `qotedy`, `qoty`, `shor`, `saiin`.

## `chor` / `daiin`

Structural similarity: 0.6627; reliability: 0.9332; normalized graphemic distance: 1.0000; counts: 211/848.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.8914 | 0.9873 |
| Left context | 0.5748 | 0.9046 |
| Right context | 0.5218 | 0.9077 |

- Primary component: positional agreement (0.891).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.575.
- Largest right-context difference: chol is more frequent for chor (absolute probability difference 0.029).

Position summaries (A/B): line-start 0.0569/0.1899, line-end 0.0332/0.1545, mean 3.318/4.248, median 2.000/4.000. Position JS similarity: 0.8914.

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
| cthy | 0.0196 | 0.0153 | +0.0043 |
| daiin | 0.0441 | 0.0153 | +0.0288 |
| cthor | 0.0147 | 0.0112 | +0.0035 |
| or | 0.0147 | 0.0112 | +0.0035 |
| chckhy | 0.0098 | 0.0126 | -0.0027 |
| chey | 0.0098 | 0.0181 | -0.0083 |
| chol | 0.0392 | 0.0098 | +0.0295 |
| chor | 0.0343 | 0.0098 | +0.0246 |
| cthol | 0.0147 | 0.0098 | +0.0049 |
| ol | 0.0098 | 0.0098 | +0.0000 |
| chy | 0.0245 | 0.0084 | +0.0161 |
| dy | 0.0098 | 0.0084 | +0.0014 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chol | 0.0392 | 0.0098 | +0.0295 |
| daiin | 0.0441 | 0.0153 | +0.0288 |
| chor | 0.0343 | 0.0098 | +0.0246 |
| aiin | 0.0245 | 0.0028 | +0.0217 |
| cheky | 0.0196 | 0.0028 | +0.0168 |
| chy | 0.0245 | 0.0084 | +0.0161 |
| ar | 0.0196 | 0.0042 | +0.0154 |
| kar | 0.0147 | 0.0000 | +0.0147 |
| shey | 0.0000 | 0.0139 | -0.0139 |
| sheey | 0.0147 | 0.0014 | +0.0133 |
| dal | 0.0049 | 0.0153 | -0.0104 |
| chal | 0.0098 | 0.0000 | +0.0098 |

Context diagnostics: predecessor Jaccard 0.1242, JS 0.3757, entropy A/B 4.835/5.613, effective vocabulary A/B 125.79/273.86; successor Jaccard 0.1175, JS 0.3727, entropy A/B 4.734/5.710, effective vocabulary A/B 113.74/302.02.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `ar`, `cheol`, `oky`, `shor`, `cthy`, `qokol`, `y`, `chy`, `dal`; right `al`, `dy`, `ol`, `chckhy`, `cthol`, `cthor`, `or`, `s`, `chey`, `ar`, `cheky`, `cthy`.

Shared unobserved high-frequency contexts (descriptive absence only): left `shedy`, `okal`, `qol`, `okar`, `okain`, `okedy`, `ain`, `lchedy`, `oteedy`, `air`, `cheody`, `qoteedy`; right `qokeedy`, `qokain`, `qokaiin`, `l`, `r`, `ain`, `lchedy`, `qokey`, `qotedy`, `air`, `kaiin`, `saiin`.

## `chey` / `ol`

Structural similarity: 0.6529; reliability: 0.9332; normalized graphemic distance: 1.0000; counts: 351/560.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9507 | 0.9873 |
| Left context | 0.6984 | 0.9046 |
| Right context | 0.3096 | 0.9077 |

- Primary component: positional agreement (0.951).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.698.
- Largest right-context difference: shedy is more frequent for ol (absolute probability difference 0.046).

Position summaries (A/B): line-start 0.0171/0.0554, line-end 0.0541/0.0750, mean 4.880/5.220, median 3.000/4.000. Position JS similarity: 0.9507.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokain | 0.0203 | 0.0246 | -0.0043 |
| aiin | 0.0290 | 0.0189 | +0.0101 |
| or | 0.0203 | 0.0189 | +0.0014 |
| chol | 0.0174 | 0.0170 | +0.0004 |
| ol | 0.0261 | 0.0151 | +0.0110 |
| qokaiin | 0.0174 | 0.0151 | +0.0023 |
| dar | 0.0145 | 0.0151 | -0.0006 |
| daiin | 0.0377 | 0.0132 | +0.0244 |
| ar | 0.0116 | 0.0170 | -0.0054 |
| r | 0.0116 | 0.0151 | -0.0035 |
| shedy | 0.0116 | 0.0113 | +0.0003 |
| cheor | 0.0145 | 0.0095 | +0.0050 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0377 | 0.0132 | +0.0244 |
| dain | 0.0232 | 0.0076 | +0.0156 |
| okar | 0.0000 | 0.0113 | -0.0113 |
| ol | 0.0261 | 0.0151 | +0.0110 |
| okain | 0.0203 | 0.0095 | +0.0108 |
| y | 0.0145 | 0.0038 | +0.0107 |
| aiin | 0.0290 | 0.0189 | +0.0101 |
| sol | 0.0116 | 0.0019 | +0.0097 |
| dol | 0.0000 | 0.0095 | -0.0095 |
| raiin | 0.0087 | 0.0000 | +0.0087 |
| t | 0.0087 | 0.0000 | +0.0087 |
| qokedy | 0.0029 | 0.0113 | -0.0084 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0241 | 0.0174 | +0.0067 |
| ol | 0.0151 | 0.0154 | -0.0004 |
| keey | 0.0181 | 0.0116 | +0.0065 |
| okaiin | 0.0151 | 0.0116 | +0.0035 |
| kain | 0.0181 | 0.0097 | +0.0084 |
| keedy | 0.0090 | 0.0116 | -0.0025 |
| dain | 0.0120 | 0.0077 | +0.0043 |
| dal | 0.0120 | 0.0077 | +0.0043 |
| chey | 0.0060 | 0.0174 | -0.0114 |
| or | 0.0060 | 0.0135 | -0.0075 |
| dar | 0.0090 | 0.0058 | +0.0032 |
| l | 0.0120 | 0.0058 | +0.0063 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0000 | 0.0463 | -0.0463 |
| chedy | 0.0000 | 0.0425 | -0.0425 |
| qokain | 0.0241 | 0.0000 | +0.0241 |
| aiin | 0.0030 | 0.0270 | -0.0240 |
| cheey | 0.0000 | 0.0212 | -0.0212 |
| sheey | 0.0000 | 0.0212 | -0.0212 |
| qol | 0.0211 | 0.0000 | +0.0211 |
| qokeey | 0.0211 | 0.0019 | +0.0192 |
| kaiin | 0.0030 | 0.0193 | -0.0163 |
| qokaiin | 0.0181 | 0.0019 | +0.0161 |
| qokeedy | 0.0181 | 0.0019 | +0.0161 |
| cheol | 0.0000 | 0.0154 | -0.0154 |

Context diagnostics: predecessor Jaccard 0.1803, JS 0.4977, entropy A/B 5.003/5.400, effective vocabulary A/B 148.80/221.44; successor Jaccard 0.1582, JS 0.3412, entropy A/B 5.146/5.098, effective vocabulary A/B 171.74/163.65.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `chor`, `dal`, `kar`, `l`, `lkaiin`, `okal`, `sheor`, `okaiin`, `otain`, `chey`, `qokal`, `qokar`; right `chol`, `dar`, `kor`, `qokar`, `r`, `keedy`, `dain`, `dal`, `l`, `lchedy`, `qoky`, `or`.

Shared unobserved high-frequency contexts (descriptive absence only): left `o`, `chdy`, `okedy`, `oty`, `char`, `air`, `cheody`, `cheo`, `qoteedy`, `qokchy`, `dor`, `qotchy`; right `oty`, `shy`, `am`, `shor`, `okeol`, `okey`, `odaiin`, `d`, `qotain`, `qotal`, `cthol`, `qokchy`.

## `lchedy` / `qokar`

Structural similarity: 0.6585; reliability: 0.9065; normalized graphemic distance: 1.0000; counts: 116/159.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9736 | 0.9811 |
| Left context | 0.5639 | 0.8669 |
| Right context | 0.4379 | 0.8714 |

- Primary component: positional agreement (0.974).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.564.
- Largest right-context difference: chedy is more frequent for lchedy (absolute probability difference 0.035).

Position summaries (A/B): line-start 0.0517/0.0440, line-end 0.1466/0.0252, mean 4.543/4.868, median 4.000/4.000. Position JS similarity: 0.9736.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0545 | 0.0658 | -0.0112 |
| qokeedy | 0.0455 | 0.0395 | +0.0060 |
| chdy | 0.0273 | 0.0197 | +0.0075 |
| chey | 0.0364 | 0.0197 | +0.0166 |
| dal | 0.0273 | 0.0132 | +0.0141 |
| okeey | 0.0182 | 0.0132 | +0.0050 |
| ol | 0.0273 | 0.0132 | +0.0141 |
| qokedy | 0.0273 | 0.0132 | +0.0141 |
| qokeey | 0.0364 | 0.0132 | +0.0232 |
| cheedy | 0.0091 | 0.0132 | -0.0041 |
| dy | 0.0091 | 0.0132 | -0.0041 |
| qoty | 0.0091 | 0.0132 | -0.0041 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| okedy | 0.0273 | 0.0000 | +0.0273 |
| oteedy | 0.0273 | 0.0000 | +0.0273 |
| shedy | 0.0000 | 0.0263 | -0.0263 |
| shey | 0.0091 | 0.0329 | -0.0238 |
| qokeey | 0.0364 | 0.0132 | +0.0232 |
| qokaiin | 0.0000 | 0.0197 | -0.0197 |
| al | 0.0182 | 0.0000 | +0.0182 |
| cheey | 0.0182 | 0.0000 | +0.0182 |
| lkedy | 0.0182 | 0.0000 | +0.0182 |
| qotchedy | 0.0182 | 0.0000 | +0.0182 |
| chey | 0.0364 | 0.0197 | +0.0166 |
| dal | 0.0273 | 0.0132 | +0.0141 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0404 | 0.0710 | -0.0306 |
| chckhy | 0.0303 | 0.0258 | +0.0045 |
| chedy | 0.0606 | 0.0258 | +0.0348 |
| qokain | 0.0202 | 0.0194 | +0.0008 |
| qokey | 0.0303 | 0.0129 | +0.0174 |
| chey | 0.0101 | 0.0194 | -0.0093 |
| okaiin | 0.0101 | 0.0129 | -0.0028 |
| okar | 0.0101 | 0.0387 | -0.0286 |
| ol | 0.0101 | 0.0323 | -0.0222 |
| qokal | 0.0101 | 0.0194 | -0.0093 |
| chol | 0.0101 | 0.0065 | +0.0036 |
| lchedy | 0.0101 | 0.0065 | +0.0036 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0606 | 0.0258 | +0.0348 |
| shedy | 0.0404 | 0.0710 | -0.0306 |
| qokaiin | 0.0303 | 0.0000 | +0.0303 |
| qokedy | 0.0303 | 0.0000 | +0.0303 |
| qokeedy | 0.0303 | 0.0000 | +0.0303 |
| okar | 0.0101 | 0.0387 | -0.0286 |
| qokeey | 0.0303 | 0.0065 | +0.0239 |
| ol | 0.0101 | 0.0323 | -0.0222 |
| lar | 0.0202 | 0.0000 | +0.0202 |
| lkaiin | 0.0202 | 0.0000 | +0.0202 |
| lkchedy | 0.0202 | 0.0000 | +0.0202 |
| ar | 0.0000 | 0.0194 | -0.0194 |

Context diagnostics: predecessor Jaccard 0.1118, JS 0.3435, entropy A/B 4.186/4.520, effective vocabulary A/B 65.76/91.79; successor Jaccard 0.1154, JS 0.3016, entropy A/B 4.161/4.385, effective vocabulary A/B 64.12/80.25.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `okeey`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `aiin`, `chol`, `or`, `ar`, `s`, `dar`, `qokain`, `chor`, `dain`, `l`, `r`, `chy`; right `daiin`, `aiin`, `dain`, `o`, `qokar`, `sheey`, `okain`, `oteey`, `qol`, `okedy`, `cthy`, `dol`.

## `lchedy` / `qol`

Structural similarity: 0.6556; reliability: 0.8977; normalized graphemic distance: 1.0000; counts: 116/142.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9663 | 0.9790 |
| Left context | 0.5352 | 0.8546 |
| Right context | 0.4654 | 0.8595 |

- Primary component: positional agreement (0.966).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.535.
- Largest left-context difference: shedy is more frequent for qol (absolute probability difference 0.064).

Position summaries (A/B): line-start 0.0517/0.1197, line-end 0.1466/0.0634, mean 4.543/3.915, median 4.000/4.000. Position JS similarity: 0.9663.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0545 | 0.1040 | -0.0495 |
| chey | 0.0364 | 0.0560 | -0.0196 |
| qokeedy | 0.0455 | 0.0320 | +0.0135 |
| qokedy | 0.0273 | 0.0240 | +0.0033 |
| cheey | 0.0182 | 0.0240 | -0.0058 |
| dal | 0.0273 | 0.0160 | +0.0113 |
| okeey | 0.0182 | 0.0160 | +0.0022 |
| oteey | 0.0091 | 0.0240 | -0.0149 |
| qoky | 0.0091 | 0.0160 | -0.0069 |
| sheedy | 0.0091 | 0.0240 | -0.0149 |
| shey | 0.0091 | 0.0160 | -0.0069 |
| y | 0.0091 | 0.0160 | -0.0069 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0000 | 0.0640 | -0.0640 |
| chedy | 0.0545 | 0.1040 | -0.0495 |
| qokeey | 0.0364 | 0.0000 | +0.0364 |
| qol | 0.0000 | 0.0320 | -0.0320 |
| chdy | 0.0273 | 0.0000 | +0.0273 |
| okedy | 0.0273 | 0.0000 | +0.0273 |
| ol | 0.0273 | 0.0000 | +0.0273 |
| oteedy | 0.0273 | 0.0000 | +0.0273 |
| sheey | 0.0000 | 0.0240 | -0.0240 |
| chey | 0.0364 | 0.0560 | -0.0196 |
| al | 0.0182 | 0.0000 | +0.0182 |
| lkedy | 0.0182 | 0.0000 | +0.0182 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0606 | 0.1128 | -0.0522 |
| shedy | 0.0404 | 0.0752 | -0.0348 |
| cheedy | 0.0101 | 0.0301 | -0.0200 |
| chey | 0.0101 | 0.0376 | -0.0275 |
| okaiin | 0.0101 | 0.0150 | -0.0049 |
| ol | 0.0101 | 0.0301 | -0.0200 |
| l | 0.0101 | 0.0075 | +0.0026 |
| okeey | 0.0101 | 0.0075 | +0.0026 |
| qokal | 0.0101 | 0.0075 | +0.0026 |
| r | 0.0101 | 0.0075 | +0.0026 |
| raiin | 0.0101 | 0.0075 | +0.0026 |
| rchedy | 0.0101 | 0.0075 | +0.0026 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0606 | 0.1128 | -0.0522 |
| sheedy | 0.0000 | 0.0451 | -0.0451 |
| shedy | 0.0404 | 0.0752 | -0.0348 |
| chckhy | 0.0303 | 0.0000 | +0.0303 |
| qokaiin | 0.0303 | 0.0000 | +0.0303 |
| qokedy | 0.0303 | 0.0000 | +0.0303 |
| qokeedy | 0.0303 | 0.0000 | +0.0303 |
| qokeey | 0.0303 | 0.0000 | +0.0303 |
| qokey | 0.0303 | 0.0000 | +0.0303 |
| cheey | 0.0000 | 0.0301 | -0.0301 |
| qol | 0.0000 | 0.0301 | -0.0301 |
| chey | 0.0101 | 0.0376 | -0.0275 |

Context diagnostics: predecessor Jaccard 0.1579, JS 0.3804, entropy A/B 4.186/4.010, effective vocabulary A/B 65.76/55.16; successor Jaccard 0.0876, JS 0.2496, entropy A/B 4.161/3.950, effective vocabulary A/B 64.12/51.93.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `okeey`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `aiin`, `or`, `ar`, `s`, `dar`, `qokain`, `qokaiin`, `chor`, `okaiin`, `dain`, `shol`, `l`; right `daiin`, `ar`, `al`, `dal`, `chor`, `dain`, `o`, `qokar`, `otaiin`, `okal`, `okain`, `otal`.

## `okaiin` / `ol`

Structural similarity: 0.6885; reliability: 0.9332; normalized graphemic distance: 0.8333; counts: 216/560.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9552 | 0.9873 |
| Left context | 0.6039 | 0.9046 |
| Right context | 0.5065 | 0.9077 |

- Primary component: positional agreement (0.955).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.604.
- Largest right-context difference: otaiin is more frequent for okaiin (absolute probability difference 0.033).

Position summaries (A/B): line-start 0.0880/0.0554, line-end 0.0833/0.0750, mean 4.245/5.220, median 4.000/4.000. Position JS similarity: 0.9552.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokain | 0.0203 | 0.0246 | -0.0043 |
| aiin | 0.0406 | 0.0189 | +0.0217 |
| or | 0.0152 | 0.0189 | -0.0037 |
| ol | 0.0305 | 0.0151 | +0.0153 |
| daiin | 0.0406 | 0.0132 | +0.0274 |
| ain | 0.0102 | 0.0113 | -0.0012 |
| ar | 0.0102 | 0.0170 | -0.0069 |
| chol | 0.0102 | 0.0170 | -0.0069 |
| dar | 0.0102 | 0.0151 | -0.0050 |
| chey | 0.0254 | 0.0095 | +0.0159 |
| okain | 0.0102 | 0.0095 | +0.0007 |
| qokar | 0.0102 | 0.0095 | +0.0007 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0406 | 0.0132 | +0.0274 |
| okaiin | 0.0305 | 0.0076 | +0.0229 |
| aiin | 0.0406 | 0.0189 | +0.0217 |
| chey | 0.0254 | 0.0095 | +0.0159 |
| ol | 0.0305 | 0.0151 | +0.0153 |
| chckhy | 0.0152 | 0.0000 | +0.0152 |
| air | 0.0102 | 0.0000 | +0.0102 |
| cheody | 0.0102 | 0.0000 | +0.0102 |
| okeor | 0.0102 | 0.0000 | +0.0102 |
| otey | 0.0102 | 0.0000 | +0.0102 |
| qokeeody | 0.0102 | 0.0000 | +0.0102 |
| qokaiin | 0.0051 | 0.0151 | -0.0100 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0303 | 0.0463 | -0.0160 |
| chedy | 0.0253 | 0.0425 | -0.0172 |
| ol | 0.0202 | 0.0154 | +0.0048 |
| cheey | 0.0152 | 0.0212 | -0.0061 |
| daiin | 0.0152 | 0.0174 | -0.0022 |
| okaiin | 0.0303 | 0.0116 | +0.0187 |
| chey | 0.0101 | 0.0174 | -0.0073 |
| or | 0.0101 | 0.0135 | -0.0034 |
| sheey | 0.0101 | 0.0212 | -0.0111 |
| ar | 0.0101 | 0.0097 | +0.0004 |
| al | 0.0152 | 0.0077 | +0.0074 |
| chckhy | 0.0303 | 0.0058 | +0.0245 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| otaiin | 0.0354 | 0.0019 | +0.0334 |
| aiin | 0.0000 | 0.0270 | -0.0270 |
| chckhy | 0.0303 | 0.0058 | +0.0245 |
| kaiin | 0.0000 | 0.0193 | -0.0193 |
| okaiin | 0.0303 | 0.0116 | +0.0187 |
| chedy | 0.0253 | 0.0425 | -0.0172 |
| shedy | 0.0303 | 0.0463 | -0.0160 |
| cheol | 0.0000 | 0.0154 | -0.0154 |
| cthy | 0.0152 | 0.0000 | +0.0152 |
| shckhy | 0.0152 | 0.0000 | +0.0152 |
| cheody | 0.0152 | 0.0019 | +0.0132 |
| okal | 0.0152 | 0.0019 | +0.0132 |

Context diagnostics: predecessor Jaccard 0.1528, JS 0.4392, entropy A/B 4.759/5.400, effective vocabulary A/B 116.67/221.44; successor Jaccard 0.1478, JS 0.3815, entropy A/B 4.733/5.098, effective vocabulary A/B 113.65/163.65.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `dain`, `okain`, `okal`, `okol`, `otal`, `qokar`, `qol`, `saiin`, `taiin`, `ain`, `dar`, `shey`; right `ar`, `oteedy`, `y`, `or`, `al`, `chol`, `chey`, `daiin`.

Shared unobserved high-frequency contexts (descriptive absence only): left `o`, `shol`, `oty`, `shy`, `cthy`, `cheo`, `qokchy`, `dor`, `qotchy`, `raiin`, `cheedy`, `ykeey`; right `qokedy`, `otal`, `qol`, `oty`, `okeedy`, `qokey`, `am`, `qotedy`, `qoty`, `shor`, `qotaiin`, `saiin`.

## `okain` / `ol`

Structural similarity: 0.6828; reliability: 0.9230; normalized graphemic distance: 0.8000; counts: 141/560.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9482 | 0.9849 |
| Left context | 0.5725 | 0.8902 |
| Right context | 0.5276 | 0.8938 |

- Primary component: positional agreement (0.948).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.572.
- Largest right-context difference: chey is more frequent for okain (absolute probability difference 0.037).

Position summaries (A/B): line-start 0.0851/0.0554, line-end 0.0922/0.0750, mean 4.617/5.220, median 4.000/4.000. Position JS similarity: 0.9482.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokain | 0.0388 | 0.0246 | +0.0142 |
| aiin | 0.0465 | 0.0189 | +0.0276 |
| or | 0.0310 | 0.0189 | +0.0121 |
| ar | 0.0310 | 0.0170 | +0.0140 |
| qokaiin | 0.0388 | 0.0151 | +0.0236 |
| qokedy | 0.0310 | 0.0113 | +0.0197 |
| shedy | 0.0155 | 0.0113 | +0.0042 |
| chedy | 0.0155 | 0.0095 | +0.0061 |
| chey | 0.0155 | 0.0095 | +0.0061 |
| qokeey | 0.0310 | 0.0095 | +0.0216 |
| chol | 0.0078 | 0.0170 | -0.0093 |
| okain | 0.0078 | 0.0095 | -0.0017 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0465 | 0.0189 | +0.0276 |
| qokaiin | 0.0388 | 0.0151 | +0.0236 |
| qokeey | 0.0310 | 0.0095 | +0.0216 |
| qokedy | 0.0310 | 0.0113 | +0.0197 |
| char | 0.0155 | 0.0000 | +0.0155 |
| dar | 0.0000 | 0.0151 | -0.0151 |
| r | 0.0000 | 0.0151 | -0.0151 |
| qokain | 0.0388 | 0.0246 | +0.0142 |
| ar | 0.0310 | 0.0170 | +0.0140 |
| daiin | 0.0000 | 0.0132 | -0.0132 |
| or | 0.0310 | 0.0189 | +0.0121 |
| ain | 0.0000 | 0.0113 | -0.0113 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0391 | 0.0425 | -0.0034 |
| shedy | 0.0234 | 0.0463 | -0.0229 |
| chey | 0.0547 | 0.0174 | +0.0373 |
| cheey | 0.0156 | 0.0212 | -0.0056 |
| ol | 0.0391 | 0.0154 | +0.0236 |
| shey | 0.0234 | 0.0135 | +0.0099 |
| okaiin | 0.0156 | 0.0116 | +0.0040 |
| ar | 0.0234 | 0.0097 | +0.0138 |
| aiin | 0.0078 | 0.0270 | -0.0192 |
| cheedy | 0.0078 | 0.0116 | -0.0038 |
| cheol | 0.0078 | 0.0154 | -0.0076 |
| sheey | 0.0078 | 0.0212 | -0.0134 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chey | 0.0547 | 0.0174 | +0.0373 |
| y | 0.0312 | 0.0039 | +0.0274 |
| ol | 0.0391 | 0.0154 | +0.0236 |
| qokain | 0.0234 | 0.0000 | +0.0234 |
| shedy | 0.0234 | 0.0463 | -0.0229 |
| kaiin | 0.0000 | 0.0193 | -0.0193 |
| aiin | 0.0078 | 0.0270 | -0.0192 |
| daiin | 0.0000 | 0.0174 | -0.0174 |
| s | 0.0000 | 0.0174 | -0.0174 |
| chear | 0.0156 | 0.0000 | +0.0156 |
| okar | 0.0156 | 0.0000 | +0.0156 |
| ar | 0.0234 | 0.0097 | +0.0138 |

Context diagnostics: predecessor Jaccard 0.0984, JS 0.3541, entropy A/B 4.426/5.400, effective vocabulary A/B 83.60/221.44; successor Jaccard 0.1355, JS 0.3907, entropy A/B 4.365/5.098, effective vocabulary A/B 78.67/163.65.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `chedy`, `chey`, `otain`, `shedy`; right `chcthy`, `chdy`, `okaiin`.

Shared unobserved high-frequency contexts (descriptive absence only): left `o`, `shol`, `okedy`, `qokol`, `oty`, `shy`, `chody`, `air`, `cheody`, `cthy`, `cheo`, `qoteedy`; right `qokedy`, `otal`, `qol`, `oty`, `okeedy`, `cthy`, `qokey`, `otain`, `shy`, `qotedy`, `qoty`, `shor`.

## `qokaiin` / `qol`

Structural similarity: 0.7398; reliability: 0.9236; normalized graphemic distance: 0.7143; counts: 265/142.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9756 | 0.9851 |
| Left context | 0.6875 | 0.8910 |
| Right context | 0.5563 | 0.8946 |

- Primary component: positional agreement (0.976).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.688.
- Largest right-context difference: chedy is more frequent for qol (absolute probability difference 0.077).

Position summaries (A/B): line-start 0.1094/0.1197, line-end 0.0453/0.0634, mean 3.921/3.915, median 4.000/4.000. Position JS similarity: 0.9756.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0720 | 0.0640 | +0.0080 |
| chedy | 0.0424 | 0.1040 | -0.0616 |
| chey | 0.0254 | 0.0560 | -0.0306 |
| qokeedy | 0.0212 | 0.0320 | -0.0108 |
| sheedy | 0.0169 | 0.0240 | -0.0071 |
| sheey | 0.0169 | 0.0240 | -0.0071 |
| shey | 0.0339 | 0.0160 | +0.0179 |
| cheey | 0.0127 | 0.0240 | -0.0113 |
| oteey | 0.0085 | 0.0240 | -0.0155 |
| qokedy | 0.0085 | 0.0240 | -0.0155 |
| chckhy | 0.0085 | 0.0080 | +0.0005 |
| olchedy | 0.0085 | 0.0080 | +0.0005 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0424 | 0.1040 | -0.0616 |
| qol | 0.0000 | 0.0320 | -0.0320 |
| chey | 0.0254 | 0.0560 | -0.0306 |
| shey | 0.0339 | 0.0160 | +0.0179 |
| qokeey | 0.0169 | 0.0000 | +0.0169 |
| otedy | 0.0000 | 0.0160 | -0.0160 |
| qokey | 0.0000 | 0.0160 | -0.0160 |
| qoky | 0.0000 | 0.0160 | -0.0160 |
| sshey | 0.0000 | 0.0160 | -0.0160 |
| oteey | 0.0085 | 0.0240 | -0.0155 |
| qokedy | 0.0085 | 0.0240 | -0.0155 |
| aiiin | 0.0127 | 0.0000 | +0.0127 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0356 | 0.1128 | -0.0772 |
| shedy | 0.0356 | 0.0752 | -0.0396 |
| ol | 0.0316 | 0.0301 | +0.0015 |
| chey | 0.0237 | 0.0376 | -0.0139 |
| shey | 0.0158 | 0.0226 | -0.0067 |
| or | 0.0158 | 0.0150 | +0.0008 |
| cheey | 0.0079 | 0.0301 | -0.0222 |
| cheol | 0.0079 | 0.0150 | -0.0071 |
| otain | 0.0079 | 0.0150 | -0.0071 |
| chcthy | 0.0079 | 0.0075 | +0.0004 |
| chdy | 0.0119 | 0.0075 | +0.0043 |
| dy | 0.0079 | 0.0075 | +0.0004 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0356 | 0.1128 | -0.0772 |
| sheedy | 0.0000 | 0.0451 | -0.0451 |
| shedy | 0.0356 | 0.0752 | -0.0396 |
| qol | 0.0000 | 0.0301 | -0.0301 |
| cheedy | 0.0040 | 0.0301 | -0.0261 |
| sheey | 0.0000 | 0.0226 | -0.0226 |
| cheey | 0.0079 | 0.0301 | -0.0222 |
| okain | 0.0198 | 0.0000 | +0.0198 |
| okal | 0.0198 | 0.0000 | +0.0198 |
| shckhy | 0.0198 | 0.0000 | +0.0198 |
| checkhy | 0.0158 | 0.0000 | +0.0158 |
| chol | 0.0158 | 0.0000 | +0.0158 |

Context diagnostics: predecessor Jaccard 0.1221, JS 0.4018, entropy A/B 4.800/4.010, effective vocabulary A/B 121.48/55.16; successor Jaccard 0.1043, JS 0.3369, entropy A/B 4.821/3.950, effective vocabulary A/B 124.12/51.93.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left ``; right `cheol`, `otain`, `or`.

Shared unobserved high-frequency contexts (descriptive absence only): left `or`, `s`, `qokain`, `al`, `chor`, `dain`, `l`, `cheol`, `r`, `qokar`, `otar`, `otaiin`; right `daiin`, `s`, `qokeedy`, `qokeey`, `qokain`, `o`, `qoky`, `oteey`, `okedy`, `lchedy`, `cthy`, `qokey`.

## `daiin` / `dol`

Structural similarity: 0.6672; reliability: 0.9018; normalized graphemic distance: 0.8000; counts: 848/109.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9436 | 0.9801 |
| Left context | 0.6111 | 0.8602 |
| Right context | 0.4468 | 0.8650 |

- Primary component: positional agreement (0.944).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.611.
- Largest right-context difference: ol is more frequent for dol (absolute probability difference 0.039).

Position summaries (A/B): line-start 0.1899/0.1284, line-end 0.1545/0.0550, mean 4.248/5.018, median 4.000/4.000. Position JS similarity: 0.9436.

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
| daiin | 0.0153 | 0.0194 | -0.0041 |
| shey | 0.0139 | 0.0194 | -0.0055 |
| dain | 0.0126 | 0.0194 | -0.0069 |
| or | 0.0112 | 0.0291 | -0.0180 |
| dar | 0.0098 | 0.0194 | -0.0097 |
| ol | 0.0098 | 0.0485 | -0.0388 |
| y | 0.0098 | 0.0291 | -0.0194 |
| chckhy | 0.0126 | 0.0097 | +0.0028 |
| chedy | 0.0139 | 0.0097 | +0.0042 |
| dal | 0.0153 | 0.0097 | +0.0056 |
| chy | 0.0084 | 0.0097 | -0.0013 |
| dy | 0.0084 | 0.0097 | -0.0013 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ol | 0.0098 | 0.0485 | -0.0388 |
| shedy | 0.0070 | 0.0388 | -0.0319 |
| y | 0.0098 | 0.0291 | -0.0194 |
| chey | 0.0181 | 0.0000 | +0.0181 |
| checthy | 0.0014 | 0.0194 | -0.0180 |
| oty | 0.0014 | 0.0194 | -0.0180 |
| sheey | 0.0014 | 0.0194 | -0.0180 |
| or | 0.0112 | 0.0291 | -0.0180 |
| cthy | 0.0153 | 0.0000 | +0.0153 |
| dair | 0.0042 | 0.0194 | -0.0152 |
| cheol | 0.0056 | 0.0194 | -0.0138 |
| cthor | 0.0112 | 0.0000 | +0.0112 |

Context diagnostics: predecessor Jaccard 0.0909, JS 0.3593, entropy A/B 5.613/4.225, effective vocabulary A/B 273.86/68.40; successor Jaccard 0.0891, JS 0.3231, entropy A/B 5.710/4.304, effective vocabulary A/B 302.02/74.01.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left ``; right `cheol`, `daiin`, `dain`, `dair`, `dar`, `shey`, `shol`.

Shared unobserved high-frequency contexts (descriptive absence only): left `shedy`, `qokaiin`, `qokar`, `okal`, `qol`, `okar`, `okain`, `ain`, `lchedy`, `oteedy`, `cheor`, `char`; right `qokeedy`, `qokain`, `qokaiin`, `l`, `r`, `okain`, `ain`, `lchedy`, `qokey`, `qotedy`, `air`, `kaiin`.

## `aiin` / `ar`

Structural similarity: 0.6715; reliability: 0.9332; normalized graphemic distance: 0.7500; counts: 504/403.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9577 | 0.9873 |
| Left context | 0.6439 | 0.9046 |
| Right context | 0.4128 | 0.9077 |

- Primary component: positional agreement (0.958).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.644.
- Largest left-context difference: or is more frequent for aiin (absolute probability difference 0.091).

Position summaries (A/B): line-start 0.0000/0.0099, line-end 0.0813/0.0794, mean 6.583/6.633, median 6.000/6.000. Position JS similarity: 0.9577.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.1171 | 0.0501 | +0.0669 |
| or | 0.1131 | 0.0226 | +0.0905 |
| ar | 0.0536 | 0.0201 | +0.0335 |
| r | 0.0496 | 0.0150 | +0.0346 |
| dar | 0.0139 | 0.0251 | -0.0112 |
| ol | 0.0278 | 0.0125 | +0.0152 |
| char | 0.0099 | 0.0150 | -0.0051 |
| chor | 0.0099 | 0.0100 | -0.0001 |
| otar | 0.0099 | 0.0301 | -0.0202 |
| ches | 0.0139 | 0.0075 | +0.0064 |
| lor | 0.0099 | 0.0075 | +0.0024 |
| o | 0.0139 | 0.0075 | +0.0064 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| or | 0.1131 | 0.0226 | +0.0905 |
| s | 0.1171 | 0.0501 | +0.0669 |
| r | 0.0496 | 0.0150 | +0.0346 |
| ar | 0.0536 | 0.0201 | +0.0335 |
| okar | 0.0000 | 0.0301 | -0.0301 |
| al | 0.0020 | 0.0226 | -0.0206 |
| otar | 0.0099 | 0.0301 | -0.0202 |
| otain | 0.0000 | 0.0175 | -0.0175 |
| d | 0.0159 | 0.0000 | +0.0159 |
| ol | 0.0278 | 0.0125 | +0.0152 |
| qokain | 0.0000 | 0.0150 | -0.0150 |
| ain | 0.0020 | 0.0150 | -0.0131 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ol | 0.0216 | 0.0243 | -0.0027 |
| al | 0.0194 | 0.0485 | -0.0291 |
| y | 0.0130 | 0.0189 | -0.0059 |
| am | 0.0108 | 0.0189 | -0.0081 |
| chey | 0.0216 | 0.0108 | +0.0108 |
| okain | 0.0130 | 0.0108 | +0.0022 |
| cheey | 0.0086 | 0.0108 | -0.0021 |
| otar | 0.0194 | 0.0081 | +0.0114 |
| oteey | 0.0086 | 0.0081 | +0.0006 |
| shedy | 0.0086 | 0.0081 | +0.0006 |
| chedy | 0.0065 | 0.0081 | -0.0016 |
| or | 0.0065 | 0.0270 | -0.0205 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0000 | 0.0728 | -0.0728 |
| al | 0.0194 | 0.0485 | -0.0291 |
| okal | 0.0216 | 0.0000 | +0.0216 |
| or | 0.0065 | 0.0270 | -0.0205 |
| ain | 0.0000 | 0.0189 | -0.0189 |
| air | 0.0000 | 0.0189 | -0.0189 |
| ar | 0.0043 | 0.0216 | -0.0172 |
| aiiin | 0.0000 | 0.0162 | -0.0162 |
| okaiin | 0.0173 | 0.0054 | +0.0119 |
| otar | 0.0194 | 0.0081 | +0.0114 |
| chey | 0.0216 | 0.0108 | +0.0108 |
| d | 0.0108 | 0.0000 | +0.0108 |

Context diagnostics: predecessor Jaccard 0.1600, JS 0.4439, entropy A/B 4.485/5.116, effective vocabulary A/B 88.70/166.71; successor Jaccard 0.1382, JS 0.3642, entropy A/B 5.437/4.969, effective vocabulary A/B 229.86/143.92.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `lkar`, `los`, `ytar`, `daiin`, `dair`, `tar`, `lor`, `sor`, `chor`, `ches`, `o`, `char`; right `chcthy`, `cheky`, `chor`, `okol`, `shol`, `yteey`, `otedy`, `sheey`, `chedy`, `oky`, `shey`, `o`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `qokal`, `cheey`, `okeey`, `dy`, `chy`, `chdy`, `okedy`, `lchedy`, `okeedy`, `oteedy`, `qokol`; right `qokedy`, `qokal`, `dain`, `qokar`, `qoky`, `qol`, `lchedy`, `sheol`, `qokey`, `dol`, `sho`, `dam`.

## `qokain` / `qol`

Structural similarity: 0.7614; reliability: 0.9236; normalized graphemic distance: 0.6667; counts: 279/142.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9601 | 0.9851 |
| Left context | 0.6971 | 0.8910 |
| Right context | 0.6268 | 0.8946 |

- Primary component: positional agreement (0.960).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.697.
- Largest right-context difference: chedy is more frequent for qol (absolute probability difference 0.064).

Position summaries (A/B): line-start 0.0896/0.1197, line-end 0.0358/0.0634, mean 4.165/3.915, median 4.000/4.000. Position JS similarity: 0.9601.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0709 | 0.1040 | -0.0331 |
| shedy | 0.0433 | 0.0640 | -0.0207 |
| chey | 0.0315 | 0.0560 | -0.0245 |
| qokedy | 0.0236 | 0.0240 | -0.0004 |
| sheey | 0.0236 | 0.0240 | -0.0004 |
| qokeedy | 0.0197 | 0.0320 | -0.0123 |
| shey | 0.0669 | 0.0160 | +0.0509 |
| otedy | 0.0118 | 0.0160 | -0.0042 |
| qo | 0.0118 | 0.0080 | +0.0038 |
| chckhy | 0.0079 | 0.0080 | -0.0001 |
| checthy | 0.0079 | 0.0080 | -0.0001 |
| cheey | 0.0079 | 0.0240 | -0.0161 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shey | 0.0669 | 0.0160 | +0.0509 |
| chedy | 0.0709 | 0.1040 | -0.0331 |
| qol | 0.0000 | 0.0320 | -0.0320 |
| chey | 0.0315 | 0.0560 | -0.0245 |
| shedy | 0.0433 | 0.0640 | -0.0207 |
| sheedy | 0.0039 | 0.0240 | -0.0201 |
| cheey | 0.0079 | 0.0240 | -0.0161 |
| oteey | 0.0079 | 0.0240 | -0.0161 |
| dal | 0.0000 | 0.0160 | -0.0160 |
| qoky | 0.0000 | 0.0160 | -0.0160 |
| sshey | 0.0000 | 0.0160 | -0.0160 |
| y | 0.0000 | 0.0160 | -0.0160 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0483 | 0.1128 | -0.0645 |
| shedy | 0.0335 | 0.0752 | -0.0417 |
| ol | 0.0483 | 0.0301 | +0.0183 |
| chey | 0.0260 | 0.0376 | -0.0116 |
| shey | 0.0223 | 0.0226 | -0.0003 |
| cheedy | 0.0149 | 0.0301 | -0.0152 |
| okaiin | 0.0149 | 0.0150 | -0.0002 |
| otar | 0.0149 | 0.0150 | -0.0002 |
| sheey | 0.0149 | 0.0226 | -0.0077 |
| cheey | 0.0112 | 0.0301 | -0.0189 |
| cheol | 0.0112 | 0.0150 | -0.0039 |
| otain | 0.0112 | 0.0150 | -0.0039 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0483 | 0.1128 | -0.0645 |
| shedy | 0.0335 | 0.0752 | -0.0417 |
| sheedy | 0.0074 | 0.0451 | -0.0377 |
| chckhy | 0.0372 | 0.0000 | +0.0372 |
| qol | 0.0000 | 0.0301 | -0.0301 |
| ar | 0.0223 | 0.0000 | +0.0223 |
| cheey | 0.0112 | 0.0301 | -0.0189 |
| dar | 0.0186 | 0.0000 | +0.0186 |
| okain | 0.0186 | 0.0000 | +0.0186 |
| ol | 0.0483 | 0.0301 | +0.0183 |
| cheedy | 0.0149 | 0.0301 | -0.0152 |
| aiin | 0.0000 | 0.0150 | -0.0150 |

Context diagnostics: predecessor Jaccard 0.1354, JS 0.4286, entropy A/B 4.564/4.010, effective vocabulary A/B 95.94/55.16; successor Jaccard 0.1179, JS 0.4069, entropy A/B 4.617/3.950, effective vocabulary A/B 101.20/51.93.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `otedy`, `qokey`; right `cheol`, `okaiin`, `otain`, `otar`.

Shared unobserved high-frequency contexts (descriptive absence only): left `ol`, `or`, `ar`, `s`, `dar`, `qokaiin`, `al`, `chor`, `dain`, `l`, `cheol`, `r`; right `chol`, `s`, `qokeedy`, `qokaiin`, `dain`, `qokar`, `otaiin`, `oteey`, `lchedy`, `okeedy`, `cthy`, `qokey`.

## `chol` / `cthy`

Structural similarity: 0.6832; reliability: 0.8970; normalized graphemic distance: 0.7500; counts: 395/103.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9271 | 0.9790 |
| Left context | 0.4803 | 0.8535 |
| Right context | 0.6424 | 0.8585 |

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

Shared unobserved high-frequency contexts (descriptive absence only): left `ar`, `qokeedy`, `qokain`, `o`, `r`, `otedy`, `otar`, `oteey`, `qol`, `chckhy`, `okain`, `okedy`; right `shey`, `qokedy`, `qokal`, `r`, `okeey`, `qokar`, `okar`, `oteey`, `okedy`, `ain`, `lchedy`, `okeedy`.

## `lchedy` / `qokeey`

Structural similarity: 0.7030; reliability: 0.9069; normalized graphemic distance: 0.6667; counts: 116/307.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9578 | 0.9812 |
| Left context | 0.6228 | 0.8676 |
| Right context | 0.5286 | 0.8720 |

- Primary component: positional agreement (0.958).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.623.
- Largest left-context difference: shedy is more frequent for qokeey (absolute probability difference 0.047).

Position summaries (A/B): line-start 0.0517/0.1042, line-end 0.1466/0.0358, mean 4.543/4.065, median 4.000/4.000. Position JS similarity: 0.9578.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0545 | 0.0691 | -0.0145 |
| qokeey | 0.0364 | 0.0436 | -0.0073 |
| qokeedy | 0.0455 | 0.0327 | +0.0127 |
| chey | 0.0364 | 0.0255 | +0.0109 |
| cheey | 0.0182 | 0.0182 | +0.0000 |
| okeey | 0.0182 | 0.0327 | -0.0145 |
| cheol | 0.0091 | 0.0109 | -0.0018 |
| keey | 0.0091 | 0.0145 | -0.0055 |
| lchedy | 0.0091 | 0.0109 | -0.0018 |
| sheedy | 0.0091 | 0.0109 | -0.0018 |
| shey | 0.0091 | 0.0182 | -0.0091 |
| chdy | 0.0273 | 0.0073 | +0.0200 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0000 | 0.0473 | -0.0473 |
| dal | 0.0273 | 0.0000 | +0.0273 |
| ol | 0.0273 | 0.0036 | +0.0236 |
| chdy | 0.0273 | 0.0073 | +0.0200 |
| okedy | 0.0273 | 0.0073 | +0.0200 |
| oteedy | 0.0273 | 0.0073 | +0.0200 |
| qokedy | 0.0273 | 0.0073 | +0.0200 |
| lkedy | 0.0182 | 0.0000 | +0.0182 |
| chedy | 0.0545 | 0.0691 | -0.0145 |
| okeey | 0.0182 | 0.0327 | -0.0145 |
| al | 0.0182 | 0.0036 | +0.0145 |
| daiin | 0.0000 | 0.0145 | -0.0145 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokedy | 0.0303 | 0.0304 | -0.0001 |
| qokeedy | 0.0303 | 0.0304 | -0.0001 |
| qokeey | 0.0303 | 0.0405 | -0.0102 |
| chedy | 0.0606 | 0.0169 | +0.0437 |
| qokaiin | 0.0303 | 0.0135 | +0.0168 |
| qokain | 0.0202 | 0.0135 | +0.0067 |
| qoky | 0.0202 | 0.0135 | +0.0067 |
| chckhy | 0.0303 | 0.0101 | +0.0202 |
| qokey | 0.0303 | 0.0101 | +0.0202 |
| chey | 0.0101 | 0.0101 | -0.0000 |
| l | 0.0101 | 0.0101 | -0.0000 |
| lchedy | 0.0101 | 0.0135 | -0.0034 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0606 | 0.0169 | +0.0437 |
| daiin | 0.0000 | 0.0338 | -0.0338 |
| shedy | 0.0404 | 0.0068 | +0.0336 |
| lar | 0.0202 | 0.0000 | +0.0202 |
| lkchedy | 0.0202 | 0.0000 | +0.0202 |
| chckhy | 0.0303 | 0.0101 | +0.0202 |
| qokey | 0.0303 | 0.0101 | +0.0202 |
| lkaiin | 0.0202 | 0.0034 | +0.0168 |
| qokaiin | 0.0303 | 0.0135 | +0.0168 |
| okeey | 0.0101 | 0.0236 | -0.0135 |
| okain | 0.0000 | 0.0135 | -0.0135 |
| qotedy | 0.0000 | 0.0135 | -0.0135 |

Context diagnostics: predecessor Jaccard 0.1814, JS 0.4485, entropy A/B 4.186/4.699, effective vocabulary A/B 65.76/109.79; successor Jaccard 0.1304, JS 0.3847, entropy A/B 4.161/4.953, effective vocabulary A/B 64.12/141.57.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `cheey`; right ``.

Shared unobserved high-frequency contexts (descriptive absence only): left `chol`, `or`, `ar`, `s`, `qokaiin`, `chor`, `l`, `r`, `chy`, `otedy`, `otar`, `otaiin`; right `dy`, `al`, `cheey`, `chy`, `chdy`, `otar`, `qol`, `ain`, `cthy`, `sho`, `shy`, `oky`.

## `chedy` / `qokeey`

Structural similarity: 0.6743; reliability: 0.9332; normalized graphemic distance: 0.6667; counts: 506/307.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9273 | 0.9873 |
| Left context | 0.3819 | 0.9046 |
| Right context | 0.7135 | 0.9077 |

- Primary component: positional agreement (0.927).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.714.
- Largest left-context difference: chedy is more frequent for qokeey (absolute probability difference 0.049).

Position summaries (A/B): line-start 0.0119/0.1042, line-end 0.0692/0.0358, mean 5.401/4.065, median 5.000/4.000. Position JS similarity: 0.9273.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0200 | 0.0691 | -0.0491 |
| shedy | 0.0160 | 0.0473 | -0.0313 |
| daiin | 0.0200 | 0.0145 | +0.0055 |
| qokeedy | 0.0140 | 0.0327 | -0.0187 |
| lchedy | 0.0120 | 0.0109 | +0.0011 |
| qokeey | 0.0100 | 0.0436 | -0.0336 |
| okeedy | 0.0080 | 0.0109 | -0.0029 |
| dar | 0.0140 | 0.0073 | +0.0067 |
| okedy | 0.0080 | 0.0073 | +0.0007 |
| qokedy | 0.0220 | 0.0073 | +0.0147 |
| aiin | 0.0060 | 0.0073 | -0.0013 |
| cheol | 0.0040 | 0.0109 | -0.0069 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0200 | 0.0691 | -0.0491 |
| ol | 0.0440 | 0.0036 | +0.0404 |
| qokeey | 0.0100 | 0.0436 | -0.0336 |
| shedy | 0.0160 | 0.0473 | -0.0313 |
| qol | 0.0300 | 0.0000 | +0.0300 |
| okeey | 0.0040 | 0.0327 | -0.0287 |
| chey | 0.0000 | 0.0255 | -0.0255 |
| qokain | 0.0260 | 0.0036 | +0.0224 |
| l | 0.0220 | 0.0000 | +0.0220 |
| qokeedy | 0.0140 | 0.0327 | -0.0187 |
| qokal | 0.0220 | 0.0036 | +0.0184 |
| cheey | 0.0000 | 0.0182 | -0.0182 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokeey | 0.0403 | 0.0405 | -0.0002 |
| qokedy | 0.0297 | 0.0304 | -0.0007 |
| daiin | 0.0212 | 0.0338 | -0.0126 |
| qokeedy | 0.0212 | 0.0304 | -0.0092 |
| chedy | 0.0212 | 0.0169 | +0.0043 |
| qokaiin | 0.0212 | 0.0135 | +0.0077 |
| qokain | 0.0382 | 0.0135 | +0.0247 |
| lchedy | 0.0127 | 0.0135 | -0.0008 |
| okeey | 0.0106 | 0.0236 | -0.0130 |
| ol | 0.0106 | 0.0169 | -0.0063 |
| qokal | 0.0127 | 0.0101 | +0.0026 |
| qokey | 0.0127 | 0.0101 | +0.0026 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qol | 0.0276 | 0.0000 | +0.0276 |
| qokain | 0.0382 | 0.0135 | +0.0247 |
| qokar | 0.0212 | 0.0068 | +0.0145 |
| okeey | 0.0106 | 0.0236 | -0.0130 |
| daiin | 0.0212 | 0.0338 | -0.0126 |
| lsheey | 0.0000 | 0.0101 | -0.0101 |
| oteey | 0.0000 | 0.0101 | -0.0101 |
| qotain | 0.0127 | 0.0034 | +0.0094 |
| okain | 0.0042 | 0.0135 | -0.0093 |
| qokeedy | 0.0212 | 0.0304 | -0.0092 |
| dy | 0.0085 | 0.0000 | +0.0085 |
| okar | 0.0085 | 0.0000 | +0.0085 |

Context diagnostics: predecessor Jaccard 0.1354, JS 0.3382, entropy A/B 5.066/4.699, effective vocabulary A/B 158.50/109.79; successor Jaccard 0.2065, JS 0.5185, entropy A/B 5.125/4.953, effective vocabulary A/B 168.14/141.57.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `aiin`, `oteey`, `shol`, `okedy`, `cheol`, `okeedy`, `sheedy`, `lchedy`, `dar`, `daiin`; right `dar`, `lchey`, `otey`, `oty`, `raiin`, `rol`, `shedy`, `okeedy`, `chckhy`, `l`, `or`, `s`.

Shared unobserved high-frequency contexts (descriptive absence only): left `s`, `chy`, `otaiin`, `sho`, `oty`, `cheor`, `char`, `cho`, `cheo`, `kaiin`, `chcthy`, `oky`; right `al`, `cheey`, `ain`, `cthy`, `sho`, `cheor`, `air`, `dair`, `kaiin`, `shor`, `cho`, `cheo`.

## `qokedy` / `qol`

Structural similarity: 0.6804; reliability: 0.9236; normalized graphemic distance: 0.6667; counts: 276/142.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9567 | 0.9851 |
| Left context | 0.6555 | 0.8910 |
| Right context | 0.4290 | 0.8946 |

- Primary component: positional agreement (0.957).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.655.
- Largest right-context difference: chedy is more frequent for qol (absolute probability difference 0.071).

Position summaries (A/B): line-start 0.0725/0.1197, line-end 0.0362/0.0634, mean 3.986/3.915, median 4.000/4.000. Position JS similarity: 0.9567.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0820 | 0.0640 | +0.0180 |
| chedy | 0.0547 | 0.1040 | -0.0493 |
| qokeedy | 0.0469 | 0.0320 | +0.0149 |
| qokedy | 0.0625 | 0.0240 | +0.0385 |
| sheedy | 0.0195 | 0.0240 | -0.0045 |
| otedy | 0.0234 | 0.0160 | +0.0074 |
| shey | 0.0156 | 0.0160 | -0.0004 |
| qokey | 0.0117 | 0.0160 | -0.0043 |
| daiin | 0.0078 | 0.0080 | -0.0002 |
| dy | 0.0078 | 0.0080 | -0.0002 |
| sheey | 0.0078 | 0.0240 | -0.0162 |
| ychedy | 0.0078 | 0.0080 | -0.0002 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chey | 0.0039 | 0.0560 | -0.0521 |
| chedy | 0.0547 | 0.1040 | -0.0493 |
| qokedy | 0.0625 | 0.0240 | +0.0385 |
| qokeey | 0.0352 | 0.0000 | +0.0352 |
| qol | 0.0000 | 0.0320 | -0.0320 |
| okedy | 0.0234 | 0.0000 | +0.0234 |
| cheey | 0.0039 | 0.0240 | -0.0201 |
| oteey | 0.0039 | 0.0240 | -0.0201 |
| shedy | 0.0820 | 0.0640 | +0.0180 |
| sheey | 0.0078 | 0.0240 | -0.0162 |
| y | 0.0000 | 0.0160 | -0.0160 |
| qokeedy | 0.0469 | 0.0320 | +0.0149 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0414 | 0.1128 | -0.0714 |
| shedy | 0.0338 | 0.0752 | -0.0414 |
| ol | 0.0226 | 0.0301 | -0.0075 |
| qol | 0.0113 | 0.0301 | -0.0188 |
| chdy | 0.0075 | 0.0075 | +0.0000 |
| dy | 0.0150 | 0.0075 | +0.0075 |
| lchey | 0.0075 | 0.0075 | +0.0000 |
| or | 0.0075 | 0.0150 | -0.0075 |
| otar | 0.0075 | 0.0150 | -0.0075 |
| otedy | 0.0263 | 0.0075 | +0.0188 |
| oteedy | 0.0113 | 0.0075 | +0.0038 |
| qokal | 0.0188 | 0.0075 | +0.0113 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0414 | 0.1128 | -0.0714 |
| qokedy | 0.0602 | 0.0000 | +0.0602 |
| qokeedy | 0.0564 | 0.0000 | +0.0564 |
| shedy | 0.0338 | 0.0752 | -0.0414 |
| sheedy | 0.0038 | 0.0451 | -0.0414 |
| chey | 0.0038 | 0.0376 | -0.0338 |
| cheedy | 0.0000 | 0.0301 | -0.0301 |
| cheey | 0.0000 | 0.0301 | -0.0301 |
| dal | 0.0263 | 0.0000 | +0.0263 |
| qokain | 0.0226 | 0.0000 | +0.0226 |
| otedy | 0.0263 | 0.0075 | +0.0188 |
| qol | 0.0113 | 0.0301 | -0.0188 |

Context diagnostics: predecessor Jaccard 0.1495, JS 0.4375, entropy A/B 4.499/4.010, effective vocabulary A/B 89.96/55.16; successor Jaccard 0.1256, JS 0.3243, entropy A/B 4.606/3.950, effective vocabulary A/B 100.07/51.93.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `qokey`, `shey`; right `or`, `otar`.

Shared unobserved high-frequency contexts (descriptive absence only): left `ol`, `aiin`, `or`, `ar`, `s`, `dar`, `chor`, `okaiin`, `dain`, `shol`, `l`, `cheol`; right `chor`, `o`, `oteey`, `oty`, `okeedy`, `cthy`, `sho`, `shy`, `oky`, `cheor`, `cheody`, `dam`.

## `dal` / `ol`

Structural similarity: 0.6704; reliability: 0.9332; normalized graphemic distance: 0.6667; counts: 242/560.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9390 | 0.9873 |
| Left context | 0.4907 | 0.9046 |
| Right context | 0.5815 | 0.9077 |

- Primary component: positional agreement (0.939).
- Similarity is concentrated: the next component, successor-distribution overlap, is 0.581.
- Largest left-context difference: daiin is more frequent for dal (absolute probability difference 0.034).

Position summaries (A/B): line-start 0.0331/0.0554, line-end 0.2025/0.0750, mean 6.707/5.220, median 5.000/4.000. Position JS similarity: 0.9390.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0171 | 0.0189 | -0.0018 |
| qokain | 0.0171 | 0.0246 | -0.0075 |
| ol | 0.0171 | 0.0151 | +0.0020 |
| daiin | 0.0470 | 0.0132 | +0.0338 |
| chol | 0.0128 | 0.0170 | -0.0042 |
| qokedy | 0.0299 | 0.0113 | +0.0186 |
| shedy | 0.0128 | 0.0113 | +0.0015 |
| chedy | 0.0171 | 0.0095 | +0.0076 |
| chey | 0.0171 | 0.0095 | +0.0076 |
| qokal | 0.0214 | 0.0095 | +0.0119 |
| al | 0.0085 | 0.0076 | +0.0010 |
| cheol | 0.0128 | 0.0057 | +0.0071 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| daiin | 0.0470 | 0.0132 | +0.0338 |
| or | 0.0000 | 0.0189 | -0.0189 |
| qokedy | 0.0299 | 0.0113 | +0.0186 |
| ar | 0.0000 | 0.0170 | -0.0170 |
| r | 0.0000 | 0.0151 | -0.0151 |
| qokal | 0.0214 | 0.0095 | +0.0119 |
| okal | 0.0171 | 0.0057 | +0.0114 |
| ain | 0.0000 | 0.0113 | -0.0113 |
| okar | 0.0000 | 0.0113 | -0.0113 |
| cheey | 0.0128 | 0.0019 | +0.0109 |
| dar | 0.0043 | 0.0151 | -0.0108 |
| qokaiin | 0.0043 | 0.0151 | -0.0108 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0415 | 0.0463 | -0.0049 |
| chedy | 0.0207 | 0.0425 | -0.0217 |
| daiin | 0.0311 | 0.0174 | +0.0137 |
| s | 0.0155 | 0.0174 | -0.0018 |
| ol | 0.0155 | 0.0154 | +0.0001 |
| or | 0.0207 | 0.0135 | +0.0072 |
| chey | 0.0104 | 0.0174 | -0.0070 |
| ar | 0.0104 | 0.0097 | +0.0007 |
| chor | 0.0155 | 0.0097 | +0.0059 |
| chy | 0.0104 | 0.0097 | +0.0007 |
| al | 0.0155 | 0.0077 | +0.0078 |
| chdy | 0.0155 | 0.0077 | +0.0078 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0052 | 0.0270 | -0.0218 |
| chedy | 0.0207 | 0.0425 | -0.0217 |
| cheey | 0.0000 | 0.0212 | -0.0212 |
| sheey | 0.0000 | 0.0212 | -0.0212 |
| dar | 0.0259 | 0.0058 | +0.0201 |
| dy | 0.0207 | 0.0039 | +0.0169 |
| kaiin | 0.0052 | 0.0193 | -0.0141 |
| daiin | 0.0311 | 0.0174 | +0.0137 |
| dair | 0.0155 | 0.0019 | +0.0136 |
| y | 0.0155 | 0.0039 | +0.0117 |
| cheedy | 0.0000 | 0.0116 | -0.0116 |
| kedy | 0.0000 | 0.0116 | -0.0116 |

Context diagnostics: predecessor Jaccard 0.1197, JS 0.3462, entropy A/B 4.969/5.400, effective vocabulary A/B 143.82/221.44; successor Jaccard 0.1408, JS 0.4086, entropy A/B 4.756/5.098, effective vocabulary A/B 116.25/163.65.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `al`, `otar`, `saiin`, `cheol`, `shedy`, `chol`, `chedy`, `chey`, `okal`, `ol`, `aiin`; right `ar`, `chol`, `chy`, `dol`, `qokar`, `sheedy`, `al`, `chdy`, `chor`, `lchedy`, `ol`, `y`.

Shared unobserved high-frequency contexts (descriptive absence only): left `o`, `shy`, `char`, `air`, `cheody`, `cthy`, `qoteedy`, `dor`, `odaiin`, `qotchy`, `raiin`, `cheedy`; right `qokain`, `otal`, `oty`, `cthy`, `qokey`, `otain`, `am`, `qotedy`, `qoty`, `shor`, `qotaiin`, `saiin`.

## `ain` / `al`

Structural similarity: 0.6588; reliability: 0.9048; normalized graphemic distance: 0.6667; counts: 113/261.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9179 | 0.9807 |
| Left context | 0.6234 | 0.8645 |
| Right context | 0.4350 | 0.8691 |

- Primary component: positional agreement (0.918).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.623.
- Largest left-context difference: r is more frequent for ain (absolute probability difference 0.097).

Position summaries (A/B): line-start 0.0088/0.0038, line-end 0.1062/0.1111, mean 5.106/7.536, median 5.000/6.000. Position JS similarity: 0.9179.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ar | 0.0625 | 0.0692 | -0.0067 |
| or | 0.1071 | 0.0308 | +0.0764 |
| s | 0.0625 | 0.0308 | +0.0317 |
| sar | 0.0536 | 0.0269 | +0.0266 |
| dar | 0.0268 | 0.0231 | +0.0037 |
| r | 0.1161 | 0.0192 | +0.0968 |
| otar | 0.0179 | 0.0192 | -0.0014 |
| ol | 0.0179 | 0.0154 | +0.0025 |
| ain | 0.0089 | 0.0115 | -0.0026 |
| cheo | 0.0089 | 0.0115 | -0.0026 |
| dair | 0.0089 | 0.0231 | -0.0141 |
| okar | 0.0089 | 0.0115 | -0.0026 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| r | 0.1161 | 0.0192 | +0.0968 |
| or | 0.1071 | 0.0308 | +0.0764 |
| aiin | 0.0000 | 0.0346 | -0.0346 |
| s | 0.0625 | 0.0308 | +0.0317 |
| sar | 0.0536 | 0.0269 | +0.0266 |
| daiin | 0.0000 | 0.0192 | -0.0192 |
| air | 0.0268 | 0.0077 | +0.0191 |
| qol | 0.0179 | 0.0000 | +0.0179 |
| rar | 0.0179 | 0.0000 | +0.0179 |
| sor | 0.0179 | 0.0000 | +0.0179 |
| otaiin | 0.0000 | 0.0154 | -0.0154 |
| dair | 0.0089 | 0.0231 | -0.0141 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ar | 0.0594 | 0.0388 | +0.0206 |
| ol | 0.0594 | 0.0172 | +0.0422 |
| al | 0.0297 | 0.0129 | +0.0168 |
| o | 0.0198 | 0.0129 | +0.0069 |
| chedy | 0.0099 | 0.0259 | -0.0160 |
| ches | 0.0099 | 0.0129 | -0.0030 |
| r | 0.0099 | 0.0129 | -0.0030 |
| chol | 0.0099 | 0.0086 | +0.0013 |
| okeey | 0.0198 | 0.0086 | +0.0112 |
| y | 0.0198 | 0.0086 | +0.0112 |
| aiiin | 0.0099 | 0.0043 | +0.0056 |
| aiin | 0.0099 | 0.0043 | +0.0056 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ol | 0.0594 | 0.0172 | +0.0422 |
| s | 0.0000 | 0.0259 | -0.0259 |
| chey | 0.0297 | 0.0043 | +0.0254 |
| shey | 0.0297 | 0.0043 | +0.0254 |
| dar | 0.0000 | 0.0216 | -0.0216 |
| ar | 0.0594 | 0.0388 | +0.0206 |
| chl | 0.0198 | 0.0000 | +0.0198 |
| okaiin | 0.0198 | 0.0000 | +0.0198 |
| okan | 0.0198 | 0.0000 | +0.0198 |
| olkeey | 0.0198 | 0.0000 | +0.0198 |
| keedy | 0.0000 | 0.0172 | -0.0172 |
| al | 0.0297 | 0.0129 | +0.0168 |

Context diagnostics: predecessor Jaccard 0.1158, JS 0.4334, entropy A/B 3.657/4.683, effective vocabulary A/B 38.76/108.12; successor Jaccard 0.1031, JS 0.2773, entropy A/B 4.167/4.978, effective vocabulary A/B 64.53/145.16.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `lr`, `ol`, `os`, `sain`, `otar`; right `o`, `okeey`, `y`.

Shared unobserved high-frequency contexts (descriptive absence only): left `chedy`, `shedy`, `qokeedy`, `qokeey`, `shey`, `y`, `qokal`, `shol`, `cheey`, `okeey`, `dy`, `cheol`; right `qokeedy`, `qokain`, `qokaiin`, `chor`, `shol`, `qokar`, `okal`, `qoky`, `chckhy`, `okain`, `okar`, `okedy`.

## `ain` / `ar`

Structural similarity: 0.6572; reliability: 0.9048; normalized graphemic distance: 0.6667; counts: 113/403.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9264 | 0.9807 |
| Left context | 0.5425 | 0.8645 |
| Right context | 0.5028 | 0.8691 |

- Primary component: positional agreement (0.926).
- Similarity is concentrated: the next component, predecessor-distribution overlap, is 0.542.
- Largest left-context difference: r is more frequent for ain (absolute probability difference 0.101).

Position summaries (A/B): line-start 0.0088/0.0099, line-end 0.1062/0.0794, mean 5.106/6.633, median 5.000/6.000. Position JS similarity: 0.9264.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| s | 0.0625 | 0.0501 | +0.0124 |
| dar | 0.0268 | 0.0251 | +0.0017 |
| or | 0.1071 | 0.0226 | +0.0846 |
| ar | 0.0625 | 0.0201 | +0.0424 |
| otar | 0.0179 | 0.0301 | -0.0122 |
| r | 0.1161 | 0.0150 | +0.1010 |
| ol | 0.0179 | 0.0125 | +0.0053 |
| ain | 0.0089 | 0.0150 | -0.0061 |
| okar | 0.0089 | 0.0301 | -0.0211 |
| otain | 0.0089 | 0.0175 | -0.0086 |
| ches | 0.0089 | 0.0075 | +0.0014 |
| dair | 0.0089 | 0.0075 | +0.0014 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| r | 0.1161 | 0.0150 | +0.1010 |
| or | 0.1071 | 0.0226 | +0.0846 |
| sar | 0.0536 | 0.0025 | +0.0511 |
| ar | 0.0625 | 0.0201 | +0.0424 |
| al | 0.0000 | 0.0226 | -0.0226 |
| air | 0.0268 | 0.0050 | +0.0218 |
| okar | 0.0089 | 0.0301 | -0.0211 |
| os | 0.0179 | 0.0000 | +0.0179 |
| qol | 0.0179 | 0.0000 | +0.0179 |
| lr | 0.0179 | 0.0025 | +0.0154 |
| rar | 0.0179 | 0.0025 | +0.0154 |
| char | 0.0000 | 0.0150 | -0.0150 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| al | 0.0297 | 0.0485 | -0.0188 |
| ol | 0.0594 | 0.0243 | +0.0351 |
| ar | 0.0594 | 0.0216 | +0.0378 |
| or | 0.0198 | 0.0270 | -0.0072 |
| y | 0.0198 | 0.0189 | +0.0009 |
| chey | 0.0297 | 0.0108 | +0.0189 |
| aiiin | 0.0099 | 0.0162 | -0.0063 |
| aiin | 0.0099 | 0.0728 | -0.0629 |
| ain | 0.0099 | 0.0189 | -0.0090 |
| air | 0.0099 | 0.0189 | -0.0090 |
| aly | 0.0099 | 0.0108 | -0.0009 |
| am | 0.0099 | 0.0189 | -0.0090 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0099 | 0.0728 | -0.0629 |
| ar | 0.0594 | 0.0216 | +0.0378 |
| ol | 0.0594 | 0.0243 | +0.0351 |
| shey | 0.0297 | 0.0081 | +0.0216 |
| chl | 0.0198 | 0.0000 | +0.0198 |
| okan | 0.0198 | 0.0000 | +0.0198 |
| olkeey | 0.0198 | 0.0000 | +0.0198 |
| chey | 0.0297 | 0.0108 | +0.0189 |
| al | 0.0297 | 0.0485 | -0.0188 |
| cheol | 0.0198 | 0.0027 | +0.0171 |
| okeey | 0.0198 | 0.0027 | +0.0171 |
| o | 0.0198 | 0.0054 | +0.0144 |

Context diagnostics: predecessor Jaccard 0.0982, JS 0.3797, entropy A/B 3.657/5.116, effective vocabulary A/B 38.76/166.71; successor Jaccard 0.1212, JS 0.3996, entropy A/B 4.167/4.969, effective vocabulary A/B 64.53/143.92.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `ol`, `sain`, `sor`; right `o`, `okaiin`, `shedy`, `y`.

Shared unobserved high-frequency contexts (descriptive absence only): left `qokeedy`, `y`, `qokal`, `cheey`, `okeey`, `dy`, `chy`, `sheey`, `chdy`, `qoky`, `okedy`, `lchedy`; right `qokeey`, `dy`, `qokedy`, `qokain`, `dal`, `qokal`, `qokar`, `okal`, `qoky`, `chckhy`, `oty`, `lchedy`.

## `qokar` / `qol`

Structural similarity: 0.7126; reliability: 0.9231; normalized graphemic distance: 0.6000; counts: 159/142.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9456 | 0.9850 |
| Left context | 0.6357 | 0.8903 |
| Right context | 0.5565 | 0.8939 |

- Primary component: positional agreement (0.946).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.636.
- Largest right-context difference: chedy is more frequent for qol (absolute probability difference 0.087).

Position summaries (A/B): line-start 0.0440/0.1197, line-end 0.0252/0.0634, mean 4.868/3.915, median 4.000/4.000. Position JS similarity: 0.9456.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0658 | 0.1040 | -0.0382 |
| qokeedy | 0.0395 | 0.0320 | +0.0075 |
| shedy | 0.0263 | 0.0640 | -0.0377 |
| chey | 0.0197 | 0.0560 | -0.0363 |
| shey | 0.0329 | 0.0160 | +0.0169 |
| dal | 0.0132 | 0.0160 | -0.0028 |
| okeey | 0.0132 | 0.0160 | -0.0028 |
| qokedy | 0.0132 | 0.0240 | -0.0108 |
| sheedy | 0.0132 | 0.0240 | -0.0108 |
| cheedy | 0.0132 | 0.0080 | +0.0052 |
| daiin | 0.0132 | 0.0080 | +0.0052 |
| dy | 0.0132 | 0.0080 | +0.0052 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0658 | 0.1040 | -0.0382 |
| shedy | 0.0263 | 0.0640 | -0.0377 |
| chey | 0.0197 | 0.0560 | -0.0363 |
| qol | 0.0000 | 0.0320 | -0.0320 |
| cheey | 0.0000 | 0.0240 | -0.0240 |
| oteey | 0.0000 | 0.0240 | -0.0240 |
| sheey | 0.0000 | 0.0240 | -0.0240 |
| chdy | 0.0197 | 0.0000 | +0.0197 |
| qokaiin | 0.0197 | 0.0000 | +0.0197 |
| shey | 0.0329 | 0.0160 | +0.0169 |
| qoky | 0.0000 | 0.0160 | -0.0160 |
| sshey | 0.0000 | 0.0160 | -0.0160 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0710 | 0.0752 | -0.0042 |
| ol | 0.0323 | 0.0301 | +0.0022 |
| chedy | 0.0258 | 0.1128 | -0.0870 |
| chey | 0.0194 | 0.0376 | -0.0182 |
| shey | 0.0194 | 0.0226 | -0.0032 |
| or | 0.0194 | 0.0150 | +0.0043 |
| okaiin | 0.0129 | 0.0150 | -0.0021 |
| otar | 0.0129 | 0.0150 | -0.0021 |
| sheedy | 0.0129 | 0.0451 | -0.0322 |
| chcthy | 0.0129 | 0.0075 | +0.0054 |
| chl | 0.0129 | 0.0075 | +0.0054 |
| chy | 0.0129 | 0.0075 | +0.0054 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.0258 | 0.1128 | -0.0870 |
| okar | 0.0387 | 0.0000 | +0.0387 |
| sheedy | 0.0129 | 0.0451 | -0.0322 |
| cheedy | 0.0000 | 0.0301 | -0.0301 |
| qol | 0.0000 | 0.0301 | -0.0301 |
| chckhy | 0.0258 | 0.0000 | +0.0258 |
| cheey | 0.0065 | 0.0301 | -0.0236 |
| sheey | 0.0000 | 0.0226 | -0.0226 |
| ar | 0.0194 | 0.0000 | +0.0194 |
| ary | 0.0194 | 0.0000 | +0.0194 |
| checkhy | 0.0194 | 0.0000 | +0.0194 |
| qokain | 0.0194 | 0.0000 | +0.0194 |

Context diagnostics: predecessor Jaccard 0.1065, JS 0.3555, entropy A/B 4.520/4.010, effective vocabulary A/B 91.79/55.16; successor Jaccard 0.1745, JS 0.4130, entropy A/B 4.385/3.950, effective vocabulary A/B 80.25/51.93.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `dal`, `okeey`; right `okaiin`, `otar`, `or`.

Shared unobserved high-frequency contexts (descriptive absence only): left `aiin`, `or`, `ar`, `s`, `dar`, `qokain`, `al`, `chor`, `dain`, `l`, `r`, `qokar`; right `daiin`, `dar`, `qokeedy`, `qokedy`, `qokaiin`, `dain`, `o`, `qokar`, `okain`, `oteey`, `okedy`, `cthy`.

## `qol` / `qotain`

Structural similarity: 0.6640; reliability: 0.8393; normalized graphemic distance: 0.6667; counts: 142/60.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9056 | 0.9632 |
| Left context | 0.6057 | 0.7734 |
| Right context | 0.4808 | 0.7812 |

- Primary component: positional agreement (0.906).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.606.
- Largest right-context difference: shedy is more frequent for qol (absolute probability difference 0.075).

Position summaries (A/B): line-start 0.1197/0.0500, line-end 0.0634/0.0667, mean 3.915/5.050, median 4.000/5.000. Position JS similarity: 0.9056.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.1040 | 0.1053 | -0.0013 |
| shedy | 0.0640 | 0.0526 | +0.0114 |
| qokeedy | 0.0320 | 0.0526 | -0.0206 |
| cheey | 0.0240 | 0.0351 | -0.0111 |
| sheey | 0.0240 | 0.0351 | -0.0111 |
| sheedy | 0.0240 | 0.0175 | +0.0065 |
| shey | 0.0160 | 0.0175 | -0.0015 |
| chckhy | 0.0080 | 0.0175 | -0.0095 |
| chol | 0.0080 | 0.0175 | -0.0095 |
| qokshedy | 0.0080 | 0.0175 | -0.0095 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chey | 0.0560 | 0.0000 | +0.0560 |
| chody | 0.0000 | 0.0351 | -0.0351 |
| otar | 0.0000 | 0.0351 | -0.0351 |
| qoteey | 0.0000 | 0.0351 | -0.0351 |
| qol | 0.0320 | 0.0000 | +0.0320 |
| oteey | 0.0240 | 0.0000 | +0.0240 |
| qokedy | 0.0240 | 0.0000 | +0.0240 |
| qokeedy | 0.0320 | 0.0526 | -0.0206 |
| chchy | 0.0000 | 0.0175 | -0.0175 |
| cheed | 0.0000 | 0.0175 | -0.0175 |
| kalkal | 0.0000 | 0.0175 | -0.0175 |
| lkeed | 0.0000 | 0.0175 | -0.0175 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chedy | 0.1128 | 0.1250 | -0.0122 |
| chey | 0.0376 | 0.0179 | +0.0197 |
| ol | 0.0301 | 0.0179 | +0.0122 |
| cheol | 0.0150 | 0.0179 | -0.0028 |
| okaiin | 0.0150 | 0.0179 | -0.0028 |
| otar | 0.0150 | 0.0179 | -0.0028 |
| chcthy | 0.0075 | 0.0179 | -0.0103 |
| oteedy | 0.0075 | 0.0714 | -0.0639 |
| sheckhy | 0.0075 | 0.0179 | -0.0103 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| shedy | 0.0752 | 0.0000 | +0.0752 |
| oteedy | 0.0075 | 0.0714 | -0.0639 |
| shcthy | 0.0000 | 0.0536 | -0.0536 |
| sheedy | 0.0451 | 0.0000 | +0.0451 |
| ar | 0.0000 | 0.0357 | -0.0357 |
| okedy | 0.0000 | 0.0357 | -0.0357 |
| qokain | 0.0000 | 0.0357 | -0.0357 |
| cheedy | 0.0301 | 0.0000 | +0.0301 |
| cheey | 0.0301 | 0.0000 | +0.0301 |
| qol | 0.0301 | 0.0000 | +0.0301 |
| sheey | 0.0226 | 0.0000 | +0.0226 |
| shey | 0.0226 | 0.0000 | +0.0226 |

Context diagnostics: predecessor Jaccard 0.0917, JS 0.3321, entropy A/B 4.010/3.617, effective vocabulary A/B 55.16/37.23; successor Jaccard 0.0833, JS 0.2559, entropy A/B 3.950/3.550, effective vocabulary A/B 51.93/34.81.

Shared unobserved high-frequency contexts (descriptive absence only): left `ol`, `aiin`, `or`, `ar`, `s`, `dar`, `qokaiin`, `al`, `chor`, `okaiin`, `dain`, `shol`; right `daiin`, `chol`, `s`, `dar`, `qokeedy`, `qokeey`, `al`, `qokedy`, `qokaiin`, `dal`, `chor`, `dain`.

## `okar` / `otain`

Structural similarity: 0.6620; reliability: 0.8809; normalized graphemic distance: 0.6000; counts: 140/96.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9381 | 0.9751 |
| Left context | 0.4224 | 0.8309 |
| Right context | 0.6255 | 0.8367 |

- Primary component: positional agreement (0.938).
- Similarity is multidimensional: successor-distribution overlap also contributes 0.626.
- Largest right-context difference: chdy is more frequent for okar (absolute probability difference 0.038).

Position summaries (A/B): line-start 0.0786/0.0208, line-end 0.0571/0.0938, mean 5.579/5.510, median 5.000/5.000. Position JS similarity: 0.9381.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| aiin | 0.0233 | 0.0319 | -0.0087 |
| otain | 0.0233 | 0.0319 | -0.0087 |
| shedy | 0.0233 | 0.0426 | -0.0193 |
| daiin | 0.0310 | 0.0213 | +0.0097 |
| qokaiin | 0.0233 | 0.0213 | +0.0020 |
| okaiin | 0.0155 | 0.0213 | -0.0058 |
| qokain | 0.0155 | 0.0319 | -0.0164 |
| chedy | 0.0310 | 0.0106 | +0.0204 |
| qokal | 0.0155 | 0.0106 | +0.0049 |
| qokar | 0.0465 | 0.0106 | +0.0359 |
| qokedy | 0.0233 | 0.0106 | +0.0126 |
| chedal | 0.0078 | 0.0106 | -0.0029 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| qokar | 0.0465 | 0.0106 | +0.0359 |
| otaiin | 0.0078 | 0.0426 | -0.0348 |
| oteey | 0.0000 | 0.0319 | -0.0319 |
| chckhy | 0.0000 | 0.0213 | -0.0213 |
| l | 0.0000 | 0.0213 | -0.0213 |
| otan | 0.0000 | 0.0213 | -0.0213 |
| qol | 0.0000 | 0.0213 | -0.0213 |
| shey | 0.0000 | 0.0213 | -0.0213 |
| chedy | 0.0310 | 0.0106 | +0.0204 |
| shedy | 0.0233 | 0.0426 | -0.0193 |
| qokain | 0.0155 | 0.0319 | -0.0164 |
| chdy | 0.0155 | 0.0000 | +0.0155 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| ar | 0.0909 | 0.0805 | +0.0104 |
| ol | 0.0455 | 0.0345 | +0.0110 |
| shedy | 0.0303 | 0.0460 | -0.0157 |
| otar | 0.0379 | 0.0230 | +0.0149 |
| al | 0.0227 | 0.0230 | -0.0003 |
| okedy | 0.0303 | 0.0115 | +0.0188 |
| shey | 0.0152 | 0.0115 | +0.0037 |
| y | 0.0303 | 0.0115 | +0.0188 |
| ain | 0.0076 | 0.0115 | -0.0039 |
| char | 0.0076 | 0.0115 | -0.0039 |
| chcthy | 0.0076 | 0.0115 | -0.0039 |
| chedy | 0.0076 | 0.0115 | -0.0039 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| chdy | 0.0379 | 0.0000 | +0.0379 |
| chey | 0.0000 | 0.0345 | -0.0345 |
| otain | 0.0000 | 0.0345 | -0.0345 |
| okar | 0.0076 | 0.0345 | -0.0269 |
| otal | 0.0076 | 0.0345 | -0.0269 |
| chckhy | 0.0000 | 0.0230 | -0.0230 |
| okedy | 0.0303 | 0.0115 | +0.0188 |
| y | 0.0303 | 0.0115 | +0.0188 |
| shedy | 0.0303 | 0.0460 | -0.0157 |
| okain | 0.0076 | 0.0230 | -0.0154 |
| okol | 0.0152 | 0.0000 | +0.0152 |
| oky | 0.0152 | 0.0000 | +0.0152 |

Context diagnostics: predecessor Jaccard 0.1234, JS 0.3161, entropy A/B 4.488/4.167, effective vocabulary A/B 88.90/64.53; successor Jaccard 0.1716, JS 0.4143, entropy A/B 4.261/3.992, effective vocabulary A/B 70.85/54.19.

Shared unobserved high-frequency contexts (descriptive absence only): left `ol`, `chol`, `or`, `qokeedy`, `dar`, `y`, `al`, `chor`, `o`, `dy`, `cheol`, `r`; right `aiin`, `s`, `qokeedy`, `qokeey`, `qokain`, `dal`, `chor`, `cheey`, `qokal`, `l`, `shol`, `cheol`.

## Negative controls

Controls match unordered log-counts, normalized graphemic distance, and reliability, while favoring structural similarity near the full-corpus median. They are decomposed with exactly the target metrics.

| Target | Control | Structural | Reliability | Distance | Match cost |
|---|---|---:|---:|---:|---:|
| or/s | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 1.8470 |
| or/s | am/qokeey | 0.2799 | 0.8816 | 1.0000 | 1.9238 |
| or/s | ar/qotchy | 0.2760 | 0.8509 | 1.0000 | 2.0633 |
| chol/daiin | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 2.5890 |
| chol/daiin | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 2.7430 |
| chol/daiin | aiin/sor | 0.2553 | 0.8316 | 1.0000 | 2.8114 |
| r/s | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 1.0141 |
| r/s | am/qokeey | 0.2799 | 0.8816 | 1.0000 | 1.0909 |
| r/s | am/qokedy | 0.2926 | 0.8816 | 1.0000 | 1.3239 |
| ol/y | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 1.9169 |
| ol/y | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 2.0708 |
| ol/y | aiin/sor | 0.2553 | 0.8316 | 1.0000 | 2.1393 |
| dar/ol | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 1.9740 |
| dar/ol | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 2.1279 |
| dar/ol | aiin/sor | 0.2553 | 0.8316 | 1.0000 | 2.1964 |
| ar/ol | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 2.1947 |
| ar/ol | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 2.3486 |
| ar/ol | aiin/sor | 0.2553 | 0.8316 | 1.0000 | 2.4171 |
| chor/daiin | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 1.9642 |
| chor/daiin | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 2.1181 |
| chor/daiin | aiin/sor | 0.2553 | 0.8316 | 1.0000 | 2.1866 |
| chey/ol | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 2.0569 |
| chey/ol | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 2.2108 |
| chey/ol | aiin/sor | 0.2553 | 0.8316 | 1.0000 | 2.2793 |
| lchedy/qokar | chy/sar | 0.2865 | 0.8764 | 1.0000 | 0.6994 |
| lchedy/qokar | cheey/sar | 0.2769 | 0.8764 | 1.0000 | 0.7129 |
| lchedy/qokar | cthy/saiin | 0.2946 | 0.8792 | 1.0000 | 0.7298 |
| lchedy/qol | cthy/saiin | 0.2946 | 0.8792 | 1.0000 | 0.5999 |
| lchedy/qol | chdy/sar | 0.2846 | 0.8696 | 1.0000 | 0.6737 |
| lchedy/qol | am/cthy | 0.2691 | 0.8481 | 1.0000 | 0.7934 |
| okaiin/ol | sar/shedy | 0.2985 | 0.8764 | 0.8000 | 1.7811 |
| okaiin/ol | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 1.9065 |
| okaiin/ol | shedy/sor | 0.2571 | 0.8316 | 0.8000 | 1.9935 |
| okain/ol | sar/shedy | 0.2985 | 0.8764 | 0.8000 | 1.2700 |
| okain/ol | shedy/sor | 0.2571 | 0.8316 | 0.8000 | 1.4824 |
| okain/ol | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 1.5287 |
| qokaiin/qol | sar/shey | 0.2989 | 0.8764 | 0.7500 | 1.1467 |
| qokaiin/qol | al/sol | 0.2964 | 0.8584 | 0.6667 | 1.3367 |
| qokaiin/qol | r/shor | 0.3028 | 0.8910 | 0.7500 | 1.3888 |
| daiin/dol | sar/shedy | 0.2985 | 0.8764 | 0.8000 | 1.3865 |
| daiin/dol | shedy/sor | 0.2571 | 0.8316 | 0.8000 | 1.5989 |
| daiin/dol | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 1.6452 |
| aiin/ar | sar/shedy | 0.2985 | 0.8764 | 0.8000 | 2.3308 |
| aiin/ar | shedy/sor | 0.2571 | 0.8316 | 0.8000 | 2.5432 |
| aiin/ar | sar/shey | 0.2989 | 0.8764 | 0.7500 | 2.6788 |
| qokain/qol | sar/shey | 0.2989 | 0.8764 | 0.7500 | 1.1978 |
| qokain/qol | al/sol | 0.2964 | 0.8584 | 0.6667 | 1.2928 |
| qokain/qol | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 1.4532 |
| chol/cthy | sar/shedy | 0.2985 | 0.8764 | 0.8000 | 0.8461 |
| chol/cthy | sar/shey | 0.2989 | 0.8764 | 0.7500 | 1.0062 |
| chol/cthy | shedy/sor | 0.2571 | 0.8316 | 0.8000 | 1.0585 |
| lchedy/qokeey | sar/shey | 0.2989 | 0.8764 | 0.7500 | 1.0593 |
| lchedy/qokeey | am/qokeedy | 0.2501 | 0.8816 | 1.0000 | 1.1240 |
| lchedy/qokeey | al/sol | 0.2964 | 0.8584 | 0.6667 | 1.1542 |
| chedy/qokeey | sar/shedy | 0.2985 | 0.8764 | 0.8000 | 2.2301 |
| chedy/qokeey | shedy/sor | 0.2571 | 0.8316 | 0.8000 | 2.4425 |
| chedy/qokeey | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 2.4888 |
| qokedy/qol | sar/shey | 0.2989 | 0.8764 | 0.7500 | 1.2015 |
| qokedy/qol | al/sol | 0.2964 | 0.8584 | 0.6667 | 1.2820 |
| qokedy/qol | otam/qokain | 0.2671 | 0.8255 | 0.6667 | 1.3377 |
| dal/ol | sar/shedy | 0.2985 | 0.8764 | 0.8000 | 2.0943 |
| dal/ol | shedy/sor | 0.2571 | 0.8316 | 0.8000 | 2.3067 |
| dal/ol | aiin/qotchy | 0.2678 | 0.8509 | 1.0000 | 2.3530 |
| ain/al | sar/shey | 0.2989 | 0.8764 | 0.7500 | 0.9929 |
| ain/al | otam/qokain | 0.2671 | 0.8255 | 0.6667 | 1.1291 |
| ain/al | r/shor | 0.3028 | 0.8910 | 0.7500 | 1.2046 |
| ain/ar | sar/shedy | 0.2985 | 0.8764 | 0.8000 | 1.1001 |
| ain/ar | sar/shey | 0.2989 | 0.8764 | 0.7500 | 1.3002 |
| ain/ar | shedy/sor | 0.2571 | 0.8316 | 0.8000 | 1.3126 |
| qokar/qol | r/shor | 0.3028 | 0.8910 | 0.7500 | 1.2293 |
| qokar/qol | sar/shol | 0.2958 | 0.8764 | 0.7500 | 1.4516 |
| qokar/qol | sar/sheey | 0.3029 | 0.8711 | 0.8000 | 1.5421 |
| qol/qotain | sheey/sor | 0.2671 | 0.8267 | 0.8000 | 0.5590 |
| qol/qotain | shy/sor | 0.2621 | 0.7970 | 0.6667 | 0.6216 |
| qol/qotain | chckhy/pchedy | 0.2655 | 0.7872 | 0.6667 | 0.6619 |
| okar/otain | r/shor | 0.3028 | 0.8910 | 0.7500 | 0.9237 |
| okar/otain | am/sar | 0.2751 | 0.8290 | 0.6667 | 1.0275 |
| okar/otain | dam/saiin | 0.3053 | 0.8662 | 0.8000 | 1.0535 |

## Family decomposition

A family is a connected component; only listed edges define direct structural-distant links. Complete matrices, including non-edge pairs, are in `family_decomposition.yaml`.

### Family 1

Tokens: `aiin`, `ain`, `al`, `ar`, `chey`, `dal`, `dar`, `okaiin`, `okain`, `ol`, `y`. Structural medoid: `ol`. Peripheral token(s): `dal`.

Edges:

- `aiin` / `ar`: similarity 0.6715, reliability 0.9332, distance 0.7500
- `ain` / `al`: similarity 0.6588, reliability 0.9048, distance 0.6667
- `ain` / `ar`: similarity 0.6572, reliability 0.9048, distance 0.6667
- `ar` / `ol`: similarity 0.6639, reliability 0.9332, distance 1.0000
- `chey` / `ol`: similarity 0.6529, reliability 0.9332, distance 1.0000
- `dal` / `ol`: similarity 0.6704, reliability 0.9332, distance 0.6667
- `dar` / `ol`: similarity 0.6746, reliability 0.9332, distance 1.0000
- `okaiin` / `ol`: similarity 0.6885, reliability 0.9332, distance 0.8333
- `okain` / `ol`: similarity 0.6828, reliability 0.9230, distance 0.8000
- `ol` / `y`: similarity 0.6788, reliability 0.9332, distance 1.0000

### Family 2

Tokens: `chedy`, `lchedy`, `qokaiin`, `qokain`, `qokar`, `qokedy`, `qokeey`, `qol`, `qotain`. Structural medoid: `qokain`. Peripheral token(s): `chedy`.

Edges:

- `chedy` / `qokeey`: similarity 0.6743, reliability 0.9332, distance 0.6667
- `lchedy` / `qokar`: similarity 0.6585, reliability 0.9065, distance 1.0000
- `lchedy` / `qokeey`: similarity 0.7030, reliability 0.9069, distance 0.6667
- `lchedy` / `qol`: similarity 0.6556, reliability 0.8977, distance 1.0000
- `qokaiin` / `qol`: similarity 0.7398, reliability 0.9236, distance 0.7143
- `qokain` / `qol`: similarity 0.7614, reliability 0.9236, distance 0.6667
- `qokar` / `qol`: similarity 0.7126, reliability 0.9231, distance 0.6000
- `qokedy` / `qol`: similarity 0.6804, reliability 0.9236, distance 0.6667
- `qol` / `qotain`: similarity 0.6640, reliability 0.8393, distance 0.6667

### Family 3

Tokens: `chol`, `chor`, `cthy`, `daiin`, `dol`. Structural medoid: `chol`. Peripheral token(s): `dol`.

Edges:

- `chol` / `cthy`: similarity 0.6832, reliability 0.8970, distance 0.7500
- `chol` / `daiin`: similarity 0.7202, reliability 0.9332, distance 1.0000
- `chor` / `daiin`: similarity 0.6627, reliability 0.9332, distance 1.0000
- `daiin` / `dol`: similarity 0.6672, reliability 0.9018, distance 0.8000

### Family 4

Tokens: `okar`, `otain`. Structural medoid: `okar`. Peripheral token(s): `okar`, `otain`.

Edges:

- `okar` / `otain`: similarity 0.6620, reliability 0.8809, distance 0.6000

### Family 5

Tokens: `or`, `r`, `s`. Structural medoid: `s`. Peripheral token(s): `r`.

Edges:

- `or` / `s`: similarity 0.7588, reliability 0.9332, distance 1.0000
- `r` / `s`: similarity 0.6989, reliability 0.9332, distance 1.0000

## Limits

Observed absence is not proof of a prohibition. Context observations at line boundaries have no neighbor and therefore context totals can be below token counts. Pair rows are statistically dependent because tokens recur across pairs. Control matching is descriptive and does not make pairs independent.
