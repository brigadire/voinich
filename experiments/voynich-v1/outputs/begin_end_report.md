# Directed paired-token candidate report

The analysis covers 39026 token occurrences, 5385 lines, and 1 inferred pages. 539 tokens met the minimum frequency. The main ranking contains 1000 non-local candidate pairs; 100 likely local pairs are separated from it.

> Page boundaries were not present in the corpus. Page-scope results therefore use the entire corpus as one page and should be treated as provisional. Supply a corpus with blank-line, form-feed, or supported page markers for page-level inference.

Tokens containing `?` are excluded from the main ranking by default because they contain uncertain signs. `@NNN;` forms are preserved as ordinary complete tokens. The labels opening and closing candidate describe direction only and do not assign semantics.

## Metrics

For every occurrence of the first token, the analyzer finds the nearest later occurrence of the second token within the same line or page. It reports fixed windows, full-scope coverage, distance summaries, reverse direction, page balance relative to pairs in similar frequency bins, four neutral four-event orders, immediate adjacency, and boundary-preserving permutation significance. Low-count pairs receive the explicit reliability factor `min(count)/(min(count)+20)`. The final score is only a documented sorting aid; all component metrics remain in the YAML output.

## Best overall non-local pairs

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `or` | `aiin` | 0.4221 | 0.4221 | 0.9512 | line | 0.2308 | 0.5949 | 0.1812 | 15.036 | 0.009901 | 0.000 | 40/39 |
| `chol` | `daiin` | 0.4202 | 0.4202 | 0.9518 | line | 0.1899 | 0.7316 | 0.1403 | 6.186 | 0.009901 | 0.000 | 56/36 |
| `sho` | `daiin` | 0.4101 | 0.4101 | 0.8621 | line | 0.2640 | 0.8080 | 0.2557 | 5.709 | 0.009901 | 0.000 | 11/18 |
| `s` | `aiin` | 0.4096 | 0.4096 | 0.9459 | line | 0.2171 | 0.4857 | 0.1775 | 11.792 | 0.009901 | 0.000 | 17/51 |
| `qokeedy` | `chedy` | 0.4027 | 0.4027 | 0.9388 | line | 0.1433 | 0.6971 | 0.0781 | 5.959 | 0.009901 | 0.000 | 40/26 |
| `r` | `aiin` | 0.4018 | 0.4018 | 0.8942 | line | 0.1953 | 0.4911 | 0.1754 | 6.714 | 0.009901 | 0.000 | 25/12 |
| `qokain` | `ol` | 0.3966 | 0.3966 | 0.9331 | line | 0.1505 | 0.6882 | 0.1220 | 5.674 | 0.009901 | 0.000 | 32/29 |
| `qokedy` | `chedy` | 0.3933 | 0.3933 | 0.9324 | line | 0.1558 | 0.8007 | 0.0965 | 5.970 | 0.009901 | 0.000 | 35/35 |
| `ol` | `aiin` | 0.3932 | 0.3932 | 0.9618 | line | 0.1143 | 0.4536 | 0.0310 | 5.027 | 0.009901 | 0.000 | 45/33 |
| `saiin` | `chedy` | 0.3931 | 0.3931 | 0.8639 | line | 0.1732 | 0.5118 | 0.1693 | 6.073 | 0.009901 | 0.000 | 15/12 |

## Best pairs within a line

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `shes` | `dy` | 0.4286 | 0.2644 | 0.4118 | line | 0.4286 | 0.5714 | 0.4249 | 9.604 | 0.009901 | 0.000 | 1/0 |
| `shes` | `y` | 0.4286 | 0.2291 | 0.4118 | line | 0.4286 | 0.5000 | 0.4253 | 7.110 | 0.009901 | 0.000 | 1/0 |
| `qokchol` | `daiin` | 0.4278 | 0.2603 | 0.4286 | line | 0.4667 | 0.8000 | 0.4667 | 5.343 | 0.009901 | 0.000 | 2/0 |
| `qotor` | `daiin` | 0.3679 | 0.2836 | 0.5652 | line | 0.3846 | 0.8846 | 0.3799 | 4.728 | 0.009901 | 0.000 | 1/2 |
| `pol` | `shedy` | 0.3401 | 0.2559 | 0.4872 | line | 0.3684 | 0.6842 | 0.3638 | 6.143 | 0.009901 | 0.000 | 1/1 |
| `sh` | `s` | 0.3324 | 0.2961 | 0.5238 | line | 0.4091 | 0.7273 | 0.4005 | 9.034 | 0.009901 | 0.000 | 1/0 |
| `pol` | `qokaiin` | 0.3158 | 0.2562 | 0.4872 | line | 0.3158 | 0.4211 | 0.3158 | 8.023 | 0.009901 | 0.000 | 2/1 |
| `qokeed` | `chedy` | 0.3077 | 0.2478 | 0.4737 | line | 0.3333 | 0.7222 | 0.3314 | 5.911 | 0.009901 | 0.000 | 1/0 |
| `ty` | `dy` | 0.3008 | 0.2664 | 0.4872 | line | 0.4211 | 0.7368 | 0.4174 | 8.818 | 0.009901 | 0.000 | 1/0 |
| `okees` | `ar` | 0.2895 | 0.2370 | 0.4872 | line | 0.3158 | 0.6316 | 0.3059 | 5.058 | 0.009901 | 0.000 | 0/2 |

## Best pairs within a page

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `qotchy` | `daiin` | 0.8710 | 0.3658 | 0.7561 | line | 0.2742 | 0.9032 | 0.2648 | 5.131 | 0.009901 | 0.000 | 5/4 |
| `qotor` | `daiin` | 0.8462 | 0.2836 | 0.5652 | line | 0.3846 | 0.8846 | 0.3799 | 4.728 | 0.009901 | 0.000 | 1/2 |
| `sho` | `daiin` | 0.8000 | 0.4101 | 0.8621 | line | 0.2640 | 0.8080 | 0.2557 | 5.709 | 0.009901 | 0.000 | 11/18 |
| `chor` | `daiin` | 0.7773 | 0.3123 | 0.9134 | line | 0.1706 | 0.8199 | 0.1376 | 3.458 | 0.009901 | 0.000 | 29/24 |
| `cthy` | `daiin` | 0.7767 | 0.2502 | 0.8374 | line | 0.1845 | 0.8447 | 0.1550 | 2.963 | 0.009901 | 0.000 | 10/17 |
| `qokedy` | `chedy` | 0.7609 | 0.3933 | 0.9324 | line | 0.1558 | 0.8007 | 0.0965 | 5.970 | 0.009901 | 0.000 | 35/35 |
| `qokedy` | `shedy` | 0.7536 | 0.3728 | 0.9324 | line | 0.1232 | 0.7862 | 0.0264 | 5.586 | 0.009901 | 0.000 | 29/42 |
| `dor` | `daiin` | 0.7500 | 0.3629 | 0.7727 | line | 0.2647 | 0.7794 | 0.2517 | 5.040 | 0.009901 | 0.000 | 7/7 |
| `shor` | `daiin` | 0.7500 | 0.2820 | 0.8276 | line | 0.1771 | 0.8021 | 0.1735 | 2.976 | 0.0198 | 0.000 | 15/6 |
| `shy` | `daiin` | 0.7449 | 0.2332 | 0.8305 | line | 0.1735 | 0.8061 | 0.1652 | 2.505 | 0.0198 | 0.000 | 9/10 |

## Strongest directionality

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `qokchol` | `daiin` | 0.4667 | 0.2603 | 0.4286 | line | 0.4667 | 0.8000 | 0.4667 | 5.343 | 0.009901 | 0.000 | 2/0 |
| `shes` | `y` | 0.4253 | 0.2291 | 0.4118 | line | 0.4286 | 0.5000 | 0.4253 | 7.110 | 0.009901 | 0.000 | 1/0 |
| `shes` | `dy` | 0.4249 | 0.2644 | 0.4118 | line | 0.4286 | 0.5714 | 0.4249 | 9.604 | 0.009901 | 0.000 | 1/0 |
| `ty` | `dy` | 0.4174 | 0.2664 | 0.4872 | line | 0.4211 | 0.7368 | 0.4174 | 8.818 | 0.009901 | 0.000 | 1/0 |
| `sh` | `s` | 0.4005 | 0.2961 | 0.5238 | line | 0.4091 | 0.7273 | 0.4005 | 9.034 | 0.009901 | 0.000 | 1/0 |
| `qotor` | `daiin` | 0.3799 | 0.2836 | 0.5652 | line | 0.3846 | 0.8846 | 0.3799 | 4.728 | 0.009901 | 0.000 | 1/2 |
| `pol` | `shedy` | 0.3638 | 0.2559 | 0.4872 | line | 0.3684 | 0.6842 | 0.3638 | 6.143 | 0.009901 | 0.000 | 1/1 |
| `okeeo` | `l` | 0.3447 | 0.2355 | 0.5000 | line | 0.3500 | 0.5500 | 0.3447 | 9.887 | 0.009901 | 0.000 | 0/7 |
| `qokeed` | `chedy` | 0.3314 | 0.2478 | 0.4737 | line | 0.3333 | 0.7222 | 0.3314 | 5.911 | 0.009901 | 0.000 | 1/0 |
| `pol` | `qokaiin` | 0.3158 | 0.2562 | 0.4872 | line | 0.3158 | 0.4211 | 0.3158 | 8.023 | 0.009901 | 0.000 | 2/1 |

## Most expressed page balance

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `aiiin` | `otal` | 0.0000 | 0.2979 | 0.7015 | line | 0.1277 | 0.3830 | 0.1131 | 6.057 | 0.009901 | 0.000 | 7/5 |
| `aiin` | `al` | 0.0000 | 0.3584 | 0.9288 | line | 0.0694 | 0.3472 | -0.0493 | 5.089 | 0.009901 | 0.000 | 31/39 |
| `aiin` | `am` | 0.0000 | 0.3408 | 0.8113 | line | 0.0397 | 0.1548 | 0.0164 | 6.035 | 0.009901 | 0.000 | 14/9 |
| `aiin` | `ches` | 0.0000 | 0.2391 | 0.6774 | line | 0.0179 | 0.0615 | -0.2202 | 3.987 | 0.009901 | 0.000 | 4/1 |
| `aiin` | `okal` | 0.0000 | 0.3223 | 0.8864 | line | 0.0456 | 0.2222 | -0.0377 | 4.561 | 0.009901 | 0.000 | 23/16 |
| `aiin` | `otaiin` | 0.0000 | 0.3144 | 0.8857 | line | 0.0476 | 0.2579 | -0.0685 | 4.472 | 0.009901 | 0.000 | 21/17 |
| `aiin` | `otar` | 0.0000 | 0.2427 | 0.8837 | line | 0.0397 | 0.2202 | -0.0853 | 3.247 | 0.009901 | 0.000 | 21/20 |
| `aiin` | `oteey` | 0.0000 | 0.2393 | 0.8788 | line | 0.0357 | 0.2123 | -0.0608 | 3.070 | 0.0198 | 0.000 | 23/14 |
| `aiir` | `al` | 0.0000 | 0.2299 | 0.5918 | line | 0.1724 | 0.4483 | 0.1609 | 4.337 | 0.009901 | 0.000 | 3/5 |
| `aiir` | `ches` | 0.0000 | 0.2299 | 0.5918 | line | 0.0690 | 0.2069 | 0.0690 | 4.614 | 0.009901 | 0.000 | 4/2 |

## Pairs with nesting-like order contrast

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `chain` | `kain` | 1.0000 | 0.2505 | 0.4872 | line | 0.1053 | 0.1053 | 0.1053 | 7.077 | 0.009901 | 0.000 | 3/0 |
| `shes` | `dy` | 1.0000 | 0.2644 | 0.4118 | line | 0.4286 | 0.5714 | 0.4249 | 9.604 | 0.009901 | 0.000 | 1/0 |
| `ydaiin` | `okey` | 1.0000 | 0.2578 | 0.4737 | line | 0.1667 | 0.1667 | 0.1667 | 7.747 | 0.009901 | 0.000 | 3/0 |
| `sal` | `opchdy` | 0.8000 | 0.2387 | 0.5000 | line | 0.0408 | 0.0612 | 0.0408 | 5.560 | 0.009901 | 0.000 | 4/0 |
| `checthy` | `qokeeey` | 0.7500 | 0.2673 | 0.5833 | line | 0.0667 | 0.1000 | 0.0667 | 6.674 | 0.009901 | 0.000 | 3/0 |
| `dchor` | `ytedy` | 0.7500 | 0.2392 | 0.5455 | line | 0.0370 | 0.0370 | 0.0370 | 4.359 | 0.05941 | 0.000 | 3/0 |
| `ckhey` | `chody` | 0.6667 | 0.2820 | 0.5833 | line | 0.1071 | 0.2857 | 0.1071 | 5.400 | 0.009901 | 0.000 | 6/0 |
| `olor` | `dchol` | 0.6667 | 0.2393 | 0.5455 | line | 0.0333 | 0.0333 | 0.0333 | 5.686 | 0.0396 | 0.000 | 2/0 |
| `qokchol` | `daiin` | 0.6667 | 0.2603 | 0.4286 | line | 0.4667 | 0.8000 | 0.4667 | 5.343 | 0.009901 | 0.000 | 2/0 |
| `qokeor` | `daiiin` | 0.6667 | 0.2410 | 0.5122 | line | 0.0476 | 0.0476 | 0.0476 | 7.000 | 0.0297 | 0.000 | 2/0 |

## Likely local pairs (reported separately)

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `aiir` | `aly` | 1.0000 | 0.1703 | 0.5652 | line | 0.0345 | 0.0345 | 0.0345 | 3.391 | 0.08911 | 0.000 | 0/0 |
| `aral` | `om` | 1.0000 | 0.2056 | 0.5000 | line | 0.0500 | 0.0500 | 0.0500 | 5.686 | 0.0396 | 0.000 | 2/1 |
| `checthy` | `chety` | 1.0000 | 0.1773 | 0.5238 | line | 0.0333 | 0.0333 | 0.0333 | 4.359 | 0.05941 | 0.000 | 0/0 |
| `ched` | `okchey` | 1.0000 | 0.1779 | 0.4737 | line | 0.0556 | 0.0556 | 0.0556 | 5.686 | 0.0396 | 0.000 | 0/0 |
| `chedain` | `qoeedy` | 1.0000 | 0.1827 | 0.4872 | line | 0.0526 | 0.0526 | 0.0526 | 7.000 | 0.0297 | 0.000 | 0/0 |
| `cheeor` | `ary` | 1.0000 | 0.1862 | 0.4444 | line | 0.0625 | 0.0625 | 0.0625 | 4.899 | 0.0495 | 0.000 | 0/0 |
| `cheeor` | `olar` | 1.0000 | 0.1736 | 0.4444 | line | 0.0625 | 0.0625 | 0.0625 | 4.899 | 0.0495 | 0.000 | 0/0 |
| `cheos` | `qokeeey` | 1.0000 | 0.1820 | 0.5833 | line | 0.0270 | 0.0270 | 0.0270 | 3.958 | 0.06931 | 0.000 | 0/0 |
| `ches` | `os` | 1.0000 | 0.2354 | 0.6226 | line | 0.0476 | 0.0476 | -0.0130 | 5.019 | 0.009901 | 0.000 | 3/2 |
| `chkaiin` | `sh` | 1.0000 | 0.1733 | 0.4737 | line | 0.0556 | 0.0556 | 0.0556 | 4.899 | 0.0495 | 0.000 | 0/0 |

## Interpretation limits

A small permutation p-value indicates stability against the selected constrained shuffle, not a grammatical construction. Frequent-token effects, transcription uncertainty, multiple testing, page-boundary quality, and corpus heterogeneity remain possible explanations. Candidates are therefore targets for follow-up inspection rather than identified operators.
