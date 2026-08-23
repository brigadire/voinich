# Structural pair decomposition

Structural similarity is reproduced unchanged from the existing pair dataset. All statements below are formal corpus descriptions; no token meaning is inferred. Context similarities and differences use full distributions, while tables are display-limited. Entropy uses natural logarithms and effective vocabulary is `exp(entropy)`.

## `x014952` / `x030834`

Structural similarity: 0.6859; reliability: 0.9008; normalized graphemic distance: 0.7143; counts: 149/155.

| Component | Similarity | Reliability |
|---|---:|---:|
| Position | 0.9685 | 0.9920 |
| Left context | 0.6881 | 0.8543 |
| Right context | 0.4011 | 0.8563 |

- Primary component: positional agreement (0.968).
- Similarity is multidimensional: predecessor-distribution overlap also contributes 0.688.
- Largest left-context difference: x015012 is more frequent for x014952 (absolute probability difference 0.059).

Position summaries (A/B): line-start 0.0470/0.0452, line-end 0.0604/0.0645, mean 5.805/5.877, median 6.000/6.000. Position JS similarity: 0.9685.

### Common predecessors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x015012 | 0.1197 | 0.0608 | +0.0589 |
| x015013 | 0.0493 | 0.0541 | -0.0048 |
| x015014 | 0.0493 | 0.0473 | +0.0020 |
| x013056 | 0.0282 | 0.0338 | -0.0056 |
| x015015 | 0.0282 | 0.0338 | -0.0056 |
| x028464 | 0.0211 | 0.0405 | -0.0194 |
| x031243 | 0.0282 | 0.0203 | +0.0079 |
| x028466 | 0.0141 | 0.0203 | -0.0062 |
| x015016 | 0.0282 | 0.0135 | +0.0147 |
| x028469 | 0.0141 | 0.0135 | +0.0006 |
| x013057 | 0.0070 | 0.0338 | -0.0267 |
| x013059 | 0.0070 | 0.0338 | -0.0267 |

### Largest predecessor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x015012 | 0.1197 | 0.0608 | +0.0589 |
| x013872 | 0.0000 | 0.0338 | -0.0338 |
| x031149 | 0.0282 | 0.0000 | +0.0282 |
| x013873 | 0.0000 | 0.0270 | -0.0270 |
| x013057 | 0.0070 | 0.0338 | -0.0267 |
| x013059 | 0.0070 | 0.0338 | -0.0267 |
| x028467 | 0.0211 | 0.0000 | +0.0211 |
| x028581 | 0.0211 | 0.0000 | +0.0211 |
| x013874 | 0.0000 | 0.0203 | -0.0203 |
| x028464 | 0.0211 | 0.0405 | -0.0194 |
| x015016 | 0.0282 | 0.0135 | +0.0147 |
| x028388 | 0.0211 | 0.0068 | +0.0144 |

### Common successors

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x028392 | 0.0214 | 0.0207 | +0.0007 |
| x000079 | 0.0143 | 0.0483 | -0.0340 |
| x000080 | 0.0143 | 0.0207 | -0.0064 |
| x000082 | 0.0143 | 0.0207 | -0.0064 |
| x000081 | 0.0286 | 0.0138 | +0.0148 |
| x018697 | 0.0143 | 0.0138 | +0.0005 |
| x028393 | 0.0214 | 0.0138 | +0.0076 |
| x028395 | 0.0214 | 0.0138 | +0.0076 |
| x000906 | 0.0071 | 0.0138 | -0.0067 |
| x014155 | 0.0071 | 0.0207 | -0.0135 |
| x015012 | 0.0071 | 0.0138 | -0.0067 |
| x000078 | 0.0214 | 0.0069 | +0.0145 |

### Largest successor differences

| Token | P(A) | P(B) | A−B |
|---|---:|---:|---:|
| x000079 | 0.0143 | 0.0483 | -0.0340 |
| x014154 | 0.0000 | 0.0207 | -0.0207 |
| x000081 | 0.0286 | 0.0138 | +0.0148 |
| x000078 | 0.0214 | 0.0069 | +0.0145 |
| x015015 | 0.0143 | 0.0000 | +0.0143 |
| x028396 | 0.0143 | 0.0000 | +0.0143 |
| x028464 | 0.0143 | 0.0000 | +0.0143 |
| x000083 | 0.0000 | 0.0138 | -0.0138 |
| x000816 | 0.0000 | 0.0138 | -0.0138 |
| x013058 | 0.0000 | 0.0138 | -0.0138 |
| x026874 | 0.0000 | 0.0138 | -0.0138 |
| x014155 | 0.0071 | 0.0207 | -0.0135 |

Context diagnostics: predecessor Jaccard 0.1586, JS 0.4685, entropy A/B 4.078/4.113, effective vocabulary A/B 59.03/61.14; successor Jaccard 0.1274, JS 0.2994, entropy A/B 4.719/4.674, effective vocabulary A/B 112.03/107.08.

Shared rare observed contexts (at least two observations per token and probability at most 0.02 on both sides): left `x028469`; right `x018697`.

Shared unobserved high-frequency contexts (descriptive absence only): left `x028392`, `x028393`, `x028394`, `x001068`, `x028902`, `x028395`, `x019044`, `x000078`, `x001069`, `x028903`, `x019045`, `x000079`; right `x001068`, `x013872`, `x019044`, `x013873`, `x001069`, `x001070`, `x019046`, `x013056`, `x001071`, `x015013`, `x030834`, `x014952`.

## Negative controls

Controls match unordered log-counts, normalized graphemic distance, and reliability, while favoring structural similarity near the full-corpus median. They are decomposed with exactly the target metrics.

| Target | Control | Structural | Reliability | Distance | Match cost |
|---|---|---:|---:|---:|---:|
| x014952/x030834 | x015014/x023779 | 0.2709 | 0.7840 | 0.7143 | 1.6351 |
| x014952/x030834 | x015013/x023779 | 0.2797 | 0.7990 | 0.7143 | 1.6448 |
| x014952/x030834 | x019047/x023779 | 0.2933 | 0.7985 | 0.7143 | 1.7029 |

## Family decomposition

A family is a connected component; only listed edges define direct structural-distant links. Complete matrices, including non-edge pairs, are in `family_decomposition.yaml`.

### Family 1

Tokens: `x014952`, `x030834`. Structural medoid: `x014952`. Peripheral token(s): `x014952`, `x030834`.

Edges:

- `x014952` / `x030834`: similarity 0.6859, reliability 0.9008, distance 0.7143

## Limits

Observed absence is not proof of a prohibition. Context observations at line boundaries have no neighbor and therefore context totals can be below token counts. Pair rows are statistically dependent because tokens recur across pairs. Control matching is descriptive and does not make pairs independent.
