# Inverse-transposition metric audit (task54)

This audit freezes the first search objective before any Voynich search. The
search receives only a transformed token stream. Doyle and Voynich controls
are calibration/validation inputs, never optimization targets.

## Control design

For each corpus, retain the canonical token sequence and generate Task46
`transposition` controls at widths 2, 4 and 8 (natural and keyed orders, seed
1). The inverse search is run blind on each transformed control. The original
sequence is opened only by the separate `validate` command, after ranking.

## Classification

| metric family | classification | use in objective | reason |
|---|---|---|---|
| token count, vocabulary, frequency spectrum, hapax | transposition-invariant | no | pure token permutation preserves them exactly |
| successor transition concentration (mean Simpson concentration) | transposition-sensitive | yes | adjacency and successor distributions change under column read |
| relation significance (mean normalized successor entropy complement) | transposition-sensitive | yes | measures concentration of token relations without token names |
| repeated bigram rate | transposition-sensitive | yes | direct local sequence structure |
| repeated trigram rate | weakly sensitive | yes, low-risk equal family weight | higher-order repetitions are sparse and can be unstable |
| character/lexical likelihood, edit distance to Doyle/Voynich | unsuitable | no | language/oracle overfitting |
| metrics depending on line count or output formatting | unstable | no | line policy is a boundary control, not a cipher signal |

## Frozen objective

`structural-v1` is the arithmetic mean of the four displayed metrics, with
equal weights. No weights, thresholds, or candidate widths are changed after
looking at Voynich candidates. Ties are resolved by candidate id. The CLI
does not accept an oracle flag in search mode; `validate` reports exact
byte/token-sequence recovery separately.

This is a validation framework, not a decryptor. A high score means only that
the candidate has more concentrated/repeated token adjacency under this
pre-registered structural proxy.
