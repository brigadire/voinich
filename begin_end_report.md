# Directed paired-token candidate report

The analysis covers 38887 token occurrences, 5371 lines, and 1 inferred pages. 537 tokens met the minimum frequency. The main ranking contains 1000 non-local candidate pairs; 100 likely local pairs are separated from it.

> Page boundaries were not present in the corpus. Page-scope results therefore use the entire corpus as one page and should be treated as provisional. Supply a corpus with blank-line, form-feed, or supported page markers for page-level inference.

Tokens containing `?` are excluded from the main ranking by default because they contain uncertain signs. `@NNN;` forms are preserved as ordinary complete tokens. The labels opening and closing candidate describe direction only and do not assign semantics.

## Metrics

For every occurrence of the first token, the analyzer finds the nearest later occurrence of the second token within the same line or page. It reports fixed windows, full-scope coverage, distance summaries, reverse direction, page balance relative to pairs in similar frequency bins, four neutral four-event orders, immediate adjacency, and boundary-preserving permutation significance. Low-count pairs receive the explicit reliability factor `min(count)/(min(count)+20)`. The final score is only a documented sorting aid; all component metrics remain in the YAML output.

## Best overall non-local pairs

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `or` | `aiin` | 0.4224 | 0.4224 | 0.9510 | line | 0.2320 | 0.5979 | 0.1824 | 12.921 | 0.009901 | 0.000 | 40/39 |
| `chol` | `daiin` | 0.4198 | 0.4198 | 0.9518 | line | 0.1899 | 0.7316 | 0.1403 | 6.531 | 0.009901 | 0.000 | 55/36 |
| `sho` | `daiin` | 0.4101 | 0.4101 | 0.8621 | line | 0.2640 | 0.8080 | 0.2557 | 5.893 | 0.009901 | 0.000 | 11/18 |
| `s` | `aiin` | 0.4096 | 0.4096 | 0.9459 | line | 0.2171 | 0.4857 | 0.1775 | 11.764 | 0.009901 | 0.000 | 17/51 |
| `qokeedy` | `chedy` | 0.4047 | 0.4047 | 0.9387 | line | 0.1438 | 0.6961 | 0.0803 | 7.165 | 0.009901 | 0.000 | 40/24 |
| `r` | `aiin` | 0.4018 | 0.4018 | 0.8936 | line | 0.1964 | 0.4940 | 0.1766 | 7.451 | 0.009901 | 0.000 | 25/12 |
| `ol` | `aiin` | 0.3934 | 0.3934 | 0.9618 | line | 0.1149 | 0.4560 | 0.0316 | 5.086 | 0.009901 | 0.000 | 45/33 |
| `qokedy` | `chedy` | 0.3932 | 0.3932 | 0.9324 | line | 0.1558 | 0.8007 | 0.0963 | 7.516 | 0.009901 | 0.000 | 35/35 |
| `ar` | `aiin` | 0.3925 | 0.3925 | 0.9526 | line | 0.1517 | 0.5920 | 0.0803 | 7.942 | 0.009901 | 0.000 | 47/48 |
| `o` | `aiin` | 0.3898 | 0.3898 | 0.9083 | line | 0.1515 | 0.4343 | 0.1218 | 5.084 | 0.009901 | 0.000 | 20/28 |

## Best pairs within a line

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `shes` | `dy` | 0.4286 | 0.2644 | 0.4118 | line | 0.4286 | 0.5714 | 0.4249 | 8.769 | 0.009901 | 0.000 | 1/0 |
| `shes` | `y` | 0.4286 | 0.2291 | 0.4118 | line | 0.4286 | 0.5000 | 0.4253 | 7.359 | 0.009901 | 0.000 | 1/0 |
| `qokchol` | `daiin` | 0.4278 | 0.2598 | 0.4286 | line | 0.4667 | 0.8000 | 0.4667 | 4.980 | 0.009901 | 0.000 | 2/0 |
| `qotor` | `daiin` | 0.3679 | 0.2929 | 0.5652 | line | 0.3846 | 0.8846 | 0.3799 | 5.067 | 0.009901 | 0.000 | 1/2 |
| `pol` | `shedy` | 0.3401 | 0.2559 | 0.4872 | line | 0.3684 | 0.6842 | 0.3638 | 6.778 | 0.009901 | 0.000 | 1/1 |
| `sh` | `s` | 0.3324 | 0.2961 | 0.5238 | line | 0.4091 | 0.7273 | 0.4005 | 9.615 | 0.009901 | 0.000 | 1/0 |
| `pol` | `qokaiin` | 0.3158 | 0.2563 | 0.4872 | line | 0.3158 | 0.4211 | 0.3158 | 6.602 | 0.009901 | 0.000 | 2/1 |
| `qokeed` | `chedy` | 0.3077 | 0.2333 | 0.4737 | line | 0.3333 | 0.7222 | 0.3313 | 4.493 | 0.009901 | 0.000 | 1/0 |
| `ty` | `dy` | 0.3008 | 0.2664 | 0.4872 | line | 0.4211 | 0.7368 | 0.4174 | 11.673 | 0.009901 | 0.000 | 1/0 |
| `okees` | `ar` | 0.2895 | 0.2371 | 0.4872 | line | 0.3158 | 0.6316 | 0.3058 | 5.731 | 0.009901 | 0.000 | 0/2 |

## Best pairs within a page

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `qotchy` | `daiin` | 0.8710 | 0.3636 | 0.7561 | line | 0.2742 | 0.9032 | 0.2647 | 4.950 | 0.009901 | 0.000 | 5/4 |
| `qotor` | `daiin` | 0.8462 | 0.2929 | 0.5652 | line | 0.3846 | 0.8846 | 0.3799 | 5.067 | 0.009901 | 0.000 | 1/2 |
| `sho` | `daiin` | 0.8000 | 0.4101 | 0.8621 | line | 0.2640 | 0.8080 | 0.2557 | 5.893 | 0.009901 | 0.000 | 11/18 |
| `chor` | `daiin` | 0.7773 | 0.3085 | 0.9134 | line | 0.1706 | 0.8199 | 0.1376 | 3.388 | 0.009901 | 0.000 | 29/24 |
| `cthy` | `daiin` | 0.7767 | 0.2623 | 0.8374 | line | 0.1845 | 0.8447 | 0.1550 | 3.204 | 0.009901 | 0.000 | 10/17 |
| `qokedy` | `chedy` | 0.7609 | 0.3932 | 0.9324 | line | 0.1558 | 0.8007 | 0.0963 | 7.516 | 0.009901 | 0.000 | 35/35 |
| `qokedy` | `shedy` | 0.7536 | 0.3727 | 0.9324 | line | 0.1232 | 0.7862 | 0.0262 | 5.081 | 0.009901 | 0.000 | 29/42 |
| `shy` | `daiin` | 0.7526 | 0.2506 | 0.8291 | line | 0.1753 | 0.8144 | 0.1682 | 2.842 | 0.0198 | 0.000 | 9/10 |
| `dor` | `daiin` | 0.7500 | 0.3337 | 0.7727 | line | 0.2647 | 0.7794 | 0.2517 | 4.370 | 0.009901 | 0.000 | 7/7 |
| `shor` | `daiin` | 0.7500 | 0.2694 | 0.8276 | line | 0.1771 | 0.8021 | 0.1735 | 2.722 | 0.009901 | 0.000 | 15/6 |

## Strongest directionality

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `qokchol` | `daiin` | 0.4667 | 0.2598 | 0.4286 | line | 0.4667 | 0.8000 | 0.4667 | 4.980 | 0.009901 | 0.000 | 2/0 |
| `shes` | `y` | 0.4253 | 0.2291 | 0.4118 | line | 0.4286 | 0.5000 | 0.4253 | 7.359 | 0.009901 | 0.000 | 1/0 |
| `shes` | `dy` | 0.4249 | 0.2644 | 0.4118 | line | 0.4286 | 0.5714 | 0.4249 | 8.769 | 0.009901 | 0.000 | 1/0 |
| `ty` | `dy` | 0.4174 | 0.2664 | 0.4872 | line | 0.4211 | 0.7368 | 0.4174 | 11.673 | 0.009901 | 0.000 | 1/0 |
| `sh` | `s` | 0.4005 | 0.2961 | 0.5238 | line | 0.4091 | 0.7273 | 0.4005 | 9.615 | 0.009901 | 0.000 | 1/0 |
| `qotor` | `daiin` | 0.3799 | 0.2929 | 0.5652 | line | 0.3846 | 0.8846 | 0.3799 | 5.067 | 0.009901 | 0.000 | 1/2 |
| `pol` | `shedy` | 0.3638 | 0.2559 | 0.4872 | line | 0.3684 | 0.6842 | 0.3638 | 6.778 | 0.009901 | 0.000 | 1/1 |
| `okeeo` | `l` | 0.3446 | 0.2355 | 0.5000 | line | 0.3500 | 0.5500 | 0.3446 | 8.406 | 0.009901 | 0.000 | 0/7 |
| `qokeed` | `chedy` | 0.3313 | 0.2333 | 0.4737 | line | 0.3333 | 0.7222 | 0.3313 | 4.493 | 0.009901 | 0.000 | 1/0 |
| `pol` | `qokaiin` | 0.3158 | 0.2563 | 0.4872 | line | 0.3158 | 0.4211 | 0.3158 | 6.602 | 0.009901 | 0.000 | 2/1 |

## Most expressed page balance

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `aiiin` | `ary` | 0.0000 | 0.2506 | 0.5556 | line | 0.0426 | 0.0638 | 0.0426 | 6.674 | 0.009901 | 0.000 | 5/1 |
| `aiiin` | `otal` | 0.0000 | 0.2979 | 0.7015 | line | 0.1277 | 0.3830 | 0.1131 | 6.188 | 0.009901 | 0.000 | 7/5 |
| `aiin` | `al` | 0.0000 | 0.3581 | 0.9278 | line | 0.0694 | 0.3472 | -0.0512 | 5.418 | 0.009901 | 0.000 | 31/38 |
| `aiin` | `am` | 0.0000 | 0.3408 | 0.8113 | line | 0.0397 | 0.1548 | 0.0164 | 5.925 | 0.009901 | 0.000 | 14/9 |
| `aiin` | `ches` | 0.0000 | 0.2273 | 0.6774 | line | 0.0179 | 0.0615 | -0.2202 | 3.696 | 0.009901 | 0.000 | 4/1 |
| `aiin` | `okal` | 0.0000 | 0.3100 | 0.8857 | line | 0.0456 | 0.2222 | -0.0382 | 4.336 | 0.009901 | 0.000 | 23/16 |
| `aiin` | `otaiin` | 0.0000 | 0.3109 | 0.8857 | line | 0.0476 | 0.2579 | -0.0685 | 4.407 | 0.009901 | 0.000 | 21/17 |
| `aiin` | `otam` | 0.0000 | 0.2341 | 0.7059 | line | 0.0179 | 0.1071 | -0.0863 | 3.477 | 0.009901 | 0.000 | 7/2 |
| `aiin` | `otar` | 0.0000 | 0.2952 | 0.8837 | line | 0.0397 | 0.2202 | -0.0853 | 4.237 | 0.009901 | 0.000 | 21/20 |
| `aiin` | `oteey` | 0.0000 | 0.2460 | 0.8780 | line | 0.0357 | 0.2123 | -0.0615 | 3.199 | 0.0198 | 0.000 | 23/14 |

## Pairs with nesting-like order contrast

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `chain` | `kain` | 1.0000 | 0.2505 | 0.4872 | line | 0.1053 | 0.1053 | 0.1053 | 5.126 | 0.0198 | 0.000 | 3/0 |
| `shes` | `dy` | 1.0000 | 0.2644 | 0.4118 | line | 0.4286 | 0.5714 | 0.4249 | 8.769 | 0.009901 | 0.000 | 1/0 |
| `ydaiin` | `okey` | 1.0000 | 0.2578 | 0.4737 | line | 0.1667 | 0.1667 | 0.1667 | 6.888 | 0.009901 | 0.000 | 3/0 |
| `sal` | `opchdy` | 0.8000 | 0.2387 | 0.5000 | line | 0.0408 | 0.0612 | 0.0408 | 5.360 | 0.009901 | 0.000 | 4/0 |
| `checthy` | `qokeeey` | 0.7500 | 0.2673 | 0.5833 | line | 0.0667 | 0.1000 | 0.0667 | 8.947 | 0.009901 | 0.000 | 3/0 |
| `ckhey` | `chody` | 0.6667 | 0.2820 | 0.5833 | line | 0.1071 | 0.2857 | 0.1071 | 5.384 | 0.009901 | 0.000 | 6/0 |
| `qokchol` | `daiin` | 0.6667 | 0.2598 | 0.4286 | line | 0.4667 | 0.8000 | 0.4667 | 4.980 | 0.009901 | 0.000 | 2/0 |
| `tchey` | `tal` | 0.6667 | 0.2426 | 0.5238 | line | 0.0455 | 0.0909 | 0.0455 | 5.686 | 0.0396 | 0.000 | 2/0 |
| `oteos` | `chdar` | 0.6000 | 0.2574 | 0.5455 | line | 0.0769 | 0.1538 | 0.0769 | 7.077 | 0.009901 | 0.000 | 3/0 |
| `oteos` | `tal` | 0.6000 | 0.2657 | 0.5652 | line | 0.0769 | 0.1538 | 0.0769 | 6.040 | 0.009901 | 0.000 | 3/0 |

## Likely local pairs (reported separately)

| opening candidate | closing candidate | section metric | score | reliability | scope | P(line) | P(page) | directionality | ranking z | p | balance sd | AABB/ABAB |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|
| `aral` | `om` | 1.0000 | 0.2056 | 0.5000 | line | 0.0500 | 0.0500 | 0.0500 | 7.000 | 0.0297 | 0.000 | 2/1 |
| `aral` | `opchdy` | 1.0000 | 0.1838 | 0.5000 | line | 0.0500 | 0.0500 | 0.0000 | 7.000 | 0.0297 | 0.000 | 0/0 |
| `chckhey` | `ldy` | 1.0000 | 0.1667 | 0.5556 | line | 0.0345 | 0.0345 | 0.0345 | 3.180 | 0.09901 | 0.000 | 0/0 |
| `chdal` | `olar` | 1.0000 | 0.1762 | 0.4595 | line | 0.0526 | 0.0526 | 0.0526 | 4.899 | 0.0495 | 0.000 | 0/0 |
| `chedain` | `otair` | 1.0000 | 0.1825 | 0.4872 | line | 0.0526 | 0.0526 | 0.0526 | 4.899 | 0.0495 | 0.000 | 0/0 |
| `cheeor` | `ary` | 1.0000 | 0.1611 | 0.4444 | line | 0.0625 | 0.0625 | 0.0625 | 3.958 | 0.06931 | 0.000 | 0/0 |
| `cheeor` | `olar` | 1.0000 | 0.1736 | 0.4444 | line | 0.0625 | 0.0625 | 0.0625 | 4.899 | 0.0495 | 0.000 | 0/0 |
| `cheos` | `ytedy` | 1.0000 | 0.1617 | 0.5455 | line | 0.0270 | 0.0270 | 0.0270 | 3.645 | 0.07921 | 0.000 | 0/0 |
| `ches` | `os` | 1.0000 | 0.2071 | 0.6226 | line | 0.0476 | 0.0476 | -0.0130 | 4.243 | 0.0198 | 0.000 | 3/2 |
| `ches` | `otaly` | 1.0000 | 0.1924 | 0.5122 | line | 0.0238 | 0.0238 | 0.0238 | 5.686 | 0.0396 | 0.000 | 3/4 |

## Interpretation limits

A small permutation p-value indicates stability against the selected constrained shuffle, not a grammatical construction. Frequent-token effects, transcription uncertainty, multiple testing, page-boundary quality, and corpus heterogeneity remain possible explanations. Candidates are therefore targets for follow-up inspection rather than identified operators.
