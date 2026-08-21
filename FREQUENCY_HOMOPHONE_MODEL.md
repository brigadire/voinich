# Frequency-dependent homophone allocation (frequency-v1)

## Motivation

`frequency-v1` tests the general historical-cryptographic idea that common
plaintext units can receive more cipher alternatives, while rare units
receive fewer. It is an allocation model only; occurrence selection remains
an independent `uniform` or `triangular-v1` selection model.

## Frozen definition

Let `V` be the number of plaintext types. Sort types by decreasing raw
occurrence count, breaking equal-frequency ties lexicographically. For the
one-based rank `r` (1 is the most frequent), define:

```
q(r) = (V-r)/(V-1)                 if V > 1, otherwise 0
H(r) = 1 + floor((Hmax-1) * q(r))
```

Thus the most frequent type gets `Hmax`, the least frequent type gets `1`,
and every allocation is in `[1,Hmax]`. This rank-quantile definition is
independent of corpus size and was fixed before any Doyle/Voynich comparison.

## Example

For frequencies `alpha=10`, `beta=3`, `gamma=1`, `delta=1` and `Hmax=4`,
the sorted allocation is alpha=4, beta=3, delta=2, gamma=1 (lexical tie
break). A hapax can receive more than one homophone only when its rank is not
the final rank; this follows from the rank rule and is not a special case.

## Invariants, diagnostics, and limitations

The transformation preserves token count and maps each occurrence to exactly
one opaque cipher token. Allocation is monotone in frequency, bounded, and
reversible through the existing mapping. Each run writes
`<output>.homophone_allocation.tsv` with token, raw frequency, rank, quantile,
and allocated H. The CLI reports allocated/observed homophones, H
distributions, and hapax types with `H>1`.

Frequency uses whitespace-tokenized raw types. It does not model a particular
historical cipher and does not tune parameters to Voynich metrics. Selection
probability is independent of allocation.

## Usage and versioning

```sh
go run ./corpus-transform -corpus INPUT -method homophonic \
  -homophone-model frequency -homophones 8 \
  -homophone-selection uniform -seed 1 -output OUTPUT
```

`frequency` is `homophonic-frequency-v1`; `fixed` remains
`homophonic-global-v1`. Selection versions are `uniform-v1` and
`triangular-v1`. Changing the formula requires a new model version and
artifact series.
